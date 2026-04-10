package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	cmq "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/mq"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/store"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/admin"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/cronjob"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/inbox"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/mgnt"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/actions"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/download_pool"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	normalizeutil "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils/normalize"
)

// MQHandler 消息队列Handler接口
type MQHandler interface {
	// Subscribe 订阅nsq消息
	Subscribe()
}

var (
	mqOnce sync.Once
	mqh    MQHandler
)

type mqHandler struct {
	mongo     mod.Store
	mq        cmq.MQClient
	mgnt      mgnt.MgntHandler
	inbox     inbox.Handler
	admin     admin.AdminHandler
	userMgnt  drivenadapters.UserManagement
	taskCache rds.TaskCache
}

// UserMsg usermsg struct
type UserMsg struct {
	ID string `json:"id"`
}

// DocflowAuditResult 审核结果消息结构体
type DocflowAuditResult struct {
	ApplyID           string   `json:"apply_id"` // 申请ID，返回发起申请时传入的申请ID
	Result            string   `json:"result"`   // 审核结果，pass审核通过 reject审核不通过 undone撤销申请
	FinallyAuditorIDs []string `json:"finally_auditor_ids"`
	AllAuditorIDs     []string `json:"all_auditor_ids"`
}

// NewMQHandler 创建document 消息队列handler对象
func NewMQHandler() MQHandler {
	mqOnce.Do(func() {
		go cronjob.NewCronJob().Start()
		mqh = &mqHandler{
			mongo:     mod.GetStore(),
			mq:        cmq.NewMQClient(),
			mgnt:      mgnt.NewMgnt(),
			inbox:     inbox.NewInbox(),
			admin:     admin.NewAdmin(),
			userMgnt:  drivenadapters.NewUserManagement(),
			taskCache: rds.GetTaskCache(),
		}
	})
	return mqh
}

// Subscribe 订阅nsq消息
func (m *mqHandler) Subscribe() {
	var docTopics = []string{
		common.TopicFileCopy,
		common.TopicFileMove,
		common.TopicFileRemove,
		common.TopicFileUpload,
		common.TopicFileCreate,
		common.TopicFolderCreate,
		common.TopicFolderMove,
		common.TopicFolderRemove,
		common.TopicFolderCopy,
		common.TopicFileEdit,
		common.TopicFileReversion,
		common.TopicFileDelete,
		common.TopicFileRename,
		common.TopicFileRestore,
	}

	var userTopics = []string{
		common.TopicUserDelete,
		common.TopicUserFreeze,
		common.TopicUserCreate,
		common.TopicNameModify,
		common.TopicDeptDelete,
		common.TopicDeptCreate,
		common.TopicDeptMoved,
		common.TopicUserMoved,
		common.TopicUserAddDept,
		common.TopicUserRemoveDept,
	}

	var tagChangeTopics = []string{
		common.TopicTagTreeAdded,
		common.TopicTagTreeEdited,
		common.TopicTagTreeDeleted,
	}

	var topicMap = map[string]func([]byte) error{
		common.TopicTagTreeCreated:                                m.handleTagTreeCreateNotify(common.TopicTagTreeCreated),
		common.TopicWorkflowResult:                                m.handleAuditResult,
		common.TopicWorkflowAuditor:                               m.handleWorkflowAuditor,
		common.TopicWorkflowSecurityPolicyPermResult:              m.handleAuditResult,
		common.TopicWorkflowSecurityPolicyPermAuditor:             m.handleWorkflowAuditor,
		common.TopicWorkflowSecurityPolicyUploadResult:            m.handleAuditResult,
		common.TopicWorkflowSecurityPolicyUploadAuditor:           m.handleWorkflowAuditor,
		common.TopicWorkflowSecurityPolicyDeleteResult:            m.handleAuditResult,
		common.TopicWorkflowSecurityPolicyDeleteAuditor:           m.handleWorkflowAuditor,
		common.TopicWorkflowSecurityPolicyCopyResult:              m.handleAuditResult,
		common.TopicWorkflowSecurityPolicyCopyAuditor:             m.handleWorkflowAuditor,
		common.TopicWorkflowSecurityPolicyMoveResult:              m.handleAuditResult,
		common.TopicWorkflowSecurityPolicyMoveAuditor:             m.handleWorkflowAuditor,
		common.TopicWorkflowSecurityPolicyRenameResult:            m.handleAuditResult,
		common.TopicWorkflowSecurityPolicyRenameAuditor:           m.handleWorkflowAuditor,
		common.TopicWorkflowSecurityPolicyFolderPropertiesResult:  m.handleAuditResult,
		common.TopicWorkflowSecurityPolicyFolderPropertiesAuditor: m.handleWorkflowAuditor,
		common.TopicSecurityPolicyProcCreate:                      m.handleSecurityPolicyProcCreate,
		common.TopicSecurityPolicyProcStatus:                      m.handleSecurityPolicyProcStatus,
		common.TopicSecurityPolicyFlowDelete:                      m.handleSecurityPolicyFlowDelete,
		common.TopicDocSetGraphInfoResult:                         m.handleGraphInfoResult,
		common.TopicDatahubIndexNotify:                            m.handleDatahubIndexNotify,
		common.TopicFlowDelete:                                    m.handleFlowOperation(common.TopicFlowDelete),
		common.TopicFlowActive:                                    m.handleFlowOperation(common.TopicFlowActive),
		common.TopicFlowSuspended:                                 m.handleFlowOperation(common.TopicFlowSuspended),
		common.TopicDatahubIndexNotify:                            m.handleDatahubIndexNotify,
		common.TopicOperatorDelete:                                m.handleOperatorDelete,

		common.TopicContentPipelineFulltextResult:         m.handleContentPipelineFulltextResult,
		common.TopicContentPipelineDocFormatConvertResult: m.handleContentPipelineDocFormatConvertResult,
		common.TopicAsyncTaskResult:                       m.handleAsyncTaskResult,
		common.TopicFlowFileDownloadResult:                m.handleFlowFileDownloadResult,
	}

	for _, val := range docTopics {
		topic := val
		go func() {
			m.mq.Subscribe(topic, common.ChannelMessage, m.handleDocNotify(topic))
		}()
	}

	for _, val := range userTopics {
		topic := val
		go func() {
			m.mq.Subscribe(topic, common.ChannelMessage, m.handleUserInfoNotify(topic))
		}()
	}

	for _, val := range tagChangeTopics {
		topic := val
		go func() {
			m.mq.Subscribe(topic, common.ChannelMessage, m.handleTagInfoChangeNotify(topic))
		}()
	}

	for t, f := range topicMap {
		// 重新赋值防止闭包
		topic, fn := t, f
		go func() {
			m.mq.Subscribe(topic, common.ChannelMessage, fn)
		}()
	}
}

