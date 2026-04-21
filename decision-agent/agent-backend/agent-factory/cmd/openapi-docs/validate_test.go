package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/internal/openapidoc"
	"github.com/stretchr/testify/require"
)

func TestValidateDefaultsMatchGeneratedSpec(t *testing.T) {
	t.Parallel()

	doc, err := openapidoc.LoadOpenAPIDocFile(filepath.Join("..", "..", defaultOutJSONPath))
	require.NoError(t, err)

	paths, ops := openapidoc.CountPathsAndOperations(doc)
	require.Equal(t, defaultExpectPaths, paths)
	require.Equal(t, defaultExpectOps, ops)
}

func TestValidateDefaultsRemoveObservabilityFromGeneratedSpec(t *testing.T) {
	t.Parallel()

	doc, err := openapidoc.LoadOpenAPIDocFile(filepath.Join("..", "..", defaultOutJSONPath))
	require.NoError(t, err)

	raw, err := doc.MarshalJSON()
	require.NoError(t, err)

	content := string(raw)
	require.NotContains(t, content, "/api/agent-factory/v1/observability/")
	require.NotContains(t, content, `"name":"可观测性"`)
	require.True(t, strings.Contains(content, `"name":"Agent运行（V1）"`) || strings.Contains(content, `"name": "Agent运行（V1）"`))
}

func TestValidateDefaultsRemoveBenchmarkAgentListFromGeneratedSpec(t *testing.T) {
	t.Parallel()

	doc, err := openapidoc.LoadOpenAPIDocFile(filepath.Join("..", "..", defaultOutJSONPath))
	require.NoError(t, err)

	raw, err := doc.MarshalJSON()
	require.NoError(t, err)

	content := string(raw)
	require.NotContains(t, content, "Agent列表（benchmarch）")
	require.NotContains(t, content, "agent_list_for_benchmark")
}
