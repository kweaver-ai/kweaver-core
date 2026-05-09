package driveradapters

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/logger"
)

func (stubKnSearchHandler) SearchSchema(c *gin.Context) {
	c.String(http.StatusOK, "search_schema")
}

func TestRestPublicHandler_RegistersSearchSchemaRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	convey.Convey("restPublicHandler registers public /kn/search_schema route", t, func() {
		engine := gin.New()
		routerGroup := engine.Group("/api/agent-retrieval/v1")

		handler := &restPublicHandler{
			Hydra:                          stubPublicHydra{},
			KnRetrievalHandler:             stubSemanticSearchHandler{},
			MCPHandler:                     http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
			KnLogicPropertyResolverHandler: stubLogicPropertyResolverHandler{},
			KnActionRecallHandler:          stubActionRecallHandler{},
			KnQueryObjectInstanceHandler:   stubQueryObjectInstanceHandler{},
			KnQuerySubgraphHandler:         stubQuerySubgraphHandler{},
			KnSearchHandler:                stubKnSearchHandler{},
			KnFindSkillsHandler:            stubKnFindSkillsHandler{},
			Logger:                         logger.DefaultLogger(),
		}
		handler.RegisterRouter(routerGroup)

		req := httptest.NewRequest(http.MethodPost, "/api/agent-retrieval/v1/kn/search_schema", http.NoBody)
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
		convey.So(w.Body.String(), convey.ShouldEqual, "search_schema")
	})
}
