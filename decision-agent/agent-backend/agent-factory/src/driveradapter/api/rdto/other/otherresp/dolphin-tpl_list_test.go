package otherresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/dolphintpleo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum/builtinagentenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/other/otherreq"
	"github.com/stretchr/testify/assert"
)

func TestNewDolphinTplListResp(t *testing.T) {
	t.Parallel()

	resp := NewDolphinTplListResp()
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.PreDolphin)
	assert.NotNil(t, resp.PostDolphin)
	assert.Empty(t, resp.PreDolphin)
	assert.Empty(t, resp.PostDolphin)
}

func TestDolphinTplListResp_LoadFromConfig(t *testing.T) {
	t.Parallel()

	t.Run("with valid config", func(t *testing.T) {
		t.Parallel()

		resp := NewDolphinTplListResp()
		req := &otherreq.DolphinTplListReq{
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
			BuiltInAgentKey: builtinagentenum.AgentKeyDocQA,
		}

		err := resp.LoadFromConfig(req)
		assert.NoError(t, err)
	})

	t.Run("with DocQA agent", func(t *testing.T) {
		t.Parallel()

		resp := NewDolphinTplListResp()
		req := &otherreq.DolphinTplListReq{
			Config:          &daconfvalobj.Config{},
			BuiltInAgentKey: builtinagentenum.AgentKeyDocQA,
		}

		err := resp.LoadFromConfig(req)
		assert.NoError(t, err)
	})

	t.Run("with empty config", func(t *testing.T) {
		t.Parallel()

		resp := NewDolphinTplListResp()
		req := &otherreq.DolphinTplListReq{
			Config:          &daconfvalobj.Config{},
			BuiltInAgentKey: builtinagentenum.AgentKeyDocQA,
		}

		err := resp.LoadFromConfig(req)
		assert.NoError(t, err)
	})
}

func TestDolphinTplListResp_Fields(t *testing.T) {
	t.Parallel()

	preEo := &dolphintpleo.DolphinTplEo{}
	postEo := &dolphintpleo.DolphinTplEo{}

	resp := &DolphinTplListResp{
		PreDolphin:  []*dolphintpleo.DolphinTplEo{preEo},
		PostDolphin: []*dolphintpleo.DolphinTplEo{postEo},
	}

	assert.Len(t, resp.PreDolphin, 1)
	assert.Len(t, resp.PostDolphin, 1)
}

func TestDolphinTplListResp_Empty(t *testing.T) {
	t.Parallel()

	resp := &DolphinTplListResp{
		PreDolphin:  []*dolphintpleo.DolphinTplEo{},
		PostDolphin: []*dolphintpleo.DolphinTplEo{},
	}

	assert.Empty(t, resp.PreDolphin)
	assert.Empty(t, resp.PostDolphin)
}

func TestDolphinTplListResp_WithAllBuiltInAgentKeys(t *testing.T) {
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

			resp := NewDolphinTplListResp()
			req := &otherreq.DolphinTplListReq{
				Config:          &daconfvalobj.Config{},
				BuiltInAgentKey: key,
			}

			err := resp.LoadFromConfig(req)
			assert.NoError(t, err)
		})
	}
}

func TestDolphinTplListResp_StructInitialization(t *testing.T) {
	t.Parallel()

	resp := &DolphinTplListResp{}

	// Verify struct is initialized
	assert.NotNil(t, resp)
	// PreDolphin and PostDolphin are nil when not using NewDolphinTplListResp
	assert.Nil(t, resp.PreDolphin)
	assert.Nil(t, resp.PostDolphin)
}

func TestDolphinTplListResp_LoadFromConfig_WithMemoryRetrieve(t *testing.T) {
	t.Parallel()

	resp := NewDolphinTplListResp()
	isEnabled := true
	req := &otherreq.DolphinTplListReq{
		Config: &daconfvalobj.Config{
			MemoryCfg: &daconfvalobj.MemoryCfg{
				IsEnabled: isEnabled,
			},
			Input:  &daconfvalobj.Input{},
			Output: &daconfvalobj.Output{},
		},
		BuiltInAgentKey: builtinagentenum.AgentKeyDocQA,
	}

	err := resp.LoadFromConfig(req)
	assert.NoError(t, err)
	assert.Greater(t, len(resp.PreDolphin), 0, "PreDolphin should contain MemoryRetrieve template")
}

func TestDolphinTplListResp_LoadFromConfig_WithDocRetrieve(t *testing.T) {
	t.Parallel()

	resp := NewDolphinTplListResp()
	req := &otherreq.DolphinTplListReq{
		Config: &daconfvalobj.Config{
			DataSource: &datasourcevalobj.RetrieverDataSource{
				Doc: []*datasourcevalobj.DocSource{
					{
						DsID:   "test-ds-id",
						Fields: []*datasourcevalobj.DocSourceField{{Name: "test-field"}},
					},
				},
			},
			Input:  &daconfvalobj.Input{},
			Output: &daconfvalobj.Output{},
		},
		BuiltInAgentKey: builtinagentenum.AgentKeySimpleChat, // Not DocQA to avoid disabling doc retrieve
	}

	err := resp.LoadFromConfig(req)
	assert.NoError(t, err)
	assert.Greater(t, len(resp.PreDolphin), 0, "PreDolphin should contain DocRetrieve template")
}

func TestDolphinTplListResp_LoadFromConfig_WithRelatedQuestions(t *testing.T) {
	t.Parallel()

	resp := NewDolphinTplListResp()
	isEnabled := true
	req := &otherreq.DolphinTplListReq{
		Config: &daconfvalobj.Config{
			RelatedQuestion: &daconfvalobj.RelatedQuestion{
				IsEnabled: isEnabled,
			},
			Input:  &daconfvalobj.Input{},
			Output: &daconfvalobj.Output{},
		},
		BuiltInAgentKey: builtinagentenum.AgentKeyDocQA,
	}

	err := resp.LoadFromConfig(req)
	assert.NoError(t, err)
	assert.Greater(t, len(resp.PostDolphin), 0, "PostDolphin should contain RelatedQuestions template")
}
