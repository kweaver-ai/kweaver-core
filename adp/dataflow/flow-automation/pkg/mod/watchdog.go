package mod

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
)

const DefFailedReason = "force failed by watch dog because it execute too long"

// DefWatchDog
type DefWatchDog struct {
	dagScheduledTimeout time.Duration
	dagRunningTimeout   time.Duration
	wg                  sync.WaitGroup
	closeCh             chan struct{}
}

// NewDefWatchDog
func NewDefWatchDog(dagScheduledTimeout, dagRunningTimeout time.Duration) *DefWatchDog {
	return &DefWatchDog{
		dagScheduledTimeout: dagScheduledTimeout,
		dagRunningTimeout:   dagRunningTimeout,
		closeCh:             make(chan struct{}),
	}
}

// Init
func (wd *DefWatchDog) Init() {
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleExpiredTaskIns)
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleLeftBehindDagIns)
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleLongRunningDagIns)
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleLongBlockedDagIns)
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleBlockedTaskIns)
}

// Close
func (wd *DefWatchDog) Close() {
	close(wd.closeCh)
	wd.wg.Wait()
}

func (wd *DefWatchDog) watchWrapper(do func() error) {
	timerCh := time.Tick(time.Second)
	closed := false
	for !closed {
		select {
		case <-wd.closeCh:
			closed = true
		case <-timerCh:
			if err := do(); err != nil {
				wd.handleErr(err)
			}
		}
	}
	wd.wg.Done()
}

func (wd *DefWatchDog) handleExpiredTaskIns() error {
	taskIns, err := GetStore().ListTaskInstance(context.TODO(), &ListTaskInstanceInput{
		Status:  []entity.TaskInstanceStatus{entity.TaskInstanceStatusRunning},
		Expired: true,
	})
	if err != nil {
		return err
	}
	if len(taskIns) == 0 {
		return nil
	}

	for i := range taskIns {
		if err := GetStore().PatchDagIns(context.TODO(), &entity.DagInstance{
			BaseInfo: entity.BaseInfo{ID: taskIns[i].DagInsID},
			Status:   entity.DagInstanceStatusFailed,
			Reason:   DefFailedReason,
			EndedAt:  time.Now().Unix(),
		}); err != nil {
			return fmt.Errorf("patch expired dag instance[%s] failed: %s", taskIns[i].DagInsID, err)
		}

		if err := GetStore().PatchTaskIns(context.TODO(), &entity.TaskInstance{
			BaseInfo: entity.BaseInfo{ID: taskIns[i].ID},
			Status:   entity.TaskInstanceStatusFailed,
			Reason:   DefFailedReason,
		}); err != nil {
			return fmt.Errorf("patch expired task[%s] failed: %s", taskIns[i].ID, err)
		}
	}
	return nil
}

func (wd *DefWatchDog) handleBlockedTaskIns() error {
	taskIns, err := GetStore().ListTaskInstance(context.TODO(), &ListTaskInstanceInput{
		Status:          []entity.TaskInstanceStatus{entity.TaskInstanceStatusBlocked},
		ActionName:      []string{common.IntelliinfoTranfer, common.InternalToolPy3Opt, common.OpContentEntity, common.OpEcoconfigReindex},
		ActionNameRegex: "^@operator/",
		Expired:         true,
	})
	if err != nil {
		return err
	}
	if len(taskIns) == 0 {
		return nil
	}

	for i := range taskIns {
		if err := GetStore().PatchDagIns(context.TODO(), &entity.DagInstance{
			BaseInfo: entity.BaseInfo{ID: taskIns[i].DagInsID},
			Status:   entity.DagInstanceStatusFailed,
			Reason:   DefFailedReason,
			EndedAt:  time.Now().Unix(),
		}); err != nil {
			return fmt.Errorf("patch expired dag instance[%s] failed: %s", taskIns[i].DagInsID, err)
		}

		if err := GetStore().PatchTaskIns(context.TODO(), &entity.TaskInstance{
			BaseInfo: entity.BaseInfo{ID: taskIns[i].ID},
			Status:   entity.TaskInstanceStatusFailed,
			Reason:   DefFailedReason,
		}); err != nil {
			return fmt.Errorf("patch expired task[%s] failed: %s", taskIns[i].ID, err)
		}
	}
	return nil
}

func (wd *DefWatchDog) handleLeftBehindDagIns() error {
	dagIns, err := GetStore().ListDagInstance(context.TODO(), &ListDagInstanceInput{
		Status:        []entity.DagInstanceStatus{entity.DagInstanceStatusScheduled},
		UpdatedEnd:    time.Now().Add(-1 * wd.dagScheduledTimeout).Unix(),
		ExcludeModeVM: true,
	})
	if err != nil {
		return err
	}
	if len(dagIns) == 0 {
		return nil
	}

	for i := range dagIns {
		debugMode := os.Getenv("DEBUG")
		if debugMode == "true" {
			log.Infof("handleLeftBehindDagIns dagins: %s", dagIns[i].ID)
		}
		dagIns[i].Status = entity.DagInstanceStatusInit
	}
	if err := GetStore().BatchUpdateDagIns(context.TODO(), dagIns); err != nil {
		return err
	}
	return nil
}

func (wd *DefWatchDog) handleLongRunningDagIns() error {
	dagIns, err := GetStore().ListDagInstance(context.TODO(), &ListDagInstanceInput{
		Status:        []entity.DagInstanceStatus{entity.DagInstanceStatusRunning},
		UpdatedEnd:    time.Now().Add(-1 * wd.dagRunningTimeout).Unix(),
		ExcludeModeVM: true,
	})
	if err != nil {
		return err
	}
	if len(dagIns) == 0 {
		return nil
	}

	for i := range dagIns {
		dagIns[i].Status = entity.DagInstanceStatusInit
	}
	if err := GetStore().BatchUpdateDagIns(context.TODO(), dagIns); err != nil {
		return err
	}
	return nil
}

func (wd *DefWatchDog) handleLongBlockedDagIns() error {
	dagIns, err := GetStore().ListDagInstance(context.TODO(), &ListDagInstanceInput{
		Status:        []entity.DagInstanceStatus{entity.DagInstanceStatusBlocked},
		UpdatedEnd:    time.Now().Add(-1 * wd.dagRunningTimeout).Unix(),
		ExcludeModeVM: true,
	})
	if err != nil {
		return err
	}
	if len(dagIns) == 0 {
		return nil
	}

	for i := range dagIns {
		dagIns[i].Status = entity.DagInstanceStatusInit
	}
	if err := GetStore().BatchUpdateDagIns(context.TODO(), dagIns); err != nil {
		return err
	}
	return nil
}

func (wd *DefWatchDog) handleErr(err error) {
	log.Error("here are some errors",
		"module", "watchdog",
		"err", err)
}
