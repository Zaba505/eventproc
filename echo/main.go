package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"

	"github.com/Zaba505/eventproc/action"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

var logLevel zapcore.Level

func init() {
	flag.Var(&logLevel, "log-level", "Set log level")
	flag.Parse()
}

func main() {
	logger, err := zap.NewDevelopment(zap.IncreaseLevel(logLevel))
	if err != nil {
		panic(err)
	}

	defer zap.ReplaceGlobals(logger)()

	srv := grpc.NewServer()
	action.RegisterProcessorServer(srv, new(echoProcessor))

	pctx := context.Background()
	ctx, stop := signal.NotifyContext(pctx, os.Interrupt)
	defer stop()

	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)

		ls, err := net.Listen("tcp", ":12345")
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
