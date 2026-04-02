package global

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalVariables(t *testing.T) {
	t.Parallel()

	t.Run("global variables are declared", func(t *testing.T) {
		t.Parallel()

		_ = GConfig
		_ = GDB
	})
}

func TestGConfig_NilByDefault(t *testing.T) {
	t.Parallel()

	t.Run("GConfig is nil before initialization", func(t *testing.T) {
		t.Parallel()

		originalConfig := GConfig
		defer func() {
			GConfig = originalConfig
		}()

		GConfig = nil
		assert.Nil(t, GConfig)
	})
}

func TestGDB_NilByDefault(t *testing.T) {
	t.Parallel()

	t.Run("GDB is nil before initialization", func(t *testing.T) {
		t.Parallel()

		originalDB := GDB
		defer func() {
			GDB = originalDB
		}()

		GDB = nil
		assert.Nil(t, GDB)
	})
}
