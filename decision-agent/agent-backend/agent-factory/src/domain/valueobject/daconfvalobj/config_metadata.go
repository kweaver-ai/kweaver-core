package daconfvalobj

import (
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum/agentconfigenum"
)

type ConfigMetadata struct {
	ConfigTplVersion       agentconfigenum.ConfigTplVersionT `json:"config_tpl_version"`        // 配置版本
	ConfigLastSetTimestamp uint64                            `json:"config_last_set_timestamp"` // 配置时间戳(nanoseconds)
}

func (c *ConfigMetadata) SetConfigTplVersion(v agentconfigenum.ConfigTplVersionT) {
	if err := v.EnumCheck(); err != nil {
		panic(err)
	}

	c.ConfigTplVersion = v
}

func (c *ConfigMetadata) GetConfigTplVersion() agentconfigenum.ConfigTplVersionT {
	return c.ConfigTplVersion
}

func (c *ConfigMetadata) SetConfigLastSetTimestamp() {
	c.ConfigLastSetTimestamp = uint64(time.Now().UnixNano())
}

func (c *ConfigMetadata) GetConfigLastSetTimestamp() uint64 {
	return c.ConfigLastSetTimestamp
}
