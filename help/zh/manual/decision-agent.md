# 🤖 Decision Agent

## 📖 概述

**Decision Agent** 是面向目标的智能体：规划、检索上下文、在策略下调用工具并基于反馈迭代。核心服务包括 **agent-factory**（编排 API）、**agent-executor**、**记忆**与检索集成等。

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/agent-factory/v1` | 工厂 API（模板、智能体、运行） |
| `/api/agent-factory/v2`、`/api/agent-factory/v3` | 版本化工厂接口 |
| `/api/agent-app/v1` | 面向应用的智能体 API |

**相关模块：** [Context Loader](context-loader.md)、[Execution Factory](execution-factory.md)、[BKN 引擎](bkn.md)、[Trace AI](trace-ai.md)。

> **模型配置前置条件**：Agent 需要 LLM（大语言模型）和 Embedding 小模型。`--minimum` 安装不包含预置模型，使用前请先完成 [安装与部署 — 配置模型](../install.md#配置模型)。创建 Agent 时需通过 `--llm-id` 指定已注册的 LLM ID。

## 🚀 使用方式

推荐先 `kweaver auth login <平台地址>`（自签名证书加 `-k`），再使用下文 CLI；REST 示例见文档末尾 curl 一节。

### CLI

#### 智能体列表与详情

```bash
# 列出智能体，按名称过滤，最多 50 条，显示详情
kweaver agent list --name "客服" --limit 50 --verbose

# 获取智能体详情
kweaver agent get agt_001 --verbose

# 导出智能体配置到文件
kweaver agent get agt_001 --save-config ./agent-configs/customer-service.json
```

#### 个人空间与模板

```bash
# 列出个人空间的智能体
kweaver agent personal-list

# 列出所有可用模板
kweaver agent template-list

# 获取模板详情
kweaver agent template-get tpl_qa_assistant --verbose

# 将模板配置保存为本地文件（后续用于创建）
kweaver agent template-get tpl_qa_assistant --save-config ./templates/qa.json
```

#### 分类

```bash
# 列出智能体分类
kweaver agent category-list
```

#### 创建智能体

```bash
# 使用基本参数创建
kweaver agent create \
  --name "客服助手" \
  --profile "专业客服智能体，擅长解答产品问题和处理客户投诉" \
  --llm-id llm_gpt4o \
  --system-prompt "你是一个专业的客服助手。请用友好、专业的语气回答客户问题。"

# 使用配置文件创建（配置来自模板导出）
kweaver agent create \
  --name "知识问答助手" \
  --profile "基于知识网络的问答智能体" \
  --llm-id llm_gpt4o \
  --config ./templates/qa.json

# 使用内联 JSON 配置创建
kweaver agent create \
  --name "数据分析助手" \
  --profile "帮助用户分析数据" \
  --llm-id llm_gpt4o \
  --config '{"tools":["web_search","code_runner"],"temperature":0.3}'
```

#### 更新智能体

```bash
# 更新名称与描述
kweaver agent update agt_001 \
  --name "高级客服助手" \
  --profile "升级版客服智能体，支持多语言"

# 更新系统提示词
kweaver agent update agt_001 \
  --system-prompt "你是一个高级客服助手。请用多语言回答客户问题，优先使用客户的语言。"

# 绑定知识网络
kweaver agent update agt_001 \
  --knowledge-network-id kn_abc123

# 使用配置文件更新
kweaver agent update agt_001 \
  --config-path ./agent-configs/updated.json
```

#### 发布与取消发布

```bash
# 发布智能体到指定分类
kweaver agent publish agt_001 --category-id cat_customer_service

# 取消发布
kweaver agent unpublish agt_001
```

#### 对话

```bash
# 单次对话（非流式）
kweaver agent chat agt_001 -m '最近一周有多少新客户注册？'

# 流式对话
kweaver agent chat agt_001 -m '分析一下华东区的销售趋势' --stream

# 在已有会话中继续对话
kweaver agent chat agt_001 -m '能否按月份细分？' \
  --conversation-id conv_20250115_001

# 交互式对话模式（连续多轮）
kweaver agent chat agt_001
# > 你好，请帮我查询VIP客户列表
# > 这些客户的平均订单金额是多少？
# > quit   # 或 exit、q（与 kweaver agent chat --help 一致）
```

#### 会话管理

```bash
# 列出智能体的所有会话
kweaver agent sessions agt_001

# 查看特定会话的完整历史
kweaver agent history agt_001 conv_20250115_001
```

#### 链路追踪

第一个参数为**智能体 ID**，第二个为**会话 ID**（与 `kweaver agent chat` 输出或 `kweaver agent sessions` 中一致）。`kweaver agent --help` 总览里对 `trace` 的一行描述可能与 `kweaver agent trace --help` 不一致时，**以 `trace --help` 为准**。

```bash
# 查看会话的执行链路（格式化输出）
kweaver agent trace agt_001 conv_20250115_001 --pretty

