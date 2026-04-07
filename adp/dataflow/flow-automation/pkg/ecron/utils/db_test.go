package utils

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/robfig/cron/v3"
	uuid "github.com/satori/go.uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func newDB() *ecronDB {
	return &ecronDB{
		db:       nil,
		dataDict: common.NewDataDict(),
		jobInfo: jobInfoFields{
			fJobID: 0, fJobName: 1, fJobCronTime: 2, fJobType: 3, fContext: 4,
			fEnabled: 5, fRemarks: 6, fCreateTime: 7, fUpdateTime: 8, fTenantID: 9,
			fields: []string{0: "`f_job_id`", 1: "`f_job_name`", 2: "`f_job_cron_time`", 3: "`f_job_type`", 4: "`f_job_context`",
				5: "`f_enabled`", 6: "`f_remarks`", 7: "`f_create_time`", 8: "`f_update_time`", 9: "`f_tenant_id`",
			},
		},
		jobStatus: jobStatusFields{fExecuteID: 0, fJobID: 1, fJobType: 2, fJobName: 3, fJobStatus: 4,
			fBeginTime: 5, fEndTime: 6, fExecutor: 7, fExecuteTimes: 8, fExtInfo: 9,
			fields: []string{0: "`f_execute_id`", 1: "`f_job_id`", 2: "`f_job_type`", 3: "`f_job_name`", 4: "`f_job_status`",
				5: "`f_begin_time`", 6: "`f_end_time`", 7: "`f_executor`", 8: "`f_execute_times`", 9: "`f_ext_info`",
			},
		},
		parser:    cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
		connector: sqlx.NewDB,
	}
}

func TestNewDBService(t *testing.T) {
	Convey("NewDBService", t, func() {
		service := NewDBClient()
		assert.NotEqual(t, service, nil)
	})
}

func TestECronDBConnect(t *testing.T) {
	Convey("connect", t, func() {
		Convey("db open failed", func() {
			db, _, err := sqlx.New()
			assert.Equal(t, err, nil)
			defer db.Close()

			edb := newDB()
			edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
				return db, errors.New(common.ErrDataBaseUnavailable)
			}
			assert.NotEqual(t, edb, nil)
			ecronErr := edb.Connect()
			assert.Equal(t, ecronErr.Code, common.InternalError)
		})

		Convey("db open success", func() {
			db, _, err := sqlx.New()
			assert.Equal(t, err, nil)
			defer db.Close()

			edb := newDB()
			edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
				return db, nil
			}
			assert.NotEqual(t, edb, nil)
			ecronErr := edb.Connect()
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))
		})
	})
}

func TestECronDBRelease(t *testing.T) {
	Convey("Release", t, func() {
		Convey("db is nil", func() {
			edb := newDB()
			assert.NotEqual(t, edb, nil)
			edb.Release()
		})

		Convey("db is ok", func() {
			db, _, err := sqlx.New()
			assert.Equal(t, err, nil)
			defer db.Close()

			edb := newDB()
			edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
				return db, nil
			}
			assert.NotEqual(t, edb, nil)
			ecronErr := edb.Connect()
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))
			edb.Release()
		})
	})
}

func TestECronDBPing(t *testing.T) {
	Convey("Ping", t, func() {
		Convey("db is unavailable", func() {
			edb := newDB()
			assert.NotEqual(t, edb, nil)
			ecronErr := edb.Ping()
			assert.Equal(t, ecronErr.Cause, common.ErrDataBaseUnavailable)
		})

		Convey("db is available", func() {
			db, _, err := sqlx.New()
			assert.Equal(t, err, nil)
			defer db.Close()

			edb := newDB()
			edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
				return db, nil
			}
			assert.NotEqual(t, edb, nil)
			ecronErr := edb.Connect()
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))

			ecronErr = edb.Ping()
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))
		})
	})
}

func TestECronDBIsDataBaseAvailable(t *testing.T) {
	Convey("isDataBaseAvailable", t, func() {
		Convey("db is unavailable", func() {
			edb := newDB()
			assert.NotEqual(t, edb, nil)
			ecronErr, ok := edb.isDataBaseAvailable()
			assert.Equal(t, ecronErr.Cause, common.ErrDataBaseUnavailable)
			assert.Equal(t, ok, false)
		})

		Convey("db is available", func() {
			db, _, err := sqlx.New()
			assert.Equal(t, err, nil)
			defer db.Close()

			edb := newDB()
			edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
				return db, nil
			}
			assert.NotEqual(t, edb, nil)
			ecronErr := edb.Connect()
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))

			ecronErr, ok := edb.isDataBaseAvailable()
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))
			assert.Equal(t, ok, true)
		})
	})
}

