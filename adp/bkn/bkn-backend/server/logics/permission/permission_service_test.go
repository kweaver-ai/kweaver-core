// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package permission

import (
	"context"
	"errors"
	"testing"

	mqclient "github.com/kweaver-ai/proton-mq-sdk-go"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

// mockMQClient is a minimal in-test stub for mqclient.ProtonMQClient.
type mockMQClient struct {
	pubErr error
}

func (m *mockMQClient) Pub(topic string, msg []byte) error { return m.pubErr }
func (m *mockMQClient) Sub(topic, channel string, handler mqclient.MessageHandler, pollIntervalMilliseconds int64, maxInFlight int, opts ...mqclient.SubOpt) error {
	return nil
}
func (m *mockMQClient) Close() {}

// withAccountInfo attaches an AccountInfo to a context.
func withAccountInfo(ctx context.Context, id, typ string) context.Context {
	return context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{ID: id, Type: typ})
}

// ── NoopPermissionService ────────────────────────────────────────────────────

func Test_NoopPermissionService_CheckPermission(t *testing.T) {
	Convey("Test NoopPermissionService CheckPermission always returns nil\n", t, func() {
		svc := NewNoopPermissionService(&common.AppSetting{})
		err := svc.CheckPermission(context.Background(), interfaces.PermissionResource{Type: "kn", ID: "kn1"}, []string{"read"})
		So(err, ShouldBeNil)
	})
}

func Test_NoopPermissionService_CreateResources(t *testing.T) {
	Convey("Test NoopPermissionService CreateResources always returns nil\n", t, func() {
		svc := NewNoopPermissionService(&common.AppSetting{})
		err := svc.CreateResources(context.Background(), []interfaces.PermissionResource{{Type: "kn", ID: "kn1"}}, []string{"read"})
		So(err, ShouldBeNil)
	})
}

func Test_NoopPermissionService_DeleteResources(t *testing.T) {
	Convey("Test NoopPermissionService DeleteResources always returns nil\n", t, func() {
		svc := NewNoopPermissionService(&common.AppSetting{})
		err := svc.DeleteResources(context.Background(), "kn", []string{"kn1", "kn2"})
		So(err, ShouldBeNil)
	})
}

func Test_NoopPermissionService_FilterResources(t *testing.T) {
	Convey("Test NoopPermissionService FilterResources\n", t, func() {
		svc := NewNoopPermissionService(&common.AppSetting{})

		Convey("Returns all IDs with fullOps\n", func() {
			ids := []string{"kn1", "kn2"}
			fullOps := []string{"read", "write"}
			result, err := svc.FilterResources(context.Background(), "kn", ids, []string{"read"}, true, fullOps)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 2)
			So(result["kn1"].Operations, ShouldResemble, fullOps)
			So(result["kn2"].Operations, ShouldResemble, fullOps)
		})

		Convey("Returns empty map for empty input\n", func() {
			result, err := svc.FilterResources(context.Background(), "kn", []string{}, []string{"read"}, true, []string{"read"})
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})
	})
}

func Test_NoopPermissionService_UpdateResource(t *testing.T) {
	Convey("Test NoopPermissionService UpdateResource always returns nil\n", t, func() {
		svc := NewNoopPermissionService(&common.AppSetting{})
		err := svc.UpdateResource(context.Background(), interfaces.PermissionResource{Type: "kn", ID: "kn1"})
		So(err, ShouldBeNil)
	})
}

// ── PermissionServiceImpl ────────────────────────────────────────────────────

func newTestPermissionImpl(t *testing.T) (*PermissionServiceImpl, *gomock.Controller, *bmock.MockPermissionAccess, *mockMQClient) {
	t.Helper()
	mockCtrl := gomock.NewController(t)
	pa := bmock.NewMockPermissionAccess(mockCtrl)
	mq := &mockMQClient{}
	svc := &PermissionServiceImpl{
		appSetting: &common.AppSetting{},
		pa:         pa,
		mqClient:   mq,
	}
	return svc, mockCtrl, pa, mq
}

