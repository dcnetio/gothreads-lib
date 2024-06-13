// Package net implements the network layer for go-threads.
// Nodes exchange messages with each other via gRPC, and the format is defined under /pb.
package net

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	gonet "net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dcnetio/gothreads-lib/broadcast"
	"github.com/dcnetio/gothreads-lib/cbor"
	"github.com/dcnetio/gothreads-lib/core/app"
	lstore "github.com/dcnetio/gothreads-lib/core/logstore"
	core "github.com/dcnetio/gothreads-lib/core/net"
	"github.com/dcnetio/gothreads-lib/core/thread"
	sym "github.com/dcnetio/gothreads-lib/crypto/symmetric"
	pb "github.com/dcnetio/gothreads-lib/net/pb"
	"github.com/dcnetio/gothreads-lib/net/queue"
	"github.com/dcnetio/gothreads-lib/net/util"
	tu "github.com/dcnetio/gothreads-lib/util"
	bs "github.com/ipfs/boxo/blockstore"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
	gostream "github.com/libp2p/go-libp2p-gostream"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	pstore "github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
)

var (
	log = logging.Logger("net")

	// MaxThreadsExchanged is the maximum number of threads for the single edge exchange.
	MaxThreadsExchanged = 2 // 2 is the minimum number of threads for the exchange to work for app

	// ExchangeCompressionTimeout is the maximum duration of collecting threads for the exchange edges request.
	ExchangeCompressionTimeout = PullTimeout / 2

	// QueuePollInterval is the polling interval for the call queue.
	QueuePollInterval = time.Millisecond * 500

	// EventBusCapacity is the buffer size of local event bus listeners.
	EventBusCapacity = 1

	// notifyTimeout is the duration to wait for a subscriber to read a new record.
	notifyTimeout = time.Second * 5

	// tokenChallengeBytes is the byte length of token challenges.
	tokenChallengeBytes = 32

	// tokenChallengeTimeout is the duration of time given to an identity to complete a token challenge.
	tokenChallengeTimeout = time.Minute
)

const (
	callPriorityLow  = 1
	callPriorityHigh = 3
)

var (
	_ util.SemaphoreKey = (*semaThreadUpdate)(nil)
)

// semaphore protecting thread info updates
type semaThreadUpdate thread.ID

func (t semaThreadUpdate) Key() string {
	return "tu:" + string(t)
}

// net is an implementation of app.Net.
type net struct {
	conf Config

	format.DAGService
	host   host.Host
	bstore bs.Blockstore

	store lstore.Logstore

	rpc      *grpc.Server
	server   *server
	listener gonet.Listener
	bus      *broadcast.Broadcaster

	connectors map[thread.ID]*app.Connector
	connLock   sync.RWMutex

	semaphores      *util.SemaphorePool
	queueGetLogs    queue.CallQueue
	queueGetRecords queue.CallQueue
	threadExternal  core.ThreadExternal

	ctx    context.Context
	cancel context.CancelFunc
}

// Config is used to specify thread instance options.
type Config struct {
	NetPullingLimit           uint
	NetPullingStartAfter      time.Duration
	NetPullingInitialInterval time.Duration
	NetPullingInterval        time.Duration
	NoNetPulling              bool
	NoExchangeEdgesMigration  bool
	PubSub                    bool
	Debug                     bool
}

func (c Config) Validate() error {
	if c.NetPullingLimit <= 0 {
		return errors.New("NetPullingLimit must be greater than zero")
	}
	if c.NetPullingStartAfter <= 0 {
		return errors.New("NetPullingStartAfter must be greater than zero")
	}
	if c.NetPullingInitialInterval <= 0 {
		return errors.New("NetPullingInitialInterval must be greater than zero")
	}
	if c.NetPullingInterval <= 0 {
		return errors.New("NetPullingInterval must be greater than zero")
	}
	return nil
}

// NewNetwork creates an instance of net from the given host and thread store.
func NewNetwork(
	ctx context.Context,
	h host.Host,
	bstore bs.Blockstore,
	ds format.DAGService,
	ls lstore.Logstore,
	conf Config,
	serverOptions []grpc.ServerOption,
	dialOptions []grpc.DialOption,
	threadExtern core.ThreadExternal,
) (app.Net, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %v", err)
	}

	if err := tu.SetLogLevels(map[string]logging.LogLevel{
		"net":      tu.LevelFromDebugFlag(conf.Debug),
		"logstore": tu.LevelFromDebugFlag(conf.Debug),
	}); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	n := &net{
		conf:            conf,
		DAGService:      ds,
		host:            h,
		bstore:          bstore,
		store:           ls,
		rpc:             grpc.NewServer(serverOptions...),
		bus:             broadcast.NewBroadcaster(EventBusCapacity),
		connectors:      make(map[thread.ID]*app.Connector),
		ctx:             ctx,
		cancel:          cancel,
		semaphores:      util.NewSemaphorePool(1),
		queueGetLogs:    queue.NewFFQueue(ctx, QueuePollInterval, conf.NetPullingInterval),
		queueGetRecords: queue.NewFFQueue(ctx, QueuePollInterval, conf.NetPullingInterval),
		threadExternal:  threadExtern,
	}

	err := n.migrateHeadsIfNeeded(ctx, ls)
	if err != nil {
		return nil, err
	}

	n.server, err = newServer(n, dialOptions...)
	if err != nil {
		return nil, err
	}

	listener, err := gostream.Listen(h, thread.Protocol)
	if err != nil {
		return nil, err
	}
	n.listener = listener
	go func() {
		pb.RegisterServiceServer(n.rpc, n.server)
		if err := n.rpc.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("serve error: %v", err)
		}
	}()
	return n, nil
}

func (n *net) countRecords(ctx context.Context, tid thread.ID, rid cid.Cid) (int64, error) {
	var (
		cursor        = rid
		counter int64 = 0
	)
	sk, err := n.store.ServiceKey(tid)
	if err != nil {
		return 0, err
	}

	for cursor.Defined() {
		r, err := cbor.GetRecord(ctx, n, cursor, sk)
		if err != nil {
			return 0, err
		}
		cursor = r.PrevID()
		counter += 1
	}
	return counter, nil
}

