# CLI 示例

本目录使用安装后的 `kweaver` 命令。默认按 `.env` 中的 `KWEAVER_SDK_PACKAGE_SOURCE` 选择远程 npm 包或本地 SDK 项目包，并安装到当前目录的 `node_modules`：

```bash
make install
```

可用目标：

| 目标 | 说明 |
| --- | --- |
| `make install` | 按 `KWEAVER_SDK_PACKAGE_SOURCE` 安装 `@kweaver-ai/kweaver-sdk` 到当前示例目录。 |
| `make help` | 验证 `kweaver --help` 和 `kweaver agent --help`。 |
| `make list` | 使用 `kweaver agent list` 查询发布到广场的 Agent。 |
| `make personal-list` | 使用 `kweaver agent personal-list` 查询个人空间 Agent。 |
| `make category-list` | 使用 `kweaver agent category-list` 查询分类。 |
| `make template-list` | 使用 `kweaver agent template-list` 查询模板。 |
| `make create` | 创建测试 Agent，并把 `AGENT_ID` 写入 `../.tmp/state.env`。 |
| `make get` | 使用 `kweaver agent get` 查询详情；优先读取环境变量，其次读取 `../.tmp/state.env`。 |
| `make update` | 更新当前 Agent 的名称和描述；如果设置 `KN_ID`，同时演示知识网络绑定。 |
| `make publish` | 发布当前 Agent。 |
| `make unpublish` | 取消发布当前 Agent。 |
| `make delete` | 删除当前 Agent，并清理 state 中的 Agent 信息。 |
| `make get-by-key` | 使用 `kweaver agent get-by-key` 查询详情；没有 `AGENT_KEY` 时返回错误提示。 |
| `make template-get` | 使用 `kweaver agent template-get` 查询模板详情；没有 `TEMPLATE_ID` 时返回错误提示。 |
| `make detail-examples` | 演示 `get`、`get-by-key`、`template-get`。 |
| `make chat-examples` | 演示单轮对话、多轮对话、`sessions`、`history`、`trace`；缺少必需变量时返回错误提示，trace 路由不可用时（或 trace-ai 服务不可用）会提示并继续。 |
| `make skill-examples` | 演示 `skill list/add/remove`；缺少 `SKILL_ID` 时返回错误提示。 |
| `make update-knowledge-network` | 演示 `update --knowledge-network-id`；缺少 `KN_ID` 时返回错误提示。 |
| `make quick-check` | 执行安装、help、列表、个人空间、分类、模板查询。 |
| `make smoke` | `quick-check` 的兼容别名。 |
| `make flow` | 组合执行创建、获取详情、更新、发布、取消发布并删除。 |

使用远程 npm 包：

```bash
make install KWEAVER_SDK_PACKAGE_SOURCE=remote
```

使用本地 SDK 项目包：

```bash
make install KWEAVER_SDK_PACKAGE_SOURCE=local KWEAVER_SDK_TS_DIR="<path-to-kweaver-sdk>/packages/typescript"
```

两种方式都只安装到当前示例目录，不做全局安装。`make install` 会先卸载并清理旧的 `@kweaver-ai/kweaver-sdk`，再安装当前选择的包来源。

`create` 和 `flow` 会修改本地数据，运行前需要可用模型；`.env` 已提供默认值，可按需修改：

```bash
export KWEAVER_LLM_ID="<your-llm-id>"
export KWEAVER_LLM_NAME="<your-llm-name>"
make create
make get
make publish
make unpublish
make delete
```

CLI 创建 Agent 时会读取 [agents/agent-config.json](./agents/agent-config.json)。该文件已加入 `.gitignore`，适合保存本地可运行配置；仓库中提供 [agents/agent-config.json.example](./agents/agent-config.json.example) 作为模板。首次运行 `make create` 时，如果本地配置不存在，脚本会自动从模板初始化，也可以手工执行：

```bash
cp agents/agent-config.json.example agents/agent-config.json
```

如需临时使用其他配置文件，可以设置：

```bash
AGENT_CONFIG_FILE="<path-to-agent-config.json>" make create
```

本目录脚本正文统一使用 `kweaver ...` 命令。默认 `quick-check` 不依赖 `AGENT_ID`、`KN_ID`、`SKILL_ID` 等外部资源。

当前 CLI 尚未暴露对话重命名、标记已读、删除、流式续连、终止执行等独立命令。需要这些能力时，请参考 REST API 示例和 Cookbook 中的 API 版本。
