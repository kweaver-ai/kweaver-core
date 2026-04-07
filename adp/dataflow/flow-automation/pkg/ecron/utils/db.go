package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"

	//_ 注册mysql驱动
	jsoniter "github.com/json-iterator/go"
	_ "github.com/kweaver-ai/proton-rds-sdk-go/driver"
	"github.com/robfig/cron/v3"
)

//go:generate mockgen -package mock -source ../utils/db.go -destination ../mock/mock_db.go

// DBClient 数据库服务
type DBClient interface {
	Connect() *common.ECronError
	Release()
	Ping() *common.ECronError
	InsertJob(job common.JobInfo, visitor common.Visitor) *common.ECronError
	UpdateJob(job common.JobInfo, visitor common.Visitor) *common.ECronError
	DeleteJob(jobID string, visitor common.Visitor) *common.ECronError
	GetJob(params common.JobInfoQueryParams, visitor common.Visitor) ([]common.JobInfo, *common.ECronError)
	GetJobTotal(params common.JobTotalQueryParams, visitor common.Visitor) (common.JobTotal, *common.ECronError)
	GetJobStatus(params common.JobStatusQueryParams, visitor common.Visitor) ([]common.JobStatus, *common.ECronError)
	UpdateJobStatus(status []common.JobStatus, visitor common.Visitor) ([]string, *common.ECronError)
	BatchJobEnable(jobID []string, enable bool, updateTime string, visitor common.Visitor) *common.ECronError
	BatchJobNotify(jobID []string, notify common.JobNotify, updatetime string, visitor common.Visitor) ([]string, *common.ECronError)
	CheckJobExecuteMode(executeMode string) *common.ECronError
	Upgrade()
}

// ecronDB mysql客户端信息结构
type ecronDB struct {
	db        *sqlx.DB
	dataDict  *common.DataDict
	jobInfo   jobInfoFields
	jobStatus jobStatusFields
	parser    cron.Parser
	connector func(*sqlx.DBConfig) (*sqlx.DB, error)
}

type jobInfoFields struct {
	fJobID       int
	fJobName     int
	fJobCronTime int
	fJobType     int
	fContext     int
	fEnabled     int
	fRemarks     int
	fCreateTime  int
	fUpdateTime  int
	fTenantID    int

	fields []string
}

type jobStatusFields struct {
	fExecuteID    int
	fJobID        int
	fJobType      int
	fJobName      int
	fJobStatus    int
	fBeginTime    int
	fEndTime      int
	fExecutor     int
	fExecuteTimes int
	fExtInfo      int

	fields []string
}

var (
	dbConfig = NewConfiger()
	dbLog    = NewLogger()
	dbSQL    = NewSQLFactory()
)

// NewDBClient 加载数据库服务
func NewDBClient() DBClient {
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

// Connect 连接数据库
func (d *ecronDB) Connect() *common.ECronError {
	connInfo := sqlx.DBConfig{
		User:         dbConfig.Config().UserName,
		Password:     dbConfig.Config().UserPwd,
		Host:         dbConfig.Config().DBAddr,
		Port:         dbConfig.Config().DBPort,
		Database:     fmt.Sprintf("%s%s", dbConfig.Config().SystemId, dbConfig.Config().DBName),
		Charset:      dbConfig.Config().DBFormat,
		Timeout:      dbConfig.Config().TimeOut,
		ReadTimeout:  dbConfig.Config().ReadTimeOut,
		WriteTimeout: dbConfig.Config().WriteTimeOut,
		MaxOpenConns: dbConfig.Config().MaxOpenConns,
	}

	var sqlErr error
	d.db, sqlErr = d.connector(&connInfo)
	if sqlErr != nil {
		dbLog.Errorf("[Connect] connect mysql failed, sqlErr: %v", sqlErr)
		return NewECronError(common.ErrOpenDataBase, common.InternalError, nil)
	}
	if d.db == nil {
		dbLog.Errorf("[Connect] connect mysql failed, connInfo: %v", connInfo)
	}
	return nil
}

// Release 释放连接
func (d *ecronDB) Release() {
	if (*sqlx.DB)(nil) != d.db {
		defer d.db.Close()
	}
}

// Ping 判断数据库连通性
func (d *ecronDB) Ping() *common.ECronError {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return err
	}

	sqlErr := d.db.Ping()
	if nil != sqlErr {
		return NewECronError(sqlErr.Error(), common.InternalError, nil)
	}
	return nil
}

func (d *ecronDB) isDataBaseAvailable() (*common.ECronError, bool) {
	if (*sqlx.DB)(nil) == d.db {
		return NewECronError(common.ErrDataBaseUnavailable, common.InternalError, nil), false
	}
	return nil, true
}

// InsertJob 添加定时任务
func (d *ecronDB) InsertJob(job common.JobInfo, visitor common.Visitor) *common.ECronError {
	checkErr, context, createTime, updateTime, jobType := d.checkInsertJobBefore(job)
	if (*common.ECronError)(nil) != checkErr {
		return checkErr
	}

	sql := dbSQL.InsertJobSQL(d.jobInfo.fields)
	_, insertErr := d.db.Exec(sql,
		job.JobID,
		job.JobName,
		job.JobCronTime,
		jobType,
		string(context),
		job.Enabled,
		job.Remarks,
		createTime,
		updateTime,
		job.TenantID)
	if nil != insertErr {
		dbLog.Errorf("InsertJob failed, InsertJobSQL: %v, ERROR: %v ", sql, insertErr)
		return d.checkInsertOrUpdateJobAfter(job, NewECronError(common.ErrInsertJob, common.InternalError, nil), visitor)
	}

	return d.checkInsertOrUpdateJobAfter(job, nil, visitor)
}

