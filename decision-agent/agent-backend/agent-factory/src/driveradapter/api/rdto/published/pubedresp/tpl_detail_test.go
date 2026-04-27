package pubedresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDetailRes(t *testing.T) {
	t.Parallel()

	res := NewDetailRes()
	assert.NotNil(t, res)
	assert.Zero(t, res.ID)
	assert.Zero(t, res.TplID)
	assert.Empty(t, res.Name)
	assert.Empty(t, res.Key)
	assert.Nil(t, res.Config)
}

func TestDetailRes_LoadFromEo(t *testing.T) {
	t.Parallel()

	t.Run("valid eo", func(t *testing.T) {
		t.Parallel()

		res := NewDetailRes()
		profile := "test profile"
		builtIn := cdaenum.BuiltInYes
		eo := &pubedeo.PublishedTpl{
			PublishedTplPo: dapo.PublishedTplPo{
				ID:         123,
				TplID:      456,
				Name:       "Test Template",
				Profile:    &profile,
				Key:        "test-tpl",
				Avatar:     "🤖",
				AvatarType: cdaenum.AvatarTypeBuiltIn,
				IsBuiltIn:  &builtIn,
				Config:     "",
			},
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		err := res.LoadFromEo(eo)
		require.NoError(t, err)
		assert.Equal(t, int64(123), res.ID)
		assert.Equal(t, int64(456), res.TplID)
		assert.Equal(t, "Test Template", res.Name)
		assert.Equal(t, "test-tpl", res.Key)
	})

	t.Run("nil profile", func(t *testing.T) {
		t.Parallel()

		res := NewDetailRes()
		eo := &pubedeo.PublishedTpl{
			PublishedTplPo: dapo.PublishedTplPo{
				ID:      123,
				TplID:   456,
				Name:    "Test Template",
				Profile: nil,
				Key:     "test-tpl",
			},
		}

		err := res.LoadFromEo(eo)
		require.NoError(t, err)
		assert.Nil(t, res.Profile)
	})
}

func TestDetailRes_Fields(t *testing.T) {
	t.Parallel()

	res := &DetailRes{
		ID:          100,
		TplID:       200,
		Name:        "Detail Test",
		Key:         "detail-key",
		Avatar:      "avatar.png",
		AvatarType:  1,
		ProductKey:  "product-key",
		ProductName: "Product Name",
		PublishedAt: 1234567890,
		PublishedBy: "user-1",
	}

	assert.Equal(t, int64(100), res.ID)
	assert.Equal(t, int64(200), res.TplID)
	assert.Equal(t, "Detail Test", res.Name)
	assert.Equal(t, "detail-key", res.Key)
	assert.Equal(t, "avatar.png", res.Avatar)
}
