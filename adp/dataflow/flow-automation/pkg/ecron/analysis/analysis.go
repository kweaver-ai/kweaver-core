package analysis

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/robfig/cron/v3"
)

//go:generate mockgen -package mock -source ../analysis/analysis.go -destination ../mock/mock_analysis.go

const (
	healthPATH = "/health/ready"
	alivePATH  = "/health/alive"
)

// AnalysisService 分析服务接口
type AnalysisService interface {

	//开启服务
	Start()

	//关闭服务
	Stop()
}

// NewAnalysisService 创建分析服务对象
func NewAnalysisService() AnalysisService {
	return &eanalysis{
		cronClient:            nil,
		msmqClient:            nil,
		httpClient:            nil,
		authClient:            nil,
		mapJobInfo:            cmap.New(),
		mapJobStatus:          cmap.New(),
		mapEntryID:            cmap.New(),
		mapLostImmediateJob:   cmap.New(),
		chJobStatus:           make(chan map[string]interface{}, 1),
		chConsumeImmediateJob: make(chan common.JobInfo, 1),
		chConsumeJobMsg:       make(chan common.JobMsg, 1),
		chConsumeJobStatus:    make(chan common.JobStatus, 1),
		chStartJob:            make(chan string, 1),
		chEndJob:              make(chan string, 1),
		bStart:                false,
		startMu:               new(sync.Mutex),
	}
}

// 分析模块数据结构集
type eanalysis struct {
	cronClient            *cron.Cron                  //定时服务
	msmqClient            utils.MsmqClient            //消息队列服务
	httpClient            utils.HTTPClient            //HTTP客户端服务
	authClient            utils.OAuthClient           //OAuth客户端
	mapJobInfo            cmap.ConcurrentMap          //任务信息：key=job_id, value=jobInfo
	mapEntryID            cmap.ConcurrentMap          //任务ID和定时句柄关系：key=job_id, value=EntryID
	mapJobStatus          cmap.ConcurrentMap          //任务状态：key=execute_id, value=jobStatus
	mapLostImmediateJob   cmap.ConcurrentMap          //丢失的即时任务，任务执行太快，即时任务消息先于http回复，导致任务迷失，key=executeID，value=jobInfo
	chJobStatus           chan map[string]interface{} //任务状态通道
	chConsumeImmediateJob chan common.JobInfo         //即时任务消息消费通道
	chConsumeJobMsg       chan common.JobMsg          //任务信息消息消费通道
	chConsumeJobStatus    chan common.JobStatus       //任务状态消息消费通道
	chStartJob            chan string                 //任务开始通知
	chEndJob              chan string                 //任务结束通知
	startMu               *sync.Mutex                 //服务锁
	bStart                bool                        //是否开启服务
}

var (
	analysisLog    = utils.NewLogger()
	analysisConfig = utils.NewConfiger()

	cronAddr     = analysisConfig.Config().CronAddr
	cronPort     = analysisConfig.Config().CronPort
	cronProtocol = analysisConfig.Config().CronProtocol
	srverPort    = analysisConfig.Config().AnlysPort
	httpAccess   = common.GetHTTPAccess(cronAddr, cronPort, cronProtocol == common.HTTPS) //外部服务访问都走https

	jobFailures = common.GetIntMoreThanLowerLimit(analysisConfig.Config().JobFailures, 3)
	serviceID   = analysisConfig.Config().AnalysisServiceID

	loadSleep             = common.GetSleepDuration(analysisConfig.Config().AnalysisLoadSleep)
	refreshSleep          = common.GetSleepDuration(analysisConfig.Config().AnalysisRefreshSleep)
	jobStatusRefreshSleep = common.GetSleepDuration(analysisConfig.Config().JobStatusRefreshSleep)
	lostImmediateJobSleep = common.GetSleepDuration(analysisConfig.Config().LostImmediateJobSleep)
)

