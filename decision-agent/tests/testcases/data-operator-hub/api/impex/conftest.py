# -*- coding:UTF-8 -*-

import pytest
import allure
import pymysql
import os
import string
import random
import uuid
import subprocess

from lib.mcp import MCP
from lib.mcp_internal import InternalMCP
from lib.operator import Operator
from lib.tool_box import ToolBox
from lib.permission import Perm

from common.get_content import GetContent
from common.create_user import CreateUser
from common.get_token import GetToken
from common.delete_user import DeleteUser

configfile = "./config/env.ini"
file = GetContent(configfile)
config = file.config()

host = config["server"]["host"]
db_port = config["server"]["db_port"]
db_user = config["server"]["db_user"]
db_pwd = config["server"]["db_pwd"]
admin_password = config["admin"]["admin_password"]

@pytest.fixture(scope="session", autouse=True)
def DeleteOperatorData():
    '''清理数据库记录'''
    conn = pymysql.connect(host=host, user=db_user, password=db_pwd, port=int(db_port), database="dip_data_operator_hub")
    cursor = conn.cursor()
    try:
        cursor.execute("DELETE FROM t_op_registry")
        conn.commit()
        print(f"表 t_op_registry 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_metadata_api")
        conn.commit()
        print(f"表 t_metadata_api 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_tool")
        conn.commit()
        print(f"表 t_tool 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_toolbox")
        conn.commit()
        print(f"表 t_toolbox 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_internal_component_config")
        conn.commit()
        print(f"表 t_internal_component_config 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_mcp_server_config")
        conn.commit()
        print(f"表 t_mcp_server_config 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_mcp_server_release")
        conn.commit()
        print(f"表 t_mcp_server_release 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_mcp_server_release_history")
        conn.commit()
        print(f"表 t_mcp_server_release_history 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_operator_release")
        conn.commit()
        print(f"表 t_operator_release 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_operator_release_history")
        conn.commit()
        print(f"表 t_operator_release_history 中 {cursor.rowcount} 条记录已被删除")
    except Exception as e:
        conn.rollback()
        print(f"error: {str(e)}")
    finally:
        cursor.close()
        conn.close()

@pytest.fixture(scope="session", autouse=True)
def DeletePolicyData():
    '''清理权限策略数据后，重启authorization服务重建内置策略'''
    conn = pymysql.connect(host=host, user=db_user, password=db_pwd, port=int(db_port), database="anyshare")
    cursor = conn.cursor()
    try:
        cursor.execute("DELETE FROM t_policy")
        conn.commit()
        print(f"表 t_policy 中 {cursor.rowcount} 条记录已被删除")
    except Exception as e:
        conn.rollback()
        print(f"error: {str(e)}")
    finally:
        cursor.close()
        conn.close()
    
    os.system("kubectl -n anyshare delete pod $(kubectl get pod -n anyshare | grep authorization | awk '{print $1}')")
    command = "kubectl get pod -n anyshare | grep authorization | awk '{print $2}'"
    for i in range(60):
        result = subprocess.run(command, shell=True, capture_output=True, text=True)
        print("标准输出：", result.stdout)
        print("标准错误：", result.stderr)
        print("返回码：", result.returncode)
        if "1/1" in result.stdout:
            break
        else:
            os.system("sleep 5")

@pytest.fixture(scope="session", autouse=True)
def PermPrepare():
    '''创建组织、部门和用户'''
    name = ["t1", "t2"]
    client = CreateUser(host=host)
    orgId = client.CreateOrganization("impex")
    deps = []
    users = []
    for i in name:        
        depId = client.AddDepartment(orgId, i)
        depIds = [depId]
        userId = client.AddUser(i, depIds, orgId)
        deps.append(depId)
        users.append(userId)
    allure.attach(orgId, name="create user and department success")

    yield deps, users

    '''删除用户、部门和组织'''
    token = GetToken(host=host).get_token(host, "admin", admin_password)
    admin_token = token[1] 

    client = DeleteUser(host=host)
    for userid in users:
        client.DeleteUser(userid)
    re = client.DeleteOrganization(host, admin_token, orgId)
    assert re == 204
    allure.attach(orgId, name="delete user and department success")

