package chelper

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

func SetAccountInfoToHeaderMap(headerMap map[string]string, accountID string, accountType cenum.AccountType) {
	if headerMap == nil {
		return
	}

	// 1. 设置account-id
	headerMap[cenum.HeaderXAccountID.String()] = accountID
	// 暂时保持向后兼容（后续可以删除）
	headerMap[cenum.HeaderXAccountIDOld.String()] = accountID

	// 2. 设置account-type
	headerMap[cenum.HeaderXAccountType.String()] = accountType.String()
	// 暂时保持向后兼容（后续可以删除）
	headerMap[cenum.HeaderXAccountTypeOld.String()] = accountType.String()
}
