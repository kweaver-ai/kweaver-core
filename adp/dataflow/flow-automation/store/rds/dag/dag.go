package dagmodel

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/db"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/event"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	data "github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils/data"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"github.com/shiningrush/goevent"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

const (
	DAG_TABLENAME                = "t_flow_dag"
	DAGINSTANCE_TABLENAME        = "t_flow_dag_instance"
	TASKINSTANCE_TABLENAME       = "t_flow_task_instance"
	DAGVAR_TABLENAME             = "t_flow_dag_var"
	DAGVERSIONS_TABLENAME        = "t_flow_dag_version"
	DAGSTEPINDEX_TABLENAME       = "t_flow_dag_step"
	DAGTRIGGERINDEX_TABLENAME    = "t_flow_dag_trigger_config"
	DAGACCESSORINDEX_TABLENAME   = "t_flow_dag_accessor"
	DAGINSTANCEKEYWORD_TABLENAME = "t_flow_dag_instance_keyword"
	OUTBOXMESSAGE_TABLENAME      = "t_flow_outbox"
	INBOXMESSAGE_TABLENAME       = "t_flow_inbox"
	LOG_TABLENAME                = "t_flow_log"
)

const defaultCtxTimeout = 100 * time.Second

// Dag 流程配置数据库模型
type DagModel struct {
	ID            uint64 `json:"f_id" gorm:"column:f_id;primaryKey"`
	CreatedAt     int64  `json:"f_created_at" gorm:"column:f_created_at"`
	UpdatedAt     int64  `json:"f_updated_at" gorm:"column:f_updated_at"`
	UserID        string `json:"f_user_id" gorm:"column:f_user_id"`
	Name          string `json:"f_name" gorm:"column:f_name"`
	Desc          string `json:"f_desc" gorm:"column:f_desc;type:longtext"`
	Trigger       string `json:"f_trigger" gorm:"column:f_trigger;type:longtext"`
	Cron          string `json:"f_cron" gorm:"column:f_cron"`
	Vars          string `json:"f_vars" gorm:"column:f_vars;type:longtext"`
	Status        string `json:"f_status" gorm:"column:f_status"`
	Tasks         string `json:"f_tasks" gorm:"column:f_tasks;type:longtext"`
	Steps         string `json:"f_steps" gorm:"column:f_steps;type:longtext"`
	Description   string `json:"f_description" gorm:"column:f_description;type:longtext"`
	Shortcuts     string `json:"f_shortcuts" gorm:"column:f_shortcuts;type:longtext"`
	Accessors     string `json:"f_accessors" gorm:"column:f_accessors;type:longtext"`
	Type          string `json:"f_type" gorm:"column:f_type"`
	PolicyType    string `json:"f_policy_type" gorm:"column:f_policy_type"`
	AppInfo       string `json:"f_appinfo" gorm:"column:f_appinfo;type:longtext"`
	Priority      string `json:"f_priority" gorm:"column:f_priority"`
	Removed       bool   `json:"f_removed" gorm:"column:f_removed"`
	Emails        string `json:"f_emails" gorm:"column:f_emails;type:longtext"`
	Template      string `json:"f_template" gorm:"column:f_template"`
	Published     bool   `json:"f_published" gorm:"column:f_published"`
	TriggerConfig string `json:"f_trigger_config" gorm:"column:f_trigger_config;type:longtext"`
	SubIDs        string `json:"f_sub_ids" gorm:"column:f_sub_ids;type:longtext"`
	ExecMode      string `json:"f_exec_mode" gorm:"column:f_exec_mode"`
	Category      string `json:"f_category" gorm:"column:f_category"`
	OutPuts       string `json:"f_outputs" gorm:"column:f_outputs;type:longtext"`
	Instructions  string `json:"f_instructions" gorm:"column:f_instructions;type:longtext"`
	OperatorID    string `json:"f_operator_id" gorm:"column:f_operator_id"`
	IncValues     string `json:"f_inc_values" gorm:"column:f_inc_values;type:longtext"`
	Version       string `json:"f_version" gorm:"column:f_version;type:longtext"`
	VersionID     string `json:"f_version_id" gorm:"column:f_version_id"`
	ModifyBy      string `json:"f_modify_by" gorm:"column:f_modify_by"`
	IsDebug       bool   `json:"f_is_debug" gorm:"column:f_is_debug"`
	DeBugID       string `json:"f_debug_id" gorm:"column:f_debug_id"`
	BizDomainID   string `json:"f_biz_domain_id" gorm:"column:f_biz_domain_id"`
}

// DagVar 流程配置变量数据库模型
type DagVarModel struct {
	ID           uint64 `json:"f_id" gorm:"column:f_id;primaryKey;autoIncrement"`
	DagID        uint64 `json:"f_dag_id" gorm:"column:f_dag_id"`
	VarName      string `json:"f_var_name" gorm:"column:f_var_name"`
	DefaultValue string `json:"f_default_value" gorm:"column:f_default_value;type:text"`
	VarType      string `json:"f_var_type" gorm:"column:f_var_type"`
	Description  string `json:"f_description" gorm:"column:f_description"`
}

// DagInstanceKeyword dag instance keyword table model
type DagInstanceKeywordModel struct {
	ID       uint64 `json:"f_id" gorm:"column:f_id;primaryKey"`
	DagInsID uint64 `json:"f_dag_ins_id" gorm:"column:f_dag_ins_id"`
	Keyword  string `json:"f_keyword" gorm:"column:f_keyword"`
}

type DagStepModel struct {
	ID            uint64 `gorm:"column:f_id;primaryKey"`
	DagID         uint64 `gorm:"column:f_dag_id"`
	Operator      string `gorm:"column:f_operator"`
	SourceID      string `gorm:"column:f_source_id"`
	HasDatasource bool   `gorm:"column:f_has_datasource"`
}

type DagTriggerConfigModel struct {
	ID       uint64 `gorm:"column:f_id;primaryKey"`
	DagID    uint64 `gorm:"column:f_dag_id"`
	Operator string `gorm:"column:f_operator"`
	SourceID string `gorm:"column:f_source_id"`
}

type DagAccessorModel struct {
	ID         uint64 `gorm:"column:f_id;primaryKey"`
	DagID      uint64 `gorm:"column:f_dag_id"`
	AccessorID string `gorm:"column:f_accessor_id"`
}

type DagVersionModel struct {
	ID        uint64 `json:"f_id" gorm:"column:f_id;primaryKey"`
	CreatedAt int64  `json:"f_created_at" gorm:"column:f_created_at"`
	UpdatedAt int64  `json:"f_updated_at" gorm:"column:f_updated_at"`
	DagID     string `json:"f_dag_id" gorm:"column:f_dag_id"`
	UserID    string `json:"f_user_id" gorm:"column:f_user_id"`
	Version   string `json:"f_version" gorm:"column:f_version"`
	VersionID string `json:"f_version_id" gorm:"column:f_version_id"`
	ChangeLog string `json:"f_change_log" gorm:"column:f_change_log"`
	Config    string `json:"f_config" gorm:"column:f_config"`
	SortTime  int64  `json:"f_sort_time" gorm:"column:f_sort_time"`
}

// DagInstance 对应数据库表 t_flow_dag_instance
type DagInstanceModel struct {
	ID               uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt        int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt        int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	DagID            uint64 `gorm:"column:f_dag_id" json:"f_dag_id"`
	Trigger          string `gorm:"column:f_trigger" json:"f_trigger,omitempty"`
	Worker           string `gorm:"column:f_worker" json:"f_worker"`
	Source           string `gorm:"column:f_source" json:"f_source"`
	Vars             string `gorm:"column:f_vars" json:"f_vars,omitempty"`
	Keywords         string `gorm:"column:f_keywords" json:"f_keywords,omitempty"`
	EventPersistence int    `gorm:"column:f_event_persistence" json:"f_event_persistence,omitempty"`
	EventOssPath     string `gorm:"column:f_event_oss_path" json:"f_event_oss_path,omitempty"`
	ShareData        string `gorm:"column:f_share_data" json:"f_share_data,omitempty"`
	ShareDataExt     string `gorm:"column:f_share_data_ext" json:"f_share_data_ext,omitempty"`
	Status           string `gorm:"column:f_status" json:"f_status"`
	Reason           string `gorm:"column:f_reason" json:"f_reason,omitempty"`
	Cmd              string `gorm:"column:f_cmd" json:"f_cmd,omitempty"`
	HasCmd           bool   `gorm:"column:f_has_cmd" json:"f_has_cmd"`
	BatchRunID       string `gorm:"column:f_batch_run_id" json:"f_batch_run_id"`
	UserID           string `gorm:"column:f_user_id" json:"f_user_id"`
	EndedAt          int64  `gorm:"column:f_ended_at" json:"f_ended_at"`
	DagType          string `gorm:"column:f_dag_type" json:"f_dag_type"`
	PolicyType       string `gorm:"column:f_policy_type" json:"f_policy_type"`
	AppInfo          string `gorm:"column:f_appinfo" json:"f_app_info,omitempty"`
	Priority         string `gorm:"column:f_priority" json:"f_priority"`
	Mode             int    `gorm:"column:f_mode" json:"f_mode"`
	Dump             string `gorm:"column:f_dump" json:"f_dump,omitempty"`
	DumpExt          string `gorm:"column:f_dump_ext" json:"f_dump_ext,omitempty"`
	SuccessCallback  string `gorm:"column:f_success_callback" json:"f_success_callback,omitempty"`
	ErrorCallback    string `gorm:"column:f_error_callback" json:"f_error_callback,omitempty"`
	CallChain        string `gorm:"column:f_call_chain" json:"f_call_chain,omitempty"`
	ResumeData       string `gorm:"column:f_resume_data" json:"f_resume_data,omitempty"`
	ResumeStatus     string `gorm:"column:f_resume_status" json:"f_resume_status"`
	Version          string `gorm:"column:f_version" json:"f_version,omitempty"`
	VersionID        string `gorm:"column:f_version_id" json:"f_version_id"`
	BizDomainID      string `gorm:"column:f_biz_domain_id" json:"f_biz_domain_id"`
}

type InBoxModel struct {
	ID        uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	Msg       string `gorm:"column:f_msg" json:"f_msg"`
	Topic     string `gorm:"column:f_topic" json:"f_topic"`
	DocID     string `gorm:"column:f_docid" json:"f_doc_id"`
	Dags      string `gorm:"column:f_dag" json:"f_dags"`
}

type OutBoxModel struct {
	ID        uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	Msg       string `gorm:"column:f_msg" json:"f_msg"`
	Topic     string `gorm:"column:f_topic" json:"f_topic"`
}

type TaskInstanceModel struct {
	ID             uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt      int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt      int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	TaskID         string `gorm:"column:f_task_id" json:"f_task_id"`
	DagInsID       uint64 `gorm:"column:f_dag_ins_id" json:"f_dag_ins_id"`
	Name           string `gorm:"column:f_name" json:"f_name"`
	DependOn       string `gorm:"column:f_depend_on" json:"f_depend_on"`
	ActionName     string `gorm:"column:f_action_name" json:"f_action_name"`
	TimeoutSecs    int    `gorm:"column:f_timeout_secs" json:"f_timeout_secs"`
	Params         string `gorm:"column:f_params" json:"f_params"`
	Traces         string `gorm:"column:f_traces" json:"f_traces"`
	Status         string `gorm:"column:f_status" json:"f_status"`
	Reason         string `gorm:"column:f_reason" json:"f_reason"`
	PreChecks      string `gorm:"column:f_pre_checks" json:"f_pre_checks"`
	Results        string `gorm:"column:f_results" json:"f_results"`
	Steps          string `gorm:"column:f_steps" json:"f_steps"`
	LastModifiedAt int64  `gorm:"column:f_last_modified_at" json:"f_last_modified_at"`
	RenderedParams string `gorm:"column:f_rendered_params" json:"f_rendered_params"`
	Hash           string `gorm:"column:f_hash" json:"f_hash"`
	Settings       string `gorm:"column:f_settings" json:"f_settings"`
	MetaData       string `gorm:"column:f_metadata" json:"f_metadata"`
}

type TokenModel struct {
	ID           uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt    int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt    int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	UserID       string `gorm:"column:f_user_id" json:"f_user_id"`
	UserName     string `gorm:"column:f_user_name" json:"f_user_name"`
	RefreshToken string `gorm:"column:f_refresh_token" json:"f_refresh_token"`
	Token        string `gorm:"column:f_token" json:"f_token"`
	ExpiresIn    int    `gorm:"column:f_expires_in" json:"f_expires_in"`
	LoginIP      string `gorm:"column:f_login_ip" json:"f_login_ip"`
	IsApp        bool   `gorm:"column:f_is_app" json:"f_is_app"`
}

type ClientModel struct {
	ID           uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt    int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt    int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	ClientName   string `gorm:"column:f_client_name" json:"f_client_name"`
	ClientID     string `gorm:"column:f_client_id" json:"f_client_id"`
	ClientSecret string `gorm:"column:f_client_secret" json:"f_client_secret"`
}