// handleDocNotify 文档事件结果通知nsq消息
func (m *mqHandler) handleDocNotify(topic string) func(message []byte) error {
	return func(message []byte) error {
		if len(message) == 0 {
			return nil
		}

		var err error
		ctx, span := trace.StartConsumerSpan(context.Background())
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		log := traceLog.WithContext(ctx)

		var msg common.DocMsg
		if err = json.Unmarshal(message, &msg); err != nil {
			log.Warnf("[handleNotify] unmarshal message error fail, detail: %s", err.Error())
			return nil
		}

		itemID := utils.GetDocCurID(msg.DocID)

		topics := []string{topic}
		// 如果是更新文件topic，则将上传文件topic和新建文件topic进行合并
		if topic == common.TopicFileEdit {
			topics = append(topics, common.TopicFileUpload, common.TopicFileCreate)
		}
		inMsgs, err := m.inbox.QueryMsgs(ctx, itemID, topics)
		if err != nil {
			log.Warnf("[handleNotify] query message fail, detail: %s", err.Error())
			return err
		}
		isMsgRepeated := false
		for _, inMsg := range inMsgs {
			if time.Now().UTC().Unix() < inMsg.CreatedAt+30 {
				isMsgRepeated = true
				break
			}
		}

		// 30s内重复的消息不处理
		if isMsgRepeated {
			return nil
		}

		err = m.inbox.SaveMsg(ctx, &msg, itemID, topic)
		if err != nil {
			log.Warnf("[handleNotify] save message fail, detail: %s", err.Error())
			return err
		}
		err = m.mgnt.HandleDocEvent(ctx, &msg, topic)
		if err != nil {
			log.Warnf("[handleNotify] handle message fail, detail: %s", err.Error())
			return err
		}

		return nil
	}
}

// handleUserInfoNotify 用户事件结果通知nsq消息
func (m *mqHandler) handleUserInfoNotify(topic string) func(message []byte) error {
	return func(message []byte) error {
		if len(message) == 0 {
			return nil
		}

		var err error
		ctx, span := trace.StartConsumerSpan(context.Background())
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		log := traceLog.WithContext(ctx)

		var msg common.UserInfoMsg
		if err = json.Unmarshal(message, &msg); err != nil {
			log.Warnf("[handleUserInfoNotify] unmarshal message error fail, detail: %s", err.Error())
			return nil
		}

		if topic == common.TopicNameModify && msg.Type == common.User.ToString() {
			m.admin.UpdateAdmin(ctx, msg.ID, msg.NewName)
		}

		err = m.mgnt.HandleUserInfoEvent(ctx, &msg, topic)
		if err != nil {
			log.Warnf("[handleNotify] handle message fail, detail: %s", err.Error())
			return err
		}

		return nil
	}
}

