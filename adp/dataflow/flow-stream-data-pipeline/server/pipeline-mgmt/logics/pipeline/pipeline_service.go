package pipeline

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/kubernetes"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"flow-stream-data-pipeline/common"
	serrors "flow-stream-data-pipeline/errors"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
	"flow-stream-data-pipeline/pipeline-mgmt/logics"
	"flow-stream-data-pipeline/pipeline-mgmt/logics/deploy_dispatcher"
	"flow-stream-data-pipeline/pipeline-mgmt/logics/permission"
)

var (
	psOnce sync.Once
	ps     interfaces.PipelineMgmtService
)

var serviceInfo = &interfaces.ServiceInfo{
	ServiceName:       interfaces.ServiceName,
	ManagerDeployName: interfaces.ManagerDeployName,
	ServiceAccount:    interfaces.ServiceAccount,
}

// 管道管理的service
type pipelineMgmtService struct {
	appSetting       *common.AppSetting
	pmAccess         interfaces.PipelineMgmtAccess
	mqAccess         interfaces.MQAccess
	ibAccess         interfaces.IndexBaseAccess
	deployDispatcher interfaces.DeployDispatcherService
	ps               interfaces.PermissionService
	pipelineUpdateCh map[string]chan struct{}
}

func NewPipelineMgmtService(appSetting *common.AppSetting) interfaces.PipelineMgmtService {
	psOnce.Do(func() {
		dp, err := deploy_dispatcher.NewDeployDispatcherService(os.Getenv(interfaces.EnvPipelineNamespace))
		if err != nil {
			logger.Errorf("failed to new dispatcher, error: %s", err.Error())
		}

		ps = &pipelineMgmtService{
			appSetting:       appSetting,
			pmAccess:         logics.PMAccess,
			mqAccess:         logics.MQAccess,
			ibAccess:         logics.IBAccess,
			deployDispatcher: dp,
			ps:               permission.NewPermissionService(appSetting),
			pipelineUpdateCh: make(map[string]chan struct{}),
		}

		// userId 存入 context 中
		ctx := context.WithValue(context.Background(), interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
			ID:   interfaces.ADMIN_ID,
			Type: interfaces.ACCESSOR_TYPE_USER,
		})

		// 轮询管道的deploy状态
		go ps.WatchPipelineDeploys(ctx, appSetting.ServerSetting.WatchDeployInterval*time.Minute)
	})

	return ps
}

// CreatePipeline 创建管道
func (pmService *pipelineMgmtService) CreatePipeline(ctx context.Context, pipelineInfo *interfaces.Pipeline) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "create pipeline")
	span.SetAttributes(
		attr.Key("pipeline_id").String(pipelineInfo.PipelineID),
		attr.Key("pipeline_name").String(pipelineInfo.PipelineName))
	defer span.End()

	// 判断userid是否有创建管道的权限
	err := pmService.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_PIPELINE,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})

	if err != nil {
		return "", err
	}

	// 如果 id 为空，则生成一个 id
	if pipelineInfo.PipelineID == "" {
		pipelineInfo.PipelineID = xid.New().String()
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	pipelineInfo.Creator = accountInfo
	pipelineInfo.Updater = accountInfo

	currentTime := time.Now().UnixMilli()
	pipelineInfo.CreateTime = currentTime
	pipelineInfo.UpdateTime = currentTime
	pipelineInfo.PipelineStatus = interfaces.PipelineStatus_Running

	// 校验索引库是否存在
	httpErr := pmService.validateIndexBase(ctx, pipelineInfo)
	if httpErr != nil {
		return "", httpErr
	}

	// 先创建 topic，如果创建topic失败，则不创建管道。如果创建topic成功，管道创建失败，下次继续创建
	// 创建topic底层先检查kafka中是否存在，如果存在，则不创建，保证了这里执行的幂等性
	inputTopicName := GenerateInputTopicName(pmService.appSetting.MQSetting.Tenant, pipelineInfo.PipelineID)
	errorTopicName := GenerateErrorTopicName(pmService.appSetting.MQSetting.Tenant, pipelineInfo.PipelineID)
	// 批量创建 topic, input_topic 和 error_topic
	err = pmService.mqAccess.CreateTopicsOrPartitions(ctx, []string{inputTopicName, errorTopicName})
	if err != nil {
		logger.Errorf("failed to create topics, error: %s", err.Error())
		span.SetStatus(codes.Error, "create topics failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_CreateTopicsFailed).
			WithErrorDetails(err.Error())
	}

	// 调用driven层，创建管道
	err = pmService.pmAccess.CreatePipeline(ctx, pipelineInfo)
	if err != nil {
		logger.Errorf("failed to create pipeline, error: %s", err.Error())
		span.SetStatus(codes.Error, "create pipeline failed")
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_CreatePipelineFailed).
			WithErrorDetails(err.Error())
	}

	// 创建 deploy
	k8scfg := pmService.buildDeployConfig(ctx, pipelineInfo.PipelineID, pipelineInfo.DeploymentConfig.CpuLimit, pipelineInfo.DeploymentConfig.MemoryLimit)
	err = pmService.deployDispatcher.CreateDeploy(ctx, pipelineInfo.PipelineID, k8scfg, serviceInfo)
	if err != nil {
		// 创建 deploy 失败只打印错误日志
		logger.Errorf("failed to create pipeline deploy, error, %s", err.Error())
	}

	resrc := []interfaces.Resource{
		{
			ID:   pipelineInfo.PipelineID,
			Type: interfaces.RESOURCE_TYPE_PIPELINE,
			Name: pipelineInfo.PipelineName,
		},
	}

	// 注册资源策略
	err = pmService.ps.CreateResources(ctx, resrc, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return "", err
	}

	span.SetStatus(codes.Ok, "")
	return pipelineInfo.PipelineID, nil
}

