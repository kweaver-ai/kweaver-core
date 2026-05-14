# API 示例

本目录使用 Shell + cURL 调用 Decision Agent REST API。默认 no-auth 本地环境：

```bash
make quick-check
```

可用目标：

| 目标 | 说明 |
| --- | --- |
| `make health` | 调用 `/health/ready`。 |
| `make list` | 查询发布到广场的 Agent 列表。 |
| `make agent-config-minimal` | 生成并校验最小配置示例。 |
| `make agent-config-advanced` | 生成并校验复杂配置示例。 |
| `make auxiliary` | 调用分类、产品、临时文件扩展名等辅助接口。 |
| `make chat-examples` | 演示非流式、普通流式、增量流式和 Debug 对话；缺少必需变量时返回错误提示。 |
| `make chat-non-stream` | 单独运行非流式对话。 |
| `make chat-stream` | 单独运行普通流式对话。 |
| `make chat-incremental-stream` | 单独运行增量流式对话。 |
| `make chat-debug` | 单独运行 Debug 对话。 |
| `make chat-terminate` | 终止正在执行的一次对话 run；缺少必需变量时返回错误提示。 |
| `make chat-resume` | 恢复读取断线前仍在运行的流式对话；缺少必需变量时返回错误提示。 |
| `make conversations` | 演示对话列表、消息、已读标记；缺少必需变量时返回错误提示。 |
| `make conversation-session` | 演示会话缓存管理；缺少必需变量时返回错误提示。 |
| `make import-export` | 演示导入导出；默认会清理导入出的临时 Agent。 |
| `make quick-check` | 执行健康检查、列表查询和低风险辅助接口。 |
| `make smoke` | `quick-check` 的兼容别名。 |
| `make flow` | 创建测试 Agent、获取详情、发布、取消发布并删除。 |

`flow` 会修改本地数据，运行前需要设置可用模型：

```bash
export KWEAVER_LLM_ID="<your-llm-id>"
export KWEAVER_LLM_NAME="<your-llm-name>"
make flow
```

需要已有资源的目标使用环境变量或 `.tmp/state.env` 控制，例如 `AGENT_ID`、`AGENT_KEY`、`CONVERSATION_ID`、`KN_ID`。未设置时示例会输出错误原因并返回非 0 状态。

目录结构按能力拆分，每个子目录都有自己的 Makefile；例如 `make -C chat incremental-stream` 会执行 `chat/incremental-stream.sh`。
