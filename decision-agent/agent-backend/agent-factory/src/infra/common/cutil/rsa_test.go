package cutil

import (
	"strings"
	"testing"
)

func TestRsa2048Decrypt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		encrypted  string
		wantErr    bool
		wantPrefix string
	}{
		{
			name:      "空字符串",
			encrypted: "",
			wantErr:   true,
		},
		{
			name:      "无效的 Base64",
			encrypted: "invalid-base64",
			wantErr:   true,
		},
		{
			name: "正确的加密数据",
			encrypted: "zhMytQF/dmSfreO1Qgkdr8wBEtzi/2QcFwroQV8y+AnFjqhS6aAkVExtgk1VpjwtBk6DlmtSedTFngRLbc61aDQKKJhXTtGYucnRGgOOqD2uu" +
				"I+MaxxAk5t7Vys29XzEyHXB5OAvETjfxkNV/5jAxmQ8k29NDraxpz/yhZ/SsnviskBaGE+l/n+7EvhL2VIVhf9Yp3FB96tOxrjfApf+7a0iIN5NgM+5YjazKnN8nHAJ5Em" +
				"SINBbb+nK+7ciC+IkEBLXRms5Hv5KWpUdPP23iw55Nl3ffjXvqtUyCfVBOqItWDAd32DA7U8Qg8Ver7Tn3wScuGokNijmis0dlEbRVw==",
			wantErr:    false,
			wantPrefix: "1111",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			decrypted, err := Rsa2048Decrypt(tt.encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rsa2048Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.wantPrefix != "" && !strings.HasPrefix(decrypted, tt.wantPrefix) {
				t.Errorf("Rsa2048Decrypt() = %v, want prefix %v", decrypted, tt.wantPrefix)
			}
		})
	}
}
