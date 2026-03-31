package common

import "reflect"

// ------------------ListCommon start----------------------
// ListCommon 带 entries 的列表
type ListCommon struct {
	Entries []interface{} `json:"entries"`
}

func NewListCommon() *ListCommon {
	return &ListCommon{}
}

func (l *ListCommon) SetEntries(entries interface{}) {
	// 处理 nil 参数
	if entries == nil {
		l.Entries = []interface{}{}
		return
	}

	// 通过反射实现
	v := reflect.ValueOf(entries)
	if v.Kind() != reflect.Slice {
		panic("entries must be a slice")
	}

	// 创建一个新的 []interface{} 切片
	length := v.Len()
	result := make([]interface{}, length)

	// 将每个元素复制到新的切片中
	for i := 0; i < length; i++ {
		result[i] = v.Index(i).Interface()
	}

	l.Entries = result
}

// ------------------ListCommon end----------------------

// ------------------ListCommonWithTotal start----------------------

// ListCommonWithTotal 带总记录数的列表
type ListCommonWithTotal struct {
	ListCommon
	Total int64 `json:"total"`
}

func NewListCommonWithTotal() *ListCommonWithTotal {
	return &ListCommonWithTotal{}
}

func (l *ListCommonWithTotal) SetEntries(entries interface{}) {
	l.ListCommon.SetEntries(entries)
}

func (l *ListCommonWithTotal) SetTotal(total int64) {
	l.Total = total
}

// ------------------ListCommonWithTotal end----------------------
