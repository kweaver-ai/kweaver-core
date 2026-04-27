package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlParseFromStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yamlStr string
		wantErr bool
		wantA   string
		wantC   int
	}{
		{
			name: "valid YAML string",
			yamlStr: `
a: Easy!
b:
  c: 2
  d: [3, 4]`,
			wantErr: false,
			wantA:   "Easy!",
			wantC:   2,
		},
		{
			name:    "empty string",
			yamlStr: "",
			wantErr: false,
		},
		{
			name:    "simple string",
			yamlStr: `a: test`,
			wantErr: false,
			wantA:   "test",
		},
		{
			name:    "invalid YAML",
			yamlStr: `a: :`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			obj := struct {
				A string `yaml:"a"`
				B struct {
					C int   `yaml:"c"`
					D []int `yaml:"d"`
				}
			}{}

			err := YamlParseFromStr(tt.yamlStr, &obj)

			if tt.wantErr {
				assert.Error(t, err, "YamlParseFromStr should return error")
			} else {
				assert.NoError(t, err, "YamlParseFromStr should not return error")

				if tt.wantA != "" {
					assert.Equal(t, tt.wantA, obj.A, "A field should match")
				}

				if tt.wantC != 0 {
					assert.Equal(t, tt.wantC, obj.B.C, "B.C field should match")
				}
			}
		})
	}
}
