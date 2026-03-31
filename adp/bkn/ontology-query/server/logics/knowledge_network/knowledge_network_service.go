// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knowledge_network

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	"ontology-query/logics"
	"ontology-query/logics/object_type"
)

var (
	knServiceOnce sync.Once
	knService     interfaces.KnowledgeNetworkService
)

type knowledgeNetworkService struct {
	appSetting *common.AppSetting
	omAccess   interfaces.OntologyManagerAccess
	ots        interfaces.ObjectTypeService
	uAccess    interfaces.UniqueryAccess
}

func NewKnowledgeNetworkService(appSetting *common.AppSetting) interfaces.KnowledgeNetworkService {
	knServiceOnce.Do(func() {
		knService = &knowledgeNetworkService{
			appSetting: appSetting,
			omAccess:   logics.OMA,
			ots:        object_type.NewObjectTypeService(appSetting),
			uAccess:    logics.UA,
		}
	})
	return knService
}

func (kns *knowledgeNetworkService) SearchSubgraph(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnSource) (interfaces.ObjectSubGraph, error) {

	// 1.获取对象类信息
	ctx, span := ar_trace.Tracer.Start(ctx, "查询对象子图")
	var resps interfaces.ObjectSubGraph

	// 1. 在指定的业务知识网络下，根据起点对象类、方向、路径长度获取所有路径。
	typePaths, err := kns.omAccess.GetRelationTypePathsBaseOnSource(ctx, query.KNID, query.Branch,
		interfaces.PathsQueryBaseOnSource{
			ConceptGroups:     query.ConceptGroups,
			SourceObjecTypeId: query.SourceObjecTypeId,
			Direction:         query.Direction,
			PathLength:        query.PathLength,
		})
	if err != nil {
		logger.Errorf("GetRelationTypePathsBaseOnSource error: %s", err.Error())

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("kn_id").String(query.KNID))
		span.SetAttributes(attribute.Key("branch").String(query.Branch))
		span.SetStatus(codes.Error, "Get RelationTypePathsBaseOnSource error")
		span.End()
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get RelationTypePathsBaseOnSource error: %v", err))

		return resps, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed).WithErrorDetails(err.Error())
	}

	// 2. 检索起点对象类的对象实例
	startObjectQuery := &interfaces.ObjectQueryBaseOnObjectType{
		ActualCondition: query.ActualCondition,
		PageQuery:       query.PageQuery,
		KNID:            query.KNID,
		Branch:          query.Branch,
		ObjectTypeID:    query.SourceObjecTypeId,
		CommonQueryParameters: interfaces.CommonQueryParameters{
			IncludeTypeInfo:    true,
			IncludeLogicParams: query.IncludeLogicParams,
			IgnoringStore:      query.IgnoringStore,
			// ExcludeSystemProperties: query.ExcludeSystemProperties,
		},
	}
	if startObjectQuery.Limit == 0 {
		startObjectQuery.Limit = interfaces.DEFAULT_LIMIT
	}

	// 排序字段的校验在获取对象类的对象数据的时候校验，在当前层不用校验
	startObjects, err := kns.ots.GetObjectsByObjectTypeID(ctx, startObjectQuery)
	if err != nil {
		return resps, err
	}

	// 3. 遍历路径，跟起点对象实例沿着路径逐个对象类查询对象实例
	query.PathQuotaManager = &interfaces.PathQuotaManager{
		TotalLimit:         query.TotalLimit, // 对象路径总长度
		GlobalCount:        0,                // 对象路径数量0
		UsedQuota:          sync.Map{},
		RequestPathTypeNum: len(typePaths),
	}

	// 起点类已经查询完成，limit已经得到，后续的路径探索用系统默认的最大值进行探索
	query.PageQuery.Limit = interfaces.MAX_PATHS
	objectGraph, err := kns.buildObjectSubgraph(ctx, query, typePaths, startObjects)
	if err != nil {
		return resps, err
	}

	// 4. 组装最终结果
	objectGraph.TotalCount = startObjects.TotalCount
	objectGraph.SearchAfter = startObjects.SearchAfter
	objectGraph.CuurentPathNumber = len(objectGraph.RelationPaths)

	span.SetStatus(codes.Ok, "")
	return *objectGraph, nil
}

func (kns *knowledgeNetworkService) SearchSubgraphByTypePath(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnTypePath) (interfaces.PathsEntries, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询路径的对象子图")
	defer span.End()

	// 多个路径，并发查，各查各的，各自有过滤条件
	errCh := make(chan error, len(query.Paths.TypePaths))

	typePathsObjectCtx := &typePathsObjectsContext{
		ctx:               ctx,
		objectSubGraphMap: map[int]interfaces.ObjectSubGraph{},
		errCh:             errCh,
		wg:                &sync.WaitGroup{},
	}

	for i, path := range query.Paths.TypePaths {
		typePathsObjectCtx.wg.Add(1)
		go kns.buildObjectSubgraphByTypePaths(typePathsObjectCtx, query, path, i)
		// kns.buildObjectSubgraphByTypePaths(typePathsObjectCtx, typePathsObjectCtx.wg, query, path, i)
	}

	// 等待所有goroutine完成
	typePathsObjectCtx.wg.Wait()
	if len(typePathsObjectCtx.errCh) > 0 {
		err := <-typePathsObjectCtx.errCh
		if err != nil {
			return interfaces.PathsEntries{}, err
		}
	}

	// 组装结果
	graphs := make([]interfaces.ObjectSubGraph, len(query.Paths.TypePaths))
	for i := range query.Paths.TypePaths {
		if grahp, exist := typePathsObjectCtx.objectSubGraphMap[i]; exist {
			graphs[i] = grahp
		} else {
			graphs[i] = interfaces.ObjectSubGraph{}
		}

	}

	return interfaces.PathsEntries{Entries: graphs}, nil
}

