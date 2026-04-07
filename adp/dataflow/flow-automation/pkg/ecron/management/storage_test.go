package management

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/mock"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils"
	. "github.com/smartystreets/goconvey/convey"
)

func newManagement(m utils.MsmqClient, d utils.DBClient, e Executor, a utils.OAuthClient, h utils.HTTPServerStarter) *management {
	return &management{
		mapRequest:        make(map[string][]map[string]func(c *gin.Context)),
		msmqClient:        m,
		dbClient:          d,
		executor:          e,
		authClient:        a,
		httpServerStarter: h,
		chJobMsg:          make(chan common.JobMsg, 1),
		chJobStatus:       make(chan common.JobStatus, 1),
		chJobImmediate:    make(chan common.JobInfo, 1),
	}
}

func setGinMode(t *testing.T) func(*testing.T) {
	old := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func(t *testing.T) {
		gin.SetMode(old)
	}
}

func TestNewManagementService(t *testing.T) {
	Convey("NewManagementService", t, func() {
		service := NewManagementService()
		assert.NotEqual(t, service, nil)
	})
}

func TestManagementStart(t *testing.T) {
	Convey("Start", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		h := mock.NewMockHTTPServerStarter(ctrl)
		timer := newManagement(m, d, e, a, h)

		d.EXPECT().Upgrade().AnyTimes().Return()
		d.EXPECT().Ping().AnyTimes().Return(nil)
		d.EXPECT().Connect().AnyTimes().Return(nil)
		a.EXPECT().VerifyHydraVersion().AnyTimes().Return("/admin/oauth2/introspect", true)
		h.EXPECT().Start(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

		Convey("multimode is true", func() {
			cronSvr.MultiNode = true
			timer.Start()
		})

		Convey("multimode is false", func() {
			cronSvr.MultiNode = false
			timer.Start()
		})
	})
}

func TestManagementStop(t *testing.T) {
	Convey("Stop", t, func() {
		Convey("uninitialized", func() {
			timer := newManagement(nil, nil, nil, nil, nil)
			timer.Stop()
		})

		Convey("initialized", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mock.NewMockMsmqClient(ctrl)
			d := mock.NewMockDBClient(ctrl)
			e := mock.NewMockExecutor(ctrl)
			a := mock.NewMockOAuthClient(ctrl)
			timer := newManagement(m, d, e, a, nil)

			d.EXPECT().Release().AnyTimes()
			a.EXPECT().Release().AnyTimes()
			timer.Stop()
		})
	})
}

func TestManagementGetFreePort(t *testing.T) {
	Convey("getFreePort", t, func() {
		timer := newManagement(nil, nil, nil, nil, nil)
		port, err := timer.getFreePort()
		assert.Equal(t, err, nil)
		assert.NotEqual(t, port, 0)
	})
}

func TestManagementInit(t *testing.T) {
	Convey("init", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)

		d.EXPECT().Connect().AnyTimes().Return(nil)
		d.EXPECT().Ping().AnyTimes().Return(nil)
		d.EXPECT().Upgrade().AnyTimes().Return()
		a.EXPECT().VerifyHydraVersion().AnyTimes().Return("/admin/oauth2/introspect", true)

		timer.init()
	})
}

func TestManagementPublishJobMsg(t *testing.T) {
	Convey("publishJobMsg", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)

		m.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		timer.publishJobMsg(common.JobMsg{
			Method: common.CREATE,
			Data: common.JobInfo{
				JobID: uuid.NewV4().String(),
			},
		})
	})
}

func TestManagementgPublishJobImmediate(t *testing.T) {
	Convey("publishJobImmediate", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)
		m.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
		timer.publishJobImmediate(common.JobInfo{
			JobID: uuid.NewV4().String(),
		})
	})
}

func TestManagementgPublishJobStatus(t *testing.T) {
	Convey("publishJobStatus", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)
		m.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

		timer.publishJobStatus(common.JobStatus{
			JobID: uuid.NewV4().String(),
		})
	})
}

