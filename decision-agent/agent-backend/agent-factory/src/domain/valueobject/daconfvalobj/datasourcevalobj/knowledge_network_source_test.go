package datasourcevalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnowledgeNetworkSource_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kn   *KnowledgeNetworkSource
	}{
		{
			name: "valid with no object types",
			kn: &KnowledgeNetworkSource{
				KnowledgeNetworkID: "kn-123",
			},
		},
		{
			name: "valid with nil object types",
			kn: &KnowledgeNetworkSource{
				KnowledgeNetworkID: "kn-456",
				ObjectTypes:        nil,
			},
		},
		{
			name: "valid with empty object types",
			kn: &KnowledgeNetworkSource{
				KnowledgeNetworkID: "kn-789",
				ObjectTypes:        []*ObjectType{},
			},
		},
		{
			name: "valid with object types",
			kn: &KnowledgeNetworkSource{
				KnowledgeNetworkID: "kn-abc",
				ObjectTypes: []*ObjectType{
					{ObjectTypeID: "type-1"},
					{ObjectTypeID: "type-2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.kn.ValObjCheck()
			assert.NoError(t, err)
		})
	}
}

func TestKnowledgeNetworkSource_ValObjCheck_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		kn          *KnowledgeNetworkSource
		expectedErr string
	}{
		{
			name: "empty knowledge network id",
			kn: &KnowledgeNetworkSource{
				KnowledgeNetworkID: "",
			},
			expectedErr: "knowledge_network_id is required",
		},
		{
			name: "invalid object type",
			kn: &KnowledgeNetworkSource{
				KnowledgeNetworkID: "kn-123",
				ObjectTypes: []*ObjectType{
					{ObjectTypeID: ""},
				},
			},
			expectedErr: "object_type is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.kn.ValObjCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestKnowledgeNetworkSource_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	kn := &KnowledgeNetworkSource{}
	errMap := kn.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Len(t, errMap, 1)
	assert.Equal(t, `"knowledge_network_id"不能为空`, errMap["KnowledgeNetworkID.required"])
}

func TestObjectType_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	objType := &ObjectType{
		ObjectTypeID: "type-123",
	}

	err := objType.ValObjCheck()
	assert.NoError(t, err)
}

func TestObjectType_ValObjCheck_Error(t *testing.T) {
	t.Parallel()

	objType := &ObjectType{
		ObjectTypeID: "",
	}

	err := objType.ValObjCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "object_type_id is required")
}

func TestObjectType_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	objType := &ObjectType{}
	errMap := objType.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Len(t, errMap, 1)
	assert.Equal(t, `"object_type_id"不能为空`, errMap["ObjectTypeID.required"])
}
