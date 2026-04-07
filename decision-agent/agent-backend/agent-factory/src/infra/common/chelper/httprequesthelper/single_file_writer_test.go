package httprequesthelper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T) *Config {
	t.Helper()
	dir := t.TempDir()

	return &Config{
		Enabled:              true,
		OutputMode:           OutputModeFile,
		LogDir:               dir,
		FileNamePattern:      "test.log",
		SingleFileMaxEntries: 5,
	}
}

// ==================== NewSingleFileWriter ====================

func TestNewSingleFileWriter_Success(t *testing.T) {
	cfg := testConfig(t)
	w, err := NewSingleFileWriter(cfg)
	require.NoError(t, err)
	require.NotNil(t, w)
	assert.Equal(t, 0, w.lineCount)
	assert.Equal(t, 5, w.maxEntries)
	_ = w.Close()
}

func TestNewSingleFileWriter_ExistingFile(t *testing.T) {
	cfg := testConfig(t)
	// Pre-create the single file with 3 lines
	singleDir := filepath.Join(cfg.LogDir, "single")
	require.NoError(t, os.MkdirAll(singleDir, 0o755))
	f, err := os.Create(filepath.Join(singleDir, "all_requests.log"))
	require.NoError(t, err)

	_, _ = f.WriteString("line1\nline2\nline3\n")
	_ = f.Close()

	w, err := NewSingleFileWriter(cfg)
	require.NoError(t, err)
	assert.Equal(t, 3, w.lineCount)
	_ = w.Close()
}

// ==================== Write ====================

func TestSingleFileWriter_Write(t *testing.T) {
	cfg := testConfig(t)
	w, err := NewSingleFileWriter(cfg)
	require.NoError(t, err)

	require.NoError(t, w.Write("hello\n"))
	assert.Equal(t, 1, w.lineCount)

	require.NoError(t, w.Write("world\n"))
	assert.Equal(t, 2, w.lineCount)

	_ = w.Close()

	// Verify file contents
	data, err := os.ReadFile(cfg.GetSingleFilePath())
	require.NoError(t, err)
	assert.Equal(t, "hello\nworld\n", string(data))
}

func TestSingleFileWriter_Write_TriggersTruncate(t *testing.T) {
	cfg := testConfig(t)
	cfg.SingleFileMaxEntries = 3 // maxEntries=3, threshold=100 by default
	w, err := NewSingleFileWriter(cfg)
	require.NoError(t, err)
	// Override threshold to trigger truncation sooner
	w.threshold = 2

	// Write 6 lines → 6 > 3+2 → triggers truncation, keeps last 3
	for i := 0; i < 6; i++ {
		require.NoError(t, w.Write("line\n"))
	}

	assert.Equal(t, 3, w.lineCount) // After truncation → 3
	_ = w.Close()

	// Verify file has 3 lines
	data, err := os.ReadFile(cfg.GetSingleFilePath())
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Equal(t, 3, len(lines))
}

// ==================== Close ====================

func TestSingleFileWriter_Close_NoFile(t *testing.T) {
	cfg := testConfig(t)
	w, err := NewSingleFileWriter(cfg)
	require.NoError(t, err)
	// Close without writing — file is nil
	assert.NoError(t, w.Close())
}

func TestSingleFileWriter_Close_WithFile(t *testing.T) {
	cfg := testConfig(t)
	w, err := NewSingleFileWriter(cfg)
	require.NoError(t, err)
	require.NoError(t, w.Write("data\n"))
	assert.NoError(t, w.Close())
}

// ==================== countLines ====================

func TestSingleFileWriter_CountLines_NoFile(t *testing.T) {
	cfg := testConfig(t)
	w := &SingleFileWriter{filePath: filepath.Join(cfg.LogDir, "nonexistent.log")}
	count, err := w.countLines()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSingleFileWriter_CountLines_ExistingFile(t *testing.T) {
	cfg := testConfig(t)
	fp := filepath.Join(cfg.LogDir, "test_count.log")
	require.NoError(t, os.WriteFile(fp, []byte("a\nb\nc\n"), 0o644))

	w := &SingleFileWriter{filePath: fp}
	count, err := w.countLines()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

// ==================== readAllLines ====================

func TestSingleFileWriter_ReadAllLines(t *testing.T) {
	cfg := testConfig(t)
	fp := filepath.Join(cfg.LogDir, "test_read.log")
	require.NoError(t, os.WriteFile(fp, []byte("line1\nline2\nline3\n"), 0o644))

	w := &SingleFileWriter{filePath: fp}
	lines, err := w.readAllLines()
	assert.NoError(t, err)
	assert.Equal(t, []string{"line1", "line2", "line3"}, lines)
}

// ==================== rewriteFile ====================

func TestSingleFileWriter_RewriteFile(t *testing.T) {
	cfg := testConfig(t)
	fp := filepath.Join(cfg.LogDir, "test_rewrite.log")

	w := &SingleFileWriter{filePath: fp}
	require.NoError(t, w.rewriteFile([]string{"new1", "new2"}))

	data, err := os.ReadFile(fp)
	require.NoError(t, err)
	assert.Equal(t, "new1\nnew2\n", string(data))
}

// ==================== GetDefaultLogger ====================

func TestGetDefaultLogger(t *testing.T) {
	l := GetDefaultLogger()
	assert.NotNil(t, l)
}
