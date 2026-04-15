# Model Management

## Overview

KWeaver Core manages LLMs and small models through a unified **Model Manager** (`mf-model-manager`). No models are pre-configured by default — you must register them before using semantic search or Agent features.

| Model Type | Purpose | Required For |
|------------|---------|--------------|
| **LLM** (Large Language Model) | Agent conversation, reasoning, decisions | Decision Agent |
| **Embedding** | Vectorization, semantic search, intent recognition | `kweaver bkn search`, Agent intent recognition |
| **Reranker** | Re-ranking retrieval results | Optional, improves search accuracy |

Typical ingress prefix:

| Prefix | Role |
| --- | --- |
| `/api/mf-model-manager/v1` | Model management API |

**Related modules:** [BKN Engine](bkn.md) (semantic search uses Embedding), [Decision Agent](decision-agent.md) (uses LLM), [Context Loader](context-loader.md) (retrieval uses Embedding + Reranker).

---

## CLI

All operations below use `kweaver call`, which auto-injects auth and the platform base URL.

### LLM Management

#### Supported types and series

**LLM `model_type` (when registering an LLM)**  

Supported: **`llm`**, **`rlm`**, **`vu`**. Typical chat workloads use **`llm`**; you may omit the field — it defaults to **`llm`**.

**LLM `model_series` (when registering an LLM)**  

Supported: `tome`, `qwen`, `openai`, `internlm`, `deepseek`, `qianxun`, `claude`, `chatglm`, `llama`, `others`, `baidu`, `baidu_tianchen`. Pick the value that matches your provider (e.g. `qwen` for Tongyi Qwen, `deepseek` for DeepSeek).

**`openai` means Azure OpenAI only** — configure `api_url`, deployment name, and keys as required by Azure.  
For **any other OpenAI Chat Completions–compatible HTTP endpoint** (official OpenAI, Tencent Cloud MaaS, and other **non-Azure** hosts), use **`others`**. Do not register those endpoints under `openai`.

**Small-model `model_type` (when registering a small model)**  

Supported: **`embedding`**, **`reranker`**.

---

#### Register an LLM

Register an OpenAI-compatible LLM:

```bash
# DeepSeek
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "deepseek-chat",
  "model_series": "deepseek",
  "max_model_len": 8192,
  "model_config": {
    "api_key": "<your-api-key>",
    "api_model": "deepseek-chat",
    "api_url": "https://api.deepseek.com/chat/completions"
  }
}'

# Azure OpenAI (`model_series` must be `openai`; set `api_url` / `api_model` per Azure portal)
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "gpt-4o-azure",
  "model_series": "openai",
  "max_model_len": 128000,
  "model_config": {
    "api_key": "<your-azure-api-key>",
    "api_model": "<your-azure-deployment-name>",
    "api_url": "<azure-openai-resource-base-url-from-portal>"
  }
}'

# Official OpenAI and other OpenAI Chat Completions–compatible, non-Azure hosts — use `others`
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "gpt-4o",
  "model_series": "others",
  "max_model_len": 128000,
  "model_config": {
    "api_key": "<your-api-key>",
    "api_model": "gpt-4o",
    "api_url": "https://api.openai.com/v1/chat/completions"
  }
}'

# Qwen (Tongyi Qianwen)
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "qwen-plus",
  "model_series": "qwen",
  "max_model_len": 131072,
  "model_config": {
    "api_key": "<your-api-key>",
    "api_model": "qwen-plus",
    "api_url": "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
  }
}'

# Tencent Cloud MaaS (OpenAI Chat Completions–compatible; not Azure — use `others`)
kweaver call /api/mf-model-manager/v1/llm/add -d '{
  "model_name": "glm-5.1",
  "model_series": "others",
  "max_model_len": 128000,
  "model_config": {
    "api_key": "<your-api-key>",
    "api_model": "glm-5.1",
    "api_url": "https://tokenhub.tencentmaas.com/v1/chat/completions"
  }
}'
```

