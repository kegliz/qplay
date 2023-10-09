package app

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"kegnet.dev/qplay/internal/config"
	"kegnet.dev/qplay/internal/server"
	"kegnet.dev/qplay/internal/server/router"
)

type (
	AppServerTestSuite struct {
		suite.Suite
		TestAppServer server.Server
	}
)

func (s *AppServerTestSuite) SetupSuite() {
	c := config.NewNakedConfig()
	c.SetConfigType("yaml") // or viper.SetConfigType("YAML")

	// any approach to require this configuration into your program.
	var yamlExample = []byte(`
debug: true
templatefolder: "testdata/templates"
`)

	c.ReadConfig(bytes.NewBuffer(yamlExample))
	var err error
	log.Println("TF:" + c.GetString("templatefolder"))
	s.TestAppServer, err = NewServer(ServerOptions{
		C:       c,
		Version: "test",
	})
	if err != nil {
		panic(err)
	}
}

func (s *AppServerTestSuite) TearDownSuite() {
	s.TestAppServer.Shutdown(context.Background())
}

func (s *AppServerTestSuite) doRequest(method string, urlStr string, body io.Reader, contentType string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(method, urlStr, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	s.TestAppServer.(*appServer).router.ServeHTTP(rec, req)
	return rec
}

// test Shutdown() method
func (s *AppServerTestSuite) TestNoServerToShutdown() {
	// No server started yet
	err := s.TestAppServer.Shutdown(context.Background())
	s.ErrorIs(err, &router.ErrNoServerToShutdown{}, "shutdown should return ErrNoServerToShutdown")

}

// test / endpoint handler
func (s *AppServerTestSuite) TestRootHandler() {
	rec := s.doRequest(http.MethodGet, "/", nil, "")
	s.Equal(http.StatusOK, rec.Code, "200 GET /")
	s.Contains(rec.Body.String(), "testbéla", "200 GET /")
}

// test /health endpoint handler
func (s *AppServerTestSuite) TestHealthHandler() {
	rec := s.doRequest(http.MethodGet, "/health", nil, "")
	s.Equal(http.StatusOK, rec.Code, "200 GET /health")
	s.Contains(rec.Body.String(), "OK", "200 GET /health")
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppServerTestSuite))
}
