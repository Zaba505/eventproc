package event

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Sink is a service exposed to clients for sending events to.
type Sink struct {
	UnimplementedEventSinkServer

	clientMap map[string]*Mux
	cfg       *viper.Viper
}

func NewSink(cfg *viper.Viper, clientMap map[string]*Mux) *Sink {
	return &Sink{
		clientMap: clientMap,
		cfg:       cfg,
	}
}

const PayloadRegexKey = "payloadRegexKey"

func (s *Sink) ProcessEvent(ctx context.Context, req *EventRequest) (*EventResponse, error) {
	ev := req.GetEvent()
	if ev == nil {
		return nil, status.Error(codes.InvalidArgument, "event must not be nil")
	}

	v := s.cfg.Get(PayloadRegexKey)
	if v == nil {
		return nil, status.Error(codes.Internal, "no payload mapper config provided")
	}

	mapper, ok := v.(map[Event_Type]string)
	if !ok {
		return nil, status.Error(codes.Internal, "incorrect type provided for payload mapper config")
	}

	var clientProcId string
	for evType, clientId := range mapper {
		if evType == ev.GetType() {
			clientProcId = clientId
			break
		}
	}
	if clientProcId == "" {
		return nil, status.Error(codes.Unimplemented, "unknown payload type")
	}

	client, ok := s.clientMap[clientProcId]
	if !ok {
		return nil, status.Errorf(codes.Unimplemented, "no event processor with id: %s", clientProcId)
	}

	respEv := client.SendEvent(ctx, ev)
	if respEv == nil {
		// TODO: come up with better strategy for this
		return new(EventResponse), nil
	}

	resp := &EventResponse{
		Body: &EventResponse_Event{
			Event: respEv,
		},
	}

	return resp, nil
}

// EchoProcessor implements the EventProcessorServer interface
// and simply echoes back any events stream to it.
type EchoProcessor struct {
	UnimplementedEventProcessorServer
}

func (p *EchoProcessor) ProcessEvents(stream EventProcessor_ProcessEventsServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			zap.L().Error("unexpected error when receiving event", zap.Error(err))
			return err
		}

		zap.L().Debug("event received")
		go p.sendResponse(stream, req)
	}
}

func (p *EchoProcessor) sendResponse(stream EventProcessor_ProcessEventsServer, req *ProcessorRequest) {
	err := stream.Send(&ProcessorResponse{
		Id: req.GetId(),
		Response: &EventResponse{
			Body: &EventResponse_Event{
				Event: req.GetEvent(),
			},
		},
	})
	if err != nil {
		zap.L().Error("unexpected error when sending response", zap.Error(err))
		return
	}

	zap.L().Debug("sent response", zap.String("id", req.GetId()))
}