func TestECronDBInsertJob(t *testing.T) {
	Convey("InsertJob, db is available, exec sql return success", t, func() {
		db, mock, err := sqlx.New()
		assert.Equal(t, err, nil)
		defer db.Close()

		edb := newDB()
		edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
			return db, nil
		}
		assert.NotEqual(t, edb, nil)
		ecronErr := edb.Connect()
		assert.Equal(t, ecronErr, (*common.ECronError)(nil))

		beginTime := time.Now().Format(time.RFC3339)
		duration, _ := time.ParseDuration("1h")
		endTime := time.Now().Add(duration).Format(time.RFC3339)

		visitor := common.Visitor{}

		Convey("time illegal", func() {
			job := common.JobInfo{
				JobName: "test",
				Context: common.JobContext{
					Mode:      common.HTTP,
					BeginTime: "123",
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}

			Convey("begin time is illegal", func() {
				ecronErr := edb.InsertJob(job, visitor)
				assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			})

			Convey("end time is illegal", func() {
				job.Context.BeginTime = beginTime
				job.Context.EndTime = "2020-01-08"
				ecronErr = edb.InsertJob(job, visitor)
				assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			})

			Convey("begin time is greater than end time", func() {
				job.Context.BeginTime = endTime
				job.Context.EndTime = beginTime
				ecronErr = edb.InsertJob(job, visitor)
				assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			})
		})

		Convey("unsupported execution mode", func() {
			job := common.JobInfo{
				JobName: "test",
				Context: common.JobContext{
					Mode: "",
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}
			ecronErr := edb.InsertJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			assert.Equal(t, ecronErr.Cause, common.ErrUnsupportedExecutionMode)
		})

		Convey("cron time error", func() {
			job := common.JobInfo{
				JobName: "test",
				Context: common.JobContext{
					Mode: common.EXE,
				},
				JobType: common.TIMING,
			}
			ecronErr := edb.InsertJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			assert.Contains(t, ecronErr.Cause, common.ErrCronTime)
		})

		Convey("job name is empty", func() {
			job := common.JobInfo{
				Context: common.JobContext{
					Mode: common.EXE,
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}
			ecronErr := edb.InsertJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			assert.Equal(t, ecronErr.Cause, common.ErrJobNameEmpty)
		})

		Convey("job name does exist", func() {
			job := common.JobInfo{
				JobName: "test",
				Context: common.JobContext{
					Mode: common.EXE,
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			ecronErr := edb.InsertJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			assert.Equal(t, ecronErr.Cause, common.ErrJobNameExists)
		})

		Convey("job type is illegal", func() {
			job := common.JobInfo{
				JobName: "test",
				Context: common.JobContext{
					Mode: common.EXE,
				},
				JobCronTime: "*/10 * * * * ?",
			}
			ecronErr := edb.InsertJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			assert.Equal(t, ecronErr.Cause, common.ErrJobTypeIllegal)
		})

		Convey("normal", func() {
			job := common.JobInfo{
				JobName: "test",
				Context: common.JobContext{
					Mode:      common.HTTP,
					BeginTime: beginTime,
					EndTime:   endTime,
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}

			mock.ExpectExec("").WillReturnError(errors.New("unknown"))
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

			ecronErr := edb.InsertJob(job, visitor)
			assert.Equal(t, ecronErr.Code, common.InternalError)
		})
	})
}

func TestECronDBUpdateJob(t *testing.T) {
	Convey("UpdateJob, db is available, exec sql return success, query sql return success but row is empty", t, func() {
		db, mock, err := sqlx.New()
		assert.Equal(t, err, nil)
		defer db.Close()

		edb := newDB()
		edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
			return db, nil
		}
		assert.NotEqual(t, edb, nil)
		ecronErr := edb.Connect()
		assert.Equal(t, ecronErr, (*common.ECronError)(nil))

		beginTime := time.Now().Format(time.RFC3339)
		duration, _ := time.ParseDuration("1h")
		endTime := time.Now().Add(duration).Format(time.RFC3339)

		id := uuid.NewV4().String()

		visitor := common.Visitor{}

		Convey("time illegal", func() {
			job := common.JobInfo{
				JobName: "test",
				Context: common.JobContext{
					Mode:      common.HTTP,
					BeginTime: "123",
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}
			ecronErr := edb.UpdateJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))

			job.Context.BeginTime = beginTime
			job.Context.EndTime = "2020-01-08"
			ecronErr = edb.UpdateJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))

			job.Context.BeginTime = endTime
			job.Context.EndTime = beginTime
			ecronErr = edb.UpdateJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
		})

		Convey("job does not exist", func() {
			job := common.JobInfo{
				JobName: "test",
				Context: common.JobContext{
					Mode:      common.HTTP,
					BeginTime: beginTime,
					EndTime:   endTime,
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}
			Convey("job id is empty", func() {
				ecronErr := edb.UpdateJob(job, visitor)
				assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
				assert.Equal(t, ecronErr.Code, common.NotFound)
			})

			Convey("can't find job id", func() {
				mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(""))

				job.JobID = id
				ecronErr = edb.UpdateJob(job, visitor)
				assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
				assert.Equal(t, ecronErr.Code, common.NotFound)
			})

		})

		Convey("unsupported execution mode", func() {
			job := common.JobInfo{
				JobID:   id,
				JobName: "test",
				Context: common.JobContext{
					Mode: "",
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}

			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))

			ecronErr := edb.UpdateJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			assert.Equal(t, ecronErr.Cause, common.ErrUnsupportedExecutionMode)
		})

		Convey("job name is empty", func() {
			job := common.JobInfo{
				JobID: id,
				Context: common.JobContext{
					Mode: common.EXE,
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}
			ecronErr := edb.UpdateJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			assert.Equal(t, ecronErr.Cause, common.ErrJobNameEmpty)
		})

		Convey("cron time error", func() {
			job := common.JobInfo{
				JobID:   id,
				JobName: "test",
				Context: common.JobContext{
					Mode: common.EXE,
				},
				JobType: common.TIMING,
			}
			ecronErr := edb.InsertJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			assert.Contains(t, ecronErr.Cause, common.ErrCronTime)
		})

		Convey("job type is illegal", func() {
			job := common.JobInfo{
				JobName: "test",
				Context: common.JobContext{
					Mode: common.EXE,
				},
				JobCronTime: "*/10 * * * * ?",
			}
			ecronErr := edb.InsertJob(job, visitor)
			assert.NotEqual(t, ecronErr, (*common.ECronError)(nil))
			assert.Equal(t, ecronErr.Cause, common.ErrJobTypeIllegal)
		})

		Convey("normal", func() {
			job := common.JobInfo{
				JobID:   id,
				JobName: "test",
				Context: common.JobContext{
					Mode:      common.HTTP,
					BeginTime: beginTime,
					EndTime:   endTime,
				},
				JobCronTime: "*/10 * * * * ?",
				JobType:     common.TIMING,
			}

			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			mock.ExpectExec("").WillReturnError(errors.New("unknown"))
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(0))

			ecronErr := edb.UpdateJob(job, visitor)
			assert.Equal(t, ecronErr.Code, common.InternalError)
		})
	})
}

