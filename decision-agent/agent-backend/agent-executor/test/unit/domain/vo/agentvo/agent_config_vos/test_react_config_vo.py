"""单元测试 - domain/vo/agentvo/agent_config_vos/react_config_vo 模块"""


class TestReactConfigVo:
    """测试 ReactConfigVo 类"""

    def test_default_values(self):
        """测试默认值为关闭状态"""
        from app.domain.vo.agentvo.agent_config_vos import ReactConfigVo

        vo = ReactConfigVo()

        assert vo.disable_history_in_a_conversation is False
        assert vo.disable_llm_cache is False

    def test_init_with_all_fields(self):
        """测试使用所有字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ReactConfigVo

        vo = ReactConfigVo(
            disable_history_in_a_conversation=True,
            disable_llm_cache=True,
        )

        assert vo.disable_history_in_a_conversation is True
        assert vo.disable_llm_cache is True
