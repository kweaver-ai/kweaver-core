package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
)

var (
	sqlLog = NewLogger()
)

//go:generate mockgen -package mock -source ../utils/sql.go -destination ../mock/mock_sql.go

// SQLFactory SQL语句生产工厂
type SQLFactory interface {
	IsParameterInject(params map[string]string) (string, bool)
	GetFindInSetSQL(value interface{}) (sql string, args []interface{})
	InsertJobSQL(fields []string) string
	UpdateJobSQL(fields []string, value []interface{}, visitor common.Visitor) (sql string, args []interface{})
	DeleteJobSQL(fields []string, jobID []string, visitor common.Visitor) (sql string, args []interface{})
	DeletedJobStatusQuerySQL(jobID []string) (sql string, args []interface{})
	DeletedJobStatusUpdateSQL() string
	GetJobSQL(fields []string) string
	GetJobTotalSQL() string
	GetJobStatusSQL(fields []string) string
	GetUpdateJobStatusQuerySQL() string
	GetUpdateJobStatusUpdateSQL(fields []string) string
	GetUpdateJobStatusInsertSQL(fields []string, set []string) string
	BatchJobEnableSQL(fields []string, jobID []string, visitor common.Visitor) (sql string, args []interface{})
	BatchJobNotifyQuerySQL(jobID []string) (sql string, args []interface{})
	BatchJobNotifyUpdateSQL() string
	CheckJobIDExistSQL(fields []string, jobID []string, visitor common.Visitor) (sql string, args []interface{})
	CheckJobNameSQL(fields []string) string
	AddConditionSQL(condition *string, field string, judge string, value interface{}, args *[]interface{})
	AddWhereSQL(condition *string, args *[]interface{}) (string, []interface{})
	AddOrderBySQL(condition *string, fields map[string]string)
	AddLimitSQL(condition *string, limit int, page int, args *[]interface{})
	IsEmpty(value interface{}) bool
}

// NewSQLFactory 新建SQL工厂
func NewSQLFactory() SQLFactory {
	return &ecronSQL{}
}

type ecronSQL struct {
}

// IsParameterInject 判断用户传参是否存在sql注入攻击
func (s *ecronSQL) IsParameterInject(params map[string]string) (string, bool) {
	str := `(?:--)|(?:'#)|(?:')|(?:#)|(?:;)|(?:1=1)|(?:=)|(\b(or|and)\b)`
	re, err := regexp.Compile(str)
	if nil != err {
		sqlLog.Errorln(err)
		return "regexp compile error", false
	}

	for k, v := range params {
		if re.MatchString(v) {
			return k, true
		}
	}
	return "", false
}

// GetFindInSetSQL 获取集合查询部分的SQL
func (s *ecronSQL) GetFindInSetSQL(value interface{}) (sql string, args []interface{}) {
	switch v := value.(type) {
	case string:
		{
			return s.GetFindInSetSQL(strings.Split(v, ","))
		}
	case []string:
		{
			set := make([]string, 0)
			for _, v2 := range v {
				set = append(set, "?")
				args = append(args, v2)
			}
			sql = strings.Join(set, ",")
		}
	case []int:
		{
			set := make([]string, 0)
			for _, v2 := range v {
				set = append(set, "?")
				args = append(args, v2)
			}
			sql = strings.Join(set, ",")
		}
	}

	return
}

// GetInsertJobSQL 添加任务SQL
func (s *ecronSQL) InsertJobSQL(fields []string) string {
	set, _ := s.GetFindInSetSQL(fields)
	sql := fmt.Sprintf("insert into t_cron_job(%v) values", strings.Join(fields, ","))
	sql += "("
	sql += set
	sql += ")"
	return sql
}

// UpdateJobSQL 更新任务SQL
func (s *ecronSQL) UpdateJobSQL(fields []string, value []interface{}, visitor common.Visitor) (sql string, args []interface{}) {
	args = value
	sql = fmt.Sprintf("update t_cron_job set %v = ?, %v = ?, %v = ?, %v = ?, %v = ?, %v = ?, %v = ? where %v = ?",
		fields[0], fields[1], fields[2], fields[3], fields[4], fields[5], fields[6], fields[7])
	if !visitor.Admin {
		s.AddConditionSQL(&sql, fields[8], "=", visitor.ClientID, &args)
	}
	return
}

// DeleteJobSQL 删除任务SQL
func (s *ecronSQL) DeleteJobSQL(fields []string, jobID []string, visitor common.Visitor) (sql string, args []interface{}) {
	set, args := s.GetFindInSetSQL(jobID)
	sql = fmt.Sprintf("delete from t_cron_job where %v in (", fields[0])
	sql += set
	sql += ")"
	if !visitor.Admin {
		s.AddConditionSQL(&sql, fields[1], "=", visitor.ClientID, &args)
	}
	return
}

// DeletedJobStatusSQL 更新任务状态之任务删除标识SQL
func (s *ecronSQL) DeletedJobStatusQuerySQL(jobID []string) (sql string, args []interface{}) {
	set, args := s.GetFindInSetSQL(jobID)
	sql = "select `f_job_id`, `f_ext_info` from t_cron_job_status where `f_job_id` in ("
	sql += set
	sql += ")"
	return
}

func (s *ecronSQL) DeletedJobStatusUpdateSQL() string {
	return "update t_cron_job_status set `f_ext_info`=? where `f_job_id`=?"
}

