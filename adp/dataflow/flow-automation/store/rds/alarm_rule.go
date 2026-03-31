package rds

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	cdb "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/db"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

var (
	arOnce sync.Once
	ar     rds.AlarmRuleDao
)

type alarmRule struct {
	dbType string
	db     *gorm.DB
}

func NewAlarmRule() rds.AlarmRuleDao {
	arOnce.Do(func() {
		ar = &alarmRule{
			dbType: strings.ToUpper(os.Getenv("DB_TYPE")),
			db:     cdb.NewDB(),
		}
	})

	return ar
}

func (ar *alarmRule) ModifyAlarmRule(ctx context.Context, ruleID string, rules []*rds.AlarmRule, users []*rds.AlarmUser) error {
	if len(rules) == 0 || len(users) == 0 {
		return nil
	}
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(rules)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	tx := ar.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if ruleID != "" {
		sqlStr := `DELETE FROM t_alarm_rule WHERE f_rule_id = ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.ALARM_RULE_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", ruleID)))

		err = tx.Exec(sqlStr, ruleID).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		sqlStr = `DELETE FROM t_alarm_user WHERE f_rule_id = ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.ALARM_RULE_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", ruleID)))

		err = tx.Exec(sqlStr, ruleID).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	sqlStr := `INSERT INTO t_alarm_rule (f_id, f_rule_id, f_dag_id, f_frequency, f_threshold, f_created_at) VALUES `
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.ALARM_RULE_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))
	values := make([]interface{}, 0, len(rules)*6)
	for _, data := range rules {
		sqlStr += "(?, ?, ?, ?, ?, ?),"
		values = append(values, data.ID, data.RuleID, data.DagID, data.Frequency, data.Threshold, data.CreatedAt)
	}

	sqlStr = sqlStr[:len(sqlStr)-1]

	err = tx.Exec(sqlStr, values...).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	sqlStr = `INSERT INTO t_alarm_user (f_id, f_rule_id, f_user_id, f_user_name, f_user_type) VALUES `
	msgStr, _ = jsoniter.MarshalToString(users)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.ALARM_RULE_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))

	values = make([]interface{}, 0, len(users)*5)
	for _, data := range users {
		sqlStr += "(?, ?, ?, ?, ?),"
		values = append(values, data.ID, data.RuleID, data.UserID, data.UserName, data.UserType)
	}
	sqlStr = sqlStr[:len(sqlStr)-1]

	err = tx.Exec(sqlStr, values...).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (ar *alarmRule) GetAlarmRule(ctx context.Context, ruleID string) (*rds.AlarmRule, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr := `SELECT * FROM t_alarm_rule WHERE f_rule_id = ?`
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.ALARM_RULE_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", ruleID)))

	var alarmRule = &rds.AlarmRule{}
	err = ar.db.Raw(sqlStr, ruleID).Scan(alarmRule).Error
	if err != nil {
		return nil, err
	}
	if alarmRule.ID == uint64(0) {
		return nil, gorm.ErrRecordNotFound
	}

	return alarmRule, nil
}

func (ar *alarmRule) ListAlarmRule(ctx context.Context, opt *rds.Options) ([]*rds.AlarmRule, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr, searchSqlVal := opt.BuildQuery("SELECT * FROM t_alarm_rule WHERE 1=1")
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", searchSqlVal)))

	var alarmRules = make([]*rds.AlarmRule, 0)
	err = ar.db.Raw(sqlStr, searchSqlVal...).Scan(&alarmRules).Error

	return alarmRules, err
}

func (ar *alarmRule) ListAlarmUser(ctx context.Context, opt *rds.Options) ([]*rds.AlarmUser, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr, searchSqlVal := opt.BuildQuery("SELECT * FROM t_alarm_user WHERE 1=1")
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", searchSqlVal)))

	var alarmUsers = make([]*rds.AlarmUser, 0)
	err = ar.db.Raw(sqlStr, searchSqlVal...).Scan(&alarmUsers).Error

	return alarmUsers, err
}

func (ar *alarmRule) GroupAlarmRule(ctx context.Context) ([]*rds.AlarmRule, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr := "SELECT f_rule_id, f_frequency, f_threshold FROM t_alarm_rule GROUP BY f_rule_id"
	if strings.HasPrefix(ar.dbType, common.DBTYPEKDB) {
		sqlStr = "SELECT f_rule_id, max(f_frequency) as f_frequency, max(f_threshold) as f_threshold FROM t_alarm_rule GROUP BY f_rule_id"
	}
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr))

	var alarmRules = make([]*rds.AlarmRule, 0)
	err = ar.db.Raw(sqlStr).Scan(&alarmRules).Error

	return alarmRules, err
}

func (ar *alarmRule) ListDagIDs(ctx context.Context, ruleID string) ([]string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr := "SELECT f_dag_id FROM t_alarm_rule WHERE f_rule_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", ruleID)))

	var dagIDs = make([]string, 0)
	err = ar.db.Raw(sqlStr, ruleID).Scan(&dagIDs).Error

	return dagIDs, err
}
