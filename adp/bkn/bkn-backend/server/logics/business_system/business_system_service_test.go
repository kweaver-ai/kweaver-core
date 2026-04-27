// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package business_system

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	bmock "bkn-backend/interfaces/mock"
)

func Test_NoopBusinessSystemService(t *testing.T) {
	Convey("Test NoopBusinessSystemService\n", t, func() {
		ctx := context.Background()
		appSetting := &common.AppSetting{}
		svc := NewNoopBusinessSystemService(appSetting)

		Convey("BindResource always returns nil\n", func() {
			err := svc.BindResource(ctx, "bd1", "rid1", "kn")
			So(err, ShouldBeNil)
		})

		Convey("UnbindResource always returns nil\n", func() {
			err := svc.UnbindResource(ctx, "bd1", "rid1", "kn")
			So(err, ShouldBeNil)
		})
	})
}

func Test_BusinessSystemServiceImpl(t *testing.T) {
	Convey("Test BusinessSystemServiceImpl\n", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		bsa := bmock.NewMockBusinessSystemAccess(mockCtrl)
		svc := &BusinessSystemServiceImpl{
			appSetting: appSetting,
			bsa:        bsa,
		}

		Convey("BindResource delegates to BusinessSystemAccess\n", func() {
			Convey("Success\n", func() {
				bsa.EXPECT().BindResource(gomock.Any(), "bd1", "rid1", "kn").Return(nil)
				err := svc.BindResource(ctx, "bd1", "rid1", "kn")
				So(err, ShouldBeNil)
			})

			Convey("Returns error from access layer\n", func() {
				bsa.EXPECT().BindResource(gomock.Any(), "bd1", "rid1", "kn").Return(errors.New("access error"))
				err := svc.BindResource(ctx, "bd1", "rid1", "kn")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("UnbindResource delegates to BusinessSystemAccess\n", func() {
			Convey("Success\n", func() {
				bsa.EXPECT().UnbindResource(gomock.Any(), "bd1", "rid1", "kn").Return(nil)
				err := svc.UnbindResource(ctx, "bd1", "rid1", "kn")
				So(err, ShouldBeNil)
			})

			Convey("Returns error from access layer\n", func() {
				bsa.EXPECT().UnbindResource(gomock.Any(), "bd1", "rid1", "kn").Return(errors.New("access error"))
				err := svc.UnbindResource(ctx, "bd1", "rid1", "kn")
				So(err, ShouldNotBeNil)
			})
		})
	})
}
