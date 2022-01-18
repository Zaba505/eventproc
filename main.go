package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/Zaba505/eventproc/event"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var processorAddr string

func init() {
	flag.StringVar(&processorAddr, "processor", "", "specify the event processor service address")
	flag.Parse()

	if processorAddr == "" {
		panic("must provide an address for a backend event processor to stream incoming events to.")
	}

	// map HELLO event types to the provided processor address
	viper.Set(event.PayloadRegexKey, map[event.Event_Type]string{
		event.Event_HELLO: processorAddr,
	})
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer zap.ReplaceGlobals(logger)()

	// manually handle OS signal interrupts
	pctx := context.Background()
	ctx, stop := signal.NotifyContext(pctx, os.Interrupt)
	defer stop()

	// dial a gRPC based EventProcessor backend given its address.
	client, err := dialEventProcessor(processorAddr)
	if err != nil {
		zap.L().Error("unexpected error when dialing event processor backend", zap.Error(err))
		return
	}

	// activate gRPC stream with backend EventProcessor
	processor, err := client.ProcessEvents(ctx)
	if err != nil {
		zap.L().Error("unexpected error when calling event processor", zap.Error(err))
		return
	}

	// construct EventSink which lies at the heart of the main program
	s := event.NewSink(viper.GetViper(), map[string]*event.Mux{
		processorAddr: event.NewMux(processor),
	})

	// fire up HTTP server
	srv := buildEventSinkHTTPServer(s)
	errChan := startHTTPServer(ctx, srv)

	select {
	case <-ctx.Done():
		srv.Shutdown(pctx)
		stop()
		<-errChan
	case err := <-errChan:
		zap.L().Error("received unexpected error from server", zap.Error(err))
		stop()
		<-errChan
	}
}

// dial a gRPC based EventProcessor backend given its address.
func dialEventProcessor(addr string) (event.EventProcessorClient, error) {
	cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return event.NewEventProcessorClient(cc), nil
}

// build REST style API around event.Sink
func buildEventSinkHTTPServer(s *event.Sink) *http.Server {
	handler := event.NewHTTPHandler(s)

	router := mux.NewRouter()
	router.
		Methods(http.MethodPost).
		Path("/event").
		Handler(handler)

	return &http.Server{
		Handler: router,
	}
}

// start http server concurrently
func startHTTPServer(ctx context.Context, srv *http.Server) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		var lcfg net.ListenConfig
		ls, err := lcfg.Listen(ctx, "tcp", ":0")
		if err != nil {
			errChan <- err
			return
		}
		zap.L().Info("starting http server...", zap.String("id", ls.Addr().String()))

		err = srv.Serve(ls)
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
			return
		}
	}()

	return errChan
}
