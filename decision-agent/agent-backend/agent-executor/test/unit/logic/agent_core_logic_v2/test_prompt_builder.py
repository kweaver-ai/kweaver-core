"""单元测试 - logic/agent_core_logic_v2/prompt_builder 模块"""

import pytest
from unittest.mock import Mock, patch


class TestPromptBuilder:
    """测试 PromptBuilder 类"""

    def setup_method(self):
        """每个测试方法前的设置"""
        # Mock config
        self.mock_config = Mock()
        self.mock_config.agent_id = "test_agent_123"
        self.mock_config.agent_run_id = "run_789"
        self.mock_config.is_dolphin_mode = False
        self.mock_config.system_prompt = "You are a helpful assistant."
        self.mock_config.dolphin = "test dolphin prompt"
        self.mock_config.pre_dolphin = []
        self.mock_config.post_dolphin = []
        self.mock_config.memory = None
        self.mock_config.related_question = None
        self.mock_config.plan_mode = None

        # Mock temp_files
        self.temp_files = {}

        # Import after setup
        from app.logic.agent_core_logic_v2.prompt_builder import (
            PromptBuilder,
            has_temp_files,
        )

        self.PromptBuilder = PromptBuilder
        self.has_temp_files = has_temp_files

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_dolphin_mode(self, mock_plan_mode_logic):
        """测试构建Dolphin模式提示词"""
        self.mock_config.is_dolphin_mode = True
        self.mock_config.pre_dolphin = [
            {"enabled": True, "key": "pre1", "value": "Pre prompt 1"}
        ]
        self.mock_config.post_dolphin = [
            {"enabled": True, "key": "post1", "value": "Post prompt 1"}
        ]

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        assert "Pre prompt 1" in result
        assert "test dolphin prompt" in result
        assert "Post prompt 1" in result
        assert "self_config" in result
        assert "header" in result

    @pytest.mark.asyncio
    async def test_build_dolphin_mode_disabled_pre_dolphin(self):
        """测试Dolphin模式下禁用的pre_dolphin"""
        self.mock_config.is_dolphin_mode = True
        self.mock_config.pre_dolphin = [
            {"enabled": False, "key": "pre1", "value": "Pre prompt 1"}
        ]

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        assert "Pre prompt 1" not in result
        assert "test dolphin prompt" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_standard_mode(self, mock_plan_mode_logic):
        """测试构建标准模式提示词"""
        mock_plan_mode_logic.get_system_prompt_with_plan.return_value = (
            "Plan mode prompt"
        )

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        assert "/explore/" in result
        assert "self_config" in result
        assert "header" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_with_memory_enabled(self, mock_plan_mode_logic):
        """测试启用记忆时构建提示词"""
        self.mock_config.memory = {"is_enabled": True}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        assert "@search_memory" in result
        assert "relevant_memories" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_with_memory_disabled(self, mock_plan_mode_logic):
        """测试禁用记忆时构建提示词"""
        self.mock_config.memory = {"is_enabled": False}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        assert "@search_memory" not in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_with_plan_mode(self, mock_plan_mode_logic):
        """测试计划模式"""
        self.mock_config.plan_mode = {"is_enabled": True}
        mock_plan_mode_logic.get_system_prompt_with_plan.return_value = (
            "Plan mode system prompt"
        )

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        # Verify append_task_plan_agent was called
        self.mock_config.append_task_plan_agent.assert_called_once()

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_without_plan_mode(self, mock_plan_mode_logic):
        """测试非计划模式"""
        self.mock_config.plan_mode = None
        # Make is_plan_mode return False
        self.mock_config.is_plan_mode = Mock(return_value=False)

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        # Verify append_task_plan_agent was not called
        self.mock_config.append_task_plan_agent.assert_not_called()

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_with_related_questions(self, mock_plan_mode_logic):
        """测试启用相关问题"""
        self.mock_config.related_question = {"is_enabled": True}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        assert "related_questions" in result
        assert "3个问题和答案对" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_with_temp_files(self, mock_plan_mode_logic):
        """测试带临时文件"""
        self.mock_config.is_dolphin_mode = True
        self.temp_files = {"file1.txt": ["file content"]}

        self.mock_config.pre_dolphin = [
            {"enabled": True, "key": "temp_file_process", "value": "Process temp files"}
        ]

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        assert "Process temp files" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_without_temp_files_for_temp_process(
        self, mock_plan_mode_logic
    ):
        """测试无临时文件时跳过temp_file_process"""
        self.mock_config.is_dolphin_mode = True
        self.temp_files = {}

        self.mock_config.pre_dolphin = [
            {"enabled": True, "key": "temp_file_process", "value": "Process temp files"}
        ]

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        assert "Process temp files" not in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_get_doc_retrieval_prompt(self, mock_plan_mode_logic):
        """测试获取文档召回提示词"""
        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.get_doc_retrieval_prompt()

        assert result == ""

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_get_graph_retrieval_prompt(self, mock_plan_mode_logic):
        """测试获取图谱召回提示词"""
        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.get_graph_retrieval_prompt()

        assert result == ""

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_search_memory_prompt_enabled(self, mock_plan_mode_logic):
        """测试搜索记忆提示词 - 启用"""
        self.mock_config.memory = {"is_enabled": True}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.search_memory_prompt()

        assert "@search_memory" in result
        assert "relevant_memories" in result
        assert "limit=" in result
        assert "threshold=" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_search_memory_prompt_disabled(self, mock_plan_mode_logic):
        """测试搜索记忆提示词 - 禁用"""
        self.mock_config.memory = {"is_enabled": False}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.search_memory_prompt()

        assert result == ""

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_search_memory_prompt_none(self, mock_plan_mode_logic):
        """测试搜索记忆提示词 - memory为None"""
        self.mock_config.memory = None

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.search_memory_prompt()

        assert result == ""

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_temp_file_prompt_with_files(self, mock_plan_mode_logic):
        """测试临时文件提示词 - 有文件"""
        self.temp_files = {"file1.txt": ["content1"], "file2.pdf": ["content2"]}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.temp_file_prompt()

        assert "@process_file_intelligent" in result
        assert "file1.txt" in result
        assert "file2.pdf" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_temp_file_prompt_empty(self, mock_plan_mode_logic):
        """测试临时文件提示词 - 空文件"""
        self.temp_files = {}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.temp_file_prompt()

        assert result == ""

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_temp_file_prompt_empty_list(self, mock_plan_mode_logic):
        """测试临时文件提示词 - 空列表"""
        self.temp_files = {"file1.txt": []}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.temp_file_prompt()

        assert result == ""

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_get_context_prompt(self, mock_plan_mode_logic):
        """测试获取上下文提示词"""
        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        # Set the attributes that build() would set
        builder.doc_retrieval_prompt = ""
        builder.graph_retrieval_prompt = ""
        builder.temp_file_prompt = ""
        result = builder.get_context_prompt()

        assert "reference" in result
        assert "query" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_get_related_questions_prompt_enabled(self, mock_plan_mode_logic):
        """测试获取相关问题提示词 - 启用"""
        self.mock_config.related_question = {"is_enabled": True}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.get_related_questions_prompt()

        assert "related_questions" in result
        assert "3个问题和答案对" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_get_related_questions_prompt_disabled(self, mock_plan_mode_logic):
        """测试获取相关问题提示词 - 禁用"""
        self.mock_config.related_question = {"is_enabled": False}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.get_related_questions_prompt()

        assert result == ""

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_get_related_questions_prompt_none(self, mock_plan_mode_logic):
        """测试获取相关问题提示词 - None"""
        self.mock_config.related_question = None

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = builder.get_related_questions_prompt()

        assert result == ""

    def test_has_temp_files_true(self):
        """测试has_temp_files - 有文件"""
        temp_files = {"file1.txt": ["content1"], "file2.pdf": ["content2"]}

        result = self.has_temp_files(temp_files)
        assert result is True

    def test_has_temp_files_false_empty(self):
        """测试has_temp_files - 空字典"""
        temp_files = {}

        result = self.has_temp_files(temp_files)
        assert result is False

    def test_has_temp_files_false_empty_lists(self):
        """测试has_temp_files - 空列表"""
        temp_files = {"file1.txt": [], "file2.pdf": []}

        result = self.has_temp_files(temp_files)
        assert result is False

    def test_has_temp_files_mixed(self):
        """测试has_temp_files - 混合"""
        temp_files = {"file1.txt": [], "file2.pdf": ["content2"]}

        result = self.has_temp_files(temp_files)
        assert result is True

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_with_none_agent_run_id(self, mock_plan_mode_logic):
        """测试agent_run_id为None时构建提示词"""
        self.mock_config.agent_run_id = None

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        # Should not raise exception
        assert isinstance(result, str)

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_with_none_agent_id(self, mock_plan_mode_logic):
        """测试agent_id为None时构建提示词"""
        self.mock_config.agent_id = None

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        # Should not raise exception
        assert isinstance(result, str)

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_with_span_recording(self, mock_plan_mode_logic):
        """测试span录制时的构建"""
        mock_span = Mock()
        mock_span.is_recording = Mock(return_value=True)

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build(span=mock_span)

        # Verify span attributes were set
        assert mock_span.set_attribute.called

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_get_context_prompt_with_temp_files(self, mock_plan_mode_logic):
        """测试带临时文件的上下文提示词"""
        self.temp_files = {"file1.txt": ["content1"]}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        # Set the attributes that build() would set
        builder.doc_retrieval_prompt = ""
        builder.graph_retrieval_prompt = ""
        builder.temp_file_prompt = "temp file content"
        result = builder.get_context_prompt()

        assert "file1.txt" in result
        assert "用户上传的文件内容" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_get_context_prompt_with_doc_retrieval(self, mock_plan_mode_logic):
        """测试带文档召回的上下文提示词"""
        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        # Set the attributes that build() would set
        builder.doc_retrieval_prompt = "some doc content"
        builder.graph_retrieval_prompt = ""
        builder.temp_file_prompt = ""

        result = builder.get_context_prompt()

        assert "doc_retrieval_res" in result
        assert "文档召回的内容" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_get_context_prompt_with_graph_retrieval(self, mock_plan_mode_logic):
        """测试带图谱召回的上下文提示词"""
        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        # Set the attributes that build() would set
        builder.doc_retrieval_prompt = ""
        builder.graph_retrieval_prompt = "some graph content"
        builder.temp_file_prompt = ""

        result = builder.get_context_prompt()

        assert "graph_retrieval_res" in result
        assert "业务知识网络召回的内容" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_standard_mode_full_prompt(self, mock_plan_mode_logic):
        """测试标准模式完整提示词构建"""
        self.mock_config.memory = {"is_enabled": True}
        self.mock_config.related_question = {"is_enabled": True}
        self.temp_files = {"file1.txt": ["content"]}

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        # Verify all components are included
        assert "@search_memory" in result
        assert "file1.txt" in result
        assert "related_questions" in result
        assert "/explore/" in result

    @pytest.mark.asyncio
    @patch("app.logic.agent_core_logic_v2.prompt_builder.plan_mode_logic")
    async def test_build_dolphin_mode_empty_pre_post(self, mock_plan_mode_logic):
        """测试Dolphin模式下空的pre/post dolphin"""
        self.mock_config.is_dolphin_mode = True
        self.mock_config.pre_dolphin = []
        self.mock_config.post_dolphin = []

        builder = self.PromptBuilder(self.mock_config, self.temp_files)
        result = await builder.build()

        # Should only have dolphin prompt
        assert "test dolphin prompt" in result