func TestECronDBDeleteJob(t *testing.T) {
	Convey("DeleteJob, db is available", t, func() {
		db, mock, err := sqlx.New()
		assert.Equal(t, err, nil)
		defer db.Close()

		edb := newDB()
		edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
			return db, nil
		}
		assert.NotEqual(t, edb, nil)
		ecronErr := edb.Connect()
		assert.Equal(t, ecronErr, (*common.ECronError)(nil))

		// beginTime := time.Now().Format(time.RFC3339)
		// duration, _ := time.ParseDuration("1h")
		// endTime := time.Now().Add(duration).Format(time.RFC3339)

		id := uuid.NewV4().String()

		visitor := common.Visitor{}

		Convey("execute failed, rollback", func() {
			mock.ExpectBegin().WillReturnError(nil)
			mock.ExpectExec("").WillReturnError(errors.New("unknown"))
			mock.ExpectRollback().WillReturnError(nil)
			ecronErr = edb.DeleteJob(id, visitor)
			assert.Equal(t, ecronErr.Code, common.InternalError)
		})
	})
}

func TestECronDBGetJob(t *testing.T) {
	Convey("GetJob, db is available", t, func() {
		db, mock, err := sqlx.New()
		assert.Equal(t, err, nil)
		defer db.Close()

		edb := newDB()
		edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
			return db, nil
		}
		assert.NotEqual(t, edb, nil)
		ecronErr := edb.Connect()
		assert.Equal(t, ecronErr, (*common.ECronError)(nil))

		beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)

		id := uuid.NewV4().String()

		visitor := common.Visitor{}

		Convey("checkGetJobInfoBefore return err", func() {
			params := common.JobInfoQueryParams{
				Limit: 0,
				Page:  0,
			}

			_, ecronErr = edb.GetJob(params, visitor)
			assert.Equal(t, ecronErr.Cause, common.ErrLimitOrPageIllegal)

			params.Limit = 10
			params.Page = 1
			params.TimeStamp = "2020-01-08"
			_, ecronErr = edb.GetJob(params, visitor)
			assert.Equal(t, ecronErr.Cause, common.ErrTimeIllegal)

			params.TimeStamp = time.Now().Format(time.RFC3339)
			params.JobType = "test"
			_, ecronErr = edb.GetJob(params, visitor)
			assert.Equal(t, common.ErrJobTypeIllegal, ecronErr.Cause)
		})

		Convey("normal", func() {
			params := common.JobInfoQueryParams{
				Limit:     10,
				Page:      1,
				TimeStamp: beginTime,
			}
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(edb.jobInfo.fields).AddRow(id, "test", "*/10 * * * * ?", 1, "{}", true, "", 0, 0, ""))
			_, ecronErr := edb.GetJob(params, visitor)
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))
		})

		Convey("normal, but unmarshal job context failed", func() {
			params := common.JobInfoQueryParams{
				Limit:     10,
				Page:      1,
				TimeStamp: beginTime,
			}
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(edb.jobInfo.fields).AddRow(id, "test", "*/10 * * * * ?", 1, "", true, "", 0, 0, ""))
			_, ecronErr := edb.GetJob(params, visitor)
			assert.Contains(t, ecronErr.Cause, common.ErrUnMarshalJSON)
		})

		Convey("checkGetJobInfoAfter return err, jobInfo's id does not match", func() {
			params := common.JobInfoQueryParams{
				Limit:     10,
				Page:      1,
				TimeStamp: beginTime,
				JobID: []string{
					0: id,
				},
			}

			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(edb.jobInfo.fields).AddRow(uuid.NewV4().String(), "test", "*/10 * * * * ?", 1, "{}", true, "", 0, 0, ""))
			_, ecronErr := edb.GetJob(params, visitor)
			assert.Contains(t, ecronErr.Cause, common.ErrJobNotExist)
		})

		Convey("checkGetJobInfoAfter return nil, although jobInfo's type does not match", func() {
			params := common.JobInfoQueryParams{
				Limit:     10,
				Page:      1,
				TimeStamp: beginTime,
				JobType:   common.TIMING,
			}

			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(edb.jobInfo.fields).AddRow(uuid.NewV4().String(), "test", "*/10 * * * * ?", 2, "{}", true, "", 0, 0, ""))
			_, ecronErr := edb.GetJob(params, visitor)
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))
		})
	})
}

