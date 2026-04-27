package util

import "strings"

// DynamicFieldsHolder 是一个可复用的结构体，用于实现动态字段功能
// 其他结构体可以通过嵌入该类型来获得动态字段能力
type DynamicFieldsHolder struct {
	DynamicFields map[string]interface{} `json:"-"` // 不直接序列化
}

// NewDynamicFieldsHolder 创建一个新的 DynamicFieldsHolder 实例
func NewDynamicFieldsHolder() DynamicFieldsHolder {
	return DynamicFieldsHolder{
		DynamicFields: make(map[string]interface{}),
	}
}

// SetField 设置动态字段的值
func (d *DynamicFieldsHolder) SetField(key string, value interface{}) {
	if d.DynamicFields == nil {
		d.DynamicFields = make(map[string]interface{})
	}

	d.DynamicFields[key] = value
}

// GetField 获取动态字段的值
func (d *DynamicFieldsHolder) GetField(key string) (interface{}, bool) {
	if d.DynamicFields == nil {
		return nil, false
	}

	val, ok := d.DynamicFields[key]

	return val, ok
}

func (d *DynamicFieldsHolder) GetFields(keys []string) map[string]interface{} {
	if d.DynamicFields == nil {
		return nil
	}

	result := make(map[string]interface{})

	for _, key := range keys {
		val, ok := d.DynamicFields[key]
		if ok {
			result[key] = val
		}
	}

	return result
}

func (d *DynamicFieldsHolder) GetFieldsByPrefix(prefix string) map[string]interface{} {
	if d.DynamicFields == nil {
		return nil
	}

	result := make(map[string]interface{})

	for key, val := range d.DynamicFields {
		if strings.HasPrefix(key, prefix) {
			result[key] = val
		}
	}

	return result
}

func (d *DynamicFieldsHolder) GetFieldSliceStr(key string) []string {
	if d.DynamicFields == nil {
		return nil
	}

	val, ok := d.DynamicFields[key]
	if !ok {
		return nil
	}

	switch val := val.(type) {
	case []string:
		return val
	case []interface{}:
		var result []string

		for _, v := range val {
			result = append(result, v.(string))
		}

		return result
	default:
		return nil
	}
}

// AddDynamicFieldsToMap 将动态字段添加到指定的 map 中
func (d *DynamicFieldsHolder) AddDynamicFieldsToMap(m map[string]interface{}) {
	if d.DynamicFields == nil {
		return
	}

	for k, v := range d.DynamicFields {
		m[k] = v
	}
}