func (n *net) migrateHeadsIfNeeded(ctx context.Context, ls lstore.Logstore) (err error) {
	if n.conf.NoExchangeEdgesMigration {
		return nil
	}

	threadIds, err := ls.Threads()
	if err != nil {
		return err
	}

	log.Info("checking for heads migration")
	isMigrationNeeded := false

	for _, tid := range threadIds {
		tInfo, err := ls.GetThread(tid)
		if err != nil {
			return err
		}

		for _, l := range tInfo.Logs {
			heads, err := ls.Heads(tid, l.ID)
			if err != nil {
				return err
			}
			hslice := make([]thread.Head, 0)
			for _, h := range heads {
				// In this case we didn't migrate the thread
				if h.Counter == thread.CounterUndef && h.ID != cid.Undef {
					if !isMigrationNeeded {
						log.Info("starting migrating heads")
						isMigrationNeeded = true
					}
					counter, err := n.countRecords(ctx, tid, h.ID)
					if err != nil {
						return err
					}
					log.Debugf("counter for thread %s, log %s, head %s is %d", tid, l.ID, h.ID, counter)

					hslice = append(hslice, thread.Head{ID: h.ID, Counter: counter})
				} else {
					hslice = append(hslice, h)
				}
			}
			err = ls.SetHeads(tid, l.ID, hslice)
			if err != nil {
				return err
			}
		}
	}

	if isMigrationNeeded {
		log.Info("finished migrating heads")
	}

	return nil
}

func (n *net) Close() (err error) {
	// Wait for all thread pulls to finish
	n.semaphores.Stop()

	// Close peer connections and shutdown the server
	n.server.Lock()
	defer n.server.Unlock()
	for _, c := range n.server.conns {
		if err = c.Close(); err != nil {
			log.Errorf("error closing connection: %v", err)
		}
	}
	tu.StopGRPCServer(n.rpc)

	var errs []error
	weakClose := func(name string, c interface{}) {
		if cl, ok := c.(io.Closer); ok {
			if err = cl.Close(); err != nil {
				errs = append(errs, fmt.Errorf("%s error: %v", name, err))
			}
		}
	}
	weakClose("DAGService", n.DAGService)
	weakClose("host", n.host)
	weakClose("threadstore", n.store)
	if len(errs) > 0 {
		return fmt.Errorf("failed while closing net; err(s): %q", errs)
	}

	n.bus.Discard()
	n.cancel()
	n.listener.Close()
	return nil
}

func (n *net) Host() host.Host {
	return n.host
}

func (n *net) Store() lstore.Logstore {
	return n.store
}

func (n *net) GetHostID(_ context.Context) (peer.ID, error) {
	return n.host.ID(), nil
}

// 获取Pubsub对象
func (n *net) GetPubsub() *pubsub.PubSub {
	if n.server == nil {
		return nil
	}
	return n.server.ps
}

func (n *net) GetToken(ctx context.Context, identity thread.Identity) (tok thread.Token, err error) {
	msg := make([]byte, tokenChallengeBytes)
	if _, err = rand.Read(msg); err != nil {
		return
	}
	sctx, cancel := context.WithTimeout(ctx, tokenChallengeTimeout)
	defer cancel()
	sig, err := identity.Sign(sctx, msg)
	if err != nil {
		return
	}
	key := identity.GetPublic()
	if ok, err := key.Verify(msg, sig); !ok || err != nil {
		return tok, fmt.Errorf("bad signature")
	}
	return thread.NewToken(n.getPrivKey(), key)
}

func (n *net) CreateThread(
	_ context.Context,
	id thread.ID,
	opts ...core.NewThreadOption,
) (info thread.Info, err error) {
	args := &core.NewThreadOptions{}
	for _, opt := range opts {
		opt(args)
	}
	// @todo: Check identity key against ACL.
	identity, err := n.Validate(id, args.Token, false)
	if err != nil {
		return
	}
	if identity != nil {
		log.Debugf("creating thread with identity: %s", identity)
	} else {
		identity = thread.NewLibp2pPubKey(n.getPrivKey().GetPublic())
	}

	if err = n.ensureUniqueLog(id, args.LogKey, identity); err != nil {
		return
	}

	info = thread.Info{
		ID:  id,
		Key: args.ThreadKey,
	}
	if !info.Key.Defined() {
		info.Key = thread.NewRandomKey()
	}
	if err = n.store.AddThread(info); err != nil {
		return
	}
	if _, err = n.createLog(id, args.LogKey, identity); err != nil {
		return
	}

	return n.getThreadWithAddrs(id)
}

func (n *net) AddThread(
	ctx context.Context,
	addr ma.Multiaddr,
	opts ...core.NewThreadOption,
) (info thread.Info, err error) {
	args := &core.NewThreadOptions{}
	for _, opt := range opts {
		opt(args)
	}

	id, err := thread.FromAddr(addr)
	if err != nil {
		return
	}
	identity, err := n.Validate(id, args.Token, false)
	if err != nil {
		return
	}
	if identity != nil {
		log.Debugf("adding thread with identity: %s", identity)
	} else {
		identity = thread.NewLibp2pPubKey(n.getPrivKey().GetPublic())
	}

	if err = n.ensureUniqueLog(id, args.LogKey, identity); err != nil {
		return
	}

	threadComp, err := ma.NewComponent(thread.Name, id.String())
	if err != nil {
		return
	}
	peerAddr := addr.Decapsulate(threadComp)
	addri, err := peer.AddrInfoFromP2pAddr(peerAddr)
	if err != nil {
		return
	}

	// Check if we're trying to dial ourselves (regardless of addr)
	addFromSelf := addri.ID == n.host.ID()
	if addFromSelf {
		// Error if we don't have the thread locally
		if _, err = n.store.GetThread(id); errors.Is(err, lstore.ErrThreadNotFound) {
			err = fmt.Errorf("cannot retrieve thread from self: %v", err)
			return
		}
	}

	// Even if we already have the thread locally, we might still need to add a new log
	if err = n.store.AddThread(thread.Info{
		ID:  id,
		Key: args.ThreadKey,
	}); err != nil {
		return
	}
	if args.ThreadKey.CanRead() || args.LogKey != nil {
		if _, err = n.createLog(id, args.LogKey, identity); err != nil {
			return
		}
	}

	// Skip if trying to dial ourselves (already have the logs)
	if !addFromSelf {
		if err = n.Host().Connect(ctx, *addri); err != nil {
			return
		}

		if err = n.queueGetLogs.Call(addri.ID, id, func(ctx context.Context, p peer.ID, t thread.ID) error {
			return n.updateLogsFromPeer(ctx, p, t)
		}); err != nil {
			return
		}
	}
	return n.getThreadWithAddrs(id)
}

func (n *net) GetThread(_ context.Context, id thread.ID, opts ...core.ThreadOption) (info thread.Info, err error) {
	args := &core.ThreadOptions{}
	for _, opt := range opts {
		opt(args)
	}
	if _, err = n.Validate(id, args.Token, true); err != nil {
		return
	}
	return n.getThreadWithAddrs(id)
}

