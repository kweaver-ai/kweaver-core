# 📒 Cookbook（中文）

KWeaver 的 **场景化操作手册**：每篇是一段「**一目标 / 一组命令 / 一段输出**」的可复制教程。

> 与本目录平级的 [模块文档](../README.md) 是「按子系统组织的参考手册」，cookbook 则按 **「我想做什么事」** 的视角组织，互相引用，不重复。

## 目录

| Recipe | 一句话目标 |
| --- | --- |
| [从 CSV 一键建知识网络](./cookbook_example.md) | 用 `kweaver bkn create-from-csv` 把若干 CSV 一次性变成可查询的 KN |

## 写一篇新 Recipe 的模版

直接复制 [`_TEMPLATE.md`](./_TEMPLATE.md) 改成你的场景；可参考已写好的 [`cookbook_example.md`](./cookbook_example.md)。

文件名建议 `NN-short-slug.md`，每篇统一以下结构：

0. **元数据卡**（顶部 blockquote）：难度、耗时、涉及模块、CLI 版本要求
1. **Goal**：以「**完成后你将拥有：**...」开头，结果导向、可观测
2. **Prerequisites**：版本 / 已登录 / 业务域 / 本篇特有依赖
3. **Steps**：编号步骤 + 可执行命令；步骤多于 1 个时拆 `### 3.x`，进阶/替代路径放 `<details>` 折叠
4. **Expected output**：先一句「**判定成功的依据**」，再贴精简后的真实输出
5. **Troubleshooting**：「现象」列写**用户能直接看到的具体输出 / 报错**
6. **See also**：链回 [模块文档](../README.md)、[`examples/`](../examples/README.md) 与相关 Recipe

> 命令以 **`kweaver`** CLI 优先，必要时给出等价 `curl`；不要把私密 token / 真实业务数据写进示例里。
