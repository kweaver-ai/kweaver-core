package alarm

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"

	// "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters/thirft" // disabled: go-lib dependency
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	commonLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/log"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

const (
	alarmFrequency  = 1
	GroupIDNotFound = float64(400019003)
	UserIDNotFound  = float64(400019001)
)

type AlarmTask struct {
	RuleID     string       `json:"rule_id"`
	AlertUsers []*AlertUser `json:"alert_users"`
	DagIDs     []string     `json:"dag_ids"`
	Frequency  float32      `json:"frequency"`
	Threshold  int          `json:"threshold"`
}

type RestAlarmTask struct {
	RuleID     string       `json:"rule_id"`
	AlertUsers []*AlertUser `json:"alert_users"`
	Dags       []*Dag       `json:"dags"`
	Frequency  float32      `json:"frequency"`
	Threshold  int          `json:"threshold"`
}

type Dag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TriggerAlarmTask struct {
	TriggerTime int64
	AlarmTask   *AlarmTask
}

type AlertUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ErrorInfo 失败信息及失败次数
type ErrorInfo struct {
	DagID       string
	Reason      string
	FailedCount int64
}

func (rt *RestAlarmTask) ToAlarmTask() *AlarmTask {
	var dagIDs []string
	for _, val := range rt.Dags {
		dagIDs = append(dagIDs, val.ID)
	}

	return &AlarmTask{
		RuleID:     rt.RuleID,
		AlertUsers: rt.AlertUsers,
		DagIDs:     dagIDs,
		Frequency:  rt.Frequency,
		Threshold:  rt.Threshold,
	}
}

// Alarm 告警任务接口
type Alarm interface {
	ModifyAlarmRule(ctx context.Context, task *AlarmTask) (string, error)
	ListAlarmRule(ctx context.Context, page, limit int64) ([]*RestAlarmTask, error)
	GetAlarmRule(ctx context.Context, ruleID string) (*RestAlarmTask, error)
	ErrorAlarm(ctx context.Context)
}

type alarm struct {
	log          commonLog.Logger
	config       *common.Config
	store        mod.Store
	alarmRuleDao rds.AlarmRuleDao
	// sharemgnt    thirft.ShareMgnt // disabled: go-lib dependency
	userMgnt drivenadapters.UserManagement
	// 告警任务
	task chan *TriggerAlarmTask
}

var (
	aOnce sync.Once
	a     Alarm
)

// NewAlarm 实例化
func NewAlarm() Alarm {
	aOnce.Do(func() {
		a = &alarm{
			log:          commonLog.NewLogger(),
			config:       common.NewConfig(),
			store:        mod.GetStore(),
			alarmRuleDao: rds.GetAlarmRuleDao(),
			// sharemgnt:    thirft.NewShareMgnt(), // disabled: go-lib dependency
			userMgnt: drivenadapters.NewUserManagement(),
			task:     make(chan *TriggerAlarmTask, 100),
		}
	})

	return a
}