type SwitchModel struct {
	ID        uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	Name      string `gorm:"column:f_name" json:"f_name"`
	Status    bool   `gorm:"column:f_status" json:"f_status"`
}

type LogModel struct {
	ID        uint64 `gorm:"column:f_id;primaryKey" json:"f_id"`
	CreatedAt int64  `gorm:"column:f_created_at" json:"f_created_at"`
	UpdatedAt int64  `gorm:"column:f_updated_at" json:"f_updated_at"`
	OssID     string `gorm:"column:f_ossid" json:"f_ossid"`
	Key       string `gorm:"column:f_key" json:"f_key"`
	FileName  string `gorm:"column:f_filename" json:"f_filename"`
}

type dag struct {
	db   *gorm.DB
	isTX bool
}

var (
	dagOnce sync.Once
	dagRep  mod.Store
)

func NewDagRepository() mod.Store {
	dagOnce.Do(func() {
		dagRep = &dag{
			db: db.NewDB(),
		}
	})
	return dagRep
}

func (d *dag) dbWithContextWithTimeout(ctx context.Context, timeout time.Duration) (*gorm.DB, context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}

	newCtx, cancel := context.WithTimeout(ctx, timeout)
	return d.db.WithContext(newCtx), newCtx, cancel
}

func (d *dag) dbWithContext(ctx context.Context) (*gorm.DB, context.Context, context.CancelFunc) {
	return d.dbWithContextWithTimeout(ctx, defaultCtxTimeout)
}

// TransactionWithContext 带 Context 的事务
func (d *dag) WithTransaction(ctx context.Context, fn func(context.Context, mod.Store) error) error {
	db, txCtx, cancel := d.dbWithContext(ctx)
	defer cancel()

	txCli := &dag{
		db:   db.Begin(),
		isTX: true,
	}

	tx := txCli.db
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := fn(txCtx, txCli); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("rollback failed: %s, original: %s", rbErr.Error(), err.Error())
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit failed: %s", err.Error())
	}

	return nil
}

func (d *dag) CreateDag(ctx context.Context, dagEntity *entity.Dag) (string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(store *dag, dagEntity *entity.Dag) error {
		// 准备 SQL 语句，使用参数化查询防止 SQL 注入
		sql := `INSERT INTO t_flow_dag (
			f_id, f_created_at, f_updated_at, f_user_id, f_name, f_desc, f_trigger,
			f_cron, f_vars, f_status, f_tasks, f_steps, f_description, f_shortcuts,
			f_accessors, f_type, f_policy_type, f_appinfo, f_priority, f_removed,
			f_emails, f_template, f_published, f_trigger_config, f_sub_ids, f_exec_mode,
			f_category, f_outputs, f_instructions, f_operator_id, f_inc_values, f_version,
			f_version_id, f_modify_by, f_is_debug, f_debug_id, f_biz_domain_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		// 执行 SQL 语句
		t := ToDagModel(dagEntity, false)
		msgStr, _ := jsoniter.MarshalToString(t)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

		err = dagEntity.CheckRootNode(dagEntity.Tasks)
		if err != nil {
			return err
		}

		err = store.db.Exec(sql,
			t.ID,
			t.CreatedAt,
			t.UpdatedAt,
			t.UserID,
			t.Name,
			t.Desc,
			t.Trigger,
			t.Cron,
			t.Vars,
			t.Status,
			t.Tasks,
			t.Steps,
			t.Description,
			t.Shortcuts,
			t.Accessors,
			t.Type,
			t.PolicyType,
			t.AppInfo,
			t.Priority,
			t.Removed,
			t.Emails,
			t.Template,
			t.Published,
			t.TriggerConfig,
			t.SubIDs,
			t.ExecMode,
			t.Category,
			t.OutPuts,
			t.Instructions,
			t.OperatorID,
			t.IncValues,
			t.Version,
			t.VersionID,
			t.ModifyBy,
			t.IsDebug,
			t.DeBugID,
			t.BizDomainID,
		).Error
		if err != nil {
			return err
		}

		err = store.CreateDagVars(newCtx, BuildDagVars(dagEntity))
		if err != nil {
			return err
		}

		err = store.refreshDagIndexes(newCtx, dagEntity)
		if err != nil {
			return err
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(_ context.Context, txStore mod.Store) error {
			return fn(txStore.(*dag), dagEntity)
		})
	} else {
		err = fn(d, dagEntity)
	}

	return dagEntity.ID, err
}

func (d *dag) CreateDagVars(ctx context.Context, dagVars []*DagVarModel) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(dagVars)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(store *dag, dagVars []*DagVarModel) error {
		if len(dagVars) == 0 {
			return nil
		}

		dagID := dagVars[0].DagID
		sqlStr := `DELETE FROM t_flow_dag_var WHERE f_dag_id = ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", dagID)))
		err = store.db.Exec(sqlStr, dagID).Error
		if err != nil {
			return err
		}

		sqlStr = `INSERT INTO t_flow_dag_var (f_id, f_dag_id, f_var_name, f_default_value, f_var_type, f_description) VALUES `
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))
		values := make([]any, 0, len(dagVars)*5)
		for _, data := range dagVars {
			sqlStr += "(?, ?, ?, ?, ?, ?),"
			values = append(values, data.ID, data.DagID, data.VarName, data.DefaultValue, data.VarType, data.Description)
		}

		sqlStr = sqlStr[:len(sqlStr)-1]

		err = store.db.Exec(sqlStr, values...).Error
		if err != nil {
			return err
		}

		return nil

	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(_ context.Context, txStore mod.Store) error {
			return fn(txStore.(*dag), dagVars)
		})
	} else {
		err = fn(d, dagVars)
	}

	return err
}

func (d *dag) deleteDagInstanceKeywords(ctx context.Context, dagInsID uint64) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sqlStr := `DELETE FROM t_flow_dag_instance_keyword WHERE f_dag_ins_id = ?`
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCEKEYWORD_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", dagInsID)))
	err = db.Exec(sqlStr, dagInsID).Error
	return err
}

func (d *dag) insertDagInstanceKeywords(ctx context.Context, dagInsID uint64, keywords []string) error {
	var err error
	if len(keywords) == 0 {
		return nil
	}
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sqlStr := `INSERT INTO t_flow_dag_instance_keyword (f_id, f_dag_ins_id, f_keyword) VALUES `
	values := make([]any, 0, len(keywords)*3)
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		id, _ := utils.GetUniqueID()
		sqlStr += "(?, ?, ?),"
		values = append(values, id, dagInsID, keyword)
	}
	if len(values) == 0 {
		return nil
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	msgStr, _ := jsoniter.MarshalToString(values)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCEKEYWORD_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))
	err = db.Exec(sqlStr, values...).Error
	return err
}

func (d *dag) insertDagInstanceKeywordsBatch(ctx context.Context, rows []DagInstanceKeywordModel) error {
	var err error
	if len(rows) == 0 {
		return nil
	}
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	batchSize := 1000
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		sqlStr := `INSERT INTO t_flow_dag_instance_keyword (f_id, f_dag_ins_id, f_keyword) VALUES `
		values := make([]any, 0, len(batch)*3)
		for _, row := range batch {
			sqlStr += "(?, ?, ?),"
			values = append(values, row.ID, row.DagInsID, row.Keyword)
		}
		sqlStr = strings.TrimSuffix(sqlStr, ",")

		msgStr, _ := jsoniter.MarshalToString(values)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCEKEYWORD_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))
		if err = db.Exec(sqlStr, values...).Error; err != nil {
			return err
		}
	}

	return nil
}

func (d *dag) replaceDagInstanceKeywords(ctx context.Context, dagInsID uint64, keywords []string) error {
	if err := d.deleteDagInstanceKeywords(ctx, dagInsID); err != nil {
		return err
	}
	return d.insertDagInstanceKeywords(ctx, dagInsID, keywords)
}

func (d *dag) refreshDagIndexes(ctx context.Context, dag *entity.Dag) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if dag == nil {
		return nil
	}

	dagID, parseErr := strconv.ParseUint(dag.ID, 10, 64)
	if parseErr != nil {
		return parseErr
	}

	stepRows := BuildDagStepIndex(dag)
	// triggerRows := BuildDagTriggerConfigIndex(dag)
	accessorRows := BuildDagAccessorIndex(dag)

	deleteIndexes := func(table string) error {
		sqlStr := fmt.Sprintf("DELETE FROM %s WHERE f_dag_id = ?", table)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, table), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, fmt.Sprintf("%v", dagID)))
		return db.Exec(sqlStr, dagID).Error
	}

	if err = deleteIndexes(DAGSTEPINDEX_TABLENAME); err != nil {
		return err
	}
	// if err = deleteIndexes(DAGTRIGGERINDEX_TABLENAME); err != nil {
	// 	return err
	// }
	if err = deleteIndexes(DAGACCESSORINDEX_TABLENAME); err != nil {
		return err
	}

	insertStepRows := func(rows []*DagStepModel) error {
		if len(rows) == 0 {
			return nil
		}
		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id, f_dag_id, f_operator, f_source_id, f_has_datasource) VALUES ", DAGSTEPINDEX_TABLENAME)
		values := make([]any, 0, len(rows)*5)
		for _, row := range rows {
			sqlStr += "(?, ?, ?, ?, ?),"
			values = append(values, row.ID, row.DagID, row.Operator, row.SourceID, row.HasDatasource)
		}
		sqlStr = strings.TrimSuffix(sqlStr, ",")
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGSTEPINDEX_TABLENAME), attribute.String(trace.DB_SQL, sqlStr))
		return db.Exec(sqlStr, values...).Error
	}

	// insertTriggerRows := func(rows []*DagTriggerConfigIndex) error {
	// 	if len(rows) == 0 {
	// 		return nil
	// 	}
	// 	sqlStr := fmt.Sprintf("INSERT INTO %s (f_id, f_dag_id, f_operator, f_source_id) VALUES ", DAGTRIGGERINDEX_TABLENAME)
	// 	values := make([]any, 0, len(rows)*4)
	// 	for _, row := range rows {
	// 		sqlStr += "(?, ?, ?, ?),"
	// 		values = append(values, row.ID, row.DagID, row.Operator, row.SourceID)
	// 	}
	// 	sqlStr = strings.TrimSuffix(sqlStr, ",")
	// 	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGTRIGGERINDEX_TABLENAME), attribute.String(trace.DB_SQL, sqlStr))
	// 	return d.db.Exec(sqlStr, values...).Error
	// }

	insertAccessorRows := func(rows []*DagAccessorModel) error {
		if len(rows) == 0 {
			return nil
		}
		sqlStr := fmt.Sprintf("INSERT INTO %s (f_id, f_dag_id, f_accessor_id) VALUES ", DAGACCESSORINDEX_TABLENAME)
		values := make([]any, 0, len(rows)*3)
		for _, row := range rows {
			sqlStr += "(?, ?, ?),"
			values = append(values, row.ID, row.DagID, row.AccessorID)
		}
		sqlStr = strings.TrimSuffix(sqlStr, ",")
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGACCESSORINDEX_TABLENAME), attribute.String(trace.DB_SQL, sqlStr))
		return db.Exec(sqlStr, values...).Error
	}

	if err = insertStepRows(stepRows); err != nil {
		return err
	}
	// if err = insertTriggerRows(triggerRows); err != nil {
	// 	return err
	// }
	if err = insertAccessorRows(accessorRows); err != nil {
		return err
	}

	return nil
}

