package auth

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// QueryBuilder 查询构建器 - 提供更简洁的API来使用SelectListWithAuthBatch
// 使用示例:
// result, err := NewQueryBuilder[ModelType, *ModelType](ctx, authService, accessor, resourceType, operations).
//
//	SetPage(page, pageSize).
//	SetAll(all).
//	SetQueryFunctions(queryTotal, queryBatch).
//	SetFilteredQueryFunctions(queryTotalWithIDs, queryBatchWithIDs).
//	Execute()
//
// 新版本特性：
// 1. 自动选择查询策略：权限ID少时使用分批IN查询，多时使用增量拉取
// 2. 保证排序和分页的准确性
// 3. 避免数据库IN查询限制问题
// 4. 支持跨数据库兼容
// 参数解释：
// ctx: 上下文对象，用于取消操作和传递上下文信息
// page: 当前页码（从1开始）
// pageSize: 每页数据量
// all: 是否查询所有数据（不分页）
// queryTotalFunc: 查询总数据量的函数
// queryBatchFunc: 分页查询数据的函数
// queryBatchWithIDsFunc: 基于权限ID列表分页查询数据的函数（支持IN查询）
// queryTotalWithIDsFunc: 基于权限ID列表查询总数据量的函数
// resourceListFunc: 获取用户有权限的资源ID列表的函数
type QueryBuilder[T any, PT interfaces.PtrBizIdentifiable[T]] struct {
	page                           int
	pageSize                       int
	all                            bool
	queryTotalFunc                 QueryTotalFunc
	queryBatchFunc                 QueryBatchFunc[T, PT]
	queryTotalWithIDsFunc          QueryTotalWithIDsFunc
	queryBatchWithIDsFunc          QueryBatchWithIDsFunc[T, PT]
	resourceListFunc               ResourceListFunc
	businessDomainResourceListFunc BusinessDomainResourceListFunc
}

// QueryTotalFunc 查询总数据量的函数
type QueryTotalFunc func(ctx context.Context) (int64, error)

// QueryBatchFunc 分页查询数据的函数
type QueryBatchFunc[T any, PT interfaces.PtrBizIdentifiable[T]] func(ctx context.Context, pageSize, offset int, cursorValue *T) ([]PT, error)

// QueryTotalWithIDsFunc 查询基于权限ID列表的总数据量的函数
type QueryTotalWithIDsFunc func(ctx context.Context, ids []string) (int64, error)

// QueryBatchWithIDsFunc 分页查询基于权限ID列表的数据的函数
type QueryBatchWithIDsFunc[T any, PT interfaces.PtrBizIdentifiable[T]] func(ctx context.Context, pageSize, offset int, ids []string, cursorValue *T) ([]PT, error)

// ResourceListFunc 获取用户有权限的资源ID列表的函数
type ResourceListFunc func(ctx context.Context) ([]string, error)

// BusinessDomainResourceListFunc 获取业务域资源ID列表的函数类型
type BusinessDomainResourceListFunc func(ctx context.Context) ([]string, error)

// NewQueryBuilder 创建一个新的查询构建器
func NewQueryBuilder[T any, PT interfaces.PtrBizIdentifiable[T]]() *QueryBuilder[T, PT] {
	return &QueryBuilder[T, PT]{
		page:     1,                          // 默认第一页
		pageSize: interfaces.DefaultPageSize, // 默认每页大小
		all:      false,                      // 默认不分页
	}
}

// SetPage 设置分页参数
func (b *QueryBuilder[T, PT]) SetPage(page, pageSize int) *QueryBuilder[T, PT] {
	if page > 0 {
		b.page = page
	}
	if pageSize > 0 {
		b.pageSize = pageSize
	}
	return b
}

// SetAll 设置是否返回全部数据
func (b *QueryBuilder[T, PT]) SetAll(all bool) *QueryBuilder[T, PT] {
	b.all = all
	return b
}

// SetQueryFunctions 设置基本查询函数
func (b *QueryBuilder[T, PT]) SetQueryFunctions(
	queryTotal QueryTotalFunc,
	queryBatch QueryBatchFunc[T, PT],
) *QueryBuilder[T, PT] {
	b.queryTotalFunc = queryTotal
	b.queryBatchFunc = queryBatch
	return b
}

// SetAuthFilter 设置权限过滤函数
func (b *QueryBuilder[T, PT]) SetAuthFilter(resourceListFunc ResourceListFunc) *QueryBuilder[T, PT] {
	b.resourceListFunc = resourceListFunc
	return b
}

// SetBusinessDomainFilter 设置业务域资源过滤函数
func (b *QueryBuilder[T, PT]) SetBusinessDomainFilter(businessDomainResourceListFunc BusinessDomainResourceListFunc) *QueryBuilder[T, PT] {
	b.businessDomainResourceListFunc = businessDomainResourceListFunc
	return b
}