// handleTagInfoChangeNotify 标签树变化通知nsq消息
func (m *mqHandler) handleTagInfoChangeNotify(topic string) func(message []byte) error {
	return func(message []byte) error {
		if len(message) == 0 {
			return nil
		}

		var err error
		ctx, span := trace.StartConsumerSpan(context.Background())
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		log := traceLog.WithContext(ctx)

		var msg []common.TagInfo
		if err = json.Unmarshal(message, &msg); err != nil {
			log.Warnf("[handleTagInfoChangeNotify] unmarshal message error fail, detail: %s", err.Error())
			return nil
		}

		for _, tag := range msg {
			err = m.mgnt.HandleTagInfoChangeEvent(ctx, &tag, message, topic)
			if err != nil {
				log.Warnf("[handleTagInfoChangeNotify] handle message fail, detail: %s", err.Error())
				return err
			}
		}

		return nil
	}
}

// handleTagTreeCreateNotify 标签树创建通知nsq消息
func (m *mqHandler) handleTagTreeCreateNotify(topic string) func(message []byte) error {
	return func(message []byte) error {
		if len(message) == 0 {
			return nil
		}

		var err error
		ctx, span := trace.StartConsumerSpan(context.Background())
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		log := traceLog.WithContext(ctx)

		var msg common.TagInfo
		if err = json.Unmarshal(message, &msg); err != nil {
			log.Warnf("[handleTagTreeCreateNotify] unmarshal message error fail, detail: %s", err.Error())
			return nil
		}
		if msg.Name == "" {
			msg.Name = msg.Path
		}

		err = m.mgnt.HandleTagTreeCreateEvent(ctx, &msg, topic)
		if err != nil {
			log.Warnf("[handleTagTreeCreateNotify] handle message fail, detail: %s", err.Error())
			return err
		}

		return nil
	}
}

// handleAuditResult 审核结果通知nsq消息
func (m *mqHandler) handleAuditResult(message []byte) error {
	if len(message) == 0 {
		return nil
	}

	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var msg DocflowAuditResult
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Warnf("[handleAuditResult] unmarshal message error fail, detail: %s", err.Error())
		return nil
	}

	fmt.Println("get workflow audit result msg:", msg)
	if err := m.mgnt.ContinueBlockInstances(ctx, []string{msg.ApplyID}, map[string]interface{}{
		"result":              msg.Result,
		"all_auditor_ids":     msg.AllAuditorIDs,
		"finally_auditor_ids": msg.FinallyAuditorIDs,
	}, entity.TaskInstanceStatusSuccess); err != nil {
		log.Warnf("handle worflow audit result faild, detail: %s", err.Error())
		return err
	}

	return nil
}

// handleWorkflowAuditor 匹配审核员结果通知nsq消息
func (m *mqHandler) handleWorkflowAuditor(message []byte) error {
	if len(message) == 0 {
		return nil
	}

	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var msg mgnt.AuditorInfo
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Warnf("[handleWorkflowAuditor] unmarshal message error fail, detail: %s", err.Error())
		return nil
	}

	fmt.Println("get workflow auditor msg:", msg)

	if err := m.mgnt.HandleAuditorsMacth(ctx, &msg); err != nil {
		log.Warnf("handle worflow audit result faild, detail: %s", err.Error())
	}

	return nil
}

func (m *mqHandler) handleSecurityPolicyProcCreate(message []byte) error {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	err = common.JSONSchemaValid(message, "security-policy/proc-params.json")

	if err != nil {
		log.Warnf("[handleSecurityPolicyProcCreate] validate message failed, details: %s", err.Error())
		return err
	}

	var procParams mgnt.ProcParams
	err = json.Unmarshal(message, &procParams)

	if err != nil {
		log.Warnf("[handleSecurityPolicyProcCreate] unmarshal message failed, details: %s", err.Error())
		return err
	}

	_, err = m.mgnt.StartSecurityPolicyFlowProc(ctx, procParams)

	if err != nil {
		log.Warnf("[handleSecurityPolicyProcCreate] StartSecurityPolicyFlowProc failed, details: %s, procParams: %v", err.Error(), procParams)
		return err
	}

	return nil
}

type UpdateProcStatusParams struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (m *mqHandler) handleSecurityPolicyProcStatus(message []byte) error {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	err = common.JSONSchemaValid(message, "security-policy/update-proc-status-message.json")

	if err != nil {
		log.Warnf("[handleSecurityPolicyProcStatus] validate message failed, details: %s", err.Error())
		return err
	}

	var params UpdateProcStatusParams

	err = json.Unmarshal(message, &params)
	if err != nil {
		log.Warnf("[handleSecurityPolicyProcStatus] unmarshal message failed, details: %s", err.Error())
		return err
	}

	err = m.mgnt.StopSecurityPolicyFlowProc(ctx, params.ID, nil)

	if err != nil {
		log.Warnf("[handleSecurityPolicyProcCreate] StopSecurityPolicyFlowProc failed, details: %s, pid: %v", err.Error(), params.ID)
		return err
	}

	return nil
}

type DeleteFlowParams struct {
	ID string `json:"id"`
}

