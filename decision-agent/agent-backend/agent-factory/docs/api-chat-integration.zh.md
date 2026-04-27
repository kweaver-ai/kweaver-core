# Agent Factory API Chat 对接文档

本文档基于以下信息整理当前 Agent Factory 的 API Chat 对接方式：
- 当前 Agent Factory 代码
- 当前 OpenAPI 产物
- DIP 中 `test_query` 的真实页面
- 已观测到的 processed 流式响应结构

这是当前 API Chat 对接的主文档，供以下 README 引用：
- `agent-backend/agent-factory/README.zh.md`
- `decision-agent/README.zh.md`

## 1. 相关入口

Agent 页面：
- 编辑页：
  `https://{ip}/studio/agent/development/my-agent-list/config?agentId=01K9TSWQ0SH5RYT60SJ1S783AC&filterParams=%7B%22mode%22%3A%22myAgent%22%7D`
- 聊天页：
  `https://{ip}/studio/agent/development/my-agent-list/usage?id=01K9TSWQ0SH5RYT60SJ1S783AC&version=v0&agentAppType=common&conversation_id=01KPAKR3GMEJAR7E6T5XMVBFTH`

OpenAPI 产物：
- `./api/README.md`
- `./api/agent-factory.html`
- `./api/agent-factory.json`
- `./api/agent-factory.yaml`

本地服务启动后的运行时入口：
- `http://{ip}/scalar`
- `http://{ip}/scalar/doc.json`
- `http://{ip}/scalar/doc.yaml`

关键源码：
- `../src/driveradapter/api/httphandler/agenthandler/api_chat.go`
- `../src/driveradapter/api/httphandler/agenthandler/chat.go`
- `../src/driveradapter/api/httphandler/agenthandler/get_api_doc.go`
- `../src/driveradapter/api/httphandler/agenthandler/define.go`
- `../src/driveradapter/api/rdto/agent/req/chat_req.go`
- `../src/domain/service/agentrunsvc/chat.go`
- `../src/domain/service/agentrunsvc/api_doc.go`

---

## 2. 先分清三个标识

| 字段 | 典型来源 | 用途 | 备注 |
|---|---|---|---|
| `agent_id` | Agent 对象的 ID | 标识具体 Agent 实体 | 普通网页 Chat 常用 |
| `agent_key` | Agent 对外 Key | API Chat 请求体中必填 | `APIChat` handler 会强校验 |
| `app_key` | 应用 Key | 路径参数 | 历史遗留标识，现在没有此概念了，一般对应 `agent_id` 或 `agent_key` ，具体请看接口文档|

---

## 3. 普通 Chat 与 API Chat 的区别

`api_chat.go` 中的注释表明：API Chat 与普通外部 Chat 的主体逻辑几乎一致，核心差异主要在 URL，以及长期 token 的使用场景。

| 维度 | 普通 Chat | API Chat |
|---|---|---|
| 服务内路由 | `/v1/app/{app_key}/chat/completion` | `/v1/app/{app_key}/api/chat/completion` 或 `/v1/api/chat/completion` |
| 常见外部网关路由 | `/api/agent-factory/v1/app/{app_key}/chat/completion` | `/api/agent-factory/v1/app/{app_key}/api/chat/completion` 或 `/api/agent-factory/v1/api/chat/completion` |
| body 主标识 | `agent_id` | `agent_key` |
| 发布要求 | 正常发布可用 | 必须已发布为 API Agent，否则返回 403 |
| 返回形式 | SSE / 非流式 JSON | SSE / 非流式 JSON |

推荐 API Chat 地址(新版本已去掉路径参数 `app_key`，暂时兼容app_key，详情请看接口文档)：

```http
POST /api/agent-factory/v1/api/chat/completion
```


---

## 4. 普通 Chat 请求

抓到的请求地址：

```http
POST https://{ip}/api/agent-factory/v1/app/01K9TSWQ0SH5RYT60SJ04BXF89/chat/completion
```

抓到的请求头（token 已脱敏）：

```http
Authorization: Bearer ***
Content-Type: application/json; charset=utf-8
accept: text/event-stream
responseType: text/event-stream
x-language: zh-CN
x-business-domain: bd_public
```

抓到的请求体：

```json
{
  "query": "抓取curl用测试",
  "conversation_id": "01KPAKR3GMEJAR7E6T5XMVBFTH",
  "agent_id": "01K9TSWQ0SH5RYT60SJ1S783AC",
  "agent_version": "v0",
  "stream": true,
  "inc_stream": true,
  "chat_option": {
    "is_need_history": true,
    "is_need_doc_retrival_post_process": true,
    "is_need_progress": true,
    "enable_dependency_cache": true
  }
}
```

