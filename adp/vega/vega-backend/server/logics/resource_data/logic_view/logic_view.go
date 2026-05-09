package logic_view

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/antlr4-go/antlr/v4"
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/filter_condition"
	"vega-backend/logics/permission"
	"vega-backend/logics/query"
	"vega-backend/logics/resource"
	lvdsl "vega-backend/logics/resource_data/logic_view/dsl"
	lvsql "vega-backend/logics/resource_data/logic_view/sql"
	"vega-backend/logics/resource_data/logic_view/sql/parsing"
)

var (
	lvServiceOnce sync.Once
	lvService     interfaces.LogicViewService
)

type logicViewService struct {
	appSetting *common.AppSetting
	cs         interfaces.CatalogService
	rs         interfaces.ResourceService
	ps         interfaces.PermissionService
	qs         interfaces.SQLQueryService
}

// NewLogicViewService creates a new ResourceDataService.
func NewLogicViewService(appSetting *common.AppSetting) interfaces.LogicViewService {
	lvServiceOnce.Do(func() {
		lvService = &logicViewService{
			appSetting: appSetting,
			cs:         catalog.NewCatalogService(appSetting),
			rs:         resource.NewResourceService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			qs:         query.NewSQLQueryService(appSetting),
		}
	})
	return lvService
}

func (lvs *logicViewService) Query(ctx context.Context, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Query logic view")
	defer span.End()

	logger.Debugf("Query logic view, resourceID: %s, params: %v",
		resource.ID, params)

	view := &interfaces.LogicView{
		Resource: *resource,
	}

	switch resource.LogicType {
	case interfaces.LogicType_Derived:
		return lvs.queryDerivedLogicView(ctx, view, params)
	case interfaces.LogicType_Composite:
		return lvs.queryCompositeLogicView(ctx, view, params)
	default:
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(fmt.Sprintf("The logic type of the custom view '%s' is not supported", resource.ID))
	}
}

func (lvs *logicViewService) queryDerivedLogicView(ctx context.Context, view *interfaces.LogicView,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query derived logic view")
	defer span.End()

	var inputNode *interfaces.LogicDefinitionNode
	for _, node := range view.LogicDefinition {
		if node.Type == interfaces.LogicDefinitionNodeType_Resource {
			inputNode = node
			break
		}
	}

	var nodeCfg interfaces.ResourceNodeCfg
	if err := mapstructure.Decode(inputNode.Config, &nodeCfg); err != nil {
		span.SetStatus(codes.Error, "Decode resource node config failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to decode resource node config: %v", err))
	}
	fromResourceFilterCond := nodeCfg.Filters

	fromResource, err := lvs.rs.GetByID(ctx, nodeCfg.ResourceID)
	if err != nil {
		span.SetStatus(codes.Error, "Get source resource failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to get source resource %s: %v", nodeCfg.ResourceID, err))
	}
	if fromResource == nil {
		span.SetStatus(codes.Error, "Source resource not found")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
			WithErrorDetails(fmt.Sprintf("source resource %s not found", nodeCfg.ResourceID))
	}

	catalog, err := lvs.cs.GetByID(ctx, fromResource.CatalogID, true)
	if err != nil {
		span.SetStatus(codes.Error, "Get catalog failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to get catalog: %v", err))
	}
	if catalog == nil {
		span.SetStatus(codes.Error, "Catalog not found")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_CatalogNotFound).
			WithErrorDetails(fmt.Sprintf("catalog %s not found", fromResource.CatalogID))
	}

	fieldMap := map[string]*interfaces.Property{}
	outputFields := make([]string, 0, len(view.SchemaDefinition))
	for _, prop := range view.SchemaDefinition {
		fieldMap[prop.Name] = prop
		outputFields = append(outputFields, prop.Name)
	}
	params.OutputFields = outputFields

	// 合并资源和查询的 FilterCondCfg, 需要判断下是否为nil
	var mergedFilterCond *interfaces.FilterCondCfg
	if fromResourceFilterCond != nil && params.FilterCondCfg != nil {
		mergedFilterCond = &interfaces.FilterCondCfg{
			Operation: filter_condition.OperationAnd,
			SubConds:  []*interfaces.FilterCondCfg{fromResourceFilterCond, params.FilterCondCfg},
		}
	} else if fromResourceFilterCond != nil {
		mergedFilterCond = fromResourceFilterCond
	} else if params.FilterCondCfg != nil {
		mergedFilterCond = params.FilterCondCfg
	}

	actualFilterCond, err := filter_condition.NewFilterCondition(ctx, mergedFilterCond, fieldMap)
	if err != nil {
		span.SetStatus(codes.Error, "Create filter condition failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
	}
	params.ActualFilterCond = actualFilterCond

	// 交给 executePhysicalQuery 处理 SQL push-down
	return executePhysicalQuery(ctx, catalog, fromResource, params)
}

