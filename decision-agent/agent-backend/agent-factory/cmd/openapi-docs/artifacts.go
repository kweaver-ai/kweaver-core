package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/internal/openapidoc"
	pkgerrors "github.com/pkg/errors"
)

type docOutputPaths struct {
	PublicJSONPath       string
	PublicYAMLPath       string
	PublicHTMLPath       string
	PublicRedocHTMLPath  string
	PublicFaviconPath    string
	PublicUIDirPath      string
	RuntimeJSONPath      string
	RuntimeYAMLPath      string
	RuntimeHTMLPath      string
	RuntimeRedocHTMLPath string
	RuntimeFaviconPath   string
	RuntimeUIDirPath     string
	FaviconSourcePath    string
	UISourceDirPath      string
}

type mirroredArtifactPaths struct {
	PublicJSONPath       string
	PublicYAMLPath       string
	PublicHTMLPath       string
	PublicRedocHTMLPath  string
	PublicFaviconPath    string
	PublicUIDirPath      string
	RuntimeJSONPath      string
	RuntimeYAMLPath      string
	RuntimeHTMLPath      string
	RuntimeRedocHTMLPath string
	RuntimeFaviconPath   string
	RuntimeUIDirPath     string
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

	if err := writeArtifact("public html", paths.PublicHTMLPath, artifacts.PublicHTML); err != nil {
		return err
	}

	if err := writeArtifact("runtime html", paths.RuntimeHTMLPath, artifacts.RuntimeHTML); err != nil {
		return err
	}

	if err := writeArtifact("public redoc html", paths.PublicRedocHTMLPath, artifacts.PublicRedocHTML); err != nil {
		return err
	}

	if err := writeArtifact("runtime redoc html", paths.RuntimeRedocHTMLPath, artifacts.RuntimeRedocHTML); err != nil {
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

	if optionalPath(paths.UISourceDirPath) != "" {
		if err := writeMirroredDirectory("ui", paths.UISourceDirPath, paths.PublicUIDirPath, paths.RuntimeUIDirPath); err != nil {
			return err
		}
	}

	return nil
}

func writeArtifact(label string, path string, data []byte) error {
	if optionalPath(path) == "" {
		return nil
	}

	if err := openapidoc.WriteFile(path, data); err != nil {
		return pkgerrors.Wrapf(err, "write %s", label)
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

func writeMirroredDirectory(label string, sourceDir string, publicDir string, runtimeDir string) error {
	if optionalPath(publicDir) != "" {
		if err := copyDirectory(sourceDir, publicDir); err != nil {
			return pkgerrors.Wrapf(err, "write public %s directory", label)
		}
	}

	if optionalPath(runtimeDir) != "" {
		if err := copyDirectory(sourceDir, runtimeDir); err != nil {
			return pkgerrors.Wrapf(err, "write runtime %s directory", label)
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

	if optionalPath(paths.PublicUIDirPath) != "" && optionalPath(paths.RuntimeUIDirPath) != "" {
		if err := validateMirroredDirectory("ui", paths.PublicUIDirPath, paths.RuntimeUIDirPath); err != nil {
			return err
		}
	}

	return nil
}

func copyDirectory(sourceDir string, targetDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return pkgerrors.Wrap(err, "walk source directory")
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return pkgerrors.Wrap(err, "resolve relative path")
		}
		if relativePath == "." {
			return nil
		}

		targetPath := filepath.Join(targetDir, relativePath)
		if info.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return pkgerrors.Wrap(err, "create mirrored directory")
			}
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return pkgerrors.Wrap(err, "read source file")
		}

		if err := openapidoc.WriteFile(targetPath, data); err != nil {
			return pkgerrors.Wrap(err, "write mirrored file")
		}

		return nil
	})
}

func validateMirroredDirectory(label string, publicDir string, runtimeDir string) error {
	publicFiles, err := readDirectoryFiles(publicDir)
	if err != nil {
		return pkgerrors.Wrapf(err, "read public %s directory", label)
	}

	runtimeFiles, err := readDirectoryFiles(runtimeDir)
	if err != nil {
		return pkgerrors.Wrapf(err, "read runtime %s directory", label)
	}

	if len(publicFiles) != len(runtimeFiles) {
		return pkgerrors.Errorf("%s directory file counts differ between %s and %s", label, publicDir, runtimeDir)
	}

	for relativePath, publicData := range publicFiles {
		runtimeData, ok := runtimeFiles[relativePath]
		if !ok {
			return pkgerrors.Errorf("%s file %s is missing in %s", label, relativePath, runtimeDir)
		}

		if !bytes.Equal(publicData, runtimeData) {
			return pkgerrors.Errorf("%s file %s differs between %s and %s", label, relativePath, publicDir, runtimeDir)
		}
	}

	return nil
}

func readDirectoryFiles(rootDir string) (map[string][]byte, error) {
	files := make(map[string][]byte)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return pkgerrors.Wrap(err, "walk mirrored directory")
		}
		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return pkgerrors.Wrap(err, "resolve mirrored relative path")
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return pkgerrors.Wrap(err, "read mirrored file")
		}
		files[relativePath] = data

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
