# -*- coding:UTF-8 -*-
import allure
import requests

from common.get_content import GetContent
from common.request import Request

class Perm():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"] + "/api/authorization/v1"

    '''创建角色'''
    def CreateRole(self, data, headers):
        url = f"{self.base_url}/roles"
        return Request.post(self, url, data, headers)
    
    '''删除角色'''
    def DeleteRole(self, headers):
        url = f"{self.base_url}/roles"
        return Request.pathdelete(self, url, headers)
    
    '''添加/删除角色成员'''
    def ManageMember(self, roleid, data, headers):
        url = f"{self.base_url}/role-members/{roleid}"
        return Request.post(self, url, data, headers)
    
    '''设置权限'''
    def SetPerm(self, data, headers):
        url = f"{self.base_url}/policy"
        return Request.post(self, url, data, headers)