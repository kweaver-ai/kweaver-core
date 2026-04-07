package cpmsresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckRunResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := &CheckRunResp{
		IsAllowed: true,
	}

	assert.True(t, resp.IsAllowed)
}

func TestCheckRunResp_NotAllowed(t *testing.T) {
	t.Parallel()

	resp := &CheckRunResp{
		IsAllowed: false,
	}

	assert.False(t, resp.IsAllowed)
}

func TestCheckRunResp_DefaultValue(t *testing.T) {
	t.Parallel()

	resp := &CheckRunResp{}

	// Default value of bool is false
	assert.False(t, resp.IsAllowed)
}
