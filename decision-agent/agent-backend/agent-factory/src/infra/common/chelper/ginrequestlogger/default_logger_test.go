package ginrequestlogger

import (
	"sync"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httprequesthelper"
	"github.com/stretchr/testify/assert"
)

func TestInitDefaultRequestLogger_Success(t *testing.T) {
	// 不使用 t.Parallel(): 修改 singleton 全局状态
	// Reset the singleton before testing
	defaultRequestLogger = nil
	defaultRequestLoggerOnce = *(new(sync.Once))

	config := &httprequesthelper.Config{
		Enabled:             true,
		OutputMode:          httprequesthelper.OutputModeConsole,
		MaxBodySize:         1024,
		IncludeHeaders:      true,
		IncludeResponseBody: true,
	}

	err := InitDefaultRequestLogger(config)

	assert.NoError(t, err)
	assert.NotNil(t, GetDefaultRequestLogger())
}

func TestInitDefaultRequestLogger_Idempotent(t *testing.T) {
	// 不使用 t.Parallel(): 修改 singleton 全局状态
	// Reset the singleton before testing
	defaultRequestLogger = nil
	defaultRequestLoggerOnce = *(new(sync.Once))

	config1 := &httprequesthelper.Config{
		Enabled:             true,
		OutputMode:          httprequesthelper.OutputModeConsole,
		IncludeHeaders:      true,
		IncludeResponseBody: false,
	}

	err1 := InitDefaultRequestLogger(config1)
	assert.NoError(t, err1)

	// Second call should be ignored
	config2 := &httprequesthelper.Config{
		Enabled:             true,
		OutputMode:          httprequesthelper.OutputModeConsole,
		IncludeHeaders:      false,
		IncludeResponseBody: true,
	}

	err2 := InitDefaultRequestLogger(config2)
	assert.NoError(t, err2)

	// Logger should still exist (first config used)
	logger := GetDefaultRequestLogger()
	assert.NotNil(t, logger)
}

func TestGetDefaultRequestLogger_NotInitialized(t *testing.T) {
	// 不使用 t.Parallel(): 修改 singleton 全局状态
	// Reset the singleton
	defaultRequestLogger = nil
	defaultRequestLoggerOnce = *(new(sync.Once))

	logger := GetDefaultRequestLogger()
	assert.Nil(t, logger)
}

func TestInitDefaultRequestLogger_WithNilConfig(t *testing.T) {
	// 不使用 t.Parallel(): 修改 singleton 全局状态
	// Reset the singleton
	defaultRequestLogger = nil
	defaultRequestLoggerOnce = *(new(sync.Once))

	err := InitDefaultRequestLogger(nil)
	// Should not panic, behavior depends on NewRequestLogger implementation
	// Just verify it returns without panicking
	assert.NoError(t, err)
}

func TestInitDefaultRequestLogger_WithDefaultConfig(t *testing.T) {
	// 不使用 t.Parallel(): 修改 singleton 全局状态
	// Reset the singleton
	defaultRequestLogger = nil
	defaultRequestLoggerOnce = *(new(sync.Once))

	config := httprequesthelper.DefaultConfig()
	config.OutputMode = httprequesthelper.OutputModeConsole

	err := InitDefaultRequestLogger(config)

	assert.NoError(t, err)
	assert.NotNil(t, GetDefaultRequestLogger())
}

func TestInitDefaultRequestLogger_InvalidLogDir(t *testing.T) {
	// 不使用 t.Parallel(): 修改 singleton 全局状态
	// Reset the singleton before testing
	defaultRequestLogger = nil
	defaultRequestLoggerOnce = *(new(sync.Once))

	config := &httprequesthelper.Config{
		Enabled:    true,
		OutputMode: httprequesthelper.OutputModeFile,
		LogDir:     "/dev/null/invalid/subdir", // Can't create subdirectory under /dev/null
	}

	err := InitDefaultRequestLogger(config)
	assert.Error(t, err)
	assert.Nil(t, GetDefaultRequestLogger())
}
