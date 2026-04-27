# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/api_tool_pkg/arun_stream_param_processor.py"""

from unittest.mock import MagicMock, patch


class TestProcessRequestBody:
    """测试 _process_request_body 函数"""

    def test_process_request_body_empty(self):
        """测试空请求体"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            _process_request_body,
        )

        body = {}
        _process_request_body(None, {}, {}, body)
        assert body == {}

    def test_process_request_body_no_content(self):
        """测试无 content 的请求体"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            _process_request_body,
        )

        request_body = {"description": "No content"}
        body = {}
        _process_request_body(request_body, {}, {}, body)
        assert body == {}

    def test_process_request_body_simple(self):
        """测试简单请求体"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            _process_request_body,
        )

        request_body = {
            "content": {
                "application/json": {
                    "schema": {
                        "type": "object",
                        "properties": {
                            "name": {"type": "string"},
                            "age": {"type": "integer"},
                        },
                    }
                }
            }
        }
        api_spec = {}
        tool_input = {"name": "test", "age": 25}
        body = {}

        _process_request_body(request_body, api_spec, tool_input, body)

        assert body["name"] == "test"
        assert body["age"] == 25

    def test_process_request_body_with_ref(self):
        """测试带 $ref 的请求体"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            _process_request_body,
        )

        request_body = {
            "content": {
                "application/json": {
                    "schema": {"$ref": "#/components/schemas/UserRequest"}
                }
            }
        }
        api_spec = {
            "components": {
                "schemas": {
                    "UserRequest": {
                        "type": "object",
                        "properties": {
                            "username": {"type": "string"},
                            "email": {"type": "string"},
                        },
                    }
                }
            }
        }
        tool_input = {"username": "user1", "email": "user@example.com"}
        body = {}

        _process_request_body(request_body, api_spec, tool_input, body)

        assert body["username"] == "user1"
        assert body["email"] == "user@example.com"


class TestProcessParams:
    """测试 process_params 函数"""

    def test_process_params_empty(self):
        """测试空参数"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            process_params,
        )

        mock_gvp = MagicMock()
        mock_gvp.get_var_value.return_value = {}
        mock_gvp.get_all_variables.return_value = {}

        path_params, query_params, body, headers = process_params(
            {}, {}, mock_gvp, [], {}
        )

        assert path_params == {}
        assert query_params == {}
        assert body == {}
        assert headers == {}

    def test_process_params_path_params(self):
        """测试路径参数"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            process_params,
        )

        tool_input = {"id": "123", "name": "test"}
        api_spec = {"parameters": [{"name": "id", "in": "path"}]}
        mock_gvp = MagicMock()
        mock_gvp.get_var_value.return_value = {}
        mock_gvp.get_all_variables.return_value = {}

        path_params, query_params, body, headers = process_params(
            tool_input, api_spec, mock_gvp, [], {}
        )

        assert path_params["id"] == "123"

    def test_process_params_query_params(self):
        """测试查询参数"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            process_params,
        )

        tool_input = {"filter": "active", "page": 1}
        api_spec = {
            "parameters": [
                {"name": "filter", "in": "query"},
                {"name": "page", "in": "query"},
            ]
        }
        mock_gvp = MagicMock()
        mock_gvp.get_var_value.return_value = {}
        mock_gvp.get_all_variables.return_value = {}

        path_params, query_params, body, headers = process_params(
            tool_input, api_spec, mock_gvp, [], {}
        )

        assert query_params["filter"] == "active"
        assert query_params["page"] == 1

    def test_process_params_header_params(self):
        """测试头部参数"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            process_params,
        )

        tool_input = {"Authorization": "Bearer token"}
        api_spec = {
            "parameters": [
                {"name": "Authorization", "in": "header", "schema": {"type": "string"}}
            ]
        }
        mock_gvp = MagicMock()
        mock_gvp.get_var_value.return_value = {}
        mock_gvp.get_all_variables.return_value = {}

        path_params, query_params, body, headers = process_params(
            tool_input, api_spec, mock_gvp, [], {}
        )

        assert headers["Authorization"] == "Bearer token"

    def test_process_params_with_request_body(self):
        """测试带请求体参数"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            process_params,
        )

        tool_input = {"username": "user1", "password": "pass123"}
        api_spec = {
            "request_body": {
                "content": {
                    "application/json": {
                        "schema": {
                            "type": "object",
                            "properties": {
                                "username": {"type": "string"},
                                "password": {"type": "string"},
                            },
                        }
                    }
                }
            }
        }
        mock_gvp = MagicMock()
        mock_gvp.get_var_value.return_value = {}
        mock_gvp.get_all_variables.return_value = {}

        path_params, query_params, body, headers = process_params(
            tool_input, api_spec, mock_gvp, [], {}
        )

        assert body["username"] == "user1"
        assert body["password"] == "pass123"

    def test_process_params_with_user_account_headers(self):
        """测试用户账号头部"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            process_params,
        )

        tool_input = {}
        api_spec = {}
        mock_gvp = MagicMock()
        mock_gvp.get_var_value.return_value = {
            "X-User-Account-Id": "user123",
            "X-User-Account-Type": "user",
        }
        mock_gvp.get_all_variables.return_value = {}

        with patch(
            "app.common.tool_v2.api_tool_pkg.arun_stream_param_processor.has_user_account"
        ) as mock_has:
            mock_has.return_value = True
            with patch(
                "app.common.tool_v2.api_tool_pkg.arun_stream_param_processor.get_user_account_id"
            ) as mock_get_id:
                mock_get_id.return_value = "user123"
                with patch(
                    "app.common.tool_v2.api_tool_pkg.arun_stream_param_processor.set_user_account_id"
                ) as mock_set_id:
                    with patch(
                        "app.common.tool_v2.api_tool_pkg.arun_stream_param_processor.has_user_account_type"
                    ) as mock_has_type:
                        mock_has_type.return_value = True
                        with patch(
                            "app.common.tool_v2.api_tool_pkg.arun_stream_param_processor.get_user_account_type"
                        ) as mock_get_type:
                            mock_get_type.return_value = "user"
                            with patch(
                                "app.common.tool_v2.api_tool_pkg.arun_stream_param_processor.set_user_account_type"
                            ) as mock_set_type:
                                path_params, query_params, body, headers = (
                                    process_params(
                                        tool_input, api_spec, mock_gvp, [], {}
                                    )
                                )

    def test_process_params_mixed_params(self):
        """测试混合参数"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            process_params,
        )

        tool_input = {"id": "path123", "filter": "active", "data": {"key": "value"}}
        api_spec = {
            "parameters": [
                {"name": "id", "in": "path"},
                {"name": "filter", "in": "query"},
            ],
            "request_body": {
                "content": {
                    "application/json": {
                        "schema": {
                            "type": "object",
                            "properties": {"data": {"type": "object"}},
                        }
                    }
                }
            },
        }
        mock_gvp = MagicMock()
        mock_gvp.get_var_value.return_value = {}
        mock_gvp.get_all_variables.return_value = {}

        path_params, query_params, body, headers = process_params(
            tool_input, api_spec, mock_gvp, [], {}
        )

        assert path_params["id"] == "path123"
        assert query_params["filter"] == "active"
        assert body["data"] == {"key": "value"}


class TestModuleImports:
    """测试模块导入"""

    def test_import_process_params(self):
        """测试导入 process_params"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            process_params,
        )

        assert callable(process_params)

    def test_import_process_request_body(self):
        """测试导入 _process_request_body"""
        from app.common.tool_v2.api_tool_pkg.arun_stream_param_processor import (
            _process_request_body,
        )

        assert callable(_process_request_body)
