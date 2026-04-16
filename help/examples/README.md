# 示例与模板

## `sample-agent.import.json`

轻量「分析助手」导入模板（无 Dolphin、`skills.tools` 为空）。导入后在 Studio 中把内嵌配置里的 **`REPLACE_WITH_YOUR_LLM_NAME`** 换成你的模型名，绑定业务知识网络，在技能中挂载 **Context Loader 工具箱**，再发布。

接口：`POST /api/agent-factory/v3/agent-inout/import`（`multipart`：`file` + `import_type`）。

供应链全链路演示仍可使用 `deploy/auto_cofig/agent.json`；`deploy/auto_cofig/auto_config.sh` 为历史一键脚本，将逐步废弃，此处不再扩展。
