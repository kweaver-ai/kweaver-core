package pubedeo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestPublishedTpl_NewPublishedTpl(t *testing.T) {
	t.Parallel()

	tpl := &PublishedTpl{
		PublishedTplPo: dapo.PublishedTplPo{
			ID: 123,
		},
		ProductName: "Test Product",
	}

	assert.NotNil(t, tpl)
	assert.Equal(t, int64(123), tpl.ID)
	assert.Equal(t, "Test Product", tpl.ProductName)
}

func TestPublishedTpl_WithConfig(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.NewConfig()
	tpl := &PublishedTpl{
		Config: config,
	}

	assert.NotNil(t, tpl.Config)
	assert.NotNil(t, tpl)
}

func TestPublishedTpl_Empty(t *testing.T) {
	t.Parallel()

	tpl := &PublishedTpl{}

	assert.NotNil(t, tpl)
	assert.Nil(t, tpl.Config)
	assert.Empty(t, tpl.ProductName)
}