func TestManagementIsEcronDBAvailable(t *testing.T) {
	Convey("isEcronDBAvailable", t, func() {
		Convey("uninitialized", func() {
			pt := newManagement(nil, nil, nil, nil, nil)
			err := pt.isEcronDBAvailable()
			assert.Equal(t, err.Cause, common.ErrDataBaseUnavailable)
			assert.Equal(t, err.Code, common.InternalError)
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)

		Convey("initialized", func() {
			timer := newManagement(m, d, e, a, nil)
			d.EXPECT().Ping().AnyTimes().Return(nil)
			err := timer.isEcronDBAvailable()
			assert.Equal(t, err, (*common.ECronError)(nil))
		})
	})
}

func TestManagementCode(t *testing.T) {
	Convey("code", t, func() {
		timer := newManagement(nil, nil, nil, nil, nil)

		Convey("code, err is nil, created is true", func() {
			c := timer.code(nil, true)
			assert.Equal(t, c, 201)
		})

		Convey("code, err is nil, created is false", func() {
			c := timer.code(nil, false)
			assert.Equal(t, c, 200)
		})

		Convey("code, err is not nil, crreated is false, and code length is less than 3", func() {
			c := timer.code(&common.ECronError{Code: 12}, false)
			assert.Equal(t, c, 500)
		})

		Convey("code, err is not nil, created is false, and code length is normal", func() {
			c := timer.code(&common.ECronError{Code: 400001}, false)
			assert.Equal(t, c, 400)
		})
	})
}

func TestManagementAuth(t *testing.T) {
	Convey("auth", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)

		Convey("normal", func() {
			timer := newManagement(m, d, e, a, nil)

			a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
			a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			_, err := timer.auth(fmt.Sprintf("%v %v", common.Bearer, "123456"), "", "")
			assert.Equal(t, err, (*common.ECronError)(nil))
		})

		Convey("authClient is null", func() {
			timer := newManagement(m, d, e, nil, nil)
			_, err := timer.auth("123456", "", "")
			assert.Equal(t, err.Cause, common.ErrAuthClientUnavailable)
		})
	})
}

func TestManagementExecuteJob(t *testing.T) {
	Convey("executeJob failed", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)
		e.EXPECT().ExecuteJob(gomock.Any()).AnyTimes().Return(true, errors.New("failed"))
		id := uuid.NewV4().String()
		executeID := uuid.NewV4().String()
		timer.executeJob(common.JobInfo{
			JobID:   id,
			JobType: common.TIMING,
			Context: common.JobContext{
				ExecuteID: executeID,
			},
		})

		time.Sleep(time.Second)
		immediateJob := <-timer.chJobImmediate
		status := <-timer.chJobStatus
		assert.Equal(t, immediateJob.Context.ExecuteID, executeID)
		assert.Equal(t, immediateJob.JobID, id)
		assert.Equal(t, status.ExecuteID, executeID)
		assert.Equal(t, status.JobStatus, common.FAILURE)
	})
}

func TestManagementHandleWebhook(t *testing.T) {
	Convey("handleWebhook", t, func() {
		id := uuid.NewV4().String()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)

		timer.handleWebhook(id, map[string]interface{}{"hello": "world"})
		time.Sleep(time.Second)
		status := <-timer.chJobStatus
		assert.Equal(t, status.ExecuteID, id)
		assert.Equal(t, status.JobStatus, common.SUCCESS)
		assert.Equal(t, status.ExtInfo["result"], map[string]interface{}{"hello": "world"})
	})
}

