package writer

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	// _ "github.com/AISHU-Technology/proton-rds-sdk-go/driver" // Disabled: DM8 and KDB dialects not available
	// dm "devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton_dm_dialect_go"
	// kdb "devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton_kdb_dialect_go"
	"github.com/go-sql-driver/mysql"
	_ "github.com/sijms/go-ora/v2"
	mysqld "gorm.io/driver/mysql"
	postgresd "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// UniversalDatabaseWriter 通用数据库写入器
type UniversalDatabaseWriter struct {
	registry *DatabaseDriverRegistry
	// useDistributedConnection 是否使用分布式连接管理
	// true: 使用每个driver的GetConnection方法
	// false: 使用集中式的getDBConnection方法 (默认)
	useDistributedConnection bool
	// useDistributedTableCreation 是否使用分布式表创建管理
	// true: 使用每个driver的CreateTableIfNotExists方法
	// false: 使用集中式的createTableIfNotExists方法 (默认)
	useDistributedTableCreation bool
}

func NewUniversalDatabaseWriter(registry *DatabaseDriverRegistry) *UniversalDatabaseWriter {
	return &UniversalDatabaseWriter{registry: registry}
}

// SetDistributedConnection 设置是否使用分布式连接管理
func (w *UniversalDatabaseWriter) SetDistributedConnection(useDistributed bool) {
	w.useDistributedConnection = useDistributed
}

// SetDistributedTableCreation 设置是否使用分布式表创建管理
func (w *UniversalDatabaseWriter) SetDistributedTableCreation(useDistributed bool) {
	w.useDistributedTableCreation = useDistributed
}

// NewUniversalDatabaseWriterWithOptions 创建带有选项的数据库写入器
func NewUniversalDatabaseWriterWithOptions(registry *DatabaseDriverRegistry, useDistributedConnection bool) *UniversalDatabaseWriter {
	return &UniversalDatabaseWriter{
		registry:                 registry,
		useDistributedConnection: useDistributedConnection,
	}
}

// NewUniversalDatabaseWriterWithFullOptions 创建带有完整选项的数据库写入器
func NewUniversalDatabaseWriterWithFullOptions(registry *DatabaseDriverRegistry, useDistributedConnection, useDistributedTableCreation bool) *UniversalDatabaseWriter {
	return &UniversalDatabaseWriter{
		registry:                    registry,
		useDistributedConnection:    useDistributedConnection,
		useDistributedTableCreation: useDistributedTableCreation,
	}
}

