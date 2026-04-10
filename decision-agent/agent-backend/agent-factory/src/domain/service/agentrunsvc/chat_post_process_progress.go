package agentsvc

import (
	"context"
	"fmt"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/conversationmsgvo"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
)

// NOTE: key 为assistantMessageID，value 为bool ,判断是否已经获取过中断前的progress
var isInterruptPreProgressGetMap sync.Map = sync.Map{}

func (agentSvc *agentSvc) handleProgressOld(ctx context.Context, req *agentreq.ChatReq, progresses []*agentrespvo.Progress, chunkIndex int) ([]*agentrespvo.Progress, error) {
	if chunkIndex == 0 {
		ctx, _ = oteltrace.StartInternalSpan(ctx)
		defer oteltrace.EndSpan(ctx, nil)
		oteltrace.SetAttributes(ctx,
			attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
			attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
			attribute.String(otelconst.AttrUserID, req.UserID),
			attribute.String("stream.chunk_position", "first"),
		)
		oteltrace.SetConversationID(ctx, req.ConversationID)
	}

	setInterface, _ := progressSet.Load(req.AssistantMessageID)

	// 1. 初始化 set
	set, _ := setInterface.(map[string]bool)
	if set == nil {
		set = make(map[string]bool)
		progressSet.Store(req.AssistantMessageID, set)
		agentSvc.logger.Infof("[handleProgress] progressSet is nil, store new set for assistantMessageID: %s", req.AssistantMessageID)
	}

	var currentProgress *agentrespvo.Progress

	// 2. 遍历 progresses
	for _, progress := range progresses {
		if progress.Status == "completed" || progress.Status == "failed" {
			if _, ok := set[progress.ID]; !ok {
				if v, ok := progressMap.Load(req.AssistantMessageID); !ok {
					progressMap.Store(req.AssistantMessageID, []*agentrespvo.Progress{progress})
				} else {
					progressMap.Store(req.AssistantMessageID, append(v.([]*agentrespvo.Progress), progress))
				}

				set[progress.ID] = true
			}
		} else if progress.Status == "processing" {
			currentProgress = progress
		}
	}

	// 3. NOTE： 如果是中断，还需要将中断前的结果拿到并拼接
	ans, err := agentSvc.forResumeInterrupt(ctx, req)
	if err != nil {
		return nil, err
	}

	// 4. append
	if v, ok := progressMap.Load(req.AssistantMessageID); ok {
		ans = append(ans, v.([]*agentrespvo.Progress)...)
	}

	// 5. append currentProgress
	if currentProgress != nil {
		ans = append(ans, currentProgress)
	}

	return ans, nil
}