func TestECronDBGetJobTotal(t *testing.T) {
	Convey("GetJobTotal, db is available", t, func() {
		db, mock, err := sqlx.New()
		assert.Equal(t, err, nil)
		defer db.Close()

		edb := newDB()
		edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
			return db, nil
		}
		assert.NotEqual(t, edb, nil)
		ecronErr := edb.Connect()
		assert.Equal(t, ecronErr, (*common.ECronError)(nil))

		beginTime := time.Now().Format(time.RFC3339)
		duration, _ := time.ParseDuration("1h")
		endTime := time.Now().Add(duration).Format(time.RFC3339)

		//id := uuid.NewV4().String()

		visitor := common.Visitor{}

		Convey("checkGetJobTotalBefore return err", func() {
			Convey("begin time is illegal", func() {
				params := common.JobTotalQueryParams{
					BeginTime: "2020-01-08",
				}

				_, ecronErr = edb.GetJobTotal(params, visitor)
				assert.Equal(t, ecronErr.Cause, common.ErrTimeIllegal)
			})

			Convey("begin time is greater than end time", func() {
				params := common.JobTotalQueryParams{
					BeginTime: endTime,
					EndTime:   beginTime,
				}

				_, ecronErr = edb.GetJobTotal(params, visitor)
				assert.Contains(t, ecronErr.Cause, common.ErrBeginTimeGreaterThanEndTime)
			})
		})

		Convey("normal", func() {
			params := common.JobTotalQueryParams{}
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(100))
			total, ecronErr := edb.GetJobTotal(params, visitor)
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))
			assert.Equal(t, total.Total, 100)
		})
	})
}

