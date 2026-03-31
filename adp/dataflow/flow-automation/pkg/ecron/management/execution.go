package management

import (
	"errors"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils"
	"github.com/mitchellh/mapstructure"
)

//go:generate mockgen -package mock -source ../management/execution.go -destination ../mock/mock_execution.go

// Executor 任务执行服务
type Executor interface {
	ExecuteJob(job common.JobInfo) (reply bool, err error)
}

// NewExecutor 创建任务执行服务
func NewExecutor() Executor {
	return &executor{
		httpClient:    utils.NewHTTPClient(),
		commandRunner: NewCommandRunner(),
		mapDecoder:    NewMapDecoder(),
	}
}

// Executor 任务执行者结构
type executor struct {
	httpClient    utils.HTTPClient
	commandRunner CommandRunner
	mapDecoder    MapDecoder
}

type kubernetes struct {
	Addr  string `json:"addr"`
	Port  int    `json:"port"`
	Image string `json:"image"`
}

var (
	exeLog = utils.NewLogger()
)

// ExecuteJob 执行任务
func (e *executor) ExecuteJob(job common.JobInfo) (reply bool, err error) {
	if nil == e.httpClient {
		err = errors.New(common.ErrHTTPClientUnavailable)
		return
	}

	switch job.Context.Mode {
	case common.HTTP:
		fallthrough
	case common.HTTPS:
		return e.httpJob(job)
	case common.EXE:
		return e.exeJob(job)
	}

	err = errors.New(common.ErrUnsupportedExecutionMode)
	return
}

func (e *executor) httpJob(job common.JobInfo) (reply bool, err error) {
	url := job.Context.Exec
	headers := job.Context.Info.Headers
	body := job.Context.Info.Body
	method := job.Context.Info.Method
	switch method {
	case "GET":
		err = e.httpClient.Get(url, headers, nil)
	case "DELETE":
		err = e.httpClient.Delete(url, headers, nil)
	case "PUT":
		err = e.httpClient.Put(url, headers, body, nil)
	default:
		err = e.httpClient.Post(url, headers, body, nil)
	}
	if nil != err {
		reply = true
	}
	return
}

func (e *executor) exeJob(job common.JobInfo) (reply bool, err error) {
	//	$ kubectl proxy
	// $ curl -X POST -H 'Content-Type: application/yaml' --data '
	// apiVersion: batch/v1
	// kind: Job
	// metadata:
	//   name: example-job
	// spec:
	//   template:
	//     metadata:
	//       name: example-job
	//     spec:
	//       containers:
	//       - name: pi
	//         image: perl
	//         command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
	//       restartPolicy: Never
	// ' http://127.0.0.1:8001/apis/batch/v1/namespaces/default/jobs
	//time := time.Now().Add(5)
	//cronTime := fmt.Sprintf("%v %v %v %v %v ? %v", time.Second(), time.Minute(), time.Hour(), time.Day(), time.Month(), time.Year())
	path := job.Context.Exec
	params := job.Context.Info.Params
	headers := job.Context.Info.Headers
	body := job.Context.Info.Body
	k8s := job.Context.Info.Kubernetes

	if nil == k8s || 0 == len(k8s) {
		name, args, _ := e.getCmd(path, params)
		exeLog.Infof("[exeJob] getCmd params, name: %v, args: %v", name, args)
		return true, e.commandRunner.Run(name, args...)
	}

	k8sInfo := kubernetes{}
	err = e.mapDecoder.Decode(k8s, &k8sInfo)
	if nil != err {
		return false, err
	}

	var reqParam map[string]interface{}
	if nil != body && len(body) > 0 {
		reqParam = body
	} else {
		_, _, command := e.getCmd(path, params)
		reqParam = map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "Job",
			"metadata": map[string]interface{}{
				"name": job.JobName,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": job.JobName,
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							0: map[string]interface{}{
								"name":    job.JobName,
								"image":   k8sInfo.Image,
								"command": command,
							},
						},
					},
				},
				"restartPolicy": "Never",
			},
		}
	}

	url := fmt.Sprintf("http://%v/apis/batch/v1/namespaces/default/jobs", common.FormatAddress(k8sInfo.Addr, k8sInfo.Port))
	if nil == headers {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"
	err = e.httpClient.Post(url, headers, reqParam, nil)
	return true, err
}

func (e *executor) getCmd(path string, params map[string]string) (name string, args []string, command string) {
	args = make([]string, 0)
	cmdFields := make([]string, 0)
	execFields := strings.Split(path, " ")
	for i, v := range execFields {
		if 0 == i {
			name = v
		} else {
			args = append(args, v)
		}
		cmdFields = append(cmdFields, v)
	}

	if nil != params {
		for _, v := range params {
			args = append(args, v)
			cmdFields = append(cmdFields, v)
		}
	}
	fields, _ := jsoniter.Marshal(cmdFields)
	return name, args, string(fields)
}

// decodeMap 包装 mapstructure.Decode 以便于测试
func decodeMap(input interface{}, output interface{}) error {
	return mapstructure.Decode(input, output)
}
