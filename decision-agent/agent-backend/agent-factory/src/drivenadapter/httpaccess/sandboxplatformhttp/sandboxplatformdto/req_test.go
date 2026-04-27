package sandboxplatformdto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSessionReq_StructFields(t *testing.T) {
	t.Parallel()

	id := "session-123"
	version := "1.0.0"
	failOnError := true
	allowConflicts := false

	req := &CreateSessionReq{
		ID:                    &id,
		TemplateID:            "template-456",
		Timeout:               600,
		CPU:                   "2",
		Memory:                "1Gi",
		Disk:                  "10Gi",
		EnvVars:               map[string]string{"KEY1": "value1", "KEY2": "value2"},
		Event:                 map[string]interface{}{"type": "start"},
		Dependencies:          []DependencySpec{{Name: "numpy", Version: &version}},
		InstallTimeout:        300,
		FailOnDependencyError: &failOnError,
		AllowVersionConflicts: &allowConflicts,
	}

	assert.Equal(t, "session-123", *req.ID)
	assert.Equal(t, "template-456", req.TemplateID)
	assert.Equal(t, 600, req.Timeout)
	assert.Equal(t, "2", req.CPU)
	assert.Equal(t, "1Gi", req.Memory)
	assert.Equal(t, "10Gi", req.Disk)
	assert.Len(t, req.EnvVars, 2)
	assert.NotNil(t, req.Event)
	assert.Len(t, req.Dependencies, 1)
	assert.Equal(t, "numpy", req.Dependencies[0].Name)
	assert.Equal(t, "1.0.0", *req.Dependencies[0].Version)
}

func TestCreateSessionReq_Minimal(t *testing.T) {
	t.Parallel()

	req := &CreateSessionReq{
		TemplateID: "template-789",
	}

	assert.Equal(t, "template-789", req.TemplateID)
	assert.Nil(t, req.ID)
	assert.Nil(t, req.EnvVars)
	assert.Nil(t, req.Dependencies)
}

func TestDependencySpec_StructFields(t *testing.T) {
	t.Parallel()

	version := "2.1.0"

	dep := &DependencySpec{
		Name:    "pandas",
		Version: &version,
	}

	assert.Equal(t, "pandas", dep.Name)
	assert.NotNil(t, dep.Version)
	assert.Equal(t, "2.1.0", *dep.Version)
}

func TestDependencySpec_WithoutVersion(t *testing.T) {
	t.Parallel()

	dep := &DependencySpec{
		Name: "requests",
	}

	assert.Equal(t, "requests", dep.Name)
	assert.Nil(t, dep.Version)
}

func TestCreateSessionReq_BooleanPointerFields(t *testing.T) {
	t.Parallel()

	failOnError := true
	allowConflicts := false

	req := &CreateSessionReq{
		TemplateID:            "template-001",
		FailOnDependencyError: &failOnError,
		AllowVersionConflicts: &allowConflicts,
	}

	assert.NotNil(t, req.FailOnDependencyError)
	assert.True(t, *req.FailOnDependencyError)
	assert.NotNil(t, req.AllowVersionConflicts)
	assert.False(t, *req.AllowVersionConflicts)
}

func TestCreateSessionReq_EmptyDependencies(t *testing.T) {
	t.Parallel()

	req := &CreateSessionReq{
		TemplateID:   "template-002",
		Dependencies: []DependencySpec{},
	}

	assert.NotNil(t, req.Dependencies)
	assert.Len(t, req.Dependencies, 0)
}

func TestCreateSessionReq_MultipleDependencies(t *testing.T) {
	t.Parallel()

	v1 := "1.0.0"
	v2 := "2.0.0"

	req := &CreateSessionReq{
		TemplateID: "template-003",
		Dependencies: []DependencySpec{
			{Name: "numpy", Version: &v1},
			{Name: "pandas", Version: &v2},
		},
	}

	assert.Len(t, req.Dependencies, 2)
	assert.Equal(t, "numpy", req.Dependencies[0].Name)
	assert.Equal(t, "pandas", req.Dependencies[1].Name)
}
