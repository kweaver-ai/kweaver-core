package cenum

import "errors"

type YesNoInt8 int8

const (
	YesNoInt8Yes YesNoInt8 = 1
	YesNoInt8No  YesNoInt8 = 0
)

func (t YesNoInt8) EnumCheck() (err error) {
	if t != YesNoInt8Yes && t != YesNoInt8No {
		err = errors.New("[YesNoInt8]: invalid yes/no int8")
		return
	}

	return
}

func (t YesNoInt8) IsYes() bool {
	return t == YesNoInt8Yes
}

func (t YesNoInt8) IsNo() bool {
	return t == YesNoInt8No
}

func (t YesNoInt8) Bool() bool {
	return t == YesNoInt8Yes
}
