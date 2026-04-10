package otherreq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum/builtinagentenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

func TestDolphinTplListReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &DolphinTplListReq{}
	errMap := req.GetErrMsgMap()

	assert.NotEmpty(t, errMap)
	assert.Equal(t, `"config"不能为空`, errMap["Config.required"])
}

func TestDolphinTplListReq_CustomCheck(t *testing.T) {
	t.Parallel()

	t.Run("valid request with config", func(t *testing.T) {
		t.Parallel()

		req := &DolphinTplListReq{
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		err := req.CustomCheck()
		assert.NoError(t, err)
	})

	t.Run("valid request with built-in agent key", func(t *testing.T) {
		t.Parallel()

		req := &DolphinTplListReq{
			Config:          &daconfvalobj.Config{},
			BuiltInAgentKey: builtinagentenum.AgentKeyDocQA,
		}

		err := req.CustomCheck()
		assert.NoError(t, err)
	})

	t.Run("request without config", func(t *testing.T) {
		t.Parallel()

		req := &DolphinTplListReq{
			Config: nil,
		}

		// This would fail at binding validation layer, not in CustomCheck
		err := req.CustomCheck()
		assert.NoError(t, err)
	})
}

func TestDolphinTplListReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &DolphinTplListReq{
		Config: &daconfvalobj.Config{
			Input: &daconfvalobj.Input{},
		},
		BuiltInAgentKey: builtinagentenum.AgentKeyDocQA,
	}

	assert.NotNil(t, req.Config)
	assert.Equal(t, builtinagentenum.AgentKeyDocQA, req.BuiltInAgentKey)
}

func TestDolphinTplListReq_Empty(t *testing.T) {
	t.Parallel()

	req := &DolphinTplListReq{}

	assert.Nil(t, req.Config)
	assert.Empty(t, req.BuiltInAgentKey)
}

func TestDolphinTplListReq_WithAllBuiltInAgentKeys(t *testing.T) {
	t.Parallel()

	agentKeys := []builtinagentenum.AgentKey{
		builtinagentenum.AgentKeyDocQA,
		builtinagentenum.AgentKeyGraphQA,
		builtinagentenum.AgentKeyOnlineSearch,
		builtinagentenum.AgentKeyPlan,
		builtinagentenum.AgentKeySimpleChat,
		builtinagentenum.AgentKeySummary,
		builtinagentenum.AgentKeyDeepSearch,
	}

	for _, key := range agentKeys {
		t.Run(key.String(), func(t *testing.T) {
			t.Parallel()

			req := &DolphinTplListReq{
				Config:          &daconfvalobj.Config{},
				BuiltInAgentKey: key,
			}

			err := req.CustomCheck()
			assert.NoError(t, err)
			assert.Equal(t, key, req.BuiltInAgentKey)
		})
	}
}

func TestNewDolphinTplListReq(t *testing.T) {
	t.Parallel()

	req := &DolphinTplListReq{
		Config: &daconfvalobj.Config{},
	}

	assert.NotNil(t, req)
	assert.NotNil(t, req.Config)
}
