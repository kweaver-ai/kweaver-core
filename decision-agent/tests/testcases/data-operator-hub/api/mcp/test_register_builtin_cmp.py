# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import uuid

from lib.mcp_internal import InternalMCP
from lib.mcp import MCP

@allure.feature("MCP服务管理接口测试：注册内置MCP服务")
class TestRegisterBuiltinMCP:
    
    client = InternalMCP()
    mcp_client = MCP()
    mcp_id = ""

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, UserHeaders):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        mcp_id = str(uuid.uuid4())
        TestRegisterBuiltinMCP.mcp_id = mcp_id
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
        # print(data)
        result = self.client.Register(data, UserHeaders)
        # print(result)
        assert result[0] == 200

    @allure.title("注册内置MCP，传参正确，注册成功，默认状态为published，类型为system")
    def test_register_builtin_mcp_01(self, UserHeaders, Headers):
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
        assert "mcp_id" in result[1]
        assert result[1]["status"] == "published"

        result = self.mcp_client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        base_info = result[1]["base_info"]
        assert base_info["is_internal"] == True
        assert base_info["category"] == "system"

    @allure.title("更新内置工具，传参正确，更新成功")
    def test_update_builtin_mcp_02(self, Headers, UserHeaders):
        new_name = "updated_" + ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "mcp_id": TestRegisterBuiltinMCP.mcp_id,
            "name": new_name,
            "description": "updated description",
            "mode": "sse",
            "source": "intenal",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Register(data, UserHeaders)
        assert result[0] == 200
        
        detail_res = self.mcp_client.GetMCPDetail(TestRegisterBuiltinMCP.mcp_id, Headers)
        assert detail_res[0] == 200
        assert detail_res[1]["base_info"]["name"] == new_name
        assert detail_res[1]["base_info"]["description"] == "updated description"

    @allure.title("更新内置工具，名称不合法，更新失败")
    @pytest.mark.parametrize("name", ["me~!@#$%^&*()_+{}|:<>?,./;'\\[]-=`《》？，。：”【】、", "a" * 51])
    def test_update_builtin_mcp_03(self, name, UserHeaders):
        data = {
            "mcp_id": TestRegisterBuiltinMCP.mcp_id,
            "name": name,
            "description": "updated description",
            "mode": "sse",
            "source": "intenal",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        # print(data)
        result = self.client.Register(data, UserHeaders)
        assert result[0] == 400

    @allure.title("更新内置工具，描述不合法，更新失败")
    def test_update_builtin_mcp_04(self, UserHeaders):
        data = {
            "mcp_id": TestRegisterBuiltinMCP.mcp_id,
            "description": "a" * 256,
            "name": ''.join(random.choice(string.ascii_letters) for i in range(8)),
            "mode": "sse",
            "source": "intenal",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "config_version": "1.0.0",
            "config_source": "auto"
        }
        result = self.client.Register(data, UserHeaders)
        assert result[0] == 400

    @allure.title("更新内置工具，版本格式错误，更新失败")
    @pytest.mark.parametrize("version", ["v1", "first_version"])
    def test_update_builtin_mcp_06(self, version, UserHeaders):
        data = {
            "mcp_id": TestRegisterBuiltinMCP.mcp_id,
            "name": ''.join(random.choice(string.ascii_letters) for i in range(8)),
            "description": "updated description 11",
            "mode": "sse",
            "source": "intenal",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "config_version": version,
            "config_source": "auto"
        }
        result = self.client.Register(data, UserHeaders)
        assert result[0] == 400

    @allure.title("注册内置MCP，存在同名mcp，注册失败")
    def test_update_builtin_mcp_07(self, UserHeaders, Headers):
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
        result = self.mcp_client.RegisterMCP(data, Headers)
        assert result[0] == 200
        data = {
            "status": "published"
        }
        result = self.mcp_client.MCPReleaseAction(result[1]["mcp_id"], data, Headers)
        assert result[0] == 200

        mcp_id = str(uuid.uuid4())
        TestRegisterBuiltinMCP.mcp_id = mcp_id
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
        assert result[0] == 400

    @allure.title("手动更新内置MCP，不加保护锁，自动更新后手动更新内容被覆盖")
    def test_update_builtin_mcp_08(self, Headers, UserHeaders):
        # 手动更新
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "mcp_id": TestRegisterBuiltinMCP.mcp_id,
            "name": name,
            "description": "updated description 111",
            "mode": "sse",
            "source": "intenal",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "config_version": "1.0.0",
            "config_source": "manual",
            "protected_flag": False
        }
        result = self.client.Register(data, UserHeaders)
        assert result[0] == 200
        # 验证更新后内容
        detail_res = self.mcp_client.GetMCPDetail(TestRegisterBuiltinMCP.mcp_id, Headers)
        assert detail_res[0] == 200
        assert detail_res[1]["base_info"]["name"] == name
        assert detail_res[1]["base_info"]["description"] == "updated description 111"
        assert detail_res[1]["base_info"]["category"] == "system"
        assert detail_res[1]["base_info"]["url"] == "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL"

        # 自动更新
        name1 = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "mcp_id": TestRegisterBuiltinMCP.mcp_id,
            "name": name1,
            "description": "updated description 222",
            "mode": "sse",
            "source": "intenal",
            "url": "https://mcp.amap.com/sse?key=5dc290f1ad89616a",
            "config_version": "1.0.0",
            "config_source": "auto",
            "protected_flag": False
        }
        result = self.client.Register(data, UserHeaders)
        assert result[0] == 200
        # 验证更新后内容，自动更新后手动更新内容被覆盖
        detail_res = self.mcp_client.GetMCPDetail(TestRegisterBuiltinMCP.mcp_id, Headers)
        assert detail_res[0] == 200
        assert detail_res[1]["base_info"]["name"] == name1
        assert detail_res[1]["base_info"]["description"] == "updated description 222"
        assert detail_res[1]["base_info"]["category"] == "system"
        assert detail_res[1]["base_info"]["url"] == "https://mcp.amap.com/sse?key=5dc290f1ad89616a"

    @allure.title("手动更新内置MCP，加保护锁，自动更新后手动更新内容被保留")
    def test_update_builtin_mcp_09(self, Headers, UserHeaders):
        # 手动更新
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "mcp_id": TestRegisterBuiltinMCP.mcp_id,
            "name": name,
            "description": "updated description 111",
            "mode": "sse",
            "source": "intenal",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "config_version": "1.0.0",
            "config_source": "manual",
            "protected_flag": True
        }
        result = self.client.Register(data, UserHeaders)
        assert result[0] == 200
        # 验证更新后内容
        detail_res = self.mcp_client.GetMCPDetail(TestRegisterBuiltinMCP.mcp_id, Headers)
        assert detail_res[0] == 200
        assert detail_res[1]["base_info"]["name"] == name
        assert detail_res[1]["base_info"]["description"] == "updated description 111"
        assert detail_res[1]["base_info"]["url"] == "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL"

        # 自动更新
        name1 = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "mcp_id": TestRegisterBuiltinMCP.mcp_id,
            "name": name1,
            "description": "updated description 222",
            "mode": "sse",
            "source": "intenal",
            "url": "https://mcp.amap.com/sse?key=5dc290f1ad89616a",
            "config_version": "1.0.0",
            "config_source": "auto",
            "protected_flag": True
        }
        result = self.client.Register(data, UserHeaders)
        assert result[0] == 200
        # 验证更新后内容，自动更新后手动更新内容保留
        detail_res = self.mcp_client.GetMCPDetail(TestRegisterBuiltinMCP.mcp_id, Headers)
        assert detail_res[0] == 200
        assert detail_res[1]["base_info"]["name"] == name
        assert detail_res[1]["base_info"]["description"] == "updated description 111"
        assert detail_res[1]["base_info"]["url"] == "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL"