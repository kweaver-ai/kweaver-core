'''
创建业务域A和业务域B
用户a1属于业务域A；用户b1属于业务域B
用户a1和b1拥有所有算子/工具/mcp的新建、查看、公开访问和删除权限
AI管理员在公共域创建2个算子/工具箱/mcp。以及1个内置算子/工具箱/mcp
用户a1和b1分别创建1个算子/工具箱/mcp，以及1个内置算子/工具箱/mcp
用户a1和b1不能看到（管理及市场页面）或删除对方的算子/工具箱/mcp
用户a1和b1都能看到（管理及市场页面）和删除公共域的算子/工具箱/mcp
【备注】当前可在其他域创建、获取、删除、注册内置资源，暂时屏蔽掉相关用例，待问题修复后再验证（调用内部接口关联业务域资源时，若透传user会导致无权限，不传则未校验权限）
'''
# -*- coding:UTF-8 -*-

from lib import mcp_internal
import pytest
import allure
import requests
import uuid
import string
import random

from common.get_content import GetContent
from common import operator_db
from common.create_user import CreateUser
from common.get_token import GetToken
from common.delete_user import DeleteUser
from lib.mcp import MCP
from lib.mcp_internal import InternalMCP
from lib.operator import Operator
from lib.tool_box import ToolBox
from lib.permission import Perm

configfile = "./config/env.ini"
file = GetContent(configfile)
config = file.config()

host = config["server"]["host"]
db_port = config["server"]["db_port"]
db_user = config["server"]["db_user"]
db_pwd = config["server"]["db_pwd"]
admin_password = config["admin"]["admin_password"]

mcp_client = MCP()
mcp_internal = InternalMCP()
op_client = Operator()
toolbox_client = ToolBox()
perm_client = Perm()

@pytest.fixture(scope="session", autouse=True)
def DeleteDatabaseData():
    '''清理数据库记录'''
    operator_db.delete_operator_data()

@pytest.fixture(scope="session", autouse=True)
def DomainPrepare():
    token = GetToken(host=host).get_token(host, "admin", admin_password)
    admin_token = token[1]
    headers = {
        "Authorization": f"Bearer {admin_token}"
    }
    domain_url = f"https://{host}/api/business-system/v1/business-domain"
    # 创建组织、用户和业务域
    name = ["a1", "b1"]
    client = CreateUser(host=host)
    orgId = client.CreateOrganization("domain")
    user_list = []
    for i in name:
        depIds = [orgId]
        userId = client.AddUser(i, depIds, orgId)
        user_list.append(userId)
    allure.attach(orgId, name="create user and department success")
    domain_data_A = {
        "name": "业务域A",
        "products": ["dip"],
        "members": [{
            "id": user_list[0],
            "type": "user",
            "role": "developer"
        }]
    }
    domain_data_B = {
        "name": "业务域B",
        "products": ["dip"],
        "members": [{
            "id": user_list[1],
            "type": "user",
            "role": "developer"
        }]
    }
    domain_list = []
    for data in [domain_data_A, domain_data_B]:
        result = requests.post(url=domain_url, json=data, verify=False, headers=headers)
        assert result.status_code == 201
        domain_info = result.json()
        domain_list.append(domain_info["id"])
    allure.attach(domain_list, name="create domains success")

    yield user_list, domain_list

    # 删除用户、组织、业务域
    client = DeleteUser(host=host)
    for userid in user_list:
        client.DeleteUser(userid)
    re = client.DeleteOrganization(host, admin_token, orgId)
    assert re == 204
    allure.attach(orgId, name="delete user and department success")
    for domain_id in [domain_list[0], domain_list[1]]:
        resource_url = f"https://{host}/api/business-system/v1/resource"
        # 获取业务域中的资源
        get_res_url = f"{resource_url}?bd_id={domain_id}&limit=-1"
        result = requests.get(url=get_res_url, verify=False, headers=headers)
        assert result.status_code == 200
        resources = result.json()
        resource_info = resources["items"]
        # 取消资源关联
        for resource in resource_info:
            resource_id = resource["id"]
            resource_type = resource["type"]
            del_res_url = f"{resource_url}?bd_id={domain_id}&id={resource_id}&type={resource_type}"
            result = requests.delete(url=del_res_url, verify=False, headers=headers)
            assert result.status_code == 200
        # 删除业务域
        del_url = f"{domain_url}/{domain_id}"    
        result = requests.delete(url=del_url, verify=False, headers=headers)
        assert result.status_code == 200
    allure.attach(orgId, name="delete domains success")

