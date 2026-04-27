package cutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		yamlStr  string
		setup    func() (string, error)
		wantErr  bool
		wantData interface{}
	}{
		{
			name: "valid YAML file",
			yamlStr: `
a: Easy!
b:
  c: 2
  d: [3, 4]`,
			setup: func() (string, error) {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "test.yaml")
				err := os.WriteFile(filePath, []byte(`
a: Easy!
b:
  c: 2
  d: [3, 4]`), 0o644)
				if err != nil {
					return "", err
				}
				return filePath, nil
			},
			wantErr: false,
			wantData: struct {
				A string `yaml:"a"`
				B struct {
					C int   `yaml:"c"`
					D []int `yaml:"d"`
				}
			}{
				A: "Easy!",
				B: struct {
					C int   `yaml:"c"`
					D []int `yaml:"d"`
				}{
					C: 2,
					D: []int{3, 4},
				},
			},
		},
		{
			name: "non-existent file",
			setup: func() (string, error) {
				return "/non/existent/file.yaml", nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filePath, err := tt.setup()
			if err != nil {
				t.Fatal(err)
			}

			obj := struct {
				A string `yaml:"a"`
				B struct {
					C int   `yaml:"c"`
					D []int `yaml:"d"`
				}
			}{}

			err = YamlParse(filePath, &obj)

			if tt.wantErr {
				assert.Error(t, err, "YamlParse should return error")
			} else {
				assert.NoError(t, err, "YamlParse should not return error")
				assert.Equal(t, tt.wantData.(struct {
					A string `yaml:"a"`
					B struct {
						C int   `yaml:"c"`
						D []int `yaml:"d"`
					}
				}).A, obj.A)
				assert.Equal(t, tt.wantData.(struct {
					A string `yaml:"a"`
					B struct {
						C int   `yaml:"c"`
						D []int `yaml:"d"`
					}
				}).B.C, obj.B.C)
				assert.Equal(t, tt.wantData.(struct {
					A string `yaml:"a"`
					B struct {
						C int   `yaml:"c"`
						D []int `yaml:"d"`
					}
				}).B.D, obj.B.D)
			}
		})
	}
}
