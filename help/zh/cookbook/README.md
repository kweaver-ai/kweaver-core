# 📒 Cookbook（中文）

KWeaver 的 **场景化操作手册**：每篇是一段「**一目标 / 一组命令 / 一段输出**」的可复制教程。

> 与本目录平级的 [模块文档](../README.md) 是「按子系统组织的参考手册」，cookbook 则按 **「我想做什么事」** 的视角组织，互相引用，不重复。

## 目录

| Recipe | 一句话目标 |
| --- | --- |
| [从 CSV 一键建知识网络](./cookbook_example.md) | 用 `kweaver bkn create-from-csv` 把若干 CSV 一次性变成可查询的 KN |

## 写一篇新 Recipe 的模版

每篇文件名建议 `NN-short-slug.md`，统一 6 段：

1. **Goal**：一句话说"做完会得到什么"
2. **Prerequisites**：版本 / 已登录 / 业务域 等
3. **Steps**：编号步骤 + 可执行命令
4. **Expected output**：贴一段精简后的真实输出
5. **Troubleshooting**：常见 401 / 403 / 业务域错配等
6. **See also**：链回 [模块文档](../README.md) 与 [`examples/`](../examples/README.md) 的相关条目

> 命令以 **`kweaver`** CLI 优先，必要时给出等价 `curl`；不要把私密 token / 真实业务数据写进示例里。
