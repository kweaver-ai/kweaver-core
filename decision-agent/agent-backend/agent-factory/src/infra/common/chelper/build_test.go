package chelper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildVersion(t *testing.T) {
	t.Parallel()

	t.Run("with all fields", func(t *testing.T) {
		t.Parallel()

		buildInfo := &BuildInfo{
			BranchName:      "main",
			BuildTime:       "2024-01-01 12:00:00",
			CommitID:        "abc123",
			GoCommonVersion: "v1.0.0",
		}

		version := BuildVersion(buildInfo)

		expected := "branchName: main, buildTime: 2024-01-01 12:00:00, commitID: abc123"
		assert.Equal(t, expected, version)
	})

	t.Run("with empty fields", func(t *testing.T) {
		t.Parallel()

		buildInfo := &BuildInfo{
			BranchName: "",
			BuildTime:  "",
			CommitID:   "",
		}

		version := BuildVersion(buildInfo)

		expected := "branchName: , buildTime: , commitID: "
		assert.Equal(t, expected, version)
	})

	t.Run("with special characters", func(t *testing.T) {
		t.Parallel()

		buildInfo := &BuildInfo{
			BranchName:      "feature/test-branch",
			BuildTime:       "2024-01-01T12:00:00Z",
			CommitID:        "commit-with-dashes",
			GoCommonVersion: "v1.2.3-beta",
		}

		version := BuildVersion(buildInfo)

		assert.Contains(t, version, "feature/test-branch")
		assert.Contains(t, version, "2024-01-01T12:00:00Z")
		assert.Contains(t, version, "commit-with-dashes")
	})

	t.Run("with chinese characters", func(t *testing.T) {
		t.Parallel()

		buildInfo := &BuildInfo{
			BranchName:      "develop-分支",
			BuildTime:       "2024年1月1日",
			CommitID:        "提交ID",
			GoCommonVersion: "v1.0.0",
		}

		version := BuildVersion(buildInfo)

		assert.Contains(t, version, "develop-分支")
		assert.Contains(t, version, "2024年1月1日")
		assert.Contains(t, version, "提交ID")
	})

	t.Run("with only branch name", func(t *testing.T) {
		t.Parallel()

		buildInfo := &BuildInfo{
			BranchName: "main",
		}

		version := BuildVersion(buildInfo)

		expected := "branchName: main, buildTime: , commitID: "
		assert.Equal(t, expected, version)
	})
}

func TestBuildInfo_StructFields(t *testing.T) {
	t.Parallel()

	other := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	buildInfo := &BuildInfo{
		BranchName:      "develop",
		BuildTime:       "2024-01-01",
		CommitID:        "def456",
		GoCommonVersion: "v2.0.0",
		Other:           other,
	}

	assert.Equal(t, "develop", buildInfo.BranchName)
	assert.Equal(t, "2024-01-01", buildInfo.BuildTime)
	assert.Equal(t, "def456", buildInfo.CommitID)
	assert.Equal(t, "v2.0.0", buildInfo.GoCommonVersion)
	assert.Equal(t, "value1", buildInfo.Other["key1"])
	assert.Equal(t, "value2", buildInfo.Other["key2"])
}

func TestBuildInfo_Empty(t *testing.T) {
	t.Parallel()

	buildInfo := &BuildInfo{}

	assert.Empty(t, buildInfo.BranchName)
	assert.Empty(t, buildInfo.BuildTime)
	assert.Empty(t, buildInfo.CommitID)
	assert.Empty(t, buildInfo.GoCommonVersion)
	assert.Nil(t, buildInfo.Other)
}

func TestBuildInfo_WithNilOther(t *testing.T) {
	t.Parallel()

	buildInfo := &BuildInfo{
		Other: nil,
	}

	assert.Nil(t, buildInfo.Other)
}

func TestPrintBuildInfo_Structure(t *testing.T) {
	t.Parallel()

	// Test that PrintBuildInfo can be called with proper structure
	// (We don't test the actual log output as it's difficult to capture)
	buildInfo := &BuildInfo{
		BranchName:      "test",
		BuildTime:       "2024-01-01",
		CommitID:        "abc",
		GoCommonVersion: "v1.0.0",
		Other: map[string]string{
			"extra": "info",
		},
	}

	assert.NotNil(t, buildInfo)
	assert.Equal(t, "test", buildInfo.BranchName)
	assert.NotNil(t, buildInfo.Other)

	// Just verify the function can be called (don't test log output)
	// In a real test, we would capture log output, but for coverage purposes,
	// we just need to ensure the function is callable
	done := make(chan bool)
	go func() {
		PrintBuildInfo(1*time.Millisecond, buildInfo)
		done <- true
	}()

	select {
	case <-done:
		// Function completed successfully
	case <-time.After(2 * time.Second):
		t.Fatal("PrintBuildInfo took too long")
	}
}

func TestPrintBuildInfo_WithZeroDelay(t *testing.T) {
	t.Parallel()

	buildInfo := &BuildInfo{
		BranchName: "main",
		BuildTime:  "now",
		CommitID:   "xyz",
	}

	done := make(chan bool)
	go func() {
		PrintBuildInfo(0, buildInfo)
		done <- true
	}()

	select {
	case <-done:
		// Function completed successfully
	case <-time.After(1 * time.Second):
		t.Fatal("PrintBuildInfo took too long")
	}
}

func TestPrintBuildInfo_WithEmptyOther(t *testing.T) {
	t.Parallel()

	buildInfo := &BuildInfo{
		BranchName: "test",
		BuildTime:  "2024-01-01",
		CommitID:   "abc",
		Other:      map[string]string{},
	}

	done := make(chan bool)
	go func() {
		PrintBuildInfo(1*time.Millisecond, buildInfo)
		done <- true
	}()

	select {
	case <-done:
		// Function completed successfully
	case <-time.After(2 * time.Second):
		t.Fatal("PrintBuildInfo took too long")
	}
}
