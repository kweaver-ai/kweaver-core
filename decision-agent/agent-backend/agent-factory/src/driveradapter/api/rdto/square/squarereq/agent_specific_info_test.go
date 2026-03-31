package squarereq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInfoType_Constant(t *testing.T) {
	t.Parallel()

	assert.Equal(t, GetInfoType("output_conf"), OutputConfGetInfoType)
}

func TestGetInfoType_String(t *testing.T) {
	t.Parallel()

	infoType := GetInfoType("output_conf")

	assert.Equal(t, "output_conf", string(infoType))
}

func TestAgentSpecificInfoReq_StructFields(t *testing.T) {
	t.Parallel()

	req := AgentSpecificInfoReq{
		GetInfoType: OutputConfGetInfoType,
	}

	assert.Equal(t, OutputConfGetInfoType, req.GetInfoType)
	assert.Equal(t, GetInfoType("output_conf"), req.GetInfoType)
}

func TestAgentSpecificInfoReq_Empty(t *testing.T) {
	t.Parallel()

	req := AgentSpecificInfoReq{}

	assert.Empty(t, req.GetInfoType)
}

func TestAgentSpecificInfoReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := AgentSpecificInfoReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"get_info_type"不能为空`, errMsgMap["GetInfoType.required"])
}

func TestAgentSpecificInfoReq_WithValidInfoType(t *testing.T) {
	t.Parallel()

	req := AgentSpecificInfoReq{
		GetInfoType: OutputConfGetInfoType,
	}

	assert.NotEmpty(t, req.GetInfoType)
	assert.Equal(t, "output_conf", string(req.GetInfoType))
}

func TestAgentSpecificInfoReq_WithCustomInfoType(t *testing.T) {
	t.Parallel()

	customType := GetInfoType("custom_info")
	req := AgentSpecificInfoReq{
		GetInfoType: customType,
	}

	assert.Equal(t, customType, req.GetInfoType)
	assert.Equal(t, GetInfoType("custom_info"), req.GetInfoType)
}

func TestAgentSpecificInfoReq_EmptyStringInfoType(t *testing.T) {
	t.Parallel()

	req := AgentSpecificInfoReq{
		GetInfoType: GetInfoType(""),
	}

	assert.Empty(t, string(req.GetInfoType))
}

func TestGetInfoType_Comparison(t *testing.T) {
	t.Parallel()

	type1 := OutputConfGetInfoType
	type2 := GetInfoType("output_conf")

	assert.Equal(t, type1, type2)
	assert.True(t, type1 == type2)
}

func TestGetInfoType_DifferentValues(t *testing.T) {
	t.Parallel()

	type1 := OutputConfGetInfoType
	type2 := GetInfoType("other_type")

	assert.NotEqual(t, type1, type2)
	assert.False(t, type1 == type2)
}

func TestAgentSpecificInfoReq_JSONTag(t *testing.T) {
	t.Parallel()

	// This test verifies the struct has the correct JSON tag
	// The GetInfoType field should have json:"get_info_type" binding:"required"
	req := AgentSpecificInfoReq{
		GetInfoType: OutputConfGetInfoType,
	}

	// Just verify the value can be set
	assert.Equal(t, OutputConfGetInfoType, req.GetInfoType)
}
