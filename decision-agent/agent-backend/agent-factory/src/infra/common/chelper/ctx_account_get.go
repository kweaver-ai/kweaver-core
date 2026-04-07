package chelper

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// GetAccountTypeFromHeaderMap 从headerMap中获取accountType
func GetAccountTypeFromHeaderMap(headerMap map[string]string) (accountType cenum.AccountType, isExist bool, err error) {
	if headerMap == nil {
		err = errors.New("headerMap is nil")
		return
	}

	// 1. 从headerMap中获取accountType
	accountTypeStr := ""
	if accountTypeStr = headerMap[cenum.HeaderXAccountType.String()]; accountTypeStr == "" {
		accountTypeStr = headerMap[cenum.HeaderXAccountTypeOld.String()]
	}

	// 2. 如果accountTypeStr为空，返回
	if accountTypeStr == "" {
		return
	}

	// 3. 如果accountTypeStr不为空，转换为AccountType
	accountType = cenum.AccountType(accountTypeStr)

	// 4. 如果accountType不合法，返回
	if err = accountType.EnumCheck(); err != nil {
		return
	}

	// 5. 如果accountType合法，设置isExist为true
	isExist = true

	return
}

// GetAccountTypeFromContext 从context中获取accountType
func GetAccountTypeFromContext(c *gin.Context) (accountType cenum.AccountType, isExist bool, err error) {
	if c == nil {
		err = errors.New("c is nil")
		return
	}

	// 1. 从header中获取accountType
	accountTypeStr := c.GetHeader(cenum.HeaderXAccountType.String())
	if accountTypeStr == "" {
		// 暂时保持向后兼容（后续可以删除）
		accountTypeStr = c.GetHeader(cenum.HeaderXAccountTypeOld.String())
	}

	// 2. 如果accountTypeStr为空，返回
	if accountTypeStr == "" {
		return
	}

	// 3. 如果accountTypeStr不为空，转换为AccountType
	accountType = cenum.AccountType(accountTypeStr)

	// 4. 如果accountType不合法，返回
	if err = accountType.EnumCheck(); err != nil {
		return
	}

	// 5. 如果accountType合法，设置isExist为true
	isExist = true

	return
}

// GetAccountIDFromHeaderMap 从headerMap中获取accountID
func GetAccountIDFromHeaderMap(headerMap map[string]string) (accountID string, isExist bool, err error) {
	if headerMap == nil {
		err = errors.New("headerMap is nil")
		return
	}

	// 1. 从headerMap中获取accountID
	if accountID = headerMap[cenum.HeaderXAccountID.String()]; accountID == "" {
		// 暂时保持向后兼容（后续可以删除）
		accountID = headerMap[cenum.HeaderXAccountIDOld.String()]
	}

	// 2. 如果accountID为空，返回
	if accountID == "" {
		return
	}

	// 3. 如果accountID不为空，设置isExist为true
	isExist = true

	return
}

// GetAccountIDFromContext 从context中获取accountID
func GetAccountIDFromContext(c *gin.Context) (accountID string, isExist bool, err error) {
	if c == nil {
		err = errors.New("c is nil")
		return
	}

	// 1. 从header中获取accountID
	accountID = c.GetHeader(cenum.HeaderXAccountID.String())
	if accountID == "" {
		// 暂时保持向后兼容（后续可以删除）
		accountID = c.GetHeader(cenum.HeaderXAccountIDOld.String())
	}

	// 2. 如果accountID为空，返回
	if accountID == "" {
		return
	}

	// 3. 如果accountID不为空，设置isExist为true
	isExist = true

	return
}
