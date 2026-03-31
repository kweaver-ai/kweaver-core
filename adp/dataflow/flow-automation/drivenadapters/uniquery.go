package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/uniquery.go -destination ../tests/mock_drivenadapters/uniquery_mock.go

const (
	defaultDataViewID   = "__dip_flow_o11y_data"
	defaultStatusSize   = 10
	defaultAdminUser    = "266c6a42-6131-4d62-8f39-853e7093701c"
	DefaultDataViewName = "dip_flow_o11y_data"
)

type UniqueryDataViewSortOption struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type UniqueryDataViewOptions struct {
	Start          int64                         `json:"start,omitempty"`
	End            int64                         `json:"end,omitempty"`
	Sort           []*UniqueryDataViewSortOption `json:"sort,omitempty"`
	Offset         int                           `json:"offset,omitempty"`
	Limit          int                           `json:"limit,omitempty"`
	NeedTotal      bool                          `json:"need_total"`
	UseSearchAfter bool                          `json:"use_search_after"`
	SearchAfter    []any                         `json:"search_after,omitempty"`
	PitID          string                        `json:"pit_id,omitempty"`
	PitKeepAlive   string                        `json:"pit_keep_alive,omitempty"`
	Format         string                        `json:"format,omitempty"`
	QueryType      string                        `json:"query_type,omitempty"`
	Sql            string                        `json:"sql,omitempty"`
}

type UniqueryDataViewResBody struct {
	Entries     []any  `json:"entries,omitempty"`
	TotalCount  int    `json:"total_count,omitempty"`
	SearchAfter []any  `json:"search_after,omitempty"`
	PitID       string `json:"pit_id,omitempty"`
}

// UniqueryDriven AR uniquery 查询接口
type UniqueryDriven interface {
	QueryDagStatusCount(ctx context.Context, startTime, endTime int64, token string, qfiled QueryFileds) (MetricQueryRes, error)
	QueryDagRunTimeAvg(ctx context.Context, startTime, endTime int64, token string, qfiled QueryFileds) (MetricQueryRes, error)
	CheckDataViewExist(ctx context.Context, viewName string) (bool, error)
	UniqueryDataView(ctx context.Context, id string, opts *UniqueryDataViewOptions, userID, userType string) (resBody *UniqueryDataViewResBody, err error)
	GetPipelineByID(ctx context.Context, id string, userID, userType string) (pipeline *Pipeline, err error)
	GetDataViewByID(ctx context.Context, id string, userID, userType string) (dataview *DataView, err error)
}

type uniquery struct {
	baseURL          string
	dataModelBaseURL string
	pipelineBaseURL  string
	dataViewID       string
	httpClient       otelHttp.HTTPClient
	httpClient2      HTTPClient2
	oauthHttpClient  otelHttp.OAuth2Client
}

var uq sync.Once
var uqInstance UniqueryDriven

type DSLConfig struct {
	MetricType    string        `json:"metric_type"`
	QueryType     string        `json:"query_type"`
	Instant       bool          `json:"instant"`
	DataSource    DataSource    `json:"data_source"`
	Filters       []interface{} `json:"filters"`
	FormulaConfig FormulaConfig `json:"formula_config"`
	Time          int64         `json:"time"`
	LookBackDelta string        `json:"look_back_delta"`
}

type FormulaConfig struct {
	Buckets       []BucketConfig `json:"buckets"`
	Aggregation   Aggregation    `json:"aggregation"`
	DateHistogram struct{}       `json:"date_histogram"`
	QueryString   string         `json:"query_string"`
}

type BucketConfig struct {
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Field      string     `json:"field"`
	Order      string     `json:"order"`
	Direction  string     `json:"direction"`
	Size       int        `json:"size"`
	DataSource DataSource `json:"data_source"`
}

type DataSource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type Aggregation struct {
	Field string `json:"field,omitempty"`
	Type  string `json:"type"`
}

// MetricQueryRes 指标模型查询结果
type MetricQueryRes struct {
	Datas []DataEntry `json:"datas"`
}

type DataEntry struct {
	Labels LabelData     `json:"labels"`
	Times  []int64       `json:"times"`
	Values []interface{} `json:"values"`
}

type LabelData struct {
	DagID  string `json:"dagId"`
	Status string `json:"status"`
}

