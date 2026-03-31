package perm

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

//go:generate mockgen -package mock_logics -source ../../logics/perm/perm_policy.go -destination ../../tests/mock_logics/perm_policy_mock.go

const (
	DataAdminID            = "00990824-4bf7-11f0-8fa7-865d5643e61f"
	DataFlowResourceType   = "data_flow"
	OperatorResourceType   = "operator"
	DataAdminResourceID    = "*"
	O11yResourceID         = "dataflow_page:o11y"
	RunWithAppOperation    = "run_with_app"
	CreateOperation        = "create"
	ModeifyOperation       = "modify"
	DeleteOperation        = "delete"
	ViewOperation          = "view"
	ManualExecOperation    = "manual_exec"
	RunStatisticsOperation = "run_statistics"
	ListOperation          = "list"
	DisplayOperation       = "display"

	// 算子相关权限定义
	OpExecuteOperation = "execute"
	// 旧权限Admin校验
	OldOnlyAdminOperation = "only_admin"
	OldAdminOperation     = "admin"
	OldShareOperation     = "share"
	// OldPublishOperatiuon 原AS publish-api接口使用
	OldPublishOperatiuon = "publish"
	// OldAppTokenOperation 原AS app-token接口使用,工作流运行接口应用账户支持所有流程执行
	OldAppTokenOperation = "app_token"
	OldOwnerOperation    = "owner"
)

var (
	Operations = []string{ListOperation, CreateOperation, ModeifyOperation, DeleteOperation, ViewOperation, ManualExecOperation, RunStatisticsOperation, RunWithAppOperation}
)

// PermPolicyHandler 权限操作接口
type PermPolicyHandler interface {
	IsDataAdmin(ctx context.Context, userID, userType string) (bool, error)
	IsUseAppAccount(ctx context.Context, userid, id, name string) (bool, error)
	CheckPerm(ctx context.Context, userID, userType string, resourceIDs []string, opts ...string) (bool, error)
	OperationCheck(ctx context.Context, accessorID, accessorType, resourceID string, opts ...string) (bool, error)
	OperationCheckWithResType(ctx context.Context, accessorID, accessorType, resourceID, resourceType string, opts ...string) error
	ResourceFilter(ctx context.Context, userID, userType string, resourceIDs []string, opts ...string) ([]string, error)
	CreatePolicy(ctx context.Context, userID, userType, userName, resourceID, resourceName string, allowOpts, denyOpts []string) error
	DeletePolicy(ctx context.Context, resourceIDs ...string) error
	UpdatePolicy(ctx context.Context, policyIDs []string, allowOpts, denyOpts []string) error
	MinPermList(ctx context.Context, userID, userType string, resourceIDs []string) ([]string, error)
	ListResource(ctx context.Context, userID, userType string, resourceType string, opts ...string) (*ResourceList, error)
	HandlePolicyNameChange(id, name, rType string)
}

type permPolicy struct {
	auth    drivenadapters.AuthorizationDriven
	publish entity.PushMessage
}

var ppOnce sync.Once
var permPolicyHandler PermPolicyHandler

// MinPermCheckReq 资源列表最小权限判断请求体参数
type MinPermCheckReq struct {
	ResourceIDs []string `json:"resource_ids"`
}

type ResourceList []string

func (rl *ResourceList) ToMap(fillterTypes ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(*rl))
	for _, v := range *rl {
		if v == DataAdminResourceID || !strings.Contains(v, ":") {
			continue
		}
		arr := strings.Split(v, ":")
		if len(fillterTypes) > 0 && !utils.Contains(fillterTypes, arr[1]) {
			continue
		}

		result[arr[0]] = struct{}{}
	}
	return result
}

// NewPermPolicy 实例化
func NewPermPolicy() PermPolicyHandler {
	ppOnce.Do(func() {
		permPolicyHandler = &permPolicy{
			auth:    drivenadapters.NewAuthorization(),
			publish: mod.NewMQHandler().Publish,
		}
	})

	return permPolicyHandler
}

// IsDataAdmin 判断用户是否为数据管理员
func (pp *permPolicy) IsDataAdmin(ctx context.Context, userID, userType string) (bool, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	hasDataFLowPerm, err := pp.OperationCheck(ctx, userID, userType, DataAdminResourceID, Operations...)
	if err != nil {
		return false, err
	}

	hasO11yPerm, err := pp.OperationCheck(ctx, userID, userType, O11yResourceID, DisplayOperation)
	if err != nil {
		return false, err
	}

	return hasDataFLowPerm && hasO11yPerm, nil
}

