package writer

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// OracleExecutor Oracle执行器实现
//
// 注意: 当前实现仅提供Oracle数据库的执行逻辑框架
// 实际执行需要与Oracle特定的驱动程序配合使用
type OracleExecutor struct{}

func (e *OracleExecutor) ExecuteInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
	return e.executeOracleInsert(ctx, dbConn, tableInfo, driver, data)
}

func (e *OracleExecutor) ExecuteUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
	return e.executeOracleUpdate(ctx, dbConn, tableInfo, driver, data, where)
}

func (e *OracleExecutor) ExecuteDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
	return e.executeOracleDelete(ctx, dbConn, tableInfo, driver, where)
}

// executeOracleInsert 执行Oracle插入操作
func (e *OracleExecutor) executeOracleInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
	fullTableName := driver.GetFullTableName(tableInfo)

	// 检查表是否存在
	var count int64
	if err := dbConn.Table(fullTableName).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check table %s existence: %w", fullTableName, err)
	}

	batchSize := DefaultBatchSize
	if tableInfo.Options != nil && tableInfo.Options.BatchSize > 0 {
		batchSize = tableInfo.Options.BatchSize
	}

	// 使用批量插入
	successCount, failedRecords, failureReasons := e.executeBatchInsertWithDetails(ctx, dbConn, tableInfo, driver, data, batchSize)

	// 验证写入结果
	var newCount int64
	if err := dbConn.Table(fullTableName).Count(&newCount).Error; err != nil {
		return nil, fmt.Errorf("failed to verify write result: %w", err)
	}

	return &ExecutionResult{
		AffectedRows:   successCount,
		Operation:      OperationInsert,
		Table:          fullTableName,
		Success:        len(failedRecords) == 0,
		BeforeCount:    count,
		AfterCount:     newCount,
		SuccessCount:   successCount,
		FailedCount:    int64(len(failedRecords)),
		TotalProcessed: int64(len(data)),
		FailedRecords:  failedRecords,
		FailureReasons: failureReasons,
	}, nil
}

// executeOracleUpdate 执行Oracle更新操作
func (e *OracleExecutor) executeOracleUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
	fullTableName := driver.GetFullTableName(tableInfo)

	if where == nil {
		return nil, fmt.Errorf("where condition is required for update operation")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty for update operation")
	}
	updateData := data[0]

	result := dbConn.Table(fullTableName).Where(where).Updates(updateData)
	if result.Error != nil {
		return nil, fmt.Errorf("Oracle update failed: %w", result.Error)
	}

	return &ExecutionResult{
		AffectedRows:   result.RowsAffected,
		Operation:      OperationUpdate,
		Table:          fullTableName,
		Success:        true,
		SuccessCount:   result.RowsAffected,
		FailedCount:    0,
		TotalProcessed: 1,
		FailedRecords:  []map[string]interface{}{},
		FailureReasons: map[string]int{},
	}, nil
}

// executeOracleDelete 执行Oracle删除操作
func (e *OracleExecutor) executeOracleDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
	fullTableName := driver.GetFullTableName(tableInfo)

	if where == nil {
		return nil, fmt.Errorf("where condition is required for delete operation")
	}

	result := dbConn.Table(fullTableName).Where(where).Delete("")
	if result.Error != nil {
		return nil, fmt.Errorf("Oracle delete failed: %w", result.Error)
	}

	return &ExecutionResult{
		AffectedRows:   result.RowsAffected,
		Operation:      OperationDelete,
		Table:          fullTableName,
		Success:        true,
		SuccessCount:   result.RowsAffected,
		FailedCount:    0,
		TotalProcessed: 1,
		FailedRecords:  []map[string]interface{}{},
		FailureReasons: map[string]int{},
	}, nil
}

