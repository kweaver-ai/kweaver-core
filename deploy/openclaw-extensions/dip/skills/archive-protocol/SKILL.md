---
name: archive-protocol
version: "1.0.1"
description: 全局归档协议。只要任务需要写入任何文件（含 PLAN.md、报告、JSON 等归档物），必须按本技能执行 Session→ARCHIVE_ID、TIMESTAMP、双轨路径（根段须为 archives/）、回读校验与状态回执；WebUI 的 archive_grid 必须用 Markdown 中语言标识为 json 的围栏代码块输出。
metadata:
  {
    "openclaw": {}
  }
---

# 全局归档协议

只要当前任务需要写入任何文件，就必须执行本协议。

归档物包括但不限于：`PLAN.md`、`report.md`、`summary.md`、`result.json`、`notes.md`。

**优先级**：本协议高于任何业务协议。定时计划类任务在 `schedule-plan` 技能中约定，且只能调用本归档规则，不能覆盖。

## 【身份文件】

若任务需要人格/身份设定，只认：

- `SOUL.md`
- `IDENTITY.md`

未确认可读或已注入上下文前，不得声称「已成功读取」。

## 【ARCHIVE_ID 规则】

`ARCHIVE_ID` 的唯一来源是 `session_status` 工具返回结果中的 `Session`。

必须执行：

1. 调用 `session_status`, 并且不允许传递任何参数
2. 读取 `Session`
3. 使用 `:` 作为分隔符切分 `Session`
4. 取切分结果的最后一段作为 `ARCHIVE_ID`

禁止：

- 禁止从其他来源生成或推断 `ARCHIVE_ID`
- 禁止在 `Session` 为空、缺失或无法切分时伪造 `ARCHIVE_ID`
- 若`session_status`工具调用失败，立即重试，不允许捏造`ARCHIVE_ID`
- 若持续失败，立即中止归档并返回：`ARCHIVE_STATUS: BLOCKED` / `ARCHIVE_REASON: invalid Session`

## 【TIMESTAMP 规则】

生成归档文件前，必须执行：`date "+%Y-%m-%d-%H-%M-%S"`

时间格式必须固定为：`YYYY-MM-DD-HH-MM-SS`（禁止带空格或冒号）。

## 【归档路径规则】

路径执行双轨制，严禁混淆：

1. **计划文件（PLAN.md 专属）**：`archives/{ARCHIVE_ID}/PLAN.md`
2. **普通归档物（其他所有生成物）**：`archives/{ARCHIVE_ID}/{TIMESTAMP}/{ORIGIN_NAME}`

**根段必须为 `archives`（禁止多一层 `arch`）**

- 所有上述路径在会话工作区内均为**相对根**，且**必须以**字面量 **`archives/`** 作为归档树的第一段（小写完整单词 `archives`，不是 `arch`）。
- **禁止**在 `archives` 前再拼 `arch/`、`archive/` 或其它目录，形成错误形态如：`arch/archives/...`、`.../arch/archives/{ARCHIVE_ID}/...`。
- **禁止**用 `arch` 当作 `archives` 的缩写；写路径时心中核对：第一段只能是 `archives`，紧接 `{ARCHIVE_ID}`。

## 【写入后强制校验】

任何归档文件写入后，必须立即回读校验。未完成校验前，不得声称「已成功归档」。

校验内容：文件存在、路径正确、内容非空、关键字段存在、内容与当前任务一致；**路径正确**须包含：相对归档树以 `archives/` 起头，**不得**含 `arch/archives/` 等错误多段。

## 【状态回执与 WebUI 卡片渲染】

归档结束后，必须按以下格式返回状态，并提供用于 WebUI 渲染的 JSON。

**JSON 输出形式（强制）**

- 承载 `archive_grid` 的 JSON **必须**以 Markdown **围栏代码块**输出：起始行为三个反引号紧跟语言标识 **`json`**，结束行为单独一行的三个反引号（即标准的 ` ```json` … ` ``` ` 结构）。
- 禁止：裸 JSON（无围栏）、无语言标签的围栏、语言标签不是 `json`（如 `text`、`plaintext`、留空）。
- 每有一个待渲染的归档文件，对应**一个**上述 ` ```json` 代码块（块内为合法 JSON 对象，含 `type: "archive_grid"` 等字段）。

- **失败**：`ARCHIVE_STATUS: BLOCKED` | `ARCHIVE_REASON: <原因>`（失败时若无卡片数据，可不输出 JSON 围栏块；若有结构化回执，仍须遵守上文围栏规则。）
- **成功**：
  1. 输出文本行：`ARCHIVE_STATUS: OK` | `ARCHIVE_ROOT: archives/{ARCHIVE_ID}/`
  2. 按【JSON 输出形式（强制）】输出 `archive_grid`，每一个归档的文件对应一个独立的 ` ```json` 代码块，示例形态如下：

```json
{
  "type": "archive_grid",
  "data": {
    "type": "file",
    "archive_root": "archives/{ARCHIVE_ID}",
    "subpath": "{TIMESTAMP}/{ORIGIN_NAME}",
    "name": "{ORIGIN_NAME}"
  }
}
```