// GetJobSQL 获取任务信息SQL
func (s *ecronSQL) GetJobSQL(fields []string) string {
	return fmt.Sprintf("select %v from t_cron_job", strings.Join(fields, ","))
}

func (s *ecronSQL) GetJobTotalSQL() string {
	return "select count(*) as total from t_cron_job"
}

func (s *ecronSQL) GetJobStatusSQL(fields []string) string {
	return fmt.Sprintf("select %s from t_cron_job_status", strings.Join(fields, ","))
}

func (s *ecronSQL) GetUpdateJobStatusQuerySQL() string {
	return "select count(*) from t_cron_job_status where `f_execute_id`=?"
}

func (s *ecronSQL) GetUpdateJobStatusUpdateSQL(fields []string) string {
	return fmt.Sprintf("update t_cron_job_status set %v=?,%v=?,%v=?,%v=?,%v=?,%v=?,%v=?,%v=?,%v=? where `f_execute_id`=?", fields[0], fields[1], fields[2], fields[3], fields[4], fields[5], fields[6], fields[7], fields[8])
}

func (s *ecronSQL) GetUpdateJobStatusInsertSQL(fields []string, set []string) string {
	insertSqlStr := fmt.Sprintf("insert into t_cron_job_status (%v) values ", strings.Join(fields, ","))
	insertSqlStr += strings.Join(set, ",")
	return insertSqlStr
}

func (s *ecronSQL) BatchJobEnableSQL(fields []string, jobID []string, visitor common.Visitor) (sql string, args []interface{}) {
	set, args := s.GetFindInSetSQL(jobID)
	sql = fmt.Sprintf("update t_cron_job set %v = ?, %v = ? where %v in (", fields[0], fields[1], fields[2])
	sql += set
	sql += ")"
	if !visitor.Admin {
		s.AddConditionSQL(&sql, fields[3], "=", visitor.ClientID, &args)
	}
	return
}

func (s *ecronSQL) BatchJobNotifyQuerySQL(jobID []string) (sql string, args []interface{}) {
	set, args := s.GetFindInSetSQL(jobID)
	sql = "select `f_job_id`, `f_tenant_id`, `f_job_context` from t_cron_job where `f_job_id` in ("
	sql += set
	sql += ")"
	return
}

func (s *ecronSQL) BatchJobNotifyUpdateSQL() string {
	return "update t_cron_job set `f_job_context`=?, `f_update_time`=? where `f_job_id`=?"
}

func (s *ecronSQL) CheckJobIDExistSQL(fields []string, jobID []string, visitor common.Visitor) (sql string, args []interface{}) {
	set, args := s.GetFindInSetSQL(jobID)
	sql = fmt.Sprintf("select %v from t_cron_job where %v in (", fields[0], fields[1])
	sql += set
	sql += ")"
	if !visitor.Admin {
		s.AddConditionSQL(&sql, fields[2], "=", visitor.ClientID, &args)
	}
	return
}

func (s *ecronSQL) CheckJobNameSQL(fields []string) string {
	sql := fmt.Sprintf("select count(*) from t_cron_job where %v = ? and %v = ?", fields[0], fields[1])
	return sql
}

// AddConditionSQL 追加条件SQL，condition=待返回的SQL, field=数据库字段, judge=SQL条件判断，value=判断值，args=返回的动态参数列表
func (s *ecronSQL) AddConditionSQL(condition *string, field string, judge string, value interface{}, args *[]interface{}) {
	if s.IsEmpty(value) {
		return
	}

	if !s.IsEmpty(*condition) {
		*condition += " and "
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		{
			set, tmp := s.GetFindInSetSQL(value)
			*condition += fmt.Sprintf(" %v %v (", field, judge)
			*condition += set
			*condition += ")"

			*args = append(*args, tmp...)
		}
	case reflect.Bool:
		*condition += fmt.Sprintf(" %v %v ? ", field, judge)
		if value.(bool) {
			*args = append(*args, 1)
			return
		}
		*args = append(*args, 0)
	default:
		*condition += fmt.Sprintf(" %v %v ? ", field, judge)
		*args = append(*args, value)
	}
}

func (s *ecronSQL) AddWhereSQL(condition *string, args *[]interface{}) (string, []interface{}) {
	if !s.IsEmpty(*condition) {
		*condition = " where " + *condition
	}
	return *condition, *args
}

func (s *ecronSQL) AddOrderBySQL(condition *string, fields map[string]string) {
	*condition += " order by "
	for k, v := range fields {
		*condition += fmt.Sprintf("%v %v,", k, v)
	}
	*condition = strings.TrimRight(*condition, ",")
}

func (s *ecronSQL) AddLimitSQL(condition *string, limit int, page int, args *[]interface{}) {
	if limit >= 0 && page > 0 {
		*condition += " limit ?, ? "
		*args = append(*args, (page-1)*limit)
		*args = append(*args, limit)
	}
}

// IsEmpty 判断数据值是否为空，支持int, string, string slice, int slice，bool无论如何都不可能空
func (s *ecronSQL) IsEmpty(value interface{}) bool {
	switch v := value.(type) {
	case int64:
		return (0 == v)
	case int32:
		return (0 == v)
	case int:
		return (0 == v)
	case string:
		return (len(v) == 0)
	case []string:
		return (len(v) == 0)
	case []int:
		return (len(v) == 0)
	}
	return false
}