func (n *net) getThreadWithAddrs(id thread.ID) (info thread.Info, err error) {
	var tinfo thread.Info
	var peerID *ma.Component
	var threadID *ma.Component
	tinfo, err = n.store.GetThread(id)
	if err != nil {
		return
	}
	peerID, err = ma.NewComponent("p2p", n.host.ID().String())
	if err != nil {
		return
	}
	threadID, err = ma.NewComponent("thread", tinfo.ID.String())
	if err != nil {
		return
	}
	addrs := n.host.Addrs()
	res := make([]ma.Multiaddr, len(addrs))
	for i := range addrs {
		res[i] = addrs[i].Encapsulate(peerID).Encapsulate(threadID)
	}
	tinfo.Addrs = res
	return tinfo, nil
}

func (n *net) PullThread(ctx context.Context, id thread.ID, opts ...core.ThreadOption) error {
	args := &core.ThreadOptions{}
	for _, opt := range opts {
		opt(args)
	}
	if _, err := n.Validate(id, args.Token, true); err != nil {
		return err
	}
	return n.pullThread(ctx, id)
}

type timestampedRecord struct {
	record     core.Record
	counter    int64
	createtime int64
	logid      peer.ID
}

func (n *net) pullThread(ctx context.Context, tid thread.ID) error {
	ts := n.semaphores.Get(semaThreadUpdate(tid))
	ts.Acquire()
	recs, err := n.pullThreadDeal(ctx, tid)
	ts.Release()
	if err != nil {
		return err
	}
	connector, appConnected := n.getConnector(tid)
	//遍历所有记录，按时间排序后，逐条处理
	var tRecords []timestampedRecord
	for lid, rs := range recs {
		if appConnected {
			indexCounter := rs.counter - int64(len(rs.records)) + 1
			for i, r := range rs.records {
				block, err := r.GetBlock(ctx, n)
				if err != nil {
					return err
				}
				event, ok := block.(*cbor.Event)
				if !ok {
					event, err = cbor.EventFromNode(block)
					if err != nil {
						return fmt.Errorf("invalid event: %w", err)
					}
				}
				header, err := event.GetHeader(ctx, n, nil)
				if err != nil {
					return err
				}
				body, err := event.GetBody(ctx, n, nil)
				if err != nil {
					return err
				}
				// store internal blocks locally, record envelope will be added by the caller after successful processing
				if err = n.AddMany(ctx, []format.Node{event, header, body}); err != nil {
					return err
				}
				tRecord := NewRecord(r, tid, lid)
				counter := indexCounter + int64(i)
				createtime, err := connector.GetNetRecordCreateTime(ctx, tRecord)
				if err != nil {
					return err
				}

				tRecords = append(tRecords, timestampedRecord{
					record:     r,
					counter:    counter,
					createtime: createtime,
					logid:      lid,
				})
			}
		} else {
			if err = n.putRecords(ctx, tid, lid, rs.records, rs.counter); err != nil {
				return err
			}
		}
	}
	//按时间排序
	sort.Slice(tRecords, func(i, j int) bool {
		return tRecords[i].createtime < tRecords[j].createtime
	})
	//逐条处理
	for _, r := range tRecords {
		if err = n.putRecords(ctx, tid, r.logid, []core.Record{r.record}, r.counter); err != nil {
			return err
		}
	}
	return nil
}

// pullThread for the new records. This method is thread-safe.
func (n *net) pullThreadDeal(ctx context.Context, tid thread.ID) (map[peer.ID]peerRecords, error) {
	offsets, peers, err := n.threadOffsets(tid)
	if err != nil {
		return nil, err
	}
	if n.threadExternal != nil { //通过外部方式确定备份节点的位置
		if extPeers, err := n.threadExternal.GetPeers(n.ctx, tid); err == nil && len(extPeers) > 0 {
			peers = extPeers
		}
	}
	pulledRecs := make(map[peer.ID]peerRecords)
	for {
		// Pull from peers
		recs, err := n.server.getRecords(peers, tid, offsets, n.conf.NetPullingLimit)
		if err != nil {
			return nil, err
		}
		continueFlag := false
		for lid, rs := range recs {
			if len(rs.records) > 0 {
				records := append(pulledRecs[lid].records, rs.records...)
				pulledRecs[lid] = peerRecords{
					records: records,
					counter: rs.counter,
				}
				record := rs.records[len(rs.records)-1]
				offsets[lid] = thread.Head{
					ID:      record.Cid(),
					Counter: rs.counter,
				}
			}
			if len(rs.records) == int(n.conf.NetPullingLimit) {
				continueFlag = true
			}
		}
		if !continueFlag {
			break
		}

	}
	return pulledRecs, nil
}

func (n *net) DeleteThread(ctx context.Context, id thread.ID, opts ...core.ThreadOption) error {
	args := &core.ThreadOptions{}
	for _, opt := range opts {
		opt(args)
	}
	if _, err := n.Validate(id, args.Token, false); err != nil {
		return err
	}
	if _, ok := n.getConnectorProtected(id, args.APIToken); !ok {
		return fmt.Errorf("cannot delete thread: %w", app.ErrThreadInUse)
	}

	log.Debugf("deleting thread %s...", id)
	ts := n.semaphores.Get(semaThreadUpdate(id))

	// Must block in case the thread is being pulled
	ts.Acquire()
	err := n.deleteThread(ctx, id)
	ts.Release()

	return err
}

// deleteThread cleans up all the persistent and in-memory bits of a thread. This includes:
// - Removing all record and event nodes.
// - Deleting all logstore keys, addresses, and heads.
// - Cancelling the pubsub subscription and topic.
// Local subscriptions will not be cancelled and will simply stop reporting.
// This method is internal and *not* thread-safe. It assumes we currently own the thread-lock.
func (n *net) deleteThread(ctx context.Context, id thread.ID) error {
	info, err := n.store.GetThread(id)
	if err != nil {
		return err
	}
	for _, lg := range info.Logs { // Walk logs, removing record and event nodes
		head := lg.Head.ID
		for head.Defined() {
			head, err = n.deleteRecord(ctx, head, info.Key.Service())
			if err != nil {
				return err
			}
		}
	}

	return n.store.DeleteThread(id) // Delete logstore keys, addresses, heads, and metadata
}