type StatusCnt struct {
	Total    int64 `json:"total"`
	Success  int64 `json:"success"`
	Failed   int64 `json:"failed"`
	Blocked  int64 `json:"blocked"`
	Canceled int64 `json:"canceled"`
	Running  int64 `json:"running"`
	Init     int64 `json:"init"`
}

// ToMap 获取当前时间范围内运行状态统计Map
func (mqr *MetricQueryRes) ToMap() map[string]StatusCnt {
	statusMap := make(map[string]StatusCnt)
	for _, data := range mqr.Datas {
		item := statusMap[data.Labels.DagID]
		value := int64(data.Values[0].(float64))
		switch data.Labels.Status {
		case "success":
			item.Success += value
		case "failed":
			item.Failed += value
		case "canceled":
			item.Canceled += value
		}

		item.Total += value
		statusMap[data.Labels.DagID] = item
	}
	return statusMap
}

// TotalCnt 获取当前时间范围内运行总数以及成功失败总数
func (mqr *MetricQueryRes) TotalCnt() StatusCnt {
	cnt := StatusCnt{}
	for _, data := range mqr.Datas {
		value := int64(data.Values[0].(float64))
		switch data.Labels.Status {
		case "success":
			cnt.Success += value
		case "failed":
			cnt.Failed += value
		case "canceled":
			cnt.Canceled += value
		}
		cnt.Total += value
	}
	return cnt
}

// GetDagIDs 获取当前结果所有DagID
func (mqr *MetricQueryRes) GetDagIDs() []string {
	dagIDs := []string{}
	for _, data := range mqr.Datas {
		dagIDs = append(dagIDs, data.Labels.DagID)
	}
	return dagIDs
}

// AvgTimeToMap 将数据流每个流程耗时从列表转换成map
func (mqr *MetricQueryRes) AvgTimeToMap() map[string]float64 {
	avgTimeMap := map[string]float64{}
	for _, data := range mqr.Datas {
		value := data.Values[0].(float64)
		avgTimeMap[data.Labels.DagID] += math.Round(value*10) / 10
	}
	return avgTimeMap
}

type QueryFileds struct {
	DagIDs      []string
	DagType     string
	BizDomainID string
}

type DataViews struct {
	Entries []DataView `json:"entries"`
}

type DataViewField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type DataView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TableName string `json:"meta_table_name"`
	Fields    []*DataViewField
}

// NewUniqueQuery 实例化对象
func NewUniquery() UniqueryDriven {
	uq.Do(func() {
		config := common.NewConfig()

		uqInstance = &uniquery{
			dataViewID:       defaultDataViewID,
			baseURL:          fmt.Sprintf("http://%s:%s", config.MdlUniquery.Host, config.MdlUniquery.Port),
			dataModelBaseURL: fmt.Sprintf("http://%s:%s", config.MdlDataModel.Host, config.MdlDataModel.Port),
			pipelineBaseURL:  fmt.Sprintf("http://%s:%s", config.MdlDataPipeline.Host, config.MdlDataPipeline.Port),
			httpClient:       NewOtelHTTPClient(),
			httpClient2:      NewHTTPClient2(),
			oauthHttpClient:  NewOauthOtelHTTPClient(),
		}
	})
	return uqInstance
}

// BuildQuery 统一构建查询条件
func BuildQuery(qfiled QueryFileds) string {
	query := fmt.Sprintf("Body.Type:%q AND Body.flow-automation.operation:%q", "flow-automation", "end")
	if len(qfiled.DagIDs) > 0 {
		query = fmt.Sprintf("%s AND Body.flow-automation.object.dagId: (\"%s\")", query, strings.Join(qfiled.DagIDs, `", "`))
	}

	if qfiled.DagType != "" {
		query = fmt.Sprintf("%s AND Body.flow-automation.object.dagType:%q", query, qfiled.DagType)
	}

	if qfiled.BizDomainID == "" || qfiled.BizDomainID == common.BizDomainDefaultID {
		query = fmt.Sprintf("%s AND ((NOT _exists_:Body.flow-automation.object.biz_domain_id) OR Body.flow-automation.object.biz_domain_id:%q)", query, qfiled.BizDomainID)
	} else {
		query = fmt.Sprintf("%s AND Body.flow-automation.object.biz_domain_id:%q", query, qfiled.BizDomainID)
	}

	return query
}

