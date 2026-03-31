package writer

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// SQLServerExecutor SQL Server执行器实现
type SQLServerExecutor struct{}

func (e *SQLServerExecutor) ExecuteInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
	return e.executeSQLServerInsert(ctx, dbConn, tableInfo, driver, data)
}

func (e *SQLServerExecutor) ExecuteUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
	return e.executeSQLServerUpdate(ctx, dbConn, tableInfo, driver, data, where)
}

func (e *SQLServerExecutor) ExecuteDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
	return e.executeSQLServerDelete(ctx, dbConn, tableInfo, driver, where)
}

// executeSQLServerInsert 执行SQL Server插入操作
func (e *SQLServerExecutor) executeSQLServerInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
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

// executeSQLServerUpdate 执行SQL Server更新操作
func (e *SQLServerExecutor) executeSQLServerUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
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
		return nil, fmt.Errorf("SQL Server update failed: %w", result.Error)
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

// executeSQLServerDelete 执行SQL Server删除操作
func (e *SQLServerExecutor) executeSQLServerDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
	fullTableName := driver.GetFullTableName(tableInfo)

	if where == nil {
		return nil, fmt.Errorf("where condition is required for delete operation")
	}

	result := dbConn.Table(fullTableName).Where(where).Delete("")
	if result.Error != nil {
		return nil, fmt.Errorf("SQL Server delete failed: %w", result.Error)
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

// executeBatchInsertWithDetails 执行批量插入并返回详细的失败信息 - SQL Server版本
func (e *SQLServerExecutor) executeBatchInsertWithDetails(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, batchSize int) (int64, []map[string]interface{}, map[string]int) {
	var successCount int64
	var failedRecords []map[string]interface{}
	failureReasons := make(map[string]int)

	// 如果批量大小为1，使用单个插入模式以获取更详细的错误信息
	if batchSize == 1 || len(data) <= MinBatchDataThreshold {
		return e.executeIndividualInserts(ctx, dbConn, tableInfo, driver, data, failureReasons)
	}

	fullTableName := driver.GetFullTableName(tableInfo)

	// 尝试批量插入
	result := dbConn.Table(fullTableName).CreateInBatches(data, batchSize)
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

// executeIndividualInserts 执行单个记录插入，返回详细的失败信息 - SQL Server版本
func (e *SQLServerExecutor) executeIndividualInserts(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, failureReasons map[string]int) (int64, []map[string]interface{}, map[string]int) {
	var successCount int64
	var failedRecords []map[string]interface{}

	fullTableName := driver.GetFullTableName(tableInfo)

	for i, record := range data {
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
		result := tx.Table(fullTableName).Create(record)
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

// categorizeInsertError 分类插入错误类型 (SQL Server版本)
func (e *SQLServerExecutor) categorizeInsertError(err error) string {
	errStr := strings.ToLower(err.Error())

	// SQL Server错误分类
	if strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "cannot insert duplicate key") ||
		strings.Contains(errStr, "2627") { // SQL Server duplicate key error code
		return "duplicate_key"
	}

	if strings.Contains(errStr, "foreign key") ||
		strings.Contains(errStr, "reference constraint") ||
		strings.Contains(errStr, "547") { // SQL Server foreign key constraint error code
		return "foreign_key_constraint"
	}

	if strings.Contains(errStr, "cannot insert the value null") ||
		strings.Contains(errStr, "null value") ||
		strings.Contains(errStr, "515") { // SQL Server null constraint error code
		return "null_constraint"
	}

	if strings.Contains(errStr, "string or binary data would be truncated") ||
		strings.Contains(errStr, "8152") { // SQL Server data truncation error code
		return "data_too_long"
	}

	if strings.Contains(errStr, "invalid column name") ||
		strings.Contains(errStr, "207") { // SQL Server invalid column name error code
		return ErrorTypeFieldNotExist
	}

	if strings.Contains(errStr, "conversion failed") ||
		strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "245") || // SQL Server type conversion error code
		strings.Contains(errStr, "8114") { // SQL Server data type error code
		return "data_type_mismatch"
	}

	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "query timeout") {
		return "timeout"
	}

	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "communication") {
		return "connection_error"
	}

	// 默认错误类型
	return "unknown_error"
}