// ModifyAlarmRule 变更告警规则
func (a *alarm) ModifyAlarmRule(ctx context.Context, task *AlarmTask) (string, error) {
	var groupIDs, userIDs []string
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	_log := traceLog.WithContext(ctx)

	// 判断Dag是否存在
	dags, err := a.store.ListDag(ctx, &mod.ListDagInput{DagIDs: task.DagIDs, Type: "all"})
	if err != nil {
		_log.Warnf("[logic.ModifyAlarmRule] ListDag err, detail: %s", err.Error())
		return "", err
	}

	if len(dags) != len(task.DagIDs) {
		var idMap = map[string]struct{}{}
		for _, dag := range dags {
			idMap[dag.ID] = struct{}{}
		}
		var ids []string
		for _, dagID := range task.DagIDs {
			if _, ok := idMap[dagID]; !ok {
				ids = append(ids, dagID)
			}
		}
		return "", errors.NewIError(errors.TaskNotFound, "", map[string]interface{}{"ids": ids})
	}

	// 获取dagID绑定过的告警规则ID列表
	DBAlarmRules, err := a.alarmRuleDao.ListAlarmRule(ctx, &rds.Options{
		SearchOptions: []*rds.SearchOption{
			{Col: "f_dag_id", Val: task.DagIDs, Condition: "IN"},
		},
	})
	if err != nil {
		_log.Warnf("[logic.ModifyAlarmRule] ListAlarmRule err, detail: %s", err.Error())
		return "", err
	}

	var ruleMap = map[string][]string{}
	if task.RuleID != "" {
		// 更新操作，校验更新的告警规则是否存在
		_, err := a.alarmRuleDao.GetAlarmRule(ctx, task.RuleID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return "", errors.NewIError(errors.AlarmRuleNotFound, "", nil)
			}
			_log.Warnf("[logic.ModifyAlarmRule] GetAlarmRule err, detail: %s", err.Error())
			return "", err
		}

		for _, val := range DBAlarmRules {
			ruleID := fmt.Sprintf("%v", val.RuleID)
			if ruleID != task.RuleID {
				if _, ok := ruleMap[ruleID]; !ok {
					ruleMap[ruleID] = []string{}
				}
				ruleMap[ruleID] = append(ruleMap[ruleID], fmt.Sprintf("%v", val.DagID))
			}
		}
	} else {
		// 创建操作
		if len(DBAlarmRules) != 0 {
			for _, val := range DBAlarmRules {
				ruleID := fmt.Sprintf("%v", val.RuleID)
				if _, ok := ruleMap[ruleID]; !ok {
					ruleMap[ruleID] = []string{}
				}
				ruleMap[ruleID] = append(ruleMap[ruleID], fmt.Sprintf("%v", val.DagID))
			}
		}
		ruleID, _ := utils.GetUniqueID()
		task.RuleID = fmt.Sprintf("%v", ruleID)
	}

	var detail = []interface{}{}
	for key, val := range ruleMap {
		detail = append(detail, map[string]interface{}{
			"rule_id": key,
			"dag_ids": val,
		})
	}
	if len(detail) > 0 {
		return "", errors.NewIError(errors.AlarmRuleAlreadyExists, "", detail)
	}

	for _, val := range task.AlertUsers {
		switch val.Type {
		case common.Group.ToString():
			groupIDs = append(groupIDs, val.ID)
		case common.User.ToString():
			userIDs = append(userIDs, val.ID)
		}
	}

	var params = map[string][]string{}
	if len(groupIDs) > 0 {
		params["group_ids"] = groupIDs
	}

	if len(userIDs) > 0 {
		params["user_ids"] = userIDs
	}

	namesInfo, err := a.userMgnt.BatchGetNames(params)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[logic.CreateAlarmRule] BatchGetNames err, detail: %s", err.Error())
		return task.RuleID, err
	}
	nameMap := namesInfo.ToMap(common.User.ToString(), common.Group.ToString())

	var DBAlarmTasks []*rds.AlarmRule
	var DBAlarmUsers []*rds.AlarmUser
	ruleIDInt64, _ := strconv.ParseUint(task.RuleID, 10, 64)
	for _, val := range task.DagIDs {
		id, _ := utils.GetUniqueID()
		dagIDInt64, _ := strconv.ParseUint(val, 10, 64)
		now := time.Now().UnixNano() / 1e6
		frequency := fmt.Sprintf("%.1f", task.Frequency)
		frequencyFloat, _ := strconv.ParseFloat(frequency, 64)
		frequencyFloat *= 60
		DBAlarmTask := &rds.AlarmRule{
			ID:        id,
			RuleID:    ruleIDInt64,
			DagID:     dagIDInt64,
			Frequency: int(frequencyFloat),
			Threshold: task.Threshold,
			CreatedAt: now,
		}
		DBAlarmTasks = append(DBAlarmTasks, DBAlarmTask)
	}

	for _, alterUser := range task.AlertUsers {
		uid, _ := utils.GetUniqueID()
		DBAlarmUser := &rds.AlarmUser{
			ID:       uid,
			RuleID:   ruleIDInt64,
			UserID:   alterUser.ID,
			UserName: nameMap[alterUser.ID],
			UserType: alterUser.Type,
		}
		DBAlarmUsers = append(DBAlarmUsers, DBAlarmUser)
	}

	err = a.alarmRuleDao.ModifyAlarmRule(ctx, task.RuleID, DBAlarmTasks, DBAlarmUsers)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[logic.ModifyAlarmRule] ModifyAlarmRule err, detail: %s", err.Error())
		return "", err
	}

	return task.RuleID, err
}

