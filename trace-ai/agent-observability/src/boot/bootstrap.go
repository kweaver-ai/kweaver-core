package boot

import (
	"context"
	"net/http"

	docs "github.com/kweaver-ai/agent-observability/docs/swagger"
	"github.com/kweaver-ai/agent-observability/src/conf"
	"github.com/kweaver-ai/agent-observability/src/domain/service/tracesvc"
	"github.com/kweaver-ai/agent-observability/src/drivenadapter/httpaccess/opensearchtraceaccess"
	"github.com/kweaver-ai/agent-observability/src/driveradapter/api/httphandler"
	"github.com/kweaver-ai/agent-observability/src/infra/opensearch"
	"github.com/kweaver-ai/agent-observability/src/infra/server/httpserver"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type App struct {
	server *httpserver.Server
}

const APIBasePath = "/api/agent-observability/v1"

func NewApp() *App {
	httpServerConfig := conf.NewHTTPServerConfig()
	openSearchConfig := conf.NewOpenSearchConfig()
	docs.SwaggerInfo.BasePath = APIBasePath

	openSearchClient := opensearch.New(
		openSearchConfig.Endpoint,
		opensearch.AuthConfig{
			Enabled:  openSearchConfig.Auth.Enabled,
			Username: openSearchConfig.Auth.Username,
			Password: openSearchConfig.Auth.Password,
		},
		openSearchConfig.Timeout,
	)
	traceDetailClient := opensearchtraceaccess.New(openSearchClient, openSearchConfig.TraceIndex)
	traceQueryService := tracesvc.New(traceDetailClient)
	traceHandler := httphandler.NewTraceHandler(traceQueryService)

	mux := http.NewServeMux()
	mux.HandleFunc(APIBasePath+"/traces/_search", traceHandler.SearchTraces)
	mux.HandleFunc(APIBasePath+"/traces/by-conversation", traceHandler.SearchTracesByConversationID)
	mux.Handle(APIBasePath+"/swagger/", httpSwagger.Handler(
		httpSwagger.URL(APIBasePath+"/swagger/doc.json"),
	))

	return &App{
		server: httpserver.New(httpServerConfig.Address, mux),
	}
}

func (a *App) Start() error {
	return a.server.Start()
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.server != nil {
		return a.server.Shutdown(ctx)
	}

	return nil
}
