package analysis

import (
	"fmt"
	"sync"
	"testing"

	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/mock"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/robfig/cron/v3"
	uuid "github.com/satori/go.uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newAnalysis(m utils.MsmqClient, h utils.HTTPClient, a utils.OAuthClient) *eanalysis {
	return &eanalysis{
		cronClient:            nil,
		msmqClient:            m,
		httpClient:            h,
		authClient:            a,
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

func TestNewAnalysisService(t *testing.T) {
	Convey("NewAnalysisService", t, func() {
		service := NewAnalysisService()
		assert.NotEqual(t, service, nil)
	})
}

func TestAnalysisStop(t *testing.T) {
	Convey("Stop", t, func() {
		Convey("unitialized", func() {
			pa := newAnalysis(nil, nil, nil)
			pa.Stop()
		})

		Convey("initialized", func() {
			pa := newAnalysis(nil, nil, nil)
			pa.Stop()
		})
	})
}

func TestAnalysisInitCronService(t *testing.T) {
	Convey("initCronService", t, func() {
		Convey("without init", func() {
			pa := newAnalysis(nil, nil, nil)
			assert.Equal(t, pa.cronClient, (*cron.Cron)(nil))
			pa.initCronClient()
			assert.NotEqual(t, pa.cronClient, nil)
		})

		Convey("repeated init", func() {
			pa := newAnalysis(nil, nil, nil)
			assert.NotEqual(t, pa.cronClient, nil)
			pa.initCronClient()
			assert.NotEqual(t, pa.cronClient, nil)
		})
	})
}

func TestAnalysisInitMsmqService(t *testing.T) {
	Convey("initMsmqService", t, func() {
		Convey("without init", func() {
			pa := newAnalysis(nil, nil, nil)
			assert.Equal(t, pa.msmqClient, nil)
		})
	})
}

func TestAnalysisInitHTTPService(t *testing.T) {
	Convey("initHTTPService", t, func() {
		Convey("without init", func() {
			pa := newAnalysis(nil, nil, nil)
			assert.Equal(t, pa.httpClient, nil)
			pa.initHTTPClient()
			assert.NotEqual(t, pa.httpClient, nil)
		})

		Convey("repeated init", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mock.NewMockMsmqClient(ctrl)
			h := mock.NewMockHTTPClient(ctrl)
			a := mock.NewMockOAuthClient(ctrl)
			pa := newAnalysis(m, h, a)

			assert.NotEqual(t, pa.httpClient, nil)
			pa.initHTTPClient()
			assert.NotEqual(t, pa.httpClient, nil)
		})
	})
}

func TestAnalysisInitAuthClient(t *testing.T) {
	Convey("initAuthClient", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		h := mock.NewMockHTTPClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, h, a)

		assert.NotEqual(t, pa.httpClient, nil)
		pa.initAuthClient()
		assert.NotEqual(t, pa.httpClient, nil)
	})
}

func TestAnalysisGetJobTotal(t *testing.T) {
	Convey("getJobTotal", t, func() {
		timestamp := time.Now().Format(time.RFC3339)

		Convey("without init", func() {
			pa := newAnalysis(nil, nil, nil)
			total, err := pa.getJobTotal()
			assert.Equal(t, total.Total, 0)
			assert.NotEqual(t, err, nil)
		})

		Convey("normal, Get command return success", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mock.NewMockMsmqClient(ctrl)
			h := mock.NewMockHTTPClient(ctrl)
			a := mock.NewMockOAuthClient(ctrl)
			pa := newAnalysis(m, h, a)

			a.EXPECT().GetSecret().AnyTimes().Return("")
			a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

			h.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(url string, headers map[string]string, respParam interface{}) error {
					params := common.JobTotal{
						Total:     100,
						TimeStamp: timestamp,
					}
					body, _ := jsoniter.Marshal(params)
					return jsoniter.Unmarshal(body, respParam)
				},
			)

			total, err := pa.getJobTotal()
			assert.Equal(t, total.Total, 100)
			assert.Equal(t, total.TimeStamp, timestamp)
			assert.Equal(t, err, nil)
		})
	})
}

