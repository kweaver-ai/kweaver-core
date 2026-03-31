# KeyboardInterrupt 警告问题说明

## 问题描述

在使用 PyCharm 调试器停止 uvicorn 服务时，会看到以下警告：

```
Task exception was never retrieved
future: <Task finished name='Task-1' coro=<Server.serve() done, defined at ...> exception=KeyboardInterrupt()>
Traceback (most recent call last):
  ...
KeyboardInterrupt
```

## 问题原因

这是 **uvicorn + PyCharm 调试器 + asyncio** 的已知交互问题：

1. **uvicorn 的信号处理**：uvicorn 使用 `signal.raise_signal()` 来处理 SIGINT
2. **PyCharm 调试器**：使用 `pydevd_nest_asyncio` 来支持异步调试
3. **asyncio 任务**：当程序被中断时，正在运行的异步任务没有被正确清理

这个警告**不影响程序功能**，只是一个日志输出。

## 已实施的解决方案

### 1. 添加信号处理器 (main.py)

```python
def signal_handler(signum, frame):
    """信号处理器"""
    print("\n接收到停止信号，正在关闭服务...")
    sys.exit(0)

signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)
```

### 2. 添加 asyncio 异常处理器 (main.py)

```python
def asyncio_exception_handler(loop, context):
    """Asyncio 异常处理器 - 抑制 KeyboardInterrupt 警告"""
    exception = context.get("exception")

    # 忽略 KeyboardInterrupt 和 CancelledError
    if isinstance(exception, (KeyboardInterrupt, asyncio.CancelledError)):
        return

    # 其他异常正常处理
    if exception:
        print(f"Asyncio 异常: {exception}")

loop = asyncio.get_event_loop()
loop.set_exception_handler(asyncio_exception_handler)
```

### 3. 清理应用任务 (app/router/__init__.py)

```python
@app.on_event("shutdown")
async def shutdown_event():
    # 关闭可观测模块
    shutdown_observability()

    # 清理所有正在运行的 rebuild 任务
    try:
        from app.logic.agent_core_logic_v2.session.session_rebuild_service import rebuild_service

        tasks_to_cancel = list(rebuild_service.rebuild_tasks.values())
        for task in tasks_to_cancel:
            if not task.done():
                task.cancel()

        if tasks_to_cancel:
            await asyncio.gather(*tasks_to_cancel, return_exceptions=True)

        rebuild_service.rebuild_tasks.clear()
    except Exception:
        pass
```

## 为什么警告仍然出现？

即使实施了上述解决方案，警告可能仍然会出现，原因是：

1. **uvicorn 内部任务**：uvicorn 自己创建的异步任务在被中断时产生警告
2. **PyCharm 调试器干预**：调试器改变了正常的信号处理流程
3. **时序问题**：`shutdown_event` 可能在 uvicorn 的内部任务被取消之前就执行了

## 完全消除警告的方法

### 方法 1：不使用 PyCharm 调试器（推荐用于生产环境）

```bash
# 直接运行，不通过调试器
python main.py
```

使用 Ctrl+C 停止时，警告会少很多或完全消失。

### 方法 2：使用 uvicorn 命令行（开发环境）

```bash
uvicorn app.router:app --host 0.0.0.0 --port 8080 --reload
```

### 方法 3：配置 PyCharm 不显示此类警告

在 PyCharm 中：
1. Settings → Editor → Inspections
2. 搜索 "Exception"
3. 禁用或降低 "Unhandled exception" 的严重级别

### 方法 4：使用环境变量抑制警告

```bash
export PYTHONWARNINGS="ignore::ResourceWarning"
python main.py
```

## 最佳实践

### 开发环境
- 接受这个警告的存在（它不影响功能）
- 或使用 `uvicorn` 命令行而不是 PyCharm 调试器

### 生产环境
- 使用进程管理器（systemd、supervisor）
- 不会遇到这个问题

## 总结

这个 `KeyboardInterrupt` 警告是：
- ✅ **正常现象**：uvicorn + PyCharm 调试器的已知问题
- ✅ **不影响功能**：只是日志输出
- ✅ **可以忽略**：在开发环境中
- ✅ **生产环境无此问题**：使用进程管理器时不会出现

我们已经实施了最佳的清理逻辑，但由于 uvicorn 内部实现的限制，无法完全消除这个警告。

## 参考资料

- [uvicorn Issue #1579](https://github.com/encode/uvicorn/issues/1579)
- [Python asyncio - Exception handling](https://docs.python.org/3/library/asyncio-exceptions.html)
- [PyCharm Debugging Async Code](https://www.jetbrains.com/help/pycharm/debugging-async-code.html)