// UpdateJob 更新定时任务
func (d *ecronDB) UpdateJob(job common.JobInfo, visitor common.Visitor) *common.ECronError {
	checkErr, context, updateTime, jobType := d.checkUpdateJobBefore(job, visitor)
	if (*common.ECronError)(nil) != checkErr {
		return checkErr
	}

	sql, args := dbSQL.UpdateJobSQL(
		[]string{
			d.jobInfo.fields[d.jobInfo.fJobName],
			d.jobInfo.fields[d.jobInfo.fJobCronTime],
			d.jobInfo.fields[d.jobInfo.fJobType],
			d.jobInfo.fields[d.jobInfo.fContext],
			d.jobInfo.fields[d.jobInfo.fEnabled],
			d.jobInfo.fields[d.jobInfo.fRemarks],
			d.jobInfo.fields[d.jobInfo.fUpdateTime],
			d.jobInfo.fields[d.jobInfo.fJobID],
			d.jobInfo.fields[d.jobInfo.fTenantID],
		},
		[]interface{}{
			job.JobName,
			job.JobCronTime,
			jobType,
			string(context),
			job.Enabled,
			job.Remarks,
			updateTime,
			job.JobID}, visitor)
	_, updateErr := d.db.Exec(sql, args...)
	if nil != updateErr {
		dbLog.Errorf("UpdateJob failed, UpdateJobSQL: %v, ERROR: %v", sql, updateErr)
		return d.checkInsertOrUpdateJobAfter(job, NewECronError(common.ErrUpdateJob, common.InternalError, nil), visitor)
	}

	return d.checkInsertOrUpdateJobAfter(job, nil, visitor)
}

// DeleteJob 删除定时任务
func (d *ecronDB) DeleteJob(jobID string, visitor common.Visitor) *common.ECronError {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return err
	}

	if err := d.checkInject(map[string]string{"job_id": jobID}, "DeleteJob()"); (*common.ECronError)(nil) != err {
		return err
	}

	ids := strings.Split(jobID, ",")
	dsql, dargs := dbSQL.DeleteJobSQL([]string{d.jobInfo.fields[d.jobInfo.fJobID], d.jobInfo.fields[d.jobInfo.fTenantID]}, ids, visitor)

	jobStatusList, err := d.getDeletedJobStatusExtInfo(ids)
	if err != nil {
		return err
	}

	tr, sqlErr := d.db.Begin()
	if nil != sqlErr {
		_ = tr.Rollback()
		dbLog.Errorln(sqlErr)
		return NewECronError(common.ErrTransactionBegin, common.InternalError, nil)
	}

	_, sqlErr = d.db.Exec(dsql, dargs...)
	if nil != sqlErr {
		_ = tr.Rollback()
		dbLog.Errorf("DeleteJob failed, DeleteJobSQL: %v, ERROR: %v", dsql, sqlErr)
		return NewECronError(common.ErrDeleteJob, common.InternalError, nil)
	}

	for _, job := range jobStatusList {
		if tenantID, ok := job.ExtInfo[common.TenantID].(string); ok {
			if tenantID != visitor.ClientID {
				continue
			}
		} else {
			continue
		}

		if _, ok := job.ExtInfo[common.IsDeleted]; ok {
			job.ExtInfo[common.IsDeleted] = 1
		} else {
			job.ExtInfo[common.IsDeleted] = 1
		}
		extInfo, jsonErr := jsoniter.Marshal(job.ExtInfo)
		if err, ok := d.marshalJSONError("DeleteJob()", jsonErr); ok {
			return err
		}

		updateSqlStr := dbSQL.DeletedJobStatusUpdateSQL()
		_, sqlErr = d.db.Exec(updateSqlStr, string(extInfo), job.JobID)
		if nil != sqlErr {
			_ = tr.Rollback()
			dbLog.Errorln(sqlErr, updateSqlStr, string(extInfo), job.JobID)
			return NewECronError(common.ErrUpdateJobDeletedFlag, common.InternalError, nil)
		}
	}

	sqlErr = tr.Commit()
	if nil != sqlErr {
		_ = tr.Rollback()
		dbLog.Errorln(sqlErr)
		return NewECronError(common.ErrCommit, common.InternalError, nil)
	}

	return nil
}

func (d *ecronDB) getDeletedJobStatusExtInfo(jobID []string) (jobStatusList []common.JobStatus, err *common.ECronError) {
	querySqlStr, queryArgs := dbSQL.DeletedJobStatusQuerySQL(jobID)
	rows, sqlErr := d.db.Query(querySqlStr, queryArgs...)
	defer func() {
		if nil != rows {
			rows.Close()
		}
	}()

	if nil != sqlErr {
		dbLog.Errorln(sqlErr, querySqlStr, queryArgs)
		return nil, NewECronError(common.ErrQueryJob, common.InternalError, nil)
	}

	var extinfo string
	status := common.JobStatus{}
	jobStatusList = make([]common.JobStatus, 0)

	for rows.Next() {
		extinfo = ""
		status = common.JobStatus{}

		sqlErr := rows.Scan(&status.JobID, &extinfo)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, querySqlStr)
			return nil, NewECronError(common.ErrScanFieldValue, common.InternalError, nil)
		}

		jsonErr := jsoniter.Unmarshal([]byte(extinfo), &status.ExtInfo)
		if err, ok := d.unmarshalJSONError("DeleteJob()", jsonErr); ok {
			return nil, err
		}

		jobStatusList = append(jobStatusList, status)
	}

	return jobStatusList, nil
}

