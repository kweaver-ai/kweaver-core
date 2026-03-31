package cenum

import (
	"testing"
)

func TestYesNoInt8_EnumCheck(t *testing.T) {
	t.Parallel()

	t.Run("valid yes", func(t *testing.T) {
		t.Parallel()

		err := YesNoInt8Yes.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for Yes, got %v", err)
		}
	})

	t.Run("valid no", func(t *testing.T) {
		t.Parallel()

		err := YesNoInt8No.EnumCheck()
		if err != nil {
			t.Errorf("Expected no error for No, got %v", err)
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		t.Parallel()

		val := YesNoInt8(2)
		err := val.EnumCheck()

		if err == nil {
			t.Error("Expected error for invalid value")
		}
	})

	t.Run("negative value", func(t *testing.T) {
		t.Parallel()

		val := YesNoInt8(-1)
		err := val.EnumCheck()

		if err == nil {
			t.Error("Expected error for negative value")
		}
	})
}

func TestYesNoInt8_IsYes(t *testing.T) {
	t.Parallel()

	t.Run("yes value", func(t *testing.T) {
		t.Parallel()

		if !YesNoInt8Yes.IsYes() {
			t.Error("Expected IsYes to return true for Yes")
		}
	})

	t.Run("no value", func(t *testing.T) {
		t.Parallel()

		if YesNoInt8No.IsYes() {
			t.Error("Expected IsYes to return false for No")
		}
	})
}

func TestYesNoInt8_IsNo(t *testing.T) {
	t.Parallel()

	t.Run("no value", func(t *testing.T) {
		t.Parallel()

		if !YesNoInt8No.IsNo() {
			t.Error("Expected IsNo to return true for No")
		}
	})

	t.Run("yes value", func(t *testing.T) {
		t.Parallel()

		if YesNoInt8Yes.IsNo() {
			t.Error("Expected IsNo to return false for Yes")
		}
	})
}

func TestYesNoInt8_Bool(t *testing.T) {
	t.Parallel()

	t.Run("yes value", func(t *testing.T) {
		t.Parallel()

		if !YesNoInt8Yes.Bool() {
			t.Error("Expected Bool to return true for Yes")
		}
	})

	t.Run("no value", func(t *testing.T) {
		t.Parallel()

		if YesNoInt8No.Bool() {
			t.Error("Expected Bool to return false for No")
		}
	})
}

func TestYesNoInt8_Constants(t *testing.T) {
	t.Parallel()

	t.Run("YesNoInt8Yes constant", func(t *testing.T) {
		t.Parallel()

		if YesNoInt8Yes != 1 {
			t.Errorf("Expected YesNoInt8Yes to be 1, got %d", YesNoInt8Yes)
		}
	})

	t.Run("YesNoInt8No constant", func(t *testing.T) {
		t.Parallel()

		if YesNoInt8No != 0 {
			t.Errorf("Expected YesNoInt8No to be 0, got %d", YesNoInt8No)
		}
	})
}
