# Execution Factory

## 概述

**Execution Factory** 负责注册与执行智能体可调用的**算子**、**工具**与**技能（Skill）**，在策略控制下将 LLM 规划落地为具体业务动作（HTTP 工具、代码执行、MCP 等）。

核心能力：

| 能力 | 说明 |
| --- | --- |
| 算子（Operator） | 平台内置或用户注册的可执行单元，通过统一接口调用 |
| 工具（Tool） | HTTP/MCP/代码类工具，由智能体运行时按需调用 |
| 技能包（Skill） | 可分发、可安装的能力包，包含文档与元数据 |

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/agent-operator-integration/v1` | 算子集成与工具执行面 |

**相关模块：** [Decision Agent](decision-agent.md)、[Dataflow](dataflow.md)（自动化与 code-runner 路径）。

### CLI

#### 算子调用

通过 `kweaver call` 直接调用已注册的算子 API：

```bash
# 调用算子 — 传入 JSON 请求体
kweaver call POST /api/agent-operator-integration/v1/operators/op_text_extract/invoke \
  -d '{"input": {"url": "https://example.com/report.pdf"}, "params": {"format": "markdown"}}'

# 列出所有已注册算子
kweaver call GET /api/agent-operator-integration/v1/operators

# 获取算子详情
kweaver call GET /api/agent-operator-integration/v1/operators/op_text_extract

# 列出工具
kweaver call GET /api/agent-operator-integration/v1/tools

# 调用工具
kweaver call POST /api/agent-operator-integration/v1/tools/tool_web_search/invoke \
  -d '{"query": "KWeaver 最新版本", "limit": 5}'
```

#### 技能包管理

```bash
# 列出已安装的技能包
kweaver skill list

# 浏览技能市场
kweaver skill market

# 注册新技能包（上传 ZIP）
kweaver skill register --zip-file ./my-skill-v1.0.zip

# 查看技能包目录结构
kweaver skill content my-skill

# 读取技能包中的指定文件（渐进式读取）
kweaver skill read-file my-skill SKILL.md
kweaver skill read-file my-skill src/main.py

# 从市场安装技能包
kweaver skill install my-skill
kweaver skill install my-skill --version 1.2.0
```

#### 端到端流程

```bash
# 1. 浏览市场，寻找合适的技能包
kweaver skill market

# 2. 安装技能包
kweaver skill install data-quality-checker

# 3. 确认已安装
kweaver skill list

# 4. 查看技能说明
kweaver skill read-file data-quality-checker SKILL.md

# 5. 通过算子接口调用技能
kweaver call POST /api/agent-operator-integration/v1/operators/data_quality_checker/invoke \
  -d '{"input": {"table": "orders", "rules": ["not_null", "unique"]}}'
```

---

### Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<访问地址>")

operators = client.execution_factory.list_operators()
for op in operators["data"]:
    print(op["id"], op["name"], op["type"])

op_detail = client.execution_factory.get_operator("op_text_extract")
print(f"名称: {op_detail['name']}")
print(f"输入参数: {op_detail['input_schema']}")
print(f"输出参数: {op_detail['output_schema']}")

result = client.execution_factory.invoke(
    operator_id="op_text_extract",
    input={"url": "https://example.com/report.pdf"},
    params={"format": "markdown"}
)
print(result["output"])

tools = client.execution_factory.list_tools()
for tool in tools["data"]:
    print(tool["id"], tool["name"])

tool_result = client.execution_factory.invoke_tool(
    tool_id="tool_web_search",
    input={"query": "KWeaver 最新版本", "limit": 5}
)
for item in tool_result["results"]:
    print(item["title"], item["url"])

skills = client.skill.list()
for s in skills["data"]:
    print(s["name"], s["version"], s["status"])

market = client.skill.market()
for s in market["data"]:
    print(s["name"], s["description"], s["downloads"])

client.skill.install("data-quality-checker", version="1.2.0")

content = client.skill.read_file("data-quality-checker", "SKILL.md")
print(content)
```

---

### TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<访问地址>' });

const operators = await client.executionFactory.listOperators();
operators.data.forEach((op) => console.log(op.id, op.name, op.type));

const opDetail = await client.executionFactory.getOperator('op_text_extract');
console.log('名称:', opDetail.name);
console.log('输入参数:', opDetail.inputSchema);

const result = await client.executionFactory.invoke({
  operatorId: 'op_text_extract',
  input: { url: 'https://example.com/report.pdf' },
  params: { format: 'markdown' },
});
console.log(result.output);

const tools = await client.executionFactory.listTools();
tools.data.forEach((tool) => console.log(tool.id, tool.name));

const toolResult = await client.executionFactory.invokeTool({
  toolId: 'tool_web_search',
  input: { query: 'KWeaver 最新版本', limit: 5 },
});
toolResult.results.forEach((item) => console.log(item.title, item.url));

const skills = await client.skill.list();
skills.data.forEach((s) => console.log(s.name, s.version, s.status));

const market = await client.skill.market();
market.data.forEach((s) => console.log(s.name, s.description));

await client.skill.install('data-quality-checker', { version: '1.2.0' });

const content = await client.skill.readFile('data-quality-checker', 'SKILL.md');
console.log(content);
```

---

### curl

```bash
# 列出算子
curl -sk "https://<访问地址>/api/agent-operator-integration/v1/operators" \
  -H "Authorization: Bearer $(kweaver token)"

# 获取算子详情
curl -sk "https://<访问地址>/api/agent-operator-integration/v1/operators/op_text_extract" \
  -H "Authorization: Bearer $(kweaver token)"

# 调用算子
curl -sk -X POST "https://<访问地址>/api/agent-operator-integration/v1/operators/op_text_extract/invoke" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "input": {"url": "https://example.com/report.pdf"},
    "params": {"format": "markdown"}
  }'

# 列出工具
curl -sk "https://<访问地址>/api/agent-operator-integration/v1/tools" \
  -H "Authorization: Bearer $(kweaver token)"

# 调用工具
curl -sk -X POST "https://<访问地址>/api/agent-operator-integration/v1/tools/tool_web_search/invoke" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "KWeaver 最新版本",
    "limit": 5
  }'

# 列出已安装技能包
curl -sk "https://<访问地址>/api/agent-operator-integration/v1/skills" \
  -H "Authorization: Bearer $(kweaver token)"

# 注册技能包
curl -sk -X POST "https://<访问地址>/api/agent-operator-integration/v1/skills" \
  -H "Authorization: Bearer $(kweaver token)" \
  -F "file=@./my-skill-v1.0.zip"

# 从市场安装技能包
curl -sk -X POST "https://<访问地址>/api/agent-operator-integration/v1/skills/install" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "data-quality-checker",
    "version": "1.2.0"
  }'
```