@pytest.fixture(scope="session", autouse=True)
def SetPerm(Headers, DomainPrepare):
    # 为用户配置资源权限
    user_list = DomainPrepare[0]
    data = [
        {
            "accessor": {"id": user_list[0], "name": "a1", "type": "user"},
            "resource": {"id": "*", "type": "mcp", "name": "mcp新建、查看、编辑、下架、公开访问、删除权限"},
            "operation": {"allow": [{"id": "create"}, {"id": "view"}, {"id": "modify"}, {"id": "unpublish"}, {"id": "public_access"}, { "id": "delete"}], "deny": []}
        },
        {
            "accessor": {"id": user_list[1], "name": "b1", "type": "user"},
            "resource": {"id": "*", "type": "mcp", "name": "mcp新建、查看、编辑、下架、公开访问、删除权限"},
            "operation": {"allow": [{"id": "create"}, {"id": "view"}, {"id": "modify"}, {"id": "unpublish"}, {"id": "public_access"}, { "id": "delete"}], "deny": []}
        },
        {
            "accessor": {"id": user_list[0], "name": "a1", "type": "user"},
            "resource": {"id": "*", "type": "operator", "name": "算子新建、查看、编辑、下架、公开访问、删除权限"},
            "operation": {"allow": [{"id": "create"}, {"id": "view"}, {"id": "modify"}, {"id": "unpublish"}, {"id": "public_access"}, { "id": "delete"}], "deny": []}
        },
        {
            "accessor": {"id": user_list[1], "name": "b1", "type": "user"},
            "resource": {"id": "*", "type": "operator", "name": "算子新建、查看、编辑、下架、公开访问、删除权限"},
            "operation": {"allow": [{"id": "create"}, {"id": "view"}, {"id": "modify"}, {"id": "unpublish"}, {"id": "public_access"}, { "id": "delete"}], "deny": []}
        },
        {
            "accessor": {"id": user_list[0], "name": "a1", "type": "user"},
            "resource": {"id": "*", "type": "tool_box", "name": "工具箱新建、查看、编辑、下架、公开访问、删除权限"},
            "operation": {"allow": [{"id": "create"}, {"id": "view"}, {"id": "modify"}, {"id": "unpublish"}, {"id": "public_access"}, { "id": "delete"}], "deny": []}
        },
        {
            "accessor": {"id": user_list[1], "name": "b1", "type": "user"},
            "resource": {"id": "*", "type": "tool_box", "name": "工具箱新建、查看、编辑、下架、公开访问、删除权限"},
            "operation": {"allow": [{"id": "create"}, {"id": "view"}, {"id": "modify"}, {"id": "unpublish"}, {"id": "public_access"}, { "id": "delete"}], "deny": []}
        }
    ]
    result = perm_client.SetPerm(data, Headers)
    assert "20" in str(result[0])

