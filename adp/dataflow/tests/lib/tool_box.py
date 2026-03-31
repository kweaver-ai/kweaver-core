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

    '''创建工具箱 (Multipart)'''
    def CreateToolboxMultipart(self, files, data, headers):
        url = self.base_url
        return Request.post_multipart(self, url, files, data, headers)

    '''更新工具箱'''
    def UpdateToolbox(self, box_id, data, headers):
        url = f"{self.base_url}/{box_id}"
        return Request.post(self, url, data, headers)

    '''更新工具箱 (Multipart)'''
    def UpdateToolboxMultipart(self, box_id, files, data, headers):
        url = f"{self.base_url}/{box_id}"
        return Request.post_multipart(self, url, files, data, headers)

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

    '''创建工具 (Multipart)'''
    def CreateToolMultipart(self, box_id, files, data, headers):
        url = f"{self.base_url}/{box_id}/tool"
        return Request.post_multipart(self, url, files, data, headers)

    '''更新工具'''
    def UpdateTool(self, box_id, tool_id, data, headers):
        url = f"{self.base_url}/{box_id}/tool/{tool_id}"
        return Request.post(self, url, data, headers)

    '''更新工具 (Multipart)'''
    def UpdateToolMultipart(self, box_id, tool_id, files, data, headers):
        url = f"{self.base_url}/{box_id}/tool/{tool_id}"
        return Request.post_multipart(self, url, files, data, headers)

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
    def DebugTool(self, box_id, tool_id, data, headers, params=None):
        url = f"{self.base_url}/{box_id}/tool/{tool_id}/debug"
        if params:
            return Request.query_post(self, url, params, data, headers)
        return Request.post(self, url, data, headers)

    '''工具执行代理接口'''
    def ProxyTool(self, box_id, tool_id, data, headers, params=None):
        url = f"{self.base_url}/{box_id}/proxy/{tool_id}"
        if params:
            return Request.query_post(self, url, params, data, headers)
        return Request.post(self, url, data, headers)

    '''算子转换成工具'''
    def ConvertOperatorToTool(self, data, headers):
        url = f"{self.base_url.replace('/tool-box', '/operator/convert/tool')}"
        return Request.post(self, url, data, headers)

    '''创建/更新内置工具'''
    def Builtin(self, data, headers):
        url = f"{self.base_url}/intcomp"
        return Request.post(self, url, data, headers)

    '''创建/更新内置工具 (Multipart)'''
    def BuiltinMultipart(self, files, data, headers):
        url = f"{self.base_url}/intcomp"
        return Request.post_multipart(self, url, files, data, headers)

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

    '''获取代码模板'''
    def GetTemplate(self, template_type, headers):
        """
        获取代码模板
        根据最新API文档：/v1/template/{template_type}
        :param template_type: 模板类型，如 "python"
        :param headers: 请求头
        :return: (status_code, response_data) 响应包含 template_type 和 code_template
        """
        url = f"{self.base_url.replace('/tool-box', '/template')}"
        if template_type:
            url = f"{url}/{template_type}"
        return Request.get(self, url, headers)

    '''执行函数'''
    def ExecuteFunction(self, data, headers):
        """
        执行函数块
        根据最新API文档：/v1/function/execute
        :param data: 请求数据，包含 code (string) 和 event (object)
        :param headers: 请求头
        :return: (status_code, response_data) 响应包含 stdout, stderr, result, metrics
        """
        url = f"{self.base_url.replace('/tool-box', '/function/execute')}"
        return Request.post(self, url, data, headers)