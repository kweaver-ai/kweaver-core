package dainject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDaConfSvc_PanicsWithoutGlobalConfig(t *testing.T) {
	t.Parallel()

	// The NewDaConfSvc function requires global.GConfig to be initialized
	// In test environment without global config, it will panic
	assert.Panics(t, func() {
		_ = NewDaConfSvc()
	})
}