@pytest.fixture(scope="session", autouse=True)
def TestDomainData(DomainPrepare):
    # 创建跨域的算子/工具/mcp资源
    domain_list = DomainPrepare[1]
    token_list = []
    for user in ["A0", "a1", "b1"]:
        token = GetToken(host=host).get_token(host, user, "111111")
        token_list.append(token[1])
    pub_domain_headers = {
        "Authorization": f"Bearer {token_list[0]}",
        "x-business-domain": "bd_public"
    }
    A_domain_headers = {
        "Authorization": f"Bearer {token_list[1]}",
        "x-business-domain": domain_list[0]
    }
    B_domain_headers = {
        "Authorization": f"Bearer {token_list[2]}",
        "x-business-domain": domain_list[1]
    }
    resource = "./resource/openapi/compliant/edit-test1.yaml"
    internal_resource = "./resource/openapi/compliant/edit-test2.yaml"
    api_data = GetContent(resource).yamlfile()
    internal_data =  GetContent(internal_resource).yamlfile()
    operator_list = []
    box_list = []
    mcp_list = []
    # AI管理员创建公共域资源，用户创建各自域内资源
    for headers in [pub_domain_headers, A_domain_headers, B_domain_headers]:
        operator_id, box_id, mcp_id = _resource_create_and_pub(api_data, headers)
        internal_operator_id, internal_box_id, internal_mcp_id = _internal_resource_create_and_pub(internal_data, headers)
        operator_list.append(operator_id)
        operator_list.append(internal_operator_id)
        box_list.append(box_id)
        box_list.append(internal_box_id)
        mcp_list.append(mcp_id)
        mcp_list.append(internal_mcp_id)
    headers = [pub_domain_headers, A_domain_headers, B_domain_headers]
    return domain_list, headers, operator_list, box_list, mcp_list

def _resource_create_and_pub(data, headers):
    name = "test_" + ''.join(random.choice(string.ascii_letters) for i in range(8))
    for path in data["paths"]:
        for method in data["paths"][path]:
            if "summary" in data["paths"][path][method]:
                op_name = ''.join(random.choice(string.ascii_letters) for i in range(8))
                data["paths"][path][method]["summary"] = f"Test_operator_{op_name}"
    # 算子
    payload = {
        "data": str(data),
        "operator_metadata_type": "openapi",
        "direct_publish": True
    }
    re = op_client.RegisterOperator(payload, headers)
    assert re[0] == 200
    operator_id = re[1][0]["operator_id"]
    # 工具箱 
    payload = {
        "box_name": name,
        "data": data,
        "metadata_type": "openapi"
    }
    result = toolbox_client.CreateToolbox(payload, headers)
    assert result[0] == 200
    box_id = result[1]["box_id"]
    result = toolbox_client.UpdateToolboxStatus(box_id, {"status": "published"}, headers)
    assert result[0] == 200
    # MCP
    payload = {
        "name": name,
        "description": "test mcp server 1",
        "mode": "stream",
        "url": "https://mcp.map.baidu.com/mcp?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL"
    }
    result = mcp_client.RegisterMCP(payload, headers)
    assert result[0] == 200
    mcp_id = result[1]["mcp_id"]
    result = mcp_client.MCPReleaseAction(mcp_id, {"status": "published"}, headers)
    assert result[0] == 200
    return operator_id, box_id, mcp_id

def _internal_resource_create_and_pub(data, headers):
    name = ''.join(random.choice(string.ascii_letters) for i in range(8))
    # 算子
    operator_id = str(uuid.uuid4())
    payload = {
        "operator_id": operator_id,
        "name": name,
        "data": data,
        "metadata_type": "openapi",
        "operator_type": "basic",
        "execution_mode": "sync",
        "source": "intenal",
        "config_source": "auto",
        "config_version": "1.0.0"
    }

    result = op_client.RegisterBuiltinOperator(payload, headers)
    assert result[0] == 200
    # 工具箱
    box_id = str(uuid.uuid4())
    payload = {
        "box_id": box_id,
        "box_name": name,
        "box_desc": "test description",
        "data": data,
        "metadata_type": "openapi",
        "source": "internal",
        "config_version": "1.0.0",
        "config_source": "auto"
    }
    result = toolbox_client.Builtin(payload, headers)
    assert result[0] == 200
    # MCP
    mcp_id = str(uuid.uuid4())
    payload = {
        "mcp_id": mcp_id,
        "name": name,
        "description": "test builtin mcp server",
        "mode": "sse",
        "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
        "command": "ls",
        "args": ["a", "b", "c"],
        "headers": {
            "Content-Type": "application/json"
        },
        "env": {},
        "source": "intenal",
        "protected_flag": False,
        "config_version": "1.0.0",
        "config_source": "auto"
    }
    result = mcp_internal.Register(payload, headers)
    assert result[0] == 200
    return operator_id, box_id, mcp_id