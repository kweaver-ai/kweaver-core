package auth

import (
	"sync"

	"vega-backend/common"
	"vega-backend/interfaces"
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
