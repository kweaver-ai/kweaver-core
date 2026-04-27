package cdaenum

import "github.com/pkg/errors"

type ModelType string

const (
	// ModelTypeLlm llm
	ModelTypeLlm ModelType = "llm"
	// ModelTypeRlm rlm
	ModelTypeRlm ModelType = "rlm"
)

func (t ModelType) EnumCheck() (err error) {
	if t != ModelTypeLlm && t != ModelTypeRlm {
		err = errors.New("[ModelType]: invalid model type")
		return
	}

	return
}
