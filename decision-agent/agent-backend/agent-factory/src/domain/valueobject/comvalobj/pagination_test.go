package comvalobj

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagination_StructFields(t *testing.T) {
	t.Parallel()

	t.Run("creates pagination with values", func(t *testing.T) {
		t.Parallel()

		p := Pagination{
			Offset: 10,
			Limit:  20,
		}

		assert.Equal(t, 10, p.Offset)
		assert.Equal(t, 20, p.Limit)
	})

	t.Run("creates pagination with zero values", func(t *testing.T) {
		t.Parallel()

		p := Pagination{}

		assert.Zero(t, p.Offset)
		assert.Zero(t, p.Limit)
	})

	t.Run("serializes to JSON", func(t *testing.T) {
		t.Parallel()

		p := Pagination{
			Offset: 5,
			Limit:  15,
		}

		data, err := json.Marshal(p)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "\"offset\":5")
		assert.Contains(t, string(data), "\"limit\":15")
	})

	t.Run("deserializes from JSON", func(t *testing.T) {
		t.Parallel()

		jsonStr := `{"offset":10,"limit":25}`

		var p Pagination
		err := json.Unmarshal([]byte(jsonStr), &p)

		assert.NoError(t, err)
		assert.Equal(t, 10, p.Offset)
		assert.Equal(t, 25, p.Limit)
	})
}

func TestPagination_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("negative offset", func(t *testing.T) {
		t.Parallel()

		p := Pagination{
			Offset: -5,
			Limit:  10,
		}

		assert.Equal(t, -5, p.Offset)
		assert.Equal(t, 10, p.Limit)
	})

	t.Run("negative limit", func(t *testing.T) {
		t.Parallel()

		p := Pagination{
			Offset: 0,
			Limit:  -1,
		}

		assert.Equal(t, 0, p.Offset)
		assert.Equal(t, -1, p.Limit)
	})

	t.Run("large values", func(t *testing.T) {
		t.Parallel()

		p := Pagination{
			Offset: 1000000,
			Limit:  100000,
		}

		assert.Equal(t, 1000000, p.Offset)
		assert.Equal(t, 100000, p.Limit)
	})

	t.Run("max int values", func(t *testing.T) {
		t.Parallel()

		p := Pagination{
			Offset: 2147483647,
			Limit:  2147483647,
		}

		assert.Equal(t, 2147483647, p.Offset)
		assert.Equal(t, 2147483647, p.Limit)
	})
}

func TestPagination_JSONEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("marshal and unmarshal roundtrip", func(t *testing.T) {
		t.Parallel()

		original := Pagination{
			Offset: 100,
			Limit:  50,
		}

		data, err := json.Marshal(original)
		assert.NoError(t, err)

		var decoded Pagination
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)

		assert.Equal(t, original.Offset, decoded.Offset)
		assert.Equal(t, original.Limit, decoded.Limit)
	})

	t.Run("unmarshal empty JSON", func(t *testing.T) {
		t.Parallel()

		data := []byte(`{}`)

		var p Pagination
		err := json.Unmarshal(data, &p)
		assert.NoError(t, err)

		assert.Equal(t, 0, p.Offset)
		assert.Equal(t, 0, p.Limit)
	})

	t.Run("marshal zero values to JSON", func(t *testing.T) {
		t.Parallel()

		p := Pagination{}

		data, err := json.Marshal(p)
		assert.NoError(t, err)

		expected := `{"offset":0,"limit":0}`
		assert.JSONEq(t, expected, string(data))
	})
}

func TestPagination_Pointer(t *testing.T) {
	t.Parallel()

	t.Run("create pointer to pagination", func(t *testing.T) {
		t.Parallel()

		p := &Pagination{
			Offset: 5,
			Limit:  10,
		}

		assert.NotNil(t, p)
		assert.Equal(t, 5, p.Offset)
		assert.Equal(t, 10, p.Limit)
	})

	t.Run("nil pagination pointer", func(t *testing.T) {
		t.Parallel()

		var p *Pagination

		assert.Nil(t, p)
	})
}
