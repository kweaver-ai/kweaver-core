package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/coderunner.go -destination ../tests/mock_drivenadapters/coderunner_mock.go

// CodeRunner method interface
type CodeRunner interface {
	// RunPyCode 运行python代码
	RunPyCode(ctx context.Context, code string, inputParams, outputParams []map[string]any) (res interface{}, err error)
	// AsyncRunPyCode 异步运行python代码
	AsyncRunPyCode(ctx context.Context, code string, inputParams, outputParams []map[string]any, callback string) (res interface{}, err error)
	// CreateFile 创建文件
	CreateFile(ctx context.Context, params CreateFileReq) (string, error)
	// UpdateFile 更新文件
	UpdateFile(ctx context.Context, params UpdateFileReq) (string, error)
	// RecognizeTextByBuildIn 内置ocr识别文字
	RecognizeTextByBuildIn(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error)
	// RecognizeTextByExternal 外置ocr识别文字
	RecognizeTextByExternal(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error)
	// GetRecognizitionResult 获取外置ocr识别结果ss
	GetRecognizitionResult(ctx context.Context, taskID, recType string) (int, map[string]interface{}, error)
	// DeleteRecognizeTask 批量删除ocr任务
	DeleteRecognizeTask(ctx context.Context, taskIDs []string) error
	// ExtractTags 提取标签
	ExtractTags(ctx context.Context, content string, rule interface{}) ([]string, error)
}

type coderunner struct {
	// CodeRunner私有地址
	crPrivateAddress string
	// 数据流工具私有地址
	dftPrivateAddress string
	httpClient        otelHttp.HTTPClient
}

var (
	cOnce           sync.Once
	c               CodeRunner
	anyshareAddress string
)

// CreateFileReq 创建文件接口请求体
type CreateFileReq struct {
	FileType   string      `json:"type"`
	Name       string      `json:"name"`
	Docid      string      `json:"docid"`
	Ondup      int         `json:"ondup"`
	NewType    string      `json:"new_type,omitempty"`
	SourceType string      `json:"source_type"`
	Content    interface{} `json:"content"`
}

// UpdateFileReq 更新文件接口请求体
type UpdateFileReq struct {
	FileType   string      `json:"type"`
	DocID      string      `json:"docid"`
	NewType    string      `json:"new_type,omitempty"`
	InsertType string      `json:"insert_type"`
	InsertPos  int         `json:"insert_pos,omitempty"`
	Content    interface{} `json:"content"`
}

// NewCodeRunner 创建获取用户服务
func NewCodeRunner() CodeRunner {
	cOnce.Do(func() {
		config := common.NewConfig()
		anyshareAddress = fmt.Sprintf("%s://%s:%v", config.AccessAddress.Schema, config.AccessAddress.Host, config.AccessAddress.Port)
		c = &coderunner{
			crPrivateAddress:  fmt.Sprintf("http://%s:%v", config.CodeRunner.PrivateHost, config.CodeRunner.PrivatePort),
			dftPrivateAddress: fmt.Sprintf("http://%s:%v", config.DataFlowTools.PrivateHost, config.DataFlowTools.PrivatePort),
			httpClient:        NewOtelHTTPClient(),
		}
	})
	return c
}

func convertSliceMapValuesToJSON(data []map[string]any) ([]map[string]string, error) {
	result := make([]map[string]string, len(data))

	for i, m := range data {
		convertedMap := make(map[string]string)
		for key, value := range m {
			// 检查值类型
			if str, ok := value.(string); ok {
				// 已经是字符串，直接使用
				convertedMap[key] = str
			} else {
				// 非字符串类型，转换为JSON字符串
				jsonValue, err := json.Marshal(value)
				if err != nil {
					return nil, fmt.Errorf("error converting key %s at index %d: %v", key, i, err)
				}
				convertedMap[key] = string(jsonValue)
			}
		}
		result[i] = convertedMap
	}

	return result, nil
}

// RunPyCode 执行py代码
func (c *coderunner) RunPyCode(ctx context.Context, code string, inputParams, outputParams []map[string]any) (res interface{}, err error) {
	target := fmt.Sprintf("%s/api/coderunner/v1/pycode/run-by-params", c.crPrivateAddress)
	jsonMap, _ := convertSliceMapValuesToJSON(inputParams)
	body := map[string]interface{}{"code": code, "input_params": jsonMap, "output_params": outputParams}
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	c.setHeader(ctx, headers)
	_, respParam, err := c.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RunPyCode failed: %v, url: %v", err, target)
		return nil, err
	}

	result := respParam.(map[string]interface{})

	return result, nil
}

// AsyncRunPyCode 执行py代码
func (c *coderunner) AsyncRunPyCode(ctx context.Context, code string, inputParams, outputParams []map[string]any, callback string) (res interface{}, err error) {
	target := fmt.Sprintf("%s/api/coderunner/v1/pycode/async-run-by-params", c.crPrivateAddress)
	jsonMap, _ := convertSliceMapValuesToJSON(inputParams)
	body := map[string]interface{}{"code": code, "input_params": jsonMap, "output_params": outputParams, "callback": callback}
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	c.setHeader(ctx, headers)
	_, respParam, err := c.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RunPyCode failed: %v, url: %v", err, target)
		return nil, err
	}

	// result := respParam.(map[string]interface{})

	return respParam, nil
}

