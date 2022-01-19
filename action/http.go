package action

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"
)

// NewHTTPHandler wraps a Gateway service to expose it over an HTTP based API.
func NewHTTPHandler(s *Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		act, err := decodeActionFromJSON(req.Body)
		if err != nil {
			zap.L().Error("unexpected error when decoding request body", zap.Error(err))
			http.Error(w, "unexpected error when decoding request body", 500)
			return
		}

		actReq := &ActionRequest{
			Action: &act,
		}
		resp, err := s.ProcessAction(req.Context(), actReq)
		if err != nil {
			zap.L().Error("unexpected error when processing event", zap.Error(err))
			http.Error(w, "unexpected error when processing event", 500)
			return
		}

		switch x := resp.GetBody().(type) {
		case *ActionResponse_Content:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)

			_, err = w.Write(x.Content)
			if err != nil {
				zap.L().Error("unexpected error when writing event to response body", zap.Error(err))
				return
			}
		case *ActionResponse_WasProcessed:
			w.WriteHeader(204)
		}
	}
}

func decodeActionFromJSON(r io.ReadCloser) (act Action, err error) {
	defer r.Close()

	var b []byte
	b, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}

	var v map[string]interface{}
	err = json.Unmarshal(b, &v)
	if err != nil {
		return
	}

	typ, ok := v["type"]
	if !ok {
		err = errors.New("expected field type to be set")
		return
	}

	payload, ok := v["payload"]
	if !ok {
		err = errors.New("expected field payload to be set")
		return
	}

	act.Type = getType(typ)
	act.Payload, err = json.Marshal(payload)
	return
}

func getType(typ interface{}) Action_Type {
	switch x := typ.(type) {
	case string:
		return Action_Type(Action_Type_value[x])
	case int:
		return Action_Type(x)
	default:
		return -1
	}
}
