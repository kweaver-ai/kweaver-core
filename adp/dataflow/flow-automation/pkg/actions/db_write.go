package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils/writer"
)

// Database type constants (deprecated, use writer package instead)
const (
	DatabaseTypeMySQL      = writer.DatabaseTypeMySQL
	DatabaseTypePostgreSQL = writer.DatabaseTypePostgreSQL
	DatabaseTypePostgres   = writer.DatabaseTypePostgres
	DatabaseTypeDM8        = writer.DatabaseTypeDM8
	DatabaseTypeKDB        = writer.DatabaseTypeKDB
	DatabaseTypeSQLServer  = writer.DatabaseTypeSQLServer
	DatabaseTypeMSSQL      = writer.DatabaseTypeMSSQL
	DatabaseTypeOracle     = writer.DatabaseTypeOracle
)

// Database connection constants (deprecated, use writer package instead)
const (
	DefaultMaxIdleConns     = writer.DefaultMaxIdleConns
	DefaultMaxOpenConns     = writer.DefaultMaxOpenConns
	DefaultBatchSize        = writer.DefaultBatchSize
	DefaultVarcharLength    = writer.DefaultVarcharLength
	DefaultDecimalPrecision = writer.DefaultDecimalPrecision
	DefaultDecimalScale     = writer.DefaultDecimalScale
	DefaultEnumLength       = writer.DefaultEnumLength
	PrimaryKeyFlag          = writer.PrimaryKeyFlag
)

// Database operation constants (deprecated, use writer package instead)
const (
	OperationInsert = writer.OperationInsert
	OperationUpdate = writer.OperationUpdate
	OperationDelete = writer.OperationDelete
	OperationAppend = writer.OperationAppend
)

// Mathematical constants (deprecated, use writer package instead)
const (
	DecimalBase           = writer.DecimalBase
	MinBatchDataThreshold = writer.MinBatchDataThreshold
)

// Database connection types (deprecated, use writer package instead)
type DBConn = writer.DBConn
type FieldAttr = writer.FieldAttr
type FieldMapping = writer.FieldMapping
type SyncOptions = writer.SyncOptions
type DatabaseConfig = writer.DatabaseConfig

// DatabaseWrite 任务参数（仅使用新结构）
type DatabaseWrite struct {
	// 基本数据源信息
	DatasourceType string `json:"datasource_type"` // mysql/postgresql/dm8/kdb
	DatasourceID   string `json:"datasource_id,omitempty"`
	DatasourceName string `json:"datasource_name,omitempty"`

	// 目标表信息
	TableExist  bool   `json:"table_exist,omitempty"`
	TableName   string `json:"table_name"`
	OperateType string `json:"operate_type"` // append

	// 目标端连接信息
	Conn *DBConn `json:"conn"`

	// 字段映射
	SyncModelFields []FieldMapping `json:"sync_model_fields,omitempty"`

	// 写入选项
	SyncOptions *SyncOptions `json:"sync_options,omitempty"`

	// 业务数据（如仍需直接写入）
	Data  interface{} `json:"data,omitempty"`
	Where interface{} `json:"where,omitempty"`
}

// Name 操作名称
func (a *DatabaseWrite) Name() string {
	return common.DatabaseWriteOpt
}

// ParameterNew 初始化参数
func (a *DatabaseWrite) ParameterNew() interface{} {
	return &DatabaseWrite{}
}

// 移除getFullTableName方法，表名生成现在由各个database driver负责

