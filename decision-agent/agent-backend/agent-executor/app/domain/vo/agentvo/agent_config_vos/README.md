# Agent Config Skills VO 结构说明

## 概述

本目录包含了Agent配置中skills字段的VO（Value Object）结构定义，用于替代原来的字典结构，提供更好的类型安全和代码可维护性。

## 文件结构

```
agent_config_vos/
├── __init__.py              # 导出所有VO类和枚举
├── skill_vo.py              # 技能配置主VO
├── tool_vo.py               # 工具技能VO
├── agent_vo.py              # Agent技能VO
├── mcp_vo.py                # MCP技能VO
├── skill_input_vo.py        # 技能输入参数VO
└── README.md                # 本文档
```

## 类结构

### 1. SkillVo (skill_vo.py)

技能配置的主VO类，包含三种类型的技能：

```python
class SkillVo(BaseModel):
    tools: Optional[List[ToolVo]] = []      # 工具列表
    agents: Optional[List[AgentVo]] = []    # Agent列表
    mcps: Optional[List[McpVo]] = []        # MCP列表
```

### 2. ToolVo (tool_vo.py)

工具技能配置：

```python
class ToolVo(BaseModel):
    tool_id: str                                    # 工具ID
    tool_box_id: str                                # 工具箱ID
    tool_input: Optional[List[SkillInputVo]] = []   # 输入参数配置
    intervention: Optional[bool] = False            # 是否启用干预
    result_process_strategies: Optional[List[...]]  # 结果处理策略
```

### 3. AgentVo (agent_vo.py)

Agent技能配置：

```python
class AgentVo(BaseModel):
    agent_key: str                                  # Agent key
    agent_version: Optional[str] = "latest"         # Agent版本
    agent_input: Optional[List[AgentInputVo]] = []  # 输入参数配置
    intervention: Optional[bool] = False            # 是否启用干预
    data_source_config: Optional[DataSourceConfigVo] = None  # 数据源配置
    llm_config: Optional[LlmConfigVo] = None        # 大模型配置
    ...
```

### 4. McpVo (mcp_vo.py)

MCP技能配置：

```python
class McpVo(BaseModel):
    mcp_server_id: str  # MCP server ID
```

### 5. SkillInputVo (skill_input_vo.py)

技能输入参数配置，支持嵌套参数结构：

```python
class SkillInputVo(BaseModel):
    enable: Optional[bool]          # 是否启用
    input_name: str                 # 输入参数名称（必传）
    input_type: str                 # 输入参数类型（必传）
    input_desc: Optional[str]       # 输入参数描述
    map_type: Optional[str]         # 映射类型
                                    # 可选值: fixedValue/var/model/auto
    map_value: Optional[Any]        # 映射值
    children: Optional[List["SkillInputVo"]]  # 子参数列表（用于嵌套参数）

```

**支持的映射类型**：
- `fixedValue`: 固定值映射
- `var`: 变量映射（从变量池获取）
- `model`: 选择模型映射
- `auto`: 模型自动生成映射（默认值）

## 使用示例

### 1. 创建SkillVo对象

```python
from app.domain.vo.agentvo.agent_config_vos import (
    SkillVo, ToolSkillVo, AgentSkillVo, McpSkillVo, SkillInputVo
)

# 创建空的技能配置
skill = SkillVo()

# 创建带工具的技能配置
tool = ToolSkillVo(
    tool_id="test_tool",
    tool_box_id="test_toolbox",
    tool_input=[
        SkillInputVo(
            enable=True,
            input_name="query",
            input_type="string",
            map_type="auto",
            map_value="test"
        )
    ]
)
skill = SkillVo(tools=[tool])

# 创建带嵌套参数的工具配置
tool_with_nested = ToolSkillVo(
    tool_id="complex_tool",
    tool_box_id="test_toolbox",
    tool_input=[
        SkillInputVo(
            input_name="action",
            input_type="string",
            input_desc="工具行为类型",
            map_type="auto",
            enable=True
        ),
        SkillInputVo(
            input_name="config",
            input_type="object",
            input_desc="工具配置参数",
            children=[
                SkillInputVo(
                    input_name="session_id",
                    input_type="string",
                    input_desc="会话ID",
                    map_type="var",
                    map_value="self_config.conversation_id",
                    enable=True
                ),
                SkillInputVo(
                    input_name="retry_times",
                    input_type="integer",
                    input_desc="重试次数",
                    map_type="auto",
                    enable=False
                )
            ]
        )
    ]
)
skill = SkillVo(tools=[tool_with_nested])
```

