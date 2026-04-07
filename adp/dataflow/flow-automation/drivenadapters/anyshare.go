package drivenadapters

import (
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/anyshare.go -destination ../tests/mock_drivenadapters/anyshare_mock.go

// Anyshare method interface
type Anyshare interface {
	// 获取anyshare 集群地址
	ClusterAccess() (ClusterAccess, error)
}

var (
	anyshareOnce sync.Once
	anyshare     Anyshare
)

type anyshareClient struct {
	log        commonLog.Logger
	baseURL    string
	httpClient HTTPClient
}

// ClusterAccess 集群访问地址
type ClusterAccess struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// NewAnyshare deploy服务
func NewAnyshare() Anyshare {
	anyshareOnce.Do(
		func() {
			config := common.NewConfig()
			anyshare = &anyshareClient{
				log:        commonLog.NewLogger(),
				baseURL:    fmt.Sprintf("http://%s:%s", config.DeployService.Host, config.DeployService.Port),
				httpClient: NewHTTPClient(),
			}
		})

	return anyshare
}

func (ac *anyshareClient) ClusterAccess() (ClusterAccess, error) {
	target := fmt.Sprintf("%s/api/deploy-manager/v1/access-addr/app", ac.baseURL)
	respParam, err := ac.httpClient.Get(target, nil)
	if err != nil {
		ac.log.Errorf("get cluster access failed: %v, url: %v", err, target)
		return ClusterAccess{}, err
	}

	// Check if respParam is nil
	if respParam == nil {
		ac.log.Errorf("get cluster access returned nil response, url: %v", target)
		return ClusterAccess{}, fmt.Errorf("cluster access response is nil")
	}

	// Type assert to map
	respMap, ok := respParam.(map[string]interface{})
	if !ok {
		ac.log.Errorf("get cluster access response is not a map, url: %v, type: %T", target, respParam)
		return ClusterAccess{}, fmt.Errorf("cluster access response is not a map")
	}

	// Check if host and port exist
	hostVal, hostExists := respMap["host"]
	portVal, portExists := respMap["port"]

	if !hostExists || !portExists {
		ac.log.Errorf("get cluster access response missing host or port, url: %v", target)
		return ClusterAccess{}, fmt.Errorf("cluster access response missing host or port")
	}

	// Check if host and port are nil
	if hostVal == nil || portVal == nil {
		ac.log.Errorf("get cluster access host or port is nil, url: %v", target)
		return ClusterAccess{}, fmt.Errorf("cluster access host or port is nil")
	}

	// Type assert host and port to string
	host, hostOk := hostVal.(string)
	port, portOk := portVal.(string)

	if !hostOk || !portOk {
		ac.log.Errorf("get cluster access host or port is not a string, url: %v, host type: %T, port type: %T", target, hostVal, portVal)
		return ClusterAccess{}, fmt.Errorf("cluster access host or port is not a string")
	}

	clusterAccess := ClusterAccess{
		Host: host,
		Port: port,
	}
	return clusterAccess, nil
}