func executePhysicalQuery(ctx context.Context, catalog *interfaces.Catalog, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Execute physical query")
	defer span.End()

	logger.Debugf("executePhysicalQuery, resourceID: %s, catalogID: %s, params: %v",
		resource.ID, resource.CatalogID, params)

	switch resource.Category {
	case interfaces.ResourceCategoryTable:
		return executeTableQuery(ctx, catalog, resource, params)
	case interfaces.ResourceCategoryIndex:
		return executeIndexQuery(ctx, catalog, resource, params)
	default:
		span.SetStatus(codes.Error, "Unsupported resource category")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(fmt.Sprintf("unsupported resource category: %s", resource.Category))
	}
}

func (lvs *logicViewService) queryCompositeLogicView(ctx context.Context, view *interfaces.LogicView,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query composite logic view")
	defer span.End()

	// input resource 的 category 决定生成 SQL 还是 DSL
	isDSL := false
	catalogMap := map[string]struct{}{}
	refResources := make(map[string]*interfaces.Resource, 0)
	for _, logicNode := range view.LogicDefinition {
		if logicNode.Type == interfaces.LogicDefinitionNodeType_Resource {
			var nodeCfg interfaces.ResourceNodeCfg
			if err := mapstructure.Decode(logicNode.Config, &nodeCfg); err != nil {
				span.SetStatus(codes.Error, "Decode resource node config failed")
				return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
					WithErrorDetails(fmt.Sprintf("failed to decode resource node config: %v", err))
			}

			fromResource, err := lvs.rs.GetByID(ctx, nodeCfg.ResourceID)
			if err != nil {
				span.SetStatus(codes.Error, "Get source resource failed")
				return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
					WithErrorDetails(fmt.Sprintf("failed to get source resource %s: %v", nodeCfg.ResourceID, err))
			}
			if fromResource == nil {
				span.SetStatus(codes.Error, "Source resource not found")
				return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
					WithErrorDetails(fmt.Sprintf("source resource %s not found", nodeCfg.ResourceID))
			}
			refResources[nodeCfg.ResourceID] = fromResource

			if fromResource.Category == interfaces.ResourceCategoryIndex {
				isDSL = true
			}
			catalogMap[fromResource.CatalogID] = struct{}{}
		}
	}
	view.RefResources = refResources

	view.IsSingleSource = len(catalogMap) == 1

	if isDSL {
		return lvs.executeCompositeViewByDSL(ctx, view, params)
	} else {
		return lvs.executeCompositeViewBySQL(ctx, view, params)
	}
}

func (lvs *logicViewService) executeCompositeViewByDSL(ctx context.Context, view *interfaces.LogicView,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Query composite view data")
	defer span.End()

	// 获取索引列表, 视图 ID 到索引列表的映射
	_, indices, viewIndicesMap, err := lvs.getIndicesByView(view)
	if err != nil {
		span.SetStatus(codes.Error, "Get indices failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).WithErrorDetails(err.Error())
	}

	// 如果索引列表为空，则返回空数据, 不需要下面拼接dsl
	if len(indices) == 0 {
		span.SetStatus(codes.Ok, "No indices found")
		return nil, 0, nil
	}

	generator := lvdsl.NewlogicViewDSLGenerator(view)
	dsl, httpErr := generator.BuildDSL(ctx, *params, view, viewIndicesMap)
	if httpErr != nil {
		span.SetStatus(codes.Error, "Convert to DSL failed")
		return nil, 0, httpErr
	}

	logger.Infof("executeCompositeViewByDSL DSL: [%s]", dsl)

	if view.IsSingleSource {
		var resourceType string
		for _, ref := range view.RefResources {
			catalog, err := lvs.cs.GetByID(ctx, ref.CatalogID, true)
			if err != nil {
				return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
					WithErrorDetails(fmt.Sprintf("failed to get catalog: %v", err))
			}
			if catalog == nil {
				return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_CatalogNotFound).
					WithErrorDetails(fmt.Sprintf("catalog %s not found", ref.CatalogID))
			}
			resourceType = catalog.ConnectorType
			break
		}

		// dsl 转为 map
		dslBytes, err := sonic.Marshal(dsl)
		if err != nil {
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(fmt.Sprintf("failed to marshal dsl: %v", err))
		}

		var dslMap map[string]any
		err = sonic.Unmarshal(dslBytes, &dslMap)
		if err != nil {
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(fmt.Sprintf("failed to unmarshal dsl: %v", err))
		}

		req := interfaces.SQLQueryRequest{
			Query:        dslMap,
			ResourceType: resourceType,
			QueryType:    params.QueryType,
			StreamSize:   params.Limit,
			QueryTimeout: int(params.Timeout),
		}
		res, err := lvs.qs.Execute(ctx, &req)
		if err != nil {
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(err.Error())
		}
		return res.Entries, res.TotalCount, nil
	} else {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusNotImplemented, rest.PublicError_NotImplemented).
			WithErrorDetails("composite view execution is not implemented")
	}
}

