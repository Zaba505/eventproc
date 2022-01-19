package action

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Gateway is a service exposed to clients for sending actions to.
type Gateway struct {
	UnimplementedGatewayServer

	clientMap map[string]*Mux
	cfg       *viper.Viper
}

func NewGateway(cfg *viper.Viper, clientMap map[string]*Mux) *Gateway {
	return &Gateway{
		clientMap: clientMap,
		cfg:       cfg,
	}
}

const TypeToProcessorMapKey = "actionToProcessorKey"

func (s *Gateway) ProcessAction(ctx context.Context, req *ActionRequest) (*ActionResponse, error) {
	act := req.GetAction()
	if act == nil {
		zap.L().Error("action must not be nil")
		return nil, status.Error(codes.InvalidArgument, "action must not be nil")
	}

	v := s.cfg.Get(TypeToProcessorMapKey)
	if v == nil {
		zap.L().Error("no action type mapper config provided")
		return nil, status.Error(codes.Internal, "no action type mapper config provided")
	}

	mapper, ok := v.(map[Action_Type]string)
	if !ok {
		zap.L().Error("incorrect type provided for action type mapper config")
		return nil, status.Error(codes.Internal, "incorrect type provided for action type mapper config")
	}

	var clientProcId string
	for actType, clientId := range mapper {
		if actType == act.GetType() {
			clientProcId = clientId
			break
		}
	}
	if clientProcId == "" {
		zap.L().Error("unknown payload type", zap.String("type", act.GetType().String()))
		return nil, status.Error(codes.Unimplemented, "unknown action type")
	}

	client, ok := s.clientMap[clientProcId]
	if !ok {
		zap.L().Error("no processor found", zap.String("processorId", clientProcId))
		return nil, status.Errorf(codes.Unimplemented, "no processor found")
	}

	respAction := client.SendAction(ctx, act)
	if respAction == nil {
		zap.L().Debug("received nil response action")
		// TODO: come up with better strategy for this
		return new(ActionResponse), nil
	}

	resp := &ActionResponse{
		Body: &ActionResponse_Content{
			Content: respAction.GetPayload(),
		},
	}

	return resp, nil
}
