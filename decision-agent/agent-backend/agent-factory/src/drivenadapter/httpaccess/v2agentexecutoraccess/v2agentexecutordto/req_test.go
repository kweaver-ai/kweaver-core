package v2agentexecutordto

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestV2AgentCallReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &V2AgentCallReq{
		AgentID:      "agent-123",
		AgentVersion: "1.0.0",
		AgentInput:   map[string]interface{}{"question": "test"},
		UserID:       "user-456",
		VisitorType:  constant.RealName,
		Token:        "token-789",
		XAccountID:   "account-101",
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Equal(t, "1.0.0", req.AgentVersion)
	assert.Equal(t, "user-456", req.UserID)
	assert.Equal(t, "token-789", req.Token)
	assert.Equal(t, "account-101", req.XAccountID)
}

func TestAgentOptions_StructFields(t *testing.T) {
	t.Parallel()

	retrieverFields := RetrieverDataSource{
		Kg:  []*KgSource{{KgID: "kg-1"}},
		Doc: []*DocSource{{ID: "doc-1"}},
	}

	options := &AgentOptions{
		Stream:                 true,
		Debug:                  false,
		Retry:                  true,
		DynamicRetrieverFields: retrieverFields,
		Step:                   "step1",
		ConversationID:         "conv-123",
		AgentRunID:             "run-456",
		IsNeedProgress:         true,
		EnableDependencyCache:  false,
	}

	assert.True(t, options.Stream)
	assert.False(t, options.Debug)
	assert.True(t, options.Retry)
	assert.Equal(t, "step1", options.Step)
	assert.Equal(t, "conv-123", options.ConversationID)
	assert.True(t, options.IsNeedProgress)
	assert.False(t, options.EnableDependencyCache)
}

func TestKgSource_StructFields(t *testing.T) {
	t.Parallel()

	fieldProps := map[string][]string{
		"prop1": {"value1", "value2"},
	}

	kg := &KgSource{
		KgID:            "kg-123",
		Fields:          []string{"field1", "field2"},
		OutputFields:    []string{"out1"},
		FieldProperties: fieldProps,
	}

	assert.Equal(t, "kg-123", kg.KgID)
	assert.Len(t, kg.Fields, 2)
	assert.Contains(t, kg.Fields, "field1")
	assert.Len(t, kg.OutputFields, 1)
	assert.NotNil(t, kg.FieldProperties)
}

func TestDocFields_StructFields(t *testing.T) {
	t.Parallel()

	fields := &DocFields{
		Name:   "test_field",
		Path:   "/path/to/field",
		Source: "test_source",
	}

	assert.Equal(t, "test_field", fields.Name)
	assert.Equal(t, "/path/to/field", fields.Path)
	assert.Equal(t, "test_source", fields.Source)
}

func TestDocSource_StructFields(t *testing.T) {
	t.Parallel()

	docFields := []*DocFields{
		{Name: "field1", Path: "/path1", Source: "src1"},
		{Name: "field2", Path: "/path2", Source: "src2"},
	}

	doc := &DocSource{
		FileSource: "local",
		ID:         "doc-123",
		Name:       "Test Document",
		DsID:       "ds-456",
		Fields:     docFields,
		DataSets:   []string{"dataset1", "dataset2"},
		Address:    "localhost",
		Port:       8080,
		AsUserID:   "user-789",
		Disabled:   false,
	}

	assert.Equal(t, "local", doc.FileSource)
	assert.Equal(t, "doc-123", doc.ID)
	assert.Equal(t, "Test Document", doc.Name)
	assert.Len(t, doc.Fields, 2)
	assert.Equal(t, "field1", doc.Fields[0].Name)
	assert.Equal(t, "localhost", doc.Address)
	assert.Equal(t, 8080, doc.Port)
	assert.False(t, doc.Disabled)
}

func TestRetrieverDataSource_StructFields(t *testing.T) {
	t.Parallel()

	dataSource := &RetrieverDataSource{
		Kg: []*KgSource{
			{KgID: "kg-1", Fields: []string{"f1"}},
			{KgID: "kg-2", Fields: []string{"f2"}},
		},
		Doc: []*DocSource{
			{ID: "doc-1"},
			{ID: "doc-2"},
		},
	}

	assert.Len(t, dataSource.Kg, 2)
	assert.Len(t, dataSource.Doc, 2)
	assert.Equal(t, "kg-1", dataSource.Kg[0].KgID)
	assert.Equal(t, "doc-1", dataSource.Doc[0].ID)
}

func TestConfig_StructFields(t *testing.T) {
	t.Parallel()

	config := Config{
		Config: daconfvalobj.Config{},
	}

	assert.NotNil(t, config.Config)
}

func TestV2AgentCallReq_WithAccountType(t *testing.T) {
	t.Parallel()

	req := &V2AgentCallReq{
		XAccountID:   "app-account",
		XAccountType: cenum.AccountTypeApp,
		UserID:       "user-123",
	}

	assert.Equal(t, "app-account", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeApp, req.XAccountType)
}

func TestAgentOptions_WithResumeInfo(t *testing.T) {
	t.Parallel()

	interruptHandle := &InterruptHandle{
		FrameID:       "frame-123",
		SnapshotID:    "snapshot-456",
		ResumeToken:   "token-789",
		InterruptType: "tool_call",
		CurrentBlock:  1,
		RestartBlock:  false,
	}

	resumeInfo := &AgentResumeInfo{
		ResumeHandle: interruptHandle,
		Action:       "confirm",
		ModifiedArgs: []ModifiedArg{{Key: "arg1", Value: "value1"}},
	}

	options := &AgentOptions{
		ResumeInfo: resumeInfo,
	}

	assert.NotNil(t, options.ResumeInfo)
	assert.Equal(t, "confirm", options.ResumeInfo.Action)
	assert.NotNil(t, options.ResumeInfo.ResumeHandle)
	assert.Equal(t, "frame-123", options.ResumeInfo.ResumeHandle.FrameID)
}
