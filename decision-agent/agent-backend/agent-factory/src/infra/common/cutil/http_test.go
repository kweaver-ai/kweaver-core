package cutil

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHttpErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		statusCode  int
		wantIsError bool
	}{
		{
			name:        "2xx success",
			statusCode:  http.StatusOK,
			wantIsError: false,
		},
		{
			name:        "201 success",
			statusCode:  http.StatusCreated,
			wantIsError: false,
		},
		{
			name:        "299 success",
			statusCode:  http.StatusMultipleChoices - 1,
			wantIsError: false,
		},
		{
			name:        "300 redirect",
			statusCode:  http.StatusMultipleChoices,
			wantIsError: true,
		},
		{
			name:        "400 client error",
			statusCode:  http.StatusBadRequest,
			wantIsError: true,
		},
		{
			name:        "500 server error",
			statusCode:  http.StatusInternalServerError,
			wantIsError: true,
		},
		{
			name:        "199 error",
			statusCode:  http.StatusOK - 1,
			wantIsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			response := &http.Response{StatusCode: tt.statusCode}
			result := IsHttpErr(response)

			assert.Equal(t, tt.wantIsError, result, "IsHttpErr should return expected result")
		})
	}
}

func TestSetTpTlsInsecureSkipVerify(t *testing.T) {
	t.Parallel()

	tp := &http.Transport{}

	SetTpTlsInsecureSkipVerify(tp)

	assert.NotNil(t, tp.TLSClientConfig, "TLSClientConfig should be set")
	assert.True(t, tp.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify should be true")
}

func TestSetTpTlsInsecureSkipVerify_NilConfig(t *testing.T) {
	t.Parallel()

	tp := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	SetTpTlsInsecureSkipVerify(tp)

	assert.True(t, tp.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify should be true")
}

func TestGetHTTPAccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		addr     string
		port     int
		protocol string
		want     string
	}{
		{
			name:     "with port",
			addr:     "localhost",
			port:     8080,
			protocol: "http",
			want:     "http://localhost:8080",
		},
		{
			name:     "with port 0",
			addr:     "localhost",
			port:     0,
			protocol: "http",
			want:     "http://localhost",
		},
		{
			name:     "with https protocol",
			addr:     "example.com",
			port:     443,
			protocol: "https",
			want:     "https://example.com:443",
		},
		{
			name:     "IPv6 address with port",
			addr:     "::1",
			port:     8080,
			protocol: "http",
			want:     "http://[::1]:8080",
		},
		{
			name:     "IPv6 address with port 0",
			addr:     "::1",
			port:     0,
			protocol: "http",
			want:     "http://[::1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := GetHTTPAccess(tt.addr, tt.port, tt.protocol)
			assert.Equal(t, tt.want, result, "GetHTTPAccess should return expected result")
		})
	}
}

func TestDoubleStackHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "IPv4 address",
			host: "192.168.1.1",
			want: "192.168.1.1",
		},
		{
			name: "IPv6 address",
			host: "::1",
			want: "[::1]",
		},
		{
			name: "IPv6 full address",
			host: "2001:db8::1",
			want: "[2001:db8::1]",
		},
		{
			name: "IPv6 with port (no brackets)",
			host: "::1:8080",
			want: "[::1:8080]",
		},
		{
			name: "hostname",
			host: "localhost",
			want: "localhost",
		},
		{
			name: "with leading/trailing spaces",
			host: "  localhost  ",
			want: "localhost",
		},
		{
			name: "IPv6 with brackets",
			host: "[::1]",
			want: "[::1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := DoubleStackHost(tt.host)
			assert.Equal(t, tt.want, result, "DoubleStackHost should return expected result")
		})
	}
}