func (m *mqHandler) handleSecurityPolicyFlowDelete(message []byte) error {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	err = common.JSONSchemaValid(message, "security-policy/delete-flow-message.json")

	if err != nil {
		log.Warnf("[handleSecurityPolicyFlowDelete] validate message failed, details: %s", err.Error())
		return err
	}

	var params DeleteFlowParams

	err = json.Unmarshal(message, &params)
	if err != nil {
		log.Warnf("[handleSecurityPolicyFlowDelete] unmarshal message failed, details: %s", err.Error())
		return err
	}

	err = m.mgnt.DeleteSecurityPolicyFlow(ctx, params.ID, nil)

	if err != nil {
		log.Warnf("[handleSecurityPolicyFlowDelete] DeleteSecurityPolicyFlow failed, details: %s, pid: %v", err.Error(), params.ID)
		return err
	}

	return nil
}

func (m *mqHandler) handleGraphInfoResult(message []byte) error {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var res struct {
		OssID     string `json:"oss_id"`
		ObjectKey string `json:"object_key"`
		ObjectID  string `json:"object_id"`
		Version   string `json:"version"`
		Status    string `json:"status"`
		Message   string `json:"message"`
	}
	err = json.Unmarshal(message, &res)

	if err != nil {
		log.Warnf("handle docset.graph_info.result faild, detail: %s", err.Error())
		return err
	}

	redis := store.NewRedis()
	client := redis.GetClient()
	key := fmt.Sprintf("%s%s", entity.ContentEntityKeyPrefix, res.ObjectID)

	taskIDMap, err := client.HGetAll(ctx, key).Result()

	if err != nil {
		log.Warnf("handle docset.graph_info.result faild, detail: %s", err.Error())
		return err
	}

	if len(taskIDMap) == 0 {
		_ = client.Del(ctx, key).Err()
		return nil
	}

	var taskIDs = make([]string, 0)

	for id, _ := range taskIDMap {
		taskIDs = append(taskIDs, id)
	}

	taskInstances, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{IDs: taskIDs})

	if err != nil {
		log.Warnf("handle docset.graph_info.result ListTaskInstance faild, taskId: detail: %s", err.Error())
		return nil
	}

	var deleteKeys = make([]string, 0)
	efast := drivenadapters.NewEfast()
	for _, taskIns := range taskInstances {
		result, ok := normalizeutil.AsMap(taskIns.Results)
		if !ok {
			continue
		}
		if __subdocParams, ok := result["__subdocParams"].(string); ok {
			subdocParams := drivenadapters.DocSetSubdocParams{}
			err1 := json.Unmarshal([]byte(__subdocParams), &subdocParams)
			if err1 != nil {
				deleteKeys = append(deleteKeys, taskIns.ID)
				continue
			}

			res, err1 := efast.DocSetSubdoc(ctx, subdocParams, -1)

			if err1 != nil {
				deleteKeys = append(deleteKeys, taskIns.ID)
				if err2 := m.mgnt.ContinueBlockInstances(ctx, []string{taskIns.ID}, map[string]interface{}{
					"__subdocParams": "{}",
					"doc_id":         subdocParams.DocID,
					"rev":            subdocParams.Version,
					"status":         "error",
					"err_msg":        err1.Error(),
					"data":           "{}",
					"url":            "",
				}, entity.TaskInstanceStatusSuccess); err2 != nil {
					log.Warnf("handle docset.graph_info.result faild, taskId: %s, docId: %s, detail: %s", taskIns.ID, subdocParams.DocID, err2.Error())
				}
				continue
			}

			if res.Status == "processing" {
				continue
			}

			deleteKeys = append(deleteKeys, taskIns.ID)
			data := res.Data
			if data == nil || data == "" {
				data = "{}"
			}

			if err2 := m.mgnt.ContinueBlockInstances(ctx, []string{taskIns.ID}, map[string]interface{}{
				"__subdocParams": "{}",
				"doc_id":         res.DocID,
				"rev":            res.Rev,
				"status":         res.Status,
				"err_msg":        res.ErrMsg,
				"data":           data,
				"url":            res.Url,
			}, entity.TaskInstanceStatusSuccess); err2 != nil {
				log.Warnf("handle docset.graph_info.result faild, taskId: %s, docId: %s, detail: %s", taskIns.ID, res.DocID, err2.Error())
			}
		}
	}

	if len(deleteKeys) == len(taskIDMap) {
		_ = client.Del(ctx, key).Err()
	} else {
		_ = client.HDel(ctx, key, deleteKeys...).Err()
	}

	return nil
}

