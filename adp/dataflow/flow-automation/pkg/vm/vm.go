package vm

import (
	"context"
	"fmt"

	commonErrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm/funcs"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm/hook"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm/opcode"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm/state"
)

type ContextKey string

const VM_CONTEXT_KEY ContextKey = "VM"

type Func = funcs.Func

type Program struct {
	Steps []*entity.Step `json:"steps"`
}

type Global interface {
	Get(vm *VM, name string, path []interface{}) interface{}
}

type VM struct {
	context context.Context
	globals map[string]Global
	funcs   map[string]funcs.Func
	extfunc funcs.Func
	hook    hook.Hook

	State        state.State           `json:"state" bson:"state"`
	PC           int                   `json:"pc" bson:"pc"`
	RegID        int                   `json:"reg_id" bson:"reg_id"`
	Stack        []any                 `json:"stack" bson:"stack"`
	Env          map[string]any        `json:"env" bson:"env"`
	Callstack    []*funcs.CallFrame    `json:"callstack" bson:"callstack"`
	Instructions []*opcode.Instruction `json:"instructions" bson:"instructions"`
	Traces       []string              `json:"traces" bson:"traces"`
	Err          *errors.Error         `json:"err" bson:"err"`
}

func NewVM() *VM {
	vm := &VM{
		context: context.Background(),
		globals: make(map[string]Global),
		funcs:   make(map[string]Func),

		State:        state.Init,
		PC:           0,
		RegID:        0,
		Stack:        make([]interface{}, 0),
		Env:          make(map[string]interface{}),
		Callstack:    make([]*funcs.CallFrame, 0),
		Instructions: make([]*opcode.Instruction, 0),
		Traces:       make([]string, 0),
	}

	vm.AddFuncs(funcs.BuiltIns)

	return vm
}

func (vm *VM) Context() context.Context {
	return vm.context
}

func (vm *VM) SetContext(ctx context.Context) {
	vm.context = context.WithValue(ctx, VM_CONTEXT_KEY, vm)
}

func (vm *VM) CallFrame() *funcs.CallFrame {
	if l := len(vm.Callstack); l > 0 {
		return vm.Callstack[l-1]
	}
	return nil
}

func (vm *VM) NewReg() string {
	name := fmt.Sprintf("__reg%d", vm.RegID)
	vm.RegID += 1
	return name
}

func (vm *VM) AddFunc(name string, f Func) {
	if _, ok := vm.funcs[name]; ok {
		return
	}
	vm.funcs[name] = f
}

func (vm *VM) AddFuncs(funcs map[string]Func) {
	for name, f := range funcs {
		vm.AddFunc(name, f)
	}
}

func (vm *VM) SetExtfunc(f Func) {
	vm.extfunc = f
}

func (vm *VM) SetHook(h hook.Hook) {
	vm.hook = h
}

func (vm *VM) AddGlobal(name string, global Global) {
	if _, ok := vm.globals[name]; ok {
		return
	}

	vm.globals[name] = global
}

func (vm *VM) AddGlobals(globals map[string]Global) {
	for name, global := range globals {
		vm.AddGlobal(name, global)
	}
}

func (vm *VM) IsGlobal(name string) bool {
	_, ok := vm.globals[name]
	return ok
}

func (vm *VM) Load(program *Program) (err error) {
	vm.Reset()
	g := NewGenerator(vm)
	err = g.Generate(program)

	if err != nil {
		return
	}

	vm.Instructions = append(vm.Instructions, g.Instructions...)
	return
}

func (vm *VM) LoadInstructions(instructions []*opcode.Instruction) {
	vm.Reset()
	vm.Instructions = append(vm.Instructions, instructions...)
}

func (vm *VM) Reset() {
	vm.State = state.Init
	vm.PC = 0
	vm.Stack = make([]interface{}, 0)
	vm.Env = make(map[string]interface{})
	vm.Callstack = make([]*funcs.CallFrame, 0)
	vm.Traces = make([]string, 0)
	vm.Err = nil
}

func (vm *VM) Run(params ...interface{}) {
	if vm.State != state.Init {
		return
	}
	vm.State = state.Run
	vm.Start()
}

func (vm *VM) Trace(msg string) {
	vm.Traces = append(vm.Traces, msg)
}

func (vm *VM) Result() (state.State, any, error) {
	if vm.State == state.Error {
		return vm.State, nil, vm.Err
	}

	if vm.State == state.Done || vm.PC == len(vm.Instructions) {
		var result any = nil
		if l := len(vm.Stack); l > 0 {
			result = vm.Stack[l-1]
		}
		return state.Done, result, nil
	}

	return vm.State, nil, nil
}

func (vm *VM) Retry() {
	if l := len(vm.Callstack); l > 0 && vm.State == state.Error {
		frame := vm.Callstack[l-1]
		vm.Callstack = vm.Callstack[0 : l-1]
		vm.call(frame.Name, frame.Label, frame.Title, frame.NRets, frame.Args...)
		vm.State = state.Run
		vm.Start()
	}
}

func (vm *VM) Resume(rets ...interface{}) {

	l := len(vm.Callstack)
	if vm.State != state.Wait || l == 0 {
		return
	}

	frame := vm.Callstack[l-1]

	if h, ok := vm.hook.(hook.Resume); ok {
		h.HookResume(frame.Name, frame.Label, rets)
	}

	vm.Callstack = vm.Callstack[0 : l-1]

	if frame.NRets > len(rets) {
		for i := len(rets); i < frame.NRets; i += 1 {
			rets = append(rets, nil)
		}
	}

	for i := frame.NRets - 1; i >= 0; i -= 1 {
		vm.Stack = append(vm.Stack, rets[i])
	}

	vm.State = state.Run
	vm.Start()
}

func (vm *VM) ResumeError(err error) {
	l := len(vm.Callstack)
	if vm.State != state.Wait || l == 0 {
		return
	}

	frame := vm.Callstack[l-1]
	vm.State = state.Error
	if vmErr, ok := err.(*errors.Error); ok {
		vm.Err = vmErr
	} else {
		ierror := commonErrors.ParseError(err).(*commonErrors.IError)
		detail := map[string]any{
			"name":  frame.Name,
			"args":  frame.Args,
			"cause": ierror,
		}
		vm.Err = errors.New(errors.FunctionCallError, frame.Label, fmt.Sprintf("call '%s' failed", frame.Name), detail, vm.Traces)
		vm.Traces = nil
	}

	if h, ok := vm.hook.(hook.ResumeError); ok {
		h.HookResumeError(frame.Name, frame.Label, vm.Err)
	}

	if h, ok := vm.hook.(hook.VMStop); ok {
		h.HookVMStop()
	}
}

func (vm *VM) pushStack(value interface{}) {
	vm.Stack = append(vm.Stack, value)
}

func (vm *VM) popStack() interface{} {

	if l := len(vm.Stack); l > 0 {
		value := vm.Stack[l-1]
		vm.Stack = vm.Stack[0 : l-1]
		return value
	}

	return nil
}

func (vm *VM) Start() {
	for {

		if vm.State == state.Run && vm.PC >= len(vm.Instructions) {
			vm.State = state.Done
		}

		if vm.State != state.Run || vm.PC < 0 || vm.PC >= len(vm.Instructions) {
			break
		}

		instr := vm.Instructions[vm.PC]
		vm.PC += 1
		switch instr.OpCode {
		case opcode.Push:
			vm.pushStack(instr.Value)
		case opcode.Pop:
			_ = vm.popStack()
		case opcode.Save:
			vm.Env[instr.Name] = vm.popStack()
		case opcode.Load:
			vm.pushStack(vm.Env[instr.Name])
		case opcode.Del:
			delete(vm.Env, instr.Name)
		case opcode.Jmp:
			vm.PC = instr.Pos
		case opcode.Jnz:
			if l := len(vm.Stack); l > 0 {
				v := vm.Stack[l-1]
				if flag, ok := v.(bool); ok && flag {
					vm.PC = instr.Pos
					continue
				}
			} else {
				vm.Err = errors.New(errors.StackOverflowError, "", "stack overflow", nil, vm.Traces)
				vm.Traces = nil
				vm.State = state.Error
			}
		case opcode.List:
			l := len(vm.Stack)
			size := instr.Size
			list := make([]interface{}, 0)

			for i := 1; i <= size; i += 1 {
				list = append(list, vm.Stack[l-i])
			}
			vm.Stack = append(vm.Stack[0:l-size], list)
		case opcode.Dict:
			l := len(vm.Stack)
			size := instr.Size
			dict := make(map[string]interface{})

			for i := 1; i <= size*2; i += 2 {
				key := fmt.Sprintf("%v", vm.Stack[l-i])
				value := vm.Stack[l-i-1]
				dict[key] = value
			}
			vm.Stack = append(vm.Stack[0:l-size*2], dict)
		case opcode.LoadGlobal:
			l := len(vm.Stack)
			path := vm.Stack[l-1].([]interface{})
			value := vm.loadGlobal(instr.Name, path)
			vm.Stack = append(vm.Stack[0:l-1], value)
		case opcode.Call:
			l := len(vm.Stack)
			args := make([]interface{}, 0)

			for i := 1; i <= instr.Pos; i += 1 {
				args = append(args, vm.Stack[l-i])
			}
			vm.Stack = vm.Stack[0 : l-instr.Pos]

			var label, title string

			switch v := instr.Value.(type) {
			case string:
				label = v
			case map[string]any:
				label = v["id"].(string)
				title = v["title"].(string)
			}

			vm.call(instr.Name, label, title, instr.Size, args...)
		case opcode.Return:
			vm.State = state.Done
		case opcode.Mark:
			vm.mark(instr)
		case opcode.LoopTrace:
			params := instr.Value.([]any)

			switch l := len(params); l {
			case 1:
				indexName := params[0].(string)
				indexValue, ok := vm.Env[indexName]
				if !ok {
					continue
				}

				loopTraceVar := fmt.Sprintf("%s_i%v", instr.Name, indexValue)
				vm.Env[loopTraceVar] = map[string]any{
					"index": indexValue,
				}
			case 3:
				indexName, stepID, traceVar := params[0].(string), params[1].(string), params[2].(string)
				indexValue, ok := vm.Env[indexName]
				if !ok {
					continue
				}
				traceValue, ok := vm.Env[traceVar]
				if !ok {
					continue
				}

				loopTraceVar := fmt.Sprintf("%s_i%v_s%s", instr.Name, indexValue, stepID)
				vm.Env[loopTraceVar] = traceValue
			default:
				continue
			}
		}
	}

	if h, ok := vm.hook.(hook.VMStop); ok {
		h.HookVMStop()
	}
}

