package agentconfigreq

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum/agentconfigenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/customvalidator"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{}

	errMap := req.GetErrMsgMap()

	// Verify the error message map is not nil
	assert.NotNil(t, errMap)
}

func TestUpdateReq_CustomCheck_NonInternalAPI_WithUpdatedBy(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{}
	req.IsInternalAPI = false
	req.UpdatedBy = "user-123"

	err := req.CustomCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updated_by is valid when is_private_api is false")
}

func TestUpdateReq_CustomCheck_InternalAPI_WithoutUpdatedBy(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{}
	req.IsInternalAPI = true
	req.UpdatedBy = ""

	err := req.CustomCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updated_by is required when is_private_api is true")
}

func TestUpdateReq_CustomCheck_InternalAPI_WithUpdatedBy(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{}
	req.IsInternalAPI = true
	req.UpdatedBy = "user-123"

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestUpdateReq_CustomCheck_NonInternalAPI_WithoutUpdatedBy(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{}
	req.IsInternalAPI = false
	req.UpdatedBy = ""

	err := req.CustomCheck()

	assert.NoError(t, err)
}

func TestUpdateReq_D2e_WithAllFields(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Name:       "Test Agent",
		Profile:    "Test Profile",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
		Config:     daconfvalobj.NewConfig(),
		CreatedBy:  "user-123",
		UpdatedBy:  "user-123",
	}

	eo, err := req.D2e()

	assert.NoError(t, err)
	require.NotNil(t, eo)
	assert.Equal(t, "Test Agent", eo.Name)
	assert.Equal(t, "product-123", eo.ProductKey)
	assert.Equal(t, "Test Profile", eo.GetProfileStr())
	assert.Equal(t, "avatar-123", eo.Avatar)
}

func TestUpdateReq_D2e_WithConfig(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Name:       "Test Agent",
		Profile:    "Test Profile",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
	}

	// Create a valid Config
	config := daconfvalobj.NewConfig()
	config.Metadata.SetConfigTplVersion(agentconfigenum.ConfigTplVersionV1)
	req.Config = config

	eo, err := req.D2e()

	assert.NoError(t, err)
	require.NotNil(t, eo)
	assert.Equal(t, "Test Agent", eo.Name)
	assert.NotNil(t, eo.Config)
}

func TestUpdateReq_D2e_WithIsBuiltIn(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Name:       "Test Agent",
		Profile:    "Test Profile",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
		Config:     daconfvalobj.NewConfig(),
	}

	builtIn := cdaenum.BuiltInYes
	req.IsBuiltIn = &builtIn

	eo, err := req.D2e()

	assert.NoError(t, err)
	require.NotNil(t, eo)
	assert.Equal(t, cdaenum.BuiltInYes, *eo.IsBuiltIn)
}

func TestUpdateReq_D2e_WithCreatedBy(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Name:       "Test Agent",
		Profile:    "Test Profile",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
		Config:     daconfvalobj.NewConfig(),
	}
	req.CreatedBy = "creator-123"

	eo, err := req.D2e()

	assert.NoError(t, err)
	require.NotNil(t, eo)
	assert.Equal(t, "creator-123", eo.CreatedBy)
}

func TestUpdateReq_IsChanged_DifferentName(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Name:       "Updated Agent Name",
		Profile:    "Test Profile",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
		Config:     daconfvalobj.NewConfig(),
	}

	profile := "Test Profile"
	configStr, _ := cutil.JSON().MarshalToString(req.Config)

	oldPo := &dapo.DataAgentPo{
		Name:       "Original Name",
		Profile:    &profile,
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
		Config:     configStr,
	}

	result := req.IsChanged(oldPo)

	assert.True(t, result)
}

func TestUpdateReq_IsChanged_SameData(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Name:       "Test Agent",
		Profile:    "Test Profile",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
		Config:     daconfvalobj.NewConfig(),
	}

	profile := "Test Profile"
	configStr, _ := cutil.JSON().MarshalToString(req.Config)

	oldPo := &dapo.DataAgentPo{
		Name:       "Test Agent",
		Profile:    &profile,
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
		Config:     configStr,
	}

	result := req.IsChanged(oldPo)

	// Since we use the same config data, it should be considered not changed
	assert.False(t, result)
}

func TestUpdateReq_D2e_WithStatus(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Name:       "Test Agent",
		Profile:    "Test Profile",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
		Config:     daconfvalobj.NewConfig(),
	}

	// Note: Status is not directly settable in UpdateReq
	// It might come from the Config or be set during D2e conversion

	eo, err := req.D2e()

	assert.NoError(t, err)
	require.NotNil(t, eo)
	// Status might have a default value
}

func TestUpdateReq_D2e_NilConfig_Panics(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{
		Name:       "Test Agent",
		Profile:    "Test Profile",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar-123",
		ProductKey: "product-123",
		// Config is nil - should panic
	}

	assert.Panics(t, func() {
		_, _ = req.D2e()
	})
}