### 2. 在AgentConfigVo中使用

```python
from app.domain.vo.agentvo.agent_config import AgentConfigVo

# 使用None（会自动转换为SkillVo对象）
config = AgentConfigVo(
    input={},
    llms=[],
    skills=None  # validator会将None转换为SkillVo()
)

# 使用字典（会自动转换为SkillVo对象）
config = AgentConfigVo(
    input={},
    llms=[],
    skills={
        "tools": [...],
        "agents": [...],
        "mcps": [...]
    }
)

# 直接使用SkillVo对象
config = AgentConfigVo(
    input={},
    llms=[],
    skills=SkillVo(tools=[...], agents=[...])
)
```

### 3. 访问技能配置

```python
# 访问工具列表
for tool in config.skills.tools:
    print(tool.tool_id, tool.tool_box_id)

# 访问Agent列表
for agent in config.skills.agents:
    print(agent.agent_key, agent.agent_version)

# 访问MCP列表
for mcp in config.skills.mcps:
    print(mcp.mcp_server_id)
```

### 4. 添加技能

```python
# 添加工具
new_tool = ToolVo(tool_id="new_tool", tool_box_id="toolbox")
config.skills.tools.append(new_tool)

# 添加Agent
new_agent = AgentVo(agent_key="new_agent", agent_version="latest")
config.skills.agents.append(new_agent)
```

## 枚举类型

### InputTypeEnum
- `STRING`: 字符串类型
- `FILE`: 文件类型
- `OBJECT`: 对象类型

### MapTypeEnum
- `FIXED_VALUE`: 固定值
- `VAR`: 引用变量
- `MODEL`: 选择模型
- `AUTO`: 模型生成

### DataSourceTypeEnum
- `INHERIT_MAIN`: 继承主Agent数据源
- `SELF_CONFIGURED`: 使用自身配置

### SpecificInheritEnum
- `DOCS_ONLY`: 仅继承文档数据源
- `GRAPH_ONLY`: 仅继承图谱数据源
- `ALL`: 继承所有类型数据源

### LlmConfigTypeEnum
- `INHERIT_MAIN`: 继承主Agent大模型
- `SELF_CONFIGURED`: 使用自身配置

## 向后兼容性

为了保持向后兼容性，`AgentConfigVo`中的`skills`字段支持以下输入：

1. **None**: 自动转换为空的`SkillVo()`对象
2. **字典**: 自动转换为`SkillVo`对象
3. **SkillVo对象**: 直接使用

这确保了现有代码可以继续使用字典形式传入skills配置。

## 注意事项

1. **动态属性**: 在某些处理流程中（如`process_skills`），会向VO对象动态添加属性（如`tool_info`、`HOST_AGENT_OPERATOR`等）。这些属性通过`__dict__`直接设置。

2. **类型安全**: 使用VO对象后，IDE可以提供更好的代码补全和类型检查。

3. **验证**: Pydantic会自动验证字段类型和必填项，提供更好的数据完整性保证。

## 迁移指南

### 从字典访问迁移到对象访问

**旧代码**:
```python
for tool in config.skills["tools"]:
    tool_id = tool["tool_id"]
    tool_box_id = tool["tool_box_id"]
```

**新代码**:
```python
for tool in config.skills.tools:
    tool_id = tool.tool_id
    tool_box_id = tool.tool_box_id
```

### 从字典操作迁移到对象操作

**旧代码**:
```python
config.skills["tools"].append({
    "tool_id": "test",
    "tool_box_id": "toolbox"
})
```

**新代码**:
```python
config.skills.tools.append(ToolVo(
    tool_id="test",
    tool_box_id="toolbox"
))
```

## 测试

测试文件位于: `test/vo_test/test_skill_vo.py`

运行测试:
```bash
pytest test/vo_test/test_skill_vo.py -v
```
