# TTFT (Time To First Token) 功能说明

## 功能概述
已为 Agent V2 的流式输出添加 TTFT（Time To First Token）字段，用于测量从请求开始到第一个token输出的时间。

## 实现细节

### 1. 时间记录
在 `run_agent_v2.py` 的 `run_agent` 函数中记录请求开始时间：
```python
start_time = time.time()
```

### 2. TTFT 计算
在 `output.py` 中添加了 `add_ttft` 方法：
- 在第一个chunk输出时计算TTFT（毫秒）
- TTFT = (当前时间 - 开始时间) × 1000
- 将TTFT值添加到每个输出chunk的最外层

### 3. 关键特性
- **一致性**：同一次API调用的所有SSE响应消息中的TTFT值相同
- **位置**：TTFT字段添加在每个响应对象的最外层
- **单位**：毫秒（ms）

## 输出示例

每个SSE响应的数据格式：
```json
{
  "answer": { ... },
  "status": "False",
  "ttft": 1234
}
```

其中：
- `ttft`: 从请求开始到第一个token的时间（毫秒）
- 同一次请求的所有响应中，`ttft` 值保持不变

## 修改的文件

1. **app/router/agent_controller_pkg/run_agent_v2.py**
   - 启用 `start_time` 记录
   - 将 `start_time` 传递给 `result_output` 方法

2. **app/logic/agent_core_logic_v2/output.py**
   - 添加 `time` 模块导入
   - 新增 `add_ttft` 方法
   - 修改 `result_output` 方法签名，添加 `start_time` 参数
   - 在输出管道中应用 `add_ttft` 处理

## 测试建议

1. 发起一个Agent V2请求
2. 观察SSE流式响应
3. 验证每个响应消息都包含 `ttft` 字段
4. 验证同一次请求的所有响应中 `ttft` 值相同
5. 验证 `ttft` 值合理（通常在几百到几千毫秒之间）