func TestUpdateReq_D2e_WithDifferentAvatarTypes(t *testing.T) {
	t.Parallel()

	avatarTypes := []cdaenum.AvatarType{
		cdaenum.AvatarTypeBuiltIn,
		cdaenum.AvatarTypeUserUploaded,
		cdaenum.AvatarTypeAIGenerated,
	}

	for _, avatarType := range avatarTypes {
		req := &UpdateReq{
			Name:       "Test Agent",
			Profile:    "Test Profile",
			AvatarType: avatarType,
			Avatar:     "avatar-123",
			ProductKey: "product-123",
			Config:     daconfvalobj.NewConfig(),
		}

		eo, err := req.D2e()

		assert.NoError(t, err)
		require.NotNil(t, eo)
		assert.Equal(t, avatarType, eo.AvatarType)
	}
}

func updateReqTestRegisterCheckAgentAndTplName(t *testing.T) {
	t.Helper()

	v, ok := binding.Validator.Engine().(*validator.Validate)
	require.True(t, ok)

	_ = v.RegisterValidation("checkAgentAndTplName", customvalidator.CheckAgentAndTplName)
}

func TestUpdateReq_Validate(t *testing.T) {
	t.Parallel()

	updateReqTestRegisterCheckAgentAndTplName(t)

	t.Run("missing required fields should return wrapped error", func(t *testing.T) {
		t.Parallel()

		req := &UpdateReq{}

		err := req.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "[UpdateReq] invalid")
	})

	t.Run("valid request should pass", func(t *testing.T) {
		t.Parallel()

		req := updateReqTestNewReq(false)

		err := req.Validate()

		assert.NoError(t, err)
	})
}

func TestUpdateReq_ReqCheckWithCtx(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		prepareReq  func() *UpdateReq
		wantErr     bool
		errContains string
	}{
		{
			name: "invalid avatar type",
			prepareReq: func() *UpdateReq {
				req := updateReqTestNewReq(true)
				req.AvatarType = 0
				return req
			},
			wantErr:     true,
			errContains: "avatar_type is invalid",
		},
		{
			name: "empty product key",
			prepareReq: func() *UpdateReq {
				req := updateReqTestNewReq(true)
				req.ProductKey = ""
				return req
			},
			wantErr:     true,
			errContains: "product_key is required",
		},
		{
			name: "doc datasource with ChatBI product should fail",
			prepareReq: func() *UpdateReq {
				req := updateReqTestNewReq(true)
				req.ProductKey = string(cdaenum.ProductChatBI)
				req.Config.DataSource = updateReqTestNewValidDocDataSource()
				return req
			},
			wantErr:     true,
			errContains: "data source is invalid",
		},
		{
			name: "valid request should pass",
			prepareReq: func() *UpdateReq {
				return updateReqTestNewReq(true)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.prepareReq().ReqCheckWithCtx(context.Background())

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestUpdateReq_IsConfigChanged(t *testing.T) {
	t.Parallel()

	req := &UpdateReq{}

	t.Run("metadata difference should be ignored", func(t *testing.T) {
		t.Parallel()

		oldConfig := `{"input":{"fields":[]},"metadata":{"config_tpl_version":"v1"}}`
		newConfig := `{"input":{"fields":[]},"metadata":{"config_tpl_version":"v2"}}`

		changed, err := req.IsConfigChanged(oldConfig, newConfig)

		require.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("invalid old json should return error", func(t *testing.T) {
		t.Parallel()

		changed, err := req.IsConfigChanged("not-json", `{"input":{"fields":[]}}`)

		require.Error(t, err)
		assert.False(t, changed)
	})
}

func updateReqTestNewReq(isInternalAPI bool) *UpdateReq {
	req := &UpdateReq{
		Name:          "AgentName",
		Profile:       "Agent Profile",
		AvatarType:    cdaenum.AvatarTypeBuiltIn,
		Avatar:        "avatar-1",
		ProductKey:    string(cdaenum.ProductDIP),
		Config:        createReqTestValidConfig(!isInternalAPI),
		IsInternalAPI: isInternalAPI,
	}

	if isInternalAPI {
		req.UpdatedBy = "updater-1"
	}

	return req
}

func updateReqTestNewValidDocDataSource() *datasourcevalobj.RetrieverDataSource {
	retrievalSlicesNum := 100
	maxSlicePerCite := 5
	rerankTopK := 10
	sliceHeadNum := 2
	sliceTailNum := 0
	documentsNum := 8
	docThreshold := -5.5
	retrievalMaxLength := 1000

	return &datasourcevalobj.RetrieverDataSource{
		Doc: []*datasourcevalobj.DocSource{
			{
				DsID: "doc-1",
				Fields: []*datasourcevalobj.DocSourceField{
					{
						Name:   "test_field",
						Path:   "test/path",
						Source: "gns://92EE2D87255142B78A6F1DFB6BBB836B/B08AC060A758422583A851C601C0A89B",
						Type:   cdaenum.DocSourceFieldTypeFile,
					},
				},
			},
		},
		AdvancedConfig: &datasourcevalobj.RetrieverAdvancedConfig{
			Doc: &datasourcevalobj.DocAdvancedConfig{
				RetrievalSlicesNum: &retrievalSlicesNum,
				MaxSlicePerCite:    &maxSlicePerCite,
				RerankTopK:         &rerankTopK,
				SliceHeadNum:       &sliceHeadNum,
				SliceTailNum:       &sliceTailNum,
				DocumentsNum:       &documentsNum,
				DocumentThreshold:  &docThreshold,
				RetrievalMaxLength: &retrievalMaxLength,
			},
		},
	}
}
