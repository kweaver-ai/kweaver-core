package helpers

import "os"

const (
	EnvPrefix     = "AGENT_FACTORY_"
	EnvIsLocalDev = EnvPrefix + "LOCAL_DEV"  // AGENT_FACTORY_LOCAL_DEV
	isDebugMode   = EnvPrefix + "DEBUG_MODE" // AGENT_FACTORY_DEBUG_MODE

	isSQLPrint = EnvPrefix + "SQL_PRINT" // AGENT_FACTORY_SQL_PRINT

	projPath = EnvPrefix + "PROJECT_PATH" // AGENT_FACTORY_PROJECT_PATH

	skipOauthVerify = EnvPrefix + "SKIP_OAUTH_VERIFY"
)

var mockIsLocalDev bool

func IsLocalDev() bool {
	return os.Getenv(EnvIsLocalDev) == "true" || mockIsLocalDev
}

func IsAaronLocalDev() bool {
	return os.Getenv(EnvIsLocalDev+"_AARON") == "true"
}

func SetIsLocalDev() {
	mockIsLocalDev = true
}

func IsDebugMode() bool {
	return os.Getenv(isDebugMode) == "true"
}

func IsOprLogShowLogForDebug() bool {
	return IsDebugMode()
}

func IsSQLPrint() bool {
	return os.Getenv(isSQLPrint) == "true"
}

func ProjectPathByEnv() string {
	return os.Getenv(projPath)
}

func IsSkipOauthVerify() bool {
	return os.Getenv(skipOauthVerify) == "true"
}
