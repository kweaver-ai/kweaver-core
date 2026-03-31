package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

var (
	personalConfigOnce sync.Once
	personalConfig     PersonalConfig
)

func NewPersonalConfig() PersonalConfig {
	personalConfigOnce.Do(func() {

		config := common.NewConfig()
		privateHost := config.PersonalConfig.PrivateHost
		if privateHost == "" {
			privateHost = "personal-config-private"
		}
		privatePort := config.PersonalConfig.PrivatePort
		if privatePort == 0 {
			privatePort = 8082
		}
		personalConfig = &PersonalConfigImpl{
			baseURL:    fmt.Sprintf("http://%s:%d", config.PersonalConfig.PrivateHost, config.PersonalConfig.PrivatePort),
			httpClient: otelHttp.NewOtelHttpClient(),
		}
	})
	return personalConfig
}

type PersonalConfigService struct {
	Name    string `json:"name"`
	Version string `json:"verion"`
}

type PersonalConfigModule struct {
	Name     string                  `json:"name"`
	Version  string                  `json:"verion"`
	Services []PersonalConfigService `json:"services"`
}

type PersonalConfig interface {
	GetModuleByName(ctx context.Context, name string) (PersonalConfigModule, error)
}

type PersonalConfigImpl struct {
	baseURL    string
	httpClient otelHttp.HTTPClient
}

func (c *PersonalConfigImpl) GetModuleByName(ctx context.Context, name string) (mod PersonalConfigModule, err error) {
	target := fmt.Sprintf("%s/api/personal-config/v1/deployment/module/%s", c.baseURL, name)
	respCode, result, err := c.httpClient.Get(ctx, target, map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	})

	if err != nil {
		if respCode != http.StatusNotFound {
			traceLog.WithContext(ctx).Warnf("GetModuleByName url: %v", target)
		}
		return
	}

	bytes, err := json.Marshal(result)

	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &mod)

	return
}
