package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kegliz/qplay/internal/config"
	"github.com/kegliz/qplay/internal/logger"
	"github.com/kegliz/qplay/internal/server/router"

	"github.com/kegliz/qplay/internal/server"
)

type (
	ServerOptions struct {
		C       *config.Config
		Version string
	}

	appServer struct {
		logger *logger.Logger
		router *router.Router
		//qs      qservice.Service
		version string
	}

	appServerOptions struct {
		logger *logger.Logger
		router *router.Router
		//qs      qservice.Service
		version string
	}
)

// newAppServer creates a new appServer.
func newAppServer(options appServerOptions) *appServer {
	a := &appServer{
		logger: options.logger,
		router: options.router,
		//qs:      options.qs,
		version: options.version,
	}
	a.router.SetRoutes(a.routes())
	return a
}

// Listen implements server.Server.
// Listen implements server.Server.
func (a *appServer) Listen(port int, localOnly bool) error {
	a.logger.Debug().Str("version", a.version).Msg("debug quantum playground server")
	a.logger.Info().
		Int("port", port).
		Bool("localOnly", localOnly).
		Msg("Starting quantum playground service")
	return a.router.Start(port, localOnly)
}

// Shutdown implements server.Server.
func (a *appServer) Shutdown(ctx context.Context) error {
	return a.router.Shutdown(ctx)
}

func NewServer(options ServerOptions) (server.Server, error) {
	l, r := server.NewLoggerAndRouter(server.EngineOptions{
		Debug: options.C.GetBool("debug"),
	})
	// qs := qservice.NewService(qservice.ServiceOptions{
	// 	Logger: l,
	// 	Store:  qservice.NewProgramStore(),
	// })
	app := newAppServer(appServerOptions{
		logger: l,
		router: r,
		//qs:      qs,
		version: options.Version,
	})

	return app, nil
}

func (a *appServer) getLoggerFromContext(c *gin.Context) (*logger.Logger, error) {
	if loggerInstance, ok := c.Get("logger"); ok {
		if loggerInstance, ok := loggerInstance.(*logger.Logger); ok {
			return loggerInstance, nil
		}
	}
	err := errors.New("logger not found in context")
	a.logger.Error().Err(err).Send()
	c.String(http.StatusInternalServerError, internalServerErrorMsg)
	return nil, err
}
