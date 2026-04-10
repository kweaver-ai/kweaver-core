package cdaenum

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type BitUnit string

const (
	KB BitUnit = "KB"
	MB BitUnit = "MB"
	GB BitUnit = "GB"
)

func (b BitUnit) EnumCheck() (err error) {
	if !cutil.ExistsGeneric([]BitUnit{KB, MB, GB}, b) {
		err = errors.New("[BitUnit]: invalid unit")
		return
	}

	return
}
