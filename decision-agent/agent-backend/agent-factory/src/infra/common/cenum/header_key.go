package cenum

import (
	"fmt"
)

type HeaderKey string

// x-account headers
const (
	HeaderXAccountType HeaderKey = "x-account-type"
	// 暂时保持向后兼容（后续可以删除）
	HeaderXAccountTypeOld HeaderKey = "x-visitor-type"

	HeaderXAccountID HeaderKey = "x-account-id"
	// 暂时保持向后兼容（后续可以删除）
	HeaderXAccountIDOld HeaderKey = "x-user"
)

// 业务域相关header
const (
	HeaderXBizDomainID HeaderKey = "x-business-domain"
)

// to string
func (h HeaderKey) String() string {
	return string(h)
}

// EnumCheck 验证header key是否有效
func (h HeaderKey) EnumCheck() error {
	switch h {
	case HeaderXAccountType, HeaderXAccountID, HeaderXAccountTypeOld, HeaderXAccountIDOld, HeaderXBizDomainID:
		return nil
	default:
		return fmt.Errorf("[HeaderKey][EnumCheck]: 无效的header key: %s", h)
	}
}
