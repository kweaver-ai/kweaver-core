package vm

import (
	"fmt"
	"reflect"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm/opcode"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"github.com/mitchellh/mapstructure"
)

type Variable struct{}

type Loop struct {
	Name           string
	Array          string
	Max            string
	Index          string
	Start          int
	JumpEnds       []int
	OutputTokenMap map[string][]*LoopOutputToken
}

type LoopOutputToken struct {
	*Token
	Key string
}

type LoopOutput struct {
	Key   string `mapstructure:"key"`
	Value string `mapstructure:"value"`
}

type Generator struct {
	vm           *VM
	regId        int
	variables    map[string]*Variable
	loops        []*Loop
	Instructions []*opcode.Instruction
}

func NewGenerator(vm *VM) *Generator {
	return &Generator{
		vm:        vm,
		regId:     0,
		loops:     make([]*Loop, 0),
		variables: make(map[string]*Variable),
	}
}

func (g *Generator) append(instructions ...*opcode.Instruction) {
	g.Instructions = append(g.Instructions, instructions...)
}

func (g *Generator) newReg() string {
	reg := g.vm.NewReg()
	g.variables[reg] = nil
	return reg
}

func (g *Generator) Generate(program *Program) (err error) {
	err = g.GenerateTrigger(program.Steps[0])
	if err != nil {
		return
	}

	return g.GenerateSteps(program.Steps[1:])
}

func (g *Generator) GenerateTrigger(step *entity.Step) (err error) {
	err = g.GenerateValue(step.Parameters)

	if err != nil {
		return
	}

	id := step.ID

	if step.DataSource != nil {
		id = step.DataSource.ID
	}

	name := fmt.Sprintf("__%v", id)
	g.append(
		&opcode.Instruction{OpCode: opcode.Call, Name: step.Operator, Pos: 1, Size: 1, Value: step.ID},
		&opcode.Instruction{OpCode: opcode.Save, Name: name},
	)
	g.variables[name] = nil
	return
}

func (g *Generator) GenerateSteps(steps []*entity.Step) (err error) {
	for _, step := range steps {
		err = g.GenerateStep(step)
		if err != nil {
			return
		}
	}

	return
}

func (g *Generator) GenerateStep(step *entity.Step) (err error) {
	switch step.Operator {
	case common.InternalAssignOpt:
		return g.GenerateAssign(step)
	case common.BranchOpt:
		for _, branch := range step.Branches {
			err = g.GenerateBranch(&branch)
			if err != nil {
				return
			}
		}
	case common.Loop:
		return g.GenerateLoop(step)
	case common.InternalReturnOpt:
		return g.GenerateReturn(step)
	default:
		return g.GenerateCall(step)
	}

	return
}

func (g *Generator) GenerateAssign(step *entity.Step) (err error) {
	targetRaw, ok := step.Parameters["target"]

	if !ok {
		return errors.New(errors.SyntaxError, step.ID,
			"parameters.target is required",
			nil, nil)
	}

	targetString, ok := targetRaw.(string)

	if !ok {
		return errors.New(errors.SyntaxError, step.ID,
			"parameters.target must be variable",
			nil, nil)
	}

	tokens := Parse(targetString)

	if len(tokens) != 1 || tokens[0].Type != TokenVariable {
		return errors.New(errors.SyntaxError, step.ID,
			"parameters.target must be variable",
			nil, nil)
	}

	target := tokens[0]
	targetName, _ := target.Value.(string)

	if _, exists := g.variables[targetName]; !exists {
		return errors.New(errors.SyntaxError, step.ID,
			fmt.Sprintf("variable %s not exists", targetName),
			nil, nil)
	}

	value, ok := step.Parameters["value"]

	if !ok {
		return errors.New(errors.SyntaxError, step.ID,
			"parameters.value is required",
			nil, nil)
	}

	err = g.GenerateValue(value)

	if err != nil {
		return
	}

	if len(target.AccessList) == 0 {
		g.Instructions = append(g.Instructions,
			&opcode.Instruction{
				OpCode: opcode.Mark,
				Name:   opcode.MARK_BEFORE_ASSIGN,
				Value:  []any{step.ID, targetName},
			},
			&opcode.Instruction{
				OpCode: opcode.Save,
				Name:   targetName,
			},
		)
		return
	}

	err = g.GenerateAccessList(target.AccessList)

	if err != nil {
		return
	}

	g.Instructions = append(g.Instructions,
		&opcode.Instruction{
			OpCode: opcode.Load,
			Name:   targetName,
		},
		&opcode.Instruction{
			OpCode: opcode.Call,
			Name:   "set",
			Pos:    3,
			Size:   1,
		},
		&opcode.Instruction{
			OpCode: opcode.Mark,
			Name:   opcode.MARK_BEFORE_ASSIGN,
			Value:  []any{step.ID, targetName},
		},
		&opcode.Instruction{
			OpCode: opcode.Save,
			Name:   targetName,
		},
	)

	return
}

func (g *Generator) GenerateBranch(branch *entity.Branch) (err error) {

	if len(branch.Conditions) == 0 {
		g.append(&opcode.Instruction{OpCode: opcode.Mark, Name: opcode.MARK_BRANCH_START, Value: branch.ID})
		for _, step := range branch.Steps {
			err = g.GenerateStep(&step)

			if err != nil {
				return
			}
		}
		return
	}

	for _, conditions := range branch.Conditions {
		for _, condition := range conditions {
			err = g.GenerateCondition(&condition)

			if err != nil {
				return
			}
		}
		g.append(&opcode.Instruction{OpCode: opcode.Call, Name: "and", Pos: len(conditions), Size: 1})
	}

	g.append(&opcode.Instruction{OpCode: opcode.Call, Name: "or", Pos: len(branch.Conditions), Size: 1})
	g.append(&opcode.Instruction{OpCode: opcode.Jnz, Pos: len(g.Instructions) + 3})

	g.append(&opcode.Instruction{OpCode: opcode.Mark, Name: opcode.MARK_BRANCH_SKIP, Value: branch.ID})
	jmpEndIndex := len(g.Instructions)
	g.append(&opcode.Instruction{OpCode: opcode.Jmp, Pos: -1})
	g.append(&opcode.Instruction{OpCode: opcode.Mark, Name: opcode.MARK_BRANCH_START, Value: branch.ID})

	for _, step := range branch.Steps {
		err = g.GenerateStep(&step)

		if err != nil {
			return
		}
	}

	g.Instructions[jmpEndIndex].Pos = len(g.Instructions)
	g.append(&opcode.Instruction{OpCode: opcode.Pop})
	return
}

func (g *Generator) GenerateCondition(condition *entity.TaskCondition) (err error) {
	err = g.GenerateValue(condition.Parameter.B)

	if err != nil {
		return
	}

	err = g.GenerateValue(condition.Parameter.A)

	if err != nil {
		return
	}

	g.append(&opcode.Instruction{OpCode: opcode.Call, Name: string(condition.Op), Pos: 2, Size: 1})
	return
}

