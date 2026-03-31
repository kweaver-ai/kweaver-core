package drivenadapters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/ecron.go -destination ../tests/mock_drivenadapters/ecron_mock.go

// taskContextStruct 定时任务
type taskContextStruct struct {
	Method string `json:"method"`
}

// JobContextStruct 定时任务
type jobContextStruct struct {
	Mode      string                 `json:"mode"`
	Exec      string                 `json:"exec"`
	Info      taskContextStruct      `json:"info"`
	ExecuteID string                 `json:"execute_id"`
	Notify    map[string]interface{} `json:"notify"`
	CreateAt  string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// ecronAddJobBody 注册定时任务请求Body
type ecronAddJobBody struct {
	JobName     string           `json:"job_name"`
	JobCronTime string           `json:"job_cron_time"`
	JobType     string           `json:"job_type"`
	Enabled     bool             `json:"enabled"`
	Remarks     string           `json:"remarks"`
	JobContext  jobContextStruct `json:"job_context"`
}

// ECron method interface
type ECron interface {
	// RunPyCode 运行python代码
	RegisterCronJob(ctx context.Context, id string, exec string, cronTime string) (jobID string, exist bool, err error)
	PostEcronJobEnd(ctx context.Context, cronJobHook string) (err error)
	DeleteEcronJob(ctx context.Context, jobID string) (err error)
	UpdateCronJob(ctx context.Context, id, exec, cronTime, jobID string) (err error)
}

type ecron struct {
	address      string
	clientID     string
	clientSecret string
	httpClient   otelHttp.HTTPClient
}

var (
	eOnce sync.Once
	e     ECron
)

const (
	// Rfc3339TimeFormt Rfc3339时间格式精确6位
	Rfc3339TimeFormt string = "2006-01-02T15:04:05.999999-07:00"
)

// NewECron 创建获取用户服务
func NewECron() ECron {
	eOnce.Do(func() {
		config := common.NewConfig()
		e = &ecron{
			address:      fmt.Sprintf("http://%s:%v", config.ECron.Host, config.ECron.Port),
			httpClient:   NewOtelHTTPClient(),
			clientID:     config.OAuth.ClientID,
			clientSecret: config.OAuth.ClientSecret,
		}
	})
	return e
}

// RegisterCronJob 注册定时任务
func (e *ecron) RegisterCronJob(ctx context.Context, id, exec, cronTime string) (jobID string, exist bool, err error) {
	taskContxt := taskContextStruct{
		Method: "POST",
	}
	jobContext := jobContextStruct{
		Mode:     "http",
		Exec:     exec,
		Info:     taskContxt,
		CreateAt: time.Now().Format(Rfc3339TimeFormt),
	}
	reqbody := ecronAddJobBody{
		JobName: fmt.Sprintf("workflow_%s", id),
		// JobCronTime: "0 0 0 * * ?",
		JobCronTime: cronTime,
		JobType:     "scheduled",
		JobContext:  jobContext,
		Enabled:     true,
		Remarks:     "",
	}
	target := fmt.Sprintf("%s/api/ecron-management/v1/job", e.address)
	tokenInfo, _, err := NewHydraPublic().RequestTokenWithCredential(e.clientID, e.clientSecret, []string{"all"})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RequestTokenWithCredential failed: %v, url: %v", err, target)
		return
	}

	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"Authorization": fmt.Sprintf("Bearer %s", tokenInfo.Token),
	}

	code, resp, err := e.httpClient.Post(ctx, target, headers, reqbody)
	if code == 409 {
		exist = true
		traceLog.WithContext(ctx).Infof("INFO: RegisterCronJob is existed")
		return
	}
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RegisterCronJob failed: %v, url: %v", err, target)
		return
	}
	jobID = resp.(map[string]interface{})["job_id"].(string)

	return
}

// UpdateCronJob 更新定时任务
func (e *ecron) UpdateCronJob(ctx context.Context, id, exec, cronTime, jobID string) (err error) {
	taskContxt := taskContextStruct{
		Method: "POST",
	}
	jobContext := jobContextStruct{
		Mode:     "http",
		Exec:     exec,
		Info:     taskContxt,
		CreateAt: time.Now().Format(Rfc3339TimeFormt),
	}
	reqbody := ecronAddJobBody{
		JobName:     fmt.Sprintf("workflow_%s", id),
		JobCronTime: cronTime,
		JobType:     "scheduled",
		JobContext:  jobContext,
		Enabled:     true,
		Remarks:     "",
	}
	target := fmt.Sprintf("%s/api/ecron-management/v1/job/%s", e.address, jobID)
	tokenInfo, _, err := NewHydraPublic().RequestTokenWithCredential(e.clientID, e.clientSecret, []string{"all"})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RequestTokenWithCredential failed: %v, url: %v", err, target)
		return
	}

	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"Authorization": fmt.Sprintf("Bearer %s", tokenInfo.Token),
	}

	_, _, err = e.httpClient.Put(ctx, target, headers, reqbody)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RegisterCronJob failed: %v, url: %v", err, target)
		return
	}

	return
}

// PostEcronJobEnd 通知ecron任务结束
func (e *ecron) PostEcronJobEnd(ctx context.Context, cronJobHook string) (err error) {
	target := fmt.Sprintf("%s%s", e.address, cronJobHook)
	tokenInfo, _, err := NewHydraPublic().RequestTokenWithCredential(e.clientID, e.clientSecret, []string{"all"})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RequestTokenWithCredential failed: %v, url: %v", err, target)
		return
	}

	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"Authorization": fmt.Sprintf("Bearer %s", tokenInfo.Token),
	}
	traceLog.WithContext(ctx).Infof("webhook.url:%v\n", target)
	_, _, err = e.httpClient.Post(ctx, target, headers, nil)
	return
}

// DeleteEcronJobEnd 删除定时任务
func (e *ecron) DeleteEcronJob(ctx context.Context, jobID string) (err error) {
	target := fmt.Sprintf("%s/api/ecron-management/v1/job/%s", e.address, jobID)
	tokenInfo, _, err := NewHydraPublic().RequestTokenWithCredential(e.clientID, e.clientSecret, []string{"all"})
	if err != nil {
		traceLog.WithContext(ctx).Warnf("RequestTokenWithCredential failed: %v, url: %v", err, target)
		return
	}

	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"Authorization": fmt.Sprintf("Bearer %s", tokenInfo.Token),
	}
	_, err = e.httpClient.Delete(ctx, target, headers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("DeleteEcronJob failed: %v, url: %v", err, target)
		return
	}
	return
}
