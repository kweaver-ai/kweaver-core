package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/internal/openapidoc"
	"github.com/stretchr/testify/require"
)

func TestDefaultDocPathsUseNewLayout(t *testing.T) {
	t.Parallel()

	require.Equal(t, "cmd/openapi-docs/generated/swagger/swagger.json", defaultSwaggerPath)
	require.Equal(t, "cmd/openapi-docs/assets/overlay.yaml", defaultOverlayPath)
	require.Equal(t, "cmd/openapi-docs/assets/baseline/agent-factory.json", defaultBaselinePath)
	require.Equal(t, "docs/api/favicon.png", defaultPublicFaviconPath)
	require.Equal(t, "src/infra/server/apidocs/assets/agent-factory.json", defaultRuntimeJSONPath)
	require.Equal(t, "src/infra/server/apidocs/assets/favicon.png", defaultRuntimeFaviconPath)
	require.Equal(t, defaultRuntimeFaviconPath, defaultFaviconSourcePath)
}

func TestWriteGeneratedArtifactsWritesPublicAndRuntimeCopies(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	faviconSourcePath := filepath.Join(rootDir, "source", "favicon.png")
	require.NoError(t, os.MkdirAll(filepath.Dir(faviconSourcePath), 0o755))
	require.NoError(t, os.WriteFile(faviconSourcePath, []byte("favicon"), 0o644))

	outputs := docOutputPaths{
		PublicJSONPath:    filepath.Join(rootDir, "docs", "api", "agent-factory.json"),
		PublicYAMLPath:    filepath.Join(rootDir, "docs", "api", "agent-factory.yaml"),
		PublicHTMLPath:    filepath.Join(rootDir, "docs", "api", "agent-factory.html"),
		PublicFaviconPath: filepath.Join(rootDir, "docs", "api", "favicon.png"),
		RuntimeJSONPath:   filepath.Join(rootDir, "runtime", "assets", "agent-factory.json"),
		RuntimeYAMLPath:   filepath.Join(rootDir, "runtime", "assets", "agent-factory.yaml"),
		RuntimeHTMLPath:   filepath.Join(rootDir, "runtime", "assets", "agent-factory.html"),
		RuntimeFaviconPath: filepath.Join(
			rootDir, "runtime", "assets", "favicon.png",
		),
		FaviconSourcePath: faviconSourcePath,
	}

	artifacts := &openapidoc.BuildArtifacts{
		JSON: []byte("{\"openapi\":\"3.0.2\"}\n"),
		YAML: []byte("openapi: 3.0.2\n"),
		HTML: []byte("<html>scalar</html>\n"),
	}

	require.NoError(t, writeGeneratedArtifacts(outputs, artifacts))

	requireFileContent(t, outputs.PublicJSONPath, artifacts.JSON)
	requireFileContent(t, outputs.PublicYAMLPath, artifacts.YAML)
	requireFileContent(t, outputs.PublicHTMLPath, artifacts.HTML)
	requireFileContent(t, outputs.PublicFaviconPath, []byte("favicon"))
	requireFileContent(t, outputs.RuntimeJSONPath, artifacts.JSON)
	requireFileContent(t, outputs.RuntimeYAMLPath, artifacts.YAML)
	requireFileContent(t, outputs.RuntimeHTMLPath, artifacts.HTML)
	requireFileContent(t, outputs.RuntimeFaviconPath, []byte("favicon"))
}

func TestValidateMirroredArtifactsRejectsDifferentCopies(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	paths := mirroredArtifactPaths{
		PublicJSONPath:    filepath.Join(rootDir, "docs", "api", "agent-factory.json"),
		PublicYAMLPath:    filepath.Join(rootDir, "docs", "api", "agent-factory.yaml"),
		PublicHTMLPath:    filepath.Join(rootDir, "docs", "api", "agent-factory.html"),
		PublicFaviconPath: filepath.Join(rootDir, "docs", "api", "favicon.png"),
		RuntimeJSONPath:   filepath.Join(rootDir, "runtime", "assets", "agent-factory.json"),
		RuntimeYAMLPath:   filepath.Join(rootDir, "runtime", "assets", "agent-factory.yaml"),
		RuntimeHTMLPath:   filepath.Join(rootDir, "runtime", "assets", "agent-factory.html"),
		RuntimeFaviconPath: filepath.Join(
			rootDir, "runtime", "assets", "favicon.png",
		),
	}

	writeMirroredFixture(t, paths, "same-json", "same-yaml", "same-html", "same-favicon")
	require.NoError(t, os.WriteFile(paths.RuntimeHTMLPath, []byte("different-html"), 0o644))

	err := validateMirroredArtifacts(paths)
	require.Error(t, err)
	require.ErrorContains(t, err, "html")
}

func TestValidateMirroredArtifactsAcceptsMatchingCopies(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	paths := mirroredArtifactPaths{
		PublicJSONPath:    filepath.Join(rootDir, "docs", "api", "agent-factory.json"),
		PublicYAMLPath:    filepath.Join(rootDir, "docs", "api", "agent-factory.yaml"),
		PublicHTMLPath:    filepath.Join(rootDir, "docs", "api", "agent-factory.html"),
		PublicFaviconPath: filepath.Join(rootDir, "docs", "api", "favicon.png"),
		RuntimeJSONPath:   filepath.Join(rootDir, "runtime", "assets", "agent-factory.json"),
		RuntimeYAMLPath:   filepath.Join(rootDir, "runtime", "assets", "agent-factory.yaml"),
		RuntimeHTMLPath:   filepath.Join(rootDir, "runtime", "assets", "agent-factory.html"),
		RuntimeFaviconPath: filepath.Join(
			rootDir, "runtime", "assets", "favicon.png",
		),
	}

	writeMirroredFixture(t, paths, "same-json", "same-yaml", "same-html", "same-favicon")

	require.NoError(t, validateMirroredArtifacts(paths))
}

func writeMirroredFixture(t *testing.T, paths mirroredArtifactPaths, jsonContent string, yamlContent string, htmlContent string, faviconContent string) {
	t.Helper()

	require.NoError(t, openapidoc.WriteFile(paths.PublicJSONPath, []byte(jsonContent)))
	require.NoError(t, openapidoc.WriteFile(paths.RuntimeJSONPath, []byte(jsonContent)))
	require.NoError(t, openapidoc.WriteFile(paths.PublicYAMLPath, []byte(yamlContent)))
	require.NoError(t, openapidoc.WriteFile(paths.RuntimeYAMLPath, []byte(yamlContent)))
	require.NoError(t, openapidoc.WriteFile(paths.PublicHTMLPath, []byte(htmlContent)))
	require.NoError(t, openapidoc.WriteFile(paths.RuntimeHTMLPath, []byte(htmlContent)))
	require.NoError(t, openapidoc.WriteFile(paths.PublicFaviconPath, []byte(faviconContent)))
	require.NoError(t, openapidoc.WriteFile(paths.RuntimeFaviconPath, []byte(faviconContent)))
}

func requireFileContent(t *testing.T, path string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
