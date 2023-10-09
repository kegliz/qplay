package app

import (
	"net/http"

	"kegnet.dev/qplay/internal/server/router"
)

func (a *appServer) routes() []*router.Route {
	return []*router.Route{
		{
			Name:        "root",
			Method:      http.MethodGet,
			Pattern:     "/",
			HandlerFunc: a.RootHandler,
		},
		{
			Name:        "health",
			Method:      http.MethodGet,
			Pattern:     "/health",
			HandlerFunc: a.HealthHandler,
		},
	}
}
