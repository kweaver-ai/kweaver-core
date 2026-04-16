# Agent Factory API Chat Integration Documentation

This document summarizes the current Agent Factory API Chat integration methods based on:
- Current Agent Factory code
- Current OpenAPI artifacts
- Real page of `test_query` in DIP
- Observed processed streaming response structure

This is the main documentation for current API Chat integration, referenced by the following READMEs:
- `agent-backend/agent-factory/README.zh.md`
- `decision-agent/README.zh.md`

## 1. Related Entry Points

Agent Pages:
- Edit Page:
  `https://{ip}/studio/agent/development/my-agent-list/config?agentId=01K9TSWQ0SH5RYT60SJ1S783AC&filterParams=%7B%22mode%22%3A%22myAgent%22%7D`
- Chat Page:
  `https://{ip}/studio/agent/development/my-agent-list/usage?id=01K9TSWQ0SH5RYT60SJ1S783AC&version=v0&agentAppType=common&conversation_id=01KPAKR3GMEJAR7E6T5XMVBFTH`

OpenAPI Artifacts:
- `./api/README.md`
- `./api/agent-factory.html`
- `./api/agent-factory.json`
- `./api/agent-factory.yaml`

Runtime entry points after starting local service:
- `http://{ip}/swagger/index.html`
- `http://{ip}/swagger/doc.json`
- `http://{ip}/swagger/doc.yaml`

Key Source Code:
- `../src/driveradapter/api/httphandler/agenthandler/api_chat.go`
- `../src/driveradapter/api/httphandler/agenthandler/chat.go`
- `../src/driveradapter/api/httphandler/agenthandler/get_api_doc.go`
- `../src/driveradapter/api/httphandler/agenthandler/define.go`
- `../src/driveradapter/api/rdto/agent/req/chat_req.go`
- `../src/domain/service/agentrunsvc/chat.go`
- `../src/domain/service/agentrunsvc/api_doc.go`

---

## 2. Understanding Three Identifiers

| Field | Typical Source | Purpose | Notes |
|---|---|---|---|
| `agent_id` | Agent object ID | Identifies specific Agent entity | Commonly used in web Chat |
| `agent_key` | Agent external Key | Required in API Chat request body | Strongly validated by `APIChat` handler |
| `app_key` | Application Key | Path parameter | Legacy identifier, no longer has this concept, generally corresponds to `agent_id` or `agent_key`, see API documentation for details |

---

## 3. Differences Between Regular Chat and API Chat

Comments in `api_chat.go` indicate: API Chat and regular external Chat share almost identical main logic, with core differences mainly in URL and long-term token usage scenarios.

| Dimension | Regular Chat | API Chat |
|---|---|---|
| Internal Service Route | `/v1/app/{app_key}/chat/completion` | `/v1/app/{app_key}/api/chat/completion` or `/v1/api/chat/completion` |
| Common External Gateway Route | `/api/agent-factory/v1/app/{app_key}/chat/completion` | `/api/agent-factory/v1/app/{app_key}/api/chat/completion` or `/api/agent-factory/v1/api/chat/completion` |
| Body Main Identifier | `agent_id` | `agent_key` |
| Publication Requirement | Normal publication available | Must be published as API Agent, otherwise returns 403 |
| Return Format | SSE / Non-streaming JSON | SSE / Non-streaming JSON |

Recommended API Chat endpoint (new version has removed path parameter `app_key`, temporarily compatible with app_key, see API documentation for details):

```http
POST /api/agent-factory/v1/api/chat/completion
```

---

## 4. Regular Chat Request

Captured request URL:

```http
POST https://{ip}/api/agent-factory/v1/app/01K9TSWQ0SH5RYT60SJ04BXF89/chat/completion
```

Captured request headers (token desensitized):

```http
Authorization: Bearer ***
Content-Type: application/json; charset=utf-8
accept: text/event-stream
responseType: text/event-stream
x-language: zh-CN
x-business-domain: bd_public
```

Captured request body:

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

This is highly valuable reference because API Chat body semantics are almost identical, except API Chat needs to replace `agent_id` with `agent_key`.

---

## 5. API Chat Request Body

Minimum usable body:

```json
{
  "agent_key": "your_agent_key",
  "query": "你好",
  "stream": true
}
```

Common fields:

| Field | Description |
|---|---|
| `agent_key` | Required for API Chat |
| `agent_version` | Optional; default `latest` |
| `query` | Required user question |
| `conversation_id` | Used for multi-turn conversations |
| `stream` | Strongly recommended to explicitly pass |
| `inc_stream` | Only meaningful when `stream=true` |
| `history` | Optional custom history |
| `custom_querys` | Optional custom input variables |
| `temp_files` | Optional temporary files |
| `temporary_area_id` | Optional temporary area ID |
| `chat_option` | Controls history, progress, retrieval post-processing, cache, etc. |

