package skill

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	infracommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	infralock "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/lock"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/redis/go-redis/v9"
)

const skillIndexBuildBatchSize = 200
const skillIndexBuildCreateLockKey = "lock:skill:index:build:create"
const skillIndexBuildCreateLockExpiry = 15 * time.Second

type skillIndexBuildCreateLocker interface {
	Lock(ctx context.Context) (bool, error)
	Unlock(ctx context.Context)
}

type skillIndexBuildService struct {
	logger       interfaces.Logger
	taskRepo     model.ISkillIndexBuildTaskDB
	skillRepo    model.ISkillRepository
	releaseRepo  model.ISkillReleaseDB
	indexSync    interfaces.SkillIndexSyncService
	enqueuer     skillIndexBuildTaskEnqueuer
	inspector    skillIndexBuildInspector
	createLocker func() skillIndexBuildCreateLocker
}

var (
	skillIndexBuildOnce sync.Once
	skillIndexBuildInst interfaces.SkillIndexBuildService
)

func NewSkillIndexBuildService() interfaces.SkillIndexBuildService {
	skillIndexBuildOnce.Do(func() {
		conf := config.NewConfigLoader()
		enqueuer, err := newSkillIndexBuildTaskEnqueuer()
		if err != nil {
			panic(err)
		}
		inspector, err := newSkillIndexBuildInspector()
		if err != nil {
			panic(err)
		}
		skillIndexBuildInst = &skillIndexBuildService{
			logger:       conf.GetLogger(),
			taskRepo:     dbaccess.NewSkillIndexBuildTaskDB(),
			skillRepo:    dbaccess.NewSkillRepositoryDB(),
			releaseRepo:  dbaccess.NewSkillReleaseDB(),
			indexSync:    NewSkillIndexSyncService(),
			enqueuer:     enqueuer,
			inspector:    inspector,
			createLocker: newSkillIndexBuildCreateLockerFactory(),
		}
	})
	return skillIndexBuildInst
}

func (s *skillIndexBuildService) CreateTask(ctx context.Context, req *interfaces.CreateSkillIndexBuildTaskReq) (*interfaces.CreateSkillIndexBuildTaskResp, error) {
	resp, err := s.createTask(ctx, req.UserID, req.ExecuteType)
	if err != nil {
		return nil, err
	}
	return &interfaces.CreateSkillIndexBuildTaskResp{
		TaskID:      resp.TaskID,
		Status:      resp.Status,
		ExecuteType: resp.ExecuteType,
	}, nil
}

func (s *skillIndexBuildService) createTask(ctx context.Context, userID string, executeType interfaces.SkillIndexBuildExecuteType) (*interfaces.RetrySkillIndexBuildTaskResp, error) {
	if s.createLocker != nil {
		locker := s.createLocker()
		if locker != nil {
			ok, err := locker.Lock(ctx)
			if err != nil {
				return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			}
			if !ok {
				return nil, infraerrors.DefaultHTTPError(ctx, http.StatusConflict, "skill index build task is being created")
			}
			defer locker.Unlock(ctx)
		}
	}

	runningTask, err := s.taskRepo.SelectRunningTask(ctx, nil)
	if err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	if runningTask != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusConflict, "skill index build task is already running")
	}

	task := &model.SkillIndexBuildTaskDB{
		TaskID:      uuid.NewString(),
		Status:      interfaces.SkillIndexBuildStatusPending.String(),
		ExecuteType: executeType.String(),
		CreateUser:  userID,
		MaxRetry:    10,
	}
	if executeType == interfaces.SkillIndexBuildExecuteTypeIncremental {
		lastTask, lastErr := s.taskRepo.SelectLatestCompletedIncrementalTask(ctx, nil)
		if lastErr != nil {
			return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, lastErr.Error())
		}
		if lastTask != nil {
			task.CursorUpdateTime = lastTask.CursorUpdateTime
			task.CursorSkillID = lastTask.CursorSkillID
		}
	}
	if err = s.taskRepo.Insert(ctx, nil, task); err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}

	if err = s.enqueuer.Enqueue(copyTaskContext(ctx), newSkillIndexBuildTaskPayload(ctx, task.TaskID)); err != nil {
		task.Status = interfaces.SkillIndexBuildStatusFailed.String()
		task.ErrorMsg = err.Error()
		task.LastFinishedTime = time.Now().UnixNano()
		_ = s.taskRepo.UpdateByTaskID(ctx, nil, task)
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}

	return &interfaces.RetrySkillIndexBuildTaskResp{
		TaskID:      task.TaskID,
		Status:      interfaces.SkillIndexBuildStatusPending,
		ExecuteType: task.ExecuteType,
	}, nil
}

