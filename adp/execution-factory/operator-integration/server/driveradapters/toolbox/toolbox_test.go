package toolbox

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestCreateToolBox(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)
	toolService := mocks.NewMockIToolService(ctrl)
	validator := mocks.NewMockValidator(ctrl)
	handler := &toolBoxHandler{
		Logger:      mockLogger,
		ToolService: toolService,
		Validator:   validator,
	}
	path := "/tool-box"
	contentType := "Content-Type"
	applicationJSON := "application/json"
	applicationUrlencoded := "application/x-www-form-urlencoded"
	headers := map[string]string{
		contentType: applicationJSON,
	}
	Convey("TestCreateToolBox", t, func() {
		Convey("TestCreateToolBox: application/json 参数为空", func() {
			recorder := mocks.MockPostRequest(path, headers, http.NoBody, handler.CreateToolBox)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})
		Convey("TestCreateToolBox: application/json 参数错误", func() {
			recorder := mocks.MockPostRequest(path, headers, strings.NewReader(`{}`), handler.CreateToolBox)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})
		Convey("TestCreateToolBox: application/x-www-form-urlencoded 参数错误", func() {
			headers[contentType] = applicationUrlencoded
			headers["user_id"] = "1"
			formData := url.Values{}
			formData.Add("metadata_type", "test")
			formData.Add("data", "1")
			recorder := mocks.MockPostRequest(path, headers,
				strings.NewReader(formData.Encode()),
				handler.CreateToolBox)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}
