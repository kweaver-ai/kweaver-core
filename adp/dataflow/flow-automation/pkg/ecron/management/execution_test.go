package management

import (
	"errors"
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/mock"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newExecutor(h utils.HTTPClient, c CommandRunner, m MapDecoder) *executor {
	return &executor{
		httpClient:    h,
		commandRunner: c,
		mapDecoder:    m,
	}
}

func TestExecuteJob(t *testing.T) {
	Convey("ExecuteJob", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := mock.NewMockHTTPClient(ctrl)
		c := mock.NewMockCommandRunner(ctrl)
		m := mock.NewMockMapDecoder(ctrl)
		executor := newExecutor(h, c, m)

		Convey("not supported execution mode", func() {
			reqParam := common.JobInfo{
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
			}
			reply, err := executor.ExecuteJob(reqParam)
			assert.Equal(t, err.Error(), common.ErrUnsupportedExecutionMode)
			assert.Equal(t, reply, false)
		})

		Convey("mode", func() {
			Convey("http delete success", func() {
				h.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

				reqParam := common.JobInfo{
					JobName:     "test",
					JobType:     common.TIMING,
					JobCronTime: "*/10 * * * * ?",
					Context: common.JobContext{
						Mode: common.HTTP,
						Info: common.JobContextInfo{
							Method: "DELETE",
						},
					},
				}
				reply, err := executor.ExecuteJob(reqParam)
				assert.Equal(t, err, nil)
				assert.Equal(t, reply, false)
			})

			Convey("https post failed", func() {
				h.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(errors.New("failed"))

				reqParam := common.JobInfo{
					JobName:     "test",
					JobType:     common.TIMING,
					JobCronTime: "*/10 * * * * ?",
					Context: common.JobContext{
						Mode: common.HTTPS,
					},
				}
				reply, err := executor.ExecuteJob(reqParam)
				assert.NotEqual(t, err, nil)
				assert.Equal(t, reply, true)
			})
		})

		Convey("exe mode, kubernetes job", func() {
			job := common.JobInfo{
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
				Context: common.JobContext{
					Mode: common.EXE,
					Exec: "python -a 1 -b 2",
					Info: common.JobContextInfo{
						Kubernetes: map[string]interface{}{
							"addr":  "localhost",
							"port":  12345,
							"image": "acr/aishu.cn/as/abc:1.0.0",
						},
					},
				},
			}

			Convey("mapstructure decode failed", func() {
				m.EXPECT().Decode(gomock.Any(), gomock.Any()).Return(errors.New("failed"))

				reply, err := executor.ExecuteJob(job)
				assert.NotEqual(t, err, nil)
				assert.Equal(t, reply, false)
			})

			h.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			m.EXPECT().Decode(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

			Convey("normal, no body", func() {
				reply, err := executor.ExecuteJob(job)
				assert.Equal(t, err, nil)
				assert.Equal(t, reply, true)
			})

			Convey("normal, has body", func() {
				job.Context.Info.Body = map[string]interface{}{
					"apiVersion": "batch/v1",
				}
				reply, err := executor.ExecuteJob(job)
				assert.Equal(t, err, nil)
				assert.Equal(t, reply, true)
			})
		})

		Convey("exe mode, local job", func() {
			h.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
			c.EXPECT().Run(gomock.Any(), gomock.Any()).Return(nil)

			job := common.JobInfo{
				JobName:     "test",
				JobType:     common.TIMING,
				JobCronTime: "*/10 * * * * ?",
				Context: common.JobContext{
					Mode: common.EXE,
					Exec: "python -a 1 -b 2",
					Info: common.JobContextInfo{
						Params: map[string]string{
							"-c": "3",
						},
					},
				},
			}
			reply, err := executor.ExecuteJob(job)
			assert.Equal(t, err, nil)
			assert.Equal(t, reply, true)
		})
	})
}

func TestGetCmd(t *testing.T) {
	Convey("getCmd", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := mock.NewMockHTTPClient(ctrl)
		c := mock.NewMockCommandRunner(ctrl)
		m := mock.NewMockMapDecoder(ctrl)
		executor := newExecutor(h, c, m)

		name, args, command := executor.getCmd("python -a 1 -b 2", map[string]string{
			"p1": "-c",
		})
		assert.Equal(t, name, "python")
		assert.Equal(t, args, []string{"-a", "1", "-b", "2", "-c"})
		assert.Equal(t, command, "[\"python\",\"-a\",\"1\",\"-b\",\"2\",\"-c\"]")
	})
}
