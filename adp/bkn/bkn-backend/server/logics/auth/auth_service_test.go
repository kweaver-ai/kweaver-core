// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package auth

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func newGinContext(accountID, accountType string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, accountID)
	c.Request.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, accountType)
	return c
}

func Test_NoopAuthService_VerifyToken(t *testing.T) {
	Convey("Test NoopAuthService VerifyToken\n", t, func() {
		svc := NewNoopAuthService(&common.AppSetting{})
		ctx := context.Background()

		Convey("Returns visitor built from headers, no error\n", func() {
			c := newGinContext("user-1", "user")

			visitor, err := svc.VerifyToken(ctx, c)

			So(err, ShouldBeNil)
			So(visitor.ID, ShouldEqual, "user-1")
			So(string(visitor.Type), ShouldEqual, "user")
		})

		Convey("Returns empty visitor when headers are absent\n", func() {
			c := newGinContext("", "")

			visitor, err := svc.VerifyToken(ctx, c)

			So(err, ShouldBeNil)
			So(visitor.ID, ShouldEqual, "")
		})
	})
}

func Test_HydraAuthService_VerifyToken(t *testing.T) {
	Convey("Test hydraAuthService VerifyToken\n", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		aa := bmock.NewMockAuthAccess(mockCtrl)
		svc := &hydraAuthService{
			appSetting: &common.AppSetting{},
			aa:         aa,
		}
		ctx := context.Background()
		c := newGinContext("user-1", "user")

		Convey("Success: delegates to AuthAccess and returns visitor\n", func() {
			expected := hydra.Visitor{ID: "user-1", Type: hydra.VisitorType("user")}
			aa.EXPECT().VerifyToken(ctx, c).Return(expected, nil)

			visitor, err := svc.VerifyToken(ctx, c)

			So(err, ShouldBeNil)
			So(visitor.ID, ShouldEqual, "user-1")
		})

		Convey("Failed: AuthAccess returns error\n", func() {
			authErr := errors.New("invalid token")
			aa.EXPECT().VerifyToken(ctx, c).Return(hydra.Visitor{}, authErr)

			visitor, err := svc.VerifyToken(ctx, c)

			So(err, ShouldEqual, authErr)
			So(visitor.ID, ShouldEqual, "")
		})
	})
}