func (d *dag) CreateDagVersion(ctx context.Context, dagVersion *entity.DagVersion) (string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	defer func() { trace.TelemetrySpanEnd(span, err) }()

	sqlStr := `INSERT INTO t_flow_dag_version (
		f_id, f_created_at, f_updated_at, f_dag_id,
		f_user_id, f_version, f_version_id, f_change_log, f_config, f_sort_time)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	t := ToDagVersionModel(dagVersion)
	msgStr, _ := jsoniter.MarshalToString(t)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVERSIONS_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))
	err = db.Exec(sqlStr,
		t.ID,
		t.CreatedAt,
		t.UpdatedAt,
		t.DagID,
		t.UserID,
		t.Version,
		t.VersionID,
		t.ChangeLog,
		t.Config,
		t.SortTime,
	).Error
	if err != nil {
		return "", err
	}

	return dagVersion.ID, nil
}

func (d *dag) UpdateDag(ctx context.Context, dagEntity *entity.Dag) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(store *dag, dagEntity *entity.Dag) error {
		// 准备 SQL 语句，使用参数化查询防止 SQL 注入
		sql := `UPDATE t_flow_dag SET
			f_created_at = ?, f_updated_at = ?, f_user_id = ?, f_name = ?, f_desc = ?,
			f_trigger = ?, f_cron = ?, f_vars = ?, f_status = ?, f_tasks = ?, f_steps = ?,
			f_description = ?, f_shortcuts = ?, f_accessors = ?, f_type = ?, f_policy_type = ?,
			f_appinfo = ?, f_priority = ?, f_removed = ?, f_emails = ?, f_template = ?, f_published = ?,
			f_trigger_config = ?, f_sub_ids = ?, f_exec_mode = ?, f_category = ?, f_outputs = ?,
			f_instructions = ?, f_operator_id = ?, f_inc_values = ?, f_version = ?, f_version_id = ?,
			f_modify_by = ?, f_is_debug = ?, f_debug_id = ?, f_biz_domain_id = ?
			WHERE f_id = ?`

		// 执行 SQL 语句
		t := ToDagModel(dagEntity, true)
		msgStr, _ := jsoniter.MarshalToString(t)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

		err = dagEntity.CheckRootNode(dagEntity.Tasks)
		if err != nil {
			return err
		}

		err = store.db.Exec(sql,
			t.CreatedAt,
			t.UpdatedAt,
			t.UserID,
			t.Name,
			t.Desc,
			t.Trigger,
			t.Cron,
			t.Vars,
			t.Status,
			t.Tasks,
			t.Steps,
			t.Description,
			t.Shortcuts,
			t.Accessors,
			t.Type,
			t.PolicyType,
			t.AppInfo,
			t.Priority,
			t.Removed,
			t.Emails,
			t.Template,
			t.Published,
			t.TriggerConfig,
			t.SubIDs,
			t.ExecMode,
			t.Category,
			t.OutPuts,
			t.Instructions,
			t.OperatorID,
			t.IncValues,
			t.Version,
			t.VersionID,
			t.ModifyBy,
			t.IsDebug,
			t.DeBugID,
			t.BizDomainID,
			t.ID,
		).Error
		if err != nil {
			return err
		}

		err = store.CreateDagVars(newCtx, BuildDagVars(dagEntity))
		if err != nil {
			return err
		}

		err = store.refreshDagIndexes(newCtx, dagEntity)
		if err != nil {
			return err
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(_ context.Context, txStore mod.Store) error {
			return fn(txStore.(*dag), dagEntity)
		})
	} else {
		err = fn(d, dagEntity)
	}

	return err
}

func (d *dag) GetDag(ctx context.Context, dagId string) (*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	return d.GetDagByFields(newCtx, map[string]interface{}{"f_id": dagId})
}

func (d *dag) GetDagByFields(ctx context.Context, params map[string]interface{}) (*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if _, ok := params["f_removed"]; !ok {
		params["f_removed"] = map[string]interface{}{"$lt": 1}
	}

	result, err := NewConverter(DAG_TABLENAME, WithAutoConvert(true)).Convert(params)
	if err != nil {
		return nil, err
	}

	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, result.SQL), attribute.String(trace.DB_QUERY, query))

	dag := &DagModel{}
	err = db.Raw(result.SQL, result.Params...).Scan(dag).Error
	if err != nil {
		return nil, err
	}

	if dag.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	dest := &entity.Dag{}
	err = ToEntity(dag, dest)
	return dest, err
}

func (d *dag) GetDagWithOptionalVersion(ctx context.Context, dagID, versionID string) (*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	var sql string
	if versionID != "" {
		var config entity.Config
		sql = `SELECT f_config FROM t_flow_dag_version WHERE f_dag_id = ? AND f_version_id = ?`
		err = db.Raw(sql, dagID, versionID).Scan(&config).Error
		if err != nil {
			return nil, err
		}

		return config.ParseToDag()
	} else {
		return d.GetDag(newCtx, dagID)
	}
}

func (d *dag) DeleteDag(ctx context.Context, id ...string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	msgStr, _ := jsoniter.MarshalToString(id)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(store *dag, ids ...string) error {
		if len(ids) == 0 {
			return nil
		}

		sqlStr := `DELETE FROM t_flow_dag WHERE f_id IN ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, msgStr))

		err = store.db.Exec(sqlStr, ids).Error
		if err != nil {
			return err
		}

		sqlStr = `DELETE FROM t_flow_dag_var WHERE f_dag_id IN ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, msgStr))
		err = store.db.Exec(sqlStr, ids).Error
		if err != nil {
			return err
		}

		sqlStr = `DELETE FROM t_flow_dag_version WHERE f_dag_id IN ?`
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVERSIONS_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_QUERY, msgStr))
		err = store.db.Exec(sqlStr, ids).Error
		if err != nil {
			return err
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(_ context.Context, txStore mod.Store) error {
			return fn(txStore.(*dag), id...)
		})
	} else {
		err = fn(d, id...)
	}

	return err
}

func (d *dag) ListDagInstance(ctx context.Context, input *mod.ListDagInstanceInput) ([]*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	var conds []string
	var args []interface{}

	if len(input.DagIDs) > 0 {
		dagIDs := make([]interface{}, len(input.DagIDs))
		for i, v := range input.DagIDs {
			id, _ := strconv.ParseUint(v, 10, 64)
			dagIDs[i] = id
		}
		conds = append(conds, "f_dag_id IN ?")
		args = append(args, dagIDs)
	}

	if len(input.Status) > 0 {
		conds = append(conds, "f_status IN ?")
		args = append(args, input.Status)
	}

	if len(input.UserIDs) > 0 {
		conds = append(conds, "f_user_id IN ?")
		args = append(args, input.UserIDs)
	}

	if len(input.Priority) > 0 {
		conds = append(conds, "f_priority IN ?")
		args = append(args, input.Priority)
	}

	if input.Worker != "" {
		conds = append(conds, "f_worker = ?")
		args = append(args, input.Worker)
	}
	if input.HasCmd {
		conds = append(conds, "f_has_cmd = ?")
		args = append(args, true)
	}
	if input.UpdatedEnd > 0 {
		conds = append(conds, "f_updated_at <= ?")
		args = append(args, input.UpdatedEnd)
	}
	if input.ExcludeModeVM {
		conds = append(conds, "f_mode < 1")
	}
	if input.TimeRange != nil {
		col := camelToFSnake(input.TimeRange.Field)
		conds = append(conds, fmt.Sprintf("%s >= ? AND %s <= ?", col, col))
		args = append(args, input.TimeRange.Begin, input.TimeRange.End)
	}
	if input.MatchQuery != nil {
		switch input.MatchQuery.Field {
		case "vars.batch_run_id.value":
			conds = append(conds, "f_batch_run_id = ?")
			args = append(args, input.MatchQuery.Value)
		case "keywords":
			if like, ok := buildKeywordLike(input.MatchQuery.Value); ok {
				conds = append(conds, "EXISTS (SELECT 1 FROM t_flow_dag_instance_keyword dik WHERE dik.f_dag_ins_id = di.f_id AND dik.f_keyword LIKE ?)")
				args = append(args, like)
			}
		}
	}

	// 构建 SQL
	var sql string
	if len(input.SelectField) > 0 {
		selectFileds := []string{}
		for _, v := range input.SelectField {
			selectFileds = append(selectFileds, camelToFSnake(v))
		}
		sql = fmt.Sprintf("SELECT %s FROM t_flow_dag_instance di", strings.Join(selectFileds, ", "))
	} else {
		sql = "SELECT * FROM t_flow_dag_instance di"
	}

	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}

	if input.SortBy != "" {
		sortBy := camelToFSnake(input.SortBy)
		dir := utils.IfNot(input.Order == 1, "ASC", "DESC")
		sql += fmt.Sprintf(" ORDER BY %s %s", sortBy, dir)
	}

	if input.Limit > 0 {
		sql += " LIMIT ? OFFSET ?"
		args = append(args, input.Limit, input.Limit*input.Offset)
	} else if input.Limit < 0 {
		sql += " LIMIT ?"
		args = append(args, -1)
	}

	query, _ := jsoniter.MarshalToString(args)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, query))

	dagInstances := make([]*DagInstanceModel, 0)
	err = db.Raw(sql, args...).Scan(&dagInstances).Error
	if err != nil {
		return nil, err
	}

	dest := make([]*entity.DagInstance, 0)
	for _, dag := range dagInstances {
		dagIns := &entity.DagInstance{}
		err = ToEntity(dag, dagIns)
		if err != nil {
			return nil, err
		}
		dest = append(dest, dagIns)
	}

	return dest, nil

}

func (d *dag) CreateDagIns(ctx context.Context, dagIns *entity.DagInstance) (string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(txCtx context.Context, store *dag) error {
		sql := `INSERT INTO t_flow_dag_instance (
		f_id, f_created_at, f_updated_at, f_dag_id, f_trigger, f_worker, f_source,
		f_vars, f_keywords, f_event_persistence, f_event_oss_path, f_share_data, f_share_data_ext,
		f_status, f_reason, f_cmd, f_has_cmd, f_batch_run_id, f_user_id, f_ended_at, f_dag_type, f_policy_type, f_appinfo,
		f_priority, f_mode, f_dump, f_dump_ext, f_success_callback, f_error_callback, f_call_chain,
		f_resume_data, f_resume_status, f_version, f_version_id, f_biz_domain_id)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		// 执行 SQL 语句
		t := ToDagInstanceModel(dagIns, false)
		msgStr, _ := jsoniter.MarshalToString(t)
		trace.SetAttributes(txCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

		if err = store.db.Exec(sql,
			t.ID,
			t.CreatedAt,
			t.UpdatedAt,
			t.DagID,
			t.Trigger,
			t.Worker,
			t.Source,
			t.Vars,
			t.Keywords,
			t.EventPersistence,
			t.EventOssPath,
			t.ShareData,
			t.ShareDataExt,
			t.Status,
			t.Reason,
			t.Cmd,
			t.HasCmd,
			t.BatchRunID,
			t.UserID,
			t.EndedAt,
			t.DagType,
			t.PolicyType,
			t.AppInfo,
			t.Priority,
			t.Mode,
			t.Dump,
			t.DumpExt,
			t.SuccessCallback,
			t.ErrorCallback,
			t.CallChain,
			t.ResumeData,
			t.ResumeStatus,
			t.Version,
			t.VersionID,
			t.BizDomainID,
		).Error; err != nil {
			return err
		}

		return store.insertDagInstanceKeywords(txCtx, t.ID, dagIns.Keywords)
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(txCtx context.Context, txStore mod.Store) error {
			return fn(txCtx, txStore.(*dag))
		})
	} else {
		err = fn(newCtx, d)
	}

	return dagIns.ID, err
}

func (d *dag) GetHistoryDagByVersionID(ctx context.Context, dagID, versionID string) (*entity.DagVersion, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `SELECT
		f_id, f_created_at, f_updated_at, f_dag_id, f_user_id, f_version,
		f_version_id, f_change_log, f_config, f_sort_time FROM t_flow_dag_version
		WHERE f_dag_id = ? AND f_version_id = ?`
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVERSIONS_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, dagID), attribute.String(trace.DB_QUERY, versionID))

	dagVersion := &DagVersionModel{}
	err = db.Raw(sql, dagID, versionID).Scan(dagVersion).Error
	if err != nil {
		return nil, err
	}

	dest := &entity.DagVersion{}
	err = ToEntity(dagVersion, dest)
	if err != nil {
		return nil, err
	}

	return dest, nil
}

// BatchCreatOutBoxMessage
func (d *dag) BatchCreatOutBoxMessage(ctx context.Context, outBox []*entity.OutBox) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if len(outBox) == 0 {
		return nil
	}

	sqlStr := `INSERT INTO t_flow_outbox (f_id, f_topic, f_msg, f_created_at, f_updated_at) VALUES `

	values := make([]any, 0, len(outBox)*5)
	for _, data := range outBox {
		sqlStr += "(?, ?, ?, ?, ?),"
		values = append(values, data.ID, data.Topic, data.Msg, data.CreatedAt, data.UpdatedAt)
	}

	sqlStr = sqlStr[:len(sqlStr)-1]

	msgStr, _ := jsoniter.MarshalToString(values)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))

	err = db.Exec(sqlStr, values...).Error

	return err
}

// BatchCreateDag
func (d *dag) BatchCreateDag(ctx context.Context, dags []*entity.Dag) ([]*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(dags) == 0 {
		return dags, nil
	}

	fn := func(ctx context.Context, d mod.Store) error {
		for _, dag := range dags {
			dag.Initial()
			// check task's connection
			_, err = mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
			if err != nil {
				return err
			}

			_, err = d.CreateDag(ctx, dag)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(_ context.Context, txStore mod.Store) error {
			return fn(newCtx, txStore)
		})
	} else {
		err = fn(newCtx, d)
	}

	if err != nil {
		return nil, err
	}

	return dags, nil
}

