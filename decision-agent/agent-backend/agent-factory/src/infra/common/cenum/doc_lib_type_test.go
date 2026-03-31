package cenum

import (
	"testing"
)

func TestDocLibType_EnumCheck(t *testing.T) {
	t.Parallel()

	t.Run("valid personal doc lib", func(t *testing.T) {
		t.Parallel()

		err := DocLibTypeStrPersonal.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for DocLibTypeStrPersonal, got %v", err)
		}
	})

	t.Run("valid department doc lib", func(t *testing.T) {
		t.Parallel()

		err := DocLibTypeStrDepartment.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for DocLibTypeStrDepartment, got %v", err)
		}
	})

	t.Run("valid custom doc lib", func(t *testing.T) {
		t.Parallel()

		err := DocLibTypeStrCustom.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for DocLibTypeStrCustom, got %v", err)
		}
	})

	t.Run("valid knowledge doc lib", func(t *testing.T) {
		t.Parallel()

		err := DocLibTypeStrKnowledge.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for DocLibTypeStrKnowledge, got %v", err)
		}
	})

	t.Run("invalid doc lib type", func(t *testing.T) {
		t.Parallel()

		docType := DocLibType("invalid")
		err := docType.EnumCheck()

		if err == nil {
			t.Error("Expected error for invalid doc lib type")
		}
	})

	t.Run("empty doc lib type", func(t *testing.T) {
		t.Parallel()

		docType := DocLibType("")
		err := docType.EnumCheck()

		if err == nil {
			t.Error("Expected error for empty doc lib type")
		}
	})
}

func TestDocLibType_Constants(t *testing.T) {
	t.Parallel()

	t.Run("DocLibTypeStrPersonal constant", func(t *testing.T) {
		t.Parallel()

		if DocLibTypeStrPersonal != "user_doc_lib" {
			t.Errorf("Expected DocLibTypeStrPersonal to be 'user_doc_lib', got '%s'", DocLibTypeStrPersonal)
		}
	})

	t.Run("DocLibTypeStrDepartment constant", func(t *testing.T) {
		t.Parallel()

		if DocLibTypeStrDepartment != "department_doc_lib" {
			t.Errorf("Expected DocLibTypeStrDepartment to be 'department_doc_lib', got '%s'", DocLibTypeStrDepartment)
		}
	})

	t.Run("DocLibTypeStrCustom constant", func(t *testing.T) {
		t.Parallel()

		if DocLibTypeStrCustom != "custom_doc_lib" {
			t.Errorf("Expected DocLibTypeStrCustom to be 'custom_doc_lib', got '%s'", DocLibTypeStrCustom)
		}
	})

	t.Run("DocLibTypeStrKnowledge constant", func(t *testing.T) {
		t.Parallel()

		if DocLibTypeStrKnowledge != "knowledge_doc_lib" {
			t.Errorf("Expected DocLibTypeStrKnowledge to be 'knowledge_doc_lib', got '%s'", DocLibTypeStrKnowledge)
		}
	})
}
