# -*- coding:UTF-8 -*-
import allure
import requests

from common.get_content import GetContent
from common.request import Request

class MCP():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"] + "/api/agent-operator-integration/v1/mcp"

    '''解析SSE MCPServer'''
    def ParseSSE(self, data, headers):
        url = f"{self.base_url}/parse/sse"
        return Request.post(self, url, data, headers)

    '''添加MCP Server配置'''
    def RegisterMCP(self, data, headers):
        url = f"{self.base_url}"
        allure.attach(url, name="Request URL")
        allure.attach(str(data), name="Request Body")
        resp = requests.post(url, json=data, headers=headers, verify=False)

        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")
        # print(resp.status_code, resp.text)

        if resp.text == "":
            return [resp.status_code, resp.text]
        else:
            return [resp.status_code, resp.json()]

    '''删除MCP Server配置'''
    def DeleteMCP(self, mcp_id, headers):
        url = f"{self.base_url}/{mcp_id}"
        return Request.delete(self, url, None, headers)

    '''获取MCP Server列表'''
    def GetMCPList(self, params, headers):
        url = f"{self.base_url}/list"
        return Request.query(self, url, params, headers)

    '''获取MCP Server详情'''
    def GetMCPDetail(self, mcp_id, headers):
        url = f"{self.base_url}/{mcp_id}"
        return Request.get(self, url, headers)

    '''编辑MCP Server配置'''
    def EditMCP(self, mcp_id, data, headers):
        url = f"{self.base_url}/{mcp_id}"
        return Request.put(self, url, data, headers)

    '''MCP服务发布操作'''
    def MCPReleaseAction(self, mcp_id, data, headers):
        url = f"{self.base_url}/{ mcp_id}/status"
        return Request.post(self, url, data, headers)

    '''MCP工具调试'''
    def MCPToolDebug(self, mcp_id, name, data, headers):
        url = f"{self.base_url}/{mcp_id}/tool/{name}/debug"
        return Request.post(self, url, data, headers)

    '''获取已发布的MCP列表'''
    def GetMCPMarketList(self, params, headers):
        url = f"{self.base_url}/market/list"
        return Request.query(self, url, params, headers)

    '''获取已发布的MCP服务市场详情'''
    def GetMCPMarketDetail(self, mcp_id, headers):
        url = f"{self.base_url}/market/{mcp_id}"
        return Request.get(self, url, headers)

    '''获取指定MCP服务下的工具列表'''
    def GetMCPToolList(self, mcp_id, headers):
        url = f"{self.base_url}/proxy/{mcp_id}/tools"
        return Request.get(self, url, headers)

    '''调用指定MCP服务下的工具'''
    def CallMCPtool(self, mcp_id, data, headers):
        url = f"{self.base_url}/proxy/{mcp_id}/tool/call"
        return Request.post(self, url, data, headers)

    '''批量获取已发布的MCP服务市场详情'''
    def BatchGetMCPMarketDetail(self, mcp_ids, fields, headers):
        url = f"{self.base_url}/market/batch/{mcp_ids}/{fields}"
        return Request.get(self, url, headers)