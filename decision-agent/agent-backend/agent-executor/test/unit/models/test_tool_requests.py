"""单元测试 - models/tool_requests 模块"""

import pytest
from pydantic import ValidationError


class TestZhipuSearchRequest:
    """测试 ZhipuSearchRequest 模型"""

    def test_default_initialization(self):
        """测试默认初始化"""
        from app.models.tool_requests import ZhipuSearchRequest

        request = ZhipuSearchRequest(query="test query")

        assert request.query == "test query"

    def test_with_all_fields(self):
        """测试所有字段"""
        from app.models.tool_requests import ZhipuSearchRequest

        request = ZhipuSearchRequest(query="machine learning")

        assert request.query == "machine learning"

    def test_query_is_required(self):
        """测试query是必填项"""
        from app.models.tool_requests import ZhipuSearchRequest

        with pytest.raises(ValidationError):
            ZhipuSearchRequest()

    def test_model_dump(self):
        """测试模型序列化"""
        from app.models.tool_requests import ZhipuSearchRequest

        request = ZhipuSearchRequest(query="test")
        data = request.model_dump()

        assert data["query"] == "test"

    def test_model_dump_json(self):
        """测试JSON序列化"""
        from app.models.tool_requests import ZhipuSearchRequest

        request = ZhipuSearchRequest(query="test")
        json_str = request.model_dump_json()

        assert "test" in json_str


class TestGetSchemaRequest:
    """测试 GetSchemaRequest 模型"""

    def test_default_initialization(self):
        """测试默认初始化"""
        from app.models.tool_requests import GetSchemaRequest

        request = GetSchemaRequest(database="test_db")

        assert request.database == "test_db"

    def test_with_all_fields(self):
        """测试所有字段"""
        from app.models.tool_requests import GetSchemaRequest

        request = GetSchemaRequest(database="my_database")

        assert request.database == "my_database"

    def test_database_is_required(self):
        """测试database是必填项"""
        from app.models.tool_requests import GetSchemaRequest

        with pytest.raises(ValidationError):
            GetSchemaRequest()

    def test_model_dump(self):
        """测试模型序列化"""
        from app.models.tool_requests import GetSchemaRequest

        request = GetSchemaRequest(database="test_db")
        data = request.model_dump()

        assert data["database"] == "test_db"


