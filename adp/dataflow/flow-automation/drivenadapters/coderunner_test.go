package drivenadapters

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func NewMockCodeRunner(clients *HttpClientMock) CodeRunner {
	InitARLog()
	return &coderunner{
		crPrivateAddress:  "http://localhost:8080",
		dftPrivateAddress: "http://localhost:8081",
		httpClient:        clients.httpClient,
	}
}

func TestRunPyCode(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cr := NewMockCodeRunner(httpClient)
	ctx := context.WithValue(context.Background(), common.Authorization, "Bearer token")

	Convey("TestRunPyCode", t, func() {
		mockResp := map[string]interface{}{"result": "ok"}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		res, err := cr.RunPyCode(ctx, "print(1)", []map[string]any{{"a": 1}}, []map[string]any{{"b": 2}})
		assert.Equal(t, err, nil)
		assert.Equal(t, res.(map[string]interface{})["result"], "ok")
	})
}

func TestAsyncRunPyCode(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cr := NewMockCodeRunner(httpClient)
	ctx := context.WithValue(context.Background(), common.Authorization, "Bearer token")

	Convey("TestAsyncRunPyCode", t, func() {
		mockResp := map[string]interface{}{"task_id": "123"}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		res, err := cr.AsyncRunPyCode(ctx, "print(1)", nil, nil, "http://callback")
		assert.Equal(t, err, nil)
		assert.Equal(t, res.(map[string]interface{})["task_id"], "123")
	})
}

func TestCreateFile(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cr := NewMockCodeRunner(httpClient)
	ctx := context.WithValue(context.Background(), common.Authorization, "Bearer token")

	Convey("TestCreateFile", t, func() {
		mockResp := map[string]interface{}{"docid": "file_id"}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		docID, err := cr.CreateFile(ctx, CreateFileReq{Name: "test.txt"})
		assert.Equal(t, err, nil)
		assert.Equal(t, docID, "file_id")
	})
}

func TestUpdateFile(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cr := NewMockCodeRunner(httpClient)
	ctx := context.WithValue(context.Background(), common.Authorization, "Bearer token")

	Convey("TestUpdateFile", t, func() {
		mockResp := map[string]interface{}{"docid": "file_id"}
		httpClient.httpClient.EXPECT().Put(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		docID, err := cr.UpdateFile(ctx, UpdateFileReq{DocID: "file_id"})
		assert.Equal(t, err, nil)
		assert.Equal(t, docID, "file_id")
	})
}

func TestRecognizeTextByBuildIn(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cr := NewMockCodeRunner(httpClient)
	ctx := context.WithValue(context.Background(), common.Authorization, "Bearer token")

	Convey("TestRecognizeTextByBuildIn", t, func() {
		mockResp := map[string]interface{}{"text": "parsed"}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		res, err := cr.RecognizeTextByBuildIn(ctx, map[string]interface{}{"file": "f"})
		assert.Equal(t, err, nil)
		assert.Equal(t, res["text"], "parsed")
	})
}

func TestRecognizeTextByExternal(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cr := NewMockCodeRunner(httpClient)
	ctx := context.WithValue(context.Background(), common.Authorization, "Bearer token")

	Convey("TestRecognizeTextByExternal", t, func() {
		mockResp := map[string]interface{}{"task_id": "t1"}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		res, err := cr.RecognizeTextByExternal(ctx, map[string]interface{}{"file": "f"})
		assert.Equal(t, err, nil)
		assert.Equal(t, res["task_id"], "t1")
	})
}

func TestGetRecognizitionResult(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cr := NewMockCodeRunner(httpClient)
	ctx := context.WithValue(context.Background(), common.Authorization, "Bearer token")

	Convey("TestGetRecognizitionResult", t, func() {
		Convey("Accepted", func() {
			httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(http.StatusAccepted, nil, nil)
			code, res, err := cr.GetRecognizitionResult(ctx, "t1", "type")
			assert.Equal(t, err, nil)
			assert.Equal(t, code, http.StatusAccepted)
			assert.Equal(t, res, nil)
		})

		Convey("Success", func() {
			mockResp := map[string]interface{}{"result": "done"}
			httpClient.httpClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).Return(200, mockResp, nil)
			code, res, err := cr.GetRecognizitionResult(ctx, "t1", "type")
			assert.Equal(t, err, nil)
			assert.Equal(t, code, 200)
			assert.Equal(t, res["result"], "done")
		})
	})
}

func TestDeleteRecognizeTask(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cr := NewMockCodeRunner(httpClient)
	ctx := context.WithValue(context.Background(), common.Authorization, "Bearer token")

	Convey("TestDeleteRecognizeTask", t, func() {
		httpClient.httpClient.EXPECT().Delete(ctx, gomock.Any(), gomock.Any()).Return(nil, nil)
		err := cr.DeleteRecognizeTask(ctx, []string{"t1"})
		assert.Equal(t, err, nil)
	})
}

func TestExtractTags(t *testing.T) {
	httpClient := NewHttpClientMock(t)
	cr := NewMockCodeRunner(httpClient)
	ctx := context.WithValue(context.Background(), common.Authorization, "Bearer token")

	Convey("TestExtractTags", t, func() {
		mockResp := map[string]interface{}{"tags": []interface{}{"tag1", "tag2"}}
		httpClient.httpClient.EXPECT().Post(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(200, mockResp, nil)

		res, err := cr.ExtractTags(ctx, "content", "rules")
		assert.Equal(t, err, nil)
		assert.Equal(t, len(res), 2)
		assert.Equal(t, res[0], "tag1")
	})
}
