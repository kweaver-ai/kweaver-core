package pools

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
)

var Pools = PoolRegister{
	mu:    sync.RWMutex{},
	pools: map[string]*PoolManager{},
}

type PoolRegister struct {
	mu    sync.RWMutex
	pools map[string]*PoolManager
}

func (p *PoolRegister) Get(key string) *PoolManager {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pools[key]
}

func (p *PoolRegister) Register(key string, pool *PoolManager) {
	if p == nil {
		return
	}

	var old *PoolManager

	p.mu.Lock()
	old = p.pools[key]
	p.pools[key] = pool
	p.mu.Unlock()

	// 如果旧值存在且不是同一个实例，再关闭
	if old != nil && old != pool && !old.IsClosed() {
		old.Shutdown(false)
	}
}

// ShutdownAll
func (p *PoolRegister) ShutdownAll() {
	if p == nil {
		return
	}

	var olds []*PoolManager

	p.mu.Lock()
	for k, v := range p.pools {
		olds = append(olds, v)
		delete(p.pools, k)
	}
	p.mu.Unlock()

	for _, v := range olds {
		if v != nil && !v.IsClosed() {
			v.Shutdown(false)
		}
	}
}

// PoolType 池类型
type PoolType int

const (
	// NormalPool 普通池（传递无参函数）
	NormalPool PoolType = iota
	// FuncPool 函数池（传递带参数的函数）
	FuncPool
)

// PoolConfig 池配置
type PoolConfig struct {
	// 池名称
	Name string
	// 池类型
	Type PoolType
	// 池容量（最大goroutine数量）
	Capacity int
	// goroutine空闲时间，超过此时间则回收
	ExpiryDuration time.Duration
	// 是否预分配goroutine队列的内存
	PreAlloc bool
	// 最大阻塞任务数，0为不限制
	MaxBlockingTasks int
	// 是否为非阻塞模式，默认阻塞
	Nonblocking bool
	// panic处理器
	PanicHandler func(interface{})
	// 任务处理函数（仅FuncPool类型需要）
	TaskHandler func(interface{})
}

// PoolOption 定义配置选项函数类型
type PoolOption func(*PoolConfig)

// WithName 设置池名称
func WithName(name string) PoolOption {
	return func(c *PoolConfig) {
		c.Name = name
	}
}

// WithType 设置池类型
func WithType(poolType PoolType) PoolOption {
	return func(c *PoolConfig) {
		c.Type = poolType
	}
}

// WithCapacity 设置池容量
func WithCapacity(capacity int) PoolOption {
	return func(c *PoolConfig) {
		c.Capacity = capacity
	}
}

// WithExpiryDuration 设置goroutine空闲时间
func WithExpiryDuration(duration time.Duration) PoolOption {
	return func(c *PoolConfig) {
		c.ExpiryDuration = duration
	}
}

// WithPreAlloc 设置是否预分配内存
func WithPreAlloc(preAlloc bool) PoolOption {
	return func(c *PoolConfig) {
		c.PreAlloc = preAlloc
	}
}

// WithMaxBlockingTasks 设置最大阻塞任务数
func WithMaxBlockingTasks(max int) PoolOption {
	return func(c *PoolConfig) {
		c.MaxBlockingTasks = max
	}
}

// WithNonblocking 设置是否非阻塞模式
func WithNonblocking(nonblocking bool) PoolOption {
	return func(c *PoolConfig) {
		c.Nonblocking = nonblocking
	}
}

// WithPanicHandler 设置panic处理器
func WithPanicHandler(handler func(interface{})) PoolOption {
	return func(c *PoolConfig) {
		c.PanicHandler = handler
	}
}

// WithTaskHandler 设置任务处理函数
func WithTaskHandler(handler func(interface{})) PoolOption {
	return func(c *PoolConfig) {
		c.TaskHandler = handler
	}
}

