package driveradapters

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"flow-stream-data-pipeline/common"
	ferrors "flow-stream-data-pipeline/errors"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
	fmock "flow-stream-data-pipeline/pipeline-mgmt/interfaces/mock"
)

func mockNewPipelineMgmtRestHandler(appSetting *common.AppSetting,
	as interfaces.AuthService, pmService interfaces.PipelineMgmtService) (r *restHandler) {

	r = &restHandler{
		appSetting:          appSetting,
		as:                  as,
		pipelineMgmtService: pmService,
	}
	return r
}

func Test_PipelineRestHandler_CreatePipeline(t *testing.T) {
	Convey("Test CreatePipeline", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				CpuMax:           8,
				MemoryMax:        8096,
				MaxPipelineCount: 100,
			},
		}
		asMock := fmock.NewMockAuthService(mockCtrl)
		pmService := fmock.NewMockPipelineMgmtService(mockCtrl)
		handler := mockNewPipelineMgmtRestHandler(appSetting, asMock, pmService)
		handler.RegisterPublic(engine)

		asMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/flow-stream-data-pipeline/v1/pipelines"

		pipelineTest := &interfaces.Pipeline{
			PipelineID:       "xyz",
			PipelineName:     "test1",
			OutputType:       "index_base",
			IndexBase:        "test",
			DeploymentConfig: &interfaces.DeploymentConfig{CpuLimit: 1, MemoryLimit: 1024},
		}

		Convey("Create failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, caused by the error from func validatePipeline", func() {
			pipelineTest1 := &interfaces.Pipeline{
				PipelineID:   "xyz",
				PipelineName: "",
				OutputType:   "index_base",
				IndexBase:    "test",
			}

			reqParamByte, _ := sonic.Marshal(pipelineTest1)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, caused by the error from pipeline count is over limit", func() {
			pmService.EXPECT().GetPipelineTotals(gomock.Any(), gomock.Any()).Return(100, nil)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Create failed, caused by check pipeline id failed", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)
			pmService.EXPECT().GetPipelineTotals(gomock.Any(), gomock.Any()).Return(5, nil)
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", false, expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Create failed, caused by data view id is duplicated", func() {
			pmService.EXPECT().GetPipelineTotals(gomock.Any(), gomock.Any()).Return(5, nil)
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", true, nil)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Create failed, caused by check pipeline name failed", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)
			pmService.EXPECT().GetPipelineTotals(gomock.Any(), gomock.Any()).Return(5, nil)
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)
			pmService.EXPECT().CheckPipelineExistByName(gomock.Any(), gomock.Any()).Return("", false, expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Create failed, caused by data view name duplicated", func() {
			pmService.EXPECT().GetPipelineTotals(gomock.Any(), gomock.Any()).Return(5, nil)
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)
			pmService.EXPECT().CheckPipelineExistByName(gomock.Any(), gomock.Any()).Return("", true, nil)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("Create failed, caused by the error from method CreatePipelines", func() {
			pmService.EXPECT().GetPipelineTotals(gomock.Any(), gomock.Any()).Return(5, nil)
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)
			pmService.EXPECT().CheckPipelineExistByName(gomock.Any(), gomock.Any()).Return("", false, nil)
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)
			pmService.EXPECT().CreatePipeline(gomock.Any(), gomock.Any()).Return("", expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Create succeed", func() {
			pmService.EXPECT().GetPipelineTotals(gomock.Any(), gomock.Any()).Return(5, nil)
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)
			pmService.EXPECT().CheckPipelineExistByName(gomock.Any(), gomock.Any()).Return("", false, nil)
			pmService.EXPECT().CreatePipeline(gomock.Any(), gomock.Any()).Return("1", nil)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

func Test_PipelineRestHandler_DeletePipelines(t *testing.T) {
	Convey("Test DeletePipelines", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		asMock := fmock.NewMockAuthService(mockCtrl)
		pmService := fmock.NewMockPipelineMgmtService(mockCtrl)
		handler := mockNewPipelineMgmtRestHandler(appSetting, asMock, pmService)
		handler.RegisterPublic(engine)

		asMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/flow-stream-data-pipeline/v1/pipelines/aa"

		Convey("Delete failed, caused by the error from method CheckPipelineExistByID", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", false, expectedHttpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete failed, caused by the error from pipeline not found", func() {
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Delete failed, caused by the error from method DeletePipeline", func() {
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", true, nil)

			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)
			pmService.EXPECT().DeletePipeline(gomock.Any(), gomock.Any()).Return(expectedHttpErr)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Delete succeed", func() {
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", true, nil)
			pmService.EXPECT().DeletePipeline(gomock.Any(), gomock.Any()).Return(nil)

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_PipelineRestHandler_UpdatePipeline(t *testing.T) {
	Convey("Test UpdatePipeline", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				CpuMax:    8,
				MemoryMax: 8096,
			},
		}
		asMock := fmock.NewMockAuthService(mockCtrl)
		pmService := fmock.NewMockPipelineMgmtService(mockCtrl)
		handler := mockNewPipelineMgmtRestHandler(appSetting, asMock, pmService)
		handler.RegisterPublic(engine)

		asMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/flow-stream-data-pipeline/v1/pipelines/1"

		pipelineTest := &interfaces.Pipeline{
			PipelineID:       "xyz",
			PipelineName:     "test1",
			OutputType:       "index_base",
			IndexBase:        "test",
			DeploymentConfig: &interfaces.DeploymentConfig{CpuLimit: 1, MemoryLimit: 1024},
		}

		Convey("Update failed, caused by the error from method ShouldBindJSON", func() {
			expectedErr := errors.New("some errors")
			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from func ValidatePipeline", func() {
			pipelineTest := &interfaces.Pipeline{
				PipelineID:   "xyz",
				PipelineName: "",
				OutputType:   "index_base",
				IndexBase:    "test",
			}

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from method UpdatePipeline", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)
			pmService.EXPECT().UpdatePipeline(gomock.Any(), gomock.Any()).Return(expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update succeed", func() {
			pmService.EXPECT().UpdatePipeline(gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

func Test_PipelineRestHandler_GetPipelines(t *testing.T) {
	Convey("Test GetPipelines", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		asMock := fmock.NewMockAuthService(mockCtrl)
		pmService := fmock.NewMockPipelineMgmtService(mockCtrl)
		handler := mockNewPipelineMgmtRestHandler(appSetting, asMock, pmService)
		handler.RegisterPublic(engine)

		asMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/flow-stream-data-pipeline/v1/pipelines/aa"

		Convey("Get failed, caused by the error from method GetPipeline", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)
			pmService.EXPECT().GetPipeline(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, false, expectedHttpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Get succeed", func() {
			pmService.EXPECT().GetPipeline(gomock.Any(), gomock.Any(), gomock.Any()).Return(&interfaces.Pipeline{}, true, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_PipelineRestHandler_ListPipelines(t *testing.T) {
	Convey("Test ListPipelines", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		asMock := fmock.NewMockAuthService(mockCtrl)
		pmService := fmock.NewMockPipelineMgmtService(mockCtrl)
		handler := mockNewPipelineMgmtRestHandler(appSetting, asMock, pmService)
		handler.RegisterPublic(engine)

		asMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/flow-stream-data-pipeline/v1/pipelines"

		Convey("List failed, caused by the error from builtin", func() {
			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0&builtin=foo"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("List failed, caused by the error from func validatePaginationQueryParameters", func() {
			url = url + "?direction=desc&sort=update_time&limit=1000&offset=a"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("List failed, caused by the error from method ListPipelines", func() {
			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0"
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)

			pmService.EXPECT().ListPipelines(gomock.Any(), gomock.Any()).Return(nil, 0, expectedHttpErr)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("List succeed", func() {
			url = url + "?direction=desc&sort=update_time&limit=1000&offset=0"
			expectedEntries := []*interfaces.Pipeline{}
			pmService.EXPECT().ListPipelines(gomock.Any(), gomock.Any()).Return(expectedEntries, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_PipelineRestHandler_UpdatePipelineStatus(t *testing.T) {
	Convey("Test UpdatePipelineStatus", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			ServerSetting: common.ServerSetting{
				CpuMax:    8,
				MemoryMax: 8096,
			},
		}

		asMock := fmock.NewMockAuthService(mockCtrl)
		pmService := fmock.NewMockPipelineMgmtService(mockCtrl)
		handler := mockNewPipelineMgmtRestHandler(appSetting, asMock, pmService)
		handler.RegisterPublic(engine)

		asMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/flow-stream-data-pipeline/v1/pipelines/a/attrs/status,status_details"

		pipelineTest := &interfaces.PipelineStatusParamter{
			Status:  interfaces.PipelineStatus_Running,
			Details: "test",
		}

		Convey("Update failed, caused by the error from method CheckPipelineExistByID", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", false, expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update failed, caused by the error from pipeline not found", func() {
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", false, nil)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("Update failed, caused by the error from method ShouldBindJSON", func() {
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", true, nil)
			expectedErr := errors.New("some errors")
			patch := ApplyMethodReturn(&gin.Context{}, "ShouldBindJSON", expectedErr)
			defer patch.Reset()

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from func ValidatePipeline", func() {
			pipelineTest := &interfaces.PipelineStatusParamter{
				Status:  "foo",
				Details: "",
			}

			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", true, nil)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Update failed, caused by the error from method UpdatePipeline", func() {
			expectedHttpErr := rest.NewHTTPError(testCtx, http.StatusInternalServerError, ferrors.StreamDataPipeline_InternalError_DataBase)
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", true, nil)
			pmService.EXPECT().UpdatePipelineStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedHttpErr)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})

		Convey("Update succeed", func() {
			pmService.EXPECT().CheckPipelineExistByID(gomock.Any(), gomock.Any()).Return("", true, nil)
			pmService.EXPECT().UpdatePipelineStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(pipelineTest)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}