func TestManagementGetJobTotal(t *testing.T) {
	Convey("getJobTotal", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		beginTime := time.Now().Format(time.RFC3339)
		duration, _ := time.ParseDuration("1h")
		endTime := time.Now().Add(duration).Format(time.RFC3339)
		//id := uuid.NewV4().String()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)

		d.EXPECT().Ping().AnyTimes().Return(nil)

		Convey("check params", func() {
			r.GET(jobTotalPATH, func(c *gin.Context) {
				params := common.JobTotalQueryParams{}
				params.BeginTime = c.Query("begin_at")
				params.EndTime = c.Query("end_at")
				timer.response(params, nil, false, c)
			})

			reqParam := common.JobTotalQueryParams{
				BeginTime: beginTime,
				EndTime:   endTime,
			}
			target := fmt.Sprintf("%v?begin_at=%v&end_at=%v",
				jobTotalPATH, url.QueryEscape(reqParam.BeginTime), url.QueryEscape(reqParam.EndTime))
			req := httptest.NewRequest("GET", target, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			respParam := common.JobTotalQueryParams{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.Equal(t, reqParam, respParam)
		})

		Convey("check result", func() {
			r.GET(jobTotalPATH, timer.getJobTotal)

			d.EXPECT().GetJobTotal(gomock.Any(), gomock.Any()).AnyTimes().Return(common.JobTotal{Total: 100, TimeStamp: beginTime}, nil)
			a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
			a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			reqParam := common.JobTotalQueryParams{
				BeginTime: beginTime,
				EndTime:   endTime,
			}
			target := fmt.Sprintf("%v?begin_at=%v&end_at=%v",
				jobTotalPATH, url.QueryEscape(reqParam.BeginTime), url.QueryEscape(reqParam.EndTime))
			req := httptest.NewRequest("GET", target, nil)
			req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			respParam := common.JobTotal{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.Equal(t, common.JobTotal{Total: 100, TimeStamp: beginTime}, respParam)
		})
	})
}

func TestManagementGetJobInfo(t *testing.T) {
	Convey("getJobInfo", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)
		id := uuid.NewV4().String()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)

		d.EXPECT().Ping().AnyTimes().Return(nil)

		Convey("check params", func() {
			r.GET(jobInfoPATH, func(c *gin.Context) {
				params := common.JobInfoQueryParams{}
				params.Limit, _ = strconv.Atoi(c.Query("limit"))
				params.Page, _ = strconv.Atoi(c.Query("page"))
				params.TimeStamp = c.Query("timestamp")
				params.JobID = c.QueryArray("job_id")
				params.JobType = c.Query("job_type")

				if 0 == params.Limit && 0 == params.Page && 0 == len(params.JobID) {
					params.Limit = 10
					params.Page = 1
				}

				timer.response(params, nil, false, c)
			})

			Convey("with all params", func() {
				reqParam := common.JobInfoQueryParams{
					Limit:     10,
					Page:      1,
					TimeStamp: beginTime,
					JobID: []string{
						0: uuid.NewV4().String(),
						1: uuid.NewV4().String(),
					},
					JobType: common.TIMING,
				}
				target := fmt.Sprintf("%v?limit=%v&page=%v&timestamp=%v&job_id=%v&job_id=%v&job_type=%v",
					jobInfoPATH, reqParam.Limit, reqParam.Page, url.QueryEscape(reqParam.TimeStamp), reqParam.JobID[0], reqParam.JobID[1], reqParam.JobType)
				req := httptest.NewRequest("GET", target, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				respParam := common.JobInfoQueryParams{}
				err := jsoniter.Unmarshal(respBody, &respParam)
				assert.Equal(t, nil, err)
				assert.Equal(t, reqParam, respParam)
			})

			Convey("without limit and page, but length of job_id is above 0. just ignore them", func() {
				reqParam := common.JobInfoQueryParams{
					TimeStamp: beginTime,
					JobID: []string{
						0: uuid.NewV4().String(),
						1: uuid.NewV4().String(),
					},
					JobType: common.TIMING,
				}
				target := fmt.Sprintf("%v?timestamp=%v&job_id=%v&job_id=%v&job_type=%v",
					jobInfoPATH, url.QueryEscape(reqParam.TimeStamp), reqParam.JobID[0], reqParam.JobID[1], reqParam.JobType)
				req := httptest.NewRequest("GET", target, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				respParam := common.JobInfoQueryParams{}
				err := jsoniter.Unmarshal(respBody, &respParam)
				assert.Equal(t, nil, err)
				assert.Equal(t, reqParam, respParam)
			})

			Convey("without limit and page, and length of job_id is also 0, default value of them will be selected", func() {
				reqParam := common.JobInfoQueryParams{
					Limit:     10,
					Page:      1,
					TimeStamp: beginTime,
					JobType:   common.TIMING,
					JobID:     []string(nil),
				}
				target := fmt.Sprintf("%v?timestamp=%v&job_type=%v",
					jobInfoPATH, url.QueryEscape(reqParam.TimeStamp), reqParam.JobType)
				req := httptest.NewRequest("GET", target, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				respParam := common.JobInfoQueryParams{}
				err := jsoniter.Unmarshal(respBody, &respParam)
				assert.Equal(t, nil, err)
				assert.Equal(t, reqParam, respParam)
			})

			Convey("only job_id", func() {
				reqParam := common.JobInfoQueryParams{
					JobID: []string{
						0: uuid.NewV4().String(),
						1: uuid.NewV4().String(),
					},
				}
				target := fmt.Sprintf("%v?job_id=%v&job_id=%v",
					jobInfoPATH, reqParam.JobID[0], reqParam.JobID[1])
				req := httptest.NewRequest("GET", target, nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				respParam := common.JobInfoQueryParams{}
				err := jsoniter.Unmarshal(respBody, &respParam)
				assert.Equal(t, nil, err)
				assert.Equal(t, reqParam, respParam)
			})
		})

		Convey("check result", func() {
			r.GET(jobInfoPATH, timer.getJobInfo)

			d.EXPECT().GetJob(gomock.Any(), gomock.Any()).AnyTimes().Return([]common.JobInfo{0: {JobID: id}}, nil)

			Convey("query jobs by id", func() {
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := common.JobInfoQueryParams{
					JobID: []string{
						0: id,
					},
				}
				target := fmt.Sprintf("%v?job_id=%v",
					jobInfoPATH, reqParam.JobID[0])
				req := httptest.NewRequest("GET", target, nil)
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				respParam := []common.JobInfo{}
				err := jsoniter.Unmarshal(respBody, &respParam)
				assert.Equal(t, nil, err)
				assert.Equal(t, len(respParam), 1)
				assert.Equal(t, respParam[0].JobID, id)
			})

			Convey("query jobs by page", func() {
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				target := fmt.Sprintf("%v?limit=0&page=0", jobInfoPATH)
				req := httptest.NewRequest("GET", target, nil)
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				respParam := []common.JobInfo{}
				err := jsoniter.Unmarshal(respBody, &respParam)
				assert.Equal(t, nil, err)
				assert.Equal(t, len(respParam), 1)
				assert.Equal(t, respParam[0].JobID, id)
			})
		})
	})
}

func TestManagementGetJobStatus(t *testing.T) {
	Convey("getJobStatus", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		beginTime := time.Now().Format(time.RFC3339)
		duration, _ := time.ParseDuration("1h")
		endTime := time.Now().Add(duration).Format(time.RFC3339)
		id := uuid.NewV4().String()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)

		d.EXPECT().Ping().AnyTimes().Return(nil)

		Convey("check params", func() {
			r.GET(jobStatusPATH, func(c *gin.Context) {
				params := common.JobStatusQueryParams{}
				params.JobID = c.Query("job_id")
				params.JobType = c.Query("job_type")
				params.JobStatus = c.Query("job_status")
				params.BeginTime = c.Query("begin_at")
				params.EndTime = c.Query("end_at")
				timer.response(params, nil, false, c)
			})

			reqParam := common.JobStatusQueryParams{
				JobID:     uuid.NewV4().String(),
				JobType:   common.TIMING,
				JobStatus: common.SUCCESS,
				BeginTime: beginTime,
				EndTime:   endTime,
			}
			target := fmt.Sprintf("%v?job_id=%v&job_type=%v&job_status=%v&begin_at=%v&end_at=%v",
				jobStatusPATH, reqParam.JobID, reqParam.JobType, reqParam.JobStatus, url.QueryEscape(reqParam.BeginTime), url.QueryEscape(reqParam.EndTime))
			req := httptest.NewRequest("GET", target, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			respParam := common.JobStatusQueryParams{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.Equal(t, reqParam, respParam)
		})

		Convey("check result", func() {
			r.GET(jobStatusPATH, timer.getJobStatus)

			d.EXPECT().GetJobStatus(gomock.Any(), gomock.Any()).AnyTimes().Return([]common.JobStatus{0: {JobID: id}}, nil)
			a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
			a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			reqParam := common.JobStatusQueryParams{
				JobID:     id,
				JobType:   common.TIMING,
				JobStatus: common.SUCCESS,
				BeginTime: beginTime,
				EndTime:   endTime,
			}
			target := fmt.Sprintf("%v?job_id=%v&job_type=%v&job_status=%v&begin_at=%v&end_at=%v",
				jobStatusPATH, reqParam.JobID, reqParam.JobType, reqParam.JobStatus, url.QueryEscape(reqParam.BeginTime), url.QueryEscape(reqParam.EndTime))
			req := httptest.NewRequest("GET", target, nil)
			req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			respParam := []common.JobStatus{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.Equal(t, len(respParam), 1)
			assert.Equal(t, respParam[0].JobID, id)
		})
	})
}

func TestManagementPostJobInfo(t *testing.T) {
	Convey("postJobInfo", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)
		id := uuid.NewV4().String()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)
		stubNSQConnected(true, t)

		d.EXPECT().Ping().AnyTimes().Return(nil)

		Convey("check params", func() {
			r.POST(jobInfoPATH, func(c *gin.Context) {
				job := common.JobInfo{}
				err := timer.request(c, &job)
				assert.Equal(t, (*common.ECronError)(nil), err)
				job.JobID = id
				job.CreateTime = beginTime
				job.UpdateTime = beginTime
				timer.response(job, nil, true, c)
			})

			reqParam := common.JobInfo{
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
			}
			reqParamByte, _ := jsoniter.Marshal(reqParam)

			target := fmt.Sprintf("%v", jobInfoPATH)
			req := httptest.NewRequest("POST", target, bytes.NewReader(reqParamByte))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusCreated)

			reqParam.JobID = id
			reqParam.CreateTime = beginTime
			reqParam.UpdateTime = beginTime

			respParam := common.JobInfo{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.Equal(t, reqParam, respParam)
		})

		Convey("check result", func() {
			r.POST(jobInfoPATH, timer.postJobInfo)

			d.EXPECT().InsertJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
			a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			reqParam := common.JobInfo{
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
			}
			reqParamByte, _ := jsoniter.Marshal(reqParam)

			target := fmt.Sprintf("%v", jobInfoPATH)
			req := httptest.NewRequest("POST", target, bytes.NewReader(reqParamByte))
			req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusCreated)

			reqParam.JobID = id
			reqParam.CreateTime = beginTime
			reqParam.UpdateTime = beginTime

			respParam := common.JobInfo{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.NotEqual(t, len(respParam.JobID), 0)
		})
		Convey("check nsq disconnected", func() {
			r.POST(jobInfoPATH, timer.postJobInfo)

			d.EXPECT().InsertJob(gomock.Any(), gomock.Any()).AnyTimes().Return(utils.NewECronError(
				common.ErrMSMQClientUnavailable, common.InternalError, nil))
			a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
			a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			reqParam := common.JobInfo{
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
			}
			reqParamByte, _ := jsoniter.Marshal(reqParam)

			stubNSQConnected(false, t)
			target := fmt.Sprintf("%v", jobInfoPATH)
			req := httptest.NewRequest("POST", target, bytes.NewReader(reqParamByte))
			req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
		})
	})
}

func TestManagementPostJobExecution(t *testing.T) {
	Convey("postJobExecution", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		//beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)
		//id := uuid.NewV4().String()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)

		d.EXPECT().Ping().AnyTimes().Return(nil)

		r.POST(jobExecutionPATH, timer.postJobExecution)

		Convey("unsupported execution mode", func() {
			d.EXPECT().CheckJobExecuteMode(gomock.Any()).AnyTimes().Return(utils.NewECronError(
				common.ErrUnsupportedExecutionMode, common.InternalError, nil))
			a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
			a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			reqParam := common.JobInfo{
				JobID:       uuid.NewV4().String(),
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
				Context: common.JobContext{
					Mode: "",
				},
			}
			reqParamByte, _ := jsoniter.Marshal(reqParam)

			target := fmt.Sprintf("%v", jobExecutionPATH)
			req := httptest.NewRequest("POST", target, bytes.NewReader(reqParamByte))
			req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusInternalServerError)

			respParam := common.ECronError{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.Equal(t, respParam.Cause, common.ErrUnsupportedExecutionMode)
			assert.Equal(t, respParam.Code, common.InternalError)
		})

		Convey("normal job execution, return job status", func() {
			d.EXPECT().CheckJobExecuteMode(gomock.Any()).AnyTimes().Return(nil)
			e.EXPECT().ExecuteJob(gomock.Any()).AnyTimes().Return(false, nil)
			a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
			a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			reqParam := common.JobInfo{
				JobID:       uuid.NewV4().String(),
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
				Context: common.JobContext{
					Mode: common.HTTP,
				},
			}
			reqParamByte, _ := jsoniter.Marshal(reqParam)

			target := fmt.Sprintf("%v", jobExecutionPATH)
			req := httptest.NewRequest("POST", target, bytes.NewReader(reqParamByte))
			req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()

			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			respParam := common.JobStatus{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.Equal(t, respParam.ExecuteTimes, 1)
		})

		Convey("abnormal job execution, return job status", func() {
			d.EXPECT().CheckJobExecuteMode(gomock.Any()).AnyTimes().Return(nil)
			e.EXPECT().ExecuteJob(gomock.Any()).AnyTimes().Return(false, nil)
			a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
			a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

			executeID := uuid.NewV4().String()
			reqParam := common.JobInfo{
				JobID:       uuid.NewV4().String(),
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
				Context: common.JobContext{
					Mode:      common.HTTP,
					ExecuteID: executeID,
				},
			}
			reqParamByte, _ := jsoniter.Marshal(reqParam)

			target := fmt.Sprintf("%v", jobExecutionPATH)
			req := httptest.NewRequest("POST", target, bytes.NewReader(reqParamByte))
			req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			respParam := common.JobStatus{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.Equal(t, respParam.ExecuteID, executeID)
		})
	})
}

func stubNSQConnected(flag bool, t *testing.T) func(*testing.T) {
	old := nsqConnected
	nsqConnected = flag
	return func(t *testing.T) {
		nsqConnected = old
	}
}

func TestManagementPostWebhook(t *testing.T) {
	Convey("postWebhook", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		//beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)
		id := uuid.NewV4().String()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)

		stubNSQConnected(true, t)

		a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
		a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

		r.POST("/api/webhook", timer.postWebhook)

		reqParam := common.JobStatus{
			JobID:     uuid.NewV4().String(),
			JobName:   "test",
			JobType:   common.TIMING,
			ExecuteID: id,
			JobStatus: common.SUCCESS,
		}
		reqParamByte, _ := jsoniter.Marshal(reqParam)

		target := fmt.Sprintf("%v", "/api/webhook")
		Convey("postWebhook", func() {
			req := httptest.NewRequest("POST", target, bytes.NewReader(reqParamByte))
			req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			//respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)
		})
		Convey("postWebhook nsq disconnected", func() {
			stubNSQConnected(false, t)
			req := httptest.NewRequest("POST", target, bytes.NewReader(reqParamByte))
			req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
		})

	})

}

