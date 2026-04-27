package cenvhelper

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// set SERVICE_NAME
	os.Setenv("SERVICE_NAME", "mock_svc_name")

	initEnv()

	// 运行测试
	code := m.Run()

	// 清理所有测试环境变量
	cleanupTestEnv()

	// 退出
	os.Exit(code)
}

func cleanupTestEnv() {
	// 清理测试环境变量
	os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")
	os.Unsetenv("MOCK_SVC_NAME_DEBUG_MODE")
	os.Unsetenv("MOCK_SVC_NAME_RUN_SCENARIO")
	os.Unsetenv("aaron_local_dev")
}

func TestIsLocalDev(t *testing.T) {
	t.Parallel()

	// 在测试开始前清理环境变量
	cleanupTestEnv()
	defer cleanupTestEnv()

	// 测试用例1: 无运行场景参数，环境变量为 true
	t.Run("no scenarios and env var is true", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_LOCAL_DEV", "true")

		defer os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")

		if !IsLocalDev() {
			t.Error("IsLocalDev() should return true when env var is 'true'")
		}
	})

	// 测试用例2: 无运行场景参数，环境变量为 false
	t.Run("no scenarios and env var is false", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_LOCAL_DEV", "false")

		defer os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")

		if IsLocalDev() {
			t.Error("IsLocalDev() should return false when env var is 'false'")
		}
	})

	// 测试用例3: 无运行场景参数，环境变量未设置
	t.Run("no scenarios and env var not set", func(t *testing.T) {
		t.Parallel()
		os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")

		if IsLocalDev() {
			t.Error("IsLocalDev() should return false when env var is not set")
		}
	})

	// 测试用例4: 有运行场景参数，但 LOCAL_DEV 为 false
	t.Run("with scenarios but LOCAL_DEV is false", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_LOCAL_DEV", "false")

		defer os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")

		if IsLocalDev(RunScenario_Aaron_Local_Dev) {
			t.Error("IsLocalDev() should return false when LOCAL_DEV is 'false'")
		}
	})

	// 测试用例5: 有运行场景参数，LOCAL_DEV 为 true，但 RUN_SCENARIO 未设置
	t.Run("with scenarios, LOCAL_DEV true, but RUN_SCENARIO not set", func(t *testing.T) {
		os.Setenv("MOCK_SVC_NAME_LOCAL_DEV", "true")
		os.Unsetenv("MOCK_SVC_NAME_RUN_SCENARIO") // 确保未设置

		defer func() {
			os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")
		}()

		if IsLocalDev(RunScenario_Aaron_Local_Dev) {
			t.Error("IsLocalDev() should return false when RUN_SCENARIO env is not set")
		}
	})

	// 测试用例6: 有运行场景参数，LOCAL_DEV 为 true，RUN_SCENARIO 匹配
	t.Run("with scenarios, LOCAL_DEV true, RUN_SCENARIO matches", func(t *testing.T) {
		os.Setenv("MOCK_SVC_NAME_LOCAL_DEV", "true")
		os.Setenv("MOCK_SVC_NAME_RUN_SCENARIO", "aaron_local_dev")

		defer func() {
			os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")
			os.Unsetenv("MOCK_SVC_NAME_RUN_SCENARIO")
		}()

		if !IsLocalDev(RunScenario_Aaron_Local_Dev) {
			t.Error("IsLocalDev() should return true when RUN_SCENARIO matches")
		}
	})

	// 测试用例7: 完整的成功场景
	t.Run("complete success scenario", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_LOCAL_DEV", "true")
		os.Setenv("MOCK_SVC_NAME_RUN_SCENARIO", "aaron_local_dev")

		defer func() {
			os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")
			os.Unsetenv("MOCK_SVC_NAME_RUN_SCENARIO")
		}()

		if !IsLocalDev(RunScenario_Aaron_Local_Dev) {
			t.Error("IsLocalDev() should return true when all conditions are met")
		}
	})

	// 测试用例8: 运行场景不匹配
	t.Run("scenario does not match", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_LOCAL_DEV", "true")
		os.Setenv("MOCK_SVC_NAME_RUN_SCENARIO", "other_scenario")

		defer func() {
			os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")
			os.Unsetenv("MOCK_SVC_NAME_RUN_SCENARIO")
		}()

		if IsLocalDev(RunScenario_Aaron_Local_Dev) {
			t.Error("IsLocalDev() should return false when scenario does not match")
		}
	})
}

