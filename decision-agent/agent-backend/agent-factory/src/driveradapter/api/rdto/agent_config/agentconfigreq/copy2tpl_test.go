package agentconfigreq

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopy2TplReq_StructFields(t *testing.T) {
	t.Parallel()

	req := Copy2TplReq{
		Name: "Test Template",
	}

	assert.Equal(t, "Test Template", req.Name)
}

func TestCopy2TplReq_Empty(t *testing.T) {
	t.Parallel()

	req := Copy2TplReq{}

	assert.Empty(t, req.Name)
}

func TestCopy2TplReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := Copy2TplReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Empty(t, errMsgMap)
}

func TestCopy2TplReq_ReqCheck_Valid(t *testing.T) {
	t.Parallel()

	req := Copy2TplReq{
		Name: "Valid Template Name",
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestCopy2TplReq_ReqCheck_Empty(t *testing.T) {
	t.Parallel()

	req := Copy2TplReq{}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestCopy2TplReq_ReqCheck_NameTooLong(t *testing.T) {
	t.Parallel()

	req := Copy2TplReq{
		Name: strings.Repeat("a", 51), // Max is 50
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "模板名称长度不能超过")
}

func TestCopy2TplReq_ReqCheck_NameMaxLength(t *testing.T) {
	t.Parallel()

	req := Copy2TplReq{
		Name: strings.Repeat("a", 50), // Exactly at max
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestCopy2TplReq_WithDifferentNames(t *testing.T) {
	t.Parallel()

	names := []string{
		"Template 1",
		"中文模板",
		"Template with numbers 123",
		"Template with special chars !@#$%",
	}

	for _, name := range names {
		req := Copy2TplReq{
			Name: name,
		}
		err := req.ReqCheck()
		assert.NoError(t, err)
		assert.Equal(t, name, req.Name)
	}
}

func TestCopy2TplReq_WithUnicodeName(t *testing.T) {
	t.Parallel()

	// Test that unicode characters are counted correctly
	// Each Chinese character is 3 bytes but 1 rune
	req := Copy2TplReq{
		Name: strings.Repeat("中", 50), // 50 Chinese characters
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestCopy2TplReq_WithMixedUnicodeName(t *testing.T) {
	t.Parallel()

	req := Copy2TplReq{
		Name: "Template模板123!@#",
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestCopy2TplReq_WithNameAtBoundary(t *testing.T) {
	t.Parallel()

	// Test exactly at the boundary
	req := Copy2TplReq{
		Name: strings.Repeat("a", 49) + "b", // 50 characters
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestCopy2TplReq_WithNameOverBoundary(t *testing.T) {
	t.Parallel()

	// Test just over the boundary
	req := Copy2TplReq{
		Name: strings.Repeat("a", 50) + "b", // 51 characters
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "模板名称长度不能超过50")
}
