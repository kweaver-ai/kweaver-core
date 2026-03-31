package httpproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Proxy 泛型代理接口定义
type Proxy[T any, R any] interface {
	Forward(request T) (response R, err error)
}

// JSONPostProxy 实现HTTP JSON POST代理（泛型版本）
type JSONPostProxy[T any, R any] struct {
	TargetURL string       // 目标服务URL
	Client    *http.Client // 可自定义的HTTP客户端
	Token     string       // 可选的Bearer Token
}

// NewJSONPostProxy 创建代理实例的构造函数
func NewJSONPostProxy[T any, R any](targetURL string) *JSONPostProxy[T, R] {
	return &JSONPostProxy[T, R]{
		TargetURL: targetURL,
		Client: &http.Client{
			Timeout: 10 * time.Second, // 默认10秒超时
		},
	}
}

// SetToken 设置Bearer Token
func (p *JSONPostProxy[T, R]) SetToken(token string) {
	p.Token = token
}

// SetClient 自定义HTTP客户端
func (p *JSONPostProxy[T, R]) SetClient(client *http.Client) {
	p.Client = client
}

// Forward 实现代理接口
func (p *JSONPostProxy[T, R]) Forward(req T) (R, error) {
	// 创建零值响应用于错误返回
	var zero R

	// 1. 序列化请求数据为JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return zero, fmt.Errorf("JSON编码失败: %w", err)
	}

	// 2. 创建HTTP请求
	httpReq, err := http.NewRequest(
		"POST",
		p.TargetURL,
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return zero, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置JSON内容类型
	httpReq.Header.Set("Content-Type", "application/json")

	// 3. 添加Authorization头（如果设置了Token）
	if p.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.Token)
	}

	// 4. 发送请求
	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return zero, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 5. 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 尝试读取错误响应体
		errorBody, _ := io.ReadAll(resp.Body)
		return zero, fmt.Errorf("服务返回错误: %d - %s", resp.StatusCode, string(errorBody))
	}

	// 6. 解析响应体
	var respData R
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return zero, fmt.Errorf("JSON解析失败: %w", err)
	}

	return respData, nil
}
