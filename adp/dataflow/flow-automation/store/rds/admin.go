package rds

import (
	"context"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	cdb "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

var (
	caOnce sync.Once
	ca     rds.ContentAmdinDao
)

type caDB struct {
	db *gorm.DB
}

func NewContentAmdin() rds.ContentAmdinDao {
	caOnce.Do(func() {
		ca = &caDB{
			db: cdb.NewDB(),
		}
	})

	return ca
}

func (ca *caDB) CreateAdmin(ctx context.Context, datas []*rds.ContentAdmin) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(datas)

	var queryBuilder strings.Builder
	queryBuilder.WriteString("INSERT INTO t_content_admin (f_id, f_user_id, f_user_name) VALUES ")
	values := make([]interface{}, 0, len(datas)*3)
	for i, data := range datas {
		if i > 0 {
			queryBuilder.WriteString(",")
		}
		queryBuilder.WriteString("(?, ?, ?)")
		values = append(values, data.ID, data.UserID, data.UserName)
	}
	sql := queryBuilder.String()
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.CONTENT_ADMIN_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ca.db.Exec(sql, values...).Error
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[caDB.CreateAdmin] create content admin failed, detail: %s", err.Error())
	}
	return err
}

func (ca *caDB) CheckAdminExistByUSerID(ctx context.Context, userID string) (bool, error) {
	var (
		err   error
		count int64
	)
	newCtx, span := trace.StartInternalSpan(ctx)

	sql := "SELECT COUNT(f_id) FROM t_content_admin WHERE f_user_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.CONTENT_ADMIN_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, userID))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ca.db.Raw(sql, userID).Scan(&count).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[caDB.CheckAdminExistByUSerID] check admin exist by userid failed, detail: %s", err.Error())
		return false, err
	}
	if count <= 0 {
		return false, nil
	}
	return true, nil
}

func (ca *caDB) ListAdmins(ctx context.Context) ([]rds.ContentAdmin, error) {
	var (
		err    error
		admins []rds.ContentAdmin
	)
	newCtx, span := trace.StartInternalSpan(ctx)

	sql := "SELECT f_id, f_user_id, f_user_name FROM t_content_admin"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.CONTENT_ADMIN_TABLENAME), attribute.String(trace.DB_SQL, sql))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ca.db.Raw(sql).Scan(&admins).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[caDB.ListAdmin] list admin failed, detail: %s", err.Error())
		return admins, err
	}

	return admins, nil
}

func (ca *caDB) ListAdminsByUserID(ctx context.Context, userIDs []string) ([]rds.ContentAdmin, error) {
	var (
		err    error
		admins []rds.ContentAdmin
	)
	newCtx, span := trace.StartInternalSpan(ctx)

	sql := "SELECT f_id, f_user_id, f_user_name FROM t_content_admin WHERE f_user_id IN ?"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.CONTENT_ADMIN_TABLENAME), attribute.String(trace.DB_SQL, sql))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ca.db.Raw(sql, userIDs).Scan(&admins).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[caDB.ListAdmin] list admin failed, detail: %s", err.Error())
		return admins, err
	}

	return admins, nil
}

func (ca *caDB) DeleteAdminByID(ctx context.Context, ID string) error {
	var err error

	newCtx, span := trace.StartInternalSpan(ctx)

	sql := "DELETE FROM t_content_admin WHERE f_id= ?"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.CONTENT_ADMIN_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, ID))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ca.db.Exec(sql, ID).Error
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[caDB.DeleteAdminByID] delete admin by id failed, detail: %s", err.Error())
	}
	return err
}

func (ca *caDB) UpdateAdminByUserID(ctx context.Context, userID, userName string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)

	sql := "UPDATE t_content_admin SET f_user_name = ? WHERE f_user_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.CONTENT_ADMIN_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, userID))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ca.db.Exec(sql, userName, userID).Error
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[caDB.UpdateAdminByUserID] update admin by userid failed, detail: %s", err.Error())
	}
	return err
}