// BatchCreateDagIns
func (d *dag) BatchCreateDagIns(ctx context.Context, dagIns []*entity.DagInstance) ([]*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(dagIns) == 0 {
		return dagIns, nil
	}

	fn := func(txCtx context.Context, store *dag) error {
		batchSize := 1000
		for i := 0; i < len(dagIns); i += batchSize {
			end := i + batchSize
			if end > len(dagIns) {
				end = len(dagIns)
			}
			batch := dagIns[i:end]

			sqlStr := `INSERT INTO t_flow_dag_instance (
		f_id, f_created_at, f_updated_at, f_dag_id, f_trigger, f_worker, f_source,
		f_vars, f_keywords, f_event_persistence, f_event_oss_path, f_share_data, f_share_data_ext,
		f_status, f_reason, f_cmd, f_has_cmd, f_batch_run_id, f_user_id, f_ended_at, f_dag_type, f_policy_type, f_appinfo,
		f_priority, f_mode, f_dump, f_dump_ext, f_success_callback, f_error_callback, f_call_chain,
		f_resume_data, f_resume_status, f_version, f_version_id, f_biz_domain_id)
		VALUES `

			values := make([]any, 0, len(batch)*35)
			keywordRows := make([]DagInstanceKeywordModel, 0)
			for _, data := range batch {
				sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
				t := ToDagInstanceModel(data, false)
				values = append(values,
					t.ID,
					t.CreatedAt,
					t.UpdatedAt,
					t.DagID,
					t.Trigger,
					t.Worker,
					t.Source,
					t.Vars,
					t.Keywords,
					t.EventPersistence,
					t.EventOssPath,
					t.ShareData,
					t.ShareDataExt,
					t.Status,
					t.Reason,
					t.Cmd,
					t.HasCmd,
					t.BatchRunID,
					t.UserID,
					t.EndedAt,
					t.DagType,
					t.PolicyType,
					t.AppInfo,
					t.Priority,
					t.Mode,
					t.Dump,
					t.DumpExt,
					t.SuccessCallback,
					t.ErrorCallback,
					t.CallChain,
					t.ResumeData,
					t.ResumeStatus,
					t.Version,
					t.VersionID,
					t.BizDomainID,
				)
				keywordRows = append(keywordRows, buildDagInstanceKeywordRows(t.ID, data.Keywords)...)
			}

			sqlStr = strings.TrimSuffix(sqlStr, ",")

			msgStr, _ := jsoniter.MarshalToString(values)
			trace.SetAttributes(txCtx, attribute.String(trace.TABLE_NAME, DAGVAR_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))

			if err = store.db.Exec(sqlStr, values...).Error; err != nil {
				return err
			}

			if err = store.insertDagInstanceKeywordsBatch(txCtx, keywordRows); err != nil {
				return err
			}
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(txCtx context.Context, txStore mod.Store) error {
			return fn(txCtx, txStore.(*dag))
		})
	} else {
		err = fn(newCtx, d)
	}
	if err != nil {
		return nil, err
	}

	return dagIns, nil
}

// BatchCreateTaskIns
func (d *dag) BatchCreateTaskIns(ctx context.Context, taskIns []*entity.TaskInstance) ([]*entity.TaskInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(taskIns) == 0 {
		return taskIns, nil
	}

	fn := func(txCtx context.Context, store *dag) error {
		batchSize := 1000
		for i := 0; i < len(taskIns); i += batchSize {
			end := i + batchSize
			if end > len(taskIns) {
				end = len(taskIns)
			}
			batch := taskIns[i:end]

			sqlStr := `INSERT INTO t_flow_task_instance (
			f_id, f_created_at, f_updated_at, f_expired_at, f_task_id, f_dag_ins_id, f_name, f_depend_on,
			f_action_name, f_timeout_secs, f_params, f_traces, f_status, f_reason, f_pre_checks,
			f_results, f_steps, f_last_modified_at, f_rendered_params, f_hash, f_settings, f_metadata
		) VALUES `

			values := make([]any, 0, len(batch)*21)
			for _, data := range batch {
				t := ToTaskInstanceModel(data, false)
				sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
				values = append(values,
					t.ID,
					t.CreatedAt,
					t.UpdatedAt,
					t.UpdatedAt+int64(t.TimeoutSecs),
					t.TaskID,
					t.DagInsID,
					t.Name,
					t.DependOn,
					t.ActionName,
					t.TimeoutSecs,
					t.Params,
					t.Traces,
					t.Status,
					t.Reason,
					t.PreChecks,
					t.Results,
					t.Steps,
					t.LastModifiedAt,
					t.RenderedParams,
					t.Hash,
					t.Settings,
					t.MetaData,
				)
			}
			sqlStr = strings.TrimSuffix(sqlStr, ",")

			msgStr, _ := jsoniter.MarshalToString(values)
			trace.SetAttributes(txCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))

			if err = store.db.Exec(sqlStr, values...).Error; err != nil {
				return err
			}
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(txCtx context.Context, txStore mod.Store) error {
			return fn(txCtx, txStore.(*dag))
		})
	} else {
		err = fn(newCtx, d)
	}
	if err != nil {
		return nil, err
	}

	return taskIns, nil
}

// BatchDeleteDagIns
func (d *dag) BatchDeleteDagIns(ctx context.Context, ids []string) error {
	return d.delete(ctx, ids, DAGINSTANCE_TABLENAME)
}

func (d *dag) delete(ctx context.Context, ids []string, tableName string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `DELETE FROM ` + tableName + ` WHERE f_id IN ?`
	msgStr, _ := jsoniter.MarshalToString(ids)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr))

	err = db.Exec(sql, ids).Error

	return err
}

// BatchDeleteDagWithTransaction
func (d *dag) BatchDeleteDagWithTransaction(ctx context.Context, ids []string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(store *dag) error {
		sql := `UPDATE t_flow_dag SET f_removed = 1 WHERE f_id IN (?) AND f_type NOT IN (?)`
		msgStr, _ := jsoniter.MarshalToString(ids)
		trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr), attribute.String(trace.DB_QUERY, common.DagTypeSecurityPolicy))

		err = store.db.Exec(sql, ids, common.DagTypeSecurityPolicy).Error
		if err != nil {
			return err
		}

		batchSize := 1000
		lastID := uint64(0) // 用于游标的最后一个ID

		for {
			// 获取要删除的 DAG 实例 ID（使用游标分页）
			var dagInsIDs []uint64
			var query string
			var queryArgs = []interface{}{ids}
			var cnt int64

			if lastID == 0 {
				// 第一次查询
				query = `SELECT f_id, f_dag_type
				FROM t_flow_dag_instance
				WHERE f_dag_id IN ?
				ORDER BY f_id ASC
				LIMIT ?`
				queryArgs = append(queryArgs, batchSize)
			} else {
				// 后续查询，使用游标
				query = `SELECT f_id, f_dag_type
				FROM t_flow_dag_instance
				WHERE f_dag_id IN ? AND f_id > ?
				ORDER BY f_id ASC
				LIMIT ?`
				queryArgs = append(queryArgs, lastID, batchSize)
			}

			rows, err := store.db.Raw(query, queryArgs...).Rows()
			if err != nil {
				return fmt.Errorf("failed to query dag instances: %w", err)
			}

			for rows.Next() {
				var id uint64
				var dagType string
				if err := rows.Scan(&id, &dagType); err != nil {
					rows.Close()
					return err
				}
				cnt++
				if dagType == common.DagTypeSecurityPolicy {
					continue
				}
				dagInsIDs = append(dagInsIDs, id)
				lastID = id // 更新游标位置
			}
			rows.Close()

			if cnt == 0 {
				break // 没有更多记录可删除
			}

			// 删除 DAG 实例
			deleteDagInsQuery := `DELETE FROM t_flow_dag_instance WHERE f_id IN ?`
			err = store.db.Exec(deleteDagInsQuery, dagInsIDs).Error
			if err != nil {
				return err
			}

			// 删除相关的任务实例
			deleteTaskInsQuery := `DELETE FROM t_flow_task_instance WHERE f_dag_ins_id IN ?`
			err = store.db.Exec(deleteTaskInsQuery, dagInsIDs).Error
			if err != nil {
				return err
			}
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(_ context.Context, txStore mod.Store) error {
			return fn(txStore.(*dag))
		})
	} else {
		err = fn(d)
	}

	return err
}

// BatchDeleteTaskIns
func (d *dag) BatchDeleteTaskIns(ctx context.Context, ids []string) error {
	return d.delete(ctx, ids, TASKINSTANCE_TABLENAME)
}

// BatchUpdateDagIns
func (d *dag) BatchUpdateDagIns(ctx context.Context, dagIns []*entity.DagInstance) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(txCtx context.Context, store *dag) error {
		batchSize := 1000

		for i := 0; i < len(dagIns); i += batchSize {
			end := i + batchSize
			if end > len(dagIns) {
				end = len(dagIns)
			}

			batch := dagIns[i:end]

			for _, dagIns := range batch {
				dagIns.Update()

				// 方法1：使用UPDATE完全替换（假设文档必须存在）
				sql := `UPDATE t_flow_dag_instance SET
				f_updated_at = ?, f_trigger = ?, f_worker = ?, f_source = ?,
				f_vars = ?, f_keywords = ?, f_event_persistence = ?, f_event_oss_path = ?, f_share_data = ?, f_share_data_ext = ?,
				f_status = ?, f_reason = ?, f_cmd = ?, f_has_cmd = ?, f_batch_run_id = ?, f_user_id = ?, f_ended_at = ?, f_dag_type = ?, f_policy_type = ?, f_appinfo = ?,
				f_priority = ?, f_mode = ?, f_dump = ?, f_dump_ext = ?, f_success_callback = ?, f_error_callback = ?, f_call_chain = ?,
				f_resume_data = ?, f_resume_status = ?, f_version = ?, f_version_id = ?, f_biz_domain_id = ?
				WHERE f_id = ?`

				// 执行 SQL 语句
				t := ToDagInstanceModel(dagIns, false)

				err = store.db.Exec(sql,
					t.UpdatedAt,
					t.Trigger,
					t.Worker,
					t.Source,
					t.Vars,
					t.Keywords,
					t.EventPersistence,
					t.EventOssPath,
					t.ShareData,
					t.ShareDataExt,
					t.Status,
					t.Reason,
					t.Cmd,
					t.HasCmd,
					t.BatchRunID,
					t.UserID,
					t.EndedAt,
					t.DagType,
					t.PolicyType,
					t.AppInfo,
					t.Priority,
					t.Mode,
					t.Dump,
					t.DumpExt,
					t.SuccessCallback,
					t.ErrorCallback,
					t.CallChain,
					t.ResumeData,
					t.ResumeStatus,
					t.Version,
					t.VersionID,
					t.BizDomainID,
					t.ID,
				).Error
				if err != nil {
					return err
				}

				if err = store.replaceDagInstanceKeywords(txCtx, t.ID, dagIns.Keywords); err != nil {
					return err
				}
			}
		}

		return nil
	}

	if !d.isTX {
		return d.WithTransaction(newCtx, func(txCtx context.Context, txStore mod.Store) error {
			return fn(txCtx, txStore.(*dag))
		})
	}

	return fn(newCtx, d)
}

// BatchUpdateTaskIns
func (d *dag) BatchUpdateTaskIns(taskIns []*entity.TaskInstance) error {
	if len(taskIns) == 0 {
		return nil
	}

	fn := func(store *dag) error {
		for i := range taskIns {
			taskIns[i].Update()
			t := ToTaskInstanceModel(taskIns[i], true)

			sql := `UPDATE t_flow_task_instance SET
			f_updated_at = ?, f_expired_at, f_task_id = ?, f_dag_ins_id = ?, f_name = ?, f_depend_on = ?,
			f_action_name = ?, f_timeout_secs = ?, f_params = ?, f_traces = ?, f_status = ?,
			f_reason = ?, f_pre_checks = ?, f_results = ?, f_steps = ?, f_last_modified_at = ?,
			f_rendered_params = ?, f_hash = ?, f_settings = ?, f_metadata = ?
			WHERE f_id = ?`

			if err := store.db.Exec(sql,
				t.UpdatedAt,
				t.UpdatedAt+int64(t.TimeoutSecs),
				t.TaskID,
				t.DagInsID,
				t.Name,
				t.DependOn,
				t.ActionName,
				t.TimeoutSecs,
				t.Params,
				t.Traces,
				t.Status,
				t.Reason,
				t.PreChecks,
				t.Results,
				t.Steps,
				t.LastModifiedAt,
				t.RenderedParams,
				t.Hash,
				t.Settings,
				t.MetaData,
				t.ID,
			).Error; err != nil {
				return err
			}
		}

		return nil
	}

	if !d.isTX {
		return d.WithTransaction(context.Background(), func(_ context.Context, txStore mod.Store) error {
			return fn(txStore.(*dag))
		})
	}

	return fn(d)
}

// Close
func (d *dag) Close() {
	if d == nil || d.db == nil {
		return
	}
	sqlDB, err := d.db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

// CreatOutBoxMessage
func (d *dag) CreatOutBoxMessage(ctx context.Context, outBox *entity.OutBox) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `INSERT INTO t_flow_outbox (f_id, f_topic, f_msg, f_created_at, f_updated_at) VALUES (?, ?, ?, ?, ?)`

	msgStr, _ := jsoniter.MarshalToString(outBox)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, OUTBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr), attribute.String(trace.DB_QUERY, msgStr))

	err = db.Exec(sql,
		outBox.ID,
		outBox.Topic,
		outBox.Msg,
		outBox.CreatedAt,
		outBox.UpdatedAt,
	).Error

	return err
}