func (a *alarm) ListAlarmRule(ctx context.Context, page, limit int64) ([]*RestAlarmTask, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	_log := traceLog.WithContext(ctx)

	order := "desc"
	orderBy := "f_created_at"
	opt := &rds.Options{
		Order:   &order,
		OrderBy: &orderBy,
		Page:    &page,
		Limit:   &limit,
	}
	alarmRules, err := a.alarmRuleDao.ListAlarmRule(ctx, opt)
	if err != nil {
		_log.Warnf("[logic.ListAlarmRule] ListAlarmRule err, detail: %s", err.Error())
		return nil, err
	}

	var dagIDs []string
	for _, val := range alarmRules {
		dagIDs = append(dagIDs, fmt.Sprintf("%v", val.DagID))
	}

	dags, err := a.store.ListDag(ctx, &mod.ListDagInput{DagIDs: dagIDs, Type: "all"})
	if err != nil {
		_log.Warnf("[logic.ListAlarmRule] ListDag err, detail: %s", err.Error())
		return nil, err
	}

	var dagMap = map[string]string{}
	for _, val := range dags {
		dagMap[val.ID] = val.Name
	}

	var taskMap = map[string]*RestAlarmTask{}
	var ruleIDs []string
	for _, val := range alarmRules {
		ruleID := fmt.Sprintf("%v", val.RuleID)
		if _, ok := taskMap[ruleID]; !ok {
			taskMap[ruleID] = &RestAlarmTask{
				RuleID:     ruleID,
				AlertUsers: []*AlertUser{},
				Dags:       []*Dag{},
				Frequency:  float32(val.Frequency) / 60,
				Threshold:  val.Threshold,
			}
			ruleIDs = append(ruleIDs, ruleID)
		}

		dagID := fmt.Sprintf("%v", val.DagID)
		taskMap[ruleID].Dags = append(taskMap[ruleID].Dags, &Dag{ID: dagID, Name: dagMap[dagID]})
	}

	alarmUsers, err := a.alarmRuleDao.ListAlarmUser(ctx, &rds.Options{
		SearchOptions: []*rds.SearchOption{
			{Col: "f_rule_id", Val: ruleIDs, Condition: "IN"},
		},
	})
	if err != nil {
		_log.Warnf("[logic.ListAlarmRule] ListAlarmUser err, detail: %s", err.Error())
		return nil, err
	}

	var userMap = map[string][]*AlertUser{}
	for _, val := range alarmUsers {
		ruleID := fmt.Sprintf("%v", val.RuleID)
		userMap[ruleID] = append(userMap[ruleID], &AlertUser{ID: val.UserID, Name: val.UserName, Type: val.UserType})
	}

	var alarmTasks = make([]*RestAlarmTask, 0)
	for key, val := range taskMap {
		ruleID := fmt.Sprintf("%v", val.RuleID)
		taskMap[key].AlertUsers = userMap[ruleID]
		alarmTasks = append(alarmTasks, taskMap[key])
	}

	return alarmTasks, nil
}