// QueryDagStatusCount 获取数据流状态数量统计
func (u *uniquery) QueryDagStatusCount(ctx context.Context, startTime, endTime int64, token string, qfiled QueryFileds) (MetricQueryRes, error) {
	url := fmt.Sprintf("%s/api/mdl-uniquery/v1/metric-model", u.baseURL)
	var metricQueryRes MetricQueryRes

	// 默认返回前1000w条数据
	size := 10000000
	if len(qfiled.DagIDs) > 0 {
		size = len(qfiled.DagIDs) * defaultStatusSize
	}
	query := BuildQuery(qfiled)

	startTimeStr := fmt.Sprintf("%v", startTime)

	if len(startTimeStr) > 13 {
		startTimeStr = startTimeStr[:13]
	} else {
		startTimeStr = fmt.Sprintf("%v%s", startTimeStr, strings.Repeat("0", 13-len(startTimeStr)))
	}
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		return metricQueryRes, err
	}

	endTimeStr := fmt.Sprintf("%v", endTime)
	if len(endTimeStr) > 13 {
		endTimeStr = endTimeStr[:13]
	} else {
		endTimeStr = fmt.Sprintf("%v%s", endTimeStr, strings.Repeat("0", 13-len(endTimeStr)))
	}
	endTime, err = strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return metricQueryRes, err
	}

	dsl := &DSLConfig{
		MetricType: "atomic",
		QueryType:  "dsl_config",
		Instant:    true,
		DataSource: DataSource{
			Type: "data_view",
			ID:   u.dataViewID,
		},
		FormulaConfig: FormulaConfig{
			Buckets: []BucketConfig{
				{
					Name:      "dagId",
					Type:      "terms",
					Field:     "Body.flow-automation.object.dagId",
					Order:     "field",
					Direction: "desc",
					Size:      size,
					DataSource: DataSource{
						Type: "dataView",
						ID:   u.dataViewID,
					},
				},
				{
					Name:      "status",
					Type:      "terms",
					Field:     "Body.flow-automation.object.status",
					Order:     "field",
					Direction: "desc",
					Size:      defaultStatusSize,
					DataSource: DataSource{
						Type: "dataView",
						ID:   u.dataViewID,
					},
				},
			},
			Aggregation: Aggregation{
				Type: "doc_count",
			},
			QueryString: query,
		},
		Time:          endTime,
		LookBackDelta: fmt.Sprintf("%dms", endTime-startTime),
	}

	dsfBytes, _ := json.Marshal(dsl)
	if !strings.HasPrefix(token, "Bearer") {
		token = fmt.Sprintf("Bearer %s", token)
	}

	headers := map[string]string{"Content-Type": "application/json", "Authorization": token, "X-Http-Method-Override": "GET"}
	_, resp, err := u.httpClient.Post(ctx, url, headers, dsfBytes)
	if err != nil {
		return metricQueryRes, err
	}

	bytes, _ := json.Marshal(resp)

	_ = json.Unmarshal(bytes, &metricQueryRes)

	return metricQueryRes, nil
}

