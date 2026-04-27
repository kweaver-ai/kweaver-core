package capierr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew400Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	detail := "Invalid input"

	err := New400Err(ctx, detail)

	assert.NotNil(t, err)
}

func TestNew401Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	detail := "Unauthorized access"

	err := New401Err(ctx, detail)

	assert.NotNil(t, err)
}

func TestNew403Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	detail := "Forbidden"

	err := New403Err(ctx, detail)

	assert.NotNil(t, err)
}

func TestNew404Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	detail := "Resource not found"

	err := New404Err(ctx, detail)

	assert.NotNil(t, err)
}

func TestNew405Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	detail := "Method not allowed"

	err := New405Err(ctx, detail)

	assert.NotNil(t, err)
}

func TestNew409Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	detail := "Conflict detected"

	err := New409Err(ctx, detail)

	assert.NotNil(t, err)
}

func TestNew500Err(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	detail := "Internal server error"

	err := New500Err(ctx, detail)

	assert.NotNil(t, err)
}

func TestNew400Err_WithNilDetail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	err := New400Err(ctx, nil)

	assert.NotNil(t, err)
}

func TestNew401Err_WithStructDetail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type ErrorDetail struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}

	detail := ErrorDetail{Field: "name", Message: "Name is required"}

	err := New401Err(ctx, detail)

	assert.NotNil(t, err)
}

func TestErrorCodes(t *testing.T) {
	t.Parallel()

	// Test that error code constants are defined
	assert.NotEmpty(t, DataAgentConfigLlmRequired)
	assert.NotEmpty(t, DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize)
}
