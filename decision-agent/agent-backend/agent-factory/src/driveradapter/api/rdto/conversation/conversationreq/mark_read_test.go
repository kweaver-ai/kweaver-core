package conversationreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkReadReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &MarkReadReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, 1, len(errMsgMap))
	assert.Equal(t, `"latest_read_index"不能为空`, errMsgMap["LastestReadIdx.required"])
}

func TestMarkReadReq_ReqCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         *MarkReadReq
		expectError bool
	}{
		{
			name: "valid request with index",
			req: &MarkReadReq{
				LastestReadIdx: 5,
			},
			expectError: false,
		},
		{
			name: "valid request with zero index",
			req: &MarkReadReq{
				LastestReadIdx: 0,
			},
			expectError: false,
		},
		{
			name: "valid request with negative index",
			req: &MarkReadReq{
				LastestReadIdx: -1,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.req.ReqCheck()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMarkReadReq_New(t *testing.T) {
	t.Parallel()

	req := &MarkReadReq{}

	assert.NotNil(t, req)
	assert.Equal(t, 0, req.LastestReadIdx)
}

func TestMarkReadReq_WithIndex(t *testing.T) {
	t.Parallel()

	req := &MarkReadReq{
		LastestReadIdx: 10,
	}

	assert.Equal(t, 10, req.LastestReadIdx)
}

func TestMarkReadReq_GetErrMsgMapConsistency(t *testing.T) {
	t.Parallel()

	req1 := &MarkReadReq{}
	req2 := &MarkReadReq{}

	map1 := req1.GetErrMsgMap()
	map2 := req2.GetErrMsgMap()

	// Ensure the method returns consistent results
	assert.Equal(t, map1, map2)
	assert.Equal(t, `"latest_read_index"不能为空`, map1["LastestReadIdx.required"])
}
