"""单元测试 - logic/agent_core_logic_v2/dolphin_enhance 模块"""

from app.domain.vo.agentvo import AgentConfigVo
from app.domain.vo.agentvo.agent_config_vos import (
    AgentSkillVo,
    SkillVo,
    ToolSkillVo,
)
from app.logic.agent_core_logic_v2.dolphin_enhance import (
    build_latest_callable_name_map,
    materialize_dolphin_from_enhance,
    parse_dolphin_enhance_name_map,
    refresh_dolphin_from_enhance,
)


class TestDolphinEnhance:
    """测试 dolphin_enhance 相关逻辑"""

    def test_parse_dolphin_enhance_name_map(self):
        """测试从 dolphin_enhance 解析 callable 映射"""
        name_map = parse_dolphin_enhance_name_map(
            '@获取agent详情{tool:tool-1}(key="DocQA_Agent")->res\n'
            '@问答助手{agent:agent-1}(query=input)->answer\n'
            '@普通语句(query=input)->noop'
        )

        assert name_map == {
            ("tool", "tool-1"): "获取agent详情",
            ("agent", "agent-1"): "问答助手",
        }

    def test_build_latest_callable_name_map(self):
        """测试从已构建 skills 提取最新 callable 名称"""
        tool = ToolSkillVo(tool_id="tool-1", tool_box_id="toolbox-1")
        tool.__dict__["tool_info"] = {"name": "获取Agent详情"}

        agent = AgentSkillVo(agent_key="agent-1")
        agent.inner_dto.agent_info = {"name": "知识问答助手"}

        latest_name_map = build_latest_callable_name_map(
            SkillVo(tools=[tool], agents=[agent])
        )

        assert latest_name_map == {
            ("tool", "tool-1"): "获取Agent详情",
            ("agent", "agent-1"): "知识问答助手",
        }

    def test_materialize_dolphin_from_enhance_uses_latest_names(self):
        """测试根据最新 callable 名称重新物化普通 dolphin"""
        dolphin = materialize_dolphin_from_enhance(
            dolphin='@获取agent详情(key="DocQA_Agent")->res\n@问答助手(query=input)->answer',
            dolphin_enhance='@获取agent详情{tool:tool-1}(key="DocQA_Agent")->res\n'
            '@问答助手{agent:agent-1}(query=input)->answer',
            latest_name_map={
                ("tool", "tool-1"): "获取Agent详情",
                ("agent", "agent-1"): "知识问答助手",
            },
        )

        assert dolphin == (
            '@获取Agent详情(key="DocQA_Agent")->res\n'
            "@知识问答助手(query=input)->answer"
        )

    def test_materialize_dolphin_from_enhance_keeps_old_name_when_id_missing(self):
        """测试找不到最新 callable 时保留原始名称"""
        dolphin = materialize_dolphin_from_enhance(
            dolphin='@获取agent详情(key="DocQA_Agent")->res',
            dolphin_enhance='@获取agent详情{tool:tool-1}(key="DocQA_Agent")->res',
            latest_name_map={},
        )

        assert dolphin == '@获取agent详情(key="DocQA_Agent")->res'

    def test_refresh_dolphin_from_enhance_updates_config(self):
        """测试刷新 AgentConfigVo.dolphin"""
        config = AgentConfigVo(
            is_dolphin_mode=True,
            dolphin='@获取agent详情(key="DocQA_Agent")->res\n@问答助手(query=input)->answer',
            dolphin_enhance='@获取agent详情{tool:tool-1}(key="DocQA_Agent")->res\n'
            '@问答助手{agent:agent-1}(query=input)->answer',
            skills={
                "tools": [{"tool_id": "tool-1", "tool_box_id": "toolbox-1"}],
                "agents": [{"agent_key": "agent-1"}],
            },
        )
        config.skills.tools[0].__dict__["tool_info"] = {"name": "获取Agent详情"}
        config.skills.agents[0].inner_dto.agent_info = {"name": "知识问答助手"}

        refresh_dolphin_from_enhance(config)

        assert config.dolphin == (
            '@获取Agent详情(key="DocQA_Agent")->res\n'
            "@知识问答助手(query=input)->answer"
        )
