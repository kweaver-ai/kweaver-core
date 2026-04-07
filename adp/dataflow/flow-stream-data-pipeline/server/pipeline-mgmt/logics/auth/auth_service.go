package auth

import (
	"sync"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

var (
	authServiceOnce sync.Once
	authService     interfaces.AuthService
)

func NewAuthService(appSetting *common.AppSetting) interfaces.AuthService {
	authServiceOnce.Do(func() {
		if !common.GetAuthEnabled() {
			authService = NewNoopAuthService(appSetting)
		} else {
			authService = NewHydraAuthService(appSetting)
		}
	})
	return authService
}
