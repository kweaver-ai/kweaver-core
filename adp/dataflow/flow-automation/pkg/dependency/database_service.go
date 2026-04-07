package dependency

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils/writer"
	"gorm.io/gorm"
)

type TableMetadata = writer.TableMetadata
type ColumnInfo = writer.ColumnInfo

// DatabaseTableService 数据库表查询服务接口
type DatabaseTableService interface {
	ListTables(ctx context.Context, dataSourceID, token, ipStr string) ([]TableMetadata, error)
	ListTableColumns(ctx context.Context, dataSourceID, tableName, schema, token, ipStr string) ([]ColumnInfo, error)
}

// DatabaseTableServiceImpl 数据库表查询服务实现
type DatabaseTableServiceImpl struct {
	dataSourceResolver drivenadapters.IDataSource
}

var (
	databaseTableServiceOnce sync.Once
	databaseTableService     DatabaseTableService
)

func NewDatabaseTableService() DatabaseTableService {
	databaseTableServiceOnce.Do(func() {
		databaseTableService = &DatabaseTableServiceImpl{
			dataSourceResolver: drivenadapters.NewDataSource(),
		}
	})

	return databaseTableService
}

// ListTables 列出数据库中的所有表
func (s *DatabaseTableServiceImpl) ListTables(ctx context.Context, dataSourceID, token, ipStr string) ([]TableMetadata, error) {
	// 获取数据源信息
	dataSourceInfo, err := s.dataSourceResolver.GetDataSourceCatalog(ctx, dataSourceID, token, ipStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve data source: %w", err)
	}

	if dataSourceInfo == nil {
		return nil, fmt.Errorf("data source info is nil")
	}

	// 获取数据库类型
	dbType := dataSourceInfo.Type
	if dbType == "" {
		return nil, fmt.Errorf("database type is empty")
	}

	// 对password进行RSA解密
	decodedPassword := dataSourceInfo.BinData.Password
	if dataSourceInfo.BinData.Password != "" {
		decryptedPassword, err := utils.DecryptPassword(dataSourceInfo.BinData.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password for %s: %w", dataSourceID, err)
		}
		decodedPassword = decryptedPassword
	}

	// 构建数据库配置
	config := &writer.DatabaseConfig{
		Host:     dataSourceInfo.BinData.Host,
		Port:     dataSourceInfo.BinData.Port,
		Username: dataSourceInfo.BinData.Account,
		Password: decodedPassword,
		Database: dataSourceInfo.BinData.DatabaseName,
		Driver:   dbType,
	}

	// 获取数据库连接
	dbConn, err := s.getDatabaseConnection(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}
	defer func() {
		if dbConn != nil {
			sqlDB, _ := dbConn.DB()
			if sqlDB != nil {
				_ = sqlDB.Close()
			}
		}
	}()

	// 获取对应的driver
	if !writer.IsDatabaseTypeSupported(dbType) {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	driver, _ := writer.GetDatabaseDriver(dbType)

	// 调用ListTables方法
	writerTables, err := driver.ListTables(dbConn, dataSourceInfo.BinData.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	// 转换类型
	tables := make([]TableMetadata, len(writerTables))
	for i, wt := range writerTables {
		tables[i] = TableMetadata{
			Name:    wt.Name,
			Schema:  wt.Schema,
			Type:    wt.Type,
			Comment: wt.Comment,
		}
	}

	return tables, nil
}

// ListTableColumns 查询指定表的字段信息
func (s *DatabaseTableServiceImpl) ListTableColumns(ctx context.Context, dataSourceID, tableName, schema, token, ipStr string) ([]ColumnInfo, error) {
	// 获取数据源信息
	dataSourceInfo, err := s.dataSourceResolver.GetDataSourceCatalog(ctx, dataSourceID, token, ipStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve data source: %w", err)
	}

	if dataSourceInfo == nil {
		return nil, fmt.Errorf("data source info is nil")
	}

	// 获取数据库类型
	dbType := dataSourceInfo.Type
	if dbType == "" {
		return nil, fmt.Errorf("database type is empty")
	}

	// 对password进行RSA解密
	decodedPassword := dataSourceInfo.BinData.Password
	if dataSourceInfo.BinData.Password != "" {
		decryptedPassword, err := utils.DecryptPassword(dataSourceInfo.BinData.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password for %s: %w", dataSourceID, err)
		}
		decodedPassword = decryptedPassword
	}

	// 如果没有指定schema，使用数据源中的schema
	if schema == "" {
		schema = dataSourceInfo.BinData.Schema
	}

	// 构建数据库配置
	config := &writer.DatabaseConfig{
		Host:     dataSourceInfo.BinData.Host,
		Port:     dataSourceInfo.BinData.Port,
		Username: dataSourceInfo.BinData.Account,
		Password: decodedPassword,
		Database: dataSourceInfo.BinData.DatabaseName,
		Driver:   dbType,
	}

	// 获取数据库连接
	dbConn, err := s.getDatabaseConnection(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}
	defer func() {
		if dbConn != nil {
			sqlDB, _ := dbConn.DB()
			if sqlDB != nil {
				_ = sqlDB.Close()
			}
		}
	}()

	// 获取对应的driver
	if !writer.IsDatabaseTypeSupported(dbType) {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	driver, _ := writer.GetDatabaseDriver(dbType)

	// 调用ListTableColumns方法
	writerColumns, err := driver.ListTableColumns(dbConn, tableName, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to list table columns: %w", err)
	}

	// 转换类型
	columns := make([]ColumnInfo, len(writerColumns))
	for i, wc := range writerColumns {
		columns[i] = ColumnInfo{
			Name:         wc.Name,
			DataType:     wc.DataType,
			DataLength:   wc.DataLength,
			IsNullable:   wc.IsNullable,
			PrimaryKey:   wc.PrimaryKey,
			DefaultValue: wc.DefaultValue,
			Comment:      wc.Comment,
			Precision:    wc.Precision,
			Scale:        wc.Scale,
		}
	}

	return columns, nil
}

// getDatabaseConnection 建立数据库连接
func (s *DatabaseTableServiceImpl) getDatabaseConnection(ctx context.Context, config *writer.DatabaseConfig) (*gorm.DB, error) {
	// 获取对应的driver
	if !writer.IsDatabaseTypeSupported(config.Driver) {
		return nil, fmt.Errorf("unsupported database type: %s", config.Driver)
	}

	driver, _ := writer.GetDatabaseDriver(config.Driver)

	// 建立连接
	return driver.GetConnection(config)
}
