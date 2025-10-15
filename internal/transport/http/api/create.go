package api

import (
	"net/http"

	"github.com/valyala/fasthttp"
)

func CreateController(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(http.StatusOK)
}