func TestAnalysisGetJobInfoByPage(t *testing.T) {
	Convey("getJobInfoByPage", t, func() {
		page := 1
		limit := 10
		timestamp := time.Now().Format(time.RFC3339)

		Convey("without init", func() {
			pa := newAnalysis(nil, nil, nil)
			pos, err := pa.getJobInfoByPage(page, limit, timestamp)
			assert.Equal(t, pos, 0)
			assert.NotEqual(t, err, nil)
		})

		Convey("normal, Get command return success", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mock.NewMockMsmqClient(ctrl)
			h := mock.NewMockHTTPClient(ctrl)
			a := mock.NewMockOAuthClient(ctrl)
			pa := newAnalysis(m, h, a)

			a.EXPECT().GetSecret().AnyTimes().Return("")
			a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

			h.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(url string, headers map[string]string, respParam interface{}) error {
					params := []common.JobInfo{
						0: {
							JobID: uuid.NewV4().String(),
						},
					}
					body, _ := jsoniter.Marshal(params)
					return jsoniter.Unmarshal(body, respParam)
				},
			)

			pos, err := pa.getJobInfoByPage(page, limit, timestamp)
			assert.Equal(t, pos, 10)
			assert.Equal(t, err, nil)
			assert.Equal(t, pa.mapJobInfo.Count(), 1)
		})
	})
}

func TestAnalysisRefresh(t *testing.T) {
	Convey("refresh", t, func() {
		Convey("without init", func() {
			pa := newAnalysis(nil, nil, nil)
			_ = pa.refresh()
		})
	})
}

func TestAnalysisCronJob(t *testing.T) {
	Convey("cronJob", t, func() {
		Convey("without init", func() {
			pa := newAnalysis(nil, nil, nil)
			pa.cronJob(common.JobInfo{})
		})

		id := uuid.NewV4().String()

		Convey("normal", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mock.NewMockMsmqClient(ctrl)
			h := mock.NewMockHTTPClient(ctrl)
			a := mock.NewMockOAuthClient(ctrl)
			pa := newAnalysis(m, h, a)

			a.EXPECT().GetSecret().AnyTimes().Return("")
			a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)
			pa.initCronClient()

			Convey("already cron the job", func() {
				pa.mapEntryID.Set(id, cron.EntryID(1))
				pa.cronJob(common.JobInfo{
					JobID: id,
				})
			})

			Convey("cron a new job, but failed", func() {
				pa.cronJob(common.JobInfo{
					JobID: id,
				})
			})
		})
	})
}

func TestAnalysisReadyToExecute(t *testing.T) {
	Convey("readyToExecute", t, func() {
		b, _ := time.ParseDuration("1m")
		begin := time.Now().Add(b).Format(time.RFC3339)
		e, _ := time.ParseDuration("-1m")
		end := time.Now().Add(e).Format(time.RFC3339)
		pa := newAnalysis(nil, nil, nil)

		Convey("readyToExecute a disabled job", func() {
			flag := pa.readyToExecute(common.JobInfo{})
			assert.Equal(t, flag, false)
		})

		Convey("readyToExecute a enabled job, but now time is less than begin time", func() {
			flag := pa.readyToExecute(common.JobInfo{
				Enabled: true,
				Context: common.JobContext{
					BeginTime: begin,
				},
			})
			assert.Equal(t, flag, false)
		})

		Convey("readyToExecute a enabled job, but now time is greater than end time", func() {
			flag := pa.readyToExecute(common.JobInfo{
				Enabled: true,
				Context: common.JobContext{
					EndTime: end,
				},
			})
			assert.Equal(t, flag, false)
		})

		Convey("readyToExecute a enabled job, and time is normal", func() {
			flag := pa.readyToExecute(common.JobInfo{
				Enabled: true,
				Context: common.JobContext{
					BeginTime: time.Now().Format(time.RFC3339),
					EndTime:   time.Now().Format(time.RFC3339),
				},
			})
			assert.Equal(t, flag, true)
		})

		Convey("readyToExecute a enabled job, but time is empty", func() {
			flag := pa.readyToExecute(common.JobInfo{
				Enabled: true,
			})
			assert.Equal(t, flag, true)
		})
	})
}

