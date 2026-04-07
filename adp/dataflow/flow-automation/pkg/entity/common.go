package entity

import (
	"context"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/store"
)

// Accessor 访问者
type Accessor struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// BaseInfo
type BaseInfo struct {
	ID        string `yaml:"id" json:"id" bson:"_id"`
	CreatedAt int64  `yaml:"createdAt" json:"createdAt" bson:"createdAt"`
	UpdatedAt int64  `yaml:"updatedAt" json:"updatedAt" bson:"updatedAt"`
}

// GetBaseInfo getter
func (b *BaseInfo) GetBaseInfo() *BaseInfo {
	return b
}

// Initial base info
func (b *BaseInfo) Initial() {
	if b.ID == "" {
		b.ID = store.NextStringID()
	}
	b.CreatedAt = time.Now().Unix()
	b.UpdatedAt = time.Now().Unix()
}

// Update
func (b *BaseInfo) Update() {
	b.UpdatedAt = time.Now().Unix()
}

// BaseInfoGetter
type BaseInfoGetter interface {
	GetBaseInfo() *BaseInfo
}

// CtxKey 上下文key
type CtxKey string

const (
	// CtxKeyRunningTaskIns 运行中任务上下文key
	CtxKeyRunningTaskIns CtxKey = "running-task"
)

// CtxWithRunningTaskIns 上下文中任务实例
func CtxWithRunningTaskIns(ctx context.Context, task *TaskInstance) context.Context {
	return context.WithValue(ctx, CtxKeyRunningTaskIns, task)
}

// CtxRunningTaskIns 上下文中任务实例
func CtxRunningTaskIns(ctx context.Context) (*TaskInstance, bool) {
	ins, ok := ctx.Value(CtxKeyRunningTaskIns).(*TaskInstance)
	return ins, ok
}
