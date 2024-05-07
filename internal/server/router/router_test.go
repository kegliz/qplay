package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kegliz/qplay/internal/server/logger"
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	suite.Suite
	BasePath       string
	BasePathRouter *Router
}

func (s *RouterTestSuite) doRequest(method string, urlStr string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(method, urlStr, nil)
	s.BasePathRouter.ServeHTTP(rec, req)
	return rec
}

func (s *RouterTestSuite) SetupSuite() {

	//_, fn, _, _ := runtime.Caller(0)
	//dir := path.Join(path.Dir(fn), "../../..") //TODO it should be configurable
	//err := os.Chdir(dir)
	//if err != nil {
	//panic(err)
	//}

	log := logger.NewLogger(logger.LoggerOptions{
		Debug: true,
	})
	testRoutes := []*Route{
		{
			Name:    "root",
			Method:  http.MethodGet,
			Pattern: "/",
			HandlerFunc: func(c *gin.Context) {
				c.Data(http.StatusOK, "text/html", []byte("200"))
			},
		},
		{
			Name:        "objWithParam",
			Method:      http.MethodGet,
			Pattern:     "/obj/:param",
			HandlerFunc: func(c *gin.Context) { c.Data(http.StatusOK, "text/html", []byte(c.Param("param"))) },
		},
	}
	s.BasePath = "/my/crazy/path"
	s.BasePathRouter = NewRouter(RouterOptions{
		Logger:         log,
		BasePath:       s.BasePath,
		TemplateFolder: "testdata/templates",
	})
	s.BasePathRouter.SetRoutes(testRoutes)
}

func (s *RouterTestSuite) TestBasePathRouter() {
	rec := s.doRequest(http.MethodGet, s.BasePath+"/")
	s.Equal(http.StatusOK, rec.Code, "200 GET "+s.BasePath+"/")

	rec = s.doRequest(http.MethodGet, "/belabacsi")
	s.Equal(http.StatusNotFound, rec.Code, "404 GET "+"/belabacsi")

	rec = s.doRequest(http.MethodGet, s.BasePath+"/obj/rudolf")
	s.Equal(http.StatusOK, rec.Code, "200 GET "+s.BasePath+"/obj/rudolf")
	s.Equal("rudolf", rec.Body.String(), "200 GET "+s.BasePath+"/obj/rudolf")
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
