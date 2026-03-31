package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		host   string
		wanted string
	}{
		{
			name:   "带端口的IP",
			host:   "192.168.1.1:8080",
			wanted: "[192.168.1.1:8080]",
		},
		{
			name:   "IPv4地址",
			host:   "192.168.1.1",
			wanted: "192.168.1.1",
		},
		{
			name:   "域名",
			host:   "example.com",
			wanted: "example.com",
		},
		{
			name:   "localhost",
			host:   "localhost:3000",
			wanted: "[localhost:3000]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ParseHost(tt.host)
			assert.Equal(t, tt.wanted, result)
		})
	}
}