# 紧凑输出（适合管道处理）
kweaver agent trace agt_001 conv_20250115_001 --compact
```

#### 删除

```bash
# 删除智能体（需确认）
kweaver agent delete agt_001

# 跳过确认直接删除
kweaver agent delete agt_001 -y
```

#### 端到端流程

下文用 **`agt_002`** 作为示例智能体 ID：第 3 步创建成功后，请将其**全部替换**为命令输出中的真实 ID。第 7 步的 **`conv_xxx`** 须与第 6 步对话返回的会话 ID 一致，或通过 `kweaver agent sessions agt_002` 查看。

```bash
# 1. 查看可用模板
kweaver agent template-list

# 2. 选择模板并保存配置
kweaver agent template-get tpl_qa_assistant --save-config ./qa-config.json

# 3. 基于模板创建智能体（记下返回的 agent id，下文以 agt_002 为例）
kweaver agent create \
  --name "产品问答助手" \
  --profile "回答关于产品功能的问题" \
  --llm-id llm_gpt4o \
  --config ./qa-config.json

# 4. 绑定知识网络
kweaver agent update agt_002 --knowledge-network-id kn_abc123

# 5. 发布到市场
kweaver agent publish agt_002 --category-id cat_product

# 6. 对话测试（记下 conversation_id，或通过 sessions 查询）
kweaver agent chat agt_002 -m '这个产品支持哪些数据库？' --stream

# 7. 查看执行链路（须同时传入智能体 ID 与会话 ID）
kweaver agent trace agt_002 conv_20250116_001 --pretty

# 8. 清理（如不再需要）
kweaver agent unpublish agt_002
kweaver agent delete agt_002 -y
```

---

### Python SDK

```python
from kweaver_sdk import KWeaverClient

# 使用已 login 的 ~/.kweaver/ 凭据（需先 kweaver auth login；具体构造方式以 kweaver-sdk 为准）
client = KWeaverClient()

agents = client.agent.list(name="客服", limit=50)
for agt in agents["data"]:
    print(agt["id"], agt["name"], agt["status"])

detail = client.agent.get("agt_001", verbose=True)
print(f"名称: {detail['name']}")
print(f"LLM: {detail['llm_id']}")
print(f"知识网络: {detail.get('knowledge_network_id', '未绑定')}")

templates = client.agent.template_list()
for tpl in templates["data"]:
    print(tpl["id"], tpl["name"], tpl["description"])

tpl_config = client.agent.template_get("tpl_qa_assistant")
print(f"模板配置: {tpl_config['config']}")

new_agent = client.agent.create(
    name="数据分析助手",
    profile="帮助用户分析业务数据",
    llm_id="llm_gpt4o",
    system_prompt="你是一个数据分析专家。",
    config={"tools": ["code_runner", "web_search"], "temperature": 0.3}
)
print(f"新智能体 ID: {new_agent['id']}")

client.agent.update(new_agent["id"], knowledge_network_id="kn_abc123")

client.agent.publish(new_agent["id"], category_id="cat_analysis")

reply = client.agent.chat(
    agent_id=new_agent["id"],
    message="最近一个月的销售趋势如何？"
)
print(f"回复: {reply['content']}")
print(f"会话 ID: {reply['conversation_id']}")

follow_up = client.agent.chat(
    agent_id=new_agent["id"],
    message="能否按区域细分？",
    conversation_id=reply["conversation_id"]
)
print(f"回复: {follow_up['content']}")

for chunk in client.agent.chat_stream(
    agent_id=new_agent["id"],
    message="生成一份销售报告"
):
    print(chunk["delta"], end="", flush=True)

sessions = client.agent.sessions(new_agent["id"])
for s in sessions["data"]:
    print(s["conversation_id"], s["created_at"], s["message_count"])

history = client.agent.history(new_agent["id"], reply["conversation_id"])
for msg in history["messages"]:
    print(f"[{msg['role']}] {msg['content'][:80]}")

trace = client.agent.trace(new_agent["id"], reply["conversation_id"])
for span in trace["spans"]:
    print(f"  {'  ' * span['depth']}{span['operation']} ({span['duration_ms']}ms)")

categories = client.agent.category_list()
for cat in categories["data"]:
    print(cat["id"], cat["name"])
```

---

### TypeScript SDK

> 💡 更多可运行示例见随 `@kweaver-ai/kweaver-sdk` 包发布的示例目录。

```typescript
import { KWeaverClient } from '@kweaver-ai/kweaver-sdk';
import type { ProgressItem } from '@kweaver-ai/kweaver-sdk';