// GetJob 获取定时任务
func (d *ecronDB) GetJob(params common.JobInfoQueryParams, visitor common.Visitor) ([]common.JobInfo, *common.ECronError) {
	if err := d.checkGetJobInfoBefore(params); nil != err {
		return nil, err
	}

	sql := dbSQL.GetJobSQL(d.jobInfo.fields)
	condition, args := d.getJobInfoQueryCondition(params, visitor)
	sql += condition
	rows, sqlErr := d.db.Query(sql, args...)
	defer func() {
		if nil != rows {
			rows.Close() //1、判断是否为空再关闭，2、如果不关闭而数据行并没有被scan的话，连接一直会被占用直到超时断开
		}
	}()

	if nil != sqlErr {
		dbLog.Errorln(sqlErr, sql, args)
		return nil, NewECronError(common.ErrQueryJob, common.InternalError, nil)
	}

	jobList := make([]common.JobInfo, 0)
	job := common.JobInfo{}
	var jobType int
	context := make([]byte, 0)
	var createTime int64
	var updateTime int64

	for rows.Next() {
		job = common.JobInfo{}
		jobType = 0
		context = make([]byte, 0)
		createTime = 0
		updateTime = 0

		sqlErr := rows.Scan(&job.JobID, &job.JobName, &job.JobCronTime, &jobType, &context, &job.Enabled, &job.Remarks, &createTime, &updateTime, &job.TenantID)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, sql)
			return nil, NewECronError(common.ErrScanFieldValue, common.InternalError, nil)
		}

		jsonErr := jsoniter.Unmarshal(context, &job.Context)
		if err, ok := d.unmarshalJSONError("GetJob()", jsonErr); ok {
			return nil, err
		}

		job.JobType, _ = d.dataDict.DJobType.IntToString(jobType)
		job.CreateTime = common.TimeStampToString(createTime)
		job.UpdateTime = common.TimeStampToString(updateTime)
		jobList = append(jobList, job)
	}

	if err := d.checkGetJobInfoAfter(params, jobList); nil != err {
		return nil, err
	}

	return jobList, nil
}

// GetJobTotal 获取任务总数
func (d *ecronDB) GetJobTotal(params common.JobTotalQueryParams, visitor common.Visitor) (common.JobTotal, *common.ECronError) {
	jobTotal := common.JobTotal{
		Total:     0,
		TimeStamp: time.Now().Format(time.RFC3339),
	}

	if err := d.checkGetJobTotalBefore(params); nil != err {
		return jobTotal, err
	}

	sql := dbSQL.GetJobTotalSQL()
	condition, args := d.getJobTotalQueryCondition(params, jobTotal.TimeStamp, visitor)
	sql += condition
	rows, sqlErr := d.db.Query(sql, args...)
	defer func() {
		if nil != rows {
			rows.Close() //1、判断是否为空再关闭，2、如果不关闭而数据行并没有被scan的话，连接一直会被占用直到超时断开
		}
	}()

	if nil != sqlErr {
		dbLog.Errorln(sqlErr, sql, args)
		return jobTotal, NewECronError(common.ErrQueryJobTotal, common.InternalError, nil)
	}

	for rows.Next() {
		sqlErr := rows.Scan(&jobTotal.Total)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, sql)
			return jobTotal, NewECronError(common.ErrScanFieldValue, common.InternalError, nil)
		}
	}

	return jobTotal, nil
}

// GetJobStatus 获取任务状态
func (d *ecronDB) GetJobStatus(params common.JobStatusQueryParams, visitor common.Visitor) ([]common.JobStatus, *common.ECronError) {
	if err := d.checkGetJobStatusBefore(params, visitor); nil != err {
		return nil, err
	}

	sql := dbSQL.GetJobStatusSQL(d.jobStatus.fields[d.jobStatus.fExecuteID:])
	condition, args := d.getJobStatusQueryCondition(params, visitor)
	sql += condition
	rows, sqlErr := d.db.Query(sql, args...)
	defer func() {
		if nil != rows {
			rows.Close() //1、判断是否为空再关闭，2、如果不关闭而数据行并没有被scan的话，连接一直会被占用直到超时断开
		}
	}()

	if nil != sqlErr {
		dbLog.Infof("[GetJobStatus] sql: %v, args: %v, sqlErr: %v", sql, args, sqlErr)
		return nil, NewECronError(common.ErrQueryJobStatus, common.InternalError, nil)
	}

	jobStatusList := make([]common.JobStatus, 0)
	status := common.JobStatus{}

	var jobtype int
	var jobStatus int
	var executor string
	var extinfo string
	var beginTime int64
	var endTime int64

	for rows.Next() {
		status = common.JobStatus{}
		jobtype = 0
		jobStatus = 0
		executor = ""
		extinfo = ""
		beginTime = 0
		endTime = 0

		sqlErr := rows.Scan(&status.ExecuteID, &status.JobID, &jobtype, &status.JobName, &jobStatus, &beginTime, &endTime, &executor, &status.ExecuteTimes, &extinfo)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, sql)
			return nil, NewECronError(common.ErrScanFieldValue, common.InternalError, nil)
		}

		jsonErr := jsoniter.Unmarshal([]byte(executor), &status.Executor)
		if err, ok := d.unmarshalJSONError("GetJobStatus()", jsonErr); ok {
			return nil, err
		}

		jsonErr = jsoniter.Unmarshal([]byte(extinfo), &status.ExtInfo)
		if err, ok := d.unmarshalJSONError("GetJobStatus()", jsonErr); ok {
			return nil, err
		}

		if isDeleted, ok := status.ExtInfo[common.IsDeleted].(float64); ok {
			if int(isDeleted) == 1 {
				continue
			}
		} else {
			dbLog.Infoln("[GetJobStatus] get isDeleted failed")
			continue
		}

		if tenantID, ok := status.ExtInfo[common.TenantID].(string); ok {
			if !visitor.Admin && tenantID != visitor.ClientID {
				continue
			}
		} else {
			dbLog.Infoln("[GetJobStatus] get tenant_id failed")
			continue
		}

		status.JobType, _ = d.dataDict.DJobType.IntToString(jobtype)
		status.JobStatus, _ = d.dataDict.DJobStatus.IntToString(jobStatus)
		status.BeginTime = common.TimeStampToString(beginTime)
		status.EndTime = common.TimeStampToString(endTime)

		jobStatusList = append(jobStatusList, status)
	}

	return jobStatusList, nil
}

