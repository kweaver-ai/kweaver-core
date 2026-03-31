package httphelper

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type DetailMap = map[string]interface{}

// CommonResp 通用响应
// 参考：https://confluence.aishu.cn/pages/viewpage.action?pageId=190114672
type CommonResp struct {
	Code        int       `json:"code"`        // 错误码（前三位：标准http错误码，中间三位为服务器特定码，后三位服务中自定义码）
	Cause       string    `json:"cause"`       // 错误原因，产生错误的具体原因
	Message     string    `json:"message"`     // 错误信息
	Description string    `json:"description"` // 符合国际化要求的错误描述
	Solution    string    `json:"solution"`    // 符合国际化要求的针对当前错误的操作提示
	Detail      DetailMap `json:"detail"`
	Debug       string    `json:"debug,omitempty"` // CAPP部分项目用到，其它项目可能没有这个
}

type CommonRespError CommonResp

func (r *CommonRespError) Error() string {
	bys, err := cutil.JSON().Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(bys)
}

const (
	RetryInterval = 5
)

// buildQueryParams 将查询参数转换为URL查询字符串
//
// 支持的输入类型：
// - nil: 返回空字符串
// - map[string]interface{}: 将map的键值对转换为URL参数
// - 结构体: 使用反射读取结构体字段，支持_query标签指定键名
// - 指针: 递归处理指向的实际值
//
// _query标签处理：
// - 使用_query标签指定URL参数名（如 `_query:"user_name"`）
// - 如果_query标签为"-"，则跳过该字段
// - 如果没有_query标签，则使用字段名作为参数名
// - 支持逗号分隔的选项，但只使用第一部分作为参数名
//
// 输出格式：
// - 使用url.Values进行URL编码
// - 特殊字符会自动转义（如空格→+，+→%2B等）
//
// 示例：
//
//	map[string]interface{}{"name": "test", "age": 25} → "name=test&age=25"
//	struct{Name string `_query:"user_name"`} → "user_name=value"
//	struct{Name string `_query:"name,omitempty"`} → "name=value"
func (c *httpClient) buildQueryParams(queryData interface{}) string {
	if queryData == nil {
		return ""
	}

	values := url.Values{}

	// 使用反射处理不同类型的查询参数
	v := reflect.ValueOf(queryData)

	switch v.Kind() {
	case reflect.Map:
		// 处理map类型：只支持字符串键的map
		if v.Type().Key().Kind() == reflect.String {
			for _, key := range v.MapKeys() {
				value := v.MapIndex(key)
				values.Set(key.String(), fmt.Sprintf("%v", value.Interface()))
			}
		}
	case reflect.Struct:
		// 处理结构体类型：遍历所有导出字段
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			// 跳过非导出字段（小写字段名）
			if field.PkgPath != "" {
				continue
			}

			// 获取_query标签作为键名，如果没有则使用字段名
			key := field.Name
			queryTag := field.Tag.Get("_query")

			// 如果_query标签为"-"，则跳过该字段
			if queryTag == "-" {
				continue
			}

			// 如果有_query标签且不为"-"，则使用标签的第一部分作为键名
			if queryTag != "" {
				// 处理_query标签，如:"name,omitempty" → "name"
				if tagParts := strings.Split(queryTag, ","); len(tagParts) > 0 {
					key = tagParts[0]
				}
			}

			// 将字段值转换为字符串并添加到URL参数中
			values.Set(key, fmt.Sprintf("%v", v.Field(i).Interface()))
		}
	case reflect.Ptr:
		// 处理指针类型：递归处理指向的实际值
		if !v.IsNil() {
			return c.buildQueryParams(v.Elem().Interface())
		}
	}

	// 使用url.Values.Encode()生成URL编码的查询字符串
	return values.Encode()
}