func (n *net) AddReplicator(
	ctx context.Context,
	id thread.ID,
	paddr ma.Multiaddr,
	opts ...core.ThreadOption,
) (pid peer.ID, err error) {
	args := &core.ThreadOptions{}
	for _, opt := range opts {
		opt(args)
	}
	if _, err = n.Validate(id, args.Token, true); err != nil {
		return
	}

	info, err := n.store.GetThread(id)
	if err != nil {
		return
	}

	// Extract peer portion
	p2p, err := paddr.ValueForProtocol(ma.P_P2P)
	if err != nil {
		return
	}
	pid, err = peer.Decode(p2p)
	if err != nil {
		return
	}

	// Update local addresses
	addr, err := ma.NewMultiaddr("/" + ma.ProtocolWithCode(ma.P_P2P).Name + "/" + p2p)
	if err != nil {
		return
	}
	managedLogs, err := n.store.GetManagedLogs(info.ID)
	if err != nil {
		return
	}
	for _, lg := range managedLogs {
		if err = n.store.AddAddr(info.ID, lg.ID, addr, pstore.PermanentAddrTTL); err != nil {
			return
		}
	}
	info, err = n.store.GetThread(info.ID) // Update info
	if err != nil {
		return
	}

	// Check if we're dialing ourselves (regardless of addr)
	if pid != n.host.ID() {
		// If not, update peerstore address
		var dialable ma.Multiaddr
		dialable, err = getDialable(paddr)
		if err == nil {
			n.host.Peerstore().AddAddr(pid, dialable, pstore.PermanentAddrTTL)
		}

		// Send all logs to the new replicator
		for _, l := range info.Logs {
			if err = n.server.pushLog(ctx, info.ID, l, pid, info.Key.Service(), nil); err != nil {
				for _, lg := range managedLogs {
					// Rollback this log only and then bail
					if lg.ID == l.ID {
						if err := n.store.SetAddrs(info.ID, lg.ID, lg.Addrs, pstore.PermanentAddrTTL); err != nil {
							log.Errorf("error rolling back log address change: %s", err)
						}
						break
					}
				}
				return
			}
		}
	}

	// Send the updated log(s) to peers
	var addrs []ma.Multiaddr
	for _, l := range info.Logs {
		addrs = append(addrs, l.Addrs...)
	}
	peers, err := n.uniquePeers(addrs)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	for _, p := range peers {
		wg.Add(1)
		go func(pid peer.ID) {
			defer wg.Done()
			for _, lg := range managedLogs {
				if err = n.server.pushLog(ctx, info.ID, lg, pid, nil, nil); err != nil {
					log.Errorf("error pushing log %s to %s: %v", lg.ID, pid, err)
				}
			}
		}(p)
	}

	wg.Wait()
	return pid, nil
}

func (n *net) uniquePeers(addrs []ma.Multiaddr) ([]peer.ID, error) {
	var pm = make(map[peer.ID]struct{}, len(addrs))
	for _, addr := range addrs {
		pid, ok, err := n.callablePeer(addr)
		if err != nil {
			return nil, err
		} else if !ok {
			// skip calling itself
			continue
		}
		pm[pid] = struct{}{}
	}
	var ps = make([]peer.ID, 0, len(pm))
	for pid := range pm {
		ps = append(ps, pid)
	}
	return ps, nil
}

// callablePeer attempts to obtain external peer ID from the multiaddress.
func (n *net) callablePeer(addr ma.Multiaddr) (peer.ID, bool, error) {
	p, err := addr.ValueForProtocol(ma.P_P2P)
	if err != nil {
		return "", false, err
	}

	pid, err := peer.Decode(p)
	if err != nil {
		return "", false, err
	}

	if pid.String() == n.host.ID().String() {
		return pid, false, nil
	}

	return pid, true, nil
}

func getDialable(addr ma.Multiaddr) (ma.Multiaddr, error) {
	parts := strings.Split(addr.String(), "/"+ma.ProtocolWithCode(ma.P_P2P).Name)
	return ma.NewMultiaddr(parts[0])
}

func (n *net) CreateRecord(
	ctx context.Context,
	id thread.ID,
	body format.Node,
	opts ...core.ThreadOption,
) (tr core.ThreadRecord, err error) {
	args := &core.ThreadOptions{}
	for _, opt := range opts {
		opt(args)
	}
	identity, err := n.Validate(id, args.Token, false)
	if err != nil {
		return
	}
	if identity == nil {
		identity = thread.NewLibp2pPubKey(n.getPrivKey().GetPublic())
	}
	con, ok := n.getConnectorProtected(id, args.APIToken)
	if !ok {
		return nil, fmt.Errorf("cannot create record: %w", app.ErrThreadInUse)
	} else if con != nil {
		if err = con.ValidateNetRecordBody(ctx, body, identity); err != nil {
			return
		}
	}

	lg, err := n.getOrCreateLog(id, identity)
	if err != nil {
		return
	}
	r, err := n.newRecord(ctx, id, lg, body, identity)
	if err != nil {
		return
	}
	tr = NewRecord(r, id, lg.ID)
	head := thread.Head{
		ID:      tr.Value().Cid(),
		Counter: lg.Head.Counter + 1,
	}
	if err = n.store.SetHead(id, lg.ID, head); err != nil {
		return
	}

	//设置loginfo关键记录(counter为10000的倍数的记录）point
	n.SetThreadLogPoint(id, lg.ID, lg.Head.Counter+1, tr.Value().Cid())
	log.Debugf("created record %s (thread=%s, log=%s)", tr.Value().Cid(), id, lg.ID)
	if err = n.bus.SendWithTimeout(tr, notifyTimeout); err != nil {
		return
	}
	if err = n.server.pushRecord(ctx, id, lg.ID, tr.Value(), lg.Head.Counter+1); err != nil {
		return
	}
	return tr, nil
}

func (n *net) AddRecord(
	ctx context.Context,
	id thread.ID,
	lid peer.ID,
	rec core.Record,
	opts ...core.ThreadOption,
) error {
	args := &core.ThreadOptions{}
	for _, opt := range opts {
		opt(args)
	}
	if _, err := n.Validate(id, args.Token, false); err != nil {
		return err
	}

	logpk, err := n.store.PubKey(id, lid)
	if err != nil {
		return err
	}
	if logpk == nil {
		return lstore.ErrLogNotFound
	}

	if knownRecord, err := n.isKnown(ctx, rec.Cid()); err != nil {
		return err
	} else if knownRecord {
		return nil
	}

	if err = rec.Verify(logpk); err != nil {
		return err
	}
	if err = n.putRecords(ctx, id, lid, []core.Record{rec}, thread.CounterUndef); err != nil {
		return err
	}
	return n.server.pushRecord(ctx, id, lid, rec, thread.CounterUndef)
}

func (n *net) GetRecord(
	ctx context.Context,
	id thread.ID,
	rid cid.Cid,
	opts ...core.ThreadOption,
) (core.Record, error) {
	args := &core.ThreadOptions{}
	for _, opt := range opts {
		opt(args)
	}
	if _, err := n.Validate(id, args.Token, true); err != nil {
		return nil, err
	}
	return n.getRecord(ctx, id, rid)
}

