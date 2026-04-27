package drivenadapters

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

func MockNewIndexBaseAccess(mockCtrl *gomock.Controller) (interfaces.IndexBaseAccess, *rmock.MockHTTPClient) {

	appSetting := &common.AppSetting{
		IndexBaseUrl: "http://localhost:13012/api/mdl-index-base/v1",
	}
	httpClient := rmock.NewMockHTTPClient(mockCtrl)
	iba := &indexBaseAccess{
		appSetting: appSetting,
		httpClient: httpClient,
	}

	return iba, httpClient
}

func Test_IndexBaseAccess_GetIndexBasesByTypes(t *testing.T) {
	Convey("Test GetIndexBasesByTypes", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		iba, httpClient := MockNewIndexBaseAccess(mockCtrl)

		baseResp := []interfaces.IndexBase{
			{
				BaseType: "a",
			},
			{
				BaseType: "b",
			},
		}
		okResp, _ := sonic.Marshal(baseResp)

		Convey("failed, caused by http error", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, nil, fmt.Errorf("method failed"))

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by status != 200", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusInternalServerError, okResp, nil)

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("failed, caused by unmarshal error", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			patch := ApplyFuncReturn(sonic.Unmarshal, errors.New("error"))
			defer patch.Reset()

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})

			So(indexbase, ShouldBeNil)
			So(err.Error(), ShouldEqual, "error")
		})

		Convey("failed, caused by the length of Data is less than requested", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b", "c"})

			So(indexbase, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			httpClient.EXPECT().GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				http.StatusOK, okResp, nil)

			indexbase, err := iba.GetIndexBasesByTypes(testCtx, []string{"a", "b"})
			So(indexbase, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}
