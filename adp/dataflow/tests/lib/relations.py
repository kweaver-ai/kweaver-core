# -*- coding:UTF-8 -*-

from common.get_content import GetContent
from common.request import Request

class Relations():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"] + "/api/agent-operator-integration/internal-v1/relations"

    '''删除关联关系'''
    def DeleteRelation(self, id, headers):
        url = f"{self.base_url}/{id}"
        return Request.delete(self, url, headers)

    '''获取关联关系详情'''
    def GetRelation(self, id, headers):
        url = f"{self.base_url}/{id}"
        return Request.get(self, url, headers)

    '''批量删除关联关系'''
    def BatchDeleteRelations(self, data, headers):
        url = f"{self.base_url}/delete"
        return Request.post(self, url, data, headers)

    '''根据源资源查询关联关系'''
    def GetRelationsBySource(self, headers, params):
        url = f"{self.base_url}/source"
        return Request.get(self, url, headers, params)

    '''根据目标资源查询关联关系'''
    def GetRelationsByTarget(self, headers, params):
        url = f"{self.base_url}/target"
        return Request.get(self, url, headers, params) 