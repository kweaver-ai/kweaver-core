---
name: schedule-plan
version: "1.0.0"
description: 定时计划协议。仅当用户请求创建定时计划、定时任务、提醒、自动化安排、周期或延迟任务时生效；包含 ORA 拆解、PLAN.md 持久化、用户书面确认 PLAN 后再创建定时任务、已落地计划的修改须同步更新 PLAN.md、任务消息首要指令与 PLAN 模板。须与 archive-protocol 技能一并遵守。
metadata:
  {
    "openclaw": {}
  }
---

# 定时计划协议

**仅当**用户请求「创建定时计划 / 定时任务 / 提醒 / 自动化安排 / 周期任务 / 延迟任务」时生效。

**与归档的关系**：定时计划只能调用 `archive-protocol` 技能中的全局归档规则，不能覆盖归档规则（含 `ARCHIVE_ID`、`TIMESTAMP`、路径双轨制与写入校验）。

所有涉及 `PLAN.md` 或其它文件的写入，须先按 `archive-protocol` 技能执行。

## 【先确认 PLAN，再创建定时任务】

生成并写入 `PLAN.md` 后，**不得**在同一轮对话中、也**不得**在用户尚未明确同意的情况下，调用任何「创建定时任务 / Cron / 调度 / 提醒注册」等能力。

必须执行的交互顺序：

1. 将 `PLAN.md` 中的关键内容摘要给用户（或明确告知路径并请用户阅读），并**主动询问**：是否确认按此执行、是否有修改意见。
2. 若用户提出修改：先按意见更新 `PLAN.md`（遵守 `archive-protocol`），再重复上一步，直至用户满意。
3. 仅在用户**明确确认**后（例如明确表示「确认」「可以创建定时任务了」「按这个建 Cron」等），才创建定时任务；此时任务 message 仍须遵守下文【任务消息首要指令】。

禁止：在用户只说「做个定时任务」但未确认 PLAN 正文前，就预先注册调度。

## 【已创建计划的修改】

若定时任务已存在，或用户要求变更已写入 `PLAN.md`、已落地的计划（含频率、步骤、判定标准、归档方式等），**必须**先同步更新 `archives/{ARCHIVE_ID}/PLAN.md` 中与该 `Task_ID` 对应的条目（遵守 `archive-protocol`），再按需调整调度参数或任务 message。禁止仅在对话中达成新口径而 `PLAN.md` 仍为旧内容。

## 【需求理解规则】

当用户需求较笼统时，必须：直接给出一个预设基准方案，并提供 2-3 个结构化选项，最后追加：`其他：[由用户补充]`。禁止让用户做开放式命题作文。

## 【ORA 拆解】

在用户确认计划细节后，必须展示：

- **O (Objective)**: 解决什么痛点，如何体现 `SOUL.md` 价值观。
- **R (Result)**: 成功判定标准、交付物、可见成果。
- **A (Action)**: 触发频率、执行逻辑、失败处理、输出物归档方式。

此处的「确认」指 ORA 维度对齐；`PLAN.md` 写入后的**全文确认与改稿**见【先确认 PLAN，再创建定时任务】，通过后方可创建定时任务。

## 【任务消息首要指令】

在用户已确认 `PLAN.md` 并**实际创建定时任务**时，任务 message 中的首要指令必须是：

`从路径 archives/{ARCHIVE_ID}/PLAN.md 中读取 Task_ID: {{TASK_UUID}} 的详细计划并按步骤执行`

（`ARCHIVE_ID` 的来源与路径规则见 `archive-protocol` 技能。）

## 【PLAN.md 持久化】

必须写入路径：`archives/{ARCHIVE_ID}/PLAN.md`

必须包含：`Task_ID`, `Role_Context`, `Archive_Path`, `Status`, `Objective`, `Result`, `Action_Steps`。

`Action_Steps` 的第 1 条必须是：

`1. [读取本文件指令]：从路径 archives/{ARCHIVE_ID}/PLAN.md 中读取 Task_ID: {{TASK_UUID}} 的详细计划并按步骤执行`

【Cron】
创建 Cron 任务时，强制参数：
1. `sessionTarget` 必须是 "isolated"
2. `payload.kind` 必须是 "agentTurn"

## 【PLAN.md 模板】

### [Task_ID: {{TASK_UUID}}] {{Task_Name}}

- **Role_Context**: {{Current_Persona}} (Aligned with `SOUL.md` & `IDENTITY.md`)
- **Archive_Path**: archives/{{ARCHIVE_ID}}/PLAN.md
- **Status**: Active

#### ORA Detail

- **Objective**: {{O_content}}
- **Result**: {{R_content}}
- **Action_Steps**:
  1. **[读取本文件指令]**：从路径 `archives/{{ARCHIVE_ID}}/PLAN.md` 中读取 `Task_ID: {{TASK_UUID}}` 的详细计划并按步骤执行。
  2. **[上下文校验]**：确认当前上下文已对齐 `SOUL.md` 与 `IDENTITY.md`。
  3. **[执行动作]**：根据读取到的详细步骤执行业务逻辑。
  4. **[归档校验]**：验证输出物已存入 `archives/{{ARCHIVE_ID}}/{{TIMESTAMP}}/` 目录。

## 【执行顺序（计划任务）】

1. 按 `archive-protocol` 调用 `session_status` 提取 `ARCHIVE_ID`，生成 `TIMESTAMP`。
2. 与用户对齐并确认 ORA 要点后，将计划写入 `archives/{ARCHIVE_ID}/PLAN.md`（遵守该技能的双轨规则）。**若为已有计划的变更**：先完成【已创建计划的修改】中的 `PLAN.md` 同步更新，再继续后续步骤。
3. **暂停创建定时任务**：向用户展示或指引阅读 PLAN，征询确认与修改意见；按需修订 `PLAN.md` 并再次征询，直至用户明确确认。
4. 用户明确确认后，再创建定时任务；任务 message 的首条指令须为「读取 `PLAN.md` 中对应 `Task_ID`」类表述（见【任务消息首要指令】）。
5. 任何其它产出物写入 `archives/{ARCHIVE_ID}/{TIMESTAMP}/` 并做回读校验。
