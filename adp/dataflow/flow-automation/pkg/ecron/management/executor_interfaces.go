package management

import (
	"os/exec"
)

//go:generate mockgen -package mock -source executor_interfaces.go -destination ../mock/mock_executor_interfaces.go

// CommandRunner 定义命令执行接口
type CommandRunner interface {
	Run(name string, args ...string) error
}

// DefaultCommandRunner 默认的命令执行器实现
type DefaultCommandRunner struct{}

// NewCommandRunner 创建默认的命令执行器
func NewCommandRunner() CommandRunner {
	return &DefaultCommandRunner{}
}

// Run 实现 CommandRunner 接口
func (r *DefaultCommandRunner) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// MapDecoder 定义 map 解码接口
type MapDecoder interface {
	Decode(input interface{}, output interface{}) error
}

// DefaultMapDecoder 默认的 map 解码器实现
type DefaultMapDecoder struct{}

// NewMapDecoder 创建默认的 map 解码器
func NewMapDecoder() MapDecoder {
	return &DefaultMapDecoder{}
}

// Decode 实现 MapDecoder 接口
func (d *DefaultMapDecoder) Decode(input interface{}, output interface{}) error {
	// 使用 mapstructure 库进行解码
	// 注意：这里需要导入 mapstructure
	return decodeMap(input, output)
}