// Start 开启分析服务
func (a *eanalysis) Start() {
	if a.isStart() {
		analysisLog.Infoln("already start")
		return
	}

	analysisLog.Infof("start analysis service, id = %v, version = %v", serviceID, common.Version)
	a.init()
	a.start()
	a.setStart(true)
}

// Stop 关闭分析服务
func (a *eanalysis) Stop() {
	if nil != a.cronClient {
		a.cronClient.Stop()
	}

	if nil != a.authClient {
		a.authClient.Release()
	}
}

func (a *eanalysis) init() {
	a.initCronClient()
	a.initMsmqClient()
	a.initHTTPClient()
	a.initAuthClient()
}

func (a *eanalysis) getHealth(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.String(http.StatusOK, "OK")
}

func (a *eanalysis) getAlive(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.String(http.StatusOK, "OK")
}

func (a *eanalysis) healthz(port int) error {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET(healthPATH, a.getHealth)
	r.GET(alivePATH, a.getAlive)
	return r.Run(fmt.Sprintf(":%d", port))
}

func (a *eanalysis) initCronClient() {
	if nil == a.cronClient {
		a.cronClient = cron.New(cron.WithSeconds())
	}

	if nil == a.cronClient {
		analysisLog.Fatalln(common.ErrCronClientUnavailable)
	}

	a.cronClient.Start()
}

func (a *eanalysis) initMsmqClient() {
	if nil == a.msmqClient {
		a.msmqClient = utils.NewMsmqClient()
	}
	if nil == a.msmqClient {
		analysisLog.Fatalln(common.ErrMSMQClientUnavailable)
	}
}

func (a *eanalysis) initHTTPClient() {
	if nil == a.httpClient {
		a.httpClient = utils.NewHTTPClient()
	}
	if nil == a.httpClient {
		analysisLog.Fatalln(common.ErrHTTPClientUnavailable)
	}
}

func (a *eanalysis) initAuthClient() {
	if nil == a.authClient {
		a.authClient = utils.NewOAuthClient()
	}

	if nil == a.authClient {
		analysisLog.Fatalln(common.ErrAuthClientUnavailable)
	}
}

func (a *eanalysis) start() {
	go func() {
		//等程序加载工作结束后，才能开始消费消息
		for {
			err := a.load()
			if nil != err {
				analysisLog.Errorln("load resource failed, err =", err)
				time.Sleep(loadSleep)
				continue
			}
			break
		}

		for {
			err := a.refresh()
			if nil != err {
				analysisLog.Errorln("refresh resource failed, err =", err)
				time.Sleep(refreshSleep)
				continue
			}
			break
		}
	}()
	gin.SetMode(gin.ReleaseMode)
	err := a.healthz(srverPort)
	if nil != err {
		analysisLog.Errorln("healthz failed, err =", err)
	}
}

func (a *eanalysis) setStart(bStart bool) {
	a.startMu.Lock()
	defer a.startMu.Unlock()
	a.bStart = bStart
}

func (a *eanalysis) isStart() bool {
	a.startMu.Lock()
	defer a.startMu.Unlock()
	return a.bStart
}

// 分析服务调用API路径
var (
	managementPrefix    = "/api/ecron-management/v1"
	jobTotalPath        = fmt.Sprintf("%s/jobtotal", managementPrefix)
	jobInfoByPagePath   = fmt.Sprintf("%s/job", managementPrefix)
	jobStatusWithIDPath = fmt.Sprintf("%s/jobstatus/", managementPrefix)
	jobExecutionPath    = fmt.Sprintf("%s/jobexecution", managementPrefix)
)

// 加载任务列表
func (a *eanalysis) load() error {
	analysisLog.Infoln("begin load()")

	jobTotal, err := a.getJobTotal()
	if nil != err {
		return err
	}

	for pos, page, limit := 0, 1, 100; pos < jobTotal.Total; page++ {
		pos, err = a.getJobInfoByPage(limit, page, jobTotal.TimeStamp)
		if nil != err {
			break
		}
	}

	if nil == err {
		go func() {
			for _, v := range a.mapJobInfo.Items() {
				switch job := v.(type) {
				case common.JobInfo:
					if job.Enabled {
						a.cronJob(job)
					}
				default:
					analysisLog.Errorln(job)
				}
			}
		}()
	}

	analysisLog.Infoln("end load()")

	return nil
}