func TestAnalysisImmediateJob(t *testing.T) {
	Convey("immediateJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		id := uuid.NewV4().String()
		job := common.JobInfo{
			JobID: id,
		}

		msg, _ := jsoniter.Marshal(job)
		err := pa.immediateJob(msg)
		assert.Equal(t, nil, err)
	})
}

func TestAnalysisReadyToExecuteimmediateJob(t *testing.T) {
	Convey("readyToExecuteImmediateJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		id := []string{
			uuid.NewV4().String(),
			uuid.NewV4().String(),
		}
		executeID := uuid.NewV4().String()
		pa.mapJobStatus.Set(executeID, common.JobStatus{
			ExecuteID: executeID,
			JobID:     id[0],
			JobStatus: common.EXECUTING,
			ExtInfo: map[string]interface{}{
				common.IsDeleted: 0,
			},
			ExecuteTimes: jobFailures,
		})
		pa.mapJobInfo.Set(id[0], common.JobInfo{JobID: id[0]})
		pa.mapJobInfo.Set(id[1], common.JobInfo{JobID: id[1]})

		Convey("ready to send immediate job, but can't find job", func() {
			job := common.JobInfo{
				JobID: uuid.NewV4().String(),
				Context: common.JobContext{
					ExecuteID: executeID,
				},
			}
			_, flag := pa.readyToExecuteImmediateJob(job)
			assert.Equal(t, flag, false)
		})

		Convey("ready to send immediate job, but can't find executeID", func() {
			eid := uuid.NewV4().String()
			job := common.JobInfo{
				JobID: id[0],
				Context: common.JobContext{
					ExecuteID: eid,
				},
			}
			_, flag := pa.readyToExecuteImmediateJob(job)
			_, ok := pa.mapLostImmediateJob.Get(eid)
			pa.mapLostImmediateJob.Remove(eid)
			assert.Equal(t, flag, false)
			assert.Equal(t, ok, true)
		})

		Convey("executeID and jobID are confused", func() {
			job := common.JobInfo{
				JobID: id[1],
				Context: common.JobContext{
					ExecuteID: executeID,
				},
			}
			_, flag := pa.readyToExecuteImmediateJob(job)
			assert.Equal(t, flag, false)
		})

		Convey("the job is deleted, interrupt a job", func() {
			job := common.JobInfo{
				JobID: id[0],
				Context: common.JobContext{
					ExecuteID: executeID,
				},
			}
			v, _ := pa.mapJobStatus.Get(executeID)
			status := v.(common.JobStatus)
			status.ExtInfo[common.IsDeleted] = 1
			pa.mapJobStatus.Set(executeID, status)
			pa.readyToExecuteImmediateJob(job)
			after, ok := pa.mapJobStatus.Get(executeID)
			assert.Equal(t, after.(common.JobStatus).JobStatus, common.INTERRUPT)
			assert.Equal(t, after.(common.JobStatus).ExtInfo[common.IsDeleted], 1)
			assert.Equal(t, ok, true)
		})

		Convey("the job executed too many times, abandon a job", func() {
			job := common.JobInfo{
				JobID: id[0],
				Context: common.JobContext{
					ExecuteID: executeID,
				},
			}
			pa.readyToExecuteImmediateJob(job)
			jobID := <-pa.chEndJob
			after, ok := pa.mapJobStatus.Get(executeID)
			assert.Equal(t, after.(common.JobStatus).JobStatus, common.ABANDON)
			assert.Equal(t, ok, true)
			assert.Equal(t, jobID, id[0])
		})
	})
}

