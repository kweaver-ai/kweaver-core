# CLI 用户指南

CLI 用户指南面向安装后的 `kweaver` 命令用户，提供两种阅读方式：

## 聚合版

- [aggregate.md](./aggregate.md)
- 适合在一个页面中连续阅读全部 CLI 内容。
- `aggregate.md` 是通过脚本生成的文件，请不要直接修改。
- 如需调整聚合版内容，请修改 [index.md](./index.md) 或对应子章节文件后重新生成。

生成命令：

```bash
make -C docs/user_manual/cli aggregate
```

## 分文件版

- [index.md](./index.md)
- 适合按章节阅读和维护。
- 子章节包括安装运行、快速开始、Agent 生命周期命令、对话/Trace、本地调试流程。

## 可运行示例

- [../examples/cli](../examples/cli/README.md)
- 示例使用 Shell + `kweaver` 命令，并提供独立 Makefile。
- 默认运行 `make quick-check`；`make smoke` 是兼容别名。完整创建、查询、更新、发布、取消发布、删除流程运行 `make flow`；单步操作可运行 `make create`、`make update`、`make publish`、`make unpublish`、`make delete`。