// DeletePipeline 删除管道
func (pmService *pipelineMgmtService) DeletePipeline(ctx context.Context, pipelineID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "delete pipeline")
	span.SetAttributes(
		attr.Key("pipeline_id").String(pipelineID))
	defer span.End()

	// 先获取资源序列
	matchResouces, err := pmService.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_PIPELINE,
		[]string{pipelineID}, []string{interfaces.OPERATION_TYPE_DELETE},
		false, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return err
	}

	// 资源过滤后的数量跟请求的数量不等，说明管道没有权限
	if len(matchResouces) != 1 {
		return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for pipeline's delete operation")
	}

	// 删除 topic,，删除 input_topic 和 error_topic
	inputTopicName := GenerateInputTopicName(pmService.appSetting.MQSetting.Tenant, pipelineID)
	errorTopicName := GenerateErrorTopicName(pmService.appSetting.MQSetting.Tenant, pipelineID)
	err = pmService.mqAccess.DeleteTopics(ctx, []string{inputTopicName, errorTopicName})
	if err != nil {
		logger.Errorf("failed to delete topic of pipeline '%s', error: %v", pipelineID, err)
		span.SetStatus(codes.Error, "delete topic failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DeleteTopicsFailed).
			WithErrorDetails(err.Error())
	}

	// 删除管道信息
	err = pmService.pmAccess.DeletePipeline(ctx, pipelineID)
	if err != nil {
		logger.Errorf("failed to delete pipeline '%s', error: %v", pipelineID, err)
		span.SetStatus(codes.Error, "delete pipeline failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).
			WithErrorDetails(err.Error())
	}

	delete(pmService.pipelineUpdateCh, pipelineID)
	// 删除 deploy pod 资源
	err = pmService.deployDispatcher.DeleteDeploy(ctx, pipelineID, serviceInfo)
	if err != nil {
		logger.Errorf("failed to delete deploy of pipeline '%s', error: %v", pipelineID, err)
	}

	//  清除资源策略
	err = pmService.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_PIPELINE, []string{pipelineID})
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdatePipeline 修改管道信息
func (pmService *pipelineMgmtService) UpdatePipeline(ctx context.Context, pipeline *interfaces.Pipeline) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "update pipeline")
	span.SetAttributes(
		attr.Key("pipeline_id").String(pipeline.PipelineID),
		attr.Key("pipeline_name").String(pipeline.PipelineName))
	defer span.End()

	// 校验是否有更新管道的权限
	err := pmService.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_PIPELINE,
		ID:   pipeline.PipelineID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// 校验索引库是否存在
	httpErr := pmService.validateIndexBase(ctx, pipeline)
	if httpErr != nil {
		return httpErr
	}

	// pipeline_id对应的管道资源是否存在
	oldPipeline, exist, err := pmService.pmAccess.GetPipeline(ctx, pipeline.PipelineID)
	if err != nil {
		logger.Errorf("failed to retrieve pipeline, error: %s", err.Error())
		span.SetStatus(codes.Error, "retrieve pipeline failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).
			WithErrorDetails(err.Error())
	}
	if !exist {
		logger.Errorf("pipeline %s does not exist", pipeline.PipelineID)
		span.SetStatus(codes.Error, "pipeline does not exist")
		return rest.NewHTTPError(ctx, http.StatusNotFound, serrors.StreamDataPipeline_NotFound_Pipeline).
			WithErrorDetails(fmt.Sprintf("pipeline %s does not exist", pipeline.PipelineID))
	}

	// 如果修改了管道名称，判断管道名称是否重名
	if pipeline.PipelineName != oldPipeline.PipelineName {
		_, exist, err := pmService.pmAccess.CheckPipelineExistByName(ctx, pipeline.PipelineName)
		if err != nil {
			logger.Errorf("failed to get pipeline by pipeline name %s, error: %s", pipeline.PipelineName, err.Error())
			span.SetStatus(codes.Error, "get pipeline by pipeline name failed")
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).
				WithErrorDetails(err.Error())
		}
		if exist {
			logger.Errorf("pipeline name %s already exists", pipeline.PipelineName)
			span.SetStatus(codes.Error, "pipeline name already exists")
			return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_Duplicated_PipelineName).
				WithErrorDetails(fmt.Sprintf("pipeline name %s already exists", pipeline.PipelineName))
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	pipeline.Updater = accountInfo
	pipeline.UpdateTime = time.Now().UnixMilli()

	// 调用driven层，修改管道信息
	err = pmService.pmAccess.UpdatePipeline(ctx, pipeline)
	if err != nil {
		logger.Errorf("failed to update pipeline, error: %s", err.Error())
		span.SetStatus(codes.Error, "update pipeline failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).
			WithErrorDetails(err.Error())
	}

	// 如果 cpu 和 memory 限额变化，编辑 deploy pod 资源
	if pipeline.DeploymentConfig.CpuLimit != oldPipeline.DeploymentConfig.CpuLimit ||
		pipeline.DeploymentConfig.MemoryLimit != oldPipeline.DeploymentConfig.MemoryLimit {

		k8scfg := pmService.buildDeployConfig(ctx, pipeline.PipelineID, pipeline.DeploymentConfig.CpuLimit, pipeline.DeploymentConfig.MemoryLimit)
		err = pmService.deployDispatcher.UpdateDeploy(ctx, pipeline.PipelineID, k8scfg, serviceInfo)
		if err != nil {
			logger.Errorf("updatePipeline resource error, %s", err.Error())
		}
	} else {
		if _, ok := pmService.pipelineUpdateCh[pipeline.PipelineID]; !ok {
			pmService.pipelineUpdateCh[pipeline.PipelineID] = make(chan struct{}, 100)
		}
		pmService.pipelineUpdateCh[pipeline.PipelineID] <- struct{}{}
	}

	// 请求更新资源名称的接口，更新资源的名称
	err = pmService.ps.UpdateResource(ctx, interfaces.Resource{
		ID:   pipeline.PipelineID,
		Type: interfaces.RESOURCE_TYPE_PIPELINE,
		Name: pipeline.PipelineName,
	})
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (pmService *pipelineMgmtService) validateIndexBase(ctx context.Context, pipeline *interfaces.Pipeline) error {
	// 如果不使用数据里的__index_base，则配置的索引库类型不能为空且存在
	if !pipeline.UseIndexBaseInData {
		if pipeline.IndexBase == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_IndexBase).
				WithErrorDetails("base type cannot be empty when 'use_index_base_in_data' is false")
		}

		_, err := pmService.ibAccess.GetIndexBasesByTypes(ctx, []string{pipeline.IndexBase})
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				serrors.StreamDataPipeline_InternalError_GetIndexBaseByTypeFailed).WithErrorDetails(err.Error())
		}

		// 使用数据里的 __index_base, 索引库类型修改为 "" 存入表
	} else {
		pipeline.IndexBase = ""
	}

	return nil
}