// Run 执行数据库写入操作
func (a *DatabaseWrite) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() {
		trace.TelemetrySpanEnd(span, err)
	}()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "database write start", entity.TraceOpPersistAfterAction)
	input := params.(*DatabaseWrite)

	// 验证配置
	if err := a.validateDatabaseConfig(input); err != nil {
		ctx.Trace(ctx.Context(), "database config validation failed: "+err.Error(), entity.TraceOpPersistAfterAction)
		return nil, fmt.Errorf("database config validation failed: %w", err)
	}

	// 如果没有直接提供连接信息，通过datasource_id解析
	if input.Conn == nil && input.DatasourceID != "" {
		ctx.Trace(ctx.Context(), "resolving connection info from datasource_id: "+input.DatasourceID, entity.TraceOpPersistAfterAction)
		conn, err := a.resolveConnByDatasourceID(input, token)
		if err != nil {
			ctx.Trace(ctx.Context(), "failed to resolve connection by datasource ID: "+err.Error(), entity.TraceOpPersistAfterAction)
			return nil, fmt.Errorf("failed to resolve connection by datasource ID: %w", err)
		}
		input.Conn = conn
		ctx.Trace(ctx.Context(), "connection info resolved successfully", entity.TraceOpPersistAfterAction)
	}

	// 解析数据（可选）
	ctx.Trace(ctx.Context(), "parsing input data", entity.TraceOpPersistAfterAction)
	data, err := a.parseData(ctx, input.Data)
	if err != nil {
		ctx.Trace(ctx.Context(), "failed to parse data: "+err.Error(), entity.TraceOpPersistAfterAction)
		return nil, fmt.Errorf("failed to parse data: %w", err)
	}

	// 应用字段映射（source.name -> target.name）
	var mappedData []map[string]interface{}
	var fieldMappings []FieldMapping

	if len(input.SyncModelFields) > 0 {
		fieldMappings = input.SyncModelFields
		ctx.Trace(ctx.Context(), fmt.Sprintf("using provided field mapping for %d fields", len(fieldMappings)), entity.TraceOpPersistAfterAction)
	} else {
		// 如果没有提供字段映射，从数据中自动推断
		ctx.Trace(ctx.Context(), "no field mapping provided, inferring fields from data", entity.TraceOpPersistAfterAction)
		inferredMappings, err := a.inferFieldMappings(data)
		if err != nil {
			ctx.Trace(ctx.Context(), "failed to infer field mappings: "+err.Error(), entity.TraceOpPersistAfterAction)
			return nil, fmt.Errorf("failed to infer field mappings: %w", err)
		}
		fieldMappings = inferredMappings
		ctx.Trace(ctx.Context(), fmt.Sprintf("inferred %d field mappings from data", len(fieldMappings)), entity.TraceOpPersistAfterAction)
	}

	// 使用字段映射转换数据
	mappedData, err = a.transformDataByMapping(data, fieldMappings)
	if err != nil {
		ctx.Trace(ctx.Context(), "failed to transform data by mapping: "+err.Error(), entity.TraceOpPersistAfterAction)
		return nil, fmt.Errorf("failed to transform data by mapping: %w", err)
	}
	ctx.Trace(ctx.Context(), fmt.Sprintf("field mapping completed, resulting in %d records", len(mappedData)), entity.TraceOpPersistAfterAction)

	// 检查是否有数据要写入
	if len(mappedData) == 0 {
		ctx.Trace(ctx.Context(), "no data to write - mappedData is empty", entity.TraceOpPersistAfterAction)
		return map[string]interface{}{
			"affected_rows":   0,
			"operation":       input.OperateType,
			"table":           input.TableName, // 表名生成由driver负责，这里使用原始表名
			"success":         true,
			"success_count":   0,
			"failed_count":    0,
			"total_processed": 0,
			"failed_records":  []map[string]interface{}{},
			"failure_reasons": map[string]int{},
			"message":         "no data to write",
		}, nil
	}

	ctx.Trace(ctx.Context(), fmt.Sprintf("prepared %d records for database operation: %s", len(mappedData), input.OperateType), entity.TraceOpPersistAfterAction)

	var where interface{}
	if input.Where != nil {
		where, err = a.parseData(ctx, input.Where)
		if err != nil {
			ctx.Trace(ctx.Context(), "failed to parse where condition: "+err.Error(), entity.TraceOpPersistAfterAction)
			return nil, fmt.Errorf("failed to parse where condition: %w", err)
		}
	}

	// 使用新的writer包执行数据库操作
	tableInfo := &writer.TableInfo{
		DatasourceType: input.DatasourceType,
		TableName:      input.TableName,
		TableExist:     input.TableExist,
		Conn:           input.Conn,
		Fields:         fieldMappings,
		Options:        input.SyncOptions,
	}

	operation := strings.ToLower(input.OperateType)
	result, err := writer.GetGlobalWriterFullyDistributed().Execute(ctx.Context(), tableInfo, mappedData, where, operation)
	if err != nil {
		return nil, fmt.Errorf("database operation failed: %w", err)
	}

	taskID := ctx.GetTaskID()
	ctx.ShareData().Set(taskID, result)
	ctx.Trace(ctx.Context(), "run end")

	if !result.Success {
		return result, fmt.Errorf("database operation failed: %w", result.FailureReasons)
	}
	return result, nil
}

