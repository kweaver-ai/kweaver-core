package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	cmq "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/mq"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/mod"
)

const (
	defaultInternal = 30 * time.Second
	restartInternal = 3 * time.Second
)

// OutBoxHandler interface
type OutBox interface {
	StartPushMessage(ctx context.Context)
}

type outbox struct {
	log        commonLog.Logger
	store      mod.Store
	mq         cmq.MQClient
	httpClient otelHttp.HTTPClient
	config     *common.Config
	retryMap   map[string]int64
}

var (
	obOnce sync.Once
	ob     OutBox
)

// NewOutBox 实例化outbox
func NewOutBox() OutBox {
	obOnce.Do(func() {
		ob = &outbox{
			log:        commonLog.NewLogger(),
			store:      mod.GetStore(),
			mq:         cmq.NewMQClient(),
			httpClient: drivenadapters.NewOtelHTTPClient(),
			config:     common.NewConfig(),
			retryMap:   make(map[string]int64),
		}
	})
	return ob
}

// StartPushMessage 开启消息推送服务
func (o *outbox) StartPushMessage(ctx context.Context) {
	go func() {
		o.log.Infof("[StartPushMessage] push message thread start...")
		timer := time.NewTimer(defaultInternal)
		defer func() {
			timer.Stop()
			o.log.Errorf("[StartPushMessage] thread closed...")
			if rErr := recover(); rErr != nil {
				o.log.Errorf("[StartPushMessage] panic occurred, detail: %v", rErr)
				time.Sleep(restartInternal)
				go o.StartPushMessage(ctx)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				input := &entity.OutBoxInput{
					CreateTime: time.Now().Unix(),
					Limit:      1000,
				}
				messages, err := o.store.ListOutBoxMessage(ctx, input)
				if err != nil {
					traceLog.WithContext(ctx).Warnf("[StartPushMessage] ListOutBoxMessage error: %v", err)
					timer.Reset(defaultInternal)
					continue
				}
				if len(messages) == 0 {
					timer.Reset(defaultInternal)
					continue
				}
				var ids []string
				tasksIDMap := make(map[string]string, 0)
				retryList := make([]*entity.OutBox, 0)
				for _, message := range messages {
					// 失败的nsq消息可能存在旧的topic数据，需适配
					err = o.mq.Publish(message.Topic, []byte(message.Msg))
					if err != nil {
						traceLog.WithContext(ctx).Warnf("[StartPushMessage] Publish error: %v", err)
						retryTime := o.retryMap[message.ID]
						if retryTime >= 5 {
							// 超过最大重试次数，丢弃消息
							delete(o.retryMap, message.ID)
							ids = append(ids, message.ID)
							tasksID := o.parseMessage(message.Msg)
							if tasksID != "" {
								tasksIDMap[tasksID] = err.Error()
							}
							continue
						}

						// 计算下次执行时间
						currentTime := time.Unix(message.CreatedAt, 0)
						currentTime = currentTime.Add((1 << retryTime) * time.Minute)
						message.CreatedAt = currentTime.Unix()
						message.UpdatedAt = currentTime.Unix()

						// 重试次数+1
						retryTime++
						o.retryMap[message.ID] = retryTime
						retryList = append(retryList, message)
						continue
					}
					ids = append(ids, message.ID)
				}

				o.batchUpdateOutBoxMessage(ctx, retryList)

				if len(ids) == 0 {
					timer.Reset(defaultInternal)
					continue
				}

				err = o.store.DeleteOutBoxMessage(ctx, ids)
				if err != nil {
					traceLog.WithContext(ctx).Warnf("[StartPushMessage] DeleteOutBoxMessage error: %v", err)
					timer.Reset(defaultInternal)
					continue
				}

				// 清理阻塞的任务记录
				if len(tasksIDMap) != 0 {
					o.resetFailedTaskStatus(tasksIDMap)
				}

				timer.Reset(defaultInternal)
			}
		}
	}()
}

func (o *outbox) batchUpdateOutBoxMessage(ctx context.Context, retryList []*entity.OutBox) {
	if len(retryList) == 0 {
		return
	}
	var ids []string
	for _, val := range retryList {
		ids = append(ids, val.ID)
	}
	err := o.store.DeleteOutBoxMessage(ctx, ids)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[batchUpdateOutBoxMessage] DeleteOutBoxMessage error: %v", err)
		o.clearUselessData(ids)
		return
	}
	err = o.store.BatchCreatOutBoxMessage(ctx, retryList)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[batchUpdateOutBoxMessage] BatchCreatOutBoxMessage error: %v", err)
		o.clearUselessData(ids)
	}
}

func (o *outbox) clearUselessData(ids []string) {
	for _, id := range ids {
		delete(o.retryMap, id)
	}
}

func (o *outbox) parseMessage(msg string) string {
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(msg), &data)
	if _, ok := data["apply_id"]; ok {
		return fmt.Sprintf("%v", data["apply_id"])
	}
	if _data, ok := data["process"]; ok {
		if process, ok := _data.(map[string]interface{}); ok {
			return fmt.Sprintf("%v", process["apply_id"])
		}
	}

	return ""
}

func (o *outbox) resetFailedTaskStatus(tasksIDMap map[string]string) {
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8",
	}

	for key, val := range tasksIDMap {
		url := fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/continue/%s", o.config.ContentAutomation.PrivateHost, o.config.ContentAutomation.PrivatePort, key)
		payload := errors.NewIError(errors.InternalError, errors.ErrorDepencyService, map[string]interface{}{"err": val})
		payloadBytes, _ := json.Marshal(payload)
		o.httpClient.Post(context.Background(), url, headers, payloadBytes)
		time.Sleep(1 * time.Second)
	}
}
