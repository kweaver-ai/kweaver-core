package model

import (
	"context"
	"database/sql"
)

// OutboxMessageDB outbox消息表
type OutboxMessageDB struct {
	ID          int64  `json:"id" db:"f_id"`                       // 主键
	EventID     string `json:"event_id" db:"f_event_id"`           // 事件ID
	EventType   string `json:"type" db:"f_event_type"`             // 事件类型
	Topic       string `json:"topic" db:"f_topic"`                 // 消息Topic
	Payload     string `json:"payload" db:"f_payload"`             // 事件负载内容(message)
	Status      string `json:"status" db:"f_status"`               // 消息状态(待处理、失败)
	CreatedAt   int64  `json:"created_at" db:"f_created_at"`       // 创建时间
	UpdatedAt   int64  `json:"updated_at" db:"f_updated_at"`       // 更新时间
	NextRetryAt int64  `json:"next_retry_at" db:"f_next_retry_at"` // 下次重试时间
	RetryCount  int    `json:"retry_count" db:"f_retry_count"`     // 重试次数
}

const (
	OutboxMessageStatusPending string = "pending" // 待处理
	OutboxMessageStatusFailed  string = "failed"  // 失败
)

// IOutboxMessage outbox消息接口
type IOutboxMessage interface {
	Insert(ctx context.Context, tx *sql.Tx, message *OutboxMessageDB) (eventID string, err error)
	UpdateByEventID(ctx context.Context, tx *sql.Tx, message *OutboxMessageDB) error
	GetByStatus(ctx context.Context, status string, limit int) ([]*OutboxMessageDB, error)
	DeleteByEventID(ctx context.Context, tx *sql.Tx, eventID string) error
}
