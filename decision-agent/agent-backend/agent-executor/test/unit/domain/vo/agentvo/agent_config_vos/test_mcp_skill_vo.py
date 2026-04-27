"""单元测试 - domain/vo/agentvo/agent_config_vos/mcp_skill_vo 模块"""

import pytest
import json


class TestMcpSkillVo:
    """测试 McpSkillVo 类"""

    def test_init_with_required_field(self):
        """测试使用必填字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp_server_123")

        assert vo.mcp_server_id == "mcp_server_123"

    def test_is_pydantic_model(self):
        """测试是Pydantic模型"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo
        from pydantic import BaseModel

        assert issubclass(McpSkillVo, BaseModel)

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp_server_123")
        data = vo.model_dump()

        assert data["mcp_server_id"] == "mcp_server_123"

    def test_model_dump_json(self):
        """测试JSON序列化"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp_server_123")
        json_str = vo.model_dump_json()

        assert "mcp_server_123" in json_str

    def test_with_empty_string_id(self):
        """测试空字符串server_id"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="")
        assert vo.mcp_server_id == ""

    def test_with_special_characters(self):
        """测试特殊字符"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp-server_123@test")
        assert vo.mcp_server_id == "mcp-server_123@test"

    def test_with_unicode_id(self):
        """测试Unicode字符"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp_测试_123")
        assert vo.mcp_server_id == "mcp_测试_123"

    def test_with_numeric_string(self):
        """测试数字字符串"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="12345")
        assert vo.mcp_server_id == "12345"

    def test_with_underscore(self):
        """测试下划线"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp_server_id")
        assert vo.mcp_server_id == "mcp_server_id"

    def test_with_hyphen(self):
        """测试连字符"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp-server-id")
        assert vo.mcp_server_id == "mcp-server-id"

    def test_copy(self):
        """测试复制"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo1 = McpSkillVo(mcp_server_id="server_1")
        vo2 = vo1.copy()

        assert vo2.mcp_server_id == "server_1"

    def test_equality(self):
        """测试相等性"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo1 = McpSkillVo(mcp_server_id="server_1")
        vo2 = McpSkillVo(mcp_server_id="server_1")
        vo3 = McpSkillVo(mcp_server_id="server_2")

        assert vo1 == vo2
        assert vo1 != vo3

    def test_from_dict(self):
        """测试从字典创建"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        data = {"mcp_server_id": "server_123"}
        vo = McpSkillVo(**data)

        assert vo.mcp_server_id == "server_123"

    def test_model_schema(self):
        """测试模型schema"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        schema = McpSkillVo.model_json_schema()
        assert "properties" in schema
        assert "mcp_server_id" in schema["properties"]

    def test_field_description(self):
        """测试字段描述"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        field_info = McpSkillVo.model_fields["mcp_server_id"]
        assert field_info.description is not None

    def test_with_uuid_like_id(self):
        """测试UUID格式的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="550e8400-e29b-41d4-a716-446655440000")
        assert vo.mcp_server_id == "550e8400-e29b-41d4-a716-446655440000"

    def test_with_long_id(self):
        """测试长ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        long_id = "mcp_server_" + "x" * 100
        vo = McpSkillVo(mcp_server_id=long_id)
        assert vo.mcp_server_id == long_id

    def test_json_serialization(self):
        """测试JSON序列化"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="server_1")
        json_str = json.dumps(vo.model_dump())

        data = json.loads(json_str)
        assert data["mcp_server_id"] == "server_1"

    def test_model_dump_excludes_none(self):
        """测试序列化时排除None（实际上没有可选字段）"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="server_1")
        data = vo.model_dump()

        # 只有一个字段
        assert len(data) == 1
        assert "mcp_server_id" in data

    def test_field_required(self):
        """测试字段必填"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo
        from pydantic import ValidationError

        with pytest.raises(ValidationError):
            McpSkillVo()

    def test_repr(self):
        """测试字符串表示"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="server_1")
        repr_str = repr(vo)

        assert "mcp_server_id" in repr_str or "McpSkillVo" in repr_str

    def test_with_whitespace_id(self):
        """测试带空格的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp server 123")
        assert vo.mcp_server_id == "mcp server 123"

    def test_model_rebuild(self):
        """测试模型重建"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        McpSkillVo.model_rebuild()

        vo = McpSkillVo(mcp_server_id="server_1")
        assert vo.mcp_server_id == "server_1"