// UpdateJobStatus 更新定时任务状态
func (d *ecronDB) UpdateJobStatus(status []common.JobStatus, visitor common.Visitor) ([]string, *common.ECronError) {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return nil, err
	}

	jobStatusMap, err := d.clearJobStatus(status)
	if (*common.ECronError)(nil) != err {
		return nil, err
	}

	failedIDs := make([]string, 0)
	insertSet := make([]string, 0)
	insertArgs := make([]interface{}, 0)
	querySQLStr := dbSQL.GetUpdateJobStatusQuerySQL()
	updateSQLStr := dbSQL.GetUpdateJobStatusUpdateSQL(d.jobStatus.fields[d.jobStatus.fJobID:])
	for executeID, argsList := range jobStatusMap {
		count, err := d.getJobStatusQueryCount(querySQLStr, executeID)
		if err != nil {
			dbLog.Errorln("getJobStatusQueryCount error, querySQLStr: %v, executeID: %v, err: %v", querySQLStr, executeID, err)
			failedIDs = append(failedIDs, executeID)
			continue
		}

		if count > 0 {
			argsList = append(argsList, executeID)
			_, sqlErr := d.db.Exec(updateSQLStr, argsList...)
			if nil != sqlErr {
				sqlLog.Errorln("UpdateJobStatus error, updateSQLStr: %v, argsList: %v, sqlErr: %v", updateSQLStr, argsList, sqlErr)
				failedIDs = append(failedIDs, executeID)
			}
		} else {
			insertSet = append(insertSet, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			insertArgs = append(insertArgs, executeID)
			insertArgs = append(insertArgs, argsList...)
		}
	}

	if len(insertSet) == 0 {
		return failedIDs, nil
	}
	insertSqlStr := dbSQL.GetUpdateJobStatusInsertSQL(d.jobStatus.fields[d.jobStatus.fExecuteID:], insertSet)
	_, sqlErr := d.db.Exec(insertSqlStr, insertArgs...)
	if nil != sqlErr {
		sqlLog.Errorln("UpdateJobStatus error, insertSqlStr: %v, insertArgs: %v, sqlErr: %v", insertSqlStr, insertArgs, sqlErr)
		if len(insertSet) == len(jobStatusMap) {
			return nil, NewECronError(common.ErrDataBaseExecError, common.InternalError, nil)
		}
		return failedIDs, nil
	}

	return failedIDs, nil
}

func (d *ecronDB) getJobStatusQueryCount(querySQLStr string, executeID string) (count int, err *common.ECronError) {
	rows, sqlErr := d.db.Query(querySQLStr, executeID)
	defer func() {
		if nil != rows {
			rows.Close()
		}
	}()

	if nil != sqlErr {
		dbLog.Errorln(sqlErr, querySQLStr, executeID)
		return 0, NewECronError(common.ErrQueryJobTotal, common.InternalError, nil)
	}

	for rows.Next() {
		sqlErr := rows.Scan(&count)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, querySQLStr)
			return 0, NewECronError(common.ErrScanFieldValue, common.InternalError, nil)
		}
	}

	return count, nil
}

// BatchJobEnable 批量启用/禁用定时任务
func (d *ecronDB) BatchJobEnable(jobID []string, enable bool, updateTime string, visitor common.Visitor) *common.ECronError {
	if err := d.checkBatchBefore(jobID, visitor); nil != err {
		return err
	}

	time, err := common.StringToTimeStamp(updateTime)
	if nil != err {
		dbLog.Errorf("BatchJobEnable() %v", err)
	}

	sql, argID := dbSQL.BatchJobEnableSQL([]string{
		d.jobInfo.fields[d.jobInfo.fEnabled],
		d.jobInfo.fields[d.jobInfo.fUpdateTime],
		d.jobInfo.fields[d.jobInfo.fJobID],
		d.jobInfo.fields[d.jobInfo.fTenantID]}, jobID, visitor)

	args := make([]interface{}, 0)
	args = append(args, enable)
	args = append(args, time)
	args = append(args, argID...)

	_, sqlErr := d.db.Exec(sql, args...)
	if nil != sqlErr {
		sqlLog.Errorln(sqlErr, sql, args)
		return NewECronError(common.ErrBatchJobEnable, common.InternalError, nil)
	}

	return nil
}

// BatchJobNotify 批量修改定时任务通知地址
func (d *ecronDB) BatchJobNotify(jobID []string, notify common.JobNotify, updateTime string, visitor common.Visitor) ([]string, *common.ECronError) {
	if err := d.checkBatchBefore(jobID, visitor); nil != err {
		return nil, err
	}

	time, err := common.StringToTimeStamp(updateTime)
	if nil != err {
		dbLog.Errorf("BatchJobNotify() %v", err)
	}

	jobInfoList, err1 := d.getBatchJobNotifyContext(jobID)
	if err1 != nil {
		return nil, err1
	}

	failedIDs := make([]string, 0)
	updateSqlStr := dbSQL.BatchJobNotifyUpdateSQL()
	for _, job := range jobInfoList {
		if !visitor.Admin && job.TenantID != visitor.ClientID {
			continue
		}

		job.Context.Notify.Webhook = notify.Webhook
		context, jsonErr := jsoniter.Marshal(job.Context)
		if err, ok := d.marshalJSONError("BatchJobNotify()", jsonErr); ok {
			dbLog.Errorf("BatchJobNotify() %v", err)
			failedIDs = append(failedIDs, job.JobID)
			continue
		}

		_, sqlErr := d.db.Exec(updateSqlStr, string(context), time, job.JobID)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, updateSqlStr, string(context), job.JobID)
			failedIDs = append(failedIDs, job.JobID)
		}
	}

	return failedIDs, nil
}

func (d *ecronDB) getBatchJobNotifyContext(jobID []string) (jobInfoList []common.JobInfo, err *common.ECronError) {
	querySqlStr, queryArgs := dbSQL.BatchJobNotifyQuerySQL(jobID)
	rows, sqlErr := d.db.Query(querySqlStr, queryArgs...)
	defer func() {
		if nil != rows {
			rows.Close()
		}
	}()

	if nil != sqlErr {
		dbLog.Errorln(sqlErr, querySqlStr, queryArgs)
		return nil, NewECronError(common.ErrQueryJob, common.InternalError, nil)
	}

	var context string
	status := common.JobInfo{}
	jobInfoList = make([]common.JobInfo, 0)

	for rows.Next() {
		sqlErr := rows.Scan(&status.JobID, &status.TenantID, &context)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, querySqlStr)
			return nil, NewECronError(common.ErrScanFieldValue, common.InternalError, nil)
		}

		jsonErr := jsoniter.Unmarshal([]byte(context), &status.Context)
		if err, ok := d.unmarshalJSONError("getBatchJobNotifyContext()", jsonErr); ok {
			return nil, err
		}

		jobInfoList = append(jobInfoList, status)
	}

	return jobInfoList, nil
}