func (s *skillIndexBuildService) GetTask(ctx context.Context, req *interfaces.GetSkillIndexBuildTaskReq) (*interfaces.SkillIndexBuildTaskResp, error) {
	task, err := s.taskRepo.SelectByTaskID(ctx, nil, req.TaskID)
	if err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	if task == nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, "skill index build task not found")
	}
	resp := toSkillIndexBuildTaskResp(task)
	resp.QueueState = s.getQueueState(ctx, task.TaskID)
	return resp, nil
}

func (s *skillIndexBuildService) QueryTaskList(ctx context.Context, req *interfaces.QuerySkillIndexBuildTaskListReq) (*interfaces.QuerySkillIndexBuildTaskListResp, error) {
	filter := map[string]interface{}{
		"all":          req.All,
		"limit":        req.PageSize,
		"offset":       (req.Page - 1) * req.PageSize,
		"status":       req.Status.String(),
		"execute_type": req.ExecuteType,
		"create_user":  req.CreateUser,
	}
	sortField := "f_update_time"
	switch req.SortBy {
	case "create_time":
		sortField = "f_create_time"
	case "name":
		sortField = "f_task_id"
	}
	sortOrder := ormhelper.SortOrder(strings.ToUpper(req.SortOrder))
	if !sortOrder.IsValid() {
		sortOrder = ormhelper.SortOrderDesc
	}
	sort := &ormhelper.SortParams{
		Fields: []ormhelper.SortField{{
			Field: sortField,
			Order: sortOrder,
		}},
	}
	total, err := s.taskRepo.CountByWhereClause(ctx, nil, filter)
	if err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	taskList, err := s.taskRepo.SelectListPage(ctx, nil, filter, sort, nil)
	if err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	resp := &interfaces.QuerySkillIndexBuildTaskListResp{
		CommonPageResult: buildCommonPageResult(req.Page, req.PageSize, total),
		Data:             make([]*interfaces.SkillIndexBuildTaskResp, 0, len(taskList)),
	}
	for _, task := range taskList {
		item := toSkillIndexBuildTaskResp(task)
		item.QueueState = s.getQueueState(ctx, task.TaskID)
		resp.Data = append(resp.Data, item)
	}
	return resp, nil
}

