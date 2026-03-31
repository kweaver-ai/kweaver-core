# -*- coding:UTF-8 -*-

import requests
import json
import warnings
import os
warnings.filterwarnings("ignore")

from eisoo import tclients
from eisoo.tclients import TClient
from ShareMgnt.ttypes import *

from common.get_token import GetToken
from common.get_content import GetContent

class DeleteUser:
    def __init__(self, host):
        self.sharemgnt_ip = "sharemgnt.anyshare.svc.cluster.local"
        self.efast_ip = "efast-efast.anyshare.svc.cluster.local"
        self.host = host

    # def DeleteUserDoc(self, userId, deleterId):
    #     '''
    #     删除个人文档库
    #     '''
    #     with tclients.TClient('EFAST', self.efast_ip, timeout_s=1800) as client:
    #         result = ''
    #         try:
    #             result = client.EFAST_DeleteUserDoc(userId, deleterId)
    #             return ['success', result]
    #         except Exception as e:
    #             result = ['get an exception', e]
    #             return result
            
    def DeleteUser(self, userId):
        '''
        删除用户
        '''
        with tclients.TClient('ShareMgnt', self.sharemgnt_ip, timeout_s=1800) as client:
            try:
                client.Usrm_DelUser(userId)
                return 'delete user success'
            except Exception as e:
                print(f"delete user failed: {e}")
                return  e
            
    def DeleteOrganization(self, ip, token, orgId):
        '''
        删除组织，默认删除该组织下的所有部门
        '''
        requrl = 'https://%s/api/user-management/v1/management/departments/%s' % (ip, orgId)
        dict1 = {}
        dict1['Authorization'] = "Bearer %s" % token
        r = requests.request('DELETE', requrl, verify=False, headers=dict1)
        if r.status_code == 204:
            return r.status_code
        else:
            return r.status_code, json.loads(r.content)
        
if __name__ == '__main__':
    # Read admin password from config file
    configfile = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "config/env.ini")
    file = GetContent(configfile)
    config = file.config()
    admin_password = config["admin"]["admin_password"]

    token = GetToken("192.168.232.15").get_token("192.168.232.15", "admin", admin_password)
    admin_id = token[0]
    admin_token = token[1] 

    client = DeleteUser("192.168.232.15")
    client.DeleteUser(["628cf266-1c40-11f0-a823-be0b43c166e7"])
    re = client.DeleteOrganization("192.168.232.15", admin_token, "e3364cba-1e51-11f0-987c-be0b43c166e7")
    assert re == 204
       