func TestECronDBGetJobStatus(t *testing.T) {
	Convey("GetJobStatus, db is available", t, func() {
		db, mock, err := sqlx.New()
		assert.Equal(t, err, nil)
		defer db.Close()

		edb := newDB()
		edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
			return db, nil
		}
		assert.NotEqual(t, edb, nil)
		ecronErr := edb.Connect()
		assert.Equal(t, ecronErr, (*common.ECronError)(nil))

		beginTime := time.Now().Format(time.RFC3339)
		duration, _ := time.ParseDuration("1h")
		endTime := time.Now().Add(duration).Format(time.RFC3339)

		id := uuid.NewV4().String()

		visitor := common.Visitor{}

		Convey("checkGetJobStatusBefore return err", func() {
			Convey("params is empty", func() {
				params := common.JobStatusQueryParams{}
				_, ecronErr = edb.GetJobStatus(params, visitor)
				assert.Contains(t, ecronErr.Cause, common.ErrQueryParameterIsNull)
			})

			Convey("begin time is illegal", func() {
				params := common.JobStatusQueryParams{
					BeginTime: "2020-01-08",
				}

				_, ecronErr = edb.GetJobStatus(params, visitor)
				assert.Equal(t, ecronErr.Cause, common.ErrTimeIllegal)
			})

			Convey("begin time is greater than end time", func() {
				params := common.JobStatusQueryParams{
					BeginTime: endTime,
					EndTime:   beginTime,
				}

				_, ecronErr = edb.GetJobStatus(params, visitor)
				assert.Contains(t, ecronErr.Cause, common.ErrBeginTimeGreaterThanEndTime)
			})

			Convey("job does not exist", func() {
				params := common.JobStatusQueryParams{
					BeginTime: beginTime,
					EndTime:   endTime,
					JobID:     id,
				}
				mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(""))
				_, ecronErr = edb.GetJobStatus(params, visitor)
				assert.Contains(t, ecronErr.Cause, common.ErrJobNotExist)
			})

			Convey("job type illegal", func() {
				params := common.JobStatusQueryParams{
					BeginTime: beginTime,
					EndTime:   endTime,
				}
				params.JobType = "test"
				_, ecronErr = edb.GetJobStatus(params, visitor)
				assert.Equal(t, common.ErrJobTypeIllegal, ecronErr.Cause)
			})

			Convey("job status illegal", func() {
				params := common.JobStatusQueryParams{
					BeginTime: beginTime,
					EndTime:   endTime,
				}
				params.JobStatus = "test"
				_, ecronErr = edb.GetJobStatus(params, visitor)
				assert.Equal(t, common.ErrJobStatusIllegal, ecronErr.Cause)
			})
		})

		Convey("scan error", func() {
			params := common.JobStatusQueryParams{
				BeginTime: beginTime,
				EndTime:   endTime,
				JobID:     id,
			}
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(edb.jobStatus.fields[edb.jobStatus.fExecuteID:]).AddRow(uuid.NewV4().String(), id, "", "Test", 1, 0, 0, "[]", 1, "{}"))
			_, ecronErr = edb.GetJobStatus(params, visitor)
			assert.Contains(t, ecronErr.Cause, common.ErrScanFieldValue)
		})

		Convey("normal", func() {
			params := common.JobStatusQueryParams{
				BeginTime: beginTime,
				EndTime:   endTime,
				JobID:     id,
			}
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(edb.jobStatus.fields[edb.jobStatus.fExecuteID:]).AddRow(uuid.NewV4().String(), id, 1, "Test", 1, 0, 0, "[]", 1, "{}"))

			_, ecronErr = edb.GetJobStatus(params, visitor)
			assert.Equal(t, ecronErr, (*common.ECronError)(nil))
		})
	})
}

func TestECronDBUpdateJobStatus(t *testing.T) {
	Convey("UpdateJobStatus, db is available", t, func() {
		db, _, err := sqlx.New()
		assert.Equal(t, err, nil)
		defer db.Close()

		edb := newDB()
		edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
			return db, nil
		}
		assert.NotEqual(t, edb, nil)
		ecronErr := edb.Connect()
		assert.Equal(t, ecronErr, (*common.ECronError)(nil))

		visitor := common.Visitor{}

		Convey("status is null, but return nil", func() {
			_, ecronErr = edb.UpdateJobStatus(nil, visitor)
			assert.Equal(t, ecronErr.Cause, common.ErrStatusEmpty)
		})
	})
}