// CreateClient
func (d *dag) CreateClient(clientName string, clientID string, clientSecret string) error {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	id, err := utils.GetUniqueID()
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	sql := `INSERT INTO t_flow_client (f_id, f_created_at, f_updated_at, f_client_name, f_client_id, f_client_secret) VALUES (?, ?, ?, ?, ?, ?)`
	return db.Exec(sql, id, now, now, clientName, clientID, clientSecret).Error
}

// CreateInbox
func (d *dag) CreateInbox(ctx context.Context, msg *entity.InBox) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `INSERT INTO t_flow_inbox (f_id, f_msg, f_topic, f_docid, f_dag, f_created_at, f_updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`

	msgStr, _ := jsoniter.MarshalToString(msg)

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, INBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	t := ToInboxModel(msg)
	err = db.Exec(sql,
		t.ID,
		t.Msg,
		t.Topic,
		t.DocID,
		t.Dags,
		t.CreatedAt,
		t.UpdatedAt,
	).Error

	return err
}

// CreateLogs
func (d *dag) CreateLogs(ctx context.Context, ossLogs []*entity.Log) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if len(ossLogs) == 0 {
		return nil
	}

	fn := func(txCtx context.Context, store *dag) error {
		batchSize := 1000
		for i := 0; i < len(ossLogs); i += batchSize {
			end := min(i+batchSize, len(ossLogs))
			batch := ossLogs[i:end]

			sqlStr := `INSERT INTO t_flow_log (f_id, f_created_at, f_updated_at, f_ossid, f_key, f_filename) VALUES `
			values := make([]any, 0, len(batch)*6)
			for _, logItem := range batch {
				logItem.Initial()
				id, _ := strconv.ParseUint(logItem.ID, 10, 64)
				sqlStr += "(?, ?, ?, ?, ?, ?),"
				values = append(values, id, logItem.CreatedAt, logItem.UpdatedAt, logItem.OssID, logItem.Key, logItem.FileName)
			}
			sqlStr = strings.TrimSuffix(sqlStr, ",")

			msgStr, _ := jsoniter.MarshalToString(values)
			trace.SetAttributes(txCtx, attribute.String(trace.TABLE_NAME, LOG_TABLENAME), attribute.String(trace.DB_SQL, sqlStr), attribute.String(trace.DB_Values, msgStr))

			if err = store.db.Exec(sqlStr, values...).Error; err != nil {
				return err
			}
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(txCtx context.Context, txStore mod.Store) error {
			return fn(txCtx, txStore.(*dag))
		})
	} else {
		err = fn(newCtx, d)
	}

	return err
}

// CreateTaskIns
func (d *dag) CreateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `INSERT INTO t_flow_task_instance (
		f_id, f_created_at, f_updated_at, f_expired_at, f_task_id, f_dag_ins_id, f_name, f_depend_on,
		f_action_name, f_timeout_secs, f_params, f_traces, f_status, f_reason, f_pre_checks,
		f_results, f_steps, f_last_modified_at, f_rendered_params, f_hash, f_settings, f_metadata
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	t := ToTaskInstanceModel(taskIns, false)
	msgStr, _ := jsoniter.MarshalToString(t)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	return db.Exec(sql,
		t.ID,
		t.CreatedAt,
		t.UpdatedAt,
		t.UpdatedAt+int64(t.TimeoutSecs),
		t.TaskID,
		t.DagInsID,
		t.Name,
		t.DependOn,
		t.ActionName,
		t.TimeoutSecs,
		t.Params,
		t.Traces,
		t.Status,
		t.Reason,
		t.PreChecks,
		t.Results,
		t.Steps,
		t.LastModifiedAt,
		t.RenderedParams,
		t.Hash,
		t.Settings,
		t.MetaData,
	).Error
}

// CreateToken
func (d *dag) CreateToken(token *entity.Token) error {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	baseInfo := token.GetBaseInfo()
	baseInfo.Initial()
	id, _ := strconv.ParseUint(baseInfo.ID, 10, 64)

	sql := `INSERT INTO t_flow_token (
		f_id, f_created_at, f_updated_at, f_user_id, f_user_name, f_refresh_token,
		f_token, f_expires_in, f_login_ip, f_is_app
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if err := db.Exec(sql,
		id,
		baseInfo.CreatedAt,
		baseInfo.UpdatedAt,
		token.UserID,
		token.UserName,
		token.RefreshToken,
		token.Token,
		token.ExpiresIn,
		token.LoginIP,
		token.IsApp,
	).Error; err != nil {
		return err
	}
	return nil
}

// DeleteDagInsByID
func (d *dag) DeleteDagInsByID(ctx context.Context, params map[string]interface{}) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	delete(params, "dagInsIDs")

	var id interface{}
	if i, ok := params["_id"]; ok {
		if idStr, ok := i.(string); ok {
			id, err = strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				return err
			}
		} else {
			id = i
		}
	}

	var status []string
	if s, ok := params["status"]; ok {
		if statusSlice, ok := s.([]string); ok {
			status = statusSlice
		}
	}

	// 安全地获取 updatedAt
	var updatedAt interface{}
	if ua, ok := params["updatedAt"]; ok {
		updatedAt = ua
	}

	sql := "DELETE FROM t_flow_dag_instance WHERE f_id <= ? AND f_status IN ? AND f_updated_at <= ?"
	msyBytes, _ := jsoniter.MarshalToString(params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msyBytes))

	err = db.Exec(sql, id, status, updatedAt).Error

	return err
}

// DeleteInbox
func (d *dag) DeleteInbox(ctx context.Context, ids []string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `DELETE FROM t_flow_inbox WHERE f_id IN ?`

	msgStr, _ := jsoniter.MarshalToString(ids)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, INBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr))

	err = db.Exec(sql, ids).Error

	return err
}

// DeleteOutBoxMessage
func (d *dag) DeleteOutBoxMessage(ctx context.Context, ids []string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `DELETE FROM t_flow_outbox WHERE f_id IN ?`

	msgStr, _ := jsoniter.MarshalToString(ids)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, OUTBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr))

	err = db.Exec(sql, ids).Error

	return err
}

// DeleteTaskInsByDagInsID
func (d *dag) DeleteTaskInsByDagInsID(ctx context.Context, dagInsID string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	id, _ := strconv.ParseUint(dagInsID, 10, 64)
	sql := `DELETE FROM t_flow_task_instance WHERE f_dag_ins_id = ?`
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, dagInsID))

	return db.Exec(sql, id).Error
}

// DeleteTaskInsByID
func (d *dag) DeleteTaskInsByID(ctx context.Context, params map[string]interface{}) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	var status []string
	if s, ok := params["status"]; ok {
		if statusSlice, ok := s.([]string); ok {
			status = statusSlice
		}
	}

	var dagInsIDs []string
	if ids, ok := params["dagInsIDs"]; ok {
		if idSlice, ok := ids.([]string); ok {
			dagInsIDs = idSlice
		}
	}

	var dagInsU64 []uint64
	for _, id := range dagInsIDs {
		v, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			continue
		}
		dagInsU64 = append(dagInsU64, v)
	}

	// 安全地获取 updatedAt
	var updatedAt interface{}
	if ua, ok := params["updatedAt"]; ok {
		updatedAt = ua
	}

	// 安全地获取 _id 参数
	var maxID uint64
	if maxIDStr, ok := params["_id"].(string); ok && maxIDStr != "" {
		var err error
		maxID, err = strconv.ParseUint(maxIDStr, 10, 64)
		if err != nil {
			return err
		}
	}

	sql := `DELETE FROM t_flow_task_instance WHERE f_id <= ? AND f_dag_ins_id IN ? AND f_status IN ? AND f_updated_at <= ?`
	msgStr, _ := jsoniter.MarshalToString(params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, msgStr))

	err = db.Exec(sql, maxID, dagInsU64, status, updatedAt).Error

	return err
}

// DeleteToken
func (d *dag) DeleteToken(id string) error {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	sql := `DELETE FROM t_flow_token WHERE f_id = ?`
	uid, _ := strconv.ParseUint(id, 10, 64)
	return db.Exec(sql, uid).Error
}

// DisdinctDagInstance
func (d *dag) DisdinctDagInstance(input *mod.ListDagInstanceInput) ([]interface{}, error) {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	var conds []string
	var args []interface{}

	if len(input.DagIDs) > 0 {
		var ids []uint64
		for _, v := range input.DagIDs {
			id, _ := strconv.ParseUint(v, 10, 64)
			ids = append(ids, id)
		}
		conds = append(conds, "f_dag_id IN ?")
		args = append(args, ids)
	}
	if len(input.Status) > 0 {
		var status []string
		for _, v := range input.Status {
			status = append(status, string(v))
		}
		conds = append(conds, "f_status IN ?")
		args = append(args, status)
	}
	if input.Worker != "" {
		conds = append(conds, "f_worker = ?")
		args = append(args, input.Worker)
	}
	if input.UpdatedEnd > 0 {
		conds = append(conds, "f_updated_at <= ?")
		args = append(args, input.UpdatedEnd)
	}
	if input.ExcludeModeVM {
		conds = append(conds, "f_mode <> ?")
		args = append(args, entity.DagInstanceModeVM)
	}

	field := camelToFSnake(input.DistinctField)
	sql := fmt.Sprintf("SELECT DISTINCT %s FROM t_flow_dag_instance", field)
	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}
	if input.SortBy != "" {
		dir := utils.IfNot(input.Order == 1, "ASC", "DESC")
		sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(input.SortBy), dir)
	}
	if input.Limit > 0 {
		sql += " LIMIT ? OFFSET ?"
		args = append(args, input.Limit, input.Limit*input.Offset)
	}

	var res []interface{}
	var dagIns []*DagInstanceModel

	if err := db.Raw(sql, args...).Scan(&dagIns).Error; err != nil {
		return nil, err
	}

	for _, v := range dagIns {
		res = append(res, v.UserID)
	}
	return res, nil
}

// GetClient
func (d *dag) GetClient(clientName string) (client *entity.Client, err error) {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	sql := `SELECT f_id, f_created_at, f_updated_at, f_client_name, f_client_id, f_client_secret FROM t_flow_client WHERE f_client_name = ?`
	model := &ClientModel{}
	if err = db.Raw(sql, clientName).Scan(model).Error; err != nil {
		return nil, err
	}
	if model.ID == 0 {
		return &entity.Client{}, nil
	}
	dest := &entity.Client{}
	if err = ToEntity(model, dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// GetDagCount
func (d *dag) GetDagCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if params == nil {
		params = map[string]interface{}{}
	}
	params["type"] = bson.M{"$ne": common.DagTypeSecurityPolicy}

	result, err := NewConverter(DAG_TABLENAME, WithAutoConvert(true)).ConvertConds(params)
	if err != nil {
		return 0, err
	}

	conds := result.Conds
	if conds == "" {
		conds = "1=1"
	}
	conds += " AND f_removed < 1 AND f_is_debug < 1"

	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", DAG_TABLENAME, conds)
	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, query))

	var count int64
	err = db.Raw(sql, result.Params...).Scan(&count).Error
	return count, err
}

// GetDagInstance
func (d *dag) GetDagInstance(ctx context.Context, dagInsId string) (*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	return d.GetDagInstanceByFields(newCtx, map[string]interface{}{"f_id": dagInsId})
}

// GetDagInstanceByFields
func (d *dag) GetDagInstanceByFields(ctx context.Context, params map[string]interface{}) (*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	result, err := NewConverter(DAGINSTANCE_TABLENAME, WithAutoConvert(true)).Convert(params)
	if err != nil {
		return nil, err
	}

	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, result.SQL), attribute.String(trace.DB_QUERY, query))

	dagIns := &DagInstanceModel{}
	err = db.Raw(result.SQL, result.Params...).Scan(dagIns).Error
	if err != nil {
		return nil, err
	}
	if dagIns.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	dest := &entity.DagInstance{}
	err = ToEntity(dagIns, dest)
	return dest, err
}

// GetDagInstanceCount
func (d *dag) GetDagInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	logCleanTTL, _ := params["log_clean_ttl"].(time.Duration)
	logCleanTTL = utils.IfNot(logCleanTTL == 0, defaultCtxTimeout, logCleanTTL)
	db, _, cancel := d.dbWithContextWithTimeout(newCtx, logCleanTTL)
	defer cancel()

	delete(params, "log_clean_ttl")
	baseParams := make(map[string]interface{}, len(params))
	for k, v := range params {
		baseParams[k] = v
	}

	var extraConds []string
	var extraArgs []interface{}
	if kw, ok := baseParams["keywords"]; ok {
		delete(baseParams, "keywords")
		if like, ok := buildKeywordLike(kw); ok {
			extraConds = append(extraConds, "EXISTS (SELECT 1 FROM t_flow_dag_instance_keyword dik WHERE dik.f_dag_ins_id = di.f_id AND dik.f_keyword LIKE ?)")
			extraArgs = append(extraArgs, like)
		}
	}

	conv := NewConverter(DAGINSTANCE_TABLENAME, WithAutoConvert(true), WithFieldMap(map[string]string{
		"vars.batch_run_id.value": "f_batch_run_id",
	}))
	result, err := conv.ConvertConds(baseParams)
	if err != nil {
		return 0, err
	}

	var conds []string
	var args []interface{}
	if result.Conds != "" {
		conds = append(conds, result.Conds)
		args = append(args, result.Params...)
	}
	if len(extraConds) > 0 {
		conds = append(conds, extraConds...)
		args = append(args, extraArgs...)
	}
	if len(conds) == 0 {
		conds = append(conds, "1=1")
	}

	sql := fmt.Sprintf("SELECT COUNT(*) FROM t_flow_dag_instance di WHERE %s", strings.Join(conds, " AND "))

	query, _ := jsoniter.MarshalToString(args)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, query))

	var count int64
	err = db.Raw(sql, args...).Scan(&count).Error
	return count, err
}

