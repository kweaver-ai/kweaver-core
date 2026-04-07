# -*- coding:UTF-8 -*-

import requests
import json
import warnings
import os
import sys
warnings.filterwarnings("ignore")

from eisoo import tclients
from eisoo.tclients import TClient
from ShareMgnt.ttypes import *

from common.get_token import GetToken
from common.get_content import GetContent

class CreateUser:
    def __init__(self, host):
        self.sharemgnt_ip = "sharemgnt.anyshare.svc.cluster.local"
        self.host = host

        # Read admin password from config file
        configfile = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "config/env.ini")
        file = GetContent(configfile)
        config = file.config()
        admin_password = config["admin"]["admin_password"]

        client = GetToken(self.host)
        result = client.get_token(self.host, "admin", admin_password)
        admin_token = result[1]

        self.headers = {
            'Authorization': 'Bearer %s' % admin_token
        }

    def DoclibUser(self, user_id, quota_allocated, storage_location):
        req_url = 'https://%s/api/efast/v1/doc-lib/user' % self.host
        uploadjson = {
            'user_id': user_id,
            'quota_allocated': quota_allocated,
            'storage_location': storage_location
        }
        try:
            r = requests.post(req_url, json=uploadjson, headers=self.headers, verify=False, allow_redirects=False)
            allure.attach(r.url, 'url', allure.attachment_type.TEXT)
            allure.attach(str(r.status_code), 'status_code', allure.attachment_type.TEXT)
            allure.attach(str(r.content), 'content', allure.attachment_type.TEXT)
            return r.status_code, r.headers, json.loads(r.content)
        except Exception as e:
            return 500, {}, {'error': str(e)}

    def DoclibDepartment(self, name, user_id, user_name, quota_allocated, storage_location, org_id):
        req_url = 'https://%s/api/efast/v1/doc-lib/department' % self.host
        owners = [{"id": user_id, "name": user_name, "type": "user"}]
        perm = {"allow":["display","preview","cache","download","create","modify","delete"],"expires_at":"1970-01-01T00:00:00Z"}
        upload_json = {"name": name, "quota_allocated": 300 * 1024 * 1024 * 1024, "owned_by": owners,
                    "storage_location": {"type": "unspecified"}, "perm": perm, "department_id": org_id}
        r = requests.request('POST', req_url, json=upload_json, verify=False, headers=self.headers, allow_redirects=False)
        if r.status_code < 200 or r.status_code > 299:
            raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))
        return json.loads(r.content)

    def CreateOrganization(self, orgName):
        '''
        创建组织
        '''
        addorginfo = ncTAddOrgParam()
        addorginfo.orgName = orgName
        addorginfo.email =''
        addorginfo.ossId =''
        with tclients.TClient('ShareMgnt', self.sharemgnt_ip, timeout_s=1800) as client:
            try:
                org_id = client.Usrm_CreateOrganization(addorginfo)
                return org_id
            except Exception as e:
                print ("create organization faild", e)
                return e

    def AddDepartment(self, parentId, departName):
        '''
        新建部门
        '''
        adddepartinfo = ncTAddDepartParam()
        adddepartinfo.parentId = parentId
        adddepartinfo.ossId = ''
        adddepartinfo.priority = 999
        adddepartinfo.departName = departName
        adddepartinfo.email =''
        with tclients.TClient('ShareMgnt', self.sharemgnt_ip, timeout_s=1800) as client:
            try:
                dep_id = client.Usrm_AddDepartment(adddepartinfo)
                return dep_id
            except Exception as e:
                print ("create department faild", e)
                return e

    def AddUser(self, loginName, departmentIds, org_id):
        '''
        新建用户
        '''
        with tclients.TClient('ShareMgnt', self.sharemgnt_ip, timeout_s=1800) as client:
            userInfo = ncTUsrmUserInfo(loginName=loginName,
                                                userType=1,
                                                departmentIds=departmentIds,
                                                space=50 * 1024 * 1024 * 1024,
                                                pwdControl=False)
            addUserInfo = ncTUsrmAddUserInfo(user=userInfo, password='111111')
            storage_location = {'type': 'unspecified'}
            # import pdb;pdb.set_trace()
            try:
                user_id = client.Usrm_AddUser(addUserInfo, '266c6a42-6131-4d62-8f39-853e7093701c')
                self.DoclibUser(user_id, 50 * 1024 * 1024 * 1024, storage_location)
                # if loginName == "A0":
                #     self.DoclibDepartment("AT-Test", user_id, loginName, 3000*1024*1024*1024, storage_location, org_id)
                return user_id
            except Exception as e:
                print ("create user faild", e)
                return e

    # def AddCustomDocLib(self, admin_token, userid, name):
    #     '''
    #     创建自定义文档库
    #     '''
    #     req_url = 'https://%s/api/efast/v1/doc-lib/custom' % self.host
    #     owners = [{"id": userid, "type": "user"}]
    #     upload_json = {"name": name, "quota_allocated": 200*1024*1024*1024, "owned_by": owners, "storage_location": {"type":"unspecified"}}
    #     r = requests.request('POST', req_url, json=upload_json, verify=False, headers=self.headers, allow_redirects=False)
    #     if r.status_code < 200 or r.status_code > 299:
    #         raise Exception("status code: %s\nbody: %s" % (r.status_code, r.content))
    #     return json.loads(r.content)

if __name__ == '__main__':
    client = CreateUser("192.168.232.15")
    orgId = client.CreateOrganization("AISHU")
    depId = client.AddDepartment(orgId, "测试部")
    depIds = [depId]
    userId = client.AddUser("A0", depIds, orgId)
    print(userId)
    # userid: 835418d8-1779-11f0-ad6b-1e06663d5d82
