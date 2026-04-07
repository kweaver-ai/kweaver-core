# -*- coding:UTF-8 -*-

import pytest
import pymysql
import os
import subprocess

from common.get_content import GetContent

configfile = "./config/env.ini"
file = GetContent(configfile)
config = file.config()

host = config["server"]["host"]
db_port = config["server"]["db_port"]
db_user = config["server"]["db_user"]
db_pwd = config["server"]["db_pwd"]

@pytest.fixture(scope="session", autouse=True)
def delete_category_data():
    '''清空算子分类数据，重启agent-operator-integration服务以重建内置类型'''
    conn = pymysql.connect(host=host, user=db_user, password=db_pwd, port=int(db_port), database="dip_data_operator_hub")
    cursor = conn.cursor()
    try:
        cursor.execute("DELETE FROM t_category")
        conn.commit()
        print(f"表 t_category 中 {cursor.rowcount} 条记录已被删除")
    except Exception as e:
        conn.rollback()
        print(f"error: {str(e)}")
    finally:
        cursor.close()

    os.system("kubectl -n anyshare delete pod $(kubectl get pod -n anyshare | grep agent-operator-integration | awk '{print $1}')")
    command = "kubectl get pod -n anyshare | grep agent-operator-integration | awk '{print $2}'"
    for i in range(60):
        result = subprocess.run(command, shell=True, capture_output=True, text=True)
        print("标准输出：", result.stdout)
        print("标准错误：", result.stderr)
        print("返回码：", result.returncode)
        if "1/1" in result.stdout:
            break
        else:
            os.system("sleep 5")
        