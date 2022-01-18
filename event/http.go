package event

import (
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"
	json "google.golang.org/protobuf/encoding/protojson"
)

// NewHTTPHandler wraps a EventSink service to expose it over an HTTP based API.
func NewHTTPHandler(s *Sink) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ev, err := decodeEvent(req)
		if err != nil {
			zap.L().Error("unexpected error when decoding request body", zap.Error(err))
			http.Error(w, "unexpected error when decoding request body", 500)
			return
		}

		evReq := &EventRequest{
			Event: &ev,
		}
		resp, err := s.ProcessEvent(req.Context(), evReq)
		if err != nil {
			zap.L().Error("unexpected error when processing event", zap.Error(err))
			http.Error(w, "unexpected error when processing event", 500)
			return
		}

		switch x := resp.GetBody().(type) {
		case *EventResponse_Event:
			respEv := x.Event
			if respEv == nil {
				respEv = new(Event)
			}

			b, err := json.Marshal(respEv)
			if err != nil {
				zap.L().Error("unexpected error when marshalling event", zap.Error(err))
				http.Error(w, "unexpected error when marshalling event", 500)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)

			_, err = w.Write(b)
			if err != nil {
				zap.L().Error("unexpected error when writing event to response body", zap.Error(err))
				return
			}
		case *EventResponse_EventConsumed:
			w.WriteHeader(204)
		}
	}
}

func decodeEvent(req *http.Request) (ev Event, err error) {
	defer req.Body.Close()

	var b []byte
	b, err = ioutil.ReadAll(req.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &ev)
	return
}
