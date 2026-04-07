package drivenadapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	traceCommon "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/common"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils/ptr"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func InitARLog() {
	if commonLog.NewLogger() == nil {
		logout := "1"
		logDir := "/var/log/contentAutoMation/ut"
		logName := "contentAutoMation.log"
		commonLog.InitLogger(logout, logDir, logName)
	}
	traceLog.InitARLog(&traceCommon.TelemetryConf{LogLevel: "all"})
}

func NewMockAgentOperatorIntegration(clients *HttpClientMock) AgentOperatorIntegration {
	InitARLog()
	return &agentOperatorIntegration{
		privateURL: "http://localhost:8080",
		httpClient: clients.httpClient,
	}
}

func TestGetOperatorInfo(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	oper := NewMockAgentOperatorIntegration(httpClient)
	user := &UserInfo{}

	Convey("TestGetOperatorInfo", t, func() {
		Convey("Get Operator Info Error", func() {
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := oper.GetOperatorInfo(context.Background(), "id", "version", "", user)
			assert.NotEqual(t, err, nil)
		})
		Convey("Get Operator Info Success", func() {
			mockResp := map[string]interface{}{
				"name":        "name",
				"operator_id": "operator_id",
				"version":     "version",
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			info, err := oper.GetOperatorInfo(context.Background(), "id", "version", "", user)
			assert.Equal(t, err, nil)
			assert.Equal(t, info.Name, "name")
			assert.Equal(t, info.OperatorID, "operator_id")
			assert.Equal(t, info.Version, "version")
		})
	})
}

func TestLatestOperatorInfo(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	oper := NewMockAgentOperatorIntegration(httpClient)
	user := &UserInfo{}

	Convey("TestLatestOperatorInfo", t, func() {
		Convey("Latest Operator Info Error", func() {
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := oper.LatestOperatorInfo(context.Background(), "operator_id", "", user)
			assert.NotEqual(t, err, nil)
		})
		Convey("Latest Operator Info Success", func() {
			mockResp := map[string]interface{}{
				"name":        "name",
				"operator_id": "operator_id",
				"version":     "version",
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			info, err := oper.LatestOperatorInfo(context.Background(), "operator_id", "", user)
			assert.Equal(t, err, nil)
			assert.Equal(t, info.Name, "name")
			assert.Equal(t, info.OperatorID, "operator_id")
			assert.Equal(t, info.Version, "version")
		})
	})
}

func TestRegisterOperator(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	oper := NewMockAgentOperatorIntegration(httpClient)

	data := &RegisterOperatorReq{}
	user := &UserInfo{}

	Convey("TestRegisterOperator", t, func() {
		Convey("Register Operator Error", func() {
			httpClient.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := oper.RegisterOperator(context.Background(), data, user)
			assert.NotEqual(t, err, nil)
		})
		Convey("Register Operator Success", func() {
			mockResp := []interface{}{
				map[string]interface{}{
					"operator_id": "operator_id",
					"version":     "version",
					"status":      "published",
					"error":       map[string]interface{}{},
				},
			}
			httpClient.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			resp, err := oper.RegisterOperator(context.Background(), data, user)
			assert.Equal(t, err, nil)
			assert.Equal(t, len(resp), 1)
		})
	})
}

func TestUpdateOperator(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	oper := NewMockAgentOperatorIntegration(httpClient)

	data := &UpdateOperatorReq{}
	user := &UserInfo{}

	Convey("TestUpdateOperator", t, func() {
		Convey("Update Operator Error", func() {
			httpClient.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := oper.UpdateOperator(context.Background(), data, user)
			assert.NotEqual(t, err, nil)
		})
		Convey("Update Operator Success", func() {
			mockResp := []interface{}{
				map[string]interface{}{
					"operator_id": "operator_id",
					"version":     "version",
					"status":      "published",
					"error":       map[string]interface{}{},
				},
			}
			httpClient.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			resp, err := oper.UpdateOperator(context.Background(), data, user)
			assert.Equal(t, err, nil)
			assert.Equal(t, len(resp), 1)
		})
	})
}

func TestOperatorList(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	oper := NewMockAgentOperatorIntegration(httpClient)

	user := &UserInfo{}

	query := &QueryParams{
		Page:     ptr.Int64(0),
		PageSize: ptr.Int64(10),
	}

	Convey("TestOperatorList", t, func() {
		Convey("OperatorList Error", func() {
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("error"))
			_, err := oper.OperatorList(context.Background(), query, user)
			assert.NotEqual(t, err, nil)
		})
		Convey("OperatorList Success", func() {
			mockResp := map[string]interface{}{
				"total":     1,
				"page":      0,
				"page_size": 10,
				"data": []interface{}{
					map[string]interface{}{
						"operator_id": "operator_id",
						"version":     "version",
						"status":      "published",
					},
				},
			}
			httpClient.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			resp, err := oper.OperatorList(context.Background(), query, user)
			assert.Equal(t, err, nil)
			assert.Equal(t, resp.Total, int64(1))
			assert.Equal(t, resp.Data[0].OperatorID, "operator_id")
		})
	})
}
