package cenum

import (
	"testing"
)

func TestUserType_String(t *testing.T) {
	t.Parallel()

	t.Run("string conversion", func(t *testing.T) {
		t.Parallel()

		userType := UserType("test-type")
		result := userType.String()

		if result != "test-type" {
			t.Errorf("Expected 'test-type', got '%s'", result)
		}
	})

	t.Run("empty type", func(t *testing.T) {
		t.Parallel()

		userType := UserType("")
		result := userType.String()

		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})
}

func TestUserType_EnumCheck(t *testing.T) {
	t.Parallel()

	t.Run("valid authenticated user", func(t *testing.T) {
		t.Parallel()

		err := AuthUser.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for AuthUser, got %v", err)
		}
	})

	t.Run("valid anonymous user", func(t *testing.T) {
		t.Parallel()

		err := AnonymousUser.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for AnonymousUser, got %v", err)
		}
	})

	t.Run("valid app", func(t *testing.T) {
		t.Parallel()

		err := App.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for App, got %v", err)
		}
	})

	t.Run("valid internal service", func(t *testing.T) {
		t.Parallel()

		err := InternalService.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for InternalService, got %v", err)
		}
	})

	t.Run("invalid user type", func(t *testing.T) {
		t.Parallel()

		userType := UserType("invalid")
		err := userType.EnumCheck()

		if err == nil {
			t.Error("Expected error for invalid user type")
		}
	})

	t.Run("empty user type", func(t *testing.T) {
		t.Parallel()

		userType := UserType("")
		err := userType.EnumCheck()

		if err == nil {
			t.Error("Expected error for empty user type")
		}
	})
}

func TestUserType_Constants(t *testing.T) {
	t.Parallel()

	t.Run("AuthUser constant", func(t *testing.T) {
		t.Parallel()

		if AuthUser != "authenticated_user" {
			t.Errorf("Expected AuthUser to be 'authenticated_user', got '%s'", AuthUser)
		}
	})

	t.Run("AnonymousUser constant", func(t *testing.T) {
		t.Parallel()

		if AnonymousUser != "anonymous_user" {
			t.Errorf("Expected AnonymousUser to be 'anonymous_user', got '%s'", AnonymousUser)
		}
	})

	t.Run("App constant", func(t *testing.T) {
		t.Parallel()

		if App != "app" {
			t.Errorf("Expected App to be 'app', got '%s'", App)
		}
	})

	t.Run("InternalService constant", func(t *testing.T) {
		t.Parallel()

		if InternalService != "internal_service" {
			t.Errorf("Expected InternalService to be 'internal_service', got '%s'", InternalService)
		}
	})
}
