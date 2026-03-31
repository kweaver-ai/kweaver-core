package agenttplreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdatePublishInfoReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, 0, len(errMsgMap))
}

func TestUpdatePublishInfoReq_New(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{}

	assert.NotNil(t, req)
	assert.Nil(t, req.CategoryIDs)
}

func TestUpdatePublishInfoReq_WithCategoryIDs(t *testing.T) {
	t.Parallel()

	categoryIDs := []string{"cat1", "cat2", "cat3"}
	req := &UpdatePublishInfoReq{
		CategoryIDs: categoryIDs,
	}

	assert.Equal(t, categoryIDs, req.CategoryIDs)
	assert.Equal(t, 3, len(req.CategoryIDs))
}

func TestUpdatePublishInfoReq_WithEmptyCategoryIDs(t *testing.T) {
	t.Parallel()

	req := &UpdatePublishInfoReq{
		CategoryIDs: []string{},
	}

	assert.NotNil(t, req.CategoryIDs)
	assert.Equal(t, 0, len(req.CategoryIDs))
}

func TestUpdatePublishInfoReq_GetErrMsgMapConsistency(t *testing.T) {
	t.Parallel()

	req1 := &UpdatePublishInfoReq{}
	req2 := &UpdatePublishInfoReq{
		CategoryIDs: []string{"cat1"},
	}

	map1 := req1.GetErrMsgMap()
	map2 := req2.GetErrMsgMap()

	// Both should return empty maps
	assert.Equal(t, 0, len(map1))
	assert.Equal(t, 0, len(map2))
}
