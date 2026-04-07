package httpserver

import "context"

// IServer HTTP服务器接口
type IServer interface {
	Start()
	Shutdown(ctx context.Context) error
}