func (kns *knowledgeNetworkService) buildObjectSubgraphByTypePaths(
	typePathsObjectCtx *typePathsObjectsContext,
	query *interfaces.SubGraphQueryBaseOnTypePath,
	path interfaces.QueryRelationTypePath, pathIndex int) {

	defer typePathsObjectCtx.wg.Done()

	ctx, span := ar_trace.Tracer.Start(typePathsObjectCtx.ctx, "查询路径的对象子图")
	defer span.End()

	// 1. 查各个边的关系类信息，补充 typeEdge 里的 RelationType
	typePath := interfaces.RelationTypePath{
		ObjectTypes: path.ObjectTypes,
	}
	for j, edge := range path.Edges {
		// 获取关系类信息
		relationType, exists, err := kns.omAccess.GetRelationType(ctx, query.KNID, query.Branch, edge.RelationTypeId)
		if err != nil {
			logger.Errorf("Get relation type error: %s", err.Error())

			// 添加异常时的 trace 属性
			span.SetAttributes(attribute.Key("rt_id").String(edge.RelationTypeId))
			span.SetStatus(codes.Error, "Get relation type error")
			span.End()
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Get relation type error: %v", err))

			err = rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyQuery_KnowledgeNetwork_InternalError_GetRelationTypeFailed).WithErrorDetails(err.Error())

			typePathsObjectCtx.errCh <- err
			return
		}
		if !exists {
			logger.Debugf("relation type %d not found!", edge.RelationTypeId)

			// 添加异常时的 trace 属性
			span.SetAttributes(attribute.Key("rt_id").String(edge.RelationTypeId))
			span.SetStatus(codes.Error, "relation type not found!")
			span.End()
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("relation type [%s] not found!", edge.RelationTypeId))

			err = rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_KnowledgeNetwork_RelationTypeNotFound)

			typePathsObjectCtx.errCh <- err
			return
		}
		path.Edges[j].RelationType = relationType
		// 记录方向。路径的边的方向和对应关系类的方向一致，则认为是正向，否则反向
		if path.Edges[j].SourceObjectTypeId == path.Edges[j].RelationType.SourceObjectTypeID {
			path.Edges[j].Direction = interfaces.DIRECTION_FORWARD
		} else {
			path.Edges[j].Direction = interfaces.DIRECTION_BACKWARD
		}
	}
	typePath.TypeEdges = path.Edges

	// 2. 检索起点对象类的对象实例
	startObjectQuery := &interfaces.ObjectQueryBaseOnObjectType{
		ActualCondition: path.ObjectTypes[0].ActualCondition,
		PageQuery:       path.ObjectTypes[0].PageQuery,
		KNID:            query.KNID,
		Branch:          query.Branch,
		ObjectTypeID:    path.Edges[0].SourceObjectTypeId,
		CommonQueryParameters: interfaces.CommonQueryParameters{
			IncludeTypeInfo:    true,
			IncludeLogicParams: query.IncludeLogicParams,
			IgnoringStore:      query.IgnoringStore,
			// ExcludeSystemProperties: query.CommonQueryParameters.ExcludeSystemProperties,
		},
	}
	if startObjectQuery.Limit == 0 {
		startObjectQuery.Limit = interfaces.DEFAULT_LIMIT
	}
	startObjects, err := kns.ots.GetObjectsByObjectTypeID(ctx, startObjectQuery)
	if err != nil {
		typePathsObjectCtx.errCh <- err
		return
	}

	// 3. 构建查询
	subGraphquery := &interfaces.SubGraphQueryBaseOnSource{
		KNID:              query.KNID,
		Branch:            query.Branch,
		SourceObjecTypeId: path.Edges[0].SourceObjectTypeId,
		ActualCondition:   path.ObjectTypes[0].ActualCondition,
		PageQuery: interfaces.PageQuery{
			Limit: path.Limit,
		},
		CommonQueryParameters: query.CommonQueryParameters,
		PathQuotaManager: &interfaces.PathQuotaManager{
			TotalLimit:         int64(path.Limit), // 对象路径总长度
			GlobalCount:        0,                 // 对象路径数量0
			UsedQuota:          sync.Map{},
			RequestPathTypeNum: 1,
		}, // 共享配额管理器，需要加锁保护
	}
	// 初始化状态
	baseState := &interfaces.BatchQueryState{
		Visited:   map[string]bool{}, // 用于防止循环路径
		BatchSize: 50,                // 每批查询的节点数量
	}
	subGraphquery.BatchQueryState = *baseState

	// 从起点开始沿着路径构建子图。todo: 需修正。各个对象类的过滤条件需加上
	typePathObjectCtx := &typePathObjectsContext{
		ctx:           typePathsObjectCtx.ctx,
		relationPaths: []interfaces.RelationPath{},
		objectsMap:    map[string]interfaces.ObjectInfoInSubgraph{},
		errCh:         typePathsObjectCtx.errCh,
		wg:            typePathsObjectCtx.wg,
		mu:            sync.Mutex{},
	}

	kns.buildSingleTypePathObjects(typePathObjectCtx, typePath, subGraphquery, startObjects)

	// 组装当前点的ctx
	typePathsObjectCtx.objectSubGraphMap[pathIndex] = interfaces.ObjectSubGraph{
		RelationPaths:     typePathObjectCtx.relationPaths,
		Objects:           typePathObjectCtx.objectsMap,
		TotalCount:        startObjects.TotalCount,
		SearchAfter:       startObjects.SearchAfter,
		CuurentPathNumber: len(typePathObjectCtx.relationPaths),
	}
}

// 多条路径下的数据查询
type typePathsObjectsContext struct {
	ctx               context.Context
	objectSubGraphMap map[int]interfaces.ObjectSubGraph // key是typePath的下标
	// relationPathsMap  map[int][]interfaces.RelationPath
	// objectsMap        map[int]map[string]interfaces.ObjectInfoInSubgraph
	errCh chan error
	wg    *sync.WaitGroup
}

type typePathObjectsContext struct {
	ctx           context.Context
	relationPaths []interfaces.RelationPath
	objectsMap    map[string]interfaces.ObjectInfoInSubgraph
	errCh         chan error
	wg            *sync.WaitGroup
	mu            sync.Mutex
}

// 从起点对象开始构建所有路径的对象子图
func (kns *knowledgeNetworkService) buildObjectSubgraph(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnSource,
	typePaths []interfaces.RelationTypePath,
	startObjects interfaces.Objects,
) (*interfaces.ObjectSubGraph, error) {

	logger.Infof("开始构建对象子图 - 概念路径数: %d, 起点对象数: %d, 总限制: %d",
		len(typePaths), len(startObjects.Datas), query.PathQuotaManager.TotalLimit)

	errCh := make(chan error, len(typePaths))
	typePathObjectCtx := &typePathObjectsContext{
		ctx:           ctx,
		relationPaths: []interfaces.RelationPath{},
		objectsMap:    map[string]interfaces.ObjectInfoInSubgraph{},
		errCh:         errCh,
		wg:            &sync.WaitGroup{},
		mu:            sync.Mutex{},
	}

	// 初始化状态
	baseState := &interfaces.BatchQueryState{
		Visited:   map[string]bool{}, // 用于防止循环路径
		BatchSize: 50,                // 每批查询的节点数量
	}
	query.BatchQueryState = *baseState

	// 为每个概念路径生成对象路径。 可优化，各个概念路径并行运行
	for i := range typePaths {
		typePathObjectCtx.wg.Add(1)
		go kns.buildObjectSubgraphBySource(typePathObjectCtx, typePaths[i], query, startObjects)
		// kns.buildSingleTypePathObjects(typePathObjectCtx, typePaths[i], query, startObjects)
	}

	// 等待所有goroutine完成
	typePathObjectCtx.wg.Wait()
	if len(typePathObjectCtx.errCh) > 0 {
		err := <-typePathObjectCtx.errCh
		if err != nil {
			return nil, err
		}
	}

	return &interfaces.ObjectSubGraph{
		Objects:       typePathObjectCtx.objectsMap,
		RelationPaths: typePathObjectCtx.relationPaths,
	}, nil
}

func (kns *knowledgeNetworkService) buildObjectSubgraphBySource(
	typePathObjectCtx *typePathObjectsContext,
	typePath interfaces.RelationTypePath,
	query *interfaces.SubGraphQueryBaseOnSource,
	startObjects interfaces.Objects,
) {

	defer typePathObjectCtx.wg.Done()
	kns.buildSingleTypePathObjects(typePathObjectCtx, typePath, query, startObjects)
}