func (g *Generator) GenerateLoop(step *entity.Step) (err error) {

	mode, hasMode := step.Parameters["mode"]
	isInfinit := false

	if !hasMode || mode != "array" {
		mode = "limit"
	}

	loop := &Loop{
		Name:           fmt.Sprintf("__%v", step.ID),
		Start:          0,
		JumpEnds:       []int{},
		Max:            g.newReg(),
		Index:          g.newReg(),
		OutputTokenMap: g.parseLoopOutputs(step),
	}

	if mode == "limit" {
		limit, ok := step.Parameters["limit"]

		if ok {
			err = g.GenerateValue(limit)
			if err != nil {
				return
			}
			g.append(&opcode.Instruction{OpCode: opcode.Save, Name: loop.Max})
		} else {
			isInfinit = true
		}
	} else {
		arr, ok := step.Parameters["array"]

		if !ok {
			return
		}

		loop.Array = g.newReg()
		err = g.GenerateValue(arr)
		if err != nil {
			return
		}

		g.append(
			&opcode.Instruction{OpCode: opcode.Call, Name: "array", Pos: 1, Size: 1},
			&opcode.Instruction{OpCode: opcode.Save, Name: loop.Array},
			&opcode.Instruction{OpCode: opcode.Load, Name: loop.Array},
			&opcode.Instruction{OpCode: opcode.Call, Name: "len", Pos: 1, Size: 1},
			&opcode.Instruction{OpCode: opcode.Save, Name: loop.Max},
		)
	}

	g.loops = append([]*Loop{loop}, g.loops...)

	g.append(
		&opcode.Instruction{OpCode: opcode.Push, Value: 0},
		&opcode.Instruction{OpCode: opcode.Save, Name: loop.Index},
	)

	g.variables[loop.Name] = nil

	g.append(
		&opcode.Instruction{OpCode: opcode.Dict, Size: 0},
		&opcode.Instruction{OpCode: opcode.Save, Name: loop.Name},
	)

	loop.Start = len(g.Instructions)

	if !isInfinit {
		g.append(
			&opcode.Instruction{OpCode: opcode.Load, Name: loop.Max},
			&opcode.Instruction{OpCode: opcode.Load, Name: loop.Index},
			&opcode.Instruction{OpCode: opcode.Call, Name: "lt", Pos: 2, Size: 1},
		)

		g.append(&opcode.Instruction{OpCode: opcode.Jnz, Pos: len(g.Instructions) + 2})

		loop.JumpEnds = append(loop.JumpEnds, len(g.Instructions))

		g.append(
			&opcode.Instruction{OpCode: opcode.Jmp, Pos: -1},
			&opcode.Instruction{OpCode: opcode.Pop},
		)
	}

	if mode == "array" {
		g.append(
			&opcode.Instruction{OpCode: opcode.Load, Name: loop.Index},
			&opcode.Instruction{OpCode: opcode.List, Size: 1},
			&opcode.Instruction{OpCode: opcode.Load, Name: loop.Array},
			&opcode.Instruction{OpCode: opcode.Call, Name: "get", Pos: 2, Size: 1},
			&opcode.Instruction{OpCode: opcode.Push, Value: "value"},
			&opcode.Instruction{OpCode: opcode.List, Size: 1},
		)
	}

	g.append(
		&opcode.Instruction{OpCode: opcode.Load, Name: loop.Index},
		&opcode.Instruction{OpCode: opcode.Push, Value: "index"},
		&opcode.Instruction{OpCode: opcode.List, Size: 1},
		&opcode.Instruction{OpCode: opcode.Load, Name: loop.Name},
		&opcode.Instruction{OpCode: opcode.Call, Name: "set", Pos: 3, Size: 1},
	)

	if mode == "array" {
		g.append(
			&opcode.Instruction{OpCode: opcode.Call, Name: "set", Pos: 3, Size: 1},
		)
	}

	g.append(
		&opcode.Instruction{OpCode: opcode.Mark, Name: opcode.MARK_LOOP_START, Value: step.ID},
		&opcode.Instruction{OpCode: opcode.Save, Name: loop.Name},
		&opcode.Instruction{OpCode: opcode.LoopTrace, Name: loop.Name, Value: []any{loop.Index}},
	)

	for _, step := range step.Steps {
		err = g.GenerateStep(&step)
		if err != nil {
			return
		}
	}

	g.GenerateLoopOutputs(loop.Name, loop)

	g.append(
		&opcode.Instruction{OpCode: opcode.Load, Name: loop.Name},
		&opcode.Instruction{OpCode: opcode.Mark, Name: opcode.MARK_LOOP_END, Value: step.ID},
		&opcode.Instruction{OpCode: opcode.Pop},
	)

	g.append(
		&opcode.Instruction{OpCode: opcode.Load, Name: loop.Index},
		&opcode.Instruction{OpCode: opcode.Push, Value: 1},
		&opcode.Instruction{OpCode: opcode.Call, Name: "add", Pos: 2, Size: 1},
		&opcode.Instruction{OpCode: opcode.Save, Name: loop.Index},
		&opcode.Instruction{OpCode: opcode.Jmp, Pos: loop.Start},
	)

	for _, jmpEndIndex := range loop.JumpEnds {
		g.Instructions[jmpEndIndex].Pos = len(g.Instructions)
	}

	if !isInfinit {
		g.append(&opcode.Instruction{OpCode: opcode.Pop})
	}

	g.append(
		&opcode.Instruction{OpCode: opcode.Del, Name: loop.Max},
		&opcode.Instruction{OpCode: opcode.Del, Name: loop.Index},
	)

	g.loops = g.loops[1:]

	return g.GenerateAllLoopOutputs(loop.Name)
}

