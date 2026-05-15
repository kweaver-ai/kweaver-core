# Agent 模式

Agent 模式决定 Agent 在运行时如何组织提示词、工具调用、编排逻辑和输出结果。当前主要包括默认模式、ReAct 模式和 Dolphin 模式。

## 默认模式

默认模式背后对应的是 Dolphin 语法的`/explore`模式（探索模式），适合简单、直接的 Agent。会根据用户的`query`、`system_prompt`、`tools`等自动探索执行。

可以简单理解成是一条 Dolphin 语句。

默认模式也承担历史配置的兼容职责。

## ReAct 模式

ReAct 是优先推荐的 Agent 模式。它适合需要工具调用、分步思考和更稳定执行流程的 Agent，同时不要求用户编写 Dolphin 语法。

在配置上，ReAct 模式通常使用 `mode: "react"`，并通过 `react_config` 控制历史、缓存等运行策略。`react_config` 只应在 ReAct 模式下配置。

`react_config`是可选的，如果不配置，和`默认模式`的行为一致。

## Dolphin 模式（多条 Dolphin 语句）

Dolphin 模式适合高级编排场景。它允许通过 Dolphin 语法描述更细粒度的执行流程，例如先调用工具、再进行判断、再组织输出，或者为特定业务场景生成一段自定义编排代码。

在链路上，Agent Factory 保存和发布 Dolphin 相关配置；Agent Executor 在运行 Agent 时读取这些配置；Dolphin SDK 负责解释或执行 Dolphin 编排逻辑。

Dolphin 模式通常涉及 `dolphin`、`pre_dolphin`、`post_dolphin` 等配置。启用 Dolphin 模式时，需要至少存在可执行的 Dolphin 内容；同时 Dolphin 模式不能和 `plan_mode` 同时启用。

## Dolphin 语法文档

Dolphin 语法文档作为独立参考资料维护，不会拼接进聚合文档。阅读 Dolphin 模式时，可以在新页面打开：

<a href="./dolphin-syntax.md" target="_blank">Dolphin 语法文档</a>

该文档保持来源内容不变，后续可随 Dolphin 语法文档同步更新。
