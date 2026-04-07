# -*- coding:UTF-8 -*-

import pytest
import pymysql

from common.get_content import GetContent

configfile = "./config/env.ini"
file = GetContent(configfile)
config = file.config()

host = config["server"]["host"]
db_port = config["server"]["db_port"]
db_user = config["server"]["db_user"]
db_pwd = config["server"]["db_pwd"]

@pytest.fixture(scope="session", autouse=True)
def delete_operator_data():
    '''清理算子记录'''
    conn = pymysql.connect(host=host, user=db_user, password=db_pwd, port=int(db_port), database="dip_data_operator_hub")
    cursor = conn.cursor()
    try:
        cursor.execute("DELETE FROM t_mcp_server_config")
        conn.commit()
        print(f"表 t_mcp_server_config 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_mcp_server_release")
        conn.commit()
        print(f"表 t_mcp_server_release 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_mcp_server_release_history")
        conn.commit()
        print(f"表 t_mcp_server_release_history 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_internal_component_config")
        conn.commit()
        print(f"表 t_internal_component_config 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_mcp_tool")
        conn.commit()
        print(f"表 t_mcp_tool 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_resource_deploy")
        conn.commit()
        print(f"表 t_resource_deploy 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_tool")
        conn.commit()
        print(f"表 t_tool 中 {cursor.rowcount} 条记录已被删除")

        cursor.execute("DELETE FROM t_toolbox")
        conn.commit()
        print(f"表 t_toolbox 中 {cursor.rowcount} 条记录已被删除")
    except Exception as e:
        conn.rollback()
        print(f"error: {str(e)}")
    finally:
        cursor.close()