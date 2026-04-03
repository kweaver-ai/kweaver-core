import json

import requests

from app.logs.stand_log import StandLogger
from app.mydb.my_pymysql_pool import connect_execute_close_db


# 获取用户名函数
def user(user_id, **kwargs):
    try:
        user_cache = kwargs.get('user_cache', {})
        if user_id in user_cache:
            return user_cache[user_id]
        response = requests.get(f"http://kg-user-rbac:6900/api/rbac/v1/user?field=userId&value={user_id}")
        # response = requests.get(f"http://10.4.134.232:6900/api/rbac/v1/user?field=userId&value={user_id}")
        response = json.loads(response.text)
        response = response['res']
        user_name = response['username']
        user_cache[user_id] = user_name
        return user_name
    except Exception:
        return ""

@connect_execute_close_db
def get_user_name_by_user_id_list(user_id_list, connection, cursor):
    for i in range(0, len(user_id_list)):
        if user_id_list[i] == "":
            user_id_list[i] = "0"
        user_id_list[i] = "'" + user_id_list[i] + "'"
    if not user_id_list or user_id_list == [""]:
        return ()
    sql = """select username, account_id from account where account_id in ({})""".format(",".join(map(str, user_id_list)))
    StandLogger.info_log(sql)
    cursor.execute(sql)
    res = cursor.fetchall()
    return res

@connect_execute_close_db
def get_admin_user_id(connection, cursor):
    sql = """select account_id from account where username='admin'"""
    StandLogger.info_log(sql)
    cursor.execute(sql)
    res = cursor.fetchall()
    if res == [] or res == ():
        return ""
    return res[0]["account_id"]
