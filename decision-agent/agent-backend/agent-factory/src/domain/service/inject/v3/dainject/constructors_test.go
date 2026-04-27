package dainject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV3Inject_Constructors(t *testing.T) {
	t.Parallel()

	initV3InjectGlobalConfig(t)
	resetV3InjectSingletons()

	t.Run("NewBizDomainSvc", func(t *testing.T) {
		t.Parallel()

		first := NewBizDomainSvc()
		second := NewBizDomainSvc()

		assert.NotNil(t, first)
		assert.Same(t, first, second)
	})

	t.Run("NewPermissionSvc", func(t *testing.T) {
		t.Parallel()

		first := NewPermissionSvc()
		second := NewPermissionSvc()

		assert.NotNil(t, first)
		assert.Same(t, first, second)
	})

	t.Run("NewPublishedSvc", func(t *testing.T) {
		t.Parallel()

		first := NewPublishedSvc()
		second := NewPublishedSvc()

		assert.NotNil(t, first)
		assert.Same(t, first, second)
	})

	t.Run("NewReleaseSvc", func(t *testing.T) {
		t.Parallel()

		first := NewReleaseSvc()
		second := NewReleaseSvc()

		assert.NotNil(t, first)
		assert.Same(t, first, second)
	})

	t.Run("NewPersonalSpaceSvc", func(t *testing.T) {
		t.Parallel()

		first := NewPersonalSpaceSvc()
		second := NewPersonalSpaceSvc()

		assert.NotNil(t, first)
		assert.Same(t, first, second)
	})

	t.Run("NewDaTplSvc", func(t *testing.T) {
		t.Parallel()

		first := NewDaTplSvc()
		second := NewDaTplSvc()

		assert.NotNil(t, first)
		assert.Same(t, first, second)
	})

	t.Run("NewAgentInOutSvc", func(t *testing.T) {
		t.Parallel()

		first := NewAgentInOutSvc()
		second := NewAgentInOutSvc()

		assert.NotNil(t, first)
		assert.Same(t, first, second)
	})
}