func (d *ecronDB) checkInsertJobBefore(job common.JobInfo) (err *common.ECronError, context []byte, createTime int64, updateTime int64, jobType int) {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return err, nil, 0, 0, 0
	}

	if err := d.checkInject(map[string]string{
		"job_id":        job.JobID,
		"job_name":      job.JobName,
		"job_cron_time": job.JobCronTime,
		"remarks":       job.Remarks}, "InsertJob()"); (*common.ECronError)(nil) != err {
		return err, nil, 0, 0, 0
	}

	if 0 == len(job.JobName) {
		return NewECronError(common.ErrJobNameEmpty, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: "job_name",
			},
		}), nil, 0, 0, 0
	}

	if err := d.checkTimeInterval(job.Context.BeginTime, job.Context.EndTime); (*common.ECronError)(nil) != err {
		return err, nil, 0, 0, 0
	}

	if err := d.CheckJobExecuteMode(job.Context.Mode); (*common.ECronError)(nil) != err {
		return err, nil, 0, 0, 0
	}

	if err := d.checkCronTime(job.JobCronTime); (*common.ECronError)(nil) != err {
		return err, nil, 0, 0, 0
	}

	context, jsonErr := jsoniter.Marshal(job.Context)
	if err, ok := d.marshalJSONError("InsertJob()", jsonErr); ok {
		return err, context, 0, 0, 0
	}

	createTime, tsErr := common.StringToTimeStamp(job.CreateTime)
	if nil != tsErr {
		dbLog.Errorf("InsertJob() %v", tsErr)
	}

	updateTime, tsErr = common.StringToTimeStamp(job.UpdateTime)
	if nil != tsErr {
		dbLog.Errorf("InsertJob() %v", tsErr)
	}

	jt, err := d.checkJobType(job.JobType)
	if (*common.ECronError)(nil) != err {
		return err, nil, 0, 0, 0
	}

	return nil, context, createTime, updateTime, jt
}

func (d *ecronDB) checkInsertOrUpdateJobAfter(job common.JobInfo, insertErr *common.ECronError, visitor common.Visitor) *common.ECronError {
	if (*common.ECronError)(nil) == insertErr {
		return nil
	}

	//判断一下出错原因是否是键值冲突，特别提醒：任务名称唯一且不能为空
	if err := d.checkJobName(job.JobName, visitor); (*common.ECronError)(nil) != err {
		return err
	}

	return insertErr
}

func (d *ecronDB) checkUpdateJobBefore(job common.JobInfo, visitor common.Visitor) (err *common.ECronError, context []byte, updateTime int64, jobType int) {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return err, nil, 0, 0
	}

	if err := d.checkInject(map[string]string{
		"job_id":        job.JobID,
		"job_name":      job.JobName,
		"job_cron_time": job.JobCronTime,
		"remarks":       job.Remarks}, "UpdateJob()"); (*common.ECronError)(nil) != err {
		return err, nil, 0, 0
	}

	if 0 == len(job.JobName) {
		return NewECronError(common.ErrJobNameEmpty, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: "job_name",
			},
		}), nil, 0, 0
	}

	if err := d.checkTimeInterval(job.Context.BeginTime, job.Context.EndTime); (*common.ECronError)(nil) != err {
		return err, nil, 0, 0
	}

	if err := d.CheckJobExecuteMode(job.Context.Mode); nil != err {
		return err, nil, 0, 0
	}

	if err := d.checkCronTime(job.JobCronTime); (*common.ECronError)(nil) != err {
		return err, nil, 0, 0
	}

	if err := d.checkSingleJobIDExist(job.JobID, visitor); (*common.ECronError)(nil) != err {
		return err, nil, 0, 0
	}

	context, jsonErr := jsoniter.Marshal(job.Context)
	if err, ok := d.marshalJSONError("UpdateJob()", jsonErr); ok {
		return err, nil, 0, 0
	}

	updateTime, tsErr := common.StringToTimeStamp(job.UpdateTime)
	if nil != tsErr {
		dbLog.Errorf("UpdateJob() %v", tsErr)
	}

	jt, err := d.checkJobType(job.JobType)
	if (*common.ECronError)(nil) != err {
		return err, nil, 0, 0
	}

	return nil, context, updateTime, jt
}

func (d *ecronDB) checkGetJobStatusBefore(params common.JobStatusQueryParams, visitor common.Visitor) *common.ECronError {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return err
	}

	if err := d.checkInject(map[string]string{"job_id": params.JobID}, "GetJobStatus()"); (*common.ECronError)(nil) != err {
		return err
	}

	if dbSQL.IsEmpty(params.JobID) && dbSQL.IsEmpty(params.JobType) && dbSQL.IsEmpty(params.JobStatus) && dbSQL.IsEmpty(params.BeginTime) && dbSQL.IsEmpty(params.EndTime) {
		return NewECronError(common.ErrQueryParameterIsNull, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: "job_id",
				1: "job_type",
				2: "job_status",
				3: "begin_at",
				4: "end_at",
			},
		})
	}

	if err := d.checkTimeInterval(params.BeginTime, params.EndTime); nil != err {
		return err
	}

	//如果job_id不存在，抛出异常
	if len(params.JobID) > 0 {
		if err := d.checkSingleJobIDExist(params.JobID, visitor); (*common.ECronError)(nil) != err {
			dbLog.Errorf("[checkSingleJobIDExist] params.JobID: %v, ERROR: %v", params.JobID, err)
			return err
		}
	}

	if len(params.JobType) > 0 {
		if _, err := d.checkJobType(params.JobType); (*common.ECronError)(nil) != err {
			dbLog.Errorf("[checkJobType] ERROR: %v", err)
			return err
		}
	}

	if len(params.JobStatus) > 0 {
		if _, err := d.checkJobStatus(params.JobStatus); (*common.ECronError)(nil) != err {
			dbLog.Errorf("[checkJobStatus] ERROR: %v", err)
			return err
		}
	}

	return nil
}

