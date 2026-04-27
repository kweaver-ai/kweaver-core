package util

import (
	"testing"
)

func TestNewDynamicFieldsHolder(t *testing.T) {
	t.Parallel()

	holder := NewDynamicFieldsHolder()
	if holder.DynamicFields == nil {
		t.Error("DynamicFields should be initialized")
	}

	if len(holder.DynamicFields) != 0 {
		t.Error("DynamicFields should be empty")
	}
}

func TestDynamicFieldsHolder_SetField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		key       string
		value     interface{}
		wantValue interface{}
	}{
		{"设置字符串", "name", "test", "test"},
		{"设置整数", "age", 25, 25},
		{"设置浮点数", "price", 19.99, 19.99},
		{"设置布尔值", "active", true, true},
		{"设置字符串slice", "tags", []string{"a", "b"}, nil},
		{"设置nil", "null", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			holder := NewDynamicFieldsHolder()
			holder.SetField(tt.key, tt.value)

			val, ok := holder.GetField(tt.key)
			if !ok {
				t.Errorf("GetField(%s) should return ok=true", tt.key)
			}

			if tt.key == "tags" {
				sliceVal, ok := val.([]string)
				if !ok || len(sliceVal) != 2 {
					t.Errorf("GetField(%s) = %v, want []string with 2 elements", tt.key, val)
				}

				if sliceVal[0] != "a" || sliceVal[1] != "b" {
					t.Errorf("GetField(%s) = %v, want [a b]", tt.key, sliceVal)
				}
			} else if val != tt.wantValue {
				t.Errorf("GetField(%s) = %v, want %v", tt.key, val, tt.wantValue)
			}
		})
	}

	t.Run("set field on nil DynamicFields", func(t *testing.T) {
		t.Parallel()

		holder := DynamicFieldsHolder{}
		holder.SetField("key", "value")

		val, ok := holder.GetField("key")
		if !ok {
			t.Error("GetField should return ok=true after SetField on nil map")
		}

		if val != "value" {
			t.Errorf("GetField(key) = %v, want value", val)
		}
	})
}

func TestDynamicFieldsHolder_GetField(t *testing.T) {
	t.Parallel()

	holder := NewDynamicFieldsHolder()
	holder.SetField("key1", "value1")
	holder.SetField("key2", 42)

	tests := []struct {
		name    string
		key     string
		wantVal interface{}
		wantOk  bool
	}{
		{"获取存在的key", "key1", "value1", true},
		{"获取另一个存在的key", "key2", 42, true},
		{"获取不存在的key", "key3", nil, false},
		{"获取空key", "", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			val, ok := holder.GetField(tt.key)
			if ok != tt.wantOk {
				t.Errorf("GetField(%s) ok = %v, want %v", tt.key, ok, tt.wantOk)
			}

			if val != tt.wantVal {
				t.Errorf("GetField(%s) = %v, want %v", tt.key, val, tt.wantVal)
			}
		})
	}

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()

		holder := DynamicFieldsHolder{}

		val, ok := holder.GetField("key")
		if ok {
			t.Error("GetField on nil map should return ok=false")
		}

		if val != nil {
			t.Error("GetField on nil map should return nil")
		}
	})
}

func TestDynamicFieldsHolder_GetFields(t *testing.T) {
	t.Parallel()

	holder := NewDynamicFieldsHolder()
	holder.SetField("key1", "value1")
	holder.SetField("key2", 42)
	holder.SetField("key3", true)

	tests := []struct {
		name string
		keys []string
		want int
	}{
		{"获取所有存在的key", []string{"key1", "key2"}, 2},
		{"获取部分存在的key", []string{"key1", "key2", "key4"}, 2},
		{"获取不存在的key", []string{"key4", "key5"}, 0},
		{"空slice", []string{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := holder.GetFields(tt.keys)
			if len(result) != tt.want {
				t.Errorf("GetFields() = %d keys, want %d", len(result), tt.want)
			}
		})
	}

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()

		holder := DynamicFieldsHolder{}

		result := holder.GetFields([]string{"key1"})
		if result != nil {
			t.Error("GetFields on nil map should return nil")
		}
	})
}

