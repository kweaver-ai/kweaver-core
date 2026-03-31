package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/agent_operator_integration.go -destination ../tests/mock_drivenadapters/agent_operator_integration_mock.go

type ExtendInfo map[string]interface{}

type RetryConditions struct {
	ErrorCodes []string `json:"error_codes"`
	StatusCode []int    `json:"status_code"`
}

type RetryPolicy struct {
	BackoffFactor   int             `json:"backoff_factor"`
	InitialDelay    int             `json:"initial_delay"`
	MaxAttempts     int             `json:"max_attempts"`
	MaxDelay        int             `json:"max_delay"`
	RetryConditions RetryConditions `json:"retry_conditions"`
}

type OperatorExecuteControl struct {
	RetryPolicy RetryPolicy `json:"retry_policy"`
	Timeout     int         `json:"timeout"`
}

type OperatorInfo struct {
	Category      string `json:"category"`
	ExecutionMode string `json:"execution_mode"`
	OperatorType  string `json:"operator_type"`
	Source        string `json:"source"`
	IsDataSource  bool   `json:"is_data_source"`
}

type APISpec struct {
	Parameters   []*Parameter `json:"parameters"`    // 结构化参数
	RequestBody  any          `json:"request_body"`  // 请求体结构
	Responses    any          `json:"responses"`     // 响应结构
	Schemas      any          `json:"schemas"`       // 引用的结构体定义
	Callbacks    any          `json:"callbacks"`     // 回调函数定义
	Security     any          `json:"security"`      // 安全要求
	Tags         []string     `json:"tags"`          // 标签
	ExternalDocs any          `json:"external_docs"` // 外部文档
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // path/query/header/cookie
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Ref         string `json:"$ref,omitempty"` // 引用字段
}

type OperatorMetadata struct {
	APISpec     *APISpec `json:"api_spec"`
	CreateTime  int64    `json:"create_time"`
	CreateUser  string   `json:"create_user"`
	Description string   `json:"description"`
	ID          int      `json:"id"`
	IsDeleted   bool     `json:"is_deleted"`
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	ServerURL   string   `json:"server_url"`
	Summary     string   `json:"summary"`
	UpdateTime  int64    `json:"update_time"`
	UpdateUser  string   `json:"update_user"`
	Version     string   `json:"version"`
}

type OperatorResponse struct {
	CreateTime             int64                  `json:"create_time"`
	CreateUser             string                 `json:"create_user"`
	ExtendInfo             ExtendInfo             `json:"extend_info"`
	Metadata               OperatorMetadata       `json:"metadata"`
	MetadataType           string                 `json:"metadata_type"`
	Name                   string                 `json:"name"`
	OperatorID             string                 `json:"operator_id"`
	OperatorExecuteControl OperatorExecuteControl `json:"operator_execute_control"`
	OperatorInfo           OperatorInfo           `json:"operator_info"`
	Status                 string                 `json:"status"`
	UpdateTime             int64                  `json:"update_time"`
	UpdateUser             string                 `json:"update_user"`
	Version                string                 `json:"version"`
}

type RegisterOperatorReq struct {
	Data                 string                 `json:"data"`
	OperatorMetadataType string                 `json:"operator_metadata_type"`
	OperatorInfo         *OperatorInfo          `json:"operator_info"`
	ExtendInfo           map[string]interface{} `json:"extend_info"`
	DirectPublish        bool                   `json:"direct_publish"`
	UserToken            string                 `json:"user_token"`
	BizDomainID          string                 `json:"-"`
}

type OperatorModifyResp struct {
	Status     string         `json:"status"`
	OperatorID string         `json:"operator_id"`
	Version    string         `json:"version"`
	Error      map[string]any `json:"error"`
}

type OperatorList struct {
	Total    int64               `json:"total"`
	Page     int64               `json:"page"`
	PageSize int64               `json:"page_size"`
	Data     []*OperatorResponse `json:"data"`
}