func TestECronDBBatchJobEnable(t *testing.T) {
	Convey("BatchJobEnable, db is available", t, func() {
		db, mock, err := sqlx.New()
		assert.Equal(t, err, nil)
		defer db.Close()

		edb := newDB()
		edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
			return db, nil
		}
		assert.NotEqual(t, edb, nil)
		ecronErr := edb.Connect()
		assert.Equal(t, ecronErr, (*common.ECronError)(nil))

		beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)

		id := []string{
			0:  uuid.NewV4().String(),
			1:  uuid.NewV4().String(),
			2:  uuid.NewV4().String(),
			3:  uuid.NewV4().String(),
			4:  uuid.NewV4().String(),
			5:  uuid.NewV4().String(),
			6:  uuid.NewV4().String(),
			7:  uuid.NewV4().String(),
			8:  uuid.NewV4().String(),
			9:  uuid.NewV4().String(),
			10: uuid.NewV4().String(),
			11: uuid.NewV4().String(),
		}

		visitor := common.Visitor{}

		Convey("checkBatchJobEnableBefore return err", func() {
			//此处mock两次query，是因为查询ID是否存在API会执行两次，系统默认每10个ID查询一次，以防sql过长导致查询失败
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id[0]))
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id[0]))
			ecronErr = edb.BatchJobEnable(id, true, beginTime, visitor)
			assert.Contains(t, ecronErr.Cause, common.ErrJobNotExist)
		})

		Convey("normal, but update failed", func() {
			id = []string{
				0: uuid.NewV4().String(),
			}
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id[0]))
			mock.ExpectExec("").WillReturnError(errors.New("unknown"))
			ecronErr = edb.BatchJobEnable(id, true, beginTime, visitor)
			assert.Equal(t, ecronErr.Code, common.InternalError)
		})
	})
}

func TestECronDBBatchJobNotify(t *testing.T) {
	Convey("BatchJobNotify, db is available", t, func() {
		db, mock, err := sqlx.New()
		assert.Equal(t, err, nil)
		defer db.Close()

		edb := newDB()
		edb.connector = func(connInfo *sqlx.DBConfig) (*sqlx.DB, error) {
			return db, nil
		}
		assert.NotEqual(t, edb, nil)
		ecronErr := edb.Connect()
		assert.Equal(t, ecronErr, (*common.ECronError)(nil))

		beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)

		id := []string{
			0:  uuid.NewV4().String(),
			1:  uuid.NewV4().String(),
			2:  uuid.NewV4().String(),
			3:  uuid.NewV4().String(),
			4:  uuid.NewV4().String(),
			5:  uuid.NewV4().String(),
			6:  uuid.NewV4().String(),
			7:  uuid.NewV4().String(),
			8:  uuid.NewV4().String(),
			9:  uuid.NewV4().String(),
			10: uuid.NewV4().String(),
			11: uuid.NewV4().String(),
		}

		visitor := common.Visitor{}

		Convey("checkBatchJobNotifyBefore return err", func() {
			//此处mock两次query，是因为查询ID是否存在API会执行两次，系统默认每10个ID查询一次，以防sql过长导致查询失败
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id[0]))
			mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id[0]))
			_, ecronErr = edb.BatchJobNotify(id, common.JobNotify{Webhook: "https://123"}, beginTime, visitor)
			assert.Contains(t, ecronErr.Cause, common.ErrJobNotExist)
		})
	})
}

func TestECronDBGetJobInfoQueryCondition(t *testing.T) {
	Convey("getJobInfoQueryCondition", t, func() {
		Convey("job id is not empty", func() {
			edb := newDB()
			assert.NotEqual(t, edb, nil)

			id := []string{
				0: uuid.NewV4().String(),
				1: uuid.NewV4().String(),
			}

			visitor := common.Visitor{}

			params := common.JobInfoQueryParams{
				JobID:   id,
				JobType: common.TIMING,
			}
			sql, args := edb.getJobInfoQueryCondition(params, visitor)
			assert.Equal(t, sql, fmt.Sprintf(" where  %v in (?,?)", edb.jobInfo.fields[edb.jobInfo.fJobID]))
			assert.Equal(t, args, []interface{}{id[0], id[1]})
		})

		Convey("without job id", func() {
			edb := newDB()
			assert.NotEqual(t, edb, nil)

			timestamp := time.Now().Format(time.RFC3339)
			params := common.JobInfoQueryParams{
				JobType:   common.TIMING,
				TimeStamp: timestamp,
				Limit:     10,
				Page:      2,
			}

			jt, _ := edb.dataDict.DJobType.StringToInt(params.JobType)
			tt, _ := common.StringToTimeStamp(params.TimeStamp)

			visitor := common.Visitor{}

			Convey("others", func() {
				visitor.ClientID = "123456"
				visitor.Admin = false
				sql, args := edb.getJobInfoQueryCondition(params, visitor)
				assert.Equal(t, sql, " where  `f_job_type` = ?  and  `f_create_time` <= ?  and  `f_tenant_id` = ?  order by `f_create_time` desc limit ?, ? ")
				assert.Equal(t, args, []interface{}{jt, tt, "123456", (params.Page - 1) * params.Limit, params.Limit})
			})

			Convey("normal", func() {
				visitor.ClientID = "123456"
				visitor.Admin = true
				sql, args := edb.getJobInfoQueryCondition(params, visitor)
				assert.Equal(t, sql, " where  `f_job_type` = ?  and  `f_create_time` <= ?  order by `f_create_time` desc limit ?, ? ")
				assert.Equal(t, args, []interface{}{jt, tt, (params.Page - 1) * params.Limit, params.Limit})
			})
		})
	})
}

