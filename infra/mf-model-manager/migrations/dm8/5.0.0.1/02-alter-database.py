#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# 定义表结构变更列表
# 操作对象包含 COLUMN、INDEX、UNIQUE INDEX
# 对象名对应操作对象的名称，如果操作对象为 COLUMN，则为字段名；如果操作对象为 INDEX/UNIQUE INDEX，则为索引名
# 操作类型包含 ADD、DROP、MODIFY
# 对象属性如果是 COLUMN，包含字段类型、是否为空、默认值；如果是 INDEX/UNIQUE INDEX，包含索引列(联合索引列之间用逗号分隔)、排序方式
# 对象属性、字段注释没有时填空字符串
# 特例：删除表，只需数据库名，表名，操作对象为TABLE，操作类型为DROP，其他全填空字符串
# 特例：删除库，只需数据库名，操作对象为DB，，操作类型为DROP，其他全填空字符串
ALTER_TABLE_DICT = [
    # 数据库名，   表名，              操作对象， 操作类型, 对象名，            对象属性，                            字段注释
    ["model_management", "t_small_model", "COLUMN", "ADD", "f_adapter", "int DEFAULT 0", ""],
    ["model_management", "t_small_model", "COLUMN", "ADD", "f_adapter_code", "VARCHAR(15000 CHAR)", ""],
]

# ！！！以下注释不可删除
# === TEMPLATE START ===
import os
import rdsdriver


def get_conn(user, password, host, port):
    try:
        conn = rdsdriver.connect(host=host,
                                 port=int(port),
                                 user=user,
                                 password=password,
                                 autocommit=True)
    except Exception as e:
        print("connect database error: %s", str(e))
        raise e
    return conn


def column_exists(conn_cursor, db_name, table_name, column_name):
    # 查询字段是否存在
    conn_cursor.execute(f"SELECT COLUMN_NAME FROM ALL_TAB_COLUMNS WHERE OWNER='{db_name}' AND TABLE_NAME='{table_name}' AND COLUMN_NAME='{column_name}';")

    # 如果结果集中有数据，表示字段存在
    exists = bool(conn_cursor.fetchall())
    return exists


def alter_column(conn_cursor, db_name, table_name, action_type, column_name, column_props, field_comment):
    if action_type == "ADD" and column_exists(conn_cursor, db_name, table_name, column_name):
        pass
    elif action_type == "MODIFY" and field_comment == "DROP AUTO_INCREMENT":
        # DROP IDENTITY 不支持幂等
        conn_cursor.execute(f"SELECT TABLEDEF('{db_name}','{table_name}');")
        result = conn_cursor.fetchall()
        if 'IDENTITY' in str(result):
            conn_cursor.execute(f'ALTER TABLE "{db_name}"."{table_name}" DROP IDENTITY;')
    else:
        conn_cursor.execute(f'ALTER TABLE "{db_name}"."{table_name}" {action_type} "{column_name}" {column_props};')


def drop_column(conn_cursor, db_name, table_name, column_name):
    if column_exists(conn_cursor, db_name, table_name, column_name):
        conn_cursor.execute(f'ALTER TABLE "{db_name}"."{table_name}" DROP "{column_name}";')


def add_index(conn_cursor, db_name, table_name, index_name, obj_props, index_type):
    conn_cursor.execute(f'CREATE {index_type} IF NOT EXISTS {table_name}_{index_name} ON "{db_name}"."{table_name}"({obj_props});')


def drop_index(conn_cursor, db_name, table_name, index_name):
    conn_cursor.execute(f'DROP INDEX IF EXISTS "{db_name}".{table_name}_{index_name};')


def drop_table(conn_cursor, db_name, table_name):
    conn_cursor.execute(f'DROP TABLE IF EXISTS "{db_name}"."{table_name}";')


def drop_db(conn_cursor, db_name):
    conn_cursor.execute(f'DROP SCHEMA IF EXISTS "{db_name}";')


if __name__ == "__main__":
    conn = get_conn(os.environ["DB_USER"], os.environ["DB_PASSWD"],
                    os.environ["DB_HOST"], os.environ["DB_PORT"])
    conn_cursor = conn.cursor()
    system_id = os.environ.get("SYSTEM_ID", "")
    try:
        for alter in ALTER_TABLE_DICT:
            db_name, table_name, obj_type, action_type, obj_name, obj_props, field_comment = alter
            db_name = system_id + db_name
            if obj_type == "COLUMN":
                if action_type == "DROP":
                    drop_column(conn_cursor, db_name, table_name, obj_name)
                elif action_type == "ADD" or action_type == "MODIFY":
                    alter_column(conn_cursor, db_name, table_name, action_type, obj_name, obj_props, field_comment)
            elif obj_type == "INDEX" or obj_type == "UNIQUE INDEX":
                if action_type == "DROP":
                    drop_index(conn_cursor, db_name, table_name, obj_name)
                elif action_type == "ADD":
                    add_index(conn_cursor, db_name, table_name, obj_name, obj_props, obj_type)
            elif obj_type == "TABLE":
                if action_type == "DROP":
                    drop_table(conn_cursor, db_name, table_name)
            elif obj_type == "DB":
                if action_type == "DROP":
                    drop_db(conn_cursor, db_name)
    except Exception:
        raise Exception()
    finally:
        conn_cursor.close()
        conn.close()

# === TEMPLATE END ===
