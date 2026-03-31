package mocks

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	url "net/url"
	"strings"

	gin "github.com/gin-gonic/gin"
)

func MockPostRequest(url string, headers map[string]string, body io.Reader, handler func(c *gin.Context)) (recorder *httptest.ResponseRecorder) {
	// 创建一个带有中间件的路由组
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use()
	router.Handle(http.MethodPost, url, func(c *gin.Context) {
		handler(c)
		c.Next()
	})
	// 创建请求并触发中间件
	recorder = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, url, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	router.ServeHTTP(recorder, req)
	return recorder
}

func MockGetRequest(path string, headers map[string]string, query map[string]interface{}, pathParams []string, handler func(c *gin.Context)) (recorder *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use()

	router.Handle(http.MethodGet, path, func(c *gin.Context) {
		// 设置路径参数
		for i, param := range pathParams {
			paramName := strings.Split(path, "/")[i+1][1:] // 提取 :param 格式的参数名
			c.Params = append(c.Params, gin.Param{Key: paramName, Value: param})
		}
		handler(c)
		c.Next()
	})
	formattedPath := path
	for _, param := range pathParams {
		// 找到第一个占位符的位置（如 :id）
		start := strings.Index(formattedPath, ":")
		if start == -1 {
			break
		}
		end := strings.Index(formattedPath[start:], "/")
		if end == -1 {
			end = len(formattedPath)
		} else {
			end += start
		}
		// 替换占位符为实际参数
		formattedPath = formattedPath[:start] + param + formattedPath[end:]
	}
	// 构造请求路径（移除了错误的路径拼接）
	queryString := url.Values{}
	for key, value := range query {
		queryString.Add(key, fmt.Sprintf("%v", value))
	}

	if len(queryString) > 0 {
		formattedPath += "?" + queryString.Encode()
	}

	req := httptest.NewRequest(http.MethodGet, formattedPath, http.NoBody)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
