package mod

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/event"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils/data"
	"github.com/panjf2000/ants/v2"
	"github.com/shiningrush/goevent"
)

// DefDispatcher
type DefDispatcher struct {
	closeCh chan struct{}

	wg           sync.WaitGroup
	log          commonLog.Logger
	listinsCount int
}

// NewDefDispatcher
func NewDefDispatcher(listinsCount int) *DefDispatcher {
	return &DefDispatcher{
		closeCh:      make(chan struct{}),
		log:          commonLog.NewLogger(),
		listinsCount: listinsCount,
	}
}

// Init
func (d *DefDispatcher) Init() {
	d.wg.Add(1)
	go d.WatchInitDags()
}

// WatchInitDags
func (d *DefDispatcher) WatchInitDags() {
	closed := false
	timerCh := time.Tick(time.Second)
	for !closed {
		select {
		case <-d.closeCh:
			closed = true
		case <-timerCh:
			start := time.Now()
			e := &event.DispatchInitDagInsCompleted{}
			if err := d.Do(); err != nil {
				d.handlerErr(err)
				e.Error = err
			}
			e.ElapsedMs = time.Now().Sub(start).Milliseconds()
			goevent.Publish(e)
		}
	}
	d.wg.Done()
}

// Do dispatch
func (d *DefDispatcher) Do() error {
	// 根据优先级分别进行调度
	users, err := GetStore().DisdinctDagInstance(&ListDagInstanceInput{
		Status: []entity.DagInstanceStatus{
			entity.DagInstanceStatusInit,
		},
		DistinctField: "userid",
	})
	if err != nil {
		return err
	}
	nodes, err := GetKeeper().AliveNodes()
	if err != nil {
		d.log.Errorf("[Diapatch] get alive node err: %v", err.Error())
		return err
	}
	if len(nodes) == 0 {
		return data.ErrNoAliveNodes
	}
	for _, userid := range users {
		err := d.pushScheduled(userid.(string), nodes)
		if err != nil {
			d.log.Errorf("[Diapatch pushScheduled] userid:%v, err: %v", userid, err.Error())
			continue
		}
	}

	return nil
}

func (d *DefDispatcher) pushScheduled(userid string, nodes []string) error {
	var ins = make([]*entity.DagInstance, 0)
	var prioritySlice = []string{
		common.PriorityHighest,
		common.PriorityHigh,
		common.PriorityMedium,
		common.PriorityLow,
		common.PriorityLowest,
	}
	var wg = &sync.WaitGroup{}
	var pool, _ = ants.NewPool(2)
	defer pool.Release()

	for _, p := range prioritySlice {
		wg.Add(1)
		goFunc := func(priority string) func() {
			return func() {
				defer wg.Done()
				cons := []interface{}{priority}
				if priority == common.PriorityLowest {
					cons = append(cons, nil)
				}
				count, err := GetStore().GetDagInstanceCount(context.TODO(), map[string]interface{}{
					"userid": userid,
					"status": entity.DagInstanceStatusScheduled,
				})
				if err != nil {
					d.log.Errorf("[Diapatch GetDagInstanceCount] err: %v", err.Error())
				}
				if count > int64(d.listinsCount) {
					debugMode := os.Getenv("DEBUG")
					if debugMode == "true" {
						d.log.Warnf("[Diapatch] overflow user limit")
					}
					return
				}
				_ins, err := GetStore().ListDagInstance(context.TODO(), &ListDagInstanceInput{
					Status: []entity.DagInstanceStatus{
						entity.DagInstanceStatusInit,
					},
					Limit:    int64(d.listinsCount),
					UserIDs:  []string{userid},
					Priority: cons,
				})
				if err != nil {
					d.log.Errorf("[Diapatch ListDagInstance] err: %v", err.Error())
				}

				for i := range _ins {
					_ins[i].Status = entity.DagInstanceStatusScheduled
					_ins[i].Worker = nodes[i%len(nodes)]
				}

				ins = append(ins, _ins...)
			}
		}
		err := pool.Submit(goFunc(p))
		if err != nil {
			d.log.Errorf("[Diapatch Submit] err: %v", err.Error())
			continue
		}
	}

	wg.Wait()

	if len(ins) == 0 {
		return nil
	}

	if err := GetStore().BatchUpdateDagIns(context.TODO(), ins); err != nil {
		d.log.Errorf("[Diapatch BatchUpdateDagIns] err: %v", err.Error())
		return err
	}

	return nil
}

func (d *DefDispatcher) handlerErr(err error) {
	log.Errorf("dispatch failed",
		"module", "dispatch",
		"err", err)
}

// Close component
func (d *DefDispatcher) Close() {
	close(d.closeCh)
	d.wg.Wait()
}
