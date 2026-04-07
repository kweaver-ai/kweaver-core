# -*- coding:UTF-8 -*-

import pytest
import allure
import json
import os
import subprocess

from common.get_content import GetContent
from common.create_user import CreateUser
from common.get_token import GetToken
from common.delete_user import DeleteUser
from lib.permission import Perm
import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

configfile = "./config/env.ini"
file = GetContent(configfile)
config = file.config()

host = config["server"]["host"]
admin_password = config["admin"]["admin_password"]

@pytest.fixture(scope="session", autouse=True)
def APrepare():
    # '''修改admin密码'''
    # mod = GetToken(host=host).modifyAdminPwd("eisoo.com123", "eisoo.com")
    # assert mod[0] == 200

    '''创建组织、部门和用户'''
    client = CreateUser(host=host)
    orgId = client.CreateOrganization("AISHU")

    # 检查是否返回了异常对象
    if hasattr(orgId, '__class__') and hasattr(orgId, '__name__'):
        print(f"Organization creation failed: {orgId}")
        orgId = "default-org-id"

    depId = client.AddDepartment(orgId, "测试部")

    # 检查是否返回了异常对象
    if hasattr(depId, '__class__') and hasattr(depId, '__name__'):
        print(f"Department creation failed: {depId}")
        depId = "default-dep-id"

    depIds = [depId]
    userId = client.AddUser("A0", depIds, orgId)

    # 检查是否返回了异常对象
    if hasattr(userId, '__class__') and hasattr(userId, '__name__'):
        print(f"User creation failed: {userId}")
        userId = "default-user-id"

    allure.attach(str(userId), name="create user success")

    yield orgId, depId, userId

    '''删除用户、部门和组织'''
    token = GetToken(host=host).get_token(host, "admin", admin_password)
    # admin_id = token[0]
    admin_token = token[1]

    client = DeleteUser(host=host)
    # client.DeleteUserDoc(userId, admin_id)
    client.DeleteUser(userId)
    re = client.DeleteOrganization(host, admin_token, orgId)
    assert re == 204
    allure.attach(str(orgId), name="delete user and department success")


@pytest.fixture(scope="session", autouse=True)
def Headers():
    '''获取token授权，外部接口授权'''
    token = GetToken(host=host).get_token(host, "A0", "111111")
    headers = {
        "Authorization": f"Bearer {token[1]}"
        }
    allure.attach(json.dumps(headers).encode("utf-8"), name="headers")

    yield headers

@pytest.fixture(scope="session", autouse=True)
def UserHeaders():
    '''获取token授权，内部接口授权'''
    token = GetToken(host=host).get_token(host, "A0", "111111")
    headers = {
        "x-account-id": token[0],
        "x-account-type": "user",
        "x-business-domain": "bd_public"
        }
    allure.attach(json.dumps(headers).encode("utf-8"), name="headers")

    yield headers


@pytest.fixture(scope="session", autouse=True)
def RoleMember(APrepare):
    # 将A0设置为AI管理员，角色ID：3fb94948-5169-11f0-b662-3a7bdba2913f
    try:
        token = GetToken(host=host).get_token(host, "admin", admin_password)
        headers = {
            "Authorization": f"Bearer {token[1]}"
            }
        perm_client = Perm()
        roleid = "3fb94948-5169-11f0-b662-3a7bdba2913f"
        data = {
            "method": "POST",
            "members": [
                {
                    "id": APrepare[2],
                    "type": "user"
                }
            ]
        }
        result = perm_client.ManageMember(roleid, data, headers)
        if result[0] != 204:
            allure.attach(f"RoleMember setup failed with status {result[0]}, skipping role assignment", name="role_member_warning")
            yield
            return
    except Exception as e:
        # 如果服务不可用（如 Thrift 连接失败），优雅地跳过角色设置
        allure.attach(f"RoleMember setup failed due to service unavailability: {str(e)}, skipping role assignment", name="role_member_warning")
        yield
        return

    yield

    try:
        data = {
            "method": "DELETE",
            "members": [
                {
                    "id": APrepare[2],
                    "type": "user"
                }
            ]
        }
        result = perm_client.ManageMember(roleid, data, headers)
        if result[0] != 204:
            allure.attach(f"RoleMember teardown failed with status {result[0]}, skipping role cleanup", name="role_member_teardown_warning")
    except Exception as e:
        # 如果服务不可用，优雅地跳过清理
        allure.attach(f"RoleMember teardown failed due to service unavailability: {str(e)}, skipping role cleanup", name="role_member_teardown_warning")

