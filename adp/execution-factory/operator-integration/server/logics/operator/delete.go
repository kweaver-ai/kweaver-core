package operator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// DeleteOperator 删除算子
func (m *operatorManager) DeleteOperator(ctx context.Context, req interfaces.OperatorDeleteReq, userID string) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 待删除算子校验
	if len(req) == 0 {
		return errors.DefaultHTTPError(ctx, http.StatusBadRequest, "operator delete list is empty")
	}
	operatorIDs := []string{}
	for _, item := range req {
		operatorIDs = append(operatorIDs, item.OperatorID)
	}
	// 检查算子是否全部存在
	operatorList, err := m.DBOperatorManager.SelectByOperatorIDs(ctx, operatorIDs)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("select operator failed, OperatorIDs: %v, err: %v", operatorIDs, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select operator failed")
		return err
	}
	if len(operatorList) == 0 || len(operatorList) != len(operatorIDs) {
		err = errors.DefaultHTTPError(ctx, http.StatusNotFound, "operator not found")
		return err
	}
	accessor, err := m.AuthService.GetAccessor(ctx, userID)
	if err != nil {
		return
	}
	checkOperatorIDs, err := m.AuthService.ResourceFilterIDs(ctx, accessor, operatorIDs, interfaces.AuthResourceTypeOperator,
		interfaces.AuthOperationTypeDelete)
	if err != nil {
		return
	}
	if len(checkOperatorIDs) != len(operatorList) {
		err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtOperatorDeleteForbidden,
			fmt.Sprintf("current user %s has no permission to delete operator %v", userID, operatorIDs))
		return
	}
	for _, operator := range operatorList {
		// 只有未发布、已下架算子可以删除
		if operator.Status != string(interfaces.BizStatusUnpublish) && operator.Status != string(interfaces.BizStatusOffline) {
			return errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOperatorDeleteForbidden,
				fmt.Sprintf("current operator status %s, can not be deleted", operator.Status))
		}
	}
	// 获取事务
	tx, err := m.DBTx.GetTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				m.Logger.Errorf("rollback failed, err: %v", e)
			}
		} else {
			e := tx.Commit()
			if e != nil {
				m.Logger.Errorf("commit failed, err: %v", e)
			}
		}
	}()
	deleteList := []string{}
	for _, item := range operatorList {
		err = m.deleteOperator(ctx, tx, item, accessor)
		if err != nil {
			return
		}
		deleteList = append(deleteList, item.OperatorID)
	}
	// 取消关联业务域
	businessDomainID, _ := common.GetBusinessDomainFromCtx(ctx)
	err = m.BusinessDomainService.BatchDisassociateResource(ctx, businessDomainID, deleteList, interfaces.AuthResourceTypeOperator)
	if err != nil {
		return
	}
	// 删除资源权限策略
	err = m.AuthService.DeletePolicy(ctx, deleteList, interfaces.AuthResourceTypeOperator)
	return
}

func (m *operatorManager) deleteOperator(ctx context.Context, tx *sql.Tx, item *model.OperatorRegisterDB, accessor *interfaces.AuthAccessor) (err error) {
	// 删除线上版本
	err = m.OpReleaseDB.DeleteByOpID(ctx, tx, item.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("delete operator release failed, OperatorID: %s, err: %v", item.OperatorID, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "delete operator release failed")
		return
	}
	// 获取待删除历史版本
	histories, err := m.OpReleaseHistoryDB.SelectByOpID(ctx, item.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("select operator history failed, OperatorID: %s, err: %v", item.OperatorID, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select operator history failed")
		return
	}
	metadataList := []string{item.MetadataVersion}
	if len(histories) > 0 {
		for _, historyDB := range histories {
			metadataList = append(metadataList, historyDB.MetadataVersion)
		}
		// 删除历史版本
		err = m.OpReleaseHistoryDB.DeleteByOpID(ctx, tx, item.OperatorID)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("delete operator history failed, OperatorID: %s, err: %v", item.OperatorID, err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "delete operator history failed")
			return
		}
	}
	// 删除元数据
	err = m.MetadataService.BatchDeleteMetadata(ctx, tx, interfaces.MetadataType(item.MetadataType), metadataList)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("delete metadata failed, MetadataVersions: %v, err: %v", metadataList, err)
		return
	}
	// 删除注册表
	err = m.DBOperatorManager.DeleteByOperatorID(ctx, tx, item.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("delete operator failed, OperatorID: %s, err: %v", item.OperatorID, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "delete operator failed")
		return
	}
	err = m.publishOperatorDeleteEvent(ctx, item, accessor.ID)
	if err != nil {
		return
	}
	go func(operator *model.OperatorRegisterDB) {
		// 发送删除算子审计日志
		accountAuth, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			m.Logger.WithContext(ctx).Errorf("get account auth context from ctx failed")
			return
		}
		m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuth.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationDelete,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectOperator,
				Name: operator.Name,
				ID:   operator.OperatorID,
			},
		})
	}(item)
	return
}

// 发送删除事件通知
func (m *operatorManager) publishOperatorDeleteEvent(ctx context.Context, operatorDB *model.OperatorRegisterDB, updateUser string) (err error) {
	// 通知删除
	extendInfo := map[string]interface{}{}
	if operatorDB.ExtendInfo != "" {
		err = json.Unmarshal([]byte(operatorDB.ExtendInfo), &extendInfo)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("unmarshal operator extend info failed, OperatorID: %s, err: %v", operatorDB.OperatorID, err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "unmarshal operator extend info failed")
			return
		}
	}
	err = m.MQClient.Publish(ctx, interfaces.OperatorDeleteEventTopic, utils.ObjectToByte(&interfaces.OperatorDeleteEvent{
		OperatorID:   operatorDB.OperatorID,
		Version:      operatorDB.MetadataVersion,
		Status:       interfaces.BizStatus(operatorDB.Status),
		IsInternal:   operatorDB.IsInternal,
		OperatorType: interfaces.OperatorType(operatorDB.OperatorType),
		IsDataSource: operatorDB.IsDataSource,
		ExtendInfo:   extendInfo,
		UpdateUser:   updateUser,
	}))
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("publish operator delete event failed, OperatorID: %s, err: %v", operatorDB.OperatorID, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, "publish operator delete event failed")
	}
	return
}
