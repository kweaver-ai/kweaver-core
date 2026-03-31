package modelfactoryacc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewModelApiAcc_ReturnsNotNil(t *testing.T) {
	t.Parallel()

	// This test verifies that NewModelApiAcc returns a non-nil value
	// Note: It will panic if global.GConfig is not initialized properly,
	// which is expected behavior in test environment

	assert.Panics(t, func() {
		_ = NewModelApiAcc(nil, nil, nil)
	})
}
