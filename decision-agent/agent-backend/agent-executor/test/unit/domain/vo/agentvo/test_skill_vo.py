# -*- coding:utf-8 -*-
import pytest
from app.domain.vo.agentvo.agent_config_vos import (
    SkillVo,
    ToolSkillVo,
    AgentSkillVo,
    McpSkillVo,
    SkillInputVo,
    AgentInputVo,
    DataSourceConfigVo,
    LlmConfigVo,
    OutputConfigVo,
    OutputVariablesVo,
    DefaultFormatEnum,
    DataSourceTypeEnum,
    LlmConfigTypeEnum,
)
from app.domain.vo.agentvo.agent_config import AgentConfigVo


def test_skill_vo_creation():
    """测试SkillVo的创建"""
    skill = SkillVo()
    assert skill.tools == []
    assert skill.agents == []
    assert skill.mcps == []


def test_skill_vo_with_data():
    """测试带数据的SkillVo创建"""
    tool = ToolSkillVo(
        tool_id="test_tool_id",
        tool_box_id="test_toolbox_id",
        tool_input=[
            SkillInputVo(
                enable=True,
                input_name="query",
                input_type="string",
                map_type="auto",
                map_value="test",
            )
        ],
    )

    agent = AgentSkillVo(
        agent_key="test_agent",
        agent_version="latest",
        agent_input=[
            AgentInputVo(
                enable=True,
                input_name="input1",
                input_type="string",
                map_type="auto",
                map_value="value1",
            )
        ],
        data_source_config=DataSourceConfigVo(type=DataSourceTypeEnum.SELF_CONFIGURED),
        llm_config=LlmConfigVo(type=LlmConfigTypeEnum.SELF_CONFIGURED),
    )

    mcp = McpSkillVo(mcp_server_id="test_mcp_id")

    skill = SkillVo(tools=[tool], agents=[agent], mcps=[mcp])

    assert len(skill.tools) == 1
    assert len(skill.agents) == 1
    assert len(skill.mcps) == 1
    assert skill.tools[0].tool_id == "test_tool_id"
    assert skill.agents[0].agent_key == "test_agent"
    assert skill.mcps[0].mcp_server_id == "test_mcp_id"


def test_agent_config_vo_with_skill():
    """测试AgentConfigVo中的skills字段"""
    config = AgentConfigVo(input={}, llms=[], skills=None)

    # validator应该将None转换为SkillVo对象
    assert config.skills is not None
    assert isinstance(config.skills, SkillVo)
    assert config.skills.tools == []
    assert config.skills.agents == []
    assert config.skills.mcps == []


def test_agent_config_vo_with_dict_skill():
    """测试AgentConfigVo接受字典形式的skills"""
    config = AgentConfigVo(
        input={},
        llms=[],
        skills={
            "tools": [
                {"tool_id": "tool1", "tool_box_id": "toolbox1", "tool_input": []}
            ],
            "agents": [],
            "mcps": [],
        },
    )

    # validator应该将字典转换为SkillVo对象
    assert isinstance(config.skills, SkillVo)
    assert len(config.skills.tools) == 1
    assert config.skills.tools[0].tool_id == "tool1"


def test_append_task_plan_agent():
    """测试append_task_plan_agent方法"""
    config = AgentConfigVo(input={}, llms=[], plan_mode={"is_enabled": True})

    config.append_task_plan_agent()

    assert len(config.skills.agents) == 1
    assert config.skills.agents[0].agent_key == "Task_Plan_Agent"
    assert config.skills.agents[0].agent_version == "latest"


def test_agent_config_vo_with_output():
    """测试AgentConfigVo中的output字段"""
    config = AgentConfigVo(input={}, llms=[])

    # 默认情况下output应该有默认值
    assert config.output is not None
    assert config.output.default_format == "markdown"
    assert config.output.variables is None


def test_agent_config_vo_with_dict_output():
    """测试AgentConfigVo接受字典形式的output"""
    config = AgentConfigVo(
        input={},
        llms=[],
        output={
            "default_format": "json",
            "variables": {
                "answer_var": "custom_answer",
                "doc_retrieval_var": "custom_doc",
                "other_vars": ["var1", "var2"],
            },
        },
    )

    # validator应该将字典转换为OutputConfigVo对象
    assert isinstance(config.output, OutputConfigVo)
    assert config.output.default_format == DefaultFormatEnum.JSON
    assert config.output.variables.answer_var == "custom_answer"
    assert config.output.variables.doc_retrieval_var == "custom_doc"
    assert config.output.variables.other_vars == ["var1", "var2"]

    # 测试get_all_vars方法
    all_vars = config.output.get_all_vars()
    expected_vars = [
        "custom_answer",
        "custom_doc",
        "graph_retrieval_res",
        "related_questions",
        "var1",
        "var2",
    ]
    assert all_vars == expected_vars


def test_agent_config_vo_with_output_object():
    """测试AgentConfigVo接受OutputConfigVo对象"""
    output_config = OutputConfigVo(
        default_format="markdown", variables=OutputVariablesVo()
    )

    config = AgentConfigVo(input={}, llms=[], output=output_config)

    # 应该直接返回传入的对象
    assert config.output is output_config
    assert config.output.default_format == DefaultFormatEnum.MARKDOWN

    # 测试get_all_vars方法（使用默认值）
    all_vars = config.output.get_all_vars()
    expected_vars = [
        "answer",
        "doc_retrieval_res",
        "graph_retrieval_res",
        "related_questions",
    ]
    assert all_vars == expected_vars


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