func (d *ecronDB) checkGetJobTotalBefore(params common.JobTotalQueryParams) *common.ECronError {
	if err := d.checkTimeInterval(params.BeginTime, params.EndTime); nil != err {
		return err
	}

	if err, ok := d.isDataBaseAvailable(); !ok {
		return err
	}

	return nil
}

func (d *ecronDB) checkGetJobInfoBefore(params common.JobInfoQueryParams) *common.ECronError {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return err
	}

	if err := d.checkInject(map[string]string{"job_id": strings.Join(params.JobID, ",")}, "GetJob()"); (*common.ECronError)(nil) != err {
		return err
	}

	if 0 == len(params.JobID) {
		//只根据job id查找，无视其他条件，更不会判断
		if params.Page < 1 || params.Limit < 0 {
			return NewECronError(common.ErrLimitOrPageIllegal, common.BadRequest, map[string]interface{}{
				common.DetailParameters: []string{
					0: "page",
					1: "limit",
				},
			})
		}

		if err := d.checkTime("timestamp", params.TimeStamp); (*common.ECronError)(nil) != err {
			return err
		}

		if len(params.JobType) > 0 {
			if _, err := d.checkJobType(params.JobType); (*common.ECronError)(nil) != err {
				return err
			}
		}
	}

	return nil
}

func (d *ecronDB) checkGetJobInfoAfter(params common.JobInfoQueryParams, jobList []common.JobInfo) *common.ECronError {
	if len(params.JobID) > 0 {
		tmpID := make([]string, 0)
		for _, v1 := range params.JobID {
			bExist := false
			for _, v2 := range jobList {
				if v1 == v2.JobID {
					bExist = true
					continue
				}
			}
			if !bExist {
				tmpID = append(tmpID, v1)
			}
		}

		if len(tmpID) > 0 {
			return NewECronError(common.ErrJobNotExist, common.NotFound, map[string]interface{}{
				common.DetailConflicts: []string{
					0: "job_id",
				},
			})
		}
	}

	return nil
}

func (d *ecronDB) checkBatchBefore(jobID []string, visitor common.Visitor) *common.ECronError {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return err
	}

	if err := d.checkInject(map[string]string{"job_id": strings.Join(jobID, ",")}, "Batch()"); (*common.ECronError)(nil) != err {
		return err
	}

	//如果有不存在的任务ID，抛出异常
	notExistID := make([]string, 0)
	if err := d.checkJobIDExist(jobID, &notExistID, visitor); nil != err {
		return err
	}

	if len(notExistID) > 0 {
		return NewECronError(common.ErrJobNotExist, common.NotFound, map[string]interface{}{
			common.DetailConflicts: []string{
				0: "job_id",
			},
		})
	}

	return nil
}

func (d *ecronDB) checkTimeInterval(beginTime string, endTime string) *common.ECronError {
	if err := d.checkTime("begin_at", beginTime); nil != err {
		return err
	}

	if err := d.checkTime("end_at", endTime); nil != err {
		return err
	}

	begin, _ := common.StringToTimeStamp(beginTime)
	end, _ := common.StringToTimeStamp(endTime)
	if len(beginTime) > 0 && len(endTime) > 0 && begin > end {
		return NewECronError(common.ErrBeginTimeGreaterThanEndTime, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: "begin_at",
				1: "end_at",
			},
		})
	}

	return nil
}

func (d *ecronDB) checkTime(key string, value string) *common.ECronError {
	v, _ := common.StringToTimeStamp(value)
	if len(value) > 0 && 0 == v {
		return NewECronError(common.ErrTimeIllegal, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: key,
			},
		})
	}

	return nil
}

func (d *ecronDB) checkCronTime(cronTime string) *common.ECronError {
	_, err := d.parser.Parse(cronTime)
	if nil != err {
		return NewECronError(common.ErrCronTime, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: "job_cron_time",
			},
		})
	}
	return nil
}

// CheckJobExecuteMode 检查任务执行方式
func (d *ecronDB) CheckJobExecuteMode(executeMode string) *common.ECronError {
	if _, ok := d.dataDict.DJobExecution.StringToInt(executeMode); !ok {
		return NewECronError(common.ErrUnsupportedExecutionMode, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: "mode",
			},
		})
	}

	return nil
}

func (d *ecronDB) Upgrade() {
	upgradeSQL := []string{
		"drop index index_job_type on t_cron_job",
		"drop index index_create_time on t_cron_job",
		"drop index index_update_time on t_cron_job",
		"drop index index_is_deleted on t_cron_job_status",
		"drop index index_job_type on t_cron_job_status",
		"drop index index_end_time on t_cron_job_status",
		"alter table t_cron_job add index index_time(`f_create_time`, `f_update_time`)",
	}

	for _, v := range upgradeSQL {
		_, _ = d.db.Exec(v)
	}
}

func (d *ecronDB) checkJobType(jobType string) (int, *common.ECronError) {
	v, ok := d.dataDict.DJobType.StringToInt(jobType)
	if !ok {
		return 0, NewECronError(common.ErrJobTypeIllegal, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: "job_type",
			},
		})
	}

	return v, nil
}

func (d *ecronDB) checkJobStatus(jobStatus string) (int, *common.ECronError) {
	v, ok := d.dataDict.DJobStatus.StringToInt(jobStatus)
	if !ok {
		return 0, NewECronError(common.ErrJobStatusIllegal, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: "job_status",
			},
		})
	}

	return v, nil
}

func (d *ecronDB) checkJobName(jobName string, visitor common.Visitor) *common.ECronError {

	if 0 == len(jobName) {
		return NewECronError(common.ErrJobNameEmpty, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: "job_name",
			},
		})
	}

	if err, ok := d.isDataBaseAvailable(); !ok {
		return err
	}

	sql := dbSQL.CheckJobNameSQL([]string{d.jobInfo.fields[d.jobInfo.fJobName], d.jobInfo.fields[d.jobInfo.fTenantID]})
	args := []interface{}{0: jobName, 1: visitor.ClientID}
	rows, sqlErr := d.db.Query(sql, args...)
	defer func() {
		if nil != rows {
			rows.Close() //1、判断是否为空再关闭，2、如果不关闭而数据行并没有被scan的话，连接一直会被占用直到超时断开
		}
	}()

	if nil != sqlErr {
		dbLog.Errorln(sqlErr)
		dbLog.Errorln(sql)
		return NewECronError(common.ErrQueryJobName, common.InternalError, nil)
	}

	count := 0
	for rows.Next() {
		count = 0
		sqlErr := rows.Scan(&count)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, sql)
			return NewECronError(common.ErrScanFieldValue, common.InternalError, nil)
		}
	}

	if count > 0 {
		return NewECronError(common.ErrJobNameExists, common.Conflict, map[string]interface{}{
			common.DetailConflicts: []string{
				0: "job_name",
			},
		})
	}

	return nil
}

