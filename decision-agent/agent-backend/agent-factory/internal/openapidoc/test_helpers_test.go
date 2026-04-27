package openapidoc

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func mustLoadDoc(t *testing.T, data string) *openapi3.T {
	t.Helper()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(strings.TrimSpace(data)))
	require.NoError(t, err)

	return doc
}