// ListPipelines 获取管道列表
func (pmService *pipelineMgmtService) ListPipelines(ctx context.Context, param *interfaces.ListPipelinesQuery) ([]*interfaces.Pipeline, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "list pipelines")
	defer span.End()

	//调用driven层，获取管道信息列表
	pipelineList, err := pmService.pmAccess.ListPipelines(ctx, param)
	if err != nil {
		logger.Errorf("failed to list pipelines, error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("failed to list pipelines, error: %v", err))
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).
			WithErrorDetails(err.Error())
	}

	if len(pipelineList) == 0 {
		return pipelineList, 0, nil
	}

	//调用driven层，获取管道总数
	// total, err := pmService.pmAccess.GetPipelinesTotal(ctx, pipelineQuery)
	// if err != nil {
	// 	logger.Errorf("failed to get the total of pipelines, error: %s", err.Error())
	// 	span.SetStatus(codes.Error, "get the total of pipelines failed")
	// 	return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).
	// 		WithErrorDetails(err.Error())
	// }

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, pl := range pipelineList {
		resMids = append(resMids, pl.PipelineID)
	}
	matchResoucesMap, err := pmService.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_PIPELINE, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return pipelineList, 0, err
	}

	// 遍历对象
	results := make([]*interfaces.Pipeline, 0)
	for _, pl := range pipelineList {
		if resrc, exist := matchResoucesMap[pl.PipelineID]; exist {
			pl.Operations = resrc.Operations // 用户当前有权限的操作
			results = append(results, pl)
		}
	}

	// limit = -1,则返回所有
	if param.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if param.Offset < 0 || param.Offset >= len(results) {
		return nil, 0, nil
	}
	// 计算结束位置
	end := param.Offset + param.Limit
	if end > len(results) {
		end = len(results)
	}

	span.SetStatus(codes.Ok, "")
	return results[param.Offset:end], len(results), nil

}

