# CLI 示例

本目录使用安装后的 `kweaver` 命令。默认从本地 SDK 项目安装：

```bash
make install
```

可用目标：

| 目标 | 说明 |
| --- | --- |
| `make install` | 安装本地 `@kweaver-ai/kweaver-sdk` 包到当前示例目录。 |
| `make help` | 验证 `kweaver --help` 和 `kweaver agent --help`。 |
| `make list` | 使用 `kweaver agent list` 查询发布到广场的 Agent。 |
| `make personal-list` | 使用 `kweaver agent personal-list` 查询个人空间 Agent。 |
| `make category-list` | 使用 `kweaver agent category-list` 查询分类。 |
| `make template-list` | 使用 `kweaver agent template-list` 查询模板。 |
| `make detail-examples` | 演示 `get`、`get-by-key`、`template-get`；缺少对应环境变量时自动跳过。 |
| `make chat-examples` | 演示 `chat`、`sessions`、`history`、`trace`；缺少 `AGENT_ID` 或 `CONVERSATION_ID` 时自动跳过，trace 路由不可用时（或 trace-ai 服务不可用）会提示并继续。 |
| `make skill-examples` | 演示 `skill list/add/remove`；缺少 `AGENT_ID` 或 `SKILL_ID` 时自动跳过。 |
| `make update-knowledge-network` | 演示 `update --knowledge-network-id`；缺少 `AGENT_ID` 或 `KN_ID` 时自动跳过。 |
| `make quick-check` | 执行安装、help、列表、个人空间、分类、模板查询。 |
| `make smoke` | `quick-check` 的兼容别名。 |
| `make flow` | 创建测试 Agent、获取详情、发布、取消发布并删除。 |

如需覆盖 SDK 包位置：

```bash
make install KWEAVER_SDK_TS_DIR="<path-to-kweaver-sdk>/packages/typescript"
```

如需直接使用远程 npm 包：

```bash
make install SDK_PACKAGE="@kweaver-ai/kweaver-sdk"
```

`flow` 会修改本地数据，运行前需要设置：

```bash
export KWEAVER_LLM_ID="<your-llm-id>"
export KWEAVER_LLM_NAME="<your-llm-name>"
make flow
```

本目录脚本正文统一使用 `kweaver ...` 命令。默认 `quick-check` 不依赖 `KN_ID`、`SKILL_ID` 等外部资源。