func (d *ecronDB) checkJobIDExist(jobID []string, notExistID *[]string, visitor common.Visitor) *common.ECronError {
	partID := make([]string, 0)
	for i, v := range jobID {
		if i > 0 && 0 == i%10 {
			partID = append(partID, v)

			err := d.checkJobPartIDExist(partID, notExistID, visitor)
			if nil != err {
				return err
			}

			partID = partID[0:0]
		} else {
			partID = append(partID, v)
		}
	}

	return d.checkJobPartIDExist(partID, notExistID, visitor)
}

func (d *ecronDB) checkJobPartIDExist(partID []string, notExistID *[]string, visitor common.Visitor) *common.ECronError {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return err
	}

	sql, args := dbSQL.CheckJobIDExistSQL([]string{
		d.jobInfo.fields[d.jobInfo.fJobID],
		d.jobInfo.fields[d.jobInfo.fJobID],
		d.jobInfo.fields[d.jobInfo.fTenantID]}, partID, visitor)
	rows, sqlErr := d.db.Query(sql, args...)
	defer func() {
		if nil != rows {
			rows.Close() //1、判断是否为空再关闭，2、如果不关闭而数据行并没有被scan的话，连接一直会被占用直到超时断开
		}
	}()

	if nil != sqlErr {
		dbLog.Errorln(sqlErr, sql, args)
		return NewECronError(common.ErrQueryJobID, common.InternalError, nil)
	}

	existID := make(map[string]string)
	for rows.Next() {
		id := ""
		sqlErr := rows.Scan(&id)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, sql)
			return NewECronError(common.ErrScanFieldValue, common.InternalError, nil)
		}

		existID[id] = id
	}

	for _, v := range partID {
		if _, ok := existID[v]; ok {
			continue
		}
		*notExistID = append(*notExistID, v)
	}
	return nil
}

func (d *ecronDB) checkSingleJobIDExist(jobID string, visitor common.Visitor) *common.ECronError {
	if err, ok := d.isDataBaseAvailable(); !ok {
		return err
	}

	if 0 == len(jobID) {
		return NewECronError(common.ErrJobNotExist, common.NotFound, map[string]interface{}{
			common.DetailConflicts: []string{
				0: "job_id",
			},
		})
	}

	sql, args := dbSQL.CheckJobIDExistSQL([]string{
		d.jobInfo.fields[d.jobInfo.fJobID],
		d.jobInfo.fields[d.jobInfo.fJobID],
		d.jobInfo.fields[d.jobInfo.fTenantID]}, []string{jobID}, visitor)
	rows, sqlErr := d.db.Query(sql, args...)
	defer func() {
		if nil != rows {
			rows.Close() //1、判断是否为空再关闭，2、如果不关闭而数据行并没有被scan的话，连接一直会被占用直到超时断开
		}
	}()

	if nil != sqlErr {
		return NewECronError(common.ErrQueryJobID, common.InternalError, nil)
	}

	id := ""
	for rows.Next() {
		id = ""
		sqlErr := rows.Scan(&id)
		if nil != sqlErr {
			dbLog.Errorln(sqlErr, sql)
			return NewECronError(common.ErrScanFieldValue, common.InternalError, nil)
		}
	}

	if 0 == len(id) {
		return NewECronError(common.ErrJobNotExist, common.NotFound, map[string]interface{}{
			common.DetailConflicts: []string{
				0: "job_id",
			},
		})
	}

	return nil
}

func (d *ecronDB) checkInject(params map[string]string, funcName string) *common.ECronError {
	if k, b := dbSQL.IsParameterInject(params); b {
		sqlLog.Errorln("inject parameter", funcName, k)
		return NewECronError(common.ErrInvalidParameter, common.BadRequest, map[string]interface{}{
			common.DetailParameters: []string{
				0: k,
			},
		})
	}
	return nil
}

func (d *ecronDB) clearJobStatus(src []common.JobStatus) (jobStatusMap map[string][]interface{}, err *common.ECronError) {
	count := 0
	jobStatusMap = make(map[string][]interface{})
	for _, v := range src {
		if err := d.checkInject(map[string]string{
			"execute_id": v.ExecuteID,
			"job_id":     v.JobID,
			"job_name":   v.JobName}, "UpdateJobStatus()"); (*common.ECronError)(nil) != err {
			return nil, err
		}

		executor, jsonErr := jsoniter.Marshal(v.Executor)
		if err, ok := d.marshalJSONError("UpdateJobStatus()", jsonErr); ok {
			return nil, err
		}

		extInfo, jsonErr := jsoniter.Marshal(v.ExtInfo)
		if err, ok := d.marshalJSONError("UpdateJobStatus()", jsonErr); ok {
			return nil, err
		}

		beginTime, err := common.StringToTimeStamp(v.BeginTime)
		if nil != err {
			dbLog.Errorln(v)
			return nil, NewECronError(common.ErrTimeIllegal, common.InternalError, nil)
		}

		endTime, err := common.StringToTimeStamp(v.EndTime)
		if nil != err {
			dbLog.Errorln(v)
			return nil, NewECronError(common.ErrTimeIllegal, common.InternalError, nil)
		}

		//不符合的任务类型和状态，不入库
		jobType, jtErr := d.checkJobType(v.JobType)
		if (*common.ECronError)(nil) != jtErr {
			dbLog.Errorln(v)
			continue
		}
		jobStatus, jsErr := d.checkJobStatus(v.JobStatus)
		if (*common.ECronError)(nil) != jsErr {
			dbLog.Errorln(v)
			continue
		}

		args := make([]interface{}, 0)
		args = append(args, v.JobID)
		args = append(args, jobType)
		args = append(args, v.JobName)
		args = append(args, jobStatus)
		args = append(args, beginTime)
		args = append(args, endTime)
		args = append(args, string(executor))
		args = append(args, v.ExecuteTimes)
		args = append(args, string(extInfo))

		jobStatusMap[v.ExecuteID] = args

		count++
	}

	if 0 == count {
		return nil, NewECronError(common.ErrStatusEmpty, common.InternalError, nil)
	}

	return
}

