# 示例与模板（中文版）

## `sample-agent.import.json`

轻量「分析助手」导入模板（无 Dolphin）。`config.skills.tools` 已预绑 **Context Loader 工具箱**中的全部工具（`tool_box_id` 与各 `tool_id` 与仓库 `adp/context-loader/agent-retrieval/docs/release/toolset/context_loader_toolset.adp` 一致，共 9 个工具）。

导入后在 Studio 中把内嵌配置里的 **`REPLACE_WITH_YOUR_LLM_NAME`** 换成你的模型名，并绑定业务知识网络后发布即可。

若你环境中通过 impex 导入工具箱后 **UUID 与仓库 ADP 不一致**，需用 Studio 重新勾选工具，或从 `kweaver call '/api/agent-operator-integration/v1/tool-box/list?...'` 的响应里替换 `tool_box_id` / `tool_id`。

接口：`POST /api/agent-factory/v3/agent-inout/import`（`multipart`：`file` + `import_type`）。

供应链全链路演示仍可使用 `deploy/auto_cofig/agent.json`；`deploy/auto_cofig/auto_config.sh` 为历史一键脚本，将逐步废弃，此处不再扩展。
