package skill

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	infralock "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/lock"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/redis/go-redis/v9"
)

const skillIndexBuildBatchSize = 200
const skillIndexBuildAssignLockExpiry = 5 * time.Second
const skillIndexBuildRunningTimeout = 30 * time.Minute

type skillIndexBuildAssignLocker interface {
	Lock(ctx context.Context) (bool, error)
	Unlock(ctx context.Context)
}

type skillIndexBuildService struct {
	logger              interfaces.Logger
	taskRepo            model.ISkillIndexBuildTaskDB
	skillRepo           model.ISkillRepository
	releaseRepo         model.ISkillReleaseDB
	indexSync           interfaces.SkillIndexSyncService
	assignLockerFactory func(taskID string) skillIndexBuildAssignLocker
}

var (
	skillIndexBuildOnce sync.Once
	skillIndexBuildInst interfaces.SkillIndexBuildService
)

// NewSkillIndexBuildService 创建技能索引构建服务
func NewSkillIndexBuildService() interfaces.SkillIndexBuildService {
	skillIndexBuildOnce.Do(func() {
		conf := config.NewConfigLoader()
		skillIndexBuildInst = &skillIndexBuildService{
			logger:              conf.GetLogger(),
			taskRepo:            dbaccess.NewSkillIndexBuildTaskDB(),
			skillRepo:           dbaccess.NewSkillRepositoryDB(),
			releaseRepo:         dbaccess.NewSkillReleaseDB(),
			indexSync:           NewSkillIndexSyncService(),
			assignLockerFactory: newSkillIndexBuildAssignLockerFactory(),
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
		MaxRetry:    3,
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
	switch interfaces.SkillIndexBuildStatus(task.Status) {
	case interfaces.SkillIndexBuildStatusPending, interfaces.SkillIndexBuildStatusRunning:
	default:
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusConflict, "skill index build task is not cancellable")
	}

	task.Status = interfaces.SkillIndexBuildStatusCanceled.String()
	task.ErrorMsg = "task canceled by user"
	task.LastFinishedTime = time.Now().UnixNano()
	if err = s.taskRepo.UpdateByTaskID(ctx, nil, task); err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}

	resp := &interfaces.CancelSkillIndexBuildTaskResp{
		TaskID: req.TaskID,
	}
	resp.Action = "cancel_task"
	return resp, nil
}

func (s *skillIndexBuildService) RetryTask(ctx context.Context, req *interfaces.RetrySkillIndexBuildTaskReq) (*interfaces.RetrySkillIndexBuildTaskResp, error) {
	task, err := s.taskRepo.SelectByTaskID(ctx, nil, req.TaskID)
	if err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	if task == nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, "skill index build task not found")
	}
	switch interfaces.SkillIndexBuildStatus(task.Status) {
	case interfaces.SkillIndexBuildStatusFailed, interfaces.SkillIndexBuildStatusCanceled, interfaces.SkillIndexBuildStatusCompleted:
	default:
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusConflict, "only failed or canceled skill index build task can be retried")
	}
	runningTask, err := s.taskRepo.SelectRunningTask(ctx, nil)
	if err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	if runningTask != nil && runningTask.TaskID != task.TaskID {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusConflict, "skill index build task is already running")
	}

	task.Status = interfaces.SkillIndexBuildStatusPending.String()
	task.TotalCount = 0
	task.SuccessCount = 0
	task.DeleteCount = 0
	task.FailedCount = 0
	task.RetryCount = 0
	task.ErrorMsg = ""
	task.LastFinishedTime = 0
	task.CursorUpdateTime = 0
	task.CursorSkillID = ""
	if interfaces.SkillIndexBuildExecuteType(task.ExecuteType) == interfaces.SkillIndexBuildExecuteTypeIncremental {
		lastTask, lastErr := s.taskRepo.SelectLatestCompletedIncrementalTask(ctx, nil)
		if lastErr != nil {
			return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, lastErr.Error())
		}
		if lastTask != nil {
			task.CursorUpdateTime = lastTask.CursorUpdateTime
			task.CursorSkillID = lastTask.CursorSkillID
		}
	}
	if err = s.taskRepo.UpdateByTaskID(ctx, nil, task); err != nil {
		return nil, infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return &interfaces.RetrySkillIndexBuildTaskResp{
		TaskID:       task.TaskID,
		SourceTaskID: req.TaskID,
		Status:       interfaces.SkillIndexBuildStatusPending,
		ExecuteType:  task.ExecuteType,
	}, nil
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
	if shouldStopTask(task) {
		return nil
	}
	if interfaces.SkillIndexBuildStatus(task.Status) == interfaces.SkillIndexBuildStatusPending {
		task.Status = interfaces.SkillIndexBuildStatusRunning.String()
		task.ErrorMsg = ""
		if err = s.taskRepo.UpdateByTaskID(ctx, nil, task); err != nil {
			s.logger.WithContext(ctx).Errorf("mark skill index build task running failed, task_id=%s, err=%v", taskID, err)
			return err
		}
	} else if interfaces.SkillIndexBuildStatus(task.Status) != interfaces.SkillIndexBuildStatusRunning {
		return nil
	}
	if err = s.indexSync.EnsureInitialized(ctx); err != nil {
		return s.failTask(ctx, task, err)
	}

	cursorUpdateTime := task.CursorUpdateTime
	cursorSkillID := task.CursorSkillID
	for {
		shouldRun, checkErr := s.shouldContinueTask(ctx, task.TaskID)
		if checkErr != nil {
			return checkErr
		}
		if !shouldRun {
			return nil
		}

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
		s.logger.WithContext(ctx).Infof("process %d skills", len(skills))
		for _, skill := range skills {
			shouldRun, checkErr = s.shouldContinueTask(ctx, task.TaskID)
			if checkErr != nil {
				return checkErr
			}
			if !shouldRun {
				s.logger.WithContext(ctx).Debugf("stop skill index build task, task_id=%s", task.TaskID)
				return nil
			}

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

func (s *skillIndexBuildService) shouldContinueTask(ctx context.Context, taskID string) (bool, error) {
	task, err := s.refreshRunningTask(ctx, taskID)
	if err != nil {
		return false, err
	}
	if shouldStopTask(task) {
		return false, nil
	}
	return true, nil
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
	if task != nil && task.MaxRetry > 0 && task.RetryCount < task.MaxRetry {
		task.Status = interfaces.SkillIndexBuildStatusPending.String()
		task.TotalCount = 0
		task.SuccessCount = 0
		task.DeleteCount = 0
		task.FailedCount = 0
		task.RetryCount++
		task.ErrorMsg = ""
		task.LastFinishedTime = 0
		task.CursorUpdateTime = 0
		task.CursorSkillID = ""
		if interfaces.SkillIndexBuildExecuteType(task.ExecuteType) == interfaces.SkillIndexBuildExecuteTypeIncremental {
			lastTask, lastErr := s.taskRepo.SelectLatestCompletedIncrementalTask(ctx, nil)
			if lastErr != nil {
				s.logger.WithContext(ctx).Errorf("load latest completed incremental task failed, task_id=%s, err=%v", task.TaskID, lastErr)
			} else if lastTask != nil {
				task.CursorUpdateTime = lastTask.CursorUpdateTime
				task.CursorSkillID = lastTask.CursorSkillID
			}
		}
		if updateErr := s.taskRepo.UpdateByTaskID(ctx, nil, task); updateErr != nil {
			s.logger.WithContext(ctx).Errorf("reschedule failed skill index build task status failed, task_id=%s, err=%v", task.TaskID, updateErr)
		}
		s.logger.WithContext(ctx).Warnf("skill index build task scheduled for retry, task_id=%s, retry_count=%d, err=%v", task.TaskID, task.RetryCount, err)
		return err
	}
	task.Status = interfaces.SkillIndexBuildStatusFailed.String()
	task.ErrorMsg = err.Error()
	task.LastFinishedTime = time.Now().UnixNano()
	if updateErr := s.taskRepo.UpdateByTaskID(ctx, nil, task); updateErr != nil {
		s.logger.WithContext(ctx).Errorf("update failed skill index build task status failed, task_id=%s, err=%v", task.TaskID, updateErr)
	}
	s.logger.WithContext(ctx).Errorf("skill index build task failed, task_id=%s, err=%v", task.TaskID, err)
	return err
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

func (s *skillIndexBuildService) refreshRunningTask(ctx context.Context, taskID string) (*model.SkillIndexBuildTaskDB, error) {
	task, err := s.taskRepo.SelectByTaskID(ctx, nil, taskID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("refresh skill index build task failed, task_id=%s, err=%v", taskID, err)
		return nil, err
	}
	return task, nil
}

func shouldStopTask(task *model.SkillIndexBuildTaskDB) bool {
	if task == nil {
		return true
	}
	status := interfaces.SkillIndexBuildStatus(task.Status)
	return status == interfaces.SkillIndexBuildStatusCanceled || status == interfaces.SkillIndexBuildStatusCompleted || status == interfaces.SkillIndexBuildStatusFailed
}

func (s *skillIndexBuildService) tryStartPendingTask(ctx context.Context, taskID string) (bool, error) {
	if taskID == "" {
		return false, nil
	}
	locker := s.newAssignLocker(taskID)
	if locker == nil {
		return false, nil
	}
	ok, err := locker.Lock(ctx)
	if err != nil || !ok {
		return false, err
	}
	defer locker.Unlock(ctx)

	task, err := s.taskRepo.SelectByTaskID(ctx, nil, taskID)
	if err != nil {
		return false, err
	}
	if task == nil || interfaces.SkillIndexBuildStatus(task.Status) != interfaces.SkillIndexBuildStatusPending {
		return false, nil
	}
	task.Status = interfaces.SkillIndexBuildStatusRunning.String()
	task.ErrorMsg = ""
	if err = s.taskRepo.UpdateByTaskID(ctx, nil, task); err != nil {
		return false, err
	}
	return true, nil
}

func (s *skillIndexBuildService) recoverStaleRunningTask(ctx context.Context) error {
	task, err := s.taskRepo.SelectRunningTask(ctx, nil)
	if err != nil || task == nil {
		return err
	}
	if interfaces.SkillIndexBuildStatus(task.Status) != interfaces.SkillIndexBuildStatusRunning {
		return nil
	}
	if time.Since(time.Unix(0, task.UpdateTime)) <= skillIndexBuildRunningTimeout {
		return nil
	}
	task.Status = interfaces.SkillIndexBuildStatusFailed.String()
	task.ErrorMsg = "stale running task recovered as failed"
	task.LastFinishedTime = time.Now().UnixNano()
	return s.taskRepo.UpdateByTaskID(ctx, nil, task)
}

func (s *skillIndexBuildService) newAssignLocker(taskID string) skillIndexBuildAssignLocker {
	if s.assignLockerFactory == nil {
		return nil
	}
	return s.assignLockerFactory(taskID)
}

func newSkillIndexBuildAssignLockerFactory() func(taskID string) skillIndexBuildAssignLocker {
	conf := config.NewConfigLoader()
	redisCli, _, err := conf.RedisConfig.GetClient()
	if err != nil || redisCli == nil {
		return func(taskID string) skillIndexBuildAssignLocker { return nil }
	}
	instanceID := conf.Project.GetMachineID()
	return func(taskID string) skillIndexBuildAssignLocker {
		return newSkillIndexBuildAssignLocker(redisCli, taskID, instanceID)
	}
}

func newSkillIndexBuildAssignLocker(redisCli *redis.Client, taskID, instanceID string) skillIndexBuildAssignLocker {
	if redisCli == nil || taskID == "" {
		return nil
	}
	lockKey := "lock:skill:index:build:assign:" + taskID
	return infralock.NewRedisLocker(redisCli, lockKey, instanceID, skillIndexBuildAssignLockExpiry)
}
