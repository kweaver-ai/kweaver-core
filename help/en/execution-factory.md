# Execution Factory

## Overview

The **Execution Factory** registers and runs **operators**, **tools**, and **skills** that agents invoke under policy. It bridges LLM planning to concrete business actions (HTTP tools, code runners, MCP integrations, etc.).

Typical ingress prefix:

| Prefix | Role |
| --- | --- |
| `/api/agent-operator-integration/v1` | Operator integration and tool execution surface |

**Related modules:** [Decision Agent](decision-agent.md), [Dataflow](dataflow.md) (automation and code-runner paths).

## CLI

### Calling Operator APIs

The `kweaver call` command invokes any registered operator API by path, acting as a lightweight HTTP client wired into the platform auth and routing.

```bash
# GET an operator endpoint
kweaver call GET /api/agent-operator-integration/v1/operators

# POST with a JSON body
kweaver call POST /api/agent-operator-integration/v1/operators/op-weather/invoke \
  --data '{"city": "Shanghai", "units": "metric"}'

# PUT to update an operator definition
kweaver call PUT /api/agent-operator-integration/v1/operators/op-weather \
  --data '{"description": "Updated weather lookup", "timeout": 30}'

# DELETE an operator
kweaver call DELETE /api/agent-operator-integration/v1/operators/op-weather
```

### Skill Management

Skills are packaged bundles of tools, prompts, and configuration that can be registered, discovered from the marketplace, and installed into your workspace.

```bash
# List installed skills
kweaver skill list

# Search the skill marketplace
kweaver skill market
kweaver skill market --query "data analysis"

# Register a new skill from a local SKILL.md
kweaver skill register ./my-skill/SKILL.md

# View skill content (metadata and tool definitions)
kweaver skill content <skill_id>

# Read a file within a skill package
kweaver skill read-file <skill_id> SKILL.md
kweaver skill read-file <skill_id> tools/analyze.py

# Install a skill from the marketplace
kweaver skill install <skill_id>
kweaver skill install <skill_id> --version 1.2.0
```

### End-to-End Example

```bash
# 1. Browse the marketplace for a useful skill
kweaver skill market --query "knowledge graph"

# 2. Install a skill
kweaver skill install skill-kg-analyzer

# 3. Verify it appears in local skills
kweaver skill list

# 4. View what tools the skill provides
kweaver skill content skill-kg-analyzer

# 5. Read the skill's documentation
kweaver skill read-file skill-kg-analyzer SKILL.md

# 6. Call an operator API to list registered operators
kweaver call GET /api/agent-operator-integration/v1/operators

# 7. Invoke an operator directly
kweaver call POST /api/agent-operator-integration/v1/operators/op-sql-query/invoke \
  --data '{"sql": "SELECT COUNT(*) FROM orders WHERE status = '\''pending'\''"}'
```

---

## Python SDK

```python
from kweaver_sdk import KWeaverClient

client = KWeaverClient(base_url="https://<access-address>")

# List all registered operators
operators = client.execution_factory.list_operators()
for op in operators:
    print(op["id"], op["name"], op["type"])

# Invoke an operator
result = client.execution_factory.invoke(
    operator_id="op-weather",
    input={"city": "Shanghai", "units": "metric"},
)
print(result["status"], result["output"])

# List installed skills
skills = client.skill.list()
for s in skills:
    print(s["id"], s["name"], s["version"])

# Search the marketplace
market_results = client.skill.market(query="data analysis")
for s in market_results:
    print(s["id"], s["name"], s["description"])

# Register a skill
registered = client.skill.register(path="./my-skill/SKILL.md")
print(registered["id"])

# Get skill content
content = client.skill.content("skill-kg-analyzer")
print(content["tools"])

# Read a file within the skill
file_content = client.skill.read_file("skill-kg-analyzer", "SKILL.md")
print(file_content)

# Install from marketplace
client.skill.install("skill-kg-analyzer", version="1.2.0")
```

---

## TypeScript SDK

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';

const client = new KWeaverClient({ baseUrl: 'https://<access-address>' });

// List operators
const operators = await client.executionFactory.listOperators();
operators.forEach((op) => console.log(op.id, op.name, op.type));

// Invoke an operator
const result = await client.executionFactory.invoke({
  operatorId: 'op-weather',
  input: { city: 'Shanghai', units: 'metric' },
});
console.log(result.status, result.output);

// List skills
const skills = await client.skill.list();
skills.forEach((s) => console.log(s.id, s.name, s.version));

// Marketplace search
const marketResults = await client.skill.market({ query: 'data analysis' });
marketResults.forEach((s) => console.log(s.id, s.name));

// Register a skill
const registered = await client.skill.register({
  path: './my-skill/SKILL.md',
});

// Get skill content
const content = await client.skill.content('skill-kg-analyzer');
console.log(content.tools);

// Read a file in a skill
const fileContent = await client.skill.readFile(
  'skill-kg-analyzer',
  'SKILL.md',
);

// Install from marketplace
await client.skill.install('skill-kg-analyzer', { version: '1.2.0' });
```

---

## curl

```bash
# List operators
curl -sk "https://<access-address>/api/agent-operator-integration/v1/operators" \
  -H "Authorization: Bearer $(kweaver token)"

# Get operator details
curl -sk "https://<access-address>/api/agent-operator-integration/v1/operators/op-weather" \
  -H "Authorization: Bearer $(kweaver token)"

# Invoke an operator
curl -sk -X POST "https://<access-address>/api/agent-operator-integration/v1/operators/op-weather/invoke" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{"city": "Shanghai", "units": "metric"}'

# Register a new operator
curl -sk -X POST "https://<access-address>/api/agent-operator-integration/v1/operators" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "weather-lookup",
    "type": "http",
    "description": "Look up current weather by city",
    "endpoint": "https://api.weather.example.com/v1/current",
    "method": "GET",
    "parameters": [
      {"name": "city", "type": "string", "required": true},
      {"name": "units", "type": "string", "default": "metric"}
    ]
  }'

# Delete an operator
curl -sk -X DELETE "https://<access-address>/api/agent-operator-integration/v1/operators/op-weather" \
  -H "Authorization: Bearer $(kweaver token)"

# List skills
curl -sk "https://<access-address>/api/agent-operator-integration/v1/skills" \
  -H "Authorization: Bearer $(kweaver token)"

# Search marketplace
curl -sk "https://<access-address>/api/agent-operator-integration/v1/skills/market?query=data+analysis" \
  -H "Authorization: Bearer $(kweaver token)"

# Get skill content
curl -sk "https://<access-address>/api/agent-operator-integration/v1/skills/skill-kg-analyzer/content" \
  -H "Authorization: Bearer $(kweaver token)"

# Read a file within a skill
curl -sk "https://<access-address>/api/agent-operator-integration/v1/skills/skill-kg-analyzer/files/SKILL.md" \
  -H "Authorization: Bearer $(kweaver token)"

# Install a skill
curl -sk -X POST "https://<access-address>/api/agent-operator-integration/v1/skills/skill-kg-analyzer/install" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{"version": "1.2.0"}'
```