func (a *alarm) GetAlarmRule(ctx context.Context, ruleID string) (*RestAlarmTask, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	_log := traceLog.WithContext(ctx)

	opt := &rds.Options{
		SearchOptions: []*rds.SearchOption{
			{Col: "f_rule_id", Val: ruleID, Condition: "="},
		},
	}
	alarmRules, err := a.alarmRuleDao.ListAlarmRule(ctx, opt)
	if err != nil {
		_log.Warnf("[logic.GetAlarmRule] ListAlarmRule err, detail: %s", err.Error())
		return nil, err
	}

	if len(alarmRules) == 0 {
		return nil, errors.NewIError(errors.AlarmRuleNotFound, "", nil)
	}

	alarmRule := alarmRules[0]

	alarmTask := &RestAlarmTask{
		RuleID:     fmt.Sprintf("%v", alarmRule.RuleID),
		AlertUsers: []*AlertUser{},
		Dags:       []*Dag{},
		Frequency:  float32(alarmRule.Frequency) / 60,
		Threshold:  alarmRule.Threshold,
	}

	var dagIDs []string
	for _, val := range alarmRules {
		dagIDs = append(dagIDs, fmt.Sprintf("%v", val.DagID))
	}

	dags, err := a.store.ListDag(ctx, &mod.ListDagInput{DagIDs: dagIDs, Type: "all"})
	if err != nil {
		_log.Warnf("[logic.GetAlarmRule] ListDag err, detail: %s", err.Error())
		return nil, err
	}

	for _, val := range dags {
		alarmTask.Dags = append(alarmTask.Dags, &Dag{ID: val.ID, Name: val.Name})
	}

	alarmUsers, err := a.alarmRuleDao.ListAlarmUser(ctx, &rds.Options{
		SearchOptions: []*rds.SearchOption{
			{Col: "f_rule_id", Val: ruleID, Condition: "="},
		},
	})
	if err != nil {
		_log.Warnf("[logic.GetAlarmRule] ListAlarmUser err, detail: %s", err.Error())
		return nil, err
	}

	for _, val := range alarmUsers {
		alarmTask.AlertUsers = append(alarmTask.AlertUsers, &AlertUser{ID: val.UserID, Name: val.UserName, Type: val.UserType})
	}

	return alarmTask, nil
}

// ErrorAlarm 告警邮件提醒
func (a *alarm) ErrorAlarm(ctx context.Context) {
	go func(ctx context.Context) {
		defer func() {
			if rErr := recover(); rErr != nil {
				time.Sleep(1 * time.Minute)
				a.ErrorAlarm(ctx)
			}
		}()

		a.errorAlarm(ctx)
	}(ctx)
}

// errorAlarm 告警邮件提醒
// 由于CheckAlarmTask和SendAlarmEmail是两个独立的goroutine，所以需要使用errChan来传递错误信息
func (a *alarm) errorAlarm(ctx context.Context) {
	errChan := make(chan interface{}, 2)
	subCtx, cancel := context.WithCancel(ctx)
	go a.CheckAlarmTask(subCtx, a.task, errChan)
	go a.SendAlarmEmail(subCtx, a.task, errChan)
	val := <-errChan
	cancel()
	panic(val)
}

// CheckAlarmTask 检查绑定的工作流是否触发告警规则
func (a *alarm) CheckAlarmTask(ctx context.Context, task chan *TriggerAlarmTask, errChan chan interface{}) {
	a.log.Infof("[CheckAlarmTask] check alarm task thread start...")
	var pastTime int64
	defer func() {
		if rErr := recover(); rErr != nil {
			a.log.Warnf("[logic.CheckAlarmTask] recover err, detail: %v", rErr)
			errChan <- rErr
		}
	}()

	timer := time.NewTimer(alarmFrequency * time.Minute)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			var err error
			ctx, span := trace.StartInternalSpan(ctx)
			_log := traceLog.WithContext(ctx)

			pastTime += alarmFrequency
			alarmRules, err := a.alarmRuleDao.GroupAlarmRule(ctx)
			if err != nil {
				_log.Warnf("[logic.CheckAlarmTask] GroupAlarmRule err, detail: %s", err.Error())
				trace.TelemetrySpanEnd(span, err)
				timer.Reset(alarmFrequency * time.Minute)
				continue
			}

			now := time.Now().Unix()
			for _, val := range alarmRules {
				if pastTime%int64(val.Frequency) == 0 {
					task <- &TriggerAlarmTask{
						TriggerTime: now,
						AlarmTask: &AlarmTask{
							RuleID:    fmt.Sprintf("%v", val.RuleID),
							Threshold: val.Threshold,
						},
					}
				}
			}
			timer.Reset(alarmFrequency * time.Minute)
		}
	}
}

