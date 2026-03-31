package operator

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func mockPostRequest(url, contentType string, body io.Reader, handler func(c *gin.Context)) (recorder *httptest.ResponseRecorder) {
	// 创建一个带有中间件的路由组
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use()
	router.Handle(http.MethodPost, url, func(c *gin.Context) {
		handler(c)
		c.Next()
	})
	// 创建请求并触发中间件
	recorder = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("user_id", "user_id")
	router.ServeHTTP(recorder, req)
	return recorder
}

func mockGetRequest(path string, query map[string]interface{}, pathParams []string, headers map[string]string, handler func(c *gin.Context)) (recorder *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use()

	router.Handle(http.MethodGet, path, func(c *gin.Context) {
		for i, param := range pathParams {
			paramName := strings.Split(path, "/")[i+1][1:] // 提取 :param 格式的参数名
			c.Params = append(c.Params, gin.Param{Key: paramName, Value: param})
		}
		handler(c)
		c.Next()
	})
	formattedPath := path
	for _, param := range pathParams {
		// 找到第一个占位符的位置（如 :id）
		start := strings.Index(formattedPath, ":")
		if start == -1 {
			break
		}
		end := strings.Index(formattedPath[start:], "/")
		if end == -1 {
			end = len(formattedPath)
		} else {
			end += start
		}
		// 替换占位符为实际参数
		formattedPath = formattedPath[:start] + param + formattedPath[end:]
	}
	// 构造请求路径（移除了错误的路径拼接）
	queryString := url.Values{}
	for key, value := range query {
		queryString.Add(key, fmt.Sprintf("%v", value))
	}

	if len(queryString) > 0 {
		formattedPath += "?" + queryString.Encode()
	}

	req := httptest.NewRequest(http.MethodGet, formattedPath, http.NoBody)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestRegisterOperator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOperatorManager := mocks.NewMockOperatorManager(ctrl)
	mockHydra := mocks.NewMockHydra(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	handler := &operatorHandle{
		OperatorManager: mockOperatorManager,
		Hydra:           mockHydra,
		Logger:          mockLogger,
		Validator:       mockValidator,
	}

	mockValidator.EXPECT().ValidateOperatorImportSize(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	path := "/operator/register"
	applicationJSON := "application/json"
	localFile := "../../tests/file/json/auth.json"
	mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).Return().AnyTimes()
	Convey("TestRegisterOperator:参数校验", t, func() {
		Convey("空参数请求", func() {
			// mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			// mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidatorStruct")).Times(1)
			recorder := mockPostRequest(path, applicationJSON, http.NoBody, handler.OperatorRegister)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldNotEqual, http.StatusOK)
		})
		Convey("传参格式为：multipart/form-json，MetadataType为空", func() {
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidatorStruct")).Times(1)
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			// 文件部分
			part, _ := writer.CreateFormFile("data", "auth.json")
			data, err := os.ReadFile(localFile)
			So(err, ShouldBeNil)
			_, _ = part.Write(data)
			_ = writer.Close()
			recorder := mockPostRequest(path, writer.FormDataContentType(),
				body, handler.OperatorRegister)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldNotEqual, http.StatusOK)
		})
		Convey("传参格式为：multipart/form-data；认证失败", func() {
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockHydra.EXPECT().Introspect(gomock.Any()).Return(nil, errors.New("mock"))
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			// 添加必填字段
			mustWriteField := func(fieldName, value string) {
				err := writer.WriteField(fieldName, value)
				So(err, ShouldBeNil)
			}
			mustWriteField("operator_metadata_type", "openapi")
			// 文件部分
			part, _ := writer.CreateFormFile("data", "auth.json")
			data, err := os.ReadFile(localFile)
			So(err, ShouldBeNil)
			_, _ = part.Write(data)
			_ = writer.Close()
			recorder := mockPostRequest(path, writer.FormDataContentType(),
				body, handler.OperatorRegister)
			fmt.Println(recorder.Body.String())
		})
		// Convey("传参格式为：application/json，MetadataType为空", func() {
		// 	recorder := mockPostRequest(path, applicationJSON, bytes.NewBufferString(utils.ObjectToJSON(&interfaces.OperatorRegisterReq{
		// 		Data: "{}",
		// 	})), handler.OperatorRegister)
		// 	fmt.Println(recorder.Body.String())
		// 	So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		// })
		// Convey("传参格式为：application/json，无效传参", func() {
		// 	mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 	recorder := mockPostRequest(path, applicationJSON, bytes.NewBufferString("nil"), handler.OperatorRegister)
		// 	fmt.Println(recorder.Body.String())
		// 	So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		// })
		// Convey("无效用户token", func() {
		// 	mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
		// 	mockHydra.EXPECT().Introspect(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock")).Times(1)
		// 	recorder := mockPostRequest(path, applicationJSON, bytes.NewBufferString(`{
		// 		"user_token": "invalid_token",
		// 		"operator_metadata_type": "openapi",
		// 		"data": "test"
		// 	}`), handler.OperatorRegister)
		// 	fmt.Println(recorder.Body.String())
		// })
		// Convey("operator_metadata_type 类型无效", func() {
		// 	mockHydra.EXPECT().Introspect(gomock.Any(), gomock.Any()).Return(&interfaces.TokenInfo{}, nil)
		// 	recorder := mockPostRequest(path, applicationJSON, bytes.NewBufferString(`{
		// 		"user_token": "invalid_token",
		// 		"operator_metadata_type": "api",
		// 		"data": "test"
		// 	}`), handler.OperatorRegister)
		// 	fmt.Println(recorder.Body.String())
		// 	So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		// })
		Convey("合法注册请求", func() {
			tokenInfo := &interfaces.TokenInfo{VisitorID: "user123"}
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			mockHydra.EXPECT().Introspect(gomock.Any()).Return(tokenInfo, nil).Times(1)
			mockOperatorManager.EXPECT().RegisterOperatorByOpenAPI(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.OperatorRegisterResp{}, nil).Times(1)
			recorder := mockPostRequest(path, applicationJSON, bytes.NewBufferString(`{
				"user_token": "valid_token",
				"operator_metadata_type": "openapi",
				"data": "test"
			}`), handler.OperatorRegister)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusOK)
			fmt.Println(recorder.Body.String())
		})
	})
}

