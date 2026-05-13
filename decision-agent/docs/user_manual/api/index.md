# Decision Agent REST 接入指南

Decision Agent REST 接入指南适合需要查看底层 REST 协议、排障或做自定义集成的开发者。业务接入优先使用 SDK；当前重点是 TypeScript SDK，后续会继续补充 Python 和 Go SDK 文档。示例默认从 `decision-agent` 目录运行。

如果需要先统一产品概念和术语，请阅读 [Agent 概念指南](../concepts/README.md)。

## 整体流程

一个典型接入场景可以这样串起来：先创建一个 Agent，并用最小化 `config` 跑通保存；随后按业务需要更新 `config`，例如绑定知识网络或调整 LLM；确认配置可用后发布到广场；发布后通过广场详情接口获取 `agent_key`；业务系统使用 `agent_key` 发起非流式、普通流式或增量流式对话；遇到配置页调试或排障场景时使用 Debug 对话；本地验证完成后可删除测试 Agent。

推荐阅读顺序：

1. [Agent 生命周期](./agent-lifecycle.md)：创建、查看、更新和删除 Agent。
2. [Agent 配置](./agent-config.md) 与 [Agent config 示例](./agent-config-examples.md)：准备可创建、可更新的 `config`。
3. [发布与广场](./publish-and-market.md)：发布 Agent 并解析 `agent_key`。
4. [对话](./chat-and-conversation.md) 与 [增量流式](./incremental-streaming.md)：完成实际调用。
5. [Debug 对话](./debug-chat.md) 和 [对话、会话与执行](./conversation-session-run.md)：排障和理解运行期概念。

配套可运行示例见 [../examples/api](../examples/api/README.md)。示例使用 Shell + cURL，并提供 `make quick-check` 与 `make flow`。

## 概念关系

| 英文字段/概念 | 中文名称 | 说明 |
| --- | --- | --- |
| `conversation` / `conversation_id` | 对话 | 面向用户的多轮上下文容器。继续同一个 `conversation_id` 会复用历史上下文。 |
| `session` / `conversation_session_id` | 会话 | 运行期会话，主要用于缓存、续连和临时状态。普通 REST 对接通常可以忽略。 |
| `run` / `agent_run_id` | 一次执行 | 一次 Agent 接口调用或一次 Agent 执行过程，用于追踪、中断、恢复和观测。 |

一个对话可以对应多个会话，一个会话可以包含一个或多个 run。文档后续统一使用“对话”对应 `conversation`，“会话”对应 `session`，“一次执行”对应 `run`。

## 最小化 no-auth 环境

本地 no-auth 调试时，只需要服务地址和基础请求函数：

```bash
cd <path-to-decision-agent>
export AF_BASE_URL="${AF_BASE_URL:-http://127.0.0.1:13020}"

af_curl() {
  curl -fsS -H "X-Language: zh-cn" "$@"
}

af_curl "$AF_BASE_URL/health/ready"
```

后续 API 示例都假设当前 shell 已定义 `af_curl`。

## 完整环境

如果部署开启鉴权或业务域校验，可以使用完整环境：

```bash
cd <path-to-decision-agent>
export AF_BASE_URL="${AF_BASE_URL:-http://127.0.0.1:13020}"
export AF_BD="${AF_BD:-bd_public}"
export AF_TOKEN="${AF_TOKEN:-${KWEAVER_TOKEN:-__NO_AUTH__}}"

af_curl() {
  if [ "$AF_TOKEN" = "__NO_AUTH__" ]; then
    curl -fsS -H "X-Business-Domain: $AF_BD" -H "X-Language: zh-cn" "$@"
  else
    curl -fsS \
      -H "Authorization: Bearer $AF_TOKEN" \
      -H "Token: $AF_TOKEN" \
      -H "X-Business-Domain: $AF_BD" \
      -H "X-Language: zh-cn" \
      "$@"
  fi
}

af_curl "$AF_BASE_URL/health/ready"
```

## 文档索引

- [Agent 生命周期](./agent-lifecycle.md)
- [Agent 配置](./agent-config.md)
- [Agent config 示例](./agent-config-examples.md)
- [发布与广场](./publish-and-market.md)
- [对话](./chat-and-conversation.md)
- [增量流式](./incremental-streaming.md)
- [Debug 对话](./debug-chat.md)
- [对话、会话与执行](./conversation-session-run.md)
- [导入导出与辅助接口](./import-export-and-auxiliary.md)

## API 版本

| 前缀 | 主要用途 |
| --- | --- |
| `/api/agent-factory/v3` | Agent 管理、发布、个人空间、广场、权限等管理接口。 |
| `/api/agent-factory/v1` | 对话执行、会话管理、run 管理等运行接口。 |