func (a *alarm) SendAlarmEmail(ctx context.Context, task chan *TriggerAlarmTask, errChan chan interface{}) {
	a.log.Infof("[SendAlarmEmail] send alarm email thread start...")
	defer func() {
		if rErr := recover(); rErr != nil {
			a.log.Warnf("[logic.SendAlarmEmail] recover err, detail: %v", rErr)
			errChan <- rErr
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case val := <-task:
			var err error
			ctx, span := trace.StartInternalSpan(ctx)
			_log := traceLog.WithContext(ctx)

			// 获取所有DagID
			var restAlarmTask *RestAlarmTask
			restAlarmTask, err = a.GetAlarmRule(ctx, val.AlarmTask.RuleID)
			if err != nil {
				_log.Warnf("[logic.SendAlarmEmail] GetAlarmRule err, detail: %s", err.Error())
				trace.TelemetrySpanEnd(span, err)
				continue
			}
			val.AlarmTask = restAlarmTask.ToAlarmTask()
			// 获取范围内的dagIns
			end := val.TriggerTime
			begin := end - int64(val.AlarmTask.Frequency*60*60)
			dagInss, err := a.store.GroupDagInstance(ctx, &mod.GroupInput{
				SearchOptions: []*mod.SearchOption{
					{Field: "dagId", Value: val.AlarmTask.DagIDs, Condition: "$in"},
					{Field: "status", Value: []entity.DagInstanceStatus{entity.DagInstanceStatusFailed}, Condition: "$in"},
				},
				TimeRange: &mod.TimeRangeSearch{
					Begin: begin,
					End:   end,
					Field: "endedAt",
				},
				GroupBy: "dagId",
				IsFirst: true,
				IsSum:   true,
				Order:   -1,
				SortBy:  "endedAt",
			})
			if err != nil {
				_log.Warnf("[logic.SendAlarmEmail] GroupDagInstance err, detail: %s", err.Error())
				trace.TelemetrySpanEnd(span, err)
				continue
			}

			if len(dagInss) == 0 {
				trace.TelemetrySpanEnd(span, err)
				continue
			}

			var dagIDs, dagInsIDs []string
			var dagFailedReasonMap = map[string]*ErrorInfo{}
			var dagInsMap = map[string]string{}
			for _, dagIns := range dagInss {
				if dagIns.Total < int64(val.AlarmTask.Threshold) {
					continue
				}
				id := dagIns.DagIns.ID
				dagID := dagIns.DagIns.DagID
				dagInsMap[id] = dagID
				dagIDs = append(dagIDs, dagID)
				dagInsIDs = append(dagInsIDs, id)
				dagFailedReasonMap[dagID] = &ErrorInfo{DagID: dagID, Reason: "", FailedCount: dagIns.Total}
			}

			taskInss, err := a.store.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
				Status:    []entity.TaskInstanceStatus{entity.TaskInstanceStatusFailed},
				DagInsIDs: dagInsIDs,
				SortBy:    "updatedAt",
				Order:     -1,
			})
			if err != nil || len(taskInss) == 0 {
				if err != nil {
					_log.Warnf("[logic.SendAlarmEmail] ListTaskInstance err, detail: %s", err.Error())
					trace.TelemetrySpanEnd(span, err)
				}
				continue
			}

			// 每个流程执行失败对应的失败原因
			for _, val := range taskInss {
				// 最近一次失败原因
				dagID, ok := dagInsMap[val.DagInsID]
				if !ok {
					continue
				}
				if val, ok := dagFailedReasonMap[dagID]; !ok || ok && val.Reason != "" {
					continue
				}

				var reasonMap bson.M
				reasonBytes, _ := bson.Marshal(val.Reason)
				_ = bson.Unmarshal(reasonBytes, &reasonMap)
				if reason, ok := reasonMap["description"]; ok {
					dagFailedReasonMap[dagID].Reason = fmt.Sprintf("%v", reason)
				} else {
					dagFailedReasonMap[dagID].Reason = fmt.Sprintf("%v", val.Reason)
				}
			}

			userMails, err := a.ListUserMail(ctx, val.AlarmTask.AlertUsers)
			if err != nil {
				_log.Warnf("[logic.SendAlarmEmail] ListUserMail err, detail: %s", err.Error())
				trace.TelemetrySpanEnd(span, err)
				continue
			}

			if len(userMails) == 0 {
				continue
			}

			// 获取dag列表
			dags, err := a.store.ListDag(ctx, &mod.ListDagInput{DagIDs: dagIDs, Type: "all"})
			if err != nil {
				_log.Warnf("[logic.SendAlarmEmail] ListDag err, detail: %s", err.Error())
				trace.TelemetrySpanEnd(span, err)
				continue
			}

			if len(dags) == 0 {
				continue
			}

			for _, dag := range dags {
				errorInfo := dagFailedReasonMap[dag.ID]
				subject := common.GetEmailSubject(common.NotifyToExecutor)
				data := map[string]interface{}{
					"Name":        dag.Name,
					"Reason":      errorInfo.Reason,
					"FailedCount": errorInfo.FailedCount,
				}
				if val.AlarmTask.Frequency < 1 {
					data["IsHour"] = false
					data["Minutes"] = int(val.AlarmTask.Frequency * 60)
				} else {
					data["IsHour"] = true
					data["Hours"] = val.AlarmTask.Frequency
				}
				contents, err := utils.GetEmailTemplate("alarm.html", data)
				if err != nil {
					_log.Warnf("[logic.SendAlarmEmail] GetEmailTemplate err, detail: %s", err.Error())
					trace.TelemetrySpanEnd(span, err)
					continue
				}

				// disabled: go-lib dependency
				// go func(ctx context.Context, subject, contents string, toEmailList []string) {
				// 	err := a.sharemgnt.SendEmailWithImage(ctx, subject, contents, &common.IMG, toEmailList)
				// 	if err != nil {
				// 		_log.Warnf("[logic.SendAlarmEmail] SendEmailWithImage err, detail: %s", err.Error())
				// 		trace.TelemetrySpanEnd(span, err)
				// 		return
				// 	}
				// }(ctx, subject, contents, userMails)
				_ = subject
				_ = contents
				_ = userMails
			}

		}
	}
}

