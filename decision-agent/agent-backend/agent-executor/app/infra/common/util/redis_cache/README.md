# Redis缓存工具

优先使用JSON序列化，失败时自动降级使用Pickle的Redis缓存工具。

## 快速开始

```python
from app.infra.common.util.redis_cache import RedisCache

# 创建缓存实例
cache = RedisCache(db=3)

# 设置缓存（自动JSON序列化）
await cache.set("user:123", {"name": "Alice", "age": 30}, ttl=3600)

# 获取缓存（自动反序列化）
user = await cache.get("user:123")
print(user)  # {"name": "Alice", "age": 30}

# 删除缓存
await cache.delete("user:123")
```

## 核心特性

- ✅ **智能序列化**：优先JSON，自动降级Pickle
- ✅ **类型安全**：完整的类型注解
- ✅ **异步高效**：基于async/await
- ✅ **功能完整**：TTL、分布式锁、Lua脚本
- ✅ **生产就绪**：异常处理、日志记录

## API概览

### 基础操作
- `await cache.set(key, value, ttl)` - 设置缓存
- `await cache.get(key)` - 获取缓存
- `await cache.delete(key)` - 删除缓存
- `await cache.exists(key)` - 检查存在

### TTL管理
- `await cache.expire(key, ttl)` - 设置过期时间
- `await cache.ttl(key)` - 获取剩余时间

### 高级操作
- `await cache.set_nx(key, value, ex)` - 分布式锁
- `await cache.eval(script, keys, args)` - Lua脚本

## 序列化支持

### JSON序列化（优先）
支持类型：dict, list, str, int, float, bool, datetime, Enum, Pydantic模型

### Pickle序列化（备选）
支持类型：几乎所有Python对象（包括DolphinAgent、TriditionalToolkit等）

## 完整文档

详见：[Redis缓存实现总结](../docs/redis_cache_implementation_summary.md)
