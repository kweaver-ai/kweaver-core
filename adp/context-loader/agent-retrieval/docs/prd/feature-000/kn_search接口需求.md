# 需求：代理 kn_search 接口

## 目标
在当前服务中代理 data-retrieval 服务的 kn_search 接口，统一对外提供服务，便于 Agent/大模型调用。

## 接口信息

### 代理接口
- **路由**: `POST /kn/kn_search`
- **目标接口**: `POST http://data-retrieval:9100/tools/kn_search`

### 功能说明
基于知识网络的智能检索工具，支持传入完整的问题或一个或多个关键词，能够检索问题或关键词的属性信息和上下文信息：
- 支持概念召回：返回相关的对象类型、关系类型、操作类型
- 支持语义实例召回：返回匹配的实例节点数据
- 支持多轮对话：通过 session_id 维护历史召回记录
- 支持灵活的召回配置：概念召回、语义实例召回、属性过滤等参数可配置

### 请求参数
- **Header**:
  - `x-account-id` (string, 可选): 账户ID，用于内部服务调用时传递账户信息
  - `x-account-type` (string, 可选, 默认 "user"): 账户类型，可选值：user(用户), app(应用), anonymous(匿名)
  - `Content-Type` (string, 可选, 默认 "application/json"): 内容类型
- **Body**:
  - `query` (string, 必需): 用户查询问题或关键词，多个关键词之间用空格隔开
  - `kn_ids` (array, 必需): 指定的知识网络配置列表，每个配置包含 `knowledge_network_id` 字段
  - `session_id` (string, 可选): 会话ID，用于维护多轮对话存储的历史召回记录
  - `additional_context` (string, 可选): 额外的上下文信息，用于二次检索时提供更精确的检索信息
  - `retrieval_config` (object, 可选): 召回配置参数，用于控制不同类型的召回场景（概念召回、语义实例召回、属性过滤）。如果不提供，将使用系统默认配置
    - `concept_retrieval` (object, 可选): 概念召回配置参数
      - `top_k` (integer, 默认 10): 概念召回返回最相关关系类型数量
      - `skip_llm` (boolean, 默认 true): 是否跳过LLM筛选相关关系类型
      - `return_union` (boolean, 默认 false): 概念召回多轮检索时是否返回并集
      - `include_sample_data` (boolean, 默认 false): 是否获取对象类型的样例数据
      - `schema_brief` (boolean, 默认 true): 概念召回时是否返回精简schema
      - `enable_coarse_recall` (boolean, 默认 true): 是否启用对象/关系粗召回
      - `coarse_object_limit` (integer, 默认 2000): 对象类型粗召回的最大返回数量
      - `coarse_relation_limit` (integer, 默认 300): 关系类型粗召回的最大返回数量
      - `coarse_min_relation_count` (integer, 默认 5000): 启用粗召回的关系类型总数阈值
      - `enable_property_brief` (boolean, 默认 true): 是否对返回的对象属性做相关性裁剪
      - `per_object_property_top_k` (integer, 默认 8): 每个对象类型最多保留的属性数量
      - `global_property_top_k` (integer, 默认 30): 全局最多保留的属性总数量
    - `semantic_instance_retrieval` (object, 可选): 语义实例召回配置参数
      - `initial_candidate_count` (integer, 默认 50): 语义实例召回的初始召回数量上限
      - `per_type_instance_limit` (integer, 默认 5): 每个对象类型最终返回的实例数量上限
      - `max_semantic_sub_conditions` (integer, 默认 10): 语义实例召回构造查询条件时 sub_conditions 的最大数量上限
      - `semantic_field_keep_ratio` (number, 默认 0.2): 语义字段筛选保留比例
      - `semantic_field_keep_min` (integer, 默认 5): 语义字段筛选最少保留字段数
      - `semantic_field_keep_max` (integer, 默认 15): 语义字段筛选最多保留字段数
      - `semantic_field_rerank_batch_size` (integer, 默认 128): 字段语义打分时的批处理大小
      - `min_direct_relevance` (number, 默认 0.3): 直接相关性最低阈值（0-1之间）
      - `enable_global_final_score_ratio_filter` (boolean, 默认 true): 是否启用全局 final_score 相对阈值过滤
      - `global_final_score_ratio` (number, 默认 0.25): 全局 final_score 相对阈值比例（0~1）
      - `exact_name_match_score` (number, 默认 0.85): 多关键词检索场景下的实例名完全相等保底分（0~1）
    - `property_filter` (object, 可选): 实例属性过滤配置
      - `max_properties_per_instance` (integer, 默认 20): 每个实例最多返回的属性字段数量
      - `max_property_value_length` (integer, 默认 500): 属性值的最大长度（字符数）
      - `enable_property_filter` (boolean, 默认 true): 是否启用实例属性过滤
  - `only_schema` (boolean, 可选, 默认 false): 是否只召回概念（schema），不召回语义实例。如果为True，则只返回object_types、relation_types和action_types，不返回nodes

### 响应结构
- `object_types` (array): 对象类型列表（概念召回时返回）
  - 精简模式（schema_brief=True）: 包含 concept_id, concept_name, comment, data_properties（仅name和display_name）, logic_properties（仅name和display_name）, sample_data（当include_sample_data=True时）
  - 完整模式（schema_brief=False）: 包含完整字段（包括primary_keys, display_key, sample_data等）
- `relation_types` (array): 关系类型列表（概念召回时返回），包含 concept_id, concept_name, source_object_type_id, target_object_type_id
- `action_types` (array): 操作类型列表（概念召回时返回）
  - 精简模式: 包含 id, name, action_type, object_type_id, object_type_name, comment, tags, kn_id
- `nodes` (array, 可选): 语义实例召回结果，与条件召回节点风格对齐的扁平列表
  - 每个节点至少包含 object_type_id、<object_type_id>_name、unique_identities
- `message` (string, 可选): 提示信息（例如未召回到实例数据时返回原因说明）

## 特殊要求
1. **错误信息优化**: 错误信息需包装成大模型可理解的内容
2. **未来扩展**: 可能封装成 MCP Server 对外提供服务