这很有参考价值，因为 API Chat 的 body 语义几乎相同，只是 API Chat 需要把 `agent_id` 换成 `agent_key`。

---

## 5. API Chat 请求体

最小可用 body：

```json
{
  "agent_key": "your_agent_key",
  "query": "你好",
  "stream": true
}
```

常用字段：

| 字段 | 说明 |
|---|---|
| `agent_key` | API Chat 必填 |
| `agent_version` | 可选；默认 `latest` |
| `query` | 必填用户问题 |
| `conversation_id` | 多轮对话时使用 |
| `stream` | 强烈建议显式传入 |
| `inc_stream` | 仅在 `stream=true` 时有意义 |
| `history` | 可选自定义历史 |
| `custom_querys` | 可选自定义输入变量 |
| `temp_files` | 可选临时文件 |
| `temporary_area_id` | 可选临时区 ID |
| `chat_option` | 控制历史、进度、检索后处理、缓存等行为 |

与当前页面行为接近的一组 `chat_option`：

```json
{
  "chat_option": {
    "is_need_history": true,
    "is_need_doc_retrival_post_process": true,
    "is_need_progress": true,
    "enable_dependency_cache": true
  }
}
```

---

## 6. API Chat cURL 示例

流式示例：

```bash
curl 'https://{ip}/api/agent-factory/v1/api/chat/completion' \
  -X POST \
  -H 'Authorization: ${AUTH_HEADER}' \
  -H 'Content-Type: application/json; charset=utf-8' \
  -H 'Accept: text/event-stream' \
  -H 'x-language: zh-CN' \
  -H 'x-business-domain: bd_public' \
  --data-raw '{
    "agent_key": "<agent_key>",
    "agent_version": "v0",
    "query": "你好，介绍一下你自己",
    "conversation_id": "<optional_conversation_id>",
    "stream": true,
    "inc_stream": true,
    "executor_version": "v2",
    "chat_option": {
      "is_need_history": true,
      "is_need_doc_retrival_post_process": true,
      "is_need_progress": true,
      "enable_dependency_cache": true
    }
  }'
```

非流式示例：

```bash
curl 'https://{ip}/api/agent-factory/v1/api/chat/completion' \
  -X POST \
  -H 'Authorization: ${AUTH_HEADER}' \
  -H 'Content-Type: application/json' \
  -H 'x-language: zh-CN' \
  -H 'x-business-domain: bd_public' \
  --data-raw '{
    "agent_key": "<agent_key>",
    "agent_version": "v0",
    "query": "给我一段摘要",
    "stream": false,
    "executor_version": "v2"
  }'
```

---

## 7. 返回结构怎么读

核心返回形态：

```json
{
  "conversation_id": "...",
  "user_message_id": "...",
  "assistant_message_id": "...",
  "agent_run_id": "...",
  "message": {
    "content_type": "prompt",
    "content": {
      "final_answer": {
        "query": "...",
        "answer": {
          "text": "最终答案文本",
          "cites": {}
        },
        "thinking": "可选思考过程",
        "skill_process": []
      },
      "middle_answer": []
    }
  }
}
```

最常用字段：
- `conversation_id`
- `user_message_id`
- `assistant_message_id`
- `agent_run_id`
- `message.content.final_answer.answer.text`
- `message.content.final_answer.thinking`
- `message.content.final_answer.skill_process`
- `message.content.middle_answer.progress`

---

## 8. 完整流式输出 vs 增量流式输出

已观测到的 processed 流式样本表明：Agent Factory 在内部会先把流式结果标准化为**完整快照**，而前端可以把它消费成**增量 delta**。

### 8.1 什么叫完整流式快照

每个 chunk 都可以是“当前时刻的完整响应对象”，而不是单纯的文本增量。

一个代表性的完整快照形态如下：

```json
{
  "conversation_id": "...",
  "agent_run_id": "...",
  "assistant_message_id": "...",
  "message": {
    "content_type": "prompt",
    "content": {
      "final_answer": {
        "query": "hi",
        "answer": {
          "text": "你好！我是你的AI助手..."
        },
        "thinking": "",
        "skill_process": null,
        "output_variables_config": {
          "answer_var": "answer",
          "doc_retrieval_var": "doc_retrieval_res",
          "graph_retrieval_var": "graph_retrieval_res",
          "related_questions_var": "related_questions",
          "other_vars": null,
          "middle_output_vars": null
        }
      },
      "middle_answer": {
        "progress": [],
        "doc_retrieval": null,
        "graph_retrieval": null,
        "other_variables": {}
      }
    }
  },
  "error": null
}
```

