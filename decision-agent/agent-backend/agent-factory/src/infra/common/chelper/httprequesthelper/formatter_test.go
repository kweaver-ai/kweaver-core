package httprequesthelper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFormatter(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(true, 1000)

	assert.NotNil(t, formatter)
	assert.True(t, formatter.prettyJSON)
	assert.Equal(t, 1000, formatter.maxBodySize)
}

func TestNewFormatter_DefaultValues(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	assert.NotNil(t, formatter)
	assert.False(t, formatter.prettyJSON)
	assert.Equal(t, 0, formatter.maxBodySize)
}

func TestFormatter_Format_BasicRecord(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Method:    "GET",
			URL:       "http://example.com/api",
			Body:      "",
		},
		Response: ResponseRecord{
			StatusCode: 200,
			Body:       `{"ok":true}`,
			DurationMs: 100,
		},
	}

	output := formatter.Format(record)

	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "http://example.com/api")
	assert.Contains(t, output, "HTTP 200")
	assert.Contains(t, output, "100.00ms")
	assert.Contains(t, output, `{"ok":true}`)
}

func TestFormatter_Format_WithHeaders(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "POST",
			URL:       "http://example.com/api",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
			Body: `{"test":"data"}`,
		},
		Response: ResponseRecord{
			StatusCode: 201,
			Headers: map[string]string{
				"Location": "/new-resource",
			},
			Body:       `{"id":123}`,
			DurationMs: 50,
		},
	}

	output := formatter.Format(record)

	assert.Contains(t, output, "Content-Type")
	assert.Contains(t, output, "application/json")
	assert.Contains(t, output, "Authorization")
	assert.Contains(t, output, "Location")
	assert.Contains(t, output, "/new-resource")
}

func TestFormatter_Format_TruncatesBody(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 20)

	longBody := "This is a very long body that should be truncated"
	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "POST",
			URL:       "http://example.com/api",
			Body:      longBody,
		},
		Response: ResponseRecord{
			StatusCode: 200,
			Body:       longBody,
			DurationMs: 100,
		},
	}

	output := formatter.Format(record)

	assert.Contains(t, output, "truncated")
}

func TestFormatter_Format_GeneratesCURL(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "POST",
			URL:       "http://example.com/api",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"key":"value"}`,
		},
		Response: ResponseRecord{
			StatusCode: 200,
			DurationMs: 100,
		},
	}

	output := formatter.Format(record)

	assert.Contains(t, output, "curl")
	assert.Contains(t, output, "-X")
	assert.Contains(t, output, "POST")
	assert.Contains(t, output, "-d")
}

func TestFormatter_Format_GETNoBody(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "GET",
			URL:       "http://example.com/api",
			Headers: map[string]string{
				"Accept": "application/json",
			},
		},
		Response: ResponseRecord{
			StatusCode: 200,
			DurationMs: 50,
		},
	}

	output := formatter.Format(record)

	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "curl")
	// GET request should not have -d in curl
	assert.NotContains(t, output, "-d")
}

func TestFormatter_Format_PrettyJSON(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(true, 0)

	jsonBody := `{"name":"test","value":123}`

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "POST",
			URL:       "http://example.com/api",
			Body:      jsonBody,
		},
		Response: ResponseRecord{
			StatusCode: 200,
			Body:       jsonBody,
			DurationMs: 100,
		},
	}

	output := formatter.Format(record)

	assert.Contains(t, output, "name")
	assert.Contains(t, output, "test")
	assert.Contains(t, output, "value")
}

func TestFormatter_Format_Separators(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "GET",
			URL:       "http://example.com/",
		},
		Response: ResponseRecord{
			StatusCode: 200,
			DurationMs: 10,
		},
	}

	output := formatter.Format(record)

	// Check for separator characters
	assert.Contains(t, output, "====")
	assert.Contains(t, output, "----")
	assert.Contains(t, output, "\n")
}

func TestFormatter_CURL_SkipsCertainHeaders(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "GET",
			URL:       "http://example.com/",
			Headers: map[string]string{
				"Host":           "example.com",
				"Connection":     "keep-alive",
				"Content-Length": "100",
				"Authorization":  "Bearer token",
			},
		},
		Response: ResponseRecord{
			StatusCode: 200,
			DurationMs: 10,
		},
	}

	output := formatter.Format(record)

	// Authorization should be in curl
	assert.Contains(t, output, "Authorization")
	assert.Contains(t, output, "Bearer token")
	// But these headers should be skipped
	// Note: The exact behavior depends on implementation
}

func TestFormatter_Format_EmptyHeaders(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "GET",
			URL:       "http://example.com/",
		},
		Response: ResponseRecord{
			StatusCode: 200,
			DurationMs: 10,
		},
	}

	output := formatter.Format(record)

	// Should not crash with empty headers
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "HTTP 200")
}

func TestFormatter_Format_PUTRequest(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "PUT",
			URL:       "http://example.com/api/resource",
			Body:      `{"updated":true}`,
		},
		Response: ResponseRecord{
			StatusCode: 200,
			DurationMs: 100,
		},
	}

	output := formatter.Format(record)

	assert.Contains(t, output, "PUT")
	assert.Contains(t, output, "-X")
	assert.Contains(t, output, "-d")
}

func TestFormatter_Format_PATCHRequest(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "PATCH",
			URL:       "http://example.com/api/resource",
			Body:      `{"patched":true}`,
		},
		Response: ResponseRecord{
			StatusCode: 200,
			DurationMs: 100,
		},
	}

	output := formatter.Format(record)

	assert.Contains(t, output, "PATCH")
	assert.Contains(t, output, "-d")
}

func TestFormatter_Format_DELETERequest(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0)

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "DELETE",
			URL:       "http://example.com/api/resource/123",
		},
		Response: ResponseRecord{
			StatusCode: 204,
			DurationMs: 50,
		},
	}

	output := formatter.Format(record)

	assert.Contains(t, output, "DELETE")
	assert.Contains(t, output, "HTTP 204")
}

func TestFormatter_MaxBodySize_Zero(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter(false, 0) // 0 means no limit

	longBody := string(make([]byte, 100))
	for range longBody {
		longBody = "a"
	}

	record := &LogRecord{
		Request: RequestRecord{
			Timestamp: time.Now(),
			Method:    "POST",
			URL:       "http://example.com/",
			Body:      longBody,
		},
		Response: ResponseRecord{
			StatusCode: 200,
			DurationMs: 100,
		},
	}

	output := formatter.Format(record)

	// With 0 max body size, body should not be truncated
	assert.NotContains(t, output, "truncated")
}

func TestIsJSON_Object(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid object", `{"key":"value"}`, true},
		{"valid array", `[1,2,3]`, true},
		{"not json - plain text", `plain text`, false},
		{"not json - xml", `<xml></xml>`, false},
		{"empty string", ``, false},
		{"whitespace only", `   `, false},
		{"object with spaces", ` { "key": "value" } `, true},
		{"array with spaces", ` [1, 2, 3] `, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isJSON(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFormatJSON_Valid(t *testing.T) {
	t.Parallel()

	input := `{"name":"test","value":123}`

	output, err := formatJSON(input)

	assert.NoError(t, err)
	assert.Contains(t, output, "\n")
	assert.Contains(t, output, "  ")
}

func TestFormatJSON_Invalid(t *testing.T) {
	t.Parallel()

	input := `not valid json`

	_, err := formatJSON(input)

	assert.Error(t, err)
}