func (kns *knowledgeNetworkService) buildSingleTypePathObjects(
	typePathObjectCtx *typePathObjectsContext,
	typePath interfaces.RelationTypePath,
	query *interfaces.SubGraphQueryBaseOnSource,
	startObjects interfaces.Objects,
) {

	logger.Debugf("处理概念路径 - ID: %d, 边数: %d", typePath.ID, len(typePath.TypeEdges))

	// 在开始处理前检查全局限制
	if !logics.CanGenerate(query.PathQuotaManager, typePath.ID) {
		logger.Debugf("路径ID %d 已达到限制，跳过处理", typePath.ID)
		return
	}

	// 为每个goroutine创建独立的状态副本，避免并发冲突
	localState := &interfaces.BatchQueryState{
		Visited:   make(map[string]bool),
		BatchSize: query.BatchSize,
	}

	localQuery := &interfaces.SubGraphQueryBaseOnSource{
		KNID:                  query.KNID,
		Branch:                query.Branch,
		SourceObjecTypeId:     query.SourceObjecTypeId,
		Direction:             query.Direction,
		PathLength:            query.PathLength,
		IncludeIncompletePath: query.IncludeIncompletePath,
		Condition:             query.Condition,
		ActualCondition:       query.ActualCondition,
		PageQuery:             query.PageQuery,
		PathQuotaManager:      query.PathQuotaManager, // 共享配额管理器，需要加锁保护
		BatchQueryState:       *localState,
		CommonQueryParameters: query.CommonQueryParameters,
	}

	var (
		// localPaths      []interfaces.RelationPath
		localObjectsMap = make(map[string]interfaces.ObjectInfoInSubgraph)
	)

	// 批量扩展对象路径
	currentObjectPaths, err := kns.expandObjectPathsBatch(typePathObjectCtx.ctx, localQuery, typePath,
		startObjects, localObjectsMap)
	if err != nil {
		typePathObjectCtx.errCh <- err
		return
	}

	// 合并结果到主数据结构（需要加锁）
	// 合并结果前再次检查，避免在扩展过程中其他goroutine已经达到限制
	if len(currentObjectPaths) > 0 {
		typePathObjectCtx.mu.Lock()
		defer typePathObjectCtx.mu.Unlock()

		// 检查合并后是否会超过限制. 按需合并
		currentGlobal := atomic.LoadInt64(&query.PathQuotaManager.GlobalCount)
		if currentGlobal > query.PathQuotaManager.TotalLimit {
			// 合并一批超，那么就合并差的那部分进去，直到超
			fixedNum := query.PathQuotaManager.TotalLimit - int64(len(typePathObjectCtx.relationPaths))
			// 限制 fixedNum 不超过当前批次的实际数量，避免数组越界
			if fixedNum > int64(len(currentObjectPaths)) {
				fixedNum = int64(len(currentObjectPaths))
			}
			for i := int64(0); i < fixedNum; i++ {
				typePathObjectCtx.relationPaths = append(typePathObjectCtx.relationPaths, currentObjectPaths[i])
				for _, edge := range currentObjectPaths[i].Relations {
					typePathObjectCtx.objectsMap[edge.SourceObjectId] = localObjectsMap[edge.SourceObjectId]
					typePathObjectCtx.objectsMap[edge.TargetObjectId] = localObjectsMap[edge.TargetObjectId]
				}
			}
			logger.Debugf("添加当前批次达到全局限制，只合并[%d]路径", fixedNum)
			return
		}

		typePathObjectCtx.relationPaths = append(typePathObjectCtx.relationPaths, currentObjectPaths...)
		for k, v := range localObjectsMap {
			typePathObjectCtx.objectsMap[k] = v
		}
	}
}