// 获取任务总数
func (a *eanalysis) getJobTotal() (jobTotal common.JobTotal, err error) {
	if nil == a.httpClient {
		err = errors.New(common.ErrHTTPClientUnavailable)
		return
	}
	target := fmt.Sprintf("%s%s", httpAccess, jobTotalPath)
	err = a.httpClient.Get(target, a.addToken(map[string]string{}), &jobTotal)
	return
}

// 分页获取任务信息
func (a *eanalysis) getJobInfoByPage(limit int, page int, timestamp string) (pos int, err error) {
	if nil == a.httpClient {
		err = errors.New(common.ErrHTTPClientUnavailable)
		return
	}
	target := fmt.Sprintf("%s%s?limit=%d&page=%d&timestamp=%s", httpAccess, jobInfoByPagePath, limit, page, url.QueryEscape(timestamp))
	resp := make([]common.JobInfo, 0)
	err = a.httpClient.Get(target, a.addToken(map[string]string{}), &resp)
	if nil != err {
		return
	}

	analysisLog.Infof("the number of obtained jobs is %v", len(resp))

	{
		for _, v := range resp {
			a.mapJobInfo.Set(v.JobID, v)
		}
	}

	pos = page * limit
	return
}

// 刷任务信息或任务状态
func (a *eanalysis) refresh() error {
	analysisLog.Infoln("begin refresh()")

	//订阅消息
	if nil != a.msmqClient {
		a.msmqClient.Subscribe(common.TopicImmediateJob, common.ChannelImmediateJob, a.immediateJob)
		a.msmqClient.Subscribe(common.TopicCronJob, common.ChannelCronJob, a.refreshJob)
		a.msmqClient.Subscribe(common.TopicJobStatus, common.ChannelJobStatus, a.refreshStatus)
	}

	//定期持久化任务状态
	go func() {
		for {
			mapStatus := a.mapJobStatus.Items()
			a.chJobStatus <- mapStatus

			go func() {
				m := <-a.chJobStatus
				go a.handleStatus(m)
			}()

			time.Sleep(jobStatusRefreshSleep)
		}
	}()

	//处理迷失的即时任务
	go func() {
		for {
			mapJob := a.mapLostImmediateJob.Items()
			for k, v := range mapJob {
				if a.mapJobStatus.Has(k) {
					switch job := v.(type) {
					case common.JobInfo:
						a.executeImmediateJob(job)
						a.mapLostImmediateJob.Remove(k)
					}
				}
			}
			time.Sleep(lostImmediateJobSleep)
		}
	}()

	//任务开始和结束
	go func() {
		for {
			select {
			case id := <-a.chStartJob:
				{
					analysisLog.Infoln("start a job, remove it from cron")
					go a.removeCronJob(id)
				}
			case id := <-a.chEndJob:
				go func() {
					analysisLog.Infoln("end a job, recron it")
					if v, ok := a.mapJobInfo.Get(id); ok {
						switch job := v.(type) {
						case common.JobInfo:
							a.cronJob(job)
						}
					}
				}()
			}
		}
	}()

	analysisLog.Infoln("end refresh()")

	return nil
}

