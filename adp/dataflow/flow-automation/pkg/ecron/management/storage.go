package management

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils/rest"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	uuid "github.com/satori/go.uuid"
)

var (
	storageLog    = utils.NewLogger()
	storageConfig = utils.NewConfiger()
	serviceID     = storageConfig.Config().ManagementServiceID
	cacheHandler  = common.NewCacheConfig()

	cronSvr = common.ServerInfo{
		ServiceID: serviceID,
		Addr:      storageConfig.Config().CronAddr,
		Port:      storageConfig.Config().CronPort,
		SSLOn:     storageConfig.Config().SSLOn,
		CertFile:  storageConfig.Config().SSLCertFile,
		KeyFile:   storageConfig.Config().SSLKeyFile,
		MultiNode: storageConfig.Config().MultiNode,
	}

	webhook = storageConfig.Config().Webhook

	dbClientPingSleep = common.GetSleepDuration(storageConfig.Config().DBClientPingSleep)

	nsqConnected = true //定时服务监测NSQ标志
)

//go:generate mockgen -package mock -source ../management/storage.go -destination ../mock/mock_storage.go

// ManagementService 定时服务接口
type ManagementService interface {

	//开启服务
	Start()

	//关闭服务
	Stop()
}

// NewManagementService 创建定时服务对象
func NewManagementService() ManagementService {
	return &management{
		mapRequest:        make(map[string][]map[string]func(c *gin.Context)),
		msmqClient:        nil,
		dbClient:          nil,
		executor:          nil,
		authClient:        nil,
		httpServerStarter: utils.NewHTTPServerStarter(),
		chJobMsg:          make(chan common.JobMsg, 1),
		chJobStatus:       make(chan common.JobStatus, 1),
		chJobImmediate:    make(chan common.JobInfo, 1),
	}
}

type management struct {
	mapRequest        map[string][]map[string]func(c *gin.Context)
	msmqClient        utils.MsmqClient
	dbClient          utils.DBClient
	executor          Executor
	authClient        utils.OAuthClient
	httpServerStarter utils.HTTPServerStarter
	chJobMsg          chan common.JobMsg
	chJobStatus       chan common.JobStatus
	chJobImmediate    chan common.JobInfo
}

func (t *management) Start() {
	if cronSvr.MultiNode {
		port, err := t.getFreePort()
		if nil == err {
			cronSvr.Port = port
		} else {
			storageLog.Errorf(fmt.Sprintf("get free port failed, %v", err))
		}
	}

	storageLog.Infof("start management service, id = %v, version = %v", cronSvr.ServiceID, common.Version)

	t.init()
	gin.SetMode(gin.ReleaseMode)
	err := t.httpServerStarter.Start(cronSvr, t.mapRequest)
	if nil != err {
		storageLog.Errorln(err)
	}
}

func (t *management) Stop() {
	if nil != t.dbClient {
		t.dbClient.Release()
	}

	if nil != t.authClient {
		t.authClient.Release()
	}
}

// 定时服务API路径键值
var (
	pathJobID     = "job_id"
	pathExecuteID = "execute_id"
)

// 定时服务开放API路径
var (
	managementPrefix    = "/api/ecron-management/v1"
	jobTotalPATH        = fmt.Sprintf("%s/jobtotal", managementPrefix)
	jobInfoPATH         = fmt.Sprintf("%s/job", managementPrefix)
	jobInfoWithIDPATH   = fmt.Sprintf("%s/:%s", jobInfoPATH, pathJobID)
	jobStatusPATH       = fmt.Sprintf("%s/jobstatus", managementPrefix)
	jobStatusWithIDPATH = fmt.Sprintf("%s/:%s", jobStatusPATH, pathExecuteID)
	jobExecutionPATH    = fmt.Sprintf("%s/jobexecution", managementPrefix)
	jobEnablePATH       = fmt.Sprintf("%s/:%s/enable", jobInfoPATH, pathJobID)
	jobNotifyPATH       = fmt.Sprintf("%s/:%s/notify", jobInfoPATH, pathJobID)
	webhookWithIDPATH   = fmt.Sprintf("%s/:%s", webhook, pathExecuteID)
	healthPATH          = "/health/ready"
	alivePATH           = "/health/alive"
)

