package ormhelper

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// buildFieldMap 构建字段映射表
func buildFieldMap(structType reflect.Type) map[string]int {
	fieldMap := make(map[string]int)
	numField := structType.NumField()
	for i := 0; i < numField; i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("db")
		if tag == "" {
			tag = field.Tag.Get("json")
		}
		if tag != "" && tag != "-" {
			// 处理tag中的选项，只取字段名
			if idx := strings.Index(tag, ","); idx != -1 {
				tag = tag[:idx]
			}
			fieldMap[tag] = i
		}
	}
	return fieldMap
}

// prepareScanTargets 准备扫描目标
func prepareScanTargets(structValue reflect.Value, columns []string, fieldMap map[string]int) []interface{} {
	scanTargets := make([]interface{}, len(columns))
	for i, column := range columns {
		if fieldIndex, exists := fieldMap[column]; exists {
			fieldValue := structValue.Field(fieldIndex)
			if fieldValue.CanSet() {
				scanTargets[i] = fieldValue.Addr().Interface()
			} else {
				var dummy interface{}
				scanTargets[i] = &dummy
			}
		} else {
			var dummy interface{}
			scanTargets[i] = &dummy
		}
	}
	return scanTargets
}

// structScanner 结构体扫描器
type structScanner struct{}

// NewScanner 创建新的扫描器
func NewScanner() Scanner {
	return &structScanner{}
}

// ScanOne 扫描单行到结构体
// 由于sql.Row没有Columns()方法，无法获取列信息进行字段映射
// 这个方法现在已废弃，建议使用ScanOneWithColumns
func (s *structScanner) ScanOne(row *sql.Row, dest interface{}) error {
	return fmt.Errorf("ScanOne method is deprecated due to lack of column information in sql.Row. Use ScanOneWithColumns instead")
}

// ScanOneWithColumns 扫描单行到结构体（支持字段映射）
func (s *structScanner) ScanOneWithColumns(row *sql.Row, dest interface{}, columns []string) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	destType := destValue.Type()

	// 创建字段映射
	fieldMap := buildFieldMap(destType)

	// 准备扫描目标
	scanTargets := prepareScanTargets(destValue, columns, fieldMap)

	return row.Scan(scanTargets...)
}

// ScanMany 扫描多行到结构体切片
func (s *structScanner) ScanMany(rows *sql.Rows, dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	// 获取切片元素类型
	sliceType := destValue.Type()
	elemType := sliceType.Elem()

	// 如果是指针类型，获取实际的结构体类型
	structType := elemType
	isPointer := false
	if elemType.Kind() == reflect.Ptr {
		isPointer = true
		structType = elemType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return fmt.Errorf("slice element must be struct or pointer to struct")
	}

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 创建字段映射
	fieldMap := buildFieldMap(structType)

	// 扫描所有行
	results := reflect.MakeSlice(sliceType, 0, 0)
	for rows.Next() {
		// 创建新的结构体实例
		var elemValue reflect.Value
		if isPointer {
			elemValue = reflect.New(structType)
		} else {
			elemValue = reflect.New(structType).Elem()
		}

		structValue := elemValue
		if isPointer {
			structValue = elemValue.Elem()
		}

		// 准备扫描目标
		scanTargets := prepareScanTargets(structValue, columns, fieldMap)

		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}

		results = reflect.Append(results, elemValue)
	}

	destValue.Set(results)
	return rows.Err()
}
