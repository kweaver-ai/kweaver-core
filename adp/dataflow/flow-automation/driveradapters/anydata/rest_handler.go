package anydata

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/gin-gonic/gin"
)

var (
	once sync.Once
	rh   RESTHandler
)

type RESTHandler interface {
	RegisterAPI(engine *gin.RouterGroup)
}

type restHandler struct {
	ad drivenadapters.AnyData
}

func NewRestHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			ad: drivenadapters.NewAnyData(),
		}
	})

	return rh
}

var (
	PublicEndpoints    = [][]string{}
	ProtectedEndpoints = [][]string{
		{"GET", "model-factory/v1/llm-source"},
		{"GET", "agent-factory/v2/agent"},
		{"GET", "agent-factory/v2/agent/list"},
	}
	AdminEndpoints = [][]string{
		{"GET", "builder/*any"},
	}
)

func (rh *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	group := engine.Group("anydata")
	for _, endpoint := range PublicEndpoints {
		group.Handle(endpoint[0], endpoint[1], rh.anydataProxy)
	}

	for _, endpoint := range ProtectedEndpoints {
		group.Handle(endpoint[0], endpoint[1], middleware.TokenAuth(), rh.anydataProxy)
	}

	for _, endpoint := range AdminEndpoints {
		group.Handle(endpoint[0], endpoint[1], middleware.TokenAuth(), middleware.CheckAdmin(), rh.anydataProxy)
	}
}

func (rh *restHandler) anydataProxy(c *gin.Context) {
	baseURL := rh.ad.GetBaseURL()
	appID := rh.ad.GetAppID()

	if baseURL == "" {
		c.JSON(http.StatusServiceUnavailable, errors.NewIError(errors.UnAvailable, "AnyData", ""))
		return
	}

	prefix := os.Getenv("API_PREFIX")

	method := c.Request.Method
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery

	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimPrefix(path, "/anydata")

	targetURL, _ := url.Parse(baseURL)
	targetURL = targetURL.JoinPath("/api", path)
	targetURL.RawQuery = query

	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	reqHeader := make(http.Header)
	reqHeader.Set("appid", appID)
	reqHeader.Set("content-type", "application/json;charset=UTF-8")
	reqHeader.Set("accept-language", "zh-CN")

	req, err := http.NewRequest(method, targetURL.String(), bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewIError(errors.InternalError, "AnyData", err))
		return
	}
	req.Header = reqHeader

	client := drivenadapters.NewRawHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errors.NewIError(errors.UnAvailable, "AnyData", err))
		return
	}
	defer resp.Body.Close()

	c.Status(resp.StatusCode)
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(c.Writer, resp.Body)
}