func (t *management) getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if nil != err {
		return 0, err
	}

	listen, err := net.ListenTCP("tcp", addr)
	if nil != err {
		return 0, err
	}
	defer listen.Close()
	return listen.Addr().(*net.TCPAddr).Port, nil
}

func (t *management) init() {
	t.initRequest()
	t.initMsmqClient()
	t.initDBClient()
	t.initExecutor()
	t.initAuthClient()
	t.initHydraConfig()
	go func() {
		for {
			t.keepDBAlive()
			time.Sleep(initSleep)
		}
	}()
}

var (
	initSleep = time.Second * 5
)

func (t *management) keepDBAlive() {
	if (*common.ECronError)(nil) != t.dbClient.Ping() {
		t.dbClient.Connect()
	}
}

func (t *management) initHydraConfig() {
	for {
		realVerifyTokenPath, success := t.authClient.VerifyHydraVersion()
		if !success {
			time.Sleep(initSleep)
		} else {
			tmpHydraConfig := common.HydraConfig{
				VerifyTokenPath: realVerifyTokenPath,
			}
			cacheHandler.SetHydraConfig(tmpHydraConfig)
			storageLog.Infof("[verifyHydraVersion] verifyHydraVersion success, realVerifyTokenPath: %v", realVerifyTokenPath)
			break
		}
	}
}

func (t *management) initRequest() {
	t.mapRequest = map[string][]map[string]func(c *gin.Context){
		"GET": {
			0: {jobTotalPATH: t.getJobTotal},
			1: {jobInfoPATH: t.getJobInfo},
			2: {jobStatusPATH: t.getJobStatus},
			3: {healthPATH: t.getHealth},
			4: {alivePATH: t.getAlive},
		},
		"POST": {
			0: {jobInfoPATH: t.postJobInfo},
			1: {jobExecutionPATH: t.postJobExecution},
			2: {webhookWithIDPATH: t.postWebhook},
		},
		"PUT": {
			0: {jobInfoWithIDPATH: t.putJobInfo},
			1: {jobStatusWithIDPATH: t.putJobStatus},
			2: {jobEnablePATH: t.putJobEnable},
			3: {jobNotifyPATH: t.putJobNotify},
		},
		"DELETE": {
			0: {jobInfoWithIDPATH: t.deleteJobInfo},
		},
	}
}

func (t *management) initMsmqClient() {
	//创建消息队列topic和channel
	if nil == t.msmqClient {
		t.msmqClient = utils.NewMsmqClient()
	}
	if nil == t.msmqClient {
		storageLog.Fatalln(common.ErrMSMQClientUnavailable)
		return
	}

	//发布消息
	go func() {
		for {
			select {
			case msg := <-t.chJobMsg:
				go t.publishJobMsg(msg)
			case msg := <-t.chJobImmediate:
				go t.publishJobImmediate(msg)
			case msg := <-t.chJobStatus:
				go t.publishJobStatus(msg)
			}
		}
	}()
}

func (t *management) publishJobMsg(msg common.JobMsg) {
	storageLog.Infof("get cron job message [%v,%v,%v,%v]", msg.Data.JobID, msg.Data.JobName, msg.Data.JobType, msg.Method)
	data, err := jsoniter.Marshal(msg)
	if err != nil {
		storageLog.Infof("jsoniter.Marshal JobMsg err: %v", err)
		return
	}
	storageLog.Infof("publish cron job message [%v] %v", msg.Data.JobID, t.msmqClient.Publish(common.TopicCronJob, data))
}

