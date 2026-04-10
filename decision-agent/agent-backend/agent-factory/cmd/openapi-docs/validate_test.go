package main

import (
	"path/filepath"
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
