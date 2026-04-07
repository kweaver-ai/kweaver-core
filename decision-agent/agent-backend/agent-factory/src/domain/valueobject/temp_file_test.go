package valueobject

import (
	"testing"
)

func TestTempFile(t *testing.T) {
	t.Parallel()

	tempFile := &TempFile{
		ID:      "file-123",
		Type:    "pdf",
		Name:    "document.pdf",
		Details: map[string]interface{}{"pages": 10},
	}

	if tempFile.ID != "file-123" {
		t.Errorf("ID = %q, want %q", tempFile.ID, "file-123")
	}

	if tempFile.Type != "pdf" {
		t.Errorf("Type = %q, want %q", tempFile.Type, "pdf")
	}

	if tempFile.Name != "document.pdf" {
		t.Errorf("Name = %q, want %q", tempFile.Name, "document.pdf")
	}

	if tempFile.Details == nil {
		t.Error("Details should not be nil")
	}
}

func TestTempFile_Empty(t *testing.T) {
	t.Parallel()

	tempFile := &TempFile{}

	if tempFile.ID != "" {
		t.Errorf("ID should be empty")
	}

	if tempFile.Type != "" {
		t.Errorf("Type should be empty")
	}

	if tempFile.Name != "" {
		t.Errorf("Name should be empty")
	}
}

func TestTempFile_NilDetails(t *testing.T) {
	t.Parallel()

	tempFile := &TempFile{
		ID:   "file-123",
		Type: "pdf",
		Name: "document.pdf",
	}

	if tempFile.Details != nil {
		t.Errorf("Details should be nil")
	}
}
