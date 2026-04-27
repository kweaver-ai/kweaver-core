package dbaccess

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/db"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type outboxMessageDB struct {
	dbPool *sqlx.DB
	logger interfaces.Logger
	dbName string
	orm    *ormhelper.DB
}

var (
	outboxMessageOnce sync.Once
	outboxMessage     model.IOutboxMessage
)

const (
	tbOutboxMessageTableName = "t_outbox_message"
)

// NewOutboxMessageDB outbox消息事件表
func NewOutboxMessageDB() model.IOutboxMessage {
	outboxMessageOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		dbPool := db.NewDBPool()
		dbName := confLoader.GetDBName()
		orm := ormhelper.New(dbPool, dbName)
		outboxMessage = &outboxMessageDB{
			dbPool: dbPool,
			logger: confLoader.GetLogger(),
			dbName: dbName,
			orm:    orm,
		}
	})
	return outboxMessage
}

// Insert 添加消息事件
func (outboxMessage *outboxMessageDB) Insert(ctx context.Context, tx *sql.Tx, message *model.OutboxMessageDB) (eventID string, err error) {
	if message.EventID == "" {
		message.EventID = uuid.New().String()
	}
	eventID = message.EventID
	orm := outboxMessage.orm
	if tx != nil {
		orm = outboxMessage.orm.WithTx(tx)
	}
	row, err := orm.Insert().Into(tbOutboxMessageTableName).Values(map[string]interface{}{
		"f_event_id":      message.EventID,
		"f_event_type":    message.EventType,
		"f_topic":         message.Topic,
		"f_payload":       message.Payload,
		"f_status":        message.Status,
		"f_created_at":    time.Now().UnixNano(),
		"f_updated_at":    time.Now().UnixNano(),
		"f_next_retry_at": message.NextRetryAt,
		"f_retry_count":   message.RetryCount,
	}).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "insert outbox message failed, message: %v", message)
		return
	}
	ok, err := checkAffected(row)
	if err != nil {
		err = errors.Wrapf(err, "insert outbox message failed, message: %v", message)
		return
	}
	if !ok {
		err = errors.Errorf("insert outbox message failed, message: %v", message)
	}
	return
}

// UpdateByEventID 更新消息事件
func (outboxMessage *outboxMessageDB) UpdateByEventID(ctx context.Context, tx *sql.Tx, message *model.OutboxMessageDB) (err error) {
	orm := outboxMessage.orm
	if tx != nil {
		orm = outboxMessage.orm.WithTx(tx)
	}
	message.UpdatedAt = time.Now().UnixNano()
	_, err = orm.Update(tbOutboxMessageTableName).SetData(map[string]interface{}{
		"f_topic":         message.Topic,
		"f_payload":       message.Payload,
		"f_status":        message.Status,
		"f_updated_at":    time.Now().UnixNano(),
		"f_retry_count":   message.RetryCount,
		"f_next_retry_at": message.NextRetryAt,
	}).WhereEq("f_event_id", message.EventID).
		WhereEq("f_event_type", message.EventType).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "update outbox message failed, message: %v", message)
	}
	return
}

// GetByStatus 获取消息事件
func (outboxMessage *outboxMessageDB) GetByStatus(ctx context.Context, status string, limit int) (messages []*model.OutboxMessageDB, err error) {
	orm := outboxMessage.orm
	messages = []*model.OutboxMessageDB{}
	err = orm.Select().From(tbOutboxMessageTableName).
		WhereEq("f_status", status).
		WhereLt("f_next_retry_at", time.Now().UnixNano()).Limit(limit).Get(ctx, &messages)
	if err != nil {
		err = errors.Wrapf(err, "get outbox message failed, status: %s", status)
	}
	return
}

// DeleteByEventID 删除消息事件
func (outboxMessage *outboxMessageDB) DeleteByEventID(ctx context.Context, tx *sql.Tx, eventID string) (err error) {
	orm := outboxMessage.orm
	if tx != nil {
		orm = outboxMessage.orm.WithTx(tx)
	}
	_, err = orm.Delete().From(tbOutboxMessageTableName).WhereEq("f_event_id", eventID).Execute(ctx)
	if err != nil {
		err = errors.Wrapf(err, "delete outbox message failed, event_id: %s", eventID)
	}
	return
}
