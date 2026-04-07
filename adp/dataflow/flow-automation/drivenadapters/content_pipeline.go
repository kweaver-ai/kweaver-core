package drivenadapters

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
)

var (
	contentpipelineOnce sync.Once
	contentpipelineIns  ContentPipeline
)

type ContentPipeline interface {
	NewJob(ctx context.Context, req *NewJobReq) (*JobItem, error)
}

func NewContentPipeline() ContentPipeline {
	contentpipelineOnce.Do(func() {
		config := common.NewConfig()
		contentpipelineIns = &contentpipeline{
			privateBaseURL: fmt.Sprintf("http://%s:%s", config.ContentPipeline.PrivateHost, config.ContentPipeline.PrivatePort),
			httpClient2:    NewHTTPClient2(),
		}
	})

	return contentpipelineIns
}

type contentpipeline struct {
	privateBaseURL string
	httpClient2    HTTPClient2
}

// SourceData 资源数据
type SourceData struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// StepReq 步骤请求信息
type StepReq struct {
	Key             string      `json:"key"`
	Parameters      interface{} `json:"parameters"`
	Priority        int64       `json:"priority,omitempty"`
	TimeoutCalParam interface{} `json:"timeoutCalParam,omitempty"`
}

// TaskReq 任务请求信息
type TaskReq struct {
	Key      string     `json:"key"`
	Passback string     `json:"passback,omitempty"`
	Steps    []*StepReq `json:"steps"`
}

// JobItem Job项
type JobItem struct {
	CreatedAt int64       `json:"created_at"`
	ID        string      `json:"id"`
	Source    *SourceData `json:"source"`
	Version   string      `json:"version"`
}

// NewJobReq 新建Job请求
type NewJobReq struct {
	Passback string      `json:"passback,omitempty"`
	Source   *SourceData `json:"source"`
	Tasks    []*TaskReq  `json:"tasks"`
}

func (p *contentpipeline) NewJob(ctx context.Context, req *NewJobReq) (*JobItem, error) {
	target := fmt.Sprintf("%s/api/pipeline/v1/jobs", p.privateBaseURL)

	var result JobItem

	_, err := p.httpClient2.Post(ctx, target, map[string]string{}, req, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
