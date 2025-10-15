package api

import (
	"fmt"

	"github.com/elysiandb/elysian-gate/internal/balancer"
	"github.com/valyala/fasthttp"
)

func DestroyController(ctx *fasthttp.RequestCtx) {
	status, body, err := balancer.SendWriteRequestToMaster("DELETE", string(ctx.Path()), "")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadGateway)
		ctx.SetBody([]byte(fmt.Sprintf("Error forwarding delete: %v", err)))
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody([]byte(body))
}
