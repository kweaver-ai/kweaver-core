# 快速开始

以下步骤假设 KWeaver Core 已 [部署](installation/deploy.md) 并完成 [验证](installation/verify.md)。

## 1. 安装 `kweaver` CLI

```bash
npm install -g @kweaver-ai/kweaver-sdk
kweaver --help
```

## 2. 登录

```bash
kweaver auth login https://<访问地址> -k
```

`<访问地址>` 与安装时 `--access_address` 一致。最小化安装下认证策略可能较宽松，以 CLI 提示为准。

## 3. 查看业务知识网络（BKN）

```bash
kweaver bkn list
```

建模、实例与 API 详见 [BKN 引擎](bkn.md)。

## 4. 数据面（VEGA）

配置好连接与视图后，可通过 VEGA 相关 API 或 SDK 查询。详见 [VEGA 引擎](vega.md)。

## 5. 运行 Decision Agent

创建或选择 Agent 模板，通过 CLI/SDK 对话或触发运行。详见 [Decision Agent](decision-agent.md)。

## 6. 使用 Trace AI 观测

在链路追踪开启时，可通过 Trace AI 或 observability 查询服务查看与智能体运行关联的 Span。详见 [Trace AI](trace-ai.md)。

## 接下来读什么

| 目标 | 文档 |
| --- | --- |
| 业务知识建模 | [bkn.md](bkn.md) |
| 统一数据访问 | [vega.md](vega.md) |
| 流程编排 | [dataflow.md](dataflow.md) |
| 智能体工具 | [execution-factory.md](execution-factory.md) |
| 上下文组装 | [context-loader.md](context-loader.md) |
| 安全与治理 | [isf.md](isf.md) |

各模块文档在同一文件中提供 **CLI**、**Python SDK**、**TypeScript SDK** 与 **curl** 示例。
