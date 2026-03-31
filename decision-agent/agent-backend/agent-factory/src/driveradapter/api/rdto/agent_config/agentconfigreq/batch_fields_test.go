package agentconfigreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchFieldsReqField_String(t *testing.T) {
	t.Parallel()

	field := BatchFieldsReqFieldName

	assert.Equal(t, string(field), field.String())
}

func TestBatchFieldsReqField_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		field   BatchFieldsReqField
		wantErr bool
	}{
		{
			name:    "valid field",
			field:   BatchFieldsReqFieldName,
			wantErr: false,
		},
		{
			name:    "invalid field",
			field:   "invalid_field",
			wantErr: true,
		},
		{
			name:    "empty field",
			field:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.field.ValObjCheck()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid field")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewBatchFieldsReqField(t *testing.T) {
	t.Parallel()

	// This tests that we can create a BatchFieldsReqField from a string
	field := BatchFieldsReqField("name")

	assert.Equal(t, BatchFieldsReqFieldName, field)
}

func TestBatchFieldsReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &BatchFieldsReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, "agent_ids不能为空", errMsgMap["AgentIDs.required"])
	assert.Equal(t, "fields不能为空", errMsgMap["Fields.required"])
}

func TestBatchFieldsReq_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     *BatchFieldsReq
		wantErr bool
	}{
		{
			name: "valid request",
			req: &BatchFieldsReq{
				AgentIDs: []string{"agent1", "agent2"},
				Fields:   []BatchFieldsReqField{BatchFieldsReqFieldName},
			},
			wantErr: false,
		},
		{
			name: "empty agent_ids",
			req: &BatchFieldsReq{
				AgentIDs: []string{},
				Fields:   []BatchFieldsReqField{BatchFieldsReqFieldName},
			},
			wantErr: true,
		},
		{
			name: "nil agent_ids",
			req: &BatchFieldsReq{
				AgentIDs: nil,
				Fields:   []BatchFieldsReqField{BatchFieldsReqFieldName},
			},
			wantErr: true,
		},
		{
			name: "empty fields",
			req: &BatchFieldsReq{
				AgentIDs: []string{"agent1"},
				Fields:   []BatchFieldsReqField{},
			},
			wantErr: true,
		},
		{
			name: "nil fields",
			req: &BatchFieldsReq{
				AgentIDs: []string{"agent1"},
				Fields:   nil,
			},
			wantErr: true,
		},
		{
			name: "invalid field in fields",
			req: &BatchFieldsReq{
				AgentIDs: []string{"agent1"},
				Fields:   []BatchFieldsReqField{"invalid_field"},
			},
			wantErr: true,
		},
		{
			name: "mixed valid and invalid fields",
			req: &BatchFieldsReq{
				AgentIDs: []string{"agent1"},
				Fields:   []BatchFieldsReqField{BatchFieldsReqFieldName, "invalid_field"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.req.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBatchFieldsReq_Validate_WithMultipleAgentIDs(t *testing.T) {
	t.Parallel()

	req := &BatchFieldsReq{
		AgentIDs: []string{"agent1", "agent2", "agent3", "agent4", "agent5"},
		Fields:   []BatchFieldsReqField{BatchFieldsReqFieldName, BatchFieldsReqFieldName},
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestBatchFieldsReq_Validate_AllFields(t *testing.T) {
	t.Parallel()

	req := &BatchFieldsReq{
		AgentIDs: []string{"agent1"},
		Fields:   []BatchFieldsReqField{BatchFieldsReqFieldName, BatchFieldsReqFieldName, BatchFieldsReqFieldName},
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestBatchFieldsReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &BatchFieldsReq{
		AgentIDs: []string{"agent-123", "agent-456"},
		Fields:   []BatchFieldsReqField{BatchFieldsReqFieldName},
	}

	assert.Equal(t, []string{"agent-123", "agent-456"}, req.AgentIDs)
	assert.Len(t, req.Fields, 1)
	assert.Equal(t, BatchFieldsReqFieldName, req.Fields[0])
}

func TestBatchFieldsReq_Empty(t *testing.T) {
	t.Parallel()

	req := &BatchFieldsReq{}

	assert.Empty(t, req.AgentIDs)
	assert.Nil(t, req.Fields)
}