// 从视图获取索引列表, 返回 catalogName， viewIndicesMap（视图id到索引列表的映射）
func (lvs *logicViewService) getIndicesByView(view *interfaces.LogicView) (string, []string, map[string][]string, error) {
	var catalog string
	catalogMap := map[string]struct{}{}
	indices := []string{}
	viewIndicesMap := map[string][]string{}
	// 判断多个 view node 的 catalog 是否一致
	for _, ref := range view.RefResources {
		sourceIdentifier := strings.Split(ref.SourceIdentifier, ".")
		indices = append(indices, sourceIdentifier[len(sourceIdentifier)-1])
		viewIndicesMap[ref.ID] = append(viewIndicesMap[ref.ID], sourceIdentifier[len(sourceIdentifier)-1])

		catalog = ref.CatalogID
		catalogMap[catalog] = struct{}{}

	}

	if len(catalogMap) > 1 {
		return "", nil, nil, fmt.Errorf("custom view %s has different catalog %v", view.Name, catalogMap)
	}

	return catalog, indices, viewIndicesMap, nil

}

func (lvs *logicViewService) executeCompositeViewBySQL(ctx context.Context, view *interfaces.LogicView,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
	// 理想状态：从生成器直接获取 SQL 构建器
	ldGenerator := lvsql.NewlogicDefinitionSQLGenerator(view)
	builder, err := ldGenerator.NewQueryBuilder(ctx, view)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to initialize query builder: %v", err))
	}

	// 统一应用查询参数（过滤、排序、分页）
	if err := builder.ApplyParams(ctx, params, view); err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to apply query parameters: %v", err))
	}

	finalSql := builder.Build()
	logger.Infof("executeCompositeViewBySQL Final SQL: [%s]", finalSql)

	if view.IsSingleSource {
		var resourceType string
		for _, ref := range view.RefResources {
			catalog, err := lvs.cs.GetByID(ctx, ref.CatalogID, true)
			if err != nil {
				return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
					WithErrorDetails(fmt.Sprintf("failed to get catalog: %v", err))
			}
			if catalog == nil {
				return nil, 0, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_CatalogNotFound).
					WithErrorDetails(fmt.Sprintf("catalog %s not found", ref.CatalogID))
			}
			resourceType = catalog.ConnectorType
			break
		}

		req := interfaces.SQLQueryRequest{
			Query:        finalSql,
			ResourceType: resourceType,
			QueryType:    params.QueryType,
			StreamSize:   params.Limit,
			QueryTimeout: int(params.Timeout),
		}
		res, err := lvs.qs.Execute(ctx, &req)
		if err != nil {
			return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
				WithErrorDetails(err.Error())
		}
		return res.Entries, res.TotalCount, nil
	} else {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusNotImplemented, rest.PublicError_NotImplemented).
			WithErrorDetails("composite view execution is not implemented")
	}
}

func executeIndexQuery(ctx context.Context, catalog *interfaces.Catalog, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Execute index query")
	defer span.End()

	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		span.SetStatus(codes.Error, "Create connector failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to create connector: %v", err))
	}

	if err := connector.Connect(ctx); err != nil {
		span.SetStatus(codes.Error, "Connect to data source failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to connect to data source: %v", err))
	}
	defer func() { _ = connector.Close(ctx) }()

	indexConnector, ok := connector.(connectors.IndexConnector)
	if !ok {
		span.SetStatus(codes.Error, "Connector does not support index operations")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(fmt.Sprintf("connector %s does not support index operations", catalog.ConnectorType))
	}

	result, err := indexConnector.ExecuteQuery(ctx, resource.Name, resource, params)
	if err != nil {
		span.SetStatus(codes.Error, "Execute query failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to execute query: %v", err))
	}
	return result.Rows, result.Total, nil
}

