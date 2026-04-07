package writer

import "strings"

// DatabaseDriverRegistry 数据库驱动注册器
type DatabaseDriverRegistry struct {
	drivers   map[string]DatabaseDriver
	executors map[string]DatabaseExecutor
}

func NewDatabaseDriverRegistry() *DatabaseDriverRegistry {
	return &DatabaseDriverRegistry{
		drivers:   make(map[string]DatabaseDriver),
		executors: make(map[string]DatabaseExecutor),
	}
}

func (r *DatabaseDriverRegistry) Register(name string, driver DatabaseDriver, executor DatabaseExecutor) {
	r.drivers[name] = driver
	r.executors[name] = executor
}

func (r *DatabaseDriverRegistry) GetDriver(name string) (DatabaseDriver, bool) {
	driver, exists := r.drivers[name]
	return driver, exists
}

func (r *DatabaseDriverRegistry) GetExecutor(name string) (DatabaseExecutor, bool) {
	executor, exists := r.executors[name]
	return executor, exists
}

func (r *DatabaseDriverRegistry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.drivers))
	for name := range r.drivers {
		types = append(types, name)
	}
	return types
}

// 全局变量
var (
	globalRegistry               *DatabaseDriverRegistry
	globalWriter                 *UniversalDatabaseWriter
	globalWriterDistributed      *UniversalDatabaseWriter
	globalWriterFullyDistributed *UniversalDatabaseWriter
)

// GetSupportedDatabaseTypes 获取支持的数据库类型
func GetSupportedDatabaseTypes() []string {
	if globalRegistry != nil {
		return globalRegistry.GetSupportedTypes()
	}
	return []string{}
}

// RegisterDatabaseDriver 注册新的数据库驱动
func RegisterDatabaseDriver(name string, driver DatabaseDriver, executor DatabaseExecutor) error {
	if globalRegistry == nil {
		globalRegistry = NewDatabaseDriverRegistry()
	}
	globalRegistry.Register(name, driver, executor)
	return nil
}

// GetDatabaseDriver 获取指定数据库的驱动
func GetDatabaseDriver(name string) (DatabaseDriver, bool) {
	if globalRegistry == nil {
		return nil, false
	}
	return globalRegistry.GetDriver(name)
}

// GetDatabaseExecutor 获取指定数据库的执行器
func GetDatabaseExecutor(name string) (DatabaseExecutor, bool) {
	if globalRegistry == nil {
		return nil, false
	}
	return globalRegistry.GetExecutor(name)
}

// GetGlobalWriter 获取全局数据库写入器 (集中式连接管理)
func GetGlobalWriter() *UniversalDatabaseWriter {
	return globalWriter
}

// GetGlobalWriterDistributed 获取全局分布式数据库写入器 (分布式连接管理)
func GetGlobalWriterDistributed() *UniversalDatabaseWriter {
	return globalWriterDistributed
}

// GetGlobalWriterFullyDistributed 获取全局完整分布式数据库写入器 (分布式连接+表创建)
func GetGlobalWriterFullyDistributed() *UniversalDatabaseWriter {
	return globalWriterFullyDistributed
}

// IsDatabaseTypeSupported 检查数据库类型是否支持
func IsDatabaseTypeSupported(dbType string) bool {
	if globalRegistry == nil {
		return false
	}
	_, exists := globalRegistry.GetDriver(strings.ToLower(dbType))
	return exists
}
