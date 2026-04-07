package perm

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

// PermCheckerMap 所有权限校验checker定义集合
var permCheckerMap = map[string]DagPermChecker{}

// RegisterChecker 注册权限校验checker
func RegisterChecker(checkType string, checker DagPermChecker) {
	permCheckerMap[checkType] = checker
}

// DagPermChecker 数据流权限校验接口
type DagPermChecker interface {
	Check(ctx context.Context, dag *entity.Dag, userInfo *drivenadapters.UserInfo, opts ...string) (isAdmin bool, err error)
}

// DataFlowDagPermChecker 类型校验
type DataFlowDagPermChecker struct {
	PermPolicy PermPolicyHandler
}

// Check 校验权限
func (c *DataFlowDagPermChecker) Check(ctx context.Context, dag *entity.Dag, userInfo *drivenadapters.UserInfo, opts ...string) (bool, error) {
	dagType := dag.Type
	if dagType == "" {
		dagType = common.DagTypeDefault
	}
	resourceID := fmt.Sprintf("%s:%s", dag.ID, dagType)
	_, err := c.PermPolicy.CheckPerm(ctx, userInfo.UserID, userInfo.AccountType, []string{resourceID}, opts...)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ComBoOperatorPermChecker 组合算子类型校验
type ComBoOperatorPermChecker struct {
	PermPolicy PermPolicyHandler
}

// Check 校验权限
func (c *ComBoOperatorPermChecker) Check(ctx context.Context, dag *entity.Dag, userInfo *drivenadapters.UserInfo, opts ...string) (bool, error) {
	err := c.PermPolicy.OperationCheckWithResType(ctx, userInfo.UserID, userInfo.AccountType, dag.OperatorID, OperatorResourceType, opts...)
	if err != nil {
		// if ierr.Is(err, ierr.PublicErrorType, ierr.PErrorForbidden) {
		// 	return false, nil
		// }

		return false, err
	}
	return true, nil
}

// ObservabilityPermChecker 可观测性概览权限校验器
type ObservabilityPermChecker struct {
	PermPolicy PermPolicyHandler
}

// Check 校验权限
func (o *ObservabilityPermChecker) Check(ctx context.Context, dag *entity.Dag, userInfo *drivenadapters.UserInfo, opts ...string) (bool, error) {
	_, err := o.PermPolicy.CheckPerm(ctx, userInfo.UserID, userInfo.AccountType, []string{O11yResourceID}, opts...)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DefaultDagPermChecker 默认校验
type DefaultDagPermChecker struct {
	Usermgnt     drivenadapters.UserManagement
	IsAccessible func(dagAccessors *[]entity.Accessor, userAccessors []string) bool
}

// Check 校验权限
func (c *DefaultDagPermChecker) Check(ctx context.Context, dag *entity.Dag, userInfo *drivenadapters.UserInfo, opts ...string) (bool, error) {
	// publish-api 类型的流程判断
	if utils.Contains(opts, OldPublishOperatiuon) {
		return true, nil
	}

	// 工作流 应用账号运行流程不执行后续权限校验
	if utils.Contains(opts, OldAppTokenOperation) {
		return true, nil
	}

	// 如果是仅Admin权限，则校验后直接返回结果
	if utils.Contains(opts, OldOnlyAdminOperation) {
		isAdmin := utils.IsAdminRole(userInfo.Roles)
		if isAdmin {
			return true, nil
		}
		return false, nil
	}

	// 如果包含Admin，则需要校验
	if utils.Contains(opts, OldAdminOperation) {
		isAdmin := utils.IsAdminRole(userInfo.Roles)
		if isAdmin {
			return true, nil
		}
	}

	if utils.Contains(opts, OldOwnerOperation) {
		return dag.UserID == userInfo.UserID, nil
	}

	if dag.UserID != userInfo.UserID && utils.Contains(opts, OldShareOperation) {
		accessors, gerr := c.Usermgnt.GetUserAccessorIDs(userInfo.UserID)
		if gerr != nil {
			log.Errorf("[logic.ListDag] GetUserAccessorIDs err, detail: %s", gerr.Error())
			return false, aerr.NewIError(aerr.InternalError, aerr.ErrorDepencyService, gerr.Error())
		}
		if !c.IsAccessible(&dag.Accessors, accessors) {
			return false, aerr.NewIError(aerr.TaskNotFound, "", map[string]string{"dagId": dag.ID})
		}
	}

	return true, nil
}

// OperationProvider 可操作列表获取接口
type OperationProvider interface {
	GetOperations(dagType string) (bool, []string)
}

// MapOperationProvider
type MapOperationProvider struct {
	OpMap map[string][]string
}

// GetOperations 获取可操作列表
func (m *MapOperationProvider) GetOperations(dagType string) (bool, []string) {
	if ops, ok := m.OpMap[dagType]; ok {
		return ok, ops
	}
	return false, []string{}
}
