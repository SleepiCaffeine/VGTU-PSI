// Package api exposes the services over HTTP.
package api

import (
	"io/fs"
	"net/http"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/api/gen"
)

const apiBasePath = "/api"

// Options configures the HTTP layer.
type Options struct {
	DemoMode bool
	Static   fs.FS
}

// NewRouter serves the API under /api and the front end everywhere else.
func NewRouter(services Services, options Options) http.Handler {
	mux := http.NewServeMux()
	strict := gen.NewStrictHandlerWithOptions(
		handlers{services: services, demoMode: options.DemoMode},
		nil,
		gen.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  writeRequestError,
			ResponseErrorHandlerFunc: writeServiceError,
		},
	)
	gen.HandlerWithOptions(strict, gen.StdHTTPServerOptions{
		BaseURL:          apiBasePath,
		BaseRouter:       mux,
		Middlewares:      []gen.MiddlewareFunc{requireTenantContext},
		ErrorHandlerFunc: writeRequestError,
	})
	mux.Handle("/", http.FileServerFS(options.Static))
	return mux
}
