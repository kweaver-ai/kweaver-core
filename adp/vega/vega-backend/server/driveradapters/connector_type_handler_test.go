// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"go.uber.org/mock/gomock"

	"vega-backend/interfaces"
	mock_interfaces "vega-backend/interfaces/mock"
)

// putUpdateConnectorType 构造 PUT /connector-types/:type 的测试上下文，驱动 handler。
func putUpdateConnectorType(t *testing.T, pathType, body string, ctsSetup func(m *mock_interfaces.MockConnectorTypeService)) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAuth := mock_interfaces.NewMockAuthService(ctrl)
	mockAuth.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).
		Return(hydra.Visitor{ID: "u1", Type: hydra.VisitorType_User}, nil).
		AnyTimes()

	mockCTS := mock_interfaces.NewMockConnectorTypeService(ctrl)
	if ctsSetup != nil {
		ctsSetup(mockCTS)
	}

	r := &restHandler{as: mockAuth, cts: mockCTS}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "type", Value: pathType}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/vega-backend/v1/connector-types/"+pathType,
		strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	r.UpdateConnectorType(c)
	return w
}

// case 3：请求体 type 与路径不一致 → 409 + TypeMismatch
func TestUpdateConnectorType_BodyTypeMismatch_409(t *testing.T) {
	body := `{"type":"postgres","name":"MySQL","mode":"local","category":"table"}`
	w := putUpdateConnectorType(t, "mysql", body, func(m *mock_interfaces.MockConnectorTypeService) {
		// 冲突应在调用任何 service 方法前返回，此处无 EXPECT。
	})

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "VegaBackend.ConnectorType.TypeMismatch") {
		t.Fatalf("expected TypeMismatch error_code, got body=%s", w.Body.String())
	}
}

// case 1：请求体 type == 路径 → 跳过冲突，进入正常更新流程
func TestUpdateConnectorType_BodyTypeMatchesPath_OK(t *testing.T) {
	body := `{"type":"mysql","name":"MySQL","mode":"local","category":"table"}`
	w := putUpdateConnectorType(t, "mysql", body, func(m *mock_interfaces.MockConnectorTypeService) {
		m.EXPECT().GetByType(gomock.Any(), "mysql").
			Return(&interfaces.ConnectorType{Type: "mysql", Name: "MySQL", Mode: "local", Category: "table"}, nil)
		// name 未变化，故不调用 CheckExistByName
		m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	})

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d, body=%s", w.Code, w.Body.String())
	}
}

// case 2：请求体省略 type → 400 InvalidParameter.Type（K8s 风格严格契约）
func TestUpdateConnectorType_BodyTypeOmitted_400(t *testing.T) {
	body := `{"name":"MySQL","mode":"local","category":"table"}`
	w := putUpdateConnectorType(t, "mysql", body, func(m *mock_interfaces.MockConnectorTypeService) {
		// 校验失败应在调用任何 service 方法前返回，此处无 EXPECT。
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "VegaBackend.ConnectorType.InvalidParameter.Type") {
		t.Fatalf("expected InvalidParameter.Type error_code, got body=%s", w.Body.String())
	}
}
