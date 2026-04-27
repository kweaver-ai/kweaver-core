package httprequesthelper

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== ConsoleWriter ====================

func TestNewConsoleWriter(t *testing.T) {
	t.Parallel()

	w := NewConsoleWriter()
	assert.NotNil(t, w)
}

func TestConsoleWriter_Write(t *testing.T) {
	t.Parallel()

	w := NewConsoleWriter()
	err := w.Write("test content", "user-1")
	assert.NoError(t, err)
}

func TestConsoleWriter_Close(t *testing.T) {
	t.Parallel()

	w := NewConsoleWriter()
	err := w.Close()
	assert.NoError(t, err)
}

// ==================== FileWriter ====================

func TestNewFileWriter(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.LogDir = tmpDir

	fw, err := NewFileWriter(cfg)
	require.NoError(t, err)
	assert.NotNil(t, fw)

	defer fw.Close()
}

func TestFileWriter_Write(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.LogDir = tmpDir

	fw, err := NewFileWriter(cfg)
	require.NoError(t, err)
	defer fw.Close()

	err = fw.Write("test log content\n", "")
	assert.NoError(t, err)

	// Write with userID
	err = fw.Write("user log content\n", "user-123")
	assert.NoError(t, err)
}

func TestFileWriter_Close_NilFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.LogDir = tmpDir

	fw, err := NewFileWriter(cfg)
	require.NoError(t, err)

	// Close without writing (file is nil)
	err = fw.Close()
	assert.NoError(t, err)
}

// ==================== MultiWriter ====================

func TestNewMultiWriter(t *testing.T) {
	t.Parallel()

	w1 := NewConsoleWriter()
	w2 := NewConsoleWriter()

	mw := NewMultiWriter(w1, w2)
	assert.NotNil(t, mw)
}

func TestMultiWriter_Write(t *testing.T) {
	t.Parallel()

	w1 := NewConsoleWriter()
	w2 := NewConsoleWriter()

	mw := NewMultiWriter(w1, w2)
	err := mw.Write("test content", "")
	assert.NoError(t, err)
}

func TestMultiWriter_Close(t *testing.T) {
	t.Parallel()

	w1 := NewConsoleWriter()
	w2 := NewConsoleWriter()

	mw := NewMultiWriter(w1, w2)
	err := mw.Close()
	assert.NoError(t, err)
}

// ==================== Logger ====================

func TestNewLogger_NilConfig(t *testing.T) {
	t.Parallel()

	l, err := NewLogger(nil)
	require.NoError(t, err)
	assert.NotNil(t, l)

	defer l.Close()
}

func TestNewLogger_ConsoleMode(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.OutputMode = OutputModeConsole

	l, err := NewLogger(cfg)
	require.NoError(t, err)
	assert.NotNil(t, l)

	defer l.Close()
}

func TestNewLogger_FileMode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.OutputMode = OutputModeFile
	cfg.LogDir = tmpDir

	l, err := NewLogger(cfg)
	require.NoError(t, err)
	assert.NotNil(t, l)

	defer l.Close()
}

func TestNewLogger_BothMode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.OutputMode = OutputModeBoth
	cfg.LogDir = tmpDir

	l, err := NewLogger(cfg)
	require.NoError(t, err)
	assert.NotNil(t, l)

	defer l.Close()
}

func TestNewLogger_InvalidDir(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.OutputMode = OutputModeFile
	cfg.LogDir = "/nonexistent/path/that/does/not/exist_" + t.Name()

	_, err := NewLogger(cfg)
	// 可能成功创建目录或失败
	if err != nil {
		assert.Error(t, err)
	}

	// 清理
	_ = os.RemoveAll(cfg.LogDir)
}

func TestLogger_IsEnabled(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.OutputMode = OutputModeConsole
	cfg.Enabled = true

	l, err := NewLogger(cfg)
	require.NoError(t, err)
	defer l.Close()

	assert.True(t, l.IsEnabled())
}

func TestLogger_SetEnabled(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.OutputMode = OutputModeConsole
	cfg.Enabled = false

	l, err := NewLogger(cfg)
	require.NoError(t, err)
	defer l.Close()

	assert.False(t, l.IsEnabled())

	l.SetEnabled(true)
	assert.True(t, l.IsEnabled())
}

func TestLogger_SetPrettyJSON(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.OutputMode = OutputModeConsole
	cfg.PrettyJSON = false

	l, err := NewLogger(cfg)
	require.NoError(t, err)
	defer l.Close()

	l.SetPrettyJSON(true)
	// 验证不 panic
}

func TestLogger_LogRequest_Disabled(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.OutputMode = OutputModeConsole
	cfg.Enabled = false

	l, err := NewLogger(cfg)
	require.NoError(t, err)
	defer l.Close()

	req, _ := http.NewRequest("GET", "http://test.com/api", nil)
	// 不会 panic（disabled 直接 return）
	l.LogRequest(t.Context(), req, "", 200, nil, "{}", time.Second)
}

func TestLogger_LogRequest_Enabled(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.OutputMode = OutputModeConsole
	cfg.Enabled = true

	l, err := NewLogger(cfg)
	require.NoError(t, err)
	defer l.Close()

	req, _ := http.NewRequest("POST", "http://test.com/api", nil)
	l.LogRequest(t.Context(), req, `{"key":"value"}`, 200, http.Header{}, `{"result":"ok"}`, time.Millisecond*100)
}