// 定时管理任务
func (a *eanalysis) cronJob(job common.JobInfo) {
	if nil == a.cronClient {
		analysisLog.Errorln("cronJob failed,", common.ErrCronClientUnavailable)
		return
	}

	//一个任务同时只能被定时管理一次

	if a.mapEntryID.Has(job.JobID) {
		analysisLog.Infof("already cron job[%v,%v,%v,%v]", job.JobID, job.JobName, job.JobCronTime, job.JobType)
		return
	}

	id, err := a.cronClient.AddFunc(job.JobCronTime, func() {
		if !a.readyToExecute(job) {
			return
		}

		if nil == a.httpClient {
			analysisLog.Errorln("cronClient.AddFunc", common.ErrHTTPClientUnavailable)
			return
		}

		target := fmt.Sprintf("%s%s", httpAccess, jobExecutionPath)
		resp := common.JobStatus{}
		err := a.httpClient.Post(target, a.addToken(map[string]string{}), job, &resp)

		//获取定时服务即时回复，增加或更新任务状态
		//由定时管理器触发的任务一定是一个新的任务
		if nil == err {
			if nil != resp.ExtInfo {
				resp.ExtInfo = make(map[string]interface{})
			}
			resp.ExtInfo[common.IsDeleted] = 0
			resp.ExtInfo[common.TenantID] = job.TenantID
			a.mapJobStatus.Set(resp.ExecuteID, resp)
			a.chStartJob <- job.JobID
		}

		analysisLog.Infof("send a job[%v,%v,%v,%v], err = %v", job.JobID, job.JobName, job.JobCronTime, job.JobType, err)
	})

	if nil == err {
		a.mapEntryID.Set(job.JobID, id)
	}

	analysisLog.Infof("cron a job[%v,%v,%v,%v], err = %v", job.JobID, job.JobName, job.JobCronTime, job.JobType, err)
}

func (a *eanalysis) readyToExecute(job common.JobInfo) bool {
	if !job.Enabled {
		analysisLog.Infof("ready to send a job[%v,%v,%v,%v], but it's disabled", job.JobID, job.JobName, job.JobCronTime, job.JobType)
		return false
	}

	now, _ := common.StringToTimeStamp(time.Now().Format(time.RFC3339))
	begin, _ := common.StringToTimeStamp(job.Context.BeginTime)
	end, _ := common.StringToTimeStamp(job.Context.EndTime)

	if len(job.Context.BeginTime) > 0 && now < begin {
		analysisLog.Infof("ready to send a job[%v,%v,%v,%v], but it's begin time is[%v]", job.JobID, job.JobName, job.JobCronTime, job.JobType, job.Context.BeginTime)
		return false
	}

	if len(job.Context.EndTime) > 0 && now > end {
		analysisLog.Infof("ready to send a job[%v,%v,%v,%v], but the job has expired[%v]", job.JobID, job.JobName, job.JobCronTime, job.JobType, job.Context.EndTime)
		return false
	}

	return true
}

func (a *eanalysis) immediateJob(message []byte) error {
	job := common.JobInfo{}
	err := jsoniter.Unmarshal(message, &job)
	if nil != err {
		analysisLog.Errorln("immediateJob() unmarshal immediate job failed", err)
		return err
	}

	a.chConsumeImmediateJob <- job

	go func() {
		chJob := <-a.chConsumeImmediateJob
		go a.executeImmediateJob(chJob)
	}()

	return nil
}