func executeTableQuery(ctx context.Context, catalog *interfaces.Catalog, resource *interfaces.Resource,
	params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Execute table query")
	defer span.End()

	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		span.SetStatus(codes.Error, "Create connector failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to create connector: %v", err))
	}

	if err := connector.Connect(ctx); err != nil {
		span.SetStatus(codes.Error, "Connect to data source failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to connect to data source: %v", err))
	}
	defer func() { _ = connector.Close(ctx) }()

	tableConnector, ok := connector.(connectors.TableConnector)
	if !ok {
		span.SetStatus(codes.Error, "Connector does not support table operations")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(fmt.Sprintf("connector %s does not support table operations", catalog.ConnectorType))
	}

	result, err := tableConnector.ExecuteQuery(ctx, resource, params)
	if err != nil {
		span.SetStatus(codes.Error, "Execute query failed")
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(fmt.Sprintf("failed to execute query: %v", err))
	}
	return result.Rows, result.Total, nil
}

// FieldInfo 表示SQL查询输出的字段信息
type FieldInfo struct {
	Name      string `json:"name"`       // 字段名或表达式
	Alias     string `json:"alias"`      // 字段别名（如果没有别名为空）
	IsStar    bool   `json:"is_star"`    // 是否是*通配符
	IsComplex bool   `json:"is_complex"` // 是否是复杂表达式（函数、CASE等）
}

// QueryAnalysis 表示SQL查询的分析结果
type QueryAnalysis struct {
	Fields       []FieldInfo `json:"fields"`
	HasStar      bool        `json:"has_star"`
	HasUnion     bool        `json:"has_union"`
	HasJoin      bool        `json:"has_join"`
	HasAggregate bool        `json:"has_aggregate"`
	HasSubquery  bool        `json:"has_subquery"`
	HasCase      bool        `json:"has_case"`
	Error        error       `json:"error,omitempty"`
}

// String 返回分析结果的字符串表示
func (q *QueryAnalysis) String() string {
	if q.Error != nil {
		return fmt.Sprintf("分析错误: %v", q.Error)
	}

	result := fmt.Sprintf("查询字段 (%d 个):\n", len(q.Fields))
	for i, field := range q.Fields {
		fieldDesc := field.Name
		if field.Alias != "" {
			fieldDesc = fmt.Sprintf("%s AS %s", field.Name, field.Alias)
		}
		if field.IsStar {
			fieldDesc = "*"
		}
		result += fmt.Sprintf("  %d. %s\n", i+1, fieldDesc)
	}

	result += "\n查询特征:\n"
	result += fmt.Sprintf("  - 包含UNION: %t\n", q.HasUnion)
	result += fmt.Sprintf("  - 包含JOIN: %t\n", q.HasJoin)
	result += fmt.Sprintf("  - 包含聚合函数: %t\n", q.HasAggregate)
	result += fmt.Sprintf("  - 包含子查询: %t\n", q.HasSubquery)
	result += fmt.Sprintf("  - 包含CASE表达式: %t\n", q.HasCase)

	return result
}

// SQLFieldParser SQL字段解析器
type SQLFieldParser struct {
	listener *sqlFieldListener
}

// NewSQLFieldParser 创建新的SQL字段解析器
func NewSQLFieldParser() *SQLFieldParser {
	return &SQLFieldParser{
		listener: newSqlFieldListener(),
	}
}

// Parse 解析SQL语句并返回字段分析结果
func (p *SQLFieldParser) Parse(sql string) *QueryAnalysis {
	// 创建输入流
	input := antlr.NewInputStream(sql)

	// 创建词法分析器
	lexer := parsing.NewSqlBaseLexer(input)

	// 创建令牌流
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// 创建语法分析器
	parser := parsing.NewSqlBaseParser(stream)

	// 添加错误监听器
	parser.RemoveErrorListeners()
	errorListener := newErrorListener()
	parser.AddErrorListener(errorListener)

	// 构建解析树 - 根据实际语法规则调整起始规则
	tree := parser.Query() // 或者可能是 parser.Query() 或 parser.Statement()

	// 遍历解析树
	antlr.ParseTreeWalkerDefault.Walk(p.listener, tree)

	analysis := p.listener.getAnalysis()
	if errorListener.hasErrors() {
		analysis.Error = fmt.Errorf("SQL语法错误: %s", strings.Join(errorListener.getErrors(), "; "))
	}

	return analysis
}

