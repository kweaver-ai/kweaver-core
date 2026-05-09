# <Recipe 标题：动词开头，能一句话说清产出>

> - **难度**：⭐ 入门 / ⭐⭐ 进阶 / ⭐⭐⭐ 专家
> - **耗时**：约 N 分钟
> - **涉及模块**：`<bkn|datasource|dataflow|...>`
> - **CLI 版本**：`kweaver >= x.y`

## 1. Goal（目标）

> 用「**完成后你将拥有：**...」开头，结果导向、可观测，避免重复标题。

## 2. Prerequisites（前置条件）

- 已通过 `kweaver auth login <平台地址>` 登录。
- 业务域：`kweaver config show` 确认；不对就 `kweaver config set-bd <uuid>`。
- <列出本 Recipe 特有的依赖：数据源 / 文件 / 已有 KN 等>

## 3. Steps（操作步骤）

> 步骤多于 1 个时用 `### 3.x 标题` 拆分；每个 `### 3.x` 内最多一段说明 + 一个代码块；再多就拆 `### 3.x.y`。
> 「替代/进阶」路径放 `<details>` 折叠，避免打断主线。

### 3.1 <步骤名>

```bash
# 必要的 kweaver CLI 命令
```

### 3.2 <步骤名>

```bash
# ...
```

如有可调参数较多，加一张速查表：

| 参数 | 是否必填 | 说明 |
| --- | --- | --- |
| `<param>` | 是/否 | <一句话> |

## 4. Expected output（期望输出）

> **判定成功的依据**：<一行明确的可观察结果，例如 `total > 0` 且 `datas[0]` 含 X 字段>

```jsonc
{
  // 贴一段精简后的真实输出；删掉敏感字段
}
```

## 5. Troubleshooting（常见问题）

> 「现象」列写**用户能直接看到的具体输出/报错**，便于复制搜索。

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| `<具体报错或输出>` | <原因> | <一行命令或操作> |

## 6. See also（延伸阅读）

- 参考：[<手册条目>](../manual/<x>.md) · [快速开始](../quick-start.md)
- 完整示例项目：[`examples/<NN-slug>/`](../../../examples/<NN-slug>/)
- 相关 Recipe：[<另一篇 cookbook>](./<other-recipe>.md)
