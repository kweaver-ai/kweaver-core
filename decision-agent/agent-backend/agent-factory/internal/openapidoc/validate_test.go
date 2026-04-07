package openapidoc

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestValidateOpenAPIRejectsInvalidDoc(t *testing.T) {
	t.Parallel()

	doc := &openapi3.T{
		OpenAPI: "3.0.2",
		Info:    &openapi3.Info{Title: "broken", Version: "1.0.0"},
	}

	err := ValidateOpenAPI(context.Background(), doc)
	require.Error(t, err)
}
