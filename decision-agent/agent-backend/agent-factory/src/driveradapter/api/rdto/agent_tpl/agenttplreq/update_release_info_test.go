package agenttplreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateReleaseInfoReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &UpdateReleaseInfoReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, 0, len(errMsgMap))
}

func TestUpdateReleaseInfoReq_New(t *testing.T) {
	t.Parallel()

	req := &UpdateReleaseInfoReq{}

	assert.NotNil(t, req)
	assert.Empty(t, req.ReleaseNote)
	assert.Empty(t, req.Version)
}

func TestUpdateReleaseInfoReq_WithValues(t *testing.T) {
	t.Parallel()

	req := &UpdateReleaseInfoReq{
		ReleaseNote: "Version 1.0.0 release",
		Version:     "1.0.0",
	}

	assert.Equal(t, "Version 1.0.0 release", req.ReleaseNote)
	assert.Equal(t, "1.0.0", req.Version)
}

func TestUpdateReleaseInfoReq_WithEmptyValues(t *testing.T) {
	t.Parallel()

	req := &UpdateReleaseInfoReq{
		ReleaseNote: "",
		Version:     "",
	}

	assert.Empty(t, req.ReleaseNote)
	assert.Empty(t, req.Version)
}

func TestUpdateReleaseInfoReq_WithOnlyReleaseNote(t *testing.T) {
	t.Parallel()

	req := &UpdateReleaseInfoReq{
		ReleaseNote: "Bug fixes and improvements",
	}

	assert.Equal(t, "Bug fixes and improvements", req.ReleaseNote)
	assert.Empty(t, req.Version)
}

func TestUpdateReleaseInfoReq_WithOnlyVersion(t *testing.T) {
	t.Parallel()

	req := &UpdateReleaseInfoReq{
		Version: "2.0.0",
	}

	assert.Empty(t, req.ReleaseNote)
	assert.Equal(t, "2.0.0", req.Version)
}

func TestUpdateReleaseInfoReq_GetErrMsgMapConsistency(t *testing.T) {
	t.Parallel()

	req1 := &UpdateReleaseInfoReq{}
	req2 := &UpdateReleaseInfoReq{
		ReleaseNote: "test",
		Version:     "1.0.0",
	}

	map1 := req1.GetErrMsgMap()
	map2 := req2.GetErrMsgMap()

	// Both should return empty maps
	assert.Equal(t, 0, len(map1))
	assert.Equal(t, 0, len(map2))
}
