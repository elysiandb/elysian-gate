package boot

import (
	"fmt"
	"log"
	"time"

	"github.com/elysiandb/elysian-gate/internal/configuration"
	"github.com/elysiandb/elysian-gate/internal/routing"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var server *fasthttp.Server

func InitHTTP() {
	r := router.New()
	routing.RegisterRoutes(r)

	server = &fasthttp.Server{
		Handler:      r.Handler,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Name:         "Elysiangate",
	}

	go func() {
		host := configuration.Config.Gateway.HTTP.Host
		port := configuration.Config.Gateway.HTTP.Port
		log.Printf("Starting ElysianGate HTTP server on %s:%d", host, port)

		if err := server.ListenAndServe(fmt.Sprintf("%s:%d", host, port)); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()
}