func (s *skillIndexBuildService) CancelTask(ctx context.Context, req *interfaces.CancelSkillIndexBuildTaskReq) (*interfaces.CancelSkillIndexBuildTaskResp, error) {
	task, err := s.taskRepo.SelectByTaskID(ctx, nil, req.TaskID)
	if err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	if task == nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, "skill index build task not found")
	}
	if s.inspector == nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "skill index build inspector is not initialized")
	}
	info, err := s.inspector.GetTaskInfo(skillIndexBuildQueueName, req.TaskID)
	if err != nil {
		if errors.Is(err, asynq.ErrTaskNotFound) || errors.Is(err, asynq.ErrQueueNotFound) {
			return nil, infraerrors.DefaultHTTPError(ctx, http.StatusConflict, "skill index build task is not cancellable")
		}
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	resp := &interfaces.CancelSkillIndexBuildTaskResp{
		TaskID:     req.TaskID,
		QueueState: asynqTaskStateToString(info.State),
	}
	switch info.State {
	case asynq.TaskStateActive:
		if err = s.inspector.CancelProcessing(req.TaskID); err != nil {
			return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		}
		task.ErrorMsg = "cancel requested by user"
		if updateErr := s.taskRepo.UpdateByTaskID(ctx, nil, task); updateErr != nil {
			return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, updateErr.Error())
		}
		resp.Action = "cancel_processing"
		return resp, nil
	case asynq.TaskStatePending, asynq.TaskStateScheduled, asynq.TaskStateRetry, asynq.TaskStateArchived:
		if err = s.inspector.DeleteTask(skillIndexBuildQueueName, req.TaskID); err != nil {
			return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		}
		task.Status = interfaces.SkillIndexBuildStatusFailed.String()
		task.ErrorMsg = "task canceled by user"
		task.LastFinishedTime = time.Now().UnixNano()
		if updateErr := s.taskRepo.UpdateByTaskID(ctx, nil, task); updateErr != nil {
			return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, updateErr.Error())
		}
		resp.Action = "delete_queue_task"
		return resp, nil
	default:
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusConflict, "skill index build task is not cancellable")
	}
}

func (s *skillIndexBuildService) RetryTask(ctx context.Context, req *interfaces.RetrySkillIndexBuildTaskReq) (*interfaces.RetrySkillIndexBuildTaskResp, error) {
	task, err := s.taskRepo.SelectByTaskID(ctx, nil, req.TaskID)
	if err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	if task == nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, "skill index build task not found")
	}
	if task.Status != interfaces.SkillIndexBuildStatusFailed.String() {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusConflict, "only failed skill index build task can be retried")
	}
	resp, err := s.createTask(ctx, req.UserID, interfaces.SkillIndexBuildExecuteType(task.ExecuteType))
	if err != nil {
		return nil, err
	}
	resp.SourceTaskID = req.TaskID
	return resp, nil
}

func (s *skillIndexBuildService) RecoverRunningTasks(ctx context.Context) error {
	return nil
}

func (s *skillIndexBuildService) runTask(ctx context.Context, taskID string) error {
	task, err := s.taskRepo.SelectByTaskID(ctx, nil, taskID)
	if err != nil || task == nil {
		if err != nil {
			s.logger.WithContext(ctx).Errorf("load skill index build task failed, task_id=%s, err=%v", taskID, err)
			return err
		}
		return nil
	}
	task.Status = interfaces.SkillIndexBuildStatusRunning.String()
	task.ErrorMsg = ""
	if err = s.taskRepo.UpdateByTaskID(ctx, nil, task); err != nil {
		s.logger.WithContext(ctx).Errorf("mark skill index build task running failed, task_id=%s, err=%v", taskID, err)
		return err
	}
	if err = s.indexSync.EnsureInitialized(ctx); err != nil {
		return s.failTask(ctx, task, err)
	}

	cursorUpdateTime := task.CursorUpdateTime
	cursorSkillID := task.CursorSkillID
	for {
		skills, scanErr := s.skillRepo.SelectSkillBuildPage(ctx, nil, cursorUpdateTime, cursorSkillID, skillIndexBuildBatchSize)
		if scanErr != nil {
			return s.failTask(ctx, task, scanErr)
		}
		if len(skills) == 0 {
			task.Status = interfaces.SkillIndexBuildStatusCompleted.String()
			task.LastFinishedTime = time.Now().UnixNano()
			if updateErr := s.taskRepo.UpdateByTaskID(ctx, nil, task); updateErr != nil {
				s.logger.WithContext(ctx).Errorf("complete skill index build task failed, task_id=%s, err=%v", task.TaskID, updateErr)
				return updateErr
			}
			return nil
		}

		for _, skill := range skills {
			task.TotalCount++
			action, actionErr := s.handleSkill(ctx, skill)
			if actionErr != nil {
				task.FailedCount++
				s.logger.WithContext(ctx).Errorf("process skill index build item failed, task_id=%s, skill_id=%s, err=%v", task.TaskID, skill.SkillID, actionErr)
			} else {
				switch action {
				case "upsert":
					task.SuccessCount++
				case "delete":
					task.DeleteCount++
				}
			}
			task.CursorUpdateTime = skill.UpdateTime
			task.CursorSkillID = skill.SkillID
			cursorUpdateTime = skill.UpdateTime
			cursorSkillID = skill.SkillID
		}
		if err = s.taskRepo.UpdateByTaskID(ctx, nil, task); err != nil {
			return s.failTask(ctx, task, err)
		}
	}
}

