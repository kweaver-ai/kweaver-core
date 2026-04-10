package rds

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	cdb "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/db"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type UpdateParams = rds.UpdateParams
type UpdateCondition = rds.UpdateCondition
type QueryCondition = rds.QueryCondition
type ListParams = rds.ListParams

var (
	amOnce sync.Once
	am     rds.AiModelDao
)

type aiDB struct {
	db *gorm.DB
}

func NewAiModel() rds.AiModelDao {
	amOnce.Do(func() {
		am = &aiDB{
			db: cdb.NewDB(),
		}
	})

	return am
}

func (ai *aiDB) GetModelInfoByID(ctx context.Context, conditions *QueryCondition) (rds.AiModel, error) {
	var (
		err error
		log rds.AiModel
	)

	newCtx, span := trace.StartInternalSpan(ctx)
	conParts, cons := getUpdateSql(*conditions)
	whereClause := strings.Join(conParts, " and ")
	sql := fmt.Sprintf("SELECT f_id, f_name, f_description, f_train_status, f_status, f_rule, f_userid, f_type, f_created_at, f_updated_at from t_model where %s", whereClause)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", cons)))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ai.db.Raw(sql, cons...).Scan(&log).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[aiDB.GetModelInfoByID] get model info by id failed, detail: %s", err.Error())
		return log, err
	}
	if log.ID == 0 {
		return log, gorm.ErrRecordNotFound
	}
	return log, err
}

func (ai *aiDB) ListModelInfo(ctx context.Context, params *ListParams, offset, limit int64) ([]rds.AiModel, error) {
	var (
		err       error
		sql       string
		tx        *gorm.DB
		tagsRules []rds.AiModel
		rawParam  []interface{}
	)
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql = "select f_id, f_status, f_name, f_userid, f_type, f_created_at, f_updated_at from t_model"
	if params != nil {
		sql = fmt.Sprintf("%v where", sql)
		if params.UserID != nil {
			sql = fmt.Sprintf("%v and f_userid = ?", sql)
			rawParam = append(rawParam, *params.UserID)
		}
		if params.Status != nil {
			sql = fmt.Sprintf("%v and f_status = ?", sql)
			rawParam = append(rawParam, *params.Status)
		} else {
			sql = fmt.Sprintf("%v and f_status > -1", sql)
		}
	}
	sql = strings.Replace(sql, "and ", "", 1)
	if limit == -1 {
		tx = ai.db.Raw(sql, rawParam...)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql))
	} else {
		rawParam = append(rawParam, offset, limit)
		sql = fmt.Sprintf("%v %v", sql, "limit ?, ?")
		tx = ai.db.Raw(sql, rawParam...)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", rawParam...)))
	}
	err = tx.Scan(&tagsRules).Error
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[aiDB.ListTagsRule] list model info failed, detail: %s", err.Error())
	}
	return tagsRules, err
}

func (ai *aiDB) DeleteModelInfoByID(ctx context.Context, conditions *QueryCondition) error {
	var (
		err error
	)

	newCtx, span := trace.StartInternalSpan(ctx)
	conParts, cons := getUpdateSql(*conditions)
	whereClause := strings.Join(conParts, " and ")
	sql := fmt.Sprintf("delete from t_model where %s", whereClause)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", cons)))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ai.db.Exec(sql, cons...).Error
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[aiDB.DeleteModelInfoByID] delete model info failed, detail: %s", err.Error())
	}
	return err
}

func (ai *aiDB) UpdateModelInfo(ctx context.Context, conditions *UpdateCondition, data *UpdateParams) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(data)

	if data == nil || conditions == nil {
		return nil
	}
	setParts, args := getUpdateSql(*data)
	conParts, cons := getUpdateSql(*conditions)

	setClause := strings.Join(setParts, ", ")
	whereClause := strings.Join(conParts, " and ")
	setClause = fmt.Sprintf("%s, f_updated_at = ?", setClause)
	sql := fmt.Sprintf("update t_model set %s where %s", setClause, whereClause)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	updatedAt := time.Now().UnixMicro() / 1000
	args = append(args, updatedAt)
	err = ai.db.Exec(sql, append(args, cons...)...).Error
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[db.UpdateModelInfo] update tags rule failed, detail: %s", err.Error())
	}
	return err
}

func (ai *aiDB) CreateTagsRule(ctx context.Context, data *rds.AiModel) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(data)

	sql := "insert into t_model (f_id, f_rule, f_status, f_name, f_description, f_userid, f_type, f_created_at, f_updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ai.db.Exec(sql, data.ID, data.Rule, data.Status, data.Name, data.Description, data.UserID, data.Type, data.CreatedAt, data.UpdatedAt).Error
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[db.CreateTagsRule] create tag rule failed, detail: %s", err.Error())
	}
	return err
}