func (pmService *pipelineMgmtService) GetPipelineTotals(ctx context.Context, pipelineQuery *interfaces.ListPipelinesQuery) (int, error) {
	total, err := pmService.pmAccess.GetPipelinesTotal(ctx, pipelineQuery)
	if err != nil {
		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).
			WithErrorDetails(err.Error())
	}
	return total, nil
}

// 根据名称检查管道是否存在，暴露 exist 参数，方便内部模块调用时根据exist决定后续行为
func (pmService *pipelineMgmtService) CheckPipelineExistByName(ctx context.Context, name string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Check pipeline exist by name")
	defer span.End()

	viewID, exist, err := pmService.pmAccess.CheckPipelineExistByName(ctx, name)
	if err != nil {
		errDetails := fmt.Sprintf("Check pipeline exist by name %s error: %s", name, err.Error())
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, "check pipeline exist by name failed")
		return viewID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			serrors.StreamDataPipeline_InternalError_CheckPipelineIfExistFailed).WithErrorDetails(errDetails)
	}

	span.SetStatus(codes.Ok, "")
	return viewID, exist, nil
}

// 单个查询，暴露 exist 参数，方便内部模块调用自己决定对存在与否的行为
func (pmService *pipelineMgmtService) CheckPipelineExistByID(ctx context.Context, ID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Check pipeline exist by id")
	defer span.End()

	viewName, exist, err := pmService.pmAccess.CheckPipelineExistByID(ctx, ID)
	if err != nil {
		errDetails := fmt.Sprintf("Check pipeline exist by id %s error: %s", ID, err.Error())
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, "check pipeline exist by id failed")
		return viewName, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			serrors.StreamDataPipeline_InternalError_CheckPipelineIfExistFailed).WithErrorDetails(errDetails)
	}

	span.SetStatus(codes.Ok, "")
	return viewName, exist, nil
}

// GetPipeline 获取管道详情
func (pmService *pipelineMgmtService) GetPipeline(ctx context.Context, pipelineID string, isListen bool) (*interfaces.Pipeline, bool, error) {
	if isListen {
		return pmService.listenPipeline(ctx, pipelineID)
	} else {
		return pmService.getPipelineByPipelineID(ctx, pipelineID)
	}
}