func (t *management) publishJobImmediate(msg common.JobInfo) {
	storageLog.Infof("get immediate job message [%v,%v,%v,%v]", msg.JobID, msg.JobName, msg.JobType, msg.Context.ExecuteID)
	data, err := jsoniter.Marshal(msg)
	if err != nil {
		storageLog.Infof("jsoniter.Marshal JobInfo err: %v", err)
		return
	}
	storageLog.Infof("publish immediate job message [%v] %v", msg.Context.ExecuteID, t.msmqClient.Publish(common.TopicImmediateJob, data))
}

func (t *management) publishJobStatus(msg common.JobStatus) {
	storageLog.Infof("get job status message[%v,%v]", msg.ExecuteID, msg.JobStatus)
	data, err := jsoniter.Marshal(msg)
	if err != nil {
		storageLog.Infof("jsoniter.Marshal JobStatus err: %v", err)
		return
	}
	storageLog.Infof("publish job status message [%v] %v", msg.ExecuteID, t.msmqClient.Publish(common.TopicJobStatus, data))
}

func (t *management) initDBClient() {
	//加载数据库操作句柄
	if nil == t.dbClient {
		t.dbClient = utils.NewDBClient()
	}
	if nil == t.dbClient {
		storageLog.Fatalln("[initDBClient] NewDBClient error")
		return
	}

	if (*common.ECronError)(nil) == t.dbClient.Connect() {
		t.dbClient.Upgrade()
		storageLog.Infoln("[initDBClient] ", common.InfoDataBaseConnected)
	} else {
		storageLog.Errorln("[initDBClient] ", common.ErrDataBaseDisconnected)
	}
}

func (t *management) initExecutor() {
	if nil == t.executor {
		t.executor = NewExecutor()
	}
	if nil == t.executor {
		storageLog.Fatalln(common.ErrExecutorUnavailable)
	}
}

func (t *management) initAuthClient() {
	if nil == t.authClient {
		t.authClient = utils.NewOAuthClient()
	}

	if nil == t.authClient {
		storageLog.Fatalln(common.ErrAuthClientUnavailable)
	}
}

func (t *management) isEcronDBAvailable() *common.ECronError {
	if nil == t.dbClient {
		return utils.NewECronError(common.ErrDataBaseUnavailable, common.InternalError, nil)
	}

	return nil
}

func (t *management) request(c *gin.Context, obj interface{}) *common.ECronError {
	if (*gin.Context)(nil) == c {
		return utils.NewECronError(common.ErrGinContextUnavailable, common.InternalError, nil)
	}
	if err := rest.GetJSONValue(c, &obj); err != nil {
		storageLog.Errorln(err)
		return utils.NewECronError(fmt.Sprintf("%v", err), common.BadRequest, nil)
	}
	return nil
}

func (t *management) response(src interface{}, err *common.ECronError, bCreate bool, c *gin.Context) {
	respBody := make([]byte, 0)
	if (*common.ECronError)(nil) != err {
		respBody, _ = jsoniter.Marshal(err)
	} else if nil != src {
		respBody, _ = jsoniter.Marshal(src)
	}

	c.Writer.Header().Set(common.ContentType, common.ApplicationJSON)
	c.String(t.code(err, bCreate), string(respBody))
}

func (t *management) code(err *common.ECronError, bCreate bool) int {
	if (*common.ECronError)(nil) == err {
		if bCreate {
			return http.StatusCreated
		}
		return http.StatusOK
	}

	code := strconv.Itoa(err.Code)
	if len(code) >= 3 {
		res, atoiErr := strconv.Atoi(code[0:3])
		if nil != atoiErr || 0 == res {
			return http.StatusInternalServerError
		}
		return res
	}

	return http.StatusInternalServerError
}

func (t *management) auth(token, secret, code string) (visitor common.Visitor, ecronErr *common.ECronError) {
	if nil == t.authClient {
		return common.Visitor{}, utils.NewECronError(common.ErrAuthClientUnavailable, common.InternalError, map[string]interface{}{
			common.DetailParameters: common.Authorization,
		})
	}

	// 优先校验是否是管理员
	visitor.Admin, ecronErr = t.authClient.VerifyCode(secret, code)
	if !visitor.Admin {
		visitor, ecronErr = t.authClient.VerifyToken(token)
	}

	return
}