// GetInbox
func (d *dag) GetInbox(ctx context.Context, id string) (*entity.InBox, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `SELECT f_id, f_msg, f_topic, f_docid, f_dag, f_created_at, f_updated_at FROM t_flow_inbox WHERE f_id = ?`

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, INBOXMESSAGE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, id))

	inbox := &InBoxModel{}
	err = db.Raw(sql, id).Scan(inbox).Error
	if err != nil {
		return nil, err
	}

	var docMsg common.DocMsg
	_ = jsoniter.UnmarshalFromString(inbox.Msg, &docMsg)
	var dags []string
	_ = jsoniter.UnmarshalFromString(inbox.Dags, &dags)

	return &entity.InBox{
		BaseInfo: entity.BaseInfo{
			ID:        fmt.Sprintf("%v", inbox.ID),
			CreatedAt: inbox.CreatedAt,
			UpdatedAt: inbox.UpdatedAt,
		},
		Msg:   docMsg,
		Topic: inbox.Topic,
		DocID: inbox.DocID,
		Dags:  dags,
	}, nil
}

// GetSwitchStatus
func (d *dag) GetSwitchStatus() (bool, error) {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	sql := `SELECT f_id, f_created_at, f_updated_at, f_name, f_status FROM t_flow_switch WHERE f_name = ?`
	sw := &SwitchModel{}
	if err := db.Raw(sql, entity.SwitchName).Scan(sw).Error; err != nil {
		return false, fmt.Errorf("get switch status failed: %w", err)
	}
	if sw.ID == 0 {
		return true, nil
	}
	return sw.Status, nil
}

// GetTaskIns
func (d *dag) GetTaskIns(ctx context.Context, taskIns string) (*entity.TaskInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `SELECT * FROM t_flow_task_instance WHERE f_id = ?`
	id, _ := strconv.ParseUint(taskIns, 10, 64)
	model := &TaskInstanceModel{}
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, taskIns))

	if err = db.Raw(sql, id).Scan(model).Error; err != nil {
		return nil, err
	}
	if model.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	dest := &entity.TaskInstance{}
	if err = ToEntity(model, dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// GetTaskInstanceCount
func (d *dag) GetTaskInstanceCount(ctx context.Context, params map[string]interface{}) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	logCleanTTL, _ := params["log_clean_ttl"].(time.Duration)
	logCleanTTL = utils.IfNot(logCleanTTL == 0, defaultCtxTimeout, logCleanTTL)
	db, _, cancel := d.dbWithContextWithTimeout(newCtx, logCleanTTL)
	defer cancel()

	delete(params, "log_clean_ttl")
	result, err := NewConverter(TASKINSTANCE_TABLENAME, WithAutoConvert(true)).ConvertConds(params)
	if err != nil {
		return 0, err
	}

	conds := result.Conds
	if conds == "" {
		conds = "1=1"
	}
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", TASKINSTANCE_TABLENAME, conds)
	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, query))

	var count int64
	err = db.Raw(sql, result.Params...).Scan(&count).Error
	return count, err
}

// GetTokenByUserID
func (d *dag) GetTokenByUserID(userID string) (*entity.Token, error) {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	sql := `SELECT f_id, f_created_at, f_updated_at, f_user_id, f_user_name, f_refresh_token, f_token, f_expires_in, f_login_ip, f_is_app FROM t_flow_token WHERE f_user_id = ?`
	model := &TokenModel{}
	if err := db.Raw(sql, userID).Scan(model).Error; err != nil {
		return &entity.Token{}, fmt.Errorf("get token failed: %w", err)
	}
	if model.ID == 0 {
		return &entity.Token{}, nil
	}
	dest := &entity.Token{}
	if err := ToEntity(model, dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// GroupDagInstance
func (d *dag) GroupDagInstance(ctx context.Context, input *mod.GroupInput) ([]*entity.DagInstanceGroup, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql, args, qerr := BuildGroupDagInstanceQuery(input)
	if qerr != nil {
		err = qerr
		return nil, qerr
	}
	if sql == "" {
		return nil, fmt.Errorf("sql is empty")
	}

	if input != nil && !input.IsFirst {
		type totalRow struct {
			Total int64 `gorm:"column:total"`
		}
		rows := make([]totalRow, 0)
		if err = db.Raw(sql, args...).Scan(&rows).Error; err != nil {
			return nil, err
		}
		result := make([]*entity.DagInstanceGroup, 0, len(rows))
		for _, row := range rows {
			result = append(result, &entity.DagInstanceGroup{Total: row.Total})
		}
		return result, nil
	}

	type groupRow struct {
		Total int64 `gorm:"column:total"`
		DagInstanceModel
	}
	rows := make([]groupRow, 0)
	if err = db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]*entity.DagInstanceGroup, 0, len(rows))
	for _, row := range rows {
		dagIns := &entity.DagInstance{}
		if err = ToEntity(&row.DagInstanceModel, dagIns); err != nil {
			return nil, err
		}
		result = append(result, &entity.DagInstanceGroup{
			Total:  row.Total,
			DagIns: dagIns,
		})
	}

	return result, nil
}

// ListDag
func (d *dag) ListDag(ctx context.Context, input *mod.ListDagInput) ([]*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if input == nil {
		return nil, nil
	}

	var conds []string
	var args []interface{}

	// default type exclude
	if input.Type != "" {
		if input.Type != "all" {
			conds = append(conds, "f_type = ?")
			args = append(args, input.Type)
		}
	} else {
		conds = append(conds, "f_type NOT IN ?")
		args = append(args, []string{
			common.DagTypeSecurityPolicy,
			common.DagTypeDataFlow,
			common.DagTypeDataFlowForBot,
			common.DagTypeComboOperator,
		})
	}

	if input.TriggerType != "" {
		conds = append(conds, "f_trigger = ?")
		args = append(args, input.TriggerType)
	}
	if len(input.TriggerTypes) > 0 {
		conds = append(conds, "f_trigger IN ?")
		args = append(args, input.TriggerTypes)
	}
	if input.UserID != "" && input.Scope != "all" {
		conds = append(conds, "f_user_id = ?")
		args = append(args, input.UserID)
	}
	if input.KeyWord != "" {
		conds = append(conds, "f_name LIKE ?")
		args = append(args, "%"+input.KeyWord+"%")
	}
	if len(input.DagIDs) > 0 {
		conds = append(conds, "f_id IN ?")
		args = append(args, parseUint64Slice(input.DagIDs))
	}
	if len(input.Status) > 0 {
		var status []string
		for _, s := range input.Status {
			status = append(status, string(s))
		}
		conds = append(conds, "f_status IN ?")
		args = append(args, status)
	}

	conds = append(conds, "f_removed < 1", "f_is_debug < 1")

	indexCond, indexArgs := BuildDagIndexSubquery(input)
	if indexCond != "" {
		conds = append(conds, indexCond)
		args = append(args, indexArgs...)
	}

	sql := "SELECT * FROM t_flow_dag"
	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}
	if input.SortBy != "" {
		sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(input.SortBy), utils.IfNot(input.Order == 1, "ASC", "DESC"))
	}
	if input.Limit > 0 {
		sql += " LIMIT ? OFFSET ?"
		args = append(args, input.Limit, input.Limit*input.Offset)
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql))

	models := make([]*DagModel, 0)
	if err = db.Raw(sql, args...).Scan(&models).Error; err != nil {
		return nil, err
	}

	var res []*entity.Dag
	for _, model := range models {
		dag := &entity.Dag{}
		if err = ToEntity(model, dag); err != nil {
			return nil, err
		}
		res = append(res, dag)
	}
	return res, nil
}

// ListDagByFields
func (d *dag) ListDagByFields(ctx context.Context, filter bson.M, opt options.FindOptions) ([]*entity.Dag, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if filter == nil {
		filter = bson.M{}
	}
	if _, ok := filter["removed"]; !ok {
		filter["removed"] = bson.M{"$ne": true}
	}
	if _, ok := filter["is_debug"]; !ok {
		filter["is_debug"] = bson.M{"$ne": true}
	}

	result, err := NewConverter(DAG_TABLENAME, WithAutoConvert(true)).Convert(filter)
	if err != nil {
		return nil, err
	}

	sql := result.SQL
	args := result.Params

	if opt.Sort != nil {
		switch v := opt.Sort.(type) {
		case map[string]interface{}:
			for k, order := range v {
				dir := "ASC"
				if ord, ok := order.(int); ok && ord <= 0 {
					dir = "DESC"
				}
				sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(k), dir)
				break
			}
		case bson.D:
			if len(v) > 0 {
				dir := "ASC"
				if v[0].Value.(int32) <= 0 {
					dir = "DESC"
				}
				sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(v[0].Key), dir)
			}
		}
	}
	if opt.Limit != nil && *opt.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, *opt.Limit)
		if opt.Skip != nil && *opt.Skip > 0 {
			sql += " OFFSET ?"
			args = append(args, *opt.Skip)
		}
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql))

	models := make([]*DagModel, 0)
	if err = db.Raw(sql, args...).Scan(&models).Error; err != nil {
		return nil, err
	}
	var res []*entity.Dag
	for _, model := range models {
		dag := &entity.Dag{}
		if err = ToEntity(model, dag); err != nil {
			return nil, err
		}
		res = append(res, dag)
	}
	return res, nil
}

// ListDagCount
func (d *dag) ListDagCount(ctx context.Context, input *mod.ListDagInput) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if input == nil {
		return 0, nil
	}

	var conds []string
	var args []interface{}

	conds = append(conds, "f_type NOT IN ?")
	args = append(args, []string{
		common.DagTypeSecurityPolicy,
		common.DagTypeDataFlow,
		common.DagTypeDataFlowForBot,
		common.DagTypeComboOperator,
	})

	if input.Type != "" {
		if input.Type == "all" {
			conds = conds[:0]
			args = args[:0]
		} else {
			conds = []string{"f_type = ?"}
			args = []interface{}{input.Type}
		}
	}

	if input.TriggerType != "" {
		conds = append(conds, "f_trigger = ?")
		args = append(args, input.TriggerType)
	}
	if input.UserID != "" {
		conds = append(conds, "f_user_id = ?")
		args = append(args, input.UserID)
	}
	if input.KeyWord != "" {
		conds = append(conds, "f_name LIKE ?")
		args = append(args, "%"+input.KeyWord+"%")
	}
	if len(input.DagIDs) > 0 {
		conds = append(conds, "f_id IN ?")
		args = append(args, parseUint64Slice(input.DagIDs))
	}
	if len(input.Status) > 0 {
		var status []string
		for _, s := range input.Status {
			status = append(status, string(s))
		}
		conds = append(conds, "f_status IN ?")
		args = append(args, status)
	}
	if input.BizDomainID != "" {
		if input.BizDomainID == common.BizDomainDefaultID {
			conds = append(conds, "(f_biz_domain_id = '' OR f_biz_domain_id = ?)")
			args = append(args, common.BizDomainDefaultID)
		} else {
			conds = append(conds, "f_biz_domain_id = ?")
			args = append(args, input.BizDomainID)
		}
	}

	conds = append(conds, "f_removed < 1", "f_is_debug < 1")

	if len(input.Sources) != 0 && len(input.Trigger) > 0 {
		conds = append(conds, "f_id IN (SELECT f_dag_id FROM t_flow_dag_step WHERE f_operator IN ? AND f_source_id IN ?)")
		args = append(args, input.Trigger, input.Sources)
	}

	if input.Accessors != nil && input.UserID == "" {
		conds = append(conds, "f_id IN (SELECT f_dag_id FROM t_flow_dag_accessor WHERE f_accessor_id IN ?)")
		args = append(args, input.Accessors)
	}

	sql := "SELECT COUNT(*) FROM t_flow_dag"
	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql))

	var count int64
	err = db.Raw(sql, args...).Scan(&count).Error
	return count, err
}

// ListDagCountByFields
func (d *dag) ListDagCountByFields(ctx context.Context, filter bson.M) (int64, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if filter == nil {
		filter = bson.M{}
	}
	if _, ok := filter["removed"]; !ok {
		filter["removed"] = bson.M{"$ne": true}
	}
	if _, ok := filter["is_debug"]; !ok {
		filter["is_debug"] = bson.M{"$ne": true}
	}

	result, err := NewConverter(DAG_TABLENAME, WithAutoConvert(true)).ConvertConds(filter)
	if err != nil {
		return 0, err
	}
	conds := result.Conds
	if conds == "" {
		conds = "1=1"
	}
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", DAG_TABLENAME, conds)
	query, _ := jsoniter.MarshalToString(result.Params)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_QUERY, query))

	var count int64
	err = db.Raw(sql, result.Params...).Scan(&count).Error
	return count, err
}