// SetFilteredQueryFunctions 设置带权限过滤的查询函数
func (b *QueryBuilder[T, PT]) SetFilteredQueryFunctions(
	queryTotalWithIDs QueryTotalWithIDsFunc,
	queryBatchWithIDs QueryBatchWithIDsFunc[T, PT],
) *QueryBuilder[T, PT] {
	b.queryTotalWithIDsFunc = queryTotalWithIDs
	b.queryBatchWithIDsFunc = queryBatchWithIDs
	return b
}

// Execute 执行权限查询
func (b *QueryBuilder[T, PT]) Execute(ctx context.Context) (*interfaces.QueryResponse[T], error) {
	// 参数验证
	if b.queryTotalFunc == nil || b.queryBatchFunc == nil {
		return nil, fmt.Errorf("queryTotalFunc and queryBatchFunc are required")
	}
	// 调用SelectListWithAuthBatchWithThresholds执行查询
	return b.SelectListWithAuthBatchWithThresholds(ctx)
}

// getFilteredResourceIDs 获取经过业务域和权限双重过滤的资源ID列表
// 业务域过滤作为第一层，权限过滤作为第二层
// 核心原则：业务域是第一层过滤限制，即使有权限访问所有资源，如果业务域没有结果，也必须返回空列表
func (b *QueryBuilder[T, PT]) getFilteredResourceIDs(ctx context.Context) ([]string, bool, error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, nil)

	var (
		businessDomainIDs      []string
		authorizedIDs          []string
		hasFullAccess          bool
		businessDomainBypassed bool
	)

	// 串行调用业务域资源列表函数
	if b.businessDomainResourceListFunc != nil {
		ids, err := b.businessDomainResourceListFunc(ctx)
		if err != nil {
			return nil, false, err
		}
		businessDomainIDs = ids
		for _, id := range ids {
			if id == interfaces.ResourceIDAll {
				businessDomainBypassed = true
				break
			}
		}
	}

	// 早期终止策略：如果业务域没有返回任何资源ID，直接返回空列表
	// 这是第一层过滤的核心体现，无论权限如何，业务域为空时直接返回空
	if len(businessDomainIDs) == 0 && b.businessDomainResourceListFunc != nil && !businessDomainBypassed {
		return []string{}, false, nil
	}

	// 串行调用权限资源列表函数
	if b.resourceListFunc != nil {
		ids, err := b.resourceListFunc(ctx)
		if err != nil {
			return nil, false, err
		}
		authorizedIDs = ids
		// 检查是否有全部权限
		for _, id := range ids {
			if id == interfaces.ResourceIDAll {
				hasFullAccess = true
				break
			}
		}
	} else {
		// 如果没有设置权限过滤函数，默认有全部权限
		hasFullAccess = true
	}

	// 计算过滤后的资源ID列表
	var filteredIDs []string

	if hasFullAccess {
		// 如果有权限访问所有资源，则只应用业务域过滤
		// 业务域作为第一层过滤，即使有权限，也只能访问业务域内的资源
		if businessDomainBypassed {
			return nil, true, nil
		}
		filteredIDs = businessDomainIDs
	} else {
		// 有限权限情况下，计算业务域和权限的交集
		// utils.CalculateIntersection函数内部会处理空列表的情况
		if businessDomainBypassed {
			filteredIDs = authorizedIDs
		} else {
			filteredIDs = utils.CalculateIntersection(businessDomainIDs, authorizedIDs)
		}
	}

	return filteredIDs, false, nil
}

