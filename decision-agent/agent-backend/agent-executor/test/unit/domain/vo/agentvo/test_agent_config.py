"""单元测试 - domain/vo/agentvo/agent_config 模块"""

from app.domain.vo.agentvo.agent_config import AgentConfigVo


class TestAgentConfigVo:
    """测试 AgentConfigVo 模型"""

    def test_default_initialization(self):
        """测试默认初始化"""
        config = AgentConfigVo()
        assert config.input is None
        assert config.llms is None
        assert config.skills is not None  # Validator converts None to SkillVo()
        assert config.data_source == {}
        assert config.system_prompt is None
        assert config.is_dolphin_mode is False
        assert config.dolphin is None
        assert config.dolphin_enhance is None
        assert config.pre_dolphin == []
        assert config.post_dolphin == []
        assert config.output is not None  # Validator creates default OutputConfigVo
        assert config.memory == {}
        assert config.related_question == {}
        assert config.plan_mode is None
        assert config.metadata is not None  # Validator creates default ConfigMetadataVo
        assert config.non_dolphin_mode_config is None
        assert config.agent_id is None
        assert config.conversation_id is not None  # Validator auto-generates if None
        assert config.agent_run_id is None
        assert config.output_vars is None
        assert config.incremental_output is False
        assert config.agent_version is None

    def test_with_agent_id(self):
        """测试设置agent_id"""
        config = AgentConfigVo(agent_id="agent_123")
        assert config.agent_id == "agent_123"

    def test_with_conversation_id(self):
        """测试设置conversation_id"""
        config = AgentConfigVo(conversation_id="conv_456")
        assert config.conversation_id == "conv_456"

    def test_with_conversation_id_auto_generated(self):
        """测试自动生成conversation_id"""
        config1 = AgentConfigVo()
        config2 = AgentConfigVo()
        # Both should have auto-generated IDs
        assert config1.conversation_id is not None
        assert config2.conversation_id is not None
        # Note: IDs might be the same if created very quickly

    def test_with_agent_run_id(self):
        """测试设置agent_run_id"""
        config = AgentConfigVo(agent_run_id="run_789")
        assert config.agent_run_id == "run_789"

    def test_with_output_vars(self):
        """测试设置output_vars"""
        config = AgentConfigVo(output_vars=["result", "status"])
        assert config.output_vars == ["result", "status"]

    def test_with_incremental_output(self):
        """测试设置incremental_output"""
        config = AgentConfigVo(incremental_output=True)
        assert config.incremental_output is True

    def test_with_system_prompt(self):
        """测试设置system_prompt"""
        prompt = "You are a helpful assistant."
        config = AgentConfigVo(system_prompt=prompt)
        assert config.system_prompt == prompt

    def test_with_dolphin_mode(self):
        """测试dolphin模式"""
        config = AgentConfigVo(
            is_dolphin_mode=True,
            dolphin="some_config",
            dolphin_enhance='@获取agent详情{tool:tool-1}(key="DocQA_Agent")->res',
        )
        assert config.is_dolphin_mode is True
        assert config.dolphin == "some_config"
        assert (
            config.dolphin_enhance
            == '@获取agent详情{tool:tool-1}(key="DocQA_Agent")->res'
        )

    def test_with_plan_mode(self):
        """测试plan_mode"""
        # plan_mode has specific validation, max_steps must be bool
        plan_mode = {"is_enabled": True, "max_steps": True}
        config = AgentConfigVo(plan_mode=plan_mode)
        assert config.plan_mode == plan_mode

    def test_with_all_fields(self):
        """测试所有字段都有值"""
        config = AgentConfigVo(
            input={"query": "test"},
            llms=[{"model": "gpt-4"}],
            data_source={"type": "database"},
            system_prompt="Be helpful",
            agent_id="agent_123",
            conversation_id="conv_456",
            agent_run_id="run_789",
            output_vars=["result"],
            incremental_output=True,
            agent_version="v1.0",
        )
        assert config.input == {"query": "test"}
        assert config.agent_id == "agent_123"
        assert config.incremental_output is True

    def test_is_plan_mode_true(self):
        """测试is_plan_mode为True"""
        config = AgentConfigVo(plan_mode={"is_enabled": True})
        assert config.is_plan_mode() is True

    def test_is_plan_mode_false(self):
        """测试is_plan_mode为False"""
        config = AgentConfigVo(plan_mode={"is_enabled": False})
        assert config.is_plan_mode() is False

    def test_is_plan_mode_none(self):
        """测试plan_mode为None"""
        config = AgentConfigVo(plan_mode=None)
        # The is_plan_mode method returns None when plan_mode is None
        # (due to short-circuit evaluation: None and ... returns None)
        assert config.is_plan_mode() is None or config.is_plan_mode() is False

    def test_is_plan_mode_missing_key(self):
        """测试plan_mode缺少is_enabled键"""
        config = AgentConfigVo(plan_mode={"other_key": True})
        assert config.is_plan_mode() is False

    def test_get_config_last_set_timestamp_with_value(self):
        """测试获取config_last_set_timestamp有值"""
        from app.domain.vo.agentvo.agent_config_vos.config_metadata_vo import (
            ConfigMetadataVo,
        )

        metadata = ConfigMetadataVo(config_last_set_timestamp=1234567890)
        config = AgentConfigVo(metadata=metadata)
        assert config.get_config_last_set_timestamp() == 1234567890

    def test_get_config_last_set_timestamp_default(self):
        """测试获取config_last_set_timestamp默认值"""
        config = AgentConfigVo()
        assert config.get_config_last_set_timestamp() == 0

    def test_pre_dolphin_none_converts_to_empty_list(self):
        """测试pre_dolphin None转换为空列表"""
        config = AgentConfigVo(pre_dolphin=None)
        assert config.pre_dolphin == []

    def test_post_dolphin_none_converts_to_empty_list(self):
        """测试post_dolphin None转换为空列表"""
        config = AgentConfigVo(post_dolphin=None)
        assert config.post_dolphin == []

    def test_pre_dolphin_with_values(self):
        """测试pre_dolphin有值"""
        pre = [{"step": "prepare"}]
        config = AgentConfigVo(pre_dolphin=pre)
        assert config.pre_dolphin == pre

    def test_output_default_format(self):
        """测试output默认格式"""
        config = AgentConfigVo()
        assert config.output is not None

    def test_model_dump(self):
        """测试模型序列化"""
        config = AgentConfigVo(agent_id="test_agent")
        data = config.model_dump()
        assert data["agent_id"] == "test_agent"
        assert isinstance(data, dict)

    def test_model_dump_json(self):
        """测试JSON序列化"""
        config = AgentConfigVo(agent_id="test_agent")
        json_str = config.model_dump_json()
        assert "test_agent" in json_str

    def test_skills_dict_to_skillvo(self):
        """测试字典转换为SkillVo"""
        skills_dict = {"tools": [], "agents": [], "mcps": []}
        config = AgentConfigVo(skills=skills_dict)
        assert config.skills is not None

    def test_output_dict_to_outputconfigvo(self):
        """测试字典转换为OutputConfigVo"""
        output_dict = {"default_format": "json"}
        config = AgentConfigVo(output=output_dict)
        assert config.output is not None

    def test_non_dolphin_mode_config_dict_to_vo(self):
        """测试字典转换为 NonDolphinModeConfigVo"""
        config = AgentConfigVo(
            non_dolphin_mode_config={
                "disable_history_in_a_conversation": True,
                "disable_llm_cache": True,
            }
        )

        assert config.non_dolphin_mode_config is not None
        assert config.disable_history_in_a_conversation() is True
        assert config.disable_llm_cache() is True
