package action

import (
	"bytes"
	"io/ioutil"
	"sync"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

var bufPool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func NewFastHTTPHandler(g *Gateway) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		b := bufPool.Get().(*bytes.Buffer)
		defer func() {
			b.Reset()
			bufPool.Put(b)
		}()

		err := ctx.Request.BodyWriteTo(b)
		if err != nil {
			zap.L().Error("unexpected error when reading request body")
			return
		}

		act, err := decodeActionFromJSON(ioutil.NopCloser(b))
		if err != nil {
			zap.L().Error("unexpected error when decoding action from json")
			return
		}

		actReq := &ActionRequest{
			Action: &act,
		}
		resp, err := g.ProcessAction(ctx, actReq)
		if err != nil {
			zap.L().Error("unexpected error when processing event", zap.Error(err))
			return
		}

		switch x := resp.GetBody().(type) {
		case *ActionResponse_Content:
			ctx.SetStatusCode(200)
			ctx.Success("application/json", x.Content)
		case *ActionResponse_WasProcessed:
			ctx.SetStatusCode(204)
		}
	}
}