func (pmService *pipelineMgmtService) getPipelineByPipelineID(ctx context.Context, pipelineID string) (*interfaces.Pipeline, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "get pipeline by pipeline id")
	span.SetAttributes(attr.Key("pipeline_id").String(pipelineID))
	defer span.End()

	// 先获取资源序列
	matchResouces, err := pmService.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_PIPELINE, []string{pipelineID},
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return nil, false, err
	}

	// 资源过滤后的数量跟请求的数量不等，说明没有权限
	if len(matchResouces) != 1 {
		return nil, false, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails("Access denied: insufficient permissions for pipeline's view_detail operation.")
	}

	pipeline, exist, err := pmService.pmAccess.GetPipeline(ctx, pipelineID)
	if err != nil {
		logger.Errorf("failed to retrieve pipeline, error: %s", err.Error())
		err = rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).WithErrorDetails(err.Error())
		span.SetStatus(codes.Error, "get pipeline by pipeline id failed")
		return nil, false, err
	}
	if !exist {
		err = rest.NewHTTPError(ctx, http.StatusNotFound, serrors.StreamDataPipeline_NotFound_Pipeline).
			WithErrorDetails("pipeline resource is not exist")
		span.SetStatus(codes.Error, "pipeline resource is not exist")
		return nil, false, err
	}

	pipeline.Operations = matchResouces[pipelineID].Operations
	// 补充 topic 信息
	pipeline.InputTopic = GenerateInputTopicName(pmService.appSetting.MQSetting.Tenant, pipelineID)
	pipeline.OutputTopic = GenerateOutputTopicName(pmService.appSetting.MQSetting.Tenant, pipeline.IndexBase)
	pipeline.ErrorTopic = GenerateErrorTopicName(pmService.appSetting.MQSetting.Tenant, pipelineID)

	span.SetStatus(codes.Ok, "")
	return pipeline, true, nil
}

// 每30s监听pipeline的 updatech 是否变化
func (pmService *pipelineMgmtService) listenPipeline(ctx context.Context, pipelineID string) (*interfaces.Pipeline, bool, error) {
	timeOut := 30 * time.Second
	select {
	case <-time.After(timeOut):
		// timeout 配置未变化
		return pmService.getPipelineByPipelineID(ctx, pipelineID)
	case <-pmService.pipelineUpdateCh[pipelineID]:
		// 将最新的信息返回给客户端
		return pmService.getPipelineByPipelineID(ctx, pipelineID)
	}
}