func (ai *aiDB) GetInferSchema(ctx context.Context, trainID string) (string, error) {
	var (
		err    error
		schema string
	)
	newCtx, span := trace.StartInternalSpan(ctx)
	sql := "select f_rule from t_model where f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, trainID))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ai.db.Raw(sql, trainID).Scan(&schema).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[db.GetInferSchema] get schema failed, detail: %s", err.Error())
	}
	return schema, err
}

func (ai *aiDB) CreateTrainFile(ctx context.Context, data *rds.AiModel, trainFile *rds.TrainFileOSSInfo) error {
	var err error
	tx := ai.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(data)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := "insert into t_model (f_id, f_train_status, f_status, f_rule, f_type, f_created_at, f_updated_at) values (?, ?, ?, ?, ?, ?, ?)"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))
	err = tx.Exec(sql, data.ID, data.TrainStatus, data.Status, data.Rule, data.Type, data.CreatedAt, data.UpdatedAt).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[db.CreateTrainLog] insert train log failed, detail: %s", err.Error())
		tx.Rollback()
		return err
	}

	sql = "insert into t_train_file (f_id, f_train_id, f_oss_id, f_key, f_created_at) values (?, ?, ?, ?, ?)"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME))
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))
	err = tx.Exec(sql, trainFile.ID, trainFile.TrainID, trainFile.OSSID, trainFile.Key, trainFile.CreatedAt).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[db.CreateTrainLog] insert train file info failed, detail: %s", err.Error())
		return err
	}

	return tx.Commit().Error
}

func (ai *aiDB) GetTrainFileInfo(ctx context.Context, trainID string) (rds.TrainFileOSSInfo, error) {
	var (
		err  error
		info rds.TrainFileOSSInfo
	)

	newCtx, span := trace.StartInternalSpan(ctx)
	sql := "select f_id, f_train_id, f_oss_id, f_key from t_train_file where f_train_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, trainID))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ai.db.Raw(sql, trainID).Scan(&info).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[db.GetTrainFileInfo] get train file info failed, detail: %s", err.Error())
		return info, err
	}

	if info.TrainID == 0 {
		return info, gorm.ErrRecordNotFound
	}
	return info, err
}

func (ai *aiDB) UpdateTrainLog(ctx context.Context, data *rds.AiModel) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(data)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sql := "update t_model set f_train_status = ?, f_rule = ?, f_userid = ?, f_updated_at = ? where f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	err = ai.db.Exec(sql, data.TrainStatus, data.Rule, data.UserID, data.UpdatedAt, data.ID).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[db.UpdateTrainLog] update train log failed, detail: %s", err.Error())
	}
	return err
}

func (ai *aiDB) CheckDupName(ctx context.Context, name string) (bool, error) {
	var (
		err   error
		count int64
	)

	newCtx, span := trace.StartInternalSpan(ctx)
	sql := "select count(*) from t_model where f_name = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, name))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ai.db.Raw(sql, name).Scan(&count).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[db.CheckDupName] get model name count failed, detail: %s", err.Error())
		return false, err
	}

	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (ai *aiDB) VerifyTaskUnique(ctx context.Context) (bool, error) {
	var (
		err   error
		count int64
	)

	newCtx, span := trace.StartInternalSpan(ctx)
	sql := "select count(*) from t_model where f_status > -1 and f_type = 1"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ai.db.Raw(sql).Scan(&count).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[db.CheckDupName] get model name count failed, detail: %s", err.Error())
		return false, err
	}

	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (ai *aiDB) GetModelTypeByID(ctx context.Context, id string) (int, error) {
	var (
		err      error
		taskType int
	)

	newCtx, span := trace.StartInternalSpan(ctx)
	sql := "select f_type from t_model where f_id = ?"
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, rds.AI_MODEL_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, id))
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	err = ai.db.Raw(sql, id).Scan(&taskType).Error
	if err != nil {
		traceLog.WithContext(newCtx).Warnf("[db.CheckDupName] get model type failed, detail: %s", err.Error())
		return taskType, err
	}

	return taskType, nil
}

func getUpdateSql(data interface{}) ([]string, []interface{}) {
	setParts := []string{}
	args := []interface{}{}
	reflectDataType := reflect.TypeOf(data)
	reflectDataValue := reflect.ValueOf(data)
	for i := 0; i < reflectDataType.NumField(); i++ {
		field := reflectDataType.Field(i)
		value := reflectDataValue.Field(i)
		jsonTag := field.Tag.Get("column")
		if !value.IsNil() {
			setParts = append(setParts, fmt.Sprintf("%s = ?", jsonTag))
			args = append(args, value.Interface())
		}
	}

	return setParts, args
}
