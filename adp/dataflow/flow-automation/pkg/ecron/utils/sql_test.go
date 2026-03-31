package utils

import (
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func newSQLFactory() *ecronSQL {
	return &ecronSQL{}
}

func TestIsParameterInject(t *testing.T) {
	Convey("IsParameterInject", t, func() {
		f := newSQLFactory()

		params := []map[string]string{
			0:  {"-": "-"},
			1:  {"--": "--"},
			2:  {"'": "'"},
			3:  {"''": "''"},
			4:  {"'#": "'#"},
			5:  {"'##": "'##"},
			6:  {"#": "#"},
			7:  {"##": "##"},
			8:  {"#'": "#'"},
			9:  {";": ";"},
			10: {";--": ";--"},
			11: {"1=1": "1=1"},
			12: {"=": "="},
			13: {"or": "or"},
			14: {"and": "and"},
		}

		_, flag := f.IsParameterInject(params[0])
		assert.Equal(t, false, flag)
		for i := 1; i < len(params); i++ {
			_, flag := f.IsParameterInject(params[i])
			assert.Equal(t, true, flag)
		}
	})
}

func TestInsertJobSQL(t *testing.T) {
	Convey("test InsertJobSQL", t, func() {
		f := newSQLFactory()

		fields := []string{
			0: "`f_job_id`",
			1: "`f_job_name`",
			2: "`f_job_cron_time`",
			3: "`f_job_type`",
		}

		sqlStr := f.InsertJobSQL(fields)
		assert.Equal(t, sqlStr, "insert into t_cron_job(`f_job_id`,`f_job_name`,`f_job_cron_time`,`f_job_type`) values(?,?,?,?)")
	})
}

func TestUpdateJobSQL(t *testing.T) {
	Convey("test UpdateJobSQL", t, func() {
		f := newSQLFactory()

		fields := []string{
			0: "`f_job_id`",
			1: "`f_job_name`",
			2: "`f_job_cron_time`",
			3: "`f_job_type`",
			4: "`f_job_context`",
			5: "`f_enabled`",
			6: "`f_remarks`",
			7: "`f_create_time`",
		}

		value := []interface{}{
			0: "111",
			1: "test",
			2: "0/5 * * * *",
			3: 2,
			4: "aaa",
			5: 1,
			6: "bbb",
			7: 123456789,
		}

		visitor := common.Visitor{
			ClientID: "111",
			Name:     "a",
			Admin:    true,
		}

		sqlStr, args := f.UpdateJobSQL(fields, value, visitor)
		expectSQLStr := "update t_cron_job set `f_job_id` = ?, `f_job_name` = ?, `f_job_cron_time` = ?, `f_job_type` = ?, `f_job_context` = ?, `f_enabled` = ?, `f_remarks` = ? where `f_create_time` = ?"

		assert.Equal(t, sqlStr, expectSQLStr)
		assert.Equal(t, len(args), 8)
	})
}

func TestDeleteJobSQL(t *testing.T) {
	Convey("test DeleteJobSQL", t, func() {
		f := newSQLFactory()

		fields := []string{
			0: "`f_job_id`",
			1: "`f_job_name`",
			2: "`f_job_cron_time`",
			3: "`f_job_type`",
			4: "`f_job_context`",
			5: "`f_enabled`",
			6: "`f_remarks`",
			7: "`f_create_time`",
		}

		jobID := []string{
			0: "111",
			1: "222",
			2: "333",
		}

		visitor := common.Visitor{
			ClientID: "111",
			Name:     "a",
			Admin:    true,
		}

		sqlStr, args := f.DeleteJobSQL(fields, jobID, visitor)
		expectSQLStr := "delete from t_cron_job where `f_job_id` in (?,?,?)"

		assert.Equal(t, sqlStr, expectSQLStr)
		assert.Equal(t, len(args), 3)
	})
}

func TestDeletedJobStatusQuerySQL(t *testing.T) {
	Convey("test DeletedJobStatusQuerySQL", t, func() {
		f := newSQLFactory()

		jobID := []string{
			0: "111",
			1: "222",
			2: "333",
		}

		sqlStr, args := f.DeletedJobStatusQuerySQL(jobID)
		expectSQLStr := "select `f_job_id`, `f_ext_info` from t_cron_job_status where `f_job_id` in (?,?,?)"

		assert.Equal(t, sqlStr, expectSQLStr)
		assert.Equal(t, len(args), 3)
	})
}

func TestDeletedJobStatusUpdateSQL(t *testing.T) {
	Convey("test DeletedJobStatusUpdateSQL", t, func() {
		f := newSQLFactory()

		sqlStr := f.DeletedJobStatusUpdateSQL()
		expectSQLStr := "update t_cron_job_status set `f_ext_info`=? where `f_job_id`=?"

		assert.Equal(t, sqlStr, expectSQLStr)
	})
}

func TestGetJobSQL(t *testing.T) {
	Convey("test GetJobSQL", t, func() {
		f := newSQLFactory()

		fields := []string{
			0: "`f_job_id`",
			1: "`f_job_name`",
			2: "`f_job_cron_time`",
			3: "`f_job_type`",
		}

		sqlStr := f.GetJobSQL(fields)
		expectSQLStr := "select `f_job_id`,`f_job_name`,`f_job_cron_time`,`f_job_type` from t_cron_job"

		assert.Equal(t, sqlStr, expectSQLStr)
	})
}

func TestGetJobStatusSQL(t *testing.T) {
	Convey("test GetJobStatusSQL", t, func() {
		f := newSQLFactory()

		fields := []string{
			0: "`f_job_id`",
			1: "`f_job_name`",
			2: "`f_job_cron_time`",
			3: "`f_job_type`",
		}

		sqlStr := f.GetJobStatusSQL(fields)
		expectSQLStr := "select `f_job_id`,`f_job_name`,`f_job_cron_time`,`f_job_type` from t_cron_job_status"

		assert.Equal(t, sqlStr, expectSQLStr)
	})
}

func TestGetUpdateJobStatusInsertSQL(t *testing.T) {
	Convey("test GetUpdateJobStatusInsertSQL", t, func() {
		f := newSQLFactory()

		fields := []string{
			0: "`f_job_id`",
			1: "`f_job_name`",
			2: "`f_job_cron_time`",
			3: "`f_job_type`",
		}

		set := []string{
			0: "(?,?,?,?)",
			1: "(?,?,?,?)",
			2: "(?,?,?,?)",
			3: "(?,?,?,?)",
		}

		sqlStr := f.GetUpdateJobStatusInsertSQL(fields, set)
		expectSQLStr := "insert into t_cron_job_status (`f_job_id`,`f_job_name`,`f_job_cron_time`,`f_job_type`) values (?,?,?,?),(?,?,?,?),(?,?,?,?),(?,?,?,?)"

		assert.Equal(t, sqlStr, expectSQLStr)
	})
}

func TestBatchJobEnableSQL(t *testing.T) {
	Convey("test BatchJobEnableSQL", t, func() {
		f := newSQLFactory()

		fields := []string{
			0: "`f_enabled`",
			1: "`f_update_time`",
			2: "`f_job_id`",
			3: "`f_tenant_id`",
		}

		jobID := []string{
			0: "111",
			1: "222",
			2: "333",
		}

		visitor := common.Visitor{
			ClientID: "111",
			Name:     "a",
			Admin:    true,
		}

		sqlStr, args := f.BatchJobEnableSQL(fields, jobID, visitor)
		expectSQLStr := "update t_cron_job set `f_enabled` = ?, `f_update_time` = ? where `f_job_id` in (?,?,?)"

		assert.Equal(t, sqlStr, expectSQLStr)
		assert.Equal(t, len(args), 3)
	})
}

func TestBatchJobNotifyQuerySQL(t *testing.T) {
	Convey("test BatchJobNotifyQuerySQL", t, func() {
		f := newSQLFactory()

		jobID := []string{
			0: "111",
			1: "222",
			2: "333",
		}

		sqlStr, args := f.BatchJobNotifyQuerySQL(jobID)
		expectSQLStr := "select `f_job_id`, `f_tenant_id`, `f_job_context` from t_cron_job where `f_job_id` in (?,?,?)"

		assert.Equal(t, sqlStr, expectSQLStr)
		assert.Equal(t, len(args), 3)
	})
}