func (n *net) getRecord(ctx context.Context, id thread.ID, rid cid.Cid) (core.Record, error) {
	sk, err := n.store.ServiceKey(id)
	if err != nil {
		return nil, err
	}
	if sk == nil {
		return nil, fmt.Errorf("a service-key is required to get records")
	}
	return cbor.GetRecord(ctx, n, rid, sk)
}

// Record implements core.Record. The most basic component of a Log.
type Record struct {
	core.Record
	threadID thread.ID
	logID    peer.ID
}

// NewRecord returns a record with the given values.
func NewRecord(r core.Record, id thread.ID, lid peer.ID) core.ThreadRecord {
	return &Record{Record: r, threadID: id, logID: lid}
}

func (r *Record) Value() core.Record {
	return r
}

func (r *Record) ThreadID() thread.ID {
	return r.threadID
}

func (r *Record) LogID() peer.ID {
	return r.logID
}

func (n *net) Subscribe(ctx context.Context, opts ...core.SubOption) (<-chan core.ThreadRecord, error) {
	args := &core.SubOptions{}
	for _, opt := range opts {
		opt(args)
	}

	filter := make(map[thread.ID]struct{})
	for _, id := range args.ThreadIDs {
		if err := id.Validate(); err != nil {
			return nil, err
		}
		if id.Defined() {
			if _, err := n.Validate(id, args.Token, true); err != nil {
				return nil, err
			}
			filter[id] = struct{}{}
		}
	}
	return n.subscribe(ctx, filter)
}

func (n *net) subscribe(ctx context.Context, filter map[thread.ID]struct{}) (<-chan core.ThreadRecord, error) {
	channel := make(chan core.ThreadRecord)
	go func() {
		defer close(channel)
		listener := n.bus.Listen()
		defer listener.Discard()
		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-listener.Channel():
				if !ok {
					return
				}
				if rec, ok := i.(*Record); ok {
					if len(filter) > 0 {
						if _, ok := filter[rec.threadID]; ok {
							channel <- rec
						}
					} else {
						channel <- rec
					}
				} else {
					log.Warn("listener received a non-record value")
				}
			}
		}
	}()
	return channel, nil
}

func (n *net) ConnectApp(a app.App, id thread.ID) (*app.Connector, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	info, err := n.getThreadWithAddrs(id)
	if err != nil {
		return nil, fmt.Errorf("error getting thread %s: %v", id, err)
	}
	con, err := app.NewConnector(a, n, info)
	if err != nil {
		return nil, fmt.Errorf("error making connector %s: %v", id, err)
	}
	n.addConnector(id, con)
	return con, nil
}

// @todo: Handle thread ACL checks against ID and readOnly.
func (n *net) Validate(id thread.ID, token thread.Token, readOnly bool) (thread.PubKey, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	return token.Validate(n.getPrivKey())
}

func (n *net) addConnector(id thread.ID, conn *app.Connector) {
	n.connLock.Lock()
	n.connectors[id] = conn
	n.connLock.Unlock()
}

func (n *net) getConnector(id thread.ID) (*app.Connector, bool) {
	n.connLock.RLock()
	defer n.connLock.RUnlock()

	conn, exist := n.connectors[id]
	return conn, exist
}

// getConnectorProtected returns the connector tied to the thread if it exists
// and whether or not the token is valid.
func (n *net) getConnectorProtected(id thread.ID, token core.Token) (*app.Connector, bool) {
	c, exist := n.getConnector(id)
	if !exist {
		return nil, true // thread is not owned by a connector
	}
	if !token.Equal(c.Token()) {
		return nil, false
	}
	return c, true
}

// PutRecord adds an existing record. This method is thread-safe.
func (n *net) PutRecord(ctx context.Context, id thread.ID, lid peer.ID, rec core.Record, counter int64) error {
	if err := id.Validate(); err != nil {
		return err
	}
	return n.putRecords(ctx, id, lid, []core.Record{rec}, counter)
}

// putRecords adds existing records. This method is thread-safe.
func (n *net) putRecords(ctx context.Context, tid thread.ID, lid peer.ID, recs []core.Record, counter int64) error {
	chain, head, err := n.loadRecords(ctx, tid, lid, recs, counter)
	if err != nil {
		return fmt.Errorf("loading records failed: %w", err)
	} else if len(chain) == 0 {
		return nil
	}

	ts := n.semaphores.Get(semaThreadUpdate(tid))
	ts.Acquire()
	defer ts.Release()

	// check the head again, as some other process could change the log concurrently
	if current, err := n.currentHead(tid, lid); err != nil {
		return fmt.Errorf("fetching head failed: %w", err)
	} else if !current.ID.Equals(head.ID) {
		// fast-forward the chain up to the updated head
		var headReached bool
		head = current
		for i := 0; i < len(chain); i++ {
			if chain[i].Value().Cid().Equals(current.ID) {
				chain = chain[i+1:]
				headReached = true
				break
			}
		}
		if !headReached {
			// entire chain already processed
			return nil
		}
	}

	var (
		connector, appConnected = n.getConnector(tid)
		identity                = &thread.Libp2pPubKey{}
		readKey                 *sym.Key
		validate                bool
		// setting new counters for heads
		updatedCounter = head.Counter
	)

	if appConnected {
		var err error
		if readKey, err = n.store.ReadKey(tid); err != nil {
			return err
		} else if readKey != nil {
			validate = true
		}
	}

	for _, record := range chain {
		if validate {
			block, err := record.Value().GetBlock(ctx, n)
			if err != nil {
				return err
			}

			event, ok := block.(*cbor.Event)
			if !ok {
				event, err = cbor.EventFromNode(block)
				if err != nil {
					return fmt.Errorf("invalid event: %w", err)
				}
			}

			dbody, err := event.GetBody(ctx, n, readKey)
			if err != nil {
				return err
			}

			if err = identity.UnmarshalBinary(record.Value().PubKey()); err != nil {
				return err
			}

			if err = connector.ValidateNetRecordBody(ctx, dbody, identity); err != nil {
				userErr := err

				// remove stored internal blocks
				header, err := event.GetHeader(ctx, n, nil)
				if err != nil {
					return err
				}

				body, err := event.GetBody(ctx, n, nil)
				if err != nil {
					return err
				}

				if err := n.RemoveMany(ctx, []cid.Cid{event.Cid(), header.Cid(), body.Cid()}); err != nil {
					return fmt.Errorf("removing invalid blocks: %w", err)
				}

				return userErr
			}
		}

		updatedCounter++
		if err := n.store.SetHead(
			tid,
			lid,
			thread.Head{
				ID:      record.Value().Cid(),
				Counter: updatedCounter,
			}); err != nil {
			return fmt.Errorf("setting log head failed: %w", err)
		}
		//设置loginfo关键记录(counter为10000的倍数的记录）point
		n.SetThreadLogPoint(tid, lid, updatedCounter, record.Value().Cid())
		if appConnected {
			if err := connector.HandleNetRecord(ctx, record); err != nil {
				// Future improvement notes.
				// If record handling fails there are two options available:
				// 1. Just interrupt and return error (current behaviour). Log head remains moved and some events
				//    from the record possibly won't reach reducers/listeners or even get dispatched.
				// 2. Rollback log head to the previous record. In this case record handling will be retried until
				//    success, but reducers must guarantee its idempotence and there is a chance of getting stuck
				//    with bad event and not making any progress at all.
				return fmt.Errorf("handling record failed: %w", err)
			}
		}

		// add record envelope to the blockstore, indicating it was successfully processed
		if err := n.Add(ctx, record.Value()); err != nil {
			return fmt.Errorf("adding record to the blockstore failed: %w", err)
		}

		// Generally broadcasting should not block for too long, i.e. we have to run it
		// under the semaphore to ensure consistent order seen by the listeners. Record
		// bursts could be overcome by adjusting listener buffers (EventBusCapacity).
		if err = n.bus.SendWithTimeout(record, notifyTimeout); err != nil {
			return err
		}
	}

	return nil
}

