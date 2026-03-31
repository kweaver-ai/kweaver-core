package helpers

import (
	"os"
	"testing"
)

func TestEnvConstants(t *testing.T) {
	t.Parallel()

	t.Run("EnvPrefix constant", func(t *testing.T) {
		t.Parallel()

		expected := "AGENT_FACTORY_"
		if EnvPrefix != expected {
			t.Errorf("Expected EnvPrefix to be '%s', got '%s'", expected, EnvPrefix)
		}
	})

	t.Run("EnvIsLocalDev constant", func(t *testing.T) {
		t.Parallel()

		expected := "AGENT_FACTORY_LOCAL_DEV"
		if EnvIsLocalDev != expected {
			t.Errorf("Expected EnvIsLocalDev to be '%s', got '%s'", expected, EnvIsLocalDev)
		}
	})

	t.Run("constants are not empty", func(t *testing.T) {
		t.Parallel()

		constants := []string{
			EnvPrefix,
			EnvIsLocalDev,
			isDebugMode,
			isSQLPrint,
			projPath,
			skipOauthVerify,
		}

		for _, constant := range constants {
			if constant == "" {
				t.Error("Expected constant to not be empty")
			}
		}
	})
}

func TestIsLocalDev(t *testing.T) {
	// 不使用 t.Parallel(): 子测试修改共享环境变量和 mockIsLocalDev
	// Save original env value
	originalValue := os.Getenv(EnvIsLocalDev)

	// Clean up after test
	defer func() {
		if originalValue != "" {
			os.Setenv(EnvIsLocalDev, originalValue)
		} else {
			os.Unsetenv(EnvIsLocalDev)
		}
		// Reset mock
		mockIsLocalDev = false
	}()

	t.Run("env var true", func(t *testing.T) {
		os.Setenv(EnvIsLocalDev, "true")
		defer os.Unsetenv(EnvIsLocalDev)

		if !IsLocalDev() {
			t.Error("Expected IsLocalDev to return true")
		}
	})

	t.Run("env var false", func(t *testing.T) {
		os.Setenv(EnvIsLocalDev, "false")
		defer os.Unsetenv(EnvIsLocalDev)

		if IsLocalDev() {
			t.Error("Expected IsLocalDev to return false")
		}
	})

	t.Run("env var not set", func(t *testing.T) {
		os.Unsetenv(EnvIsLocalDev)

		if IsLocalDev() {
			t.Error("Expected IsLocalDev to return false when env not set")
		}
	})

	t.Run("mock is local dev", func(t *testing.T) {
		os.Unsetenv(EnvIsLocalDev)
		SetIsLocalDev()

		defer func() { mockIsLocalDev = false }()

		if !IsLocalDev() {
			t.Error("Expected IsLocalDev to return true when mock is set")
		}
	})
}

func TestIsAaronLocalDev(t *testing.T) {
	// 不使用 t.Parallel(): 子测试修改共享环境变量
	envVar := EnvIsLocalDev + "_AARON"
	originalValue := os.Getenv(envVar)

	defer func() {
		if originalValue != "" {
			os.Setenv(envVar, originalValue)
		} else {
			os.Unsetenv(envVar)
		}
	}()

	t.Run("env var true", func(t *testing.T) {
		os.Setenv(envVar, "true")
		defer os.Unsetenv(envVar)

		if !IsAaronLocalDev() {
			t.Error("Expected IsAaronLocalDev to return true")
		}
	})

	t.Run("env var false", func(t *testing.T) {
		os.Setenv(envVar, "false")
		defer os.Unsetenv(envVar)

		if IsAaronLocalDev() {
			t.Error("Expected IsAaronLocalDev to return false")
		}
	})

	t.Run("env var not set", func(t *testing.T) {
		os.Unsetenv(envVar)

		if IsAaronLocalDev() {
			t.Error("Expected IsAaronLocalDev to return false when env not set")
		}
	})
}