func TestManagementPutJobInfo(t *testing.T) {
	Convey("putJobInfo", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)
		id := uuid.NewV4().String()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)
		stubNSQConnected(true, t)

		d.EXPECT().Ping().AnyTimes().Return(nil)

		Convey("check params", func() {
			r.PUT(jobInfoWithIDPATH, func(c *gin.Context) {
				job := common.JobInfo{}
				if err := timer.request(c, &job); nil != err {
					timer.response(nil, err, false, c)
					return
				}

				job.JobID = c.Param(pathJobID)
				job.UpdateTime = beginTime

				timer.response(job, nil, false, c)
			})

			reqParam := common.JobInfo{
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
			}
			reqParamByte, _ := jsoniter.Marshal(reqParam)

			target := fmt.Sprintf("%v/%v", jobInfoPATH, id)
			req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			respParam := common.JobInfo{}
			err := jsoniter.Unmarshal(respBody, &respParam)
			assert.Equal(t, nil, err)
			assert.Equal(t, respParam.JobID, id)
			assert.Equal(t, respParam.UpdateTime, beginTime)
		})

		Convey("check result", func() {
			r.PUT(jobInfoWithIDPATH, timer.putJobInfo)

			Convey("update job failed", func() {
				d.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(utils.NewECronError("unknown", common.InternalError, nil))
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := common.JobInfo{
					JobName:     "test",
					JobType:     common.TIMING,
					JobCronTime: "*/10 * * * * ?",
				}
				reqParamByte, _ := jsoniter.Marshal(reqParam)

				target := fmt.Sprintf("%v/%v", jobInfoPATH, id)
				req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
			})

			Convey("update job success", func() {
				d.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := common.JobInfo{
					JobName:     "test",
					JobType:     common.TIMING,
					JobCronTime: "*/10 * * * * ?",
				}
				reqParamByte, _ := jsoniter.Marshal(reqParam)

				target := fmt.Sprintf("%v/%v", jobInfoPATH, id)
				req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				time.Sleep(time.Second)
				msg := <-timer.chJobMsg
				assert.Equal(t, msg.Method, common.UPDATE)
			})
			Convey("update job nsq dis connected", func() {
				d.EXPECT().UpdateJob(gomock.Any(), gomock.Any()).AnyTimes().Return(utils.NewECronError(
					common.ErrMSMQClientUnavailable, common.InternalError, nil))
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := common.JobInfo{
					JobName:     "test",
					JobType:     common.TIMING,
					JobCronTime: "*/10 * * * * ?",
				}
				reqParamByte, _ := jsoniter.Marshal(reqParam)

				stubNSQConnected(false, t)
				target := fmt.Sprintf("%v/%v", jobInfoPATH, id)
				req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
			})
		})
	})
}

