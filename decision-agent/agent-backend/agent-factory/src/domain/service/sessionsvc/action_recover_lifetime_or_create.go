package sessionsvc

import (
	"context"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/redisaccess/sessionredisacc"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/session/sessionreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/grhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/ctype"
)

// HandleRecoverLifetimeOrCreate 处理recover_lifetime_or_create操作
// 恢复session的有效期或创建新的session
// 返回: startTime-会话开始时间, ttl-剩余过期时间(秒), err-错误
func (s *sessionSvc) HandleRecoverLifetimeOrCreate(ctx context.Context, req sessionreq.ManageReq, visitorInfo *ctype.VisitorInfo, isTriggerCacheUpsert bool) (startTime int64, ttl int, err error) {
	defaultTTL := sessionredisacc.SessionTTL

	defer func() {
		if err != nil {
			startTime = 0
			ttl = 0
		}
	}()

	// 1. 刷新session的有效期
	exists, existingStartTime, err := s.sessionRedisAcc.RefreshSession(ctx, req.ConversationID, defaultTTL)
	if err != nil {
		s.logger.Errorf("[sessionSvc][HandleRecoverLifetimeOrCreate] failed to refresh session: %v", err)
		return
	}

	// 2. 判断session是否存在
	if exists {
		// session已存在，TTL已在RefreshSession中刷新，使用已有的startTime
		startTime = existingStartTime
	} else {
		// session不存在，创建新的session
		startTime = time.Now().Unix()

		_, err = s.sessionRedisAcc.SetSession(ctx, req.ConversationID, startTime, defaultTTL)
		if err != nil {
			s.logger.Errorf("[sessionSvc][HandleRecoverLifetimeOrCreate] failed to set session: %v", err)
			return
		}
	}

	// 实时获取TTL以确保准确性
	ttl, err = s.sessionRedisAcc.GetSessionTTL(ctx, req.ConversationID)
	if err != nil {
		s.logger.Errorf("[sessionSvc][HandleRecoverLifetimeOrCreate] failed to get session ttl: %v", err)
		return
	}

	// 3. 异步触发agent缓存的创建或更新
	if isTriggerCacheUpsert {
		_ctx := context.Background()

		grhelper.GoSafe(s.logger, func() error {
			return s.triggerAgentCacheUpsert(_ctx, req, visitorInfo)
		})
	}

	return
}