func (a *eanalysis) readyToExecuteImmediateJob(job common.JobInfo) (immediateJob common.JobInfo, flag bool) {
	if nil == a.httpClient {
		analysisLog.Errorln(common.ErrHTTPClientUnavailable)
		return
	}

	vJob, ok := a.mapJobInfo.Get(job.JobID)
	if !ok {
		analysisLog.Errorf("ready to send immediate job, but can't find job[%v]", job.JobID)
		return
	}

	vStatus, ok := a.mapJobStatus.Get(job.Context.ExecuteID)
	if !ok {
		analysisLog.Errorf("ready to send immediate job, but can't find executeID[%v]", job.Context.ExecuteID)
		a.mapLostImmediateJob.Set(job.Context.ExecuteID, job)
		return
	}

	//以防出现执行流水号相同，而任务ID不匹配的情况（以假乱真）
	switch status := vStatus.(type) {
	case common.JobStatus:
		if job.JobID != status.JobID {
			analysisLog.Errorf("%v. executeID[%v] job.JobID[%v] status.JobID[%v]", common.ErrExecuteIDAndJobIDConfused, job.Context.ExecuteID, job.JobID, status.JobID)
			return
		}
	}

	//任务执行失败后，重试过程中任务被删除，理应中断即时任务
	{
		cb := func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
			if exist {
				switch status := valueInMap.(type) {
				case common.JobStatus:
					if deleted, ok := status.ExtInfo[common.IsDeleted]; ok {
						if deleted.(int) == 1 {
							status.JobStatus = common.INTERRUPT
							status.EndTime = time.Now().Format(time.RFC3339)
						}
					}
					return status
				}

			}
			return valueInMap
		}
		res := a.mapJobStatus.Upsert(job.Context.ExecuteID, nil, cb)
		switch status := res.(type) {
		case common.JobStatus:
			if status.JobStatus == common.INTERRUPT {
				analysisLog.Infof("%v. interrupt a job[%v,%v,%v,%v], executeID %v", common.ErrJobExecutedTooManyTimes, job.JobID, job.JobName, job.JobCronTime, job.JobType, job.Context.ExecuteID)
				return
			}
		}
	}

	//超过执行次数的任务将被放弃
	{
		cb := func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
			if exist {
				switch status := valueInMap.(type) {
				case common.JobStatus:
					if status.ExecuteTimes >= jobFailures {
						status.JobStatus = common.ABANDON
						status.EndTime = time.Now().Format(time.RFC3339)
					}
					return status
				}
			}
			return valueInMap
		}
		res := a.mapJobStatus.Upsert(job.Context.ExecuteID, nil, cb)
		switch status := res.(type) {
		case common.JobStatus:
			if status.JobStatus == common.ABANDON {
				go func() {
					a.chEndJob <- job.JobID
				}()
				analysisLog.Infof("%v. abandon a job[%v,%v,%v,%v], executeID %v", common.ErrJobExecutedTooManyTimes, job.JobID, job.JobName, job.JobCronTime, job.JobType, job.Context.ExecuteID)
				return
			}
		}
	}

	//补充任务信息
	switch v := vJob.(type) {
	case common.JobInfo:
		immediateJob = v
		immediateJob.JobType = job.JobType
		immediateJob.Context.ExecuteID = job.Context.ExecuteID
		flag = true
	}

	return
}

func (a *eanalysis) executeImmediateJob(job common.JobInfo) {
	if immediateJob, flag := a.readyToExecuteImmediateJob(job); flag {
		target := fmt.Sprintf("%s%s", httpAccess, jobExecutionPath)
		resp := common.JobStatus{}
		err := a.httpClient.Post(target, a.addToken(map[string]string{}), immediateJob, &resp)

		analysisLog.Infof("send a job[%v,%v,%v,%v,%v], err = %v", job.JobID, job.JobName, job.JobCronTime, job.JobType, job.Context.ExecuteID, err)

		//即时任务，不存在新增状态，只有更新
		if nil == err && resp.ExecuteID == job.Context.ExecuteID {
			cb := func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
				if exist {
					switch status := valueInMap.(type) {
					case common.JobStatus:
						status.Executor = append(status.Executor, resp.Executor...)
						status.ExecuteTimes++
						return status
					}
				}
				return valueInMap
			}
			a.mapJobStatus.Upsert(resp.ExecuteID, nil, cb)
		}
	}
}

func (a *eanalysis) refreshJob(message []byte) error {
	jobMsg := common.JobMsg{}
	err := jsoniter.Unmarshal(message, &jobMsg)
	if nil != err {
		analysisLog.Errorln("refreshJob() unmarshal cron job failed", err)
		return err
	}

	a.chConsumeJobMsg <- jobMsg

	go func() {
		msg := <-a.chConsumeJobMsg
		go func() {
			switch msg.Method {
			case common.CREATE:
				a.insertJob(msg.Data)
			case common.UPDATE: //全量更新
				a.updateJob(msg.Data)
			case common.DELETE:
				a.deleteJob(msg.Data)
			case common.ENABLE: //启用/禁用任务
				a.enableJob(msg.Data)
			case common.NOTIFY: //修改任务通知地址
				a.notifyJob(msg.Data)
			}
		}()
	}()

	return nil
}