// executeBatchInsertWithDetails 执行批量插入并返回详细的失败信息 - Oracle版本
func (e *OracleExecutor) executeBatchInsertWithDetails(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, batchSize int) (int64, []map[string]interface{}, map[string]int) {
	var successCount int64
	var failedRecords []map[string]interface{}
	failureReasons := make(map[string]int)

	// 如果批量大小为1，使用单个插入模式以获取更详细的错误信息
	if batchSize == 1 || len(data) <= MinBatchDataThreshold {
		return e.executeIndividualInserts(ctx, dbConn, tableInfo, driver, data, failureReasons)
	}

	fullTableName := driver.GetFullTableName(tableInfo)

	// 预处理批量数据
	processedData := make([]map[string]interface{}, len(data))
	for i, record := range data {
		processedData[i] = e.preprocessOracleData(record, tableInfo)
	}

	// 尝试批量插入（由于包含特殊表达式，我们需要使用自定义的批量插入）
	result := e.executeBatchRawInsert(dbConn, fullTableName, processedData)
	if result.Error != nil {
		// 如果批量插入失败，回退到单个插入模式
		return e.executeIndividualInserts(ctx, dbConn, tableInfo, driver, data, failureReasons)
	}

	successCount = result.RowsAffected

	// 检查是否有部分失败的情况（通过比较预期插入数量和实际影响行数）
	expectedInserts := int64(len(data))
	if successCount < expectedInserts {
		failedCount := expectedInserts - successCount
		failureReasons["partial_batch_failure"] = int(failedCount)
	}

	return successCount, failedRecords, failureReasons
}

// executeIndividualInserts 执行单个记录插入，返回详细的失败信息 - Oracle版本
func (e *OracleExecutor) executeIndividualInserts(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, failureReasons map[string]int) (int64, []map[string]interface{}, map[string]int) {
	var successCount int64
	var failedRecords []map[string]interface{}

	fullTableName := driver.GetFullTableName(tableInfo)

	for i, record := range data {
		// 预处理数据，转换日期和时间戳格式
		processedRecord := e.preprocessOracleData(record, tableInfo)

		// 创建新的事务用于单个记录插入
		tx := dbConn.Begin()
		if tx.Error != nil {
			failureReasons["transaction_start_failed"]++
			failedRecords = append(failedRecords, map[string]interface{}{
				"index":   i,
				"record":  record,
				"reason":  "transaction_start_failed",
				"details": tx.Error.Error(),
			})
			continue
		}

		// 尝试插入单个记录
		result := e.executeRawInsert(tx, fullTableName, processedRecord)
		if result.Error != nil {
			tx.Rollback()

			// 分类错误类型
			errorType := e.categorizeInsertError(result.Error)
			failureReasons[errorType]++

			failedRecords = append(failedRecords, map[string]interface{}{
				"index":   i,
				"record":  record,
				"reason":  errorType,
				"details": result.Error.Error(),
			})

		} else {
			// 提交事务
			if err := tx.Commit().Error; err != nil {
				failureReasons["transaction_commit_failed"]++
				failedRecords = append(failedRecords, map[string]interface{}{
					"index":   i,
					"record":  record,
					"reason":  "transaction_commit_failed",
					"details": err.Error(),
				})
			} else {
				successCount++
			}
		}
	}

	return successCount, failedRecords, failureReasons
}

// preprocessOracleData 预处理 Oracle 数据，转换日期和时间戳格式，并验证数值范围
func (e *OracleExecutor) preprocessOracleData(record map[string]interface{}, tableInfo *TableInfo) map[string]interface{} {
	// 创建记录的副本以避免修改原始数据
	processed := make(map[string]interface{})
	for k, v := range record {
		processed[k] = v
	}

	// 如果没有字段信息，无法进行类型判断，直接返回
	if tableInfo.Fields == nil {
		return processed
	}

	// 遍历字段，根据字段类型处理数据
	for _, field := range tableInfo.Fields {
		fieldName := field.Target.Name
		if value, exists := processed[fieldName]; exists && value != nil {
			dataType := strings.ToLower(field.Target.DataType)

			// 处理日期和时间戳类型
			if strValue, ok := value.(string); ok && strValue != "" {
				// 检查是否为日期时间类型
				if dataType == "date" ||
					strings.Contains(dataType, "timestamp") ||
					strings.Contains(dataType, "timestamptz") ||
					strings.Contains(dataType, "time zone") {
					processed[fieldName] = e.convertDateTimeForOracle(strValue, dataType)
				}
			}

			// 处理数值类型，验证和截断超出范围的数值
			if e.isNumericType(dataType) {
				processed[fieldName] = e.validateAndClampNumericValue(value, dataType)
			}
		}
	}

	return processed
}

