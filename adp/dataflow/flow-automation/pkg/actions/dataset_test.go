package actions

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
)

func TestDatasetWriteDocs_Name(t *testing.T) {
	node := &DatasetWriteDocs{}
	assert.Equal(t, common.OpDatasetWriteDocs, node.Name())
}

func TestDatasetWriteDocs_ParameterNew(t *testing.T) {
	node := &DatasetWriteDocs{}
	params := node.ParameterNew()
	_, ok := params.(*DatasetWriteDocs)
	assert.Equal(t, true, ok)
}

func TestNormalizeDatasetDocuments(t *testing.T) {
	// 测试数组输入
	docs := []map[string]any{
		{"id": "doc1", "name": "test1"},
		{"id": "doc2", "name": "test2"},
	}
	result := normalizeDatasetDocuments(docs)
	assert.Equal(t, 2, len(result))

	// 测试单个对象输入
	singleDoc := map[string]any{"id": "doc1", "name": "test"}
	result = normalizeDatasetDocuments(singleDoc)
	assert.Equal(t, 1, len(result))

	// 测试 JSON 字符串输入
	jsonStr := `[{"id": "doc1", "name": "test1"}]`
	result = normalizeDatasetDocuments(jsonStr)
	assert.Equal(t, 1, len(result))

	// 测试空输入
	result = normalizeDatasetDocuments(nil)
	assert.Equal(t, 0, len(result))

	// 测试 []any 输入
	docsAny := []any{
		map[string]any{"id": "doc1", "name": "test1"},
		map[string]any{"id": "doc2", "name": "test2"},
	}
	result = normalizeDatasetDocuments(docsAny)
	assert.Equal(t, 2, len(result))
}
