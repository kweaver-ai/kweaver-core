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
    ["model_management", "t_small_model", "COLUMN", "ADD", "f_adapter_code", "varchar(15000)", ""],
]

# ！！！以下注释不可删除
# === TEMPLATE START ===
import os
import select
import time

import rdsdriver
import subprocess

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
    conn_cursor.execute(f"SHOW COLUMNS FROM `{db_name}`.`{table_name}` LIKE '{column_name}';")

    # 如果结果集中有数据，表示字段存在
    exists = bool(conn_cursor.fetchall())
    return exists


def alter_column(conn_cursor, db_name, table_name, action_type, column_name, column_props, field_comment):
    if action_type == "ADD" and column_exists(conn_cursor, db_name, table_name, column_name):
        pass
    else:
        if field_comment == "" or field_comment == "DROP AUTO_INCREMENT" or field_comment == "IGNORE":
            conn_cursor.execute(f"ALTER TABLE `{db_name}`.`{table_name}` {action_type} COLUMN `{column_name}` {column_props};")
        else:
            conn_cursor.execute(f"ALTER TABLE `{db_name}`.`{table_name}` {action_type} COLUMN `{column_name}` {column_props} COMMENT '{field_comment}';")

def add_column(conn_cursor, db_name, table_name, action_type, column_name, column_props, field_comment):
    if action_type == "ADD" and column_exists(conn_cursor, db_name, table_name, column_name):
        pass
    else:
        if field_comment == "":
            conn_cursor.execute(f"ALTER TABLE `{db_name}`.`{table_name}` {action_type} COLUMN `{column_name}` {column_props} ;")
        else:
            conn_cursor.execute(f"ALTER TABLE `{db_name}`.`{table_name}` {action_type} COLUMN `{column_name}` {column_props} COMMENT '{field_comment}';")


def drop_column(conn_cursor, db_name, table_name, column_name):
    if column_exists(conn_cursor, db_name, table_name, column_name):
        conn_cursor.execute(f"ALTER TABLE `{db_name}`.`{table_name}` DROP COLUMN {column_name};")


def index_exists(conn_cursor, db_name, table_name, index_name):
    # 查询索引是否存在
    conn_cursor.execute(f"SHOW INDEX FROM `{db_name}`.`{table_name}` WHERE Key_name = '{index_name}';")

    # 如果结果集中有数据，表示索引存在
    exists = bool(conn_cursor.fetchall())
    return exists


def delete_pt_osc_triggers_and_pt_osc_table(conn_cursor, db_name, table_name):
    # 检查是否有残留触发器和残留表
    query = f"""
            SELECT TABLE_NAME
            FROM information_schema.TABLES
            WHERE TABLE_SCHEMA = '{db_name}' AND TABLE_NAME LIKE '\\_%\\_new';
    """
    conn_cursor.execute(query)
    tmp_table_records = conn_cursor.fetchall()

    # 删除pt-osc产生的临时表
    for tmp_table_record in tmp_table_records:
        tmp_table = tmp_table_record[0]
        drop_query = f"DROP TABLE IF EXISTS `{db_name}`.`{tmp_table}`;"
        conn_cursor.execute(drop_query)
        print(f"Deleted table {tmp_table}")

    query = f"""
    SELECT TRIGGER_NAME FROM information_schema.TRIGGERS
    WHERE TRIGGER_SCHEMA = '{db_name}' AND EVENT_OBJECT_TABLE = '{table_name}' and TRIGGER_NAME LIKE 'pt_osc%';
    """
    conn_cursor.execute(query)
    triggers = conn_cursor.fetchall()
    # 删除pt-osc产生的零时触发器
    for trigger in triggers:
        trigger_name = trigger[0]
        query_trigger_name_sql = f"""
        SELECT TRIGGER_NAME FROM information_schema.TRIGGERS
        WHERE TRIGGER_SCHEMA = '{db_name}' AND EVENT_OBJECT_TABLE = '{table_name}' and TRIGGER_NAME LIKE '{trigger_name}';
        """
        conn_cursor.execute(query_trigger_name_sql)
        result = conn_cursor.fetchone()
        if result:
            drop_query = f"DROP TRIGGER IF EXISTS `{db_name}`.`{trigger_name}`;"
            conn_cursor.execute(drop_query)
            print(f"Deleted trigger {trigger_name} on table {table_name}")


def execute_command(cmd_list, conn_cursor, db_name, table_name):
    try:
        # 执行py文件
        # print(f"debug cmd : {cmd_list}")
        process = subprocess.Popen(
            cmd_list,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            encoding='utf-8',
            env={**os.environ, 'PYTHONUNBUFFERED': '1'}
        )

        while True:
            reads = [process.stdout.fileno(), process.stderr.fileno()]
            ret = select.select(reads, [], [])

            for fd in ret[0]:
                if fd == process.stdout.fileno():
                    output = process.stdout.readline()
                    if output:
                        print(output.strip())
                if fd == process.stderr.fileno():
                    error = process.stderr.readline()
                    if error:
                        print(error.strip())

            if process.poll() is not None:
                break
            time.sleep(0.1)

        # 确保所有输出都已经被处理
        for remaining_output in iter(process.stdout.readline, ''):
            print(remaining_output.strip())
        for remaining_error in iter(process.stderr.readline, ''):
            print(remaining_error.strip())
        # 检查返回码
        return_code = process.wait()
        if return_code != 0:
            raise Exception(f"{process.stderr}")
    except Exception as e:
        raise e


