# -*- coding:UTF-8 -*-

import pytest
import allure
import pymysql
import os
import subprocess

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
def delete_operator_data():
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

        # cursor.execute("DELETE FROM t_relations")
        # conn.commit()
        # print(f"表 t_relations 中 {cursor.rowcount} 条记录已被删除")

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
        conn.close()

@pytest.fixture(scope="session", autouse=True)
def delete_policy_data():
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
    name = ["a", "b", "c", "d"]
    client = CreateUser(host=host)
    orgId = client.CreateOrganization("permisson")
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