func TestIsDebugMode(t *testing.T) {
	// 不使用 t.Parallel(): 子测试修改共享环境变量
	originalValue := os.Getenv(isDebugMode)
	defer func() {
		if originalValue != "" {
			os.Setenv(isDebugMode, originalValue)
		} else {
			os.Unsetenv(isDebugMode)
		}
	}()

	t.Run("env var true", func(t *testing.T) {
		os.Setenv(isDebugMode, "true")
		defer os.Unsetenv(isDebugMode)

		if !IsDebugMode() {
			t.Error("Expected IsDebugMode to return true")
		}
	})

	t.Run("env var false", func(t *testing.T) {
		os.Setenv(isDebugMode, "false")
		defer os.Unsetenv(isDebugMode)

		if IsDebugMode() {
			t.Error("Expected IsDebugMode to return false")
		}
	})

	t.Run("env var not set", func(t *testing.T) {
		os.Unsetenv(isDebugMode)

		if IsDebugMode() {
			t.Error("Expected IsDebugMode to return false when env not set")
		}
	})
}

func TestIsOprLogShowLogForDebug(t *testing.T) {
	t.Parallel()

	t.Run("test function exists", func(t *testing.T) {
		t.Parallel()
		// This function calls IsDebugMode internally
		result := IsOprLogShowLogForDebug()
		// Result depends on environment
		_ = result
	})
}

func TestIsSQLPrint(t *testing.T) {
	// 不使用 t.Parallel(): 子测试修改共享环境变量
	originalValue := os.Getenv(isSQLPrint)
	defer func() {
		if originalValue != "" {
			os.Setenv(isSQLPrint, originalValue)
		} else {
			os.Unsetenv(isSQLPrint)
		}
	}()

	t.Run("env var true", func(t *testing.T) {
		os.Setenv(isSQLPrint, "true")
		defer os.Unsetenv(isSQLPrint)

		if !IsSQLPrint() {
			t.Error("Expected IsSQLPrint to return true")
		}
	})

	t.Run("env var false", func(t *testing.T) {
		os.Setenv(isSQLPrint, "false")
		defer os.Unsetenv(isSQLPrint)

		if IsSQLPrint() {
			t.Error("Expected IsSQLPrint to return false")
		}
	})

	t.Run("env var not set", func(t *testing.T) {
		os.Unsetenv(isSQLPrint)

		if IsSQLPrint() {
			t.Error("Expected IsSQLPrint to return false when env not set")
		}
	})
}

func TestProjectPathByEnv(t *testing.T) {
	// 不使用 t.Parallel(): 子测试修改共享环境变量
	originalValue := os.Getenv(projPath)
	defer func() {
		if originalValue != "" {
			os.Setenv(projPath, originalValue)
		} else {
			os.Unsetenv(projPath)
		}
	}()

	t.Run("env var set", func(t *testing.T) {
		expectedPath := "/custom/path"
		os.Setenv(projPath, expectedPath)

		defer os.Unsetenv(projPath)

		result := ProjectPathByEnv()
		if result != expectedPath {
			t.Errorf("Expected '%s', got '%s'", expectedPath, result)
		}
	})

	t.Run("env var not set", func(t *testing.T) {
		os.Unsetenv(projPath)

		result := ProjectPathByEnv()
		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})
}

func TestIsSkipOauthVerify(t *testing.T) {
	// 不使用 t.Parallel(): 子测试修改共享环境变量
	originalValue := os.Getenv(skipOauthVerify)
	defer func() {
		if originalValue != "" {
			os.Setenv(skipOauthVerify, originalValue)
		} else {
			os.Unsetenv(skipOauthVerify)
		}
	}()

	t.Run("env var true", func(t *testing.T) {
		os.Setenv(skipOauthVerify, "true")
		defer os.Unsetenv(skipOauthVerify)

		if !IsSkipOauthVerify() {
			t.Error("Expected IsSkipOauthVerify to return true")
		}
	})

	t.Run("env var false", func(t *testing.T) {
		os.Setenv(skipOauthVerify, "false")
		defer os.Unsetenv(skipOauthVerify)

		if IsSkipOauthVerify() {
			t.Error("Expected IsSkipOauthVerify to return false")
		}
	})

	t.Run("env var not set", func(t *testing.T) {
		os.Unsetenv(skipOauthVerify)

		if IsSkipOauthVerify() {
			t.Error("Expected IsSkipOauthVerify to return false when env not set")
		}
	})
}
