package common

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func TestResponseFormatContextHelpers(t *testing.T) {
	convey.Convey("SetResponseFormatToCtx and GetResponseFormatFromCtx", t, func() {
		type responseFormat string

		ctx := context.Background()
		ctx = SetResponseFormatToCtx(ctx, responseFormat("toon"))

		v, ok := GetResponseFormatFromCtx(ctx)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(v, convey.ShouldEqual, responseFormat("toon"))
	})
}

func TestIsPublicAPIFromCtx(t *testing.T) {
	convey.Convey("SetPublicAPIToCtx and IsPublicAPIFromCtx", t, func() {
		ctx := context.Background()
		convey.So(IsPublicAPIFromCtx(ctx), convey.ShouldBeFalse)

		ctx = SetPublicAPIToCtx(ctx, true)
		convey.So(IsPublicAPIFromCtx(ctx), convey.ShouldBeTrue)
	})
}

func TestGetHeaderFromCtx(t *testing.T) {
	convey.Convey("GetHeaderFromCtx returns account headers when auth context exists", t, func() {
		ctx := context.Background()
		authCtx := &interfaces.AccountAuthContext{
			AccountID:   "user-1",
			AccountType: interfaces.AccessorType("tenant"),
		}
		ctx = SetAccountAuthContextToCtx(ctx, authCtx)

		header := GetHeaderFromCtx(ctx)
		convey.So(header[string(interfaces.HeaderXAccountID)], convey.ShouldEqual, "user-1")
		convey.So(header[string(interfaces.HeaderXAccountType)], convey.ShouldEqual, "tenant")
	})
}