class TestMcpSkillVoExtended:
    """Extended tests for McpSkillVo"""

    def test_init_with_various_formats(self):
        """测试各种格式的server_id"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        # CamelCase
        vo1 = McpSkillVo(mcp_server_id="McpServerId")
        assert vo1.mcp_server_id == "McpServerId"

        # snake_case
        vo2 = McpSkillVo(mcp_server_id="mcp_server_id")
        assert vo2.mcp_server_id == "mcp_server_id"

        # kebab-case
        vo3 = McpSkillVo(mcp_server_id="mcp-server-id")
        assert vo3.mcp_server_id == "mcp-server-id"

    def test_with_dots(self):
        """测试带点号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp.server.v1")
        assert vo.mcp_server_id == "mcp.server.v1"

    def test_with_colons(self):
        """测试带冒号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp:server:123")
        assert vo.mcp_server_id == "mcp:server:123"

    def test_with_slashes(self):
        """测试带斜杠的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp/server/123")
        assert vo.mcp_server_id == "mcp/server/123"

    def test_with_at_symbol(self):
        """测试带@符号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp@server@123")
        assert vo.mcp_server_id == "mcp@server@123"

    def test_with_plus_sign(self):
        """测试带+号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp+server+123")
        assert vo.mcp_server_id == "mcp+server+123"

    def test_with_equals_sign(self):
        """测试带等号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp=server=123")
        assert vo.mcp_server_id == "mcp=server=123"

    def test_with_pipe(self):
        """测试带管道符的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp|server|123")
        assert vo.mcp_server_id == "mcp|server|123"

    def test_with_tilde(self):
        """测试带波浪号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp~server~123")
        assert vo.mcp_server_id == "mcp~server~123"

    def test_with_percent(self):
        """测试带百分号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp%server%123")
        assert vo.mcp_server_id == "mcp%server%123"

    def test_with_ampersand(self):
        """测试带&符号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp&server&123")
        assert vo.mcp_server_id == "mcp&server&123"

    def test_with_hash(self):
        """测试带#号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp#server#123")
        assert vo.mcp_server_id == "mcp#server#123"

    def test_with_dollar_sign(self):
        """测试带$符号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp$server$123")
        assert vo.mcp_server_id == "mcp$server$123"

    def test_with_exclamation(self):
        """测试带感叹号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp!server!123")
        assert vo.mcp_server_id == "mcp!server!123"

    def test_with_question_mark(self):
        """测试带问号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp?server?123")
        assert vo.mcp_server_id == "mcp?server?123"

    def test_with_asterisk(self):
        """测试带星号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp*server*123")
        assert vo.mcp_server_id == "mcp*server*123"

    def test_with_parentheses(self):
        """测试带括号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp(server)123")
        assert vo.mcp_server_id == "mcp(server)123"

    def test_with_brackets(self):
        """测试带方括号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp[server]123")
        assert vo.mcp_server_id == "mcp[server]123"

    def test_with_braces(self):
        """测试带花括号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp{server}123")
        assert vo.mcp_server_id == "mcp{server}123"

    def test_with_angle_brackets(self):
        """测试带尖括号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp<server>123")
        assert vo.mcp_server_id == "mcp<server>123"

    def test_with_backslash(self):
        """测试带反斜杠的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp\\server\\123")
        assert vo.mcp_server_id == "mcp\\server\\123"

    def test_with_comma(self):
        """测试带逗号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp,server,123")
        assert vo.mcp_server_id == "mcp,server,123"

    def test_with_semicolon(self):
        """测试带分号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp;server;123")
        assert vo.mcp_server_id == "mcp;server;123"

    def test_with_single_quote(self):
        """测试带单引号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp'server'123")
        assert vo.mcp_server_id == "mcp'server'123"

    def test_with_double_quote(self):
        """测试带双引号的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id='mcp"server"123')
        assert vo.mcp_server_id == 'mcp"server"123'

    def test_with_space_only(self):
        """测试仅空格的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="   ")
        assert vo.mcp_server_id == "   "

    def test_with_tab(self):
        """测试带制表符的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp\tserver\t123")
        assert vo.mcp_server_id == "mcp\tserver\t123"

    def test_with_newline(self):
        """测试带换行符的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp\nserver\n123")
        assert vo.mcp_server_id == "mcp\nserver\n123"

    def test_with_carriage_return(self):
        """测试带回车符的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp\rserver\r123")
        assert vo.mcp_server_id == "mcp\rserver\r123"

    def test_with_null_byte(self):
        """测试带空字节的ID"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="mcp\x00server\x00123")
        assert vo.mcp_server_id == "mcp\x00server\x00123"

    def test_multiple_instances(self):
        """测试多个实例"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo1 = McpSkillVo(mcp_server_id="server_1")
        vo2 = McpSkillVo(mcp_server_id="server_2")
        vo3 = McpSkillVo(mcp_server_id="server_3")

        assert vo1.mcp_server_id == "server_1"
        assert vo2.mcp_server_id == "server_2"
        assert vo3.mcp_server_id == "server_3"

    def test_immutability_of_field(self):
        """测试字段的可变性（Pydantic v2默认可变）"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="server_1")
        vo.mcp_server_id = "server_2"

        assert vo.mcp_server_id == "server_2"

    def test_parse_obj(self):
        """测试parse_obj方法"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo.model_validate({"mcp_server_id": "server_1"})
        assert vo.mcp_server_id == "server_1"

    def test_validate_assignment(self):
        """测试赋值验证"""
        from app.domain.vo.agentvo.agent_config_vos import McpSkillVo

        vo = McpSkillVo(mcp_server_id="server_1")
        # 任何字符串值都是有效的
        vo.mcp_server_id = "new_server"
        assert vo.mcp_server_id == "new_server"