func TestIsAaronLocalDev(t *testing.T) {
	t.Parallel()

	// 在测试开始前清理环境变量
	cleanupTestEnv()
	defer cleanupTestEnv()

	// 测试用例1: Aaron 本地开发环境开启
	t.Run("aaron local dev enabled", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_LOCAL_DEV", "true")
		os.Setenv("MOCK_SVC_NAME_RUN_SCENARIO", "aaron_local_dev")

		defer func() {
			os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")
			os.Unsetenv("MOCK_SVC_NAME_RUN_SCENARIO")
		}()

		if !IsAaronLocalDev() {
			t.Error("IsAaronLocalDev() should return true when aaron local dev is enabled")
		}
	})

	// 测试用例2: Aaron 本地开发环境关闭
	t.Run("aaron local dev disabled", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_LOCAL_DEV", "false")

		defer os.Unsetenv("MOCK_SVC_NAME_LOCAL_DEV")

		if IsAaronLocalDev() {
			t.Error("IsAaronLocalDev() should return false when local dev is disabled")
		}
	})
}

func TestIsDebugMode(t *testing.T) {
	t.Parallel()

	// 在测试开始前清理环境变量
	cleanupTestEnv()
	defer cleanupTestEnv()

	// 测试用例1: debug 模式开启
	t.Run("debug mode is true", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_DEBUG_MODE", "true")

		defer os.Unsetenv("MOCK_SVC_NAME_DEBUG_MODE")

		if !IsDebugMode() {
			t.Error("IsDebugMode() should return true when env var is 'true'")
		}
	})

	// 测试用例2: debug 模式关闭
	t.Run("debug mode is false", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_DEBUG_MODE", "false")

		defer os.Unsetenv("MOCK_SVC_NAME_DEBUG_MODE")

		if IsDebugMode() {
			t.Error("IsDebugMode() should return false when env var is 'false'")
		}
	})

	// 测试用例3: debug 模式未设置
	t.Run("debug mode not set", func(t *testing.T) {
		t.Parallel()
		os.Unsetenv("MOCK_SVC_NAME_DEBUG_MODE")

		if IsDebugMode() {
			t.Error("IsDebugMode() should return false when env var is not set")
		}
	})
}

func TestIsSQLPrint(t *testing.T) {
	t.Parallel()

	cleanupTestEnv()
	defer cleanupTestEnv()

	t.Run("sql print is true", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_SQL_PRINT", "true")

		defer os.Unsetenv("MOCK_SVC_NAME_SQL_PRINT")

		if !IsSQLPrint() {
			t.Error("IsSQLPrint() should return true when env var is 'true'")
		}
	})

	t.Run("sql print is false", func(t *testing.T) {
		t.Parallel()
		os.Setenv("MOCK_SVC_NAME_SQL_PRINT", "false")

		defer os.Unsetenv("MOCK_SVC_NAME_SQL_PRINT")

		if IsSQLPrint() {
			t.Error("IsSQLPrint() should return false when env var is 'false'")
		}
	})

	t.Run("sql print not set", func(t *testing.T) {
		t.Parallel()
		os.Unsetenv("MOCK_SVC_NAME_SQL_PRINT")

		if IsSQLPrint() {
			t.Error("IsSQLPrint() should return false when env var is not set")
		}
	})
}

