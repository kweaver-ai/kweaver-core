package mgnt

import (
	"context"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierr "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

// FilterFunc 过滤函数：输入 dagIDs，返回过滤后的 dagIDs
type FilterFunc func(ctx context.Context, dagIDs []string) ([]string, error)

// ListDagOption 查询选项
type ListDagOption func(*listDagConfig)

type listDagConfig struct {
	AuthEnable bool
	filters    []FilterFunc
}

func NewListDagConfig() *listDagConfig {
	return &listDagConfig{
		AuthEnable: common.NewConfig().IsAuthEnabled(),
		filters:    []FilterFunc{},
	}
}

// WithBizDomainFilter 添加业务域过滤
func WithBizDomainFilter(bizDomain drivenadapters.BusinessDomain, bizDomainID, resourceID, dType, token string) ListDagOption {
	return func(c *listDagConfig) {
		filter := func(ctx context.Context, dagIDs []string) ([]string, error) {
			// 获取业务域下指定资源类型下的所有dag
			bizDomainParams := drivenadapters.BizDomainResourceQuery{
				BizDomainResourceParams: drivenadapters.BizDomainResourceParams{
					BizDomainID:  bizDomainID,
					ResourceID:   resourceID,
					ResourceType: perm.DataFlowResourceType,
				},
				Limit: -1,
			}
			domainResources, err := bizDomain.ListResource(ctx, bizDomainParams, token)
			if err != nil {
				return []string{}, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, err.Error())
			}

			if domainResources.Total == 0 {
				return []string{}, nil
			}

			resourceIDs := domainResources.Items.GetIDs(dType)

			// 如果是第一个过滤器，直接返回
			if dagIDs == nil {
				return resourceIDs, nil
			}

			// 否则取交集
			return utils.GetIntersection(dagIDs, resourceIDs), nil
		}
		c.filters = append(c.filters, filter)
	}
}

// WithPermissionFilter 添加权限过滤
func WithPermissionFilter(permPolicy perm.PermPolicyHandler, userInfo *drivenadapters.UserInfo, resourceType string) ListDagOption {
	return func(c *listDagConfig) {
		filter := func(ctx context.Context, dagIDs []string) ([]string, error) {
			// 数据管理员直接返回当前域下的所有dagID
			isDataAdmin, err := permPolicy.IsDataAdmin(ctx, userInfo.UserID, userInfo.AccountType)
			if err != nil {
				return []string{}, err
			}

			if isDataAdmin {
				return dagIDs, nil
			}

			ids, err := permPolicy.ListResource(ctx, userInfo.UserID, userInfo.AccountType, perm.DataFlowResourceType, perm.ListOperation)
			if err != nil {
				return nil, err
			}

			if len(*ids) == 0 {
				return []string{}, nil
			}

			filterType := resourceType
			if strings.TrimSpace(filterType) == "" {
				filterType = common.DagTypeDefault
			}

			resourceMap := ids.ToMap(filterType)
			permIDs := make([]string, 0, len(resourceMap))
			for k := range resourceMap {
				permIDs = append(permIDs, k)
			}

			// 如果是第一个过滤器，直接返回
			if dagIDs == nil {
				return permIDs, nil
			}

			// 否则取交集
			return utils.GetIntersection(dagIDs, permIDs), nil
		}

		c.filters = append(c.filters, filter)
	}
}

// WithExistenceFilter 添加存在性检查
func WithExistenceFilter() ListDagOption {
	return func(c *listDagConfig) {
		filter := func(ctx context.Context, dagIDs []string) ([]string, error) {
			if len(dagIDs) == 0 {
				return dagIDs, nil
			}

			var err error
			ctx, span := trace.StartInternalSpan(ctx)
			defer func() { trace.TelemetrySpanEnd(span, err) }()
			log := traceLog.WithContext(ctx)
			store := mod.GetStore()

			const chunkSize = 1024
			var allExistingIDs []string

			for i := 0; i < len(dagIDs); i += chunkSize {
				end := min(i+chunkSize, len(dagIDs))

				chunk := dagIDs[i:end]
				existingIDs, err := store.ListExistDagID(ctx, chunk)
				if err != nil {
					log.Warnf("[logic.WithExistenceFilter] ListExistDagID err, detail: %s", err.Error())
					return nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
				}

				allExistingIDs = append(allExistingIDs, existingIDs...)
			}

			return allExistingIDs, nil
		}
		c.filters = append(c.filters, filter)
	}
}

// WithSharedDagFilter 添加共享dag过滤，根据triggerType, type, keyword, userID, triggerExclude, accessors过滤dag
func WithSharedDagFilter(param QueryParams) ListDagOption {
	return func(c *listDagConfig) {
		filter := func(ctx context.Context, dagIDs []string) ([]string, error) {
			if len(dagIDs) == 0 {
				return dagIDs, nil
			}

			var err error
			ctx, span := trace.StartInternalSpan(ctx)
			defer func() { trace.TelemetrySpanEnd(span, err) }()
			log := traceLog.WithContext(ctx)
			store := mod.GetStore()

			const chunkSize = 1024
			var allExistingIDs []string

			listDagInput := &mod.ListDagInput{
				TriggerType:  param.TriggerType,
				TriggerTypes: param.TriggerTypes,
				Type:         param.Type,
				KeyWord:      param.KeyWord,
				SelectField:  []string{"_id"},
			}

			if param.UserID != "" {
				listDagInput.UserID = param.UserID
			}

			if len(param.TriggerExclude) > 0 {
				listDagInput.TriggerExclude = param.TriggerExclude
			}

			if len(param.Accessors) > 0 {
				listDagInput.Accessors = param.Accessors
			}

			if param.Scope != "" {
				listDagInput.Scope = param.Scope
			}

			for i := 0; i < len(dagIDs); i += chunkSize {
				end := min(i+chunkSize, len(dagIDs))

				chunk := dagIDs[i:end]
				listDagInput.DagIDs = chunk
				dags, err := store.ListDag(ctx, listDagInput)
				if err != nil {
					log.Warnf("[logic.WithSharedDagFilter] ListDag err, detail: %s", err.Error())
					return nil, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
				}

				for _, dag := range dags {
					allExistingIDs = append(allExistingIDs, dag.ID)
				}

			}

			return allExistingIDs, nil
		}
		c.filters = append(c.filters, filter)
	}
}