@pytest.fixture(scope="session", autouse=True)
def PrepareData(Headers, UserHeaders, PermPrepare):
    '''准备算子、工具箱、MCP'''
    mcp_client = MCP()
    mcp_internal_client = InternalMCP()
    operator_client = Operator()
    toolbox_client = ToolBox()
    perm_client = Perm()
    # AI管理员创建算子：
    # 4个基础算子,第1个为未发布状态，第2为下架状态，第3、4个为已发布状态；
    # 1个内置算子
    operator_ids = []
    # 创建4个基础算子
    filepath = "./resource/openapi/compliant/four_operator.yaml"
    api_data = GetContent(filepath).yamlfile()

    data = {
        "data": str(api_data),
        "operator_metadata_type": "openapi"
    }

    result = operator_client.RegisterOperator(data, Headers)
    assert result[0] == 200
    operators = result[1]
    for operator in operators:
        if operator["status"] == "success":
            operator_ids.append(operator["operator_id"])
            tool_operator_id = operator["operator_id"]
            tool_operator_version = operator["version"]
    for operator_id in operator_ids[1:]:
        update_data = [
            {
                "operator_id": operator_id,
                "status": "published"
            }
        ] 
        result = operator_client.UpdateOperatorStatus(update_data, Headers)    # 未发布 -> 已发布
        assert result[0] == 200
    update_data = [
        {
            "operator_id": operator_ids[1],
            "status": "offline"
        }
    ] 
    result = operator_client.UpdateOperatorStatus(update_data, Headers)    # 已发布 -> 下架
    assert result[0] == 200
    # 创建1个内置算子
    operator_id = str(uuid.uuid4())
    name = ''.join(random.choice(string.ascii_letters) for i in range(8))
    filepath = "./resource/openapi/compliant/test3.yaml"
    api_data = GetContent(filepath).yamlfile()
    data = {
        "operator_id": operator_id,
        "name": name,
        "data": api_data,
        "metadata_type": "openapi",
        "operator_type": "basic",
        "execution_mode": "sync",
        "source": "intenal",
        "config_source": "auto",
        "config_version": "1.0.0"
    }

    result = operator_client.RegisterBuiltinOperator(data, Headers)
    assert result[0] == 200
    operator_ids.append(result[1]["operator_id"])
    
    # AI管理员创建1个内置工具箱和31个自定义工具箱，每个工具箱包含多个工具，第31个自定义工具箱包含一个算子导入的工具
    tool_boxes = []
    toolbox_ids = []
    filepath = "./resource/openapi/compliant/capp-model-manage.json"
    json_data = GetContent(filepath).jsonfile()
    for i in range(31):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        result = toolbox_client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]
        toolbox_ids.append(box_id)
    convert_data = {
        "box_id": toolbox_ids[30],
        "operator_id": tool_operator_id,
        "operator_version": tool_operator_version
    }
    result = toolbox_client.ConvertOperatorToTool(convert_data, Headers)
    assert result[0] == 200
    operator_tool_id = result[1]["tool_id"]
    for box_id in toolbox_ids:
        result = toolbox_client.GetBoxToolsList(box_id, {"all": True}, Headers)
        tools = result[1]["tools"]
        tool_ids = []
        update_data = []
        for tool in tools:
            tool_ids.append(tool["tool_id"])
            data = {
                "tool_id": tool["tool_id"],
                "status": "enabled"
            }
            update_data.append(data)  
        result = toolbox_client.UpdateToolStatus(box_id, update_data, Headers)
        assert result[0] == 200
        box_info = {
            "box_id": box_id,
            "tool_ids": tool_ids
        }
        tool_boxes.append(box_info)
    for box_id in toolbox_ids:
        result = toolbox_client.UpdateToolboxStatus(box_id, {"status": "published"}, Headers) # 发布
        assert result[0] == 200
    # 创建内置工具箱
    box_id = str(uuid.uuid4())
    filepath = "./resource/openapi/compliant/mcp.yaml"
    yaml_data = GetContent(filepath).yamlfile()
    name = ''.join(random.choice(string.ascii_letters) for i in range(8))
    data = {
        "box_id": box_id,
        "box_name": name,
        "box_desc": "test description",
        "data": yaml_data,
        "metadata_type": "openapi",
        "source": "internal",
        "config_version": "1.0.0",
        "config_source": "auto"
    }
    result = toolbox_client.Builtin(data, Headers)
    assert result[0] == 200
    toolbox_ids.append(result[1]["box_id"])

    # AI管理员创建1个内置mcp
    # 1个自定义mcp
    # 1个从工具导入（不包含算子转工具）的mcp两种类型，每个MCP包含30个工具，工具来自不同工具箱
    # 1个包含从算子导入的工具的mcp
    mcp_ids = []
    # 创建自定义MCP
    name = ''.join(random.choice(string.ascii_letters) for i in range(8))
    data = {
        "name": name,
        "description": "test mcp server",
        "mode": "sse",
        "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
        "category": "data_analysis"
    }
    result = mcp_client.RegisterMCP(data, Headers)
    assert result[0] == 200
    mcp_ids.append(result[1]["mcp_id"])
    # 从工具箱导入MCP，mcp下包含30个工具，分别来自不同工具箱
    tool_configs = []
    for box_info in tool_boxes[:30]:
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        tool_config = {
            "box_id": box_info["box_id"],
            "box_name": name,
            "tool_id": box_info["tool_ids"][0],
            "tool_name": f"工具名称_{name}_tool",
            "description": "工具具体描述信息，包括工具的功能、使用方式等"
        }
        tool_configs.append(tool_config)
    name = ''.join(random.choice(string.ascii_letters) for i in range(8))
    data = {
        "name": name,
        "description": "add mcp server config",
        "category": "data_analysis",
        "creation_type": "tool_imported",
        "tool_configs": tool_configs
    }
    result = mcp_client.RegisterMCP(data, Headers)
    assert result[0] == 200
    mcp_ids.append(result[1]["mcp_id"])
    # 从工具箱导入MCP，导入的工具为从算子导入
    name = ''.join(random.choice(string.ascii_letters) for i in range(8))
    data = {
        "name": "mcp_" + name,
        "description": "add mcp server config",
        "category": "data_analysis",
        "creation_type": "tool_imported",
        "tool_configs": [{
            "box_id": toolbox_ids[30],
            "box_name": "box_" + name,
            "tool_id": operator_tool_id,
            "tool_name": f"工具名称_{operator_tool_id[-8:]}_tool",
            "description": "工具具体描述信息，包括工具的功能、使用方式等"
        }]
    }
    result = mcp_client.RegisterMCP(data, Headers)
    assert result[0] == 200
    mcp_ids.append(result[1]["mcp_id"])
    # 创建内置MCP
    name = ''.join(random.choice(string.ascii_letters) for i in range(8))
    mcp_id = str(uuid.uuid4())
    data = {
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
    result = mcp_internal_client.Register(data, UserHeaders)
    assert result[0] == 200
    mcp_ids.append(result[1]["mcp_id"])

    # 用户t1创建算子、工具箱、MCP，并授权给用户t2（公开访问和使用）
    # 获取用户token
    configfile = "./config/env.ini"
    file = GetContent(configfile)
    config = file.config()
    host = config["server"]["host"]
    t1_token = GetToken(host=host).get_token(host, "t1", "111111")
    t1_headers = {
        "Authorization": f"Bearer {t1_token[1]}"
    }
    
    t2_token = GetToken(host=host).get_token(host, "t2", "111111")
    t2_headers = {
        "Authorization": f"Bearer {t2_token[1]}"
    }
    # 给t1配置新建权限
    user_t1 = PermPrepare[1][0]
    data = [
            {
                "accessor": {"id": user_t1, "name": "t1", "type": "user"},
                "resource": {"id": "*", "type": "operator", "name": "算子权限"},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            },
            {
                "accessor": {"id": user_t1, "name": "t1", "type": "user"},
                "resource": {"id": "*", "type": "tool_box", "name": "工具箱权限"},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            },
            {
                "accessor": {"id": user_t1, "name": "t1", "type": "user"},
                "resource": {"id": "*", "type": "mcp", "name": "mcp权限"},
                "operation": {"allow": [{"id": "create"}], "deny": []}
            }
        ]
    result = perm_client.SetPerm(data, Headers)
    assert "20" in str(result[0])
    # 创建算子并发布
    filepath = "./resource/openapi/compliant/template.yaml"
    api_data = GetContent(filepath).yamlfile()
    data = {
        "data": str(api_data),
        "operator_metadata_type": "openapi",
        "direct_publish": True
    }
    result = operator_client.RegisterOperator(data, t1_headers)
    assert result[0] == 200
    assert result[1][0]["status"] == "success"
    t1_operator_id = result[1][0]["operator_id"]
    # 创建工具箱，启用工具并发布
    name = ''.join(random.choice(string.ascii_letters + string.digits) for i in range(8))
    filepath = "./resource/openapi/compliant/mcp.yaml"
    yaml_data = GetContent(filepath).yamlfile()
    data = {
        "box_name": name,
        "data": yaml_data,
        "metadata_type": "openapi"
    }
    result = toolbox_client.CreateToolbox(data, t1_headers)
    assert result[0] == 200
    t1_toolbox_id = result[1]["box_id"]
    result = toolbox_client.GetBoxToolsList(t1_toolbox_id, {"all": True}, t1_headers)
    tools = result[1]["tools"]
    tool_ids = []
    update_data = []
    for tool in tools:
        data = {
            "tool_id": tool["tool_id"],
            "status": "enabled"
        }
        update_data.append(data)  
    result = toolbox_client.UpdateToolStatus(t1_toolbox_id, update_data, t1_headers)
    assert result[0] == 200
    result = toolbox_client.UpdateToolboxStatus(t1_toolbox_id, {"status": "published"}, t1_headers) # 发布
    assert result[0] == 200
    # 创建MCP并发布
    name = ''.join(random.choice(string.ascii_letters + string.digits) for i in range(8))
    data = {
        "name": name,
        "description": "test mcp server",
        "mode": "sse",
        "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL"
    }
    result = mcp_client.RegisterMCP(data, t1_headers)
    assert result[0] == 200
    t1_mcp_id = result[1]["mcp_id"]
    result = mcp_client.MCPReleaseAction(t1_mcp_id, {"status": "published"}, t1_headers)
    assert result[0] == 200
    # 授权给t2
    user_t2 = PermPrepare[1][1]
    data = [
            {
                "accessor": {"id": user_t2, "name": "t2", "type": "user"},
                "resource": {"id": t1_operator_id, "type": "operator", "name": "算子权限"},
                "operation": {"allow": [{"id": "public_access"}, {"id": "execute"}], "deny": []}
            },
            {
                "accessor": {"id": user_t2, "name": "t2", "type": "user"},
                "resource": {"id": t1_toolbox_id, "type": "tool_box", "name": "工具箱权限"},
                "operation": {"allow": [{"id": "public_access"}, {"id": "execute"}], "deny": []}
            },
            {
                "accessor": {"id": user_t2, "name": "t2", "type": "user"},
                "resource": {"id": t1_mcp_id, "type": "mcp", "name": "mcp权限"},
                "operation": {"allow": [{"id": "public_access"}, {"id": "execute"}], "deny": []}
            }
        ]
    result = perm_client.SetPerm(data, t1_headers)
    assert "20" in str(result[0])

    yield operator_ids, toolbox_ids, mcp_ids, t1_operator_id, t1_toolbox_id, t1_mcp_id, t2_headers, user_t2, tool_operator_id, tool_operator_version
