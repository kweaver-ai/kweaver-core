# KnowledgeRerank 模块实现设计文档

## 文档信息

| 项目 | 内容 |
|------|------|
| **功能名称** | KnowledgeRerank（知识重排序） |
| **关联文档** | [KnowledgeRerank模块逻辑设计.md](./KnowledgeRerank模块逻辑设计.md) |
| **文档版本** | v1.0 |
| **状态** | 已实现 |

---

## 1. 概述

KnowledgeRerank 是 agent-retrieval 内的内部复用模块，对候选概念列表进行 LLM 或 Vector 模式的重排序。**无独立 API**，由 KnRetrieval 模块调用。

---

## 2. 代码结构

### 2.1 文件清单

```
agent-retrieval/server/
├── logics/
│   └── knrerank/
│       └── knowledge_rerank.go   # 知识重排器实现
├── interfaces/
│   └── drivenadapters.go         # KnowledgeRerankReq、ConceptResult、KnowledgeRerankActionType
└── infra/config/
    └── config.go                 # RerankLLMConfig
```

### 2.2 调用关系

```
KnRetrieval (logics/knretrieval/rerank.go)
    └── rerankByDataRetrieval
        └── knReranker.Rerank(ctx, &KnowledgeRerankReq{...})
            └── knrerank.KnowledgeReranker (logics/knrerank/knowledge_rerank.go)
```

---

## 3. 核心实现

### 3.1 核心类：KnowledgeReranker

**文件**：`logics/knrerank/knowledge_rerank.go`

**职责**：对概念列表进行重排序，支持 LLM 模式和 Vector 模式，采用单例模式实现。

**依赖**：
- `interfaces.DrivenMFModelAPIClient`：Chat（LLM 模式）、Rerank（Vector 模式）
- `config.RerankLLMConfig`：LLM 参数（model、temperature、top_k、top_p 等）

**主入口**：`Rerank(ctx context.Context, req *KnowledgeRerankReq) ([]*ConceptResult, error)`

### 3.2 请求/响应结构（interfaces/drivenadapters.go）

**KnowledgeRerankReq**：
- `QueryUnderstanding`：查询理解结果（OriginQuery、Intent）
- `KnowledgeConcepts`：待排序概念列表
- `Action`：`llm` 或 `vector`

**ConceptResult**：
- `ConceptType`、`ConceptID`、`ConceptName`、`ConceptDetail`
- `RerankScore`：输出字段，重排后的分数

---

## 4. 依赖服务

| 服务 | 用途 | 接口 |
|------|------|------|
| MF Model API | LLM 对话（LLM 模式） | DrivenMFModelAPIClient.Chat |
| MF Model API | 向量 Rerank（Vector 模式） | DrivenMFModelAPIClient.Rerank |

---

## 5. 配置

**RerankLLMConfig**（infra/config）：
- `Model`：LLM 模型名称
- `Temperature`、`TopK`、`TopP`：采样参数
- `FrequencyPenalty`、`PresencePenalty`
- `MaxTokens`：最大输出 Token

---

## 6. 参考文档

- [KnowledgeRerank模块逻辑设计.md](./KnowledgeRerank模块逻辑设计.md)