**Reminder:** **`openai` = Azure OpenAI**; **other OpenAI-compatible endpoints use `others`**. If a vendor documents a dedicated series (e.g. `qwen`, `deepseek`), prefer that series when it applies.

#### List LLMs

```bash
kweaver call '/api/mf-model-manager/v1/llm/list?page=1&size=50'
```

#### Test LLM Connectivity

```bash
kweaver call /api/mf-model-manager/v1/llm/test -d '{
  "model_id": "<model_id>"
}'
```

#### Chat with an LLM directly (Model Factory)

To talk to a registered LLM **without** going through Decision Agent, call the Model Factory **Chat Completions** API. `kweaver call` injects the current platform auth and business domain for you.

- **Endpoint**: `POST /api/mf-model-api/v1/chat/completions`
- **`model`**: use the registered **`model_name`** from **`llm/list`** (usually the same string as upstream `api_model`)
- **Body**: common OpenAI Chat Completions fields such as **`messages`** (`system` / `user` / `assistant`), **`stream`**, **`max_tokens`**, **`temperature`**, **`top_p`**, **`top_k`**, etc. If the model emits long “reasoning” text, set **`max_tokens`** high enough or the reply may be cut off

Non-streaming example:

```bash
kweaver call /api/mf-model-api/v1/chat/completions -X POST \
  -d '{
    "model": "<model_name>",
    "messages": [{"role": "user", "content": "Introduce yourself in one sentence."}],
    "stream": false,
    "max_tokens": 512,
    "temperature": 0.7,
    "top_p": 0.9,
    "top_k": 50
  }' --pretty
```

For streaming, set **`"stream": true`** (SSE stream; your terminal or client must handle chunked output).

#### Delete an LLM

```bash
kweaver call /api/mf-model-manager/v1/llm/delete -d '{
  "model_ids": ["<model_id>"]
}'
```

### Small Model Management (Embedding / Reranker)

#### Register an Embedding Model

```bash
# BGE-M3 (via SiliconFlow)
kweaver call /api/mf-model-manager/v1/small-model/add -d '{
  "model_name": "bge-m3",
  "model_type": "embedding",
  "model_config": {
    "api_url": "https://api.siliconflow.cn/v1/embeddings",
    "api_model": "BAAI/bge-m3",
    "api_key": "<your-api-key>"
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
    "api_key": "<your-api-key>"
  },
  "batch_size": 32,
  "max_tokens": 8191,
  "embedding_dim": 1536
}'
```

#### Register a Reranker Model (optional)

A Reranker re-ranks semantic search results for better precision:

```bash
kweaver call /api/mf-model-manager/v1/small-model/add -d '{
  "model_name": "bge-reranker-v2-m3",
  "model_type": "reranker",
  "model_config": {
    "api_url": "https://api.siliconflow.cn/v1/rerank",
    "api_model": "BAAI/bge-reranker-v2-m3",
    "api_key": "<your-api-key>"
  },
  "batch_size": 32,
  "max_tokens": 512
}'
```

#### List Small Models

```bash
kweaver call '/api/mf-model-manager/v1/small-model/list?page=1&size=50'
```

Example response:

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

#### Test Small Model Connectivity

```bash
kweaver call /api/mf-model-manager/v1/small-model/test -d '{
  "model_id": "<model_id>"
}'
```

#### Delete a Small Model

```bash
kweaver call /api/mf-model-manager/v1/small-model/delete -d '{
  "model_id": "<model_id>"
}'
```

---

## Enable BKN Semantic Search

After registering an embedding in the **model factory**, point **bkn-backend** and **ontology-query** at the same default name — the **`model_name`** from the list API.

**1.** Run the list call to read **`model_name`**, then `kubectl edit configmap bkn-backend-cm` and `ontology-query-cm` (namespace is often `kweaver`). In the YAML blob under `data`, under **`server:`**, set:

```bash
kweaver call '/api/mf-model-manager/v1/small-model/list?page=1&size=50'
```

```yaml
server:
  defaultSmallModelEnabled: true
  defaultSmallModelName: <model_name from above>
```

Edit **both** ConfigMaps; **`defaultSmallModelName` must match**. Add the line under `server:` if it is missing.