func TestECronDBGetJobTotalQueryCondition(t *testing.T) {
	Convey("getJobTotalQueryCondition", t, func() {
		edb := newDB()
		assert.NotEqual(t, edb, nil)

		beginTime := time.Now().Format(time.RFC3339)
		duration, _ := time.ParseDuration("1h")
		endTime := time.Now().Add(duration).Format(time.RFC3339)
		timestamp := time.Now().Format(time.RFC3339)

		params := common.JobTotalQueryParams{
			BeginTime: beginTime,
			EndTime:   endTime,
		}

		bt, _ := common.StringToTimeStamp(beginTime)
		et, _ := common.StringToTimeStamp(endTime)
		tt, _ := common.StringToTimeStamp(timestamp)

		visitor := common.Visitor{}

		Convey("others", func() {
			visitor.ClientID = "123456"
			visitor.Admin = false
			sql, args := edb.getJobTotalQueryCondition(params, timestamp, visitor)
			assert.Equal(t, sql, " where  `f_create_time` >= ?  and  `f_create_time` <= ?  and  `f_create_time` <= ?  and  `f_tenant_id` = ? ")
			assert.Equal(t, args, []interface{}{bt, et, tt, "123456"})
		})

		Convey("normal", func() {
			visitor.ClientID = "123456"
			visitor.Admin = true
			sql, args := edb.getJobTotalQueryCondition(params, timestamp, visitor)
			assert.Equal(t, sql, " where  `f_create_time` >= ?  and  `f_create_time` <= ?  and  `f_create_time` <= ? ")
			assert.Equal(t, args, []interface{}{bt, et, tt})
		})
	})
}

func TestECronDBGetJobStatusQueryCondition(t *testing.T) {
	Convey("getJobStatusQueryCondition", t, func() {
		edb := newDB()
		assert.NotEqual(t, edb, nil)

		id := uuid.NewV4().String()
		beginTime := time.Now().Format(time.RFC3339)
		duration, _ := time.ParseDuration("1h")
		endTime := time.Now().Add(duration).Format(time.RFC3339)

		visitor := common.Visitor{}

		Convey("only job id", func() {
			params := common.JobStatusQueryParams{
				JobID: id,
			}

			Convey("admin", func() {
				visitor.Admin = true
				visitor.ClientID = "123456"
				sql, args := edb.getJobStatusQueryCondition(params, visitor)
				assert.Equal(t, sql, " where  `f_job_id` = ?  order by `f_begin_time` desc limit ?, ? ")
				assert.Equal(t, args, []interface{}{id, 0, 1})
			})

			Convey("others", func() {
				visitor.Admin = false
				visitor.ClientID = "123456"
				sql, args := edb.getJobStatusQueryCondition(params, visitor)
				assert.Equal(t, sql, " where  `f_job_id` = ?  order by `f_begin_time` desc limit ?, ? ")
				assert.Equal(t, args, []interface{}{id, 0, 1})
			})
		})

		Convey("normal", func() {
			params := common.JobStatusQueryParams{
				JobID:     id,
				BeginTime: beginTime,
				EndTime:   endTime,
				JobType:   common.TIMING,
				JobStatus: common.SUCCESS,
			}

			bt, _ := common.StringToTimeStamp(params.BeginTime)
			et, _ := common.StringToTimeStamp(params.EndTime)
			jt, _ := edb.dataDict.DJobType.StringToInt(params.JobType)
			js, _ := edb.dataDict.DJobStatus.StringToInt(params.JobStatus)

			sql, args := edb.getJobStatusQueryCondition(params, visitor)
			assert.Equal(t, sql, " where  `f_job_id` = ?  and  `f_begin_time` >= ?  and  `f_begin_time` <= ?  and  `f_job_type` = ?  and  `f_job_status` = ?  order by `f_begin_time` desc")
			assert.Equal(t, args, []interface{}{id, bt, et, jt, js})
		})
	})
}

