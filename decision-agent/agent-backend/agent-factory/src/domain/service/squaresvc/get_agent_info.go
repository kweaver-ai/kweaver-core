package squaresvc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/p2e/daconfp2e"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"

	"github.com/pkg/errors"
)

// GetAgentInfo implements iv3portdriver.IMarketSvc.
func (svc *squareSvc) GetAgentInfo(ctx context.Context, agentInfoReq *squarereq.AgentInfoReq) (res *squareresp.AgentMarketAgentInfoResp, err error) {
	res = squareresp.NewAgentMarketAgentInfoResp()
	res.LatestVersion = string(cdaenum.StatusUnpublished)

	// 1. 获取最新 v0 版本 Agent 配置
	agentV0CfgPo, err := svc.agentConfRepo.GetByID(ctx, agentInfoReq.AgentID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "agent not found")
			return
		}

		err = errors.Wrapf(err, "[squareSvc.GetAgentInfo]:svc.agentConfRepo.GetByID(ctx, %s)", agentInfoReq.AgentID)

		return
	}

	// 2. 记录访问日志
	visitLogErr := svc.RecordVisitLog(ctx, agentInfoReq)
	if visitLogErr != nil {
		svc.Logger.Warnf("svc.RecordVisitLog(ctx, %+v) failed, err:%v\n", agentInfoReq, visitLogErr)
	}

	// 3. 如果是未发布的版本，返回 V0版本配置，同时版本号设置为 v0
	if agentInfoReq.AgentVersion == daconstant.AgentVersionUnpublished {
		var agentCfgEO *daconfeo.DataAgent

		agentCfgEO, err = daconfp2e.DataAgent(ctx, agentV0CfgPo)
		if err != nil {
			err = errors.Wrapf(err, "[squareSvc.GetAgentInfo]:daconfp2e.DataAgent(&po.DataAgentPo)")
			return
		}

		// 3.1 获取 LatestVersion
		res.LatestVersion, err = svc.getLatestVersion(ctx, agentInfoReq.AgentID, nil)
		if err != nil {
			err = errors.Wrapf(err, "[squareSvc.GetAgentInfo]:svc.getLatestVersion(ctx, %s, nil)", agentInfoReq.AgentID)
			return
		}

		res.Version = daconstant.AgentVersionUnpublished
		res.DataAgent = *agentCfgEO
		res.Config = *res.DataAgent.Config

		return
	}

	// 4. agentInfoReq.AgentVersion != daconstant.AgentVersionUnpublished时
	err = svc.notUnpublished(ctx, agentInfoReq, res)
	if err != nil {
		return
	}

	return
}

