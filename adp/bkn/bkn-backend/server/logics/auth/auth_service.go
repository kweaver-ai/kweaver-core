// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package auth

import (
	"sync"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	authServiceOnce sync.Once
	authService     interfaces.AuthService
)

func NewAuthService(appSetting *common.AppSetting) interfaces.AuthService {
	// 认证禁用时返回空实现
	authServiceOnce.Do(func() {
		if !common.GetAuthEnabled() {
			authService = NewNoopAuthService(appSetting)
		} else {
			authService = NewHydraAuthService(appSetting)
		}
	})
	return authService
}
