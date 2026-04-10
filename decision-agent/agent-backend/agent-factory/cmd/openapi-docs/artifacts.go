package main

import (
	"bytes"
	"os"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/internal/openapidoc"
	pkgerrors "github.com/pkg/errors"
)

type docOutputPaths struct {
	PublicJSONPath     string
	PublicYAMLPath     string
	PublicHTMLPath     string
	PublicFaviconPath  string
	RuntimeJSONPath    string
	RuntimeYAMLPath    string
	RuntimeHTMLPath    string
	RuntimeFaviconPath string
	FaviconSourcePath  string
}

type mirroredArtifactPaths struct {
	PublicJSONPath     string
	PublicYAMLPath     string
	PublicHTMLPath     string
	PublicFaviconPath  string
	RuntimeJSONPath    string
	RuntimeYAMLPath    string
	RuntimeHTMLPath    string
	RuntimeFaviconPath string
}

func writeGeneratedArtifacts(paths docOutputPaths, artifacts *openapidoc.BuildArtifacts) error {
	if artifacts == nil {
		return pkgerrors.New("build artifacts is nil")
	}

	if err := writeMirroredArtifact("json", paths.PublicJSONPath, paths.RuntimeJSONPath, artifacts.JSON); err != nil {
		return err
	}

	if err := writeMirroredArtifact("yaml", paths.PublicYAMLPath, paths.RuntimeYAMLPath, artifacts.YAML); err != nil {
		return err
	}

	if err := writeMirroredArtifact("html", paths.PublicHTMLPath, paths.RuntimeHTMLPath, artifacts.HTML); err != nil {
		return err
	}

	if optionalPath(paths.FaviconSourcePath) != "" {
		faviconData, err := os.ReadFile(paths.FaviconSourcePath)
		if err != nil {
			return pkgerrors.Wrap(err, "read favicon source")
		}

		if err := writeMirroredArtifact("favicon", paths.PublicFaviconPath, paths.RuntimeFaviconPath, faviconData); err != nil {
			return err
		}
	}

	return nil
}

func writeMirroredArtifact(label string, publicPath string, runtimePath string, data []byte) error {
	if optionalPath(publicPath) != "" {
		if err := openapidoc.WriteFile(publicPath, data); err != nil {
			return pkgerrors.Wrapf(err, "write public %s", label)
		}
	}

	if optionalPath(runtimePath) != "" {
		if err := openapidoc.WriteFile(runtimePath, data); err != nil {
			return pkgerrors.Wrapf(err, "write runtime %s", label)
		}
	}

	return nil
}

func validateMirroredArtifacts(paths mirroredArtifactPaths) error {
	checks := []struct {
		label       string
		publicPath  string
		runtimePath string
	}{
		{label: "json", publicPath: paths.PublicJSONPath, runtimePath: paths.RuntimeJSONPath},
		{label: "yaml", publicPath: paths.PublicYAMLPath, runtimePath: paths.RuntimeYAMLPath},
		{label: "html", publicPath: paths.PublicHTMLPath, runtimePath: paths.RuntimeHTMLPath},
		{label: "favicon", publicPath: paths.PublicFaviconPath, runtimePath: paths.RuntimeFaviconPath},
	}

	for _, check := range checks {
		if optionalPath(check.publicPath) == "" || optionalPath(check.runtimePath) == "" {
			continue
		}

		publicData, err := os.ReadFile(check.publicPath)
		if err != nil {
			return pkgerrors.Wrapf(err, "read public %s", check.label)
		}

		runtimeData, err := os.ReadFile(check.runtimePath)
		if err != nil {
			return pkgerrors.Wrapf(err, "read runtime %s", check.label)
		}

		if !bytes.Equal(publicData, runtimeData) {
			return pkgerrors.Errorf("%s copies differ between %s and %s", check.label, check.publicPath, check.runtimePath)
		}
	}

	return nil
}