// IsUseAppAccount 判断用户是否使用应用账号
func (pp *permPolicy) IsUseAppAccount(ctx context.Context, userid, id, name string) (bool, error) {
	return pp.OperationCheck(ctx, userid, common.User.ToString(), id, RunWithAppOperation)
}

// CheckPerm 批量校验当前用户是否对资源实例具有指定权限
func (pp *permPolicy) CheckPerm(ctx context.Context, userID, userType string, resourceIDs []string, opts ...string) (bool, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	isDataAdmin, err := pp.IsDataAdmin(ctx, userID, userType)
	if err != nil {
		return false, err
	}
	if isDataAdmin {
		return true, nil
	}

	ids, err := pp.ResourceFilter(ctx, userID, userType, resourceIDs, opts...)
	if err != nil {
		return false, err
	}

	_, delete := utils.Arrcmp(resourceIDs, ids)
	if len(delete) > 0 {
		return false, ierr.NewPublicRestError(ctx, ierr.PErrorForbidden, aerr.NoPermission, map[string]interface{}{"resource_ids": delete})
	}

	return true, nil
}

// OperationCheck 校验当前用户是否对资源实例具有指定权限列表权限
func (pp *permPolicy) OperationCheck(ctx context.Context, accessorID, accessorType, resourceID string, opts ...string) (bool, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	params := drivenadapters.OperationPermCheckParams{
		Accessor: drivenadapters.Vistor{
			ID:   accessorID,
			Type: accessorType,
		},
		Resource: drivenadapters.Resource{
			ID:   resourceID,
			Type: DataFlowResourceType,
			// Name: resourceName,
		},
		Operation: opts,
		Method:    "GET",
	}

	isAdmin, err := pp.auth.OperationPermCheck(ctx, params)
	if err != nil {
		log.Warnf("[logic.OperationCheck] OperationPermCheck err, detail: %s", err.Error())
		return isAdmin, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}
	return isAdmin, nil
}

