package routing

import (
	"github.com/elysiandb/elysian-gate/internal/transport/http/api"
	"github.com/fasthttp/router"
)

func RegisterRoutes(r *router.Router) {
	r.POST("/api/{entity}", api.CreateController)
}
