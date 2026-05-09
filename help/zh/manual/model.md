# 🧠 模型管理

## 📖 概述

KWeaver Core 通过**模型管理器**（`mf-model-manager`）统一管理 LLM 和小模型。平台默认不包含预置模型，需自行注册后才能使用语义搜索和 Agent 功能。

| 模型类别 | 用途 | 必需场景 |
|---------|------|---------|
| **LLM**（大语言模型） | Agent 对话、推理、决策 | Decision Agent |
| **Embedding**（嵌入模型） | 向量化、语义搜索、意图识别 | `kweaver bkn search`、Agent 意图识别 |
| **Reranker**（重排模型） | 检索结果精排 | 可选，提高检索精度 |

典型 Ingress 前缀：

| 前缀 | 作用 |
| --- | --- |
| `/api/mf-model-manager/v1` | 模型管理 API |

**相关模块：** [BKN 引擎](bkn.md)（语义搜索消费 Embedding）、[Decision Agent](decision-agent.md)（消费 LLM）、[Context Loader](context-loader.md)（检索使用 Embedding + Reranker）。

---

## 💻 CLI

以下操作使用 `kweaver call`，CLI 会自动注入认证和平台地址。

### LLM 管理

#### 支持的类型与系列

**大模型 `model_type`（注册 LLM 时）**  

支持：**`llm`**、**`rlm`**、**`vu`**。一般对话场景用 **`llm`** 即可；可省略，省略时默认为 **`llm`**。

**大模型 `model_series`（注册 LLM 时）**  

支持：`tome`、`qwen`、`openai`、`internlm`、`deepseek`、`qianxun`、`claude`、`chatglm`、`llama`、`others`、`baidu`、`baidu_tianchen`。请按实际上游选择对应系列（如通义用 `qwen`、DeepSeek 用 `deepseek` 等）。

其中 **`openai` 专指 Azure OpenAI**：按 Azure 要求填写资源地址、部署名与密钥。  
**其它「OpenAI Chat Completions」兼容的 HTTP 服务**（例如官方 OpenAI、腾讯云 MaaS 等**非 Azure** 端点）请使用 **`others`**，不要把这类端点配在 `openai` 系列下。

**小模型 `model_type`（注册小模型时）**  

支持：**`embedding`**、**`reranker`**。

---

#### 注册 LLM

注册一个 OpenAI 兼容的 LLM：

```bash
# DeepSeek
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "deepseek-chat",
  "model_series": "deepseek",
  "max_model_len": 8192,
  "model_config": {
    "api_key": "<你的 API Key>",
    "api_model": "deepseek-chat",
    "api_url": "https://api.deepseek.com/chat/completions"
  }
}'

# Azure OpenAI（model_series 必须为 openai；api_url、api_model 以 Azure 门户为准）
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "gpt-4o-azure",
  "model_series": "openai",
  "max_model_len": 128000,
  "model_config": {
    "api_key": "<你的 Azure API Key>",
    "api_model": "<Azure 上的部署名>",
    "api_url": "<Azure OpenAI 资源基址，按门户说明填写>"
  }
}'

# 官方 OpenAI 等「OpenAI Chat Completions 兼容、非 Azure」——使用 others
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "gpt-4o",
  "model_series": "others",
  "max_model_len": 128000,
  "model_config": {
    "api_key": "<你的 API Key>",
    "api_model": "gpt-4o",
    "api_url": "https://api.openai.com/v1/chat/completions"
  }
}'

# 通义千问
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "qwen-plus",
  "model_series": "qwen",
  "max_model_len": 131072,
  "model_config": {
    "api_key": "<你的 API Key>",
    "api_model": "qwen-plus",
    "api_url": "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
  }
}'

# 腾讯云 MaaS（OpenAI Chat Completions 兼容；非 Azure，请用 others）
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "glm-5.1",
  "model_series": "others",
  "max_model_len": 128000,
  "model_config": {
    "api_key": "<你的 API Key>",
    "api_model": "glm-5.1",
    "api_url": "https://tokenhub.tencentmaas.com/v1/chat/completions"
  }
}'
```

再次说明：**`openai` = Azure OpenAI**；**其它 OpenAI 兼容模式（完整 Chat Completions URL + Bearer 等）用 `others`**。厂商若提供专用系列（如 `qwen`、`deepseek`），也可优先选用对应系列。

#### 列出 LLM

```bash
kweaver call '/api/mf-model-manager/v1/llm/list?page=1&size=50'
```

#### 测试 LLM 连通性

```bash
kweaver call /api/mf-model-manager/v1/llm/test -d '{
  "model_id": "<model_id>"
}'
```

#### 通过模型工厂直接与 LLM 对话

