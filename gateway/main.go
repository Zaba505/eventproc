package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/Zaba505/eventproc/action"

	"github.com/fasthttp/router"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var processorAddr string
var logLevel zapcore.Level

func init() {
	flag.StringVar(&processorAddr, "processor", ":12345", "specify the event processor service address")
	flag.Var(&logLevel, "log-level", "Set log level")
	flag.Parse()

	if processorAddr == "" {
		panic("must provide an address for a backend event processor to stream incoming events to.")
	}

	// map HELLO event types to the provided processor address
	viper.Set(action.TypeToProcessorMapKey, map[action.Action_Type]string{
		action.Action_HELLO: processorAddr,
	})
}

func main() {
	logger, err := zap.NewDevelopment(zap.IncreaseLevel(logLevel))
	if err != nil {
		panic(err)
	}

	defer zap.ReplaceGlobals(logger)()

	// manually handle OS signal interrupts
	pctx := context.Background()
	ctx, stop := signal.NotifyContext(pctx, os.Interrupt)
	defer stop()

	// dial a gRPC based Processor backend given its address.
	client, err := dialEventProcessor(processorAddr)
	if err != nil {
		zap.L().Error("unexpected error when dialing event processor backend", zap.Error(err))
		return
	}

	// activate gRPC stream with backend Processor
	processor, err := client.ProcessActions(ctx)
	if err != nil {
		zap.L().Error("unexpected error when calling event processor", zap.Error(err))
		return
	}

	// construct EventSink which lies at the heart of the main program
	s := action.NewGateway(viper.GetViper(), map[string]*action.Mux{
		processorAddr: action.NewMux(processor),
	})

	// fire up standard library HTTP server
	httpServer := buildActionHTTPGatewayServer(s)
	httpErrChan := startHTTPServer(ctx, httpServer)

	// fire up fasthttp HTTP server
	fastHttpServer := buildFastHTTPServer(s)
	fastErrChan := startFastHTTPServer(fastHttpServer)

	// fire up gRPC server
	grpcServer := grpc.NewServer()
	action.RegisterGatewayServer(grpcServer, s)
	grpcErrChan := startGRPCServer(grpcServer)

	select {
	case <-ctx.Done():
		httpServer.Shutdown(pctx)
		grpcServer.GracefulStop()
		fastHttpServer.Shutdown()
		stop()
	case err := <-httpErrChan:
		zap.L().Error("received unexpected error from http server", zap.Error(err))
		stop()
		grpcServer.GracefulStop()
		fastHttpServer.Shutdown()
	case err := <-grpcErrChan:
		zap.L().Error("received unexpected error from grpc server", zap.Error(err))
		stop()
		httpServer.Shutdown(pctx)
		fastHttpServer.Shutdown()
	case err := <-fastErrChan:
		zap.L().Error("received unexpected error from fasthttp server", zap.Error(err))
		stop()
		httpServer.Shutdown(pctx)
		grpcServer.GracefulStop()
	}

	// make sure both gRPC and HTTP server goroutine are done executing
	<-httpErrChan
	<-grpcErrChan
}

// dial a gRPC based EventProcessor backend given its address.
func dialEventProcessor(addr string) (action.ProcessorClient, error) {
	cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return action.NewProcessorClient(cc), nil
}

// build REST style API around action.Gateway
func buildActionHTTPGatewayServer(s *action.Gateway) *http.Server {
	handler := action.NewHTTPHandler(s)

	router := mux.NewRouter()
	router.
		Methods(http.MethodPost).
		Path("/action").
		Handler(handler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	return srv
}

// start http server concurrently
func startHTTPServer(ctx context.Context, srv *http.Server) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		zap.L().Info("starting http server...", zap.String("addr", srv.Addr))
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
			return
		}
	}()

	return errChan
}

func buildFastHTTPServer(g *action.Gateway) *fasthttp.Server {
	r := router.New()

	r.POST("/action", action.NewFastHTTPHandler(g))

	return &fasthttp.Server{
		Handler: r.Handler,
	}
}

func startFastHTTPServer(srv *fasthttp.Server) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		zap.L().Debug("starting fasthttp server")
		if err := srv.ListenAndServe(":8181"); err != nil {
			errChan <- err
		}
	}()

	return errChan
}

// start grpc server concurrently
func startGRPCServer(srv *grpc.Server) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		ls, err := net.Listen("tcp", ":9090")
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

	return errChan
}
