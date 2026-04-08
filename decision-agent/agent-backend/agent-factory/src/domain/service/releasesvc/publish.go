package releasesvc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/e2p/releasee2p"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/releaseeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/pkg/errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/daconstant"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/release/releasereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/release/releaseresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// Publish implements iv3portdriver.IReleaseSvc.
func (svc *releaseSvc) Publish(ctx context.Context, req *releasereq.PublishReq) (resp *releaseresp.PublishUpsertResp,
	auditloginfo auditlogdto.AgentPublishAuditLogInfo, err error,
) {
	defer func() {
		if err != nil {
			resp = &releaseresp.PublishUpsertResp{}
		}
	}()

	// 1. 准备工具
	// 1.1 get agent config
	agentCfgPo, err := svc.agentConfigRepo.GetByID(ctx, req.AgentID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "agent not found")
			return
		}

		err = errors.Wrapf(err, "get agent config by id failed")

		return
	}

	auditloginfo = auditlogdto.AgentPublishAuditLogInfo{
		ID:   agentCfgPo.ID,
		Name: agentCfgPo.Name,
	}

	// 1.2 检查发布权限
	if !req.IsInternalAPI {
		var hasPms bool

		hasPms, err = svc.isHasPublishPermission(ctx, agentCfgPo)
		if err != nil {
			err = errors.Wrapf(err, "check publish permission failed")
			return
		}

		if !hasPms {
			err = capierr.NewCustom403Err(ctx, apierr.AgentFactoryPermissionForbidden, "do not have publish permission")
			return
		}
	}

	// // 1.3 检查名称是否重复
	// existsByName, err := svc.agentConfigRepo.ExistsByNameExcludeID(ctx, agentCfgPo.Name, req.AgentID)
	// if err != nil {
	// 	return
	// }

	// if existsByName {
	// 	err = capierr.NewCustom409Err(ctx, apierr.DataAgentConfigNameExists, "名称已存在")
	// 	return
	// }

	// 2. 发布开始
	entity := releaseeo.ReleaseEO{
		AgentID: req.AgentID,

		UserID:    req.UserID,
		AgentDesc: req.Description,

		PublishToBes:   req.PublishToBes,
		PublishToWhere: req.PublishToWhere,
	}

	entity.SetIsPmsCtrl(req.PmsControl != nil)

	currentVersion := daconstant.AgentVersionUnpublished

	// 3. get lastest version from release history
	latestReleaseHistory, err := svc.releaseHistoryRepo.
		GetLatestVersionByAgentID(ctx, req.AgentID)
	if err != nil {
		err = errors.Wrapf(err, "get latest version from release history failed")
		return
	}

	if latestReleaseHistory != nil {
		currentVersion = latestReleaseHistory.AgentVersion
	}

	// 4. generate new version
	newVersion, err := svc.generateAgentVersion(currentVersion)
	if err != nil {
		err = errors.Wrapf(err, "generate agent version failed")
		return
	}

	entity.AgentVersion = newVersion

	// 5. check agent config
	daCfgObj := &daconfvalobj.Config{}

	err = cutil.JSON().UnmarshalFromString(agentCfgPo.Config, daCfgObj)
	if err != nil {
		err = errors.Wrapf(err, "unmarshal agent config failed")
		return
	}

	err = daCfgObj.ValObjCheckWithCtx(ctx, req.IsInternalAPI)
	if err != nil {
		err = capierr.NewCustom400Err(ctx, apierr.PublishFailedByConfigError, "invalid agent config")
		return
	}

	// 6. marshal agent config
	agentCfgPoJsonStr, err := json.Marshal(agentCfgPo)
	if err != nil {
		err = errors.Wrapf(err, "marshal agent config failed")
		return
	}

	entity.AgentConfig = string(agentCfgPoJsonStr)

	// 7. get release by agent id
	po, err := svc.releaseRepo.GetByAgentID(ctx, req.AgentID)
	if err != nil {
		err = errors.Wrapf(err, "get release by agent id failed")
		return
	}

	currentTs := cutil.GetCurrentMSTimestamp()

	// 8. 开启事务
	tx, err := svc.releaseRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "begin transaction failed")
		return
	}

	defer chelper.TxRollback(tx, &err, svc.Logger)

	// 9. 如果已经发布过，则更新发布记录
	if po != nil {
		po.AgentName = agentCfgPo.Name
		po.AgentConfig = entity.AgentConfig
		po.AgentDesc = entity.AgentDesc
		po.AgentVersion = entity.AgentVersion
		po.UpdateBy = entity.UserID
		po.UpdateTime = currentTs
		po.SetPublishToBes(entity.PublishToBes)
		po.SetPublishToWhere(entity.PublishToWhere)
		po.SetIsPmsCtrl(entity.IsPmsCtrlBool())

		err = svc.releaseRepo.Update(ctx, tx, po)
		if err != nil {
			err = errors.Wrapf(err, "update release failed")
			return
		}
	} else {
		// 如果没有查到结果，表示首次发布
		po = releasee2p.ReleaseE2P(&entity)
		po.AgentName = agentCfgPo.Name
		po.CreateBy = entity.UserID
		po.CreateTime = currentTs
		po.UpdateBy = entity.UserID
		po.UpdateTime = currentTs
		po.SetPublishToBes(entity.PublishToBes)
		po.SetPublishToWhere(entity.PublishToWhere)
		po.SetIsPmsCtrl(entity.IsPmsCtrlBool())

		_, err = svc.releaseRepo.Create(ctx, tx, po)
		if err != nil {
			err = errors.Wrapf(err, "create release failed")
			return
		}
	}

	// 11. 设置响应
	resp = &releaseresp.PublishUpsertResp{
		ReleaseId:   po.ID,
		Version:     po.AgentVersion,
		PublishedAt: po.UpdateTime,
		PublishedBy: po.UpdateBy,
	}

	err = resp.FillPublishedByName(ctx, svc.umHttp)
	if err != nil {
		err = errors.Wrapf(err, "fill published by name failed")
		return
	}

	// 12. 绑定分类
	err = svc.handleCategory(ctx, req.CategoryIDs, po.ID, tx)
	if err != nil {
		err = errors.Wrapf(err, "handle category failed")
		return
	}

	// 13. 处理权限控制
	err = svc.handlePmsCtrl(ctx, req.PmsControl, po.ID, po.AgentID, tx)
	if err != nil {
		err = errors.Wrapf(err, "handle pms ctrl failed")
		return
	}

	// 14. 创建发布历史
	historyPo := &dapo.ReleaseHistoryPO{
		AgentID:      po.AgentID,
		AgentVersion: po.AgentVersion,
		AgentDesc:    po.AgentDesc,
		AgentConfig:  po.AgentConfig,

		CreateTime: currentTs,
		CreateBy:   entity.UserID,
		UpdateTime: currentTs,
		UpdateBy:   entity.UserID,
	}

	_, err = svc.releaseHistoryRepo.Create(ctx, tx, historyPo)
	if err != nil {
		err = errors.Wrapf(err, "create release history failed")
		return
	}

	// 15. 更新Agent状态
	agentCfgPo.Status = cdaenum.StatusPublished

	err = svc.agentConfigRepo.UpdateStatus(ctx, tx, agentCfgPo.Status, agentCfgPo.ID, req.UserID)
	if err != nil {
		err = errors.Wrapf(err, "update agent status to published failed")
		return
	}

	err = tx.Commit()
	if err != nil {
		err = errors.Wrapf(err, "commit transaction failed")
		return
	}

	return
}

func (svc *releaseSvc) generateAgentVersion(oldVersion string) (string, error) {
	// 去掉前缀 'v'
	versionNumberStr := strings.TrimPrefix(oldVersion, "v")

	// 将字符串转换为整数
	versionNumber, err := strconv.Atoi(versionNumberStr)
	if err != nil {
		return "", fmt.Errorf("invalid version format: %s", oldVersion)
	}

	// 增加版本号
	newVersionNumber := versionNumber + 1

	// 组合成新的版本号字符串
	newVersion := fmt.Sprintf("v%d", newVersionNumber)

	return newVersion, nil
}
