package app

import (
	"image/png"
	"net/http"

	"github.com/gin-gonic/gin"
	"kegnet.dev/qplay/internal/qservice"
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

// CreateCircuit is the handler for the /api/qprogs endpoint
func (a *appServer) CreateCircuit(c *gin.Context) {
	log := a.logger.ContextLoggingFn(c)
	log(logger.DebugLevel).Msg("serving qprog creation endpoint")
	var params qservice.ProgramValue
	if err := c.ShouldBindJSON(&params); err != nil {
		log(logger.ErrorLevel).Err(err).Msg("binding json failed")
		c.String(http.StatusBadRequest, badRequestErrorMsg)
		return
	}
	// Save the circuit
	id, err := a.qs.SaveProgram(log, &params)
	if err != nil {
		log(logger.ErrorLevel).Err(err).Msg("saving circuit failed")
		c.String(http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	c.PureJSON(http.StatusOK, qservice.ProgramIDValue{ID: id})
}

// RenderCircuit is the handler for the /api/qprogs/:id/img endpoint
func (a *appServer) RenderCircuit(c *gin.Context) {
	log := a.logger.ContextLoggingFn(c)
	log(logger.DebugLevel).Msg("serving rendering circuit img endpoint")
	id := c.Param("id")
	img, err := a.qs.RenderCircuit(log, id)
	if err != nil {
		log(logger.ErrorLevel).Err(err).Msg("rendering circuit failed")
		c.String(http.StatusInternalServerError, internalServerErrorMsg)
		return
	}
	c.Header("Content-Type", "image/png")
	png.Encode(c.Writer, img)
	c.Status(http.StatusOK)
}
