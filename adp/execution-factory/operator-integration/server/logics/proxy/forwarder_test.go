package proxy

import (
	"context"
	"testing"

	myErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestForwardStream_MissingResponseWriterDoesNotPanic(t *testing.T) {
	Convey("ForwardStream handles missing response writer without panic", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		forwarder := &forwarder{
			logger: logger.DefaultLogger(),
		}

		req := &interfaces.HTTPRequest{
			HTTPRouter: interfaces.HTTPRouter{
				Method: "GET",
				URL:    "http://example.com",
			},
		}

		So(func() {
			resp, err := forwarder.ForwardStream(context.Background(), req)

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)

			httpErr, ok := err.(*myErr.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, 500)
			So(httpErr.ErrorDetails, ShouldEqual, "response writer not found in context")
		}, ShouldNotPanic)
	})
}
