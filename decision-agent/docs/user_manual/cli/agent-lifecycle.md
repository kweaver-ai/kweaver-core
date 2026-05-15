# Agent 生命周期命令

请先完成 [CLI 安装和运行](./install-and-run.md)。本页覆盖 Agent 管理、个人空间、模板、发布和技能绑定相关子命令。完整可运行脚本见 [../examples/cli](../examples/cli/README.md)。

## `kweaver agent list`

用途：查询已发布到广场的 Agent 列表。

```bash
kweaver agent list --name "" --limit 10 --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `--name <text>` | 否 | 空字符串 | 按名称过滤。 |
| `--offset <n>` | 否 | `0` | 分页偏移量。小于 0 或非法值会回退为 `0`。 |
| `--limit <n>` | 否 | `30` | 返回条数。小于 1 或非法值会回退为 `30`。 |
| `--category-id <id>` | 否 | 空字符串 | 按分类过滤。 |
| `--custom-space-id <id>` | 否 | 空字符串 | 按自定义空间过滤。 |
| `--is-to-square <0\|1>` | 否 | `1` | 是否查询广场 Agent。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |
| `--simple` | 否 | 无 | 兼容参数，当前不会改变输出。 |

背后 API：

```bash
curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v3/published/agent" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}" \
  -H "Content-Type: text/plain;charset=UTF-8" \
  -d '{"offset":0,"limit":10,"category_id":"","name":"","custom_space_id":"","is_to_square":1}'
```

对应示例：`make -C docs/user_manual/examples/cli list`。

## `kweaver agent personal-list`

用途：查询个人空间中的 Agent 列表。

```bash
kweaver agent personal-list --size 10 --publish-status "" --publish-to-be "" --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `--name <text>` | 否 | 空字符串 | 按名称过滤。 |
| `--pagination-marker <str>` | 否 | 空字符串 | 下一页分页游标。 |
| `--publish-status <status>` | 否 | 空字符串 | 发布状态过滤，可用 `unpublished`、`published`、`published_edited`。 |
| `--publish-to-be <value>` | 否 | 空字符串 | 发布类型过滤，例如 `skill_agent`、`api_agent`。 |
| `--size <n>` | 否 | `48` | 返回条数。小于 1 或非法值会回退为 `48`。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/personal-space/agent-list?size=10" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`make -C docs/user_manual/examples/cli personal-list`。

## `kweaver agent category-list`

用途：查询 Agent 分类。

```bash
kweaver agent category-list --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/category" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`make -C docs/user_manual/examples/cli category-list`。

## `kweaver agent template-list`

用途：查询已发布 Agent 模板列表。

```bash
kweaver agent template-list --name "" --size 10 --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `--category-id <id>` | 否 | 空字符串 | 按分类过滤。 |
| `--name <text>` | 否 | 空字符串 | 按模板名称过滤。 |
| `--pagination-marker <str>` | 否 | 空字符串 | 下一页分页游标。 |
| `--size <n>` | 否 | `48` | 返回条数。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/published/agent-tpl?size=10" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`make -C docs/user_manual/examples/cli template-list`。

## `kweaver agent template-get`

用途：获取已发布 Agent 模板详情，可把模板 `config` 保存到本地文件。

```bash
kweaver agent template-get "$TEMPLATE_ID" --save-config /tmp/kweaver-template-config.json --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<template_id>` | 是 | 无 | 模板 ID。 |
| `--save-config <path>` | 否 | 无 | 保存模板 config，CLI 会在文件名中追加时间戳。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/published/agent-tpl/$TEMPLATE_ID" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`TEMPLATE_ID=<template-id> make -C docs/user_manual/examples/cli detail-examples`。

## `kweaver agent get`

用途：按 `agent_id` 获取 Agent 详情。

```bash
kweaver agent get "$AGENT_ID" --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `--save-config <path>` | 否 | 无 | 保存当前 Agent config，CLI 会在文件名中追加时间戳。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `--verbose, -v` | 否 | 关闭 | 输出完整 API 响应；默认只输出摘要。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`AGENT_ID=<agent-id> make -C docs/user_manual/examples/cli detail-examples`。

## `kweaver agent get-by-key`

用途：按 `agent_key` 获取 Agent 信息。

```bash
kweaver agent get-by-key "$AGENT_KEY"
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<key>` | 是 | 无 | Agent key。 |

背后 API：

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/by-key/$AGENT_KEY" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}"
```

对应示例：`AGENT_KEY=<agent-key> make -C docs/user_manual/examples/cli detail-examples`。

## `kweaver agent create`

用途：创建 Agent。推荐通过 `--config <path>` 传入完整 config。

```bash
kweaver agent create \
  --name "$AGENT_NAME" \
  --profile "Created from CLI lifecycle doc" \
  --product-key dip \
  --mode default \
  --config "$AGENT_CONFIG" \
  --pretty
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `--name <text>` | 是 | 无 | Agent 名称。 |
| `--profile <text>` | 是 | 无 | Agent 描述。 |
| `--config <json\|path>` | 否 | 无 | 完整 config JSON 字符串或文件路径；推荐使用文件路径。 |
| `--key <text>` | 否 | 自动生成 | Agent 唯一 key。 |
| `--product-key <text>` | 否 | `dip` | 产品 key。 |
| `--system-prompt <text>` | 否 | 空字符串 | 未使用 `--config` 时写入 config 的系统提示词。 |
| `--llm-id <id>` | 条件必填 | 无 | 未使用 `--config` 时用于生成默认 LLM 配置。 |
| `--llm-max-tokens <n>` | 否 | `4096` | 未使用 `--config` 时的大模型最大 token。 |
| `--mode <mode>` | 否 | `default` | Agent 模式：`default`、`dolphin`、`react`。会覆盖 `config.mode`。 |
| `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |

背后 API：

```bash
curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v3/agent" \
  -H "X-Business-Domain: ${KWEAVER_BUSINESS_DOMAIN:-bd_public}" \
  -H "Content-Type: application/json" \
  -d '{"name":"...","profile":"...","product_key":"dip","config":{"mode":"default","llms":[{"is_default":true,"llm_config":{"id":"...","name":"..."}}]}}'
```

对应示例：`make -C docs/user_manual/examples/cli create`。

## `kweaver agent update`

用途：更新已有 Agent 的基础信息或 config。修改已发布 Agent 后，个人空间通常会变为“已发布编辑中”，需要重新发布才会生成新的发布版本。

```bash
kweaver agent update "$AGENT_ID" \
  --name "${AGENT_NAME}_updated" \
  --profile "Updated from CLI lifecycle doc" \
  --mode default
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `--name <text>` | 否 | 保持原值 | 更新 Agent 名称。 |
| `--profile <text>` | 否 | 保持原值 | 更新 Agent 描述。 |
| `--system-prompt <text>` | 否 | 保持原值 | 更新 `config.system_prompt`。 |
| `--knowledge-network-id <id>` | 否 | 无 | 将 `config.data_source.knowledge_network` 设置为只包含该知识网络 ID。 |
| `--config-path <path>` | 否 | 无 | 从文件读取完整 config。 |
| `--mode <mode>` | 否 | 原 config 或 `default` | Agent 模式：`default`、`dolphin`、`react`。 |

背后 API：

```bash
curl -X PUT "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" \
  -H "Content-Type: application/json" \
  -d '{"name":"...","profile":"...","product_key":"dip","config":{"mode":"default"}}'
```

知识网络绑定是可选场景，本地默认不运行：

```bash
AGENT_ID=<agent-id> KN_ID=<knowledge-network-id> make -C docs/user_manual/examples/cli update-knowledge-network
```

基础信息更新示例：

```bash
make -C docs/user_manual/examples/cli update
```

## `kweaver agent publish`

用途：发布 Agent。再次发布会生成新版本；只调整发布元信息请使用 API 的更新发布信息能力。

```bash
kweaver agent publish "$AGENT_ID"
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `--category-id <id>` | 否 | 空数组 | 发布分类。 |

当前 CLI 发布默认请求体会发布到广场，并发布为 `skill_agent`。如果需要更精细地选择发布类型，请使用 REST API 或 SDK。

背后 API：

```bash
curl -X POST "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/publish" \
  -H "Content-Type: application/json" \
  -d '{"category_ids":[],"description":"","publish_to_where":["square"],"publish_to_bes":["skill_agent"],"pms_control":null}'
```

对应示例：`make -C docs/user_manual/examples/cli publish`。

## `kweaver agent unpublish`

用途：取消发布 Agent。

```bash
kweaver agent unpublish "$AGENT_ID"
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |

背后 API：

```bash
curl -X PUT "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID/unpublish"
```

对应示例：`make -C docs/user_manual/examples/cli unpublish`。

## `kweaver agent delete`

用途：删除 Agent。

```bash
kweaver agent delete "$AGENT_ID" -y
```

参数：

| 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `<agent_id>` | 是 | 无 | Agent ID。 |
| `-y, --yes` | 否 | 关闭 | 跳过删除确认。 |

背后 API：

```bash
curl -X DELETE "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID"
```

对应示例：`make -C docs/user_manual/examples/cli delete`。

## `kweaver agent skill`

用途：管理 Agent config 中绑定的 skill。默认本地环境没有 `SKILL_ID`，相关示例不会进入 `quick-check`。

```bash
kweaver agent skill list "$AGENT_ID"
kweaver agent skill add "$AGENT_ID" "$SKILL_ID" --strict
kweaver agent skill remove "$AGENT_ID" "$SKILL_ID"
```

参数：

| 命令 | 参数 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `skill list` | `<agent-id>` | 是 | 无 | Agent ID。 |
| `skill list` | `--pretty` | 否 | 开启 | 缩进 JSON 输出。 |
| `skill list` | `--compact` | 否 | 关闭 | 紧凑 JSON 输出。 |
| `skill list` | `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `skill add` | `<agent-id>` | 是 | 无 | Agent ID。 |
| `skill add` | `<skill-id>...` | 是 | 无 | 一个或多个 skill ID。 |
| `skill add` | `--strict` | 否 | 关闭 | 如果 skill 未发布则报错；默认只警告并继续。 |
| `skill add` | `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |
| `skill remove` | `<agent-id>` | 是 | 无 | Agent ID。 |
| `skill remove` | `<skill-id>...` | 是 | 无 | 一个或多个 skill ID。 |
| `skill remove` | `-bd, --biz-domain <value>` | 否 | 当前配置 | 覆盖业务域。 |

背后 API：CLI 会先 `GET /api/agent-factory/v3/agent/{agent_id}` 读取当前 config，然后修改 `config.skills.skills`，最后 `PUT /api/agent-factory/v3/agent/{agent_id}` 写回。

```bash
curl "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID"
curl -X PUT "$KWEAVER_BASE_URL/api/agent-factory/v3/agent/$AGENT_ID" \
  -H "Content-Type: application/json" \
  -d '{"name":"...","profile":"...","product_key":"dip","config":{"skills":{"skills":[{"skill_id":"..."}]}}}'
```

条件示例：

```bash
AGENT_ID=<agent-id> SKILL_ID=<skill-id> make -C docs/user_manual/examples/cli skill-examples
```
