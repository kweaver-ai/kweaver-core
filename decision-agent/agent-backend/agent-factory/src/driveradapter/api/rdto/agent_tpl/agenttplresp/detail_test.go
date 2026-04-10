package agenttplresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDetailRes(t *testing.T) {
	t.Parallel()

	detail := NewDetailRes()

	assert.NotNil(t, detail)
	assert.IsType(t, &DetailRes{}, detail)
	assert.Equal(t, int64(0), detail.ID)
	assert.Empty(t, detail.Name)
}

func TestDetailRes_LoadFromEo(t *testing.T) {
	t.Parallel()

	t.Run("load from template entity", func(t *testing.T) {
		t.Parallel()

		profile := "Test profile"
		builtIn := cdaenum.BuiltInYes
		detail := NewDetailRes()
		eo := &daconfeo.DataAgentTpl{
			DataAgentTplPo: dapo.DataAgentTplPo{
				ID:         1,
				Name:       "Test Template",
				Key:        "test-tpl",
				Profile:    &profile,
				Avatar:     "avatar.png",
				AvatarType: 1,
				Status:     cdaenum.StatusPublished,
				ProductKey: "product-1",
				IsBuiltIn:  &builtIn,
				CreatedAt:  1234567890,
				UpdatedAt:  1234567899,
				CreatedBy:  "user-1",
				UpdatedBy:  "user-1",
			},
			ProductName: "Product One",
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		err := detail.LoadFromEo(eo)

		require.NoError(t, err)
		assert.Equal(t, int64(1), detail.ID)
		assert.Equal(t, "Test Template", detail.Name)
		assert.Equal(t, "test-tpl", detail.Key)
		assert.Equal(t, "Test profile", *detail.Profile)
		assert.Equal(t, "avatar.png", detail.Avatar)
		assert.Equal(t, 1, detail.AvatarType)
		assert.Equal(t, cdaenum.StatusPublished, detail.Status)
		assert.Equal(t, "product-1", detail.ProductKey)
		assert.Equal(t, "Product One", detail.ProductName)
		// Note: IsBuiltIn is converted from cdaenum.BuiltIn to int during JSON copying
		if detail.IsBuiltIn != nil {
			assert.Equal(t, int(cdaenum.BuiltInYes), *detail.IsBuiltIn)
		}

		assert.NotNil(t, detail.Config)
	})

	t.Run("with nil entity", func(t *testing.T) {
		t.Parallel()

		detail := NewDetailRes()

		assert.Panics(t, func() {
			_ = detail.LoadFromEo(nil)
		})
	})

	t.Run("with nil profile", func(t *testing.T) {
		t.Parallel()

		detail := NewDetailRes()
		builtIn := cdaenum.BuiltInNo
		eo := &daconfeo.DataAgentTpl{
			DataAgentTplPo: dapo.DataAgentTplPo{
				ID:        2,
				Name:      "Template No Profile",
				Key:       "tpl-no-profile",
				Profile:   nil,
				Status:    cdaenum.StatusUnpublished,
				IsBuiltIn: &builtIn,
			},
		}

		err := detail.LoadFromEo(eo)

		require.NoError(t, err)
		assert.Equal(t, int64(2), detail.ID)
		assert.Nil(t, detail.Profile)
	})

	t.Run("with all fields", func(t *testing.T) {
		t.Parallel()

		detail := NewDetailRes()
		profile := "Complete profile"
		builtIn := cdaenum.BuiltInYes
		publishedAt := int64(3000000000)
		publishedBy := "publisher-1"
		eo := &daconfeo.DataAgentTpl{
			DataAgentTplPo: dapo.DataAgentTplPo{
				ID:          999,
				Name:        "Complete Template",
				Key:         "complete-tpl",
				Profile:     &profile,
				Avatar:      "complete.png",
				AvatarType:  2,
				Status:      cdaenum.StatusPublished,
				ProductKey:  "product-complete",
				IsBuiltIn:   &builtIn,
				CreatedAt:   1000000000,
				UpdatedAt:   2000000000,
				CreatedBy:   "admin-1",
				UpdatedBy:   "admin-2",
				PublishedAt: &publishedAt,
				PublishedBy: &publishedBy,
			},
			ProductName: "Complete Product",
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		err := detail.LoadFromEo(eo)

		require.NoError(t, err)
		assert.Equal(t, int64(999), detail.ID)
		assert.Equal(t, "Complete Template", detail.Name)
		assert.Equal(t, "complete-tpl", detail.Key)
		assert.Equal(t, "complete.png", detail.Avatar)
		assert.Equal(t, 2, detail.AvatarType)
		assert.Equal(t, cdaenum.StatusPublished, detail.Status)
		assert.Equal(t, "product-complete", detail.ProductKey)
		assert.Equal(t, "Complete Product", detail.ProductName)
		assert.Equal(t, int64(1000000000), detail.CreatedAt)
		assert.Equal(t, int64(2000000000), detail.UpdatedAt)
		assert.Equal(t, "admin-1", detail.CreatedBy)
		assert.Equal(t, "admin-2", detail.UpdatedBy)
		assert.Equal(t, int64(3000000000), detail.PublishedAt)
		assert.Equal(t, "publisher-1", detail.PublishedBy)
	})
}

func TestDetailRes_StructFields(t *testing.T) {
	t.Parallel()

	profile := "Test profile"
	isBuiltIn := int(cdaenum.BuiltInYes)
	detail := &DetailRes{
		ID:          123,
		Name:        "Test Template",
		Profile:     &profile,
		Key:         "test-tpl",
		Avatar:      "avatar.png",
		AvatarType:  1,
		Status:      cdaenum.StatusPublished,
		ProductKey:  "product-1",
		ProductName: "Product One",
		IsBuiltIn:   &isBuiltIn,
		Config: &daconfvalobj.Config{
			Input:  &daconfvalobj.Input{},
			Output: &daconfvalobj.Output{},
		},
		CreatedAt:   1234567890,
		UpdatedAt:   1234567899,
		CreatedBy:   "user-1",
		UpdatedBy:   "user-1",
		PublishedAt: 9876543210,
		PublishedBy: "publisher-1",
	}

	assert.Equal(t, int64(123), detail.ID)
	assert.Equal(t, "Test Template", detail.Name)
	assert.Equal(t, "Test profile", *detail.Profile)
	assert.Equal(t, "test-tpl", detail.Key)
	assert.Equal(t, "avatar.png", detail.Avatar)
	assert.Equal(t, 1, detail.AvatarType)
	assert.Equal(t, cdaenum.StatusPublished, detail.Status)
	assert.Equal(t, "product-1", detail.ProductKey)
	assert.Equal(t, "Product One", detail.ProductName)
	assert.NotNil(t, detail.Config)
	assert.Equal(t, int64(1234567890), detail.CreatedAt)
	assert.Equal(t, int64(1234567899), detail.UpdatedAt)
	assert.Equal(t, "user-1", detail.CreatedBy)
	assert.Equal(t, "user-1", detail.UpdatedBy)
	assert.Equal(t, int64(9876543210), detail.PublishedAt)
	assert.Equal(t, "publisher-1", detail.PublishedBy)
}

func TestDetailRes_Empty(t *testing.T) {
	t.Parallel()

	detail := &DetailRes{}

	assert.Equal(t, int64(0), detail.ID)
	assert.Empty(t, detail.Name)
	assert.Nil(t, detail.Profile)
	assert.Empty(t, detail.Key)
	assert.Empty(t, detail.Avatar)
	assert.Equal(t, 0, detail.AvatarType)
	assert.Empty(t, detail.ProductKey)
	assert.Empty(t, detail.ProductName)
	assert.Nil(t, detail.IsBuiltIn)
	assert.Nil(t, detail.Config)
	assert.Equal(t, int64(0), detail.CreatedAt)
	assert.Equal(t, int64(0), detail.UpdatedAt)
	assert.Empty(t, detail.CreatedBy)
	assert.Empty(t, detail.UpdatedBy)
	assert.Equal(t, int64(0), detail.PublishedAt)
	assert.Empty(t, detail.PublishedBy)
}

func TestDetailRes_WithDifferentStatus(t *testing.T) {
	t.Parallel()

	statuses := []cdaenum.Status{
		cdaenum.StatusUnpublished,
		cdaenum.StatusPublished,
	}

	for _, status := range statuses {
		detail := &DetailRes{
			Status: status,
		}
		assert.Equal(t, status, detail.Status)
	}
}

func TestDetailRes_WithDifferentAvatarType(t *testing.T) {
	t.Parallel()

	avatarTypes := []int{0, 1, 2, 3}

	for _, avatarType := range avatarTypes {
		detail := &DetailRes{
			AvatarType: avatarType,
		}
		assert.Equal(t, avatarType, detail.AvatarType)
	}
}

func TestDetailRes_WithNilIsBuiltIn(t *testing.T) {
	t.Parallel()

	detail := &DetailRes{
		IsBuiltIn: nil,
	}

	assert.Nil(t, detail.IsBuiltIn)
}

func TestDetailRes_WithNilConfig(t *testing.T) {
	t.Parallel()

	detail := &DetailRes{
		Config: nil,
	}

	assert.Nil(t, detail.Config)
}

func TestDetailRes_WithEmptyProfile(t *testing.T) {
	t.Parallel()

	profile := ""
	detail := &DetailRes{
		Profile: &profile,
	}

	assert.Equal(t, "", *detail.Profile)
}

func TestDetailRes_FullStructure(t *testing.T) {
	t.Parallel()

	// Test the complete structure can be created
	detail := &DetailRes{
		ID:          456,
		Name:        "Full Template",
		Key:         "full-tpl",
		Avatar:      "full.png",
		AvatarType:  1,
		Status:      cdaenum.StatusPublished,
		ProductKey:  "product-full",
		ProductName: "Full Product",
		Config: &daconfvalobj.Config{
			Input:  &daconfvalobj.Input{},
			Output: &daconfvalobj.Output{},
		},
		CreatedAt:   1111111111,
		UpdatedAt:   2222222222,
		CreatedBy:   "creator",
		UpdatedBy:   "updater",
		PublishedAt: 3333333333,
		PublishedBy: "publisher",
	}

	assert.NotNil(t, detail)
	assert.Equal(t, "Full Template", detail.Name)
	assert.Equal(t, "full-tpl", detail.Key)
}