// sqlFieldListener 自定义字段解析监听器
type sqlFieldListener struct {
	*parsing.BaseSqlBaseListener
	analysis          *QueryAnalysis
	currentQueryLevel int
	inSelectClause    bool
	currentField      *FieldInfo
}

// newSqlFieldListener 创建新的字段监听器
func newSqlFieldListener() *sqlFieldListener {
	return &sqlFieldListener{
		analysis: &QueryAnalysis{
			Fields: make([]FieldInfo, 0),
		},
		currentQueryLevel: 0,
		inSelectClause:    false,
	}
}

// getAnalysis 获取分析结果
func (l *sqlFieldListener) getAnalysis() *QueryAnalysis {
	return l.analysis
}

// EnterQuery 进入查询
func (l *sqlFieldListener) EnterQuery(ctx *parsing.QueryContext) {
	l.currentQueryLevel++
}

// ExitQuery 退出查询
func (l *sqlFieldListener) ExitQuery(ctx *parsing.QueryContext) {
	l.currentQueryLevel--
}

// EnterQuerySpecification 进入查询规范（SELECT语句）
func (l *sqlFieldListener) EnterQuerySpecification(ctx *parsing.QuerySpecificationContext) {
	l.inSelectClause = true
}

// ExitQuerySpecification 退出查询规范（SELECT语句）
func (l *sqlFieldListener) ExitQuerySpecification(ctx *parsing.QuerySpecificationContext) {
	l.inSelectClause = false
}

// EnterSelectSingle 处理SELECT单个字段
func (l *sqlFieldListener) EnterSelectSingle(ctx *parsing.SelectSingleContext) {
	if !l.inSelectClause || l.currentQueryLevel > 1 {
		// 只处理最外层的SELECT字段
		return
	}

	// 获取选择项的文本
	itemText := l.getText(ctx)

	// 创建字段信息
	l.currentField = &FieldInfo{
		Name:      itemText,
		IsStar:    l.isStarExpression(itemText),
		IsComplex: l.isComplexExpression(itemText),
	}

	if l.isStarExpression(itemText) {
		l.analysis.HasStar = true
	}

	// 检查是否有别名
	if alias := l.extractAlias(ctx); alias != "" {
		l.currentField.Alias = alias
		l.currentField.Name = strings.TrimSuffix(l.currentField.Name, " AS "+alias)
		l.currentField.Name = strings.TrimSuffix(l.currentField.Name, " "+alias)
	}

	// 添加到分析结果
	l.analysis.Fields = append(l.analysis.Fields, *l.currentField)
	l.currentField = nil
}

// isStarExpression 检查是否是*通配符
func (l *sqlFieldListener) isStarExpression(expr string) bool {
	return strings.TrimSpace(expr) == "*"
}

// isComplexExpression 检查是否是复杂表达式
func (l *sqlFieldListener) isComplexExpression(expr string) bool {
	trimmed := strings.ToUpper(strings.TrimSpace(expr))

	// 检查函数调用
	if strings.Contains(trimmed, "(") && strings.Contains(trimmed, ")") {
		return true
	}

	// 检查CASE表达式
	if strings.HasPrefix(trimmed, "CASE") {
		l.analysis.HasCase = true
		return true
	}

	// 检查聚合函数
	if l.isAggregateFunction(trimmed) {
		l.analysis.HasAggregate = true
		return true
	}

	// 检查算术表达式
	if strings.ContainsAny(trimmed, "+-*/%") {
		return true
	}

	return false
}

// isAggregateFunction 检查是否是聚合函数
func (l *sqlFieldListener) isAggregateFunction(expr string) bool {
	upperExpr := strings.ToUpper(expr)
	aggregateFuncs := []string{
		"COUNT(", "SUM(", "AVG(", "MIN(", "MAX(",
		"GROUP_CONCAT(", "ARRAY_AGG(", "STRING_AGG(",
	}

	for _, funcName := range aggregateFuncs {
		if strings.Contains(upperExpr, funcName) {
			return true
		}
	}
	return false
}

