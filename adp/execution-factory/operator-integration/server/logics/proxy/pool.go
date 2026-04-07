package proxy

import (
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

const (
	// 清理间隔
	cleanupInterval = 1 * time.Minute
)

// 客户端键（用于区分不同客户端实例）
type clientKey struct {
	ExecutionMode interfaces.ExecutionMode `json:"execution_mode,omitempty"`
	StreamingMode interfaces.StreamingMode `json:"streaming_mode,omitempty"`
	Timeout       time.Duration            `json:"timeout,omitempty"`
}

// GetClientKey 获取客户端键
func GetClientKey(executionMode interfaces.ExecutionMode, streamingMode interfaces.StreamingMode, timeout time.Duration) clientKey {
	return clientKey{
		ExecutionMode: executionMode,
		StreamingMode: streamingMode,
		Timeout:       timeout,
	}
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxClients     int           // 最大客户端数量
	MaxTimeout     time.Duration // 最大超时时间
	DefaultTimeout time.Duration // 默认超时时间
	ClientLifetime time.Duration // 客户端生命周期
}

// ProxyClient 代理客户端信息
type ProxyClient struct {
	*http.Client
	IsStreaming   bool
	StreamingMode interfaces.StreamingMode
	CreateAt      time.Time
}

// clientPool 代理客户端池
// 客户端池结构体
type clientPool struct {
	logger      interfaces.Logger
	mu          sync.Mutex
	clients     map[clientKey]*ProxyClient
	config      PoolConfig
	stopCleanup chan struct{}
}

var (
	clientPoolInstance *clientPool
	clientPoolOnce     sync.Once
)

// NewClientPool 创建新的客户端池
func NewClientPool() *clientPool {
	clientPoolOnce.Do(func() {
		conf := config.NewConfigLoader()
		poolConfig := PoolConfig{
			MaxClients:     conf.ProxyModuleConfig.MaxClients,
			MaxTimeout:     time.Duration(conf.ProxyModuleConfig.MaxTimeout) * time.Second,
			DefaultTimeout: time.Duration(conf.ProxyModuleConfig.DefaultTimeout) * time.Second,
			ClientLifetime: time.Duration(conf.ProxyModuleConfig.ClientLifetime) * time.Second,
		}
		clientPoolInstance = &clientPool{
			mu:          sync.Mutex{},
			logger:      conf.GetLogger(),
			clients:     make(map[clientKey]*ProxyClient),
			config:      poolConfig,
			stopCleanup: make(chan struct{}),
		}
		// 启动定期清理 goroutine
		go clientPoolInstance.startCleanupTimer()
	})
	return clientPoolInstance
}

// GetClient 获取同步类型客户端
func (p *clientPool) GetClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = p.config.DefaultTimeout
	}
	if timeout > p.config.MaxTimeout {
		timeout = p.config.MaxTimeout
	}
	key := GetClientKey(interfaces.ExecutionModeSync, "", timeout)
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查客户端是否已存在
	if client, exists := p.clients[key]; exists {
		client.CreateAt = time.Now() // 更新访问时间
		return client.Client
	}

	// 如果达到最大客户端数量，移除最旧的客户端
	if len(p.clients) >= p.config.MaxClients {
		p.removeOldestClient()
	}

	// 创建新客户端
	client := &ProxyClient{
		Client: rest.NewRawHTTPClientWithOptions(rest.HTTPClientOptions{
			TimeOut: int(timeout.Seconds()),
		}),
		IsStreaming: false,
		CreateAt:    time.Now(),
	}

	p.clients[key] = client
	return client.Client
}

// getStreamClient 通用流式客户端创建方法
func (p *clientPool) GetStreamClient(streamingMode interfaces.StreamingMode, timeout time.Duration) *http.Client {
	p.logger.Debugf("get stream client, streamingMode: %v, timeout: %v", streamingMode, timeout)
	key := GetClientKey(interfaces.ExecutionModeStream, streamingMode, timeout)
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查客户端是否已存在
	if client, exists := p.clients[key]; exists {
		p.logger.Debugf("stream client exists, streamingMode: %v, timeout: %v", streamingMode, timeout)
		client.CreateAt = time.Now() // 更新访问时间
		return client.Client
	}
	p.logger.Debugf("stream client not exists, streamingMode: %v, timeout: %v", streamingMode, timeout)
	// 如果达到最大客户端数量，优先移除同步客户端
	if len(p.clients) >= p.config.MaxClients {
		p.logger.Debugf("stream client not exists, remove oldest client")
		p.removeOldestClient()
	}

	responseHeaderTimeout := p.config.DefaultTimeout
	if timeout > 0 {
		responseHeaderTimeout = timeout
	}
	// 创建新客户端
	client := &ProxyClient{
		Client: rest.NewRawHTTPClientWithOptions(rest.HTTPClientOptions{
			TimeOut:               int(timeout.Seconds()),
			ResponseHeaderTimeout: int(responseHeaderTimeout.Seconds()),
		}),
		IsStreaming:   true,
		StreamingMode: streamingMode,
		CreateAt:      time.Now(),
	}
	p.logger.Debugf("create stream client, streamingMode: %v, timeout: %v", streamingMode, timeout)

	p.clients[key] = client
	return client.Client
}

// / 移除最旧的客户端
func (p *clientPool) removeOldestClient() {
	var oldestKey *clientKey
	oldestTime := time.Time{}

	// 找到最旧的客户端
	first := true
	for key, client := range p.clients {
		if client.IsStreaming {
			continue
		}
		if first || oldestTime.IsZero() || client.CreateAt.Before(oldestTime) {
			oldestTime = client.CreateAt
			oldestKey = &key
			first = false
		}
	}
	// 如果全是流式客户端，则移除最旧的
	if oldestKey == nil {
		for key, client := range p.clients {
			if first || client.CreateAt.Before(oldestTime) {
				oldestKey = &key
				oldestTime = client.CreateAt
				first = false
			}
		}
	}

	if !oldestTime.IsZero() && oldestKey != nil {
		p.clients[*oldestKey].CloseIdleConnections()
		delete(p.clients, *oldestKey)
	}
}

// 定期清理闲置客户端
func (p *clientPool) startCleanupTimer() {
	cleanupTicker := time.NewTicker(cleanupInterval)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-cleanupTicker.C:
			p.cleanupIdleClients()
		case <-p.stopCleanup:
			return
		}
	}
}

// 清理闲置客户端
func (p *clientPool) cleanupIdleClients() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	for key, client := range p.clients {
		createdAt := client.CreateAt
		// 根据连接类型应用不同的超时策略
		if key.ExecutionMode == interfaces.ExecutionModeSync &&
			now.Sub(createdAt) > p.config.ClientLifetime {
			p.logger.Infof("cleanup idle sync client, key: %v, created at: %v", key, createdAt)
			client.CloseIdleConnections() // 关闭同步客户端的闲置连接
			delete(p.clients, key)
		}
	}
}

// Close 关闭连接池
func (p *clientPool) Close() {
	close(p.stopCleanup)
}
