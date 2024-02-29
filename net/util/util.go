package util

import (
	"sync"

	"github.com/dcnetio/gothreads-lib/core/thread"
	apipb "github.com/dcnetio/gothreads-lib/net/api/pb"
	netpb "github.com/dcnetio/gothreads-lib/net/pb"
	ma "github.com/multiformats/go-multiaddr"
)

func RecFromServiceRec(r *netpb.Log_Record) *apipb.Record {
	return &apipb.Record{
		RecordNode: r.RecordNode,
		EventNode:  r.EventNode,
		HeaderNode: r.HeaderNode,
		BodyNode:   r.BodyNode,
	}
}

func RecToServiceRec(r *apipb.Record) *netpb.Log_Record {
	return &netpb.Log_Record{
		RecordNode: r.RecordNode,
		EventNode:  r.EventNode,
		HeaderNode: r.HeaderNode,
		BodyNode:   r.BodyNode,
	}
}

// LogToProto returns a proto log from a thread log.
func LogToProto(l thread.LogInfo) *netpb.Log {
	return &netpb.Log{
		ID:      &netpb.ProtoPeerID{ID: l.ID},
		PubKey:  &netpb.ProtoPubKey{PubKey: l.PubKey},
		Addrs:   AddrsToProto(l.Addrs),
		Head:    &netpb.ProtoCid{Cid: l.Head.ID},
		Counter: l.Head.Counter,
	}
}

// LogFromProto returns a thread log from a proto log.
func LogFromProto(l *netpb.Log) thread.LogInfo {
	return thread.LogInfo{
		ID:     l.ID.ID,
		PubKey: l.PubKey.PubKey,
		Addrs:  AddrsFromProto(l.Addrs),
		Head: thread.Head{
			ID:      l.Head.Cid,
			Counter: l.Counter,
		},
	}
}

func AddrsToProto(mas []ma.Multiaddr) []netpb.ProtoAddr {
	pas := make([]netpb.ProtoAddr, len(mas))
	for i, a := range mas {
		pas[i] = netpb.ProtoAddr{Multiaddr: a}
	}
	return pas
}

func AddrsFromProto(pa []netpb.ProtoAddr) []ma.Multiaddr {
	mas := make([]ma.Multiaddr, len(pa))
	for i, a := range pa {
		mas[i] = a.Multiaddr
	}
	return mas
}

func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{inner: make(chan struct{}, capacity)}
}

type Semaphore struct {
	inner chan struct{}
}

// Blocking acquire
func (s *Semaphore) Acquire() {
	s.inner <- struct{}{}
}

// Non-blocking acquire
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.inner <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Semaphore) Release() {
	select {
	case <-s.inner:
	default:
		panic("thread semaphore inconsistency: release before acquire!")
	}
}

type SemaphoreKey interface {
	Key() string
}

func NewSemaphorePool(semaCap int) *SemaphorePool {
	return &SemaphorePool{ss: make(map[string]*Semaphore), semaCap: semaCap}
}

type SemaphorePool struct {
	ss      map[string]*Semaphore
	semaCap int
	mu      sync.Mutex
}

func (p *SemaphorePool) Get(k SemaphoreKey) *Semaphore {
	var (
		s     *Semaphore
		exist bool
		key   = k.Key()
	)

	p.mu.Lock()
	if s, exist = p.ss[key]; !exist {
		s = NewSemaphore(p.semaCap)
		p.ss[key] = s
	}
	p.mu.Unlock()

	return s
}

// Stop holds all semaphores in the pool
func (p *SemaphorePool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// grab all semaphores and hold
	for _, s := range p.ss {
		s.Acquire()
	}
}
