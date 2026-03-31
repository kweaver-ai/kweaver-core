package mcpinstance

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// InstancePool 负责 MCP 实例的运行态管理：
// - 动态按需创建：实例缺失时回源 DB 解析配置并创建
// - 并发单飞：同一 (mcpID, version) 并发只创建一次
// - 有界内存：支持最大实例数 (MaxInstances) 的 LRU 淘汰
// - 过期清理：支持按最近访问时间的 TTL 清理，提供定时清理循环
// - 活跃保护：有活跃连接 (SSE/Stream) 的实例不参与淘汰/清理
//
// 该池“不缓存配置”，仅管理运行态实例，配置解析由 resolver 负责。
var ErrMCPInstanceConfigNotFound = errors.New("mcp instance runtime config not found")

type instanceBuilder interface {
	Build(ctx context.Context, cfg *interfaces.MCPRuntimeConfig) (*interfaces.MCPServerInstance, error)
	Shutdown(ctx context.Context, instance *interfaces.MCPServerInstance) error
}

type createCall struct {
	done     chan struct{}
	instance *interfaces.MCPServerInstance
	err      error
}

// InstancePoolOptions 池行为配置
type InstancePoolOptions struct {
	MaxInstances    int           // 最大保留实例数量 (<=0 表示不限制)
	InstanceTTL     time.Duration // 最近访问超时阈值 (<=0 表示不启用 TTL 清理)
	CleanupInterval time.Duration // 定时清理周期 (<=0 表示不启用定时清理)
}

// instanceEntry LRU 节点，记录实例与最近访问时间
type instanceEntry struct {
	key        string
	instance   *interfaces.MCPServerInstance
	lastAccess time.Time
	element    *list.Element
}

// InstancePool 实例池
type InstancePool struct {
	logger           interfaces.Logger
	dbResourceDeploy model.DBResourceDeploy
	builder          instanceBuilder
	opts             InstancePoolOptions
	now              func() time.Time
	mu               sync.Mutex
	entries          map[string]*instanceEntry
	lru              *list.List
	inflight         map[string]*createCall
	stopCleanup      chan struct{}
}

var (
	pOnce sync.Once
	pool  *InstancePool
)

// initInstancePool 初始化实例池
func initInstancePool(executor interfaces.IMCPToolExecutor) *InstancePool {
	pOnce.Do(func() {
		conf := config.NewConfigLoader()
		opts := InstancePoolOptions{
			MaxInstances:    conf.MCPConfig.MaxInstances,
			InstanceTTL:     time.Duration(conf.MCPConfig.InstanceTTL) * time.Second,
			CleanupInterval: time.Duration(conf.MCPConfig.CleanupInterval) * time.Second,
		}
		// 归一化非法配置值
		if opts.MaxInstances < 0 {
			opts.MaxInstances = 0
		}
		if opts.InstanceTTL < 0 {
			opts.InstanceTTL = 0
		}
		if opts.CleanupInterval < 0 {
			opts.CleanupInterval = 0
		}

		pool = &InstancePool{
			logger:           conf.GetLogger(),
			builder:          newInstanceManager(executor, conf.GetLogger()),
			dbResourceDeploy: dbaccess.NewResourceDeployDBSingleton(),
			opts:             opts,
			now:              time.Now,
			entries:          make(map[string]*instanceEntry),
			lru:              list.New(),
			inflight:         make(map[string]*createCall),
			stopCleanup:      make(chan struct{}),
		}
		if pool.opts.CleanupInterval > 0 && pool.opts.InstanceTTL > 0 {
			go pool.startCleanupLoop()
		}
	})
	return pool
}

// GetOrCreate 如果内存没有实例，则通过 resolver 解析配置并创建
func (p *InstancePool) GetOrCreate(ctx context.Context, mcpID string, version int) (*interfaces.MCPServerInstance, error) {
	key := p.key(mcpID, version)

	p.mu.Lock()
	if e, ok := p.entries[key]; ok && e != nil && e.instance != nil {
		p.touchLocked(e)
		p.mu.Unlock()
		return e.instance, nil
	}
	if call, ok := p.inflight[key]; ok {
		p.mu.Unlock()
		<-call.done
		if call.err == nil && call.instance != nil {
			p.mu.Lock()
			if e, ok := p.entries[key]; ok && e != nil {
				p.touchLocked(e)
			}
			p.mu.Unlock()
		}
		return call.instance, call.err
	}
	p.mu.Unlock()

	loaded, err := p.resolve(ctx, mcpID, version)
	if err != nil {
		return nil, err
	}
	if loaded == nil {
		return nil, ErrMCPInstanceConfigNotFound
	}
	return p.getOrCreateFromConfig(ctx, loaded)
}

// GetOrCreateWithConfig 使用外部提供的配置创建实例（避免额外 DB 查询）
func (p *InstancePool) GetOrCreateWithConfig(ctx context.Context, cfg *interfaces.MCPRuntimeConfig) (*interfaces.MCPServerInstance, error) {
	if cfg == nil {
		return nil, ErrMCPInstanceConfigNotFound
	}
	return p.getOrCreateFromConfig(ctx, cfg)
}