func TestSQLInject(t *testing.T) {
	Convey("inject sql", t, func() {
		//连接数据库，注入测试
		realDB := &ecronDB{
			db:       nil,
			dataDict: common.NewDataDict(),
			jobInfo: jobInfoFields{
				fJobID: 0, fJobName: 1, fJobCronTime: 2, fJobType: 3, fContext: 4, fTenantID: 5, fEnabled: 6, fRemarks: 7, fCreateTime: 8, fUpdateTime: 9,
				fields: []string{0: "`f_job_id`", 1: "`f_job_name`", 2: "`f_job_cron_time`", 3: "`f_job_type`", 4: "`f_job_context`",
					5: "`f_tenant_id`", 6: "`f_enabled`", 7: "`f_remarks`", 8: "`f_create_time`", 9: "`f_update_time`",
				},
			},
			jobStatus: jobStatusFields{
				fExecuteID: 0, fJobID: 1, fJobType: 2, fJobName: 3, fJobStatus: 4, fBeginTime: 5, fEndTime: 6, fExecutor: 7, fExecuteTimes: 8, fExtInfo: 9,
				fields: []string{0: "`f_execute_id`", 1: "`f_job_id`", 2: "`f_job_type`", 3: "`f_job_name`", 4: "`f_job_status`",
					5: "`f_begin_time`", 6: "`f_end_time`", 7: "`f_executor`", 8: "`f_execute_times`", 9: "`f_ext_info`",
				},
			},
			parser:    cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
			connector: sqlx.NewDB,
		}
		err := realDB.Connect()
		if nil != err {
			return
		}

		visitor := common.Visitor{}

		//只有连接数据库成功后，才开始注入攻击测试
		Convey("query job, and do something else", func() {
			//单纯的参数化传参会返回common.NotFound
			//加入参数防注入判断后返回common.BadRequest
			params := common.JobInfoQueryParams{
				Limit:   10,
				Page:    1,
				JobType: common.TIMING,
			}

			Convey("query all job", func() {
				//如果拼接sql, select * from t_cron_job where f_job_id in ('') and 1 or 1 ;--') limit 0, 10
				params.JobID = []string{"') and 1 or 1 ;--"}
				_, err = realDB.GetJob(params, visitor)
				assert.Equal(t, common.BadRequest, err.Code)

				//如果拼接sql, select * from t_cron_job where f_job_id in ('') or 1 = 1 # limit 0, 10
				params.JobID = []string{"') or 1 = 1 #"}
				_, err = realDB.GetJob(params, visitor)
				assert.Equal(t, common.BadRequest, err.Code)

				//如果拼接sql, select * from t_cron_job where f_job_id in (''); union all select * from t_cron_job; -- limit 0, 10
				params.JobID = []string{"'); union all select * from t_cron_job; --"}
				_, err = realDB.GetJob(params, visitor)
				assert.Equal(t, common.BadRequest, err.Code)
			})

			Convey("query nothing", func() {
				//如果拼接sql, select * from t_cron_job where f_job_id in ('') # limit 0, 10
				params.JobID = []string{"') #"}
				_, err = realDB.GetJob(params, visitor)
				assert.Equal(t, common.BadRequest, err.Code)

				params.JobID = []string{"')#"}
				_, err = realDB.GetJob(params, visitor)
				assert.Equal(t, common.BadRequest, err.Code)
			})

			Convey("truncate table", func() {
				//如果拼接sql, select * from t_cron_job where f_job_id in (''); truncate table t_cron_job; --') limit 0, 10
				params.JobID = []string{"'); truncate table t_cron_job; --"}
				_, err = realDB.GetJob(params, visitor)
				assert.Equal(t, common.BadRequest, err.Code)
				params.JobID = []string{"')    ; truncate table t_cron_job; --"}
				_, err = realDB.GetJob(params, visitor)
				assert.Equal(t, common.BadRequest, err.Code)
			})

			Convey("drop database", func() {
				//如果拼接sql, select * from t_cron_job where f_job_id in (''); drop database t_cron_job; --') limit 0, 10
				params.JobID = []string{"'); drop database ecron; --"}
				_, err = realDB.GetJob(params, visitor)
				assert.Equal(t, common.BadRequest, err.Code)
			})
		})
	})
}
