package cutil

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type JSONSchemaProperty struct {
	Type        string                         `json:"type"`
	Title       string                         `json:"title"`
	Description string                         `json:"description"`
	Items       *JSONSchemaProperty            `json:"items,omitempty"`
	Properties  map[string]*JSONSchemaProperty `json:"properties,omitempty"`
}

var typeCache sync.Map // map[string]reflect.Type

func generateTypeCacheKey(schema *JSONSchemaProperty) string {
	bytes, _ := json.Marshal(schema)
	return fmt.Sprintf("%x", md5.Sum(bytes))
}

func CreateDynamicStruct(schemaStr string) (reflect.Type, error) {
	var schema map[string]*JSONSchemaProperty
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		return nil, fmt.Errorf("解析schema失败: %w", err)
	}

	var wg sync.WaitGroup

	errChan := make(chan error, len(schema))
	structTypes := &sync.Map{}

	for fieldName, prop := range schema {
		if prop.Type == "array" && prop.Items != nil && prop.Items.Type == "object" {
			wg.Add(1)

			go func(name string, p *JSONSchemaProperty) {
				defer wg.Done()

				subType, err := createStructType(p.Items.Properties, structTypes)
				if err != nil {
					errChan <- err
					return
				}

				structTypes.Store(name, reflect.SliceOf(subType))
			}(fieldName, prop)
		}
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	fields := make([]reflect.StructField, 0, len(schema))

	for fieldName, prop := range schema {
		field, err := createField(fieldName, prop, structTypes)
		if err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}

	return reflect.StructOf(fields), nil
}

func createStructType(properties map[string]*JSONSchemaProperty, structTypes *sync.Map) (reflect.Type, error) {
	cacheKey := generateTypeCacheKey(&JSONSchemaProperty{Type: "object", Properties: properties})
	if cachedType, ok := typeCache.Load(cacheKey); ok {
		return cachedType.(reflect.Type), nil
	}

	fields := make([]reflect.StructField, 0, len(properties))

	var wg sync.WaitGroup

	fieldChan := make(chan reflect.StructField, len(properties))
	errChan := make(chan error, len(properties))

	for fieldName, prop := range properties {
		wg.Add(1)

		go func(name string, p *JSONSchemaProperty) {
			defer wg.Done()

			field, err := createField(name, p, structTypes)
			if err != nil {
				errChan <- err
				return
			}
			fieldChan <- field
		}(fieldName, prop)
	}

	wg.Wait()
	close(fieldChan)
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	for field := range fieldChan {
		fields = append(fields, field)
	}

	structType := reflect.StructOf(fields)

	typeCache.Store(cacheKey, structType)

	return structType, nil
}

func createField(fieldName string, prop *JSONSchemaProperty, structTypes *sync.Map) (reflect.StructField, error) {
	var fieldType reflect.Type

	switch prop.Type {
	case "string":
		fieldType = reflect.TypeOf("")
	case "boolean":
		fieldType = reflect.TypeOf(true)
	case "number":
		fieldType = reflect.TypeOf(float64(0))
	case "integer":
		fieldType = reflect.TypeOf(int64(0))
	case "array":
		if prop.Items != nil {
			if existingType, ok := structTypes.Load(fieldName); ok {
				fieldType = existingType.(reflect.Type)
			} else {
				elemType, err := getArrayElementType(prop.Items)
				if err != nil {
					return reflect.StructField{}, err
				}

				fieldType = reflect.SliceOf(elemType)
			}
		} else {
			fieldType = reflect.TypeOf([]interface{}{})
		}
	case "object":
		if prop.Properties != nil {
			subType, err := createStructType(prop.Properties, structTypes)
			if err != nil {
				return reflect.StructField{}, err
			}

			fieldType = subType
		} else {
			fieldType = reflect.TypeOf(map[string]interface{}{})
		}
	default:
		fieldType = reflect.TypeOf((*interface{})(nil)).Elem()
	}

	return reflect.StructField{
		Name: toCamelCase(fieldName),
		Type: fieldType,
		Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, fieldName)),
	}, nil
}

func getArrayElementType(prop *JSONSchemaProperty) (reflect.Type, error) {
	switch prop.Type {
	case "string":
		return reflect.TypeOf(""), nil
	case "boolean":
		return reflect.TypeOf(true), nil
	case "number":
		return reflect.TypeOf(float64(0)), nil
	case "integer":
		return reflect.TypeOf(int64(0)), nil
	case "object":
		if prop.Properties != nil {
			return createStructType(prop.Properties, nil)
		}

		return reflect.TypeOf(map[string]interface{}{}), nil
	default:
		return reflect.TypeOf((*interface{})(nil)).Elem(), nil
	}
}

func toCamelCase(s string) string {
	caser := cases.Title(language.English)
	words := strings.Split(s, "_")

	for i := range words {
		words[i] = caser.String(words[i])
	}

	return strings.Join(words, "")
}
