package capierr

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
)

// ==================== NewCustomXxxErr functions ====================
// 使用 rest.PublicError_* 已注册的 errorCode 来测试 NewCustomXxxErr 函数

func TestNewCustom400Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	err := NewCustom400Err(ctx, rest.PublicError_BadRequest, "bad input")

	assert.NotNil(t, err)
}

func TestNewCustom401Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	err := NewCustom401Err(ctx, rest.PublicError_Unauthorized, "no auth")

	assert.NotNil(t, err)
}

func TestNewCustom403Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	err := NewCustom403Err(ctx, rest.PublicError_Forbidden, "forbidden")

	assert.NotNil(t, err)
}

func TestNewCustom404Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	err := NewCustom404Err(ctx, rest.PublicError_NotFound, "not found")

	assert.NotNil(t, err)
}

func TestNewCustom405Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	err := NewCustom405Err(ctx, rest.PublicError_MethodNotAllowed, "method")

	assert.NotNil(t, err)
}

func TestNewCustom409Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	err := NewCustom409Err(ctx, rest.PublicError_Conflict, "conflict")

	assert.NotNil(t, err)
}

func TestNewCustom500Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	err := NewCustom500Err(ctx, rest.PublicError_InternalServerError, "server error")

	assert.NotNil(t, err)
}

// ==================== Error Code Constants ====================

func TestErrorCodes_AreDefined(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, DataAgentConfigLlmRequired)
	assert.NotEmpty(t, DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize)

	assert.Contains(t, DataAgentConfigLlmRequired, "AgentFactory")
	assert.Contains(t, DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize, "AgentFactory")
}

func TestErrorCodes_ConstantValues(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "AgentFactory.DataAgentConfig.BadRequest.LlmRequired", DataAgentConfigLlmRequired)
	assert.Contains(t, DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize, "KnEntryExceedLimitSize")
}

func TestErrorCodes_Uniqueness(t *testing.T) {
	t.Parallel()

	assert.NotEqual(t, DataAgentConfigLlmRequired, DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize)
}
