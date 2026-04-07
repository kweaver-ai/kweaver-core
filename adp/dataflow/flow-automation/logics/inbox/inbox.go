// Package inbox logics 消息入队策略
package inbox

import (
	"context"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
)

// CheckAuthRes 鉴权结果信息
type CheckAuthRes struct {
	Status bool   `json:"status"`
	URL    string `json:"url"`
}

// Handler method interface
type Handler interface {
	SaveMsg(ctx context.Context, msg *common.DocMsg, itemID, topic string) error
	QueryMsgs(ctx context.Context, rev string, topics []string) ([]*entity.InBox, error)
}

var (
	iOnce sync.Once
	i     Handler
)

type inbox struct {
	config     *common.Config
	hydra      drivenadapters.HydraPublic
	hydraAdmin drivenadapters.HydraAdmin
	store      mod.Store
}

// NewInbox new inbox instance
func NewInbox() Handler {
	iOnce.Do(func() {
		i = &inbox{
			hydra:      drivenadapters.NewHydraPublic(),
			hydraAdmin: drivenadapters.NewHydraAdmin(),
			config:     common.NewConfig(),
			store:      mod.GetStore(),
		}
	})
	return i
}

// SaveMsg 暂存重复度较高的消息
func (a *inbox) SaveMsg(ctx context.Context, msg *common.DocMsg, itemID, topic string) error {
	var err error
	ctx, span := trace.StartConsumerSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if topic == common.TopicFileUpload || topic == common.TopicFileEdit || topic == common.TopicFileCreate {
		inbox := &entity.InBox{
			Msg:   *msg,
			Topic: topic,
			DocID: itemID,
		}
		err := a.store.CreateInbox(ctx, inbox)
		return err
	}
	return nil
}

func (a *inbox) QueryMsgs(ctx context.Context, itemID string, topics []string) ([]*entity.InBox, error) {
	var err error
	ctx, span := trace.StartConsumerSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	msgs, err := a.store.ListInbox(ctx, &mod.ListInboxInput{DocID: itemID, Topics: topics})

	if err != nil {
		traceLog.WithContext(ctx).Warnf("listInbox error: %v", err)
		return nil, err
	}

	return msgs, nil
}
