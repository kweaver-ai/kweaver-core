# 用 CLI 创建合同摘要 Agent

> - **难度**：⭐ 入门
> - **耗时**：约 15 分钟
> - **涉及模块**：`kweaver agent`、`chat`、`sessions`
> - **CLI 版本**：`kweaver >= 0.6`

## 1. Goal（目标）

**完成后你将拥有：** 一个通过 `kweaver` 命令创建、发布并调用的合同摘要 Agent。

## 2. Prerequisites（前置条件）

- 已安装 CLI：`make -C docs/user_manual/examples/cli install`。
- 已设置本地连接：`KWEAVER_BASE_URL`、`KWEAVER_NO_AUTH=1`。
- 已设置 `KWEAVER_LLM_ID` 和 `KWEAVER_LLM_NAME`。

## 3. Steps（操作步骤）

### 3.1 验证 CLI

```bash
make -C docs/user_manual/examples/cli help
```

### 3.2 创建、发布、删除完整流程

```bash
make -C docs/user_manual/examples/cli flow
```

### 3.3 对话与对话列表

```bash
AGENT_ID=<agent-id> make -C docs/user_manual/examples/cli/chat chat
AGENT_ID=<agent-id> make -C docs/user_manual/examples/cli/chat sessions
```

## 4. Expected output（期望输出）

> **判定成功的依据**：`kweaver agent create` 返回 `id`，`kweaver agent chat` 输出回答，并提示可用 `--conversation-id` 继续对话。

```jsonc
{
  "id": "01..."
}
```

## 5. Troubleshooting（常见问题）

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| `kweaver command is required` | 还没安装示例依赖 | 运行 `make -C docs/user_manual/examples/cli install`。 |
| `Missing value for conversation-id` | 参数缺值 | 使用 `--conversation-id <conversation-id>`。 |
| 对话列表为空 | 还没有调用过 Agent | 先运行 `chat`。 |

## 6. See also（延伸阅读）

- 场景说明：[创建合同摘要 Agent](../scenario/01-contract-summary-basic.md)
- CLI 示例：[examples/cli](../../user_manual/examples/cli/README.md)
- 用户手册：[CLI 快速开始](../../user_manual/cli/quick-start.md)

