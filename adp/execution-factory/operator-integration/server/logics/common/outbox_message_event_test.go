package common

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	mqmock "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/mq/mock"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestOutboxMessageEventPublishFallbackUsesDetachedContextOnCanceledRequest(t *testing.T) {
	Convey("Publish在请求上下文已取消时仍应写入outbox", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockMQClient := mqmock.NewMockMQClient(ctrl)
		mockOutboxMessageDB := mocks.NewMockIOutboxMessage(ctrl)

		event := &outboxMessageEvent{
			logger:          mockLogger,
			outboxMessageDB: mockOutboxMessageDB,
			mqClient:        mockMQClient,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req := &interfaces.OutboxMessageReq{
			EventID:   "evt-1",
			EventType: interfaces.OutboxMessageEventTypeAuditLog,
			Topic:     "topic.test",
			Payload:   `{"hello":"world"}`,
		}

		var insertCtxErr error
		var insertHasDeadline bool
		var insertedMessage *model.OutboxMessageDB

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return().Times(1)
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Return().AnyTimes()
		mockMQClient.EXPECT().Publish(ctx, req.Topic, []byte(req.Payload)).Return(context.Canceled).Times(1)
		mockOutboxMessageDB.EXPECT().
			Insert(gomock.Any(), (*sql.Tx)(nil), gomock.Any()).
			DoAndReturn(func(insertCtx context.Context, tx *sql.Tx, message *model.OutboxMessageDB) (string, error) {
				insertCtxErr = insertCtx.Err()
				_, insertHasDeadline = insertCtx.Deadline()
				insertedMessage = message
				return message.EventID, nil
			}).
			Times(1)

		err := event.Publish(ctx, req)

		So(err, ShouldBeNil)
		So(insertCtxErr, ShouldBeNil)
		So(insertHasDeadline, ShouldBeTrue)
		So(insertedMessage, ShouldNotBeNil)
		So(insertedMessage.EventID, ShouldEqual, req.EventID)
		So(insertedMessage.EventType, ShouldEqual, req.EventType.String())
		So(insertedMessage.Topic, ShouldEqual, req.Topic)
		So(insertedMessage.Payload, ShouldEqual, req.Payload)
		So(insertedMessage.Status, ShouldEqual, model.OutboxMessageStatusPending)
		So(insertedMessage.NextRetryAt, ShouldBeGreaterThan, time.Now().Add(-commonPollInterval).UnixNano())
	})

	Convey("Publish在MQ返回包装后的context canceled错误时仍应写入outbox", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockMQClient := mqmock.NewMockMQClient(ctrl)
		mockOutboxMessageDB := mocks.NewMockIOutboxMessage(ctrl)

		event := &outboxMessageEvent{
			logger:          mockLogger,
			outboxMessageDB: mockOutboxMessageDB,
			mqClient:        mockMQClient,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req := &interfaces.OutboxMessageReq{
			EventID:   "evt-2",
			EventType: interfaces.OutboxMessageEventTypeAuditLog,
			Topic:     "topic.test",
			Payload:   `{"hello":"wrapped"}`,
		}

		var insertCtxErr error
		var insertHasDeadline bool

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return().Times(1)
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Return().AnyTimes()
		mockMQClient.EXPECT().Publish(ctx, req.Topic, []byte(req.Payload)).Return(fmt.Errorf("mq publish failed: %w", context.Canceled)).Times(1)
		mockOutboxMessageDB.EXPECT().
			Insert(gomock.Any(), (*sql.Tx)(nil), gomock.Any()).
			DoAndReturn(func(insertCtx context.Context, tx *sql.Tx, message *model.OutboxMessageDB) (string, error) {
				insertCtxErr = insertCtx.Err()
				_, insertHasDeadline = insertCtx.Deadline()
				return message.EventID, nil
			}).
			Times(1)

		err := event.Publish(ctx, req)

		So(err, ShouldBeNil)
		So(insertCtxErr, ShouldBeNil)
		So(insertHasDeadline, ShouldBeTrue)
	})

	Convey("Publish在MQ返回context deadline exceeded错误时仍应写入outbox", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockMQClient := mqmock.NewMockMQClient(ctrl)
		mockOutboxMessageDB := mocks.NewMockIOutboxMessage(ctrl)

		event := &outboxMessageEvent{
			logger:          mockLogger,
			outboxMessageDB: mockOutboxMessageDB,
			mqClient:        mockMQClient,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		time.Sleep(time.Millisecond)

		req := &interfaces.OutboxMessageReq{
			EventID:   "evt-3",
			EventType: interfaces.OutboxMessageEventTypeAuditLog,
			Topic:     "topic.test",
			Payload:   `{"hello":"timeout"}`,
		}

		var insertCtxErr error
		var insertHasDeadline bool

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return().Times(1)
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Return().AnyTimes()
		mockMQClient.EXPECT().Publish(ctx, req.Topic, []byte(req.Payload)).Return(fmt.Errorf("mq publish timeout: %w", context.DeadlineExceeded)).Times(1)
		mockOutboxMessageDB.EXPECT().
			Insert(gomock.Any(), (*sql.Tx)(nil), gomock.Any()).
			DoAndReturn(func(insertCtx context.Context, tx *sql.Tx, message *model.OutboxMessageDB) (string, error) {
				insertCtxErr = insertCtx.Err()
				_, insertHasDeadline = insertCtx.Deadline()
				return message.EventID, nil
			}).
			Times(1)

		err := event.Publish(ctx, req)

		So(err, ShouldBeNil)
		So(insertCtxErr, ShouldBeNil)
		So(insertHasDeadline, ShouldBeTrue)
	})
}
