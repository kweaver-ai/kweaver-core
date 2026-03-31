package ormhelper

import (
	"fmt"
	"reflect"
)

// DebugFieldMapping 调试字段映射信息
func DebugFieldMapping(structPtr interface{}) {
	if structPtr == nil {
		fmt.Println("结构体指针为nil")
		return
	}

	destValue := reflect.ValueOf(structPtr)
	if destValue.Kind() != reflect.Ptr {
		fmt.Println("参数必须是结构体指针")
		return
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		fmt.Println("参数必须是结构体指针")
		return
	}

	destType := destValue.Type()
	structName := destType.Name()

	fmt.Printf("=== 结构体字段映射调试信息: %s ===\n", structName)
	fmt.Printf("字段总数: %d\n", destType.NumField())

	fieldMap := buildFieldMap(destType)

	fmt.Println("\n字段详情:")
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		dbTag := field.Tag.Get("db")
		jsonTag := field.Tag.Get("json")

		fmt.Printf("  [%d] %s %s", i, field.Name, field.Type)
		if dbTag != "" {
			fmt.Printf(" db:\"%s\"", dbTag)
		}
		if jsonTag != "" {
			fmt.Printf(" json:\"%s\"", jsonTag)
		}
		fmt.Println()
	}

	fmt.Println("\n字段映射表:")
	for dbField, index := range fieldMap {
		field := destType.Field(index)
		fmt.Printf("  \"%s\" -> [%d] %s (%s)\n", dbField, index, field.Name, field.Type)
	}

	fmt.Println("=== 调试信息结束 ===")
}

// DebugColumnMapping 调试列映射信息
func DebugColumnMapping(structPtr interface{}, columns []string) {
	if structPtr == nil || len(columns) == 0 {
		fmt.Println("参数无效")
		return
	}

	destValue := reflect.ValueOf(structPtr)
	if destValue.Kind() != reflect.Ptr {
		fmt.Println("参数必须是结构体指针")
		return
	}

	destValue = destValue.Elem()
	destType := destValue.Type()
	fieldMap := buildFieldMap(destType)

	fmt.Printf("=== 列映射调试信息 ===\n")
	fmt.Printf("数据库列数: %d\n", len(columns))
	fmt.Printf("结构体字段数: %d\n", destType.NumField())

	fmt.Println("\n列映射详情:")
	for i, column := range columns {
		if fieldIndex, exists := fieldMap[column]; exists {
			field := destType.Field(fieldIndex)
			fmt.Printf("  [%d] \"%s\" -> [%d] %s (%s) ✓\n",
				i, column, fieldIndex, field.Name, field.Type)
		} else {
			fmt.Printf("  [%d] \"%s\" -> 未找到匹配字段 ✗\n", i, column)
		}
	}

	fmt.Println("=== 列映射调试结束 ===")
}
