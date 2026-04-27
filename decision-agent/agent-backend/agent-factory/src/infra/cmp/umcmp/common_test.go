package umcmp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/stretchr/testify/assert"
)

func TestUm_getPrivateURLPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		conf     cconf.UserMgntCfg
		expected string
	}{
		{
			name:     "http protocol",
			conf:     cconf.UserMgntCfg{Protocol: "http", Host: "localhost", Port: 8080},
			expected: "http://localhost:8080/api/user-management",
		},
		{
			name:     "https protocol",
			conf:     cconf.UserMgntCfg{Protocol: "https", Host: "um.example.com", Port: 443},
			expected: "https://um.example.com:443/api/user-management",
		},
		{
			name:     "empty protocol defaults to http",
			conf:     cconf.UserMgntCfg{Protocol: "", Host: "localhost", Port: 9090},
			expected: "http://localhost:9090/api/user-management",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conf := tt.conf
			um := NewUmCmp(&conf, nil)
			result := um.getPrivateURLPrefix()
			assert.Equal(t, tt.expected, result)
		})
	}
}