func (d *ecronDB) getJobInfoQueryCondition(params common.JobInfoQueryParams, visitor common.Visitor) (condition string, args []interface{}) {
	dbSQL.AddConditionSQL(&condition, d.jobInfo.fields[d.jobInfo.fJobID], "in", params.JobID, &args)
	if !dbSQL.IsEmpty(condition) {
		//如果用户填写了ID，则无视其他条件
		if !visitor.Admin {
			dbSQL.AddConditionSQL(&condition, d.jobInfo.fields[d.jobInfo.fTenantID], "=", visitor.ClientID, &args)
		}
		return dbSQL.AddWhereSQL(&condition, &args)
	}

	jobType, _ := d.dataDict.DJobType.StringToInt(params.JobType)
	stamp, _ := common.StringToTimeStamp(params.TimeStamp)
	dbSQL.AddConditionSQL(&condition, d.jobInfo.fields[d.jobInfo.fJobType], "=", jobType, &args)
	dbSQL.AddConditionSQL(&condition, d.jobInfo.fields[d.jobInfo.fCreateTime], "<=", stamp, &args)
	if !visitor.Admin {
		dbSQL.AddConditionSQL(&condition, d.jobInfo.fields[d.jobInfo.fTenantID], "=", visitor.ClientID, &args)
	}
	dbSQL.AddWhereSQL(&condition, &args) //如果条件为空，order by是不需要where关键字的
	dbSQL.AddOrderBySQL(&condition, map[string]string{d.jobInfo.fields[d.jobInfo.fCreateTime]: "desc"})
	dbSQL.AddLimitSQL(&condition, params.Limit, params.Page, &args)
	return
}

func (d *ecronDB) getJobTotalQueryCondition(params common.JobTotalQueryParams, timeStamp string, visitor common.Visitor) (condition string, args []interface{}) {
	beginTime, _ := common.StringToTimeStamp(params.BeginTime)
	endTime, _ := common.StringToTimeStamp(params.EndTime)
	stamp, _ := common.StringToTimeStamp(timeStamp)

	dbSQL.AddConditionSQL(&condition, d.jobInfo.fields[d.jobInfo.fCreateTime], ">=", beginTime, &args)
	dbSQL.AddConditionSQL(&condition, d.jobInfo.fields[d.jobInfo.fCreateTime], "<=", endTime, &args)
	dbSQL.AddConditionSQL(&condition, d.jobInfo.fields[d.jobInfo.fCreateTime], "<=", stamp, &args)
	if !visitor.Admin {
		dbSQL.AddConditionSQL(&condition, d.jobInfo.fields[d.jobInfo.fTenantID], "=", visitor.ClientID, &args)
	}
	return dbSQL.AddWhereSQL(&condition, &args)
}

func (d *ecronDB) getJobStatusQueryCondition(params common.JobStatusQueryParams, visitor common.Visitor) (condition string, args []interface{}) {
	beginTime, _ := common.StringToTimeStamp(params.BeginTime)
	endTime, _ := common.StringToTimeStamp(params.EndTime)
	jobType, _ := d.dataDict.DJobType.StringToInt(params.JobType)
	jobStatus, _ := d.dataDict.DJobStatus.StringToInt(params.JobStatus)

	dbSQL.AddConditionSQL(&condition, d.jobStatus.fields[d.jobStatus.fJobID], "=", params.JobID, &args)
	dbSQL.AddConditionSQL(&condition, d.jobStatus.fields[d.jobStatus.fBeginTime], ">=", beginTime, &args)
	dbSQL.AddConditionSQL(&condition, d.jobStatus.fields[d.jobStatus.fBeginTime], "<=", endTime, &args)
	dbSQL.AddConditionSQL(&condition, d.jobStatus.fields[d.jobStatus.fJobType], "=", jobType, &args)
	dbSQL.AddConditionSQL(&condition, d.jobStatus.fields[d.jobStatus.fJobStatus], "=", jobStatus, &args)

	dbSQL.AddWhereSQL(&condition, &args)
	dbSQL.AddOrderBySQL(&condition, map[string]string{d.jobStatus.fields[d.jobStatus.fBeginTime]: "desc"})

	// 如果只传入jobID，便只给出该任务的最近一次执行情况
	if !dbSQL.IsEmpty(params.JobID) && dbSQL.IsEmpty(params.JobType) && dbSQL.IsEmpty(params.JobStatus) && dbSQL.IsEmpty(params.BeginTime) && dbSQL.IsEmpty(params.EndTime) {
		dbSQL.AddLimitSQL(&condition, 1, 1, &args)
	}

	return
}

// MarshalJSONError 对象转换成json错误获取函数
func (d *ecronDB) marshalJSONError(funcName string, jsonErr error) (*common.ECronError, bool) {
	if nil != jsonErr {
		dbLog.Errorln(funcName, jsonErr)
		return NewECronError(fmt.Sprintf("ecronDB.%v %v", funcName, common.ErrMarshalJSON), common.InternalError, nil), true
	}

	return nil, false
}

// UnmarshalJSONError json转换成对象错误获取函数
func (d *ecronDB) unmarshalJSONError(funcName string, jsonErr error) (*common.ECronError, bool) {
	if nil != jsonErr {
		dbLog.Errorln(funcName, jsonErr)
		return NewECronError(fmt.Sprintf("ecronDB.%v %v", funcName, common.ErrUnMarshalJSON), common.InternalError, nil), true
	}

	return nil, false
}
