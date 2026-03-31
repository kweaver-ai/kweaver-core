package initial

import (
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/stretchr/testify/assert"
)

func TestInitCluster_WithConfiguredAddress(t *testing.T) {
	// Setup configuration with pre-defined AccessAddress
	config := &common.Config{
		AccessAddress: common.AccessAddress{
			Host: "127.0.0.1",
			Port: "8080",
		},
		DeployService: common.DeployService{},
	}

	// Execute initCluster
	initCluster(config)

	// Verify that DeployService uses the values from AccessAddress
	assert.Equal(t, "127.0.0.1", config.DeployService.Host)
	assert.Equal(t, "8080", config.DeployService.Port)
}
