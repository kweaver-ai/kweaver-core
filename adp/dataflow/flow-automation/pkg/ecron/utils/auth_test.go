package utils

import (
	"errors"
	"fmt"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/mock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newOAuthClient(h HTTPClient) *oauth {
	return &oauth{
		publicAddress: "",
		adminAddress:  "",
		httpClient:    h,
		timer:         nil,
	}
}

func TestAuthInit(t *testing.T) {
	// Reset logger state before test
	logMutex.Lock()
	logHandle.logger = nil
	logMutex.Unlock()

	t.Setenv("LOGOUT", "1")
	Convey("init", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := mock.NewMockHTTPClient(ctrl)

		o := newOAuthClient(h)
		assert.NotEqual(t, o, nil)

		// Test that Release doesn't panic when timer is nil
		o.Release()
	})
}

func TestAuthRequestToken(t *testing.T) {
	Convey("RequestToken", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("request success", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			// Mock the Post method to return success response
			mockHTTP.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(url string, headers map[string]string, reqParam interface{}, respParam interface{}) error {
					data := make(map[string]interface{})
					data["access_token"] = "123456"
					data["expires_in"] = 3599
					body, _ := jsoniter.Marshal(data)
					return jsoniter.Unmarshal(body, respParam)
				})

			token, duration, err := o.RequestToken("123", "456", []string{"789"})
			assert.Equal(t, token, "123456")
			assert.Equal(t, duration, time.Duration(1e9*(3599-10)))
			assert.Equal(t, err, nil)
		})

		Convey("request failed", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			// Mock the Post method to return error
			mockHTTP.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("failed"))

			_, _, err := o.RequestToken("123", "456", []string{"789"})
			assert.NotEqual(t, err, nil)
		})
	})
}

func TestAuthVerifyToken(t *testing.T) {
	Convey("VerifyToken", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("authOn is true, o.token is empty", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			_, err := o.VerifyToken("")
			assert.Equal(t, err.Cause, common.ErrTokenEmpty)
		})

		Convey("the token parameter is invalid", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			_, err := o.VerifyToken("123")
			assert.Equal(t, err.Cause, common.ErrInvalidToken)
			assert.Equal(t, err.Code, common.Unauthorized)
		})

		Convey("http post err", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			// Mock Post to return error
			mockHTTP.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("failed"))

			_, err := o.VerifyToken(fmt.Sprintf("%v %v", common.Bearer, "123"))
			assert.Equal(t, err.Cause, "failed")
		})

		Convey("token active false", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			// Mock Post to return inactive token
			mockHTTP.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(url string, headers map[string]string, reqParam interface{}, respParam interface{}) error {
					data := map[string]interface{}{
						"active": false,
					}
					body, _ := jsoniter.Marshal(data)
					return jsoniter.Unmarshal(body, respParam)
				})

			_, err := o.VerifyToken(fmt.Sprintf("%v %v", common.Bearer, "123"))
			assert.Equal(t, err.Cause, common.ErrTokenExpired)
			assert.Equal(t, err.Code, common.Unauthorized)
		})

		Convey("no client id", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			// Mock Post to return active but no client_id
			mockHTTP.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(url string, headers map[string]string, reqParam interface{}, respParam interface{}) error {
					data := map[string]interface{}{
						"active": true,
					}
					body, _ := jsoniter.Marshal(data)
					return jsoniter.Unmarshal(body, respParam)
				})

			_, err := o.VerifyToken(fmt.Sprintf("%v %v", common.Bearer, "123"))
			assert.Equal(t, err.Cause, common.ErrInvalidToken)
			assert.Equal(t, err.Code, common.Unauthorized)
		})

		Convey("pass", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			// Mock Post to return valid token with client_id
			mockHTTP.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(url string, headers map[string]string, reqParam interface{}, respParam interface{}) error {
					data := map[string]interface{}{
						"active":    true,
						"client_id": "123456",
					}
					body, _ := jsoniter.Marshal(data)
					return jsoniter.Unmarshal(body, respParam)
				})

			visitor, err := o.VerifyToken(fmt.Sprintf("%v %v", common.Bearer, "123"))
			assert.Equal(t, err, (*common.ECronError)(nil))
			assert.Equal(t, visitor.ClientID, "123456")
		})
	})
}

func TestVerifyHydraVersion(t *testing.T) {
	Convey("VerifyHydraVersion", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		oldVerifyTokenPath := "/oauth2/introspect"
		newVerifyTokenPath := "/admin/oauth2/introspect"

		Convey("return newVerifyTokenPath", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			oldUrl := fmt.Sprintf("%s%s", o.adminAddress, oldVerifyTokenPath)
			newUrl := fmt.Sprintf("%s%s", o.adminAddress, newVerifyTokenPath)

			// Mock PostV2 to return 404 for old URL and 200 for new URL
			mockHTTP.EXPECT().PostV2(newUrl, gomock.Any(), gomock.Any()).Return(200, nil, errors.New("failed"))
			mockHTTP.EXPECT().PostV2(oldUrl, gomock.Any(), gomock.Any()).Return(404, nil, errors.New("failed")).MaxTimes(1)

			path, success := o.VerifyHydraVersion()
			assert.Equal(t, path, newVerifyTokenPath)
			assert.Equal(t, success, true)
		})

		Convey("return oldVerifyTokenPath", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			oldUrl := fmt.Sprintf("%s%s", o.adminAddress, oldVerifyTokenPath)
			newUrl := fmt.Sprintf("%s%s", o.adminAddress, newVerifyTokenPath)

			// Mock PostV2 to return 404 for new URL and 200 for old URL
			mockHTTP.EXPECT().PostV2(newUrl, gomock.Any(), gomock.Any()).Return(404, nil, errors.New("failed"))
			mockHTTP.EXPECT().PostV2(oldUrl, gomock.Any(), gomock.Any()).Return(200, nil, errors.New("failed"))

			path, success := o.VerifyHydraVersion()
			assert.Equal(t, path, oldVerifyTokenPath)
			assert.Equal(t, success, true)
		})

		Convey("return false", func() {
			mockHTTP := mock.NewMockHTTPClient(ctrl)
			o := newOAuthClient(mockHTTP)
			assert.NotEqual(t, o, nil)

			oldUrl := fmt.Sprintf("%s%s", o.adminAddress, oldVerifyTokenPath)
			newUrl := fmt.Sprintf("%s%s", o.adminAddress, newVerifyTokenPath)

			// Mock PostV2 to return 500 for both URLs
			mockHTTP.EXPECT().PostV2(newUrl, gomock.Any(), gomock.Any()).Return(500, nil, errors.New("failed"))
			mockHTTP.EXPECT().PostV2(oldUrl, gomock.Any(), gomock.Any()).Return(500, nil, errors.New("failed"))

			path, success := o.VerifyHydraVersion()
			assert.Equal(t, path, "")
			assert.Equal(t, success, false)
		})
	})
}
