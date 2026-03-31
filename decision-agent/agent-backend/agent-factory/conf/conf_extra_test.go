package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSandboxPlatformConf_Struct(t *testing.T) {
	t.Parallel()

	t.Run("create SandboxPlatformConf struct", func(t *testing.T) {
		t.Parallel()

		sp := &SandboxPlatformConf{}

		assert.NotNil(t, sp)
	})
}

func TestDocsetConf_Struct(t *testing.T) {
	t.Parallel()

	t.Run("create DocsetConf struct", func(t *testing.T) {
		t.Parallel()

		dc := &DocsetConf{}

		assert.NotNil(t, dc)
	})
}

func TestUniqueryConf_Struct(t *testing.T) {
	t.Parallel()

	t.Run("create UniqueryConf struct", func(t *testing.T) {
		t.Parallel()

		uc := &UniqueryConf{}

		assert.NotNil(t, uc)
	})
}

func TestAgentExecutorConf_Struct(t *testing.T) {
	t.Parallel()

	t.Run("create AgentExecutorConf struct", func(t *testing.T) {
		t.Parallel()

		ae := &AgentExecutorConf{}

		assert.NotNil(t, ae)
	})
}

func TestConfig_Struct(t *testing.T) {
	t.Parallel()

	t.Run("create Config struct", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		assert.NotNil(t, cfg)
	})
}
