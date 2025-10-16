// ---- /home/taymik/Projects/elysian-gate/internal/api/list.go ----
package api

import (
	"github.com/elysiandb/elysian-gate/internal/balancer"
	"github.com/valyala/fasthttp"
)

func ListController(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	query := string(ctx.URI().QueryString())

	status, body, _ := balancer.SendReadRequest(path, query)
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	ctx.SetBody(body)
}