func (p *InstancePool) getOrCreateFromConfig(ctx context.Context, cfg *interfaces.MCPRuntimeConfig) (*interfaces.MCPServerInstance, error) {
	key := p.key(cfg.MCPID, cfg.Version)

	p.mu.Lock()
	if e, ok := p.entries[key]; ok && e != nil && e.instance != nil {
		p.touchLocked(e)
		p.mu.Unlock()
		return e.instance, nil
	}
	if call, ok := p.inflight[key]; ok {
		p.mu.Unlock()
		<-call.done
		if call.err == nil && call.instance != nil {
			p.mu.Lock()
			if e, ok := p.entries[key]; ok && e != nil {
				p.touchLocked(e)
			}
			p.mu.Unlock()
		}
		return call.instance, call.err
	}
	call := &createCall{done: make(chan struct{})}
	p.inflight[key] = call
	p.mu.Unlock()

	ins, err := p.builder.Build(ctx, cfg)

	var evicted []*interfaces.MCPServerInstance
	p.mu.Lock()
	call.instance = ins
	call.err = err
	if err == nil && ins != nil {
		e := &instanceEntry{
			key:        key,
			instance:   ins,
			lastAccess: p.now(),
		}
		e.element = p.lru.PushFront(e)
		p.entries[key] = e
		evicted = p.evictLocked()
	}
	delete(p.inflight, key)
	close(call.done)
	p.mu.Unlock()

	for _, victim := range evicted {
		_ = p.builder.Shutdown(ctx, victim)
	}
	return ins, err
}

// DeleteInstance 主动删除指定实例，并调用生命周期卸载
func (p *InstancePool) DeleteInstance(ctx context.Context, mcpID string, version int) error {
	key := p.key(mcpID, version)
	p.mu.Lock()
	e, ok := p.entries[key]
	if !ok || e == nil || e.instance == nil {
		p.mu.Unlock()
		return nil
	}
	delete(p.entries, key)
	if e.element != nil {
		p.lru.Remove(e.element)
	}
	ins := e.instance
	p.mu.Unlock()
	return p.builder.Shutdown(ctx, ins)
}

// Close 关闭定时清理循环
func Close() {
	if pool == nil {
		return
	}
	pool.mu.Lock()
	ch := pool.stopCleanup
	pool.stopCleanup = nil
	pool.mu.Unlock()
	if ch != nil {
		close(ch)
	}
}

// Cleanup 执行 TTL 清理；跳过有活跃连接的实例
func (p *InstancePool) cleanup(ctx context.Context) {
	var victims []*interfaces.MCPServerInstance
	now := p.now()

	p.mu.Lock()
	if p.opts.InstanceTTL <= 0 {
		p.mu.Unlock()
		return
	}
	for el := p.lru.Back(); el != nil; {
		prev := el.Prev()
		e, _ := el.Value.(*instanceEntry)
		if e == nil || e.instance == nil {
			p.lru.Remove(el)
			el = prev
			continue
		}
		if atomic.LoadInt64(&e.instance.ActiveStreamConn) > 0 || atomic.LoadInt64(&e.instance.ActiveSSEConn) > 0 {
			el = prev
			continue
		}
		if now.Sub(e.lastAccess) <= p.opts.InstanceTTL {
			break
		}
		delete(p.entries, e.key)
		p.lru.Remove(el)
		victims = append(victims, e.instance)
		el = prev
	}
	p.mu.Unlock()

	for _, ins := range victims {
		_ = p.builder.Shutdown(ctx, ins)
	}
}

// startCleanupLoop 定时触发 Cleanup
func (p *InstancePool) startCleanupLoop() {
	ticker := time.NewTicker(p.opts.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.cleanup(context.Background())
		case <-p.stopCleanup:
			return
		}
	}
}

// evictLocked 执行 LRU 淘汰；跳过有活跃连接的实例
func (p *InstancePool) evictLocked() []*interfaces.MCPServerInstance {
	if p.opts.MaxInstances <= 0 {
		return nil
	}
	var victims []*interfaces.MCPServerInstance
	for len(p.entries) > p.opts.MaxInstances {
		var removed bool
		for el := p.lru.Back(); el != nil; el = el.Prev() {
			e, _ := el.Value.(*instanceEntry)
			if e == nil || e.instance == nil {
				p.lru.Remove(el)
				continue
			}
			if atomic.LoadInt64(&e.instance.ActiveStreamConn) > 0 || atomic.LoadInt64(&e.instance.ActiveSSEConn) > 0 {
				continue
			}
			p.lru.Remove(el)
			delete(p.entries, e.key)
			victims = append(victims, e.instance)
			removed = true
			break
		}
		if !removed {
			break
		}
	}
	return victims
}

// touchLocked 更新最近访问时间并移动到 LRU 队头
func (p *InstancePool) touchLocked(e *instanceEntry) {
	if e == nil {
		return
	}
	e.lastAccess = p.now()
	if e.element != nil {
		p.lru.MoveToFront(e.element)
	}
}

func (p *InstancePool) key(mcpID string, version int) string {
	return fmt.Sprintf("%s-%d", mcpID, version)
}

func (p *InstancePool) resolve(ctx context.Context, mcpID string, version int) (*interfaces.MCPRuntimeConfig, error) {
	list, err := p.dbResourceDeploy.SelectList(ctx, nil, &model.ResourceDeployDB{
		ResourceID: mcpID,
		Type:       interfaces.ResourceDeployTypeMCP.String(),
		Version:    version,
	})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return utils.JSONToObjectWithError[*interfaces.MCPRuntimeConfig](list[0].Config)
}