type UpdateOperatorReq struct {
	OperatorID           string                 `json:"operator_id"`
	Data                 string                 `json:"data"`
	OperatorMetadataType string                 `json:"operator_metadata_type"`
	OperatorInfo         *OperatorInfo          `json:"operator_info"`
	ExtendInfo           map[string]interface{} `json:"extend_info"`
	DirectPublish        bool                   `json:"direct_publish"`
	UserToken            string                 `json:"user_token"`
	BizDomainID          string                 `json:"-"`
}

type QueryParams struct {
	Page       *int64  `json:"page"`
	PageSize   *int64  `json:"page_size"`
	SortBy     *string `json:"sort_by"`
	SortOrder  *string `json:"sort_order"`
	OperatorID *string `json:"operator_id"`
	Name       *string `json:"name"`
	Version    *string `json:"version"`
	Status     *string `json:"status"`
	Category   *string `json:"category"`
	UserID     *string `json:"user_id"`
}

func (q *QueryParams) ToQueryString() string {
	if q == nil {
		return ""
	}
	var query []string
	v := reflect.ValueOf(q).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.IsNil() {
			continue
		}

		jsonTag := t.Field(i).Tag.Get("json")
		if jsonTag == "" {
			continue
		}

		value := fmt.Sprintf("%v", reflect.Indirect(field).Interface())
		query = append(query, fmt.Sprintf("%s=%s", jsonTag, url.PathEscape(value)))
	}

	return strings.Join(query, "&")
}

type AgentOperatorIntegration interface {
	GetOperatorInfo(ctx context.Context, operatorID, bizDomainID string, version string, userInfo *UserInfo) (res *OperatorResponse, err error)
	LatestOperatorInfo(ctx context.Context, operatorID, bizDomainID string, userInfo *UserInfo) (res *OperatorResponse, err error)
	RegisterOperator(ctx context.Context, data *RegisterOperatorReq, userInfo *UserInfo) ([]*OperatorModifyResp, error)
	UpdateOperator(ctx context.Context, data *UpdateOperatorReq, userInfo *UserInfo) ([]*OperatorModifyResp, error)
	OperatorList(ctx context.Context, query *QueryParams, userInfo *UserInfo) (*OperatorList, error)
}

var (
	agentOperatorIntegrationOnce     sync.Once
	agentOperatorIntegrationInstance AgentOperatorIntegration
)

type agentOperatorIntegration struct {
	privateURL string
	httpClient otelHttp.HTTPClient
}

func NewAgentOperatorIntegration() AgentOperatorIntegration {
	agentOperatorIntegrationOnce.Do(func() {
		config := common.NewConfig()

		agentOperatorIntegrationInstance = &agentOperatorIntegration{
			privateURL: fmt.Sprintf("http://%s:%v", config.AgentOperatorIntegration.Host, config.AgentOperatorIntegration.Port),
			httpClient: NewOtelHTTPClient(),
		}
	})

	return agentOperatorIntegrationInstance
}

// GetOperatorInfo implements AgentOperatorIntegration.
func (a *agentOperatorIntegration) GetOperatorInfo(ctx context.Context, operatorID, version, bizDomainID string, userInfo *UserInfo) (res *OperatorResponse, err error) {
	target := fmt.Sprintf("%s/api/agent-operator-integration/internal-v1/operator/history/%s/%s", a.privateURL, operatorID, version)
	headers := map[string]string{
		"Content-Type":      "application/json;charset=UTF-8",
		"X-User":            userInfo.UserID,
		"X-Visitor-Type":    userInfo.AccountType,
		"X-Business-Domain": bizDomainID,
	}

	_, resp, err := a.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetOperatorInfo] Get failed: %v, url: %v", err, target)
		return nil, err
	}

	res = &OperatorResponse{}
	respByte, _ := json.Marshal(resp)

	err = json.Unmarshal(respByte, res)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.GetOperatorInfo] Unmarshal faild, detail: %s", err)
		return nil, err
	}

	return
}

