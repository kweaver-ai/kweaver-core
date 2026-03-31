# -*- coding:UTF-8 -*-
from common.request import Request

class InternalOperator():
    def __init__(self):
        self.base_url = "http://agent-operator-integration:9000/api/agent-operator-integration/internal-v1/operator"

    '''获取算子分类'''
    def GetCategory(self, headers):
        url = f"{self.base_url}/category"
        return Request.get(self, url, headers)
    
    '''新建算子分类'''
    def CreateCategory(self, data, headers):
        url = f"{self.base_url}/category"
        return Request.post(self, url, data, headers)
    
    '''更新算子分类'''
    def UpdateCategory(self, category_type, data, headers):
        url = f"{self.base_url}/category/{category_type}"
        return Request.put(self, url, data, headers)
    
    '''删除算子分类'''
    def DeleteCategory(self, category_type, headers):
        url = f"{self.base_url}/category/{category_type}"
        return Request.pathdelete(self, url, headers)

    '''代理执行算子'''
    def ProxyOperator(self, operator_id, data, headers):
        url = f"{self.base_url}/proxy/{operator_id}"
        return Request.post(self, url, data, headers)