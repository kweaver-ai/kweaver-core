package ginrequestlogger

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/httprequesthelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequestLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *httprequesthelper.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &httprequesthelper.Config{
				Enabled:    true,
				OutputMode: httprequesthelper.OutputModeConsole,
			},
			wantErr: false,
		},
		{
			name:    "nil config uses default",
			config:  nil,
			wantErr: false,
		},
		{
			name: "disabled logger",
			config: &httprequesthelper.Config{
				Enabled:    false,
				OutputMode: httprequesthelper.OutputModeConsole,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, err := NewRequestLogger(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestRequestLogger_Close(t *testing.T) {
	t.Parallel()

	logger, err := NewRequestLogger(&httprequesthelper.Config{
		Enabled:    true,
		OutputMode: httprequesthelper.OutputModeConsole,
	})
	require.NoError(t, err)
	require.NotNil(t, logger)

	err = logger.Close()
	assert.NoError(t, err)
}

func TestRequestLogger_Close_Idempotent(t *testing.T) {
	t.Parallel()

	logger, err := NewRequestLogger(&httprequesthelper.Config{
		Enabled:    true,
		OutputMode: httprequesthelper.OutputModeConsole,
	})
	require.NoError(t, err)

	// Close multiple times should not panic
	err = logger.Close()
	require.NoError(t, err)

	err = logger.Close()
	assert.NoError(t, err)
}

func TestRequestLogger_Close_NilLogger(t *testing.T) {
	t.Parallel()

	// This test verifies that closing a nil logger causes a panic
	var logger *RequestLogger

	assert.Panics(t, func() {
		logger.Close()
	})
}

func TestNewRequestLogger_WithMaxBodySize(t *testing.T) {
	t.Parallel()

	config := &httprequesthelper.Config{
		Enabled:     true,
		OutputMode:  httprequesthelper.OutputModeConsole,
		MaxBodySize: 2048,
	}

	logger, err := NewRequestLogger(config)
	require.NoError(t, err)
	assert.NotNil(t, logger)

	err = logger.Close()
	assert.NoError(t, err)
}

func TestNewRequestLogger_WithAllOptions(t *testing.T) {
	t.Parallel()

	config := &httprequesthelper.Config{
		Enabled:             true,
		OutputMode:          httprequesthelper.OutputModeConsole,
		MaxBodySize:         4096,
		IncludeHeaders:      true,
		IncludeResponseBody: true,
	}

	logger, err := NewRequestLogger(config)
	require.NoError(t, err)
	assert.NotNil(t, logger)

	err = logger.Close()
	assert.NoError(t, err)
}

func TestNewRequestLogger_InvalidLogDir(t *testing.T) {
	t.Parallel()

	// Test error path when file output mode is used with invalid log directory
	config := &httprequesthelper.Config{
		Enabled:    true,
		OutputMode: httprequesthelper.OutputModeFile,
		// Use an invalid path that cannot be created
		// In Unix-like systems, a path containing null bytes is invalid
		// Use a path that's likely to fail (e.g., path to a location we can't write to)
		LogDir: "/dev/null/invalid/subdir", // Can't create subdirectory under /dev/null
	}

	logger, err := NewRequestLogger(config)
	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestNewRequestLogger_InvalidLogDirBothMode(t *testing.T) {
	t.Parallel()

	// Test error path when both output mode is used with invalid log directory
	config := &httprequesthelper.Config{
		Enabled:    true,
		OutputMode: httprequesthelper.OutputModeBoth,
		LogDir:     "/dev/null/invalid/subdir", // Can't create subdirectory under /dev/null
	}

	logger, err := NewRequestLogger(config)
	assert.Error(t, err)
	assert.Nil(t, logger)
}