// validateDatabaseConfig 验证数据库配置
func (a *DatabaseWrite) validateDatabaseConfig(input *DatabaseWrite) error {
	if input.DatasourceType == "" {
		return fmt.Errorf("datasource_type is required")
	}

	// 动态检查是否支持该数据库类型（通过writer包）
	if !writer.IsDatabaseTypeSupported(input.DatasourceType) {
		supportedTypes := writer.GetSupportedDatabaseTypes()
		if len(supportedTypes) > 0 {
			return fmt.Errorf("unsupported database type: %s, supported types: %v", input.DatasourceType, supportedTypes)
		} else {
			return fmt.Errorf("unsupported database type: %s, no database drivers registered", input.DatasourceType)
		}
	}

	if input.TableName == "" {
		return fmt.Errorf("table_name is required")
	}

	// 允许未传 conn，由 datasource_id 解析（未来扩展）
	if input.Conn == nil && input.DatasourceID == "" {
		return fmt.Errorf("either conn or datasource_id is required")
	}

	return nil
}

// resolveConnByDatasourceID 通过数据源ID解析连接信息
func (a *DatabaseWrite) resolveConnByDatasourceID(input *DatabaseWrite, token *entity.Token) (*DBConn, error) {
	if strings.TrimSpace(input.DatasourceID) == "" {
		return nil, fmt.Errorf("datasource_id is required but not provided")
	}

	if token == nil {
		return nil, fmt.Errorf("token is required for datasource_id resolution")
	}

	// 创建数据源客户端
	dataSourceClient := drivenadapters.NewDataSource()

	// 提取token字符串
	tokenStr := ""
	if token.Token != "" {
		tokenStr = token.Token
	}

	// 获取客户端IP地址
	ipStr := ""
	if token.LoginIP != "" {
		ipStr = token.LoginIP
	}

	// 调用GetDataSourceCatalog接口获取数据源信息
	catalog, err := dataSourceClient.GetDataSourceCatalog(context.Background(), input.DatasourceID, tokenStr, ipStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get data source catalog for %s: %w", input.DatasourceID, err)
	}

	// 检查数据源信息是否完整
	if catalog == nil {
		return nil, fmt.Errorf("data source catalog is nil for %s", input.DatasourceID)
	}

	if catalog.BinData == nil {
		return nil, fmt.Errorf("bin_data is nil in data source catalog for %s", input.DatasourceID)
	}

	// 检查必要的连接信息是否完整
	if catalog.BinData.Host == "" {
		return nil, fmt.Errorf("host is empty in data source catalog for %s", input.DatasourceID)
	}

	if catalog.BinData.Port == 0 {
		return nil, fmt.Errorf("port is invalid in data source catalog for %s", input.DatasourceID)
	}

	if catalog.BinData.DatabaseName == "" {
		return nil, fmt.Errorf("database_name is empty in data source catalog for %s", input.DatasourceID)
	}

	// 对password进行RSA解密（先base64解码，再RSA解密）
	decodedPassword := catalog.BinData.Password
	if catalog.BinData.Password != "" {
		decryptedPassword, err := utils.DecryptPassword(catalog.BinData.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password for %s: %w", input.DatasourceID, err)
		}
		decodedPassword = decryptedPassword
	}

	// 从数据源信息构建连接信息
	return &DBConn{
		Host:     catalog.BinData.Host,
		Port:     catalog.BinData.Port,
		Username: catalog.BinData.Account,
		Password: decodedPassword,
		Database: catalog.BinData.DatabaseName,
		Schema:   catalog.BinData.Schema,
	}, nil
}

