package cenum

import "errors"

type BizDomainID string

func (b BizDomainID) ToString() string {
	return string(b)
}

const (
	BizDomainPublic BizDomainID = "bd_public" // 公共业务域
)

// EnumCheck 验证业务域是否有效
func (b BizDomainID) EnumCheck() error {
	if b != BizDomainPublic {
		return errors.New("[BizDomain][EnumCheck]: 无效的业务域: " + b.ToString())
	}

	return nil
}
