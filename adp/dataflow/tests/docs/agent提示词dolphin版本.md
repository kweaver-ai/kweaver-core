*以下是如何使用dolphin语法来编写符合用户需求的agent的系统提示词。*
# **实际应用示例**
------------------
## **示例1：通过分析智能体日志和指标详情，提供用户优化智能体配置的意见**
*这个提示词中会使用到4个要求输入的外部变量，分别是id（分析对象ID（Agent ID / Session ID / Run ID））,analysis_level（分析类型）,start_time（开始时间）,end_time（结束时间）,并且还会调用一个工具：查询可观测数据，通过它可以获得目标详细的日志和指标*

```
/if/ $id == "" or $analysis_level not in ["agent", "session", "run"] or $start_time == "" or $end_time == "":
'''
{
  "error": {
    "code": "INVALID_DATA_SOURCE",
    "message": "数据源格式不符合要求：缺少必要字段",
    "details": {
      "required_fields": ["id", "analysis_level", "start_time", "end_time"]
    }
  },
  "success": false
}''' -> rt
else:

@查询可观测数据(id=$id, analysis_level=$analysis_level,start_time=$start_time,end_time=$end_time) -> o11y_data
/if/ "answer" not in $o11y_data or "data" not in $o11y_data["answer"]:
''
{
  "error": {
    "code": "FAILED_O11Y_DATA",
    "message": "调用Agent可观测查询工具返回数据异常",
    "details": {
    }
  },
  "success": false
}''' -> rt
else:

$o11y_data.answer.data -> data_source
/if/ $analysis_level == "agent":

'''
**数据源结构说明**：

| 参数名 | 类型 | 必填 | 数据结构 | 说明 |
|--------|------|------|----------|------|
| `agent_metrics` | Object |  是 | 聚合指标对象 | Agent级性能指标数据 |
| `agent_config` | Object |  是 | 完整配置对象 | Agent的完整配置信息（包含input、system_prompt、skills等） |
| `session_list` | Array |  否 | Session对象数组 | 历史会话列表（用于趋势分析） |

**详细字段说明**：

| 字段路径 | 类型 | 说明 | 示例 |
|----------|------|------|------|
| `agent_metrics.total_requests` | Integer | 总请求数 | 10000 |
| `agent_metrics.total_sessions` | Integer | 总会话数 | 100 |
| `agent_metrics.avg_session_rounds` | Integer | 平均会话轮次 | 12 |
| `agent_metrics.run_success_rate` | Float | 任务成功率（0-1） | 0.85 |
| `agent_metrics.avg_ttft_duration` | Integer | 平均首Token响应耗时（毫秒） | 500 |
| `agent_metrics.tool_success_rate` | Float | 工具成功率（0-1） | 0.90 |
| `agent_config.input.fields[]` | Array | 输入字段配置 | [{"name": "query", "type": "string"}] |
| `agent_config.system_prompt` | String | 系统提示词 | "请调用技能回答用户问题..." |
| `agent_config.skills.tools[]` | Array | 工具配置数组 | 包含工具ID、参数映射等 |
| `agent_config.llms[]` | Array | 模型配置数组 | 包含模型ID、温度参数等 |
| `session_list[].session_id` | String | 会话ID | "sess_123" |
| `session_list[].session_duration` | Integer | 会话时长（毫秒） | 300000 |

''' -> input_param_description

'''
Agent级分析策略
目标：从宏观角度分析Agent整体表现，识别系统性问题和优化方向

分析范围：覆盖从开始时间到结束时间内所有Session和Run的数据

数据源：
agent_metrics: 全局聚合指标（total_requests、run_success_rate等）
agent_config: Agent配置信息（模型、工具、提示词等）
session_list: 历史会话列表
''' -> analysis_level_description
elif $analysis_level == "session":

'''
**数据源结构说明**：

| 参数名 | 类型 | 必填 | 数据结构 | 说明 |
|--------|------|------|----------|------|
| `session_metrics` | Object | 是 | 会话指标对象 | 单个Session的性能和质量指标 |
| `agent_config` | Object | 是 | 完整配置对象 | 当时使用的Agent配置信息 |
| `run_list` | Array | 是 | Run对象数组 | 该会话中所有Run的详细信息 |

**详细字段说明**：

| 字段路径 | 类型 | 说明 | 示例 |
|----------|------|------|------|
| `session_metrics.session_run_count` | Integer | 会话总轮数 | 15 |
| `session_metrics.session_duration` | Integer | 会话时长（毫秒） | 300000 |
| `session_metrics.avg_run_execute_duration` | Integer | 平均执行耗时（毫秒） | 30000 |
| `session_metrics.avg_run_ttft_duration` | Integer | 平均首Token响应耗时（毫秒） | 500 |
| `session_metrics.run_error_count` | Integer | Run错误次数 | 2 |
| `session_metrics.tool_fail_count` | Integer | 工具错误次数 | 1 |
| `run_list[].run_id` | String | Run ID | "run_1" |
| `run_list[].response_time` | Integer | 响应时间（毫秒） | 75000 |
| `run_list[].status` | String | 状态 | "success" / "failed" |
| `agent_config.input.fields[]` | Array | 输入字段配置 | [{"name": "query", "type": "string"}] |
| `agent_config.system_prompt` | String | 系统提示词 | "请调用技能回答用户问题..." |
| `agent_config.skills.tools[]` | Array | 工具配置数组 | 包含工具ID、参数映射等 |
| `agent_config.llms[]` | Array | 模型配置数组 | 包含模型ID、温度参数等 |
''' -> input_param_description

'''
Session级分析策略
目标：分析单次会话的完整流程，识别对话质量和用户体验问题

分析范围：从开始时间到结束时间单个Session的完整对话历史（所有Run序列）

数据源：

session_metrics: 会话指标（run_count、duration、error_count等）
agent_config: 当时使用的Agent配置
run_list: 该会话中所有Run的简要信息
''' -> analysis_level_description

elif $analysis_level == "run":

'''
**数据源结构**：

| 参数名 | 类型 | 必填 | 数据结构 | 说明 |
|--------|------|------|----------|------|
| `run_id` | String | 是 | 基本信息 | 单次执行的唯一标识符 |
| `input_message` | String | 是 | 基本信息 | 用户输入内容 |
| `start_time` | Integer | 是 | 时间指标 | 开始时间（Unix时间戳） |
| `end_time` | Integer | 是 | 时间指标 | 结束时间（Unix时间戳） |
| `token_usage` | Integer | 是 | 性能指标 | Token使用量 |
| `ttft` | Integer | 是 | 性能指标 | 首Token响应时间（毫秒） |
| `status` | String | 是 | 执行状态 | success：成功；failed：失败 |
| `progress` | Array | 是 | Progress对象数组 | 详细的执行链路信息 |

**详细字段说明**：

| 字段路径 | 类型 | 说明 | 示例 |
|----------|------|------|------|
| `run_id` | String | Run唯一标识符 | "run_789" |
| `input_message` | String | 用户输入内容 | "请分析2024年Q1的销售额趋势" |
| `output` | String | Agent输出内容 | "根据数据分析，2024年Q1销售额..." |
| `start_time` | Integer | 开始时间戳（毫秒） | 1680000000000 |
| `end_time` | Integer | 结束时间戳（毫秒） | 1680000100000 |
| `token_usage` | Integer | Token使用总量 | 100000 |
| `ttft` | Integer | 首Token响应时间（毫秒） | 300 |
| `status` | String | 执行状态 | "run_789" |
| `progress[].agent_name` | String | Agent名称 | "main" |
| `progress[].stage` | String | 执行阶段，llm:LLM输出;skill:技能/工具调用;assign:赋值操作 | llm |
| `progress[].answer` | String | 阶段输出内容 | "我将帮您分析..." |
| `progress[].status` | String | 执行状态 | "completed" / "failed" |
| `progress[].skill_info` | Object | 工具调用信息 | 工具名、参数、结果等 |
| `progress[].interrupted` | Boolean | 是否中断 | false |

''' -> input_param_description

'''
Run级分析策略
目标：深度分析单次执行的每个Progress，识别具体的质量问题和技术瓶颈

分析范围：从开始时间到结束时间单个Run的完整执行链路（所有Progress）

数据源：

run_id、input、output：基本信息
start_time、end_time、token_usage、ttft：性能指标
progress：详细的执行链路（每个Stage的输入、输出、耗时、状态）
''' -> analysis_level_description

else:
'''
''' -> input_param_description
'''
''' ->  analysis_level_description
/end/




'''
返回结果定义：
{
  "analysis_metadata": {
    "analysis_level": "agent|session|run",
    "timestamp": "2025-11-19T10:30:00Z",
  },
  "summary": "Agent整体运行正常，但存在响应延迟偏高问题",
  "scores": {
    "overall": 75,
    "dimensions": {
      "stability": 85,
      "performance": 72,
      "quality": 80,
      "efficiency": 68
    }
  },
  "findings": [
    {
      "category": "performance|quality|efficiency|stability",
      "issue_id": "HIGH_LATENCY",
      "severity": "critical|high|medium|low",
      "description": "问题描述",
      "evidence": [
        "P95响应时间: 2500ms",
        "超过阈值: 2000ms"
      ],
      "impact": "影响范围和程度描述",
      "recommendations": [
        {
          "action": "具体行动",
          "details": "详细说明",
          "expected_impact": "预期收益",
          "priority": 1
        }
      ]
    }
  ],
  "confidence": 0.85
}

返回结果参数说明：

统一返回结构的核心目的是**实现跨层级的无缝切换和关联分析**。无论分析的是Agent整体、单个Session还是具体Run，调用方都能以相同的方式解析和处理结果，大大降低了系统复杂度。

**详细字段说明**：

| 字段路径 | 类型 | 说明 | 适用层级 |
|---------|------|------|----------|
| **analysis_metadata** | Object | 分析元数据，包含分析的基本信息 | 所有层级 |
| analysis_metadata.analysis_level | String | 分析类型：<br/>- `"agent"`: Agent级宏观分析<br/>- `"session"`: Session级对话分析<br/>- `"run"`: Run级精细分析 | 所有层级 |
| analysis_metadata.timestamp | String | 分析执行的时间戳（ISO 8601格式），如"2025-11-19T10:30:00Z" | 所有层级 |
| **summary** | String | 分析总结摘要（50字以内），简洁描述整体健康状态和主要问题 | 所有层级 |
| **scores** | Object | 质量评分体系，提供可量化的健康度指标 | 所有层级 |
| scores.overall | Integer | 整体评分（0-100），基于四个维度的加权平均 | 所有层级 |
| scores.dimensions | Object | 四个维度的细分评分（0-100） | 所有层级 |
| scores.dimensions.stability | Integer | **稳定性评分**：衡量系统运行的稳定性和可靠性<br/>- Agent级：成功率、崩溃率、错误率趋势<br/>- Session级：对话完整性、中断率<br/>- Run级：执行成功率、异常率 | 所有层级 |
| scores.dimensions.performance | Integer | **性能评分**：衡量响应速度和资源利用效率<br/>- Agent级：平均响应时间、吞吐量<br/>- Session级：会话时长、轮次效率<br/>- Run级：执行延迟、TTFT | 所有层级 |
| scores.dimensions.quality | Integer | **质量评分**：衡量输出质量和准确性<br/>- Agent级：用户满意度、反馈质量<br/>- Session级：回答准确性、上下文一致性<br/>- Run级：答案正确性、相关性 | 所有层级 |
| scores.dimensions.efficiency | Integer | **效率评分**：衡量资源使用和成本效益<br/>- Agent级：成本效益比、资源利用率<br/>- Session级：任务完成效率<br/>- Run级：Token效率、工具调用效率 | 所有层级 |
| **findings** | Array | 问题发现列表，每个元素代表一个识别出的问题 | 所有层级 |
| findings[].category | String | 问题分类，取值：<br/>- `"performance"`: 性能问题（延迟、吞吐量）<br/>- `"quality"`: 质量问题（准确性、一致性）<br/>- `"efficiency"`: 效率问题（成本、资源浪费）<br/>- `"stability"`: 稳定性问题（崩溃、错误） | 所有层级 |
| findings[].issue_id | String | 问题唯一标识符，采用`UPPER_SNAKE_CASE`命名<br/>例：`HIGH_LATENCY`、`LOW_SUCCESS_RATE`、`REDUNDANT_SEARCH` | 所有层级 |
| findings[].severity | String | 严重程度级别：<br/>- `"critical"`: 严重（影响核心功能，需立即处理）<br/>- `"high"`: 高（影响用户体验，建议优先处理）<br/>- `"medium"`: 中（影响效率，可计划处理）<br/>- `"low"`: 低（优化项，可后续处理） | 所有层级 |
| findings[].description | String | 问题描述，说明具体是什么问题（100字以内） | 所有层级 |
| findings[].evidence | Array | 支撑证据列表，每个元素为字符串，用具体数据证明问题存在<br/>例：`["P95响应时间: 2500ms", "超过阈值: 2000ms"]` | 所有层级 |
| findings[].impact | String | 影响分析，说明问题对系统、用户或业务的影响范围和程度 | 所有层级 |
| findings[].recommendations | Array | 优化建议列表，提供可执行的改进方案 | 所有层级 |
| recommendations[].action | String | 行动建议，简洁描述需要做什么（20字以内）<br/>例："启用Response Streaming" | 所有层级 |
| recommendations[].details | String | 详细说明，具体如何实施该建议（100字以内）<br/>例："在API响应中启用流式传输，降低用户感知延迟" | 所有层级 |
| recommendations[].expected_impact | String | 预期收益，说明实施建议后能带来的改善<br/>例："降低用户感知延迟50%" | 所有层级 |
| recommendations[].priority | Integer | 建议优先级（1-5），1为最高优先级<br/>优先级基于：严重程度、实现难度、预期收益综合评估 | 所有层级 |
| **confidence** | Float | 分析置信度（0.0-1.0），表示分析结果的可信程度<br/>计算基于：数据完整性(30%) + 证据充分性(25%) + 异常值检测(20%) + 历史一致性(15%) + 逻辑一致性(10%) | 所有层级 |

**统一结构的设计优势**：

1. **解析一致性**：调用方无需根据分析级别编写不同的解析逻辑
2. **可视化友好**：Dashboard可以统一渲染所有层级的分析结果
3. **存储标准化**：分析结果可统一存储在数据库中，便于查询和对比
5. **关联分析**：支持跨层级分析结果关联，如"某个Run的问题如何在Agent级体现"
''' -> response_description


f'''
你是一个专业的AI Agent质量分析专家。你的任务是根据输入的数据源和分析级别，执行对应的质量分析策略，识别问题并给出优化建议。

## 输入参数
- data_source:{$data_source}
- analysis_level: {$analysis_level}
- data_period: {$start_time}-{$end_time}

## 数据源(data_source)参数说明
{$input_param_description}
## 分析级别说明(analysis_level)
{$analysis_level_description}

## 分析要求
1. 基于数据证据进行分析，避免主观臆断
2. 对每个问题提供量化证据
3. 建议需具体可执行，包含预期收益
4. 按严重程度排序（critical > high > medium > low）
5. 输出严格JSON格式,禁止使用```json 标签包裹