func TestDynamicFieldsHolder_GetFieldsByPrefix(t *testing.T) {
	t.Parallel()

	holder := NewDynamicFieldsHolder()
	holder.SetField("user.name", "John")
	holder.SetField("user.age", 30)
	holder.SetField("user.email", "test@example.com")
	holder.SetField("product.id", 1)
	holder.SetField("product.name", "Test Product")
	holder.SetField("other.key", "value")

	tests := []struct {
		name   string
		prefix string
		want   int
	}{
		{"获取user前缀", "user.", 3},
		{"获取product前缀", "product.", 2},
		{"获取other前缀", "other.", 1},
		{"获取不存在的前缀", "empty.", 0},
		{"空前缀", "", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := holder.GetFieldsByPrefix(tt.prefix)
			if len(result) != tt.want {
				t.Errorf("GetFieldsByPrefix(%s) = %d keys, want %d", tt.prefix, len(result), tt.want)
			}
		})
	}

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()

		holder := DynamicFieldsHolder{}

		result := holder.GetFieldsByPrefix("user.")
		if result != nil {
			t.Error("GetFieldsByPrefix on nil map should return nil")
		}
	})
}

func TestDynamicFieldsHolder_GetFieldSliceStr(t *testing.T) {
	t.Parallel()

	holder := NewDynamicFieldsHolder()
	holder.SetField("str_slice", []string{"a", "b", "c"})
	holder.SetField("int_slice", []interface{}{"x", "y", "z"})
	holder.SetField("not_slice", "string")

	tests := []struct {
		name      string
		key       string
		want      []string
		wantNil   bool
		wantPanic bool
	}{
		{"获取[]string", "str_slice", []string{"a", "b", "c"}, false, false},
		{"获取[]interface{}字符串", "int_slice", []string{"x", "y", "z"}, false, false},
		{"获取非slice", "not_slice", nil, true, false},
		{"获取不存在的key", "missing", nil, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("GetFieldSliceStr(%s) should panic", tt.key)
					}
				}()
			}

			result := holder.GetFieldSliceStr(tt.key)
			if tt.wantNil && result != nil {
				t.Errorf("GetFieldSliceStr(%s) = %v, want nil", tt.key, result)
			}

			if !tt.wantNil && len(result) != len(tt.want) {
				t.Errorf("GetFieldSliceStr(%s) = %v, want %v", tt.key, result, tt.want)
			}

			if !tt.wantNil {
				for i, v := range result {
					if v != tt.want[i] {
						t.Errorf("GetFieldSliceStr(%s)[%d] = %v, want %v", tt.key, i, v, tt.want[i])
					}
				}
			}
		})
	}

	t.Run("panic测试", func(t *testing.T) {
		t.Parallel()
		holder.SetField("mixed", []interface{}{1, 2, 3})

		defer func() {
			if r := recover(); r == nil {
				t.Error("GetFieldSliceStr should panic for non-string []interface{}")
			}
		}()
		holder.GetFieldSliceStr("mixed")
	})

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()

		holder := DynamicFieldsHolder{}

		result := holder.GetFieldSliceStr("key")
		if result != nil {
			t.Error("GetFieldSliceStr on nil map should return nil")
		}
	})
}

func TestDynamicFieldsHolder_AddDynamicFieldsToMap(t *testing.T) {
	t.Parallel()

	holder := NewDynamicFieldsHolder()
	holder.SetField("key1", "value1")
	holder.SetField("key2", 42)

	t.Run("添加到空map", func(t *testing.T) {
		t.Parallel()

		m := make(map[string]interface{})
		holder.AddDynamicFieldsToMap(m)

		if len(m) != 2 {
			t.Errorf("map length = %d, want 2", len(m))
		}

		if m["key1"] != "value1" {
			t.Errorf("map[key1] = %v, want value1", m["key1"])
		}

		if m["key2"] != 42 {
			t.Errorf("map[key2] = %v, want 42", m["key2"])
		}
	})

	t.Run("添加到非空map", func(t *testing.T) {
		t.Parallel()

		m := map[string]interface{}{"existing": "value"}
		holder.AddDynamicFieldsToMap(m)

		if len(m) != 3 {
			t.Errorf("map length = %d, want 3", len(m))
		}

		if m["existing"] != "value" {
			t.Error("existing key should be preserved")
		}
	})

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()

		holder := DynamicFieldsHolder{}
		m := map[string]interface{}{"key": "value"}
		holder.AddDynamicFieldsToMap(m)

		if len(m) != 1 {
			t.Error("map should be unchanged when holder has nil DynamicFields")
		}
	})

	t.Run("nil input map", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Error("AddDynamicFieldsToMap should panic for nil map")
			}
		}()
		holder.AddDynamicFieldsToMap(nil)
	})
}

func TestDynamicFieldsHolder_OverwriteField(t *testing.T) {
	t.Parallel()

	holder := NewDynamicFieldsHolder()
	holder.SetField("key", "original")
	holder.SetField("key", "updated")

	val, ok := holder.GetField("key")
	if !ok {
		t.Error("GetField should return ok=true")
	}

	if val != "updated" {
		t.Errorf("key = %v, want updated", val)
	}
}
