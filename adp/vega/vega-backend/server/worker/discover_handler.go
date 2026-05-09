// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/common"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/discover_task"
	"vega-backend/logics/resource"
)

// indexDiscoverItem represents an index discover item.
type indexDiscoverItem struct {
	resource  *interfaces.Resource
	indexMeta *interfaces.IndexMeta
}

// tableDiscoverItem represents a table discover item.
type tableDiscoverItem struct {
	resource  *interfaces.Resource
	tableMeta *interfaces.TableMeta
}

// filesetDiscoverItem represents a fileset discover item.
type filesetDiscoverItem struct {
	resource *interfaces.Resource
	meta     *interfaces.FilesetMeta
}

// discoverHandler handles discover tasks.
type discoverHandler struct {
	appSetting *common.AppSetting
	rs         interfaces.ResourceService
	cs         interfaces.CatalogService
	dts        interfaces.DiscoverTaskService
}

// NewDiscoverHandler creates a new discover handler.
func NewDiscoverHandler(appSetting *common.AppSetting) *discoverHandler {
	return &discoverHandler{
		appSetting: appSetting,
		rs:         resource.NewResourceService(appSetting),
		cs:         catalog.NewCatalogService(appSetting),
		dts:        discover_task.NewDiscoverTaskService(appSetting),
	}
}

// HandleTask handles a discover task from the queue.
func (dh *discoverHandler) HandleTask(ctx context.Context, task *asynq.Task) error {
	var msg interfaces.DiscoverTaskMessage
	if err := sonic.Unmarshal(task.Payload(), &msg); err != nil {
		logger.Errorf("Failed to unmarshal task message: %v", err)
		return err
	}

	taskID := msg.TaskID
	logger.Infof("Starting discover for task: %s", taskID)

	taskInfo, err := dh.dts.GetByID(ctx, taskID)
	if err != nil {
		logger.Errorf("Failed to get task info for task %s: %v", taskID, err)
		return err
	}

	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, taskInfo.Creator)

	catalogInfo, err := dh.cs.GetByID(ctx, taskInfo.CatalogID, true)
	if err != nil {
		logger.Errorf("Failed to get catalog for task %s: %v", taskID, err)
		return err
	}

	// Update task status to running and set start time
	now := time.Now().UnixMilli()
	if err := dh.dts.UpdateStatus(ctx, taskID, interfaces.DiscoverTaskStatusRunning, "", now); err != nil {
		logger.Errorf("Failed to set start time for task %s: %v", taskID, err)
		return err
	}

	// Execute discover : 元数据采集主要逻辑
	//首先根据 catalog ID 获取 catalog 信息，
	//然后根据 catalog 信息获取 connector 信息，
	//然后根据 connector 信息获取 connector 实例，
	//然后根据 connector 实例获取 catalog 的元数据，
	//然后根据 catalog 的元数据获取 catalog 的资源信息：元数据
	result, err := dh.discoverCatalog(ctx, catalogInfo, taskID)
	if err != nil {
		// Update task status to failed
		now = time.Now().UnixMilli()
		_ = dh.dts.UpdateStatus(ctx, taskID, interfaces.DiscoverTaskStatusFailed, err.Error(), now)
		return err
	}

	// Update task result
	now = time.Now().UnixMilli()
	if err := dh.dts.UpdateResult(ctx, taskID, result, now); err != nil {
		logger.Errorf("Failed to update result for task %s: %v", taskID, err)
		return err
	}

	logger.Infof("Discover completed for task: %s, catalog: %s", taskID, catalogInfo.ID)
	return nil
}