// 批量扩展对象路径
func (kns *knowledgeNetworkService) expandObjectPathsBatch(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnSource,
	typePath interfaces.RelationTypePath,
	startObjects interfaces.Objects,
	objectsMap map[string]interfaces.ObjectInfoInSubgraph) ([]interfaces.RelationPath, error) {

	var paths []interfaces.RelationPath

	// 使用广度优先搜索进行批量扩展
	var bfs func(currentLevelObjects []interfaces.LevelObjectWithPath, depth int) error

	bfs = func(currentLevel []interfaces.LevelObjectWithPath, depth int) error {
		// 在每一层开始前检查全局限制
		if !logics.CanGenerate(query.PathQuotaManager, typePath.ID) {
			logger.Debugf("达到限制，停止扩展路径，深度: %d", depth)
			return nil
		}

		if depth >= len(typePath.TypeEdges) || len(currentLevel) == 0 {
			// 到达路径终点，保存路径
			totalPathsInThisBatch := 0
			for _, current := range currentLevel {
				paths = append(paths, current.Paths...)
				totalPathsInThisBatch += len(current.Paths)
			}

			if totalPathsInThisBatch > 0 {
				logics.RecordGenerated(query.PathQuotaManager, typePath.ID, totalPathsInThisBatch)
				logger.Debugf("路径扩展完成 - 路径ID: %d, 新增路径: %d, 深度: %d",
					typePath.ID, totalPathsInThisBatch, depth)
			}
			return nil
		}

		// 获取当前深度的边
		edge := typePath.TypeEdges[depth]
		// 获取当前边的终点对象类
		objectType := typePath.ObjectTypes[depth+1]

		// 准备批量查询
		currentLevelObjects := make([]interfaces.LevelObject, len(currentLevel))
		for i, obj := range currentLevel {
			currentLevelObjects[i] = obj.LevelObject
		}

		// 按对象分批处理，避免单次查询过大
		batchSize := query.BatchSize
		if batchSize <= 0 {
			batchSize = 50
		}

		continueBatch := true
		for i := 0; i < len(currentLevelObjects) && continueBatch; i += batchSize {
			// 在每次处理对象前检查限制
			if !logics.CanGenerate(query.PathQuotaManager, typePath.ID) {
				// 结束，不再遍历下一批
				break
			}

			end := i + batchSize
			if end > len(currentLevelObjects) {
				end = len(currentLevelObjects)
			}

			batch := currentLevelObjects[i:end]
			// 批量查询下一层对象
			nextLevelObjects, err := kns.getNextObjectsBatchByRelation(ctx, query, batch, &edge, objectType)
			if err != nil {
				return err
			}
			if len(nextLevelObjects) == 0 {
				// 无下一层的对象
				if query.IncludeIncompletePath {
					// 如果需要包含不完整路径，将当前批次中所有对象的路径添加到结果中
					batchObjectIDs := make(map[string]bool)
					for _, obj := range batch {
						batchObjectIDs[obj.ObjectID] = true
					}
					totalPathsInThisBatch := 0
					for _, currentObj := range currentLevel {
						if batchObjectIDs[currentObj.ObjectID] && len(currentObj.Paths) > 0 {
							for _, path := range currentObj.Paths {
								if len(path.Relations) > 0 {
									paths = append(paths, path)
									totalPathsInThisBatch += len(path.Relations)
								}
							}
						}
					}
					if totalPathsInThisBatch > 0 {
						logics.RecordGenerated(query.PathQuotaManager, typePath.ID, totalPathsInThisBatch)
						logger.Debugf("添加不完整路径 - 路径ID: %d, 新增路径: %d, 深度: %d",
							typePath.ID, totalPathsInThisBatch, depth)
					}
				}
				// 继续下一批的遍历
				continue
			}

			// 构建下一层对象
			var nextLevel []interfaces.LevelObjectWithPath

			for _, currentObj := range currentLevel {
				// 在每次处理对象前检查限制
				if !logics.CanGenerate(query.PathQuotaManager, typePath.ID) {
					// 结束，不再遍历下一批
					continueBatch = false
					break
				}

				nextObjects, exists := nextLevelObjects[currentObj.ObjectID]
				if !exists {
					// 当前对象没有找到下一层对象
					if query.IncludeIncompletePath {
						// 如果需要包含不完整路径，将当前对象的路径添加到结果中
						paths = append(paths, currentObj.Paths...)
						if len(currentObj.Paths) > 0 {
							logics.RecordGenerated(query.PathQuotaManager, typePath.ID, len(currentObj.Paths))
							logger.Debugf("添加不完整路径 - 路径ID: %d, 对象ID: %s, 新增路径: %d, 深度: %d",
								typePath.ID, currentObj.ObjectID, len(currentObj.Paths), depth)
						}
					}
					continue
				}

				for _, nextObject := range nextObjects.Datas {
					// 检查限制
					if !logics.CanGenerate(query.PathQuotaManager, typePath.ID) {
						break
					}

					nextObjectID, uk := logics.GetObjectID(nextObject, nextObjects.ObjectType)
					if nextObjectID == "" {
						continue
					}

					// 构建路径键来检测循环
					// pathKey := ""
					// if len(typePath.TypeEdges) == 0 {
					// 	pathKey = fmt.Sprintf("%s:%s->%s", edge.RelationTypeId, currentObj.ObjectID, nextObjectID)
					// } else {
					// 	pathKey = logics.BuildPathKey(currentPath, nextObjectID)
					// }

					// 如果对象mapp中没有，则添加
					// _, exists = objectsMap[currentObj.ObjectID]
					// if !exists {
					// 	continue
					// }

					// 当前对象未被添加过，则添加到对象映射
					_, exists = objectsMap[currentObj.ObjectID]
					if !exists {
						objInfo := interfaces.ObjectInfoInSubgraph{
							ObjectTypeId:   currentObj.ObjectType.OTID,
							ObjectTypeName: currentObj.ObjectType.OTName,
							Properties:     currentObj.ObjectData,
						}
						if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_ID, query.ExcludeSystemProperties) {
							objInfo.InstanceID = currentObj.ObjectID
						}
						if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY, query.ExcludeSystemProperties) {
							objInfo.InstanceIdentity = currentObj.ObjectUK
						}
						if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_DISPLAY, query.ExcludeSystemProperties) {
							objInfo.Display = currentObj.ObjectData[currentObj.ObjectType.DisplayKey]
						}
						objectsMap[currentObj.ObjectID] = objInfo
					}

					// 添加下一层对象到对象映射
					objInfo := interfaces.ObjectInfoInSubgraph{
						ObjectTypeId:   nextObjects.ObjectType.OTID,
						ObjectTypeName: nextObjects.ObjectType.OTName,
						Properties:     nextObject,
					}
					if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_ID, query.ExcludeSystemProperties) {
						objInfo.InstanceID = nextObjectID
					}
					if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY, query.ExcludeSystemProperties) {
						objInfo.InstanceIdentity = uk
					}
					if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_DISPLAY, query.ExcludeSystemProperties) {
						objInfo.Display = nextObject[nextObjects.ObjectType.DisplayKey]
					}
					objectsMap[nextObjectID] = objInfo

					// 为当前对象的所有路径添加新边
					newPaths, pathExisted := kns.extendPathsWithNewEdge(query, currentObj.Paths, currentObj.ObjectID, nextObjectID, edge)
					if pathExisted {
						continue
					}

					// 记录下一层对象用于继续扩展
					nextLevel = append(nextLevel, interfaces.LevelObjectWithPath{
						LevelObject: interfaces.LevelObject{
							ObjectID:   nextObjectID,
							ObjectData: nextObject,
							PathFrom:   currentObj.ObjectID,
						},
						Paths: newPaths, // 携带扩展后的路径
					})

				}
			}

			// 当前批在此层的路径已拼完，递归处理下一层
			if len(nextLevel) > 0 {
				err = bfs(nextLevel, depth+1)
				if err != nil {
					return err
				}
				continue
			}
			// 当前层的当前批无路径可扩展，则继续扩展下一批数据的路径
		}

		// 当前层没有扩展到，结束遍历，无路径。
		return nil
	}

	// 初始化第一层对象
	var initialLevel []interfaces.LevelObjectWithPath
	for _, startObjectData := range startObjects.Datas {
		startObjectID, startObjectUK := logics.GetObjectID(startObjectData, startObjects.ObjectType)
		if startObjectID == "" {
			continue
		}

		// 为每个起点对象创建初始路径（空路径）
		initialPath := interfaces.RelationPath{
			Relations: []interfaces.Relation{},
			Length:    0,
		}

		initialLevel = append(initialLevel, interfaces.LevelObjectWithPath{
			LevelObject: interfaces.LevelObject{
				ObjectID:   startObjectID,
				ObjectType: startObjects.ObjectType,
				ObjectUK:   startObjectUK,
				ObjectData: startObjectData,
				PathFrom:   "", // 起点对象没有来源
			},
			Paths: []interfaces.RelationPath{initialPath},
		})
	}

	// 开始广度优先搜索
	err := bfs(initialLevel, 0)
	if err != nil {
		return nil, err
	}

	logger.Debugf("路径扩展完成 - 路径ID: %d, 总路径数: %d", typePath.ID, len(paths))
	return paths, nil
}

// 批量根据关系获取下一层对象
func (kns *knowledgeNetworkService) getNextObjectsBatchByRelation(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnSource,
	batch []interfaces.LevelObject,
	edge *interfaces.TypeEdge,
	objectType interfaces.ObjectTypeWithKeyField) (map[string]interfaces.Objects, error) {

	// 根据关系方向确定下一个对象类
	var nextObjectTypeID string
	isForward := true
	if edge.Direction == interfaces.DIRECTION_FORWARD {
		nextObjectTypeID = edge.RelationType.TargetObjectTypeID
	} else {
		nextObjectTypeID = edge.RelationType.SourceObjectTypeID
		isForward = false
	}

	result := make(map[string]interfaces.Objects)

	// // 按对象分批处理，避免单次查询过大
	// batchSize := query.BatchSize
	// if batchSize <= 0 {
	// 	batchSize = 50
	// }

	// for i := 0; i < len(currentLevelObjects); i += batchSize {
	// 	end := i + batchSize
	// 	if end > len(currentLevelObjects) {
	// 		end = len(currentLevelObjects)
	// 	}

	// 	batch := currentLevelObjects[i:end]

	// 构建批量查询条件，还需返回间接关联时关联视图的数据
	conditions, viewDataMap, err := kns.buildBatchConditions(ctx, query, batch, edge, isForward)
	if err != nil {
		return nil, err
	}
	if len(conditions) == 0 {
		// 通过关系类的映射构建不到过滤条件，那就跳过。
		// continue
		return nil, nil
	}

	nextObjectQuery := &interfaces.ObjectQueryBaseOnObjectType{
		KNID:         query.KNID,
		Branch:       query.Branch,
		ObjectTypeID: nextObjectTypeID,
		CommonQueryParameters: interfaces.CommonQueryParameters{
			IncludeTypeInfo:    true,
			IncludeLogicParams: query.IncludeLogicParams,
			IgnoringStore:      query.IgnoringStore,
			// ExcludeSystemProperties: query.ExcludeSystemProperties, 子图查询的系统字段在子图查询中生成，不需要对象实例自己生成
		},
	}

	if len(conditions) > 1 {
		nextObjectQuery.ActualCondition = &cond.CondCfg{
			Operation: "or", // 多个对象之间是 OR 关系
			SubConds:  conditions,
		}
	} else if len(conditions) == 1 {
		nextObjectQuery.ActualCondition = conditions[0]
	}

	// 把对象类身上配置的过滤条件加上
	if objectType.ActualCondition != nil {
		nextObjectQuery.ActualCondition = &cond.CondCfg{
			Operation: "and", // 多个对象之间是 OR 关系
			SubConds:  []*cond.CondCfg{nextObjectQuery.ActualCondition, objectType.ActualCondition},
		}
	}
	// 分页和排序信息
	if len(objectType.Sort) > 0 {
		nextObjectQuery.Sort = objectType.Sort
	}
	if objectType.Limit > 0 {
		nextObjectQuery.Limit = objectType.Limit
	} else {
		nextObjectQuery.Limit = query.Limit // 适当调整限制
	}

	nextObjects, err := kns.ots.GetObjectsByObjectTypeID(ctx, nextObjectQuery)
	if err != nil {
		return nil, err
	}
	// logger.Debugf("从对象类[%s]中获取到的数据条数为[%d]", nextObjectTypeID, len(nextObjects.Datas))

	// 根据映射规则将结果映射回各个对象
	kns.mapResultsToObjects(batch, nextObjects, result, edge, isForward, viewDataMap)
	// }

	return result, nil
}