func (s *skillIndexBuildService) handleSkill(ctx context.Context, skill *model.SkillRepositoryDB) (string, error) {
	if skill == nil {
		return "", nil
	}
	if skill.IsDeleted {
		if err := s.indexSync.DeleteSkill(ctx, skill.SkillID); err != nil {
			return "", err
		}
		return "delete", nil
	}

	release, err := s.releaseRepo.SelectBySkillID(ctx, nil, skill.SkillID)
	if err != nil {
		return "", err
	}

	switch interfaces.BizStatus(skill.Status) {
	case interfaces.BizStatusPublished:
		payload := skill
		if release != nil {
			payload = releaseToSkillRepository(release)
		}
		if err = s.indexSync.UpsertSkill(ctx, payload); err != nil {
			return "", err
		}
		return "upsert", nil
	case interfaces.BizStatusEditing:
		if release == nil {
			if err = s.indexSync.DeleteSkill(ctx, skill.SkillID); err != nil {
				return "", err
			}
			return "delete", nil
		}
		if err = s.indexSync.UpsertSkill(ctx, releaseToSkillRepository(release)); err != nil {
			return "", err
		}
		return "upsert", nil
	case interfaces.BizStatusOffline, interfaces.BizStatusUnpublish:
		if err = s.indexSync.DeleteSkill(ctx, skill.SkillID); err != nil {
			return "", err
		}
		return "delete", nil
	default:
		if err = s.indexSync.DeleteSkill(ctx, skill.SkillID); err != nil {
			return "", err
		}
		return "delete", nil
	}
}

func (s *skillIndexBuildService) failTask(ctx context.Context, task *model.SkillIndexBuildTaskDB, err error) error {
	task.Status = interfaces.SkillIndexBuildStatusFailed.String()
	task.ErrorMsg = err.Error()
	task.LastFinishedTime = time.Now().UnixNano()
	if updateErr := s.taskRepo.UpdateByTaskID(ctx, nil, task); updateErr != nil {
		s.logger.WithContext(ctx).Errorf("update failed skill index build task status failed, task_id=%s, err=%v", task.TaskID, updateErr)
	}
	s.logger.WithContext(ctx).Errorf("skill index build task failed, task_id=%s, err=%v", task.TaskID, err)
	return err
}

func copyTaskContext(ctx context.Context) context.Context {
	bg := context.Background()
	if authCtx, ok := infracommon.GetAccountAuthContextFromCtx(ctx); ok {
		bg = infracommon.SetAccountAuthContextToCtx(bg, authCtx)
	}
	if businessDomain, ok := infracommon.GetBusinessDomainFromCtx(ctx); ok {
		bg = infracommon.SetBusinessDomainToCtx(bg, businessDomain)
	}
	bg = infracommon.SetPublicAPIToCtx(bg, infracommon.IsPublicAPIFromCtx(ctx))
	if language := infracommon.GetLanguageFromCtx(ctx); language != "" {
		bg = infracommon.SetLanguageToCtx(bg, language)
	}
	return bg
}