func TestManagementPutJobStatus(t *testing.T) {
	Convey("putJobStatus", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		//beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)
		id := []string{
			0: uuid.NewV4().String(),
			1: uuid.NewV4().String(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)

		d.EXPECT().Ping().AnyTimes().Return(nil)
		d.EXPECT().UpdateJobStatus(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, utils.NewECronError("unknown", common.InternalError, nil))
		a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
		a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

		r.PUT(jobStatusWithIDPATH, timer.putJobStatus)

		reqParam := []common.JobStatus{
			0: {
				JobID:     uuid.NewV4().String(),
				JobName:   "test",
				JobType:   common.TIMING,
				ExecuteID: id[0],
				JobStatus: common.SUCCESS,
			},
			1: {
				JobID:     uuid.NewV4().String(),
				JobName:   "test2",
				JobType:   common.TIMING,
				ExecuteID: id[1],
				JobStatus: common.SUCCESS,
			},
		}
		reqParamByte, _ := jsoniter.Marshal(reqParam)

		target := fmt.Sprintf("%v/%v", jobStatusPATH, strings.Join(id, ","))
		req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
		req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		result := w.Result()
		defer result.Body.Close()
		//respBody, _ := ioutil.ReadAll(result.Body)

		assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
	})
}

func TestManagementPutJobEnable(t *testing.T) {
	Convey("putJobEnable", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)
		id := []string{
			0: uuid.NewV4().String(),
			1: uuid.NewV4().String(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)
		stubNSQConnected(true, t)

		d.EXPECT().Ping().AnyTimes().Return(nil)

		Convey("check params", func() {
			r.PUT(jobEnablePATH, func(c *gin.Context) {
				jobID := c.Param("job_id")
				reqBody, _ := ioutil.ReadAll(c.Request.Body)
				defer c.Request.Body.Close()
				enable := jsoniter.Get(reqBody, "enable").ToBool()
				ids := strings.Split(jobID, ",")
				timer.response(gin.H{"job_id": ids, "enable": enable, "update_time": beginTime}, nil, false, c)
			})

			reqParam := gin.H{"enable": false}
			reqParamByte, _ := jsoniter.Marshal(reqParam)

			target := fmt.Sprintf("%v/%v/enable", jobInfoPATH, strings.Join(id, ","))
			req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			ids := jsoniter.Get(respBody, "job_id").GetInterface().([]interface{})
			assert.Equal(t, ids[0].(string), id[0])
			assert.Equal(t, ids[1].(string), id[1])
			assert.Equal(t, jsoniter.Get(respBody, "enable").ToBool(), false)
			assert.Equal(t, jsoniter.Get(respBody, "update_time").ToString(), beginTime)
		})

		Convey("check result", func() {
			r.PUT(jobEnablePATH, timer.putJobEnable)

			Convey("batch job enable failed", func() {
				d.EXPECT().BatchJobEnable(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(utils.NewECronError("unknown", common.InternalError, nil))
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := gin.H{"enable": false}
				reqParamByte, _ := jsoniter.Marshal(reqParam)

				target := fmt.Sprintf("%v/%v/enable", jobInfoPATH, strings.Join(id, ","))
				req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
			})

			Convey("batch job enable success", func() {
				d.EXPECT().BatchJobEnable(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := gin.H{"enable": false}
				reqParamByte, _ := jsoniter.Marshal(reqParam)

				target := fmt.Sprintf("%v/%v/enable", jobInfoPATH, strings.Join(id, ","))
				req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				time.Sleep(time.Second)
				msg := <-timer.chJobMsg
				assert.Equal(t, msg.Method, common.ENABLE)
			})
			Convey("batch job enable nsq disconnected", func() {
				d.EXPECT().BatchJobEnable(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(utils.NewECronError(common.ErrMSMQClientUnavailable, common.InternalError, nil))
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := gin.H{"enable": false}
				reqParamByte, _ := jsoniter.Marshal(reqParam)

				stubNSQConnected(false, t)
				target := fmt.Sprintf("%v/%v/enable", jobInfoPATH, strings.Join(id, ","))
				req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
			})
		})
	})
}

func TestManagementPutJobNotify(t *testing.T) {
	Convey("putJobNotify", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)
		id := []string{
			0: uuid.NewV4().String(),
			1: uuid.NewV4().String(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)
		stubNSQConnected(true, t)

		d.EXPECT().Ping().AnyTimes().Return(nil)

		Convey("check params", func() {
			r.PUT(jobNotifyPATH, func(c *gin.Context) {
				jobID := c.Param("job_id")
				notify := common.JobNotify{}
				if err := timer.request(c, &notify); nil != err {
					timer.response(nil, err, false, c)
					return
				}
				ids := strings.Split(jobID, ",")
				timer.response(gin.H{"job_id": ids, "notify": notify, "update_time": beginTime}, nil, false, c)
			})

			reqParam := common.JobNotify{
				Webhook: "http://hello.webhook",
			}
			reqParamByte, _ := jsoniter.Marshal(reqParam)

			target := fmt.Sprintf("%v/%v/notify", jobInfoPATH, strings.Join(id, ","))
			req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			ids := jsoniter.Get(respBody, "job_id").GetInterface().([]interface{})
			assert.Equal(t, ids[0].(string), id[0])
			assert.Equal(t, ids[1].(string), id[1])

			notify, _ := jsoniter.Marshal(jsoniter.Get(respBody, "notify").GetInterface().(map[string]interface{}))
			assert.Equal(t, notify, reqParamByte)
			assert.Equal(t, jsoniter.Get(respBody, "update_time").ToString(), beginTime)
		})

		Convey("check result", func() {
			r.PUT(jobNotifyPATH, timer.putJobNotify)

			Convey("batch job notify faild", func() {
				d.EXPECT().BatchJobNotify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, utils.NewECronError("unknown", common.InternalError, nil))
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := common.JobNotify{
					Webhook: "http://hello.webhook",
				}
				reqParamByte, _ := jsoniter.Marshal(reqParam)

				target := fmt.Sprintf("%v/%v/notify", jobInfoPATH, strings.Join(id, ","))
				req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
			})

			Convey("batch job notify nsq disconnected", func() {
				d.EXPECT().BatchJobNotify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, utils.NewECronError(common.ErrMSMQClientUnavailable, common.InternalError, nil))
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := common.JobNotify{
					Webhook: "http://hello.webhook",
				}
				reqParamByte, _ := jsoniter.Marshal(reqParam)

				stubNSQConnected(false, t)
				target := fmt.Sprintf("%v/%v/notify", jobInfoPATH, strings.Join(id, ","))
				req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
			})

			Convey("batch job notify success", func() {
				d.EXPECT().BatchJobNotify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return([]string{"111"}, nil)
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				reqParam := common.JobNotify{
					Webhook: "http://hello.webhook",
				}
				reqParamByte, _ := jsoniter.Marshal(reqParam)

				target := fmt.Sprintf("%v/%v/notify", jobInfoPATH, strings.Join(id, ","))
				req := httptest.NewRequest("PUT", target, bytes.NewReader(reqParamByte))
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				time.Sleep(time.Second)
				msg := <-timer.chJobMsg
				assert.Equal(t, msg.Method, common.NOTIFY)
			})
		})
	})
}