type TraceEventMQMessage struct {
	IndexName string `json:"index_name"` // 索引名称
	DocID     string `json:"doc_id"`     // 文档ID
	EventName string `json:"event_name"` // 事件名称
	EventID   string `json:"event_id"`   // 事件ID
	EventType int    `json:"event_type"` // 事件类型, 0:初始态事件，1：中间态事件，2：完成态事件
	Status    string `json:"status"`     // 事件状态
	ErrorMsg  string `json:"error_msg"`  // 错误信息
	EventTime int64  `json:"event_time"` // 事件时间
	CreatedAt int64  `json:"created_at"` // 创建时间
	UpdatedAt int64  `json:"updated_at"` // 更新时间
}

func (m *mqHandler) handleDatahubIndexNotify(message []byte) (err error) {
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() {
		trace.TelemetrySpanEnd(span, err)
	}()

	log := traceLog.WithContext(ctx)
	var res map[string]interface{}
	err = json.Unmarshal(message, &res)

	if err != nil {
		log.Warnf("handle %s faild, detail: %s", common.TopicDatahubIndexNotify, err.Error())
		return err
	}

	// 只处理索引构建结束事件
	if res["event_type"] != float64(2) {
		return
	}

	redis := store.NewRedis()
	client := redis.GetClient()
	key := fmt.Sprintf("%s%s", entity.EcoconfigReindexKeyPrefix, res["doc_id"])

	taskIDMap, err := client.HGetAll(ctx, key).Result()

	if err != nil {
		log.Warnf("handle %s faild, detail: %s", common.TopicDatahubIndexNotify, err.Error())
		return err
	}

	if len(taskIDMap) == 0 {
		_ = client.Del(ctx, key).Err()
		return nil
	}

	var taskIDs = make([]string, 0)

	for id, _ := range taskIDMap {
		taskIDs = append(taskIDs, id)
	}

	if err := m.mgnt.ContinueBlockInstances(ctx, taskIDs, res, entity.TaskInstanceStatusSuccess); err != nil {
		log.Warnf("handle %s faild, taskIds: %v, docId: %s, detail: %s",
			common.TopicDatahubIndexNotify, taskIDs, res["doc_id"], err.Error())
	}

	return
}

// FlowOperationParams 流程参数
type FlowOperationParams struct {
	IDs    []string `json:"ids"`
	UserID string   `json:"userid"`
}

// handleFlowOperation 通用流程操作处理
func (m *mqHandler) handleFlowOperation(topic string) func(message []byte) error {
	return func(message []byte) error {
		var err error
		ctx, span := trace.StartConsumerSpan(context.Background())
		defer func() { trace.TelemetrySpanEnd(span, err) }()
		log := traceLog.WithContext(ctx)

		var params FlowOperationParams
		err = json.Unmarshal(message, &params)
		if err != nil {
			log.Warnf("[handleFlowOperation] unmarshal message failed for topic %s, details: %s", topic, err.Error())
			return err
		}

		userInfo, err := m.userMgnt.GetUserInfo(params.UserID)
		if err != nil {
			log.Warnf("[handleFlowOperation] GetUserInfo failed, details: %s, id: %s", err.Error(), params.UserID)
			return err
		}

		// 根据不同的topic执行不同的操作
		switch topic {
		case common.TopicFlowDelete:
			for _, id := range params.IDs {
				if err := m.mgnt.DeleteDagByID(ctx, id, "", &userInfo); err != nil {
					log.Warnf("[handleFlowOperation] DeleteDagByID failed, details: %s, id: %s", err.Error(), id)
					return err
				}
			}
		case common.TopicFlowActive:
			for _, id := range params.IDs {
				status := string(entity.DagStatusNormal)
				if err := m.mgnt.UpdateDag(ctx, id, &mgnt.OptionalUpdateDagReq{
					Status: &status,
				}, &userInfo); err != nil {
					log.Warnf("[handleFlowOperation] ActivateDag failed, details: %s, id: %s", err.Error(), id)
					return err
				}
			}
		case common.TopicFlowSuspended:
			for _, id := range params.IDs {
				status := string(entity.DagStatusStopped)
				if err := m.mgnt.UpdateDag(ctx, id, &mgnt.OptionalUpdateDagReq{
					Status: &status,
				}, &userInfo); err != nil {
					log.Warnf("[handleFlowOperation] DeactivateDag failed, details: %s, id: %s", err.Error(), id)
					return err
				}
			}
		}

		return nil
	}
}

// AgentOperatorMsg 算子平台算子删除消息体
type AgentOperatorMsg struct {
	OperatorID   string         `json:"operator_id"`
	OperatorType string         `json:"operator_type"`
	ExtendInfo   map[string]any `json:"extend_info"`
}

func (m *mqHandler) handleOperatorDelete(message []byte) error {
	if len(message) == 0 {
		return nil
	}

	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var msg AgentOperatorMsg
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Warnf("[handleOperatorDelete] unmarshal message error fail, detail: %s", err.Error())
		return nil
	}

	fmt.Println("get combo operator msg:", msg)

	if msg.OperatorType != "composite" {
		return nil
	}

	err = m.mgnt.DeleteComboOperator(ctx, msg.OperatorID)
	if err != nil {
		log.Warnf("[handleOperatorDelete] DeleteComboOperator error fail, detail: %s", err.Error())
		return err
	}

	return nil
}