func (g *Generator) parseLoopOutputs(step *entity.Step) (outputTokenMap map[string][]*LoopOutputToken) {
	outputTokenMap = make(map[string][]*LoopOutputToken)
	outputs := make([]*LoopOutput, 0)

	if p, ok := step.Parameters["outputs"]; ok {
		_ = mapstructure.WeakDecode(p, &outputs)
	}

	for _, output := range outputs {
		if output.Key == "" {
			continue
		}

		if output.Value == "" {
			continue
		}

		tokens := Parse(output.Value)

		if len(tokens) != 1 || tokens[0].Type != TokenVariable {
			continue
		}

		token := tokens[0]
		stepName, ok := token.Value.(string)

		if !ok {
			continue
		}

		outputTokenMap[stepName] = append(outputTokenMap[stepName], &LoopOutputToken{
			Key:   output.Key,
			Token: &token,
		})
	}
	return
}

func (g *Generator) GenerateAllLoopOutputs(name string) (err error) {
	for _, loop := range g.loops {
		err = g.GenerateLoopOutputs(name, loop)
		if err != nil {
			return
		}
	}
	return nil
}

func (g *Generator) GenerateLoopOutputs(name string, loop *Loop) (err error) {

	tokens, ok := loop.OutputTokenMap[name]
	if !ok {
		return
	}

	for _, token := range tokens {
		if len(token.AccessList) > 0 {
			err = g.GenerateAccessList(token.AccessList)

			if err != nil {
				return err
			}

			g.append(
				&opcode.Instruction{OpCode: opcode.Load, Name: name},
				&opcode.Instruction{OpCode: opcode.Call, Name: "get", Pos: 2, Size: 1},
			)
		} else {
			g.append(&opcode.Instruction{OpCode: opcode.Load, Name: name})
		}

		g.append(
			&opcode.Instruction{OpCode: opcode.Push, Value: token.Key},
			&opcode.Instruction{OpCode: opcode.Push, Value: "outputs"},
			&opcode.Instruction{OpCode: opcode.List, Size: 2},
			&opcode.Instruction{OpCode: opcode.Load, Name: loop.Name},
			&opcode.Instruction{OpCode: opcode.Call, Name: "get", Pos: 2, Size: 1},
			&opcode.Instruction{OpCode: opcode.Call, Name: "append", Pos: 2, Size: 1},
		)

		g.append(
			&opcode.Instruction{OpCode: opcode.Push, Value: token.Key},
			&opcode.Instruction{OpCode: opcode.Push, Value: "outputs"},
			&opcode.Instruction{OpCode: opcode.List, Size: 2},
			&opcode.Instruction{OpCode: opcode.Load, Name: loop.Name},
			&opcode.Instruction{OpCode: opcode.Call, Name: "set", Pos: 3, Size: 1},
			&opcode.Instruction{OpCode: opcode.Save, Name: loop.Name},
		)
	}

	return nil
}

// 日志兼容处理，记录循环内节点每次运行的结果，变量名称为 __循环ID_i循环次数_s节点ID
func (g *Generator) GenerateLoopTrace(step *entity.Step) (err error) {
	l := len(g.loops)

	if l == 0 {
		return nil
	}

	loop := g.loops[l-1]
	name := fmt.Sprintf("__%v", step.ID)
	g.append(
		&opcode.Instruction{OpCode: opcode.LoopTrace, Name: loop.Name, Value: []any{loop.Index, step.ID, name}},
	)

	return nil
}

func (g *Generator) GenerateReturn(step *entity.Step) (err error) {
	err = g.GenerateValue(step.Parameters)
	if err != nil {
		return
	}
	g.append(
		&opcode.Instruction{OpCode: opcode.Mark, Name: opcode.MARK_BEFORE_RETURN, Value: step.ID},
		&opcode.Instruction{OpCode: opcode.Return},
	)
	return
}

