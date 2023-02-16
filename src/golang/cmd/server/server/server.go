package server

import (
	"context"
	"net/http"

	"github.com/aqueducthq/aqueduct/cmd/server/handler"
	"github.com/aqueducthq/aqueduct/cmd/server/routes"
)

type Server interface {
	// `Handlers()` returns a map of route:handler which the server handles.
	Handlers() map[string]handler.Handler
	// `AddHandler()` specifies how the server initialize with a particular handler.
	// It's wrapped in `AddAllHandler()` helper and should be called before `Run()`.
	AddHandler(route string, handler handler.Handler)
	// `Log()` specifies how an outcome of http request is logged.
	// Args:
	//	`key`: additional identifier for the log, for example, the name of the server service.
	//	`req`: the request object
	//	`statusCode`: status of the request
	//	`err`: any error generated when handling the request, which could be nil when successful.
	Log(ctx context.Context, key string, req *http.Request, statusCode int, err error)
	// `Run()` should start the server service and handle requests. This will be called in `main()`.
	Run(expose bool)
}

func GetAllHeaders(server Server) []string {
	headers := []string{
		"Accept",
		"Authorization",
		"Content-Type",
		"X-CSRF-Token",
		routes.ApiKeyHeader,
	}

	headersSet := map[string]bool{}

	for _, h := range headers {
		headersSet[h] = true
	}

	for _, handler := range server.Handlers() {
		for _, h := range handler.Headers() {
			if _, ok := headersSet[h]; !ok {
				headersSet[h] = true
				headers = append(headers, h)
			}
		}
	}

	return headers
}

func AddAllHandlers(server Server) {
	for route, handler := range server.Handlers() {
		server.AddHandler(route, handler)
	}
}
