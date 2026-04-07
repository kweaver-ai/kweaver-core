# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import uuid

from common.get_content import GetContent
from common.get_token import GetToken
from lib.mcp import MCP
from lib.mcp_internal import InternalMCP
from lib.permission import Perm


@allure.feature("算子平台权限测试：MCP权限测试")
class TestMCPPerm:
    client = MCP()
    client1 = InternalMCP()
    perm_client = Perm()
    a_headers = {}
    b_headers = {}
    c_headers = {}
    mcp_id = ""
    user_list = []

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers, PermPrepare):
        # 获取用户token
        configfile = "./config/env.ini"
        file = GetContent(configfile)
        config = file.config()
        host = config["server"]["host"]
        a_token = GetToken(host=host).get_token(host, "a", "111111")
        TestMCPPerm.a_headers = {
            "Authorization": f"Bearer {a_token[1]}"
        }
        
        b_token = GetToken(host=host).get_token(host, "b", "111111")
        TestMCPPerm.b_headers = {
            "Authorization": f"Bearer {b_token[1]}"
        }

        c_token = GetToken(host=host).get_token(host, "c", "111111")
        TestMCPPerm.c_headers = {
            "Authorization": f"Bearer {c_token[1]}"
        }

        d_token = GetToken(host=host).get_token(host, "d", "111111")
        TestMCPPerm.d_headers = {
            "Authorization": f"Bearer {d_token[1]}"
        }

        # AI管理员创建MCP
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "source": "custom",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        TestMCPPerm.mcp_id = result[1]["mcp_id"]

        '''
        为用户a设置mcp资源的新建权限
        为用户b设置该mcp实例的查看、编辑、删除权限
        为用户c设置该mcp实例的发布、下架、公共访问和使用权限
        '''
        TestMCPPerm.user_list = PermPrepare[1]
        # user_list = ['caad2ad2-56ec-11f0-bce8-8269137aaf40', 'cace09a0-56ec-11f0-8f88-8269137aaf40', 'caec81f0-56ec-11f0-9591-8269137aaf40', 'cb2e864a-56ec-11f0-b88f-8269137aaf40']
        data = [
            {
                "accessor": {"id": TestMCPPerm.user_list[0], "name": "a", "type": "user"},
                "resource": {"id": "*", "type": "mcp", "name": "新建权限"},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            },
            {
                "accessor": {"id": TestMCPPerm.user_list[1], "name": "b", "type": "user"},
                "resource": {"id": TestMCPPerm.mcp_id, "type": "mcp", "name": "mcp实例查看编辑删除权限"},
                "operation": {"allow": [{"id": "view"}, {"id": "modify"}, { "id": "delete"}], "deny": []}
            },
            {
                "accessor": {"id": TestMCPPerm.user_list[2], "name": "c", "type": "user"},
                "resource": {"id": TestMCPPerm.mcp_id, "type": "mcp", "name": "mcp实例发布下架公开访问使用权限"},
                "operation": {"allow": [{"id": "publish"}, {"id": "unpublish"}, {"id": "public_access"}, {"id": "execute"}], "deny": []}
            }
        ]
        result = self.perm_client.SetPerm(data, Headers)
        assert "20" in str(result[0])
        
    @allure.title("有新建权限，新建MCP，新建成功，创建者和AI管理员可对该MCP进行所有操作")
    def test_mcp_permission_01(self, Headers):
        # 新建
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        create_data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "source": "custom",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(create_data, TestMCPPerm.a_headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]

        headers = [Headers, TestMCPPerm.a_headers]
        for header in headers:
            # 查看
            data = { "page_size": 100 }
            result = self.client.GetMCPList(data, header)
            assert result[0] == 200
            assert mcp_id in str(result[1]["data"])

            result = self.client.GetMCPDetail(mcp_id, header)
            assert result[0] == 200

            # 编辑
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            data = {
                "name": name,
                "description": "edit test mcp server",
                "mode": "sse",
                "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
                "category": "data_process"
            }
            result = self.client.EditMCP(mcp_id, data, header)
            assert result[0] == 200

            # 使用
            result = self.client.GetMCPToolList(mcp_id, header)
            assert result[0] == 200

            data = {
                "tool_name": "map_geocode",
                "parameters": {
                    "address": "上海东方明珠"
                }
            }
            result = self.client.CallMCPtool(mcp_id, data, header)
            assert result[0] == 200

            data = {
                "parameters": {
                    "address": "上海东方明珠"
                }
            }
            result = self.client.MCPToolDebug(mcp_id, "map_geocode", data, header)
            assert result[0] == 200

            # 发布
            data = {
                "status": "published"
            }
            result = self.client.MCPReleaseAction(mcp_id, data, header)
            assert result[0] == 200

            # 公开访问
            params = {
            "page_size": 50
            }
            result = self.client.GetMCPList(params, header)
            assert result[0] == 200
            assert mcp_id in str(result[1]["data"])

            result = self.client.GetMCPMarketDetail(mcp_id, header)
            assert result[0] == 200

            result = self.client.BatchGetMCPMarketDetail(mcp_id, "name,description", header)
            assert result[0] == 200

            # 下架
            data = {
                "status": "offline"
            }
            result = self.client.MCPReleaseAction(mcp_id, data, header)
            assert result[0] == 200

            # 删除
            result = self.client.DeleteMCP(mcp_id, header)
            assert result[0] == 200

            # 重新创建
            result = self.client.RegisterMCP(create_data, TestMCPPerm.a_headers)
            assert result[0] == 200
            mcp_id = result[1]["mcp_id"]

    @allure.title("新建内置MCP，AI管理员可对该内置MCP进行所有操作，普通用户可公开访问和使用该内置MCP")
    def test_mcp_permission_02(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        mcp_id = str(uuid.uuid4())
        data = {
            "mcp_id": mcp_id,
            "name": name,
            "description": "test intcomp mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "source": "intenal",
            "config_version": "1.0.0",
            "config_source": "auto",
            "protected_flag": False
        }
        # 用户a新建内置mcp
        result = self.client1.Register(data, TestMCPPerm.a_headers)
        assert result[0] == 200

        # 普通用户可公开访问和使用
        result = self.client.GetMCPMarketDetail(mcp_id, TestMCPPerm.b_headers)
        assert result[0] == 200
        result = self.client.GetMCPToolList(mcp_id, TestMCPPerm.b_headers)
        assert result[0] == 200
        result = self.client.MCPToolDebug(mcp_id, "map_geocode", {"parameters": {"address": "上海东方明珠"}}, TestMCPPerm.b_headers)
        assert result[0] == 200
        data = {"tool_name": "map_geocode", "parameters": {"address": "上海东方明珠"}}
        result = self.client.CallMCPtool(mcp_id, data, TestMCPPerm.b_headers)
        assert result[0] == 200
        # 普通用户不可编辑
        result = self.client.EditMCP(mcp_id, {"name": "edit_fail", "mode": "sse", "url": "http://127.0.0.1"}, TestMCPPerm.b_headers)
        assert result[0] == 403
        # AI管理员给用户b配置权限
        perm_data = [{
                "accessor": {"id": TestMCPPerm.user_list[2], "name": "c", "type": "user"},
                "resource": {"id": mcp_id, "type": "mcp", "name": "配置权限"},
                "operation": {"allow": [{"id": "view"}], "deny": []}
            }]
        result = self.perm_client.SetPerm(perm_data, Headers)
        assert "20" in str(result[0])
        # AI管理员可编辑、发布、下架、删除
        result = self.client.EditMCP(mcp_id, {"name": "edit_success", "mode": "sse", "url": "http://127.0.0.1"}, Headers)
        assert result[0] == 200
        result = self.client.MCPReleaseAction(mcp_id, {"status": "published"}, Headers)
        assert result[0] == 200
        result = self.client.MCPReleaseAction(mcp_id, {"status": "offline"}, Headers)
        assert result[0] == 200
        result = self.client.DeleteMCP(mcp_id, Headers)
        assert result[0] == 200

    @allure.title("新建MCP，无权限新建失败，有权限新建成功")
    def test_mcp_permission_03(self):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "source": "custom",
            "category": "data_analysis"
        }
        # 用户b无新建权限
        result = self.client.RegisterMCP(data, TestMCPPerm.b_headers)
        assert result[0] == 403
        # 用户a有新建权限
        result = self.client.RegisterMCP(data, TestMCPPerm.a_headers)
        assert result[0] == 200

    @allure.title("获取mcp详情，无查看权限获取失败，有查看权限获取成功")
    def test_mcp_permission_04(self):
        # 用户c无查看权限
        result = self.client.GetMCPDetail(TestMCPPerm.mcp_id, TestMCPPerm.c_headers)
        assert result[0] == 403
        # 用户b有查看权限
        result = self.client.GetMCPDetail(TestMCPPerm.mcp_id, TestMCPPerm.b_headers)
        assert result[0] == 200

    @allure.title("获取MCP列表，无法获取到无查看权限的MCP，可获取到有查看权限的MCP")
    def test_mcp_permission_05(self):
        # 用户c无查看权限，其列表不应包含mcp_id
        result = self.client.GetMCPList({"all": True}, TestMCPPerm.c_headers)
        assert TestMCPPerm.mcp_id not in [m["mcp_id"] for m in result[1]["data"]]
        # 用户b有查看权限，其列表应包含mcp_id
        result = self.client.GetMCPList({"all": True}, TestMCPPerm.b_headers)
        assert TestMCPPerm.mcp_id in [m["mcp_id"] for m in result[1]["data"]]

    @allure.title("编辑MCP，无编辑权限编辑失败，有编辑权限编辑成功")
    def test_mcp_permission_06(self):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "edit test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "source": "custom",
            "category": "data_analysis"
        }
        # 用户c无编辑权限
        result = self.client.EditMCP(TestMCPPerm.mcp_id, data, TestMCPPerm.c_headers)
        assert result[0] == 403
        # 用户b有编辑权限
        result = self.client.EditMCP(TestMCPPerm.mcp_id, data, TestMCPPerm.b_headers)
        assert result[0] == 200

    @allure.title("发布、下架MCP，无权限则发布、下架失败，有权限则发布、下架成功")
    def test_mcp_permission_07(self):
        data_pub = {"status": "published"}
        data_off = {"status": "offline"}
        # 用户b无发布权限
        result = self.client.MCPReleaseAction(TestMCPPerm.mcp_id, data_pub, TestMCPPerm.b_headers)
        assert result[0] == 403
        # 用户c有发布权限
        result = self.client.MCPReleaseAction(TestMCPPerm.mcp_id, data_pub, TestMCPPerm.c_headers)
        assert result[0] == 200
        # 用户b无下架权限
        result = self.client.MCPReleaseAction(TestMCPPerm.mcp_id, data_off, TestMCPPerm.b_headers)
        assert result[0] == 403
        # 用户c有下架权限
        result = self.client.MCPReleaseAction(TestMCPPerm.mcp_id, data_off, TestMCPPerm.c_headers)
        assert result[0] == 200
        # 用户c发布mcp
        result = self.client.MCPReleaseAction(TestMCPPerm.mcp_id, data_pub, TestMCPPerm.c_headers)
        assert result[0] == 200

    @allure.title("调试MCP，无使用权限调试失败，有调试权限调试成功")
    def test_mcp_permission_08(self):
        # 用户b无使用权限
        result = self.client.MCPToolDebug(TestMCPPerm.mcp_id, "map_geocode", {"parameters": {"address": "上海东方明珠"}}, TestMCPPerm.b_headers)
        assert result[0] == 403
        # 用户c有使用权限
        result = self.client.MCPToolDebug(TestMCPPerm.mcp_id, "map_geocode", {"parameters": {"address": "上海东方明珠"}}, TestMCPPerm.c_headers)
        assert result[0] == 200

    @allure.title("获取MCP服务市场列表，无法获取到无公开访问权限的MCP，可获取到有公开访问权限的MCP")
    def test_mcp_permission_09(self):
        # 用户b获取市场列表
        result = self.client.GetMCPMarketList({"all": True}, TestMCPPerm.b_headers)
        assert TestMCPPerm.mcp_id not in [m["mcp_id"] for m in result[1]["data"]]
        # 用户c获取市场列表
        result = self.client.GetMCPMarketList({"all": True}, TestMCPPerm.c_headers)
        assert TestMCPPerm.mcp_id in [m["mcp_id"] for m in result[1]["data"]]

    @allure.title("获取MCP服务市场详情，无公开访问权限获取失败，有公开访问权限获取成功")
    def test_mcp_permission_10(self):
        # 用户b获取公共MCP详情失败
        result = self.client.GetMCPMarketDetail(TestMCPPerm.mcp_id, TestMCPPerm.b_headers)
        assert result[0] == 403
        # 用户c获取公共MCP详情成功
        result = self.client.GetMCPMarketDetail(TestMCPPerm.mcp_id, TestMCPPerm.c_headers)
        assert result[0] == 200

    @allure.title("批量获取MCP服务市场详情，无公开访问权限获取失败，有公开访问权限获取成功")
    def test_mcp_permission_11(self):        
        # 用户b批量获取失败
        result = self.client.BatchGetMCPMarketDetail(TestMCPPerm.mcp_id, "name", TestMCPPerm.b_headers)
        assert result[0] == 200
        assert result[1] == []
        # 用户c批量获取成功
        result = self.client.BatchGetMCPMarketDetail(TestMCPPerm.mcp_id, "name", TestMCPPerm.c_headers)
        assert result[0] == 200
        assert result[1][0]["mcp_id"] == TestMCPPerm.mcp_id

    @allure.title("获取指定MCP下的工具列表，无查看和公开访问权限获取失败，有查看或公开访问权限获取成功")
    def test_mcp_permission_12(self):
        # 用户d无查看和公开访问权限
        result = self.client.GetMCPToolList(TestMCPPerm.mcp_id, TestMCPPerm.d_headers)
        assert result[0] == 403
        # 用户b有查看权限
        result = self.client.GetMCPToolList(TestMCPPerm.mcp_id, TestMCPPerm.b_headers)
        assert result[0] == 200
        # 用户c有公开访问权限
        result = self.client.GetMCPToolList(TestMCPPerm.mcp_id, TestMCPPerm.c_headers)
        assert result[0] == 200

    @allure.title("调用MCP工具，无使用权限调用失败，有使用权限调用成功")
    def test_mcp_permission_13(self):
        data = {"tool_name": "map_geocode", "parameters": {"address": "上海东方明珠"}}
        # 用户b无使用权限
        result = self.client.CallMCPtool(TestMCPPerm.mcp_id, data, TestMCPPerm.b_headers)
        assert result[0] == 403
        # 用户c有使用权限
        result = self.client.CallMCPtool(TestMCPPerm.mcp_id, data, TestMCPPerm.c_headers)
        assert result[0] == 200

    @allure.title("删除MCP，无删除权限删除失败，有删除权限删除成功")
    def test_mcp_permission_14(self):
        # 已发布mcp不允许删除
        # 用户c有下架权限
        data_off = {"status": "offline"}
        result = self.client.MCPReleaseAction(TestMCPPerm.mcp_id, data_off, TestMCPPerm.c_headers)
        assert result[0] == 200
        # 用户c无权限，无法删除
        result = self.client.DeleteMCP(TestMCPPerm.mcp_id, TestMCPPerm.c_headers)
        assert result[0] == 403
        # 用户b有权限，可以删除
        result = self.client.DeleteMCP(TestMCPPerm.mcp_id, TestMCPPerm.b_headers)
        assert result[0] == 200