type ContentPipelineValue[T any] struct {
	Type  string `json:"type,omitempty"`
	Value T      `json:"value,omitempty"`
}

type ContentPipelineJob[T any] struct {
	ID       string                   `json:"id,omitempty"`
	Passback string                   `json:"passback,omitempty"`
	Source   *ContentPipelineValue[T] `json:"source,omitempty"`
}

type ContentPipelineResult[T any, R any] struct {
	Code        string                   `json:"code,omitempty"`
	Description string                   `json:"description,omitempty"`
	FinishedAt  int64                    `json:"finished_at,omitempty"`
	Job         *ContentPipelineJob[T]   `json:"job,omitempty"`
	Result      *ContentPipelineValue[R] `json:"result,omitempty"`
	Task        string                   `json:"task,omitempty"`
}

type FullTextResultValue struct {
	OssID     string `json:"oss_id,omitempty"`
	ObjectKey string `json:"object_key,omitempty"`
	Status    string `json:"status,omitempty"`
	Size      int64  `json:"size,omitempty"`
}

func (m *mqHandler) handleContentPipelineFulltextResult(message []byte) error {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() {
		trace.TelemetrySpanEnd(span, err)
	}()

	log := traceLog.WithContext(ctx)
	log.Warnf("[handleContentPipelineFulltextResult] message: %s", string(message))
	var result ContentPipelineResult[string, *FullTextResultValue]

	err = json.Unmarshal(message, &result)

	if err != nil {
		log.Warnf("[handleContentPipelineFulltextResult] failed, err %s", err.Error())
		return err
	}

	if result.Job == nil || !strings.HasPrefix(result.Job.Passback, "automation:") {
		return nil
	}

	hash := result.Job.Passback[len("automation:"):]

	taskCacheItem, err := m.getTaskCacheByHashWithRetry(ctx, hash, 3)

	if err != nil {
		log.Warnf("[handleContentPipelineFulltextResult] getTaskCacheByHashWithRetry failed, err %s, hash %s", err.Error(), hash)
		return err
	}

	if taskCacheItem == nil {
		log.Warnf("[handleContentPipelineFulltextResult] task cache not found, hash %s", hash)
		return nil
	}

	taskStatus := entity.TaskInstanceStatusSuccess
	taskResult := map[string]any{}
	update := &rds.TaskCacheItem{
		Hash:       hash,
		Status:     rds.TaskStatusSuccess,
		ModifyTime: time.Now().Unix(),
		Ext:        ".txt",
	}

	if result.Code != "" {
		update.Status = rds.TaskStatusFailed
		update.ErrMsg = result.Description

		taskStatus = entity.TaskInstanceStatusFailed
		taskResult["code"] = result.Code
		taskResult["description"] = result.Description
	} else if result.Result == nil || result.Result.Value == nil {
		log.Warnf("[handleContentPipelineFulltextResult] failed, result.Result.Value is nil")
		update.Status = rds.TaskStatusFailed
		update.ErrMsg = "result is empty"
		taskStatus = entity.TaskInstanceStatusFailed
		taskResult["description"] = "result is empty"
	} else {
		og := drivenadapters.NewOssGateWay()
		servicePrefix := false
		reader := og.NewReader(result.Result.Value.OssID, result.Result.Value.ObjectKey, drivenadapters.OssOpt{StoragePrifix: &servicePrefix})
		taskResult["url"], _ = reader.Url(ctx)
		taskResult["text"], _ = reader.Text(ctx)
	}

	err = m.taskCache.Update(ctx, update)

	if err != nil {
		log.Warnf("[handleContentPipelineFulltextResult] update task cache failed, err %s, hash %s", err.Error(), hash)
		return err
	}

	taskIns, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
		Hash: hash,
	})

	if err != nil {
		log.Warnf("[handleContentPipelineFulltextResult] ListTaskInstance failed, err %s, hash %s", err.Error(), hash)
		return err
	}

	if len(taskIns) == 0 {
		log.Infof("[handleContentPipelineFulltextResult] task not found, hash %s", hash)
		return nil
	}

	for _, ins := range taskIns {
		err := m.mgnt.ContinueBlockInstances(ctx, []string{ins.ID}, taskResult, taskStatus)
		if err != nil {
			log.Warnf("[handleContentPipelineFulltextResult] failed, id %s, err %s", ins.ID, err.Error())
		}
	}

	return nil
}

