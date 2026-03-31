"""单元测试 - domain/vo/agentvo/agent_input 模块"""

import json


class TestAgentInputVo:
    """测试 AgentInputVo 类"""

    def test_init_with_minimal_fields(self):
        """测试使用最小字段初始化"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test query")

        assert vo.query == "test query"

    def test_init_with_history(self):
        """测试带历史记录初始化"""
        from app.domain.vo.agentvo import AgentInputVo

        history = [{"role": "user", "content": "hello"}]
        vo = AgentInputVo(query="test", history=history)

        assert vo.history == history

    def test_init_with_tool(self):
        """测试带工具初始化"""
        from app.domain.vo.agentvo import AgentInputVo

        tool = {"name": "test_tool", "parameters": {}}
        vo = AgentInputVo(query="test", tool=tool)

        assert vo.tool == tool

    def test_get_value_defined_field(self):
        """测试获取已定义字段的值"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", user_id="user123")
        value = vo.get_value("user_id")

        assert value == "user123"

    def test_get_value_extra_field(self):
        """测试获取额外字段的值"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", extra_field="extra_value")
        value = vo.get_value("extra_field")

        assert value == "extra_value"

    def test_get_value_with_default(self):
        """测试获取字段值时使用默认值"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        value = vo.get_value("nonexistent", default="default_value")

        assert value == "default_value"

    def test_set_value(self):
        """测试设置字段值"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        vo.set_value("new_field", "new_value")

        assert vo.get_value("new_field") == "new_value"

    def test_extra_fields_allowed(self):
        """测试允许额外字段"""
        from app.domain.vo.agentvo import AgentInputVo

        # Should not raise validation error
        vo = AgentInputVo(query="test", custom_field="custom_value")

        assert vo.custom_field == "custom_value"

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test query", user_id="user123")
        data = vo.model_dump()

        assert data["query"] == "test query"
        assert data["user_id"] == "user123"

    def test_model_dump_json(self):
        """测试模型JSON序列化"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test query")
        json_str = vo.model_dump_json()

        data = json.loads(json_str)
        assert data["query"] == "test query"

    def test_model_dump_excludes_empty_tool(self):
        """测试model_dump排除空tool字段"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", tool={})
        data = vo.model_dump()

        assert "tool" not in data

    def test_init_with_header(self):
        """测试带header初始化"""
        from app.domain.vo.agentvo import AgentInputVo

        header = {"Authorization": "Bearer token"}
        vo = AgentInputVo(query="test", header=header)

        assert vo.header == header

    def test_init_with_self_config(self):
        """测试带self_config初始化"""
        from app.domain.vo.agentvo import AgentInputVo

        config = {"key": "value"}
        vo = AgentInputVo(query="test", self_config=config)

        assert vo.self_config == config

    def test_empty_history(self):
        """测试空历史记录"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", history=[])

        assert vo.history == []

    def test_multiple_history_entries(self):
        """测试多条历史记录"""
        from app.domain.vo.agentvo import AgentInputVo

        history = [
            {"role": "user", "content": "hello"},
            {"role": "assistant", "content": "hi there"},
            {"role": "user", "content": "how are you?"},
        ]
        vo = AgentInputVo(query="test", history=history)

        assert len(vo.history) == 3
        assert vo.history[2]["content"] == "how are you?"