func (a *alarm) ListUserMail(ctx context.Context, alertUsers []*AlertUser) ([]string, error) {
	// 获取用户邮箱
	var groupIDs, userIDs []string
	var err error
	for _, alterUser := range alertUsers {
		if alterUser.Type == common.Group.ToString() {
			groupIDs = append(groupIDs, alterUser.ID)
		} else {
			userIDs = append(userIDs, alterUser.ID)
		}
	}

	if len(groupIDs) > 0 {
		for {
			groupUserIDs, err := a.userMgnt.GetGroupUserList(groupIDs)
			if err != nil {
				parsedError, _err := errors.ExHTTPErrorParser(err)
				if _err != nil {
					return nil, err
				}
				if parsedError["code"] != GroupIDNotFound {
					return nil, err
				}
				var newGroupIDs []string
				detail, ok := parsedError["detail"].(map[string]interface{})
				if !ok {
					return nil, err
				}
				nonExistIDs := detail["ids"].([]interface{})
				for _, groupID := range groupIDs {
					if utils.ContainsInterface(nonExistIDs, interface{}(groupID)) {
						continue
					}
					newGroupIDs = append(newGroupIDs, groupID)
				}

				groupIDs = newGroupIDs
				if len(groupIDs) == 0 {
					break
				}
				continue
			}
			userIDs = append(userIDs, groupUserIDs...)
			break
		}
	}

	userIDs = utils.RemoveRepByMap(userIDs)
	if len(userIDs) == 0 {
		return []string{}, nil
	}

	var userMailList = []string{}
	for {
		userMailList, err = a.userMgnt.GetUserMailList(userIDs)
		if err != nil {
			parsedError, _err := errors.ExHTTPErrorParser(err)
			if _err != nil {
				return nil, err
			}
			if parsedError["code"] != UserIDNotFound {
				return nil, err
			}
			var newUserIDs []string
			detail, ok := parsedError["detail"].(map[string]interface{})
			if !ok {
				return nil, err
			}
			nonExistIDs := detail["ids"].([]interface{})
			for _, userID := range userIDs {
				if utils.ContainsInterface(nonExistIDs, interface{}(userID)) {
					continue
				}
				newUserIDs = append(newUserIDs, userID)
			}

			userIDs = newUserIDs
			if len(userIDs) == 0 {
				break
			}
			continue
		}
		break
	}

	return userMailList, nil
}
