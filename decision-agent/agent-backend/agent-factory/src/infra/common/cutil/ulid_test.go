package cutil

import (
	"strings"
	"testing"
)

func TestUlidMake(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{"生成ULID"},
		{"再次生成ULID"},
		{"第三次生成ULID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := UlidMake()
			if got == "" {
				t.Error("UlidMake() 返回空字符串")
			}

			if len(got) != 26 {
				t.Errorf("UlidMake() 返回长度 = %d, want 26", len(got))
			}

			if !strings.ContainsAny(got, "0123456789ABCDEFGHJKMNPQRSTVWXYZ") {
				t.Error("UlidMake() 返回格式不正确")
			}
		})
	}
}

func TestUlidMake_Uniqueness(t *testing.T) {
	t.Parallel()

	ulids := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		ulid := UlidMake()
		if ulids[ulid] {
			t.Errorf("生成的ULID重复: %s", ulid)
			break
		}

		ulids[ulid] = true
	}
}