不经过 Decision Agent 时，可对**模型工厂**的 Chat Completions 接口发请求，与已注册大模型对话。`kweaver call` 会自动带上当前平台的认证与业务域。

- **路径**：`POST /api/mf-model-api/v1/chat/completions`
- **`model`**：填写已在模型管理器中注册的 **`model_name`**（与 **`llm/list`** 中名称一致，通常与上游 `api_model` 相同）
- **正文**：OpenAI Chat Completions 常见字段，如 **`messages`**（`system` / `user` / `assistant`）、**`stream`**、**`max_tokens`**、**`temperature`**、**`top_p`**、**`top_k`** 等；若模型会输出较长「思考」内容，请把 **`max_tokens`** 设得足够大，否则可能因长度截断而看不到完整回答

单次非流式示例：

```bash
kweaver call /api/mf-model-api/v1/chat/completions -X POST \
  -d '{
    "model": "<model_name>",
    "messages": [{"role": "user", "content": "请用一句话介绍你自己。"}],
    "stream": false,
    "max_tokens": 512,
    "temperature": 0.7,
    "top_p": 0.9,
    "top_k": 50
  }' --pretty
```

流式对话可加 **`"stream": true`**（返回为 SSE 流，终端上为持续输出，需自行处理或换用支持流式的客户端）。

#### 删除 LLM

```bash
kweaver call /api/mf-model-manager/v1/llm/delete -d '{
  "model_ids": ["<model_id>"]
}'
```

### 小模型管理（Embedding / Reranker）

#### 注册 Embedding 模型

```bash
# BGE-M3（通过 SiliconFlow）
kweaver call /api/mf-model-manager/v1/small-model/add -d '{
  "model_name": "bge-m3",
  "model_type": "embedding",
  "model_config": {
    "api_url": "https://api.siliconflow.cn/v1/embeddings",
    "api_model": "BAAI/bge-m3",
    "api_key": "<你的 API Key>"
  },
  "batch_size": 32,
  "max_tokens": 512,
  "embedding_dim": 1024
}'

# OpenAI text-embedding-3-small
kweaver call /api/mf-model-manager/v1/small-model/add -d '{
  "model_name": "text-embedding-3-small",
  "model_type": "embedding",
  "model_config": {
    "api_url": "https://api.openai.com/v1/embeddings",
    "api_model": "text-embedding-3-small",
    "api_key": "<你的 API Key>"
  },
  "batch_size": 32,
  "max_tokens": 8191,
  "embedding_dim": 1536
}'
```

#### 注册 Reranker 模型（可选）

Reranker 对语义搜索结果进行精排，可提高检索精度：

```bash
kweaver call /api/mf-model-manager/v1/small-model/add -d '{
  "model_name": "bge-reranker-v2-m3",
  "model_type": "reranker",
  "model_config": {
    "api_url": "https://api.siliconflow.cn/v1/rerank",
    "api_model": "BAAI/bge-reranker-v2-m3",
    "api_key": "<你的 API Key>"
  },
  "batch_size": 32,
  "max_tokens": 512
}'
```

#### 列出小模型

```bash
kweaver call '/api/mf-model-manager/v1/small-model/list?page=1&size=50'
```

返回示例：

```json
{
  "count": 2,
  "data": [
    {
      "model_id": "2044075511382151168",
      "model_name": "bge-m3",
      "model_type": "embedding",
      "model_config": { "api_url": "...", "api_model": "BAAI/bge-m3" },
      "batch_size": 32,
      "max_tokens": 512,
      "embedding_dim": 1024
    }
  ]
}
```

#### 测试小模型连通性

```bash
kweaver call /api/mf-model-manager/v1/small-model/test -d '{
  "model_id": "<model_id>"
}'
```

#### 删除小模型

```bash
kweaver call /api/mf-model-manager/v1/small-model/delete -d '{
  "model_id": "<model_id>"
}'
```

---

## 🔧 启用 BKN 语义搜索

除在 **模型工厂** 注册 Embedding 外，还要让 **bkn-backend** 与 **ontology-query** 使用同一个默认 Embedding 名（即列表里的 **`model_name`**）。

**1.** 用下面命令查出 **`model_name`**，然后 `kubectl edit configmap bkn-backend-cm`、`ontology-query-cm`（命名空间以集群为准，常见 `kweaver`）。在 ConfigMap 里 `data` 下那段 YAML 的 **`server:`** 中设置：

```bash
kweaver call '/api/mf-model-manager/v1/small-model/list?page=1&size=50'
```

```yaml
server:
  defaultSmallModelEnabled: true
  defaultSmallModelName: <上一步的 model_name>
```

两边 ConfigMap 都要改，**`defaultSmallModelName` 相同**；没有该字段就在 `server:` 下加一行。

**2.** 保存后重启并验证；若曾用错误模型建过索引，可再执行一次 `build`。

