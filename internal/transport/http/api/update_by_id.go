package api

import (
	"fmt"

	"github.com/elysiandb/elysian-gate/internal/balancer"
	"github.com/valyala/fasthttp"
)

func UpdateByIdController(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	if q := ctx.URI().QueryString(); len(q) > 0 {
		path += "?" + string(q)
	}

	status, body, err := balancer.SendWriteRequestToMaster("PUT", path, string(ctx.PostBody()))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadGateway)
		ctx.SetBody([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody([]byte(body))
}