func (svc *squareSvc) notUnpublished(ctx context.Context, agentInfoReq *squarereq.AgentInfoReq, res *squareresp.AgentMarketAgentInfoResp) (err error) {
	// 1. 获取发布记录
	releasePo, err := svc.releaseRepo.GetByAgentID(ctx, agentInfoReq.AgentID)
	if err != nil {
		err = errors.Wrapf(err, "[squareSvc.notUnpublished]:svc.releaseRepo.GetByAgentId(ctx, %s)", agentInfoReq.AgentID)
		return
	}

	// 2. 获取 LatestVersion
	res.LatestVersion, err = svc.getLatestVersion(ctx, agentInfoReq.AgentID, releasePo)
	if err != nil {
		err = errors.Wrapf(err, "[squareSvc.notUnpublished]: svc.getLatestVersion err")
		return
	}

	// 3. 如果是已发布版本，则基于已发布版本的配置返回结果
	if releasePo != nil && agentInfoReq.AgentVersion == daconstant.AgentVersionLatest {
		agentInfoReq.AgentVersion = releasePo.AgentVersion
	}

	// 4. 如果版本号为空，返回错误
	if agentInfoReq.AgentVersion == "" {
		err = errors.Wrapf(err, "agent version is empty")
		return
	}

	// 5. 获取 指定Agent 版本的配置，检查发布状态
	historyPo, err := svc.releaseHistoryRepo.GetByAgentIdVersion(ctx, agentInfoReq.AgentID, agentInfoReq.AgentVersion)
	if err != nil {
		err = errors.Wrapf(err, "[squareSvc.notUnpublished]:svc.releaseRepo.GetByAgentIdVersion(ctx, %s, %s)", agentInfoReq.AgentID, agentInfoReq.AgentVersion)
		return
	}

	if historyPo == nil {
		err = fmt.Errorf("the agent version:%s is not exist", agentInfoReq.AgentVersion)
		return
	}

	// 6. 将historyPo.AgentConfig 转换为 DataAgentPo
	agentCfgPo := &dapo.DataAgentPo{}

	err = json.Unmarshal([]byte(historyPo.AgentConfig), agentCfgPo)
	if err != nil {
		err = errors.Wrapf(err, "[squareSvc.notUnpublished]:json.Unmarshal([]byte(%s), &agentCfg)", historyPo.AgentConfig)
		return
	}

	// 7. po 转 eo
	agentCfgEO, err := daconfp2e.DataAgent(ctx, agentCfgPo)
	if err != nil {
		err = errors.Wrapf(err, "[squareSvc.notUnpublished]:daconfp2e.DataAgent(&po.DataAgentPo)")
		return
	}

	// 8. eo 转 res
	res.DataAgent = *agentCfgEO

	res.Version = agentInfoReq.AgentVersion
	res.Description = historyPo.AgentDesc
	res.PublishedAt = historyPo.UpdateTime
	res.Config = *res.DataAgent.Config

	res.PublishInfo.LoadFromReleasePo(releasePo)

	// 9. 获取用户信息
	userIDS := []string{historyPo.UpdateBy}
	userFields := []string{"name"}

	users, err := svc.usermanagementHttpClient.GetUserInfoByUserID(ctx, userIDS, userFields)
	if err != nil {
		svc.Logger.Warnf("get user info failed, err: %v", err)
		err = nil

		return
	}

	if user, ok := users[historyPo.UpdateBy]; ok {
		res.PublishedBy = historyPo.UpdateBy
		res.PublishedByName = user.Name
	}

	return
}

// getLatestVersion 获取 Agent 的最新版本号
// - 如果 releasePo 存在（已发布），使用 releasePo.AgentVersion
// - 如果 releasePo 不存在（未发布/取消发布），从历史记录获取最新版本
// - 如果历史记录也不存在，使用 v0 作为默认值
func (svc *squareSvc) getLatestVersion(ctx context.Context, agentID string, releasePo *dapo.ReleasePO) (latestVersion string, err error) {
	// 1. 如果 releasePo 存在（已发布），使用 releasePo.AgentVersion
	if releasePo != nil {
		latestVersion = releasePo.AgentVersion
		return
	}

	// 2. 从历史记录获取最新版本
	latestHistoryPo, err := svc.releaseHistoryRepo.GetLatestVersionByAgentID(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "[squareSvc.getLatestVersion]:svc.releaseHistoryRepo.GetLatestVersionByAgentID(ctx, %s)", agentID)
		return
	}

	if latestHistoryPo != nil {
		latestVersion = latestHistoryPo.AgentVersion
		return
	}

	// 3. 如果历史记录也不存在，使用 v0 作为默认值
	latestVersion = daconstant.AgentVersionUnpublished

	return
}

// 记录访问日志，用于最近使用展示
func (svc *squareSvc) RecordVisitLog(ctx context.Context, agentInfoReq *squarereq.AgentInfoReq) (err error) {
	// 记录访问日志
	if !agentInfoReq.IsVisit {
		return
	}

	historyAgentVersion := agentInfoReq.AgentVersion
	// 访问历史中只保存 发布版本和未发布版本记录，发布版本统一记录成一条访问记录，防止访问历史记录过多
	if historyAgentVersion != daconstant.AgentVersionUnpublished {
		historyAgentVersion = daconstant.AgentVersionLatest
	}

	currentTs := cutil.GetCurrentMSTimestamp()

	visitHistoryPO := &dapo.VisitHistoryPO{
		ID:           cutil.UlidMake(),
		AgentID:      agentInfoReq.AgentID,
		AgentVersion: historyAgentVersion,
		VisitCount:   1,
		CreateTime:   currentTs,
		UpdateTime:   currentTs,
		CreateBy:     agentInfoReq.UserID,
		UpdateBy:     agentInfoReq.UserID,
	}

	err = svc.visitHistoryRepo.IncVisitCount(ctx, visitHistoryPO)
	if err != nil {
		svc.Logger.Warnf("inc visit count failed, agent id:%v, agent version: %v, err: %v", agentInfoReq.AgentID, agentInfoReq.AgentVersion, err)
	}

	return
}