// OperationCheckWithResType 校验当前用户是否对资源实例具有指定权限列表权限,自定义资源类型
func (pp *permPolicy) OperationCheckWithResType(ctx context.Context, accessorID, accessorType, resourceID, resourceType string, opts ...string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	params := drivenadapters.OperationPermCheckParams{
		Accessor: drivenadapters.Vistor{
			ID:   accessorID,
			Type: accessorType,
		},
		Resource: drivenadapters.Resource{
			ID:   resourceID,
			Type: resourceType,
		},
		Operation: opts,
		Method:    "GET",
	}

	isAdmin, err := pp.auth.OperationPermCheck(ctx, params)
	if err != nil {
		log.Warnf("[logic.OperationCheck] OperationPermCheck err, detail: %s", err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}

	if !isAdmin {
		return ierr.NewPublicRestError(ctx, ierr.PErrorForbidden, aerr.NoPermission, map[string]interface{}{"resource_ids": []interface{}{resourceID}})
	}

	return nil
}

// ResourceFilter 过滤当前用户具有权限的资源实例
func (pp *permPolicy) ResourceFilter(ctx context.Context, userID, userType string, resourceIDs []string, opts ...string) ([]string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	params := drivenadapters.ResourceFilterParmas{
		Accessor: drivenadapters.Vistor{
			ID:   userID,
			Type: userType,
		},
		Resources: []drivenadapters.Resource{},
		Operation: opts,
		Method:    "GET",
	}

	resources := []drivenadapters.Resource{}
	for _, resourceID := range resourceIDs {
		resources = append(resources, drivenadapters.Resource{
			ID:   resourceID,
			Type: DataFlowResourceType,
		})
	}

	params.Resources = resources
	filterIDs := []string{}
	filterRes, err := pp.auth.ResourceFilter(ctx, params)
	if err != nil {
		log.Warnf("[logic.ResourceFilter] ResourceFilter err, detail: %s", err.Error())
		return filterIDs, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}

	for _, v := range filterRes {
		filterIDs = append(filterIDs, v.ID)
	}

	return filterIDs, nil
}

// CreatePolicy 创建策略
func (pp *permPolicy) CreatePolicy(ctx context.Context, userID, userType, userName, resourceID, resourceName string, allowOpts, denyOpts []string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	params := drivenadapters.PermPolicyParams{
		Accessor: &drivenadapters.Vistor{
			ID:   userID,
			Type: userType,
			Name: userName,
		},
		Resource: &drivenadapters.Resource{
			ID:   resourceID,
			Type: DataFlowResourceType,
			Name: resourceName,
		},
		Operation: drivenadapters.Operation{
			Allow: []drivenadapters.OperationID{},
			Deny:  []drivenadapters.OperationID{},
		},
		Condition: "",
		ExpiresAt: "",
	}

	for _, opt := range allowOpts {
		params.Operation.Allow = append(params.Operation.Allow, drivenadapters.OperationID{
			ID: opt,
		})
	}
	for _, opt := range denyOpts {
		params.Operation.Deny = append(params.Operation.Deny, drivenadapters.OperationID{
			ID: opt,
		})
	}

	err = pp.auth.CreatePolicy(ctx, []drivenadapters.PermPolicyParams{params})
	if err != nil {
		log.Warnf("[logic.CreatePolicy] CreatePolicy err, detail: %s", err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}
	return nil
}

// UpdatePolicy 更新策略
func (pp *permPolicy) UpdatePolicy(ctx context.Context, policyIDs []string, allowOpts, denyOpts []string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	params := drivenadapters.PermPolicyParams{
		Operation: drivenadapters.Operation{
			Allow: []drivenadapters.OperationID{},
			Deny:  []drivenadapters.OperationID{},
		},
		Condition: "",
		ExpiresAt: "",
	}

	for _, opt := range allowOpts {
		params.Operation.Allow = append(params.Operation.Allow, drivenadapters.OperationID{
			ID: opt,
		})
	}

	for _, opt := range denyOpts {
		params.Operation.Deny = append(params.Operation.Deny, drivenadapters.OperationID{
			ID: opt,
		})
	}

	err = pp.auth.UpdatePolicy(ctx, policyIDs, []drivenadapters.PermPolicyParams{params})
	if err != nil {
		log.Warnf("[logic.UpdatePolicy] UpdatePolicy err, detail: %s", err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}

	return nil
}

// DeletePolicy 删除策略
func (pp *permPolicy) DeletePolicy(ctx context.Context, resourceIDs ...string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	params := drivenadapters.PermPolicyDeleteParams{
		Method: "DELETE",
	}

	for _, resourceID := range resourceIDs {
		params.Resources = append(params.Resources, drivenadapters.Resource{
			ID:   resourceID,
			Type: DataFlowResourceType,
		})
	}

	err = pp.auth.DeletePolicy(ctx, params)
	if err != nil {
		log.Warnf("[logic.DeletePolicy] DeletePolicy err, detail: %s", err.Error())
		return ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}
	return nil
}

// MinPermList 批量查询资源实例最小操作权限列表
func (pp *permPolicy) MinPermList(ctx context.Context, userID, userType string, resourceIDs []string) ([]string, error) {
	var (
		err   error
		perms []string
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	isDataAdmin, err := pp.IsDataAdmin(ctx, userID, userType)
	if err != nil {
		return perms, err
	}
	if isDataAdmin {
		perms = append(perms, Operations...)
		perms = append(perms, DisplayOperation)
		return perms, nil
	}

	params := drivenadapters.ListResourceOperationParmas{
		Accessor: drivenadapters.Vistor{
			ID:   userID,
			Type: userType,
		},
		Method: "GET",
	}

	for _, resourceID := range resourceIDs {
		params.Resources = append(params.Resources, drivenadapters.Resource{
			ID:   resourceID,
			Type: DataFlowResourceType,
		})
	}

	res, err := pp.auth.ListResourceOperation(ctx, params)
	if err != nil {
		log.Warnf("[logic.MinPermList] ListResourceOperation err, detail: %s", err.Error())
		return perms, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}

	if len(res) == 0 {
		return perms, nil
	}

	perms = res[0].Operation
	for i := 1; i < len(res); i++ {
		perms = utils.GetIntersection(perms, res[i].Operation)
	}

	return perms, nil
}

// ListResource 列举当前资源类型具有指定权限的资源列表
func (pp *permPolicy) ListResource(ctx context.Context, userID, userType string, resourceType string, opts ...string) (*ResourceList, error) {
	var (
		err error
		ids = &ResourceList{}
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	params := drivenadapters.ListResourceParams{
		Accessor: drivenadapters.Vistor{
			ID:   userID,
			Type: userType,
		},
		Resource: drivenadapters.Resource{
			Type: DataFlowResourceType,
		},
		Operation: opts,
		Method:    "GET",
	}

	res, err := pp.auth.ListResource(ctx, params)
	if err != nil {
		log.Warnf("[logic.ListResource] ListResource err, detail: %s", err.Error())
		return ids, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}

	for _, v := range res {
		*ids = append(*ids, v.ID)
	}

	return ids, nil
}

// HandlePolicyNameChange 事件处理，策略名称变更
func (pp *permPolicy) HandlePolicyNameChange(id, name, rType string) {
	msg := map[string]interface{}{
		"id":   id,
		"type": rType,
		"name": name,
	}

	msgBytes, _ := json.Marshal(msg)
	pp.publish(common.TopicAuthorizationNameModify, msgBytes)
}
