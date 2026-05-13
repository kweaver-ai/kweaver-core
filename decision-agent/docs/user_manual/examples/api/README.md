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
| `make agent-config` | 生成并校验最小配置、复杂配置示例。 |
| `make auxiliary` | 调用分类、产品、临时文件扩展名等辅助接口。 |
| `make chat-examples` | 演示非流式、普通流式、增量流式和 Debug 对话；没有 `AGENT_KEY` 时自动跳过。 |
| `make conversations` | 演示对话列表、消息、已读标记；没有 `AGENT_KEY` 或 `CONVERSATION_ID` 时自动跳过。 |
| `make conversation-session` | 演示会话缓存管理；没有 `CONVERSATION_ID` 时自动跳过。 |
| `make import-export` | 演示导入导出；没有 `AGENT_ID` 时自动跳过，默认会清理导入出的临时 Agent。 |
| `make quick-check` | 执行健康检查、列表查询和低风险辅助接口。 |
| `make smoke` | `quick-check` 的兼容别名。 |
| `make flow` | 创建测试 Agent、获取详情、发布、取消发布并删除。 |

`flow` 会修改本地数据，运行前需要设置可用模型：

```bash
export KWEAVER_LLM_ID="<your-llm-id>"
export KWEAVER_LLM_NAME="<your-llm-name>"
make flow
```

需要已有资源的目标使用环境变量控制，例如 `AGENT_ID`、`AGENT_KEY`、`CONVERSATION_ID`、`KN_ID`。未设置时示例会输出跳过原因并返回成功。