// SelectListWithAuthBatchWithThresholds 带阈值参数的权限查询函数
func (b *QueryBuilder[T, PT]) SelectListWithAuthBatchWithThresholds(ctx context.Context) (resp *interfaces.QueryResponse[T], err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 设置默认参数
	if b.page <= 0 {
		b.page = 1
	}
	if b.pageSize <= 0 {
		b.pageSize = 10
	}

	// 使用统一的方法获取经过业务域和权限双重过滤的资源ID列表
	filteredIDs, hasFullAccess, err := b.getFilteredResourceIDs(ctx)
	if err != nil {
		return nil, err
	}
	// 2. 根据权限类型选择不同的查询策略
	if hasFullAccess {
		// 有全部权限，直接使用分页查询
		totalCount, err := b.queryTotalFunc(ctx)
		if err != nil {
			return nil, err
		}
		if b.all {
			// 需要获取所有有权限的数据
			var allData []PT
			queryTimes := totalCount / int64(b.pageSize)
			if totalCount%int64(b.pageSize) != 0 {
				queryTimes++
			}
			var processTimes int64
			var cursorValue *T
			for processTimes <= queryTimes {
				// 加载一批数据
				var batchData []PT
				batchData, err = b.queryBatchFunc(ctx, b.pageSize, 0, cursorValue)
				if err != nil {
					return nil, err
				}

				if len(batchData) == 0 {
					break // 没有更多数据
				}
				allData = append(allData, batchData...)
				cursorValue = allData[len(allData)-1]
				processTimes++
			}
			// 构造响应数据
			pageData := make([]*T, len(allData))
			for i, item := range allData {
				pageData[i] = (*T)(item)
			}

			return &interfaces.QueryResponse[T]{
				Data: pageData,
				CommonPageResult: interfaces.CommonPageResult{
					TotalCount: len(pageData),
					Page:       b.page,
					PageSize:   b.pageSize,
					TotalPage:  1,
					HasNext:    false,
					HasPrev:    false,
				},
			}, nil
		}

		// 计算分页
		totalPages := int((totalCount + int64(b.pageSize) - 1) / int64(b.pageSize))
		hasNext := b.page < totalPages
		hasPrev := b.page > 1

		// 计算偏移量和限制
		offset := (b.page - 1) * b.pageSize
		var data []PT
		data, err = b.queryBatchFunc(ctx, b.pageSize, offset, nil)
		if err != nil {
			return nil, err
		}

		// 构造响应数据
		pageData := make([]*T, len(data))
		for i, item := range data {
			pageData[i] = (*T)(item)
		}

		return &interfaces.QueryResponse[T]{
			Data: pageData,
			CommonPageResult: interfaces.CommonPageResult{
				TotalCount: int(totalCount),
				Page:       b.page,
				PageSize:   b.pageSize,
				TotalPage:  totalPages,
				HasNext:    hasNext,
				HasPrev:    hasPrev,
			},
		}, nil
	}

	// 3. 有限权限情况
	// 没有权限ID时返回空结果
	if len(filteredIDs) == 0 {
		return &interfaces.QueryResponse[T]{
			Data: []*T{},
			CommonPageResult: interfaces.CommonPageResult{
				TotalCount: 0,
				Page:       b.page,
				PageSize:   b.pageSize,
				TotalPage:  0,
				HasNext:    false,
				HasPrev:    false,
			},
		}, nil
	}

	// 根据权限ID数量选择查询策略
	if len(filteredIDs) <= MaxInQuerySize {
		// 权限ID数量较少，使用分批IN查询
		return b.selectListWithBatchInQueryWithThresholds(ctx, filteredIDs)
	} else {
		// 权限ID数量较多，使用增量拉取
		return b.selectListWithIncrementalFetch(ctx, filteredIDs)
	}
}

// selectListWithBatchInQueryWithThresholds 带阈值参数的分批IN查询
func (b *QueryBuilder[T, PT]) selectListWithBatchInQueryWithThresholds(ctx context.Context, filteredIDs []string) (resp *interfaces.QueryResponse[T], err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 如果提供了数据库过滤函数，直接使用
	if b.queryBatchWithIDsFunc != nil && b.queryTotalWithIDsFunc != nil {
		// 查询有权限的数据总数
		authorizedTotalCount, err := b.queryTotalWithIDsFunc(ctx, filteredIDs)
		if err != nil {
			return nil, err
		}

		if b.all {
			// 需要获取所有有权限的数据
			var allData []PT
			allData, err = b.queryBatchWithIDsFunc(ctx, int(authorizedTotalCount), 0, filteredIDs, nil)
			if err != nil {
				return nil, err
			}

			// 构造响应数据
			pageData := make([]*T, len(allData))
			for i, item := range allData {
				pageData[i] = (*T)(item)
			}

			return &interfaces.QueryResponse[T]{
				Data: pageData,
				CommonPageResult: interfaces.CommonPageResult{
					TotalCount: len(allData),
					Page:       1,
					PageSize:   len(allData),
					TotalPage:  1,
					HasNext:    false,
					HasPrev:    false,
				},
			}, nil
		}

		// 分页查询有限权限数据, 计算分页
		totalPages := int((authorizedTotalCount + int64(b.pageSize) - 1) / int64(b.pageSize))
		hasNext := b.page < totalPages
		hasPrev := b.page > 1

		// 如果请求的页码超出范围，返回空数据
		if b.page > totalPages {
			return &interfaces.QueryResponse[T]{
				Data: []*T{},
				CommonPageResult: interfaces.CommonPageResult{
					TotalCount: int(authorizedTotalCount),
					Page:       b.page,
					PageSize:   b.pageSize,
					TotalPage:  totalPages,
					HasNext:    false,
					HasPrev:    true,
				},
			}, nil
		}

		// 计算偏移量和限制
		offset := (b.page - 1) * b.pageSize

		// 查询指定页的数据，数据库层会保证排序准确性
		pageDataList, err := b.queryBatchWithIDsFunc(ctx, b.pageSize, offset, filteredIDs, nil)
		if err != nil {
			return nil, err
		}

		// 构造响应数据
		pageData := make([]*T, len(pageDataList))
		for i, item := range pageDataList {
			pageData[i] = (*T)(item)
		}

		return &interfaces.QueryResponse[T]{
			Data: pageData,
			CommonPageResult: interfaces.CommonPageResult{
				TotalCount: int(authorizedTotalCount),
				Page:       b.page,
				PageSize:   b.pageSize,
				TotalPage:  totalPages,
				HasNext:    hasNext,
				HasPrev:    hasPrev,
			},
		}, nil
	}

	// 如果没有提供数据库过滤函数，为保证全局排序一致性，改用增量拉取
	return b.selectListWithIncrementalFetch(ctx, filteredIDs)
}

