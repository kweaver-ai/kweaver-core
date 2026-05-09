# AI 生成功能设计方案

## 1. 功能概述

本方案旨在通过集成大语言模型（LLM）能力，辅助算子开发过程。主要提供以下两类核心功能：
1. **Python 函数生成**：根据用户自然语言描述、输入/输出定义，自动生成符合规范的 Python 业务逻辑代码。
2. **元数据生成**：根据已有的 Python 代码及上下文，自动生成或补全算子的元数据（Input/Output Schema、名称、描述等）。

## 2. 总体架构

### 2.1 架构分层
功能实现主要位于 `operator-integration` 服务中，依赖外部的模型服务接口（MFModelAPI）。

* **接入层 (Driver Adapters)**: 提供 RESTful 接口，处理 HTTP 请求上下文、参数校验及流式响应（SSE）封装。
* **业务逻辑层 (Logics)**:
    * `AIGenerationService`: 核心服务，负责组装 Prompt、调用 LLM、处理响应结果。
    * `PromptLoader`: 提示词加载器，支持"本地默认模板"与"远程配置模板"的双重加载策略。
* **基础设施层 (Infra)**:
    * `MFModelAPIClient`: 统一的模型服务客户端。
    * `Config`: 包含 LLM 参数配置及 Prompt ID 映射。

### 2.2 核心流程
1. **请求接收**: 用户发起生成请求（指定类型：代码生成/元数据生成）。
2. **模板加载**: `PromptLoader` 根据类型获取 System Prompt。
   * 优先检查配置文件中是否指定了自定义 `PromptID`。
   * 若配置了且有效，从 `MFModelManager` 获取远程模板。
   * 若未配置或获取失败，回退使用本地嵌入（Embed）的默认 `.md` 模板。
3. **参数组装**: 将用户输入（Query/Code/Inputs/Outputs）填入 User Prompt 模板。
4. **模型调用**: 调用 LLM 接口（支持普通调用与流式调用）。
5. **结果处理**:
   * 代码生成：直接返回生成的 Python 代码字符串。
   * 元数据生成：尝试解析 JSON 结果并反序列化为结构体。

## 3. 接口设计

### 3.1 AI 辅助生成接口
**路径**: `POST /api/agent-operator-integration/v1/ai_generate/function/:type`

**Path 参数**:
* `type`: 模板类型
    * `python_function_generator`: Python 函数生成
    * `metadata_param_generator`: 元数据参数生成

**请求 Body (JSON)**:
```json
{
  "query": "string",          // [必填-代码生成] 用户需求描述
  "code": "string",           // [必填-元数据生成] Python 代码内容
  "inputs": [ ... ],          // [可选] 已有的输入定义
  "outputs": [ ... ],         // [可选] 已有的输出定义
  "stream": boolean           // [可选] 是否开启流式返回 (SSE)
}
```

**响应 Body (非流式)**:
```json
{
  "content": "Result Context" // 可能是字符串(代码)或JSON对象(元数据)
}
```

**响应 (流式 SSE)**:
`Content-Type: text/event-stream`
```text
data: chunk_content
...
data: [DONE]
```

### 3.2 获取当前 Prompt 模板接口
**路径**: `GET /api/agent-operator-integration/v1/ai_generate/prompt/:type`

**功能**: 返回当前生效的 System Prompt 和 User Prompt 模板，用于前端展示或调试。

**响应 Body**:
```json
{
  "prompt_id": "string",
  "name": "string",
  "description": "string",
  "system_prompt": "string",
  "user_prompt_template": "string"
}
```

## 4. 配置说明

配置主要位于 `configmap.yaml` 中的 `agent-operator-integration.yaml` 部分，在该文件的 `ai_generation_config` 节点下。

### 4.1 配置文件结构
```yaml
ai_generation_config:
  # 自定义 Prompt ID 配置
  # 如果配置了 ID，系统将尝试从模型管理服务加载对应的 Prompt 内容
  # 如果留空，则使用系统内置的默认模板
  python_function_generator_prompt_id: "your-prompt-id-for-python"
  metadata_param_generator_prompt_id: "your-prompt-id-for-metadata"

  # LLM 模型参数配置
  llm:
    model: ""                 # 模型名称，留空则使用服务侧默认
    max_tokens: 2048          # 最大 Token 数
    temperature: 0.1          # 温度 (代码生成建议低温度)
    top_k: 40
    top_p: 0.9
    frequency_penalty: 0.0
    presence_penalty: 0.0
```

### 4.2 配置热变更
目前配置变更需要重新部署或重启 Pod 才能生效（依赖 ConfigMap 挂载更新）。

1. **修改 Helm Values**: 更新 `values.yaml` 或直接编辑 ConfigMap。
2. **重载应用**: 重启 `agent-operator-integration` Pod。

## 5. 提示词 (Prompt) 管理细节

### 5.1 本地默认模板
系统内置了兜底的 Prompt 模板，位于源码 `server/logics/aigeneration/templates/` 目录：
* `Python_Function_Generator.md`: 包含严格的 Python 代码生成约束、错误处理模板、测试块要求。
* `Metadata_Param_Generator.md`: 包含元数据 Schema 定义、JSON 输出约束。

### 5.2 如何变更 Prompt
我们支持两种方式变更 Prompt，无需修改代码：

**方式一：配置远程 Prompt ID (推荐)**
1. 在模型管理平台/服务中创建新的 Prompt，调试至满意。
2. 获取该 Prompt 的 ID。
3. 修改服务的 ConfigMap，将对应功能的 `*_prompt_id` 更新为新 ID。
4. 重启服务。
   * *优点*: 可以动态管理，利用外部平台的版本控制。

**方式二：直接修改源码模板 (开发阶段)**
1. 修改 `server/logics/aigeneration/templates/` 下的 `.md` 文件。
2. 重新编译并构建镜像。
   * *优点*: 此时为硬编码默认值，适合固化核心逻辑。

## 6. 模型服务依赖
本服务依赖 `MFModelAPI` 和 `MFModelManager` 进行实际的模型交互。
* **MFModelAPI**: 用于发送 ChatCompletion 请求。
* **MFModelManager**: 用于根据 PromptID 查询 Prompt 详情。

确保 `depServices` 中关于 `mf-model-api` 和 `mf-model-manager` 的地址配置正确。