func (vm *VM) loadGlobal(name string, path []interface{}) interface{} {
	if global, ok := vm.globals[name]; ok {
		value := global.Get(vm, name, path)
		return value
	} else {
		return nil
	}
}

func (vm *VM) call(name string, label string, title string, nrets int, args ...interface{}) {
	var err error
	var wait bool
	var rets []interface{}

	vm.Callstack = append(vm.Callstack, &funcs.CallFrame{
		Name:  name,
		NRets: nrets,
		Args:  args,
		Label: label,
		Title: title,
	})

	if f, ok := vm.funcs[name]; ok {
		wait, rets, err = f.Call(vm.context, name, nrets, args...)
	} else if vm.extfunc != nil {
		wait, rets, err = vm.extfunc.Call(vm.context, name, nrets, args...)
	} else {
		err = errors.New(errors.FunctionNotFoundError, label, fmt.Sprintf("%s not found", name), nil, vm.Traces)
		vm.Traces = nil
	}

	if err != nil {
		vm.State = state.Error

		if vmErr, ok := err.(*errors.Error); ok {
			vm.Err = vmErr
		} else {
			ierror := commonErrors.ParseError(err).(*commonErrors.IError)
			detail := map[string]any{
				"name":  name,
				"args":  args,
				"cause": ierror,
			}
			vm.Err = errors.New(errors.FunctionCallError, label, fmt.Sprintf("call '%s' failed", name), detail, vm.Traces)
			vm.Traces = nil
		}
		return
	}

	if wait {
		vm.State = state.Wait
		return
	}

	vm.Callstack = vm.Callstack[0 : len(vm.Callstack)-1]

	if nrets > len(rets) {
		for i := len(rets); i < nrets; i += 1 {
			rets = append(rets, nil)
		}
	}

	for i := nrets - 1; i >= 0; i -= 1 {
		vm.Stack = append(vm.Stack, rets[i])
	}
}

func (vm *VM) mark(instr *opcode.Instruction) {
	switch instr.Name {
	case opcode.MARK_BEFORE_ASSIGN:
		h, ok := vm.hook.(hook.BeforeAssign)
		if !ok {
			return
		}

		var value any = nil

		if l := len(vm.Stack); l > 0 {
			value = vm.Stack[l-1]
		}

		arr, ok := instr.Value.([]any)
		if !ok || len(arr) < 2 {
			return
		}

		id, ok := arr[0].(string)
		if !ok {
			return
		}

		target, ok := arr[1].(string)
		if !ok {
			return
		}

		h.HookBeforeAssign(id, target, value)
	case opcode.MARK_BEFORE_RETURN:
		h, ok := vm.hook.(hook.BeforeReturn)
		if !ok {
			return
		}

		var value any
		if l := len(vm.Stack); l > 0 {
			value = vm.Stack[l-1]
		}

		id, ok := instr.Value.(string)
		if !ok {
			return
		}

		h.HookBeforeReturn(id, value)

	case opcode.MARK_LOOP_START:
		h, ok := vm.hook.(hook.LoopStart)
		if !ok {
			return
		}

		var value any
		if l := len(vm.Stack); l > 0 {
			value = vm.Stack[l-1]
		}

		id, ok := instr.Value.(string)
		if !ok {
			return
		}

		h.HookLoopStart(id, value)

	case opcode.MARK_LOOP_END:
		h, ok := vm.hook.(hook.LoopEnd)

		if !ok {
			return
		}

		var value any
		if l := len(vm.Stack); l > 0 {
			value = vm.Stack[l-1]
		}

		id, ok := instr.Value.(string)
		if !ok {
			return
		}

		h.HookLoopEnd(id, value)

	case opcode.MARK_BRANCH_START:
		h, ok := vm.hook.(hook.BranchStart)

		if !ok {
			return
		}

		id, ok := instr.Value.(string)
		if !ok {
			return
		}

		h.HookBranchStart(id)

	case opcode.MARK_BRANCH_SKIP:
		h, ok := vm.hook.(hook.BranchSkip)

		if !ok {
			return
		}

		id, ok := instr.Value.(string)
		if !ok {
			return
		}

		h.HookBranchSkip(id)
	}
}
