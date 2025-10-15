package api

import (
	"github.com/elysiandb/elysian-gate/internal/balancer"
	"github.com/valyala/fasthttp"
)

func CreateController(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	if q := ctx.URI().QueryString(); len(q) > 0 {
		path += "?" + string(q)
	}

	status, body, err := balancer.SendWriteRequestToMaster("POST", path, string(ctx.PostBody()))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadGateway)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody([]byte(body))
}
