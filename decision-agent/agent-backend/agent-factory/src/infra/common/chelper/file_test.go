package chelper

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendWriteToFile(t *testing.T) {
	t.Parallel()

	t.Run("should not write file in non-local env", func(t *testing.T) {
		t.Parallel()
		filePath := filepath.Join(t.TempDir(), "non-local.txt")
		fileTestSetupLocalEnv(t, false)

		err := AppendWriteToFile(filePath, "hello")

		require.NoError(t, err)

		_, statErr := os.Stat(filePath)
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("should append content in local env", func(t *testing.T) {
		t.Parallel()
		filePath := filepath.Join(t.TempDir(), "local-append.txt")
		fileTestSetupLocalEnv(t, true)

		err := AppendWriteToFile(filePath, "line-1\n")
		require.NoError(t, err)

		content, readErr := os.ReadFile(filePath)
		require.NoError(t, readErr)
		assert.Equal(t, "line-1\n", string(content))
	})

	t.Run("should be thread-safe and keep all writes in local env", func(t *testing.T) {
		t.Parallel()
		filePath := filepath.Join(t.TempDir(), "local-concurrent.txt")
		fileTestSetupLocalEnv(t, true)

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				_ = AppendWriteToFile(filePath, "line\n")
			}()
		}

		wg.Wait()

		content, readErr := os.ReadFile(filePath)
		require.NoError(t, readErr)
		assert.Equal(t, 10, strings.Count(string(content), "\n"))
	})
}

func fileTestSetupLocalEnv(t *testing.T, isLocal bool) {
	t.Helper()

	const (
		svcName   = "UT_APPEND_FILE"
		localKey  = "UT_APPEND_FILE_LOCAL_DEV"
		sceneKey  = "UT_APPEND_FILE_RUN_SCENARIO"
		sceneVal  = "aaron_local_dev"
		svcEnvKey = "SERVICE_NAME"
	)

	originalServiceName, hasServiceName := os.LookupEnv(svcEnvKey)
	originalLocal, hasLocal := os.LookupEnv(localKey)
	originalScene, hasScene := os.LookupEnv(sceneKey)

	restore := func(k, v string, ok bool) {
		if ok {
			_ = os.Setenv(k, v)
			return
		}

		_ = os.Unsetenv(k)
	}

	t.Cleanup(func() {
		restore(svcEnvKey, originalServiceName, hasServiceName)
		restore(localKey, originalLocal, hasLocal)
		restore(sceneKey, originalScene, hasScene)
		cenvhelper.InitEnvForTest()
	})

	_ = os.Setenv(svcEnvKey, svcName)
	if isLocal {
		_ = os.Setenv(localKey, "true")
		_ = os.Setenv(sceneKey, sceneVal)
	} else {
		_ = os.Setenv(localKey, "false")
		_ = os.Unsetenv(sceneKey)
	}

	cenvhelper.InitEnvForTest()
}
