package squareresp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentMarketAgentInfoResp(t *testing.T) {
	t.Parallel()

	resp := NewAgentMarketAgentInfoResp()

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.PublishInfo)
	assert.IsType(t, &pubedeo.AgentPublishedInfoEo{}, resp.PublishInfo)
}

func TestAgentMarketAgentInfoResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := AgentMarketAgentInfoResp{
		CategoryId:      "cat-123",
		CategoryName:    "TestCategory",
		Version:         "1.0.0",
		LatestVersion:   "1.1.0",
		Description:     "Test description",
		PublishedAt:     1640995200000,
		PublishedBy:     "user-456",
		PublishedByName: "Test User",
		PublishInfo:     &pubedeo.AgentPublishedInfoEo{},
	}

	assert.Equal(t, "cat-123", resp.CategoryId)
	assert.Equal(t, "TestCategory", resp.CategoryName)
	assert.Equal(t, "1.0.0", resp.Version)
	assert.Equal(t, "1.1.0", resp.LatestVersion)
	assert.Equal(t, "Test description", resp.Description)
	assert.Equal(t, int64(1640995200000), resp.PublishedAt)
	assert.Equal(t, "user-456", resp.PublishedBy)
	assert.Equal(t, "Test User", resp.PublishedByName)
	assert.NotNil(t, resp.PublishInfo)
}

func TestAgentMarketAgentInfoResp_EmptyValues(t *testing.T) {
	t.Parallel()

	resp := AgentMarketAgentInfoResp{}

	assert.Empty(t, resp.CategoryId)
	assert.Empty(t, resp.CategoryName)
	assert.Empty(t, resp.Version)
	assert.Empty(t, resp.LatestVersion)
	assert.Empty(t, resp.Description)
	assert.Equal(t, int64(0), resp.PublishedAt)
	assert.Empty(t, resp.PublishedBy)
	assert.Empty(t, resp.PublishedByName)
	assert.Nil(t, resp.PublishInfo)
}

func TestAgentMarketAgentInfoResp_WithConfig(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.Config{}
	resp := AgentMarketAgentInfoResp{
		CategoryId:   "cat-config",
		CategoryName: "ConfigCategory",
		Version:      "2.0.0",
		Config:       config,
		PublishInfo:  &pubedeo.AgentPublishedInfoEo{},
	}

	assert.NotNil(t, resp.Config)
	assert.NotNil(t, resp.PublishInfo)
}

func TestAgentMarketAgentInfoResp_NilPublishInfo(t *testing.T) {
	t.Parallel()

	resp := AgentMarketAgentInfoResp{
		CategoryId:   "cat-nil",
		CategoryName: "NilCategory",
		PublishInfo:  nil,
	}

	assert.Nil(t, resp.PublishInfo)
}

func TestAgentMarketAgentInfoResp_FullyPopulated(t *testing.T) {
	t.Parallel()

	publishInfo := &pubedeo.AgentPublishedInfoEo{}

	resp := AgentMarketAgentInfoResp{
		CategoryId:      "cat-full",
		CategoryName:    "FullCategory",
		Version:         "3.0.0",
		LatestVersion:   "3.1.0",
		Description:     "Fully populated agent",
		PublishedAt:     1672531200000,
		PublishedBy:     "user-999",
		PublishedByName: "Admin User",
		PublishInfo:     publishInfo,
	}

	assert.Equal(t, "cat-full", resp.CategoryId)
	assert.Equal(t, "FullCategory", resp.CategoryName)
	assert.Equal(t, "3.0.0", resp.Version)
	assert.Equal(t, "3.1.0", resp.LatestVersion)
	assert.Equal(t, "Fully populated agent", resp.Description)
	assert.Equal(t, int64(1672531200000), resp.PublishedAt)
	assert.Equal(t, "user-999", resp.PublishedBy)
	assert.Equal(t, "Admin User", resp.PublishedByName)
	assert.Same(t, publishInfo, resp.PublishInfo)
}

func TestAgentMarketAgentInfoResp_WithTimestamps(t *testing.T) {
	t.Parallel()

	timestamps := []int64{
		1640995200000, // 2022-01-01
		1643673600000, // 2022-02-01
		1646092800000, // 2022-03-01
		1648771200000, // 2022-04-01
	}

	for i, ts := range timestamps {
		resp := AgentMarketAgentInfoResp{
			PublishedAt: ts,
			PublishInfo: &pubedeo.AgentPublishedInfoEo{},
		}

		assert.Equal(t, ts, resp.PublishedAt, "Timestamp %d should match", i)
	}
}

func TestAgentMarketAgentInfoResp_VersionComparison(t *testing.T) {
	t.Parallel()

	resp := AgentMarketAgentInfoResp{
		Version:       "1.0.0",
		LatestVersion: "2.0.0",
		PublishInfo:   &pubedeo.AgentPublishedInfoEo{},
	}

	assert.NotEqual(t, resp.Version, resp.LatestVersion)
	assert.Equal(t, "1.0.0", resp.Version)
	assert.Equal(t, "2.0.0", resp.LatestVersion)
}

func TestAgentMarketAgentInfoResp_SameVersion(t *testing.T) {
	t.Parallel()

	resp := AgentMarketAgentInfoResp{
		Version:       "1.5.0",
		LatestVersion: "1.5.0",
		PublishInfo:   &pubedeo.AgentPublishedInfoEo{},
	}

	assert.Equal(t, resp.Version, resp.LatestVersion)
}
