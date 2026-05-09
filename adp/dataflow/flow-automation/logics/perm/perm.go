package perm

import (
	"context"
	"sync"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
)

//go:generate mockgen -package mock_perm -source ../../logics/perm/perm.go -destination ../../tests/mock_logics/mock_perm/perm_mock.go

// PermCheckerService 权限检查接口定义
type PermCheckerService interface {
	CheckDagAndPerm(ctx context.Context, dagID string, userInfo *drivenadapters.UserInfo, opProvider OperationProvider) (bool, error)
	CheckPerm(ctx context.Context, resourceType string, dag *entity.Dag, userInfo *drivenadapters.UserInfo, opProvider OperationProvider) (bool, error)
}

type permChecker struct {
	mongo mod.Store
}

var (
	pOnce sync.Once
	p     PermCheckerService
)

// NewPermCheckerService 创建权限检查服务
func NewPermCheckerService() PermCheckerService {
	pOnce.Do(func() {
		p = &permChecker{
			mongo: mod.GetStore(),
		}
	})

	return p
}

// CheckDagAndPerm 检查Dag是否存在并具有对应操作权限
func (p *permChecker) CheckDagAndPerm(ctx context.Context, dagID string, userInfo *drivenadapters.UserInfo, opProvider OperationProvider) (bool, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dag, err := p.mongo.GetDag(ctx, dagID)
	if err != nil {
		if aerr.IsNotFoundErr(err) {
			return false, aerr.NewIError(aerr.TaskNotFound, "", map[string]string{"dagId": dagID})
		}
		log.Errorf("[logic.checkDagAndPerm] GetDag err, detail: %s", err.Error())
		return false, aerr.NewIError(aerr.InternalError, "", nil)
	}

	// 调试模式不检查权限
	if dag.IsDebug {
		return true, nil
	}

	isAdmin, err := p.CheckPerm(ctx, dag.Type, dag, userInfo, opProvider)
	if err != nil {
		return false, err
	}

	query := map[string]interface{}{"_id": dagID}
	if !isAdmin {
		query["userid"] = userInfo.UserID
	}

	total, err := p.mongo.GetDagCount(ctx, query)
	if err != nil {
		log.Warnf("[logic.CheckDagAndPerm] GetDagCount err, detail: %s", err.Error())
		return false, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
	}
	if total == 0 {
		return false, ierr.NewPublicRestError(ctx, ierr.PErrorNotFound, aerr.DescKeyTaskNotFound, map[string]interface{}{"param": query})
	}

	return isAdmin, nil
}

// CheckPerm 检查dag是否具有对应操作权限
//
// resourceType: Dag的类型,如果为空,则使用default类型
// dag: Dag对象, 可为空，根据实际业务传递
// userInfo: 用户信息
// opProvider: 操作列表提供者,用于获取对应的操作列表
func (p *permChecker) CheckPerm(ctx context.Context, resourceType string, dag *entity.Dag, userInfo *drivenadapters.UserInfo, opProvider OperationProvider) (bool, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if resourceType == "" {
		resourceType = common.DagTypeDefault
	}

	// 选择checker
	checker, ok := permCheckerMap[resourceType]
	if !ok {
		checker = permCheckerMap[common.DagTypeDefault]
		resourceType = common.DagTypeDefault
	}

	ok, ops := opProvider.GetOperations(resourceType)
	if !ok {
		checker = permCheckerMap[common.DagTypeDefault]
		_, ops = opProvider.GetOperations(common.DagTypeDefault)
	}

	isAdmin, err := checker.Check(ctx, dag, userInfo, ops...)
	if err != nil {
		return false, err
	}

	return isAdmin, nil
}
