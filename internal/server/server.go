package server

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"
)

const timeout = 10 * time.Second

type server struct {
	httpServer *httpServer
	gRPCServer *grpcServer
}

func (s *server) ServerRun() error {
	// check if any server was created
	if s.httpServer == nil && s.gRPCServer == nil {
		fmt.Println("nothing to run!")
		return nil
	}

	idleConnectionsClosed := make(chan struct{})
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	// listen for stop signals
	go func() {
		<-ctx.Done()

		// finish HTTP server
		s.httpServer.Shutdown()

		// finish gRPC server
		s.gRPCServer.Shutdown()

		close(idleConnectionsClosed)
	}()

	// launch all created servers
	if s.httpServer != nil {
		fmt.Println("Launching HTTP server")
		go s.httpServer.RunServer()
	}
	if s.gRPCServer != nil {
		fmt.Println("Launching GRPC server")
		go s.gRPCServer.RunServer()
	}

	<-idleConnectionsClosed
	fmt.Println("server Shutdown gracefully")

	return nil
}
