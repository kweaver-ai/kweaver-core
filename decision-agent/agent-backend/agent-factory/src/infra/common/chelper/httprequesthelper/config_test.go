package httprequesthelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputMode_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode OutputMode
		want string
	}{
		{"file mode", OutputModeFile, "file"},
		{"console mode", OutputModeConsole, "console"},
		{"both mode", OutputModeBoth, "both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, string(tt.mode))
		})
	}
}

func TestDefaultConfig_ReturnsValidConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Equal(t, OutputModeFile, config.OutputMode)
	assert.Equal(t, "log/requests", config.LogDir)
	assert.Equal(t, "requests_2006-01-02.log", config.FileNamePattern)
	assert.False(t, config.PrettyJSON)
	assert.Equal(t, 10*1024, config.MaxBodySize)
	assert.True(t, config.IncludeHeaders)
	assert.True(t, config.IncludeResponseBody)
	assert.Equal(t, 0, config.SingleFileMaxEntries)
}

func TestConfig_GetLogFilePath_WithoutUserID(t *testing.T) {
	t.Parallel()

	config := &Config{
		LogDir:          "log/requests",
		FileNamePattern: "requests_2006-01-02.log",
	}

	path := config.GetLogFilePath("")
	assert.Contains(t, path, "log/requests")
	assert.Contains(t, path, "requests_")
	assert.Contains(t, path, ".log")
}

func TestConfig_GetLogFilePath_WithUserID(t *testing.T) {
	t.Parallel()

	config := &Config{
		LogDir:          "log/requests",
		FileNamePattern: "requests_2006-01-02.log",
	}

	path := config.GetLogFilePath("user123")
	assert.Contains(t, path, "log/requests")
	assert.Contains(t, path, "requests_")
	assert.Contains(t, path, "user123")
	assert.Contains(t, path, ".log")
}

func TestConfig_GetLogFilePath_CustomPattern(t *testing.T) {
	t.Parallel()

	config := &Config{
		LogDir:          "custom/dir",
		FileNamePattern: "app_%Y_%m_%d.log",
	}

	path := config.GetLogFilePath("")
	assert.Contains(t, path, "custom/dir")
	assert.Contains(t, path, "app_")
	assert.Contains(t, path, ".log")
}

func TestConfig_EnsureLogDir_ConsoleMode(t *testing.T) {
	t.Parallel()

	config := &Config{
		OutputMode: OutputModeConsole,
	}

	err := config.EnsureLogDir()
	assert.NoError(t, err)
}

func TestConfig_EnsureLogDir_FileMode(t *testing.T) {
	t.Parallel()

	config := &Config{
		OutputMode: OutputModeFile,
		LogDir:     "/tmp/test-log-dir-12345",
	}

	err := config.EnsureLogDir()
	assert.NoError(t, err)
}

func TestConfig_GetSingleFilePath(t *testing.T) {
	t.Parallel()

	config := &Config{
		LogDir: "log/requests",
	}

	path := config.GetSingleFilePath()
	assert.Equal(t, "log/requests/single/all_requests.log", path)
}

func TestConfig_GetSingleFilePath_CustomLogDir(t *testing.T) {
	t.Parallel()

	config := &Config{
		LogDir: "custom/logs",
	}

	path := config.GetSingleFilePath()
	assert.Equal(t, "custom/logs/single/all_requests.log", path)
}

func TestConfig_EnsureSingleFileDir(t *testing.T) {
	t.Parallel()

	config := &Config{
		LogDir: "/tmp/test-single-dir-12345",
	}

	err := config.EnsureSingleFileDir()
	assert.NoError(t, err)
}

func TestConfig_OutputModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode OutputMode
	}{
		{"file", OutputModeFile},
		{"console", OutputModeConsole},
		{"both", OutputModeBoth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := &Config{
				OutputMode: tt.mode,
			}
			assert.Equal(t, tt.mode, config.OutputMode)
		})
	}
}

func TestConfig_AllFieldsSettable(t *testing.T) {
	t.Parallel()

	config := &Config{
		Enabled:              false,
		OutputMode:           OutputModeBoth,
		LogDir:               "custom/log/dir",
		FileNamePattern:      "custom_%Y-%m-%d.log",
		PrettyJSON:           true,
		MaxBodySize:          2048,
		IncludeHeaders:       false,
		IncludeResponseBody:  false,
		SingleFileMaxEntries: 100,
	}

	assert.False(t, config.Enabled)
	assert.Equal(t, OutputModeBoth, config.OutputMode)
	assert.Equal(t, "custom/log/dir", config.LogDir)
	assert.Equal(t, "custom_%Y-%m-%d.log", config.FileNamePattern)
	assert.True(t, config.PrettyJSON)
	assert.Equal(t, 2048, config.MaxBodySize)
	assert.False(t, config.IncludeHeaders)
	assert.False(t, config.IncludeResponseBody)
	assert.Equal(t, 100, config.SingleFileMaxEntries)
}

func TestConfig_MaxBodySizeZero(t *testing.T) {
	t.Parallel()

	config := &Config{
		MaxBodySize: 0,
	}

	assert.Equal(t, 0, config.MaxBodySize)
}

func TestConfig_SingleFileMaxEntriesZero(t *testing.T) {
	t.Parallel()

	config := &Config{
		SingleFileMaxEntries: 0,
	}

	assert.Equal(t, 0, config.SingleFileMaxEntries)
}