// cleanJSONWhitespace 清理JSON字符串中的非标准空白字符
func cleanJSONWhitespace(jsonStr string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\u00a0', '\u2007', '\u202f': // 全角空格和其他非标准空白字符
			return ' ' // 替换为普通空格
		case '\u200b', '\u200c', '\u200d', '\u200e', '\u200f': // 零宽度字符
			return -1 // 删除
		case '\u2028', '\u2029': // 行分隔符和段落分隔符
			return '\n' // 替换为换行符
		default:
			return r
		}
	}, jsonStr)
}

// parseData 解析数据内容，支持变量引用
func (a *DatabaseWrite) parseData(ctx entity.ExecuteContext, data interface{}) (interface{}, error) {
	if data == nil {
		return nil, nil
	}
	if dataStr, ok := data.(string); ok {
		if strings.HasPrefix(dataStr, "$") {
			varName := strings.TrimPrefix(dataStr, "$")
			if value, exists := ctx.ShareData().Get(varName); exists {
				return value, nil
			}
			return nil, fmt.Errorf("variable %s not found", varName)
		} else {
			var jsonVal interface{}
			// 清理JSON字符串中的非标准空白字符
			cleanDataStr := cleanJSONWhitespace(dataStr)

			err := json.Unmarshal([]byte(cleanDataStr), &jsonVal)
			if err == nil {
				return jsonVal, nil
			}
			return dataStr, nil
		}
	}
	if dataMap, ok := data.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		for key, value := range dataMap {
			parsedValue, err := a.parseData(ctx, value)
			if err != nil {
				return nil, err
			}
			result[key] = parsedValue
		}
		return result, nil
	}
	if dataSlice, ok := data.([]interface{}); ok {
		result := make([]interface{}, len(dataSlice))
		for i, value := range dataSlice {
			parsedValue, err := a.parseData(ctx, value)
			if err != nil {
				return nil, err
			}
			result[i] = parsedValue
		}
		return result, nil
	}
	return data, nil
}

// transformDataByMapping 根据 sync_model_fields 将 data 中的源字段映射为目标字段
func (a *DatabaseWrite) transformDataByMapping(data interface{}, mappings []FieldMapping) ([]map[string]interface{}, error) {
	if data == nil {
		return []map[string]interface{}{}, nil
	}

	if len(mappings) == 0 {
		return a.convertToMapSlice(data)
	}

	// 构建映射表: source.name -> target.name
	srcToTgt := make(map[string]string, len(mappings))
	for _, m := range mappings {
		if m.Source.Name == "" || m.Target.Name == "" {
			continue
		}
		srcToTgt[m.Source.Name] = m.Target.Name
	}

	// 单对象
	if row, ok := data.(map[string]interface{}); ok {
		mappedRow := a.mapOneRow(row, srcToTgt, mappings)
		return []map[string]interface{}{mappedRow}, nil
	}

	// 对象数组（[]interface{}，每项为 map[string]interface{}）
	if arr, ok := data.([]interface{}); ok {
		out := make([]map[string]interface{}, 0, len(arr))
		for _, it := range arr {
			row, ok := it.(map[string]interface{})
			if !ok {
				continue
			}
			mappedRow := a.mapOneRow(row, srcToTgt, mappings)
			out = append(out, mappedRow)
		}
		return out, nil
	}

	// 其他类型，尝试转换
	return a.convertToMapSlice(data)
}

