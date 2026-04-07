package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	aerr "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	cstore "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/store"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils/ptr"
)

//go:generate mockgen -package mock_observability -source ../../logics/observability/observability.go -destination ../../tests/mock_logics/mock_observability/observability_mock.go

const (
	chunkSize        = 1024
	fullViewCacheKey = "observability_full_view"
	rencentCacheKey  = "observability_recent"
	runtimeCacheKey  = "observability_runtime"
)

// ObservabilityHandler 可观测接口
type ObservabilityHandler interface {
	FullView(ctx context.Context, params ObservabilityQueryParams, userInfo *drivenadapters.UserInfo) (FullViewRes, error)
	RuntimeView(ctx context.Context, params ObservabilityQueryParams, userInfo *drivenadapters.UserInfo) (RuntimeViewRes, error)
	RecentRunView(ctx context.Context, params ObservabilityQueryParams, userInfo *drivenadapters.UserInfo) ([]*RuntimeViewItem, error)
	IsVisible(ctx context.Context) bool
}

var (
	oOnce sync.Once
	o     ObservabilityHandler
)

type observability struct {
	mongo       mod.Store
	uniquery    drivenadapters.UniqueryDriven
	usermgnt    drivenadapters.UserManagement
	memoryCache cstore.LocalCache
	permChecker perm.PermCheckerService
	permPolicy  perm.PermPolicyHandler
	bizDomain   drivenadapters.BusinessDomain
}

// ObservabilityQueryParams 可观测查询参数
type ObservabilityQueryParams struct {
	StartTime   *int64  `json:"start_time,string"`
	EndTime     *int64  `json:"end_time,string"`
	Page        *int64  `json:"page"`
	Limit       *int64  `json:"limit"`
	Type        *string `json:"type"`
	Trigger     *string `json:"trigger"`
	Keyword     *string `json:"keyword"`
	BizDomainID string  `json:"-"`
}

type BasicInfo struct {
	DagCnt   int64 `json:"dag_total"`
	Cron     int64 `json:"cron"`
	Event    int64 `json:"event"`
	Manually int64 `json:"manually"`
}

type RunInfo struct {
	SuccessCnt int64 `json:"success"`
	FailedCnt  int64 `json:"failed"`
	Canceled   int64 `json:"canceled"`
	Running    int64 `json:"running"`
	Scheduled  int64 `json:"scheduled"`
	TotalCnt   int64 `json:"run_total"`
}

// FullViewRes 全量视图响应
type FullViewRes struct {
	Basic BasicInfo `json:"basic"`
	Run   RunInfo   `json:"run"`
}

// RuntimeViewRes 运行时视图响应
type RuntimeViewRes struct {
	Datas []*RuntimeViewItem `json:"datas"`
	Total int64              `json:"total"`
	Page  int64              `json:"page"`
	Limit int64              `json:"limit"`
}

// RuntimeViewItem 运行时可观测所包含信息
type RuntimeViewItem struct {
	ID            string                    `json:"id"`
	Name          string                    `json:"name"`
	Creator       string                    `json:"creator"`
	Metric        *RuntimeViewMetric        `json:"metric"`
	StatusSummary *drivenadapters.StatusCnt `json:"status_summary"`
}

// RuntimeViewMetric 运行时可观测流程指标
type RuntimeViewMetric struct {
	FailedRate     float64 `json:"failed_rate"`
	SuccessRate    float64 `json:"success_rate"`
	AvgRunDuration float64 `json:"avg_run_duration"`
}

// NewObservability observability instance
func NewObservability() ObservabilityHandler {
	oOnce.Do(func() {
		o = &observability{
			mongo:    mod.GetStore(),
			uniquery: drivenadapters.NewUniquery(),
			usermgnt: drivenadapters.NewUserManagement(),
			memoryCache: cstore.NewLocalCache(&cstore.Option{
				Expiration:      30 * time.Second,
				CleanUpInterval: 5 * time.Minute,
			}),
			permChecker: perm.NewPermCheckerService(),
			permPolicy:  perm.NewPermPolicy(),
			bizDomain:   drivenadapters.NewBusinessDomain(),
		}
		perm.RegisterChecker(common.ReSourceTypeObservability, &perm.ObservabilityPermChecker{PermPolicy: perm.NewPermPolicy()})
	})
	return o
}

