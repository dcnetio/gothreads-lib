package net

import (
	"context"

	"github.com/dcnetio/gothreads-lib/core/thread"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/grpc"
)

type ThreadExternal interface {
	// GetPeers get peers that thread stored.
	GetPeers(ctx context.Context, id thread.ID) ([]peer.ID, error)
	ConnectToPeer(ctx context.Context, h host.Host, pid peer.ID) bool
}

type ServiceExternal interface {
	RegisterServiceServers(h host.Host, s *grpc.Server) (err error)
}
