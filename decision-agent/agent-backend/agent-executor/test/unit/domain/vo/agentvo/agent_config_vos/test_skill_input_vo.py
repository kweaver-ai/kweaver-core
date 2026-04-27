"""单元测试 - domain/vo/agentvo/agent_config_vos/skill_input_vo 模块"""


class TestSkillInputVo:
    """测试 SkillInputVo 类"""

    def test_init_with_required_fields(self):
        """测试使用必填字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test_input", input_type="string")

        assert vo.input_name == "test_input"
        assert vo.input_type == "string"
        assert vo.enable is None
        assert vo.input_desc is None
        assert vo.map_type is None
        assert vo.map_value is None
        assert vo.children is None

    def test_init_with_all_fields(self):
        """测试使用所有字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            enable=True,
            input_name="test_input",
            input_type="string",
            input_desc="Test input description",
            map_type="fixedValue",
            map_value="fixed_value_123",
            children=None,
        )

        assert vo.enable is True
        assert vo.input_name == "test_input"
        assert vo.input_type == "string"
        assert vo.input_desc == "Test input description"
        assert vo.map_type == "fixedValue"
        assert vo.map_value == "fixed_value_123"

    def test_init_with_children(self):
        """测试带children初始化"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        child = SkillInputVo(input_name="child_input", input_type="int")

        vo = SkillInputVo(
            input_name="parent_input", input_type="object", children=[child]
        )

        assert vo.children is not None
        assert len(vo.children) == 1
        assert vo.children[0].input_name == "child_input"

    def test_enable_is_optional(self):
        """测试enable是可选的"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test_input", input_type="string")

        assert vo.enable is None

    def test_input_desc_is_optional(self):
        """测试input_desc是可选的"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test_input", input_type="string")

        assert vo.input_desc is None

    def test_map_type_is_optional(self):
        """测试map_type是可选的"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test_input", input_type="string")

        assert vo.map_type is None

    def test_map_value_is_optional(self):
        """测试map_value是可选的"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test_input", input_type="string")

        assert vo.map_value is None

    def test_is_pydantic_model(self):
        """测试是Pydantic模型"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo
        from pydantic import BaseModel

        assert issubclass(SkillInputVo, BaseModel)

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test_input", input_type="string")

        data = vo.model_dump()

        assert data["input_name"] == "test_input"
        assert data["input_type"] == "string"