// CreateFile 创建文件
func (c *coderunner) CreateFile(ctx context.Context, params CreateFileReq) (string, error) {
	target := fmt.Sprintf("%s/api/coderunner/v1/documents/file", c.dftPrivateAddress)
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	c.setHeader(ctx, headers)
	_, respParam, err := c.httpClient.Post(ctx, target, headers, params)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("CreateFile failed: %v, url: %v", err, target)
		return "", err
	}
	respMap := respParam.(map[string]interface{})
	docID := respMap["docid"].(string)
	return docID, nil
}

// UpdateFile 更新文件
func (c *coderunner) UpdateFile(ctx context.Context, params UpdateFileReq) (string, error) {
	target := fmt.Sprintf("%s/api/coderunner/v1/documents/content", c.dftPrivateAddress)
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	c.setHeader(ctx, headers)
	_, respParam, err := c.httpClient.Put(ctx, target, headers, params)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("UpdateFile failed: %v, url: %v", err, target)
		return "", err
	}
	respMap := respParam.(map[string]interface{})
	docID := respMap["docid"].(string)
	return docID, nil
}

// RecognizeTextByBuildIn 内置ocr识别文字
func (c *coderunner) RecognizeTextByBuildIn(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	target := fmt.Sprintf("%s/api/coderunner/v1/built-in/ocr/task", c.dftPrivateAddress)
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	c.setHeader(ctx, headers)
	_, respParam, err := c.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RecognizeTextByBuildIn failed: %v, url: %v", err, target)
		return nil, err
	}
	respMap := respParam.(map[string]interface{})
	return respMap, nil
}

// RecognizeTextByExternal 外置ocr识别文字
func (c *coderunner) RecognizeTextByExternal(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	target := fmt.Sprintf("%s/api/coderunner/v1/external/ocr/task", c.dftPrivateAddress)
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	c.setHeader(ctx, headers)
	_, respParam, err := c.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RecognizeTextByExternal failed: %v, url: %v", err, target)
		return nil, err
	}
	respMap := respParam.(map[string]interface{})
	return respMap, nil
}

// GetRecognizitionResult 获取外置ocr识别结果
func (c *coderunner) GetRecognizitionResult(ctx context.Context, taskID, recType string) (int, map[string]interface{}, error) { //nolint
	target := fmt.Sprintf("%s/api/coderunner/v1/external/ocr/result?task_id=%s&rec_type=%s", c.dftPrivateAddress, taskID, recType)
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	c.setHeader(ctx, headers)
	code, respParam, err := c.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RecognizeTextByExternal failed: %v, url: %v", err, target)
		return code, nil, err
	}
	if code == http.StatusAccepted {
		return code, nil, nil
	}
	respMap := respParam.(map[string]interface{})
	return code, respMap, nil
}

// DeleteRecognizeTask 批量删除ocr任务
func (c *coderunner) DeleteRecognizeTask(ctx context.Context, taskIDs []string) error {
	target := fmt.Sprintf("%s/api/coderunner/v1/external/ocr/task/%s", c.dftPrivateAddress, strings.Join(taskIDs, ","))
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	c.setHeader(ctx, headers)
	_, err := c.httpClient.Delete(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("DeleteRecognizeTask failed: %v, url: %v", err, target)
		return err
	}
	return nil
}

// ExtractTags 提取标签
func (c *coderunner) ExtractTags(ctx context.Context, content string, rule interface{}) ([]string, error) {
	target := fmt.Sprintf("%s/api/coderunner/v1/tag/extraction", c.dftPrivateAddress)
	headers := map[string]string{"Content-Type": "application/json;charset=UTF-8"}
	c.setHeader(ctx, headers)
	body := map[string]interface{}{
		"target": map[string]string{
			"content": content,
		},
		"rules": rule,
	}
	_, respParam, err := c.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RecognizeTextByExternal failed: %v, url: %v", err, target)
		return nil, err
	}
	respMap := respParam.(map[string]interface{})
	respTags := respMap["tags"].([]interface{})
	tags := []string{}
	for _, t := range respTags {
		tags = append(tags, t.(string))
	}
	return tags, nil
}

func (c *coderunner) setHeader(ctx context.Context, headers map[string]string) {
	token := ctx.Value(common.Authorization)
	if tokenStr, ok := token.(string); ok {
		headers[common.Authorization] = tokenStr
	} else {
		headers[common.Authorization] = ""
	}

	config := common.NewConfig()
	anyshareAddress = fmt.Sprintf("%s://%s:%v%s", config.AccessAddress.Schema, config.AccessAddress.Host, config.AccessAddress.Port, config.AccessAddress.Path)

	headers[common.AnyshareAddress] = anyshareAddress
}
