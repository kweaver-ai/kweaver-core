package policy

import (
	"context"
	"sync"
	"time"
)

// Result 泛型执行结果（最简版本）
type Result[T any] struct {
	Name string `json:"name"`
	Data T      `json:"data"`
}

// NewResult 创建结果
func NewResult[T any](name string, data T) *Result[T] {
	return &Result[T]{
		Name: name,
		Data: data,
	}
}

// ResultWrapper 获取结果信息接口
type ResultWrapper interface {
	GetName() string
}

func (r *Result[T]) GetName() string { return r.Name }

// RetryData 重试策略过程性数据
type RetryData struct {
	Attempts int `json:"attempts"`
	Max      int `json:"max"`
}

// ResultCollector 策略执行结果过程性数据收集器
type ResultCollector struct {
	Duration int64
	results  map[string]ResultWrapper
	mu       sync.RWMutex
}

// NewResultCollector 实例化
func NewResultCollector() *ResultCollector {
	return &ResultCollector{
		results: make(map[string]ResultWrapper),
	}
}

// Add 添加结果过程性数据收集器
func (c *ResultCollector) Add(result ResultWrapper) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results[result.GetName()] = result
}

// Reset 重置
func (c *ResultCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = make(map[string]ResultWrapper)
}

// GetAs 类型安全地获取结果
func GetAs[T any](c *ResultCollector, name string) (*Result[T], bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if wrapper, ok := c.results[name]; ok {
		if result, ok := wrapper.(*Result[T]); ok {
			return result, true
		}
	}
	return nil, false
}

// ExecutionInfo 策略执行信息
type ExecutionInfo struct {
	PolicyName string
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Error      error
	RetryCount int
	TimedOut   bool
	CustomData map[string]interface{}
}

type Policy interface {
	Name() string
	Do(ctx context.Context, fn func(context.Context) error) error
}

type ResultAwarePolicy interface {
	Policy
	Init()
	SetCollector(collector *ResultCollector)
}

type CompositePolicy struct {
	policies      []Policy
	collector     *ResultCollector
	failFast      bool
	onError       func(context.Context, error)            // 错误回调
	beforeExecute func(context.Context)                   // 执行前回调
	afterExecute  func(context.Context, *ResultCollector) // 执行后回调
}

// Option 配置选项函数类型
type Option func(*CompositePolicy)

func NewComposite(opts ...Option) *CompositePolicy {
	cp := &CompositePolicy{
		policies:  make([]Policy, 0),
		collector: NewResultCollector(),
		failFast:  true,
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(cp)
	}

	return cp
}

// WithRetry 添加重试策略
func WithRetry(max, delay int, fn func(error) bool) Option {
	return func(cp *CompositePolicy) {
		cp.policies = append(cp.policies, &RetryPolicy{
			Max:     max,
			Delay:   delay,
			RetryIf: fn,
		})
	}
}

// WithTimeout 添加超时策略
func WithTimeout(duration int) Option {
	return func(cp *CompositePolicy) {
		cp.policies = append(cp.policies, &TimeoutPolicy{
			Delay: duration,
		})
	}
}

// WithPolicies 批量添加策略
func WithPolicies(policies ...Policy) Option {
	return func(cp *CompositePolicy) {
		cp.policies = append(cp.policies, policies...)
	}
}

// WithFailFast 设置快速失败模式
func WithFailFast(enabled bool) Option {
	return func(cp *CompositePolicy) {
		cp.failFast = enabled
	}
}

// WithOnError 设置错误回调
func WithOnError(fn func(context.Context, error)) Option {
	return func(cp *CompositePolicy) {
		cp.onError = fn
	}
}

// WithBeforeExecute 设置执行前回调
func WithBeforeExecute(fn func(ctx context.Context)) Option {
	return func(cp *CompositePolicy) {
		cp.beforeExecute = fn
	}
}

// WithAfterExecute 设置执行后回调
func WithAfterExecute(fn func(context.Context, *ResultCollector)) Option {
	return func(cp *CompositePolicy) {
		cp.afterExecute = fn
	}
}

// Do 执行组合策略
func (c *CompositePolicy) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	c.collector.Reset()

	// 依赖注入, 哪些策略需要过程性输出数据,对应实现如下函数
	// func Init()
	// func SetCollector(collector *ResultCollector)
	for _, p := range c.policies {
		if rp, ok := p.(ResultAwarePolicy); ok {
			rp.SetCollector(c.collector)
			rp.Init()
		}
	}

	// 执行前回调
	if c.beforeExecute != nil {
		c.beforeExecute(ctx)
	}

	// 构建策略链
	wrapped := fn
	for i := len(c.policies) - 1; i >= 0; i-- {
		p := c.policies[i]
		prev := wrapped
		// 使用闭包捕获当前策略
		wrapped = func(policy Policy, prevFn func(context.Context) error) func(context.Context) error {
			return func(ctx context.Context) error {
				return policy.Do(ctx, prevFn)
			}
		}(p, prev)
	}

	// 执行
	now := time.Now()
	err := wrapped(ctx)
	c.collector.Duration = time.Since(now).Milliseconds()
	// c.collector.Duration = math.Round(time.Since(now).Seconds()*100) / 100

	// 错误回调
	if err != nil && c.onError != nil {
		c.onError(ctx, err)
	}

	// 执行后回调
	if c.afterExecute != nil {
		c.afterExecute(ctx, c.collector)
	}

	return err
}