// discoverCatalog discovers resources for a specific catalog.
// discoverCatalog 是一个发现目录资源的方法
// 它接收上下文和目录信息，返回发现结果或错误
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - catalog: 目录信息，包含目录ID和类型等
//
// 返回值:
//   - *interfaces.DiscoverResult: 发现结果，包含发现的资源信息
//   - error: 错误信息，如果发现过程中出现错误
func (dh *discoverHandler) discoverCatalog(ctx context.Context, catalog *interfaces.Catalog, taskID string) (*interfaces.DiscoverResult, error) {

	logger.Infof("Starting discover for catalog: %s", catalog.ID)

	// 验证 catalog 类型
	if catalog.Type != interfaces.CatalogTypePhysical {
		return nil, fmt.Errorf("discover only supports physical catalogs")
	}

	// 1. 创建 Connector 并连接
	connector, err := dh.createAndConnectConnector(ctx, catalog)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}
	defer func() { _ = connector.Close(ctx) }()

	// Update catalog metadata
	if meta, err := connector.GetMetadata(ctx); err == nil {
		if err := dh.cs.UpdateMetadata(ctx, catalog.ID, meta); err != nil {
			logger.Errorf("Failed to update catalog metadata: %v", err)
		}
	} else {
		logger.Warnf("Failed to get metadata: %v", err)
	}
	// 2. 根据 connector category 分发到不同的发现函数：例如mysql会到mysql.go下面进行元数据的采集，里面会有具体的实现
	category := connector.GetCategory()
	switch category {
	// table类型的会到这里，例如mysql
	case interfaces.ConnectorCategoryTable:
		return dh.discoverTableResources(ctx, catalog, connector, taskID)
	// index类型的会到这里，例如open search
	case interfaces.ConnectorCategoryIndex:
		return dh.discoverIndexResources(ctx, catalog, connector)
	// fileset类型的会到这里，例如anyshare
	case interfaces.ConnectorCategoryFileset:
		return dh.discoverFilesetResources(ctx, catalog, connector)
	default:
		return nil, fmt.Errorf("unsupported connector category for discover: %s", category)
	}
}

// discoverFilesetResources discovers fileset resources from a fileset connector.
func (dh *discoverHandler) discoverFilesetResources(ctx context.Context, catalog *interfaces.Catalog,
	connector connectors.Connector) (*interfaces.DiscoverResult, error) {

	filesetConnector, ok := connector.(connectors.FilesetConnector)
	if !ok {
		return nil, fmt.Errorf("connector does not support fileset discover")
	}

	sourceFilesets, err := filesetConnector.ListFilesets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list filesets: %w", err)
	}
	logger.Infof("Discovered %d fileset objects from source", len(sourceFilesets))

	existingResources, err := dh.rs.GetByCatalogID(ctx, catalog.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing resources: %w", err)
	}

	result, items, err := dh.reconcileFilesetResources(ctx, catalog, sourceFilesets, existingResources)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile fileset resources: %w", err)
	}

	if err := dh.enrichFilesetMetadata(ctx, items); err != nil {
		return nil, fmt.Errorf("failed to enrich fileset metadata: %w", err)
	}

	logger.Infof("Discover completed for catalog %s: new=%d, stale=%d, unchanged=%d", catalog.ID, result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, nil
}

