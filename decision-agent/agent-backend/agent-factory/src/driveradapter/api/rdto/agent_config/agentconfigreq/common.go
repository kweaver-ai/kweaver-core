package agentconfigreq

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum/agentconfigenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
)

func setDefaultValue(config *daconfvalobj.Config) {
	// 2. 给metaData设置默认值
	if config.GetConfigMetadata().GetConfigTplVersion() == "" {
		config.GetConfigMetadata().SetConfigTplVersion(agentconfigenum.ConfigTplVersionV1)
	}
}

func HandleConfig(config *daconfvalobj.Config) (err error) {
	// 1. 设置默认值
	setDefaultValue(config)

	return
}

func D2eCommonAfterD2e(eo *daconfeo.DataAgent) {
	// 1. 设置Config.Metadata.ConfigLastSetTimestamp
	eo.Config.Metadata.SetConfigLastSetTimestamp()
}
