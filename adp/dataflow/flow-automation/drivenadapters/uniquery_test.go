package drivenadapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
	mhttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/mock/mock_httpclient"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/tests/mock_httpclient"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

type HttpClientMock struct {
	httpClient      *mhttp.MockHTTPClient
	oauthHttpClient *mhttp.MockOAuth2Client
	httpClient1     *mock_httpclient.MockHTTPClient
	httpClient2     *mock_httpclient.MockHTTPClient2
}

func NewHttpClientMock(t *testing.T) *HttpClientMock {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return &HttpClientMock{
		httpClient:      mhttp.NewMockHTTPClient(ctrl),
		oauthHttpClient: mhttp.NewMockOAuth2Client(ctrl),
		httpClient1:     mock_httpclient.NewMockHTTPClient(ctrl),
		httpClient2:     mock_httpclient.NewMockHTTPClient2(ctrl),
	}
}

func NewMockUniquery(clients *HttpClientMock) UniqueryDriven {
	return &uniquery{
		baseURL:          "http://localhost:8080",
		dataViewID:       "533595539741103586",
		dataModelBaseURL: "http://localhost:8080",
		httpClient:       clients.httpClient,
	}
}

func TestUniquery_Query(t *testing.T) {
	httpClients := NewHttpClientMock(t)
	uniquery := NewMockUniquery(httpClients)

	queryRes := MetricQueryRes{
		Datas: []DataEntry{
			{
				Labels: LabelData{
					DagID:  "dagId1",
					Status: "success",
				},
				Times:  []int64{1},
				Values: []interface{}{float64(1)},
			},
		},
	}

	mockRes := map[string]interface{}{
		"datas": []map[string]interface{}{
			{
				"labels": map[string]interface{}{
					"dagId":  "dagId1",
					"status": "success",
				},
				"times":  []int64{1},
				"values": []float64{1},
			},
		},
	}

	Convey("TestUniquery_Query", t, func() {
		Convey("Query Error", func() {
			httpClients.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("query error"))
			_, err := uniquery.QueryDagStatusCount(context.Background(), 1, 2, "token", QueryFileds{DagIDs: []string{"dagId1"}})
			assert.NotEqual(t, err, nil)
		})
		Convey("Query Success", func() {
			httpClients.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockRes, nil)
			res, err := uniquery.QueryDagStatusCount(context.Background(), 1, 2, "token", QueryFileds{DagIDs: []string{"dagId1"}})
			assert.Equal(t, err, nil)
			So(res, ShouldResemble, queryRes)
		})
	})
}

func TestUniquery_QueryDagRunTimeAvg(t *testing.T) {
	httpClients := NewHttpClientMock(t)
	uniquery := NewMockUniquery(httpClients)

	queryRes := MetricQueryRes{
		Datas: []DataEntry{
			{
				Labels: LabelData{
					DagID:  "dagId1",
					Status: "success",
				},
				Times:  []int64{1},
				Values: []interface{}{1.5},
			},
		},
	}

	mockRes := map[string]interface{}{
		"datas": []map[string]interface{}{
			{
				"labels": map[string]interface{}{
					"dagId":  "dagId1",
					"status": "success",
				},
				"times":  []int64{1},
				"values": []float64{1.5},
			},
		},
	}

	Convey("TestUniquery_QueryDagRunTimeAvg", t, func() {
		Convey("Query Error", func() {
			httpClients.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("query error"))
			_, err := uniquery.QueryDagRunTimeAvg(context.Background(), 1, 2, "token", QueryFileds{DagIDs: []string{"dagId1"}})
			assert.NotEqual(t, err, nil)
		})
		Convey("Query Success", func() {
			httpClients.httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockRes, nil)
			res, err := uniquery.QueryDagRunTimeAvg(context.Background(), 1, 2, "token", QueryFileds{DagIDs: []string{"dagId1"}})
			assert.Equal(t, err, nil)
			So(res, ShouldResemble, queryRes)
		})
	})
}

func TestUniquery_CheckDataViewExist(t *testing.T) {
	httpClients := NewHttpClientMock(t)
	uniquery := NewMockUniquery(httpClients)

	mockResExist := map[string]interface{}{
		"entries": []map[string]interface{}{
			{"id": "1", "name": "view1"},
		},
	}
	mockResNotExist := map[string]interface{}{
		"entries": []map[string]interface{}{},
	}

	Convey("TestUniquery_CheckDataViewExist", t, func() {
		Convey("View Exists", func() {
			httpClients.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResExist, nil)
			ok, err := uniquery.CheckDataViewExist(context.Background(), "view1")
			assert.Equal(t, err, nil)
			assert.Equal(t, ok, true)
		})
		Convey("View Not Exist", func() {
			httpClients.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResNotExist, nil)
			ok, err := uniquery.CheckDataViewExist(context.Background(), "view2")
			assert.Equal(t, err, nil)
			assert.Equal(t, ok, false)
		})
		Convey("Get Error", func() {
			httpClients.httpClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(500, nil, fmt.Errorf("get error"))
			ok, err := uniquery.CheckDataViewExist(context.Background(), "view3")
			assert.NotEqual(t, err, nil)
			assert.Equal(t, ok, false)
		})
	})
}

func TestMetricQueryRes_Helpers(t *testing.T) {
	mqr := MetricQueryRes{
		Datas: []DataEntry{
			{
				Labels: LabelData{DagID: "dag1", Status: "success"},
				Times:  []int64{1},
				Values: []interface{}{float64(2)},
			},
			{
				Labels: LabelData{DagID: "dag1", Status: "failed"},
				Times:  []int64{2},
				Values: []interface{}{float64(3)},
			},
			{
				Labels: LabelData{DagID: "dag2", Status: "success"},
				Times:  []int64{3},
				Values: []interface{}{float64(4)},
			},
		},
	}

	t.Run("ToMap", func(t *testing.T) {
		m := mqr.ToMap()
		assert.Equal(t, m["dag1"].Success, int64(2))
		assert.Equal(t, m["dag1"].Failed, int64(3))
		assert.Equal(t, m["dag1"].Total, int64(5))
		assert.Equal(t, m["dag2"].Success, int64(4))
		assert.Equal(t, m["dag2"].Total, int64(4))
	})
	t.Run("TotalCnt", func(t *testing.T) {
		tot := mqr.TotalCnt()
		assert.Equal(t, tot.Success, int64(6))
		assert.Equal(t, tot.Failed, int64(3))
		assert.Equal(t, tot.Total, int64(9))
	})
	t.Run("GetDagIDs", func(t *testing.T) {
		ids := mqr.GetDagIDs()
		assert.Equal(t, ids, []string{"dag1", "dag1", "dag2"})
	})
	t.Run("AvgTimeToMap", func(t *testing.T) {
		avg := mqr.AvgTimeToMap()
		assert.Equal(t, avg["dag1"], 5.0) // 2+3
		assert.Equal(t, avg["dag2"], 4.0)
	})
}