// FullView 获取当前所有已创建的数据流以及数据视图
func (o *observability) FullView(ctx context.Context, params ObservabilityQueryParams, userInfo *drivenadapters.UserInfo) (FullViewRes, error) {
	var (
		err error
		res FullViewRes
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.ReSourceTypeObservability: {perm.DisplayOperation},
		},
	}

	_, err = o.permChecker.CheckPerm(ctx, common.ReSourceTypeObservability, nil, userInfo, opMap)
	if err != nil {
		return res, err
	}

	err = o.PreCheck(ctx, &params)
	if err != nil {
		return res, err
	}

	b, _ := json.Marshal(params)
	cacheKey := fmt.Sprintf("%s:%s:%s", fullViewCacheKey, userInfo.UserID, utils.ComputeHash(string(b)))

	contents, exist := o.memoryCache.Get(cacheKey)
	if exist {
		err = json.Unmarshal([]byte(contents), &res)
		if err != nil {
			log.Warnf("[logic.FullView] Unmarshal err, detail: %s", err.Error())
			return res, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
		}
		return res, nil
	}

	listDagInput := mgnt.QueryParams{Type: *params.Type}
	if params.Trigger != nil {
		listDagInput.TriggerType = *params.Trigger
	}

	// 获取当前所有已创建的数据流
	dags, _, err := mgnt.ListDagWithFilters(ctx, listDagInput,
		mgnt.WithBizDomainFilter(o.bizDomain, params.BizDomainID, "", *params.Type, userInfo.TokenID),
		mgnt.WithPermissionFilter(o.permPolicy, userInfo, *params.Type),
		mgnt.WithExistenceFilter())
	if err != nil {
		log.Warnf("[logic.FullView] ListDag err, detail: %s", err.Error())
		return res, err
	}

	if len(dags) == 0 {
		o.memoryCache.Set(cacheKey, res, 30*time.Second)
		return res, nil
	}

	dataViews, err := o.uniquery.QueryDagStatusCount(ctx, *params.StartTime, *params.EndTime, userInfo.TokenID, drivenadapters.QueryFileds{
		DagType:     *params.Type,
		BizDomainID: params.BizDomainID,
	})
	if err != nil {
		log.Warnf("[logic.FullView] QueryDagStatusCount err, detail: %s", err.Error())
		return res, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, err.Error())
	}

	dvMap := dataViews.ToMap()

	for _, dag := range dags {
		cnt := dvMap[dag.ID]
		res.Run.SuccessCnt += cnt.Success
		res.Run.FailedCnt += cnt.Failed
		res.Run.Canceled += cnt.Canceled
		res.Run.TotalCnt += cnt.Total
		switch dag.Trigger {
		case entity.TriggerManually:
			res.Basic.Manually++
		case entity.TriggerCron:
			res.Basic.Cron++
		case entity.TriggerEvent:
			res.Basic.Event++
		}
	}

	res.Basic.DagCnt = int64(len(dags))
	var dagInss []*entity.DagInstanceGroup

	for i := 0; i < len(dags); i += chunkSize {
		end := min(i+chunkSize, len(dags))

		chunk := dags[i:end]
		chunkIDs := []string{}
		for _, dag := range chunk {
			chunkIDs = append(chunkIDs, dag.ID)
		}

		// 获取数据库其他状态数据
		inss, err := o.mongo.GroupDagInstance(ctx, &mod.GroupInput{
			SearchOptions: []*mod.SearchOption{
				{Field: "dagId", Value: chunkIDs, Condition: "$in"},
				{Field: "status", Value: []entity.DagInstanceStatus{entity.DagInstanceStatusInit,
					entity.DagInstanceStatusRunning,
					entity.DagInstanceStatusBlocked,
					entity.DagInstanceStatusScheduled},
					Condition: "$in"},
			},
			TimeRange: &mod.TimeRangeSearch{
				Begin: *params.StartTime,
				End:   *params.EndTime,
				Field: "createdAt",
			},
			GroupBys:      []string{"dagId", "status"},
			IsSum:         true,
			IsFirst:       true,
			Order:         -1,
			SortBy:        "createdAt",
			ProjectFields: []string{"status"},
		})
		if err != nil {
			log.Warnf("[logic.FullView] GroupDagInstance err, detail: %s", err.Error())
			return res, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
		}
		dagInss = append(dagInss, inss...)
	}

	for _, ins := range dagInss {
		switch ins.DagIns.Status {
		case entity.DagInstanceStatusInit, entity.DagInstanceStatusBlocked, entity.DagInstanceStatusScheduled:
			res.Run.Scheduled += ins.Total
		case entity.DagInstanceStatusRunning:
			res.Run.Running += ins.Total
		}
		res.Run.TotalCnt += ins.Total
	}

	o.memoryCache.Set(cacheKey, res, 30*time.Second)

	return res, nil
}

