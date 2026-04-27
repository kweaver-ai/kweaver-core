# 异常处理器重构文档

## 重构时间
2025-11-02

## 重构目标
将单一的 `exception_handler.py` 文件重构为包结构，提高代码的可维护性和可扩展性。

## 重构前结构
```
app/router/
└── exception_handler.py  (260行，包含所有异常处理逻辑)
```

## 重构后结构
```
app/router/exception_handler/
├── __init__.py              # 统一导入和注册入口
├── common.py                # 通用导入和工具函数
├── validation_handler.py    # RequestValidationError 处理器
├── param_handler.py         # ParamException 处理器
├── permission_handler.py    # AgentPermissionException 处理器
├── code_handler.py          # CodeException 处理器
└── unknown_handler.py       # Exception 未知异常处理器
```

## 各文件职责

### 1. `__init__.py`
- 提供统一的注册函数 `register_exception_handlers(app)`
- 导出所有处理器函数供单独使用
- 集中管理异常处理器的注册逻辑

### 2. `common.py`
- 包含所有处理器共享的导入
- 避免重复导入，提高代码复用性

### 3. `validation_handler.py`
- 处理 FastAPI 请求参数验证错误 (`RequestValidationError`)
- 提供详细的参数错误信息
- 记录结构化日志，包含调用栈信息

### 4. `param_handler.py`
- 处理自定义参数异常 (`ParamException`)
- 返回 400 状态码

### 5. `permission_handler.py`
- 处理权限异常 (`AgentPermissionException`)
- 返回 403 状态码

### 6. `code_handler.py`
- 处理代码异常 (`CodeException`)
- 返回 500 状态码

### 7. `unknown_handler.py`
- 处理未知异常 (`Exception`)
- 提供兜底的异常处理
- 返回 500 状态码

## 使用方式

### 在 `app/router/__init__.py` 中注册
```python
from app.router.exception_handler import register_exception_handlers
register_exception_handlers(app)
```

### 单独使用某个处理器
```python
from app.router.exception_handler import handle_param_error
```

## 调用栈记录优化

### 问题
原始代码中 `inspect.stack()[1:6]` 只获取前5层调用栈，导致：
1. 对于深层框架（如 FastAPI），前5层都是框架代码
2. 无法看到实际的业务代码调用路径
3. 调试困难

### 改进
```python
# 改进前
for frame_info in inspect.stack()[1:6]:  # 只取5层
    filename = frame_info.filename
    if "/agent-executor/" in filename:
        filename = filename.split("/agent-executor/")[-1]
    stack_trace.append({...})  # 包含框架代码

# 改进后
for frame_info in inspect.stack()[1:16]:  # 取15层
    filename = frame_info.filename
    if "/agent-executor/" in filename:  # 只保留项目代码
        simplified_path = filename.split("/agent-executor/")[-1]
        stack_trace.append({...})
```

### 改进效果
- **增加层数**：从 5 层增加到 15 层，确保能捕获到业务代码
- **过滤框架代码**：只记录项目内的代码，减少噪音
- **简化路径**：只显示项目内的相对路径，提高可读性

### 日志输出示例

**有业务代码调用时：**
```
call_stack:
  File "app/router/exception_handler/validation_handler.py", line 166, in handle_param_error
  File "app/router/__init__.py", line 310, in <module>
  File "app/router/agent_controller_pkg/common_v2.py", line 45, in run_agent_v2
```

**参数验证错误时：**
```
call_stack: null
```
> 注意：FastAPI 的参数验证发生在请求处理之前（Pydantic 模型验证阶段），此时还未进入业务代码，所以 `call_stack` 为 `null`。这是正常现象，可以通过 `request_path` 和 `request_method` 定位到具体的 API 端点。

## 优势

1. **职责分离** - 每个文件只负责一种类型的异常处理
2. **易于维护** - 修改特定异常处理逻辑时只需关注对应文件
3. **代码复用** - 通过 `common.py` 避免重复导入
4. **统一管理** - 通过 `__init__.py` 提供统一的注册接口
5. **可扩展性** - 新增异常类型时只需添加新的处理器文件
6. **更好的调试** - 优化的调用栈记录，只显示项目相关代码

## 注意事项

1. 所有异常处理器都使用 `struct_logger` 记录日志
2. 异常处理器函数命名遵循 `handle_xxx_exception` 或 `handle_xxx_error` 模式
3. 调用栈记录只保留项目相关代码，过滤框架代码
4. 确保在应用启动时调用 `register_exception_handlers(app)` 注册所有处理器

## 相关文件
- `/app/router/exception_handler/` - 异常处理器包
- `/app/router/__init__.py` - 注册异常处理器
- `/docs/logging/struct_logger_usage.md` - 结构化日志使用文档
