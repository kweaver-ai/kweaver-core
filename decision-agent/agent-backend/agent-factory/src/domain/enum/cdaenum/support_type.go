package cdaenum

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type SupportDataType string

var ValidSupportDataTypes = []SupportDataType{
	"file",
}

type SupportDataTypes []SupportDataType

func (c SupportDataTypes) EnumCheck() (err error) {
	if len(c) == 0 {
		err = errors.New("[SupportDataTypes]: cannot be empty")
		return
	}

	for _, t := range c {
		if !cutil.ExistsGeneric(ValidSupportDataTypes, t) {
			err = errors.New("[SupportDataTypes]: invalid type")
			return
		}
	}

	return
}