// UpdatePipelineStatus 修改管道状态
func (pmService *pipelineMgmtService) UpdatePipelineStatus(ctx context.Context, pipelineID string,
	pipelineStatusInfo *interfaces.PipelineStatusParamter, isInnerRequest bool) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "update pipeline status")
	span.SetAttributes(attr.Key("pipeline_id").String(pipelineID))
	defer span.End()

	// 校验是否有更新管道的权限
	err := pmService.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_PIPELINE,
		ID:   pipelineID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// 检测管道资源是否存在
	pipeline, exist, err := pmService.pmAccess.GetPipeline(ctx, pipelineID)
	if err != nil {
		logger.Errorf("failed to retrieve pipeline, error: %s", err.Error())
		span.SetStatus(codes.Error, "retrieve pipeline failed")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).
			WithErrorDetails(err.Error())
	}
	if !exist {
		span.SetStatus(codes.Error, "pipeline resource does not exist")
		return rest.NewHTTPError(ctx, http.StatusNotFound, serrors.StreamDataPipeline_NotFound_Pipeline).
			WithErrorDetails("pipeline resource does not exist")
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	pipeline.Updater = accountInfo
	pipeline.UpdateTime = time.Now().UnixMilli()
	pipeline.PipelineStatus = pipelineStatusInfo.Status
	if pipelineStatusInfo.Status == interfaces.PipelineStatus_Error {
		pipeline.PipelineStatusDetails = pipelineStatusInfo.Details
	} else {
		pipeline.PipelineStatusDetails = ""
	}

	// 调用driven层，修改管道状态
	err = pmService.pmAccess.UpdatePipelineStatus(ctx, pipeline, isInnerRequest)
	if err != nil {
		logger.Errorf("failed to update pipeline status, error: %s", err.Error())
		span.SetStatus(codes.Error, "update pipeline status failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, serrors.StreamDataPipeline_InternalError_DataBase).
			WithErrorDetails(err.Error())
	}

	if !isInnerRequest {
		if _, ok := pmService.pipelineUpdateCh[pipelineID]; !ok {
			pmService.pipelineUpdateCh[pipelineID] = make(chan struct{}, 100)
		}
		pmService.pipelineUpdateCh[pipelineID] <- struct{}{}
	}

	// 请求更新资源名称的接口，更新资源的名称
	err = pmService.ps.UpdateResource(ctx, interfaces.Resource{
		ID:   pipeline.PipelineID,
		Type: interfaces.RESOURCE_TYPE_PIPELINE,
		Name: pipeline.PipelineName,
	})
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 每隔 10min 轮询 pipeline deploys
func (pmService *pipelineMgmtService) WatchPipelineDeploys(ctx context.Context, interval time.Duration) {
	logger.Infof("watch pipeline deploys status, interval: %v", interval)
	for {
		pmService.listPipelinesAndUpdateDeploy(ctx)
		time.Sleep(interval)
	}
}

// 获取数据库中的Pipelines，在数据库中存在，但是没有对应的deploy, 则新建deploy
// 在数据库中存在，并且有对应的deploy，查看limit是否更新成功，否则更新limit
// 在数据库中不存在，deploy存在的时候，删除deploy
func (pmService *pipelineMgmtService) listPipelinesAndUpdateDeploy(ctx context.Context) {
	pipelineQuery := &interfaces.ListPipelinesQuery{
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Limit:     interfaces.MAX_LIMIT,
			Offset:    interfaces.MIN_OFFSET,
			Sort:      interfaces.TABLE_SORT[interfaces.DEFAULT_SORT],
			Direction: interfaces.DESC_DIRECTION,
		},
	}
	pipelines, err := pmService.pmAccess.ListPipelines(ctx, pipelineQuery)
	if err != nil {
		logger.Errorf("failed to list pipelines, error: %s", err.Error())
		return
	}

	pipelineMap := make(map[string]*interfaces.Pipeline)
	for _, v := range pipelines {
		pipelineMap[v.PipelineID] = v
	}

	for _, pipeline := range pipelines {
		deploy, err := pmService.deployDispatcher.GetDeploy(ctx, pipeline.PipelineID, serviceInfo)
		if kubernetes.ResourceNotFound(err) {
			logger.Infof("The deploy of pipeline '%s' is not exist", pipeline.PipelineID)
			cErr := pmService.deployDispatcher.CreateDeploy(
				ctx,
				pipeline.PipelineID,
				pmService.buildDeployConfig(ctx, pipeline.PipelineID, pipeline.DeploymentConfig.CpuLimit, pipeline.DeploymentConfig.MemoryLimit),
				serviceInfo,
			)

			if cErr != nil {
				logger.Errorf("failed to create deploy of pipeline '%s', error: %v", pipeline.PipelineID, cErr)
			}
			continue
		}

		if err != nil {
			logger.Errorf("failed to get deploy of pipeline '%s', error: %v", pipeline.PipelineID, err)
			continue
		}

		// 如果存在 deploy, 查看 resource limit 是否一致, 以及从属关系是否失效
		if deploy != nil {
			containers := deploy.Spec.Template.Spec.Containers
			resourceLimit := containers[0].Resources.Limits
			cpuLimit, _ := resourceLimit.Cpu().AsInt64()
			memoryLimit, _ := resourceLimit.Memory().AsInt64()
			pipelineImage := os.Getenv(interfaces.EnvPipelineImage)
			logger.Debugf("update pipeline deploy, cpu limit is %d, memory limit is %d, deploy ownerreference is %v",
				cpuLimit, memoryLimit, deploy.OwnerReferences)

			if len(deploy.OwnerReferences) == 0 || cpuLimit != int64(pipeline.DeploymentConfig.CpuLimit) || memoryLimit != int64(pipeline.DeploymentConfig.MemoryLimit*1024*1024) || containers[0].Image != pipelineImage {
				logger.Infof("update deploy of pipeline '%s'", pipeline.PipelineID)
				err := pmService.deployDispatcher.UpdateDeploy(ctx, pipeline.PipelineID,
					pmService.buildDeployConfig(ctx, pipeline.PipelineID, pipeline.DeploymentConfig.CpuLimit, pipeline.DeploymentConfig.MemoryLimit),
					serviceInfo,
				)
				if err != nil {
					logger.Errorf("failed to update deploy of pipeline '%s', error: %v", pipeline.PipelineID, err)
				}
			}
		}
	}

	opts := metav1.ListOptions{
		LabelSelector: "pipeline-worker=pipeline-worker",
	}

	deployList, err := pmService.deployDispatcher.ListDeploy(ctx, opts)
	if err != nil {
		logger.Errorf("failed to list pipeline deploys, error: %s", err.Error())
		return
	}
	if deployList == nil {
		return
	}

	for _, deploy := range deployList.Items {
		pipelineID := deploy.Labels["pipeline-id"]
		if _, ok := pipelineMap[pipelineID]; !ok {
			logger.Infof("The deploy of pipeline '%s' need to be deleted, because the pipeline is not in database", pipelineID)
			err = pmService.deployDispatcher.DeleteDeploy(ctx, pipelineID, serviceInfo)
			if err != nil {
				logger.Errorf("failed to delete deploy of pipeline '%s', error: %v", pipelineID, err)
			}
		}
	}
}

// deployment 配置信息
func (pmService *pipelineMgmtService) buildDeployConfig(ctx context.Context, pipelineID string, limitCpu, limitMemory int) *interfaces.KubernetesCfg {
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	logger.Infof("build deploy config, pipelineID: %s, limitCpu: %d, limitMemory: %d, accountInfo: %v", pipelineID, limitCpu, limitMemory, accountInfo)

	var (
		image = kubernetes.Image{
			Name:     "pipeline",
			ImageURL: os.Getenv(interfaces.EnvPipelineImage),
		}
		resource = kubernetes.ResourceRequire{
			Limit: kubernetes.Resource{
				Cpu:    fmt.Sprintf("%d", limitCpu),
				Memory: fmt.Sprintf("%dMi", limitMemory),
			},
			Require: kubernetes.Resource{
				Cpu:    "0",
				Memory: "0",
			},
		}
		labels = map[string]string{
			"pipeline-id":     pipelineID,
			"pipeline-worker": interfaces.PipelineDeployLabels,
		}
		cmd  = []string{"./pipeline-worker-server"}
		args = []string{"--worker_id", pipelineID, "--account_id", accountInfo.ID, "--account_type", accountInfo.Type}
	)

	mounts := []corev1.VolumeMount{{
		Name:      "flow-stream-data-pipeline-cm",
		MountPath: "/opt/pipeline/config",
	}}

	volumes := []corev1.Volume{{
		Name: "flow-stream-data-pipeline-cm",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "flow-stream-data-pipeline-cm",
				},
			},
		},
	}}

	return &interfaces.KubernetesCfg{
		Image:     image,
		Resource:  resource,
		Namespace: os.Getenv(interfaces.EnvPipelineNamespace),

		Label:         labels,
		ContainerCmd:  cmd,
		ContainerArgs: args,
		WorkingDir:    "/opt/pipeline/",
		VolumeMounts:  mounts,
		Volumes:       volumes,
	}
}

