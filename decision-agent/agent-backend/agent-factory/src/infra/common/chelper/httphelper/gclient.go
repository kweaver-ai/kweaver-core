package httphelper

import (
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
)

var defaultStdClient *http.Client

func init() {
	if cenvhelper.IsLocalDev() {
		defaultStdClient = GetClient(0)
	} else {
		defaultStdClient = GetClient(DefaultTimeout)
	}
}

func GetNewGClientWithDefaultStdClient() (gClient *gclient.Client) {
	gClient = g.Client()
	gClient.Client = *defaultStdClient

	return
}

func GetDefaultClient() *http.Client {
	return defaultStdClient
}
