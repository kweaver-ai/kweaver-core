package driveradapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/logger"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func TestMiddlewareResponseFormat_DefaultAndValid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	convey.Convey("middlewareResponseFormat default and valid values", t, func() {
		convey.Convey("no query param defaults to json", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

			mw := middlewareResponseFormat()
			mw(c)

			formatVal, ok := common.GetResponseFormatFromCtx(c.Request.Context())
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(formatVal, convey.ShouldEqual, rest.FormatJSON)
			convey.So(w.Code, convey.ShouldEqual, 200) // default status for recorder
		})

		convey.Convey("response_format=json", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test?response_format=json", http.NoBody)

			mw := middlewareResponseFormat()
			mw(c)

			formatVal, ok := common.GetResponseFormatFromCtx(c.Request.Context())
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(formatVal, convey.ShouldEqual, rest.FormatJSON)
		})

		convey.Convey("response_format=toon", func() {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test?response_format=toon", http.NoBody)

			mw := middlewareResponseFormat()
			mw(c)

			formatVal, ok := common.GetResponseFormatFromCtx(c.Request.Context())
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(formatVal, convey.ShouldEqual, rest.FormatTOON)
		})
	})
}

func TestMiddlewareResponseFormat_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	convey.Convey("middlewareResponseFormat invalid value returns 400", t, func() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test?response_format=xml", http.NoBody)

		mw := middlewareResponseFormat()
		mw(c)

		convey.So(w.Code, convey.ShouldEqual, http.StatusBadRequest)
		convey.So(w.Body.String(), convey.ShouldContainSubstring, "invalid response_format")
	})
}

func TestMiddlewareHeaderAuthContext_LeavesMissingAccountHeadersEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	convey.Convey("middlewareHeaderAuthContext keeps missing account headers empty", t, func() {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

		mw := middlewareHeaderAuthContext()
		mw(c)

		authCtx, ok := common.GetAccountAuthContextFromCtx(c.Request.Context())
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(authCtx, convey.ShouldNotBeNil)
		convey.So(authCtx.AccountID, convey.ShouldEqual, "")
		convey.So(authCtx.AccountType, convey.ShouldEqual, interfaces.AccessorType(""))
		convey.So(authCtx.TokenInfo, convey.ShouldNotBeNil)
		convey.So(authCtx.TokenInfo.VisitorID, convey.ShouldEqual, "")
		convey.So(authCtx.TokenInfo.VisitorTyp, convey.ShouldEqual, interfaces.VisitorType(""))

		header := common.GetHeaderFromCtx(c.Request.Context())
		convey.So(header[string(interfaces.HeaderXAccountID)], convey.ShouldEqual, "")
		convey.So(header[string(interfaces.HeaderXAccountType)], convey.ShouldEqual, "")
	})
}

type stubPublicHydra struct{}

func (stubPublicHydra) Introspect(_ context.Context, _ string) (*interfaces.TokenInfo, error) {
	return &interfaces.TokenInfo{
		VisitorID:  "user-1",
		VisitorTyp: interfaces.RealName,
	}, nil
}

type stubSemanticSearchHandler struct{}

func (stubSemanticSearchHandler) SemanticSearch(c *gin.Context) {
	formatVal, ok := common.GetResponseFormatFromCtx(c.Request.Context())
	if !ok {
		c.String(http.StatusInternalServerError, "response_format missing")
		return
	}
	if formatVal != rest.FormatTOON {
		c.String(http.StatusInternalServerError, "unexpected response_format")
		return
	}
	c.String(http.StatusOK, "ok")
}

type stubLogicPropertyResolverHandler struct{}

func (stubLogicPropertyResolverHandler) ResolveLogicProperties(c *gin.Context) {
	c.Status(http.StatusOK)
}

type stubActionRecallHandler struct{}

func (stubActionRecallHandler) GetActionInfo(c *gin.Context) {
	c.Status(http.StatusOK)
}

type stubQueryObjectInstanceHandler struct{}

func (stubQueryObjectInstanceHandler) QueryObjectInstance(c *gin.Context) {
	c.Status(http.StatusOK)
}

type stubQuerySubgraphHandler struct{}

func (stubQuerySubgraphHandler) QueryInstanceSubgraph(c *gin.Context) {
	c.Status(http.StatusOK)
}

type stubKnSearchHandler struct{}

func (stubKnSearchHandler) KnSearch(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (stubKnSearchHandler) SearchSchema(c *gin.Context) {
	c.String(http.StatusOK, "search_schema")
}

type stubFindSkillsHandler struct{}

func (stubFindSkillsHandler) FindSkills(c *gin.Context) {
	c.Status(http.StatusOK)
}

func TestRestPublicHandler_AppliesResponseFormatMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	convey.Convey("restPublicHandler applies response_format middleware to public routes", t, func() {
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
			KnFindSkillsHandler:            stubFindSkillsHandler{},
			Logger:                         logger.DefaultLogger(),
		}
		handler.RegisterRouter(routerGroup)

		req := httptest.NewRequest(http.MethodPost, "/api/agent-retrieval/v1/kn/semantic-search?response_format=toon", http.NoBody)
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()

		engine.ServeHTTP(w, req)

		convey.So(w.Code, convey.ShouldEqual, http.StatusOK)
		convey.So(w.Body.String(), convey.ShouldEqual, "ok")
	})
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
			KnFindSkillsHandler:            stubFindSkillsHandler{},
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
