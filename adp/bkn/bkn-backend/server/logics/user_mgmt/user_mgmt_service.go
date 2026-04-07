// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package user_mgmt

import (
	"sync"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	umServiceOnce sync.Once
	umService     interfaces.UserMgmtService
)

func NewUserMgmtService(appSetting *common.AppSetting) interfaces.UserMgmtService {
	umServiceOnce.Do(func() {
		if !common.GetAuthEnabled() {
			umService = NewNoopUserMgmtService(appSetting)
		} else {
			umService = NewUserMgmtServiceImpl(appSetting)
		}
	})
	return umService
}
