import datetime
import json

from app.interfaces.dbaccess import AddExternalSmallModelInfo
from app.logs.stand_log import StandLogger
from app.mydb.my_pymysql_pool import connect_execute_close_db, connect_execute_commit_close_db

para_dict = {
    "update_time": "f_update_time",
    "create_time": "f_create_time",
    "name": "f_model_name"
}


class SmallModelDao:
    @connect_execute_commit_close_db
    def add_model_info(self, config_info: AddExternalSmallModelInfo, userId, connection, cursor):
        sql = """insert into t_small_model(f_model_id, f_model_name, f_model_type, f_model_config, f_create_time, f_update_time,f_create_by,f_update_by,
        f_adapter,f_adapter_code,f_batch_size,f_max_tokens,f_embedding_dim) 
                    values(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) """
        value_list = [config_info.model_id, config_info.model_name, config_info.model_type,
                      json.dumps(config_info.model_config, ensure_ascii=False), datetime.datetime.today(),
                      datetime.datetime.today(), userId, userId, config_info.adapter, config_info.adapter_code,
                      config_info.batch_size, config_info.max_tokens, config_info.embedding_dim]

        cursor.execute(sql, value_list)

    @connect_execute_commit_close_db
    def edit_model_info(self, config_info: AddExternalSmallModelInfo, userId, connection, cursor):
        sql = """update t_small_model set f_model_name = %s,f_model_type = %s,f_model_config = %s,f_adapter = %s,f_adapter_code = %s,
        f_update_time = %s,f_update_by = %s,f_batch_size = %s,f_max_tokens = %s,f_embedding_dim = %s where f_model_id = %s"""
        value_list = [config_info.model_name, config_info.model_type, json.dumps(config_info.model_config, ensure_ascii=False),
                      config_info.adapter,config_info.adapter_code, datetime.datetime.today(),
                      userId, config_info.batch_size, config_info.max_tokens, config_info.embedding_dim,
                      config_info.model_id]

        cursor.execute(sql, value_list)

    @connect_execute_close_db
    def get_model_info_by_id(self, model_id, connection, cursor):
        sql = f"""select f_model_id, f_model_name, f_model_type, f_model_config, f_create_time, f_update_time,f_adapter, 
        f_adapter_code,f_batch_size,f_max_tokens,f_embedding_dim 
                    from t_small_model where f_model_id = '{model_id}'"""

        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_model_info_by_name(self, model_name, connection, cursor):
        sql = """select f_model_id, f_model_name, f_model_type, f_model_config, f_create_time, f_update_time,f_adapter, 
        f_adapter_code,f_batch_size,f_max_tokens,f_embedding_dim 
                    from t_small_model where f_model_name = %s"""

        cursor.execute(sql, model_name)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_model_info_by_ids(self, model_ids, connection, cursor):
        placeholders = ','.join(['%s'] * len(model_ids))
        sql = f"""select f_model_id,f_model_name from t_small_model where f_model_id IN ({placeholders})"""
        cursor.execute(sql, model_ids)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_model_info_by_names(self, model_names, connection, cursor):
        placeholders = ','.join(['%s'] * len(model_names))
        sql = f"""select f_model_id,f_model_name from t_small_model where f_model_name IN ({placeholders})"""
        cursor.execute(sql, model_names)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_model_info_list(self, page, size, order, rule, model_name, model_type, model_series, permission_ids,
                            connection, cursor):
        sql = """select f_model_id, f_model_name, f_model_type, f_model_config, f_create_time, f_update_time, f_create_by,f_update_by,
        f_adapter,f_adapter_code,f_batch_size,f_max_tokens,f_embedding_dim
                    from t_small_model """
        where_list = []
        value_list = []
        if permission_ids:
            placeholders = ','.join(['%s'] * len(permission_ids))
            where_list.append(f"f_model_id in ({placeholders})")
            value_list.extend(permission_ids)
        if model_name != "":
            where_list.append("f_model_name like %s")
            value_list.append(f"%{model_name}%")
        if model_type != "":
            where_list.append("f_model_type = %s")
            value_list.append(model_type)
        where_sql = f" where {' and '.join(where_list)}" if where_list else ""
        order_sql = f" order by f_{rule} {'desc' if order == 'desc' else 'asc'}"
        limit_sql = f" limit {(int(page) - 1) * int(size)},{size}"
        sql = sql + where_sql + order_sql + limit_sql
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_model_info_total(self, model_name, model_type, model_series, permission_ids, connection, cursor):
        sql = """select f_model_id, f_model_name, f_model_type, f_model_config, f_create_time, f_update_time 
                    from t_small_model """
        where_list = []
        value_list = []
        if model_name != "":
            where_list.append("f_model_name like %s")
            value_list.append(f"%{model_name}%")
        if model_type != "":
            where_list.append("f_model_type = %s")
            value_list.append(model_type)
        if permission_ids:
            placeholders = ','.join(['%s'] * len(permission_ids))
            where_list.append(f"f_model_id in ({placeholders})")
            value_list.extend(permission_ids)
        where_sql = f" where {' and '.join(where_list)}" if where_list else ""
        sql = sql + where_sql
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def delete_model_info_by_ids(self, model_ids, connection, cursor):
        placeholders = ','.join(['%s'] * len(model_ids))
        sql = f"DELETE FROM t_small_model WHERE f_model_id IN ({placeholders})"
        cursor.execute(sql, model_ids)

    @connect_execute_close_db
    def name_check(self, name, connection, cursor):
        sql = """select f_model_id, f_model_name from t_small_model where f_model_name = %s"""
        cursor.execute(sql, name)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_all_ids(self, connection, cursor):
        sql = """select f_model_id from t_small_model limit 1000"""
        cursor.execute(sql)
        res = cursor.fetchall()
        return res


small_model_dao = SmallModelDao()
