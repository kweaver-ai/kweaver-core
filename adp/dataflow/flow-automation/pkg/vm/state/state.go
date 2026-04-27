package state

type State int

const (
	Init  State = iota // 初始化
	Run                // 运行
	Wait               // 等待
	Done               // 完成
	Error              // 错误
)