func TestAnalysisExecuteImmediateJob(t *testing.T) {
	Convey("executeImmediateJob", t, func() {
		Convey("analysis service is normal, and http post returns success", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mock.NewMockMsmqClient(ctrl)
			h := mock.NewMockHTTPClient(ctrl)
			a := mock.NewMockOAuthClient(ctrl)
			pa := newAnalysis(m, h, a)

			a.EXPECT().GetSecret().AnyTimes().Return("")
			a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

			id := []string{
				uuid.NewV4().String(),
				uuid.NewV4().String(),
			}
			executeID := uuid.NewV4().String()
			beginTime := time.Now().Format(time.RFC3339)
			job := common.JobInfo{
				JobID:   id[0],
				JobName: "test",
				JobType: common.PERIODICITY,
				Enabled: true,
				Context: common.JobContext{
					ExecuteID: executeID,
				},
			}
			status := common.JobStatus{}

			pa.mapJobStatus.Set(executeID, common.JobStatus{
				ExecuteID: executeID,
				JobID:     id[0],
				JobStatus: common.EXECUTING,
				ExtInfo: map[string]interface{}{
					common.IsDeleted: 0,
				},
				ExecuteTimes: 1,
			})
			pa.mapJobInfo.Set(id[0], common.JobInfo{JobID: id[0]})
			pa.mapJobInfo.Set(id[1], common.JobInfo{JobID: id[1]})

			h.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(url string, headers map[string]string, reqParam interface{}, respParam interface{}) error {
					status.ExecuteID = executeID
					status.Executor = append(status.Executor, map[string]interface{}{
						"executor_id":  serviceID,
						"execute_time": beginTime,
					})
					status.BeginTime = beginTime
					status.JobID = job.JobID
					status.JobName = job.JobName
					status.JobType = job.JobType
					status.JobStatus = common.EXECUTING
					status.ExecuteTimes = 1
					status.ExtInfo = map[string]interface{}{
						common.IsDeleted: 0,
					}

					body, _ := jsoniter.Marshal(status)
					return jsoniter.Unmarshal(body, respParam)
				},
			)

			pa.executeImmediateJob(job)
			after, ok := pa.mapJobStatus.Get(executeID)
			assert.Equal(t, ok, true)
			assert.Equal(t, after.(common.JobStatus).ExecuteTimes, 2)
		})
	})
}

func TestAnalysisRefreshJob(t *testing.T) {
	Convey("refreshJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		id := []string{
			0: uuid.NewV4().String(),
			1: uuid.NewV4().String(),
		}
		newID := uuid.NewV4().String()
		pa.mapJobInfo.Set(id[0], common.JobInfo{
			JobID:   id[0],
			JobName: "test",
			JobType: common.TIMING,
			Enabled: false,
			Context: common.JobContext{
				Notify: common.JobNotify{
					Webhook: "https://123",
				},
			},
		})
		pa.mapJobInfo.Set(id[1], common.JobInfo{
			JobID:   id[1],
			JobName: "test1",
			JobType: common.TIMING,
			Enabled: false,
			Context: common.JobContext{
				Notify: common.JobNotify{
					Webhook: "https://456",
				},
			},
		})

		Convey("insert", func() {
			jobMsg := common.JobMsg{
				Method: common.CREATE,
				Data: common.JobInfo{
					JobID:   newID,
					JobName: "newJob",
				},
			}
			msg, _ := jsoniter.Marshal(jobMsg)
			_ = pa.refreshJob(msg)
			time.Sleep(time.Second * 1)
			after, ok := pa.mapJobInfo.Get(newID)
			assert.Equal(t, ok, true)
			assert.Equal(t, after.(common.JobInfo).JobName, "newJob")
		})

		Convey("update, full volume updated", func() {
			jobMsg := common.JobMsg{
				Method: common.UPDATE,
				Data: common.JobInfo{
					JobID:   id[1],
					JobName: "test2",
				},
			}
			msg, _ := jsoniter.Marshal(jobMsg)
			_ = pa.refreshJob(msg)
			time.Sleep(time.Second * 1)
			after, ok := pa.mapJobInfo.Get(id[1])
			assert.Equal(t, after.(common.JobInfo), common.JobInfo{JobID: id[1], JobName: "test2"})
			assert.Equal(t, ok, true)
		})

		Convey("delete", func() {
			jobMsg := common.JobMsg{
				Method: common.DELETE,
				Data: common.JobInfo{
					JobID: newID,
				},
			}
			msg, _ := jsoniter.Marshal(jobMsg)
			_ = pa.refreshJob(msg)
			time.Sleep(time.Second * 1)
			_, ok := pa.mapJobInfo.Get(newID)
			assert.Equal(t, ok, false)
		})

		Convey("enable, modify enabled attribute", func() {
			jobMsg := common.JobMsg{
				Method: common.ENABLE,
				Data: common.JobInfo{
					JobID:   id[0],
					JobName: "test2",
					Enabled: true,
				},
			}
			msg, _ := jsoniter.Marshal(jobMsg)
			_ = pa.refreshJob(msg)
			time.Sleep(time.Second * 1)
			after, ok := pa.mapJobInfo.Get(id[0])
			assert.Equal(t, after.(common.JobInfo).JobName, "test")
			assert.Equal(t, after.(common.JobInfo).Enabled, true)
			assert.Equal(t, ok, true)
		})

		Convey("notify, modify notify attribute", func() {
			jobMsg := common.JobMsg{
				Method: common.NOTIFY,
				Data: common.JobInfo{
					JobID:   id[0],
					JobName: "test2",
					Enabled: true,
					Context: common.JobContext{
						Notify: common.JobNotify{
							Webhook: "https://456",
						},
					},
				},
			}
			msg, _ := jsoniter.Marshal(jobMsg)
			_ = pa.refreshJob(msg)
			time.Sleep(time.Second * 1)
			after, ok := pa.mapJobInfo.Get(id[0])
			assert.Equal(t, after.(common.JobInfo).JobName, "test")
			assert.Equal(t, after.(common.JobInfo).Enabled, false)
			assert.Equal(t, after.(common.JobInfo).Context.Notify.Webhook, "https://456")
			assert.Equal(t, ok, true)
		})
	})
}