func (o *observability) RuntimeView(ctx context.Context, params ObservabilityQueryParams, userInfo *drivenadapters.UserInfo) (RuntimeViewRes, error) {
	var (
		err error
		res RuntimeViewRes
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.ReSourceTypeObservability: {perm.DisplayOperation},
		},
	}

	_, err = o.permChecker.CheckPerm(ctx, common.ReSourceTypeObservability, nil, userInfo, opMap)
	if err != nil {
		return res, err
	}

	err = o.PreCheck(ctx, &params)
	if err != nil {
		return res, err
	}

	b, _ := json.Marshal(params)
	cacheKey := fmt.Sprintf("%s:%s:%s", runtimeCacheKey, userInfo.UserID, utils.ComputeHash(string(b)))

	contents, exist := o.memoryCache.Get(cacheKey)
	if exist {
		err = json.Unmarshal([]byte(contents), &res)
		if err != nil {
			return res, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
		}
		return res, nil
	}

	listDagInput := mgnt.QueryParams{
		Type:   *params.Type,
		Page:   *params.Page,
		Limit:  *params.Limit,
		SortBy: common.UpdatedAt,
		Order:  common.DESC,
	}

	if params.Keyword != nil {
		listDagInput.KeyWord = *params.Keyword
	}

	if params.Trigger != nil {
		listDagInput.TriggerType = *params.Trigger
	}

	// 获取当前所有已创建的数据流
	dags, total, err := mgnt.ListDagWithFilters(ctx, listDagInput,
		mgnt.WithBizDomainFilter(o.bizDomain, params.BizDomainID, "", *params.Type, userInfo.TokenID),
		mgnt.WithPermissionFilter(o.permPolicy, userInfo, *params.Type),
		mgnt.WithExistenceFilter())
	if err != nil {
		return res, err
	}

	res.Total = total
	res.Page = *params.Page
	res.Limit = *params.Limit

	if len(dags) == 0 {
		o.memoryCache.Set(cacheKey, res, 30*time.Second)
		return res, nil
	}

	var dagIDs []string
	dagNameMap := make(map[string]string)
	accessorIDs := make(map[string]string)
	for _, dag := range dags {
		dagIDs = append(dagIDs, dag.ID)
		dagNameMap[dag.ID] = dag.Name
		accessorIDs[dag.UserID] = common.User.ToString()
	}

	dvMap, avgMap, err := o.CollectDagMetrics(ctx, *params.StartTime, *params.EndTime, userInfo.TokenID, params.BizDomainID, dagIDs)
	if err != nil {
		return res, err
	}

	accessors, _ := o.usermgnt.GetNameByAccessorIDs(accessorIDs)

	for _, dag := range dags {
		name := dagNameMap[dag.ID]
		cntInfo, ok := dvMap[dag.ID]
		if !ok {
			cntInfo = drivenadapters.StatusCnt{}
		}
		item := RuntimeViewItem{
			ID:            dag.ID,
			Name:          name,
			Creator:       accessors[dag.UserID],
			StatusSummary: &cntInfo,
			Metric:        &RuntimeViewMetric{},
		}

		if cntInfo.Total > 0 {
			item.Metric.FailedRate = math.Round(float64(cntInfo.Failed)/float64(cntInfo.Total)*1000) / 10
			item.Metric.SuccessRate = math.Round(float64(cntInfo.Success)/float64(cntInfo.Total)*1000) / 10
		}
		item.Metric.AvgRunDuration = avgMap[dag.ID]

		res.Datas = append(res.Datas, &item)
	}

	o.memoryCache.Set(cacheKey, res, 30*time.Second)

	return res, nil
}