// mapOneRow 将一行根据映射生成新行（只输出目标字段）。未映射字段忽略。
func (a *DatabaseWrite) mapOneRow(row map[string]interface{}, srcToTgt map[string]string, mappings []FieldMapping) map[string]interface{} {
	out := make(map[string]interface{}, len(srcToTgt))

	// 创建目标字段名到字段属性的映射
	tgtToField := make(map[string]FieldAttr, len(mappings))
	for _, mapping := range mappings {
		if mapping.Target.Name != "" {
			tgtToField[mapping.Target.Name] = mapping.Target
		}
	}

	for src, tgt := range srcToTgt {
		if v, ok := row[src]; ok {
			// 应用字段级别的精度设置
			if field, exists := tgtToField[tgt]; exists {
				out[tgt] = a.applyFieldPrecision(v, field)
			} else {
				out[tgt] = v
			}
		}
	}
	return out
}

// convertToMapSlice 将任意数据转换为[]map[string]interface{}类型
func (a *DatabaseWrite) convertToMapSlice(data interface{}) ([]map[string]interface{}, error) {
	if data == nil {
		return []map[string]interface{}{}, nil
	}

	// 如果已经是[]map[string]interface{}类型
	if result, ok := data.([]map[string]interface{}); ok {
		return result, nil
	}

	// 如果是单个map[string]interface{}
	if singleMap, ok := data.(map[string]interface{}); ok {
		return []map[string]interface{}{singleMap}, nil
	}

	// 如果是[]interface{}，尝试转换每个元素
	if arr, ok := data.([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(arr))
		for _, item := range arr {
			if itemMap, ok := item.(map[string]interface{}); ok {
				result = append(result, itemMap)
			}
		}
		if len(result) > 0 {
			return result, nil
		}
	}

	// 无法转换
	return nil, fmt.Errorf("cannot convert data to []map[string]interface{}, got %T", data)
}

// fieldTypeInfo 字段类型信息，包含数据类型和精度信息
type fieldTypeInfo struct {
	DataType   string
	Precision  int // DECIMAL类型的小数位数
	DataLength int // DECIMAL类型的总位数(scale)
}

