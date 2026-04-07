package sqlhelper2

import (
	"reflect"
	"strings"
)

// AllFieldsByStruct returns all fields of a struct by tag for "select * from table" query
func AllFieldsByStruct(s interface{}, tags ...string) (fields []string) {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// default tag is db
	tag := "db"
	if len(tags) > 0 {
		tag = tags[0]
	}

	fields = make([]string, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		// 支持匿名字段
		if t.Field(i).Anonymous {
			tmpFieldType := t.Field(i).Type

			if tmpFieldType.Kind() == reflect.Ptr {
				tmpFieldType = tmpFieldType.Elem()
			}

			if tmpFieldType.Kind() == reflect.Struct {
				fields = append(fields, AllFieldsByStruct(reflect.New(tmpFieldType).Interface(), tag)...)
			}

			continue
		}

		if key, ok := t.Field(i).Tag.Lookup(tag); ok {
			if key == "" || key == "-" {
				continue
			}

			fields = append(fields, key)
		}
	}

	return fields
}

// GenSQLSelectFieldsStr generate sql select fields string
func GenSQLSelectFieldsStr(fields []string, tableFlag string) string {
	b := strings.Builder{}

	for _, v := range fields {
		if tableFlag != "" {
			b.WriteString(tableFlag)
			b.WriteString(".")
		}

		b.WriteString(v)
		b.WriteString(",")
	}

	return strings.TrimRight(b.String(), ",")
}