```bash
kubectl rollout restart deployment/bkn-backend -n kweaver
kubectl rollout restart deployment/ontology-query -n kweaver
kweaver bkn search <kn_id> "测试搜索"
# 可选：kweaver bkn build <kn_id> --wait --timeout 600
```

**排障**：`IdNotExist` 多为 `defaultSmallModelName` 与列表不一致，或只改了一侧 ConfigMap、未重启。若报 **Redis GET 超时**，检查 **mf-model-api** 与 Redis/Sentinel 或重启 `mf-model-api`。

---

## 📋 模型参数说明

### LLM 参数

| 参数 | 必填 | 说明 |
|------|:----:|------|
| `model_name` | YES | 模型显示名称，需唯一 |
| `model_series` | YES | 模型系列：`openai`、`deepseek`、`qwen`、`claude`、`tome` 等 |
| `max_model_len` | YES | 最大上下文长度（tokens） |
| `model_config.api_key` | YES | API Key |
| `model_config.api_model` | YES | 提供商侧的模型名称 |
| `model_config.api_url` | YES | Chat Completions API 端点 |

### 小模型参数

| 参数 | 必填 | 说明 |
|------|:----:|------|
| `model_name` | YES | 模型显示名称，需唯一 |
| `model_type` | YES | `embedding` 或 `reranker` |
| `model_config.api_key` | YES | API Key |
| `model_config.api_model` | YES | 提供商侧的模型名称 |
| `model_config.api_url` | YES | Embedding / Rerank API 端点 |
| `batch_size` | NO | 批量处理大小（默认 32） |
| `max_tokens` | NO | 单次最大 token 数 |
| `embedding_dim` | NO | 向量维度（仅 embedding 类型需要） |

---

## 🌐 常见模型提供商

| 提供商 | 模型类型 | 模型名称 | API 端点 |
|--------|---------|---------|---------|
| DeepSeek | LLM | `deepseek-chat` | `https://api.deepseek.com/chat/completions` |
| OpenAI | LLM | `gpt-4o` | `https://api.openai.com/v1/chat/completions` |
| 通义千问 | LLM | `qwen-plus` | `https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions` |
| 腾讯云 MaaS | LLM | `glm-5.1`（以控制台为准） | `https://tokenhub.tencentmaas.com/v1/chat/completions` |
| SiliconFlow | Embedding | `BAAI/bge-m3` | `https://api.siliconflow.cn/v1/embeddings` |
| SiliconFlow | Reranker | `BAAI/bge-reranker-v2-m3` | `https://api.siliconflow.cn/v1/rerank` |
| OpenAI | Embedding | `text-embedding-3-small` | `https://api.openai.com/v1/embeddings` |

私有部署的模型（如 vLLM、Ollama）只需将 `api_url` 指向本地端点即可，`model_series` 选 `openai` 或 `tome`。

---

## 🎯 端到端流程

```bash
# 1. 注册 LLM
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "deepseek-chat",
  "model_series": "deepseek",
  "max_model_len": 8192,
  "model_config": {
    "api_key": "<key>",
    "api_model": "deepseek-chat",
    "api_url": "https://api.deepseek.com/chat/completions"
  }
}'

# 2. 注册 Embedding
kweaver call /api/mf-model-manager/v1/small-model/add -d '{
  "model_name": "bge-m3",
  "model_type": "embedding",
  "model_config": {
    "api_url": "https://api.siliconflow.cn/v1/embeddings",
    "api_model": "BAAI/bge-m3",
    "api_key": "<key>"
  },
  "batch_size": 32,
  "max_tokens": 512,
  "embedding_dim": 1024
}'

# 3. 验证注册结果
kweaver call '/api/mf-model-manager/v1/llm/list?page=1&size=50'
kweaver call '/api/mf-model-manager/v1/small-model/list?page=1&size=50'

# 4. 测试连通性
kweaver call /api/mf-model-manager/v1/llm/test -d '{"model_id": "<llm_id>"}'
kweaver call /api/mf-model-manager/v1/small-model/test -d '{"model_id": "<embedding_id>"}'

# 5. 启用 BKN 语义搜索（kubectl）：见上文「启用 BKN 语义搜索」步骤 1、2
kubectl edit configmap bkn-backend-cm -n kweaver
kubectl edit configmap ontology-query-cm -n kweaver
kubectl rollout restart deployment/bkn-backend -n kweaver
kubectl rollout restart deployment/ontology-query -n kweaver
kweaver bkn build <kn_id> --wait --timeout 600

# 6. 验证语义搜索
kweaver bkn search <kn_id> "测试查询"

# 7. 创建 Agent（需要 llm_id）
kweaver agent create --name "测试助手" --profile "回答问题" --llm-id <llm_id>
```