// PageDags 分页查询
func PageDags(ctx context.Context, store mod.Store, dagIDs []string, queryInput *mod.ListDagInput) ([]*entity.Dag, error) {
	if len(dagIDs) <= 1024 {
		queryInput.DagIDs = dagIDs
		return store.ListDag(ctx, queryInput)
	}
	return pageDagsIncr(ctx, store, dagIDs, queryInput)
}

// pageDagsIncr 增量获取Dag列表
func pageDagsIncr(ctx context.Context, store mod.Store, dagIDs []string, queryInput *mod.ListDagInput) ([]*entity.Dag, error) {
	dagIDSet := make(map[string]struct{}, len(dagIDs))
	for _, id := range dagIDs {
		dagIDSet[id] = struct{}{}
	}

	var dags []*entity.Dag
	threshold := int(queryInput.Limit * (queryInput.Offset + 1))
	isAll := queryInput.Limit <= 0
	queryInput.Limit = utils.IfNot(queryInput.Limit <= 0, 1024, queryInput.Limit)
	page := int64(0)

	for {
		queryInput.Offset = page
		chunks, err := store.ListDag(ctx, queryInput)
		if err != nil {
			return nil, err
		}

		if len(chunks) == 0 {
			break
		}

		for _, chunk := range chunks {
			if _, ok := dagIDSet[chunk.ID]; ok {
				dags = append(dags, chunk)
				if !isAll && len(dags) >= threshold {
					goto DONE
				}
			}
		}
		page++
	}

DONE:
	if isAll {
		return dags, nil
	}

	startIndex := threshold - int(queryInput.Limit)
	if startIndex >= len(dags) {
		return []*entity.Dag{}, nil
	}

	return dags[startIndex:], nil
}

// applyFilters 依次执行所有过滤器
func (c *listDagConfig) ApplyFilters(ctx context.Context) ([]string, error) {
	var dagIDs []string = nil

	for _, filter := range c.filters {
		var err error
		dagIDs, err = filter(ctx, dagIDs)
		if err != nil {
			return nil, err
		}

		// 如果已经没有数据了，提前退出
		if len(dagIDs) == 0 {
			return dagIDs, nil
		}
	}

	return dagIDs, nil
}

func listDagsDirectly(ctx context.Context, store mod.Store, queryInput *mod.ListDagInput) ([]*entity.Dag, int64, error) {
	total, err := store.ListDagCount(ctx, queryInput)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*entity.Dag{}, 0, nil
	}

	dags, err := store.ListDag(ctx, queryInput)
	if err != nil {
		return nil, 0, err
	}

	return dags, total, nil
}

// ListDagWithFilters 根据过滤器获取Dag列表
func ListDagWithFilters(ctx context.Context, param QueryParams, opts ...ListDagOption) ([]*entity.Dag, int64, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 1. 应用选项，构建过滤器列表
	config := NewListDagConfig()

	for _, opt := range opts {
		opt(config)
	}

	listDagInput := &mod.ListDagInput{
		Offset:       param.Page,
		Limit:        param.Limit,
		Order:        -1,
		TriggerType:  param.TriggerType,
		TriggerTypes: param.TriggerTypes,
		Type:         param.Type,
		KeyWord:      param.KeyWord,
	}

	if param.UserID != "" {
		listDagInput.UserID = param.UserID
	}

	if len(param.TriggerExclude) > 0 {
		listDagInput.TriggerExclude = param.TriggerExclude
	}

	if len(param.Accessors) > 0 {
		listDagInput.Accessors = param.Accessors
	}

	if param.Scope != "" {
		listDagInput.Scope = param.Scope
	}

	if param.Order == common.ASC {
		listDagInput.Order = 1
	}

	sortMap := map[string]string{
		common.Updated_At: common.UpdatedAt,
		common.Created_At: common.CreatedAt,
		common.Name:       common.Name,
	}
	if field, ok := sortMap[param.SortBy]; ok {
		listDagInput.SortBy = field
	}

	if !config.AuthEnable {
		dags, total, err := listDagsDirectly(ctx, mod.GetStore(), listDagInput)
		if err != nil {
			log.Warnf("[logic.ListDagWithFilters] listDagsDirectly err, detail: %s", err.Error())
			return nil, 0, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
		}

		return dags, total, nil
	}

	// 2. 依次执行所有过滤器
	dagIDs, err := config.ApplyFilters(ctx)
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(dagIDs))
	if total == 0 {
		return []*entity.Dag{}, 0, nil
	}

	dags, err := PageDags(ctx, mod.GetStore(), dagIDs, listDagInput)
	if err != nil {
		log.Warnf("[logic.ListDagWithFilters] PageDags err, detail: %s", err.Error())
		return nil, 0, ierr.NewPublicRestError(ctx, ierr.PErrorInternalServerError, ierr.PErrorInternalServerError, nil)
	}

	return dags, total, nil
}