func (u *uniquery) QueryDagRunTimeAvg(ctx context.Context, startTime, endTime int64, token string, qfiled QueryFileds) (MetricQueryRes, error) {
	url := fmt.Sprintf("%s/api/mdl-uniquery/v1/metric-model", u.baseURL)
	var metricQueryRes MetricQueryRes

	// 默认返回前1000条数据
	size := 1000
	if len(qfiled.DagIDs) > 0 {
		size = len(qfiled.DagIDs)
	}

	query := BuildQuery(qfiled)

	startTimeStr := fmt.Sprintf("%v", startTime)

	if len(startTimeStr) > 13 {
		startTimeStr = startTimeStr[:13]
	} else {
		startTimeStr = fmt.Sprintf("%v%s", startTimeStr, strings.Repeat("0", 13-len(startTimeStr)))
	}
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		return metricQueryRes, err
	}

	endTimeStr := fmt.Sprintf("%v", endTime)
	if len(endTimeStr) > 13 {
		endTimeStr = endTimeStr[:13]
	} else {
		endTimeStr = fmt.Sprintf("%v%s", endTimeStr, strings.Repeat("0", 13-len(endTimeStr)))
	}
	endTime, err = strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		return metricQueryRes, err
	}

	dsl := &DSLConfig{
		MetricType: "atomic",
		QueryType:  "dsl_config",
		Instant:    true,
		DataSource: DataSource{
			Type: "data_view",
			ID:   u.dataViewID,
		},
		FormulaConfig: FormulaConfig{
			Buckets: []BucketConfig{
				{
					Name:      "dagId",
					Type:      "terms",
					Field:     "Body.flow-automation.object.dagId",
					Order:     "field",
					Direction: "desc",
					Size:      size,
					DataSource: DataSource{
						Type: "dataView",
						ID:   u.dataViewID,
					},
				},
			},
			Aggregation: Aggregation{
				Field: "Body.flow-automation.object.duration",
				Type:  "avg",
			},
			QueryString: query,
		},
		Time:          endTime,
		LookBackDelta: fmt.Sprintf("%dms", endTime-startTime),
	}

	dsfBytes, _ := json.Marshal(dsl)

	if !strings.HasPrefix(token, "Bearer") {
		token = fmt.Sprintf("Bearer %s", token)
	}

	headers := map[string]string{"Content-Type": "application/json", "Authorization": token, "X-Http-Method-Override": "GET"}
	_, resp, err := u.httpClient.Post(ctx, url, headers, dsfBytes)
	if err != nil {
		return metricQueryRes, err
	}

	bytes, _ := json.Marshal(resp)

	_ = json.Unmarshal(bytes, &metricQueryRes)

	return metricQueryRes, nil
}

// CheckDataViewExist 检查数据视图是否存在
func (u *uniquery) CheckDataViewExist(ctx context.Context, viewName string) (bool, error) {
	url := fmt.Sprintf("%s/api/mdl-data-model/in/v1/data-views?name=%s", u.dataModelBaseURL, viewName)

	headers := map[string]string{
		"content-type":   "application/json",
		"X-ACCOUNT-ID":   defaultAdminUser,
		"X-ACCOUNT-TYPE": common.User.ToString(),
	}

	_, resp, err := u.httpClient.Get(ctx, url, headers)
	if err != nil {
		return false, err
	}

	var dataViews DataViews
	bytes, _ := json.Marshal(resp)

	_ = json.Unmarshal(bytes, &dataViews)

	if len(dataViews.Entries) == 0 {
		return false, nil
	}

	return true, nil
}

func (u *uniquery) UniqueryDataView(ctx context.Context, id string, opts *UniqueryDataViewOptions, userID, userType string) (resBody *UniqueryDataViewResBody, err error) {
	target := fmt.Sprintf("%s/api/mdl-uniquery/in/v1/data-views/%s", u.baseURL, id)
	headers := map[string]string{
		"x-http-method-override": "GET",
		"content-type":           "application/json",
		"X-ACCOUNT-ID":           userID,
		"X-ACCOUNT-TYPE":         userType,
	}
	resBody = new(UniqueryDataViewResBody)
	_, err = u.httpClient2.Post(ctx, target, headers, opts, resBody)

	if err != nil {
		return nil, err
	}

	return resBody, nil
}

type Pipeline struct {
	ID          string `json:"id"`
	OutputTopic string `json:"output_topic"`
	InputTopic  string `json:"input_topic"`
	ErrorTopic  string `json:"error_topic"`
	Status      string `json:"status"`
}

func (u *uniquery) GetPipelineByID(ctx context.Context, id string, userID, userType string) (pipeline *Pipeline, err error) {
	target := fmt.Sprintf("%s/api/flow-stream-data-pipeline/in/v1/pipelines/%s", u.pipelineBaseURL, id)
	headers := map[string]string{
		"X-ACCOUNT-ID":   userID,
		"X-ACCOUNT-TYPE": userType,
	}

	pipeline = new(Pipeline)
	_, err = u.httpClient2.Get(ctx, target, headers, pipeline)

	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

func (u *uniquery) GetDataViewByID(ctx context.Context, id string, userID, userType string) (dataview *DataView, err error) {
	target := fmt.Sprintf("%s/api/mdl-data-model/in/v1/data-views/%s", u.dataModelBaseURL, id)
	headers := map[string]string{
		"X-ACCOUNT-ID":   userID,
		"X-ACCOUNT-TYPE": userType,
	}
	var items []DataView
	_, err = u.httpClient2.Get(ctx, target, headers, &items)

	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("dataview %s not found", id)
	}

	return &items[0], nil
}
