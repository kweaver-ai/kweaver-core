package mgnt

import (
	"encoding/json"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

// BuildDagVersionParams 构建版本记录所需的参数
type BuildDagVersionParams struct {
	OldDagBytes []byte
	Dag         *entity.Dag
	CurVersion  *entity.Version
	ChangeLog   string
	UserID      string
	IsCreate    bool
}

// buildDagVersions 构建待持久化的版本记录列表。
// 统一处理版本号计算、验证以及旧版本自动回填逻辑。
func (m *mgnt) buildDagVersions(params *BuildDagVersionParams) ([]*entity.DagVersion, error) {
	var versions []*entity.DagVersion
	dag := params.Dag

	// 处理回填逻辑：如果是非创建场景且旧数据没有版本，先创建一个 v0.0.0 的历史版本
	if !params.IsCreate && dag.Version == "" {

		vID, _ := utils.GetUniqueIDStr()
		versions = append(versions, &entity.DagVersion{
			DagID:     dag.ID,
			UserID:    dag.UserID,
			Version:   common.DefaultDagVersion,
			VersionID: vID,
			Config:    entity.Config(params.OldDagBytes),
			SortTime:  time.Now().UnixNano(),
		})
	}

	// 确定新版本号
	var curVersion entity.Version
	if params.CurVersion != nil && *params.CurVersion != "" {
		curVersion = *params.CurVersion
	} else if dag.Version == "" {
		curVersion = common.DefaultDagVersion
	}

	preVersion := utils.IfNot(dag.Version == "", common.DefaultDagVersion, dag.Version)

	// 如果是更新操作且用户没有提供新版本号，则自动升级版本
	if !params.IsCreate && (params.CurVersion == nil || *params.CurVersion == "") {
		next, err := preVersion.GetNextVersion()
		if err != nil {
			return nil, ierrors.NewIError(ierrors.InternalError, "", err.Error())
		}
		curVersion = entity.Version(next)
	}

	// 验证版本号是否合法：仅在更新场景下要求新版本必须大于当前版本
	if !params.IsCreate {
		cmp, err := curVersion.Compare(preVersion)
		if err != nil {
			return nil, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"version": err.Error()})
		}

		if cmp < 1 {
			return nil, ierrors.NewIError(ierrors.InvalidParameter, ierrors.IllegalSemverVersion, map[string]interface{}{
				"version": "new version must be greater than old version",
				"latest":  curVersion,
				"prev":    preVersion,
			})
		}
	}

	// 更新 DAG 对象的状态
	dag.Version = curVersion
	dag.VersionID, _ = utils.GetUniqueIDStr()

	// 创建新的版本记录
	config, _ := json.Marshal(dag)
	versions = append(versions, &entity.DagVersion{
		DagID:     dag.ID,
		UserID:    params.UserID,
		Version:   dag.Version,
		VersionID: dag.VersionID,
		ChangeLog: params.ChangeLog,
		Config:    entity.Config(config),
		SortTime:  time.Now().UnixNano(),
	})

	return versions, nil
}
