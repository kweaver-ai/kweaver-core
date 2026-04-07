package agenttplresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := CopyResp{
		ID:   12345,
		Name: "Test Template",
		Key:  "test-template-key",
	}

	assert.Equal(t, int64(12345), resp.ID)
	assert.Equal(t, "Test Template", resp.Name)
	assert.Equal(t, "test-template-key", resp.Key)
}

func TestCopyResp_Empty(t *testing.T) {
	t.Parallel()

	resp := CopyResp{}

	assert.Equal(t, int64(0), resp.ID)
	assert.Empty(t, resp.Name)
	assert.Empty(t, resp.Key)
}

func TestCopyResp_WithID(t *testing.T) {
	t.Parallel()

	ids := []int64{
		0,
		1,
		12345,
		999999,
		-1, // Should be handled if needed
	}

	for _, id := range ids {
		resp := CopyResp{
			ID: id,
		}
		assert.Equal(t, id, resp.ID)
	}
}

func TestCopyResp_WithName(t *testing.T) {
	t.Parallel()

	names := []string{
		"Test Template",
		"测试模板",
		"Template with numbers 123",
		"Template with special chars !@#$%",
		"",
	}

	for _, name := range names {
		resp := CopyResp{
			Name: name,
		}
		assert.Equal(t, name, resp.Name)
	}
}

func TestCopyResp_WithKey(t *testing.T) {
	t.Parallel()

	keys := []string{
		"test-template-key",
		"template-key-123",
		"模板-key-中文",
		"",
	}

	for _, key := range keys {
		resp := CopyResp{
			Key: key,
		}
		assert.Equal(t, key, resp.Key)
	}
}

func TestCopyResp_WithAllFields(t *testing.T) {
	t.Parallel()

	resp := CopyResp{
		ID:   98765,
		Name: "Complete Template Name",
		Key:  "complete-template-key",
	}

	assert.Equal(t, int64(98765), resp.ID)
	assert.Equal(t, "Complete Template Name", resp.Name)
	assert.Equal(t, "complete-template-key", resp.Key)
}

func TestCopyResp_WithLargeID(t *testing.T) {
	t.Parallel()

	largeID := int64(9223372036854775807) // Max int64
	resp := CopyResp{
		ID: largeID,
	}

	assert.Equal(t, largeID, resp.ID)
}

func TestCopyResp_WithNegativeID(t *testing.T) {
	t.Parallel()

	resp := CopyResp{
		ID: -12345,
	}

	assert.Equal(t, int64(-12345), resp.ID)
}

func TestCopyResp_WithChineseName(t *testing.T) {
	t.Parallel()

	resp := CopyResp{
		Name: "智能客服模板",
	}

	assert.Equal(t, "智能客服模板", resp.Name)
}

func TestCopyResp_WithMixedName(t *testing.T) {
	t.Parallel()

	resp := CopyResp{
		Name: "Template模板Name",
	}

	assert.Equal(t, "Template模板Name", resp.Name)
}

func TestCopyResp_WithSpecialCharsInName(t *testing.T) {
	t.Parallel()

	resp := CopyResp{
		Name: "Template @#$%^&*()",
	}

	assert.Equal(t, "Template @#$%^&*()", resp.Name)
}
