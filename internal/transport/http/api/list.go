package api

import (
	"fmt"

	"github.com/elysiandb/elysian-gate/internal/balancer"
	"github.com/elysiandb/elysian-gate/internal/forward"
	"github.com/valyala/fasthttp"
)

func ListController(ctx *fasthttp.RequestCtx) {
	node := balancer.GetReadRequestNode()
	if node == nil {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetBody([]byte("No available node"))
		return
	}

	url := fmt.Sprintf("http://%s:%d%s", node.HTTP.Host, node.HTTP.Port, string(ctx.Path()))
	if q := ctx.URI().QueryString(); len(q) > 0 {
		url += "?" + string(q)
	}

	status, body, err := forward.ForwardRequest("GET", url, "")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadGateway)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody([]byte(body))
}