func TestManagementDeleteJobInfo(t *testing.T) {
	Convey("deleteJobInfo", t, func() {
		test := setGinMode(t)
		defer test(t)
		r := gin.Default()

		//beginTime := time.Now().Format(time.RFC3339)
		//duration, _ := time.ParseDuration("1h")
		//endTime := time.Now().Add(duration).Format(time.RFC3339)
		id := []string{
			0: uuid.NewV4().String(),
			1: uuid.NewV4().String(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := mock.NewMockMsmqClient(ctrl)
		d := mock.NewMockDBClient(ctrl)
		e := mock.NewMockExecutor(ctrl)
		a := mock.NewMockOAuthClient(ctrl)
		timer := newManagement(m, d, e, a, nil)
		stubNSQConnected(true, t)

		d.EXPECT().Ping().AnyTimes().Return(nil)

		Convey("check params", func() {
			r.DELETE(jobInfoWithIDPATH, func(c *gin.Context) {
				jobID := c.Param("job_id")
				timer.response(gin.H{"job_id": strings.Split(jobID, ",")}, nil, false, c)
			})

			target := fmt.Sprintf("%v/%v", jobInfoPATH, strings.Join(id, ","))
			req := httptest.NewRequest("DELETE", target, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()
			respBody, _ := ioutil.ReadAll(result.Body)

			assert.Equal(t, result.StatusCode, http.StatusOK)

			ids := jsoniter.Get(respBody, "job_id").GetInterface().([]interface{})
			assert.Equal(t, ids[0].(string), id[0])
			assert.Equal(t, ids[1].(string), id[1])
		})

		Convey("check result", func() {
			r.DELETE(jobInfoWithIDPATH, timer.deleteJobInfo)

			Convey("delete job failed", func() {
				d.EXPECT().DeleteJob(gomock.Any(), gomock.Any()).AnyTimes().Return(utils.NewECronError("unknown", common.InternalError, nil))
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				target := fmt.Sprintf("%v/%v", jobInfoPATH, strings.Join(id, ","))
				req := httptest.NewRequest("DELETE", target, nil)
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
			})

			Convey("delete job success", func() {
				d.EXPECT().DeleteJob(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				target := fmt.Sprintf("%v/%v", jobInfoPATH, strings.Join(id, ","))
				req := httptest.NewRequest("DELETE", target, nil)
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusOK)

				time.Sleep(time.Second)

				msg := <-timer.chJobMsg
				assert.Equal(t, msg.Method, common.DELETE)
			})

			Convey("delete job nsq disconnected", func() {
				d.EXPECT().DeleteJob(gomock.Any(), gomock.Any()).AnyTimes().Return(utils.NewECronError(common.ErrMSMQClientUnavailable, common.InternalError, nil))
				a.EXPECT().VerifyToken(gomock.Any()).AnyTimes().Return(common.Visitor{}, nil)
				a.EXPECT().VerifyCode(gomock.Any(), gomock.Any()).AnyTimes().Return(true, nil)

				stubNSQConnected(false, t)
				target := fmt.Sprintf("%v/%v", jobInfoPATH, strings.Join(id, ","))
				req := httptest.NewRequest("DELETE", target, nil)
				req.Header.Add(common.Authorization, fmt.Sprintf("%v %v", common.Bearer, "123"))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				result := w.Result()
				defer result.Body.Close()
				//respBody, _ := ioutil.ReadAll(result.Body)

				assert.Equal(t, result.StatusCode, http.StatusInternalServerError)
			})
		})
	})
}