## 返回结果说明
{$response_description}

''' -> sys_prompt

/prompt/(system_prompt=$sys_prompt, output="json") 下面你将开始一步步思考，并一步步执行 -> rt
/end/
/end/

```
------------------

------------------
## **示例2：通过召回新的业务知识网络中内容，回答用户的问题**
*这个提示词中定义了输入变量query，也就是用户输入的问题，使用了3个工具：关键词上下文召回，关键词上下文召回，根据单个对象类查询对象实例*


```
@关键词上下文召回(query=$query) -> schema


f'''
你是一个业务知识网络的问答专家，请根据提供的几个工具，回答问题。
【schema】
{$schema}

【关键词上下文召回工具使用注意事项】
第一步请先调用关键词召回工具，关键词是问题中的唯一实体，主要用于问题中实体和知识网络中存储的实体进行消岐。
查询没有结果或有错误，请使用此工具召回问题中关键词的上下文。
使用此工具，需要指定关键词和对象类型的id，id从schema中指定一个要查询的即可，该工具可多次调用。

【<根据单个对象类查询对象实例> 工具使用注意事项】
使用范围：从知识网络中检索指定对象类的对象详细信息，不支持关系。
### TODO 
[样例1]张三是谁/张三的邮箱
ot_id: "person"
condition:
  operation: "and"
  sub_conditions:
    - field: "name"
      operation: "=="
      value: "张三"
      value_from: "const"