// 将查询结果映射回各个对象
func (kns *knowledgeNetworkService) mapResultsToObjects(currentLevelObjects []interfaces.LevelObject,
	nextObjects interfaces.Objects, result map[string]interfaces.Objects,
	edge *interfaces.TypeEdge, isForward bool,
	viewDataMap map[string][]map[string]any) {

	// 根据映射规则过滤属于每个对象的下一层对象
	for _, levelObj := range currentLevelObjects {
		filteredObjects := interfaces.Objects{
			Datas:       []map[string]any{},
			ObjectType:  nextObjects.ObjectType,
			TotalCount:  0,
			SearchAfter: nextObjects.SearchAfter,
		}

		for _, nextObj := range nextObjects.Datas {
			// 获取该对象的视图数据（如果是间接映射）
			var objectViewData []map[string]any
			if _, isIndirect := edge.RelationType.MappingRules.(interfaces.InDirectMapping); isIndirect {
				objectViewData = viewDataMap[levelObj.ObjectID]
			}

			if kns.isObjectRelated(levelObj.ObjectData, nextObj, edge, isForward, objectViewData) {
				filteredObjects.Datas = append(filteredObjects.Datas, nextObj)
				filteredObjects.TotalCount++
			}
		}

		if len(filteredObjects.Datas) > 0 {
			result[levelObj.ObjectID] = filteredObjects
		}
	}
}

// 判断对象是否关联（根据映射规则）
func (kns *knowledgeNetworkService) isObjectRelated(currentObjectData map[string]any,
	nextObject map[string]any, edge *interfaces.TypeEdge, isForward bool,
	viewData []map[string]any) bool {

	switch mappingRules := edge.RelationType.MappingRules.(type) {
	case []interfaces.Mapping:
		// 检查直接映射条件是否满足
		return logics.CheckDirectMappingConditions(currentObjectData, nextObject, mappingRules, isForward)

	case interfaces.InDirectMapping:
		// 间接映射的检查逻辑
		// 需要根据具体业务实现
		return logics.CheckIndirectMappingConditionsWithViewData(currentObjectData, nextObject, mappingRules, isForward, viewData)
	}

	return false
}

// 构建批量查询条件
func (kns *knowledgeNetworkService) buildBatchConditions(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnSource,
	currentLevelObjects []interfaces.LevelObject,
	edge *interfaces.TypeEdge,
	isForward bool) ([]*cond.CondCfg, map[string][]map[string]any, error) {

	var conditions []*cond.CondCfg
	viewDataMap := make(map[string][]map[string]any) // objectID -> []viewData

	// 先处理直接映射的情况
	directObjects := make([]interfaces.LevelObject, 0)
	indirectObjects := make([]interfaces.LevelObject, 0)

	for _, levelObj := range currentLevelObjects {
		switch edge.RelationType.MappingRules.(type) {
		case []interfaces.Mapping:
			directObjects = append(directObjects, levelObj)
		case interfaces.InDirectMapping:
			indirectObjects = append(indirectObjects, levelObj)
		}
	}

	// 处理直接映射
	if len(directObjects) > 0 {
		directConditions, err := logics.BuildDirectBatchConditions(directObjects, edge, isForward)
		if err != nil {
			return nil, nil, err
		}
		conditions = append(conditions, directConditions...)
	}

	// 处理间接映射 - 批量查询视图数据
	if len(indirectObjects) > 0 {
		indirectConditions, indirectViewData, err := kns.buildIndirectBatchConditions(ctx, query, indirectObjects, edge, isForward)
		if err != nil {
			return nil, nil, err
		}
		conditions = append(conditions, indirectConditions...)

		// 合并视图数据映射
		for k, v := range indirectViewData {
			viewDataMap[k] = v
		}
	}

	return conditions, viewDataMap, nil
}

// 构建间接映射的批量条件，并返回视图数据映射
func (kns *knowledgeNetworkService) buildIndirectBatchConditions(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnSource,
	currentLevelObjects []interfaces.LevelObject,
	edge *interfaces.TypeEdge, isForward bool) ([]*cond.CondCfg, map[string][]map[string]any, error) {

	var conditions []*cond.CondCfg
	viewDataMap := make(map[string][]map[string]any)
	mappingRules := edge.RelationType.MappingRules.(interfaces.InDirectMapping)

	if mappingRules.BackingDataSource.ID == "" {
		// 视图为空，返回异常，不请求
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("当前关系类[%s]是通过视图关联,视图ID为空,请检查关系类的配置", edge.RelationType.RTName))
	}

	// 视图到目标对象的映射关系
	var targetMappingRules []interfaces.Mapping
	if isForward {
		targetMappingRules = mappingRules.TargetMappingRules
	} else {
		targetMappingRules = mappingRules.SourceMappingRules
	}

	// 批量查询所有对象的视图数据
	batchViewData, err := kns.batchGetViewData(ctx, query, edge, currentLevelObjects, mappingRules, isForward)
	if err != nil {
		return nil, nil, err
	}

	var inValues []any
	var inField string
	// 为每个对象构建查询条件
	for _, levelObj := range currentLevelObjects {
		objectViewData, exists := batchViewData[levelObj.ObjectID]
		if !exists || len(objectViewData) == 0 {
			continue
		}

		// 保存视图数据映射，用于后续的对象关联判断
		viewDataMap[levelObj.ObjectID] = objectViewData

		// 遍历视图数据，逐个构建过过滤条件，最后用or连接
		multiConds := []*cond.CondCfg{}
		for _, viewData := range objectViewData {
			viewConditions, targetField, inValue := logics.BuildCondition(nil, targetMappingRules, isForward, viewData)
			multiConds = append(multiConds, viewConditions...)
			inValues = append(inValues, inValue)
			inField = targetField
		}

		if len(multiConds) > 1 {
			conditions = append(conditions, &cond.CondCfg{
				Operation: "or",
				SubConds:  multiConds,
			})
		} else if len(multiConds) == 1 {
			conditions = append(conditions, multiConds[0])
		}
	}

	if len(targetMappingRules) == 1 && len(inValues) > 0 {
		return []*cond.CondCfg{
			{
				Name:      inField,
				Operation: "in",
				ValueOptCfg: cond.ValueOptCfg{
					ValueFrom: "const",
					Value:     inValues,
				},
			},
		}, viewDataMap, nil
	}
	return conditions, viewDataMap, nil
}

