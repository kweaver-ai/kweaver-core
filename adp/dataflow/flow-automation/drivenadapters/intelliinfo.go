package drivenadapters

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/intelliinfo.go -destination ../tests/mock_drivenadapters/intelliinfo_mock.go

const (
	DataTransConcurrentLimit = "intelliinfo.DataTrans.ConcurrentLimit"
)

const (
	_ = iota
	V1
	V2
)

type DataTransferReqV1 struct {
	RuleID   string      `json:"rule_id"`
	Data     interface{} `json:"data"`
	CallBack string      `json:"callback"`
}

type DataTransferReqV2 struct {
	Datas    []*ADGraphTansferWrapper `json:"datas"`
	CallBack string                   `json:"callback"`
}

type ADGraphTansferWrapper struct {
	GrapID   int              `json:"graph_id"`
	Entities []*ADGraphEntity `json:"entities"`
	Edges    []*ADGraphEdge   `json:"edges"`
}

type ADGraphEdge struct {
	Name  string      `json:"name"`
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value,omitempty"`
}

type ADGraphEntity struct {
	Action      string        `json:"action"`
	Name        string        `json:"name"`
	ArrayNumKey string        `json:"array_num_key,omitempty"`
	Fields      []ADGraphProp `json:"fields"`
}

type ADGraphProp struct {
	Key             string      `json:"key"`
	Type            string      `json:"type"`
	Value           interface{} `json:"value,omitempty"`
	FixedValue      string      `json:"fixed_value,omitempty"`
	TypeFormat      string      `json:"type_format,omitempty"`
	OnNil           string      `json:"on_nil,omitempty"`
	OnTypeMissMatch string      `json:"on_type_miss_match,omitempty"`
}

// Intelliinfo method interface
type Intelliinfo interface {
	// TransferData 数据转换为智能信息数据
	TransferData(ctx context.Context, data interface{}, apiVers int, userID, userType string) (interface{}, error)
}

type intelliinfo struct {
	privateURL string
	httpClient otelHttp.HTTPClient
}

var (
	iOnce sync.Once
	i     Intelliinfo
)

// NewIntelliinfo 创建Intelliinfo服务
func NewIntelliinfo() Intelliinfo {
	iOnce.Do(func() {
		config := common.NewConfig()
		i = &intelliinfo{
			privateURL: fmt.Sprintf("http://%s:%v", config.Intelliinfo.PrivateHost, config.Intelliinfo.PrivatePort),
			httpClient: NewOtelHTTPClient(),
		}
	})
	return i
}

// TransferData 数据转换为智能信息数据
func (i *intelliinfo) TransferData(ctx context.Context, data interface{}, apiVers int, userID, userType string) (interface{}, error) {
	switch apiVers {
	case V1:
		body := data.(*DataTransferReqV1)
		return i.TransferDataV1(ctx, body.RuleID, body.CallBack, userID, userType, body.Data)
	case V2:
		body := data.(*DataTransferReqV2)
		return i.TransferDataV2(ctx, body, userID, userType)
	default:
		return nil, fmt.Errorf("unsupport api version")
	}
}

// TransferData 数据转换为智能信息数据
func (i *intelliinfo) TransferDataV1(ctx context.Context, ruleID, callback, userID, userType string, data interface{}) (interface{}, error) {
	target := fmt.Sprintf("%s/api/intelliinfo/v1/data/transfer", i.privateURL)

	headers := map[string]string{
		"Content-Type":   "application/json;charset=UTF-8",
		"X-Account-ID":   userID,
		"X-Account-Type": userType,
	}

	body := map[string]interface{}{
		"rule_id":  ruleID,
		"data":     data,
		"callback": callback,
	}
	_, res, err := i.httpClient.Post(ctx, target, headers, body)

	if err != nil {
		parseError, err0 := errors.ExHTTPErrorParser(err)
		if err0 == nil {
			if parseError["code"] == DataTransConcurrentLimit {
				traceLog.WithContext(ctx).Debugf("TransferData failed: %v, url: %v", err, target)
				return nil, err
			}
		}

		traceLog.WithContext(ctx).Warnf("TransferData failed: %v, url: %v", err, target)
		return nil, err
	}

	return res, nil
}

// TransferDataV2 图谱写入V2接口
func (i *intelliinfo) TransferDataV2(ctx context.Context, req *DataTransferReqV2, userID, userType string) (interface{}, error) {
	target := fmt.Sprintf("%s/api/intelliinfo/v2/data/transfer", i.privateURL)

	headers := map[string]string{
		"Content-Type":   "application/json;charset=UTF-8",
		"X-Account-ID":   userID,
		"X-Account-Type": userType,
	}

	_, res, err := i.httpClient.Post(ctx, target, headers, req)
	if err != nil {
		parseError, err0 := errors.ExHTTPErrorParser(err)
		if err0 == nil {
			if parseError["code"] == DataTransConcurrentLimit {
				traceLog.WithContext(ctx).Debugf("TransferDataV2 failed: %v, url: %v", err, target)
				return nil, err
			}
		}

		traceLog.WithContext(ctx).Warnf("TransferDataV2 failed: %v, url: %v", err, target)
		return nil, err
	}

	return res, nil
}
