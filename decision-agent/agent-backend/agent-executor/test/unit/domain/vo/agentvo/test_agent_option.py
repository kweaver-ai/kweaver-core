"""单元测试 - domain/vo/agentvo/agent_option 模块"""


class TestAgentRunOptionsVo:
    """测试 AgentRunOptionsVo 类"""

    def test_init_with_no_fields(self):
        """测试不使用任何字段初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo()

        assert options.output_vars is None
        assert options.incremental_output is None
        assert options.data_source is None
        assert options.llm_config is None
        assert options.tmp_files is None
        assert options.agent_id is None
        assert options.conversation_id is None
        assert options.agent_run_id is None
        assert options.is_need_progress is None
        assert options.enable_dependency_cache is None
        assert options.resume_info is None

    def test_init_with_output_vars(self):
        """测试使用output_vars初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(output_vars=["var1", "var2"])

        assert options.output_vars == ["var1", "var2"]

    def test_init_with_incremental_output(self):
        """测试使用incremental_output初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(incremental_output=True)

        assert options.incremental_output is True

    def test_init_with_data_source(self):
        """测试使用data_source初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        data_source = {"key": "value"}
        options = AgentRunOptionsVo(data_source=data_source)

        assert options.data_source == {"key": "value"}

    def test_init_with_llm_config(self):
        """测试使用llm_config初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        llm_config = {"model": "gpt-4", "temperature": 0.7}
        options = AgentRunOptionsVo(llm_config=llm_config)

        assert options.llm_config == {"model": "gpt-4", "temperature": 0.7}

    def test_init_with_tmp_files(self):
        """测试使用tmp_files初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        tmp_files = ["file1.txt", "file2.txt"]
        options = AgentRunOptionsVo(tmp_files=tmp_files)

        assert options.tmp_files == ["file1.txt", "file2.txt"]

    def test_init_with_agent_ids(self):
        """测试使用agent相关ID初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(
            agent_id="agent123", conversation_id="conv456", agent_run_id="run789"
        )

        assert options.agent_id == "agent123"
        assert options.conversation_id == "conv456"
        assert options.agent_run_id == "run789"

    def test_init_with_progress_options(self):
        """测试使用进度相关选项初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(
            is_need_progress=True, enable_dependency_cache=False
        )

        assert options.is_need_progress is True
        assert options.enable_dependency_cache is False

    def test_init_with_resume_info(self):
        """测试使用resume_info初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        resume_info = {"frame_id": "frame123"}
        options = AgentRunOptionsVo(resume_info=resume_info)

        assert options.resume_info == {"frame_id": "frame123"}

    def test_init_with_all_fields(self):
        """测试使用所有字段初始化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(
            output_vars=["var1"],
            incremental_output=True,
            data_source={"key": "value"},
            llm_config={"model": "gpt-4"},
            tmp_files=["file1.txt"],
            agent_id="agent123",
            conversation_id="conv456",
            agent_run_id="run789",
            is_need_progress=True,
            enable_dependency_cache=False,
            resume_info={"frame_id": "frame123"},
        )

        assert options.output_vars == ["var1"]
        assert options.incremental_output is True
        assert options.data_source == {"key": "value"}
        assert options.llm_config == {"model": "gpt-4"}
        assert options.tmp_files == ["file1.txt"]
        assert options.agent_id == "agent123"
        assert options.conversation_id == "conv456"
        assert options.agent_run_id == "run789"
        assert options.is_need_progress is True
        assert options.enable_dependency_cache is False
        assert options.resume_info == {"frame_id": "frame123"}

    def test_is_pydantic_model(self):
        """测试是Pydantic模型"""
        from app.domain.vo.agentvo import AgentRunOptionsVo
        from pydantic import BaseModel

        assert issubclass(AgentRunOptionsVo, BaseModel)

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(output_vars=["var1"], agent_id="agent123")

        data = options.model_dump()

        assert data["output_vars"] == ["var1"]
        assert data["agent_id"] == "agent123"

    def test_model_dump_json(self):
        """测试模型JSON序列化"""
        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(agent_id="agent123")
        json_str = options.model_dump_json()

        assert "agent123" in json_str