class TestAgentInputVoEdgeCases:
    """Edge case tests for AgentInputVo"""

    def test_empty_query(self):
        """测试空query"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="")
        assert vo.query == ""

    def test_query_with_special_characters(self):
        """测试带特殊字符的query"""
        from app.domain.vo.agentvo import AgentInputVo

        special_query = "Test with special chars: @#$%^&*()_+-=[]{}|;':\",./<>?"
        vo = AgentInputVo(query=special_query)

        assert vo.query == special_query

    def test_query_with_newlines(self):
        """测试带换行符的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "Line 1\nLine 2\nLine 3"
        vo = AgentInputVo(query=query)

        assert "\n" in vo.query

    def test_query_with_tabs(self):
        """测试带制表符的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "Col1\tCol2\tCol3"
        vo = AgentInputVo(query=query)

        assert "\t" in vo.query

    def test_query_with_unicode(self):
        """测试Unicode query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "测试查询 with 中文 and 混合 text"
        vo = AgentInputVo(query=query)

        assert vo.query == query

    def test_query_with_emoji(self):
        """测试带emoji的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "Hello! How are you? 😀 🎉 🚀"
        vo = AgentInputVo(query=query)

        assert "😀" in vo.query

    def test_very_long_query(self):
        """测试超长query"""
        from app.domain.vo.agentvo import AgentInputVo

        long_query = "a" * 10000
        vo = AgentInputVo(query=long_query)

        assert len(vo.query) == 10000

    def test_history_with_complex_content(self):
        """测试带复杂内容的历史记录"""
        from app.domain.vo.agentvo import AgentInputVo

        # History items must have role and content as strings
        # But can have extra fields since AgentInputVo allows extra fields
        # However, the history type is List[Dict[str, str]], so values must be strings
        history = [
            {
                "role": "user",
                "content": "Question",
                "extra": "extra_value",  # Extra field with string value
            }
        ]
        vo = AgentInputVo(query="test", history=history)

        assert vo.history[0]["extra"] == "extra_value"

    def test_tool_with_complex_structure(self):
        """测试带复杂结构的tool"""
        from app.domain.vo.agentvo import AgentInputVo

        tool = {
            "name": "test_tool",
            "description": "A test tool",
            "parameters": {"param1": "value1", "param2": {"nested": "value2"}},
            "required": ["param1"],
        }
        vo = AgentInputVo(query="test", tool=tool)

        assert vo.tool["parameters"]["param2"]["nested"] == "value2"

    def test_header_with_multiple_fields(self):
        """测试带多个字段的header"""
        from app.domain.vo.agentvo import AgentInputVo

        header = {
            "Authorization": "Bearer token",
            "Content-Type": "application/json",
            "Accept": "application/json",
            "User-Agent": "TestAgent/1.0",
        }
        vo = AgentInputVo(query="test", header=header)

        assert len(vo.header) == 4
        assert vo.header["Accept"] == "application/json"

    def test_self_config_nested(self):
        """测试嵌套的self_config"""
        from app.domain.vo.agentvo import AgentInputVo

        config = {"level1": {"level2": {"level3": "deep_value"}}}
        vo = AgentInputVo(query="test", self_config=config)

        assert vo.self_config["level1"]["level2"]["level3"] == "deep_value"

    def test_get_value_query(self):
        """测试获取query字段"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test query")
        assert vo.get_value("query") == "test query"

    def test_get_value_history(self):
        """测试获取history字段"""
        from app.domain.vo.agentvo import AgentInputVo

        history = [{"role": "user", "content": "test"}]
        vo = AgentInputVo(query="test", history=history)

        assert vo.get_value("history") == history

    def test_get_value_header(self):
        """测试获取header字段"""
        from app.domain.vo.agentvo import AgentInputVo

        header = {"key": "value"}
        vo = AgentInputVo(query="test", header=header)

        assert vo.get_value("header") == header

    def test_get_value_self_config(self):
        """测试获取self_config字段"""
        from app.domain.vo.agentvo import AgentInputVo

        config = {"key": "value"}
        vo = AgentInputVo(query="test", self_config=config)

        assert vo.get_value("self_config") == config

    def test_set_value_updates_query(self):
        """测试设置query值"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="original")
        vo.set_value("query", "updated")

        assert vo.query == "updated"

    def test_set_value_adds_extra_field(self):
        """测试设置额外字段"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        vo.set_value("custom_field", "custom_value")

        assert vo.get_value("custom_field") == "custom_value"

    def test_set_value_with_none(self):
        """测试设置为None"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        vo.set_value("field1", "value1")
        vo.set_value("field1", None)

        assert vo.get_value("field1") is None

    def test_set_value_with_empty_string(self):
        """测试设置为空字符串"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        vo.set_value("field1", "")

        assert vo.get_value("field1") == ""

    def test_set_value_with_zero(self):
        """测试设置为0"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        vo.set_value("count", 0)

        assert vo.get_value("count") == 0

    def test_set_value_with_false(self):
        """测试设置为False"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        vo.set_value("flag", False)

        assert vo.get_value("flag") is False

    def test_set_value_with_list(self):
        """测试设置为列表"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        items = [1, 2, 3]
        vo.set_value("items", items)

        assert vo.get_value("items") == items

    def test_set_value_with_dict(self):
        """测试设置为字典"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        data = {"key": "value"}
        vo.set_value("data", data)

        assert vo.get_value("data") == data

    def test_model_dump_with_all_fields(self):
        """测试序列化所有字段"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(
            query="test",
            history=[{"role": "user", "content": "hi"}],
            header={"Auth": "Bearer token"},
            self_config={"key": "value"},
            custom_field="custom",
        )

        data = vo.model_dump()

        assert data["query"] == "test"
        assert data["history"][0]["content"] == "hi"
        assert data["header"]["Auth"] == "Bearer token"
        assert data["self_config"]["key"] == "value"
        assert data["custom_field"] == "custom"

    def test_model_dump_with_exclude(self):
        """测试序列化时排除字段"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", history=[{"role": "user"}])

        data = vo.model_dump(exclude={"history"})

        assert "history" not in data
        assert "query" in data

    def test_model_dump_exclude_none(self):
        """测试序列化时排除None值"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", tool={})
        data = vo.model_dump(exclude_none=True)

        # Empty tool dict should be excluded by custom model_dump
        assert "tool" not in data

    def test_model_dump_mode_json(self):
        """测试JSON模式序列化"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", count=123)
        data = vo.model_dump(mode="json")

        assert data["count"] == 123

    def test_model_dump_json_with_extra_fields(self):
        """测试JSON序列化包含额外字段"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", custom="value")
        json_str = vo.model_dump_json()

        assert "custom" in json_str

    def test_copy(self):
        """测试复制VO"""
        from app.domain.vo.agentvo import AgentInputVo

        vo1 = AgentInputVo(query="test", custom_field="custom")
        vo2 = vo1.copy()

        assert vo2.query == "test"
        assert vo2.custom_field == "custom"

    def test_copy_with_update(self):
        """测试复制并更新"""
        from app.domain.vo.agentvo import AgentInputVo

        vo1 = AgentInputVo(query="original")
        vo2 = vo1.copy(update={"query": "updated"})

        assert vo2.query == "updated"
        assert vo1.query == "original"

    def test_equality(self):
        """测试相等性"""
        from app.domain.vo.agentvo import AgentInputVo

        vo1 = AgentInputVo(query="test")
        vo2 = AgentInputVo(query="test")
        vo3 = AgentInputVo(query="other")

        assert vo1 == vo2
        assert vo1 != vo3

    def test_from_dict(self):
        """测试从字典创建"""
        from app.domain.vo.agentvo import AgentInputVo

        data = {
            "query": "test query",
            "history": [{"role": "user", "content": "hello"}],
            "custom": "value",
        }

        vo = AgentInputVo(**data)

        assert vo.query == "test query"
        assert vo.history[0]["content"] == "hello"
        assert vo.custom == "value"

    def test_model_validate(self):
        """测试model_validate方法"""
        from app.domain.vo.agentvo import AgentInputVo

        data = {"query": "test"}
        vo = AgentInputVo.model_validate(data)

        assert vo.query == "test"

    def test_model_validate_json(self):
        """测试model_validate_json方法"""
        from app.domain.vo.agentvo import AgentInputVo

        json_str = '{"query": "test", "custom": "value"}'
        vo = AgentInputVo.model_validate_json(json_str)

        assert vo.query == "test"
        assert vo.custom == "value"

    def test_with_url_query(self):
        """测试URL格式的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "Check https://example.com/path?query=value for info"
        vo = AgentInputVo(query=query)

        assert "https://example.com" in vo.query

    def test_with_json_in_query(self):
        """测试JSON格式的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = '{"key": "value", "number": 123}'
        vo = AgentInputVo(query=query)

        assert vo.query == query

    def test_with_xml_in_query(self):
        """测试XML格式的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "<root><item>value</item></root>"
        vo = AgentInputVo(query=query)

        assert vo.query == query

    def test_with_code_in_query(self):
        """测试代码格式的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "def test():\n    return 'hello'"
        vo = AgentInputVo(query=query)

        assert "def test():" in vo.query

    def test_with_sql_in_query(self):
        """测试SQL格式的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "SELECT * FROM users WHERE id = 123"
        vo = AgentInputVo(query=query)

        assert "SELECT" in vo.query

    def test_history_various_roles(self):
        """测试各种角色的历史记录"""
        from app.domain.vo.agentvo import AgentInputVo

        history = [
            {"role": "system", "content": "System message"},
            {"role": "user", "content": "User message"},
            {"role": "assistant", "content": "Assistant message"},
            {"role": "function", "content": "Function result"},
        ]
        vo = AgentInputVo(query="test", history=history)

        assert len(vo.history) == 4
        assert vo.history[0]["role"] == "system"

    def test_multiple_extra_fields(self):
        """测试多个额外字段"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(
            query="test", field1="value1", field2="value2", field3="value3"
        )

        assert vo.field1 == "value1"
        assert vo.field2 == "value2"
        assert vo.field3 == "value3"

    def test_extra_field_types(self):
        """测试各种类型的额外字段"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(
            query="test",
            string_field="string",
            int_field=123,
            float_field=3.14,
            bool_field=True,
            list_field=[1, 2, 3],
            dict_field={"key": "value"},
            none_field=None,
        )

        assert vo.string_field == "string"
        assert vo.int_field == 123
        assert vo.float_field == 3.14
        assert vo.bool_field is True
        assert vo.list_field == [1, 2, 3]
        assert vo.dict_field == {"key": "value"}
        assert vo.none_field is None

    def test_repr(self):
        """测试字符串表示"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test query")
        repr_str = repr(vo)

        assert "AgentInputVo" in repr_str or "query" in repr_str

    def test_with_base64_in_query(self):
        """测试Base64格式的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "SGVsbG8gV29ybGQ="
        vo = AgentInputVo(query=query)

        assert vo.query == query

    def test_with_html_in_query(self):
        """测试HTML格式的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "<div class='test'>Hello</div>"
        vo = AgentInputVo(query=query)

        assert "<div" in vo.query

    def test_with_markdown_in_query(self):
        """测试Markdown格式的query"""
        from app.domain.vo.agentvo import AgentInputVo

        query = "# Heading\n\n**Bold** and *italic*"
        vo = AgentInputVo(query=query)

        assert "# Heading" in vo.query

    def test_empty_tool_not_in_dump(self):
        """测试空tool不在序列化结果中"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", tool={})
        data = vo.model_dump()

        assert "tool" not in data

    def test_non_empty_tool_in_dump(self):
        """测试非空tool在序列化结果中"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test", tool={"name": "test_tool"})
        data = vo.model_dump()

        assert "tool" in data
        assert data["tool"]["name"] == "test_tool"

    def test_get_value_nonexistent_with_default(self):
        """测试获取不存在的字段并返回默认值"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")

        assert vo.get_value("nonexistent", default=None) is None
        assert vo.get_value("nonexistent", default=0) == 0
        assert vo.get_value("nonexistent", default=False) is False
        assert vo.get_value("nonexistent", default="") == ""

    def test_config_attributes(self):
        """测试Config类属性"""
        from app.domain.vo.agentvo import AgentInputVo

        # Check that extra fields are allowed
        assert AgentInputVo.model_config.get("extra") == "allow"

    def test_field_description(self):
        """测试字段描述"""
        from app.domain.vo.agentvo import AgentInputVo

        # Query is required
        assert "query" in AgentInputVo.model_fields

    def test_immutability_check(self):
        """测试字段可变性（Pydantic v2默认可变）"""
        from app.domain.vo.agentvo import AgentInputVo

        vo = AgentInputVo(query="test")
        vo.query = "updated"

        assert vo.query == "updated"
