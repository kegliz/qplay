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
		{
			Name:        "api.qprogs.save",
			Method:      http.MethodPost,
			Pattern:     "/api/qprogs",
			HandlerFunc: a.CreateCircuit,
		},
		{
			Name:        "api.qprogs.render",
			Method:      http.MethodGet,
			Pattern:     "/api/qprogs/:id/img",
			HandlerFunc: a.RenderCircuit,
		},
	}
}
