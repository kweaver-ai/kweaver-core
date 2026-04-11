package mod

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	aerr "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	liberrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm/opcode"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm/state"
)

type VMExt struct {
	*vm.VM

	dagIns    *entity.DagInstance        `json:"-"`
	logger    traceLog.Logger            `json:"-"`
	userID    string                     `json:"-"`
	events    []*entity.DagInstanceEvent `json:"-"`
	eventLock sync.Mutex                 `json:"-"`
}

func NewVMExt(ctx context.Context, dagIns *entity.DagInstance, userID string) *VMExt {
	dagIns.ShareData.Save = nil
	vmIns := &VMExt{
		VM:     vm.NewVM(),
		dagIns: dagIns,
		logger: traceLog.WithContext(ctx),
		userID: userID,
		events: make([]*entity.DagInstanceEvent, 0),
	}

	vmIns.SetContext(ctx)
	vmIns.AddFuncs(compareFuncs)
	vmIns.AddGlobals(NewGlobals(dagIns))
	vmIns.SetExtfunc(NewExtFunc(vmIns, dagIns, userID))
	vmIns.SetHook(vmIns)
	return vmIns
}

func (vmIns *VMExt) canPersistEvent() bool {
	return vmIns.dagIns != nil && vmIns.dagIns.EventPersistence == entity.DagInstanceEventPersistenceSql
}

func (vmIns *VMExt) AppendEvents(events ...*entity.DagInstanceEvent) {
	if !vmIns.canPersistEvent() {
		return
	}

	vmIns.eventLock.Lock()
	defer vmIns.eventLock.Unlock()
	vmIns.events = append(vmIns.events, events...)
}

func (vmIns *VMExt) WriteEvents() error {

	if !vmIns.canPersistEvent() || len(vmIns.events) == 0 {
		return nil
	}

	vmIns.eventLock.Lock()
	defer vmIns.eventLock.Unlock()
	events := vmIns.events
	vmIns.events = make([]*entity.DagInstanceEvent, 0)
	err := vmIns.dagIns.WriteEvents(context.Background(), events)
	return err
}

func (vmIns *VMExt) LoadDag(dag *entity.Dag) (err error) {
	steps := dag.Steps

	if len(steps) == 0 {
		return
	}

	g := vm.NewGenerator(vmIns.VM)
	err = g.GenerateTrigger(&steps[0])

	if err != nil {
		return
	}

	for _, step := range steps[1:] {
		err = g.GenerateStep(&step)
		if err != nil {
			return
		}
	}

	vmIns.LoadInstructions(g.Instructions)
	vmIns.AppendEvents(&entity.DagInstanceEvent{
		Type:       rds.DagInstanceEventTypeInstructions,
		InstanceID: vmIns.dagIns.ID,
		Data:       vmIns.Instructions,
		Timestamp:  time.Now().UnixMicro(),
		Visibility: rds.DagInstanceEventVisibilityPrivate,
	})
	return err
}

func (vmIns *VMExt) HandleDagInsError(err error) {
	ctx := vmIns.Context()
	store := GetStore()
	dagIns := vmIns.dagIns

	patch := &entity.DagInstance{
		BaseInfo:         dagIns.BaseInfo,
		EventPersistence: dagIns.EventPersistence,
		EndedAt:          time.Now().Unix(),
		Status:           entity.DagInstanceStatusFailed,
	}

	if dbErr := store.PatchDagIns(ctx, patch); dbErr != nil {
		log.Warnf("[VMExt.HandleDagInsError] PatchDagIns err: %s", dbErr.Error())
	}

	go dagIns.SendErrorCallback(err)
}

