package versions

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
)

//go:generate mockgen -package mock_dag_versions -source ../../logics/versions/dag_versions.go -destination ../../tests/mock_logics/mock_dag_versions/dag_versions_mock.go

type DagVersionService interface {
	ListDagVersions(ctx context.Context, dagID string, userInfo *drivenadapters.UserInfo) ([]DagVersionSimple, error)
	GetNextVersion(ctx context.Context, dagID string) (string, error)
	RevertToVersion(ctx context.Context, params RevertDagReq, userInfo *drivenadapters.UserInfo) (string, error)
}

type dagVersion struct {
	mongo     mod.Store
	usermgnt  drivenadapters.UserManagement
	permCheck perm.PermCheckerService
}

var dOnce sync.Once
var d DagVersionService

// DagVersionSimple 历史版本简单信息
type DagVersionSimple struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	VersionID string `json:"version_id"`
	ChangeLog string `json:"change_log"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	CreatedAt int64  `json:"created_at"`
}

// RevertDagReq 版本回退请求体参数
type RevertDagReq struct {
	DagID     string `json:"dag_id"`
	VersionID string `json:"version_id"`
	Title     string `json:"title"`
	Version   string `json:"version"`
	ChangeLog string `json:"change_log"`
}

func NewDagVersionService() DagVersionService {
	dOnce.Do(func() {
		dIns := &dagVersion{
			mongo:     mod.GetStore(),
			usermgnt:  drivenadapters.NewUserManagement(),
			permCheck: perm.NewPermCheckerService(),
		}

		d = dIns
	})

	return d
}

// ListDagVersions 列举历史版本
func (d *dagVersion) ListDagVersions(ctx context.Context, dagID string, userInfo *drivenadapters.UserInfo) ([]DagVersionSimple, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	res := []DagVersionSimple{}
	// 判断是否具有查看数据流详情权限
	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow: {perm.ViewOperation},
		},
	}

	_, err = d.permCheck.CheckDagAndPerm(ctx, dagID, userInfo, opMap)
	if err != nil {
		return res, err
	}

	input := &mod.ListDagVersionInput{
		DagID:  dagID,
		Limit:  50,
		Offset: 0,
		Order:  -1,
		SortBy: "sortTime",
	}

	versions, err := d.mongo.ListDagVersions(ctx, input)
	if err != nil {
		log.Warnf("[logic.ListDagVersions] ListDagVersions err, detail: %s", err.Error())
		return res, err
	}

	var accessorIDs = make(map[string]string)

	for _, version := range versions {
		accessorIDs[version.UserID] = common.User.ToString()
	}

	accessors, _ := d.usermgnt.GetNameByAccessorIDs(accessorIDs)

	for _, version := range versions {
		res = append(res, DagVersionSimple{
			ID:        version.DagID,
			Version:   version.Version.ToString(),
			VersionID: version.VersionID,
			ChangeLog: version.ChangeLog,
			UserID:    version.UserID,
			UserName:  accessors[version.UserID],
			CreatedAt: version.CreatedAt,
		})
	}

	return res, nil
}

// GetNextVersion 获取最新版本号
func (d *dagVersion) GetNextVersion(ctx context.Context, dagID string) (string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	_, err = d.mongo.GetDag(ctx, dagID)
	if err != nil {
		if aerr.IsNotFoundErr(err) {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorNotFound, ierr.PErrorNotFound, map[string]interface{}{"dag_id": dagID})
		}
		log.Warnf("[logic.GetLatestDagVersion] GetDag err, detail: %s", err.Error())
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	dagVersion, err := d.GetLatestDagVersion(ctx, dagID)
	if err != nil {
		return "", err
	}

	version := entity.Version(common.DefaultDagVersion)
	if dagVersion != nil {
		version = dagVersion.Version
	}

	return version.GetNextVersion()
}

func (d *dagVersion) GetLatestDagVersion(ctx context.Context, dagID string) (*entity.DagVersion, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	input := &mod.ListDagVersionInput{
		DagID:  dagID,
		Limit:  1,
		Offset: 0,
		Order:  -1,
		SortBy: "sortTime",
	}

	versions, err := d.mongo.ListDagVersions(ctx, input)
	if err != nil {
		log.Warnf("[logic.GetLatestDagVersion] ListDagVersions err, detail: %s", err.Error())
		return nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	if len(versions) == 0 {
		return nil, nil
	}

	return &versions[0], nil
}

// RevertToVersion 回退指定版本
func (d *dagVersion) RevertToVersion(ctx context.Context, params RevertDagReq, userInfo *drivenadapters.UserInfo) (string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 判断是否具有编辑权限
	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow: {perm.ModeifyOperation},
			common.DagTypeDefault:  {perm.OldAdminOperation, perm.OldOwnerOperation},
		},
	}

	_, err = d.permCheck.CheckDagAndPerm(ctx, params.DagID, userInfo, opMap)
	if err != nil {
		return "", err
	}
	// 判断目标版本是否存在
	historyDag, err := d.mongo.GetHistoryDagByVersionID(ctx, params.DagID, params.VersionID)
	if err != nil {
		if aerr.IsNotFoundErr(err) {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorNotFound, aerr.DescKeyTaskNotFound, map[string]string{"dagId": params.DagID, "version": params.VersionID})
		}
		log.Warnf("[logic.RevertToVersion] GetHistoryDagByVersionID err, deail: %s", err.Error())
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	// 判断当前版本数据流名称是否已被其他流程引用
	prevDag, err := historyDag.Config.ParseToDag()
	if err != nil {
		log.Warnf("[logic.RevertToVersion] ParseToDag err, deail: %s", err.Error())
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}
	title := params.Title
	if params.Title == "" {
		title = prevDag.Name
	}
	dagInfo, err := d.mongo.GetDagByFields(ctx, map[string]interface{}{"name": title})
	if err != nil && !aerr.IsNotFoundErr(err) {
		log.Warnf("[logic.RevertToVersion] GetDagByFields err, deail: %s", err.Error())
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
	}
	if dagInfo != nil && dagInfo.ID != params.DagID {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorConflict, ierr.PErrorConflict, map[string]interface{}{"name": title})
	}
	prevDag.Name = title

	// 校验版本是否合规，若未填写则使用最新版本+1
	version := params.Version
	dagVersion, err := d.GetLatestDagVersion(ctx, params.DagID)
	if err != nil {
		return "", err
	}

	if params.Version == "" {
		version, err = dagVersion.Version.GetNextVersion()
		if err != nil {
			return "", err
		}
	} else {
		semver, err := entity.Version(version).Compare(dagVersion.Version)
		if err != nil {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
		}

		if semver < 1 {
			return "", ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, aerr.IllegalSemverVersion, map[string]interface{}{
				"version": "new version must be greater than old version",
				"latest":  params.Version,
				"prev":    dagVersion.Version,
			})
		}
	}
	prevDag.Version = entity.Version(version)

	newVersID, _ := utils.GetUniqueIDStr()
	prevDag.VersionID = newVersID
	prevDag.ModifyBy = userInfo.UserID

	config, err := json.Marshal(prevDag)
	if err != nil {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
	}

	newDagVersion := &entity.DagVersion{
		DagID:     params.DagID,
		UserID:    userInfo.UserID,
		Version:   entity.Version(version),
		VersionID: newVersID,
		ChangeLog: params.ChangeLog,
		Config:    entity.Config(config),
		SortTime:  time.Now().UnixNano(),
	}

	// 如果变更说明为空，生成默认变更说明
	if params.ChangeLog == "" {
		newDagVersion.ChangeLog = fmt.Sprintf("版本: %v 退回至 版本: %v", dagVersion.Version, historyDag.Version)
	}

	err = d.mongo.WithTransaction(ctx, func(nctx context.Context, ms mod.Store) error {
		// 更新dag
		if err = ms.UpdateDag(nctx, prevDag); err != nil {
			log.Warnf("[logic.RevertToVersion] UpdateDag err, detail: %s", err.Error())
			return err
		}

		_, err = ms.CreateDagVersion(nctx, newDagVersion)
		if err != nil {
			log.Warnf("[logic.RevertToVersion] CreateDagVersion err, detail: %s", err.Error())
			return err
		}

		return nil
	})
	if err != nil {
		return "", ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
	}

	return newVersID, nil
}
