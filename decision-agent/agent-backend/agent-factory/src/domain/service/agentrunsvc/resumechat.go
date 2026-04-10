package agentsvc

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/panichelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
)

// ResumeChat 恢复聊天（Session恢复）
func (agentSvc *agentSvc) ResumeChat(ctx context.Context, conversationID string) (chan []byte, error) {
	var err error

	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetConversationID(ctx, conversationID)

	otellog.LogDebug(ctx, "[ResumeChat] started")

	sessionInterface, ok := SessionMap.Load(conversationID)
	if !ok {
		otellog.LogError(ctx, fmt.Sprintf("[ResumeChat] conversation_id %s not found", conversationID), nil)
		agentSvc.logger.Errorf("[ResumeChat] conversation_id %s not found", conversationID)

		return nil, capierr.New400Err(ctx, "conversation_id not found")
	}

	session := sessionInterface.(*Session)
	session.Lock()
	defer session.Unlock()
	session.IsResuming = true
	// NOTE: 注册一个channel
	signal := make(chan struct{})
	if session.Signal == nil {
		session.Signal = signal
		SessionMap.Store(conversationID, session)
	} else {
		signal = session.Signal
	}

	channel := make(chan []byte)

	go func() {
		defer panichelper.Recovery(agentSvc.logger)
		defer close(channel)

		oldResp := []byte(`{}`)
		seq := new(int)
		*seq = 0

		sessionInterface, ok := SessionMap.Load(conversationID)
		if !ok {
			otellog.LogError(ctx, fmt.Sprintf("[ResumeChat] conversation_id %s not found", conversationID), nil)
			agentSvc.logger.Errorf("[ResumeChat] conversation_id %s not found", conversationID)

			return
		}

		session := sessionInterface.(*Session)
		signal = session.GetSignal()

		newResp, err := sonic.Marshal(session.GetTempMsgResp())
		if err != nil {
			otellog.LogError(ctx, fmt.Sprintf("[ResumeChat] marshal temp msg resp err: %v", err), err)
			agentSvc.logger.Errorf("[ResumeChat] marshal temp msg resp err: %v", err)

			return
		}
		// NOTE:先发送一次,把当前的tempMsgResp发送出去
		if newResp != nil {
			if err := StreamDiff(ctx, seq, oldResp, newResp, channel, 0); err != nil {
				otellog.LogError(ctx, fmt.Sprintf("[ResumeChat] stream diff err: %v", err), err)
				agentSvc.logger.Errorf("[ResumeChat] stream diff err: %v", err)

				return
			}
		}

		// NOTE: 监听信号，直到关闭
		for _, ok := <-signal; ok; _, ok = <-signal {
			// NOTE: 每当收到信号，就发送一条消息
			newResp, err := sonic.Marshal(session.GetTempMsgResp())
			if err != nil {
				otellog.LogError(ctx, fmt.Sprintf("[ResumeChat] marshal temp msg resp err: %v", err), err)
				agentSvc.logger.Errorf("[ResumeChat] marshal temp msg resp err: %v", err)

				break
			}

			if len(oldResp) == 0 {
				oldResp = newResp
			} else {
				if err := StreamDiff(ctx, seq, oldResp, newResp, channel, 0); err != nil {
					otellog.LogError(ctx, fmt.Sprintf("[ResumeChat] stream diff err: %v", err), err)
					agentSvc.logger.Errorf("[ResumeChat] stream diff err: %v", err)

					break
				}

				oldResp = newResp
			}
		}

		emitJSON(seq, channel, []interface{}{}, nil, "end")
	}()

	return channel, nil
}
