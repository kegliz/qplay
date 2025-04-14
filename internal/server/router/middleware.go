package router

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kegliz/qplay/internal/server/logger"
)

var (
	requestServedMsg string = "Request served"
	requestCount     int64
)

type CORSOptions struct {
	Origin string
}

// CORS middleware from
// https://github.com/gin-gonic/gin/issues/29#issuecomment-89132826
// https://www.moesif.com/blog/technical/cors/Authoritative-Guide-to-CORS-Cross-Origin-Resource-Sharing-for-REST-APIs/
func cors(options CORSOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // allow any origin domain
		if options.Origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", options.Origin)
		}
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

// requestWrapper is a middleware that logs the request and response and
// It injects the logger into the context.
// It is used to log the request and response.
// It is used to set the request id and request count in the context.
func requestWrapper(log *logger.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		reqCount, reqID := setupContext(c)
		l := log.SpawnForContext(reqCount, reqID)
		c.Set("logger", l)
		reqPath := c.Request.URL.Path
		l.Debug().Msgf("Incoming request: %s", reqPath)

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
			l.Info().Fields(meta).Msg(requestServedMsg)
		case status == http.StatusNotFound:
			l.Warn().Fields(meta).Msg(requestServedMsg)
		default:
			l.Error().Fields(meta).Msg(requestServedMsg)
		}
	}
}

// setupContext sets up the context for the request.
// It sets the request id and increments the request count.
func setupContext(c *gin.Context) (reqCount string, reqID string) {
	reqCount = strconv.FormatInt(atomic.AddInt64(&requestCount, 1), 10)
	c.Set("requestcount", reqCount)
	reqID = c.Request.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = uuid.Must(uuid.NewRandom()).String()
	}
	c.Set("requestid", reqID)
	c.Writer.Header().Set("X-Request-Id", reqID)
	return
}
