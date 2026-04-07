package mod

import (
	"context"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	normalizeutil "github.com/kweaver-ai/adp/autoflow/flow-automation/utils/normalize"
)

var tokenRequiredOps = map[string]bool{
	common.MannualTrigger:         true,
	common.CronTrigger:            true,
	common.CronWeekTrigger:        true,
	common.CronMonthTrigger:       true,
	common.CronCustomTrigger:      true,
	common.InternalToolPy3Opt:     true,
	common.AnyshareFileOCROpt:     true,
	common.AudioTransfer:          true,
	common.DocInfoEntityExtract:   true,
	common.OperatorTrigger:        true,
	common.DataflowDocTrigger:     true,
	common.DataflowUserTrigger:    true,
	common.DataflowDeptTrigger:    true,
	common.DataflowTagTrigger:     true,
	common.OpAnyDataCallAgent:     true,
	common.OpLLMChatCompletion:    true,
	common.OpContentEntity:        true,
	common.OpEcoconfigReindex:     true,
	common.DatabaseWriteOpt:       true,
	common.OpLLMReranker:          true,
	common.OpLLmEmbedding:         true,
	common.OpContentFileParse:     true,
	common.SubFlowCallDataflowOpt: true,
	common.SubFlowCallWorkflowOpt: true,
	"@custom":                     true,
	"@operator":                   true,
}

// TokenContext token获取的上下文信息
type TokenContext struct {
	TaskIns *entity.TaskInstance
	ActName string
	Ctx     context.Context
}

// TokenStrategy token获取策略接口
type TokenStrategy interface {
	// Match 判断是否匹配当前策略
	Match(tokenCtx *TokenContext) bool
	// GetToken 获取token
	GetToken(tokenCtx *TokenContext) (*entity.Token, error)
	// Name 策略名称
	Name() string
}

// SecurityPolicyTokenStrategy 安全策略Token
type SecurityPolicyTokenStrategy struct{}

func NewSecurityPolicyTokenStrategy() *SecurityPolicyTokenStrategy {
	return &SecurityPolicyTokenStrategy{}
}

func (s *SecurityPolicyTokenStrategy) Name() string {
	return "SecurityPolicyTokenStrategy"
}

func (s *SecurityPolicyTokenStrategy) Match(tokenCtx *TokenContext) bool {
	return tokenCtx.TaskIns.RelatedDagInstance.Trigger == entity.TriggerSecurityPolicy
}

func (s *SecurityPolicyTokenStrategy) GetToken(tokenCtx *TokenContext) (*entity.Token, error) {
	sourceID := s.extractSourceID(tokenCtx.TaskIns)

	if sourceID != "" {
		return s.getTokenByDocOwner(tokenCtx, sourceID)
	}

	return s.getDefaultAppToken(tokenCtx)
}

func (s *SecurityPolicyTokenStrategy) extractSourceID(taskIns *entity.TaskInstance) string {
	source, exists := taskIns.RelatedDagInstance.ShareData.Get("__source")
	if !exists {
		return ""
	}

	m, ok := source.(map[string]interface{})
	if !ok {
		return ""
	}

	if m["type"] == "file" || m["type"] == "folder" {
		if id, ok := m["id"].(string); ok {
			return id
		}
	}
	return ""
}

func (s *SecurityPolicyTokenStrategy) getTokenByDocOwner(tokenCtx *TokenContext, sourceID string) (*entity.Token, error) {
	docshareAdapters := drivenadapters.NewDocShare()
	owners, err := docshareAdapters.GetDocOwners(tokenCtx.Ctx, sourceID)
	if err != nil {
		return nil, ierrors.ParseError(err)
	}

	ownerMap := s.buildOwnerMap(owners)
	userIDs := s.collectUserIDs(tokenCtx.TaskIns, ownerMap)

	// 从后往前遍历，找到第一个是owner的用户获取token
	for i := len(userIDs) - 1; i >= 0; i-- {
		userID := userIDs[i]
		if _, ok := ownerMap[userID]; ok {
			tokenMgnt := NewTokenMgnt(userID)
			tokenInfo, err := tokenMgnt.GetUserToken("", userID)
			if err != nil {
				continue
			}
			return tokenInfo, nil
		}
	}

	return nil, ierrors.NewIError(ierrors.UnAuthorization, "", map[string]interface{}{"info": "no valid owner token found"})
}

