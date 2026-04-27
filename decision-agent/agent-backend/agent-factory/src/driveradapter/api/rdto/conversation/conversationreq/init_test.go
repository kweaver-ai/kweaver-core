package conversationreq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestInitReq_StructFields(t *testing.T) {
	t.Parallel()

	req := InitReq{
		AgentAPPKey:       "app-key-123",
		Title:             "Test Conversation",
		UserID:            "user-456",
		TempareaId:        "temp-789",
		VisitorType:       "user",
		AgentID:           "agent-001",
		AgentVersion:      "1.0.0",
		ExecutorVersion:   "v2",
		XAccountID:        "acc-123",
		XAccountType:      cenum.AccountTypeUser,
		XBusinessDomainID: "domain-456",
	}

	assert.Equal(t, "app-key-123", req.AgentAPPKey)
	assert.Equal(t, "Test Conversation", req.Title)
	assert.Equal(t, "user-456", req.UserID)
	assert.Equal(t, "temp-789", req.TempareaId)
	assert.Equal(t, "user", req.VisitorType)
	assert.Equal(t, "agent-001", req.AgentID)
	assert.Equal(t, "1.0.0", req.AgentVersion)
	assert.Equal(t, "v2", req.ExecutorVersion)
	assert.Equal(t, "acc-123", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeUser, req.XAccountType)
	assert.Equal(t, "domain-456", req.XBusinessDomainID)
}

func TestInitReq_Empty(t *testing.T) {
	t.Parallel()

	req := InitReq{}

	assert.Empty(t, req.AgentAPPKey)
	assert.Empty(t, req.Title)
	assert.Empty(t, req.UserID)
	assert.Empty(t, req.TempareaId)
	assert.Empty(t, req.VisitorType)
	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.Empty(t, req.ExecutorVersion)
	assert.Empty(t, req.XAccountID)
	assert.Empty(t, req.XBusinessDomainID)
}

func TestInitReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := InitReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"title"不能为空`, errMsgMap["Title.required"])
}

func TestInitReq_ReqCheck(t *testing.T) {
	t.Parallel()

	req := InitReq{
		Title: "Test Conversation",
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestInitReq_ReqCheck_Empty(t *testing.T) {
	t.Parallel()

	req := InitReq{}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestInitReq_WithAgentAPPKey(t *testing.T) {
	t.Parallel()

	keys := []string{
		"app-key-001",
		"agent-app-xyz",
		"key-中文-123",
	}

	for _, key := range keys {
		req := InitReq{
			AgentAPPKey: key,
		}
		assert.Equal(t, key, req.AgentAPPKey)
	}
}

func TestInitReq_WithUserID(t *testing.T) {
	t.Parallel()

	userIDs := []string{
		"user-001",
		"user-xyz",
		"用户123",
	}

	for _, userID := range userIDs {
		req := InitReq{
			UserID: userID,
		}
		assert.Equal(t, userID, req.UserID)
	}
}

func TestInitReq_WithTitle(t *testing.T) {
	t.Parallel()

	titles := []string{
		"Test Conversation",
		"中文会话标题",
		"Conversation with numbers 123",
		"Title with special chars !@#$%",
	}

	for _, title := range titles {
		req := InitReq{
			Title: title,
		}
		assert.Equal(t, title, req.Title)
	}
}

func TestInitReq_WithTempareaId(t *testing.T) {
	t.Parallel()

	tempareaIds := []string{
		"temp-001",
		"temp-xyz",
		"",
		"临时-123",
	}

	for _, tempareaId := range tempareaIds {
		req := InitReq{
			TempareaId: tempareaId,
		}
		assert.Equal(t, tempareaId, req.TempareaId)
	}
}

func TestInitReq_WithVisitorType(t *testing.T) {
	t.Parallel()

	visitorTypes := []string{
		"user",
		"app",
		"anonymous",
	}

	for _, vt := range visitorTypes {
		req := InitReq{
			VisitorType: vt,
		}
		assert.Equal(t, vt, req.VisitorType)
	}
}

func TestInitReq_WithAgentID(t *testing.T) {
	t.Parallel()

	agentIDs := []string{
		"agent-001",
		"agent-xyz",
		"智能体-123",
	}

	for _, agentID := range agentIDs {
		req := InitReq{
			AgentID: agentID,
		}
		assert.Equal(t, agentID, req.AgentID)
	}
}

func TestInitReq_WithAgentVersion(t *testing.T) {
	t.Parallel()

	versions := []string{
		"1.0.0",
		"2.1.3",
		"3.0.0-alpha",
		"",
	}

	for _, version := range versions {
		req := InitReq{
			AgentVersion: version,
		}
		assert.Equal(t, version, req.AgentVersion)
	}
}

func TestInitReq_WithExecutorVersion(t *testing.T) {
	t.Parallel()

	versions := []string{
		"v1",
		"v2",
		"",
	}

	for _, version := range versions {
		req := InitReq{
			ExecutorVersion: version,
		}
		assert.Equal(t, version, req.ExecutorVersion)
	}
}

func TestInitReq_WithXAccountID(t *testing.T) {
	t.Parallel()

	accountIDs := []string{
		"acc-001",
		"acc-xyz",
		"账户-123",
		"",
	}

	for _, accountID := range accountIDs {
		req := InitReq{
			XAccountID: accountID,
		}
		assert.Equal(t, accountID, req.XAccountID)
	}
}

func TestInitReq_WithXAccountType(t *testing.T) {
	t.Parallel()

	accountTypes := []cenum.AccountType{
		cenum.AccountTypeUser,
		cenum.AccountTypeApp,
		cenum.AccountTypeAnonymous,
	}

	for _, accountType := range accountTypes {
		req := InitReq{
			XAccountType: accountType,
		}
		assert.Equal(t, accountType, req.XAccountType)
	}
}

func TestInitReq_WithXBusinessDomainID(t *testing.T) {
	t.Parallel()

	domainIDs := []string{
		"domain-001",
		"domain-xyz",
		"域-123",
		"",
	}

	for _, domainID := range domainIDs {
		req := InitReq{
			XBusinessDomainID: domainID,
		}
		assert.Equal(t, domainID, req.XBusinessDomainID)
	}
}

func TestInitReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := InitReq{
		AgentAPPKey:       "app-key-abc",
		Title:             "Complete Conversation",
		UserID:            "user-def",
		TempareaId:        "temp-ghi",
		VisitorType:       "app",
		AgentID:           "agent-jkl",
		AgentVersion:      "2.5.0",
		ExecutorVersion:   "v1",
		XAccountID:        "acc-mno",
		XAccountType:      cenum.AccountTypeApp,
		XBusinessDomainID: "domain-pqr",
	}

	assert.Equal(t, "app-key-abc", req.AgentAPPKey)
	assert.Equal(t, "Complete Conversation", req.Title)
	assert.Equal(t, "user-def", req.UserID)
	assert.Equal(t, "temp-ghi", req.TempareaId)
	assert.Equal(t, "app", req.VisitorType)
	assert.Equal(t, "agent-jkl", req.AgentID)
	assert.Equal(t, "2.5.0", req.AgentVersion)
	assert.Equal(t, "v1", req.ExecutorVersion)
	assert.Equal(t, "acc-mno", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeApp, req.XAccountType)
	assert.Equal(t, "domain-pqr", req.XBusinessDomainID)

	err := req.ReqCheck()
	assert.NoError(t, err)
}