// extractAlias 提取字段别名
func (l *sqlFieldListener) extractAlias(ctx antlr.ParserRuleContext) string {
	// 根据实际的语法规则提取别名
	// 这里是一个通用实现，您可能需要根据您的g4语法调整

	children := ctx.GetChildren()
	for _, child := range children {
		if terminal, ok := child.(antlr.TerminalNode); ok {
			text := terminal.GetText()
			upperText := strings.ToUpper(text)
			if upperText == "AS" {
				// 找到AS关键字，下一个兄弟节点应该是别名
				return l.getNextSiblingText(child)
			}
		}
	}

	// 如果没有AS关键字，检查最后一个子节点是否可能是别名
	// 这需要根据具体语法规则调整
	return ""
}

// getNextSiblingText 获取下一个兄弟节点的文本
func (l *sqlFieldListener) getNextSiblingText(node antlr.Tree) string {
	parent := getParent(node)
	if parent == nil {
		return ""
	}

	children := parent.GetChildren()
	found := false
	for _, child := range children {
		if found {
			return l.getText(child)
		}
		if child == node {
			found = true
		}
	}
	return ""
}

// getParent 获取节点的父节点
func getParent(node antlr.Tree) antlr.Tree {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case antlr.ParserRuleContext:
		return n.GetParent()
	case antlr.TerminalNode:
		return n.GetParent()
	default:
		return nil
	}
}

// getText 安全获取节点的文本内容
func (l *sqlFieldListener) getText(node antlr.Tree) string {
	if node == nil {
		return ""
	}

	switch ctx := node.(type) {
	case antlr.ParserRuleContext:
		return ctx.GetText()
	case antlr.TerminalNode:
		return ctx.GetText()
	default:
		return fmt.Sprintf("%v", node)
	}
}

// EnterSetOperation 处理UNION操作
func (l *sqlFieldListener) EnterSetOperation(ctx *parsing.SetOperationContext) {
	l.analysis.HasUnion = true
}

// EnterJoinRelation 处理JOIN关系
func (l *sqlFieldListener) EnterJoinRelation(ctx *parsing.JoinRelationContext) {
	l.analysis.HasJoin = true
}

// EnterSubquery 处理子查询
func (l *sqlFieldListener) EnterSubquery(ctx *parsing.SubqueryContext) {
	if l.currentQueryLevel > 0 {
		l.analysis.HasSubquery = true
	}
}

// errorListener 自定义错误监听器
type errorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

// newErrorListener 创建错误监听器
func newErrorListener() *errorListener {
	return &errorListener{
		errors: make([]string, 0),
	}
}

// SyntaxError 处理语法错误
func (l *errorListener) SyntaxError(
	recognizer antlr.Recognizer,
	offendingSymbol interface{},
	line, column int,
	msg string,
	e antlr.RecognitionException,
) {
	errorMsg := fmt.Sprintf("第%d行第%d列: %s", line, column, msg)
	l.errors = append(l.errors, errorMsg)
}

// hasErrors 检查是否有错误
func (l *errorListener) hasErrors() bool {
	return len(l.errors) > 0
}

// getErrors 获取所有错误
func (l *errorListener) getErrors() []string {
	return l.errors
}

// GetFieldNames 获取所有字段名称（优先使用别名，没有别名使用字段名）
func (q *QueryAnalysis) GetFieldNames() []string {
	names := make([]string, len(q.Fields))
	for i, field := range q.Fields {
		if field.Alias != "" {
			names[i] = field.Alias
		} else if field.IsStar {
			names[i] = "*"
		} else {
			names[i] = field.Name
		}
	}
	return names
}

// HasComplexFields 检查是否包含复杂字段
func (q *QueryAnalysis) HasComplexFields() bool {
	for _, field := range q.Fields {
		if field.IsComplex {
			return true
		}
	}
	return false
}

// GetSimpleFieldNames 获取简单字段名称（排除复杂表达式和*）
func (q *QueryAnalysis) GetSimpleFieldNames() []string {
	var names []string
	for _, field := range q.Fields {
		if !field.IsComplex && !field.IsStar {
			if field.Alias != "" {
				names = append(names, field.Alias)
			} else {
				names = append(names, field.Name)
			}
		}
	}
	return names
}

// FormatAsJSON 将解析结果格式化为 JSON
func (info *QueryAnalysis) FormatAsJSON() string {
	jsonData, err := sonic.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "JSON 格式化失败: %v"}`, err)
	}
	return string(jsonData)
}