func (m *mqHandler) handleContentPipelineDocFormatConvertResult(message []byte) error {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() {
		trace.TelemetrySpanEnd(span, err)
	}()

	log := traceLog.WithContext(ctx)
	log.Warnf("[handleContentPipelineDocFormatConvertResult] message: %s", string(message))
	var result ContentPipelineResult[string, any]

	err = json.Unmarshal(message, &result)

	if err != nil {
		log.Warnf("[handleContentPipelineDocFormatConvertResult] failed, err %s", err.Error())
		return err
	}

	if result.Job == nil || !strings.HasPrefix(result.Job.Passback, "automation:") {
		return nil
	}

	hash := result.Job.Passback[len("automation:"):]
	log.Warnf("[handleContentPipelineDocFormatConvertResult] hash %s, %s", hash, string(message))

	taskCacheItem, err := m.getTaskCacheByHashWithRetry(ctx, hash, 3)

	if err != nil {
		log.Warnf("[handleContentPipelineDocFormatConvertResult] getTaskCacheByHashWithRetry failed, err %s, hash %s", err.Error(), hash)
		return err
	}

	if taskCacheItem == nil {
		log.Warnf("[handleContentPipelineDocFormatConvertResult] task cache not found, hash %s", hash)
		return nil
	}

	taskStatus := entity.TaskInstanceStatusSuccess
	taskResult := map[string]any{}
	update := &rds.TaskCacheItem{
		Hash:       hash,
		Status:     rds.TaskStatusSuccess,
		ModifyTime: time.Now().Unix(),
		Ext:        ".txt",
	}

	if result.Code != "" {
		update.Status = rds.TaskStatusFailed
		update.ErrMsg = result.Description

		taskStatus = entity.TaskInstanceStatusFailed
		taskResult["code"] = result.Code
		taskResult["description"] = result.Description
	} else if result.Result == nil || result.Result.Value == nil {
		log.Warnf("[handleContentPipelineDocFormatConvertResult] failed, result.Result.Value is nil")
		update.Status = rds.TaskStatusFailed
		update.ErrMsg = "result is empty"
		taskStatus = entity.TaskInstanceStatusFailed
		taskResult["description"] = "result is empty"
	} else {
		ossPath, ok := result.Result.Value.(string)
		if !ok || ossPath == "" {
			log.Warnf("[handleContentPipelineDocFormatConvertResult] failed, ossPath %s", ossPath)
			update.Status = rds.TaskStatusFailed
			update.ErrMsg = fmt.Sprintf("ossPath %s is invalid", ossPath)
			taskStatus = entity.TaskInstanceStatusFailed
			taskResult["description"] = fmt.Sprintf("ossPath %s is invalid", ossPath)
			return nil
		} else {
			og := drivenadapters.NewOssGateWay()
			servicePrefix := false
			parts := strings.SplitN(ossPath, "/", 2)
			ossID, objectKey := parts[0], parts[1]
			reader := og.NewReader(ossID, objectKey, drivenadapters.OssOpt{StoragePrifix: &servicePrefix})
			taskResult["url"], _ = reader.Url(ctx)
		}
	}

	err = m.taskCache.Update(ctx, update)

	if err != nil {
		return err
	}

	taskIns, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
		Hash: hash,
	})

	if err != nil {
		log.Warnf("[handleContentPipelineDocFormatConvertResult] ListTaskInstance failed, err %s, hash %s", err.Error(), hash)
		return err
	}

	if len(taskIns) == 0 {
		log.Infof("[handleContentPipelineDocFormatConvertResult] task not found, hash %s", hash)
		return nil
	}

	for _, ins := range taskIns {
		err := m.mgnt.ContinueBlockInstances(ctx, []string{ins.ID}, taskResult, taskStatus)

		if err != nil {
			log.Warnf("[handleContentPipelineDocFormatConvertResult] failed, id %s, err %s", ins.ID, err.Error())
		}
	}

	return nil
}

func (m *mqHandler) getTaskCacheByHashWithRetry(ctx context.Context, hash string, maxRetryTimes int) (*rds.TaskCacheItem, error) {

	for retry := 0; retry < maxRetryTimes; retry++ {
		taskCacheItem, err := m.taskCache.GetByHash(ctx, hash)
		if err != nil {
			return nil, err
		}

		if taskCacheItem != nil {
			return taskCacheItem, nil
		}

		if retry == maxRetryTimes-1 {
			break
		}

		retryTime := retry + 1
		time.Sleep(time.Duration(retryTime) * time.Second)
	}

	return nil, nil
}

