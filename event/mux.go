package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

// Mux multiplexes a single EventProcessor_EventProcessorClient over
// multiple goroutines.
type Mux struct {
	stream  EventProcessor_ProcessEventsClient
	eventCh chan streamRequest
	cache   *cache.Cache
}

type MuxOption func(*Mux)

func WithCache(cache *cache.Cache) MuxOption {
	return func(m *Mux) {
		m.cache = cache
	}
}

type streamRequest struct {
	event        *Event
	responseChan chan<- *Event
}

func NewMux(stream EventProcessor_ProcessEventsClient, opts ...MuxOption) *Mux {
	m := &Mux{
		stream:  stream,
		eventCh: make(chan streamRequest, 5),
		cache:   cache.New(5*time.Second, 10*time.Second),
	}

	for _, opt := range opts {
		opt(m)
	}

	go m.sendEvents()
	go m.receiveEvents()

	return m
}

func (m *Mux) SendEvent(ctx context.Context, ev *Event) *Event {
	responseCh := make(chan *Event, 1)

	req := streamRequest{
		event:        ev,
		responseChan: responseCh,
	}

	// TODO: pass Context.Deadline() to eventCh for setting TTL in cache
	select {
	case <-ctx.Done():
		// TODO: return error here
		return nil
	case m.eventCh <- req:
		zap.L().Debug("mux: sent event to processor goroutine")
	}

	select {
	case <-ctx.Done():
		// TODO: return error here
		return nil
	case resp := <-responseCh:
		zap.L().Debug("mux: received response from processor goroutine")
		return resp
	}
}

func (m *Mux) sendEvents() {
	for sreq := range m.eventCh {
		uid, err := uuid.NewRandom()
		if err != nil {
			close(sreq.responseChan)
			continue
		}

		id := uid.String()
		req := &ProcessorRequest{
			Id:    id,
			Event: sreq.event,
		}
		m.cache.SetDefault(id, sreq.responseChan)

		err = m.stream.Send(req)
		if err != nil {
			close(sreq.responseChan)
			continue
		}
	}
}

func (m *Mux) receiveEvents() {
	for {
		resp, err := m.stream.Recv()
		if err != nil {
			// TODO: determine what error is
			zap.L().Error("unexpected error when receiving on stream", zap.Error(err))
			continue
		}

		v, ok := m.cache.Get(resp.GetId())
		if !ok {
			// cache expired so can't relay event to client
			continue
		}

		respCh, ok := v.(chan<- *Event)
		if !ok {
			continue
		}

		r := resp.GetResponse()
		if r == nil {
			close(respCh)
			continue
		}

		if ev := r.GetEvent(); ev != nil {
			respCh <- ev
		}
		close(respCh)
	}
}
