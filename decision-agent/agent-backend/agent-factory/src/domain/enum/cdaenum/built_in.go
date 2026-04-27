package cdaenum

import "github.com/pkg/errors"

// BuiltIn 是否内置
type BuiltIn int8

const (
	// BuiltInNo 非内置
	BuiltInNo BuiltIn = 0

	// BuiltInYes 内置
	BuiltInYes BuiltIn = 1
)

func (b BuiltIn) EnumCheck() (err error) {
	if b != BuiltInNo && b != BuiltInYes {
		err = errors.New("[BuiltIn]: invalid built in")
		return
	}

	return
}

func (b *BuiltIn) IsBuiltIn() bool {
	return b != nil && *b == BuiltInYes
}