func GenerateInputTopicName(tenant string, pipelineID string) string {
	return fmt.Sprintf(interfaces.TopicInputName, tenant, pipelineID)
}

func GenerateOutputTopicName(tenant string, BaseType string) string {
	return fmt.Sprintf(interfaces.TopicOutputName, tenant, BaseType)
}

func GenerateErrorTopicName(tenant string, pipelineID string) string {
	return fmt.Sprintf(interfaces.TopicErrorName, tenant, pipelineID)
}

func (pmService *pipelineMgmtService) ListPipelineResources(ctx context.Context, param *interfaces.ListPipelinesQuery) ([]*interfaces.Resource, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "list pipeline resources")
	span.End()

	pipelines, err := pmService.pmAccess.ListPipelines(ctx, param)
	if err != nil {
		logger.Errorf("ListPipelines error: %s", err.Error())
		span.SetStatus(codes.Error, "List pipelines error")
		span.End()
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			serrors.StreamDataPipeline_InternalError_DataBase).WithErrorDetails(err.Error())
	}
	if len(pipelines) == 0 {
		return nil, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, p := range pipelines {
		resMids = append(resMids, p.PipelineID)
	}
	// 校验权限管理的操作权限
	matchResoucesMap, err := pmService.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_PIPELINE, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return nil, 0, err
	}

	// 遍历对象
	results := make([]*interfaces.Resource, 0)
	for _, p := range pipelines {
		if _, exist := matchResoucesMap[p.PipelineID]; exist {
			results = append(results, &interfaces.Resource{
				ID:   p.PipelineID,
				Type: interfaces.RESOURCE_TYPE_PIPELINE,
				Name: p.PipelineName,
			})
		}
	}

	// limit = -1,则返回所有
	if param.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if param.Offset < 0 || param.Offset >= len(results) {
		return nil, 0, nil
	}
	// 计算结束位置
	end := param.Offset + param.Limit
	if end > len(results) {
		end = len(results)
	}

	span.SetStatus(codes.Ok, "")
	return results[param.Offset:end], len(results), nil
}
