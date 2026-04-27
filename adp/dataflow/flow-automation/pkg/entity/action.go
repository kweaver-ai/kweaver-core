package entity

import (
	"context"
	"errors"
	"time"
)

// RunFunc is task concrete action
type RunFunc func(ctx ExecuteContext, params interface{}) error

// Action is the interface that user action must implement
type Action interface {
	Name() string
	Run(ctx ExecuteContext, params interface{}, token *Token) (interface{}, error)
}

// BeforeAction run before run action
type BeforeAction interface {
	RunBefore(ctx ExecuteContext, params interface{}) (TaskInstanceStatus, error)
}

// AfterAction run after run action
type AfterAction interface {
	RunAfter(ctx ExecuteContext, params interface{}) (TaskInstanceStatus, error)
}

// ParameterAction means action has parameter
type ParameterAction interface {
	ParameterNew() interface{}
}

// RetryBeforeAction execute before retry
type RetryBeforeAction interface {
	RetryBefore(ctx ExecuteContext, params interface{}) error
}

// CancelAction cancle
type CancelAction interface {
	Cancel(taskIns TaskInstance, executeMethods ExecuteMethods) error
}

var (
	EndLoop = errors.New("end loop")
)

type LoopDoOptionOp func(loop *LoopDoOption)

// LoopInterval indicate the interval of loop
func LoopInterval(duration time.Duration) LoopDoOptionOp {
	return func(loop *LoopDoOption) {
		if duration != 0 {
			loop.interval = duration
		}
	}
}

// LoopDoOption
type LoopDoOption struct {
	interval time.Duration
}

// LoopDo help you to complete loop action,for example
//
//	LoopDo(ctx, func(){
//	    log.Println("check status")
//	})
func LoopDo(ctx ExecuteContext, do func() error, ops ...LoopDoOptionOp) error {
	opt := &LoopDoOption{
		interval: time.Second,
	}
	for _, o := range ops {
		o(opt)
	}

	tick := time.Tick(opt.interval)
	for {
		select {
		case <-tick:
			if err := do(); err != nil {
				if errors.Is(err, EndLoop) {
					return nil
				}
				return err
			}
		case <-ctx.Context().Done():
			if errors.Is(ctx.Context().Err(), context.Canceled) {
				ctx.Trace(ctx.Context(), "task is canceled")
			}
			return ctx.Context().Err()
		}
	}
}
