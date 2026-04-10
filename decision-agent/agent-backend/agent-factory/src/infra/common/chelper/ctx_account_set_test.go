package chelper

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestSetAccountInfoToHeaderMap_NilHeaderMap(t *testing.T) {
	t.Parallel()

	var headerMap map[string]string = nil

	// Should not panic with nil map
	SetAccountInfoToHeaderMap(headerMap, "account123", cenum.AccountTypeUser)

	assert.Nil(t, headerMap)
}

func TestSetAccountInfoToHeaderMap_ValidInput(t *testing.T) {
	t.Parallel()

	headerMap := make(map[string]string)
	accountID := "account123"
	accountType := cenum.AccountTypeUser

	SetAccountInfoToHeaderMap(headerMap, accountID, accountType)

	// Check new headers
	assert.Equal(t, accountID, headerMap[cenum.HeaderXAccountID.String()])
	assert.Equal(t, accountType.String(), headerMap[cenum.HeaderXAccountType.String()])

	// Check old headers for backward compatibility
	assert.Equal(t, accountID, headerMap[cenum.HeaderXAccountIDOld.String()])
	assert.Equal(t, accountType.String(), headerMap[cenum.HeaderXAccountTypeOld.String()])
}

func TestSetAccountInfoToHeaderMap_WithExistingValues(t *testing.T) {
	t.Parallel()

	headerMap := map[string]string{
		"existing_key": "existing_value",
	}
	accountID := "account456"
	accountType := cenum.AccountTypeApp

	SetAccountInfoToHeaderMap(headerMap, accountID, accountType)

	// Existing values should be preserved
	assert.Equal(t, "existing_value", headerMap["existing_key"])

	// New values should be set
	assert.Equal(t, accountID, headerMap[cenum.HeaderXAccountID.String()])
	assert.Equal(t, accountType.String(), headerMap[cenum.HeaderXAccountType.String()])
}

func TestSetAccountInfoToHeaderMap_OverwritesExistingAccountHeaders(t *testing.T) {
	t.Parallel()

	accountTypeUser := cenum.AccountTypeUser
	headerMap := map[string]string{
		cenum.HeaderXAccountID.String():      "old_account",
		cenum.HeaderXAccountType.String():    accountTypeUser.String(),
		cenum.HeaderXAccountIDOld.String():   "old_account_old",
		cenum.HeaderXAccountTypeOld.String(): accountTypeUser.String(),
	}
	newAccountID := "new_account"
	newAccountType := cenum.AccountTypeApp

	SetAccountInfoToHeaderMap(headerMap, newAccountID, newAccountType)

	// Values should be overwritten
	assert.Equal(t, newAccountID, headerMap[cenum.HeaderXAccountID.String()])
	assert.Equal(t, newAccountType.String(), headerMap[cenum.HeaderXAccountType.String()])
	assert.Equal(t, newAccountID, headerMap[cenum.HeaderXAccountIDOld.String()])
	assert.Equal(t, newAccountType.String(), headerMap[cenum.HeaderXAccountTypeOld.String()])
}

func TestSetAccountInfoToHeaderMap_EmptyAccountID(t *testing.T) {
	t.Parallel()

	headerMap := make(map[string]string)

	SetAccountInfoToHeaderMap(headerMap, "", cenum.AccountTypeUser)

	// Should set empty string
	assert.Equal(t, "", headerMap[cenum.HeaderXAccountID.String()])
	accountTypeUser := cenum.AccountTypeUser
	assert.Equal(t, accountTypeUser.String(), headerMap[cenum.HeaderXAccountType.String()])
}

func TestSetAccountInfoToHeaderMap_AllAccountTypes(t *testing.T) {
	t.Parallel()

	accountTypes := []cenum.AccountType{
		cenum.AccountTypeUser,
		cenum.AccountTypeApp,
		cenum.AccountTypeAnonymous,
	}

	for _, accountType := range accountTypes {
		t.Run(accountType.String(), func(t *testing.T) {
			t.Parallel()

			headerMap := make(map[string]string)
			accountID := "test_account"

			SetAccountInfoToHeaderMap(headerMap, accountID, accountType)

			assert.Equal(t, accountID, headerMap[cenum.HeaderXAccountID.String()])
			assert.Equal(t, accountType.String(), headerMap[cenum.HeaderXAccountType.String()])
		})
	}
}

func TestSetAccountInfoToHeaderMap_MultipleCalls(t *testing.T) {
	t.Parallel()

	headerMap := make(map[string]string)

	SetAccountInfoToHeaderMap(headerMap, "account1", cenum.AccountTypeUser)
	assert.Equal(t, "account1", headerMap[cenum.HeaderXAccountID.String()])

	SetAccountInfoToHeaderMap(headerMap, "account2", cenum.AccountTypeApp)
	assert.Equal(t, "account2", headerMap[cenum.HeaderXAccountID.String()])
	accountTypeApp := cenum.AccountTypeApp
	assert.Equal(t, accountTypeApp.String(), headerMap[cenum.HeaderXAccountType.String()])
}

func TestSetAccountInfoToHeaderMap_AccountTypeAnonymous(t *testing.T) {
	t.Parallel()

	headerMap := make(map[string]string)
	accountID := "anonymous_account"
	accountTypeAnonymous := cenum.AccountTypeAnonymous

	SetAccountInfoToHeaderMap(headerMap, accountID, accountTypeAnonymous)

	assert.Equal(t, accountID, headerMap[cenum.HeaderXAccountID.String()])
	assert.Equal(t, accountTypeAnonymous.String(), headerMap[cenum.HeaderXAccountType.String()])
}
