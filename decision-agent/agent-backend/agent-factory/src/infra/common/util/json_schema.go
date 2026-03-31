package util

import (
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// ValidJsonSchema 校验json数据，返回错误字段、错误原因
func ValidJsonSchema(schemaStr, doc string) (invalidFields []string, err error) {
	jsonLoader := gojsonschema.NewStringLoader(schemaStr)

	schema, err := gojsonschema.NewSchema(jsonLoader)
	if err != nil {
		err = errors.Wrap(err, "[ValidJsonSchema]:NewSchema")
		return
	}

	documentLoader := gojsonschema.NewStringLoader(doc)

	result, err := schema.Validate(documentLoader)
	if err != nil {
		err = errors.Wrap(err, "[ValidJsonSchema]:Validate")
		return
	}

	if result.Valid() {
		return
	}

	for _, rsErr := range result.Errors() {
		invalidFields = append(invalidFields, rsErr.String())
	}

	return
}

func IsJsonschemaValid(schemaStr, doc string) (isValid bool, err error) {
	invalidFields, err := ValidJsonSchema(schemaStr, doc)
	if err != nil {
		return
	}

	isValid = len(invalidFields) == 0

	return
}
