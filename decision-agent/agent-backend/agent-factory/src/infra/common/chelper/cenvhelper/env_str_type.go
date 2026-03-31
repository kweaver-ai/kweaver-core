package cenvhelper

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type EnvStr string

func NewEnvStr(key string, envPrefix string) EnvStr {
	return EnvStr(envPrefix + key)
}

func (e *EnvStr) Value() string {
	return cutil.GetEnv(string(*e), "")
}
