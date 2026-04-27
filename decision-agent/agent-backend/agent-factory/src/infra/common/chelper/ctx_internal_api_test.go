package chelper

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestIsInternalAPIFromCtx_WithTrueValue(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), cenum.InternalAPIFlagCtxKey.String(), true) //nolint:staticcheck // SA1029

	isInternal := IsInternalAPIFromCtx(ctx)

	assert.True(t, isInternal)
}

func TestIsInternalAPIFromCtx_WithFalseValue(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), cenum.InternalAPIFlagCtxKey.String(), false) //nolint:staticcheck // SA1029

	isInternal := IsInternalAPIFromCtx(ctx)

	assert.False(t, isInternal)
}

func TestIsInternalAPIFromCtx_WithNilValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	isInternal := IsInternalAPIFromCtx(ctx)

	assert.False(t, isInternal)
}

func TestIsInternalAPIFromCtx_WithInvalidType(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), cenum.InternalAPIFlagCtxKey.String(), "not_a_bool") //nolint:staticcheck // SA1029

	assert.Panics(t, func() {
		IsInternalAPIFromCtx(ctx)
	})
}

func TestIsInternalAPIFromCtx_WithIntValue(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), cenum.InternalAPIFlagCtxKey.String(), 1) //nolint:staticcheck // SA1029

	assert.Panics(t, func() {
		IsInternalAPIFromCtx(ctx)
	})
}

func TestIsInternalAPIFromCtx_DerivedContext(t *testing.T) {
	t.Parallel()

	baseCtx := context.Background()
	ctx := context.WithValue(baseCtx, cenum.InternalAPIFlagCtxKey.String(), true) //nolint:staticcheck // SA1029

	isInternal := IsInternalAPIFromCtx(ctx)

	assert.True(t, isInternal)
}