// Execute 执行数据库操作
func (w *UniversalDatabaseWriter) Execute(ctx context.Context, tableInfo *TableInfo, data []map[string]interface{}, where interface{}, operation string) (*ExecutionResult, error) {
	// 获取数据库驱动
	dbType := strings.ToLower(tableInfo.DatasourceType)
	driver, exists := w.registry.GetDriver(dbType)
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", tableInfo.DatasourceType)
	}

	// 获取执行器
	executor, exists := w.registry.GetExecutor(dbType)
	if !exists {
		return nil, fmt.Errorf("no executor found for database type: %s", tableInfo.DatasourceType)
	}

	// 建立数据库连接
	config, err := w.buildConfig(tableInfo)
	if err != nil {
		return nil, err
	}

	// 根据配置选择连接管理方式
	var dbConn *gorm.DB
	if w.useDistributedConnection {
		dbConn, err = w.getDBConnectionByDriver(config, driver)
	} else {
		dbConn, err = w.getDBConnectionCentralized(config, driver)
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if dbConn != nil {
			sqlDB, _ := dbConn.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		}
	}()

	// 检查并创建表（如果需要）
	if !tableInfo.TableExist && len(tableInfo.Fields) > 0 {
		// 根据配置选择表创建方式
		if w.useDistributedTableCreation {
			err = w.createTableIfNotExistsByDriver(dbConn, tableInfo, driver)
		} else {
			err = w.createTableIfNotExists(dbConn, tableInfo, driver)
		}
		if err != nil {
			return nil, err
		}
	}

	// 执行操作
	switch strings.ToLower(operation) {
	case OperationInsert, OperationAppend:
		return executor.ExecuteInsert(ctx, dbConn, tableInfo, driver, data)
	case OperationUpdate:
		return executor.ExecuteUpdate(ctx, dbConn, tableInfo, driver, data, where)
	case OperationDelete:
		return executor.ExecuteDelete(ctx, dbConn, tableInfo, driver, where)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// buildConfig 从输入构建数据库配置
func (w *UniversalDatabaseWriter) buildConfig(tableInfo *TableInfo) (*DatabaseConfig, error) {
	conn := tableInfo.Conn
	if conn == nil {
		// 如果没有直接提供连接信息，尝试从数据源ID解析
		// 这里暂时不支持数据源ID解析，未来可以扩展
		return nil, fmt.Errorf("connection information is required")
	}

	// 验证必要的连接信息
	if conn.Host == "" {
		return nil, fmt.Errorf("database host is required")
	}
	if conn.Port == 0 {
		return nil, fmt.Errorf("database port is required and must be greater than 0")
	}
	if conn.Username == "" {
		return nil, fmt.Errorf("database username is required")
	}
	if conn.Database == "" {
		return nil, fmt.Errorf("database name is required")
	}

	// 设置默认参数（如果为空）
	params := conn.Params
	if params == nil {
		params = make(map[string]string)
	}

	return &DatabaseConfig{
		Host:     conn.Host,
		Port:     conn.Port,
		Username: conn.Username,
		Password: conn.Password,
		Database: conn.Database,
		Driver:   tableInfo.DatasourceType,
		Params:   params,
	}, nil
}

// getDBConnectionCentralized 集中式数据库连接实现
func (w *UniversalDatabaseWriter) getDBConnectionCentralized(config *DatabaseConfig, driver DatabaseDriver) (*gorm.DB, error) {
	var dsn string
	if strings.EqualFold(config.Driver, DatabaseTypePostgres) {
		// PostgreSQL DSN format
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database)
		for k, v := range config.Params {
			dsn += fmt.Sprintf(" %s=%s", k, v)
		}
		if _, hasSSLMode := config.Params["sslmode"]; !hasSSLMode {
			dsn += " sslmode=disable"
		}
	} else {
		// MySQL DSN format (default for other databases)
		dsnConfig := mysql.NewConfig()
		dsnConfig.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
		dsnConfig.User = config.Username
		dsnConfig.Passwd = config.Password
		dsnConfig.DBName = config.Database
		dsnConfig.Net = "tcp"
		params := map[string]string{
			"charset":   "utf8mb4",
			"parseTime": "true",
		}
		for k, v := range config.Params {
			params[k] = v
		}
		dsnConfig.Params = params
		dsn = dsnConfig.FormatDSN()
	}

	var dial gorm.Dialector
	var sqlDB *sql.DB
	var err error

	switch strings.ToLower(config.Driver) {
	// DM8 dialect is disabled due to unavailable internal dependency
	// case DatabaseTypeDM8:
	// 	sqlDB, err = sql.Open("DM8", dsn)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to open DM8 connection to %s:%d as %s (database: %s): %w",
	// 			config.Host, config.Port, config.Username, config.Database, err)
	// 	}
	// 	sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
	// 	sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
	// 	sqlDB.SetConnMaxLifetime(time.Hour)
	// 	dial = dm.New(dm.Config{Conn: sqlDB})

	// KDB dialect is disabled due to unavailable internal dependency
	// case DatabaseTypeKDB:
	// 	sqlDB, err = sql.Open("kdb", dsn)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to open KDB connection to %s:%d as %s (database: %s): %w",
	// 			config.Host, config.Port, config.Username, config.Database, err)
	// 	}
	// 	sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
	// 	sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
	// 	sqlDB.SetConnMaxLifetime(time.Hour)
	// 	dial = kdb.New(kdb.Config{Conn: sqlDB})

	case DatabaseTypePostgres:
		sqlDB, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open PostgreSQL connection to %s:%d as %s (database: %s): %w",
				config.Host, config.Port, config.Username, config.Database, err)
		}
		sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
		sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
		// Use official GORM PostgreSQL driver
		dial = postgresd.New(postgresd.Config{Conn: sqlDB})

	default:
		// Default to proton-rds for MySQL and other compatible databases
		sqlDB, err = sql.Open("proton-rds", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open MySQL connection to %s:%d as %s (database: %s): %w",
				config.Host, config.Port, config.Username, config.Database, err)
		}
		sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
		sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
		dial = mysqld.New(mysqld.Config{Conn: sqlDB})
	}

	gormDB, err := gorm.Open(dial, &gorm.Config{Logger: getLogger()})
	if err != nil {
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, fmt.Errorf("failed to create GORM connection: %w", err)
	}

	return gormDB, nil
}

// getDBConnectionByDriver 使用driver的GetConnection方法获取数据库连接
// 这是一个新的分布式连接管理方法，允许每个driver管理自己的连接逻辑
func (w *UniversalDatabaseWriter) getDBConnectionByDriver(config *DatabaseConfig, driver DatabaseDriver) (*gorm.DB, error) {
	if driverConn, err := driver.GetConnection(config); err != nil {
		// driver的GetConnection方法失败，返回错误
		return nil, fmt.Errorf("driver connection failed: %w", err)
	} else if driverConn != nil {
		return driverConn, nil
	}
	return w.getDBConnectionCentralized(config, driver)
}

// createTableIfNotExistsByDriver 使用driver的CreateTableIfNotExists方法创建表
// 这是一个新的分布式表管理方法，允许每个driver管理自己的建表逻辑
func (w *UniversalDatabaseWriter) createTableIfNotExistsByDriver(dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver) error {
	// 尝试使用driver的CreateTableIfNotExists方法（如果实现的话）
	if err := driver.CreateTableIfNotExists(dbConn, tableInfo); err == nil {
		return nil
	}

	// 如果driver没有实现CreateTableIfNotExists方法，fallback到集中式方法
	return w.createTableIfNotExists(dbConn, tableInfo, driver)
}

