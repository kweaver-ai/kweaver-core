package agentinoutreq

import (
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type ImportType string

const (
	ImportTypeUpsert ImportType = "upsert"
	ImportTypeCreate ImportType = "create"
)

func (b ImportType) EnumCheck() (err error) {
	if !cutil.ExistsGeneric([]ImportType{ImportTypeUpsert, ImportTypeCreate}, b) {
		err = errors.New("[ImportType]: invalid import type")
		return
	}

	return
}

func (b ImportType) String() string {
	return string(b)
}