// OracleSQLExpr 表示需要作为原始 SQL 插入的表达式
type OracleSQLExpr string

// convertDateTimeForOracle 将日期时间字符串转换为 Oracle 兼容的格式
func (e *OracleExecutor) convertDateTimeForOracle(value string, dataType string) interface{} {
	// 首先检查是否为时区时间戳格式
	if e.isStandardTimestampTZFormat(value) {
		return OracleSQLExpr(fmt.Sprintf("TO_TIMESTAMP_TZ('%s', 'YYYY-MM-DD HH24:MI:SS.FF6TZH:TZM')", e.escapeSQLString(value)))
	}

	switch dataType {
	case "date":
		// 对于 DATE 类型，使用 TO_DATE 函数
		if e.isStandardDateFormat(value) {
			return OracleSQLExpr(fmt.Sprintf("TO_DATE('%s', 'YYYY-MM-DD')", e.escapeSQLString(value)))
		}
	case "timestamp", "timestamp(3)":
		// 对于 TIMESTAMP 类型，使用 TO_TIMESTAMP 函数
		if e.isStandardTimestampFormat(value) {
			return OracleSQLExpr(fmt.Sprintf("TO_TIMESTAMP('%s', 'YYYY-MM-DD HH24:MI:SS.FF6')", e.escapeSQLString(value)))
		}
	case "timestamp(3) with time zone", "timestamptz", "timestamp with time zone":
		// 对于 TIMESTAMP WITH TIME ZONE 类型，使用 TO_TIMESTAMP_TZ 函数
		return OracleSQLExpr(fmt.Sprintf("TO_TIMESTAMP_TZ('%s', 'YYYY-MM-DD HH24:MI:SS.FF6TZH:TZM')", e.escapeSQLString(value)))
	}

	// 如果不匹配任何已知格式，返回原始值
	return value
}