// 批量获取视图数据
func (kns *knowledgeNetworkService) batchGetViewData(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnSource,
	edge *interfaces.TypeEdge,
	currentLevelObjects []interfaces.LevelObject,
	mappingRules interfaces.InDirectMapping, isForward bool) (map[string][]map[string]any, error) {

	result := make(map[string][]map[string]any)
	batchSize := 50 // 视图查询的批次大小
	var mappingRulesToUse []interfaces.Mapping
	if isForward {
		mappingRulesToUse = mappingRules.SourceMappingRules
	} else {
		mappingRulesToUse = mappingRules.TargetMappingRules
	}

	// 按批次处理对象
	for i := 0; i < len(currentLevelObjects); i += batchSize {
		end := i + batchSize
		if end > len(currentLevelObjects) {
			end = len(currentLevelObjects)
		}

		batch := currentLevelObjects[i:end]

		// 为批次内的所有对象构建组合查询条件
		batchConditions := []*cond.CondCfg{}
		objectMapping := make(map[int]string) // 条件索引到对象ID的映射
		var inValues []any
		var inField string
		for _, levelObj := range batch {
			objectConditions, targetField, inValue := logics.BuildCondition(nil, mappingRulesToUse, isForward, levelObj.ObjectData)
			inValues = append(inValues, inValue)
			inField = targetField

			if len(objectConditions) > 1 {
				batchConditions = append(batchConditions, &cond.CondCfg{
					Operation: "and",
					SubConds:  objectConditions,
				})
			} else if len(objectConditions) == 1 {
				batchConditions = append(batchConditions, objectConditions[0])
			} else {
				continue
			}
			objectMapping[len(batchConditions)-1] = levelObj.ObjectID
		}

		// if len(batchConditions) == 0 {
		// 	continue
		// }

		// 构建视图查询
		viewQuery := &interfaces.ViewQuery{
			NeedTotal: query.NeedTotal,
			Limit:     interfaces.MAX_LIMIT, // 查关系表时，不限制条数，所有关系都查出来
		}

		if len(mappingRulesToUse) == 1 && len(inValues) > 0 {
			viewQuery.Filters = &cond.CondCfg{
				Name:      inField,
				Operation: "in",
				ValueOptCfg: cond.ValueOptCfg{
					ValueFrom: "const",
					Value:     inValues,
				},
			}
		} else {
			if len(batchConditions) > 1 {
				viewQuery.Filters = &cond.CondCfg{
					Operation: "or",
					SubConds:  batchConditions,
				}
			} else {
				viewQuery.Filters = batchConditions[0]
			}
		}

		// 构建排序，按关联字段排序
		sort := []*interfaces.SortParams{}
		for _, mapping := range mappingRulesToUse {
			targetName := mapping.TargetProp.Name
			if !isForward {
				targetName = mapping.SourceProp.Name
			}

			sort = append(sort, &interfaces.SortParams{
				Field:     targetName,
				Direction: interfaces.ASC_DIRECTION,
			})
		}
		viewQuery.Sort = sort

		// 执行视图查询
		backingViewData, err := kns.uAccess.GetViewDataByID(ctx, mappingRules.BackingDataSource.ID, *viewQuery)
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyQuery_ObjectType_InternalError_GetViewDataByIDFailed).WithErrorDetails(err.Error())
		}

		logger.Debugf("依据关系[%s]从视图[%s]中获取到的数据条数为[%d]", edge.RelationType.RTName, mappingRules.BackingDataSource.ID, len(backingViewData.Datas))

		// 将视图数据映射回各个对象
		kns.mapViewDataToObjects(backingViewData.Datas, batchConditions, objectMapping, mappingRules, isForward, result)
	}

	return result, nil
}

// 将视图数据映射回各个对象
func (kns *knowledgeNetworkService) mapViewDataToObjects(viewData []map[string]any,
	batchConditions []*cond.CondCfg,
	objectMapping map[int]string,
	mappingRules interfaces.InDirectMapping,
	isForward bool,
	result map[string][]map[string]any) {

	for _, data := range viewData {
		// 找出这个视图数据属于哪个对象
		for condIndex, objectID := range objectMapping {
			if condIndex >= len(batchConditions) {
				continue
			}

			var mappingRulesToUse []interfaces.Mapping
			if isForward {
				mappingRulesToUse = mappingRules.SourceMappingRules
			} else {
				mappingRulesToUse = mappingRules.TargetMappingRules
			}

			// 检查视图数据是否满足该对象的查询条件
			if logics.CheckViewDataMatchesCondition(data, batchConditions[condIndex], mappingRulesToUse, isForward) {
				if result[objectID] == nil {
					result[objectID] = make([]map[string]any, 0)
				}
				result[objectID] = append(result[objectID], data)
				break // 一个视图记录只属于一个对象
			}
		}
	}
}

// 为路径集合添加新边以及检查是否存在重复路径
func (kns *knowledgeNetworkService) extendPathsWithNewEdge(query *interfaces.SubGraphQueryBaseOnSource,
	paths []interfaces.RelationPath,
	sourceObjectID string, targetObjectID string, edge interfaces.TypeEdge) ([]interfaces.RelationPath, bool) {

	var newPaths []interfaces.RelationPath
	var pathExisted bool

	for _, path := range paths {
		// 检查这个路径是否以 sourceObjectID 结尾
		if !kns.isPathEndsWith(path, sourceObjectID) {
			continue
		}

		// 创建新路径（深拷贝）
		newPath := interfaces.RelationPath{
			Relations: make([]interfaces.Relation, len(path.Relations)),
			Length:    path.Length + 1,
		}
		copy(newPath.Relations, path.Relations)

		// 添加新边
		newPath.Relations = append(newPath.Relations, interfaces.Relation{
			RelationTypeId:   edge.RelationTypeId,
			RelationTypeName: edge.RelationType.RTName,
			SourceObjectId:   sourceObjectID,
			TargetObjectId:   targetObjectID,
		})

		// 构建路径键来检测循环
		pathKey := ""
		for _, edge := range newPath.Relations {
			pathKey = fmt.Sprintf("%s-%s:%s->%s", pathKey, edge.RelationTypeId, edge.SourceObjectId, edge.TargetObjectId)
		}
		if query.BatchQueryState.Visited[pathKey] {
			logger.Warnf("检测到重复路径: %s", pathKey)
			pathExisted = true
		}
		query.BatchQueryState.Visited[pathKey] = true

		newPaths = append(newPaths, newPath)
	}

	return newPaths, pathExisted
}

// 检查路径是否以指定对象ID结尾
func (kns *knowledgeNetworkService) isPathEndsWith(path interfaces.RelationPath, objectID string) bool {
	if len(path.Relations) == 0 {
		// 空路径，检查是否是起点对象
		// 这里需要额外的逻辑来跟踪起点对象，暂时返回true
		return true
	}

	// 检查最后一条边的目标对象是否匹配
	lastEdge := path.Relations[len(path.Relations)-1]
	return lastEdge.TargetObjectId == objectID
}