**2.** Save, restart, and test; run **`bkn build`** again if indexes were built with the wrong model.

```bash
kubectl rollout restart deployment/bkn-backend -n kweaver
kubectl rollout restart deployment/ontology-query -n kweaver
kweaver bkn search <kn_id> "test query"
# optional: kweaver bkn build <kn_id> --wait --timeout 600
```

**Troubleshooting**: **`IdNotExist`** usually means `defaultSmallModelName` does not match the list, or only one ConfigMap was edited / pods not restarted. **`Redis GET` timeout**: check **mf-model-api** ↔ Redis/Sentinel or restart **mf-model-api**.

---

## Parameter Reference

### LLM Parameters

| Parameter | Required | Description |
|-----------|:--------:|-------------|
| `model_name` | YES | Display name (must be unique) |
| `model_series` | YES | Model series: `openai`, `deepseek`, `qwen`, `claude`, `tome`, etc. |
| `max_model_len` | YES | Maximum context length (tokens) |
| `model_config.api_key` | YES | API key |
| `model_config.api_model` | YES | Model name on the provider side |
| `model_config.api_url` | YES | Chat Completions API endpoint |

### Small Model Parameters

| Parameter | Required | Description |
|-----------|:--------:|-------------|
| `model_name` | YES | Display name (must be unique) |
| `model_type` | YES | `embedding` or `reranker` |
| `model_config.api_key` | YES | API key |
| `model_config.api_model` | YES | Model name on the provider side |
| `model_config.api_url` | YES | Embedding / Rerank API endpoint |
| `batch_size` | NO | Batch processing size (default 32) |
| `max_tokens` | NO | Maximum tokens per request |
| `embedding_dim` | NO | Vector dimensions (embedding type only) |

---

## Common Providers

| Provider | Model Type | Model Name | API Endpoint |
|----------|-----------|------------|--------------|
| DeepSeek | LLM | `deepseek-chat` | `https://api.deepseek.com/chat/completions` |
| OpenAI | LLM | `gpt-4o` | `https://api.openai.com/v1/chat/completions` |
| Qwen | LLM | `qwen-plus` | `https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions` |
| Tencent Cloud MaaS | LLM | `glm-5.1` (see console) | `https://tokenhub.tencentmaas.com/v1/chat/completions` |
| SiliconFlow | Embedding | `BAAI/bge-m3` | `https://api.siliconflow.cn/v1/embeddings` |
| SiliconFlow | Reranker | `BAAI/bge-reranker-v2-m3` | `https://api.siliconflow.cn/v1/rerank` |
| OpenAI | Embedding | `text-embedding-3-small` | `https://api.openai.com/v1/embeddings` |

Self-hosted models (vLLM, Ollama, etc.) work by pointing `api_url` to the local endpoint. Use `openai` or `tome` as the `model_series`.

---

## End-to-End Workflow

```bash
# 1. Register LLM
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

# 2. Register Embedding
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

# 3. Verify registration
kweaver call '/api/mf-model-manager/v1/llm/list?page=1&size=50'
kweaver call '/api/mf-model-manager/v1/small-model/list?page=1&size=50'

# 4. Test connectivity
kweaver call /api/mf-model-manager/v1/llm/test -d '{"model_id": "<llm_id>"}'
kweaver call /api/mf-model-manager/v1/small-model/test -d '{"model_id": "<embedding_id>"}'

# 5. Enable BKN semantic search (kubectl): see "Enable BKN Semantic Search" steps 1–2 above
kubectl edit configmap bkn-backend-cm -n kweaver
kubectl edit configmap ontology-query-cm -n kweaver
kubectl rollout restart deployment/bkn-backend -n kweaver
kubectl rollout restart deployment/ontology-query -n kweaver
kweaver bkn build <kn_id> --wait --timeout 600

# 6. Verify semantic search
kweaver bkn search <kn_id> "test query"

# 7. Create Agent (requires llm_id)
kweaver agent create --name "test-agent" --profile "answer questions" --llm-id <llm_id>
```
