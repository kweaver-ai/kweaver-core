# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import uuid

from lib.mcp_internal import InternalMCP
from lib.mcp import MCP

mcp_id = ""

@allure.feature("MCP服务管理接口测试：注销内置MCP服务")
class TestUnRegisterBuiltinMCP:
    
    client = InternalMCP()
    client1 = MCP()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, UserHeaders):
        global mcp_id
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        mcp_id = str(uuid.uuid4())
        data = {
            "mcp_id": mcp_id,
            "name": name,
            "description": "test builtin mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "command": "ls",
            "args": ["a", "b", "c"],
            "headers": {
                "Content-Type": "application/json"
            },
            "env": {},
            "source": "intenal",
            "protected_flag": False,
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Register(data, UserHeaders)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]

    @allure.title("注销内置MCP，内置MCP不存在，注销失败")
    def test_unregister_builtin_mcp_01(self, UserHeaders):
        mcp_id = str(uuid.uuid4())
        result = self.client.UnRegister(mcp_id, {}, UserHeaders)
        assert result[0] == 400

    @allure.title("注销非内置MCP，注销失败")
    def test_unregister_builtin_mcp_02(self, Headers, UserHeaders):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "http://localhost:8080/api/v1/tools",
            "headers": {
                "Content-Type": "application/json"
            },
            "category": "data_analysis"
        }
        result = self.client1.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]

        result = self.client.UnRegister(mcp_id, None, UserHeaders)
        assert result[0] == 400

        result = self.client1.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200

    @allure.title("注销内置MCP，内置MCP存在，注销成功")
    def test_unregister_builtin_mcp_03(self, UserHeaders, Headers):
        global mcp_id

        result = self.client.UnRegister(mcp_id, None, UserHeaders)
        assert result[0] == 200

        result = self.client1.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 404