// LatestOperatorInfo 获取指定算子id最新版本信息
func (a *agentOperatorIntegration) LatestOperatorInfo(ctx context.Context, operatorID, bizDomainID string, userInfo *UserInfo) (res *OperatorResponse, err error) {
	target := fmt.Sprintf("%s/api/agent-operator-integration/internal-v1/operator/info/%s", a.privateURL, operatorID)
	headers := map[string]string{
		"Content-Type":      "application/json;charset=UTF-8",
		"X-User":            userInfo.UserID,
		"X-Visitor-Type":    userInfo.AccountType,
		"X-Business-Domain": bizDomainID,
	}

	_, resp, err := a.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.LatestOperatorInfo] Get failed: %v, url: %v", err, target)
		return nil, err
	}

	res = &OperatorResponse{}
	respByte, _ := json.Marshal(resp)

	err = json.Unmarshal(respByte, res)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.LatestOperatorInfo] Unmarshal faild, detail: %s", err)
		return nil, err
	}

	return
}

// RegisterOperator 注册组合算子
func (a *agentOperatorIntegration) RegisterOperator(ctx context.Context, data *RegisterOperatorReq, userInfo *UserInfo) ([]*OperatorModifyResp, error) {
	target := fmt.Sprintf("%v/api/agent-operator-integration/internal-v1/operator/register", a.privateURL)
	headers := map[string]string{
		"Content-Type":      "application/json;charset=UTF-8",
		"X-User":            userInfo.UserID,
		"X-Visitor-Type":    userInfo.AccountType,
		"X-Business-Domain": data.BizDomainID,
	}

	_, res, err := a.httpClient.Post(ctx, target, headers, data)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.RegisterOperator] Post failed: %v, url: %v", err, target)
		return nil, err
	}

	var OperatorModifyResps []*OperatorModifyResp
	respByte, _ := json.Marshal(res)
	err = json.Unmarshal(respByte, &OperatorModifyResps)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.RegisterOperator] Unmarshal faild, detail: %s", err)
		return nil, err
	}

	return OperatorModifyResps, nil
}

// UpdateOperator 更新组合算子
func (a *agentOperatorIntegration) UpdateOperator(ctx context.Context, data *UpdateOperatorReq, userInfo *UserInfo) ([]*OperatorModifyResp, error) {
	target := fmt.Sprintf("%v/api/agent-operator-integration/internal-v1/operator/info/update", a.privateURL)
	headers := map[string]string{
		"Content-Type":      "application/json;charset=UTF-8",
		"X-User":            userInfo.UserID,
		"X-Visitor-Type":    userInfo.AccountType,
		"X-Business-Domain": data.BizDomainID,
	}
	_, res, err := a.httpClient.Post(ctx, target, headers, data)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.UpdateOperator] Post failed: %v, url: %v", err, target)
		return nil, err
	}

	var OperatorModifyResps []*OperatorModifyResp
	respByte, _ := json.Marshal(res)
	err = json.Unmarshal(respByte, &OperatorModifyResps)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.UpdateOperator] Unmarshal faild, detail: %s", err)
		return nil, err
	}

	return OperatorModifyResps, nil
}

func (a *agentOperatorIntegration) OperatorList(ctx context.Context, query *QueryParams, userInfo *UserInfo) (*OperatorList, error) {
	target := fmt.Sprintf("%s/api/agent-operator-integration/internal-v1/operator/market", a.privateURL)
	queryStr := query.ToQueryString()
	if queryStr != "" {
		target = fmt.Sprintf("%s?%s", target, queryStr)
	}

	headers := map[string]string{
		"Content-Type":   "application/json;charset=UTF-8",
		"X-User":         userInfo.UserID,
		"X-Visitor-Type": userInfo.AccountType,
	}

	_, resp, err := a.httpClient.Get(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.OperatorList] Get failed: %v, url: %v", err, target)
		return nil, err
	}

	var operatorList *OperatorList
	respByte, _ := json.Marshal(resp)
	err = json.Unmarshal(respByte, &operatorList)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.OperatorList] Unmarshal faild, detail: %s", err)
		return nil, err
	}

	return operatorList, nil
}
