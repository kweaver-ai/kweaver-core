package httprequesthelper

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRequestRecord_BasicFields(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	record := NewRequestRecord(req, "", false)

	assert.Equal(t, "GET", record.Method)
	assert.Contains(t, record.URL, "example.com")
	assert.Contains(t, record.URL, "/test")
	assert.Empty(t, record.Body)
	assert.False(t, record.Timestamp.IsZero())
}

func TestNewRequestRecord_WithBody(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("POST", "http://example.com/api", nil)
	if err != nil {
		t.Fatal(err)
	}

	body := `{"key":"value"}`
	record := NewRequestRecord(req, body, false)

	assert.Equal(t, body, record.Body)
}

func TestNewRequestRecord_WithHeaders(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Accept", "application/json")

	record := NewRequestRecord(req, "", true)

	assert.NotNil(t, record.Headers)
	assert.Equal(t, "application/json", record.Headers["Content-Type"])
	assert.Equal(t, "Bearer token123", record.Headers["Authorization"])
	assert.Equal(t, "application/json", record.Headers["Accept"])
}

func TestNewRequestRecord_WithoutHeaders(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	record := NewRequestRecord(req, "", false)

	assert.Nil(t, record.Headers)
}

func TestNewRequestRecord_HTTPS_URL(t *testing.T) {
	t.Parallel()

	req := &http.Request{
		Method: "GET",
		Host:   "example.com",
		URL: &url.URL{
			Scheme: "https",
			Path:   "/test",
		},
		TLS: &tls.ConnectionState{},
	}

	record := NewRequestRecord(req, "", false)

	assert.Contains(t, record.URL, "https://example.com/test")
}

func TestNewRequestRecord_HTTP_URL(t *testing.T) {
	t.Parallel()

	req := &http.Request{
		Method: "GET",
		Host:   "example.com",
		URL: &url.URL{
			Path: "/test",
		},
	}

	record := NewRequestRecord(req, "", false)

	assert.Contains(t, record.URL, "http://example.com/test")
}

func TestNewRequestRecord_QueryParams(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("GET", "http://example.com/api?key=value&foo=bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	record := NewRequestRecord(req, "", false)

	assert.Contains(t, record.URL, "key=value")
	assert.Contains(t, record.URL, "foo=bar")
}

func TestNewRequestRecord_MultiValueHeaders(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Set-Cookie", "cookie1=value1")
	req.Header.Add("Set-Cookie", "cookie2=value2")

	record := NewRequestRecord(req, "", true)

	// Should only take the first value
	assert.Equal(t, "cookie1=value1", record.Headers["Set-Cookie"])
}

func TestNewResponseRecord_BasicFields(t *testing.T) {
	t.Parallel()

	headers := http.Header{
		"Content-Type":  []string{"application/json"},
		"Cache-Control": []string{"no-cache"},
	}

	record := NewResponseRecord(200, headers, `{"result":"ok"}`, 100*time.Millisecond, true)

	assert.Equal(t, 200, record.StatusCode)
	assert.Equal(t, `{"result":"ok"}`, record.Body)
	assert.Equal(t, float64(100), record.DurationMs)
	assert.NotNil(t, record.Headers)
}

func TestNewResponseRecord_WithHeaders(t *testing.T) {
	t.Parallel()

	headers := http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"Bearer token"},
	}

	record := NewResponseRecord(200, headers, "", 0, true)

	assert.NotNil(t, record.Headers)
	assert.Equal(t, "application/json", record.Headers["Content-Type"])
	assert.Equal(t, "Bearer token", record.Headers["Authorization"])
}

func TestNewResponseRecord_WithoutHeaders(t *testing.T) {
	t.Parallel()

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	record := NewResponseRecord(200, headers, "", 0, false)

	assert.Nil(t, record.Headers)
}

func TestNewResponseRecord_NilHeaders(t *testing.T) {
	t.Parallel()

	record := NewResponseRecord(200, nil, "", 0, true)

	assert.Nil(t, record.Headers)
}

func TestNewResponseRecord_DurationMs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		expected float64
	}{
		{"zero duration", 0, 0},
		{"millisecond", 150 * time.Millisecond, 150},
		{"second", 1 * time.Second, 1000},
		{"nanosecond", 123 * time.Nanosecond, 0}, // Milliseconds() returns 0 for sub-millisecond
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			record := NewResponseRecord(200, nil, "", tt.duration, false)
			assert.Equal(t, tt.expected, record.DurationMs)
		})
	}
}

func TestNewResponseRecord_StatusCodes(t *testing.T) {
	t.Parallel()

	statusCodes := []int{200, 201, 301, 400, 404, 500, 503}

	for _, code := range statusCodes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			t.Parallel()

			record := NewResponseRecord(code, nil, "", 0, false)
			assert.Equal(t, code, record.StatusCode)
		})
	}
}

func TestNewResponseRecord_MultiValueHeaders(t *testing.T) {
	t.Parallel()

	headers := http.Header{
		"Set-Cookie": []string{"cookie1=value1", "cookie2=value2"},
	}

	record := NewResponseRecord(200, headers, "", 0, true)

	// Should only take the first value
	assert.Equal(t, "cookie1=value1", record.Headers["Set-Cookie"])
}

func TestLogRecord_Structure(t *testing.T) {
	t.Parallel()

	reqRecord := RequestRecord{
		Timestamp: time.Now(),
		Method:    "GET",
		URL:       "http://example.com",
	}

	respRecord := ResponseRecord{
		StatusCode: 200,
		Body:       `{"ok":true}`,
		DurationMs: 123.45,
	}

	logRecord := LogRecord{
		Request:  reqRecord,
		Response: respRecord,
	}

	assert.Equal(t, "GET", logRecord.Request.Method)
	assert.Equal(t, 200, logRecord.Response.StatusCode)
	assert.Equal(t, `{"ok":true}`, logRecord.Response.Body)
	assert.Equal(t, 123.45, logRecord.Response.DurationMs)
}

func TestRequestRecord_AllFields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	record := RequestRecord{
		Timestamp: now,
		Method:    "POST",
		URL:       "http://example.com/api",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"test":"data"}`,
	}

	assert.Equal(t, now, record.Timestamp)
	assert.Equal(t, "POST", record.Method)
	assert.Equal(t, "http://example.com/api", record.URL)
	assert.Equal(t, "application/json", record.Headers["Content-Type"])
	assert.Equal(t, `{"test":"data"}`, record.Body)
}

func TestResponseRecord_AllFields(t *testing.T) {
	t.Parallel()

	record := ResponseRecord{
		StatusCode: 201,
		Headers: map[string]string{
			"Location": "/new-resource",
		},
		Body:       `{"id":123}`,
		DurationMs: 456.78,
	}

	assert.Equal(t, 201, record.StatusCode)
	assert.Equal(t, "/new-resource", record.Headers["Location"])
	assert.Equal(t, `{"id":123}`, record.Body)
	assert.Equal(t, 456.78, record.DurationMs)
}