// escapeSQLString 转义 SQL 字符串中的单引号
func (e *OracleExecutor) escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// executeRawInsert 执行原始 SQL 插入
func (e *OracleExecutor) executeRawInsert(tx *gorm.DB, tableName string, record map[string]interface{}) *gorm.DB {
	// 构建 INSERT 语句
	var columns []string
	var placeholders []string
	var args []interface{}

	// 为了确保顺序一致，我们需要按键排序
	var keys []string
	for key := range record {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 按排序后的键处理记录
	for _, key := range keys {
		value := record[key]
		columns = append(columns, fmt.Sprintf("\"%s\"", key))

		// 处理 OracleSQLExpr 类型的值
		if expr, ok := value.(OracleSQLExpr); ok {
			placeholders = append(placeholders, string(expr))
		} else {
			// 对于普通值，使用参数占位符
			placeholders = append(placeholders, "?")
			args = append(args, value)
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// 执行 SQL
	return tx.Exec(sql, args...)
}

// executeBatchRawInsert 执行批量原始 SQL 插入
func (e *OracleExecutor) executeBatchRawInsert(db *gorm.DB, tableName string, records []map[string]interface{}) *gorm.DB {
	if len(records) == 0 {
		return db
	}

	// 收集所有记录中出现的字段
	fieldSet := make(map[string]bool)
	for _, record := range records {
		for key := range record {
			fieldSet[key] = true
		}
	}

	// 对字段进行排序
	var keys []string
	for key := range fieldSet {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// 构建批量 INSERT 语句
	var columns []string
	for _, key := range keys {
		columns = append(columns, fmt.Sprintf("\"%s\"", key))
	}

	var valueGroups []string
	var allArgs []interface{}

	for _, record := range records {
		var placeholders []string

		for _, key := range keys {
			value, exists := record[key]

			// 处理 OracleSQLExpr 类型的值
			if expr, ok := value.(OracleSQLExpr); ok && exists {
				placeholders = append(placeholders, string(expr))
			} else if exists {
				// 对于普通值，使用参数占位符
				placeholders = append(placeholders, "?")
				allArgs = append(allArgs, value)
			} else {
				// 如果字段不存在，使用 NULL
				placeholders = append(placeholders, "NULL")
			}
		}

		valueGroups = append(valueGroups, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(valueGroups, ", "))

	// 执行批量 SQL
	return db.Exec(sql, allArgs...)
}

// isNumericType 检查是否为数值类型
func (e *OracleExecutor) isNumericType(dataType string) bool {
	switch strings.ToLower(dataType) {
	case "number", "binary_float", "binary_double":
		return true
	case "tinyint", "smallint", "int", "integer", "bigint", "long":
		return true
	case "float", "real", "double":
		return true
	case "decimal", "numeric":
		return true
	default:
		return false
	}
}

// validateAndClampNumericValue 验证和截断超出范围的数值
func (e *OracleExecutor) validateAndClampNumericValue(value interface{}, dataType string) interface{} {
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		return e.clampIntegerValue(v, dataType)
	case uint, uint8, uint16, uint32, uint64:
		return e.clampUnsignedValue(v, dataType)
	case float32, float64:
		return e.clampFloatValue(v, dataType)
	case string:
		// 尝试将字符串转换为数值
		if numValue, err := e.parseNumericString(v); err == nil {
			return numValue
		}
		// 如果转换失败，返回原值
		return value
	default:
		// 对于其他类型，返回原值
		return value
	}
}

// clampIntegerValue 截断整数值到 Oracle 允许的范围内
func (e *OracleExecutor) clampIntegerValue(value interface{}, dataType string) interface{} {
	// Oracle NUMBER 类型支持的最大值和最小值
	const maxOracleNumber = 99999999999999999999999999999999999999.0 // 38 位 9
	const minOracleNumber = -99999999999999999999999999999999999999.0

	var floatVal float64
	switch v := value.(type) {
	case int:
		floatVal = float64(v)
	case int8:
		floatVal = float64(v)
	case int16:
		floatVal = float64(v)
	case int32:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	default:
		return value
	}

	// 截断到 Oracle NUMBER 范围
	if floatVal > maxOracleNumber {
		return maxOracleNumber
	}
	if floatVal < minOracleNumber {
		return minOracleNumber
	}

	return floatVal
}

// clampUnsignedValue 截断无符号数值到 Oracle 允许的范围内
func (e *OracleExecutor) clampUnsignedValue(value interface{}, dataType string) interface{} {
	const maxOracleNumber = 99999999999999999999999999999999999999.0

	var floatVal float64
	switch v := value.(type) {
	case uint:
		floatVal = float64(v)
	case uint8:
		floatVal = float64(v)
	case uint16:
		floatVal = float64(v)
	case uint32:
		floatVal = float64(v)
	case uint64:
		floatVal = float64(v)
	default:
		return value
	}

	// 截断到 Oracle NUMBER 范围
	if floatVal > maxOracleNumber {
		return maxOracleNumber
	}

	return floatVal
}

// clampFloatValue 截断浮点数值到 Oracle 允许的范围内
func (e *OracleExecutor) clampFloatValue(value interface{}, dataType string) interface{} {
	const maxOracleNumber = 99999999999999999999999999999999999999.0
	const minOracleNumber = -99999999999999999999999999999999999999.0

	var floatVal float64
	switch v := value.(type) {
	case float32:
		floatVal = float64(v)
	case float64:
		floatVal = v
	default:
		return value
	}

	// 检查是否为特殊浮点值
	if floatVal != floatVal { // NaN
		return 0.0
	}
	if floatVal > maxOracleNumber {
		floatVal = maxOracleNumber
	}
	if floatVal < minOracleNumber {
		floatVal = minOracleNumber
	}

	return floatVal
}

// parseNumericString 尝试将字符串解析为数值
func (e *OracleExecutor) parseNumericString(s string) (interface{}, error) {
	// 尝试解析为整数
	if intVal, err := strconv.ParseInt(s, 10, 64); err == nil {
		return e.clampIntegerValue(intVal, "bigint"), nil
	}

	// 尝试解析为浮点数
	if floatVal, err := strconv.ParseFloat(s, 64); err == nil {
		return e.clampFloatValue(floatVal, "double"), nil
	}

	return nil, fmt.Errorf("cannot parse as number: %s", s)
}

// isStandardDateFormat 检查是否为标准日期格式 (YYYY-MM-DD)
func (e *OracleExecutor) isStandardDateFormat(value string) bool {
	// 简单的格式检查 YYYY-MM-DD
	parts := strings.Split(value, "-")
	if len(parts) != 3 {
		return false
	}
	if len(parts[0]) != 4 || len(parts[1]) != 2 || len(parts[2]) != 2 {
		return false
	}
	return true
}

// isStandardTimestampFormat 检查是否为标准时间戳格式 (YYYY-MM-DD HH:MM:SS.FFFFFF)
func (e *OracleExecutor) isStandardTimestampFormat(value string) bool {
	// 检查是否包含空格和冒号
	if !strings.Contains(value, " ") || !strings.Contains(value, ":") {
		return false
	}

	parts := strings.Split(value, " ")
	if len(parts) != 2 {
		return false
	}

	// 检查日期部分
	datePart := parts[0]
	if !e.isStandardDateFormat(datePart) {
		return false
	}

	// 检查时间部分 (HH:MM:SS.FFFFFF)
	timePart := parts[1]
	timeParts := strings.Split(timePart, ":")
	if len(timeParts) < 2 || len(timeParts) > 3 {
		return false
	}

	return true
}

// isStandardTimestampTZFormat 检查是否为标准带时区时间戳格式
func (e *OracleExecutor) isStandardTimestampTZFormat(value string) bool {
	// 检查是否包含 + 或 - 时区信息
	if !strings.Contains(value, "+") && !strings.Contains(value, "-") {
		return false
	}

	// 找到时区分隔符的位置（+ 或 -）
	tzIndex := strings.LastIndex(value, "+")
	if tzIndex == -1 {
		tzIndex = strings.LastIndex(value, "-")
	}

	// 提取时区信息之前的部分
	timestampPart := value[:tzIndex]

	// 检查是否是标准时间戳格式
	return e.isStandardTimestampFormat(timestampPart)
}

// categorizeInsertError 分类插入错误类型 (Oracle版本)
func (e *OracleExecutor) categorizeInsertError(err error) string {
	errStr := strings.ToLower(err.Error())

	// Oracle错误分类
	if strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "duplicate value") ||
		strings.Contains(errStr, "ora-00001") { // Oracle duplicate key error code
		return "duplicate_key"
	}

	if strings.Contains(errStr, "foreign key") ||
		strings.Contains(errStr, "integrity constraint") ||
		strings.Contains(errStr, "ora-02291") || // Oracle foreign key constraint error code
		strings.Contains(errStr, "ora-02292") { // Oracle child record found error code
		return "foreign_key_constraint"
	}

	if strings.Contains(errStr, "cannot insert null") ||
		strings.Contains(errStr, "null value") ||
		strings.Contains(errStr, "ora-01400") { // Oracle cannot insert null error code
		return "null_constraint"
	}

	if strings.Contains(errStr, "value too large") ||
		strings.Contains(errStr, "ora-12899") { // Oracle value too large error code
		return "data_too_long"
	}

	if strings.Contains(errStr, "invalid identifier") ||
		strings.Contains(errStr, "ora-00904") { // Oracle invalid identifier error code
		return ErrorTypeFieldNotExist
	}

	if strings.Contains(errStr, "invalid number") ||
		strings.Contains(errStr, "invalid character") ||
		strings.Contains(errStr, "ora-01722") || // Oracle invalid number error code
		strings.Contains(errStr, "ora-01861") { // Oracle literal does not match format string error code
		return "data_type_mismatch"
	}

	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") {
		return "connection_error"
	}

	// 默认错误类型
	return "unknown_error"
}
