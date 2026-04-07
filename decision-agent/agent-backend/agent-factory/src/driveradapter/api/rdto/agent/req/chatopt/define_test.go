package chatopt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChatOption_StructFields(t *testing.T) {
	t.Parallel()

	opt := ChatOption{
		EnableDependencyCache:        true,
		IsNeedHistory:                true,
		IsNeedDocRetrivalPostProcess: true,
		IsNeedProgress:               true,
	}

	assert.True(t, opt.EnableDependencyCache)
	assert.True(t, opt.IsNeedHistory)
	assert.True(t, opt.IsNeedDocRetrivalPostProcess)
	assert.True(t, opt.IsNeedProgress)
}

func TestChatOption_Empty(t *testing.T) {
	t.Parallel()

	opt := ChatOption{}

	assert.False(t, opt.EnableDependencyCache)
	assert.False(t, opt.IsNeedHistory)
	assert.False(t, opt.IsNeedDocRetrivalPostProcess)
	assert.False(t, opt.IsNeedProgress)
}

func TestChatOption_Check(t *testing.T) {
	t.Parallel()

	t.Run("with normal mode and cache enabled", func(t *testing.T) {
		t.Parallel()

		opt := ChatOption{
			EnableDependencyCache: true,
		}

		err := opt.Check(false)
		assert.NoError(t, err)
	})

	t.Run("with debug mode and cache enabled", func(t *testing.T) {
		t.Parallel()

		opt := ChatOption{
			EnableDependencyCache: true,
		}

		err := opt.Check(true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "enable_dependency_cache must be false")
	})

	t.Run("with debug mode and cache disabled", func(t *testing.T) {
		t.Parallel()

		opt := ChatOption{
			EnableDependencyCache: false,
		}

		err := opt.Check(true)
		assert.NoError(t, err)
	})

	t.Run("with normal mode and cache disabled", func(t *testing.T) {
		t.Parallel()

		opt := ChatOption{
			EnableDependencyCache: false,
		}

		err := opt.Check(false)
		assert.NoError(t, err)
	})
}

func TestChatOption_WithAllFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cache       bool
		history     bool
		docProcess  bool
		progress    bool
		isDebugMode bool
		expectError bool
	}{
		{
			name:        "all flags enabled in normal mode",
			cache:       true,
			history:     true,
			docProcess:  true,
			progress:    true,
			isDebugMode: false,
			expectError: false,
		},
		{
			name:        "cache enabled in debug mode",
			cache:       true,
			history:     false,
			docProcess:  false,
			progress:    false,
			isDebugMode: true,
			expectError: true,
		},
		{
			name:        "cache disabled in debug mode",
			cache:       false,
			history:     true,
			docProcess:  true,
			progress:    true,
			isDebugMode: true,
			expectError: false,
		},
		{
			name:        "all flags disabled",
			cache:       false,
			history:     false,
			docProcess:  false,
			progress:    false,
			isDebugMode: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opt := ChatOption{
				EnableDependencyCache:        tt.cache,
				IsNeedHistory:                tt.history,
				IsNeedDocRetrivalPostProcess: tt.docProcess,
				IsNeedProgress:               tt.progress,
			}

			err := opt.Check(tt.isDebugMode)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChatOption_WithHistoryEnabled(t *testing.T) {
	t.Parallel()

	opt := ChatOption{
		IsNeedHistory: true,
	}

	assert.True(t, opt.IsNeedHistory)
	err := opt.Check(false)
	assert.NoError(t, err)
}

func TestChatOption_WithDocRetrivalPostProcess(t *testing.T) {
	t.Parallel()

	opt := ChatOption{
		IsNeedDocRetrivalPostProcess: true,
	}

	assert.True(t, opt.IsNeedDocRetrivalPostProcess)
	err := opt.Check(false)
	assert.NoError(t, err)
}

func TestChatOption_WithProgressEnabled(t *testing.T) {
	t.Parallel()

	opt := ChatOption{
		IsNeedProgress: true,
	}

	assert.True(t, opt.IsNeedProgress)
	err := opt.Check(false)
	assert.NoError(t, err)
}

func TestChatOption_Combinations(t *testing.T) {
	t.Parallel()

	combinations := []struct {
		name  string
		setup func() ChatOption
	}{
		{
			name: "history and progress",
			setup: func() ChatOption {
				return ChatOption{
					IsNeedHistory:  true,
					IsNeedProgress: true,
				}
			},
		},
		{
			name: "history and doc process",
			setup: func() ChatOption {
				return ChatOption{
					IsNeedHistory:                true,
					IsNeedDocRetrivalPostProcess: true,
				}
			},
		},
		{
			name: "doc process and progress",
			setup: func() ChatOption {
				return ChatOption{
					IsNeedDocRetrivalPostProcess: true,
					IsNeedProgress:               true,
				}
			},
		},
		{
			name: "all three",
			setup: func() ChatOption {
				return ChatOption{
					IsNeedHistory:                true,
					IsNeedDocRetrivalPostProcess: true,
					IsNeedProgress:               true,
				}
			},
		},
	}

	for _, combo := range combinations {
		t.Run(combo.name, func(t *testing.T) {
			t.Parallel()

			opt := combo.setup()
			err := opt.Check(false)
			assert.NoError(t, err)
		})
	}
}
