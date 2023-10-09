package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kegnet.dev/qplay/internal/server/logger"
)

var badRequestErrorMsg = "Bad Request - please contact the administrator"
var internalServerErrorMsg = "Internal Server Error - please contact the administrator"

// RootHandler is the handler for the / endpoint
func (a *appServer) RootHandler(c *gin.Context) {
	log := a.logger.ContextLoggingFn(c)
	log(logger.DebugLevel).Msg("serving root endpoint")

	c.HTML(http.StatusOK, "index.tmpl", gin.H{"title": "Quantum Playground DEV"})
}

// HealthHandler is the handler for the /health endpoint
func (a *appServer) HealthHandler(c *gin.Context) {
	log := a.logger.ContextLoggingFn(c)
	log(logger.DebugLevel).Msg("serving health endpoint")
	c.String(http.StatusOK, "OK")
}
