package agentsvc

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformErrorToHTTPError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name           string
		err            interface{}
		wantNil        bool
		wantStatusCode int
	}{
		{
			name:    "nil errMap returns nil",
			err:     map[string]interface{}(nil),
			wantNil: true,
		},
		{
			name:           "non-map error returns 500",
			err:            "some string error",
			wantNil:        false,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "map without error_code returns 500",
			err: map[string]interface{}{
				"error_details": "some detail",
			},
			wantNil:        false,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "error_code not string returns 500",
			err: map[string]interface{}{
				"error_code": 12345,
			},
			wantNil:        false,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "ModelExecption error code",
			err: map[string]interface{}{
				"error_code":    "AgentExecutor.DolphinSDKException.ModelExecption",
				"error_details": "model error detail",
			},
			wantNil:        false,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "SkillExecption error code",
			err: map[string]interface{}{
				"error_code":    "AgentExecutor.DolphinSDKException.SkillExecption",
				"error_details": "skill error detail",
			},
			wantNil:        false,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "BaseExecption error code",
			err: map[string]interface{}{
				"error_code":    "AgentExecutor.DolphinSDKException.BaseExecption",
				"error_details": "base error detail",
			},
			wantNil:        false,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "unknown error code falls to default",
			err: map[string]interface{}{
				"error_code":    "AgentExecutor.UnknownError",
				"error_details": "unknown detail",
			},
			wantNil:        false,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := TransformErrorToHTTPError(ctx, tt.err)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantStatusCode, result.HTTPCode)
			}
		})
	}
}