// 自动读取 ~/.kweaver/ 凭据
const client = await KWeaverClient.connect();

// 列出智能体
const agentList = await client.agents.list({ limit: 10 });
for (const a of agentList) {
  console.log(`${a.name} (${a.id}) — ${a.description ?? ''}`);
}

// 单轮对话
const agentId = agentList[0].id;
const reply = await client.agents.chat(agentId, '最近一个月的销售趋势如何？');
console.log('回复:', reply.text);

// 查看推理链路（progress chain）
if (reply.progress && reply.progress.length > 0) {
  for (const step of reply.progress) {
    console.log(`[${step.skill_info?.type ?? 'step'}] ${step.skill_info?.name ?? step.agent_name} → ${step.status}`);
  }
}

// 流式对话（实时输出）
let prevLen = 0;
const streamResult = await client.agents.stream(
  agentId,
  '能否按区域细分？',
  {
    onTextDelta: (fullText: string) => {
      process.stdout.write(fullText.slice(prevLen));
      prevLen = fullText.length;
    },
    onProgress: (progress: ProgressItem[]) => {
      for (const p of progress) {
        if (p.skill_info?.name) {
          console.log(`[progress] ${p.skill_info.name} → ${p.status ?? ''}`);
        }
      }
    },
  },
);

// 会话历史
const conversationId = reply.conversationId;
if (conversationId) {
  const messages = await client.conversations.listMessages(conversationId, { limit: 10 });
  for (const msg of messages) {
    console.log(`[${msg.role}] ${(msg.content ?? '').slice(0, 80)}`);
  }
}

// 列出会话
const sessions = await client.conversations.list(agentId, { limit: 5 });
for (const s of sessions) {
  console.log(`${s.id} — ${s.created_at ?? ''}`);
}
```

---

### curl

```bash
# 列出智能体（查询串含中文时，部分环境需对 name 做 URL 编码）
curl -sk "https://<访问地址>/api/agent-factory/v1/agents?name=客服&limit=50" \
  -H "Authorization: Bearer $(kweaver token)"

# 获取智能体详情
curl -sk "https://<访问地址>/api/agent-factory/v1/agents/agt_001" \
  -H "Authorization: Bearer $(kweaver token)"

# 列出模板
curl -sk "https://<访问地址>/api/agent-factory/v1/templates" \
  -H "Authorization: Bearer $(kweaver token)"

# 获取模板详情
curl -sk "https://<访问地址>/api/agent-factory/v1/templates/tpl_qa_assistant" \
  -H "Authorization: Bearer $(kweaver token)"

# 创建智能体
curl -sk -X POST "https://<访问地址>/api/agent-factory/v1/agents" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "数据分析助手",
    "profile": "帮助用户分析业务数据",
    "llm_id": "llm_gpt4o",
    "system_prompt": "你是一个数据分析专家。",
    "config": {
      "tools": ["code_runner", "web_search"],
      "temperature": 0.3
    }
  }'

# 更新智能体
curl -sk -X PUT "https://<访问地址>/api/agent-factory/v1/agents/agt_001" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "knowledge_network_id": "kn_abc123"
  }'

# 发布智能体
curl -sk -X POST "https://<访问地址>/api/agent-factory/v1/agents/agt_001/publish" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{"category_id": "cat_customer_service"}'

# 对话（非流式）
curl -sk -X POST "https://<访问地址>/api/agent-app/v1/agents/agt_001/chat" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "最近一周有多少新客户注册？",
    "stream": false
  }'

# 对话（流式 SSE）
curl -sk -N -X POST "https://<访问地址>/api/agent-app/v1/agents/agt_001/chat" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "message": "分析一下华东区的销售趋势",
    "stream": true
  }'

# 在已有会话中继续对话
curl -sk -X POST "https://<访问地址>/api/agent-app/v1/agents/agt_001/chat" \
  -H "Authorization: Bearer $(kweaver token)" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "能否按月份细分？",
    "conversation_id": "conv_20250115_001",
    "stream": false
  }'

# 列出会话
curl -sk "https://<访问地址>/api/agent-app/v1/agents/agt_001/sessions" \
  -H "Authorization: Bearer $(kweaver token)"

# 查看会话历史
curl -sk "https://<访问地址>/api/agent-app/v1/agents/agt_001/sessions/conv_20250115_001/history" \
  -H "Authorization: Bearer $(kweaver token)"

# 删除智能体
curl -sk -X DELETE "https://<访问地址>/api/agent-factory/v1/agents/agt_001" \
  -H "Authorization: Bearer $(kweaver token)"

# 列出分类
curl -sk "https://<访问地址>/api/agent-factory/v1/categories" \
  -H "Authorization: Bearer $(kweaver token)"
```