// ListDagInstanceInRangeTime
func (d *dag) ListDagInstanceInRangeTime(ctx context.Context, status string, begin int64, end int64) ([]*entity.DagInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := `SELECT * FROM t_flow_dag_instance WHERE f_status = ? AND f_updated_at >= ? AND f_updated_at <= ?`
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql))

	models := make([]*DagInstanceModel, 0)
	if err = db.Raw(sql, status, begin, end).Scan(&models).Error; err != nil {
		return nil, err
	}
	var res []*entity.DagInstance
	for _, model := range models {
		dagIns := &entity.DagInstance{}
		if err = ToEntity(model, dagIns); err != nil {
			return nil, err
		}
		res = append(res, dagIns)
	}
	return res, nil
}

// ListDagVersions
func (d *dag) ListDagVersions(ctx context.Context, input *mod.ListDagVersionInput) ([]entity.DagVersion, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	sql := "SELECT * FROM t_flow_dag_version WHERE 1=1"
	args := make([]interface{}, 0)
	if input.DagID != "" {
		sql += " AND f_dag_id = ?"
		args = append(args, input.DagID)
	}
	if input.SortBy != "" {
		sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(input.SortBy), utils.IfNot(input.Order == 1, "ASC", "DESC"))
	}
	if input.Limit > 0 {
		sql += " LIMIT ? OFFSET ?"
		args = append(args, input.Limit, input.Limit*input.Offset)
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGVERSIONS_TABLENAME), attribute.String(trace.DB_SQL, sql))

	var models []DagVersionModel
	if err = db.Raw(sql, args...).Scan(&models).Error; err != nil {
		return nil, err
	}
	var res []entity.DagVersion
	for _, m := range models {
		item := entity.DagVersion{}
		if err = ToEntity(&m, &item); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

// ListExistDagID
func (d *dag) ListExistDagID(ctx context.Context, dagIDs []string) ([]string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if len(dagIDs) == 0 {
		return nil, nil
	}

	sql := `SELECT f_id FROM t_flow_dag WHERE f_id IN ? AND f_removed < 1 AND f_is_debug < 1`
	ids := parseUint64Slice(dagIDs)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAG_TABLENAME), attribute.String(trace.DB_SQL, sql))

	var res []uint64
	if err = db.Raw(sql, ids).Scan(&res).Error; err != nil {
		return nil, err
	}
	var out []string
	for _, id := range res {
		out = append(out, strconv.FormatUint(id, 10))
	}
	return out, nil
}

// ListExistDagInsID
func (d *dag) ListExistDagInsID(ctx context.Context, dagInsIDs []string) ([]string, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if len(dagInsIDs) == 0 {
		return nil, nil
	}

	sql := `SELECT f_id FROM t_flow_dag_instance WHERE f_id IN ?`
	ids := parseUint64Slice(dagInsIDs)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql))

	var res []uint64
	if err = db.Raw(sql, ids).Scan(&res).Error; err != nil {
		return nil, err
	}
	var out []string
	for _, id := range res {
		out = append(out, strconv.FormatUint(id, 10))
	}
	return out, nil
}

// ListHistoryDagIns
func (d *dag) ListHistoryDagIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	status, _ := params["status"]
	updatedAt, _ := params["updatedAt"]

	lastID := uint64(0)
	batchSize := common.DefaultQuerySize

	for {
		select {
		case <-newCtx.Done():
			close(dataChannel)
			return nil
		default:
		}

		sql := `SELECT * FROM t_flow_dag_instance WHERE f_status IN ? AND f_updated_at <= ? AND f_id > ? ORDER BY f_id ASC LIMIT ?`
		var models []DagInstanceModel
		if err = db.Raw(sql, status, updatedAt, lastID, batchSize).Scan(&models).Error; err != nil {
			return err
		}
		if len(models) == 0 {
			close(dataChannel)
			return nil
		}

		var batch []bson.M
		for _, m := range models {
			lastID = m.ID
			e := &entity.DagInstance{}
			if err = ToEntity(&m, e); err != nil {
				return err
			}
			b, _ := bson.Marshal(e)
			var doc bson.M
			_ = bson.Unmarshal(b, &doc)
			batch = append(batch, doc)
		}
		if len(batch) > 0 {
			dataChannel <- batch
		}
		if len(models) < batchSize {
			close(dataChannel)
			return nil
		}
	}
}

// ListHistoryTaskIns
func (d *dag) ListHistoryTaskIns(ctx context.Context, params map[string]interface{}, dataChannel chan []bson.M) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	status, _ := params["status"]
	updatedAt, _ := params["updatedAt"]

	lastID := uint64(0)
	batchSize := common.DefaultQuerySize

	for {
		select {
		case <-newCtx.Done():
			close(dataChannel)
			return nil
		default:
		}

		sql := `SELECT * FROM t_flow_task_instance WHERE f_status IN ? AND f_updated_at <= ? AND f_id > ? ORDER BY f_id ASC LIMIT ?`
		var models []TaskInstanceModel
		if err = db.Raw(sql, status, updatedAt, lastID, batchSize).Scan(&models).Error; err != nil {
			return err
		}
		if len(models) == 0 {
			close(dataChannel)
			return nil
		}

		var batch []bson.M
		for _, m := range models {
			lastID = m.ID
			e := &entity.TaskInstance{}
			if err = ToEntity(&m, e); err != nil {
				return err
			}
			b, _ := bson.Marshal(e)
			var doc bson.M
			_ = bson.Unmarshal(b, &doc)
			batch = append(batch, doc)
		}
		if len(batch) > 0 {
			dataChannel <- batch
		}
		if len(models) < batchSize {
			close(dataChannel)
			return nil
		}
	}
}

// ListInbox
func (d *dag) ListInbox(ctx context.Context, input *mod.ListInboxInput) ([]*entity.InBox, error) {
	db, _, cancel := d.dbWithContext(ctx)
	defer cancel()

	// 构建 SQL 查询
	sqlQuery := "SELECT * FROM t_flow_inbox WHERE 1=1"
	var args []interface{}

	// 应用筛选条件
	if input != nil {
		// 根据用户ID筛选
		if input.DocID != "" {
			sqlQuery += " AND f_docid = ?"
			args = append(args, input.DocID)
		}

		// 根据消息类型筛选
		if len(input.Topics) > 0 {
			sqlQuery += " AND f_topic IN ?"
			args = append(args, input.Topics)
		}

		// 根据状态筛选
		if input.Now > 0 {
			sqlQuery += " AND f_created_at <= ?"
			// input.Now - 2*60
			args = append(args, input.Now-2*60)
		}

		// 排序
		orderBy := "f_created_at"
		order := "DESC"
		if input.SortBy != "" {
			orderBy = input.SortBy
		}
		if input.Order > 0 {
			order = "ASC"
		}
		sqlQuery += fmt.Sprintf(" ORDER BY %s %s", orderBy, order)

		// 分页处理（放在最后）
		if input.Limit >= 0 && input.Offset >= 0 {
			offset := input.Offset * input.Limit
			sqlQuery += " LIMIT ? OFFSET ?"
			args = append(args, input.Limit, offset)
		}
	}

	// 执行原生 SQL 查询
	var inboxes []*InBoxModel
	if err := db.Raw(sqlQuery, args...).Scan(&inboxes).Error; err != nil {
		return nil, err
	}

	var res []*entity.InBox
	for _, inbox := range inboxes {
		var docMsg common.DocMsg
		_ = jsoniter.UnmarshalFromString(inbox.Msg, &docMsg)
		var dags []string
		_ = jsoniter.UnmarshalFromString(inbox.Dags, &dags)

		res = append(res, &entity.InBox{
			BaseInfo: entity.BaseInfo{
				ID:        fmt.Sprintf("%v", inbox.ID),
				CreatedAt: inbox.CreatedAt,
				UpdatedAt: inbox.UpdatedAt,
			},
			Msg:   docMsg,
			Topic: inbox.Topic,
			DocID: inbox.DocID,
			Dags:  dags,
		})
	}

	return res, nil
}

// ListOutBoxMessage
func (d *dag) ListOutBoxMessage(ctx context.Context, input *entity.OutBoxInput) ([]*entity.OutBox, error) {
	var err error
	var msgs []*OutBoxModel

	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx = newCtx
	db, _, cancel := d.dbWithContext(ctx)
	defer cancel()

	sqlStr := `SELECT f_id, f_topic, f_msg, f_created_at, f_updated_at FROM t_flow_outbox WHERE 1 = 1 `

	values := make([]interface{}, 0)
	if input.CreateTime > 0 {
		sqlStr += " AND f_created_at <= ?"
		values = append(values, input.CreateTime)
	}

	if input.Limit > 0 {
		sqlStr += " LIMIT ?"
		values = append(values, input.Limit)
	}

	err = db.Raw(sqlStr, values...).Scan(&msgs).Error
	if err != nil {
		return nil, err
	}

	var res []*entity.OutBox
	for _, msg := range msgs {
		t := &entity.OutBox{}
		err = ToEntity(&msg, t)
		if err != nil {
			return nil, err
		}

		res = append(res, t)
	}

	return res, nil
}

// ListTaskInstance
func (d *dag) ListTaskInstance(ctx context.Context, input *mod.ListTaskInstanceInput) ([]*entity.TaskInstance, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	conds := make([]string, 0)
	args := make([]interface{}, 0)

	if len(input.IDs) > 0 {
		var ids []uint64
		for _, id := range input.IDs {
			v, _ := strconv.ParseUint(id, 10, 64)
			ids = append(ids, v)
		}
		conds = append(conds, "f_id IN ?")
		args = append(args, ids)
	}

	var actionConds []string
	var actionArgs []interface{}
	if len(input.ActionName) > 0 {
		actionConds = append(actionConds, "f_action_name IN ?")
		actionArgs = append(actionArgs, input.ActionName)
	}
	if input.ActionNameRegex != "" {
		actionConds = append(actionConds, "f_action_name LIKE ?")
		actionArgs = append(actionArgs, regexToLike(input.ActionNameRegex))
	}
	if len(actionConds) > 0 {
		conds = append(conds, "("+strings.Join(actionConds, " OR ")+")")
		args = append(args, actionArgs...)
	}

	if len(input.Status) > 0 {
		var status []string
		for _, s := range input.Status {
			status = append(status, string(s))
		}
		conds = append(conds, "f_status IN ?")
		args = append(args, status)
	}

	if input.Expired {
		conds = append(conds, "f_expired_at <= ?")
		args = append(args, time.Now().Unix()-5)
	}

	if input.DagInsID != "" {
		id, _ := strconv.ParseUint(input.DagInsID, 10, 64)
		conds = append(conds, "f_dag_ins_id = ?")
		args = append(args, id)
	} else if len(input.DagInsIDs) > 0 {
		var ids []uint64
		for _, id := range input.DagInsIDs {
			v, _ := strconv.ParseUint(id, 10, 64)
			ids = append(ids, v)
		}
		conds = append(conds, "f_dag_ins_id IN ?")
		args = append(args, ids)
	}

	if input.Hash != "" {
		conds = append(conds, "f_hash = ?")
		args = append(args, input.Hash)
	}

	sql := "SELECT * FROM t_flow_task_instance"
	if len(conds) > 0 {
		sql += " WHERE " + strings.Join(conds, " AND ")
	}
	if input.SortBy != "" {
		sql += fmt.Sprintf(" ORDER BY %s %s", camelToFSnake(input.SortBy), utils.IfNot(input.Order == 1, "ASC", "DESC"))
	}
	if input.Limit > 0 {
		sql += " LIMIT ? OFFSET ?"
		args = append(args, input.Limit, input.Limit*input.Offset)
	}

	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql))

	models := make([]*TaskInstanceModel, 0)
	if err = db.Raw(sql, args...).Scan(&models).Error; err != nil {
		return nil, err
	}

	var res []*entity.TaskInstance
	for _, model := range models {
		task := &entity.TaskInstance{}
		if err = ToEntity(model, task); err != nil {
			return nil, err
		}
		res = append(res, task)
	}

	return res, nil
}

// Marshal
func (d *dag) Marshal(obj interface{}) ([]byte, error) {
	return jsoniter.Marshal(obj)
}

