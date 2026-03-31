package agentsvc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStreamingLoggerDebugEnv(t *testing.T, isDebug bool, logRootDir string) {
	t.Helper()

	const (
		svcName       = "UT_STREAM_LOG"
		debugEnvKey   = "UT_STREAM_LOG_DEBUG_MODE"
		svcEnvKey     = "SERVICE_NAME"
		logRootEnvKey = "AGENT_FACTORY_LOCAL_DEV_LOG_ROOT_DIR"
	)

	originalServiceName, hasServiceName := os.LookupEnv(svcEnvKey)
	originalDebug, hasDebug := os.LookupEnv(debugEnvKey)
	originalLogRoot, hasLogRoot := os.LookupEnv(logRootEnvKey)

	restore := func(k, v string, ok bool) {
		if ok {
			_ = os.Setenv(k, v)
			return
		}

		_ = os.Unsetenv(k)
	}

	t.Cleanup(func() {
		restore(svcEnvKey, originalServiceName, hasServiceName)
		restore(debugEnvKey, originalDebug, hasDebug)
		restore(logRootEnvKey, originalLogRoot, hasLogRoot)
		cenvhelper.InitEnvForTest()
	})

	_ = os.Setenv(svcEnvKey, svcName)
	if isDebug {
		_ = os.Setenv(debugEnvKey, "true")
	} else {
		_ = os.Setenv(debugEnvKey, "false")
	}

	if logRootDir != "" {
		_ = os.Setenv(logRootEnvKey, logRootDir)
	} else {
		_ = os.Unsetenv(logRootEnvKey)
	}

	cenvhelper.InitEnvForTest()
}

func TestNewStreamingResponseLogger_NonDebugMode(t *testing.T) {
	t.Parallel()

	setupStreamingLoggerDebugEnv(t, false, "")

	logger, err := NewStreamingResponseLogger("conv-123", ExecutorResponse)

	assert.NoError(t, err)
	assert.Nil(t, logger)
}

func TestNewStreamingResponseLogger_DebugMode_CreateSuccess(t *testing.T) {
	t.Parallel()

	logRoot := t.TempDir()
	setupStreamingLoggerDebugEnv(t, true, logRoot)

	logger, err := NewStreamingResponseLogger("conv-debug", ExecutorResponse)
	require.NoError(t, err)
	require.NotNil(t, logger)
	require.NotNil(t, logger.file)

	assert.Contains(t, logger.file.Name(), filepath.Join("streaming_responses", string(ExecutorResponse)))
	assert.Contains(t, logger.file.Name(), "conv-debug")

	logger.LogChunk([]byte("hello-debug"))
	logger.Complete()

	content, readErr := os.ReadFile(logger.file.Name())
	require.NoError(t, readErr)
	assert.True(t, strings.Contains(string(content), "Chunk 1"))
	assert.True(t, strings.Contains(string(content), "Stream completed"))
}

func TestNewStreamingResponseLogger_DebugMode_MkdirAllError(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	notDir := filepath.Join(baseDir, "root-file")
	require.NoError(t, os.WriteFile(notDir, []byte("x"), 0o644))
	setupStreamingLoggerDebugEnv(t, true, notDir)

	logger, err := NewStreamingResponseLogger("conv-err", ProcessedResponse)
	require.Error(t, err)
	assert.Nil(t, logger)
}

func TestStreamingResponseLogger_LogChunk_NilLogger(t *testing.T) {
	t.Parallel()

	var l *StreamingResponseLogger

	assert.NotPanics(t, func() {
		l.LogChunk([]byte("test chunk"))
	})
}

func TestStreamingResponseLogger_Complete_NilLogger(t *testing.T) {
	t.Parallel()

	var l *StreamingResponseLogger

	assert.NotPanics(t, func() {
		l.Complete()
	})
}

func TestStreamingResponseLogger_LogChunk_NilFile(t *testing.T) {
	t.Parallel()

	l := &StreamingResponseLogger{
		file: nil,
	}

	assert.NotPanics(t, func() {
		l.LogChunk([]byte("test chunk"))
	})
}

func TestStreamingResponseLogger_Complete_NilFile(t *testing.T) {
	t.Parallel()

	l := &StreamingResponseLogger{
		file: nil,
	}

	assert.NotPanics(t, func() {
		l.Complete()
	})
}

func TestStreamingResponseLogger_LogAndComplete_WithRealFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	f, err := os.CreateTemp(tmpDir, "test-*.log")
	if err != nil {
		t.Fatal(err)
	}

	l := &StreamingResponseLogger{
		file:           f,
		conversationID: "conv-test",
		logType:        ProcessedResponse,
	}

	l.LogChunk([]byte("hello"))
	l.LogChunk([]byte("world"))
	assert.Equal(t, 2, l.chunksCount)
	assert.Equal(t, 10, l.totalBytes)

	l.Complete()
}
