package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/http"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/authorization.go -destination ../tests/mock_drivenadapters/authorization_mock.go

// AuthorizationDriven 接口定义
type AuthorizationDriven interface {
	// CreatePolicy 创建权限策略
	CreatePolicy(ctx context.Context, params []PermPolicyParams) error
	// UpdatePolicy 更新权限策略
	UpdatePolicy(ctx context.Context, plocyIDs []string, params []PermPolicyParams) error
	// DeletePolicy 删除权限策略
	DeletePolicy(ctx context.Context, params PermPolicyDeleteParams) error
	// ResourceFilter 资源权限过滤
	ResourceFilter(ctx context.Context, params ResourceFilterParmas) ([]Resource, error)
	// ListResourceOperation 获取资源可操作列表
	ListResourceOperation(ctx context.Context, params ListResourceOperationParmas) ([]ListResourceOperationRes, error)
	// OperationPermCheck 决策指定资源是否具有操作列表权限
	OperationPermCheck(ctx context.Context, params OperationPermCheckParams) (bool, error)
	// ListResource 列举当前资源类型具有指定权限的资源列表
	ListResource(ctx context.Context, params ListResourceParams) ([]Resource, error)
}

type authorization struct {
	privateAddress string
	httpClient     otelHttp.HTTPClient
}

var authOnce sync.Once
var az AuthorizationDriven

// NewAuthorization 实例化
func NewAuthorization() AuthorizationDriven {
	authOnce.Do(func() {
		config := common.NewConfig()
		if !config.IsAuthEnabled() {
			az = &mockAuthorization{}
		} else {
			az = &authorization{
				privateAddress: fmt.Sprintf("http://%s:%v", config.AuthorizationConf.PrivateHost, config.AuthorizationConf.PrivatePort),
				httpClient:     NewOtelHTTPClient(),
			}
		}
	})
	return az
}

// Vistor 权限访问者信息
type Vistor struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// Resource 资源信息
type Resource struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// Operation 授权可操作列表
type Operation struct {
	Allow []OperationID `json:"allow"`
	Deny  []OperationID `json:"deny"`
}

// OperationID 权限操作ID
type OperationID struct {
	ID string `json:"id"`
}

// PermPolicyParams 权限策略
type PermPolicyParams struct {
	Accessor  *Vistor   `json:"accessor"`
	Resource  *Resource `json:"resource"`
	Operation Operation `json:"operation"`
	Condition string    `json:"condition,omitempty"`
	ExpiresAt string    `json:"expires_at,omitempty"`
}

// CreatePolicy 创建权限策略
func (a *authorization) CreatePolicy(ctx context.Context, params []PermPolicyParams) error {
	target := fmt.Sprintf("%s/api/authorization/v1/policy", a.privateAddress)
	_, _, err := a.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, params)
	return err
}

// UpdatePolicy 更新权限策略
func (a *authorization) UpdatePolicy(ctx context.Context, plocyIDs []string, params []PermPolicyParams) error {
	target := fmt.Sprintf("%s/api/authorization/v1/policy/%s", a.privateAddress, strings.Join(plocyIDs, ","))
	_, _, err := a.httpClient.Put(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, params)
	return err
}

// PermPolicyDelete 删除权限策略请求参数
type PermPolicyDeleteParams struct {
	Method    string     `json:"method"`
	Resources []Resource `json:"resources"`
}

// DeletePolicy 删除权限策略
func (a *authorization) DeletePolicy(ctx context.Context, params PermPolicyDeleteParams) error {
	target := fmt.Sprintf("%s/api/authorization/v1/policy-delete", a.privateAddress)
	_, _, err := a.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, params)
	return err
}

// ResourceFilterParmas 资源过滤请求参数
type ResourceFilterParmas struct {
	Accessor  Vistor     `json:"accessor"`
	Resources []Resource `json:"resources"`
	Operation []string   `json:"operation"`
	Method    string     `json:"method"`
}

// ResourceFilter 资源权限过滤
func (a *authorization) ResourceFilter(ctx context.Context, params ResourceFilterParmas) ([]Resource, error) {
	target := fmt.Sprintf("%s/api/authorization/v1/resource-filter", a.privateAddress)
	_, body, err := a.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, params)
	if err != nil {
		return nil, err
	}

	byteRes, _ := json.Marshal(body)

	resources := []Resource{}
	err = json.Unmarshal(byteRes, &resources)
	return resources, err
}

// ListResourceOperationParmas 获取资源可操作列表请求参数
type ListResourceOperationParmas struct {
	Accessor  Vistor     `json:"accessor"`
	Resources []Resource `json:"resources"`
	Method    string     `json:"method"`
}

// ListResourceOperationRes 获取资源可操作列表结果
type ListResourceOperationRes struct {
	ID        string   `json:"id"`
	Operation []string `json:"operation"`
}

// ListResourceOperation 获取资源可操作列表
func (a *authorization) ListResourceOperation(ctx context.Context, params ListResourceOperationParmas) ([]ListResourceOperationRes, error) {
	target := fmt.Sprintf("%s/api/authorization/v1/resource-operation", a.privateAddress)
	_, body, err := a.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, params)
	if err != nil {
		return nil, err
	}

	byteRes, _ := json.Marshal(body)

	resources := []ListResourceOperationRes{}
	err = json.Unmarshal(byteRes, &resources)
	return resources, err

}

// OperationPermCheckParams 单次决策
type OperationPermCheckParams struct {
	Accessor  Vistor   `json:"accessor"`
	Resource  Resource `json:"resource"`
	Operation []string `json:"operation"`
	Method    string   `json:"method"`
}

// OperationPermCheck 决策指定资源是否具有操作列表权限
func (a *authorization) OperationPermCheck(ctx context.Context, params OperationPermCheckParams) (bool, error) {
	target := fmt.Sprintf("%s/api/authorization/v1/operation-check", a.privateAddress)
	_, body, err := a.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, params)
	if err != nil {
		return false, err
	}

	bodyMap := body.(map[string]interface{})
	return bodyMap["result"].(bool), nil
}

type ListResourceParams struct {
	Accessor  Vistor   `json:"accessor"`
	Resource  Resource `json:"resource"`
	Operation []string `json:"operation"`
	Method    string   `json:"method"`
}

// ListResource 资源列举
func (a *authorization) ListResource(ctx context.Context, params ListResourceParams) ([]Resource, error) {
	target := fmt.Sprintf("%s/api/authorization/v1/resource-list", a.privateAddress)
	_, body, err := a.httpClient.Post(ctx, target, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, params)
	if err != nil {
		return nil, err
	}

	byteRes, _ := json.Marshal(body)

	resources := []Resource{}
	err = json.Unmarshal(byteRes, &resources)
	return resources, err
}
