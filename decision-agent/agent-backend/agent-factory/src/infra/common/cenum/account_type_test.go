package cenum

import (
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func TestAccountType_Constants(t *testing.T) {
	t.Parallel()

	t.Run("AccountTypeUser constant", func(t *testing.T) {
		t.Parallel()

		if AccountTypeUser != "user" {
			t.Errorf("Expected AccountTypeUser to be 'user', got '%s'", AccountTypeUser)
		}
	})

	t.Run("AccountTypeApp constant", func(t *testing.T) {
		t.Parallel()

		if AccountTypeApp != "app" {
			t.Errorf("Expected AccountTypeApp to be 'app', got '%s'", AccountTypeApp)
		}
	})

	t.Run("AccountTypeAnonymous constant", func(t *testing.T) {
		t.Parallel()

		if AccountTypeAnonymous != "anonymous" {
			t.Errorf("Expected AccountTypeAnonymous to be 'anonymous', got '%s'", AccountTypeAnonymous)
		}
	})
}

func TestAccountType_String(t *testing.T) {
	t.Parallel()

	t.Run("user account type", func(t *testing.T) {
		t.Parallel()

		at := AccountTypeUser
		result := at.String()

		if result != "user" {
			t.Errorf("Expected String() to return 'user', got '%s'", result)
		}
	})

	t.Run("app account type", func(t *testing.T) {
		t.Parallel()

		at := AccountTypeApp
		result := at.String()

		if result != "app" {
			t.Errorf("Expected String() to return 'app', got '%s'", result)
		}
	})
}

func TestAccountType_EnumCheck(t *testing.T) {
	t.Parallel()

	t.Run("valid user account type", func(t *testing.T) {
		t.Parallel()

		at := AccountTypeUser

		err := at.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for valid user account type, got %v", err)
		}
	})

	t.Run("valid app account type", func(t *testing.T) {
		t.Parallel()

		at := AccountTypeApp

		err := at.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for valid app account type, got %v", err)
		}
	})

	t.Run("valid anonymous account type", func(t *testing.T) {
		t.Parallel()

		at := AccountTypeAnonymous

		err := at.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for valid anonymous account type, got %v", err)
		}
	})

	t.Run("invalid account type", func(t *testing.T) {
		t.Parallel()

		at := AccountType("invalid")
		err := at.EnumCheck()

		if err == nil {
			t.Error("Expected error for invalid account type, got nil")
		}
	})
}

func TestAccountType_LoadFromMDLVisitorType(t *testing.T) {
	t.Parallel()

	t.Run("load from real name visitor type", func(t *testing.T) {
		t.Parallel()

		var at AccountType

		at.LoadFromMDLVisitorType(rest.VisitorType_RealName)

		if at != AccountTypeUser {
			t.Errorf("Expected account type to be user, got %s", at)
		}
	})

	t.Run("load from user visitor type", func(t *testing.T) {
		t.Parallel()

		var at AccountType

		at.LoadFromMDLVisitorType(rest.VisitorType_User)

		if at != AccountTypeUser {
			t.Errorf("Expected account type to be user, got %s", at)
		}
	})

	t.Run("load from anonymous visitor type", func(t *testing.T) {
		t.Parallel()

		var at AccountType

		at.LoadFromMDLVisitorType(rest.VisitorType_Anonymous)

		if at != AccountTypeAnonymous {
			t.Errorf("Expected account type to be anonymous, got %s", at)
		}
	})

	t.Run("load from app visitor type", func(t *testing.T) {
		t.Parallel()

		var at AccountType

		at.LoadFromMDLVisitorType(rest.VisitorType_App)

		if at != AccountTypeApp {
			t.Errorf("Expected account type to be app, got %s", at)
		}
	})

	t.Run("panic on invalid visitor type", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid visitor type, but did not panic")
			}
		}()

		var at AccountType

		at.LoadFromMDLVisitorType(rest.VisitorType("invalid"))
	})
}

func TestAccountType_ToMDLVisitorType(t *testing.T) {
	t.Parallel()

	t.Run("user account type to visitor type", func(t *testing.T) {
		t.Parallel()

		at := AccountTypeUser
		result := at.ToMDLVisitorType()

		if result != rest.VisitorType_RealName {
			t.Errorf("Expected visitor type to be RealName, got %v", result)
		}
	})

	t.Run("app account type to visitor type", func(t *testing.T) {
		t.Parallel()

		at := AccountTypeApp
		result := at.ToMDLVisitorType()

		if result != rest.VisitorType_App {
			t.Errorf("Expected visitor type to be App, got %v", result)
		}
	})

	t.Run("anonymous account type to visitor type", func(t *testing.T) {
		t.Parallel()

		at := AccountTypeAnonymous
		result := at.ToMDLVisitorType()

		if result != rest.VisitorType_Anonymous {
			t.Errorf("Expected visitor type to be Anonymous, got %v", result)
		}
	})

	t.Run("panic on invalid account type", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid account type, but did not panic")
			}
		}()

		at := AccountType("invalid")
		_ = at.ToMDLVisitorType()
	})
}
