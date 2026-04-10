package capimiddleware

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
)

var hydraInstance rest.Hydra

const defaultMockUserID = "e39adc84-6de8-11f0-b206-4a2c3f0cd493"

type MockHydra struct{}

func (m *MockHydra) GetLanguage(c *gin.Context) rest.Language {
	return rest.SimplifiedChinese
}

func (m *MockHydra) VerifyToken(ctx context.Context, c *gin.Context) (rest.Visitor, error) {
	userID := defaultMockUserID
	if global.GConfig != nil &&
		global.GConfig.SwitchFields != nil &&
		global.GConfig.SwitchFields.Mock != nil &&
		global.GConfig.SwitchFields.Mock.MockUserID != "" {
		userID = global.GConfig.SwitchFields.Mock.MockUserID
	}

	return rest.Visitor{
		ID:      userID,
		TokenID: "Bearer mock token",
		Type:    rest.VisitorType_RealName,
	}, nil
}

func (m *MockHydra) Introspect(ctx context.Context, token string) (rest.TokenIntrospectInfo, error) {
	return rest.TokenIntrospectInfo{}, nil
}

func GetHydra() rest.Hydra {
	if hydraInstance != nil {
		return hydraInstance
	}

	if global.GConfig.SwitchFields.Mock.MockHydra {
		hydraInstance = &MockHydra{}
		return hydraInstance
	}

	hydraAdminSetting := rest.HydraAdminSetting{
		HydraAdminHost:     cglobal.GConfig.Hydra.HydraAdmin.Host,
		HydraAdminPort:     cglobal.GConfig.Hydra.HydraAdmin.Port,
		HydraAdminProcotol: "http",
	}
	hydraInstance = rest.NewHydra(hydraAdminSetting)

	return hydraInstance
}

func VerifyOAuthMiddleWare() gin.HandlerFunc {
	hydra := GetHydra()

	return func(c *gin.Context) {
		ctx := rest.GetLanguageCtx(c)
		visitor, err := hydra.VerifyToken(ctx, c)
		if err != nil {
			httpError := rest.NewHTTPError(ctx, http.StatusUnauthorized, rest.PublicError_Unauthorized).
				WithErrorDetails(err.Error())
			rest.ReplyError(c, httpError)
			c.Abort()

			return
		}

		ctxKey := cenum.VisitUserInfoCtxKey.String()
		c.Set(ctxKey, &visitor)

		// 设置request context
		_ctx := context.WithValue(c.Request.Context(), ctxKey, &visitor) //nolint:staticcheck // SA1029
		cutil.UpdateGinReqCtx(c, _ctx)

		// 执行后续操作
		c.Next()
	}
}
