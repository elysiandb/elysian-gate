package api

import (
	"encoding/json"
	"fmt"

	"github.com/elysiandb/elysian-gate/internal/balancer"
	"github.com/elysiandb/elysian-gate/internal/forward"
	"github.com/valyala/fasthttp"
)

func GetByIdController(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	if q := ctx.URI().QueryString(); len(q) > 0 {
		path += "?" + string(q)
	}

	node := balancer.GetReadRequestNode()
	if node == nil {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetBody([]byte(`{"error":"no available node"}`))
		return
	}

	url := fmt.Sprintf("http://%s:%d%s", node.HTTP.Host, node.HTTP.Port, path)
	status, body, err := forward.ForwardRequest("GET", url, "")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadGateway)
		ctx.SetBody([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		return
	}

	var formatted any
	if json.Unmarshal([]byte(body), &formatted) == nil {
		data, _ := json.MarshalIndent(formatted, "", "  ")
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(status)
		ctx.SetBody(data)
	} else {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(status)
		ctx.SetBody([]byte(body))
	}
}