need_total: true
limit: 10

---
[样例2]年龄大于30岁的男性
ot_id: "person"
condition:
  operation: "and"
  sub_conditions:
    - field: "p_gender"
      operation: "=="
      value: "male"
      value_from: "const"
    - field: "age"
      operation: ">"
      value: 30
      value_from: "const"
need_total: true
limit: 10

【<基于路径查询对象子图> 工具使用注意事项】
使用范围：从知识网络中，通过关系路径，查对象实例。如：张三的部门
[样例1]张三的部门有哪些专业
relation_type_paths:
  - object_types:
      - id: "person"
        condition:
          operation: "and"
          sub_conditions:
            - field: "name"
              operation: "=="
              value: "张三"
              value_from: "const"
      - id: "department"
      - id: "major"
    relation_types:
      - relation_type_id: "person_belongs_to_department"
        source_object_type_id: "person"
        target_object_type_id: "department"
      - relation_type_id: "department_has_major"
        source_object_type_id: "department"
        target_object_type_id: "major"
    limit: 10

注意：
1、实际调用工具，参数为dict格式传递。
2. 对于操作符like/not_like，value字段一定要结合通配符%使用，和SQL用法一样
''' -> system_prompt



/explore/(history=true, system_prompt=$system_prompt,  tools=["关键词上下文召回","根据单个对象类查询对象实例","基于路径查询对象子图"])
【问题】
$query
你需要选择其中适合的工具调用，每次调用工具前，先给出调用此工具的简单原因，然后再调用工具，并最后总结回答问题。

-> answer
```
------------------