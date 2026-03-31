package cenum

import (
	"testing"
)

func TestOrgObjType_String(t *testing.T) {
	t.Parallel()

	t.Run("dept type", func(t *testing.T) {
		t.Parallel()

		result := OrgObjTypeDep.String()
		if result != "dept" {
			t.Errorf("Expected 'dept', got '%s'", result)
		}
	})

	t.Run("user type", func(t *testing.T) {
		t.Parallel()

		result := OrgObjTypeUser.String()
		if result != "user" {
			t.Errorf("Expected 'user', got '%s'", result)
		}
	})

	t.Run("group type", func(t *testing.T) {
		t.Parallel()

		result := OrgObjTypeGroup.String()
		if result != "user_group" {
			t.Errorf("Expected 'user_group', got '%s'", result)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()

		orgType := OrgObjType("invalid")

		result := orgType.String()
		if result != "" {
			t.Errorf("Expected empty string for invalid type, got '%s'", result)
		}
	})
}

func TestOrgObjType_EnumCheck(t *testing.T) {
	t.Parallel()

	t.Run("valid dept", func(t *testing.T) {
		t.Parallel()

		err := OrgObjTypeDep.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for OrgObjTypeDep, got %v", err)
		}
	})

	t.Run("valid user", func(t *testing.T) {
		t.Parallel()

		err := OrgObjTypeUser.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for OrgObjTypeUser, got %v", err)
		}
	})

	t.Run("valid group", func(t *testing.T) {
		t.Parallel()

		err := OrgObjTypeGroup.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for OrgObjTypeGroup, got %v", err)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()

		orgType := OrgObjType("invalid")
		err := orgType.EnumCheck()

		if err == nil {
			t.Error("Expected error for invalid org obj type")
		}
	})

	t.Run("empty type", func(t *testing.T) {
		t.Parallel()

		orgType := OrgObjType("")
		err := orgType.EnumCheck()

		if err == nil {
			t.Error("Expected error for empty org obj type")
		}
	})
}

func TestOrgObjType_Constants(t *testing.T) {
	t.Parallel()

	t.Run("OrgObjTypeDep constant", func(t *testing.T) {
		t.Parallel()

		if OrgObjTypeDep != "dept" {
			t.Errorf("Expected OrgObjTypeDep to be 'dept', got '%s'", OrgObjTypeDep)
		}
	})

	t.Run("OrgObjTypeUser constant", func(t *testing.T) {
		t.Parallel()

		if OrgObjTypeUser != "user" {
			t.Errorf("Expected OrgObjTypeUser to be 'user', got '%s'", OrgObjTypeUser)
		}
	})

	t.Run("OrgObjTypeGroup constant", func(t *testing.T) {
		t.Parallel()

		if OrgObjTypeGroup != "user_group" {
			t.Errorf("Expected OrgObjTypeGroup to be 'user_group', got '%s'", OrgObjTypeGroup)
		}
	})
}