func (dh *discoverHandler) reconcileFilesetResources(ctx context.Context, catalog *interfaces.Catalog, source []*interfaces.FilesetMeta, existingResources []*interfaces.Resource) (*interfaces.DiscoverResult, []filesetDiscoverItem, error) {
	result := &interfaces.DiscoverResult{
		CatalogID: catalog.ID,
	}
	var items []filesetDiscoverItem

	existingMap := make(map[string]*interfaces.Resource)
	for _, r := range existingResources {
		if r.Category != interfaces.ResourceCategoryFileset {
			continue
		}
		existingMap[r.SourceIdentifier] = r
	}

	sourceMap := make(map[string]*interfaces.FilesetMeta)
	for _, fs := range source {
		sid := filesetSourceIdentifier(fs)
		sourceMap[sid] = fs
	}

	for _, fs := range source {
		sourceIdentifier := filesetSourceIdentifier(fs)
		if resource, ok := existingMap[sourceIdentifier]; ok {
			if resource.Status == interfaces.ResourceStatusStale {
				if err := dh.rs.UpdateStatus(ctx, resource.ID, interfaces.ResourceStatusActive, ""); err != nil {
					logger.Errorf("Failed to reactivate resource %s: %v", resource.ID, err)
				}
			}
			result.UnchangedCount++
			items = append(items, filesetDiscoverItem{resource: resource, meta: fs})
		} else {
			resource, err := dh.createFilesetResource(ctx, catalog, fs, sourceIdentifier)
			if err != nil {
				logger.Errorf("Failed to create fileset resource %s: %v", sourceIdentifier, err)
			} else {
				result.NewCount++
				items = append(items, filesetDiscoverItem{resource: resource, meta: fs})
			}
		}
	}

	for sourceIdentifier, existing := range existingMap {
		if _, ok := sourceMap[sourceIdentifier]; !ok {
			if existing.Status != interfaces.ResourceStatusStale {
				if err := dh.rs.UpdateStatus(ctx, existing.ID, interfaces.ResourceStatusStale, ""); err != nil {
					logger.Errorf("Failed to mark resource %s as stale: %v", existing.ID, err)
				} else {
					result.StaleCount++
				}
			}
		}
	}

	result.Message = fmt.Sprintf("Discover completed: %d new, %d stale, %d unchanged", result.NewCount, result.StaleCount, result.UnchangedCount)
	return result, items, nil
}

func filesetSourceIdentifier(fs *interfaces.FilesetMeta) string {
	if fs.DisplayPath != "" {
		return fs.DisplayPath
	}
	return fs.ID
}

func (dh *discoverHandler) createFilesetResource(ctx context.Context, catalog *interfaces.Catalog, fs *interfaces.FilesetMeta, sourceIdentifier string) (*interfaces.Resource, error) {
	meta := cloneMetadata(fs.SourceMetadata)
	req := &interfaces.ResourceRequest{
		CatalogID:        catalog.ID,
		Name:             fs.Name,
		Category:         interfaces.ResourceCategoryFileset,
		Status:           interfaces.ResourceStatusActive,
		Database:         "",
		SourceIdentifier: sourceIdentifier,
		SourceMetadata:   meta,
	}
	resource, err := dh.rs.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return resource, nil
}

