package common

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/creasty/defaults"
	validator "github.com/go-playground/validator/v10"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/lock"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/mq"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	redis "github.com/redis/go-redis/v9"
)

const (
	lockKeyTemp          = "agent-operator-integration:outbox_message_event:lock" // 锁键模板
	defultLockExpiryTime = 1 * time.Minute                                        // 默认锁过期时间
	commonPollInterval   = 3 * time.Second                                        // 扫描间隔
	queryDefaultLimit    = 100                                                    // 默认查询数量
	defaultTimeout       = 30 * time.Second                                       // 默认超时时间
)

var (
	outboxOnce  sync.Once
	outboxEvent *outboxMessageEvent
)

// OutboxMessageEvent 消息事件管理
type outboxMessageEvent struct {
	confLoader      *config.Config
	logger          interfaces.Logger
	outboxMessageDB model.IOutboxMessage
	mqClient        mq.MQClient
	redisCli        *redis.Client
	quit            chan bool
}

// NewOutboxMessageEvent 创建消息事件管理
func NewOutboxMessageEvent() *outboxMessageEvent {
	outboxOnce.Do(func() {
		conf := config.NewConfigLoader()
		cli, _, err := conf.RedisConfig.GetClient()
		if err != nil {
			panic(fmt.Sprintf("get redis client failed: %v", err))
		}
		outboxEvent = &outboxMessageEvent{
			confLoader:      conf,
			logger:          conf.GetLogger(),
			outboxMessageDB: dbaccess.NewOutboxMessageDB(),
			mqClient:        mq.NewMQClient(),
			redisCli:        cli,
			quit:            make(chan bool),
		}
	})
	return outboxEvent
}

// Start 启动 outboxMessageEvent
func (m *outboxMessageEvent) Start() error {
	m.logger.Info("[outboxMessageEvent] start scan outbox message event")
	go func() {
		ticker := time.NewTicker(commonPollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.scan(context.Background())
			case <-m.quit:
				m.logger.Info("[outboxMessageEvent] stop scan outbox message event")
				return
			}
		}
	}()
	return nil
}
func (m *outboxMessageEvent) scan(ctx context.Context) {
	var err error
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 获取分布式锁
	v := m.confLoader.Project.GetMachineID()
	locker := lock.NewRedisLocker(m.redisCli, lockKeyTemp, v, defultLockExpiryTime)
	ok, err := locker.Lock(ctx)
	if err != nil && err != redis.Nil {
		m.logger.WithContext(ctx).Warnf("[auditLogHandler] processFaildLogData get lock err: %s", err.Error())
		return
	}
	if !ok {
		return
	}
	defer locker.Unlock(ctx)
	// 获取未处理的消息
	events, err := m.outboxMessageDB.GetByStatus(ctx, model.OutboxMessageStatusPending, queryDefaultLimit)
	if err != nil {
		m.logger.WithContext(ctx).Errorf("[auditLogHandler] processFaildLogData get outbox message err: %s", err.Error())
		return
	}
	// 处理消息事件
	for _, event := range events {
		m.processOutboxEventMessage(ctx, event)
	}
}

func (m *outboxMessageEvent) processOutboxEventMessage(ctx context.Context, event *model.OutboxMessageDB) {
	// 发送消息到MQ
	err := m.mqClient.Publish(ctx, event.Topic, []byte(event.Payload))
	if err == nil {
		// 清理消息
		err = m.outboxMessageDB.DeleteByEventID(ctx, nil, event.EventID)
		if err != nil {
			m.logger.WithContext(ctx).Errorf("delete outbox message failed: %v, topic:%s, message:%s", err, event.Topic, event.Payload)
		}
		return
	}
	m.logger.WithContext(ctx).Warnf("publish outbox message failed: %v, topic:%s, message:%s", err, event.Topic, event.Payload)
	event.RetryCount++
	event.NextRetryAt = time.Now().Add(time.Duration(event.RetryCount) * commonPollInterval).UnixNano()
	err = m.outboxMessageDB.UpdateByEventID(ctx, nil, event)
	if err != nil {
		m.logger.WithContext(ctx).Errorf("update outbox message failed: %v, topic:%s, message:%s", err, event.Topic, event.Payload)
	}
}

// Stop 停止 outboxMessageEvent
func (m *outboxMessageEvent) Stop(ctx context.Context) {
	close(m.quit)
}

// Publish 发布消息事件
func (m *outboxMessageEvent) Publish(ctx context.Context, req *interfaces.OutboxMessageReq) (err error) {
	// 参数校验
	err = defaults.Set(req)
	if err != nil {
		return
	}
	err = validator.New().Struct(req)
	if err != nil {
		return
	}
	// 发送消息到MQ
	err = m.mqClient.Publish(ctx, req.Topic, []byte(req.Payload))
	if err == nil {
		return
	}
	m.logger.WithContext(ctx).Warnf("publish outbox message failed: %v, topic:%s, message:%s", err, req.Topic, req.Payload)
	if strings.Contains(err.Error(), "context canceled") {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	// 处理失败，保存消息到数据库
	event := &model.OutboxMessageDB{
		EventID:     req.EventID,
		EventType:   req.EventType.String(),
		Topic:       req.Topic,
		Payload:     req.Payload,
		NextRetryAt: time.Now().Add(commonPollInterval).UnixNano(),
		Status:      model.OutboxMessageStatusPending,
	}
	// 保存消息到数据库
	_, err = m.outboxMessageDB.Insert(ctx, nil, event)
	if err != nil {
		m.logger.WithContext(ctx).Errorf("insert outbox message failed: %v", err)
	}
	return
}
