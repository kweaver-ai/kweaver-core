"""单元测试 - domain/vo/agentvo/agent_config_vos/skill_vo 模块"""


class TestSkillVo:
    """测试 SkillVo 类"""

    def test_init_with_no_fields(self):
        """测试不使用任何字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import SkillVo

        vo = SkillVo()

        assert vo.tools == []
        assert vo.agents == []
        assert vo.mcps == []

    def test_init_with_tools(self):
        """测试使用tools初始化"""
        from app.domain.vo.agentvo.agent_config_vos import SkillVo, ToolSkillVo

        tool = ToolSkillVo(tool_id="tool1", tool_box_id="toolbox1")
        vo = SkillVo(tools=[tool])

        assert len(vo.tools) == 1
        assert vo.tools[0].tool_id == "tool1"

    def test_init_with_none_becomes_empty_list(self):
        """测试传入None变为空列表"""
        from app.domain.vo.agentvo.agent_config_vos import SkillVo

        # The validator should convert None to []
        vo = SkillVo(tools=None, agents=None, mcps=None)

        assert vo.tools == []
        assert vo.agents == []
        assert vo.mcps == []

    def test_is_pydantic_model(self):
        """测试是Pydantic模型"""
        from app.domain.vo.agentvo.agent_config_vos import SkillVo
        from pydantic import BaseModel

        assert issubclass(SkillVo, BaseModel)

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo.agent_config_vos import SkillVo

        vo = SkillVo()
        data = vo.model_dump()

        assert data["tools"] == []
        assert data["agents"] == []
        assert data["mcps"] == []
