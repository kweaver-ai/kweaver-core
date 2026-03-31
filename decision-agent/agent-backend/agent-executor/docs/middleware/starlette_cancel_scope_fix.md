# Starlette Cancel Scope 异常修复

## 问题描述

在运行时遇到以下异常：

```
2025-10-27 21:30:03 - ERROR - app/router/exception_handler.py:226
{
  'message': 'RuntimeError("Attempted to exit a cancel scope that isn\'t the current tasks\'s current cancel scope")',
  'caller': '/Users/Zhuanz/Work/as/dip_ws/agent-executor/app/router/exception_handler.py:223',
  'stack': 'Traceback (most recent call last):
    File "/Users/Zhuanz/Work/as/dip_ws/agent-executor/.venv/lib/python3.10/site-packages/starlette/middleware/errors.py", line 164, in __call__
      await self.app(scope, receive, _send)
    File "/Users/Zhuanz/Work/as/dip_ws/agent-executor/.venv/lib/python3.10/site-packages/starlette/middleware/base.py", line 183, in __call__
      async with anyio.create_task_group() as task_group:
    File "/Users/Zhuanz/Work/as/dip_ws/agent-executor/.venv/lib/python3.10/site-packages/anyio/_backends/_asyncio.py", line 792, in __aexit__
      return self.cancel_scope.__exit__(exc_type, exc_val, exc_tb)
    File "/Users/Zhuanz/Work/as/dip_ws/agent-executor/.venv/lib/python3.10/site-packages/anyio/_backends/_asyncio.py", line 467, in __exit__
      raise RuntimeError(
  RuntimeError: Attempted to exit a cancel scope that isn\'t the current tasks\'s current cancel scope',
  'time': '2025-10-27 21:30:03'
}
```

## 根本原因

### 问题触发链路

1. **Starlette 0.48.0 的 BaseHTTPMiddleware 实现**
   - 使用 `anyio.create_task_group()` 管理异步任务
   - 依赖 cancel scope 来协调 `receive` 回调和响应发送

2. **日志中间件的错误实现**
   - 旧代码通过 `Request(request.scope, receive=receive)` 创建新的 Request 实例
   - 新实例有独立的 `receive` 回调，破坏了 BaseHTTPMiddleware 的 cancel scope 管理

3. **异常触发**
   - 当中间件尝试退出 task group 时，发现当前的 cancel scope 不是预期的
   - anyio 抛出 `RuntimeError`

### 核心问题代码

```python
# ❌ 错误的实现
@app.middleware("http")
async def log_requests(request: Request, call_next):
    # 读取请求体
    body_bytes = await request.body()
    request_body = body_bytes.decode("utf-8")

    # 问题：创建新的 Request 对象并替换原始对象
    async def receive():
        return {"type": "http.request", "body": body_bytes}

    request = Request(request.scope, receive=receive)  # ❌ 破坏了 cancel scope

    response = await call_next(request)  # 传递了新的 request
    return response
```

## 解决方案

### 修复原理

利用 Starlette 内置的请求体缓存机制，避免替换 Request 对象：

- `request.body()` 第一次调用时会读取并缓存到 `request._body`
- 后续任何地方调用 `request.body()` 都从缓存返回
- 不需要手动管理 `receive` 回调

### 修复代码

```python
# ✅ 正确的实现
@app.middleware("http")
async def log_requests(request: Request, call_next):
    # 读取请求体（Starlette 会自动缓存，后续 handler 仍可读取）
    body_bytes = await request.body()
    request_body = body_bytes.decode("utf-8") if body_bytes else ""

    request_info["body"] = request_body

    # 直接使用原始 request 对象，不替换
    response = await call_next(request)  # ✅ 保持 cancel scope 完整
    return response
```

### 修改对比