// NewPoolConfig 创建新的池配置
func NewPoolConfig(opts ...PoolOption) *PoolConfig {
	config := &PoolConfig{
		// 设置默认值
		Name:             "default-pool",
		Type:             NormalPool,
		Capacity:         20,
		ExpiryDuration:   30 * time.Second,
		PreAlloc:         false,
		MaxBlockingTasks: 0,
		Nonblocking:      false,
		PanicHandler: func(i interface{}) {
			fmt.Printf("Worker panic recovered: %v\n", i)
		},
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(config)
	}

	return config
}

// PoolManager 池管理器
type PoolManager struct {
	config    *PoolConfig
	pool      *ants.Pool
	poolFunc  *ants.PoolWithFunc
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning atomic.Bool
}

// NewPoolManager 创建池管理器
func NewPoolManager(config *PoolConfig) (*PoolManager, error) {
	if config == nil {
		config = NewPoolConfig()
	}

	// 参数验证
	if config.Capacity <= 0 {
		config.Capacity = 20
	}

	if config.Type == FuncPool && config.TaskHandler == nil {
		return nil, errors.New("task handler is required for FuncPool type")
	}

	ctx, cancel := context.WithCancel(context.Background())

	pm := &PoolManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// 创建 ants 选项
	options := pm.buildAntsOptions()

	// 根据类型创建不同的池
	var err error
	switch config.Type {
	case NormalPool:
		pm.pool, err = ants.NewPool(config.Capacity, options...)
	case FuncPool:
		pm.poolFunc, err = ants.NewPoolWithFunc(config.Capacity, pm.wrapTaskHandler(), options...)
	default:
		return nil, fmt.Errorf("unsupported pool type: %d", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	pm.isRunning.Store(true)

	Pools.Register(config.Name, pm)

	return pm, nil
}

// buildAntsOptions 构建 ants 选项
func (pm *PoolManager) buildAntsOptions() []ants.Option {
	var options []ants.Option

	options = append(options, ants.WithExpiryDuration(pm.config.ExpiryDuration))
	options = append(options, ants.WithPreAlloc(pm.config.PreAlloc))
	options = append(options, ants.WithMaxBlockingTasks(pm.config.MaxBlockingTasks))
	options = append(options, ants.WithNonblocking(pm.config.Nonblocking))
	options = append(options, ants.WithPanicHandler(pm.config.PanicHandler))

	return options
}

func (pm *PoolManager) wrapTaskHandler() func(interface{}) {
	return func(i interface{}) {
		defer func() {
			if r := recover(); r != nil {
				pm.config.PanicHandler(r)
			}
		}()

		pm.config.TaskHandler(i)
	}
}

// Submit 提交不带参数的任务
func (pm *PoolManager) Submit(task func()) error {
	if !pm.isRunning.Load() {
		return errors.New("pool is not running")
	}

	if pm.config.Type != NormalPool {
		return errors.New("task handler is required for FuncPool type")
	}

	wrappedTask := func() {
		defer func() {
			if r := recover(); r != nil {
				if pm.config.PanicHandler != nil {
					pm.config.PanicHandler(r)
				}
			}
		}()

		task()
	}

	return pm.pool.Submit(wrappedTask)
}

// Invoke 提交带参数的任务
func (pm *PoolManager) Invoke(args interface{}) error {
	if !pm.isRunning.Load() {
		return errors.New("pool is not running")
	}

	if pm.config.Type != FuncPool {
		return errors.New("task handler is required for FuncPool type")
	}

	return pm.poolFunc.Invoke(args)
}

// WaitAll 等待所有任务完成
func (pm *PoolManager) WaitAll() {
	for {
		if pm.pool.Running() == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// IsClosed 检查池是否已关闭
func (pm *PoolManager) IsClosed() bool {
	return !pm.isRunning.Load()
}

// Shutdown 关闭池管理器
func (pm *PoolManager) Shutdown(graceful bool) {
	if !pm.isRunning.Load() {
		return
	}

	pm.isRunning.Store(false)
	pm.cancel()

	if graceful {
		pm.WaitAll()
	}

	switch pm.config.Type {
	case NormalPool:
		pm.pool.Release()
	case FuncPool:
		pm.poolFunc.Release()
	}
}
