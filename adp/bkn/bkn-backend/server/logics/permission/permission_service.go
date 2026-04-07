// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package permission

import (
	"sync"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	pServiceOnce sync.Once
	pService     interfaces.PermissionService
)

func NewPermissionService(appSetting *common.AppSetting) interfaces.PermissionService {
	pServiceOnce.Do(func() {
		if !common.GetAuthEnabled() {
			pService = NewNoopPermissionService(appSetting)
		} else {
			pService = NewPermissionServiceImpl(appSetting)
		}
	})
	return pService
}
