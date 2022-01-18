package main

import (
	"context"
	"net"
	"os"
	"os/signal"

	"github.com/Zaba505/eventproc/event"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer zap.ReplaceGlobals(logger)()

	srv := grpc.NewServer()
	event.RegisterEventProcessorServer(srv, new(event.EchoProcessor))

	pctx := context.Background()
	ctx, stop := signal.NotifyContext(pctx, os.Interrupt)
	defer stop()

	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)

		ls, err := net.Listen("tcp", ":0")
		if err != nil {
			errChan <- err
			return
		}
		zap.L().Info("starting grpc server...", zap.String("addr", ls.Addr().String()))

		err = srv.Serve(ls)
		if err != nil {
			errChan <- err
			return
		}
	}()

	select {
	case <-ctx.Done():
		srv.GracefulStop()
	case err := <-errChan:
		zap.L().Error("unexpected error from grpc server", zap.Error(err))
	}
	<-errChan
}
