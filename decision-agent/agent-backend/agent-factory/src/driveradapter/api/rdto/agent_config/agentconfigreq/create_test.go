package agentconfigreq

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum/agentconfigenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &CreateReq{}

	errMap := req.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Contains(t, errMap, "Key.max")
	assert.Equal(t, `"key"长度不能超过50`, errMap["Key.max"])
}

func TestCreateReq_ReqCheckWithCtx(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		prepareReq  func() *CreateReq
		wantErr     bool
		errContains string
	}{
		{
			name: "external api with created_by should return error",
			prepareReq: func() *CreateReq {
				req := createReqTestNewReq(false)
				req.CreatedBy = "creator-1"
				return req
			},
			wantErr:     true,
			errContains: "created_by is valid when is_private_api is false",
		},
		{
			name: "internal api without created_by should return error",
			prepareReq: func() *CreateReq {
				req := createReqTestNewReq(true)
				req.CreatedBy = ""
				return req
			},
			wantErr:     true,
			errContains: "created_by is required when is_private_api is true",
		},
		{
			name: "external api with built-in should return error",
			prepareReq: func() *CreateReq {
				req := createReqTestNewReq(false)
				builtInYes := cdaenum.BuiltInYes
				req.IsBuiltIn = &builtInYes
				return req
			},
			wantErr:     true,
			errContains: "is_built_in is invalid when is_private_api is false",
		},
		{
			name: "nested update req error should be wrapped",
			prepareReq: func() *CreateReq {
				req := createReqTestNewReq(true)
				req.ProductKey = ""
				return req
			},
			wantErr:     true,
			errContains: "update_req is invalid",
		},
		{
			name: "valid internal api request should pass",
			prepareReq: func() *CreateReq {
				req := createReqTestNewReq(true)
				req.CreatedBy = "creator-1"
				return req
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := tc.prepareReq()
			err := req.ReqCheckWithCtx(context.Background())

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestCreateReq_D2e(t *testing.T) {
	t.Parallel()

	t.Run("should auto-generate key when key is empty", func(t *testing.T) {
		t.Parallel()

		req := createReqTestNewReq(true)
		req.Key = ""

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.NotEmpty(t, eo.Key)
		require.NotNil(t, eo.Config)
		assert.Equal(t, agentconfigenum.ConfigTplVersionV1, eo.Config.GetConfigMetadata().GetConfigTplVersion())
		assert.Greater(t, eo.Config.GetConfigMetadata().ConfigLastSetTimestamp, uint64(0))
	})

	t.Run("should keep key when provided", func(t *testing.T) {
		t.Parallel()

		req := createReqTestNewReq(true)
		req.Key = "fixed-key"

		eo, err := req.D2e()

		require.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, "fixed-key", eo.Key)
	})
}

func createReqTestNewReq(isInternalAPI bool) *CreateReq {
	updateReq := &UpdateReq{
		Name:          "AgentName",
		Profile:       "Agent Profile",
		AvatarType:    cdaenum.AvatarTypeBuiltIn,
		Avatar:        "avatar-1",
		ProductKey:    string(cdaenum.ProductDIP),
		Config:        createReqTestValidConfig(!isInternalAPI),
		IsInternalAPI: isInternalAPI,
	}

	if isInternalAPI {
		updateReq.UpdatedBy = "updater-1"
	}

	return &CreateReq{
		UpdateReq: updateReq,
	}
}

func createReqTestValidConfig(withLlms bool) *daconfvalobj.Config {
	cfg := daconfvalobj.NewConfig()
	cfg.Input = &daconfvalobj.Input{
		Fields: daconfvalobj.Fields{
			&daconfvalobj.Field{Name: "query", Type: cdaenum.InputFieldTypeString},
		},
	}
	cfg.Output = &daconfvalobj.Output{
		DefaultFormat: cdaenum.OutputDefaultFormatJson,
		Variables:     &daconfvalobj.VariablesS{},
	}
	cfg.IsDolphinMode = cdaenum.DolphinModeDisabled
	cfg.IsDataFlowSetEnabled = 1

	if withLlms {
		cfg.Llms = []*daconfvalobj.LlmItem{
			{
				IsDefault: true,
				LlmConfig: &daconfvalobj.LlmConfig{
					Name:      "mock-llm",
					MaxTokens: 100,
				},
			},
		}
	}

	return cfg
}