func cloneMetadata(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	if m == nil {
		return out
	}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (dh *discoverHandler) enrichFilesetMetadata(ctx context.Context, items []filesetDiscoverItem) error {
	for _, item := range items {
		fs := item.meta
		resource := item.resource

		sourceMetadata := cloneMetadata(resource.SourceMetadata)
		if fs.SourceMetadata != nil {
			for k, v := range fs.SourceMetadata {
				sourceMetadata[k] = v
			}
		}
		resource.SourceMetadata = sourceMetadata
		resource.SchemaDefinition = []*interfaces.Property{}
		for _, col := range fs.Columns {
			resource.SchemaDefinition = append(resource.SchemaDefinition, &interfaces.Property{
				Name:         col.Name,
				Type:         col.Type,
				OrigType:     col.OrigType,
				DisplayName:  col.Name,
				OriginalName: col.Name,
				Description:  "",
			})
		}

		if err := dh.rs.UpdateResource(ctx, resource); err != nil {
			logger.Errorf("Failed to update fileset resource %s: %v", resource.ID, err)
			return err
		}
		logger.Infof("Enriched fileset resource %s (%s)", resource.Name, fs.ID)
	}
	return nil
}

// createAndConnectConnector creates and connects a connector for the catalog.
func (dh *discoverHandler) createAndConnectConnector(ctx context.Context, catalog *interfaces.Catalog) (connectors.Connector, error) {

	// 创建 connector
	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	// 连接
	if err := connector.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return connector, nil
}

// discoverTableResources discovers table resources from a table connector.
// 分步执行：1. 获取表名列表 2. 创建/更新 Resource 3. 逐个补齐详细元数据
func (dh *discoverHandler) discoverTableResources(ctx context.Context, catalog *interfaces.Catalog, connector connectors.Connector, taskID string) (*interfaces.DiscoverResult, error) {

	tableConnector, ok := connector.(connectors.TableConnector)
	if !ok {
		return nil, fmt.Errorf("connector does not support table discover")
	}

	// Step 1: 获取表名列表
	sourceTables, err := tableConnector.ListTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	logger.Infof("Discovered %d tables from source", len(sourceTables))

	// Step 2: 获取现有 Resources
	existingResources, err := dh.rs.GetByCatalogID(ctx, catalog.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing resources: %w", err)
	}

	// Step 3: 对比并创建/更新 Resource（基础信息）
	result, items, err := dh.reconcileTableResources(ctx, catalog, sourceTables, existingResources, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile resources: %w", err)
	}

	// Step 4: 逐个补齐详细元数据:元数据采集就是补充每一个table的元数据信息
	if err := dh.enrichTableMetadata(ctx, tableConnector, items); err != nil {
		return nil, fmt.Errorf("failed to enrich table metadata: %w", err)
	}

	logger.Infof("Discover completed for catalog %s: new=%d, stale=%d, unchanged=%d", catalog.ID, result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, nil
}

// enrichTableMetadata 为表元数据添加详细信息
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - tableConnector: 表连接器，用于获取表的元数据
//   - items: 表发现项目列表，包含表元数据和资源信息
//
// 返回值:
//   - error: 如果在处理过程中发生错误，则返回错误信息
func (dh *discoverHandler) enrichTableMetadata(ctx context.Context, tableConnector connectors.TableConnector, items []tableDiscoverItem) error {

	// 遍历所有表发现项目
	for _, item := range items {
		table := item.tableMeta   // 获取表元数据
		resource := item.resource // 获取资源信息

		// 获取详细元数据
		err := tableConnector.GetTableMeta(ctx, table)
		if err != nil {
			logger.Warnf("Failed to get metadata for table %s: %v", table.Name, err)
			return err
		}

		// 填充 Resource 元数据 ：schema_definition 字段
		resource.Database = table.Database
		resource.SchemaDefinition = []*interfaces.Property{}
		for _, column := range table.Columns {
			resource.SchemaDefinition = append(resource.SchemaDefinition, &interfaces.Property{
				Name:         column.Name,
				Type:         column.Type,
				DisplayName:  column.Name,
				OriginalName: column.Name,
				Description:  column.Description,
			})
		}
		// 填充 Resource 元数据 ：source_metadata 字段
		sourceMetadata := make(map[string]any)
		if resource.SourceMetadata != nil {
			sourceMetadata = resource.SourceMetadata
		}
		sourceMetadata["columns"] = table.Columns
		if table.TableType != "" {
			sourceMetadata["table_type"] = table.TableType
		}
		if len(table.Properties) > 0 {
			sourceMetadata["properties"] = table.Properties
		}
		if len(table.PKs) > 0 {
			sourceMetadata["primary_keys"] = table.PKs
		}
		if len(table.Indices) > 0 {
			sourceMetadata["indices"] = table.Indices
		}
		if len(table.ForeignKeys) > 0 {
			sourceMetadata["foreign_keys"] = table.ForeignKeys
		}
		resource.SourceMetadata = sourceMetadata

		// 更新 Resource
		if err := dh.rs.UpdateResource(ctx, resource); err != nil {
			logger.Errorf("Failed to update metadata for table %s: %v", table.Name, err)
			return err
		}

		logger.Debugf("Enriched table %s: properties=%v, columns=%d, indices=%d, foreign_keys=%d", table.Name, table.Properties, len(table.Columns), len(table.Indices), len(table.ForeignKeys))
	}
	return nil
}

// reconcileTableResources reconciles source tables with existing resources.
func (dh *discoverHandler) reconcileTableResources(ctx context.Context, catalog *interfaces.Catalog, sourceTables []*interfaces.TableMeta, existingResources []*interfaces.Resource, taskID string) (*interfaces.DiscoverResult, []tableDiscoverItem, error) {

	result := &interfaces.DiscoverResult{
		CatalogID: catalog.ID,
	}

	// 用于返回的 Discover Items
	var items []tableDiscoverItem

	// 构建现有资源的 map（按 SourceIdentifier 索引）
	existingMap := make(map[string]*interfaces.Resource)
	for _, r := range existingResources {
		existingMap[r.SourceIdentifier] = r
	}
	// 查出来执行的task
	task, err := dh.dts.GetByID(ctx, taskID)
	if err != nil {
		logger.Errorf("Failed to get task %s: %v", taskID, err)
		return nil, nil, err
	}
	// 构建源端表的 map
	sourceMap := make(map[string]*interfaces.TableMeta)
	for _, t := range sourceTables {
		sourceIdentifier := dh.buildSourceIdentifier(t)
		sourceMap[sourceIdentifier] = t
	}
	// 将策略转换为 map 以便快速查找
	strategyMap := make(map[string]bool)
	for _, strategy := range task.Strategies {
		strategyMap[strategy] = true
	}
	// 处理新增和保持的资源
	for _, table := range sourceTables {
		sourceIdentifier := dh.buildSourceIdentifier(table)

		if resource, ok := existingMap[sourceIdentifier]; ok {
			// 已存在，检查状态
			if len(task.Strategies) == 0 || strategyMap["update"] {
				if resource.Status == interfaces.ResourceStatusStale {
					// 之前标记为 stale，现在重新激活
					if err := dh.rs.UpdateStatus(ctx, resource.ID, interfaces.ResourceStatusActive, ""); err != nil {
						logger.Errorf("Failed to reactivate resource %s: %v", resource.ID, err)
					}
				}
				result.UnchangedCount++
				items = append(items, tableDiscoverItem{
					resource:  resource,
					tableMeta: table,
				})
			}
		} else {
			// 新增资源 - 只在策略包含 "insert" 或没有策略时处理
			if len(task.Strategies) == 0 || strategyMap["insert"] {
				resource, err := dh.createResource(ctx, catalog, table, sourceIdentifier)
				if err != nil {
					logger.Errorf("Failed to create resource %s: %v", sourceIdentifier, err)
				} else {
					result.NewCount++
					items = append(items, tableDiscoverItem{
						resource:  resource,
						tableMeta: table,
					})
				}
			}
		}
	}

	// 处理已删除的资源（标记为 stale） - 只在策略包含 "delete" 或没有策略时处理
	if len(task.Strategies) == 0 || strategyMap["delete"] {
		for sourceIdentifier, existing := range existingMap {
			if _, ok := sourceMap[sourceIdentifier]; !ok {
				// 源端不存在，标记为 stale
				if existing.Status != interfaces.ResourceStatusStale {
					if err := dh.rs.UpdateStatus(ctx, existing.ID, interfaces.ResourceStatusStale, ""); err != nil {
						logger.Errorf("Failed to mark resource %s as stale: %v", existing.ID, err)
					} else {
						result.StaleCount++
					}
				}
			}
		}
	}
	result.Message = fmt.Sprintf("Discover completed: %d new, %d stale, %d unchanged", result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, items, nil
}

// buildSourceIdentifier builds the source identifier for a table.
func (dh *discoverHandler) buildSourceIdentifier(table *interfaces.TableMeta) string {
	identifier := table.Name
	if table.Schema != "" {
		identifier = fmt.Sprintf("%s.%s", table.Schema, identifier)
	}
	if table.Database != "" {
		return fmt.Sprintf("%s.%s", table.Database, identifier)
	}
	return identifier
}

// createResource creates a new resource.
func (dh *discoverHandler) createResource(ctx context.Context, catalog *interfaces.Catalog, table *interfaces.TableMeta, sourceIdentifier string) (*interfaces.Resource, error) {

	req := &interfaces.ResourceRequest{
		CatalogID:        catalog.ID,
		Name:             sourceIdentifier,
		Description:      table.Description,
		Category:         interfaces.ResourceCategoryTable,
		Status:           interfaces.ResourceStatusActive,
		Database:         table.Database,
		SourceIdentifier: sourceIdentifier,
	}
	resource, err := dh.rs.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// discoverIndexResources discovers index resources from an index connector.
// discoverIndexResources 是一个发现索引资源的方法，它会从连接器中获取索引列表，
// 然后与现有资源进行协调，最后丰富索引的元数据信息
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - catalog: 目录接口，包含目录的相关信息
//   - connector: 连接器接口，用于与数据源进行交互
//
// 返回值:
//   - *interfaces.DiscoverResult: 发现结果，包含新资源、过期资源和未变化资源的统计信息
//   - error: 错误信息，如果在发现过程中出现错误则返回
func (dh *discoverHandler) discoverIndexResources(ctx context.Context, catalog *interfaces.Catalog, connector connectors.Connector) (*interfaces.DiscoverResult, error) {

	// 检查连接器是否实现了IndexConnector接口
	indexConnector, ok := connector.(connectors.IndexConnector)
	if !ok {
		return nil, fmt.Errorf("connector does not support index discover")
	}

	// Step 1: List Indices：获取所有的index
	sourceIndices, err := indexConnector.ListIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}
	logger.Infof("Discovered %d indices from source", len(sourceIndices))

	// Step 2: Get Existing Resources：查出db是否已存在，然后做比对
	existingResources, err := dh.rs.GetByCatalogID(ctx, catalog.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing resources: %w", err)
	}

	// Step 3: Reconcile:将index数据获取并插入：
	result, items, err := dh.reconcileIndexResources(ctx, catalog, sourceIndices, existingResources)
	if err != nil {
		return nil, fmt.Errorf("failed to reconcile resources: %w", err)
	}

	// Step 4: Enrich ： 为索引项丰富元数据信息
	if err := dh.enrichIndexMetadata(ctx, indexConnector, items); err != nil {
		return nil, fmt.Errorf("failed to enrich index metadata: %w", err)
	}

	logger.Infof("Discover completed for catalog %s: new=%d, stale=%d, unchanged=%d", catalog.ID, result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, nil
}

// reconcileIndexResources reconciles source indices with existing resources.
// reconcileIndexResources 协调索引资源，处理新资源、现有资源和过期资源
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - catalog: 目录信息，包含ID等元数据
//   - sourceIndices: 源索引元数据列表
//   - existingResources: 已存在的资源列表
//
// 返回值:
//   - *interfaces.DiscoverResult: 发现结果，包含目录ID和各类资源的统计信息
//   - []indexDiscoverItem: 索引发现项目列表，包含资源和索引元数据
//   - error: 错误信息，如果处理过程中出现错误则返回
func (dh *discoverHandler) reconcileIndexResources(ctx context.Context, catalog *interfaces.Catalog, sourceIndices []*interfaces.IndexMeta, existingResources []*interfaces.Resource) (*interfaces.DiscoverResult, []indexDiscoverItem, error) {

	// 初始化发现结果，设置目录ID
	result := &interfaces.DiscoverResult{
		CatalogID: catalog.ID,
	}

	var items []indexDiscoverItem // 索引发现项目列表

	// 创建已存在资源的映射，以源标识符为键
	existingMap := make(map[string]*interfaces.Resource)
	for _, r := range existingResources {
		existingMap[r.SourceIdentifier] = r
	}

	// 创建源索引映射，以索引名为键
	sourceMap := make(map[string]*interfaces.IndexMeta)
	for _, idx := range sourceIndices {
		sourceMap[idx.Name] = idx
	}

	// Handle new and existing
	for _, idx := range sourceIndices {
		sourceIdentifier := idx.Name //test-index

		if resource, ok := existingMap[sourceIdentifier]; ok {
			if resource.Status == interfaces.ResourceStatusStale {
				if err := dh.rs.UpdateStatus(ctx, resource.ID, interfaces.ResourceStatusActive, ""); err != nil {
					logger.Errorf("Failed to reactivate resource %s: %v", resource.ID, err)
				}
			}
			result.UnchangedCount++
			items = append(items, indexDiscoverItem{
				resource:  resource,
				indexMeta: idx,
			})
		} else {
			resource, err := dh.createIndexResource(ctx, catalog, idx)
			if err != nil {
				logger.Errorf("Failed to create resource %s: %v", sourceIdentifier, err)
			} else {
				result.NewCount++
				items = append(items, indexDiscoverItem{
					resource:  resource,
					indexMeta: idx,
				})
			}
		}
	}

	// Handle stale
	for sourceIdentifier, existing := range existingMap {
		if _, ok := sourceMap[sourceIdentifier]; !ok {
			if existing.Status != interfaces.ResourceStatusStale {
				if err := dh.rs.UpdateStatus(ctx, existing.ID, interfaces.ResourceStatusStale, ""); err != nil {
					logger.Errorf("Failed to mark resource %s as stale: %v", existing.ID, err)
				} else {
					result.StaleCount++
				}
			}
		}
	}

	result.Message = fmt.Sprintf("Discover completed: %d new, %d stale, %d unchanged", result.NewCount, result.StaleCount, result.UnchangedCount)

	return result, items, nil
}

// createIndexResource creates a new resource for an index.
func (dh *discoverHandler) createIndexResource(ctx context.Context, catalog *interfaces.Catalog, index *interfaces.IndexMeta) (*interfaces.Resource, error) {

	req := &interfaces.ResourceRequest{
		CatalogID:        catalog.ID,
		Name:             index.Name,
		Category:         interfaces.ResourceCategoryIndex,
		Status:           interfaces.ResourceStatusActive,
		SourceIdentifier: index.Name,
	}
	resource, err := dh.rs.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// enrichIndexMetadata enriches index resources with detailed metadata.
// enrichIndexMetadata 为索引项丰富元数据信息
// 参数:
//   - ctx: 上下文信息，用于控制请求的超时和取消
//   - indexConnector: 索引连接器，用于获取索引的元数据
//   - items: 需要丰富元数据的索引项列表
//
// 返回值:
//   - error: 如果在处理过程中发生错误，则返回错误信息
func (dh *discoverHandler) enrichIndexMetadata(ctx context.Context, indexConnector connectors.IndexConnector, items []indexDiscoverItem) error {

	// 遍历所有需要处理的索引项
	for _, item := range items {
		idx := item.indexMeta
		resource := item.resource

		// Get detailed metadata (mappings) : 获取index详细的信息
		if err := indexConnector.GetIndexMeta(ctx, idx); err != nil {
			logger.Warnf("Failed to get metadata for index %s: %v", idx.Name, err)
			return err
		}

		// Map fields to SchemaDefinition
		var columns []*interfaces.Property
		for _, field := range idx.Mapping {
			// {"ignore_above":256,"type":"keyword"} : 去掉type
			delete(field.Features, "type")

			columns = append(columns, &interfaces.Property{
				Name:         field.Name,
				Type:         field.Type,
				OrigType:     field.OrigType,
				DisplayName:  field.Name,
				OriginalName: field.Name,
				Description:  "",
				Attributes:   field.Features,
			})
		}
		resource.SchemaDefinition = columns

		// Populate SourceMetadata
		sourceMetadata := make(map[string]any)
		if resource.SourceMetadata != nil {
			sourceMetadata = resource.SourceMetadata
		}

		sourceMetadata["properties"] = idx.Properties
		sourceMetadata["mapping"] = idx.Mapping
		resource.SourceMetadata = sourceMetadata

		// Update Resource
		if err := dh.rs.UpdateResource(ctx, resource); err != nil {
			logger.Errorf("Failed to update metadata for index %s: %v", idx.Name, err)
			return err
		}

		// Wait a bit to avoid overwhelming the server? No, it's fine for now.
		// Just logging
		logger.Infof("Enriched index %s: fields=%d", idx.Name, len(columns))
	}
	return nil
}
