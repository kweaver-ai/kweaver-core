# -*- coding:UTF-8 -*-

from common.get_content import GetContent
from common.request import Request

class Operator():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"] + "/api/agent-operator-integration/v1/operator"

    '''注册算子'''
    def RegisterOperator(self, data, headers):
        url = self.base_url + "/register"

        return Request.post(self, url, data, headers)

    '''获取算子列表'''
    def GetOperatorList(self, params, headers):
        url = self.base_url + "/info/list"
        return Request.query(self, url, params, headers)

    '''获取算子信息'''
    def GetOperatorInfo(self, operator_id, headers):
        url = self.base_url + "/info/" + operator_id
        return Request.get(self, url, headers)

    '''编辑算子'''
    def EditOperator(self, data, headers):
        url = self.base_url + "/info"

        return Request.post(self, url, data, headers)

    '''获取算子分类'''
    def GetOperatorCategory(self, headers):
        url = self.base_url + "/category"

        return Request.get(self, url, headers)

    '''删除算子'''
    def DeleteOperator(self, data, headers):
        url = self.base_url + "/delete"

        return Request.delete(self, url, data, headers)

    '''更新算子状态'''
    def UpdateOperatorStatus(self, data, headers):
        url = self.base_url + "/status"

        return Request.post(self, url, data, headers)
    
    '''更新算子信息'''
    def UpdateOperatorInfo(self, data, headers):
        url = self.base_url + "/info/update"

        return Request.post(self, url, data, headers)

    '''算子调试'''
    def OperatorDebug(self, data, headers):
        url = self.base_url + "/debug"

        return Request.post(self, url, data, headers)

    '''获取算子历史版本详情'''
    def GetOperatorHistoryDetail(self, operator_id, version, headers, tag=None):
        url = self.base_url + f"/history/{operator_id}/{version}"
        params = {}
        if tag is not None:
            params["tag"] = tag
        return Request.query(self, url, params, headers)

    '''获取算子历史版本列表'''
    def GetOperatorHistoryList(self, operator_id, headers):
        url = self.base_url + f"/history/{operator_id}"
        return Request.get(self, url, headers)

    '''获取算子市场列表'''
    def GetOperatorMarketList(self, params, headers):
        url = self.base_url + "/market"
        return Request.query(self, url, params, headers)

    '''获取算子市场指定算子详情'''
    def GetOperatorMarketDetail(self, operator_id, headers):
        url = self.base_url + f"/market/{operator_id}"
        return Request.get(self, url, headers)
    
    '''注册或更新内置算子'''
    def RegisterBuiltinOperator(self, data, headers):
        url = self.base_url + "/intcomp"
        return Request.post(self, url, data, headers)
    
    '''注册或更新内置算子(内部接口)'''
    def InternalBuiltinOperator(self, data, headers):
        url = "http://agent-operator-integration:9000/api/agent-operator-integration/internal-v1/operator/intcomp"
        return Request.post(self, url, data, headers)
