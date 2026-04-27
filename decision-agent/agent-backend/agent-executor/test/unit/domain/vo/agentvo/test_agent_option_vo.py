"""单元测试 - domain/vo/agentvo/agent_option 模块"""

from unittest import TestCase

from app.domain.vo.agentvo.agent_option import AgentRunOptionsVo


class TestAgentRunOptionsVo(TestCase):
    """测试 AgentRunOptionsVo 类"""

    def test_init_minimal(self):
        """测试最小初始化"""
        options = AgentRunOptionsVo()
        self.assertIsNone(options.output_vars)
        self.assertIsNone(options.incremental_output)
        self.assertIsNone(options.data_source)
        self.assertIsNone(options.llm_config)
        self.assertIsNone(options.tmp_files)
        self.assertIsNone(options.agent_id)
        self.assertIsNone(options.conversation_id)
        self.assertIsNone(options.agent_run_id)
        self.assertIsNone(options.is_need_progress)
        self.assertIsNone(options.enable_dependency_cache)

    def test_init_with_output_vars(self):
        """测试带输出变量初始化"""
        output_vars = ["var1", "var2", "var3"]
        options = AgentRunOptionsVo(output_vars=output_vars)
        self.assertEqual(options.output_vars, output_vars)

    def test_init_with_incremental_output(self):
        """测试带增量输出初始化"""
        options = AgentRunOptionsVo(incremental_output=True)
        self.assertTrue(options.incremental_output)

    def test_init_with_data_source(self):
        """测试带数据源初始化"""
        data_source = {"type": "database", "config": {}}
        options = AgentRunOptionsVo(data_source=data_source)
        self.assertEqual(options.data_source, data_source)

    def test_init_with_llm_config(self):
        """测试带LLM配置初始化"""
        llm_config = {"model": "gpt-4", "temperature": 0.7}
        options = AgentRunOptionsVo(llm_config=llm_config)
        self.assertEqual(options.llm_config, llm_config)

    def test_init_with_tmp_files(self):
        """测试带临时文件初始化"""
        tmp_files = ["/path/to/file1", "/path/to/file2"]
        options = AgentRunOptionsVo(tmp_files=tmp_files)
        self.assertEqual(options.tmp_files, tmp_files)

    def test_init_with_agent_id(self):
        """测试带agent_id初始化"""
        options = AgentRunOptionsVo(agent_id="agent_123")
        self.assertEqual(options.agent_id, "agent_123")

    def test_init_with_conversation_id(self):
        """测试带conversation_id初始化"""
        options = AgentRunOptionsVo(conversation_id="conv_456")
        self.assertEqual(options.conversation_id, "conv_456")

    def test_init_with_agent_run_id(self):
        """测试带agent_run_id初始化"""
        options = AgentRunOptionsVo(agent_run_id="run_789")
        self.assertEqual(options.agent_run_id, "run_789")

    def test_init_with_is_need_progress(self):
        """测试带is_need_progress初始化"""
        options = AgentRunOptionsVo(is_need_progress=True)
        self.assertTrue(options.is_need_progress)

    def test_init_with_enable_dependency_cache(self):
        """测试带enable_dependency_cache初始化"""
        options = AgentRunOptionsVo(enable_dependency_cache=True)
        self.assertTrue(options.enable_dependency_cache)

    def test_init_with_resume_info(self):
        """测试带resume_info初始化"""
        resume_info = {"conversation_id": "conv_123", "agent_run_id": "run_456"}
        options = AgentRunOptionsVo(resume_info=resume_info)
        self.assertEqual(options.resume_info, resume_info)

    def test_init_with_all_fields(self):
        """测试带所有字段初始化"""
        options = AgentRunOptionsVo(
            output_vars=["var1", "var2"],
            incremental_output=True,
            data_source={"type": "database"},
            llm_config={"model": "gpt-4"},
            tmp_files=["file1", "file2"],
            agent_id="agent_123",
            conversation_id="conv_456",
            agent_run_id="run_789",
            is_need_progress=True,
            enable_dependency_cache=True,
        )

        self.assertEqual(options.output_vars, ["var1", "var2"])
        self.assertTrue(options.incremental_output)
        self.assertEqual(options.data_source, {"type": "database"})
        self.assertEqual(options.llm_config, {"model": "gpt-4"})
        self.assertEqual(options.tmp_files, ["file1", "file2"])
        self.assertEqual(options.agent_id, "agent_123")
        self.assertEqual(options.conversation_id, "conv_456")
        self.assertEqual(options.agent_run_id, "run_789")
        self.assertTrue(options.is_need_progress)
        self.assertTrue(options.enable_dependency_cache)

    def test_model_dump_minimal(self):
        """测试最小序列化"""
        options = AgentRunOptionsVo()
        data = options.model_dump()
        self.assertIsNone(data["output_vars"])
        self.assertIsNone(data["incremental_output"])

    def test_model_dump_with_fields(self):
        """测试带字段序列化"""
        options = AgentRunOptionsVo(
            output_vars=["var1"],
            incremental_output=True,
            agent_id="agent_123",
        )
        data = options.model_dump()
        self.assertEqual(data["output_vars"], ["var1"])
        self.assertTrue(data["incremental_output"])
        self.assertEqual(data["agent_id"], "agent_123")

    def test_model_dump_all_fields(self):
        """测试所有字段序列化"""
        options = AgentRunOptionsVo(
            output_vars=["var1", "var2"],
            incremental_output=False,
            data_source={"type": "file"},
            llm_config={"model": "gpt-4"},
            tmp_files=["file1"],
            agent_id="agent_123",
            conversation_id="conv_456",
            agent_run_id="run_789",
            is_need_progress=False,
            enable_dependency_cache=False,
        )
        data = options.model_dump()
        # 10 fields + resume_info (None) = 11 total
        self.assertEqual(len(data), 11)
        self.assertEqual(data["output_vars"], ["var1", "var2"])
        self.assertFalse(data["incremental_output"])

    def test_model_dump_exclude_none(self):
        """测试序列化排除None值"""
        options = AgentRunOptionsVo(agent_id="agent_123")
        data = options.model_dump(exclude_none=True)
        self.assertIn("agent_id", data)
        self.assertNotIn("output_vars", data)
        self.assertNotIn("conversation_id", data)

    # ========== 新增测试用例 ==========

    def test_init_with_empty_output_vars(self):
        """测试空列表输出变量"""
        options = AgentRunOptionsVo(output_vars=[])
        self.assertEqual(options.output_vars, [])

    def test_init_with_zero_values(self):
        """测试零值和 False 值"""
        options = AgentRunOptionsVo(
            incremental_output=False,
            is_need_progress=False,
            enable_dependency_cache=False,
        )
        self.assertFalse(options.incremental_output)
        self.assertFalse(options.is_need_progress)
        self.assertFalse(options.enable_dependency_cache)

    def test_init_with_complex_data_source(self):
        """测试复杂数据源配置"""
        data_source = {
            "type": "multi",
            "sources": [
                {"type": "database", "connection": "postgres://..."},
                {"type": "file", "path": "/data/file.json"},
            ],
            "merge_strategy": "concat",
        }
        options = AgentRunOptionsVo(data_source=data_source)
        self.assertEqual(options.data_source, data_source)

    def test_init_with_complex_llm_config(self):
        """测试复杂 LLM 配置"""
        llm_config = {
            "model": "gpt-4",
            "temperature": 0.7,
            "max_tokens": 2000,
            "top_p": 0.9,
            "frequency_penalty": 0.5,
            "presence_penalty": 0.5,
            "stop_sequences": ["\n", "END"],
        }
        options = AgentRunOptionsVo(llm_config=llm_config)
        self.assertEqual(options.llm_config, llm_config)

    def test_init_with_unicode_ids(self):
        """测试 Unicode 字符 ID"""
        options = AgentRunOptionsVo(
            agent_id="代理_123", conversation_id="对话_456", agent_run_id="运行_789"
        )
        self.assertEqual(options.agent_id, "代理_123")
        self.assertEqual(options.conversation_id, "对话_456")
        self.assertEqual(options.agent_run_id, "运行_789")

    def test_init_with_special_chars_in_ids(self):
        """测试包含特殊字符的 ID"""
        options = AgentRunOptionsVo(
            agent_id="agent.with.dots-and_dashes/slashes",
            conversation_id="conv@domain#tag",
            agent_run_id="run:123|version:2",
        )
        self.assertEqual(options.agent_id, "agent.with.dots-and_dashes/slashes")
        self.assertEqual(options.conversation_id, "conv@domain#tag")
        self.assertEqual(options.agent_run_id, "run:123|version:2")

    def test_init_with_nested_resume_info(self):
        """测试嵌套 resume_info"""
        resume_info = {
            "conversation_id": "conv_123",
            "agent_run_id": "run_456",
            "interrupt_data": {
                "type": "user_confirmation",
                "block_index": 5,
                "context": {"key": "value"},
            },
        }
        options = AgentRunOptionsVo(resume_info=resume_info)
        self.assertEqual(
            options.resume_info["interrupt_data"]["context"]["key"], "value"
        )

    def test_model_dump_json_serializable(self):
        """测试序列化结果为 JSON 可序列化"""
        import json

        options = AgentRunOptionsVo(
            output_vars=["var1", "var2"],
            agent_id="agent_123",
            llm_config={"model": "gpt-4"},
        )
        data = options.model_dump()
        # 应该可以序列化为 JSON
        json_str = json.dumps(data)
        self.assertIsInstance(json_str, str)

    def test_model_dump_mode_python(self):
        """测试 Python 模式序列化"""
        options = AgentRunOptionsVo(output_vars=["var1"], incremental_output=True)
        data = options.model_dump(mode="python")
        self.assertEqual(data["output_vars"], ["var1"])
        self.assertTrue(data["incremental_output"])

    def test_model_dump_exclude_fields(self):
        """测试排除特定字段"""
        options = AgentRunOptionsVo(
            output_vars=["var1"], agent_id="agent_123", llm_config={"model": "gpt-4"}
        )
        data = options.model_dump(exclude={"llm_config", "agent_id"})
        self.assertIn("output_vars", data)
        self.assertNotIn("llm_config", data)
        self.assertNotIn("agent_id", data)

    def test_model_dump_include_fields(self):
        """测试仅包含特定字段"""
        options = AgentRunOptionsVo(
            output_vars=["var1"], agent_id="agent_123", llm_config={"model": "gpt-4"}
        )
        data = options.model_dump(include={"output_vars", "agent_id"})
        self.assertIn("output_vars", data)
        self.assertIn("agent_id", data)
        self.assertNotIn("llm_config", data)

    def test_model_dump_exclude_unset(self):
        """测试排除未设置的字段"""
        options = AgentRunOptionsVo(agent_id="agent_123")
        data = options.model_dump(exclude_unset=True)
        # 由于 agent_id 是明确设置的，应该保留
        self.assertIn("agent_id", data)

    def test_model_dump_with_defaults(self):
        """测试包含默认值"""
        options = AgentRunOptionsVo()
        data = options.model_dump()
        # 所有字段应该存在但为 None
        self.assertIsNone(data["output_vars"])
        self.assertIsNone(data["agent_id"])

    def test_field_immutability_on_optional_fields(self):
        """测试可选字段的可变性"""
        options = AgentRunOptionsVo()
        self.assertIsNone(options.agent_id)
        # Pydantic v2 允许修改
        options.agent_id = "new_agent"
        self.assertEqual(options.agent_id, "new_agent")

    def test_model_schema_generation(self):
        """测试模型模式生成"""
        schema = AgentRunOptionsVo.model_json_schema()
        self.assertIn("properties", schema)
        self.assertIn("output_vars", schema["properties"])
        self.assertIn("agent_id", schema["properties"])

    def test_model_validate(self):
        """测试模型验证"""
        data = {"agent_id": "agent_123", "output_vars": ["var1", "var2"]}
        options = AgentRunOptionsVo.model_validate(data)
        self.assertEqual(options.agent_id, "agent_123")
        self.assertEqual(options.output_vars, ["var1", "var2"])

    def test_model_validate_with_extra_fields(self):
        """测试带额外字段的验证"""
        data = {"agent_id": "agent_123", "extra_field": "extra_value"}
        options = AgentRunOptionsVo.model_validate(data)
        self.assertEqual(options.agent_id, "agent_123")

    def test_copy_method(self):
        """测试 copy 方法"""
        options1 = AgentRunOptionsVo(agent_id="agent_123")
        options2 = options1.copy()
        self.assertEqual(options1.agent_id, options2.agent_id)
        options2.agent_id = "agent_456"
        self.assertEqual(options1.agent_id, "agent_123")
        self.assertEqual(options2.agent_id, "agent_456")

    def test_copy_update(self):
        """测试带更新的 copy"""
        options1 = AgentRunOptionsVo(agent_id="agent_123")
        options2 = options1.copy(update={"agent_id": "agent_456"})
        self.assertEqual(options1.agent_id, "agent_123")
        self.assertEqual(options2.agent_id, "agent_456")

    def test_model_construct(self):
        """测试 model_construct 方法"""
        options = AgentRunOptionsVo.model_construct(
            agent_id="agent_123", output_vars=["var1"]
        )
        self.assertEqual(options.agent_id, "agent_123")
        self.assertEqual(options.output_vars, ["var1"])
