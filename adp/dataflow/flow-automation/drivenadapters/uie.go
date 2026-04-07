package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

type Uie interface {
	// ListLog 列举训练日志记录
	ListTrainLog(ctx context.Context) ([]TrainLog, error)
	// GetDataSetTemplate 获取训练数据模板
	GetDataSetTemplate(ctx context.Context) (int, error)
	// StartTrainModule 模型训练
	StartTrainModule(ctx context.Context, con *[]byte) (int, error)
	// StartInfer 实体抽取
	StartInfer(ctx context.Context, schema interface{}, texts ...string) (map[string]interface{}, error)
}

type uie struct {
	privateURL string
	httpClient otelHttp.HTTPClient
}

// TrainLog 训练日志信息结构体
type TrainLog struct {
	ID     int    `json:"id"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Status string `json:"status"`
}

var (
	uieOnce  sync.Once
	uHandler Uie
)

func NewUie() Uie {
	uieOnce.Do(func() {
		config := common.NewConfig()
		uHandler = &uie{
			privateURL: fmt.Sprintf("http://%s:%v", config.Uie.PrivateHost, config.Uie.PrivatePort),
			httpClient: otelHttp.NewOtelHttpClient(),
		}
	})

	return uHandler
}

// ListLog 列举训练日志记录
func (ui *uie) ListTrainLog(ctx context.Context) ([]TrainLog, error) {
	target := fmt.Sprintf("%s/api/uie/train_log", ui.privateURL)
	header := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	_, resp, err := ui.httpClient.Get(ctx, target, header)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[ListLog] Request failed, detail: %s", err.Error())
		return nil, err
	}

	var trainLogs []TrainLog
	resMap := resp.(map[string]interface{})
	resByte, err := json.Marshal(resMap["res"])
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[ListLog] Marshal failed, detail: %s", err.Error())
		return nil, err
	}
	err = json.Unmarshal(resByte, &trainLogs)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[ListLog] Unmarshal failed, detail: %s", err.Error())
		return nil, err
	}

	return trainLogs, nil
}

// GetDataSetTemplate 获取训练数据模板
func (ui *uie) GetDataSetTemplate(ctx context.Context) (int, error) {
	target := fmt.Sprintf("%s/api/uie/get_dataset_template", ui.privateURL)
	header := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	respCode, _, err := ui.httpClient.Get(ctx, target, header)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[GetDataSetTemplate] Request failed, detail: %s", err.Error())
	}
	return respCode, err
}

// StartTrainModule 模型训练
func (ui *uie) StartTrainModule(ctx context.Context, con *[]byte) (int, error) {
	var taskID int
	target := fmt.Sprintf("%s/api/uie/train", ui.privateURL)
	fileName := strings.ReplaceAll(uuid.New().String(), "-", "")
	payload, writer, err := utils.StructFormData("data", fileName, con)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[StartTrainModule] StructFormData failed, detail: %s", err.Error())
		return taskID, err
	}

	var headers = make(map[string]string, 0)
	headers["Content-Type"] = writer.FormDataContentType()
	headers["Accept"] = " */*"

	_, resp, err := ui.httpClient.Post(ctx, target, headers, payload.Bytes())
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[StartTrainModule] Request failed, detail: %s", err.Error())
		return taskID, err
	}

	resMap := resp.(map[string]interface{})
	taskIDMap := resMap["res"].(map[string]interface{})
	if _, ok := taskIDMap["task_id"]; !ok {
		if _, tok := taskIDMap["error"]; !tok {
			return taskID, fmt.Errorf("%v", taskIDMap)
		}
		// 如果是任务正在运行错误，则taksID使用-1标识
		if strings.HasSuffix(taskIDMap["error"].(string), "is running!") {
			return -1, nil
		}
		return taskID, fmt.Errorf("%v", taskIDMap)
	}
	taskID = int(taskIDMap["task_id"].(float64))
	return taskID, err
}

// StartInfer 实体抽取
func (ui *uie) StartInfer(ctx context.Context, schema interface{}, texts ...string) (map[string]interface{}, error) {
	target := fmt.Sprintf("%s/api/uie/infer", ui.privateURL)
	header := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	body := map[string]interface{}{"schema": schema, "texts": texts}
	_, resp, err := ui.httpClient.Post(ctx, target, header, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[StartInfer] Request failed, detail: %s", err.Error())
		return nil, err
	}

	return resp.(map[string]interface{}), nil
}