func (s *SecurityPolicyTokenStrategy) buildOwnerMap(owners []drivenadapters.DocOwner) map[string]drivenadapters.Owner {
	ownerMap := make(map[string]drivenadapters.Owner)
	for _, docOwner := range owners {
		if docOwner.Owner.Type == common.APP.ToString() {
			continue
		}
		ownerMap[docOwner.Owner.ID] = docOwner.Owner
	}
	return ownerMap
}

func (s *SecurityPolicyTokenStrategy) collectUserIDs(taskIns *entity.TaskInstance, ownerMap map[string]drivenadapters.Owner) []string {
	userIDs := []string{taskIns.RelatedDagInstance.Vars["userid"].Value}

	// 添加owner的ID
	for ownerID := range ownerMap {
		userIDs = append([]string{ownerID}, userIDs...)
	}

	// 添加审批人ID
	s.appendAuditorIDs(taskIns, &userIDs)

	return userIDs
}

func (s *SecurityPolicyTokenStrategy) appendAuditorIDs(taskIns *entity.TaskInstance, userIDs *[]string) {
	workflowApprovalTaskIds, ok := taskIns.RelatedDagInstance.ShareData.Get(common.WorkflowApprovalTaskIds)
	if !ok {
		return
	}

	taskIds, ok := normalizeutil.AsSlice(workflowApprovalTaskIds)
	if !ok || len(taskIds) == 0 {
		return
	}

	for _, taskId := range taskIds {
		taskData, ok := taskIns.RelatedDagInstance.ShareData.Get("__" + taskId.(string))
		if !ok {
			continue
		}

		data, ok := taskData.(map[string]interface{})
		if !ok {
			continue
		}

		if allAuditorIds, ok := normalizeutil.AsSlice(data["all_auditor_ids"]); ok {
			for _, id := range allAuditorIds {
				if userID, ok := id.(string); ok {
					*userIDs = append(*userIDs, userID)
				}
			}
		}

		if finallyAuditorIds, ok := normalizeutil.AsSlice(data["finally_auditor_ids"]); ok {
			for _, id := range finallyAuditorIds {
				if userID, ok := id.(string); ok {
					*userIDs = append(*userIDs, userID)
				}
			}
		}
	}
}

func (s *SecurityPolicyTokenStrategy) getDefaultAppToken(tokenCtx *TokenContext) (*entity.Token, error) {
	userID := tokenCtx.TaskIns.RelatedDagInstance.Vars["userid"].Value
	tokenMgnt := NewTokenMgnt(userID)
	tokenInfo, err := tokenMgnt.GetAppToken()
	if err != nil {
		return nil, err
	}
	tokenInfo.IsApp = true
	return tokenInfo, nil
}

// DocLibQuotaScaleTokenStrategy 文档库扩容Token
type DocLibQuotaScaleTokenStrategy struct{}

func NewDocLibQuotaScaleTokenStrategy() *DocLibQuotaScaleTokenStrategy {
	return &DocLibQuotaScaleTokenStrategy{}
}

func (s *DocLibQuotaScaleTokenStrategy) Name() string {
	return "DocLibQuotaScaleTokenStrategy"
}

func (s *DocLibQuotaScaleTokenStrategy) Match(tokenCtx *TokenContext) bool {
	return tokenCtx.ActName == common.AnyshareDocLibQuotaScaleOpt
}