func (agentSvc *agentSvc) handleProgress(ctx context.Context, req *agentreq.ChatReq, progresses []*agentrespvo.Progress, chunkIndex int) (newPgs []*agentrespvo.Progress, err error) {
	if chunkIndex == 0 {
		ctx, _ = oteltrace.StartInternalSpan(ctx)
		defer oteltrace.EndSpan(ctx, nil)
		oteltrace.SetAttributes(ctx,
			attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
			attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
			attribute.String(otelconst.AttrUserID, req.UserID),
			attribute.String("stream.chunk_position", "first"),
		)
		oteltrace.SetConversationID(ctx, req.ConversationID)
	}

	aMsgID := req.AssistantMessageID

	setInterface, _ := progressSet.Load(aMsgID)

	// 1. 初始化 set
	set, _ := setInterface.(map[string]bool)
	if set == nil {
		set = make(map[string]bool)
		progressSet.Store(aMsgID, set)
		agentSvc.logger.Infof("[handleProgress] progressSet is nil, store new set for assistantMessageID: %s", aMsgID)
	}

	// 2. NOTE： 如果是中断，还需要将中断前的结果拿到并拼接
	prePgs, err := agentSvc.forResumeInterrupt(ctx, req)
	if err != nil {
		return
	}

	pgs := append(prePgs, progresses...)

	var currentProgress *agentrespvo.Progress

	// 3. 遍历 progresses
	for _, pg := range pgs {
		// DEBUG: 记录 _evidence 字段状态
		if pg.Evidence != nil && len(pg.Evidence) > 0 {
			agentSvc.logger.Infof("[handleProgress] Input progress has _evidence: id=%s, stage=%s, count=%d", pg.ID, pg.Stage, len(pg.Evidence))
		}

		// fmt.Printf("pid: %s,status: %s\n", pg.ID, pg.Status)
		if _, exist := set[pg.ID]; exist {
			continue
		}

		if pg.Status == "completed" || pg.Status == "failed" || pg.Status == "skipped" {
			if v, _exist := progressMap.Load(aMsgID); !_exist {
				progressMap.Store(aMsgID, []*agentrespvo.Progress{pg})
			} else {
				progressMap.Store(aMsgID, append(v.([]*agentrespvo.Progress), pg))
			}

			set[pg.ID] = true
		} else if pg.Status == "processing" {
			currentProgress = pg
		}
	}

	// 4. append
	if v, ok := progressMap.Load(aMsgID); ok {
		newPgs = append(newPgs, v.([]*agentrespvo.Progress)...)
	}

	// 5. append currentProgress
	if currentProgress != nil {
		newPgs = append(newPgs, currentProgress)
	}

	// DEBUG: 检查输出中的 _evidence 字段
	evidenceCount := 0
	for _, pg := range newPgs {
		if pg.Evidence != nil && len(pg.Evidence) > 0 {
			evidenceCount++
			agentSvc.logger.Infof("[handleProgress] Output progress has _evidence: id=%s, stage=%s, count=%d", pg.ID, pg.Stage, len(pg.Evidence))
		}
	}
	if evidenceCount > 0 {
		agentSvc.logger.Infof("[handleProgress] ✅ Output contains %d progress(es) with _evidence", evidenceCount)
	} else {
		agentSvc.logger.Infof("[handleProgress] ℹ️ Output contains NO progress with _evidence (total=%d)", len(newPgs))
	}

	return
}

func (agentSvc *agentSvc) forResumeInterrupt(ctx context.Context, req *agentreq.ChatReq) (ans []*agentrespvo.Progress, err error) {
	ans = make([]*agentrespvo.Progress, 0)

	if req.InterruptedAssistantMsgID != "" {
		// 0. 检查是否已经获取过中断前的progress
		if _, ok := isInterruptPreProgressGetMap.Load(req.AssistantMessageID); ok {
			return
		}

		var assistantMsgPO *dapo.ConversationMsgPO

		// 1. 获取中断前的消息
		assistantMsgPO, err = agentSvc.conversationMsgRepo.GetByID(ctx, req.InterruptedAssistantMsgID)
		if err != nil {
			err = errors.Wrapf(err, "[handleProgress] get interrupted progress err")
			return
		}

		// 2. 初始化 content 以避免 nil
		content := conversationmsgvo.AssistantContent{
			MiddleAnswer: &conversationmsgvo.MiddleAnswer{}, // 新增：初始化 MiddleAnswer
		}

		// 3. 得到 中断前的消息的content
		// NOTE: 不能将空字符串反序列化，否则会报错
		if assistantMsgPO.Content != nil && *assistantMsgPO.Content != "" {
			err = sonic.Unmarshal([]byte(*assistantMsgPO.Content), &content)
			if err != nil {
				otellog.LogWarn(ctx, fmt.Sprintf("[handleProgress] unmarshal assistant content error, id: %s, err: %v", req.InterruptedAssistantMsgID, err))
				err = errors.Wrapf(err, "[handleProgress] unmarshal assistant content error, id: %s, err: %v", req.InterruptedAssistantMsgID, err)

				return
			}
		}

		// 4. 将中断前的消息的progress append到当前ans
		if content.MiddleAnswer != nil {
			ans = append(ans, content.MiddleAnswer.Progress...)
		} else {
			agentSvc.logger.Warnf("[handleProgress] skipped appending progress for interrupted msg %s: MiddleAnswer is nil", req.InterruptedAssistantMsgID)
		}

		// 5. 标记已获取过中断前的progress
		isInterruptPreProgressGetMap.Store(req.AssistantMessageID, true)
	}

	return
}