func TestAnalysisInsertJob(t *testing.T) {
	Convey("insertJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		id := uuid.NewV4().String()
		pa.mapJobInfo.Set(id, common.JobInfo{
			JobID:   id,
			JobName: "test",
			JobType: common.TIMING,
			Enabled: false,
		})

		Convey("job does not exist, then add a new job", func() {
			newID := uuid.NewV4().String()
			job := common.JobInfo{
				JobID:   newID,
				JobName: "test2",
				JobType: common.PERIODICITY,
				Enabled: true,
			}
			pa.insertJob(job)
			_, ok := pa.mapJobInfo.Get(newID)
			assert.Equal(t, ok, true)
		})

		Convey("job does exist, nothing has been changed", func() {
			before, _ := pa.mapJobInfo.Get(id)
			job := common.JobInfo{
				JobID:   id,
				JobName: "test2",
				JobType: common.PERIODICITY,
				Enabled: true,
			}
			pa.insertJob(job)
			after, _ := pa.mapJobInfo.Get(id)
			assert.Equal(t, before, after)
			assert.NotEqual(t, job, after)
		})
	})
}

func TestAnalysisUpdateJob(t *testing.T) {
	Convey("updateJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		id := uuid.NewV4().String()
		pa.mapJobInfo.Set(id, common.JobInfo{
			JobID:   id,
			JobName: "test",
			JobType: common.TIMING,
			Enabled: false,
		})

		Convey("job does not exist, nothing has been changed", func() {
			before, _ := pa.mapJobInfo.Get(id)
			job := common.JobInfo{
				JobID:   uuid.NewV4().String(),
				JobName: "test2",
				JobType: common.PERIODICITY,
				Enabled: true,
			}
			pa.updateJob(job)
			after, _ := pa.mapJobInfo.Get(id)
			assert.Equal(t, before, after)
		})

		Convey("job does exist, then full volume update it", func() {
			before, _ := pa.mapJobInfo.Get(id)
			job := common.JobInfo{
				JobID:   id,
				JobName: "test2",
				JobType: common.PERIODICITY,
				Enabled: true,
			}
			pa.updateJob(job)
			after, _ := pa.mapJobInfo.Get(id)
			assert.NotEqual(t, before, after)
			assert.Equal(t, job, after)
		})
	})
}

func TestAnalysisDeleteJob(t *testing.T) {
	Convey("deleteJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		id := uuid.NewV4().String()
		pa.mapJobInfo.Set(id, common.JobInfo{
			JobID:   id,
			JobName: "test",
			JobType: common.TIMING,
			Enabled: false,
		})

		Convey("job does not exist, nothing has been changed", func() {
			job := common.JobInfo{
				JobID:   uuid.NewV4().String(),
				JobName: "test",
				JobType: common.TIMING,
				Enabled: true,
			}
			pa.deleteJob(job)
			_, ok := pa.mapJobInfo.Get(id)
			assert.Equal(t, ok, true)
		})

		Convey("job does exist, then delete it", func() {
			job := common.JobInfo{
				JobID: id,
			}
			pa.deleteJob(job)
			_, ok := pa.mapJobInfo.Get(id)
			assert.Equal(t, ok, false)
		})
	})
}

