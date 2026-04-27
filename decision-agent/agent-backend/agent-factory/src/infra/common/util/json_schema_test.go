package util

import (
	"testing"
)

func TestValidJsonSchema(t *testing.T) {
	t.Parallel()

	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number"},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["name", "age"]
	}`

	tests := []struct {
		name           string
		schemaStr      string
		doc            string
		wantErr        bool
		wantFieldCount int
	}{
		{
			name:      "valid document",
			schemaStr: schema,
			doc:       `{"name": "John", "age": 30}`,
			wantErr:   false,
		},
		{
			name:      "valid document with optional field",
			schemaStr: schema,
			doc:       `{"name": "Jane", "age": 25, "email": "jane@example.com"}`,
			wantErr:   false,
		},
		{
			name:           "missing required field",
			schemaStr:      schema,
			doc:            `{"age": 30}`,
			wantErr:        false,
			wantFieldCount: 1,
		},
		{
			name:           "wrong type",
			schemaStr:      schema,
			doc:            `{"name": "John", "age": "thirty"}`,
			wantErr:        false,
			wantFieldCount: 1,
		},
		{
			name:           "invalid email format",
			schemaStr:      schema,
			doc:            `{"name": "John", "age": 30, "email": "not-an-email"}`,
			wantErr:        false,
			wantFieldCount: 1,
		},
		{
			name:           "empty document",
			schemaStr:      schema,
			doc:            `{}`,
			wantErr:        false,
			wantFieldCount: 2,
		},
		{
			name:      "all valid",
			schemaStr: schema,
			doc:       `{"name": "Test", "age": 20, "email": "test@test.com"}`,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			invalidFields, err := ValidJsonSchema(tt.schemaStr, tt.doc)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidJsonSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantFieldCount > 0 && len(invalidFields) != tt.wantFieldCount {
				t.Errorf("ValidJsonSchema() invalidFields count = %d, want %d", len(invalidFields), tt.wantFieldCount)
			}
		})
	}
}

func TestValidJsonSchema_InvalidSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		schemaStr string
		doc       string
	}{
		{
			name:      "invalid JSON schema",
			schemaStr: `{invalid json}`,
			doc:       `{"name": "test"}`,
		},
		{
			name:      "empty schema",
			schemaStr: ``,
			doc:       `{"name": "test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ValidJsonSchema(tt.schemaStr, tt.doc)
			if err == nil {
				t.Error("ValidJsonSchema() should return error for invalid schema")
			}
		})
	}
}

func TestValidJsonSchema_InvalidDocument(t *testing.T) {
	t.Parallel()

	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		},
		"required": ["name"]
	}`

	tests := []struct {
		name string
		doc  string
	}{
		{
			name: "invalid JSON document",
			doc:  `{invalid json}`,
		},
		{
			name: "malformed JSON document",
			doc:  `{"name": "test",}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ValidJsonSchema(schema, tt.doc)
			if err == nil {
				t.Error("ValidJsonSchema() should return error for invalid document")
			}
		})
	}
}

func TestIsJsonschemaValid(t *testing.T) {
	t.Parallel()

	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number"}
		},
		"required": ["name", "age"]
	}`

	tests := []struct {
		name     string
		schema   string
		doc      string
		wantErr  bool
		wantBool bool
	}{
		{
			name:     "valid document",
			schema:   schema,
			doc:      `{"name": "John", "age": 30}`,
			wantErr:  false,
			wantBool: true,
		},
		{
			name:     "invalid document - missing required",
			schema:   schema,
			doc:      `{"name": "John"}`,
			wantErr:  false,
			wantBool: false,
		},
		{
			name:     "invalid document - wrong type",
			schema:   schema,
			doc:      `{"name": "John", "age": "thirty"}`,
			wantErr:  false,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isValid, err := IsJsonschemaValid(tt.schema, tt.doc)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsJsonschemaValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if isValid != tt.wantBool {
				t.Errorf("IsJsonschemaValid() = %v, want %v", isValid, tt.wantBool)
			}
		})
	}
}

func TestIsJsonschemaValid_InvalidSchema(t *testing.T) {
	t.Parallel()

	_, err := IsJsonschemaValid(`{invalid}`, `{"name": "test"}`)
	if err == nil {
		t.Error("IsJsonschemaValid() should return error for invalid schema")
	}
}

func TestValidJsonSchema_ComplexSchema(t *testing.T) {
	t.Parallel()

	schema := `{
		"type": "object",
		"properties": {
			"users": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"id": {"type": "integer"},
						"name": {"type": "string"}
					},
					"required": ["id", "name"]
				}
			}
		}
	}`

	validDoc := `{
		"users": [
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"}
		]
	}`

	invalidDoc := `{
		"users": [
			{"id": 1, "name": "Alice"},
			{"name": "Bob"}
		]
	}`

	t.Run("valid complex document", func(t *testing.T) {
		t.Parallel()

		invalidFields, err := ValidJsonSchema(schema, validDoc)
		if err != nil {
			t.Errorf("ValidJsonSchema() error = %v", err)
		}

		if len(invalidFields) != 0 {
			t.Errorf("ValidJsonSchema() should have no errors, got %d", len(invalidFields))
		}
	})

	t.Run("invalid complex document", func(t *testing.T) {
		t.Parallel()

		invalidFields, err := ValidJsonSchema(schema, invalidDoc)
		if err != nil {
			t.Errorf("ValidJsonSchema() error = %v", err)
		}

		if len(invalidFields) == 0 {
			t.Error("ValidJsonSchema() should have errors")
		}
	})
}