func (s *DocLibQuotaScaleTokenStrategy) GetToken(tokenCtx *TokenContext) (*entity.Token, error) {
	userID := tokenCtx.TaskIns.RelatedDagInstance.UserID
	if userID == "" {
		userID = tokenCtx.TaskIns.RelatedDagInstance.Vars["userid"].Value
	}

	tokenMgnt := NewTokenMgnt(userID)
	tokenInfo, err := tokenMgnt.GetAppToken()
	if err != nil {
		return nil, err
	}
	tokenInfo.IsApp = true
	return tokenInfo, nil
}

// DefaultUserTokenStrategy 默认策略Token
type DefaultUserTokenStrategy struct{}

func NewDefaultUserTokenStrategy() *DefaultUserTokenStrategy {
	return &DefaultUserTokenStrategy{}
}

func (s *DefaultUserTokenStrategy) Name() string {
	return "DefaultUserTokenStrategy"
}

func (s *DefaultUserTokenStrategy) Match(tokenCtx *TokenContext) bool {
	return true // 默认策略，始终匹配
}

func (s *DefaultUserTokenStrategy) GetToken(tokenCtx *TokenContext) (*entity.Token, error) {
	userID := tokenCtx.TaskIns.RelatedDagInstance.UserID
	if userID == "" {
		userID = tokenCtx.TaskIns.RelatedDagInstance.Vars["userid"].Value
	}

	appInfo := tokenCtx.TaskIns.RelatedDagInstance.AppInfo
	if appInfo.Enable {
		userID = common.NewConfig().OAuth.ClientID
	}

	tokenMgnt := NewTokenMgnt(userID)
	return tokenMgnt.GetUserToken("", userID)
}

// TokenStrategyManager token策略管理器
type TokenStrategyManager struct {
	strategies []TokenStrategy
	mu         sync.RWMutex
}

var tokenStrategyManager *TokenStrategyManager
var tsOnce sync.Once

// NewTokenStrategyManager 创建新的策略管理器
func NewTokenStrategyManager() *TokenStrategyManager {
	tsOnce.Do(func() {
		tokenStrategyManager = &TokenStrategyManager{
			strategies: make([]TokenStrategy, 0),
		}
		// 需要按照顺序添加先后顺序即代表策略判断
		tokenStrategyManager.Register(NewSecurityPolicyTokenStrategy())
		tokenStrategyManager.Register(NewDocLibQuotaScaleTokenStrategy())
		tokenStrategyManager.Register(NewDefaultUserTokenStrategy())
	})
	return tokenStrategyManager
}

// Register 注册策略
func (m *TokenStrategyManager) Register(strategy TokenStrategy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.strategies = append(m.strategies, strategy)
}

// GetToken 获取token（遍历策略链）
func (m *TokenStrategyManager) GetToken(tokenCtx *TokenContext) (*entity.Token, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tokenRequired := utils.IfNot(tokenRequiredOps[tokenCtx.ActName], true, false) || utils.IfNot(strings.Contains(tokenCtx.ActName, "@anyshare"), true, false)
	if !tokenRequired {
		return nil, nil
	}

	for _, strategy := range m.strategies {
		if strategy.Match(tokenCtx) {
			tokenInfo, err := strategy.GetToken(tokenCtx)
			if err != nil {
				return nil, ierrors.NewIError(
					ierrors.UnAuthorization,
					"",
					map[string]interface{}{
						"info":     err.Error(),
						"strategy": strategy.Name(),
					},
				)
			}
			if tokenInfo == nil {
				return nil, ierrors.NewIError(
					ierrors.UnAuthorization,
					"",
					map[string]interface{}{
						"info":     "get token failed",
						"strategy": strategy.Name(),
					},
				)
			}
			return tokenInfo, nil
		}
	}

	return nil, ierrors.NewIError(ierrors.UnAuthorization, "", map[string]interface{}{"info": "no matching token strategy found"})
}
