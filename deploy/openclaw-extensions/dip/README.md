# DIP OpenClaw 插件

`dip` 是一个 OpenClaw 网关扩展，当前主要提供三类能力：

1. **Agent skills**：可发现技能列表、按 agent 读写技能绑定、以及通过 Gateway 上传 `.skill`（zip）安装到仓库 `skills/`。
2. **工作区 archives**：HTTP 读取 `archives/`，以及写文件后的归档路径补齐。
3. **内置技能包**：插件目录下附带若干 skill 文档，参与发现逻辑。
4. **工作区临时上传**：将上传文件写入 `workspace/tmp`（可按会话分目录）。

插件自身打包了 2 个 skills：

- `archive-protocol`
- `schedule-plan`

## 当前实现的能力

### 1. Skills 管理

#### CLI

```text
/skills-manage [list | enable <name> | disable <name>]
```

- `list`：返回当前可发现的 skill 名称，以及全局配置中的启用状态（`skills.entries.<name>.enabled`）。
- `enable <name>` / `disable <name>`：写入 `openclaw` 配置里的 `skills.entries.<name>.enabled`。

#### HTTP

```text
GET    /v1/config/agents/skills
GET    /v1/config/agents/skills?agentId=<id>
GET    /v1/config/agents/skills/<name>/tree
GET    /v1/config/agents/skills/<name>/content?path=<relative-file-path>
GET    /v1/config/agents/skills/<name>/download?path=<relative-file-path>
POST   /v1/config/agents/skills
PUT    /v1/config/agents/skills
POST   /v1/config/agents/skills/install
DELETE /v1/config/agents/skills/<name>
```

**查询与更新 agent 技能绑定**

- `GET` 无 `agentId`：返回当前可发现的 skill id 列表（JSON：`{ "skills": string[] }`）。
- `GET` 带 `agentId`：
  - 若该 agent 在配置中显式设置了 `skills`，直接返回该数组；
  - 若未显式设置，则按发现逻辑返回该 agent 可见的 skills（JSON：`{ "agentId", "skills" }`）。
- `POST` / `PUT`：请求体为 JSON，需包含 `agentId`（string）与 `skills`（string[]），整组写回 `agents.list[].skills`（JSON：`{ "success", "agentId", "skills" }`）。

**读取技能目录树**

- `GET /v1/config/agents/skills/<name>/tree`
- 返回技能目录下的完整文件树（JSON：`{ "name", "entries" }`）。
- `entries[].type` 为 `file` 或 `directory`；目录节点额外带 `children`。

**预览技能文件**

- `GET /v1/config/agents/skills/<name>/content?path=<relative-file-path>`
- `path` 是技能目录内的相对路径，例如 `SKILL.md`、`docs/guide.md`；不传时默认 `SKILL.md`。
- 仅允许读取普通文件；路径穿越和目录读取都会返回 `400`。
- 成功返回：`{ "name", "path", "content", "bytes", "truncated" }`。
- 当前文本预览上限为 `1MB`，超出部分不返回，并标记 `truncated=true`。

**下载技能文件**

- `GET /v1/config/agents/skills/<name>/download?path=<relative-file-path>`
- `path` 是技能目录内的相对路径；不传时默认 `SKILL.md`。
- 仅允许读取普通文件；路径穿越和目录读取都会返回 `400`。
- 成功时返回原始文件字节流，并设置 `Content-Type` 与 `Content-Disposition: attachment`。

**安装 `.skill` 包（zip）**

- `POST /v1/config/agents/skills/install`
- 查询参数：`overwrite=true`（可选）。为 `true` 时，若 `skills/<name>/` 已存在则先删除再写入。
- 查询参数：`name=<slug>`（可选）。当 zip **根目录**含 **`SKILL.md`**（扁平布局，可多顶层文件/目录）时**必填**，用于指定安装目录名；若 zip 为**单一顶层目录** `<name>/` 且内含 `<name>/SKILL.md`，则技能 id 取自目录名，**不需要** `name`。
- 请求体：**原始 zip 字节**（推荐 `Content-Type: application/zip`）。
- 成功响应示例：`{ "name": "<id>", "skillPath": "<绝对路径>" }`（路径为网关进程所在机器上的落盘路径）。
- 包内结构（二选一）：**嵌套** — zip 根下仅一个顶层目录 `<name>/`，且含 `<name>/SKILL.md`；**扁平** — zip 根下含 `SKILL.md`，且通过 `name` 指定安装名。目录名需符合常见 slug 字符集。
- 解压**不引入 npm 压缩库**，通过宿主环境的 `tar -xf` 或 `unzip` 执行。运行 OpenClaw 的进程需在 `PATH` 上能调用其中之一（多数 Linux/macOS/Windows 10+ 自带可读 zip 的 `tar`；极简容器可能需自行安装 `tar` 或 `unzip`，二者并非在所有环境都保证存在）。

**卸载技能（仓库 `skills/`）**