// PatchDagIns
func (d *dag) PatchDagIns(ctx context.Context, dagIns *entity.DagInstance, mustsPatchFields ...string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if dagIns.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}

	fn := func(txCtx context.Context, store *dag) error {
		setClauses := []string{"f_updated_at = ?"}
		values := []any{time.Now().Unix()}
		updateKeywords := false

		if dagIns.EndedAt != 0 {
			setClauses = append(setClauses, "f_ended_at = ?")
			values = append(values, dagIns.EndedAt)
		}

		if dagIns.EventPersistence == 0 {
			if dagIns.ShareData != nil {
				if dagIns.ShareDataExt != nil {
					setClauses = append(setClauses, "f_share_data_ext = ?", "f_share_data = ?")
					values = append(values, marshalToString(dagIns.ShareDataExt), "")
				} else {
					setClauses = append(setClauses, "f_share_data_ext = ?", "f_share_data = ?")
					values = append(values, "", marshalToString(dagIns.ShareData))
				}
			}

			if dagIns.Dump != "" {
				if dagIns.DumpExt != nil {
					setClauses = append(setClauses, "f_dump_ext = ?", "f_dump = ?")
					values = append(values, marshalToString(dagIns.DumpExt), "")
				} else {
					setClauses = append(setClauses, "f_dump_ext = ?", "f_dump = ?")
					values = append(values, "", dagIns.Dump)
				}
			}
		} else {
			setClauses = append(setClauses, "f_event_persistence = ?")
			values = append(values, int(dagIns.EventPersistence))
		}

		if dagIns.EventOssPath != "" {
			setClauses = append(setClauses, "f_event_oss_path = ?")
			values = append(values, dagIns.EventOssPath)
		}
		if dagIns.Status != "" {
			setClauses = append(setClauses, "f_status = ?")
			values = append(values, string(dagIns.Status))
		}
		if utils.IsContain("Cmd", mustsPatchFields) || dagIns.Cmd != nil {
			setClauses = append(setClauses, "f_cmd = ?", "f_has_cmd = ?")
			values = append(values, marshalToString(dagIns.Cmd), dagIns.Cmd != nil)
		}
		if dagIns.Worker != "" {
			setClauses = append(setClauses, "f_worker = ?")
			values = append(values, dagIns.Worker)
		}
		if utils.IsContain("Reason", mustsPatchFields) || dagIns.Reason != "" {
			setClauses = append(setClauses, "f_reason = ?")
			values = append(values, dagIns.Reason)
		}
		if dagIns.ResumeData != "" {
			setClauses = append(setClauses, "f_resume_data = ?")
			values = append(values, dagIns.ResumeData)
		}
		if dagIns.ResumeStatus != "" {
			setClauses = append(setClauses, "f_resume_status = ?")
			values = append(values, string(dagIns.ResumeStatus))
		}
		if dagIns.Source != "" {
			setClauses = append(setClauses, "f_source = ?")
			values = append(values, dagIns.Source)
		}
		if len(dagIns.Keywords) > 0 {
			setClauses = append(setClauses, "f_keywords = ?")
			values = append(values, marshalToString(dagIns.Keywords))
			updateKeywords = true
		}

		sql := fmt.Sprintf("UPDATE t_flow_dag_instance SET %s WHERE f_id = ?", strings.Join(setClauses, ", "))
		values = append(values, dagIns.ID)

		msgStr, _ := jsoniter.MarshalToString(values)
		trace.SetAttributes(txCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

		if err = store.db.Exec(sql, values...).Error; err != nil {
			return err
		}

		if updateKeywords {
			dagInsID, _ := strconv.ParseUint(dagIns.ID, 10, 64)
			if err = store.replaceDagInstanceKeywords(txCtx, dagInsID, dagIns.Keywords); err != nil {
				return err
			}
		}

		return nil
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(txCtx context.Context, txStore mod.Store) error {
			return fn(txCtx, txStore.(*dag))
		})
	} else {
		err = fn(newCtx, d)
	}
	if err != nil {
		return err
	}

	goevent.Publish(&event.DagInstancePatched{
		Payload:         dagIns,
		MustPatchFields: mustsPatchFields,
	})

	return nil
}

// PatchTaskIns
func (d *dag) PatchTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	if taskIns.ID == "" {
		return fmt.Errorf("id cannot be empty")
	}

	setClauses := []string{"f_updated_at = ?"}
	values := []any{time.Now().Unix()}

	if taskIns.Status != "" {
		setClauses = append(setClauses, "f_status = ?")
		values = append(values, string(taskIns.Status))
	}
	if taskIns.Reason != "" {
		setClauses = append(setClauses, "f_reason = ?")
		values = append(values, marshalToString(taskIns.Reason))
	}
	if len(taskIns.Traces) > 0 {
		setClauses = append(setClauses, "f_traces = ?")
		values = append(values, marshalToString(taskIns.Traces))
	}
	if taskIns.Results != nil {
		setClauses = append(setClauses, "f_results = ?")
		values = append(values, marshalToString(taskIns.Results))
	}
	if taskIns.LastModifiedAt != 0 {
		setClauses = append(setClauses, "f_last_modified_at = ?")
		values = append(values, taskIns.LastModifiedAt)
	}
	if taskIns.RenderedParams != nil {
		setClauses = append(setClauses, "f_rendered_params = ?")
		values = append(values, marshalToString(taskIns.RenderedParams))
	}
	if taskIns.DependOn != nil {
		setClauses = append(setClauses, "f_depend_on = ?")
		values = append(values, marshalToString(taskIns.DependOn))
	}
	if taskIns.Hash != "" {
		setClauses = append(setClauses, "f_hash = ?")
		values = append(values, taskIns.Hash)
	}
	if taskIns.MetaData != nil {
		setClauses = append(setClauses, "f_metadata = ?")
		values = append(values, marshalToString(taskIns.MetaData))
	}

	sql := fmt.Sprintf("UPDATE t_flow_task_instance SET %s WHERE f_id = ?", strings.Join(setClauses, ", "))
	values = append(values, taskIns.ID)

	msgStr, _ := jsoniter.MarshalToString(values)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	return db.Exec(sql, values...).Error
}

// RemoveClient
func (d *dag) RemoveClient(clientName string) (err error) {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	sql := `DELETE FROM t_flow_client WHERE f_client_name = ?`
	return db.Exec(sql, clientName).Error
}

// RetryDagIns
func (d *dag) RetryDagIns(ctx context.Context, dagInsID string, taskInsIDs []string) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(store *dag) error {
		now := time.Now().Unix()
		if len(taskInsIDs) > 0 {
			var ids []uint64
			for _, id := range taskInsIDs {
				v, _ := strconv.ParseUint(id, 10, 64)
				ids = append(ids, v)
			}
			sqlTask := `UPDATE t_flow_task_instance SET f_updated_at = ?, f_status = ? WHERE f_id IN ?`
			if err = store.db.Exec(sqlTask, now, string(entity.TaskInstanceStatusInit), ids).Error; err != nil {
				return err
			}
		}

		sqlDag := `UPDATE t_flow_dag_instance SET f_updated_at = ?, f_status = ?, f_ended_at = ? WHERE f_id = ?`
		return store.db.Exec(sqlDag, now, string(entity.DagInstanceStatusInit), now, dagInsID).Error
	}

	if !d.isTX {
		return d.WithTransaction(newCtx, func(_ context.Context, txStore mod.Store) error {
			return fn(txStore.(*dag))
		})
	}
	return fn(d)
}

// SetSwitchStatus
func (d *dag) SetSwitchStatus(status bool) error {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	now := time.Now().Unix()
	sql := `INSERT INTO t_flow_switch (f_id, f_created_at, f_updated_at, f_name, f_status)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE f_status = VALUES(f_status), f_updated_at = VALUES(f_updated_at)`
	id, _ := utils.GetUniqueID()
	return db.Exec(sql, id, now, now, entity.SwitchName, status).Error
}

// Unmarshal
func (d *dag) Unmarshal(bytes []byte, ptr interface{}) error {
	return jsoniter.Unmarshal(bytes, ptr)
}

// UpdateDagIncValue
func (d *dag) UpdateDagIncValue(ctx context.Context, dagId string, incKey string, incValue any) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	var raw string
	if err = db.Raw(`SELECT f_inc_values FROM t_flow_dag WHERE f_id = ?`, dagId).Scan(&raw).Error; err != nil {
		return err
	}
	updated, err := updateJSONMapString(raw, incKey, incValue)
	if err != nil {
		return err
	}
	err = db.Exec(`UPDATE t_flow_dag SET f_inc_values = ? WHERE f_id = ?`, updated, dagId).Error

	return err
}

// UpdateDagIns
func (d *dag) UpdateDagIns(ctx context.Context, dagIns *entity.DagInstance) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	fn := func(txCtx context.Context, store *dag) error {
		sql := `UPDATE t_flow_dag_instance SET
		f_updated_at = ?, f_dag_id = ?, f_trigger = ?, f_worker = ?, f_source = ?,
		f_vars = ?, f_keywords = ?, f_event_persistence = ?, f_event_oss_path = ?, f_share_data = ?, f_share_data_ext = ?,
		f_status = ?, f_reason = ?, f_cmd = ?, f_has_cmd = ?, f_batch_run_id = ?, f_user_id = ?, f_ended_at = ?, f_dag_type = ?, f_policy_type = ?, f_appinfo = ?,
		f_priority = ?, f_mode = ?, f_dump = ?, f_dump_ext = ?, f_success_callback = ?, f_error_callback = ?, f_call_chain = ?,
		f_resume_data = ?, f_resume_status = ?, f_version = ?, f_version_id = ?, f_biz_domain_id = ?
		WHERE f_id = ?`

		t := ToDagInstanceModel(dagIns, true)
		msgStr, _ := jsoniter.MarshalToString(t)
		trace.SetAttributes(txCtx, attribute.String(trace.TABLE_NAME, DAGINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

		if err = store.db.Exec(sql,
			t.UpdatedAt,
			t.DagID,
			t.Trigger,
			t.Worker,
			t.Source,
			t.Vars,
			t.Keywords,
			t.EventPersistence,
			t.EventOssPath,
			t.ShareData,
			t.ShareDataExt,
			t.Status,
			t.Reason,
			t.Cmd,
			t.HasCmd,
			t.BatchRunID,
			t.UserID,
			t.EndedAt,
			t.DagType,
			t.PolicyType,
			t.AppInfo,
			t.Priority,
			t.Mode,
			t.Dump,
			t.DumpExt,
			t.SuccessCallback,
			t.ErrorCallback,
			t.CallChain,
			t.ResumeData,
			t.ResumeStatus,
			t.Version,
			t.VersionID,
			t.BizDomainID,
			t.ID,
		).Error; err != nil {
			return err
		}

		return store.replaceDagInstanceKeywords(txCtx, t.ID, dagIns.Keywords)
	}

	if !d.isTX {
		err = d.WithTransaction(newCtx, func(txCtx context.Context, txStore mod.Store) error {
			return fn(txCtx, txStore.(*dag))
		})
	} else {
		err = fn(newCtx, d)
	}
	if err != nil {
		return err
	}

	goevent.Publish(&event.DagInstanceUpdated{Payload: dagIns})
	return nil
}

// UpdateTaskIns
func (d *dag) UpdateTaskIns(ctx context.Context, taskIns *entity.TaskInstance) error {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	db, _, cancel := d.dbWithContext(newCtx)
	defer cancel()

	t := ToTaskInstanceModel(taskIns, true)

	sql := `UPDATE t_flow_task_instance SET
		f_updated_at = ?, f_task_id = ?, f_dag_ins_id = ?, f_name = ?, f_depend_on = ?,
		f_action_name = ?, f_timeout_secs = ?, f_params = ?, f_traces = ?, f_status = ?,
		f_reason = ?, f_pre_checks = ?, f_results = ?, f_steps = ?, f_last_modified_at = ?,
		f_rendered_params = ?, f_hash = ?, f_settings = ?, f_metadata = ?
		WHERE f_id = ?`

	msgStr, _ := jsoniter.MarshalToString(t)
	trace.SetAttributes(newCtx, attribute.String(trace.TABLE_NAME, TASKINSTANCE_TABLENAME), attribute.String(trace.DB_SQL, sql), attribute.String(trace.DB_Values, msgStr))

	return db.Exec(sql,
		t.UpdatedAt,
		t.TaskID,
		t.DagInsID,
		t.Name,
		t.DependOn,
		t.ActionName,
		t.TimeoutSecs,
		t.Params,
		t.Traces,
		t.Status,
		t.Reason,
		t.PreChecks,
		t.Results,
		t.Steps,
		t.LastModifiedAt,
		t.RenderedParams,
		t.Hash,
		t.Settings,
		t.MetaData,
		t.ID,
	).Error
}

// UpdateToken
func (d *dag) UpdateToken(token *entity.Token) error {
	db, _, cancel := d.dbWithContext(nil)
	defer cancel()

	baseInfo := token.GetBaseInfo()
	baseInfo.Update()

	sql := `UPDATE t_flow_token SET f_updated_at = ?, f_token = ?, f_expires_in = ? WHERE f_user_id = ?`
	res := db.Exec(sql, baseInfo.UpdatedAt, token.Token, token.ExpiresIn, token.UserID)
	if res.Error != nil {
		return fmt.Errorf("update token failed: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("t_flow_token has no key[ %s ] to update: %w", baseInfo.ID, data.ErrDataNotFound)
	}
	return nil
}
