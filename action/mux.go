package action

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

// Mux multiplexes a Actions over a single Processor_ProcessActionsClient.
type Mux struct {
	stream Processor_ProcessActionsClient
	cache  *cache.Cache
}

type MuxOption func(*Mux)

func WithCache(cache *cache.Cache) MuxOption {
	return func(m *Mux) {
		m.cache = cache
	}
}

func NewMux(stream Processor_ProcessActionsClient, opts ...MuxOption) *Mux {
	m := &Mux{
		stream: stream,
		cache:  cache.New(1*time.Second, 2*time.Second),
	}

	for _, opt := range opts {
		opt(m)
	}

	go m.receiveActions()

	return m
}

func (m *Mux) SendAction(ctx context.Context, act *Action) *Action {
	uid, err := uuid.NewRandom()
	if err != nil {
		return nil
	}

	id := uid.String()
	req := &ProcessorRequest{
		Id:     id,
		Action: act,
	}
	responseCh := make(chan *ProcessorResponse, 1)

	m.set(ctx, id, responseCh)
	err = m.sendAction(req)
	if err != nil {
		zap.L().Error("unexpected error when sending action to processor", zap.Error(err))
		return nil
	}
	zap.L().Debug("sent request to processor", zap.String("id", id))

	select {
	case <-ctx.Done():
		// TODO: return error here
		return nil
	case resp := <-responseCh:
		respId := resp.GetId()

		zap.L().Debug("received response from processor goroutine", zap.String("id", respId))
		if id != respId {
			zap.L().Error(
				"received processor response id doesn't match processor request id",
				zap.String("reqId", id),
				zap.String("respId", respId),
			)
			return nil
		}

		switch x := resp.GetBody().(type) {
		case *ProcessorResponse_Content:
			return &Action{
				Payload: x.Content,
			}
		case *ProcessorResponse_WasProcessed:
			return nil
		default:
			zap.L().Error("unexpected processor response body", zap.String("id", respId))
			return nil
		}
	}
}

func (m *Mux) set(ctx context.Context, id string, responseCh chan<- *ProcessorResponse) {
	// zero expiration means use cache defined default expiration
	var expiration time.Duration

	deadline, ok := ctx.Deadline()
	if ok {
		// if expiration is negative the item will be cached forever
		if expiration = deadline.Sub(time.Now()); expiration < 0 {
			expiration = 0
		}
	}

	m.cache.Set(id, responseCh, expiration)
}

func (m *Mux) sendAction(req *ProcessorRequest) error {
	return m.stream.Send(req)
}

func (m *Mux) receiveActions() {
	for {
		resp, err := m.stream.Recv()
		if err != nil {
			// TODO: determine what error is
			zap.L().Error("unexpected error when receiving processor response", zap.Error(err))
			continue
		}

		go func() {
			v, ok := m.cache.Get(resp.GetId())
			if !ok {
				// cache expired so can't relay response to client
				return
			}

			respCh, ok := v.(chan<- *ProcessorResponse)
			if !ok {
				return
			}

			respCh <- resp
			close(respCh)
		}()
	}
}
