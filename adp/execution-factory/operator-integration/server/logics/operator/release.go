package operator

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// 下架操作
func (m *operatorManager) unpublishRelease(ctx context.Context, tx *sql.Tx, operator *model.OperatorRegisterDB, userID string) (err error) {
	exist, releaseDB, err := m.OpReleaseDB.SelectByOpID(ctx, operator.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("select operator release failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select operator release failed")
		return
	}
	if !exist {
		return
	}
	releaseDB.Status = interfaces.BizStatusOffline.String()
	// 下架操作，将当前版本历史记录置为已下架
	has, historyDB, err := m.OpReleaseHistoryDB.SelectByOpIDAndMetdata(ctx, operator.OperatorID, operator.MetadataVersion)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("select operator history failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select operator history failed")
		return
	}
	if has { // 添加历史记录
		historyDB.OpRelease = utils.ObjectToJSON(releaseDB)
		err = m.OpReleaseHistoryDB.UpdateReleaseHistoryByID(ctx, tx, historyDB)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("update operator history failed, OperatorID: %s, err: %v", operator.OperatorID, err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("update operator history failed, OperatorID: %s, err: %s", operator.OperatorID, err.Error()))
		}
	} else {
		err = m.addReleaseHistory(ctx, tx, releaseDB, userID)
	}
	if err != nil {
		return
	}
	err = m.OpReleaseDB.UpdateByOpID(ctx, tx, releaseDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("update operator release failed, OperatorID: %s, err: %v", operator.OperatorID, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("update operator release failed, OperatorID: %s, err: %s", operator.OperatorID, err.Error()))
	}
	return
}

// publishRelease 发布操作
func (m *operatorManager) publishRelease(ctx context.Context, tx *sql.Tx, operator *model.OperatorRegisterDB, userID string) (err error) {
	// 检查是否存在已发布版本
	exist, releaseDB, err := m.OpReleaseDB.SelectByOpID(ctx, operator.OperatorID)
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select operator release failed")
		m.Logger.WithContext(ctx).Errorf("select operator release failed, OperatorID: %s, err: %v", operator.OperatorID, err)
		return
	}

	if exist { // 如果存在，更新记录，并将新的发布记录添加到release_history中
		operatorRegisterToReleaseModel(operator, releaseDB)
		releaseDB.ReleaseUser = userID
		releaseDB.Tag++
		err = m.OpReleaseDB.UpdateByOpID(ctx, tx, releaseDB)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("update operator release failed, OperatorID: %s, err: %v", operator.OperatorID, err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("update operator release failed, OperatorID: %s, err: %s", operator.OperatorID, err.Error()))
			return
		}
	} else { // 如果不存在，添加记录到relase/release_history中
		releaseDB = &model.OperatorReleaseDB{}
		operatorRegisterToReleaseModel(operator, releaseDB)
		releaseDB.ReleaseUser = userID
		releaseDB.Tag++
		err = m.OpReleaseDB.Insert(ctx, tx, releaseDB)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("failed to create new release, OperatorID: %s, err: %v", operator.OperatorID, err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("create new release failed, OperatorID: %s, err: %s", operator.OperatorID, err.Error()))
			return
		}
	}
	err = m.addReleaseHistory(ctx, tx, releaseDB, userID)
	return
}

func (m *operatorManager) addReleaseHistory(ctx context.Context, tx *sql.Tx, releaseDB *model.OperatorReleaseDB, userID string) (err error) {
	now := time.Now().UnixNano()
	historyDB := &model.OperatorReleaseHistoryDB{
		OpID:            releaseDB.OpID,
		MetadataVersion: releaseDB.MetadataVersion,
		MetadataType:    releaseDB.MetadataType,
		OpRelease:       utils.ObjectToJSON(releaseDB),
		Tag:             releaseDB.Tag,
		CreateTime:      now,
		CreateUser:      userID,
		UpdateTime:      now,
		UpdateUser:      userID,
	}
	// 添加记录到release_history中
	err = m.OpReleaseHistoryDB.Insert(ctx, tx, historyDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("failed to insert release history record, OperatorID: %s, err: %v", releaseDB.OpID, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to insert release history record, OperatorID: %s, err: %s", releaseDB.OpID, err.Error()))
	}
	return
}

// operatorRegisterToReleaseModel 注册配置到发布
func operatorRegisterToReleaseModel(operator *model.OperatorRegisterDB, release *model.OperatorReleaseDB) {
	release.OpID = operator.OperatorID
	release.Name = operator.Name
	release.MetadataVersion = operator.MetadataVersion
	release.MetadataType = operator.MetadataType
	release.OperatorType = operator.OperatorType
	release.ExecutionMode = operator.ExecutionMode
	release.ExecuteControl = operator.ExecuteControl
	release.ExtendInfo = operator.ExtendInfo
	release.Source = operator.Source
	release.Category = operator.Category
	release.Status = operator.Status
	release.CreateUser = operator.CreateUser
	release.CreateTime = operator.CreateTime
	release.UpdateUser = operator.UpdateUser
	release.UpdateTime = operator.UpdateTime
	release.IsInternal = operator.IsInternal
	release.IsDataSource = operator.IsDataSource
}