### 8.2 什么叫增量流式输出

增量流式输出是下游消费模式：只把新增部分渲染或向下游转发。

例子：
- 快照 N：`你好！我是你的AI助手`
- 快照 N+1：`你好！我是你的AI助手，可以帮你`
- 前端增量 delta：`，可以帮你`

### 8.3 什么时候算“增量”

一般可按以下条件理解为增量：
- `stream=true`
- 且下游 / 前端按 `inc_stream=true` 的方式消费新增 delta，而不是每次都重绘累计全文

### 8.4 增量流转的作用

增量流转主要帮助：
- 减少前端单次有效更新量
- 减少客户端整段重绘等
- 提高前端渲染性能
- 减少网络传输数据量

### 8.5 不使用增量流转的影响

如果不使用增量流转：
- 仍然可以流式返回
- 但每个事件更可能携带当前累计全文
- payload 会随着回答变长而变大
- 客户端的渲染成本也会更高
- 网络传输数据量也会更大（可能非常大）

这通常不会影响正确性，但会影响效率与体验。

---

## 9. 过程态与工具调用结构

一个代表性的结构如下：

```json
{
  "message": {
    "content": {
      "final_answer": {
        "answer": {
          "text": "这是一个关于文档问答智能体的配置信息摘要..."
        },
        "thinking": "",
        "skill_process": null
      },
      "middle_answer": {
        "progress": [
          {
            "id": "...",
            "stage": "skill",
            "status": "processing",
            "skill_info": {
              "type": "TOOL",
              "name": "获取agent详情",
              "args": [
                {
                  "name": "key",
                  "value": "DocQA_Agent",
                  "type": "str"
                }
              ]
            }
          }
        ],
        "doc_retrieval": null,
        "graph_retrieval": null,
        "other_variables": {}
      }
    }
  }
}
```

这意味着客户端可以选择：
- 只展示最终答案
- 展示进度条 / 步骤流 + 最终答案
- 两者同时展示

一个实用的渲染规则是：
- 用 `final_answer.answer.text` 驱动主答案区
- 用 `middle_answer.progress` 驱动工具 / 步骤 / 运行态 UI

---

## 10. Agent Factory 内部调用链

主路由注册：
- `../src/driveradapter/api/httphandler/agenthandler/define.go`

API Chat handler 的主要行为：
- 绑定请求 JSON
- 注入默认值（`agent_version=latest`、`executor_version=v2`）
- 从上下文读取 visitor 身份与 token
- 设置 `CallType=APIChat`
- 强校验 `agent_key`
- 调用统一服务 `agentSvc.Chat(...)`
- 按 `stream` 输出 SSE 或最终 JSON

统一服务 `../src/domain/service/agentrunsvc/chat.go` 的主要步骤：
1. 按 ID 或 Key 解析 Agent
2. 如果不是 API Agent 发布态，则 API Chat 直接 403
3. 加载历史消息和消息索引
4. 创建 / 更新用户消息与助手消息
5. 建立 stop channel 与 session
6. 创建 conversation session
7. 若启用了 sandbox，则确保 sandbox session 存在
8. 构建发给 agent-executor 的请求
9. 调用 agent-executor
10. 统一处理流式结果并输出标准化 chunk

---

## 11. API 文档入口

仓库内静态产物：
- `./api/agent-factory.html`
- `./api/agent-factory.json`
- `./api/agent-factory.yaml`

本地服务启动后的 Scalar：
- `http://localhost:13020/scalar`
- `http://localhost:13020/scalar/doc.json`
- `http://localhost:13020/scalar/doc.yaml`

单 Agent 动态文档入口（只针对发布为API的Agent）：

```http
POST /api/agent-factory/v1/app/{app_key}/api/doc
```

请求体：

```json
{
  "agent_id": "<agent_id>",
  "agent_version": "v0"
}
```

这个动态接口会自动改写：
- summary = 当前 Agent 名称
- description = 当前 Agent profile
- request example = 当前 `agent_key` / 版本 / 自定义输入字段

---

## 12. 对接注意事项

- 网页 Chat 能用，不代表 API Chat 一定可用；API Chat 依赖 API Agent 发布状态
- `inc_stream` 增量流式只在流式模式`stream`下有意义
- 想尽量对齐页面行为时，可复用真实页面请求里的 `chat_option`
- 只关心最终答案时，读取 `message.content.final_answer`的内容
- 需要展示工具 / 过程态时，读取 `message.content.middle_answer.progress`的内容