func TestProjectPathByEnv(t *testing.T) {
	t.Parallel()

	cleanupTestEnv()
	defer cleanupTestEnv()

	t.Run("project path is set", func(t *testing.T) {
		t.Parallel()

		expectedPath := "/test/project/path"
		os.Setenv("MOCK_SVC_NAME_PROJECT_PATH", expectedPath)

		defer os.Unsetenv("MOCK_SVC_NAME_PROJECT_PATH")

		result := ProjectPathByEnv()
		if result != expectedPath {
			t.Errorf("ProjectPathByEnv() should return %s, got %s", expectedPath, result)
		}
	})

	t.Run("project path is empty", func(t *testing.T) {
		t.Parallel()
		os.Unsetenv("MOCK_SVC_NAME_PROJECT_PATH")

		result := ProjectPathByEnv()
		if result != "" {
			t.Error("ProjectPathByEnv() should return empty string when env var is not set")
		}
	})
}

func TestConfigPathFromEnv(t *testing.T) {
	t.Parallel()

	cleanupTestEnv()
	defer cleanupTestEnv()

	t.Run("config path is set", func(t *testing.T) {
		t.Parallel()

		expectedPath := "/test/config/path"
		os.Setenv("MOCK_SVC_NAME_CONFIG_PATH", expectedPath)

		defer os.Unsetenv("MOCK_SVC_NAME_CONFIG_PATH")

		result := ConfigPathFromEnv()
		if result != expectedPath {
			t.Errorf("ConfigPathFromEnv() should return %s, got %s", expectedPath, result)
		}
	})

	t.Run("config path is empty", func(t *testing.T) {
		t.Parallel()
		os.Unsetenv("MOCK_SVC_NAME_CONFIG_PATH")

		result := ConfigPathFromEnv()
		if result != "" {
			t.Error("ConfigPathFromEnv() should return empty string when env var is not set")
		}
	})
}

func TestConfigPathFromEnv_PanicWhenNotInited(t *testing.T) {
	// 不使用 t.Parallel(): 此测试修改包级别共享变量 isEnvInited，不能与其他测试并行
	// Save and restore isEnvInited
	oldInited := isEnvInited
	defer func() {
		isEnvInited = oldInited
	}()

	isEnvInited = false

	defer func() {
		if r := recover(); r == nil {
			t.Error("ConfigPathFromEnv() should panic when env not inited")
		}
	}()

	ConfigPathFromEnv()
}

func TestProjectPathByEnv_PanicWhenNotInited(t *testing.T) {
	// 不使用 t.Parallel(): 此测试修改包级别共享变量 isEnvInited，不能与其他测试并行
	// Save and restore isEnvInited
	oldInited := isEnvInited
	defer func() {
		isEnvInited = oldInited
	}()

	isEnvInited = false

	defer func() {
		if r := recover(); r == nil {
			t.Error("ProjectPathByEnv() should panic when env not inited")
		}
	}()

	ProjectPathByEnv()
}

func TestIsDebugMode_PanicWhenNotInited(t *testing.T) {
	// 不使用 t.Parallel(): 此测试修改包级别共享变量 isEnvInited，不能与其他测试并行
	// Save and restore isEnvInited
	oldInited := isEnvInited
	defer func() {
		isEnvInited = oldInited
	}()

	isEnvInited = false

	defer func() {
		if r := recover(); r == nil {
			t.Error("IsDebugMode() should panic when env not inited")
		}
	}()

	IsDebugMode()
}

func TestIsSQLPrint_PanicWhenNotInited(t *testing.T) {
	// 不使用 t.Parallel(): 此测试修改包级别共享变量 isEnvInited，不能与其他测试并行
	// Save and restore isEnvInited
	oldInited := isEnvInited
	defer func() {
		isEnvInited = oldInited
	}()

	isEnvInited = false

	defer func() {
		if r := recover(); r == nil {
			t.Error("IsSQLPrint() should panic when env not inited")
		}
	}()

	IsSQLPrint()
}

func TestIsLocalDev_PanicWhenNotInited(t *testing.T) {
	// 不使用 t.Parallel(): 此测试修改包级别共享变量 isEnvInited，不能与其他测试并行
	// Save and restore isEnvInited
	oldInited := isEnvInited
	defer func() {
		isEnvInited = oldInited
	}()

	isEnvInited = false

	defer func() {
		if r := recover(); r == nil {
			t.Error("IsLocalDev() should panic when env not inited")
		}
	}()

	IsLocalDev()
}