// selectListWithIncrementalFetch 使用增量拉取进行权限过滤
// 适用于权限ID数量较多的情况（>MaxInQuerySize）
func (b *QueryBuilder[T, PT]) selectListWithIncrementalFetch(ctx context.Context, filteredIDs []string) (resp *interfaces.QueryResponse[T], err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 构建权限映射用于内存过滤
	authMap := make(map[string]bool, len(filteredIDs))
	for _, id := range filteredIDs {
		authMap[id] = true
	}

	if b.all {
		// 需要获取所有有权限的数据
		var totalCount int64
		totalCount, err = b.queryTotalFunc(ctx)
		if err != nil {
			return nil, err
		}

		// 分批加载所有数据并过滤
		filteredData := make([]PT, 0, interfaces.MaxQuerySize)
		queryTimes := totalCount / int64(interfaces.MaxQuerySize)
		if totalCount%int64(interfaces.MaxQuerySize) != 0 {
			queryTimes++
		}
		var processTimes int64
		var cursorValue *T
		for processTimes <= queryTimes {
			// 加载一批数据
			var batchData []PT
			batchData, err = b.queryBatchFunc(ctx, interfaces.MaxQuerySize, 0, cursorValue)
			if err != nil {
				return nil, err
			}

			if len(batchData) == 0 {
				break // 没有更多数据
			}
			// 过滤出有权限的数据
			for _, item := range batchData {
				if item != nil && authMap[item.GetBizID()] {
					filteredData = append(filteredData, item)
				}
			}
			cursorValue = batchData[len(batchData)-1]
			processTimes++

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}

		// 构造响应数据
		pageData := make([]*T, len(filteredData))
		for i, item := range filteredData {
			pageData[i] = (*T)(item)
		}

		return &interfaces.QueryResponse[T]{
			Data: pageData,
			CommonPageResult: interfaces.CommonPageResult{
				TotalCount: len(filteredData),
				Page:       1,
				PageSize:   len(filteredData),
				TotalPage:  1,
				HasNext:    false,
				HasPrev:    false,
			},
		}, nil
	}

	// 分页查询有限权限数据
	targetStart := (b.page - 1) * b.pageSize
	targetEnd := targetStart + b.pageSize
	var pageData []*T

	// 为了高效分页，我们需要找到足够的有权限的数据
	// offset := 0
	foundCount := 0 // 已找到的有权限的数据数量

	// 首先获取总记录数以便计算总页数
	allTotalCount, err := b.queryTotalFunc(ctx)
	if err != nil {
		return nil, err
	}

	// 分批加载数据直到找到足够的有权限的数据或者处理完所有数据
	queryTimes := allTotalCount / int64(interfaces.MaxQuerySize)
	if allTotalCount%int64(interfaces.MaxQuerySize) != 0 {
		queryTimes++
	}
	var cursorValue *T
	var processTimes int64
	for processTimes <= queryTimes && foundCount < targetEnd {
		// 加载一批数据
		batchData, err := b.queryBatchFunc(ctx, interfaces.MaxQuerySize, 0, cursorValue)
		if err != nil {
			return nil, err
		}
		if len(batchData) == 0 {
			break // 没有更多数据
		}
		// 过滤并记录有权限的数据
		for i := range batchData {
			item := batchData[i]
			if item != nil && authMap[item.GetBizID()] {
				if foundCount >= targetStart && foundCount < targetEnd {
					// 在目标范围内，加入结果集
					t := (*T)(item)
					pageData = append(pageData, t)
				}
				foundCount++
			}
			batchData[i] = nil
			cursorValue = (*T)(item)
		}
		processTimes++
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	// 计算分页信息
	totalPages := (foundCount + b.pageSize - 1) / b.pageSize
	hasNext := b.page < totalPages
	hasPrev := b.page > 1

	return &interfaces.QueryResponse[T]{
		Data: pageData,
		CommonPageResult: interfaces.CommonPageResult{
			TotalCount: foundCount,
			Page:       b.page,
			PageSize:   b.pageSize,
			TotalPage:  totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}, nil
}