// 基于一组对象实例组织关系子图
func (kns *knowledgeNetworkService) SearchSubgraphByObjects(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnObjects) (interfaces.ObjectSubGraph, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "基于一组对象实例组织关系子图")
	defer span.End()

	var result interfaces.ObjectSubGraph
	result.Objects = make(map[string]interfaces.ObjectInfoInSubgraph)
	result.IsolatedObjects = make(map[string]interfaces.ObjectInfoInSubgraph)
	result.RelationPaths = []interfaces.RelationPath{}

	// 1. 处理输入对象实例，验证并查询对象数据
	objectsByType, objectTypeMap, err := kns.processInputObjects(ctx, query)
	if err != nil {
		return result, err
	}

	if len(objectsByType) == 0 {
		return result, nil
	}

	// 2. 发现相关关系类
	allRelationTypes, err := kns.discoverRelationTypes(ctx, query, objectTypeMap)
	if err != nil {
		return result, err
	}

	// 3. 匹配关系
	relations, objectsInRelations, err := kns.matchRelations(ctx, query, objectsByType, allRelationTypes)
	if err != nil {
		return result, err
	}

	// 4. 构建子图
	result = kns.buildSubgraphFromObjects(query, objectsByType, relations, objectsInRelations)

	return result, nil
}

// 处理输入对象实例，验证并查询对象数据
func (kns *knowledgeNetworkService) processInputObjects(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnObjects) (map[string][]interfaces.LevelObject, map[string]*interfaces.ObjectType, error) {

	objectsByType := make(map[string][]interfaces.LevelObject)
	objectTypeMap := make(map[string]*interfaces.ObjectType)

	// 按对象类型分组输入对象
	typeGroups := make(map[string][]interfaces.InputObjectInstance)
	for _, entry := range query.Entries {
		typeGroups[entry.ObjectTypeID] = append(typeGroups[entry.ObjectTypeID], entry)
	}

	// 对每个对象类型，批量查询对象数据
	for otID, entries := range typeGroups {
		// 获取对象类型信息
		objectType, exists, err := kns.omAccess.GetObjectType(ctx, query.KNID, query.Branch, otID)
		if err != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed).WithErrorDetails(err.Error())
		}
		if !exists {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusNotFound,
				oerrors.OntologyQuery_ObjectType_ObjectTypeNotFound).WithErrorDetails(fmt.Sprintf("对象类型[%s]不存在", otID))
		}
		objectTypeMap[otID] = &objectType

		// 构建唯一标识条件
		instanceIdentities := make([]map[string]any, len(entries))
		for i, entry := range entries {
			instanceIdentities[i] = entry.InstanceIdentity
		}
		ukCond := logics.BuildInstanceIdentitiesCondition(instanceIdentities)

		// 查询对象数据
		objectQuery := &interfaces.ObjectQueryBaseOnObjectType{
			ActualCondition: ukCond,
			PageQuery: interfaces.PageQuery{
				Limit:     interfaces.MAX_LIMIT,
				NeedTotal: false,
			},
			KNID:         query.KNID,
			Branch:       query.Branch,
			ObjectTypeID: otID,
			CommonQueryParameters: interfaces.CommonQueryParameters{
				IncludeTypeInfo:    true,
				IncludeLogicParams: query.IncludeLogicParams,
				IgnoringStore:      query.IgnoringStore,
			},
		}

		objects, err := kns.ots.GetObjectsByObjectTypeID(ctx, objectQuery)
		if err != nil {
			return nil, nil, err
		}

		// 构建LevelObject列表
		levelObjects := make([]interfaces.LevelObject, 0, len(objects.Datas))
		for _, objData := range objects.Datas {
			objectID, uk := logics.GetObjectID(objData, objects.ObjectType)
			if objectID == "" {
				continue
			}
			levelObjects = append(levelObjects, interfaces.LevelObject{
				ObjectID:   objectID,
				ObjectUK:   uk,
				ObjectData: objData,
				ObjectType: objects.ObjectType,
			})
		}
		objectsByType[otID] = levelObjects
	}

	return objectsByType, objectTypeMap, nil
}

// 发现与输入对象类型相关的关系类
func (kns *knowledgeNetworkService) discoverRelationTypes(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnObjects, objectTypeMap map[string]*interfaces.ObjectType) (map[string]interfaces.RelationType, error) {

	// 提取所有对象类型ID集合
	objectTypeIDs := make([]string, 0, len(objectTypeMap))
	for otID := range objectTypeMap {
		objectTypeIDs = append(objectTypeIDs, otID)
	}

	// 一次性查询所有相关关系类
	relationTypes, err := kns.omAccess.ListRelationTypes(ctx, query.KNID, query.Branch, interfaces.RelationTypesQuery{
		SourceObjectTypeIDs: objectTypeIDs,
		TargetObjectTypeIDs: objectTypeIDs,
	})
	if err != nil {
		return nil, err
	}

	// 收集并过滤关系类（确保源和目标都在输入集合中）
	allRelationTypes := make(map[string]interfaces.RelationType) // key: relationTypeID
	for _, rt := range relationTypes {
		// 确保源和目标都在输入集合中
		if _, exists := objectTypeMap[rt.SourceObjectTypeID]; !exists {
			continue
		}
		if _, exists := objectTypeMap[rt.TargetObjectTypeID]; !exists {
			continue
		}
		allRelationTypes[rt.RTID] = rt
	}

	return allRelationTypes, nil
}

// 匹配对象实例之间的关系
func (kns *knowledgeNetworkService) matchRelations(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnObjects,
	objectsByType map[string][]interfaces.LevelObject,
	allRelationTypes map[string]interfaces.RelationType) ([]interfaces.Relation, map[string]bool, error) {

	relations := []interfaces.Relation{}
	objectsInRelations := make(map[string]bool) // 记录参与关系的对象ID

	// 直接遍历所有关系类
	for _, relationType := range allRelationTypes {

		logger.Debugf("匹配关系类: %s, 源对象类型: %s, 目标对象类型: %s", relationType.RTID, relationType.SourceObjectTypeID, relationType.TargetObjectTypeID)

		sourceOTID := relationType.SourceObjectTypeID
		targetOTID := relationType.TargetObjectTypeID

		sourceObjects := objectsByType[sourceOTID]
		targetObjects := objectsByType[targetOTID]

		if len(sourceObjects) == 0 || len(targetObjects) == 0 {
			continue
		}

		// 构建TypeEdge用于复用现有逻辑
		edge := &interfaces.TypeEdge{
			RelationTypeId:     relationType.RTID,
			RelationType:       relationType,
			SourceObjectTypeId: sourceOTID,
			TargetObjectTypeId: targetOTID,
			Direction:          interfaces.DIRECTION_FORWARD,
		}

		// 用关系类的起点对象集和终点对象集去匹配关系，返回匹配到的关系
		matchedRelations, err := kns.matchRelationsForPair(ctx, query, sourceObjects, targetObjects, edge)
		if err != nil {
			logger.Warnf("匹配关系失败: relationType=%s, error=%v", relationType.RTID, err)
			continue
		}

		// 添加到结果中
		for _, rel := range matchedRelations {
			relations = append(relations, rel)
			objectsInRelations[rel.SourceObjectId] = true
			objectsInRelations[rel.TargetObjectId] = true
		}
	}

	return relations, objectsInRelations, nil
}

