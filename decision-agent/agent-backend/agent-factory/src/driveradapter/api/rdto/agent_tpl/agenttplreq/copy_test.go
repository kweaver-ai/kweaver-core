package agenttplreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &CopyReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	// The map should be empty as per the implementation
	assert.Equal(t, 0, len(errMsgMap))
}

func TestCopyReq_New(t *testing.T) {
	t.Parallel()

	req := &CopyReq{}

	assert.NotNil(t, req)
}

func TestCopyReq_GetErrMsgMapConsistency(t *testing.T) {
	t.Parallel()

	req1 := &CopyReq{}
	req2 := &CopyReq{}

	map1 := req1.GetErrMsgMap()
	map2 := req2.GetErrMsgMap()

	// Ensure the method returns consistent results (both empty maps)
	assert.Equal(t, map1, map2)
	assert.Equal(t, 0, len(map1))
	assert.Equal(t, 0, len(map2))
}

func TestCopyReq_MultipleInstances(t *testing.T) {
	t.Parallel()

	req1 := &CopyReq{}
	req2 := &CopyReq{}

	// Both instances should be independent
	assert.NotNil(t, req1)
	assert.NotNil(t, req2)

	map1 := req1.GetErrMsgMap()
	map2 := req2.GetErrMsgMap()

	// Both should return empty maps
	assert.Equal(t, 0, len(map1))
	assert.Equal(t, 0, len(map2))
}