func (g *Generator) GenerateCall(step *entity.Step) (err error) {
	err = g.GenerateValue(step.Parameters)
	if err != nil {
		return
	}
	name := fmt.Sprintf("__%v", step.ID)
	var value any = step.ID
	if step.Title != "" {
		value = map[string]any{
			"id":    step.ID,
			"title": step.Title,
		}
	}
	g.append(
		&opcode.Instruction{OpCode: opcode.Call, Name: step.Operator, Pos: 1, Size: 1, Value: value},
		&opcode.Instruction{OpCode: opcode.Save, Name: name},
	)
	g.variables[name] = nil
	err = g.GenerateAllLoopOutputs(name)

	if err != nil {
		return
	}

	err = g.GenerateLoopTrace(step)

	if err != nil {
		return
	}

	return nil
}

func (g *Generator) GenerateValue(value interface{}) (err error) {
	if value == nil {
		g.append(&opcode.Instruction{OpCode: opcode.Push, Value: value})
		return
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return g.GenerateString(rv.String())

	case reflect.Slice, reflect.Array:
		l := rv.Len()
		for i := l - 1; i >= 0; i-- {
			elem := rv.Index(i).Interface()
			err = g.GenerateValue(elem)
			if err != nil {
				return
			}
		}
		g.append(&opcode.Instruction{OpCode: opcode.List, Size: l})

	case reflect.Map:
		keys := rv.MapKeys()
		l := len(keys)
		for _, key := range keys {
			val := rv.MapIndex(key).Interface()
			err = g.GenerateValue(val)
			if err != nil {
				return
			}
			err = g.GenerateValue(key.Interface())
			if err != nil {
				return
			}
		}
		g.append(&opcode.Instruction{OpCode: opcode.Dict, Size: l})

	default:
		g.append(&opcode.Instruction{OpCode: opcode.Push, Value: value})
	}
	return
}

func (g *Generator) GenerateString(s string) (err error) {
	s = utils.RemoveZeroWidthChars(s)

	// 不支持超过8k长度的字符串模板拼接
	if len(s) > 1024*8 {
		g.append(&opcode.Instruction{OpCode: opcode.Push, Value: s})
		return
	}

	tokens := Parse(s)
	l := len(tokens)

	if l == 0 {
		g.append(&opcode.Instruction{OpCode: opcode.Push, Value: s})
		return
	}

	for i := l - 1; i >= 0; i -= 1 {
		token := tokens[i]
		if token.Type != TokenVariable {
			g.append(&opcode.Instruction{OpCode: opcode.Push, Value: token.Value})
			continue
		}

		name := token.Value.(string)
		if len(token.AccessList) == 0 {
			if g.vm.IsGlobal(name) {
				g.append(&opcode.Instruction{OpCode: opcode.Push, Value: []interface{}{}})
				g.append(&opcode.Instruction{OpCode: opcode.LoadGlobal, Name: name})
			} else {
				g.append(&opcode.Instruction{OpCode: opcode.Load, Name: name})
			}
			continue
		}

		err = g.GenerateAccessList(token.AccessList)
		if err != nil {
			return
		}

		if g.vm.IsGlobal(name) {
			g.append(&opcode.Instruction{OpCode: opcode.LoadGlobal, Name: name})
		} else {
			g.append(&opcode.Instruction{OpCode: opcode.Load, Name: name})
			g.append(&opcode.Instruction{OpCode: opcode.Call, Name: "get", Pos: 2, Size: 1})
		}
	}

	if l > 1 {
		g.append(&opcode.Instruction{OpCode: opcode.Call, Name: "str", Pos: l, Size: 1})
	}

	return
}

func (g *Generator) GenerateAccessList(tokens []Token) (err error) {
	l := len(tokens)
	for i := l - 1; i >= 0; i -= 1 {
		token := tokens[i]
		g.append(&opcode.Instruction{OpCode: opcode.Push, Value: token.Value})
	}

	g.append(&opcode.Instruction{OpCode: opcode.List, Size: l})
	return
}