| 方面 | 修改前 | 修改后 |
|------|--------|--------|
| **Request 对象** | 创建新实例并替换 | 保持原始对象 |
| **receive 回调** | 自定义 receive 函数 | 使用原始 receive |
| **请求体缓存** | 依赖自定义缓存 | 依赖 Starlette 内置缓存 |
| **空值处理** | 直接 decode | `if body_bytes else ""` |
| **cancel scope** | 被破坏 | 保持完整 |

## 版本兼容性

### ✅ 修复方案在新旧版本中都适用

| 版本范围 | 兼容性 | 说明 |
|---------|--------|------|
| **0.12.x - 0.47.x** | ✅ 完全兼容 | `request.body()` 缓存机制已存在，不替换 Request 对象是最佳实践 |
| **0.48.0+** | ✅ 必须使用 | 新版 `BaseHTTPMiddleware` 使用严格的 anyio 管理，替换 Request 会触发异常 |

### 关键点

1. **不要替换 Request 对象** - 这是核心修复
2. **直接调用 `request.body()`** - Starlette 会自动缓存
3. **后续 handler 可以继续读取** - 从缓存中获取，不会重复消费流

## 测试验证

### 测试文件位置

`/Users/Zhuanz/Work/as/dip_ws/agent-executor/test/router_test/test_middleware_compatibility.py`

### 测试覆盖场景

1. **中间件读取请求体后，handler 仍能正常读取**
   ```python
   def test_middleware_body_reading():
       # 验证 Starlette 缓存机制正常工作
   ```

2. **中间件与流式响应的兼容性**
   ```python
   def test_middleware_with_streaming_response():
       # 验证不会影响流式输出
   ```

3. **GET 请求（无请求体）的兼容性**
   ```python
   def test_middleware_without_body():
       # 验证空请求体处理正确
   ```

### 运行测试

```bash
cd /Users/Zhuanz/Work/as/dip_ws/agent-executor
python -m pytest test/router_test/test_middleware_compatibility.py -v
```

### 测试结果

```
=================== 3 passed in 0.19s ====================
```

- **通过率**: 100% (3/3)
- **执行时间**: 0.19 秒
- **平台**: macOS, Python 3.10.5

## 影响范围

### 修改文件

- `app/router/__init__.py` - 日志中间件 `log_requests` 函数

### 功能影响

- ✅ 请求处理链保持原样
- ✅ Starlette 负责缓存请求体并允许后续 handler 继续读取
- ✅ 消除了 cancel scope 异常风险
- ✅ 中间件逻辑仍能记录完整请求体
- ✅ 支持流式响应、普通响应、空请求体等多种场景

### 性能影响

- 无负面影响
- 减少了不必要的 Request 对象创建
- 利用 Starlette 内置缓存，更高效

## 相关资源

### Starlette 文档

- [Request Body](https://www.starlette.io/requests/#body)
- [BaseHTTPMiddleware](https://www.starlette.io/middleware/#basehttpmiddleware)

### 相关 Issue

- [Starlette #847 - Middleware Request parse Hangs forever](https://github.com/encode/starlette/issues/847)
- [Starlette #1702 - Cache Request in scope](https://github.com/Kludex/starlette/issues/1702)

## 后续建议

1. **日志大小限制**
   - 若需记录大型请求体，考虑加入大小限制或采样
   - 避免日志过大影响性能

2. **依赖升级注意事项**
   - 升级 Starlette/FastAPI 时关注 Breaking Changes
   - 特别注意中间件相关的变更

3. **监控告警**
   - 添加对此类异常的监控
   - 及时发现类似的 cancel scope 问题

## 修复时间线

- **2025-10-27 21:30** - 问题首次出现
- **2025-10-27 23:18** - 完成根因分析
- **2025-10-28 09:51** - 修复方案实施
- **2025-10-28 10:00** - 测试验证通过

## 总结

这是一个由于不当使用 Starlette Request 对象导致的 cancel scope 管理问题。通过利用 Starlette 内置的请求体缓存机制，避免替换 Request 对象，可以彻底解决此问题，并且修复方案在所有 Starlette 版本中都适用。
