# -*- coding:UTF-8 -*-
from common.request import Request

class InternalMCP():
    def __init__(self):
        self.base_url = "http://agent-operator-integration:9000/api/agent-operator-integration/internal-v1/mcp/intcomp"

    '''注册内置MCP服务'''
    def Register(self, data, headers):
        url = f"{self.base_url}/register"
        return Request.post(self, url, data, headers)
    
    '''注销内置MCP服务'''
    def UnRegister(self, mcp_id, data, headers):
        url = f"{self.base_url}/unregister/{mcp_id}"
        return Request.post(self, url, data, headers)