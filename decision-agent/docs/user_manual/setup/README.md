# 安装与部署准备

本目录用于沉淀 Decision Agent 相关的安装、环境准备和后续部署说明。当前先提供 TypeScript SDK 本地包安装指南，后续 Decision Agent 部署、启动、环境检查等内容也会放在这里。

## 聚合版

- [aggregate.md](./aggregate.md)
- 适合在一个页面中连续阅读安装与准备说明。
- `aggregate.md` 是通过脚本生成的文件，请不要直接修改。
- 如需调整聚合版内容，请修改 [index.md](./index.md) 或对应子章节文件后重新生成。

生成命令：

```bash
make -C docs/user_manual/setup aggregate
```

## 分文件版

- [index.md](./index.md)
- 适合按主题阅读和维护。
- 当前子章节包括 TypeScript SDK 本地包安装。