func (o *observability) RecentRunView(ctx context.Context, params ObservabilityQueryParams, userInfo *drivenadapters.UserInfo) ([]*RuntimeViewItem, error) {
	var (
		err error
		res = []*RuntimeViewItem{}
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.ReSourceTypeObservability: {perm.DisplayOperation},
		},
	}

	_, err = o.permChecker.CheckPerm(ctx, common.ReSourceTypeObservability, nil, userInfo, opMap)
	if err != nil {
		return res, err
	}

	b, _ := json.Marshal(params)
	cacheKey := fmt.Sprintf("%s:%s:%s", runtimeCacheKey, userInfo.UserID, utils.ComputeHash(string(b)))

	contents, exist := o.memoryCache.Get(cacheKey)
	if exist {
		err = json.Unmarshal([]byte(contents), &res)
		if err != nil {
			log.Warnf("[logic.RecentRunView] Unmarshal err, detail: %s", err.Error())
			return res, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
		}
		return res, nil
	}

	listDagInput := mgnt.QueryParams{
		Type: common.DagTypeDataFlow,
	}

	if params.Trigger != nil {
		listDagInput.TriggerType = *params.Trigger
	}

	// 获取当前所有已创建的数据流
	dags, _, err := mgnt.ListDagWithFilters(ctx, listDagInput,
		mgnt.WithBizDomainFilter(o.bizDomain, params.BizDomainID, "", listDagInput.Type, userInfo.TokenID),
		mgnt.WithPermissionFilter(o.permPolicy, userInfo, listDagInput.Type),
		mgnt.WithExistenceFilter())
	if err != nil {
		return res, err
	}

	if len(dags) == 0 {
		o.memoryCache.Set(cacheKey, res, 30*time.Second)
		return res, nil
	}

	dagIDs, dagNameMap := []string{}, make(map[string]string)

	for _, dag := range dags {
		dagNameMap[dag.ID] = dag.Name
		dagIDs = append(dagIDs, dag.ID)
	}

	now := time.Now()
	startTime := now.Add(-time.Hour * 24 * 7).Unix()
	endTime := now.Unix()
	var dagInss []*entity.DagInstanceGroup

	for i := 0; i < len(dagIDs); i += chunkSize {
		end := min(i+chunkSize, len(dagIDs))

		chunkIDs := dagIDs[i:end]
		// 获取数据库其他状态数据
		ins, err := o.mongo.GroupDagInstance(ctx, &mod.GroupInput{
			SearchOptions: []*mod.SearchOption{
				{Field: "dagId", Value: chunkIDs, Condition: "$in"},
				{Field: "status", Value: []entity.DagInstanceStatus{entity.DagInstanceStatusRunning,
					entity.DagInstanceStatusBlocked},
					Condition: "$in"},
			},
			TimeRange: &mod.TimeRangeSearch{
				Begin: startTime,
				End:   endTime,
				Field: "createdAt",
			},
			GroupBys:      []string{"dagId"},
			IsSum:         true,
			IsFirst:       true,
			Order:         -1,
			SortBy:        "createdAt",
			Limit:         10,
			ProjectFields: []string{"dagId", "userid", "createdAt"},
		})
		if err != nil {
			log.Warnf("[logic.RecentRunView] GroupDagInstance err, detail: %s", err.Error())
			return res, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
		}

		dagInss = append(dagInss, ins...)
	}

	if len(dagInss) == 0 {
		o.memoryCache.Set(cacheKey, res, 30*time.Second)
		return res, nil
	}

	sort.Slice(dagInss, func(i, j int) bool {
		return dagInss[i].DagIns.CreatedAt > dagInss[j].DagIns.CreatedAt
	})

	dagIDs, dagInss = dagIDs[0:0], dagInss[0:min(10, len(dagInss))]
	creatorMap := make(map[string]string)
	accessorIDs := make(map[string]string)
	for _, ins := range dagInss {
		dagIDs = append(dagIDs, ins.DagIns.DagID)
		accessorIDs[ins.DagIns.UserID] = common.User.ToString()
		creatorMap[ins.DagIns.DagID] = ins.DagIns.UserID
	}

	dvMap, avgMap, err := o.CollectDagMetrics(ctx, startTime, endTime, userInfo.TokenID, params.BizDomainID, dagIDs)
	if err != nil {
		return res, err
	}

	accessors, _ := o.usermgnt.GetNameByAccessorIDs(accessorIDs)

	res = []*RuntimeViewItem{}
	for _, dagID := range dagIDs {
		name := dagNameMap[dagID]
		cntInfo, ok := dvMap[dagID]
		if !ok {
			cntInfo = drivenadapters.StatusCnt{}
		}
		item := RuntimeViewItem{
			ID:            dagID,
			Name:          name,
			Creator:       accessors[creatorMap[dagID]],
			StatusSummary: &cntInfo,
			Metric:        &RuntimeViewMetric{},
		}

		if cntInfo.Total > 0 {
			item.Metric.FailedRate = math.Round(float64(cntInfo.Failed)/float64(cntInfo.Total)*1000) / 10
			item.Metric.SuccessRate = math.Round(float64(cntInfo.Success)/float64(cntInfo.Total)*1000) / 10
		}
		item.Metric.AvgRunDuration = avgMap[dagID]

		res = append(res, &item)
	}

	o.memoryCache.Set(cacheKey, res, 30*time.Second)

	return res, nil

}