- `DELETE /v1/config/agents/skills/<slug>`（路径参数为技能 id）。
- 仅删除 **`{repoRoot}/skills/<name>/`**（或同名 `*.skill` 条目）若存在；若该 id **仅**存在于插件内置 `extensions/dip/skills/`，返回 **403**（`BUNDLED`），不删除内置包。Studio 仅会在 `skills.status` 条目 `source === "openclaw-managed"` 且目录位于 `~/.openclaw/skills/<name>/` 时调用此接口。
- 成功：`200` + `{ "name": "<id>" }`。

### 2. Skills 发现

`discoverSkillNames` 统一使用 OpenClaw 原生 SDK 的 `listSkillCommandsForAgents` 进行发现，
并对返回的 `skillName` 做去重和字典序排序。不再通过读取仓库或插件目录来列举技能。

### 3. Archives 访问

插件注册了前缀路由：

```text
GET /v1/archives...
```

当前支持的访问方式：

- 直接读取当前工作区下的 `archives/`
- 通过 `?agent=<agentId>` 切换到对应 agent 的 `workspace` 下读取 `archives/`
- 通过 `?session=<sessionKey或sessionId>` 将会话标识归一化后，直接定位到对应归档目录

当前返回行为：

- 目标是目录时：返回 JSON，包含 `path` 和 `contents`
- 目标是文件时：按扩展名返回常见 MIME 类型并直接流式输出文件内容
- 不存在返回 `404`
- 路径穿越被拦截时返回 `403`

### 4. 工作区临时上传

插件注册了上传路由：

```text
POST /v1/workspace/tmp/upload
```

请求方式：

- 请求体支持：
  - **`multipart/form-data`**（推荐，字段名固定为 `file`，可携带原始文件名）
  - **原始文件字节**（binary body，兼容模式）
- 可选查询参数：
  - `agent=<agentId>`：上传到指定 agent 的 workspace（未传时使用当前 workspace）。
  - `session=<sessionId|sessionKey>`：按会话在 `tmp/<session>/` 下分目录存放。

落盘规则：

- 基础目录：`{workspace}/tmp/`
- 会话目录：`{workspace}/tmp/{normalizedSession}/`
- 文件名：`{basename}_{sha256前12位哈希}{ext}`（未提供文件名时默认 `upload_<hash>.bin`）

成功响应示例：

```json
{
  "name": "report_2cf24dba5fb0.pdf",
  "path": "tmp/chat-1/report_2cf24dba5fb0.pdf",
  "absolutePath": "/abs/workspace/tmp/chat-1/report_2cf24dba5fb0.pdf",
  "bytes": 12840
}
```

### 4. 写文件后的归档补齐

插件监听 `after_tool_call`，只在以下工具名命中时生效：

- 名称包含 `write`
- 名称包含 `edit`
- 名称包含 `replace`

当前只处理 `event.params.path`、`file` 或 `filename` 中给出的单个文件路径，并且要求：

- 文件位于当前工作区内
- 文件真实存在
- 目标是普通文件

归档规则是当前代码里真正实现的规则：

- 如果文件名是 `plan.md`，补齐到 `archives/{sessionId}/PLAN.md`
- 其他文件补齐到 `archives/{sessionId}/{YYYY-MM-DD-HH-mm-ss}/{sanitizedFileName}`

其中：

- `sessionId` 来自 `ctx.sessionKey` 最后一段，取不到时回退到 `ctx.sessionId`
- 文件名会被标准化为小写、空白转 `_`、移除非法字符，扩展名保留为小写
- 如果原路径已经符合上述规则，则不会重复复制
- 当前实现是“复制到合规归档路径”，不会移动原文件

## 当前内置 skills

### `archive-protocol`

这是一个归档约束 skill，文档中要求在涉及文件写入时遵守：

- 从 `session_status` 的 `sessionKey` 提取 `ARCHIVE_ID`
- 生成固定格式的时间戳
- `PLAN.md` 与普通产物走不同归档路径
- 写入后必须回读校验
- 输出归档状态与用于 WebUI 的卡片 JSON

注意：这些是该 skill 文档定义的操作协议，不是插件代码主动替 agent 执行的完整流程。插件代码当前只实现了上一节描述的“写文件后归档补齐”。

### `schedule-plan`

这是一个定时任务规划协议 skill，当前文档要求：

- 仅在创建定时任务、提醒、自动化安排等场景生效
- 先生成并归档 `PLAN.md`
- 用户明确确认 `PLAN.md` 后，才允许创建定时任务
- 计划中需要包含 ORA（Objective / Result / Action）结构
- 创建的任务消息首条指令应先读取 `archives/{ARCHIVE_ID}/PLAN.md`

注意：插件代码当前没有直接提供 Cron、提醒或自动化创建接口；这里只打包了该 skill 文档。

## 安装与启用

将本目录部署到 OpenClaw 扩展目录后，在配置中启用插件：

```json
{
  "plugins": {
    "entries": {
      "dip": {
        "enabled": true
      }
    }
  }
}
```

插件元数据定义在 `openclaw.plugin.json` 中，当前会暴露：

- 插件 id：`dip`
- 插件扩展入口：`./index.ts`
- 插件内置 skills 目录：`./skills`

若使用 `POST /v1/config/agents/skills/install`，请确保网关进程所在环境可调用 `tar` 或 `unzip`（见上文「安装 `.skill` 包」）。
