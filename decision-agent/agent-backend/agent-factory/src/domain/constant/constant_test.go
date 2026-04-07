package constant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		callType CallType
		want     string
	}{
		{
			name:     "Chat",
			callType: Chat,
			want:     "chat",
		},
		{
			name:     "DebugChat",
			callType: DebugChat,
			want:     "debug_chat",
		},
		{
			name:     "APIChat",
			callType: APIChat,
			want:     "api_chat",
		},
		{
			name:     "InternalChat",
			callType: InternalChat,
			want:     "internal_chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.callType.String())
		})
	}
}

func TestCallType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, CallType("chat"), Chat)
	assert.Equal(t, CallType("debug_chat"), DebugChat)
	assert.Equal(t, CallType("api_chat"), APIChat)
	assert.Equal(t, CallType("internal_chat"), InternalChat)
}

func TestChatMode_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "normal", NormalMode)
	assert.Equal(t, "deep_thinking", DeepThinkingMode)
}

func TestVisitorType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		visitorType VisitorType
		want        string
	}{
		{
			name:        "RealName",
			visitorType: RealName,
			want:        "realname",
		},
		{
			name:        "Anonymous",
			visitorType: Anonymous,
			want:        "anonymous",
		},
		{
			name:        "Business",
			visitorType: Business,
			want:        "business",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.visitorType.String())
		})
	}
}

func TestVisitorType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, VisitorType("realname"), RealName)
	assert.Equal(t, VisitorType("anonymous"), Anonymous)
	assert.Equal(t, VisitorType("business"), Business)
}

func TestIsShowOriginResponseFromCtx_True(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), CtxKeyIsShowOriginResponse, "true")
	assert.True(t, IsShowOriginResponseFromCtx(ctx))
}

func TestIsShowOriginResponseFromCtx_False(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), CtxKeyIsShowOriginResponse, "false")
	assert.False(t, IsShowOriginResponseFromCtx(ctx))
}

func TestIsShowOriginResponseFromCtx_NotSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assert.False(t, IsShowOriginResponseFromCtx(ctx))
}

func TestIsShowOriginResponseFromCtx_OtherValue(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), CtxKeyIsShowOriginResponse, "yes")
	assert.False(t, IsShowOriginResponseFromCtx(ctx))
}