// CollectDagMetrics 获取数据流流程指标
func (o *observability) CollectDagMetrics(ctx context.Context, startTime, endTime int64, token, bizDomainID string, dagIDs []string) (map[string]drivenadapters.StatusCnt, map[string]float64, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	runtimeView, err := o.uniquery.QueryDagStatusCount(ctx, startTime, endTime, token, drivenadapters.QueryFileds{
		DagIDs:      dagIDs,
		BizDomainID: bizDomainID,
	})
	if err != nil {
		log.Warnf("[logic.CollectDagMetrics] QueryDagStatusCount err, detail: %s", err.Error())
		return nil, nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, err.Error())
	}

	runtimmAvg, err := o.uniquery.QueryDagRunTimeAvg(ctx, startTime, endTime, token, drivenadapters.QueryFileds{
		DagIDs:      dagIDs,
		BizDomainID: bizDomainID,
	})
	if err != nil {
		log.Warnf("[logic.CollectDagMetrics] QueryDagRunTimeAvg err, detail: %s", err.Error())
		return nil, nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, err.Error())
	}

	dvMap := runtimeView.ToMap()
	avgMap := runtimmAvg.AvgTimeToMap()

	// 获取数据库其他状态数据
	dagInss, err := o.mongo.GroupDagInstance(ctx, &mod.GroupInput{
		SearchOptions: []*mod.SearchOption{
			{Field: "dagId", Value: dagIDs, Condition: "$in"},
			{Field: "status", Value: []entity.DagInstanceStatus{entity.DagInstanceStatusInit,
				entity.DagInstanceStatusRunning,
				entity.DagInstanceStatusBlocked,
				entity.DagInstanceStatusScheduled},
				Condition: "$in"},
		},
		TimeRange: &mod.TimeRangeSearch{
			Begin: startTime,
			End:   endTime,
			Field: "createdAt",
		},
		GroupBys:      []string{"dagId", "status"},
		IsSum:         true,
		IsFirst:       true,
		Order:         -1,
		SortBy:        "createdAt",
		ProjectFields: []string{"dagId", "status"},
	})
	if err != nil {
		log.Warnf("[logic.CollectDagMetrics] GroupDagInstance err, detail: %s", err.Error())
		return nil, nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, aerr.DescKeyErrorDepencyService, nil)
	}

	for _, v := range dagInss {
		item := dvMap[v.DagIns.DagID]
		switch v.DagIns.Status {
		case entity.DagInstanceStatusInit, entity.DagInstanceStatusScheduled:
			item.Init += v.Total
		case entity.DagInstanceStatusRunning:
			item.Running += v.Total
		case entity.DagInstanceStatusBlocked:
			item.Blocked += v.Total
		}

		item.Total += v.Total

		dvMap[v.DagIns.DagID] = item
	}
	return dvMap, avgMap, nil
}

// PreCheck 参数预检查
func (o *observability) PreCheck(ctx context.Context, params *ObservabilityQueryParams) error {
	// 默认查询数据流流程
	if params.Type == nil {
		params.Type = ptr.String(common.DagTypeDataFlow)
	}

	// 如果查询的开始时间不设置, 默认查询最近7天数据
	if params.StartTime == nil {
		now := time.Now()
		params.StartTime = ptr.Int64(now.Add(-time.Hour * 24 * 7).Unix())
		params.EndTime = ptr.Int64(now.Unix())
	}

	// 如果查询的结束时间不设置, 默认查询当前时间
	if params.EndTime == nil {
		params.EndTime = ptr.Int64(time.Now().Unix())
	}

	if *params.StartTime >= *params.EndTime {
		return ierr.NewPublicRestError(ctx, ierr.PErrorBadRequest, ierr.PErrorBadRequest, "start_time must be less than end_time")
	}

	return nil
}

// IsVisible 概览可见性判断
func (o *observability) IsVisible(ctx context.Context) bool {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	isVisible, err := o.uniquery.CheckDataViewExist(ctx, drivenadapters.DefaultDataViewName)
	if err != nil {
		log.Warnf("[logic.IsVisible] CheckDataViewExist err, detail: %s", err.Error())
		return false
	}
	return isVisible
}