func mockRegisterValidation() {
	_ = validatorv10.New().RegisterValidation("uuid4", func(fl validatorv10.FieldLevel) bool {
		return govalidator.IsUUIDv4(fl.Field().String())
	})
}

func TestOperatorUpdateByOpenAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockOperatorManager := mocks.NewMockOperatorManager(ctrl)
	mockHydra := mocks.NewMockHydra(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mockValidator := mocks.NewMockValidator(ctrl)
	handler := &operatorHandle{
		OperatorManager: mockOperatorManager,
		Hydra:           mockHydra,
		Logger:          mockLogger,
		Validator:       mockValidator,
	}
	mockValidator.EXPECT().ValidateOperatorImportSize(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).Return().AnyTimes()
	mockOperatorID := "b2d8baf0-e31f-4cac-851d-30ad8c2e4722"
	mockOperatorVersion := "416278e0-2816-4537-a974-fbe46a3a7720"
	// 模拟服务起来时主动注册验证器
	mockRegisterValidation()
	path := "/operator/info/update"
	localFile := "../../tests/file/json/auth.json"
	applicationJSON := "application/json"
	Convey("TestOperatorUpdateByOpenAPI:参数校验", t, func() {
		Convey("空参数请求", func() {
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(mocks.MockFuncErr("ValidatorStruct")).Times(1)
			// mockHydra.EXPECT().Introspect(gomock.Any(), gomock.Any()).Return(&interfaces.TokenInfo{}, nil)
			recorder := mockPostRequest(path, applicationJSON, bytes.NewBufferString("{}"), handler.OperatorUpdateByOpenAPI)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldNotEqual, http.StatusOK)
		})
		Convey("传参格式为：multipart/form-json，MetadataType为空", func() {
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			mockHydra.EXPECT().Introspect(gomock.Any()).Return(&interfaces.TokenInfo{}, nil).AnyTimes()
			mockOperatorManager.EXPECT().UpdateOperatorByOpenAPI(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*interfaces.OperatorRegisterResp{}, nil).AnyTimes()
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			// 文件部分
			part, _ := writer.CreateFormFile("data", "auth.json")
			data, err := os.ReadFile(localFile)
			So(err, ShouldBeNil)
			_, _ = part.Write(data)
			_ = writer.Close()
			recorder := mockPostRequest(path, writer.FormDataContentType(),
				body, handler.OperatorUpdateByOpenAPI)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusOK)
		})
		Convey("传参格式为：multipart/form-data: 认证未通过", func() {
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			mockHydra.EXPECT().Introspect(gomock.Any()).Return(nil, errors.New("mock")).AnyTimes()
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			// 添加必填字段
			mustWriteField := func(fieldName, value string) {
				err := writer.WriteField(fieldName, value)
				So(err, ShouldBeNil)
			}
			mustWriteField("operator_metadata_type", "openapi")
			mustWriteField("operator_id", mockOperatorID)
			mustWriteField("version", mockOperatorVersion)

			// 文件部分
			part, _ := writer.CreateFormFile("data", "auth.json")
			data, err := os.ReadFile(localFile)
			So(err, ShouldBeNil)
			_, _ = part.Write(data)
			_ = writer.Close()
			recorder := mockPostRequest(path, writer.FormDataContentType(),
				body, handler.OperatorUpdateByOpenAPI)
			fmt.Println(recorder.Body.String())
		})
		Convey("传参格式为：multipart/form-data: 认证通过", func() {
			mockValidator.EXPECT().ValidatorStruct(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			mockHydra.EXPECT().Introspect(gomock.Any()).Return(&interfaces.TokenInfo{}, nil).AnyTimes()
			mockOperatorManager.EXPECT().UpdateOperatorByOpenAPI(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				[]*interfaces.OperatorRegisterResp{}, nil).AnyTimes()
			recorder := mockPostRequest(path, applicationJSON, bytes.NewBufferString(utils.ObjectToJSON(&interfaces.OperatorUpdateReq{
				OperatorID: mockOperatorID,
				OperatorRegisterReq: &interfaces.OperatorRegisterReq{
					Data:         "{}",
					MetadataType: "openapi",
					UserToken:    "mock_usre_token",
				},
			})), handler.OperatorUpdateByOpenAPI)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func TestOperatorQueryPage(t *testing.T) {
	Convey("TestOperatorQueryPage:参数校验", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockOperatorManager := mocks.NewMockOperatorManager(ctrl)
		mockHydra := mocks.NewMockHydra(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		handler := &operatorHandle{
			OperatorManager: mockOperatorManager,
			Hydra:           mockHydra,
			Logger:          mockLogger,
		}
		path := "/operator/info/list"
		headers := map[string]string{
			"x-business-domain": "domain123",
		}
		Convey("校验默认值，默认查询第一页，页面大小为10", func() {
			req := &interfaces.PageQueryRequest{
				BusinessDomainID: "domain123",
				Page:             1,
				PageSize:         10,
				SortOrder:        "desc",
				SortBy:           "update_time",
			}
			mockOperatorManager.EXPECT().GetOperatorQueryPage(gomock.Any(),
				req).Return(&interfaces.PageQueryResponse{}, nil)
			recorder := mockGetRequest(path, map[string]interface{}{}, []string{}, headers, handler.OperatorQueryPage)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusOK)
		})
		Convey("排序字段无效", func() {
			recorder := mockGetRequest(path, map[string]interface{}{
				"sort_by": "a",
			}, []string{}, headers, handler.OperatorQueryPage)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})
		Convey("排序规则无效", func() {
			recorder := mockGetRequest(path, map[string]interface{}{
				"sort_order": "b",
			}, []string{}, headers, handler.OperatorQueryPage)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})
		Convey("获取第二页，页面大小为30,处理失败", func() {
			mockOperatorManager.EXPECT().GetOperatorQueryPage(gomock.Any(),
				gomock.Any()).Return(nil, errors.New("mock"))
			recorder := mockGetRequest(path, map[string]interface{}{
				"page":      2,
				"page_size": 30,
				"status":    "published",
			}, []string{}, headers, handler.OperatorQueryPage)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func TestOperatorQueryByOperatorIDOrVersion(t *testing.T) {
	Convey("TestOperatorQueryByOperatorIDOrVersion", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockOperatorManager := mocks.NewMockOperatorManager(ctrl)
		mockHydra := mocks.NewMockHydra(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		handler := &operatorHandle{
			OperatorManager: mockOperatorManager,
			Hydra:           mockHydra,
			Logger:          mockLogger,
		}
		headers := map[string]string{}
		path := "/operator/info/:operator_id"
		Convey("查询成功", func() {
			mockOperatorManager.EXPECT().GetOperatorInfoByOperatorID(gomock.Any(), gomock.Any()).Return(&interfaces.OperatorDataInfo{}, nil)
			recorder := mockGetRequest(path, map[string]interface{}{
				"version": "version_mock",
			}, []string{"operator_id_mock"}, headers, handler.OperatorQueryByOperatorID)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusOK)
		})
		Convey("查询失败", func() {
			mockOperatorManager.EXPECT().GetOperatorInfoByOperatorID(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock"))
			recorder := mockGetRequest(path, map[string]interface{}{}, []string{"operator_id_mock"}, headers, handler.OperatorQueryByOperatorID)
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

// func TestOperatorCategoryList(t *testing.T) {
// 	Convey("TestOperatorCategoryList", t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()
// 		mockOperatorManager := mocks.NewMockOperatorManager(ctrl)
// 		mockHydra := mocks.NewMockHydra(ctrl)
// 		mockLogger := mocks.NewMockLogger(ctrl)
// 		handler := &operatorHandle{
// 			OperatorManager: mockOperatorManager,
// 			Hydra:           mockHydra,
// 			Logger:          mockLogger,
// 		}
// 		path := "/operator/category"
// 		Convey("查询成功", func() {
// 			mockOperatorManager.EXPECT().GetOperatorCategoryList(gomock.Any()).Return([]*interfaces.CategoryInfo{}, nil)
// 			recorder := mockGetRequest(path, map[string]interface{}{}, []string{}, handler.OperatorCategoryList)
// 			fmt.Println(recorder.Body.String())
// 			So(recorder.Code, ShouldEqual, http.StatusOK)
// 		})
// 		Convey("查询失败", func() {
// 			mockOperatorManager.EXPECT().GetOperatorCategoryList(gomock.Any()).Return(nil, errors.New("mock err"))
// 			recorder := mockGetRequest(path, map[string]interface{}{}, []string{}, handler.OperatorCategoryList)
// 			fmt.Println(recorder.Body.String())
// 			So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
// 		})
// 	})
// }

func TestOperatorStatusUpdate(t *testing.T) {
	Convey("TestOperatorCategoryList", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockOperatorManager := mocks.NewMockOperatorManager(ctrl)
		mockHydra := mocks.NewMockHydra(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		handler := &operatorHandle{
			OperatorManager: mockOperatorManager,
			Hydra:           mockHydra,
			Logger:          mockLogger,
		}
		path := "/operator/status"
		applicationJSON := "application/json"
		Convey("更新成功", func() {
			mockHydra.EXPECT().Introspect(gomock.Any()).Return(&interfaces.TokenInfo{}, nil)
			mockOperatorManager.EXPECT().UpdateOperatorStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			reqStr := `[{ "status": "published","operator_id": "11176d4f-bd5c-471d-9e80-93c5830b78f8","version": "71a889c5-b425-4d9e-93b6-d6b3230eb14b"}]`
			recorder := mockPostRequest(path, applicationJSON, bytes.NewBufferString(reqStr), func(c *gin.Context) {
				ctx := c.Request.Context()
				c.Request = c.Request.WithContext(ctx)
				handler.OperatorStatusUpdate(c)
			})
			fmt.Println(recorder.Body.String())
			So(recorder.Code, ShouldEqual, http.StatusOK)
		})
	})
}

// func TestOperatorEdit(t *testing.T) {
// 	Convey("TestOperatorEdit", t, func() {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()
// 		mockOperatorManager := mocks.NewMockOperatorManager(ctrl)
// 		mockHydra := mocks.NewMockHydra(ctrl)
// 		mockLogger := mocks.NewMockLogger(ctrl)
// 		handler := &operatorHandle{
// 			OperatorManager: mockOperatorManager,
// 			Hydra:           mockHydra,
// 			Logger:          mockLogger,
// 		}
// 		path := "/operator/info"
// 		applicationJSON := "application/json"
// 		Convey("更新成功", func() {
// 			reqJSON := `"data": "./resource/openapi/compliant/template.yaml",
//             "operator_metadata_type": "openapi",
//             "operator_info": {
//                 "operator_type": "basic",
//                 "execution_mode": "sync",
//                 "category": "other_category"
//             },
//             "operator_execute_control": {
//                 "timeout": 0,
//                 "retry_policy": {
//                     "max_attempts": 3,
//                     "initial_delay": 1000,
//                     "max_delay": 6000,
//                     "backoff_factor": 2,
//                     "retry_conditions": {
//                         "status_code": [500],
//                         "error_codes": ["string"]
//                     }
//                 }
//             },
//             "direct_publish": false,
//             "user_token": "string"
//         }`
// 			recorder := mockPostRequest(path, applicationJSON, bytes.NewBufferString(reqJSON), func(c *gin.Context) {
// 				c.Set(string(interfaces.KeyTokenInfo), &interfaces.TokenInfo{VisitorID: "mock"})
// 				handler.OperatorEdit(c)
// 			})
// 			fmt.Println(recorder.Body.String())
// 			So(recorder.Code, ShouldEqual, http.StatusOK)
// 		})
// 	})

// }
