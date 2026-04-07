// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package business_system

import (
	"context"

	"bkn-backend/common"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
)

type BusinessSystemServiceImpl struct {
	appSetting *common.AppSetting
	bsa        interfaces.BusinessSystemAccess
}

func NewBusinessSystemServiceImpl(appSetting *common.AppSetting) interfaces.BusinessSystemService {
	return &BusinessSystemServiceImpl{
		appSetting: appSetting,
		bsa:        logics.BSA,
	}
}

func (s *BusinessSystemServiceImpl) BindResource(ctx context.Context, bd_id string, rid string, rtype string) error {
	return s.bsa.BindResource(ctx, bd_id, rid, rtype)
}

func (s *BusinessSystemServiceImpl) UnbindResource(ctx context.Context, bd_id string, rid string, rtype string) error {
	return s.bsa.UnbindResource(ctx, bd_id, rid, rtype)
}