func (vmIns *VMExt) Boot() error {
	ctx := vmIns.Context()
	store := GetStore()
	dagIns, logger := vmIns.dagIns, vmIns.logger

	if dagIns == nil {
		err := errors.New("invalid dagIns")
		logger.Warnf("[VMExt.Boot] err: %s", err.Error())
		return err
	}

	if dagIns.Mode != entity.DagInstanceModeVM {
		err := fmt.Errorf("invalid dagIns: id %s, mode %v", dagIns.ID, dagIns.Mode)
		logger.Warnf("[VMExt.Boot] err: %s", err.Error())
		return err
	}

	switch dagIns.Status {
	case entity.DagInstanceStatusBlocked,
		entity.DagInstanceStatusCancled,
		entity.DagInstanceStatusFailed,
		entity.DagInstanceStatusInit,
		entity.DagInstanceStatusSuccess:
		err := fmt.Errorf("invalid dagIns: id %s, status %v", dagIns.ID, dagIns.Status)
		logger.Warnf("[VMExt.Boot] err: %s", err.Error())
		return err
	default:
		locked := dagIns.Lock(300 * time.Second)
		if locked {
			defer func() {
				dagIns.Unlock()
			}()
			if dagIns.Status == entity.DagInstanceStatusScheduled {
				if err := store.PatchDagIns(ctx, &entity.DagInstance{
					BaseInfo:         dagIns.BaseInfo,
					Status:           entity.DagInstanceStatusRunning,
					EventPersistence: dagIns.EventPersistence,
				}); err != nil {
					vmIns.HandleDagInsError(err)
					return err
				}
				dagIns.Status = entity.DagInstanceStatusRunning
			}
		} else {
			if dagIns.Status == entity.DagInstanceStatusRunning {
				if err := store.PatchDagIns(ctx, &entity.DagInstance{
					BaseInfo:         dagIns.BaseInfo,
					Status:           entity.DagInstanceStatusScheduled,
					EventPersistence: dagIns.EventPersistence,
				}); err != nil {
					vmIns.HandleDagInsError(err)
					return err
				}
			}
			return nil
		}
	}

	dag, err := store.GetDagWithOptionalVersion(ctx, dagIns.DagID, dagIns.VersionID)

	if err != nil {
		logger.Warnf("[VMExt.Boot] GetDag err, deail: %s", err.Error())

		if aerr.IsNotFoundErr(err) {
			err = liberrors.NewPublicRestError(ctx, liberrors.PErrorNotFound,
				liberrors.PErrorNotFound,
				map[string]string{"dagId": dagIns.DagID})
		} else {
			err = liberrors.NewPublicRestError(ctx, liberrors.PErrorInternalServerError,
				liberrors.PErrorInternalServerError,
				nil)
		}
		vmIns.HandleDagInsError(err)
		return err
	}

	switch dagIns.EventPersistence {
	case entity.DagInstanceEventPersistenceOss:
		return fmt.Errorf("dagIns is already archived")
	case entity.DagInstanceEventPersistenceSql:

		events, err := dagIns.ListEvents(vmIns.Context(), &rds.DagInstanceEventListOptions{
			DagInstanceID: dagIns.ID,
			Types:         []rds.DagInstanceEventType{rds.DagInstanceEventTypeInstructions, rds.DagInstanceEventTypeVM},
			LatestOnly:    true,
		})

		if err != nil {
			vmIns.HandleDagInsError(err)
			return err
		}

		var vmData, instructionsData string

		for _, event := range events {
			switch event.Type {
			case rds.DagInstanceEventTypeInstructions:
				instructionsData = event.Data.(string)
			case rds.DagInstanceEventTypeVM:
				vmData = event.Data.(string)
			}
		}

		if vmData == "" || instructionsData == "" {
			if err := vmIns.LoadDag(dag); err != nil {
				vmIns.HandleDagInsError(err)
				return err
			}
			vmIns.Run()
			return nil
		}

		if err := json.Unmarshal([]byte(vmData), vmIns); err != nil {
			err = fmt.Errorf("invalid dagIns dump: id %s", dagIns.ID)
			vmIns.HandleDagInsError(err)
			return err
		}

		vmIns.Instructions = make([]*opcode.Instruction, 0)

		if err := json.Unmarshal([]byte(instructionsData), &vmIns.Instructions); err != nil {
			err = fmt.Errorf("invalid dagIns instructions: id %s", dagIns.ID)
			vmIns.HandleDagInsError(err)
			return err
		}

	default:
		if dagIns.Dump == "" {
			if err := vmIns.LoadDag(dag); err != nil {
				vmIns.HandleDagInsError(err)
				return err
			}
			vmIns.Run()
			return nil
		}

		if err := json.Unmarshal([]byte(dagIns.Dump), vmIns); err != nil {
			err = fmt.Errorf("invalid dagIns dump: id %s", dagIns.ID)
			vmIns.HandleDagInsError(err)
			return err
		}
	}

	switch vmIns.State {
	case state.Wait:
		rets := make([]any, 1)
		if err := json.Unmarshal([]byte(dagIns.ResumeData), &rets); err != nil {
			vmIns.ResumeError(errors.New("invalid return values"))
			return err
		}

		if dagIns.ResumeStatus == entity.TaskInstanceStatusSuccess {
			vmIns.Resume(rets...)
		} else {
			vmIns.ResumeError(liberrors.NewPublicRestError(ctx,
				liberrors.PErrorInternalServerError,
				liberrors.PErrorInternalServerError, rets[0]))
		}
	default:
		vmIns.Start()
	}

	return nil
}
