# 流式响应文件日志记录

## 概述

为了优化流式响应的日志记录，避免在控制台日志中输出大量流式内容，现在将流式响应内容存储到独立的文件中。

## 模块结构

### 文件组织
流式响应处理逻辑已提取到独立模块中：
```
app/router/middleware_pkg/
├── __init__.py                 # 模块导出
├── log_requests.py            # 主中间件，处理请求日志
└── streaming_response_handler.py  # 流式响应处理模块
```

### 核心模块
- **`streaming_response_handler.py`**: 专门处理流式响应的拦截、记录和文件存储
- **`log_requests.py`**: 主中间件，调用流式响应处理模块

## 功能特性

### 文件存储
- **存储目录**: `log/streaming_responses/`
- **文件命名**: `{timestamp}_{request_id}.log`
- **编码格式**: UTF-8
- **写入模式**: 追加模式

### 日志格式
每个流式响应文件包含以下信息：
```
[2025-01-11T15:30:45.123456] Chunk 1 (128 bytes):
{实际的chunk内容}
==================================================
[2025-01-11T15:30:45.234567] Chunk 2 (256 bytes):
{实际的chunk内容}
==================================================

[2025-01-11T15:30:47.345678] Stream completed: 15 chunks, 2048 bytes total
```

### 控制台日志
在控制台日志中只记录文件路径信息：
```
🟢 [abc-123-def] 流式响应文件: log/streaming_responses/20250111_153045_abc-123-def.log
🟢 [abc-123-def] 流式响应完成: 块数=15, 总大小=2048B, 流处理时间=2150.45ms
```

## 配置要求

### 调试模式
- 流式响应文件记录**仅在调试模式**下启用
- 通过 `Config.is_debug_mode()` 控制
- 生产环境下不会创建文件，避免磁盘空间占用

### 权限要求
- 应用需要对 `log/` 目录的写权限
- 自动创建 `log/streaming_responses/` 子目录

## 错误处理

### 文件写入失败
如果文件写入失败，会在控制台记录错误信息：
```
写入流式响应日志文件失败: {错误信息}
```

### 异常安全
- 使用 `try-finally` 确保资源正确释放
- 文件操作异常不会影响正常的流式响应处理
- 解码错误使用 `errors='ignore'` 参数

## 性能考虑

### I/O 操作
- 使用追加模式写入，避免频繁的文件打开/关闭
- 每个chunk独立写入，确保数据实时性
- 文件操作在异步上下文中同步执行，对性能影响较小

### 磁盘空间
- 仅在调试模式下启用，生产环境无影响
- 文件大小取决于实际的流式响应内容
- 建议定期清理旧的日志文件

## 使用示例

### 启用调试模式
在配置文件中设置：
```yaml
app:
  debug: true
```

### 查看流式响应日志
```bash
# 查看日志目录
ls -la log/streaming_responses/

# 查看特定请求的流式响应
cat log/streaming_responses/20250111_153045_abc-123-def.log
```

## 实现细节

### 核心函数
- `_ensure_streaming_log_dir()`: 确保日志目录存在
- `_get_streaming_log_file_path()`: 生成日志文件路径
- `_write_chunk_to_file()`: 将单个chunk写入文件
- `_write_stream_completion_info()`: 写入流式响应完成信息
- `_create_streaming_wrapper()`: 包装流式响应迭代器
- `handle_streaming_response()`: 主要的流式响应处理函数

### 文件路径生成
```python
def _get_streaming_log_file_path(request_id: str) -> str:
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"{timestamp}_{request_id}.log"
    return os.path.join(STREAMING_RESPONSE_LOG_DIR, filename)
```

### 模块调用关系
```python
# log_requests.py 中的调用
if is_streaming:
    response = handle_streaming_response(response, request_id, process_time)
```

## 维护建议

1. **定期清理**: 建议设置定时任务清理超过7天的日志文件
2. **监控空间**: 在调试模式下关注磁盘空间使用情况
3. **日志轮转**: 可以考虑添加日志文件大小限制和轮转机制

## 重构优势

1. **代码组织**: 流式响应逻辑独立，便于维护和测试
2. **职责分离**: 主中间件专注于请求日志，流式处理独立模块化
3. **可复用性**: 流式响应处理逻辑可在其他地方复用
4. **可测试性**: 独立模块更容易进行单元测试