func (t *management) addWebhook(headers map[string]string, executeID string) map[string]string {
	if nil == headers {
		headers = make(map[string]string)
	}

	headers["webhook"] = fmt.Sprintf("%v/%v", webhook, executeID)
	return headers
}

func (t *management) check(token, secret, code string) (visitor common.Visitor, ecronErr *common.ECronError) {
	if ecronErr = t.isEcronDBAvailable(); (*common.ECronError)(nil) != ecronErr {
		return
	}

	return t.auth(token, secret, code)
}

func (t *management) executeJob(job common.JobInfo) {
	status := common.JobStatus{
		ExecuteID: job.Context.ExecuteID,
		JobStatus: common.SUCCESS,
	}

	//如果是http方式，增加鉴权信息和动态webhook地址
	switch job.Context.Mode {
	case common.HTTP:
		fallthrough
	case common.HTTPS:
		job.Context.Info.Headers = t.addWebhook(job.Context.Info.Headers, job.Context.ExecuteID)
	default:
		break
	}

	reply, err := t.executor.ExecuteJob(job)
	storageLog.Infof("execute a job[%v,%v,%v,%v,%v] reply: %v, result: %v", job.JobID, job.JobName, job.JobCronTime, job.JobType, job.Context.ExecuteID, reply, err)
	if nil != err {
		//如果失败，设置任务状态和任务类型
		//系统启动重试机制
		status.JobStatus = common.FAILURE

		go func() {
			immediateJob := common.JobInfo{
				JobType: common.IMMEDIATE,
				JobID:   job.JobID,
				Context: common.JobContext{
					ExecuteID: job.Context.ExecuteID,
				},
			}
			t.chJobImmediate <- immediateJob
		}()
	} else {
		status.EndTime = time.Now().Format(time.RFC3339)
	}

	go func() {
		if reply {
			t.chJobStatus <- status
		}
	}()
}

func (t *management) handleWebhook(executeID string, result map[string]interface{}) {
	if len(executeID) > 0 {
		status := common.JobStatus{
			ExecuteID: executeID,
			JobStatus: common.SUCCESS,
			EndTime:   time.Now().Format(time.RFC3339),
			ExtInfo:   map[string]interface{}{"result": result},
		}
		t.chJobStatus <- status
	}
}

func (t *management) getJobTotal(c *gin.Context) {
	visitor, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	params := common.JobTotalQueryParams{}
	params.BeginTime = c.Query("begin_at")
	params.EndTime = c.Query("end_at")
	jobTotal, err := t.dbClient.GetJobTotal(params, visitor)
	t.response(jobTotal, err, false, c)
}

func (t *management) getJobInfo(c *gin.Context) {
	visitor, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	params := common.JobInfoQueryParams{}
	params.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "1"))
	params.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	params.TimeStamp = c.Query("timestamp")
	params.JobID = c.QueryArray("job_id")
	params.JobType = c.Query("job_type")

	jobInfo, err := t.dbClient.GetJob(params, visitor)
	t.response(jobInfo, err, false, c)
}

func (t *management) getJobStatus(c *gin.Context) {
	visitor, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	params := common.JobStatusQueryParams{}
	params.JobID = c.Query("job_id")
	params.JobType = c.Query("job_type")
	params.JobStatus = c.Query("job_status")
	params.BeginTime = c.Query("begin_at")
	params.EndTime = c.Query("end_at")
	jobStatus, err := t.dbClient.GetJobStatus(params, visitor)
	t.response(jobStatus, err, false, c)
}

