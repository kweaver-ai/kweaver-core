# Agent开发指南

## 概述

Agent-App 提供了完整的Agent开发平台，支持开发者快速创建、调试和部署智能对话Agent。本指南将详细介绍如何使用Agent API搭建对话平台的最佳实践。

## 1. 创建Agent并发布为API

### 1.1 Agent创建流程

**步骤1: 定义Agent能力**
- 明确Agent的应用场景和功能范围
- 设计对话流程和用户交互模式
- 确定所需的数据源和工具集成

**步骤2: 配置Agent参数**
```json
{
  "agent_name": "智能客服助手",
  "description": "提供产品咨询和客户服务",
  "model_config": {
    "model_type": "gpt-4",
    "temperature": 0.7,
    "max_tokens": 2000
  },
  "skills": [
    "产品查询",
    "故障排查",
    "订单处理"
  ],
  "data_sources": [
    "产品知识库",
    "FAQ文档"
  ]
}
```

**步骤3: 集成工具和数据源**
- 配置外部API调用
- 集成知识库检索
- 设置文件处理能力

**步骤4: 测试和验证**
- 功能完整性测试
- 性能压力测试
- 用户体验验证

**步骤5: 发布为API**
- 生成唯一的App Key
- 配置API访问权限
- 设置使用配额和限制

### 1.2 API发布配置

**API端点配置**:
```
POST /app/{app_key}/api/chat/completion
```

**认证方式**:
- Bearer Token认证
- API Key验证
- 访问频率限制

**发布检查清单**:
- [ ] Agent功能测试通过
- [ ] API接口文档完整
- [ ] 错误处理机制完善
- [ ] 性能指标达标
- [ ] 安全配置正确

## 2. API调试界面使用

### 2.1 API调试界面各参数含义

**基础参数**:
- **app_key**: Agent应用标识，用于区分不同Agent
- **agent_id**: Agent唯一标识，指定要使用的Agent
- **agent_version**: Agent版本，支持版本管理和回滚

**对话参数**:
- **query**: 用户输入的查询内容
- **conversation_id**: 会话ID，用于多轮对话上下文
- **history**: 历史对话记录，支持自定义历史管理
- **chat_mode**: 对话模式（normal/deep_thinking）

**流式参数**:
- **stream**: 是否启用流式响应
- **inc_stream**: 是否启用增量流式传输

**高级参数**:
- **temp_files**: 临时文件列表，支持文件上传和处理
- **tool**: 工具调用配置
- **llm_config**: 大模型参数配置
- **data_source**: 数据源配置

### 2.2 调试API流程

**步骤1: 准备测试数据**
```json
{
  "agent_id": "your_agent_id",
  "query": "请介绍一下你们的产品",
  "stream": true,
  "conversation_id": "test_session_001"
}
```

**步骤2: 发送调试请求**
- 使用Postman或curl工具
- 设置正确的认证头
- 监控请求响应时间

**步骤3: 分析响应结果**
```json
{
  "conversation_id": "test_session_001",
  "user_message_id": "msg_001",
  "assistant_message_id": "msg_002",
  "message": {
    "content": {
      "final_answer": {
        "answer": {
          "text": "我们提供智能客服解决方案..."
        }
      }
    }
  }
}
```

**步骤4: 性能优化**
- 调整模型参数（temperature, max_tokens）
- 优化数据源配置
- 改进提示词设计

## 3. 使用Agent API接口

### 3.1 会话自动创建

Agent-App 采用**自动会话创建**机制，无需专门的初始化接口。当用户发起第一次对话时，系统会自动创建会话并返回会话ID。

**首次对话请求** (不提供conversation_id):
```json
{
  "agent_id": "your_agent_id",
  "query": "开始新的对话",
  "stream": true
}
```

**首次对话响应** (包含自动生成的会话ID):
```json
{
  "conversation_id": "conv_123456789",
  "user_message_id": "user_msg_001",
  "assistant_message_id": "assistant_msg_001",
  "message": {
    "id": "assistant_msg_001",
    "content": {
      "text": "您好！我是智能助手，有什么可以帮您的吗？"
    }
  }
}
```

**继续对话请求** (使用返回的conversation_id):
```json
{
  "agent_id": "your_agent_id",
  "conversation_id": "conv_123456789",
  "query": "继续刚才的话题",
  "stream": true
}
```

**会话管理功能**:
- **获取会话列表**: `GET /app/{app_key}/conversations`
- **获取会话详情**: `GET /app/{app_key}/conversations/{conversation_id}`
- **更新会话信息**: `PUT /app/{app_key}/conversations/{conversation_id}`
- **删除会话**: `DELETE /app/{app_key}/conversations/{conversation_id}`

