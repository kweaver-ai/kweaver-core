// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/gin-gonic/gin"
	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func MockNewBKNRestHandler(
	appSetting *common.AppSetting,
	as interfaces.AuthService,
	kns interfaces.KNService,
	bs interfaces.BKNService,
) *restHandler {
	return &restHandler{
		appSetting: appSetting,
		as:         as,
		kns:        kns,
		bs:         bs,
	}
}

// newValidBKNTar 生成一个最小合法的 BKN tar，用于测试
func newValidBKNTar(t *testing.T) []byte {
	net := &bknsdk.BknNetwork{
		BknNetworkFrontmatter: bknsdk.BknNetworkFrontmatter{
			Type:    "network",
			ID:      "test-net",
			Name:    "Test Network",
			Version: "1.0.0",
		},
	}
	var buf bytes.Buffer
	if err := bknsdk.WriteNetworkToTar(net, &buf); err != nil {
		t.Fatalf("failed to create test BKN tar: %v", err)
	}
	return buf.Bytes()
}

// newMultipartRequestWithContentType 构造指定 Content-Type 的文件上传请求（用于测试扩展名校验分支）
func newMultipartRequestWithContentType(t *testing.T, url, filename, contentType string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	h.Set("Content-Type", contentType)
	fw, err := mw.CreatePart(h)
	if err != nil {
		t.Fatalf("failed to create form part: %v", err)
	}
	_, _ = fw.Write(content)
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, url, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

// newMultipartRequest 构造一个包含文件的 multipart/form-data 请求
func newMultipartRequest(t *testing.T, url, filename string, content []byte) *http.Request {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, _ = fw.Write(content)
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, url, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func Test_BKNRestHandler_UploadBKN(t *testing.T) {
	Convey("Test BKNHandler UploadBKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		bs := bmock.NewMockBKNService(mockCtrl)

		handler := MockNewBKNRestHandler(appSetting, as, kns, bs)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/bkns"

		Convey("Success with business domain header\n", func() {
			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", nil)

			req := newMultipartRequest(t, url, "test.tar", newValidBKNTar(t))
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Success without business domain header (optional)\n", func() {
			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", nil)

			req := newMultipartRequest(t, url, "test.tar", newValidBKNTar(t))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Failed when no file uploaded\n", func() {
			req := httptest.NewRequest(http.MethodPost, url, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when file has invalid extension\n", func() {
			var body bytes.Buffer
			mw := multipart.NewWriter(&body)
			fw, _ := mw.CreateFormFile("file", "test.json")
			_, _ = fw.Write([]byte(`{"invalid": "content"}`))
			mw.Close()

			req := httptest.NewRequest(http.MethodPost, url, &body)
			req.Header.Set("Content-Type", mw.FormDataContentType())
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when tar content is invalid\n", func() {
			req := newMultipartRequest(t, url, "test.tar", []byte("this is not a valid tar"))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed when CreateKN returns error\n", func() {
			err := &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_KnowledgeNetwork_InternalError,
				},
			}
			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", err)

			req := newMultipartRequest(t, url, "test.tar", newValidBKNTar(t))
			req.Header.Set(interfaces.HTTP_HEADER_BUSINESS_DOMAIN, "domain1")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_BKNRestHandler_DownloadBKN(t *testing.T) {
	Convey("Test BKNHandler DownloadBKN\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		bs := bmock.NewMockBKNService(mockCtrl)

		handler := MockNewBKNRestHandler(appSetting, as, kns, bs)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		Convey("Success downloading BKN tar\n", func() {
			bs.EXPECT().ExportToTar(gomock.Any(), "kn1", interfaces.MAIN_BRANCH).Return([]byte("tar-content"), nil)

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/v1/bkns/kn1", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
			So(w.Result().Header.Get("Content-Type"), ShouldEqual, "application/octet-stream")
		})

		Convey("Failed when ExportToTar returns error\n", func() {
			bs.EXPECT().ExportToTar(gomock.Any(), "kn1", gomock.Any()).Return(nil, &rest.HTTPError{
				HTTPCode: http.StatusInternalServerError,
				Language: rest.DefaultLanguage,
				BaseError: rest.BaseError{
					ErrorCode: berrors.BknBackend_KnowledgeNetwork_InternalError,
				},
			})

			req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/v1/bkns/kn1", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
	})
}

func Test_BKNRestHandler_UploadBKN_AuthFail(t *testing.T) {
	Convey("Test BKNHandler UploadBKN returns 401 when auth fails\n", t, func() {
		test := setGinMode()
		defer test()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		engine := gin.New()
		engine.Use(gin.Recovery())

		as := bmock.NewMockAuthService(mockCtrl)
		handler := MockNewBKNRestHandler(&common.AppSetting{}, as, nil, nil)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, errors.New("invalid token"))

		req := newMultipartRequest(t, "/api/bkn-backend/v1/bkns", "test.tar", newValidBKNTar(t))
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusUnauthorized)
	})
}

func Test_BKNRestHandler_UploadBKN_ExtensionCheck(t *testing.T) {
	Convey("Test BKNHandler UploadBKN extension validation (non-octet-stream content type)\n", t, func() {
		test := setGinMode()
		defer test()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		engine := gin.New()
		engine.Use(gin.Recovery())

		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)
		handler := MockNewBKNRestHandler(&common.AppSetting{}, as, kns, nil)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/bkns"

		Convey("Failed when invalid extension with non-octet-stream content type\n", func() {
			req := newMultipartRequestWithContentType(t, url, "test.json", "application/json", []byte("content"))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Success when .tgz extension with non-octet-stream content type\n", func() {
			kns.EXPECT().CreateKN(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("kn1", nil)

			req := newMultipartRequestWithContentType(t, url, "test.tgz", "application/gzip", newValidBKNTar(t))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
