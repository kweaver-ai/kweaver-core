package v2agentexecutordto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterruptData_StructFields(t *testing.T) {
	t.Parallel()

	interruptConfig := &InterruptConfig{
		RequiresConfirmation: true,
		ConfirmationMessage:  "Please confirm",
	}

	toolArgs := []ToolArg{
		{Key: "arg1", Value: "value1", Type: "string"},
		{Key: "arg2", Value: 123, Type: "number"},
	}

	data := &InterruptData{
		ToolName:        "test_tool",
		ToolDescription: "Test tool description",
		ToolArgs:        toolArgs,
		InterruptConfig: interruptConfig,
	}

	assert.Equal(t, "test_tool", data.ToolName)
	assert.Equal(t, "Test tool description", data.ToolDescription)
	assert.Len(t, data.ToolArgs, 2)
	assert.NotNil(t, data.InterruptConfig)
	assert.True(t, data.InterruptConfig.RequiresConfirmation)
	assert.Equal(t, "Please confirm", data.InterruptConfig.ConfirmationMessage)
}

func TestInterruptData_Empty(t *testing.T) {
	t.Parallel()

	data := &InterruptData{}

	assert.Empty(t, data.ToolName)
	assert.Empty(t, data.ToolDescription)
	assert.Nil(t, data.ToolArgs)
	assert.Nil(t, data.InterruptConfig)
}

func TestToolArg_StructFields(t *testing.T) {
	t.Parallel()

	arg := ToolArg{
		Key:   "test_key",
		Value: "test_value",
		Type:  "string",
	}

	assert.Equal(t, "test_key", arg.Key)
	assert.Equal(t, "test_value", arg.Value)
	assert.Equal(t, "string", arg.Type)
}

func TestToolArg_WithNumberValue(t *testing.T) {
	t.Parallel()

	arg := ToolArg{
		Key:   "count",
		Value: 42,
		Type:  "integer",
	}

	assert.Equal(t, "count", arg.Key)
	assert.Equal(t, 42, arg.Value)
	assert.Equal(t, "integer", arg.Type)
}

func TestToolArg_WithObjectValue(t *testing.T) {
	t.Parallel()

	objValue := map[string]interface{}{
		"nested": "value",
		"number": 123,
	}

	arg := ToolArg{
		Key:   "config",
		Value: objValue,
		Type:  "object",
	}

	assert.Equal(t, "config", arg.Key)
	assert.NotNil(t, arg.Value)
	assert.Equal(t, "object", arg.Type)
}

func TestInterruptConfig_StructFields(t *testing.T) {
	t.Parallel()

	config := &InterruptConfig{
		RequiresConfirmation: false,
		ConfirmationMessage:  "",
	}

	assert.False(t, config.RequiresConfirmation)
	assert.Empty(t, config.ConfirmationMessage)
}

func TestInterruptConfig_WithConfirmation(t *testing.T) {
	t.Parallel()

	config := &InterruptConfig{
		RequiresConfirmation: true,
		ConfirmationMessage:  "Do you want to continue?",
	}

	assert.True(t, config.RequiresConfirmation)
	assert.Equal(t, "Do you want to continue?", config.ConfirmationMessage)
}

func TestInterruptData_WithMultipleToolArgs(t *testing.T) {
	t.Parallel()

	toolArgs := []ToolArg{
		{Key: "arg1", Value: "val1", Type: "string"},
		{Key: "arg2", Value: "val2", Type: "string"},
		{Key: "arg3", Value: 456, Type: "integer"},
		{Key: "arg4", Value: true, Type: "boolean"},
	}

	data := &InterruptData{
		ToolName: "multi_arg_tool",
		ToolArgs: toolArgs,
	}

	assert.Len(t, data.ToolArgs, 4)
	assert.Equal(t, "arg1", data.ToolArgs[0].Key)
	assert.Equal(t, "arg4", data.ToolArgs[3].Key)
	assert.True(t, data.ToolArgs[3].Value.(bool))
}

func TestInterruptData_WithNilInterruptConfig(t *testing.T) {
	t.Parallel()

	data := &InterruptData{
		ToolName: "tool_no_config",
		ToolArgs: []ToolArg{},
	}

	assert.Equal(t, "tool_no_config", data.ToolName)
	assert.NotNil(t, data.ToolArgs)
	assert.Nil(t, data.InterruptConfig)
}
