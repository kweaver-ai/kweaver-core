// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package business_system

import (
	"sync"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	bsServiceOnce sync.Once
	bsService     interfaces.BusinessSystemService
)

func NewBusinessSystemService(appSetting *common.AppSetting) interfaces.BusinessSystemService {
	bsServiceOnce.Do(func() {
		if !common.GetBusinessDomainEnabled() {
			bsService = NewNoopBusinessSystemService(appSetting)
		} else {
			bsService = NewBusinessSystemServiceImpl(appSetting)
		}
	})
	return bsService
}
