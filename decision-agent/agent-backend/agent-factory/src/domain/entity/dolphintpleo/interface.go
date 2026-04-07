package dolphintpleo

import "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"

type IDolphinTpl interface {
	LoadFromConfig(config *daconfvalobj.Config)
	ToString() (str string)
}