// Load, validate and cache all records in log between last provided and currentHead.
func (n *net) loadRecords(
	ctx context.Context,
	tid thread.ID,
	lid peer.ID,
	recs []core.Record,
	counter int64,
) ([]core.ThreadRecord, thread.Head, error) {
	if len(recs) == 0 {
		return nil, thread.HeadUndef, errors.New("cannot load empty record chain")
	}
	head, err := n.currentHead(tid, lid)
	if err != nil {
		return nil, thread.HeadUndef, err
	}

	// check if the last record was already loaded and processed
	var last = recs[len(recs)-1]
	// if we don't have the counter (but have some recs) then we were communicating with old version peer
	if counter == thread.CounterUndef {
		if exist, err := n.isKnown(ctx, last.Cid()); err != nil {
			return nil, thread.HeadUndef, err
		} else if exist || !last.Cid().Defined() {
			return nil, thread.HeadUndef, nil
		}
	} else if counter <= head.Counter {
		return nil, head, nil
	}

	var (
		chain    = make([]core.Record, 0, len(recs))
		complete bool
	)

	for i := len(recs) - 1; i >= 0; i-- {
		var next = recs[i]
		if c := next.Cid(); !c.Defined() || c.Equals(head.ID) {
			complete = true
			break
		}
		chain = append(chain, next)
	}

	if !complete {
		// bridge the gap between the last provided record and current head
		var c = chain[len(chain)-1].PrevID()
		for c.Defined() {
			if c.Equals(head.ID) {
				break
			}

			r, err := n.getRecord(ctx, tid, c)
			if err != nil {
				return nil, head, err
			}

			chain = append(chain, r)
			c = r.PrevID()
		}
	}

	if len(chain) == 0 {
		// fast path
		return nil, head, nil
	}

	tRecords := make([]core.ThreadRecord, 0, len(chain))
	for i := len(chain) - 1; i >= 0; i-- {
		var r = chain[i]
		block, err := r.GetBlock(ctx, n)
		if err != nil {
			return nil, head, err
		}

		event, ok := block.(*cbor.Event)
		if !ok {
			event, err = cbor.EventFromNode(block)
			if err != nil {
				return nil, head, fmt.Errorf("invalid event: %w", err)
			}
		}

		header, err := event.GetHeader(ctx, n, nil)
		if err != nil {
			return nil, head, err
		}

		body, err := event.GetBody(ctx, n, nil)
		if err != nil {
			return nil, head, err
		}

		// store internal blocks locally, record envelope will be added by the caller after successful processing
		if err = n.AddMany(ctx, []format.Node{event, header, body}); err != nil {
			return nil, head, err
		}

		tRecords = append(tRecords, NewRecord(r, tid, lid))
	}

	return tRecords, head, nil
}

func (n *net) isKnown(ctx context.Context, rec cid.Cid) (bool, error) {
	return n.bstore.Has(ctx, rec)
}

func (n *net) currentHead(tid thread.ID, lid peer.ID) (thread.Head, error) {
	var head thread.Head
	heads, err := n.store.Heads(tid, lid)
	if err != nil {
		return head, err
	}

	if len(heads) > 0 {
		head = heads[0]
	} else {
		head = thread.HeadUndef
	}

	return head, nil
}

// newRecord creates a new record with the given body as a new event body.
func (n *net) newRecord(
	ctx context.Context,
	id thread.ID,
	lg thread.LogInfo,
	body format.Node,
	pk thread.PubKey,
) (core.Record, error) {
	if lg.PrivKey == nil {
		return nil, fmt.Errorf("a private-key is required to create records")
	}
	sk, err := n.store.ServiceKey(id)
	if err != nil {
		return nil, err
	}
	if sk == nil {
		return nil, fmt.Errorf("a service-key is required to create records")
	}
	rk, err := n.store.ReadKey(id)
	if err != nil {
		return nil, err
	}
	if rk == nil {
		return nil, fmt.Errorf("a read-key is required to create records")
	}
	event, err := cbor.CreateEvent(ctx, n, body, rk)
	if err != nil {
		return nil, err
	}
	return cbor.CreateRecord(ctx, n, cbor.CreateRecordConfig{
		Block:      event,
		Prev:       lg.Head.ID,
		Key:        lg.PrivKey,
		PubKey:     pk,
		ServiceKey: sk,
	})
}

// getPrivKey returns the host's private key.
func (n *net) getPrivKey() crypto.PrivKey {
	return n.host.Peerstore().PrivKey(n.host.ID())
}

// getLocalRecords returns local records from the given thread that are ahead of
// offset but not farther than limit.
// It is possible to reach limit before offset, meaning that the caller
// will be responsible for the remaining traversal.
func (n *net) getLocalRecords(
	ctx context.Context,
	id thread.ID,
	lid peer.ID,
	offset cid.Cid,
	limit int,
	counter int64,
) ([]core.Record, error) {
	lg, err := n.store.GetLog(id, lid)
	if err != nil {
		return nil, err
	}
	if lg.Head.Counter <= counter {
		return []core.Record{}, nil
	}
	sk, err := n.store.ServiceKey(id)
	if err != nil {
		return nil, err
	}
	if sk == nil {
		return nil, fmt.Errorf("a service-key is required to get records")
	}

	var (
		cursor      = lg.Head.ID
		recs        []core.Record
		diffCounter = lg.Head.Counter - counter
	)
	// traverse the log until we reach the limit or offset
	if lg.Head.Counter > counter+int64(limit)+10000 { //If the counter at the offset location plus the limit, the counter with the head exceeds 10,000, set the cursor to the counter at the offset location plus the limit, which is the nearest 10,000-level counter.
		var pointCounter int64
		cursor, pointCounter, err = n.GetThreadLogPoint(id, lid, counter+int64(limit))
		if err != nil {
			cursor = lg.Head.ID
			diffCounter = lg.Head.Counter - counter + 1 //+1 because it needs to include the record where the offset is located
		} else {
			diffCounter = pointCounter - counter + 1 //+1 because it needs to include the record where the offset is located
		}

	}
	recsLen := 0
	for recsLen < int(diffCounter) { //Take at most diff counter records
		if !cursor.Defined() || cursor.String() == offset.String() {
			break
		}
		r, err := cbor.GetRecord(ctx, n, cursor, sk) // Important invariant: heads are always in blockstore
		if err != nil {
			// return records fetched so far
			return recs, err
		}
		recs = append([]core.Record{r}, recs...)
		recsLen++
		cursor = r.PrevID()
	}
	if recsLen > limit { //If the number of records obtained is greater than the limit, limit records will be intercepted before interception.
		recs = recs[:limit]
	}
	return recs, nil
}

