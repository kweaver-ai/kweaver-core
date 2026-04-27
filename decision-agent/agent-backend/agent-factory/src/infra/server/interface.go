package server

import "context"

type IServer interface {
	Start()
	Shutdown(ctx context.Context) error
}
