# TypeScript SDK 用户指南

SDK 文档提供两种阅读方式：

## 聚合版

- [aggregate.md](./aggregate.md)
- 适合在一个页面中连续阅读全部 SDK 内容。
- `aggregate.md` 是通过脚本生成的文件，请不要直接修改。
- 如需调整聚合版内容，请修改 [index.md](./index.md) 或对应子章节文件后重新生成。

生成命令：

```bash
make -C docs/user_manual/sdk/typescript aggregate
```

## 分文件版

- [index.md](./index.md)
- 适合按章节阅读和维护。
- 子章节包括安装运行、Client 初始化、Agents、对话与流式输出、对话。

## 可运行示例

- [../../examples/sdk/typescript](../../examples/sdk/typescript/README.md)
- 示例使用 TypeScript + `@kweaver-ai/kweaver-sdk/kweaver`，并提供独立 Makefile。
- 默认运行 `make quick-check`；`make smoke` 是兼容别名。完整创建、发布、取消发布、删除流程运行 `make flow`。
