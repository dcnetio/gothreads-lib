package queue

import (
	"context"

	"github.com/dcnetio/gothreads-lib/core/thread"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"
)

var log = logging.Logger("netqueue")

type (
	PeerCall func(context.Context, peer.ID, thread.ID) error

	CallQueue interface {
		// Make call immediately and synchronously return its result.
		Call(p peer.ID, t thread.ID, c PeerCall) error

		// Schedule call to be invoked later.
		Schedule(p peer.ID, t thread.ID, priority int, c PeerCall) bool
	}
)

type (
	ThreadPack struct {
		Peer    peer.ID
		Threads []thread.ID
	}

	ThreadPacker interface {
		// Add thread to peer's queue
		Add(pid peer.ID, tid thread.ID)

		// Start packing incoming thread requests
		Run() <-chan ThreadPack
	}
)
