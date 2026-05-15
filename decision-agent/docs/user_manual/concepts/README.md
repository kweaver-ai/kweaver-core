# Agent 概念指南

Agent 概念指南用于解释 Decision Agent 文档中反复出现的产品概念、发布逻辑、Agent 模式和术语表达。建议在阅读 API、CLI 或 SDK 接入文档前先浏览本目录。

## 聚合版

- [aggregate.md](./aggregate.md)
- 适合在一个页面中连续阅读概念说明。
- `aggregate.md` 是通过脚本生成的文件，请不要直接修改。
- 如需调整聚合版内容，请修改 [index.md](./index.md) 或对应子章节文件后重新生成。

生成命令：

```bash
make -C docs/user_manual/concepts aggregate
```

注意：<a href="./dolphin-syntax.md" target="_blank">Dolphin 语法文档</a> 是独立引用文档，不会被拼接进聚合版。

## 分文件版

- [index.md](./index.md)
- 适合按主题阅读和维护。
- 子章节包括基础概念、发布逻辑、Agent 模式、运行控制、产品术语与表达。

## Dolphin 语法参考

- <a href="./dolphin-syntax.md" target="_blank">dolphin-syntax.md</a>
- 该文件保持来源文档原貌，后续可随上游 Dolphin 语法文档同步更新。
