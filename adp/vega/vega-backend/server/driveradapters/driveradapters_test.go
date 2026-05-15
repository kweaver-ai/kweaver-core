// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"vega-backend/common"
	"vega-backend/interfaces"
	"vega-backend/worker"

	"github.com/gin-gonic/gin"
)

func setGinMode() func() {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(oldMode)
	}
}

func MockNewRestHandler(
	appSetting *common.AppSetting,
	as interfaces.AuthService,
	cs interfaces.CatalogService,
	rs interfaces.ResourceService,
	bts interfaces.BuildTaskService,
	ds interfaces.DatasetService,
	cts interfaces.ConnectorTypeService,
	dts interfaces.DiscoverTaskService,
	dss interfaces.DiscoverScheduleService,
	rds interfaces.ResourceDataService,
	sw *worker.ScheduleWorker,
) *restHandler {
	return &restHandler{
		appSetting: appSetting,
		as:         as,
		cs:         cs,
		rs:         rs,
		bts:        bts,
		ds:         ds,
		cts:        cts,
		dts:        dts,
		dss:        dss,
		rds:        rds,
		sw:         sw,
	}
}