func (a *eanalysis) insertJob(job common.JobInfo) {
	if a.mapJobInfo.Has(job.JobID) {
		analysisLog.Infof("insert a job[%v,%v,%v], but it does exist", job.JobID, job.JobName, job.JobCronTime)
		return
	}

	a.mapJobInfo.Set(job.JobID, job)
	a.removeCronJob(job.JobID) //此语句可解决一个问题：添加任务消息入消息队列成功，分析模块没来得及消费异常重启，任务会被重复定时
	a.cronJob(job)

	analysisLog.Infof("insert a job[%v,%v,%v]", job.JobID, job.JobName, job.JobCronTime)
}

func (a *eanalysis) updateJob(job common.JobInfo) {
	if !a.mapJobInfo.Has(job.JobID) {
		analysisLog.Infof("update a job[%v,%v,%v], but it does not exist", job.JobID, job.JobName, job.JobCronTime)
		return
	}

	//全量更新job，但任务的租户类型保持不变，即使管理员来修改也要保持不变
	tmp, _ := a.mapJobInfo.Get(job.JobID)
	switch v := tmp.(type) {
	case common.JobInfo:
		job.TenantID = v.TenantID
		a.mapJobInfo.Set(job.JobID, job)
		a.removeCronJob(job.JobID)
		a.cronJob(job)

		analysisLog.Infof("update a job[%v,%v,%v]", job.JobID, job.JobName, job.JobCronTime)
	}
}

func (a *eanalysis) deleteJob(job common.JobInfo) {
	if !a.mapJobInfo.Has(job.JobID) {
		analysisLog.Infof("delete a job[%v], but it does not exist", job.JobID)
		return
	}

	a.mapJobInfo.Remove(job.JobID)
	a.removeCronJob(job.JobID)
	a.deleteJobStatus(job.JobID)

	analysisLog.Infof("delete a job[%v]", job.JobID)
}

func (a *eanalysis) enableJob(job common.JobInfo) {
	if !a.mapJobInfo.Has(job.JobID) {
		analysisLog.Infof("enable a job[%v,%v], but it does not exist", job.JobID, job.Enabled)
		return
	}

	v, _ := a.mapJobInfo.Get(job.JobID)
	switch info := v.(type) {
	case common.JobInfo:
		info.Enabled = job.Enabled
		info.UpdateTime = job.UpdateTime
		a.mapJobInfo.Set(info.JobID, info)

		a.removeCronJob(info.JobID)
		a.cronJob(info)
		analysisLog.Infof("enable a job[%v,%v,%v,%v]", info.JobID, info.JobName, info.JobCronTime, info.Enabled)
	}
}

func (a *eanalysis) notifyJob(job common.JobInfo) {
	if !a.mapJobInfo.Has(job.JobID) {
		analysisLog.Infof("notify a job[%v,%v], but it does not exist", job.JobID, job.Context.Notify)
		return
	}

	v, _ := a.mapJobInfo.Get(job.JobID)
	switch info := v.(type) {
	case common.JobInfo:
		info.Context.Notify = job.Context.Notify
		info.UpdateTime = job.UpdateTime
		a.mapJobInfo.Set(info.JobID, info)

		a.removeCronJob(info.JobID)
		a.cronJob(info)

		analysisLog.Infof("notify a job[%v,%v,%v,%v]", info.JobID, info.JobName, info.JobCronTime, info.Context.Notify)
	}
}

func (a *eanalysis) removeCronJob(jobID string) {
	if v, ok := a.mapEntryID.Get(jobID); ok {
		switch id := v.(type) {
		case cron.EntryID:
			a.mapEntryID.Remove(jobID)
			if nil != a.cronClient {
				a.cronClient.Remove(id)
			}
		}
	}
}

