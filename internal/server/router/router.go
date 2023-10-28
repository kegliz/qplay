package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"kegnet.dev/qplay/internal/server/logger"
)

type (
	Router struct {
		*gin.Engine
		Logger     *logger.Logger
		Routes     []*Route
		BasePath   string
		HTTPServer *http.Server
	}

	RouterOptions struct {
		Logger          *logger.Logger
		TemplateFolder  string
		BasePath        string
		CORSAllowOrigin string
	}

	Route struct {
		Name        string
		Method      string
		Pattern     string
		HandlerFunc gin.HandlerFunc
	}

	ErrNoServerToShutdown struct{}
)

func (e *ErrNoServerToShutdown) Error() string {
	return "no server to shutdown"
}

// NewRouter creates a new router
func NewRouter(options RouterOptions) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.LoadHTMLGlob(options.TemplateFolder + "/*.tmpl")
	engine.Static("/static", "./public") //TODO it should be configurable

	engine.Use(gin.Recovery())
	engine.Use(requestWrapper(options.Logger))

	engine.Use(cors(CORSOptions{
		Origin: options.CORSAllowOrigin,
	}))

	router := &Router{
		Engine:   engine,
		Routes:   []*Route{},
		Logger:   options.Logger,
		BasePath: options.BasePath,
	}
	router.NoRoute(func(c *gin.Context) { c.JSON(404, gin.H{"error": "not found"}) })
	return router
}

// Start starts the server
func (r *Router) Start(port int) error {
	r.HTTPServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	return r.HTTPServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting any active connections
func (r *Router) Shutdown(ctx context.Context) error {
	if r.HTTPServer != nil {
		return r.HTTPServer.Shutdown(ctx)
	} else {
		return new(ErrNoServerToShutdown)
	}
}

// SetRoutes sets the routes for the router and registers them in the gin engine
func (r *Router) SetRoutes(routes []*Route) {
	r.Routes = routes
	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			r.GET(r.BasePath+route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			r.POST(r.BasePath+route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			r.PUT(r.BasePath+route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			r.DELETE(r.BasePath+route.Pattern, route.HandlerFunc)
		}
		r.Logger.Info().Msgf("Route %s %s registered", route.Method, r.BasePath+route.Pattern)
	}

}
