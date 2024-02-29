package api

import (
	"errors"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/dcnetio/gothreads-lib/common"
	pb "github.com/dcnetio/gothreads-lib/net/api/pb"
	"github.com/dcnetio/gothreads-lib/util"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
)

// CreateTestService creates a test network API gRPC service for test purpose.
// It uses either the addr passed in as host addr, or pick an available local addr if it is empty
func CreateTestService(addr string, debug bool) (hostAddr ma.Multiaddr, gRPCAddr ma.Multiaddr, stop func(), err error) {
	time.Sleep(time.Second * time.Duration(rand.Intn(5)))
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return
	}
	if addr == "" {
		hostAddr = util.FreeLocalAddr()
	} else {
		hostAddr, _ = ma.NewMultiaddr(addr)
	}
	n, err := common.DefaultNetwork(
		common.WithNetBadgerPersistence(dir),
		common.WithNetHostAddr(hostAddr),
		common.WithNetPubSub(true),
		common.WithNetDebug(debug),
	)
	if err != nil {
		return
	}
	service, err := NewService(n, Config{
		Debug: debug,
	})
	if err != nil {
		return
	}
	gRPCAddr = util.FreeLocalAddr()
	target, err := util.TCPAddrFromMultiAddr(gRPCAddr)
	if err != nil {
		return
	}
	server := grpc.NewServer()
	listener, err := net.Listen("tcp", target)
	if err != nil {
		return
	}
	go func() {

		pb.RegisterAPIServer(server, service)
		if err := server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("serve error: %v", err)
		}
	}()

	return hostAddr, gRPCAddr, func() {
		util.StopGRPCServer(server)
		if err := n.Close(); err != nil {
			return
		}
		_ = os.RemoveAll(dir)
	}, nil
}