func TestAnalysisEnableJob(t *testing.T) {
	Convey("enableJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		id := uuid.NewV4().String()
		pa.mapJobInfo.Set(id, common.JobInfo{
			JobID:   id,
			JobName: "test",
			JobType: common.TIMING,
			Enabled: false,
		})

		Convey("job does not exist, nothing has been changed", func() {
			before, _ := pa.mapJobInfo.Get(id)
			job := common.JobInfo{
				JobID:   uuid.NewV4().String(),
				JobName: "test",
				JobType: common.TIMING,
				Enabled: true,
			}
			pa.enableJob(job)
			after, _ := pa.mapJobInfo.Get(id)
			assert.Equal(t, before, after)
		})

		Convey("job does exist, just modify enabled attribute", func() {
			before, _ := pa.mapJobInfo.Get(id)
			job := common.JobInfo{
				JobID:   id,
				JobName: "test",
				JobType: common.PERIODICITY,
				Enabled: true,
			}
			pa.enableJob(job)
			v := before.(common.JobInfo)
			v.Enabled = true

			after, _ := pa.mapJobInfo.Get(id)
			assert.Equal(t, v, after)
		})
	})
}

func TestAnalysisNotifyJob(t *testing.T) {
	Convey("notifyJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		id := uuid.NewV4().String()
		pa.mapJobInfo.Set(id, common.JobInfo{
			JobID:   id,
			JobName: "test",
			JobType: common.TIMING,
			Enabled: false,
			Context: common.JobContext{
				Notify: common.JobNotify{
					Webhook: "https://123",
				},
			},
		})

		Convey("job does not exist, nothing has been changed", func() {
			before, _ := pa.mapJobInfo.Get(id)
			job := common.JobInfo{
				JobID:   uuid.NewV4().String(),
				JobName: "test",
				JobType: common.TIMING,
				Enabled: true,
				Context: common.JobContext{
					Notify: common.JobNotify{
						Webhook: "https://456",
					},
				},
			}
			pa.notifyJob(job)
			after, _ := pa.mapJobInfo.Get(id)
			assert.Equal(t, before, after)
		})

		Convey("job does exist, just modify notify's address", func() {
			before, _ := pa.mapJobInfo.Get(id)
			job := common.JobInfo{
				JobID:   id,
				JobName: "test",
				JobType: common.PERIODICITY,
				Enabled: true,
				Context: common.JobContext{
					Notify: common.JobNotify{
						Webhook: "https://789",
					},
				},
			}
			pa.notifyJob(job)

			b := before.(common.JobInfo)
			b.Context.Notify.Webhook = "https://789"

			after, _ := pa.mapJobInfo.Get(id)
			assert.Equal(t, b, after)
		})
	})
}

func TestAnalysisRemoveCronJob(t *testing.T) {
	Convey("removeCronJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		id := uuid.NewV4().String()
		pa.mapEntryID.Set(id, cron.EntryID(123))

		Convey("job does not exist, nothing has been changed", func() {
			before, _ := pa.mapEntryID.Get(id)
			pa.removeCronJob(uuid.NewV4().String())
			after, _ := pa.mapEntryID.Get(id)
			assert.Equal(t, before, after)
		})

		Convey("job does exist, remove cron job", func() {
			pa.removeCronJob(id)
			_, ok := pa.mapEntryID.Get(id)
			assert.Equal(t, ok, false)
		})
	})
}

