// Package interfaces 定义接口
// @file interfaces.go
// @description: 定义接口
package interfaces

//go:generate mockgen -source=interface.go -destination=../mocks/interface.go -package=mocks
import "context"

// App 应用接口
type App interface {
	Start() error
	Stop(context.Context)
}

// LogModelOperator 日志模型操作器
type LogModelOperator[T any] interface {
	Logger(context.Context, T)
}

const (
	DefaultBatchSize = 1000 // 默认批量大小为1000
	MaxQuerySize     = 5000 // 最大查询数量为5000
)

// ResourceObjectType 资源对象类型
type ResourceObjectType string

const (
	ResourceObjectTool     ResourceObjectType = "tool"     // 工具
	ResourceObjectOperator ResourceObjectType = "operator" // 算子
)

// ResourceDeployType 资源部署类型
type ResourceDeployType string

func (r ResourceDeployType) String() string {
	return string(r)
}

const (
	ResourceDeployTypeMCP ResourceDeployType = "mcp"
)
