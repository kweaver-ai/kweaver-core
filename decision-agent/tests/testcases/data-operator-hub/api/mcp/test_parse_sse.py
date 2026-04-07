# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random

from lib.mcp import MCP
from lib.tool_box import ToolBox
from common.get_content import GetContent
from common.get_token import GetToken

@allure.feature("MCP服务管理接口测试：解析SSE_MCPServer")
class TestParseSSE:
    
    client = MCP()
    tool_client = ToolBox()

    @allure.title("解析SSE_MCPServer，参数正确，解析成功")
    def test_parse_sse_01(self, Headers):
        data = {
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "headers": {
                "Content-Type": "application/json"
            }
        }
        result = self.client.ParseSSE(data, Headers)
        assert result[0] == 200
        assert "tools" in result[1]
        assert len(result[1]["tools"]) > 0

    @allure.title("解析SSE_MCPServer，缺少mode参数，解析失败")
    def test_parse_sse_02(self, Headers):
        data = {
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "headers": {
                "Content-Type": "application/json"
                }
        }
        result = self.client.ParseSSE(data, Headers)
        assert result[0] == 400

    @allure.title("解析SSE_MCPServer，mode参数不正确，解析失败")
    def test_parse_sse_03(self, Headers):
        data = {
            "mode": "invalid_mode",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "headers": {
                "Content-Type": "application/json"
            }
        }
        result = self.client.ParseSSE(data, Headers)
        assert result[0] == 400

    @allure.title("解析SSE_MCPServer，url格式不正确，解析失败")
    def test_parse_sse_04(self, Headers):
        data = {
            "mode": "sse",
            "url": "invalid_url",
            "headers": {
                "Content-Type": "application/json"
            }
        }
        result = self.client.ParseSSE(data, Headers)
        assert result[0] == 400

    @allure.title("解析SSE_MCPServer，url无法连接，解析失败")
    def test_parse_sse_05(self, Headers):
        data = {
            "mode": "sse",
            "url": "http://localhost:8080/api/v1/tools"
        }
        result = self.client.ParseSSE(data, Headers)
        assert result[0] == 500 

    @allure.title("解析SSE_MCPServer，缺少url参数，解析失败")
    def test_parse_sse_06(self, Headers):
        data = {
            "mode": "sse",
            "headers": {
                "Content-Type": "application/json"
                }
        }
        result = self.client.ParseSSE(data, Headers)
        assert result[0] == 400

    @allure.title("解析SSE_MCPServer，mode为stream，解析成功")
    def test_parse_sse_07(self, Headers):
        data = {
            "mode": "stream",
            "url": "https://mcp.map.baidu.com/mcp?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "headers": {
                "Content-Type": "application/json"
            }
        }
        result = self.client.ParseSSE(data, Headers)
        assert result[0] == 200
        assert "tools" in result[1]
        assert len(result[1]["tools"]) > 0

    @allure.title("工具转换成mcp，使用转换后的url进行解析，解析成功")
    def test_parse_sse_08(self, Headers):
        # 创建工具箱、启用工具并发布
        filepath = "./resource/openapi/compliant/test.json"
        json_data = GetContent(filepath).jsonfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        result = self.tool_client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]
        result = self.tool_client.GetBoxToolsList(box_id, None, Headers)
        tools = result[1]["tools"]
        update_data = []
        tools_id = []
        for tool in tools:
            data = {
                "tool_id": tool["tool_id"],
                "status": "enabled"
            }
            update_data.append(data) 
            tools_id.append(tool["tool_id"])         
        result = self.tool_client.UpdateToolStatus(box_id, update_data, Headers)
        assert result[0] == 200
        result = self.tool_client.UpdateToolboxStatus(box_id, {"status": "published"}, Headers)
        assert result[0] == 200

        tool_result = self.tool_client.GetTool(box_id, tools_id[0], Headers)
        assert tool_result[0] == 200
        tool_configs = [{
            "box_id": box_id,
            "box_name": name,
            "tool_id": tools_id[0],
            "tool_name": tool_result[1]["name"],
            "description": tool_result[1]["description"],
            "use_rule": "all"
        }]
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "add mcp server config",
            "category": "data_analysis",
            "creation_type": "tool_imported",
            "tool_configs": tool_configs
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]
        # 发布mcp
        result = self.client.MCPReleaseAction(mcp_id, {"status": "published"}, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"
        # 获取mcp详情
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        sse_url = result[1]["connection_info"]["sse_url"]
        stream_url = result[1]["connection_info"]["stream_url"]
        # 获取token
        configfile = "./config/env.ini"
        file = GetContent(configfile)
        config = file.config()
        host = config["server"]["host"]
        token = GetToken(host=host).get_token(host, "A0", "111111")
        access_token = token[1]
        sse_url = "http://" + host + sse_url + "?token=" + access_token
        stream_url = "http://" + host + stream_url + "?token=" + access_token
        # 解析url
        datas = [{"mode": "sse", "url": sse_url,  "headers": {"Content-Type": "application/json"}},
                 {"mode": "stream", "url": stream_url, "headers": {"Content-Type": "application/json"}}]
        for data in datas:
            result = self.client.ParseSSE(data, Headers)
            assert result[0] == 200
            assert "tools" in result[1]
            assert len(result[1]["tools"]) > 0

    @allure.title("自定义sse mcp，使用代理url进行解析，解析成功")
    def test_parse_sse_09(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test sse mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]
        result = self.client.MCPReleaseAction(mcp_id, {"status": "published"}, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"
        # 获取mcp详情
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        sse_url = result[1]["connection_info"]["sse_url"]
        # 获取token
        configfile = "./config/env.ini"
        file = GetContent(configfile)
        config = file.config()
        host = config["server"]["host"]
        token = GetToken(host=host).get_token(host, "A0", "111111")
        access_token = token[1]
        sse_url = "http://" + host + sse_url + "?token=" + access_token
        # 解析url
        data = {
            "mode": "sse",
            "url": sse_url,
            "headers": {
                "Content-Type": "application/json"
            }
        }
        result = self.client.ParseSSE(data, Headers)
        assert result[0] == 200
        assert "tools" in result[1]
        assert len(result[1]["tools"]) > 0

    @allure.title("自定义stream mcp，使用代理url进行解析，解析成功")
    def test_parse_sse_10(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test stream mcp server",
            "mode": "stream",
            "url": "https://mcp.map.baidu.com/mcp?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]
        result = self.client.MCPReleaseAction(mcp_id, {"status": "published"}, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"
        # 获取mcp详情
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        stream_url = result[1]["connection_info"]["stream_url"]
        # 获取token
        configfile = "./config/env.ini"
        file = GetContent(configfile)
        config = file.config()
        host = config["server"]["host"]
        token = GetToken(host=host).get_token(host, "A0", "111111")
        access_token = token[1]
        stream_url = "http://" + host + stream_url + "?token=" + access_token
        # 解析url
        data = {
            "mode": "stream",
            "url": stream_url,
            "headers": {
                "Content-Type": "application/json"
            }
        }
        result = self.client.ParseSSE(data, Headers)
        assert result[0] == 200
        assert "tools" in result[1]
        assert len(result[1]["tools"]) > 0