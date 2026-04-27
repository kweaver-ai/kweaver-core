# struct_logger.console_logger.error 用法记录

## 概述

`struct_logger.console_logger.error` 是项目中用于记录错误级别日志到控制台的方法，属于结构化日志系统的一部分。

## 导入方式

```python
from app.common.struct_logger import struct_logger
# 使用
struct_logger.console_logger.error(message, exc_info=e)

# 或者直接导入
from app.common.struct_logger import console_logger
# 使用
console_logger.error(message, exc_info=e)
```

## 项目中的使用位置

### 1. app/logic/agent_core_logic_v2/exception.py:25

```python
class ExceptionHandler:
    @classmethod
    async def handle_exception(
        cls, e: Exception, res: Dict[str, Any], headers: Dict[str, str]
    ) -> None:
        """处理异常

        Args:
            e: 异常对象
            res: 结果字典
            headers: HTTP请求头
        """

        message = "agent run failed: {}".format(repr(e))

        # 打印到控制台
        struct_logger.console_logger.error(message, exc_info=e)

        if not isinstance(res, dict):
            res = {}

        res["error"] = await get_format_error_info(headers, e)
        res["status"] = "Error"
```

**使用场景：** Agent 核心逻辑异常处理

### 2. app/router/exception_handler/unknown_handler.py:42

```python
def handle_unknown_exception(request: Request, exc: Exception):
    """
    处理未知异常

    Args:
        request: FastAPI 请求对象
        exc: 未知异常（可能是 ExceptionGroup）

    Returns:
        JSONResponse: 错误响应
    """

    # 处理 ExceptionGroup（Python 3.11+）
    actual_exc = exc
    if hasattr(exc, "__class__") and exc.__class__.__name__ == "ExceptionGroup":
        # ExceptionGroup 包含多个异常，取第一个实际异常
        if hasattr(exc, "exceptions") and exc.exceptions:
            actual_exc = exc.exceptions[0]
            struct_logger.console_logger.warning(
                f"ExceptionGroup contains {len(exc.exceptions)} exceptions, using first one",
                exception_count=len(exc.exceptions),
            )

    message = "handle_unknown_exception: {}".format(repr(actual_exc))

    # 记录异常日志
    struct_logger.console_logger.error(message, exc_info=actual_exc)

    # ... 后续处理
```

**使用场景：** 全局未知异常处理器

## 参数说明

### 必需参数
- **message** (str): 错误消息内容

### 可选参数
- **exc_info** (Exception): 异常对象，会自动提取异常信息
- **kwargs**: 其他键值对参数，会被添加到日志中

## 使用示例

### 基本用法
```python
struct_logger.console_logger.error("操作失败")
```

### 带异常信息
```python
try:
    some_operation()
except Exception as e:
    struct_logger.console_logger.error("操作失败", exc_info=e)
```

### 带额外上下文
```python
struct_logger.console_logger.error(
    "API调用失败",
    api_url="http://example.com/api",
    status_code=500,
    user_id="12345"
)
```

## 输出格式

控制台输出使用带色彩的易读格式：

```
2025-10-30 16:32:41 [error    ] 错误消息                       [agent-executor-console] caller=/path/to/file.py:25 api_url=http://example.com status_code=500
```

**特点：**
- 时间戳：`YYYY-MM-DD HH:MM:SS`
- 级别显示：红色 `[error]`
- 消息内容：对齐显示
- Logger名称：`[agent-executor-console]`
- 调用位置：`caller=文件:行号`
- 额外参数：键值对格式

## 最佳实践

### 1. 异常处理
```python
# ✅ 推荐：使用 exc_info 参数
try:
    result = some_function()
except Exception as e:
    struct_logger.console_logger.error("函数调用失败", exc_info=e)
```

### 2. 上下文信息
```python
# ✅ 推荐：提供有用的上下文
struct_logger.console_logger.error(
    "数据库查询失败",
    table="users",
    query_id="query_123",
    duration_ms=1500
)
```

### 3. 避免敏感信息
```python
# ❌ 避免：记录敏感信息
struct_logger.console_logger.error("登录失败", password="secret123")

# ✅ 推荐：脱敏处理
struct_logger.console_logger.error("登录失败", user_id="12345", ip="192.168.1.1")
```

## 与其他 Logger 的区别

| Logger | 输出位置 | 使用场景 |
|--------|----------|----------|
| `struct_logger` | 文件 + 控制台 | 生产环境，需要持久化 |
| `file_logger` | 仅文件 | 大量日志，敏感信息 |
| `console_logger` | 仅控制台 | 开发调试，临时信息 |

## 相关文档

- [结构化日志完整使用指南](/Users/Zhuanz/Work/as/dip_ws/agent-executor/docs/logging/struct_logger_usage.md)
- [struct_logger 模块源码](/Users/Zhuanz/Work/as/dip_ws/agent-executor/app/common/struct_logger/)

## 注意事项

1. **仅输出到控制台**：不会写入日志文件
2. **适合开发调试**：生产环境建议使用 `struct_logger` 或 `file_logger`
3. **性能考虑**：比文件日志稍慢，但比同时输出要快
4. **异常处理**：使用 `exc_info` 参数自动处理异常信息

## 更新记录

- 2025-11-07: 初始文档创建，记录项目中的使用情况
