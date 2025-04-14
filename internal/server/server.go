package server

import (
	"context"

	"github.com/kegliz/qplay/internal/server/logger"
	"github.com/kegliz/qplay/internal/server/router"
)

type (
	EngineOptions struct {
		Debug bool
	}

	Server interface {
		Listen(port int, localOnly bool) error
		Shutdown(ctx context.Context) error
	}
)

func NewLoggerAndRouter(options EngineOptions) (l *logger.Logger, r *router.Router) {
	l = logger.NewLogger(logger.LoggerOptions{
		Debug: options.Debug,
	})
	r = router.NewRouter(router.RouterOptions{
		Logger: l,
	})
	return
}