func (t *management) postJobInfo(c *gin.Context) {
	visitor, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	job := common.JobInfo{}
	if err := t.request(c, &job); (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	if !nsqConnected {
		err := utils.NewECronError(common.ErrMSMQClientUnavailable, common.InternalError, nil)
		t.response(nil, err, false, c)
		return
	}

	job.JobID = uuid.NewV4().String()
	job.CreateTime = time.Now().Format(time.RFC3339)
	job.UpdateTime = time.Now().Format(time.RFC3339)
	job.TenantID = visitor.ClientID

	err = t.dbClient.InsertJob(job, visitor)
	t.response(gin.H{"job_id": job.JobID}, err, true, c)

	go func() {
		//添加任务成功，发送消息
		if (*common.ECronError)(nil) == err {
			t.chJobMsg <- common.JobMsg{
				Method: common.CREATE,
				Data:   job,
			}
		}
	}()
}

func (t *management) postJobExecution(c *gin.Context) {
	_, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	job := common.JobInfo{}
	if err := t.request(c, &job); (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	if err := t.dbClient.CheckJobExecuteMode(job.Context.Mode); (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	//立即回复任务状态
	//jobExecuteID := uuid.NewV4().String()
	jobStatus := common.JobStatus{}
	if 0 == len(job.Context.ExecuteID) {
		jobStatus.ExecuteID = uuid.NewV4().String()
		jobStatus.Executor = append(jobStatus.Executor, map[string]interface{}{
			"executor_id": serviceID,
			"executed_at": time.Now().Format(time.RFC3339),
		})
		jobStatus.BeginTime = time.Now().Format(time.RFC3339)
		jobStatus.JobID = job.JobID
		jobStatus.JobName = job.JobName
		jobStatus.JobType = job.JobType
		jobStatus.JobStatus = common.EXECUTING
		jobStatus.ExecuteTimes = 1
		jobStatus.ExtInfo = map[string]interface{}{}

		//立即更新任务执行者
		job.Context.ExecuteID = jobStatus.ExecuteID
	} else {
		jobStatus.JobID = job.JobID
		jobStatus.JobType = job.JobType
		jobStatus.JobStatus = common.EXECUTING
		jobStatus.ExecuteID = job.Context.ExecuteID
		jobStatus.Executor = append(jobStatus.Executor, map[string]interface{}{
			"executor_id": serviceID,
			"executed_at": time.Now().Format(time.RFC3339),
		})
	}

	t.response(jobStatus, nil, false, c)

	go t.executeJob(job)
}

func (t *management) postWebhook(c *gin.Context) {
	_, ecronErr := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != ecronErr {
		t.response(nil, ecronErr, false, c)
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if nil != err {
		storageLog.Errorln(err)
		t.response(nil, nil, false, c)
		return
	}

	result := make(map[string]interface{})
	err = jsoniter.Unmarshal(body, &result)
	if nil != err {
		storageLog.Errorln(err)
		t.response(nil, nil, false, c)
		return
	}

	executeID := c.Param("execute_id")

	storageLog.Infof("execute_id:%v,body:%v", executeID, string(body))

	if !nsqConnected {
		err := utils.NewECronError(common.ErrMSMQClientUnavailable, common.InternalError, nil)
		t.response(nil, err, false, c)
		return
	}

	t.response(nil, nil, false, c)

	if nil == err {
		//收到任务异步执行通知，解析正确，发送消息
		go t.handleWebhook(executeID, result)
	}
}

func (t *management) putJobInfo(c *gin.Context) {
	visitor, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	job := common.JobInfo{}
	if err := t.request(c, &job); (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	if !nsqConnected {
		err := utils.NewECronError(common.ErrMSMQClientUnavailable, common.InternalError, nil)
		t.response(nil, err, false, c)
		return
	}

	job.JobID = c.Param(pathJobID)
	job.UpdateTime = time.Now().Format(time.RFC3339)

	err = t.dbClient.UpdateJob(job, visitor)
	t.response(nil, err, false, c)

	go func() {
		//更新任务成功，发送消息
		if (*common.ECronError)(nil) == err {
			t.chJobMsg <- common.JobMsg{
				Method: common.UPDATE,
				Data:   job,
			}
		}
	}()
}

func (t *management) putJobStatus(c *gin.Context) {
	visitor, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	jobStatus := make([]common.JobStatus, 0)
	if err := t.request(c, &jobStatus); (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	failedIDs := make([]string, 0)
	failedIDs, err = t.dbClient.UpdateJobStatus(jobStatus, visitor)
	resMap := make(map[string]interface{})
	resMap["failed_job_id"] = failedIDs
	t.response(resMap, err, false, c)
}

func (t *management) putJobEnable(c *gin.Context) {
	visitor, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	jobID := c.Param("job_id")
	reqBody, reqErr := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if nil != reqErr {
		respErr := utils.NewECronError(fmt.Sprintf("%v", reqErr), common.BadRequest, nil)
		t.response(nil, respErr, false, c)
		return
	}
	enable := jsoniter.Get(reqBody, "enable").ToBool()

	if !nsqConnected {
		err := utils.NewECronError(common.ErrMSMQClientUnavailable, common.InternalError, nil)
		t.response(nil, err, false, c)
		return
	}
	id := strings.Split(jobID, ",")
	updateTime := time.Now().Format(time.RFC3339)
	err = t.dbClient.BatchJobEnable(id, enable, updateTime, visitor)
	t.response(nil, err, false, c)

	go func() {
		//启用/禁用任务成功，发送消息
		if (*common.ECronError)(nil) == err {
			for _, v := range id {
				t.chJobMsg <- common.JobMsg{
					Method: common.ENABLE,
					Data:   common.JobInfo{JobID: v, Enabled: enable, UpdateTime: updateTime},
				}
			}
		}
	}()
}

func (t *management) putJobNotify(c *gin.Context) {
	visitor, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	jobID := c.Param("job_id")
	notify := common.JobNotify{}
	if err := t.request(c, &notify); (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	if !nsqConnected {
		err := utils.NewECronError(common.ErrMSMQClientUnavailable, common.InternalError, nil)
		t.response(nil, err, false, c)
		return
	}

	id := strings.Split(jobID, ",")
	updateTime := time.Now().Format(time.RFC3339)
	failedIDs := make([]string, 0)
	failedIDs, err = t.dbClient.BatchJobNotify(id, notify, updateTime, visitor)
	resMap := make(map[string]interface{})
	resMap["failed_job_id"] = failedIDs
	t.response(resMap, err, false, c)
	go func() {
		//修改任务通知地址成功，发送消息
		if (*common.ECronError)(nil) == err {
			for _, v := range id {
				if t.inSlice(failedIDs, v) {
					continue
				}
				t.chJobMsg <- common.JobMsg{
					Method: common.NOTIFY,
					Data: common.JobInfo{
						JobID:      v,
						Context:    common.JobContext{Notify: notify},
						UpdateTime: updateTime,
					},
				}
			}
		}
	}()
}

func (t *management) inSlice(arr []string, key string) bool {
	for _, k := range arr {
		if k == key {
			return true
		}
	}
	return false
}

func (t *management) deleteJobInfo(c *gin.Context) {
	visitor, err := t.check(c.GetHeader(common.Authorization), c.GetHeader(common.AdminSecret), c.GetHeader(common.AdminCode))
	if (*common.ECronError)(nil) != err {
		t.response(nil, err, false, c)
		return
	}

	if !nsqConnected {
		err := utils.NewECronError(common.ErrMSMQClientUnavailable, common.InternalError, nil)
		t.response(nil, err, false, c)
		return
	}

	jobID := c.Param("job_id")
	err = t.dbClient.DeleteJob(jobID, visitor)
	t.response(nil, err, false, c)

	go func() {
		//删除任务成功，发送消息
		id := strings.Split(jobID, ",")
		if (*common.ECronError)(nil) == err {
			for _, v := range id {
				t.chJobMsg <- common.JobMsg{
					Method: common.DELETE,
					Data:   common.JobInfo{JobID: v, TenantID: visitor.ClientID},
				}
			}
		}
	}()
}

func (t *management) getHealth(c *gin.Context) {
	t.response("OK", nil, false, c)
}

func (t *management) getAlive(c *gin.Context) {
	t.response("OK", nil, false, c)
}
