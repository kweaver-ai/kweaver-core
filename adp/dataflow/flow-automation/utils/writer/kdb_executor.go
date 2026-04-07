package writer

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// KDBExecutor KDB执行器实现
type KDBExecutor struct{}

func (e *KDBExecutor) ExecuteInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
	return e.executeKDBInsert(ctx, dbConn, tableInfo, driver, data)
}

func (e *KDBExecutor) ExecuteUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
	return e.executeKDBUpdate(ctx, dbConn, tableInfo, driver, data, where)
}

func (e *KDBExecutor) ExecuteDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
	return e.executeKDBDelete(ctx, dbConn, tableInfo, driver, where)
}

// executeKDBInsert 执行KDB插入操作
func (e *KDBExecutor) executeKDBInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
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

// executeKDBUpdate 执行KDB更新操作
func (e *KDBExecutor) executeKDBUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
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
		return nil, fmt.Errorf("KDB update failed: %w", result.Error)
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

// executeKDBDelete 执行KDB删除操作
func (e *KDBExecutor) executeKDBDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
	fullTableName := driver.GetFullTableName(tableInfo)

	if where == nil {
		return nil, fmt.Errorf("where condition is required for delete operation")
	}

	result := dbConn.Table(fullTableName).Where(where).Delete("")
	if result.Error != nil {
		return nil, fmt.Errorf("KDB delete failed: %w", result.Error)
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

// executeBatchInsertWithDetails 执行批量插入并返回详细的失败信息
func (e *KDBExecutor) executeBatchInsertWithDetails(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, batchSize int) (int64, []map[string]interface{}, map[string]int) {
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

// executeIndividualInserts 执行单个记录插入，返回详细的失败信息
func (e *KDBExecutor) executeIndividualInserts(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, failureReasons map[string]int) (int64, []map[string]interface{}, map[string]int) {
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

// categorizeInsertError 分类插入错误类型 (KDB版本)
func (e *KDBExecutor) categorizeInsertError(err error) string {
	errStr := err.Error()

	// KDB 错误分类（基于常见的数据库错误模式）
	if strings.Contains(errStr, "Duplicate entry") ||
		strings.Contains(errStr, "duplicate key value") ||
		strings.Contains(errStr, "UNIQUE constraint failed") ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "唯一键约束") {
		return "duplicate_key"
	}

	if strings.Contains(errStr, "foreign key constraint fails") ||
		strings.Contains(errStr, "foreign key") ||
		strings.Contains(errStr, "外键约束") {
		return "foreign_key_constraint"
	}

	if strings.Contains(errStr, "cannot be null") ||
		strings.Contains(errStr, "Column") ||
		strings.Contains(errStr, "NOT NULL") ||
		strings.Contains(errStr, "不能为空") {
		return "null_constraint"
	}

	if strings.Contains(errStr, "data too long") ||
		strings.Contains(errStr, "数据过长") {
		return "data_too_long"
	}

	if strings.Contains(errStr, "column") &&
		strings.Contains(errStr, "does not exist") ||
		strings.Contains(errStr, "42703") || // PostgreSQL style error code
		strings.Contains(errStr, "字段不存在") ||
		strings.Contains(errStr, "列不存在") {
		return ErrorTypeFieldNotExist
	}

	if strings.Contains(errStr, "incorrect") ||
		strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "无效") {
		return "data_type_mismatch"
	}

	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline") ||
		strings.Contains(errStr, "超时") {
		return "timeout"
	}

	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "连接") {
		return "connection_error"
	}

	// 默认错误类型
	return "unknown_error"
}