class TestOnlineSearchCiteRequest:
    """测试 OnlineSearchCiteRequest 模型"""

    def test_default_initialization(self):
        """测试默认初始化"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        request = OnlineSearchCiteRequest(
            query="test",
            model_name="model",
            search_tool="tool",
            api_key="key",
            user_id="user123",
        )

        assert request.query == "test"
        assert request.model_name == "model"
        assert request.search_tool == "tool"
        assert request.api_key == "key"
        assert request.user_id == "user123"
        assert request.stream is False  # Default value

    def test_with_stream_true(self):
        """测试启用流式"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        request = OnlineSearchCiteRequest(
            query="test",
            model_name="model",
            search_tool="tool",
            api_key="key",
            user_id="user123",
            stream=True,
        )

        assert request.stream is True

    def test_all_fields_required(self):
        """测试所有必填字段"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        with pytest.raises(ValidationError):
            OnlineSearchCiteRequest()

        with pytest.raises(ValidationError):
            OnlineSearchCiteRequest(query="test")

        with pytest.raises(ValidationError):
            OnlineSearchCiteRequest(query="test", model_name="model")

    def test_model_dump(self):
        """测试模型序列化"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        request = OnlineSearchCiteRequest(
            query="test",
            model_name="model",
            search_tool="tool",
            api_key="key",
            user_id="user123",
        )
        data = request.model_dump()

        assert data["query"] == "test"
        assert data["stream"] is False

    def test_model_dump_json(self):
        """测试JSON序列化"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        request = OnlineSearchCiteRequest(
            query="test",
            model_name="model",
            search_tool="tool",
            api_key="key",
            user_id="user123",
        )
        json_str = request.model_dump_json()

        assert "test" in json_str

    # Extended tests for coverage improvement

    def test_empty_string_query(self):
        """测试空字符串查询"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        # Empty string should be valid for query field (it's just str type, not min_length)
        request = OnlineSearchCiteRequest(
            query="",
            model_name="model",
            search_tool="tool",
            api_key="key",
            user_id="user123",
        )
        assert request.query == ""

    def test_unicode_characters(self):
        """测试Unicode字符处理"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        request = OnlineSearchCiteRequest(
            query="测试搜索",
            model_name="模型-名称",
            search_tool="工具-搜索",
            api_key="密钥-123",
            user_id="用户@#$%",
        )
        assert request.query == "测试搜索"
        assert request.model_name == "模型-名称"
        assert request.search_tool == "工具-搜索"

    def test_special_characters(self):
        """测试特殊字符处理"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        request = OnlineSearchCiteRequest(
            query="test with special chars: <script>alert('xss')</script>",
            model_name="model<>",
            search_tool="tool&name",
            api_key="key=123",
            user_id="user@example.com",
        )
        assert "<script>" in request.query
        assert request.model_name == "model<>"

    def test_long_string_values(self):
        """测试长字符串值"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        long_query = "test " * 1000
        request = OnlineSearchCiteRequest(
            query=long_query,
            model_name="model",
            search_tool="tool",
            api_key="key",
            user_id="user123",
        )
        assert len(request.query) == len(long_query)

    def test_query_with_newlines_and_tabs(self):
        """测试包含换行符和制表符的查询"""
        from app.models.tool_requests import OnlineSearchCiteRequest

        request = OnlineSearchCiteRequest(
            query="line1\nline2\ttabbed",
            model_name="model",
            search_tool="tool",
            api_key="key",
            user_id="user123",
        )
        assert "\n" in request.query
        assert "\t" in request.query


class TestZhipuSearchRequestExtended:
    """测试 ZhipuSearchRequest 模型的扩展测试"""

    def test_empty_string_query(self):
        """测试空字符串查询"""
        from app.models.tool_requests import ZhipuSearchRequest

        request = ZhipuSearchRequest(query="")
        assert request.query == ""

    def test_unicode_query(self):
        """测试Unicode查询"""
        from app.models.tool_requests import ZhipuSearchRequest

        request = ZhipuSearchRequest(query="搜索机器学习算法")
        assert request.query == "搜索机器学习算法"

    def test_special_characters_in_query(self):
        """测试查询中的特殊字符"""
        from app.models.tool_requests import ZhipuSearchRequest

        special_chars = "!@#$%^&*()_+-=[]{}|;':\",./<>?"
        request = ZhipuSearchRequest(query=special_chars)
        assert request.query == special_chars

    def test_query_with_newlines(self):
        """测试包含换行符的查询"""
        from app.models.tool_requests import ZhipuSearchRequest

        request = ZhipuSearchRequest(query="line1\nline2\nline3")
        assert "\n" in request.query


class TestGetSchemaRequestExtended:
    """测试 GetSchemaRequest 模型的扩展测试"""

    def test_empty_string_database(self):
        """测试空字符串数据库名"""
        from app.models.tool_requests import GetSchemaRequest

        request = GetSchemaRequest(database="")
        assert request.database == ""

    def test_unicode_database_name(self):
        """测试Unicode数据库名"""
        from app.models.tool_requests import GetSchemaRequest

        request = GetSchemaRequest(database="数据库_测试")
        assert request.database == "数据库_测试"

    def test_special_characters_in_database(self):
        """测试数据库名中的特殊字符"""
        from app.models.tool_requests import GetSchemaRequest

        request = GetSchemaRequest(database="db@#$%^&*()")
        assert request.database == "db@#$%^&*()"

    def test_database_with_dots_and_underscores(self):
        """测试包含点和下划线的数据库"""
        from app.models.tool_requests import GetSchemaRequest

        request = GetSchemaRequest(database="test_db.production_2024")
        assert request.database == "test_db.production_2024"
