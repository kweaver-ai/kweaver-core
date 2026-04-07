package openapidoc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadOpenAPIDocSanitizesLegacySchemaFields(t *testing.T) {
	t.Parallel()

	doc, err := loadOpenAPIDoc([]byte(`{
  "openapi": "3.0.2",
  "info": {
    "title": "Manual",
    "version": "1.0.0"
  },
  "paths": {},
  "components": {
    "schemas": {
      "Legacy": {
        "type": "object",
        "properties": {
          "avatar": {
            "type": "string",
            "example": 1,
            "required": true
          },
          "progress": {
            "type": "array",
            "example": {
              "progress": []
            }
          }
        }
      }
    }
  }
}`))
	require.NoError(t, err)

	schema := doc.Components.Schemas["Legacy"].Value.Properties["avatar"].Value
	require.Equal(t, "1", schema.Example)
	require.Empty(t, schema.Required)
	require.Nil(t, doc.Components.Schemas["Legacy"].Value.Properties["progress"].Value.Example)
}