// deleteRecord remove a record from the dag service.
func (n *net) deleteRecord(ctx context.Context, rid cid.Cid, sk *sym.Key) (prev cid.Cid, err error) {
	rec, err := cbor.GetRecord(ctx, n, rid, sk)
	if err != nil {
		return
	}
	if err = cbor.RemoveRecord(ctx, n, rec); err != nil {
		return
	}
	event, err := cbor.EventFromRecord(ctx, n, rec)
	if err != nil {
		return
	}
	if err = cbor.RemoveEvent(ctx, n, event); err != nil {
		return
	}
	return rec.PrevID(), nil
}

// startPulling periodically pulls on all threads.
func (n *net) StartPulling() {
	if n.conf.NoNetPulling {
		return
	}
	select {
	case <-time.After(n.conf.NetPullingStartAfter):
	case <-n.ctx.Done():
		return
	}

	// set pull cycle interval into initial value,
	// it will be redefined on the next iteration
	var interval = n.conf.NetPullingInitialInterval

	// group threads by peers and exchange edges efficiently
	var compressor = queue.NewThreadPacker(n.ctx, MaxThreadsExchanged, ExchangeCompressionTimeout)
	go n.startExchange(compressor)

PullCycle:
	for {
		ts, err := n.store.Threads()
		if err != nil {
			log.Errorf("error listing threads: %s", err)
			return
		}
		log.Infof("pulling %d threads", len(ts))

		if len(ts) == 0 {
			// if there are no threads served, just wait and retry
			select {
			case <-time.After(interval):
				interval = n.conf.NetPullingInterval
				continue PullCycle
			case <-n.ctx.Done():
				return
			}
		}

		var (
			period = interval
			ticker = time.NewTicker(period)
			idx    = 0
		)

		for {
			select {
			case <-ticker.C:
				var tid = ts[idx]
				if n.threadExternal == nil { //If thread external is empty, get peers locally based on thread id.
					if _, peers, err := n.threadOffsets(tid); err != nil {
						log.Errorf("error getting thread info %s: %s", tid, err)
					} else {
						for _, pid := range peers {
							compressor.Add(pid, tid)
						}
					}
				} else { //If thread external is not empty, get peers from thread external based on thread id
					if peers, err := n.threadExternal.GetPeers(n.ctx, tid); err != nil {
						log.Debugf("error getting thread peers from threadExternal %s: %s", tid, err)
					} else {
						for _, pid := range peers {
							compressor.Add(pid, tid)
						}
					}
				}
				idx++
				if idx >= len(ts) {
					ticker.Stop()
					interval = n.conf.NetPullingInterval
					continue PullCycle
				}

			case <-n.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}
}

func (n *net) startExchange(compressor queue.ThreadPacker) {
	for pack := range compressor.Run() {
		go func(p queue.ThreadPack) {
			if err := n.server.exchangeEdges(n.ctx, p.Peer, p.Threads); err != nil {
				log.Debugf("exchangeEdges with %s failed: %v", p.Peer, err)
			}
		}(pack)
	}
}

// createLog creates a new log with the given peer as host.
func (n *net) createLog(id thread.ID, key crypto.Key, identity thread.PubKey) (info thread.LogInfo, err error) {
	var ok bool
	if key == nil {
		info.PrivKey, info.PubKey, err = crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return
		}
	} else if info.PrivKey, ok = key.(crypto.PrivKey); ok {
		info.PubKey = info.PrivKey.GetPublic()
	} else if info.PubKey, ok = key.(crypto.PubKey); !ok {
		return info, fmt.Errorf("invalid log-key")
	}
	info.ID, err = peer.IDFromPublicKey(info.PubKey)
	if err != nil {
		return
	}
	addr, err := ma.NewMultiaddr("/" + ma.ProtocolWithCode(ma.P_P2P).Name + "/" + n.host.ID().String())
	if err != nil {
		return
	}
	info.Addrs = []ma.Multiaddr{addr}
	// If we're creating the log, we're 'managing' it
	info.Managed = true

	// Add to thread
	if err = n.store.AddLog(id, info); err != nil {
		return info, err
	}
	lidb, err := info.ID.MarshalBinary()
	if err != nil {
		return info, err
	}
	if err = n.store.PutBytes(id, identity.String(), lidb); err != nil {
		return info, err
	}
	return info, nil
}

// getOrCreateLog returns a log for identity under the given thread.
// If no log exists, a new one is created.
func (n *net) getOrCreateLog(id thread.ID, identity thread.PubKey) (info thread.LogInfo, err error) {
	if identity == nil {
		identity = thread.NewLibp2pPubKey(n.getPrivKey().GetPublic())
	}
	lidb, err := n.store.GetBytes(id, identity.String())
	if err != nil {
		return info, err
	}
	// Check if we have an old-style "own" (unindexed) log
	if lidb == nil && identity.Equals(thread.NewLibp2pPubKey(n.getPrivKey().GetPublic())) {
		thrd, err := n.store.GetThread(id)
		if err != nil {
			return info, err
		}
		ownLog := thrd.GetFirstPrivKeyLog()
		if ownLog != nil {
			return *ownLog, nil
		}
	} else if lidb != nil {
		lid, err := peer.IDFromBytes(*lidb)
		if err != nil {
			return info, err
		}
		return n.store.GetLog(id, lid)
	}
	return n.createLog(id, nil, identity)
}

// createExternalLogsIfNotExist creates an external logs if doesn't exists. The created
// logs will have cid.Undef as the current head. Is thread-safe.
func (n *net) createExternalLogsIfNotExist(
	tid thread.ID,
	lis []thread.LogInfo,
) error {
	ts := n.semaphores.Get(semaThreadUpdate(tid))
	ts.Acquire()
	defer ts.Release()

	for _, li := range lis {
		if currHeads, err := n.Store().Heads(tid, li.ID); err != nil {
			return err
		} else if len(currHeads) == 0 {
			li.Head = thread.HeadUndef
			if err = n.Store().AddLog(tid, li); err != nil {
				return err
			}
		} else {
			// update log addresses
			if err = n.Store().AddAddrs(tid, li.ID, li.Addrs, pstore.PermanentAddrTTL); err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureUniqueLog returns a non-nil error if a log with key already exists,
// or if a log for identity already exists for the given thread.
func (n *net) ensureUniqueLog(id thread.ID, key crypto.Key, identity thread.PubKey) (err error) {
	thrd, err := n.store.GetThread(id)
	if errors.Is(err, lstore.ErrThreadNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	var lid peer.ID
	if key != nil {
		switch key := key.(type) {
		case crypto.PubKey:
			lid, err = peer.IDFromPublicKey(key)
			if err != nil {
				return err
			}
		case crypto.PrivKey:
			lid, err = peer.IDFromPrivateKey(key)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid log key")
		}
	} else {
		lidb, err := n.store.GetBytes(id, identity.String())
		if err != nil {
			return err
		}
		if lidb == nil {
			// Check if we have an old-style "own" (unindexed) log
			if identity.Equals(thread.NewLibp2pPubKey(n.getPrivKey().GetPublic())) {
				if thrd.GetFirstPrivKeyLog().PrivKey != nil {
					return lstore.ErrThreadExists
				}
			}
			return nil
		}
		lid, err = peer.IDFromBytes(*lidb)
		if err != nil {
			return err
		}
	}
	_, err = n.store.GetLog(id, lid)
	if err == nil {
		return lstore.ErrLogExists
	}
	if !errors.Is(err, lstore.ErrLogNotFound) {
		return err
	}
	return nil
}

// updateRecordsFromPeer fetches new logs & records from the peer and adds them in the local peer store.
func (n *net) updateRecordsFromPeer(ctx context.Context, pid peer.ID, tid thread.ID) error {
	offsets, _, err := n.threadOffsets(tid)
	if err != nil {
		return fmt.Errorf("getting offsets for thread %s failed: %w", tid, err)
	}
	req, sk, err := n.server.buildGetRecordsRequest(tid, offsets, n.conf.NetPullingLimit)
	if err != nil {
		return fmt.Errorf("building GetRecords request for thread %s failed: %w", tid, err)
	}
	recs, err := n.server.getRecordsFromPeer(ctx, tid, pid, req, sk)
	if err != nil {
		return fmt.Errorf("getting records for thread %s from %s failed: %w", tid, pid, err)
	}
	for lid, rs := range recs {
		if err = n.putRecords(ctx, tid, lid, rs.records, rs.counter); err != nil {
			return fmt.Errorf("putting records from log %s (thread %s) failed: %w", lid, tid, err)
		}
	}
	for lid, rs := range recs {
		head, err := n.currentHead(tid, lid)
		if err == nil {
			if head.Counter <= rs.counter && len(rs.records) == int(n.conf.NetPullingLimit) {
				// If we got the max number of records, there might be more
				return n.updateRecordsFromPeer(ctx, pid, tid)
			}
		}
	}

	return nil
}

// updateLogsFromPeer gets new logs information from the peer and adds it in the local peer store.
func (n *net) updateLogsFromPeer(ctx context.Context, pid peer.ID, tid thread.ID) error {
	lgs, err := n.server.getLogs(ctx, tid, pid)
	if err != nil {
		return err
	}
	return n.createExternalLogsIfNotExist(tid, lgs)
}

// returns offsets and involved peers for all known thread's logs.
func (n *net) threadOffsets(tid thread.ID) (map[peer.ID]thread.Head, []peer.ID, error) {
	info, err := n.store.GetThread(tid)
	if err != nil {
		return nil, nil, err
	}
	var (
		offsets = make(map[peer.ID]thread.Head, len(info.Logs))
		addrs   []ma.Multiaddr
	)
	for _, lg := range info.Logs {
		var has bool
		if lg.Head.ID.Defined() {
			has, err = n.isKnown(context.Background(), lg.Head.ID)
			if err != nil {
				return nil, nil, err
			}
		}
		if has {
			offsets[lg.ID] = lg.Head
		} else {
			offsets[lg.ID] = thread.HeadUndef
		}
		addrs = append(addrs, lg.Addrs...)
	}
	peers, err := n.uniquePeers(addrs)
	if err != nil {
		return nil, nil, err
	}
	return offsets, peers, nil
}

// Set the key node to be recorded in the thread's log. If the incoming counter is a multiple of 10000, the key node will be recorded, otherwise it will not be recorded.
func (n *net) SetThreadLogPoint(tid thread.ID, lid peer.ID, counter int64, rcid cid.Cid) (err error) {
	if counter%10000 != 0 {
		return
	}
	//构造关键节点的记录的key
	key := fmt.Sprintf("%s/%d", lid.String(), counter)
	err = n.store.PutBytes(tid, key, rcid.Bytes())
	return
}

// Get the key nodes recorded in the thread's log, and return the key nodes that are the nearest multiple of 10,000 on counter.
func (n *net) GetThreadLogPoint(tid thread.ID, lid peer.ID, counter int64) (rcid cid.Cid, pointCounter int64, err error) {
	//向上找最近的10000倍数
	pointCounter = (counter/10000 + 1) * 10000
	//构造关键节点的记录的key
	key := fmt.Sprintf("%s/%d", lid.String(), pointCounter)
	data, err := n.store.GetBytes(tid, key)
	if err != nil || data == nil {
		return
	}
	rcid, err = cid.Cast(*data)
	return
}

// Exchange thread records with  peers that have the thread.
func (n *net) Exchange(ctx context.Context, tid thread.ID) error {
	// Get thread info.
	info, err := n.store.GetThread(tid)
	if err != nil {
		return err
	}
	// Get peers with the thread.
	addrs := make([]ma.Multiaddr, 0, len(info.Logs))
	for _, lg := range info.Logs {
		addrs = append(addrs, lg.Addrs...)
	}
	peers, err := n.uniquePeers(addrs)
	if err != nil {
		return err
	}
	// Get addresses from external object
	if n.threadExternal != nil {
		resPeers, err := n.threadExternal.GetPeers(ctx, tid)
		if err == nil && len(resPeers) > 0 {
			peers = resPeers
		}
	}
	// Exchange with each peer.
	for _, pid := range peers {
		go func(pId peer.ID) {
			n.server.exchangeEdges(n.ctx, pId, []thread.ID{tid})
		}(pid)
	}
	return nil
}