func Test_PermissionServiceImpl_CheckPermission(t *testing.T) {
	Convey("Test PermissionServiceImpl CheckPermission\n", t, func() {
		svc, mockCtrl, pa, _ := newTestPermissionImpl(t)
		defer mockCtrl.Finish()

		resource := interfaces.PermissionResource{Type: "kn", ID: "kn1"}
		ops := []string{"read"}

		Convey("Failed: missing account info in context\n", func() {
			err := svc.CheckPermission(context.Background(), resource, ops)
			So(err, ShouldNotBeNil)
		})

		Convey("Success: pa returns true\n", func() {
			ctx := withAccountInfo(context.Background(), "u1", "user")
			pa.EXPECT().CheckPermission(gomock.Any(), gomock.Any()).Return(true, nil)

			err := svc.CheckPermission(ctx, resource, ops)
			So(err, ShouldBeNil)
		})

		Convey("Failed: pa returns false (forbidden)\n", func() {
			ctx := withAccountInfo(context.Background(), "u1", "user")
			pa.EXPECT().CheckPermission(gomock.Any(), gomock.Any()).Return(false, nil)

			err := svc.CheckPermission(ctx, resource, ops)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed: pa returns error\n", func() {
			ctx := withAccountInfo(context.Background(), "u1", "user")
			pa.EXPECT().CheckPermission(gomock.Any(), gomock.Any()).Return(false, errors.New("access error"))

			err := svc.CheckPermission(ctx, resource, ops)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_PermissionServiceImpl_CreateResources(t *testing.T) {
	Convey("Test PermissionServiceImpl CreateResources\n", t, func() {
		svc, mockCtrl, pa, _ := newTestPermissionImpl(t)
		defer mockCtrl.Finish()

		resources := []interfaces.PermissionResource{{Type: "kn", ID: "kn1"}}
		ops := []string{"read"}

		Convey("Failed: missing account info\n", func() {
			err := svc.CreateResources(context.Background(), resources, ops)
			So(err, ShouldNotBeNil)
		})

		Convey("Success: pa.CreateResources returns nil\n", func() {
			ctx := withAccountInfo(context.Background(), "u1", "user")
			pa.EXPECT().CreateResources(gomock.Any(), gomock.Any()).Return(nil)

			err := svc.CreateResources(ctx, resources, ops)
			So(err, ShouldBeNil)
		})

		Convey("Failed: pa.CreateResources returns error\n", func() {
			ctx := withAccountInfo(context.Background(), "u1", "user")
			pa.EXPECT().CreateResources(gomock.Any(), gomock.Any()).Return(errors.New("create failed"))

			err := svc.CreateResources(ctx, resources, ops)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_PermissionServiceImpl_DeleteResources(t *testing.T) {
	Convey("Test PermissionServiceImpl DeleteResources\n", t, func() {
		svc, mockCtrl, pa, _ := newTestPermissionImpl(t)
		defer mockCtrl.Finish()

		Convey("Empty IDs: returns nil without calling pa\n", func() {
			err := svc.DeleteResources(context.Background(), "kn", []string{})
			So(err, ShouldBeNil)
		})

		Convey("Success: pa.DeleteResources returns nil\n", func() {
			ctx := context.Background()
			pa.EXPECT().DeleteResources(gomock.Any(), gomock.Any()).Return(nil)

			err := svc.DeleteResources(ctx, "kn", []string{"kn1"})
			So(err, ShouldBeNil)
		})

		Convey("Failed: pa.DeleteResources returns error\n", func() {
			ctx := context.Background()
			pa.EXPECT().DeleteResources(gomock.Any(), gomock.Any()).Return(errors.New("delete failed"))

			err := svc.DeleteResources(ctx, "kn", []string{"kn1"})
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_PermissionServiceImpl_FilterResources(t *testing.T) {
	Convey("Test PermissionServiceImpl FilterResources\n", t, func() {
		svc, mockCtrl, pa, _ := newTestPermissionImpl(t)
		defer mockCtrl.Finish()

		Convey("Failed: missing account info\n", func() {
			result, err := svc.FilterResources(context.Background(), "kn", []string{"kn1"}, []string{"read"}, true, []string{"read"})
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Success: returns resource ops map\n", func() {
			ctx := withAccountInfo(context.Background(), "u1", "user")
			paResult := map[string]interfaces.PermissionResourceOps{
				"kn1": {
					ResourceID: "kn1", Operations: []string{"read"},
				},
			}
			pa.EXPECT().FilterResources(gomock.Any(), gomock.Any()).Return(paResult, nil)

			result, err := svc.FilterResources(ctx, "kn", []string{"kn1"}, []string{"read"}, true, []string{"read"})
			So(err, ShouldBeNil)
			So(result["kn1"].ResourceID, ShouldEqual, "kn1")
		})

		Convey("Failed: pa.FilterResources returns error\n", func() {
			ctx := withAccountInfo(context.Background(), "u1", "user")
			pa.EXPECT().FilterResources(gomock.Any(), gomock.Any()).Return(nil, errors.New("filter error"))

			result, err := svc.FilterResources(ctx, "kn", []string{"kn1"}, []string{"read"}, true, []string{"read"})
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_PermissionServiceImpl_UpdateResource(t *testing.T) {
	Convey("Test PermissionServiceImpl UpdateResource\n", t, func() {
		svc, mockCtrl, _, mq := newTestPermissionImpl(t)
		defer mockCtrl.Finish()

		Convey("Success: mqClient.Pub returns nil\n", func() {
			err := svc.UpdateResource(context.Background(), interfaces.PermissionResource{Type: "kn", ID: "kn1", Name: "Test"})
			So(err, ShouldBeNil)
		})

		Convey("Failed: mqClient.Pub returns error\n", func() {
			mq.pubErr = errors.New("mq unavailable")
			err := svc.UpdateResource(context.Background(), interfaces.PermissionResource{Type: "kn", ID: "kn1", Name: "Test"})
			So(err, ShouldNotBeNil)
		})
	})
}
