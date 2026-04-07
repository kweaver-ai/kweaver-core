package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyStructUseJSON_Advanced(t *testing.T) {
	t.Parallel()

	type Person struct {
		Name string
		Age  int
	}

	type Employee struct {
		Person
		Company string
	}

	tests := []struct {
		name    string
		dst     interface{}
		src     interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "nested struct copy",
			dst: &Employee{
				Person: Person{
					Name: "OldName",
					Age:  99,
				},
				Company: "OldCompany",
			},
			src: Employee{
				Person: Person{
					Name: "NewName",
					Age:  30,
				},
				Company: "NewCompany",
			},
			want: &Employee{
				Person: Person{
					Name: "NewName",
					Age:  30,
				},
				Company: "NewCompany",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := CopyStructUseJSON(tt.dst, tt.src)
			if tt.wantErr {
				assert.Error(t, err, "CopyStructUseJSON should return error")
			} else {
				assert.NoError(t, err, "CopyStructUseJSON should not return error")
			}
		})
	}
}