A set of `chat_option` close to current page behavior:

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

## 6. API Chat cURL Examples

Streaming example:

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

Non-streaming example:

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

## 7. How to Read Response Structure

Core response format:

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
          "text": "Final answer text",
          "cites": {}
        },
        "thinking": "Optional thinking process",
        "skill_process": []
      },
      "middle_answer": []
    }
  }
}
```

Most commonly used fields:
- `conversation_id`
- `user_message_id`
- `assistant_message_id`
- `agent_run_id`
- `message.content.final_answer.answer.text`
- `message.content.final_answer.thinking`
- `message.content.final_answer.skill_process`
- `message.content.middle_answer.progress`

---

## 8. Complete Streaming Output vs Incremental Streaming Output

Observed processed streaming samples indicate: Agent Factory internally standardizes streaming results into **complete snapshots** first, while the frontend can consume them as **incremental deltas**.

### 8.1 What is a Complete Streaming Snapshot

Each chunk can be "the complete response object at the current moment," not just a simple text increment.

A representative complete snapshot format is as follows:

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

### 8.2 What is Incremental Streaming Output

Incremental streaming output is a downstream consumption mode: only rendering or forwarding new parts downstream.

Example:
- Snapshot N: `你好！我是你的AI助手`
- Snapshot N+1: `你好！我是你的AI助手，可以帮你`
- Frontend incremental delta: `，可以帮你`

### 8.3 When is it Considered "Incremental"

Generally understood as incremental under the following conditions:
- `stream=true`
- And downstream/frontend consumes new deltas with `inc_stream=true`, instead of redrawing the full accumulated text each time

### 8.4 Role of Incremental Streaming

Incremental streaming mainly helps:
- Reduce frontend single effective update volume
- Reduce client full segment redrawing, etc.
- Improve frontend rendering performance
- Reduce network transmission data volume

### 8.5 Impact of Not Using Incremental Streaming

If incremental streaming is not used:
- Can still return streaming
- But each event is more likely to carry the current accumulated full text
- Payload will grow as the answer gets longer
- Client rendering cost will also be higher
- Network transmission data volume will also be larger (potentially very large)

This usually doesn't affect correctness but affects efficiency and experience.

---

## 9. Process State and Tool Call Structure

A representative structure is as follows:

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

This means clients can choose:
- Only display final answer
- Display progress bar / step stream + final answer
- Display both simultaneously

A practical rendering rule is:
- Use `final_answer.answer.text` to drive the main answer area
- Use `middle_answer.progress` to drive tool / step / runtime UI

---

## 10. Agent Factory Internal Call Chain

Main route registration:
- `../src/driveradapter/api/httphandler/agenthandler/define.go`

Main behavior of API Chat handler:
- Bind request JSON
- Inject default values (`agent_version=latest`, `executor_version=v2`)
- Read visitor identity and token from context
- Set `CallType=APIChat`
- Strong validation of `agent_key`
- Call unified service `agentSvc.Chat(...)`
- Output SSE or final JSON according to `stream`

Main steps of unified service `../src/domain/service/agentrunsvc/chat.go`:
1. Parse Agent by ID or Key
2. If not API Agent published state, API Chat directly returns 403
3. Load history messages and message indexes
4. Create / update user messages and assistant messages
5. Establish stop channel and session
6. Create conversation session
7. If sandbox is enabled, ensure sandbox session exists
8. Build request sent to agent-executor
9. Call agent-executor
10. Uniformly process streaming results and output standardized chunks

---

## 11. API Documentation Entry Points

Static artifacts in repository:
- `./api/agent-factory.html`
- `./api/agent-factory.json`
- `./api/agent-factory.yaml`

Swagger after starting local service:
- `http://localhost:13020/swagger/index.html`
- `http://localhost:13020/swagger/doc.json`
- `http://localhost:13020/swagger/doc.yaml`

Single Agent dynamic documentation entry (only for Agents published as API):

```http
POST /api/agent-factory/v1/app/{app_key}/api/doc
```

Request body:

```json
{
  "agent_id": "<agent_id>",
  "agent_version": "v0"
}
```

This dynamic interface will automatically rewrite:
- summary = current Agent name
- description = current Agent profile
- request example = current `agent_key` / version / custom input fields

---

## 12. Integration Notes

- Web Chat working doesn't guarantee API Chat availability; API Chat depends on API Agent publication status
- `inc_stream` incremental streaming is only meaningful in streaming mode `stream`
- To align with page behavior as much as possible, reuse `chat_option` from real page requests
- When only caring about final answer, read content of `message.content.final_answer`
- When needing to display tools / process state, read content of `message.content.middle_answer.progress`