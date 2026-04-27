package logics

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-worker/interfaces"
	fmock "flow-stream-data-pipeline/pipeline-worker/interfaces/mock"
)

func TestIndexMgmtService_GetIndexBases(t *testing.T) {
	Convey("Test IndexMgmtAccess GetIndexBases", t, func() {
		ctx := context.Background()

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		imaMock := fmock.NewMockIndexBaseAccess(mockCtl)
		imsMock := &indexBaseService{
			appSetting: &common.AppSetting{},
			ima:        imaMock,
		}

		Convey("GetIndexBases failed", func() {
			imaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]*interfaces.IndexBaseInfo{}, fmt.Errorf("error"))

			_, err := imsMock.GetIndexBaseByBaseType(ctx, "test")
			So(err, ShouldNotBeNil)
		})

		Convey("no index base found", func() {
			imaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]*interfaces.IndexBaseInfo{}, nil)

			_, err := imsMock.GetIndexBaseByBaseType(ctx, "test")
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			imaMock.EXPECT().GetIndexBasesByTypes(gomock.Any(), gomock.Any()).
				Return([]*interfaces.IndexBaseInfo{
					{
						BaseType: "test",
					},
				}, nil)

			_, err := imsMock.GetIndexBaseByBaseType(ctx, "test")
			So(err, ShouldBeNil)
		})
	})
}
