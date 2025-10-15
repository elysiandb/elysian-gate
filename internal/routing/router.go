package routing

import (
	"github.com/elysiandb/elysian-gate/internal/transport/http/api"
	"github.com/fasthttp/router"
)

func RegisterRoutes(r *router.Router) {
	r.POST("/api/{entity}", api.CreateController)
	r.GET("/api/{entity}/{id}", api.GetByIdController)
	r.GET("/api/{entity}", api.ListController)
	r.DELETE("/api/{entity}/{id}", api.DeleteByIdController)
	r.PUT("/api/{entity}/{id}", api.UpdateByIdController)
	r.DELETE("/api/{entity}", api.DestroyController)
}
