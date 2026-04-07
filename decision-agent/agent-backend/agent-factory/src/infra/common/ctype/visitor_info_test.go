package ctype

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestVisitorInfo_StructFields(t *testing.T) {
	t.Parallel()

	t.Run("creates visitor info with all fields", func(t *testing.T) {
		t.Parallel()

		info := VisitorInfo{
			XAccountID:        "account-123",
			XAccountType:      cenum.AccountTypeUser,
			XBusinessDomainID: cenum.BizDomainPublic,
		}

		assert.Equal(t, "account-123", info.XAccountID)
		assert.Equal(t, cenum.AccountTypeUser, info.XAccountType)
		assert.Equal(t, cenum.BizDomainPublic, info.XBusinessDomainID)
	})

	t.Run("allows empty visitor info", func(t *testing.T) {
		t.Parallel()

		info := VisitorInfo{}

		assert.Empty(t, info.XAccountID)
	})
}
