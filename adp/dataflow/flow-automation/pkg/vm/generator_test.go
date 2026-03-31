package vm

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm/opcode"
)

func assertInstructionsEqual(t *testing.T, actual []*opcode.Instruction, expected []*opcode.Instruction) {
	// 验证生成的指令
	if len(actual) != len(expected) {
		t.Fatalf("指令数量不匹配, 期望 %d, 实际 %d",
			len(expected), len(actual))
	}

	for i, inst := range actual {
		expectedInst := expected[i]
		if inst.OpCode != expectedInst.OpCode {
			t.Errorf("指令 %d: 操作码不匹配, 期望 %v, 实际 %v",
				i, expectedInst.OpCode, inst.OpCode)
		}

		if inst.Name != expectedInst.Name {
			t.Errorf("指令 %d: Name不匹配, 期望 %v, 实际 %v",
				i, expectedInst.Name, inst.Name)
		}

		if inst.Pos != expectedInst.Pos {
			t.Errorf("指令 %d: Pos 不匹配, 期望 %v, 实际 %v",
				i, expectedInst.Pos, inst.Pos)
		}

		if inst.Size != expectedInst.Size {
			t.Errorf("指令 %d: Size不匹配, 期望 %v, 实际 %v",
				i, expectedInst.Size, inst.Size)
		}

		// 特殊处理类型比较
		if expectedInst.Value != nil {
			switch v := expectedInst.Value.(type) {
			case reflect.Type:
				if reflect.TypeOf(inst.Value) != v {
					t.Errorf("指令 %d: 类型不匹配, 期望 %v, 实际 %v",
						i, v, reflect.TypeOf(inst.Value))
				}
			default:
				if !reflect.DeepEqual(inst.Value, expectedInst.Value) {
					t.Errorf("指令 %d: 值不匹配, 期望 %v, 实际 %v",
						i, expectedInst.Value, inst.Value)
				}
			}
		}
	}
}

func TestGenerateValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected []*opcode.Instruction
		wantErr  bool
		errMsg   string
	}{
		{
			name:  "int",
			value: 1,
			expected: []*opcode.Instruction{
				{
					OpCode: opcode.Push,
					Value:  1,
				},
			},
		},
		{
			name:  "bool",
			value: true,
			expected: []*opcode.Instruction{
				{
					OpCode: opcode.Push,
					Value:  true,
				},
			},
		},
		{
			name:  "slice",
			value: []any{1, 2, 3},
			expected: []*opcode.Instruction{
				{
					OpCode: opcode.Push,
					Value:  3,
				},
				{
					OpCode: opcode.Push,
					Value:  2,
				},
				{
					OpCode: opcode.Push,
					Value:  1,
				},
				{
					OpCode: opcode.List,
					Size:   3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM()
			g := NewGenerator(vm)
			err := g.GenerateValue(tt.value)

			// 检查错误情况
			if tt.wantErr {
				if err == nil {
					t.Fatal("期望返回错误, 但实际没有错误")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("期望错误信息包含: %q, 实际得到: %q", tt.errMsg, err.Error())
				}
				return
			}

			// 检查正常情况
			if err != nil {
				t.Fatalf("不期望错误但得到: %v", err)
			}

			assertInstructionsEqual(t, g.Instructions, tt.expected)
		})
	}
}
