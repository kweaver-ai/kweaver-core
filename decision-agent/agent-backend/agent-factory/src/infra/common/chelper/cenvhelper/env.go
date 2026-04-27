package cenvhelper

import (
	"strings"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
)

var (
	isLocalDev  EnvStr
	isDebugMode EnvStr
	isSQLPrint  EnvStr

	projPath EnvStr

	configPath EnvStr

	runScenarioEnv EnvStr // 运行场景

	isEnvInited bool
)

// 测试时，可通过initEnv()重新初始化
func init() {
	initEnv()
}

// InitEnvForTest 供测试的 TestMain 调用, 在设置好测试环境变量后重新初始化
// 因为 init() 在 TestMain 之前执行, 若 SERVICE_NAME 未预先设置到系统级 env 中
// 则 envPrefix 会是错误的值, 导致 IsLocalDev() 等函数读取到错误的 key
func InitEnvForTest() {
	initEnv()
}

// 测试时，可通过initEnv()重新初始化
func initEnv() {
	// 1. 获取服务名
	svcName := cutil.GetEnv("SERVICE_NAME", "AGENT_FACTORY")
	if svcName == "" {
		panic("env SERVICE_NAME not set")
	}

	// 2. 获取环境前缀
	envPrefix := strings.ToUpper(svcName) + "_"
	envPrefix = strings.Replace(envPrefix, "-", "_", -1)

	// 3. 初始化环境变量
	isLocalDev = NewEnvStr("LOCAL_DEV", envPrefix)   // example: AGENT_FACTORY_LOCAL_DEV
	isDebugMode = NewEnvStr("DEBUG_MODE", envPrefix) // example: AGENT_FACTORY_DEBUG_MODE
	isSQLPrint = NewEnvStr("SQL_PRINT", envPrefix)   // example: AGENT_FACTORY_SQL_PRINT
	projPath = NewEnvStr("PROJECT_PATH", envPrefix)  // example: AGENT_FACTORY_PROJECT_PATH
	configPath = NewEnvStr("CONFIG_PATH", envPrefix) // example: AGENT_FACTORY_CONFIG_PATH

	runScenarioEnv = NewEnvStr("RUN_SCENARIO", envPrefix) // example: AGENT_FACTORY_RUN_SCENARIO

	// 4. 设置"是否初始化完成"
	isEnvInited = true
}

func IsDebugMode() bool {
	if !isEnvInited {
		panic("env not inited")
	}

	return isDebugMode.Value() == "true"
}

func IsSQLPrint() bool {
	if !isEnvInited {
		panic("env not inited")
	}

	return isSQLPrint.Value() == "true"
}

func ProjectPathByEnv() string {
	if !isEnvInited {
		panic("env not inited")
	}

	return projPath.Value()
}

func ConfigPathFromEnv() string {
	if !isEnvInited {
		panic("env not inited")
	}

	return configPath.Value()
}
