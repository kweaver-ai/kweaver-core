package agentsvc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatSSEMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
		want []byte
	}{
		{
			name: "formats data as SSE message",
			data: "test data",
			want: []byte("data: test data\n\n"),
		},
		{
			name: "handles empty data",
			data: "",
			want: []byte("data: \n\n"),
		},
		{
			name: "handles JSON data",
			data: `{"key":"value"}`,
			want: []byte("data: {\"key\":\"value\"}\n\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := formatSSEMessage(tt.data)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFormatChange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ch   Change
		want string
	}{
		{
			name: "formats change with string content",
			ch: Change{
				SeqID:   1,
				Key:     []interface{}{"path", "to", "field"},
				Content: "value",
				Action:  "upsert",
			},
			want: `{"seq_id": 1, "key": ["path","to","field"], "content": "value", "action": "upsert"}`,
		},
		{
			name: "formats change with numeric content",
			ch: Change{
				SeqID:   2,
				Key:     []interface{}{"count"},
				Content: 42,
				Action:  "append",
			},
			want: `{"seq_id": 2, "key": ["count"], "content": 42, "action": "append"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := formatChange(tt.ch)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestStreamDiff(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runDiff := func(oldJSON, newJSON []byte) (chan []byte, error) {
		lastSeq := 0
		out := make(chan []byte, 100)

		err := StreamDiff(ctx, &lastSeq, oldJSON, newJSON, out, 0)

		return out, err
	}

	t.Run("same JSON produces no changes", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"name":"test","value":123}`)
		newJSON := []byte(`{"name":"test","value":123}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.Empty(t, out)
	})

	t.Run("different objects produce changes", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"name":"test","value":123}`)
		newJSON := []byte(`{"name":"test","value":456}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)
	})

	t.Run("string append produces append action", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"text":"hello"}`)
		newJSON := []byte(`{"text":"hello world"}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)
	})

	t.Run("invalid old JSON returns error", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{invalid json}`)
		newJSON := []byte(`{"name":"test"}`)

		_, err := runDiff(oldJSON, newJSON)
		assert.Error(t, err)
	})

	t.Run("invalid new JSON returns error", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"name":"test"}`)
		newJSON := []byte(`{invalid json}`)

		_, err := runDiff(oldJSON, newJSON)
		assert.Error(t, err)
	})

	t.Run("new field added produces upsert", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"name":"test"}`)
		newJSON := []byte(`{"name":"test","value":123}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)

		result := <-out
		assert.Contains(t, string(result), `"action": "upsert"`)
	})

	t.Run("field removed produces remove", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"name":"test","value":123}`)
		newJSON := []byte(`{"name":"test"}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)

		result := <-out
		assert.Contains(t, string(result), `"action": "remove"`)
	})

	t.Run("array element added produces append", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"items":[1,2]}`)
		newJSON := []byte(`{"items":[1,2,3]}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)

		result := <-out
		assert.Contains(t, string(result), `"action": "append"`)
	})

	t.Run("array element removed produces remove", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"items":[1,2,3]}`)
		newJSON := []byte(`{"items":[1,2]}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)

		result := <-out
		assert.Contains(t, string(result), `"action": "remove"`)
	})

	t.Run("array element changed produces upsert", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"items":[1,2,3]}`)
		newJSON := []byte(`{"items":[1,5,3]}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)

		result := <-out
		assert.Contains(t, string(result), `"action": "upsert"`)
	})

	t.Run("nested object diff", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"user":{"name":"test","age":30}}`)
		newJSON := []byte(`{"user":{"name":"test","age":31}}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)
	})

	t.Run("nested array diff", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"matrix":[[1,2],[3,4]]}`)
		newJSON := []byte(`{"matrix":[[1,2],[3,5]]}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)
	})

	t.Run("type change produces upsert", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"value":123}`)
		newJSON := []byte(`{"value":"123"}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)

		result := <-out
		assert.Contains(t, string(result), `"action": "upsert"`)
	})

	t.Run("array element type change produces upsert", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"items":[1,2,3]}`)
		newJSON := []byte(`{"items":[1,"two",3]}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)

		result := <-out
		assert.Contains(t, string(result), `"action": "upsert"`)
	})

	t.Run("string replacement produces upsert", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"text":"hello"}`)
		newJSON := []byte(`{"text":"goodbye"}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.NotEmpty(t, out)

		result := <-out
		assert.Contains(t, string(result), `"action": "upsert"`)
	})

	t.Run("empty objects", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{}`)
		newJSON := []byte(`{}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.Empty(t, out)
	})

	t.Run("empty arrays", func(t *testing.T) {
		t.Parallel()

		oldJSON := []byte(`{"items":[]}`)
		newJSON := []byte(`{"items":[]}`)

		out, err := runDiff(oldJSON, newJSON)
		assert.NoError(t, err)
		assert.Empty(t, out)
	})
}
