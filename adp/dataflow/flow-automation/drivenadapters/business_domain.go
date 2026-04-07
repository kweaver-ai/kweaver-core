package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/google/go-querystring/query"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

// BusinessDomain 业务域接口
type BusinessDomain interface {
	BindResourceInternal(ctx context.Context, params BizDomainResourceParams) error
	UnBindResourceInternal(ctx context.Context, params BizDomainResourceParams) error
	ListResource(ctx context.Context, params BizDomainResourceQuery, token string) (BizDomainResourceResp, error)
	CheckerResource(ctx context.Context, params BizDomainResourceParams, token string) (bool, error)
}

var bOnce sync.Once
var bd BusinessDomain

type businessDomain struct {
	baseURL    string
	httpClient otelHttp.HTTPClient
}

func NewBusinessDomain() BusinessDomain {
	bOnce.Do(func() {
		config := common.NewConfig()
		if !config.IsBusinessDomainEnabled() {
			bd = &mockBusinessDomain{}
		} else {
			bd = &businessDomain{
				baseURL:    fmt.Sprintf("http://%s:%v", config.BusinessDomain.Host, config.BusinessDomain.Port),
				httpClient: NewOtelHTTPClient(),
			}
		}
	})

	return bd
}

// BizDomainResourceParams 业务域资源绑定或解绑参数
type BizDomainResourceParams struct {
	BizDomainID  string `json:"bd_id" url:"bd_id"`
	ResourceID   string `json:"id,omitempty" url:"id,omitempty"`
	ResourceType string `json:"type,omitempty" url:"type,omitempty"`
}

// BizDomainResourceQuery 业务域资源查询参数
type BizDomainResourceQuery struct {
	BizDomainResourceParams `json:",inline"`
	Limit                   int64 `json:"-" url:"limit"`
	OffSet                  int64 `json:"-" url:"offset,omitempty"`
}

type BizDomainResources []*BizDomainResourceParams

func (b BizDomainResources) GetIDs(dtype string) (res []string) {
	for _, v := range b {
		vs := strings.SplitN(v.ResourceID, ":", 2)

		// dtype 有值时才校验类型
		if dtype != "" {
			if len(vs) < 2 || vs[1] != dtype {
				continue
			}
		}

		res = append(res, vs[0])
	}
	return
}

type BizDomainResourceResp struct {
	Total int64              `json:"total"`
	Items BizDomainResources `json:"items"`
}

// BindResourceInternal 业务域资源绑定
func (b *businessDomain) BindResourceInternal(ctx context.Context, params BizDomainResourceParams) error {
	target := fmt.Sprintf("%s/internal/api/business-system/v1/resource", b.baseURL)

	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	_, _, err := b.httpClient.Post(ctx, target, headers, params)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.BindResourceInternal] Post failed: %v, url: %v", err, target)
		return err
	}

	return nil
}

// UnBindResourceInternal 业务域资源解绑
func (b *businessDomain) UnBindResourceInternal(ctx context.Context, params BizDomainResourceParams) error {
	v, err := query.Values(params)
	if err != nil {
		return err
	}

	target := fmt.Sprintf("%s/internal/api/business-system/v1/resource?%s", b.baseURL, v.Encode())
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	_, err = b.httpClient.Delete(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.UnBindResourceInternal] Post failed: %v, url: %v", err, target)
		return err
	}

	return nil
}

// ListResource 列举当前业务域下已注册的资源列表
func (b *businessDomain) ListResource(ctx context.Context, params BizDomainResourceQuery, token string) (BizDomainResourceResp, error) {
	res := BizDomainResourceResp{}

	v, err := query.Values(params)
	if err != nil {
		return res, err
	}

	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"Authorization": utils.IfNot(strings.HasPrefix(token, "Bearer "), token, fmt.Sprintf("Bearer %s", token)),
	}

	target := fmt.Sprintf("%s/api/business-system/v1/resource?%s", b.baseURL, v.Encode())
	_, resp, err := b.httpClient.Get(ctx, target, headers)
	if err != nil {
		return res, err
	}

	respByte, _ := json.Marshal(resp)

	err = json.Unmarshal(respByte, &res)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[drivenadapters.ListResource] Unmarshal faild, detail: %s", err)
		return res, err
	}

	return res, nil
}

// CheckerResource 检查当前业务域是否存在指定资源
func (b *businessDomain) CheckerResource(ctx context.Context, params BizDomainResourceParams, token string) (bool, error) {
	res, err := b.ListResource(ctx, BizDomainResourceQuery{
		BizDomainResourceParams: params,
	}, token)
	if err != nil {
		return false, err
	}

	if res.Total > 0 {
		return true, nil
	}

	return false, nil
}