func TestAnalysisDeleteJobStatus(t *testing.T) {
	Convey("deleteJobStatus", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		jobID := uuid.NewV4().String()
		executeID := uuid.NewV4().String()
		pa.mapJobStatus.Set(executeID, common.JobStatus{
			ExecuteID: executeID,
			JobID:     jobID,
			JobStatus: common.EXECUTING,
			ExtInfo: map[string]interface{}{
				common.IsDeleted: 0,
			},
		})

		Convey("job does not exist, nothing has been changed", func() {
			before, _ := pa.mapJobStatus.Get(executeID)
			pa.deleteJobStatus(uuid.NewV4().String())
			after, _ := pa.mapJobStatus.Get(executeID)
			assert.Equal(t, before, after)
			assert.Equal(t, after.(common.JobStatus).ExtInfo[common.IsDeleted].(int), 0)
		})

		Convey("job does exist, change job status.", func() {
			pa.deleteJobStatus(jobID)
			after, _ := pa.mapJobStatus.Get(executeID)
			fmt.Println(after)
			assert.Equal(t, after.(common.JobStatus).ExtInfo[common.IsDeleted].(int), 1)
		})
	})
}

func TestAnalysisRefreshStatus(t *testing.T) {
	Convey("refreshStatus", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, nil, a)
		pa.initHTTPClient()

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		jobID := uuid.NewV4().String()
		executeID := uuid.NewV4().String()
		beginTime := time.Now().Format(time.RFC3339)
		endTime := time.Now().Add(time.Hour * 1).Format(time.RFC3339)
		pa.mapJobStatus.Set(executeID, common.JobStatus{
			ExecuteID: executeID,
			JobID:     jobID,
			JobStatus: common.EXECUTING,
			ExtInfo: map[string]interface{}{
				common.IsDeleted: 0,
			},
			BeginTime: beginTime,
		})

		Convey("refresh a uncompleted job status", func() {
			status := common.JobStatus{
				JobStatus: common.EXECUTING,
			}
			message, _ := jsoniter.Marshal(status)
			_ = pa.refreshStatus(message)
			time.Sleep(time.Second * 1)
			after, _ := pa.mapJobStatus.Get(executeID)
			assert.Equal(t, after.(common.JobStatus).EndTime, "")
		})

		Convey("refresh a commpleted job status", func() {
			status := common.JobStatus{
				ExecuteID: executeID,
				JobStatus: common.SUCCESS,
				EndTime:   endTime,
			}
			message, _ := jsoniter.Marshal(status)
			_ = pa.refreshStatus(message)
			time.Sleep(time.Second * 1)
			after, _ := pa.mapJobStatus.Get(executeID)
			assert.Equal(t, after.(common.JobStatus).EndTime, endTime)
			assert.Equal(t, after.(common.JobStatus).JobStatus, common.SUCCESS)
		})
	})
}

func TestAnalysisHandleStatus(t *testing.T) {
	Convey("handleStatus", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		h := mock.NewMockHTTPClient(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		pa := newAnalysis(m, h, a)

		a.EXPECT().GetSecret().AnyTimes().Return("")
		a.EXPECT().GetCode(gomock.Any()).AnyTimes().Return("", nil)

		h.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

		jobID := uuid.NewV4().String()
		beginTime := time.Now().Format(time.RFC3339)
		endTime := time.Now().Add(time.Hour * 1).Format(time.RFC3339)
		executeID := make([]string, 0)
		status := make([]interface{}, 0)
		for i := 0; i < 1200; i++ {
			executeID = append(executeID, uuid.NewV4().String())
			status = append(status, common.JobStatus{
				ExecuteID: executeID[i],
				JobID:     jobID,
				JobStatus: common.EXECUTING,
				ExtInfo: map[string]interface{}{
					common.IsDeleted: 0,
				},
				BeginTime: beginTime,
				EndTime:   endTime,
			})
		}
		pa.mapJobStatus.Set(executeID[0], status)
		pa.mapJobStatus.Set(executeID[1001], status)

		Convey("less than 1000", func() {
			mapStatus := make(map[string]interface{})
			for i := 0; i < 10; i++ {
				mapStatus[executeID[i]] = status[i]
			}

			pa.handleStatus(mapStatus)

			_, ok := pa.mapJobStatus.Get(executeID[0])
			assert.Equal(t, ok, false)
		})

		Convey("more than 1000", func() {
			mapStatus := make(map[string]interface{})
			for i := 0; i < 1100; i++ {
				mapStatus[executeID[i]] = status[i]
			}
			pa.handleStatus(mapStatus)

			_, ok := pa.mapJobStatus.Get(executeID[1001])
			assert.Equal(t, ok, false)
		})
	})
}
