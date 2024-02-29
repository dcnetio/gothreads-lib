package peer

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	badger "github.com/dcnetio/gothreads-lib/go-ds-badger"
	rpc "github.com/dcnetio/gothreads-lib/go-libp2p-pubsub-rpc"
	"github.com/dcnetio/gothreads-lib/go-libp2p-pubsub-rpc/finalizer"
	"github.com/dcnetio/gothreads-lib/go-libp2p-pubsub-rpc/peer/mdns"
	ipfslite "github.com/dcnetio/ipfs-lite"
	blockstore "github.com/ipfs/boxo/blockstore"
	format "github.com/ipfs/go-ipld-format"
	golog "github.com/ipfs/go-log/v2"
	ipfsconfig "github.com/ipfs/kubo/config"
	"github.com/libp2p/go-libp2p"
	ps "github.com/libp2p/go-libp2p-pubsub"
	cconnmgr "github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	connmgr "github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
)

var log = golog.Logger("psrpc/peer")

// Config defines params for Peer configuration.
type Config struct {
	RepoPath                 string
	PrivKey                  crypto.PrivKey
	ListenMultiaddrs         []string
	AnnounceMultiaddrs       []string
	BootstrapAddrs           []string
	ConnManager              cconnmgr.ConnManager
	EnableQUIC               bool
	EnableNATPortMap         bool
	EnableMDNS               bool
	EnablePubSubPeerExchange bool
	EnablePubSubFloodPublish bool
}

func setDefaults(conf *Config) {
	if len(conf.ListenMultiaddrs) == 0 {
		conf.ListenMultiaddrs = []string{"/ip4/0.0.0.0/tcp/0"}
	}
	if conf.ConnManager == nil {
		cm, err := connmgr.NewConnManager(256, 512, connmgr.WithGracePeriod(time.Second*120))
		if err != nil {
			panic(err)
		}
		conf.ConnManager = cm
	}
}

// Info contains public information about the libp2p peer.
type Info struct {
	ID        peer.ID
	PublicKey string
	Addresses []multiaddr.Multiaddr
}

// Peer wraps libp2p peer components for pubsub rpc.
type Peer struct {
	host      host.Host
	peer      *ipfslite.Peer
	ps        *ps.PubSub
	bootstrap []peer.AddrInfo
	finalizer *finalizer.Finalizer
}

// New returns a new Peer.
func New(conf Config) (*Peer, error) {
	setDefaults(&conf)

	listenAddr, err := parseMultiaddrs(conf.ListenMultiaddrs)
	if err != nil {
		return nil, fmt.Errorf("parsing listen addresses: %v", err)
	}
	announceAddrs, err := parseMultiaddrs(conf.AnnounceMultiaddrs)
	if err != nil {
		return nil, fmt.Errorf("parsing announce addresses: %v", err)
	}
	bootstrap, err := ipfsconfig.ParseBootstrapPeers(conf.BootstrapAddrs)
	if err != nil {
		return nil, fmt.Errorf("parsing bootstrap addresses: %v", err)
	}

	opts := []libp2p.Option{
		libp2p.ConnectionManager(conf.ConnManager),
		libp2p.DisableRelay(),
	}
	if len(announceAddrs) != 0 {
		opts = append(opts, libp2p.AddrsFactory(func([]multiaddr.Multiaddr) []multiaddr.Multiaddr {
			return announceAddrs
		}))
	}
	if conf.EnableNATPortMap {
		opts = append(opts, libp2p.NATPortMap())
	}

	fin := finalizer.NewFinalizer()
	ctx, cancel := context.WithCancel(context.Background())
	fin.Add(finalizer.NewContextCloser(cancel))

	// Setup ipfslite peerstore
	repoPath := filepath.Join(conf.RepoPath, "ipfslite")
	if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("making dir: %v", err)
	}
	dstore, err := badger.NewDatastore(repoPath, &badger.DefaultOptions)
	if err != nil {
		return nil, fin.Cleanupf("creating repo: %v", err)
	}
	fin.Add(dstore)
	pstore, err := pstoremem.NewPeerstore()
	if err != nil {
		return nil, fin.Cleanup(err)
	}
	fin.Add(pstore)
	opts = append(opts, libp2p.Peerstore(pstore))

	// Setup libp2p
	lhost, dht, err := ipfslite.SetupLibp2p(ctx, conf.PrivKey, nil, listenAddr, dstore, opts...)
	if err != nil {
		return nil, fin.Cleanupf("setting up libp2p", err)
	}
	fin.Add(lhost, dht)

	// Create ipfslite peer
	lpeer, err := ipfslite.New(ctx, dstore, nil, lhost, dht, nil)
	if err != nil {
		return nil, fin.Cleanupf("creating ipfslite peer", err)
	}

	// Setup pubsub
	gps, err := ps.NewGossipSub(
		ctx,
		lhost,
		ps.WithPeerExchange(conf.EnablePubSubPeerExchange),
		ps.WithFloodPublish(conf.EnablePubSubFloodPublish),
		ps.WithDirectPeers(bootstrap),
	)
	if err != nil {
		return nil, fin.Cleanupf("starting libp2p pubsub: %v", err)
	}

	log.Infof("marketpeer %s is online", lhost.ID())

	p := &Peer{
		host:      lhost,
		peer:      lpeer,
		ps:        gps,
		bootstrap: bootstrap,
		finalizer: fin,
	}

	if conf.EnableMDNS {
		if err := mdns.Start(ctx, p.host); err != nil {
			return nil, fin.Cleanupf("enabling mdns: %v", err)
		}
	}

	return p, nil
}

// Close the peer.
func (p *Peer) Close() error {
	return p.finalizer.Cleanup(nil)
}

// Host returns the peer host.
func (p *Peer) Host() host.Host {
	return p.host
}

// Info returns the peer's public information.
func (p *Peer) Info() (*Info, error) {
	var pkey string
	if pk := p.host.Peerstore().PubKey(p.host.ID()); pk != nil {
		pkb, err := crypto.MarshalPublicKey(pk)
		if err != nil {
			return nil, fmt.Errorf("marshaling public key: %s", err)
		}
		pkey = base64.StdEncoding.EncodeToString(pkb)
	}

	return &Info{
		p.host.ID(),
		pkey,
		p.host.Addrs(),
	}, nil
}

// ListPeers returns the peers the market peer currently connects to.
func (p *Peer) ListPeers() []peer.ID {
	return p.ps.ListPeers("")
}

// Bootstrap the market peer against Config.Bootstrap network peers.
// Some well-known network peers are included by default.
func (p *Peer) Bootstrap() {
	p.peer.Bootstrap(p.bootstrap)
	log.Info("peer was bootstapped")
}

// DAGService returns the underlying format.DAGService.
func (p *Peer) DAGService() format.DAGService {
	return p.peer
}

// BlockStore returns the underlying format.DAGService.
func (p *Peer) BlockStore() blockstore.Blockstore {
	return p.peer.BlockStore()
}

// NewTopic returns a new pubsub.Topic using the peer's host.
func (p *Peer) NewTopic(ctx context.Context, topic string, subscribe bool) (*rpc.Topic, error) {
	return rpc.NewTopic(ctx, p.ps, p.host.ID(), topic, subscribe)
}

func parseMultiaddrs(strs []string) ([]multiaddr.Multiaddr, error) {
	addrs := make([]multiaddr.Multiaddr, len(strs))
	for i, a := range strs {
		addr, err := multiaddr.NewMultiaddr(a)
		if err != nil {
			return nil, fmt.Errorf("parsing multiaddress: %v", err)
		}
		addrs[i] = addr
	}
	return addrs, nil
}
