package dainject

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver"
	"github.com/stretchr/testify/assert"
)

func TestNewOtherSvc_CreatesService(t *testing.T) {
	t.Parallel()

	svc := NewOtherSvc()

	assert.NotNil(t, svc)
	assert.Implements(t, (*iv3portdriver.IOtherSvc)(nil), svc)
}

func TestNewOtherSvc_Singleton(t *testing.T) {
	t.Parallel()

	svc1 := NewOtherSvc()
	svc2 := NewOtherSvc()

	// Should return the same instance
	assert.Same(t, svc1, svc2)
}