// 统一的异步任务结果处理器
func (m *mqHandler) handleAsyncTaskResult(message []byte) error {
	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	log := traceLog.WithContext(ctx)

	// 解析统一的消息结构
	var notification actions.AsyncTaskResultNotification
	err = json.Unmarshal(message, &notification)
	if err != nil {
		log.Warnf("[handleAsyncTaskResult] Unmarshal failed, err %s", err.Error())
		return err
	}

	log.Infof("[handleAsyncTaskResult] Received notification, hash: %s, taskType: %s, status: %s",
		notification.Hash, notification.TaskType, notification.Status)

	// 根据状态确定任务实例状态和结果
	taskStatus := entity.TaskInstanceStatusSuccess
	taskResult := map[string]any{}

	if notification.Status == "failed" {
		taskStatus = entity.TaskInstanceStatusFailed
		taskResult["error"] = notification.Error
	} else {
		taskCacheItem, err := m.getTaskCacheByHashWithRetry(ctx, notification.Hash, 3)
		if err != nil {
			log.Warnf("[handleAsyncTaskResult] getTaskCacheByHashWithRetry failed, err %s, hash %s", err.Error(), notification.Hash)
			taskStatus = entity.TaskInstanceStatusFailed
			taskResult["error"] = err.Error()
		} else if taskCacheItem == nil {
			log.Warnf("[handleAsyncTaskResult] task cache not found, hash %s", notification.Hash)
			taskStatus = entity.TaskInstanceStatusFailed
			taskResult["error"] = "task cache not found"
		} else {
			loader := &actions.DefaultTaskResultLoader{}
			taskResult, err = loader.LoadResult(ctx, taskCacheItem)
			if err != nil {
				log.Warnf("[handleAsyncTaskResult] LoadResult failed, err %s, hash %s", err.Error(), notification.Hash)
				taskStatus = entity.TaskInstanceStatusFailed
				taskResult["error"] = err.Error()
			}
		}
	}

	taskIns, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
		Hash:   notification.Hash,
		Status: []entity.TaskInstanceStatus{entity.TaskInstanceStatusBlocked},
	})

	if err != nil {
		log.Warnf("[handleAsyncTaskResult] ListTaskInstance failed, hash: %s, err: %s", notification.Hash, err.Error())
		return err
	}

	if len(taskIns) == 0 {
		log.Infof("[handleAsyncTaskResult] No blocked instances found for hash: %s, taskType: %s, already processed or not exist",
			notification.Hash, notification.TaskType)
		return nil
	}

	// 恢复所有阻塞的任务实例
	for _, ins := range taskIns {
		err := m.mgnt.ContinueBlockInstances(ctx, []string{ins.ID}, taskResult, taskStatus)
		if err != nil {
			log.Warnf("[handleAsyncTaskResult] ContinueBlockInstances failed, id %s, taskType: %s, err %s",
				ins.ID, notification.TaskType, err.Error())
		} else {
			log.Infof("[handleAsyncTaskResult] Successfully continued instance %s, taskType: %s",
				ins.ID, notification.TaskType)
		}
	}

	return nil
}

// handleFlowFileDownloadResult 处理文件下载结果
func (m *mqHandler) handleFlowFileDownloadResult(message []byte) error {
	if len(message) == 0 {
		return nil
	}

	var err error
	ctx, span := trace.StartConsumerSpan(context.Background())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var result download_pool.FlowFileDownloadResult
	if err = json.Unmarshal(message, &result); err != nil {
		log.Warnf("[handleFlowFileDownloadResult] unmarshal message error: %s", err.Error())
		return nil
	}

	log.Infof("[handleFlowFileDownloadResult] Received result for file %d, status: %s",
		result.FileID, result.Status)

	// 查询恢复记录
	resumeList, err := rds.GetFlowTaskResumeDao().List(ctx, &rds.FlowTaskResumeQueryOptions{
		ResourceType: "file",
		ResourceID:   &result.FileID,
		Limit:        10,
	})
	if err != nil {
		log.Warnf("[handleFlowFileDownloadResult] query task_resume failed: %s", err.Error())
		return err
	}

	if len(resumeList) == 0 {
		log.Infof("[handleFlowFileDownloadResult] No blocked instances found for file %d", result.FileID)
		return nil
	}

	// 确定任务状态和输出
	var taskStatus entity.TaskInstanceStatus
	var taskResult map[string]interface{}

	if result.Status == "success" && result.NodeOutput != nil {
		taskStatus = entity.TaskInstanceStatusSuccess
		taskResult = result.NodeOutput
	} else {
		taskStatus = entity.TaskInstanceStatusFailed
		taskResult = map[string]interface{}{
			"error":      result.ErrorMsg,
			"error_code": result.ErrorCode,
			"status":     "failed",
		}
	}

	// 恢复所有阻塞的任务实例
	for _, resume := range resumeList {
		err := m.mgnt.ContinueBlockInstances(ctx, []string{resume.TaskInstanceID}, taskResult, taskStatus)
		if err != nil {
			log.Warnf("[handleFlowFileDownloadResult] ContinueBlockInstances failed, task_instance_id: %s, err: %s",
				resume.TaskInstanceID, err.Error())
			continue
		}

		// 恢复成功后删除恢复记录
		if delErr := rds.GetFlowTaskResumeDao().Delete(ctx, resume.ID); delErr != nil {
			log.Warnf("[handleFlowFileDownloadResult] delete task_resume failed: %s", delErr.Error())
		}

		log.Infof("[handleFlowFileDownloadResult] Successfully continued instance %s for file %d",
			resume.TaskInstanceID, result.FileID)
	}

	return nil
}