class TestSkillInputVoExtended:
    """Extended tests for SkillInputVo class"""

    def test_init_with_empty_string_fields(self):
        """Test initialization with empty string fields"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="", input_type="", input_desc="", map_type="", map_value=""
        )

        assert vo.input_name == ""
        assert vo.input_type == ""
        assert vo.input_desc == ""
        assert vo.map_type == ""
        assert vo.map_value == ""

    def test_init_with_various_map_types(self):
        """Test initialization with various map types"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        for map_type in ["fixedValue", "var", "model", "auto"]:
            vo = SkillInputVo(input_name="test", input_type="string", map_type=map_type)
            assert vo.map_type == map_type

    def test_init_with_various_map_values(self):
        """Test initialization with various map value types"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        # String value
        vo1 = SkillInputVo(
            input_name="test", input_type="string", map_value="string_value"
        )
        assert vo1.map_value == "string_value"

        # Integer value
        vo2 = SkillInputVo(input_name="test", input_type="string", map_value=123)
        assert vo2.map_value == 123

        # Dict value
        vo3 = SkillInputVo(
            input_name="test", input_type="string", map_value={"key": "value"}
        )
        assert vo3.map_value == {"key": "value"}

        # List value
        vo4 = SkillInputVo(input_name="test", input_type="string", map_value=[1, 2, 3])
        assert vo4.map_value == [1, 2, 3]

    def test_init_with_nested_children(self):
        """Test initialization with nested children"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        grandchild = SkillInputVo(input_name="grandchild", input_type="int")

        child = SkillInputVo(
            input_name="child", input_type="object", children=[grandchild]
        )

        parent = SkillInputVo(
            input_name="parent", input_type="object", children=[child]
        )

        assert parent.children[0].children[0].input_name == "grandchild"

    def test_init_with_multiple_children(self):
        """Test initialization with multiple children"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        child1 = SkillInputVo(input_name="child1", input_type="string")

        child2 = SkillInputVo(input_name="child2", input_type="int")

        child3 = SkillInputVo(input_name="child3", input_type="bool")

        vo = SkillInputVo(
            input_name="parent", input_type="object", children=[child1, child2, child3]
        )

        assert len(vo.children) == 3
        assert vo.children[0].input_name == "child1"
        assert vo.children[1].input_name == "child2"
        assert vo.children[2].input_name == "child3"

    def test_init_with_enable_true(self):
        """Test initialization with enable=True"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(enable=True, input_name="test", input_type="string")

        assert vo.enable is True

    def test_init_with_enable_false(self):
        """Test initialization with enable=False"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(enable=False, input_name="test", input_type="string")

        assert vo.enable is False

    def test_model_dump_with_all_fields(self):
        """Test model_dump with all fields"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            enable=True,
            input_name="test_input",
            input_type="string",
            input_desc="Test description",
            map_type="fixedValue",
            map_value="test_value",
            children=None,
        )

        data = vo.model_dump()

        assert data["enable"] is True
        assert data["input_name"] == "test_input"
        assert data["input_type"] == "string"
        assert data["input_desc"] == "Test description"
        assert data["map_type"] == "fixedValue"
        assert data["map_value"] == "test_value"
        assert data["children"] is None

    def test_model_dump_with_children(self):
        """Test model_dump with children"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        child = SkillInputVo(input_name="child", input_type="string")

        vo = SkillInputVo(input_name="parent", input_type="object", children=[child])

        data = vo.model_dump()

        assert len(data["children"]) == 1
        assert data["children"][0]["input_name"] == "child"

    def test_model_dump_json(self):
        """Test model_dump_json method"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="string")

        json_str = vo.model_dump_json()

        assert "test" in json_str
        assert "string" in json_str

    def test_copy(self):
        """Test copying the VO"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo1 = SkillInputVo(input_name="test", input_type="string", map_value="value")

        vo2 = vo1.copy()

        assert vo2.input_name == "test"
        assert vo2.input_type == "string"
        assert vo2.map_value == "value"

    def test_equality(self):
        """Test equality comparison"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo1 = SkillInputVo(input_name="test", input_type="string")

        vo2 = SkillInputVo(input_name="test", input_type="string")

        vo3 = SkillInputVo(input_name="other", input_type="int")

        assert vo1 == vo2
        assert vo1 != vo3

    def test_from_dict(self):
        """Test creating instance from dictionary"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        data = {
            "input_name": "test",
            "input_type": "string",
            "map_type": "var",
            "map_value": "$var",
        }

        vo = SkillInputVo(**data)

        assert vo.input_name == "test"
        assert vo.input_type == "string"
        assert vo.map_type == "var"
        assert vo.map_value == "$var"

    def test_with_special_characters_in_desc(self):
        """Test with special characters in description"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="test",
            input_type="string",
            input_desc="Test with special chars: @#$%^&*()",
        )

        assert "@" in vo.input_desc
        assert "#" in vo.input_desc

    def test_with_unicode_in_fields(self):
        """Test with unicode characters in fields"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="测试输入", input_type="字符串", input_desc="这是一个测试描述"
        )

        assert vo.input_name == "测试输入"
        assert vo.input_type == "字符串"
        assert "测试描述" in vo.input_desc


class TestSkillInputVoEdgeCases:
    """Edge case tests for SkillInputVo"""

    def test_with_none_enable(self):
        """Test with None enable value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="string", enable=None)

        assert vo.enable is None

    def test_with_very_long_input_name(self):
        """Test with very long input name"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        long_name = "a" * 1000
        vo = SkillInputVo(input_name=long_name, input_type="string")

        assert vo.input_name == long_name

    def test_with_very_long_input_desc(self):
        """Test with very long input description"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        long_desc = "a" * 5000
        vo = SkillInputVo(input_name="test", input_type="string", input_desc=long_desc)

        assert vo.input_desc == long_desc

    def test_with_deeply_nested_children(self):
        """Test with deeply nested children (5 levels)"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        level5 = SkillInputVo(input_name="level5", input_type="string")
        level4 = SkillInputVo(
            input_name="level4", input_type="object", children=[level5]
        )
        level3 = SkillInputVo(
            input_name="level3", input_type="object", children=[level4]
        )
        level2 = SkillInputVo(
            input_name="level2", input_type="object", children=[level3]
        )
        level1 = SkillInputVo(
            input_name="level1", input_type="object", children=[level2]
        )

        assert (
            level1.children[0].children[0].children[0].children[0].input_name
            == "level5"
        )

    def test_with_many_children(self):
        """Test with many children (100 items)"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        children = [
            SkillInputVo(input_name=f"child_{i}", input_type="string")
            for i in range(100)
        ]
        vo = SkillInputVo(input_name="parent", input_type="object", children=children)

        assert len(vo.children) == 100

    def test_with_complex_map_value_dict(self):
        """Test with complex map value (nested dict)"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        complex_value = {"level1": {"level2": {"level3": "value"}}}

        vo = SkillInputVo(
            input_name="test", input_type="object", map_value=complex_value
        )

        assert vo.map_value == complex_value

    def test_with_complex_map_value_list(self):
        """Test with complex map value (list of dicts)"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        list_value = [
            {"id": 1, "name": "item1"},
            {"id": 2, "name": "item2"},
            {"id": 3, "name": "item3"},
        ]

        vo = SkillInputVo(input_name="test", input_type="array", map_value=list_value)

        assert vo.map_value == list_value

    def test_with_boolean_map_value(self):
        """Test with boolean map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="boolean", map_value=True)

        assert vo.map_value is True

    def test_with_float_map_value(self):
        """Test with float map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="float", map_value=3.14159)

        assert vo.map_value == 3.14159

    def test_with_negative_map_value(self):
        """Test with negative map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="int", map_value=-100)

        assert vo.map_value == -100

    def test_with_zero_map_value(self):
        """Test with zero map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="int", map_value=0)

        assert vo.map_value == 0

    def test_with_null_map_value(self):
        """Test with null map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="string", map_value=None)

        assert vo.map_value is None

    def test_with_special_characters_in_input_name(self):
        """Test with special characters in input name"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        special_names = [
            "test-name",
            "test_name",
            "test.name",
            "test:name",
            "test/name",
            "test\\name",
            "test|name",
            "test@name",
            "test#name",
            "test$name",
            "test%name",
            "test&name",
            "test*name",
            "test?name",
            "test+name",
            "test=name",
            "test[name]",
            "test{name}",
            "test<name>",
        ]

        for name in special_names:
            vo = SkillInputVo(input_name=name, input_type="string")
            assert vo.input_name == name

    def test_with_special_characters_in_input_type(self):
        """Test with special characters in input type"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        special_types = [
            "string",
            "str-ing",
            "str_ing",
            "str.ing",
            "str:ing",
            "str/ing",
            "str\\ing",
            "str|ing",
            "str@ing",
            "str#ing",
            "str$ing",
            "str%ing",
            "str&ing",
            "str*ing",
            "str?ing",
            "str+ing",
            "str=ing",
            "str[ing]",
            "str{ing}",
            "str<ing>",
        ]

        for input_type in special_types:
            vo = SkillInputVo(input_name="test", input_type=input_type)
            assert vo.input_type == input_type

    def test_with_all_map_types(self):
        """Test with all documented map types"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        map_types = ["fixedValue", "var", "model", "auto"]

        for map_type in map_types:
            vo = SkillInputVo(input_name="test", input_type="string", map_type=map_type)
            assert vo.map_type == map_type

    def test_with_custom_map_type(self):
        """Test with custom map type"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="string", map_type="customType")

        assert vo.map_type == "customType"

    def test_with_children_having_different_map_types(self):
        """Test children having different map types"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        child1 = SkillInputVo(
            input_name="child1", input_type="string", map_type="fixedValue"
        )
        child2 = SkillInputVo(input_name="child2", input_type="string", map_type="var")
        child3 = SkillInputVo(input_name="child3", input_type="string", map_type="auto")

        vo = SkillInputVo(
            input_name="parent", input_type="object", children=[child1, child2, child3]
        )

        assert vo.children[0].map_type == "fixedValue"
        assert vo.children[1].map_type == "var"
        assert vo.children[2].map_type == "auto"

    def test_with_empty_children_list(self):
        """Test with empty children list"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="parent", input_type="object", children=[])

        assert vo.children == []

    def test_with_enable_explicit_none(self):
        """Test with explicit None for enable"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="string", enable=None)

        assert vo.enable is None

    def test_model_dump_with_none_values(self):
        """Test model_dump with None values"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="test",
            input_type="string",
            enable=None,
            input_desc=None,
            map_type=None,
            map_value=None,
        )

        data = vo.model_dump()

        assert data["enable"] is None
        assert data["input_desc"] is None
        assert data["map_type"] is None
        assert data["map_value"] is None

    def test_model_dump_exclude_unset(self):
        """Test model_dump with exclude_unset=True"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="string")

        data = vo.model_dump(exclude_unset=True)

        # Only required fields should be present
        assert "input_name" in data
        assert "input_type" in data

    def test_model_dump_exclude_none(self):
        """Test model_dump with exclude_none=True"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="test", input_type="string", enable=None, input_desc=None
        )

        data = vo.model_dump(exclude_none=True)

        # None fields should be excluded
        assert "enable" not in data
        assert "input_desc" not in data

    def test_model_dump_mode_json(self):
        """Test model_dump with mode='json'"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(input_name="test", input_type="string", map_value=123)

        data = vo.model_dump(mode="json")

        assert data["map_value"] == 123

    def test_from_dict_with_extra_fields(self):
        """Test creating from dict with extra fields (should be ignored)"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        data = {
            "input_name": "test",
            "input_type": "string",
            "extra_field": "extra_value",
        }

        vo = SkillInputVo(**data)

        assert vo.input_name == "test"
        assert vo.input_type == "string"

    def test_with_newline_in_desc(self):
        """Test with newline in description"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="test", input_type="string", input_desc="Line 1\nLine 2\nLine 3"
        )

        assert "\n" in vo.input_desc

    def test_with_tab_in_desc(self):
        """Test with tab in description"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="test", input_type="string", input_desc="Col1\tCol2\tCol3"
        )

        assert "\t" in vo.input_desc

    def test_with_emoji_in_desc(self):
        """Test with emoji in description"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="test",
            input_type="string",
            input_desc="Test with emoji: 😀 🎉 🚀",
        )

        assert "😀" in vo.input_desc

    def test_with_url_in_map_value(self):
        """Test with URL in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="test",
            input_type="string",
            map_value="https://example.com/path?query=value",
        )

        assert vo.map_value == "https://example.com/path?query=value"

    def test_with_email_in_map_value(self):
        """Test with email in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        vo = SkillInputVo(
            input_name="test", input_type="string", map_value="user@example.com"
        )

        assert vo.map_value == "user@example.com"

    def test_with_json_string_in_map_value(self):
        """Test with JSON string in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        json_str = '{"key": "value", "number": 123}'
        vo = SkillInputVo(input_name="test", input_type="string", map_value=json_str)

        assert vo.map_value == json_str

    def test_with_xml_in_map_value(self):
        """Test with XML in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        xml = "<root><item>value</item></root>"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=xml)

        assert vo.map_value == xml

    def test_with_sql_in_map_value(self):
        """Test with SQL in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        sql = "SELECT * FROM users WHERE id = 123"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=sql)

        assert vo.map_value == sql

    def test_with_html_in_map_value(self):
        """Test with HTML in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        html = "<div class='test'><span>value</span></div>"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=html)

        assert vo.map_value == html

    def test_with_css_in_map_value(self):
        """Test with CSS in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        css = ".test { color: red; font-size: 14px; }"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=css)

        assert vo.map_value == css

    def test_with_javascript_in_map_value(self):
        """Test with JavaScript in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        js = "function test() { return 'value'; }"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=js)

        assert vo.map_value == js

    def test_with_python_code_in_map_value(self):
        """Test with Python code in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        code = "def test():\n    return 'value'"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=code)

        assert vo.map_value == code

    def test_with_base64_in_map_value(self):
        """Test with base64 in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        base64 = "SGVsbG8gV29ybGQ="
        vo = SkillInputVo(input_name="test", input_type="string", map_value=base64)

        assert vo.map_value == base64

    def test_with_hex_in_map_value(self):
        """Test with hex in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        hex = "48656c6c6f20576f726c64"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=hex)

        assert vo.map_value == hex

    def test_with_uuid_in_map_value(self):
        """Test with UUID in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        uuid = "550e8400-e29b-41d4-a716-446655440000"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=uuid)

        assert vo.map_value == uuid

    def test_with_ip_address_in_map_value(self):
        """Test with IP address in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        ip = "192.168.1.1"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=ip)

        assert vo.map_value == ip

    def test_with_mac_address_in_map_value(self):
        """Test with MAC address in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        mac = "00:1A:2B:3C:4D:5E"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=mac)

        assert vo.map_value == mac

    def test_with_date_in_map_value(self):
        """Test with date in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        date = "2024-01-01"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=date)

        assert vo.map_value == date

    def test_with_time_in_map_value(self):
        """Test with time in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        time = "12:34:56"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=time)

        assert vo.map_value == time

    def test_with_datetime_in_map_value(self):
        """Test with datetime in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        datetime = "2024-01-01T12:34:56Z"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=datetime)

        assert vo.map_value == datetime

    def test_with_timestamp_in_map_value(self):
        """Test with timestamp in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        timestamp = "1704107696"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=timestamp)

        assert vo.map_value == timestamp

    def test_with_latitude_in_map_value(self):
        """Test with latitude in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        lat = 40.7128
        vo = SkillInputVo(input_name="test", input_type="float", map_value=lat)

        assert vo.map_value == lat

    def test_with_longitude_in_map_value(self):
        """Test with longitude in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        lon = -74.0060
        vo = SkillInputVo(input_name="test", input_type="float", map_value=lon)

        assert vo.map_value == lon

    def test_with_coordinates_in_map_value(self):
        """Test with coordinates in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        coords = {"lat": 40.7128, "lon": -74.0060}
        vo = SkillInputVo(input_name="test", input_type="object", map_value=coords)

        assert vo.map_value == coords

    def test_with_color_hex_in_map_value(self):
        """Test with color hex in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        color = "#FF5733"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=color)

        assert vo.map_value == color

    def test_with_color_rgb_in_map_value(self):
        """Test with color RGB in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        rgb = "rgb(255, 87, 51)"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=rgb)

        assert vo.map_value == rgb

    def test_with_phone_number_in_map_value(self):
        """Test with phone number in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        phone = "+1-555-123-4567"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=phone)

        assert vo.map_value == phone

    def test_with_credit_card_in_map_value(self):
        """Test with credit card in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        cc = "4111-1111-1111-1111"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=cc)

        assert vo.map_value == cc

    def test_with_ssn_in_map_value(self):
        """Test with SSN in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        ssn = "123-45-6789"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=ssn)

        assert vo.map_value == ssn

    def test_with_zip_code_in_map_value(self):
        """Test with ZIP code in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        zip = "12345-6789"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=zip)

        assert vo.map_value == zip

    def test_with_country_code_in_map_value(self):
        """Test with country code in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        country = "US"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=country)

        assert vo.map_value == country

    def test_with_currency_code_in_map_value(self):
        """Test with currency code in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        currency = "USD"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=currency)

        assert vo.map_value == currency

    def test_with_language_code_in_map_value(self):
        """Test with language code in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        lang = "en-US"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=lang)

        assert vo.map_value == lang

    def test_with_locale_in_map_value(self):
        """Test with locale in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        locale = "en_US.UTF-8"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=locale)

        assert vo.map_value == locale

    def test_with_timezone_in_map_value(self):
        """Test with timezone in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        tz = "America/New_York"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=tz)

        assert vo.map_value == tz

    def test_with_semantic_version_in_map_value(self):
        """Test with semantic version in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        version = "1.2.3-beta+build.123"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=version)

        assert vo.map_value == version

    def test_with_git_commit_in_map_value(self):
        """Test with git commit hash in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        commit = "a1b2c3d4e5f6g7h8i9j0"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=commit)

        assert vo.map_value == commit

    def test_with_file_path_in_map_value(self):
        """Test with file path in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        path = "/usr/local/bin/test"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=path)

        assert vo.map_value == path

    def test_with_windows_path_in_map_value(self):
        """Test with Windows path in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        path = "C:\\Users\\test\\file.txt"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=path)

        assert vo.map_value == path

    def test_with_unc_path_in_map_value(self):
        """Test with UNC path in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        path = "\\\\server\\share\\file.txt"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=path)

        assert vo.map_value == path

    def test_with_regex_in_map_value(self):
        """Test with regex in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        regex = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=regex)

        assert vo.map_value == regex

    def test_with_markdown_in_map_value(self):
        """Test with markdown in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        md = "# Heading\n\n**Bold** and *italic* text."
        vo = SkillInputVo(input_name="test", input_type="string", map_value=md)

        assert vo.map_value == md

    def test_with_yaml_in_map_value(self):
        """Test with YAML in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        yaml = "key:\n  nested: value\n  list:\n    - item1\n    - item2"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=yaml)

        assert vo.map_value == yaml

    def test_with_cron_expression_in_map_value(self):
        """Test with cron expression in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        cron = "0 0 12 * * ?"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=cron)

        assert vo.map_value == cron

    def test_with_user_agent_in_map_value(self):
        """Test with user agent in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=ua)

        assert vo.map_value == ua

    def test_with_mime_type_in_map_value(self):
        """Test with MIME type in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        mime = "application/json; charset=utf-8"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=mime)

        assert vo.map_value == mime

    def test_with_jwt_in_map_value(self):
        """Test with JWT in map value"""
        from app.domain.vo.agentvo.agent_config_vos import SkillInputVo

        jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ"
        vo = SkillInputVo(input_name="test", input_type="string", map_value=jwt)

        assert vo.map_value == jwt