// createTableIfNotExists 根据字段映射创建表（如果不存在）
func (w *UniversalDatabaseWriter) createTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver) error {
	if len(tableInfo.Fields) == 0 {
		return fmt.Errorf("no field mappings provided for table creation")
	}

	// 使用驱动生成创建表的 SQL
	createSQL, err := driver.GenerateCreateTableSQL(tableInfo)
	if err != nil {
		return fmt.Errorf("failed to generate create table SQL: %w", err)
	}

	dbTypeLower := strings.ToLower(tableInfo.DatasourceType)

	// 对于PostgreSQL，可能包含多个SQL语句（建表+注释）
	if dbTypeLower == DatabaseTypePostgreSQL || dbTypeLower == DatabaseTypePostgres {
		sqlStatements := strings.Split(createSQL, "; ")
		for _, sql := range sqlStatements {
			sql = strings.TrimSpace(sql)
			if sql == "" {
				continue
			}

			if err := dbConn.Exec(sql).Error; err != nil {
				// 如果是表已存在的错误或列注释已存在的错误，忽略
				if strings.Contains(err.Error(), "already exists") ||
					strings.Contains(err.Error(), "already_exists") ||
					strings.Contains(err.Error(), "42P07") || // PostgreSQL table exists error code
					strings.Contains(err.Error(), "42710") { // PostgreSQL comment already exists
					continue
				}
				return fmt.Errorf("failed to execute SQL: %s, error: %w", sql, err)
			}
		}
	} else {
		// MySQL和其他数据库的单条SQL执行
		if err := dbConn.Exec(createSQL).Error; err != nil {
			// 如果是表已存在的错误，忽略
			if strings.Contains(err.Error(), "already exists") ||
				strings.Contains(err.Error(), "already_exists") ||
				strings.Contains(err.Error(), "1050") { // MySQL table exists error code
				return nil
			}
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// 初始化数据库驱动
func init() {
	// 创建注册器
	globalRegistry = NewDatabaseDriverRegistry()

	// 注册MySQL驱动
	mysqlDriver := &MySQLDriver{}
	mysqlExecutor := &MySQLExecutor{}
	globalRegistry.Register(DatabaseTypeMySQL, mysqlDriver, mysqlExecutor)
	globalRegistry.Register(DatabaseTypeMariaDB, mysqlDriver, mysqlExecutor)
	globalRegistry.Register(DatabaseTypeMaria, mysqlDriver, mysqlExecutor)

	// 注册PostgreSQL驱动
	pgDriver := &PostgreSQLDriver{}
	pgExecutor := &PostgreSQLExecutor{}
	globalRegistry.Register(DatabaseTypePostgreSQL, pgDriver, pgExecutor)
	globalRegistry.Register(DatabaseTypePostgres, pgDriver, pgExecutor)

	// DM8 driver disabled due to unavailable internal dependencies
	// dm8Driver := &DM8Driver{}
	// dm8Executor := &DM8Executor{}
	// globalRegistry.Register(DatabaseTypeDM8, dm8Driver, dm8Executor)
	// globalRegistry.Register(DatabaseTypeDM, dm8Driver, dm8Executor)

	// KDB driver disabled due to unavailable internal dependencies
	// kdbDriver := &KDBDriver{}
	// kdbExecutor := &KDBExecutor{}
	// globalRegistry.Register(DatabaseTypeKDB, kdbDriver, kdbExecutor)

	// 注册SQL Server驱动
	sqlServerDriver := &SQLServerDriver{}
	sqlServerExecutor := &SQLServerExecutor{}
	globalRegistry.Register(DatabaseTypeSQLServer, sqlServerDriver, sqlServerExecutor)
	globalRegistry.Register(DatabaseTypeMSSQL, sqlServerDriver, sqlServerExecutor)

	// 注册Oracle驱动
	oracleDriver := &OracleDriver{}
	oracleExecutor := &OracleExecutor{}
	globalRegistry.Register(DatabaseTypeOracle, oracleDriver, oracleExecutor)

	globalWriter = NewUniversalDatabaseWriter(globalRegistry)

	// 创建分布式连接管理的写入器实例
	globalWriterDistributed = NewUniversalDatabaseWriterWithOptions(globalRegistry, true)

	// 创建完整分布式管理的写入器实例（连接+表创建）
	globalWriterFullyDistributed = NewUniversalDatabaseWriterWithFullOptions(globalRegistry, true, true)
}

// getLogger 根据环境变量"debug"决定日志等级
func getLogger() logger.Interface {
	debug := os.Getenv("DEBUG")
	if debug == "true" || debug == "1" || debug == "yes" {
		return logger.Default.LogMode(logger.Info)
	}
	return logger.Default.LogMode(logger.Silent)
}
