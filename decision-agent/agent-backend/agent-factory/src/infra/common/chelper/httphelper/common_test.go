package httphelper

import (
	"strings"
	"testing"
)

func TestCommonRespError_Error(t *testing.T) {
	t.Parallel()

	// 测试CommonRespError的Error方法
	tests := []struct {
		name string
		resp CommonRespError
		want string
	}{
		{
			name: "基本错误响应",
			resp: CommonRespError{
				Code:        400,
				Cause:       "invalid parameter",
				Message:     "参数错误",
				Description: "请求参数不正确",
				Solution:    "请检查参数格式",
				Detail:      DetailMap{"field": "name"},
			},
			want: `{"code":400,"cause":"invalid parameter","message":"参数错误","description":"请求参数不正确","solution":"请检查参数格式","detail":{"field":"name"}}`,
		},
		{
			name: "空错误响应",
			resp: CommonRespError{},
			want: `{"code":0,"cause":"","message":"","description":"","solution":"","detail":null}`,
		},
		{
			name: "带Debug信息的错误响应",
			resp: CommonRespError{
				Code:        500,
				Cause:       "internal error",
				Message:     "内部错误",
				Description: "服务器内部错误",
				Solution:    "请联系管理员",
				Detail:      DetailMap{},
				Debug:       "stack trace info",
			},
			want: `{"code":500,"cause":"internal error","message":"内部错误","description":"服务器内部错误","solution":"请联系管理员","detail":{},"debug":"stack trace info"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.resp.Error()
			if got != tt.want {
				t.Errorf("CommonRespError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildQueryParams(t *testing.T) {
	t.Parallel()

	// 创建httpClient实例用于测试
	client := &httpClient{}

	tests := []struct {
		name      string
		queryData interface{}
		want      string
	}{
		{
			name:      "nil值",
			queryData: nil,
			want:      "",
		},
		{
			name: "map类型",
			queryData: map[string]interface{}{
				"name":   "test",
				"age":    25,
				"active": true,
			},
			want: "active=true&age=25&name=test",
		},
		{
			name: "结构体类型",
			queryData: struct {
				Name   string `_query:"name"`
				Age    int    `_query:"age"`
				Active bool   `_query:"active"`
			}{
				Name:   "test",
				Age:    25,
				Active: true,
			},
			want: "active=true&age=25&name=test",
		},
		{
			name: "带_query标签的结构体",
			queryData: struct {
				Name     string `_query:"name,omitempty"`
				Age      int    `_query:"age,omitempty"`
				Required string `_query:"required"`
			}{
				Name:     "",
				Age:      0,
				Required: "value",
			},
			want: "age=0&name=&required=value",
		},
		// _query:"-" 跳过字段
		{
			name: "包含_query:\"-\"的结构体",
			queryData: struct {
				Name string `_query:"name"`
				Skip string `_query:"-"`
				Age  int    `_query:"age"`
			}{
				Name: "test",
				Skip: "should_be_ignored",
				Age:  25,
			},
			want: "age=25&name=test",
		},
		{
			name: "指针类型",
			queryData: &struct {
				Name string `_query:"name"`
			}{
				Name: "test",
			},
			want: "name=test",
		},
		{
			name:      "空指针",
			queryData: (*struct{})(nil),
			want:      "",
		},
		{
			name:      "空map",
			queryData: map[string]interface{}{},
			want:      "",
		},
		{
			name: "包含特殊字符的值",
			queryData: map[string]interface{}{
				"query":  "test value with spaces",
				"symbol": "a+b=c",
			},
			want: "query=test+value+with+spaces&symbol=a%2Bb%3Dc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := client.buildQueryParams(tt.queryData)
			// 注意：map的遍历顺序不确定，所以我们需要验证结果包含所有预期的键值对
			if tt.want == "" {
				if got != "" {
					t.Errorf("buildQueryParams() = %v, want %v", got, tt.want)
				}

				return
			}

			// 对于非空结果，我们检查是否包含所有预期的键值对，并且没有额外的键值对
			wantParams := parseQueryParams(tt.want)
			gotParams := parseQueryParams(got)

			// 首先检查数量是否相同
			if len(wantParams) != len(gotParams) {
				t.Errorf("buildQueryParams() returned %d parameters, want %d", len(gotParams), len(wantParams))
				t.Errorf("got params: %v", gotParams)
				t.Errorf("want params: %v", wantParams)

				return
			}

			// 然后检查每个键值对是否相同
			for key, wantValue := range wantParams {
				if gotValue, exists := gotParams[key]; !exists {
					t.Errorf("buildQueryParams() missing key %s", key)
				} else if gotValue != wantValue {
					t.Errorf("buildQueryParams() key %s = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

// 辅助函数：解析查询字符串
func parseQueryParams(query string) map[string]string {
	params := make(map[string]string)
	if query == "" {
		return params
	}

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		} else if len(parts) == 1 {
			params[parts[0]] = ""
		}
	}

	return params
}