func releaseToSkillRepository(release *model.SkillReleaseDB) *model.SkillRepositoryDB {
	if release == nil {
		return nil
	}
	return &model.SkillRepositoryDB{
		SkillID:      release.SkillID,
		Name:         release.Name,
		Description:  release.Description,
		SkillContent: release.SkillContent,
		Version:      release.Version,
		Category:     release.Category,
		Status:       release.Status,
		Source:       release.Source,
		ExtendInfo:   release.ExtendInfo,
		Dependencies: release.Dependencies,
		FileManifest: release.FileManifest,
		CreateTime:   release.CreateTime,
		CreateUser:   release.CreateUser,
		UpdateTime:   release.UpdateTime,
		UpdateUser:   release.UpdateUser,
	}
}

func toSkillIndexBuildTaskResp(task *model.SkillIndexBuildTaskDB) *interfaces.SkillIndexBuildTaskResp {
	return &interfaces.SkillIndexBuildTaskResp{
		TaskID:           task.TaskID,
		Status:           interfaces.SkillIndexBuildStatus(task.Status),
		ExecuteType:      task.ExecuteType,
		QueueState:       "",
		TotalCount:       task.TotalCount,
		SuccessCount:     task.SuccessCount,
		DeleteCount:      task.DeleteCount,
		FailedCount:      task.FailedCount,
		RetryCount:       task.RetryCount,
		MaxRetry:         task.MaxRetry,
		CursorUpdateTime: task.CursorUpdateTime,
		CursorSkillID:    task.CursorSkillID,
		ErrorMsg:         task.ErrorMsg,
		CreateUser:       task.CreateUser,
		CreateTime:       task.CreateTime,
		UpdateTime:       task.UpdateTime,
		LastFinishedTime: task.LastFinishedTime,
	}
}

func (s *skillIndexBuildService) getQueueState(ctx context.Context, taskID string) string {
	if s.inspector == nil || taskID == "" {
		return ""
	}
	info, err := s.inspector.GetTaskInfo(skillIndexBuildQueueName, taskID)
	if err != nil || info == nil {
		return ""
	}
	return asynqTaskStateToString(info.State)
}

func asynqTaskStateToString(state asynq.TaskState) string {
	switch state {
	case asynq.TaskStateActive:
		return "active"
	case asynq.TaskStatePending:
		return "pending"
	case asynq.TaskStateScheduled:
		return "scheduled"
	case asynq.TaskStateRetry:
		return "retry"
	case asynq.TaskStateArchived:
		return "archived"
	case asynq.TaskStateCompleted:
		return "completed"
	case asynq.TaskStateAggregating:
		return "aggregating"
	default:
		return ""
	}
}

func buildCommonPageResult(page, pageSize int, total int64) interfaces.CommonPageResult {
	if pageSize <= 0 {
		pageSize = interfaces.DefaultPageSize
	}
	if page <= 0 {
		page = interfaces.DefaultPage
	}
	totalPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	return interfaces.CommonPageResult{
		TotalCount: int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPage:  totalPage,
		HasNext:    page < totalPage,
		HasPrev:    page > 1,
	}
}

func newSkillIndexBuildCreateLockerFactory() func() skillIndexBuildCreateLocker {
	conf := config.NewConfigLoader()
	redisCli, _, err := conf.RedisConfig.GetClient()
	if err != nil || redisCli == nil {
		return nil
	}
	lockValue := conf.Project.GetMachineID()
	return func() skillIndexBuildCreateLocker {
		return infralock.NewRedisLocker(redisCli, skillIndexBuildCreateLockKey, lockValue, skillIndexBuildCreateLockExpiry)
	}
}

func newSkillIndexBuildCreateLockerFromClient(redisCli *redis.Client, lockValue string) func() skillIndexBuildCreateLocker {
	if redisCli == nil {
		return nil
	}
	return func() skillIndexBuildCreateLocker {
		return infralock.NewRedisLocker(redisCli, skillIndexBuildCreateLockKey, lockValue, skillIndexBuildCreateLockExpiry)
	}
}
