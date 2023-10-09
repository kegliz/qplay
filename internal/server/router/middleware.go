package router

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"kegnet.dev/qplay/internal/server/logger"
)

var (
	requestServedMsg string = "Request served"
	requestCount     int64
)

// requestWrapper is a middleware that logs the request and response.
func requestWrapper(log *logger.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		setupContext(c)
		reqPath := c.Request.URL.Path
		log.Debugc(c).Msgf("Incoming request: %s", reqPath)

		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		meta := []interface{}{
			"path", reqPath,
			"method", c.Request.Method,
			"statuscode", status,
			"latency", latency,
		}

		switch {
		case status == http.StatusOK || status == http.StatusCreated || status == http.StatusNoContent:
			log.Infoc(c).Fields(meta).Msg(requestServedMsg)
		case status == http.StatusNotFound:
			log.Warnc(c).Fields(meta).Msg(requestServedMsg)
		default:
			log.Errorc(c).Fields(meta).Msg(requestServedMsg)
		}
	}
}

// setupContext sets up the context for the request.
// It sets the request id and increments the request count.
func setupContext(c *gin.Context) {
	c.Set("requestcount", strconv.FormatInt(atomic.AddInt64(&requestCount, 1), 10))
	reqID := c.Request.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = uuid.Must(uuid.NewRandom()).String()
	}
	c.Set("requestid", reqID)
	c.Writer.Header().Set("X-Request-Id", reqID)
}
