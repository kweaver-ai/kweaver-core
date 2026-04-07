package anysharedshandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/httpproxy"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func (h *anysharedsHandler) getInfoByPath(c *gin.Context) {
	// 接收语言标识转换为 context.Context
	ctx := rest.GetLanguageCtx(c)
	reqByte, err := c.GetRawData()
	if err != nil {
		err = capierr.New400Err(ctx, err.Error())
		_ = c.Error(err)

		return
	}
	// req, _ := http.NewRequest("POST", "http://127.0.0.1:8080/api/v3/any-shareds/info-by-path", nil)
	type req struct {
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Account  string `json:"account"`
		Password string `json:"password"`
		Namepath string `json:"namepath"`
	}

	type resp interface{}

	var reqParam req

	err = json.Unmarshal(reqByte, &reqParam)
	if err != nil {
		err = capierr.New400Err(ctx, err.Error())
		_ = c.Error(err)

		return
	}

	targetURL := fmt.Sprintf("%s://%s:%d/api/efast/v1/file/getinfobypath", reqParam.Protocol, reqParam.Host, reqParam.Port)
	oauthURL := fmt.Sprintf("%s://%s:%d/oauth2/token", reqParam.Protocol, reqParam.Host, reqParam.Port)
	token, err := getAnyshareOauth2Token(ctx, oauthURL, reqParam.Account, reqParam.Password)
	// Token has been obtained successfully
	if err != nil {
		err = capierr.New500Err(ctx, err.Error())
		_ = c.Error(err)

		return
	}
	// proxy := httpproxy.NewJSONPostProxy[req, resp]("https://10.4.110.139/api/efast/v1/file/getinfobypath")
	proxy := httpproxy.NewJSONPostProxy[req, resp](targetURL)
	// proxy.SetToken("ory_at_-vpMgTvvTC80Pdo3Vv35rhRV-eCcJK3Dhqv72SpqO5k.Zs8lG9sRKpwROmzFprbyXHKs1sRtsnd_kIucGSV2oec")
	proxy.SetToken(token)

	rt, err := proxy.Forward(reqParam)
	if err != nil {
		httpErr := capierr.New500Err(ctx, err.Error())
		rest.ReplyError(c, httpErr)

		return
	}

	// // 返回成功
	rest.ReplyOK(c, http.StatusOK, rt)
}