def fk_pk_and_trigger_check(conn_cursor, db_name, table_name, index_name, obj_props, obj_type):
    # 1.检查是否存在外键,pt-osc不能使用外键，2.检查是否存在主键，没有主键的表也不使用pt-osc进行升级
    try:
        # 检查主键
        conn_cursor.execute(f"SHOW KEYS FROM `{db_name}`.`{table_name}` WHERE Key_name = 'PRIMARY'")
        primary_keys = conn_cursor.fetchall()

        # 检查外键
        conn_cursor.execute(f"""
            SELECT CONSTRAINT_NAME 
            FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
            WHERE TABLE_NAME = '{table_name}'
            AND TABLE_SCHEMA = '{db_name}'
            AND REFERENCED_TABLE_NAME IS NOT NULL
        """)
        foreign_keys = conn_cursor.fetchall()

        # 返回 True 如果有主键并且没有外键，否则返回 False
        # 检查是否有pt-os创建出来的触发器 or 临时表，有则删除
        if bool(primary_keys) and not bool(foreign_keys):
            delete_pt_osc_triggers_and_pt_osc_table(conn_cursor, db_name, table_name)
            return True
        return False
    except Exception as e:
        raise e


def add_index(conn_cursor, db_name, table_name, index_name, obj_props, obj_type):
    if index_exists(conn_cursor, db_name, table_name, index_name):
        pass
    else:
        check_flag = fk_pk_and_trigger_check(conn_cursor, db_name, table_name, index_name, obj_props, obj_type)
        if os.environ.get("ONLINE_UPGRADE") != "false" and check_flag and os.environ.get("RDS_SOURCE_TYPE") == "internal":
            cmd = "pt-online-schema-change"
            conn_arg = f"u={os.environ['DB_USER']},p={os.environ['DB_PASSWD']},h={os.environ['DB_HOST']},P={os.environ['DB_PORT']},D={db_name},t={table_name}"
            entity_arg = "--alter"
            sql_arg = f"ADD {obj_type} {index_name}({obj_props})"
            arg_list = ["--execute", "--recursion-method=none"]
            cmd_list = [cmd, conn_arg, entity_arg, sql_arg] + arg_list
            execute_command(cmd_list, conn_cursor, db_name, table_name)
        else:
            conn_cursor.execute(f"ALTER TABLE `{db_name}`.`{table_name}` ADD {obj_type} {index_name}({obj_props});")


def drop_index(conn_cursor, db_name, table_name, index_name):
    if index_exists(conn_cursor, db_name, table_name, index_name):
        conn_cursor.execute(f"ALTER TABLE `{db_name}`.`{table_name}` DROP INDEX {index_name};")


def table_exists(conn_cursor, db_name, table_name):
    # 查询表是否存在
    conn_cursor.execute(f"SHOW TABLES FROM `{db_name}` LIKE '{table_name}';")

    # 如果结果集中有数据，表示表存在
    exists = bool(conn_cursor.fetchall())
    return exists


def drop_table(conn_cursor, db_name, table_name):
    if table_exists(conn_cursor, db_name, table_name):
        conn_cursor.execute(f"DROP TABLE `{db_name}`.`{table_name}`;")


def db_exists(conn_cursor, db_name):
    # 查询表是否存在
    conn_cursor.execute(f"SHOW DATABASES LIKE '{db_name}';")

    # 如果结果集中有数据，表示表存在
    exists = bool(conn_cursor.fetchall())
    return exists


def drop_db(conn_cursor, db_name):
    if db_exists(conn_cursor, db_name):
        conn_cursor.execute(f"DROP DATABASE `{db_name}`;")


def drop_unique_index(conn_cursor, db_name, table_name, index_name):
    if index_exists(conn_cursor, db_name, table_name, index_name):
        conn_cursor.execute(f"ALTER TABLE `{db_name}`.`{table_name}` DROP INDEX {index_name};")


def add_unique_index(conn_cursor, db_name, table_name, index_name, obj_props, obj_type):
    if index_exists(conn_cursor, db_name, table_name, index_name):
        pass
    else:
        conn_cursor.execute(f"ALTER TABLE `{db_name}`.`{table_name}` ADD {obj_type} {index_name}({obj_props});")


if __name__ == "__main__":
    conn = get_conn(os.environ["DB_USER"], os.environ["DB_PASSWD"],
                    os.environ["DB_HOST"], os.environ["DB_PORT"])
    conn_cursor = conn.cursor()
    system_id = os.environ.get("SYSTEM_ID", "")
    try:
        conn_cursor.execute(f"SET tidb_allow_remove_auto_inc = ON;")
    except Exception:
        pass
    finally:
        pass
    try:
        for alter in ALTER_TABLE_DICT:
            db_name, table_name, obj_type, action_type, obj_name, obj_props, field_comment = alter
            db_name = system_id + db_name
            if obj_type == "COLUMN":
                if action_type == "DROP":
                    drop_column(conn_cursor, db_name, table_name, obj_name)
                elif action_type == "MODIFY":
                    alter_column(conn_cursor, db_name, table_name, action_type, obj_name, obj_props, field_comment)
                elif action_type == "ADD":
                    add_column(conn_cursor, db_name, table_name, action_type, obj_name, obj_props, field_comment)
            elif obj_type == "INDEX":
                if action_type == "DROP":
                    drop_index(conn_cursor, db_name, table_name, obj_name)
                elif action_type == "ADD":
                    add_index(conn_cursor, db_name, table_name, obj_name, obj_props, obj_type)
            elif obj_type == "UNIQUE INDEX":
                if action_type == "DROP":
                    drop_unique_index(conn_cursor, db_name, table_name, obj_name)
                elif action_type == "ADD":
                    add_unique_index(conn_cursor, db_name, table_name, obj_name, obj_props, obj_type)
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
