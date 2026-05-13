# CLI 用户指南

CLI 手册面向已经安装 `@kweaver-ai/kweaver-sdk` 后使用 `kweaver ...` 命令的用户。普通用户不需要进入 SDK 源码目录，也不需要使用源码调试入口。

如果需要先理解 Agent、个人空间、广场、发布和 Agent 模式，请阅读 [Agent 概念指南](../concepts/README.md)。

## 阅读顺序

1. [安装和运行](./install-and-run.md)：完成 CLI 安装、连接配置和认证。
2. [快速开始](./quick-start.md)：跑通创建、详情、发布、对话、清理。
3. [Agent 生命周期命令](./agent-lifecycle.md)：查看 Agent、个人空间、模板、创建、更新、删除。
4. [对话与 Trace](./chat-session-trace.md)：单轮对话、继续对话、流式输出、对话记录和 trace。
5. [本地调试流程](./local-debug-flow.md)：仅面向 SDK/CLI 开发者。

配套可运行示例见 [../examples/cli](../examples/cli/README.md)。示例会先安装本地 SDK 包，再通过 `kweaver ...` 命令执行 `quick-check` 或完整流程。

## 最小环境

```bash
export KWEAVER_BASE_URL="${KWEAVER_BASE_URL:-http://127.0.0.1:13020}"
export KWEAVER_NO_AUTH=1

curl -fsS "$KWEAVER_BASE_URL/health/ready"
kweaver agent --help
```

如果需要指定业务域，可以设置默认值或在单条命令中使用 `-bd`：

```bash
export KWEAVER_BUSINESS_DOMAIN="${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
kweaver config show
```

## 文档索引

- [安装和运行](./install-and-run.md)
- [快速开始](./quick-start.md)
- [Agent 生命周期命令](./agent-lifecycle.md)
- [对话与 Trace](./chat-session-trace.md)
- [本地调试流程](./local-debug-flow.md)