### 3.2 调用API接口

**标准调用流程**:

1. **首次对话** (不提供conversation_id，系统自动创建会话)
```json
{
  "agent_id": "your_agent_id",
  "query": "请帮我查询订单状态",
  "stream": true,
  "temp_files": [
    {
      "id": "file_001",
      "name": "order_info.pdf",
      "type": "doc"
    }
  ]
}
```

2. **继续对话** (使用返回的conversation_id)
```json
{
  "agent_id": "your_agent_id",
  "conversation_id": "conv_123456789",
  "query": "继续刚才的查询",
  "stream": true
}
```

3. **设置认证头**
```http
Authorization: Bearer your_token
Content-Type: application/json
```

4. **处理响应结果**
```javascript
// 流式响应处理示例
const eventSource = new EventSource('/api/chat/completion');

eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  if (data.event === 'message') {
    // 处理增量消息
    updateUI(data.data.message);
  } else if (data.event === 'end') {
    // 对话结束
    eventSource.close();
  }
};
```

### 3.3 结果返回处理

#### 仅需要最终结果
直接使用 `final_answer` 中的内容：

```json
{
  "final_answer": {
    "query": "请帮我查询订单状态",
    "answer": {
      "text": "您的订单状态是已发货，预计明天送达。",
      "cites": {
        "order_system": "订单号: ORD20231118001"
      }
    },
    "thinking": "用户询问订单状态，需要查询订单系统...",
    "skill_process": [
      {
        "agent_name": "订单查询",
        "text": "正在查询订单系统...",
        "status": "completed"
      }
    ]
  }
}
```

#### 需要展示对话过程
使用 `middle_answer` 中的 `progress` 字段：

```json
{
  "middle_answer": [
    {
      "progress": [
        {
          "step": 1,
          "description": "解析用户查询意图",
          "status": "completed",
          "result": "识别为订单状态查询"
        },
        {
          "step": 2,
          "description": "调用订单系统API",
          "status": "completed",
          "result": "获取到订单信息"
        },
        {
          "step": 3,
          "description": "生成回复内容",
          "status": "in_progress"
        }
      ]
    }
  ]
}
```

#### 高级处理
结合两者实现渐进式展示：

```javascript
function handleStreamResponse(data) {
  if (data.middle_answer && data.middle_answer.length > 0) {
    // 展示处理进度
    showProgress(data.middle_answer[0].progress);
  }

  if (data.final_answer) {
    // 展示最终结果
    showFinalResult(data.final_answer);
  }
}
```

## 4. 最佳实践

### 4.1 会话管理最佳实践

**会话创建机制**:
- **首次对话**: 不提供 `conversation_id`，系统自动创建会话
- **继续对话**: 使用返回的 `conversation_id` 维护多轮对话上下文
- **会话复用**: 合理复用现有会话，避免创建过多会话

**会话生命周期管理**:
- 及时清理过期会话
- 合理设置会话超时时间
- 实现会话数据备份

**上下文管理**:
- 控制历史记录长度
- 实现上下文压缩
- 支持会话恢复

### 4.2 性能优化建议

**请求优化**:
- 合理设置流式传输频率
- 使用增量传输减少带宽
- 实现请求重试机制

**响应处理**:
- 客户端缓存策略
- 错误降级处理
- 用户体验优化

### 4.3 错误处理策略

**客户端错误处理**:
```javascript
async function callAgentAPI(request) {
  try {
    const response = await fetch('/api/chat/completion', {
      method: 'POST',
      headers: {
        'Authorization': 'Bearer ' + token,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(request)
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.error('API调用失败:', error);
    // 实现重试逻辑或降级处理
    handleError(error);
  }
}
```

**服务端错误码处理**:
- `400`: 检查请求参数格式
- `401`: 验证认证信息
- `403`: 检查权限配置
- `404`: 确认资源存在性
- `500`: 联系技术支持

### 4.4 安全考虑

**API安全**:
- 使用HTTPS加密传输
- 实现API访问频率限制
- 定期轮换认证密钥

**数据安全**:
- 敏感信息脱敏处理
- 用户数据权限控制
- 审计日志记录

## 5. 常见问题解答

### Q: 如何处理大文件上传？
A: 使用临时区域接口先上传文件，然后在对话中引用文件ID。

### Q: 如何实现多轮对话上下文？
A: 使用conversation_id关联对话，系统会自动维护上下文。

### Q: 流式响应中断如何处理？
A: 实现重连机制，使用interrupted_assistant_message_id继续对话。

### Q: 如何优化响应速度？
A: 调整模型参数，优化数据源配置，使用合适的对话模式。

---

*最后更新: 2025-11-18*