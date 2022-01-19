package main

import (
	"github.com/Zaba505/eventproc/action"

	"go.uber.org/zap"
)

// echoProcessor implements the action.ProcessorServer interface
// and simply echoes back any action content streamed to it.
type echoProcessor struct {
	action.UnimplementedProcessorServer
}

func (p *echoProcessor) ProcessActions(stream action.Processor_ProcessActionsServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			zap.L().Error("unexpected error when receiving action", zap.Error(err))
			return err
		}

		zap.L().Debug("action received", zap.String("id", req.GetId()))
		go p.sendResponse(stream, req)
	}
}

func (p *echoProcessor) sendResponse(stream action.Processor_ProcessActionsServer, req *action.ProcessorRequest) {
	act := req.GetAction()
	if act == nil {
		act = new(action.Action)
	}

	err := stream.Send(&action.ProcessorResponse{
		Id: req.GetId(),
		Body: &action.ProcessorResponse_Content{
			Content: act.GetPayload(),
		},
	})
	if err != nil {
		zap.L().Error("unexpected error when sending response", zap.Error(err))
		return
	}

	zap.L().Debug("sent response", zap.String("id", req.GetId()))
}