// deleteJobStatus 如果任务本身被删除了，其正在执行的所有任务状态都增加任务被删除标识
func (a *eanalysis) deleteJobStatus(jobID string) {
	mapStatus := a.mapJobStatus.Items()
	for k := range mapStatus {
		cb := func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
			if exist {
				switch status := valueInMap.(type) {
				case common.JobStatus:
					if status.JobID == jobID {
						status.ExtInfo[common.IsDeleted] = 1
					}
					return status
				}
			}
			return valueInMap
		}
		a.mapJobStatus.Upsert(k, nil, cb)
	}
}

func (a *eanalysis) refreshStatus(message []byte) error {
	jobStatus := common.JobStatus{}
	err := jsoniter.Unmarshal(message, &jobStatus)
	if nil != err {
		analysisLog.Errorln("refreshStatus() unmarshal job status failed")
		return err
	}

	a.chConsumeJobStatus <- jobStatus

	go func() {
		status := <-a.chConsumeJobStatus

		//此处刷新，只关注任务的完成情况，如果拿到的任务状态是正在执行，则无视它
		if status.JobStatus == common.EXECUTING || 0 == len(status.EndTime) {
			return
		}

		cb := func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
			if exist {
				switch s := valueInMap.(type) {
				case common.JobStatus:
					s.JobStatus = status.JobStatus
					s.EndTime = status.EndTime
					return s
				}
			}
			return valueInMap
		}
		res := a.mapJobStatus.Upsert(status.ExecuteID, nil, cb)
		switch s := res.(type) {
		case common.JobStatus:
			if len(s.EndTime) > 0 {
				go func() {
					a.chEndJob <- s.JobID
				}()
				analysisLog.Infof("completed a job[%v,%v,%v]", s.JobID, s.JobName, s.JobStatus)
			}
		}
	}()

	return nil
}

// handleStatus 处理任务状态
func (a *eanalysis) handleStatus(m map[string]interface{}) {
	if 0 == len(m) {
		return
	}

	analysisLog.Infoln("get changed job status, size =", len(m))

	//每1000个更新一次
	mapStatus := make([]interface{}, 0)
	var i = 0
	for _, v := range m {
		if i > 0 && 0 == i%1000 {
			mapStatus = append(mapStatus, v)
			_ = a.updateStatus(mapStatus)
			mapStatus = make([]interface{}, 0)
		} else {
			mapStatus = append(mapStatus, v)
		}
		i++
	}
	_ = a.updateStatus(mapStatus)
}

// updateStatus 任务状态持久化
func (a *eanalysis) updateStatus(mapStatus []interface{}) error {
	if nil == a.httpClient {
		analysisLog.Errorln("updateStatus", common.ErrHTTPClientUnavailable)
		return errors.New(common.ErrHTTPClientUnavailable)
	}

	target := fmt.Sprintf("%s%s%s", httpAccess, jobStatusWithIDPath, "execute_id")
	err := a.httpClient.Put(target, a.addToken(map[string]string{"Connection": "close"}), mapStatus, nil)
	if nil != err {
		analysisLog.Errorln("send job status failed, err =", err)
	} else {
		//判断该次执行是否已结束，如果是就清除其执行状态
		for _, v := range mapStatus {
			switch status := v.(type) {
			case common.JobStatus:
				if len(status.EndTime) > 0 {
					a.mapJobStatus.Remove(status.ExecuteID)
					analysisLog.Infof("remove a completed job status[%v]", status.ExecuteID)
				}
			}
		}
	}

	return err
}

func (a *eanalysis) addToken(headers map[string]string) map[string]string {
	if nil == a.authClient {
		analysisLog.Errorln("updateStatus", common.ErrAuthClientUnavailable)
		return nil
	}

	if nil == headers {
		headers = make(map[string]string)
	}

	// 增加管理员动态密钥
	secret := a.authClient.GetSecret()
	code, _ := a.authClient.GetCode(secret)
	headers[common.AdminSecret] = secret
	headers[common.AdminCode] = code

	return headers
}