// inferFieldMappings 从数据中自动推断字段映射
func (a *DatabaseWrite) inferFieldMappings(data interface{}) ([]FieldMapping, error) {
	if data == nil {
		return []FieldMapping{}, nil
	}

	// 提取所有字段名和类型信息
	fieldTypes := make(map[string]*fieldTypeInfo)

	// 处理单个对象
	if row, ok := data.(map[string]interface{}); ok {
		a.extractFieldTypes(row, fieldTypes)
	} else if arr, ok := data.([]interface{}); ok {
		// 处理对象数组
		for _, item := range arr {
			if row, ok := item.(map[string]interface{}); ok {
				a.extractFieldTypes(row, fieldTypes)
			}
		}
	} else if arr, ok := data.([]map[string]interface{}); ok {
		// 处理map数组
		for _, row := range arr {
			a.extractFieldTypes(row, fieldTypes)
		}
	}

	// 创建字段映射
	var mappings []FieldMapping
	for fieldName, typeInfo := range fieldTypes {
		mapping := FieldMapping{
			Source: FieldAttr{
				Name:     fieldName,
				DataType: typeInfo.DataType,
			},
			Target: FieldAttr{
				Name:      fieldName,
				DataType:  typeInfo.DataType,
				Precision: typeInfo.Precision,
				DataLenth: typeInfo.DataLength,
			},
		}
		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

// extractFieldTypes 从单个数据行中提取字段类型信息
func (a *DatabaseWrite) extractFieldTypes(row map[string]interface{}, fieldTypes map[string]*fieldTypeInfo) {
	for fieldName, value := range row {
		if value == nil {
			continue
		}

		// 根据值的类型推断数据类型和精度
		typeInfo := a.inferDataTypeWithPrecision(value)

		// 如果字段已存在，需要合并精度信息（取最大值）
		if existing, exists := fieldTypes[fieldName]; exists {
			// 如果类型不一致，需要统一类型
			if existing.DataType != typeInfo.DataType {
				// 如果一个是DECIMAL一个是BIGINT，统一为DECIMAL（DECIMAL可以存储整数）
				if (typeInfo.DataType == "DECIMAL" && existing.DataType == "BIGINT") ||
					(typeInfo.DataType == "BIGINT" && existing.DataType == "DECIMAL") {
					existing.DataType = "DECIMAL"
					// 如果原来是BIGINT（Precision和DataLength都是0），需要根据当前值设置精度
					// 如果当前值是DECIMAL，使用当前值的精度；如果当前值是BIGINT，需要重新计算
					if existing.Precision == 0 && existing.DataLength == 0 {
						if typeInfo.DataType == "DECIMAL" {
							// 当前值是DECIMAL，直接使用其精度
							existing.Precision = typeInfo.Precision
							existing.DataLength = typeInfo.DataLength
						} else {
							// 当前值是BIGINT，但typeInfo也是BIGINT，需要根据实际值计算
							// 由于BIGINT的typeInfo中Precision和DataLength都是0，我们需要根据value重新计算
							if floatVal, ok := value.(float64); ok {
								precision, scale := a.calculateDecimalPrecision(floatVal)
								existing.Precision = precision
								existing.DataLength = scale
							} else if floatVal, ok := value.(float32); ok {
								precision, scale := a.calculateDecimalPrecision(float64(floatVal))
								existing.Precision = precision
								existing.DataLength = scale
							}
						}
					}
				} else {
					// 其他类型不一致的情况，保持原有类型
					continue
				}
			}
			// 合并精度信息：取最大值
			if existing.DataType == "DECIMAL" {
				if typeInfo.Precision > existing.Precision {
					existing.Precision = typeInfo.Precision
				}
				if typeInfo.DataLength > existing.DataLength {
					existing.DataLength = typeInfo.DataLength
				}
			}
		} else {
			fieldTypes[fieldName] = typeInfo
		}
	}
}

// inferDataTypeWithPrecision 根据Go类型推断数据库数据类型和精度信息
func (a *DatabaseWrite) inferDataTypeWithPrecision(value interface{}) *fieldTypeInfo {
	switch v := value.(type) {
	case bool:
		return &fieldTypeInfo{DataType: "BOOLEAN"}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return &fieldTypeInfo{DataType: "BIGINT"}
	case float32:
		// 检查是否是整数（使用更精确的方法）
		if a.isInteger(float64(v)) {
			return &fieldTypeInfo{DataType: "BIGINT"}
		}
		// 计算精度和标度
		precision, scale := a.calculateDecimalPrecision(float64(v))
		return &fieldTypeInfo{
			DataType:   "DECIMAL",
			Precision:  precision,
			DataLength: scale,
		}
	case float64:
		// 检查是否是整数（使用更精确的方法）
		if a.isInteger(v) {
			return &fieldTypeInfo{DataType: "BIGINT"}
		}
		// 计算精度和标度
		precision, scale := a.calculateDecimalPrecision(v)
		return &fieldTypeInfo{
			DataType:   "DECIMAL",
			Precision:  precision,
			DataLength: scale,
		}
	case string:
		// 根据字符串长度判断使用VARCHAR还是TEXT
		if len(v) > 255 {
			return &fieldTypeInfo{DataType: "TEXT"}
		}
		return &fieldTypeInfo{DataType: "VARCHAR(255)"}
	case []byte:
		return &fieldTypeInfo{DataType: "BLOB"}
	default:
		// 对于复杂类型或其他未知类型，使用TEXT存储JSON序列化结果
		return &fieldTypeInfo{DataType: "TEXT"}
	}
}

// isInteger 检查float64值是否表示一个整数
// 使用字符串比较来避免float64精度问题
func (a *DatabaseWrite) isInteger(value float64) bool {
	// 使用格式化字符串来避免精度损失
	str := fmt.Sprintf("%.0f", value)

	// 尝试解析为int64（适用于大多数情况）
	if intVal, err := strconv.ParseInt(str, 10, 64); err == nil {
		return float64(intVal) == value
	}

	// 如果int64范围不够，尝试解析为uint64
	if uintVal, err := strconv.ParseUint(str, 10, 64); err == nil {
		return float64(uintVal) == value
	}

	// 如果都失败，使用字符串方法：检查是否有小数点和小数部分
	// 将float转换为字符串，检查是否包含非零小数部分
	strWithDecimal := fmt.Sprintf("%.20f", value)
	strWithDecimal = strings.TrimRight(strWithDecimal, "0")
	strWithDecimal = strings.TrimRight(strWithDecimal, ".")

	// 如果去除尾部0后没有小数点，说明是整数
	return !strings.Contains(strWithDecimal, ".")
}

// calculateDecimalPrecision 计算DECIMAL类型的精度和标度
// 返回 (precision小数位数, scale总位数)
func (a *DatabaseWrite) calculateDecimalPrecision(value float64) (int, int) {
	// 将float转换为字符串来解析精度，使用足够的小数位数
	str := fmt.Sprintf("%.20f", value)

	// 移除尾部的0和小数点
	str = strings.TrimRight(str, "0")
	str = strings.TrimRight(str, ".")

	// 分离整数部分和小数部分
	parts := strings.Split(str, ".")
	integerPart := parts[0]

	// 处理空字符串或只有负号的情况
	if integerPart == "" || integerPart == "-" {
		integerPart = "0"
	}

	// 计算整数部分位数
	integerDigits := len(integerPart)
	if len(integerPart) > 0 && integerPart[0] == '-' {
		integerDigits-- // 减去负号
	}
	// 如果整数部分是0，至少算1位
	if integerDigits == 0 {
		integerDigits = 1
	}

	// 计算小数部分位数
	decimalDigits := 0
	if len(parts) > 1 {
		decimalDigits = len(parts[1])
	}

	// 精度（小数位数）- 对于整数，precision应该为0，而不是默认值
	precision := decimalDigits
	// 只有当确实有小数部分时，才设置最小精度
	if precision == 0 {
		precision = 0 // 整数的小数位数为0
	} else if precision < writer.DefaultDecimalPrecision {
		precision = writer.DefaultDecimalPrecision
	}
	// 限制最大精度为38（大多数数据库的最大精度）
	if precision > 38 {
		precision = 38
	}

	// 标度（总位数）= 整数位数 + 小数位数
	scale := integerDigits + precision
	// 对于整数，scale应该至少等于整数位数，但不应该被限制为DefaultDecimalScale
	// 只有当scale小于整数位数时才调整
	if scale < integerDigits {
		scale = integerDigits
	}
	// 限制最大标度为38
	if scale > 38 {
		scale = 38
	}

	return precision, scale
}

// inferDataType 根据Go类型推断数据库数据类型（保留用于向后兼容）
func (a *DatabaseWrite) inferDataType(value interface{}) string {
	typeInfo := a.inferDataTypeWithPrecision(value)
	return typeInfo.DataType
}

// applyFieldPrecision 根据字段映射应用数据精度设置
func (a *DatabaseWrite) applyFieldPrecision(value interface{}, field FieldAttr) interface{} {
	// 如果是DECIMAL类型，使用字段级别的precision设置
	if strings.EqualFold(field.DataType, "DECIMAL") && field.Precision >= 0 {
		switch v := value.(type) {
		case float64:
			// 对float64类型应用小数位数精度
			multiplier := writer.DecimalBase
			for i := 0; i < field.Precision; i++ {
				multiplier *= 10
			}
			return float64(int64(v*float64(multiplier)+0.5)) / float64(multiplier)
		case float32:
			// 对float32类型应用小数位数精度
			multiplier := writer.DecimalBase
			for i := 0; i < field.Precision; i++ {
				multiplier *= 10
			}
			return float32(int64(float64(v)*float64(multiplier)+0.5)) / float32(multiplier)
		}
	}

	// 对于其他类型，保持原值
	return value
}

// 实现Action接口
var _ entity.Action = (*DatabaseWrite)(nil)
