package cdaenum

import (
	"testing"
)

func TestStatusThreeState_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    StatusThreeState
		wantError bool
	}{
		{"未发布状态", StatusThreeStateUnpublished, false},
		{"已发布状态", StatusThreeStatePublished, false},
		{"发布后有修改状态", StatusThreeStatePublishedEdited, false},
		{"无效状态", StatusThreeState("invalid"), true},
		{"空字符串", StatusThreeState(""), true},
		{"部分匹配", StatusThreeState("published_edited_"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.status.EnumCheck()
			hasError := err != nil

			if hasError != tt.wantError {
				t.Errorf("StatusThreeState.EnumCheck() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.wantError && err != nil {
				errMsg := err.Error()
				if errMsg != "[StatusThreeState]: invalid status" {
					t.Errorf("Error message = %q, want %q", errMsg, "[StatusThreeState]: invalid status")
				}
			}
		})
	}
}

func TestStatusThreeState_IsPublished(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status StatusThreeState
		want   bool
	}{
		{"未发布状态", StatusThreeStateUnpublished, false},
		{"已发布状态", StatusThreeStatePublished, true},
		{"发布后有修改状态", StatusThreeStatePublishedEdited, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.status.IsPublished()
			if got != tt.want {
				t.Errorf("StatusThreeState.IsPublished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusThreeState_StringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{"未发布", "unpublished"},
		{"已发布", "published"},
		{"发布后有修改", "published_edited"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.value != tt.value { //nolint:staticcheck
				t.Errorf("StatusThreeState value = %q, want %q", tt.value, tt.value)
			}
		})
	}
}
