# -*- coding:UTF-8 -*-

from common.get_content import GetContent
from common.request import Request

class ToolBox():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"] + "/api/agent-operator-integration/v1/tool-box"

    '''创建工具箱'''
    def CreateToolbox(self, data, headers):
        url = self.base_url
        return Request.post(self, url, data, headers)

    '''更新工具箱'''
    def UpdateToolbox(self, box_id, data, headers):
        url = f"{self.base_url}/{box_id}"
        return Request.post(self, url, data, headers)

    '''获取工具箱信息'''
    def GetToolbox(self, box_id, headers):
        url = f"{self.base_url}/{box_id}"
        return Request.get(self, url, headers)

    '''删除工具箱'''
    def DeleteToolbox(self, box_id, headers):
        url = f"{self.base_url}/{box_id}"
        return Request.pathdelete(self, url, headers)

    '''获取工具箱列表'''
    def GetToolboxList(self, params, headers):
        url = f"{self.base_url}/list"
        return Request.query(self, url, params, headers)

    '''更新工具箱状态'''
    def UpdateToolboxStatus(self, box_id, data, headers):
        url = f"{self.base_url}/{box_id}/status"
        return Request.post(self, url, data, headers)

    '''创建工具'''
    def CreateTool(self, box_id, data, headers):
        url = f"{self.base_url}/{box_id}/tool"
        return Request.post(self, url, data, headers)

    '''更新工具'''
    def UpdateTool(self, box_id, tool_id, data, headers):
        url = f"{self.base_url}/{box_id}/tool/{tool_id}"
        return Request.post(self, url, data, headers)

    '''获取工具信息'''
    def GetTool(self, box_id, tool_id, headers):
        url = f"{self.base_url}/{box_id}/tool/{tool_id}"
        return Request.get(self, url, headers)

    '''批量删除工具'''
    def BatchDeleteTools(self, box_id, data, headers):
        url = f"{self.base_url}/{box_id}/tools/batch-delete"
        return Request.post(self, url, data, headers)

    '''获取工具箱中的工具列表'''
    def GetBoxToolsList(self, box_id, params, headers):
        url = f"{self.base_url}/{box_id}/tools/list"
        return Request.query(self, url, params, headers)

    '''更新工具状态'''
    def UpdateToolStatus(self, box_id, data, headers):
        url = f"{self.base_url}/{box_id}/tools/status"
        return Request.post(self, url, data, headers)

    '''获取所有工具列表'''
    def GetMarketToolsList(self, params, headers):
        url = f"{self.base_url}/market/tools"
        return Request.query(self, url, params, headers)

    '''工具调试'''
    def DebugTool(self, box_id, tool_id, data, headers):
        url = f"{self.base_url}/{box_id}/tool/{tool_id}/debug"
        return Request.post(self, url, data, headers)

    '''工具执行代理接口'''
    def ProxyTool(self, box_id, tool_id, data, headers):
        url = f"{self.base_url}/{box_id}/proxy/{tool_id}"
        return Request.post(self, url, data, headers)

    '''算子转换成工具'''
    def ConvertOperatorToTool(self, data, headers):
        url = f"{self.base_url.replace('/tool-box', '/operator/convert/tool')}"
        return Request.post(self, url, data, headers)

    '''创建/更新内置工具'''
    def Builtin(self, data, headers):
        url = f"{self.base_url}/intcomp"
        return Request.post(self, url, data, headers)

    '''获取工具箱市场详情'''
    def GetMarketDetail(self, box_id, fields, headers):
        url = f"{self.base_url}/market/{box_id}/{fields}"
        return Request.get(self, url, headers)

    '''创建/更新内置工具（内部接口）'''
    def InternalBuiltin(self, data, headers):
        url = "http://agent-operator-integration:9000/api/agent-operator-integration/internal-v1/tool-box/intcomp"
        return Request.post(self, url, data, headers)

    '''获取市场工具箱信息'''
    def GetMarketToolbox(self, box_id, headers):
        url = f"{self.base_url}/market/{box_id}"
        return Request.get(self, url, headers)

    '''获取市场工具箱列表'''
    def GetMarketToolboxList(self, params, headers):
        url = f"{self.base_url}/market"
        return Request.query(self, url, params, headers)