package appruntime

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerName_Constant(t *testing.T) {
	assert.Equal(t, "agent-factory", ServerName)
	assert.NotEmpty(t, ServerName)
}

func TestServerVersion_Constant(t *testing.T) {
	assert.Equal(t, "1.0.0", ServerVersion)
	assert.NotEmpty(t, ServerVersion)
}

func TestLanguageGo_Constant(t *testing.T) {
	assert.Equal(t, "go", LanguageGo)
	assert.NotEmpty(t, LanguageGo)
}

func TestGoVersion_Constant(t *testing.T) {
	assert.Equal(t, runtime.Version(), GoVersion)
	assert.NotEmpty(t, GoVersion)
	assert.Contains(t, GoVersion, "go")
}

func TestGoArch_Constant(t *testing.T) {
	assert.Equal(t, runtime.GOARCH, GoArch)
	assert.NotEmpty(t, GoArch)
	// Common architectures are amd64, arm64, etc.
	validArchs := []string{"amd64", "arm64", "386", "arm", "ppc64", "ppc64le", "mips", "mipsle", "mips64", "mips64le", "riscv64", "s390x"}
	assert.Contains(t, validArchs, GoArch)
}

func TestTraceInstrumentationName_Constant(t *testing.T) {
	assert.Equal(t, "Opentelemetry@1.39.0/exporter", TraceInstrumentationName)
	assert.NotEmpty(t, TraceInstrumentationName)
	assert.Contains(t, TraceInstrumentationName, "Opentelemetry")
	assert.Contains(t, TraceInstrumentationName, "1.39.0")
}

func TestServerName_Format(t *testing.T) {
	// Check if server name follows kebab-case format
	assert.Contains(t, ServerName, "-")
	assert.NotContains(t, ServerName, "_")
}

func TestServerVersion_Format(t *testing.T) {
	// Check if version follows semantic versioning format (major.minor.patch)
	assert.Contains(t, ServerVersion, ".")
	assert.Equal(t, "1.0.0", ServerVersion)
}

func TestLanguageGo_Value(t *testing.T) {
	assert.Equal(t, "go", LanguageGo)
}

func TestGoVersion_MatchesRuntime(t *testing.T) {
	assert.Equal(t, runtime.Version(), GoVersion)
}

func TestGoArch_MatchesRuntime(t *testing.T) {
	assert.Equal(t, runtime.GOARCH, GoArch)
}

func TestAllConstants_AreNotEmpty(t *testing.T) {
	assert.NotEmpty(t, ServerName)
	assert.NotEmpty(t, ServerVersion)
	assert.NotEmpty(t, LanguageGo)
	assert.NotEmpty(t, GoVersion)
	assert.NotEmpty(t, GoArch)
	assert.NotEmpty(t, TraceInstrumentationName)
}

func TestTraceInstrumentationName_Structure(t *testing.T) {
	// Check format: Name@version/exporter
	assert.Contains(t, TraceInstrumentationName, "@")
	assert.Contains(t, TraceInstrumentationName, "/")
}

func TestServerName_IsKebabCase(t *testing.T) {
	// Should be all lowercase with hyphens
	for _, c := range ServerName {
		assert.True(t, c == '-' || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	}
}

func TestServerVersion_Parts(t *testing.T) {
	// Version should be in format major.minor.patch
	version := ServerVersion
	assert.Contains(t, version, ".")
	assert.NotContains(t, version, "-") // No pre-release identifier
}

func TestGoArch_IsValid(t *testing.T) {
	// Common valid architectures
	validArchs := map[string]bool{
		"386":      true,
		"amd64":    true,
		"arm":      true,
		"arm64":    true,
		"ppc64":    true,
		"ppc64le":  true,
		"mips":     true,
		"mipsle":   true,
		"mips64":   true,
		"mips64le": true,
		"riscv64":  true,
		"s390x":    true,
		"wasm":     true,
	}

	assert.True(t, validArchs[GoArch], "GoArch should be a valid architecture")
}

func TestAllConstants_HaveExpectedValues(t *testing.T) {
	assert.Equal(t, "agent-factory", ServerName)
	assert.Equal(t, "1.0.0", ServerVersion)
	assert.Equal(t, "go", LanguageGo)
	assert.Equal(t, "Opentelemetry@1.39.0/exporter", TraceInstrumentationName)
}

func TestServerVersion_IsSemVer(t *testing.T) {
	// Semantic versioning format: X.Y.Z
	version := ServerVersion
	dotCount := 0

	for _, c := range version {
		if c == '.' {
			dotCount++
		}
	}

	assert.Equal(t, 2, dotCount, "Version should have exactly 2 dots (X.Y.Z format)")
}

func TestTraceInstrumentationName_Components(t *testing.T) {
	// Should have format: Tool@version/exporter
	assert.Contains(t, TraceInstrumentationName, "Opentelemetry")
	assert.Contains(t, TraceInstrumentationName, "@")
	assert.Contains(t, TraceInstrumentationName, "/")
	assert.Contains(t, TraceInstrumentationName, "exporter")
}