// 匹配一对对象类型之间的关系
func (kns *knowledgeNetworkService) matchRelationsForPair(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnObjects,
	sourceObjects []interfaces.LevelObject,
	targetObjects []interfaces.LevelObject,
	edge *interfaces.TypeEdge) ([]interfaces.Relation, error) {

	// 判断关系类型，分别处理直接关联和间接关联
	switch edge.RelationType.MappingRules.(type) {
	case []interfaces.Mapping:
		// 直接关联：直接使用已有的 targetObjects 进行匹配，避免数据库查询
		return kns.matchDirectRelations(sourceObjects, targetObjects, edge)

	case interfaces.InDirectMapping:
		// 间接关联：需要查询视图数据，但只匹配输入的目标对象
		return kns.matchIndirectRelations(ctx, query, sourceObjects, targetObjects, edge)
	}

	return []interfaces.Relation{}, nil
}

// 匹配直接关联关系（无需查询数据库）
func (kns *knowledgeNetworkService) matchDirectRelations(
	sourceObjects []interfaces.LevelObject,
	targetObjects []interfaces.LevelObject,
	edge *interfaces.TypeEdge) ([]interfaces.Relation, error) {

	relations := []interfaces.Relation{}
	mappingRules := edge.RelationType.MappingRules.([]interfaces.Mapping)

	// 直接遍历源对象和目标对象进行匹配
	for _, sourceObj := range sourceObjects {
		for _, targetObj := range targetObjects {
			if logics.CheckDirectMappingConditions(sourceObj.ObjectData, targetObj.ObjectData, mappingRules, true) {
				relations = append(relations, interfaces.Relation{
					RelationTypeId:   edge.RelationTypeId,
					RelationTypeName: edge.RelationType.RTName,
					SourceObjectId:   sourceObj.ObjectID,
					TargetObjectId:   targetObj.ObjectID,
				})
			}
		}
	}

	return relations, nil
}

// 匹配间接关联关系（需要查询视图数据，但只匹配输入的目标对象）
func (kns *knowledgeNetworkService) matchIndirectRelations(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnObjects,
	sourceObjects []interfaces.LevelObject,
	targetObjects []interfaces.LevelObject,
	edge *interfaces.TypeEdge) ([]interfaces.Relation, error) {

	relations := []interfaces.Relation{}

	// 查询视图数据（这是必要的，因为视图数据不在输入对象中）
	_, viewDataMap, err := kns.buildBatchConditionsForObjects(ctx, query, sourceObjects, edge, true)
	if err != nil {
		return nil, err
	}

	// 直接使用输入的目标对象进行匹配（而不是查询数据库）
	for _, sourceObj := range sourceObjects {
		objectViewData := viewDataMap[sourceObj.ObjectID]

		for _, targetObj := range targetObjects {
			mappingRules := edge.RelationType.MappingRules.(interfaces.InDirectMapping)
			if logics.CheckIndirectMappingConditionsWithViewData(
				sourceObj.ObjectData,
				targetObj.ObjectData,
				mappingRules,
				true,
				objectViewData) {
				relations = append(relations, interfaces.Relation{
					RelationTypeId:   edge.RelationTypeId,
					RelationTypeName: edge.RelationType.RTName,
					SourceObjectId:   sourceObj.ObjectID,
					TargetObjectId:   targetObj.ObjectID,
				})
			}
		}
	}

	return relations, nil
}

// 为对象构建批量查询条件（适配SubGraphQueryBaseOnObjects）
func (kns *knowledgeNetworkService) buildBatchConditionsForObjects(ctx context.Context,
	query *interfaces.SubGraphQueryBaseOnObjects,
	currentLevelObjects []interfaces.LevelObject,
	edge *interfaces.TypeEdge,
	isForward bool) ([]*cond.CondCfg, map[string][]map[string]any, error) {

	var conditions []*cond.CondCfg
	viewDataMap := make(map[string][]map[string]any)

	// 先处理直接映射的情况
	directObjects := make([]interfaces.LevelObject, 0)
	indirectObjects := make([]interfaces.LevelObject, 0)

	for _, levelObj := range currentLevelObjects {
		switch edge.RelationType.MappingRules.(type) {
		case []interfaces.Mapping:
			directObjects = append(directObjects, levelObj)
		case interfaces.InDirectMapping:
			indirectObjects = append(indirectObjects, levelObj)
		}
	}

	// 处理直接映射
	if len(directObjects) > 0 {
		directConditions, err := logics.BuildDirectBatchConditions(directObjects, edge, isForward)
		if err != nil {
			return nil, nil, err
		}
		conditions = append(conditions, directConditions...)
	}

	// 处理间接映射 - 批量查询视图数据
	if len(indirectObjects) > 0 {
		// 构建一个临时的SubGraphQueryBaseOnSource用于调用buildIndirectBatchConditions
		tempQuery := &interfaces.SubGraphQueryBaseOnSource{
			KNID:   query.KNID,
			Branch: query.Branch,
			PageQuery: interfaces.PageQuery{
				NeedTotal: false,
			},
			CommonQueryParameters: interfaces.CommonQueryParameters{
				IncludeLogicParams: query.IncludeLogicParams,
				IgnoringStore:      query.IgnoringStore,
			},
		}
		indirectConditions, indirectViewData, err := kns.buildIndirectBatchConditions(ctx, tempQuery, indirectObjects, edge, isForward)
		if err != nil {
			return nil, nil, err
		}
		conditions = append(conditions, indirectConditions...)

		// 合并视图数据映射
		for k, v := range indirectViewData {
			viewDataMap[k] = v
		}
	}

	return conditions, viewDataMap, nil
}

// 构建子图
func (kns *knowledgeNetworkService) buildSubgraphFromObjects(query *interfaces.SubGraphQueryBaseOnObjects,
	objectsByType map[string][]interfaces.LevelObject,
	relations []interfaces.Relation,
	objectsInRelations map[string]bool) interfaces.ObjectSubGraph {

	result := interfaces.ObjectSubGraph{
		Objects:         make(map[string]interfaces.ObjectInfoInSubgraph),
		IsolatedObjects: make(map[string]interfaces.ObjectInfoInSubgraph),
		RelationPaths:   []interfaces.RelationPath{},
	}

	// 构建对象映射表
	for _, levelObjects := range objectsByType {
		for _, levelObj := range levelObjects {
			objInfo := interfaces.ObjectInfoInSubgraph{
				ObjectTypeId:   levelObj.ObjectType.OTID,
				ObjectTypeName: levelObj.ObjectType.OTName,
				Properties:     levelObj.ObjectData,
			}
			if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_ID, query.ExcludeSystemProperties) {
				objInfo.InstanceID = levelObj.ObjectID
			}
			if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY, query.ExcludeSystemProperties) {
				objInfo.InstanceIdentity = levelObj.ObjectUK
			}
			if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_DISPLAY, query.ExcludeSystemProperties) {
				objInfo.Display = levelObj.ObjectData[levelObj.ObjectType.DisplayKey]
			}

			// 判断是否为孤立对象
			if objectsInRelations[levelObj.ObjectID] {
				result.Objects[levelObj.ObjectID] = objInfo
			} else {
				result.IsolatedObjects[levelObj.ObjectID] = objInfo
			}
		}
	}

	// 构建关系路径（长度为1）
	for _, rel := range relations {
		result.RelationPaths = append(result.RelationPaths, interfaces.RelationPath{
			Relations: []interfaces.Relation{rel},
			Length:    1,
		})
	}

	return result
}
