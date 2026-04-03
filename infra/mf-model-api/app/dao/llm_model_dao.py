import datetime
import json
from app.mydb.my_pymysql_pool import connect_execute_close_db, connect_execute_commit_close_db


class ModelDao():
    # 通过模型id获取模型名称
    @connect_execute_close_db
    def get_model_name_by_id(self, model_id, connection, cursor):
        sql = "select f_model_name from t_llm_model where f_model_id={}".format(model_id)

        cursor.execute(sql)
        res = cursor.fetchall()
        return res[0]["f_model_name"]

    @connect_execute_close_db
    def get_model_id_by_name(self, model_name, connection, cursor):
        sql = "select f_model_id from t_llm_model where f_model_name=%s"
        cursor.execute(sql, model_name)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_model_series_by_id(self, model_id, connection, cursor):
        sql = "select f_model_series from t_llm_model where f_model_id={}".format(model_id)

        cursor.execute(sql)
        res = cursor.fetchall()
        return res[0]["f_model_series"]

    @connect_execute_close_db
    def get_model_model_from_model_list_by_id(self, model_id, connection, cursor):
        sql = "select f_model from t_llm_model where f_model_id={}".format(model_id)

        cursor.execute(sql)
        res = cursor.fetchall()
        return res[0]["f_model"]

    @connect_execute_close_db
    def get_all_model_list(self, connection, cursor):
        sql = """select f_create_by,f_create_time,f_model,
                       f_model_config,f_model_id,f_model_name,f_model_series,f_model_type,
                       f_update_by,f_update_time from t_llm_model"""

        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_data_from_model_list_by_id(self, model_id, connection, cursor):
        sql = """select f_create_by,f_create_time,f_model,f_model_config,f_model_id,
                            f_model_name,f_model_series,f_model_type,f_update_by,f_update_time,
                            f_max_model_len, f_model_parameters,f_model_type,f_quota 
                            from t_llm_model where f_model_id='{}'""".format(model_id)

        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_data_from_model_list_by_name_id(self, model_name, model_id, connection, cursor):
        sql = """select f_create_by,f_model,f_model_config,f_model_id,f_quota,
                                f_model_name,f_model_series,f_model_type,f_update_by,f_max_model_len, f_model_parameters 
                                from t_llm_model"""
        if model_name:
            sql += f" where f_model_name='{model_name}'"
        else:
            sql += f" where f_model_id='{model_id}'"
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def delete_model_by_id(self, model_ids, connection, cursor):
        # 使用参数化查询
        placeholders = ','.join(['%s'] * len(model_ids))
        sql = f"DELETE FROM t_llm_model WHERE f_model_id IN ({placeholders})"
        cursor.execute(sql, model_ids)

    @connect_execute_close_db
    def get_all_data_from_model_param(self, connection, cursor):
        sql = """select f_box_component,f_box_lab_cn,f_box_lab_us,f_box_mark_cn,f_box_mark_us,f_max,
                            f_max_mes_cn,f_max_mes_us,f_param_field,f_param_id,f_param_type,f_pat_mes_cn,f_pat_mes_us,
                            f_pattern,f_req,f_req_mes_cn,f_req_mes_us from t_model_param"""

        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_data_from_model_param_by_param_id(self, param_id, connection, cursor):
        sql = """select f_box_component,f_box_lab_cn,f_box_lab_us,f_box_mark_cn,f_box_mark_us,f_max,
                            f_max_mes_cn,f_max_mes_us,f_param_field,f_param_id,f_param_type,f_pat_mes_cn,f_pat_mes_us,
                            f_pattern,f_req,f_req_mes_cn,f_req_mes_us from t_model_param
                            where f_param_id='{}'""".format(param_id)

        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_model_by_name(self, model_name, connection, cursor):
        sql = """select f_create_by,f_create_time,f_model,
                       f_model_config,f_model_id,f_model_name,f_model_series,f_model_type,
                       f_update_by,f_update_time, f_max_model_len, f_model_parameters 
                from t_llm_model where f_model_name=%s;"""

        cursor.execute(sql, model_name)
        res = cursor.fetchall()
        return res

    # 检查模型是否已经绑定
    @connect_execute_close_db
    def check_model_is_exist(self, model_id, connection, cursor):
        sql = """select count(f_model_id) from t_llm_model where f_model_id = %s"""
        cursor.execute(sql, model_id)
        res = cursor.fetchall()
        if "COUNT(f_model_id)" in res[0].keys():
            res[0]["count(f_model_id)"] = res[0]["COUNT(f_model_id)"]
        if res[0]["count(f_model_id)"] > 0:
            return True
        return False

    # 检查是否在同一用户下已有base和model都相同的模型，或者在非admin用户的情况下，admin用户是否已有base和model都相同的模型
    @connect_execute_close_db
    def check_model_unique(self, base, model, user_id, api_key, connection, cursor):
        sql = """select f_model_id, f_model_config from t_llm_model where f_model='{}'""".format(
            model)

        cursor.execute(sql)
        models = cursor.fetchall()
        for item in models:
            model_config = json.loads(item["f_model_config"])
            if "api_base" in model_config.keys() and model_config["api_base"] == base and \
                    (api_key is None or ("api_key" in model_config.keys() and model_config["api_key"] == api_key)):
                return True
            if "api_url" in model_config.keys() and model_config["api_url"] == base and \
                    (api_key is None or ("api_key" in model_config.keys() and model_config["api_key"] == api_key)):
                return True
        return False

    @connect_execute_close_db
    def get_model_default_paras(self, connection, cursor):
        sql = """select f_model_id,f_model_name,f_model_series,f_model from t_llm_model"""

        cursor.execute(sql)
        res = cursor.fetchall()
        model_id_list = []
        for item in res:
            model_id_list.append(item["f_model_id"])
        res_dict = {}
        for item in res:
            res_dict[item["f_model_id"]] = {"model_name": item["f_model_name"], "model_series": item["f_model_series"],
                                            "model": item["f_model"]}
        return res_dict

    @connect_execute_close_db
    def get_all_tome_model_list(self, connection, cursor):
        sql = """select f_model_config,f_model_id,f_model_name from t_llm_model where f_model_series = 'tome'"""

        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_ten_minutes_ago_monitor_data(self, model_ids, ten_minutes_ago_format, connection, cursor):
        sql = f"""select f_model_id, f_generation_tokens_total,f_prompt_tokens_total from t_model_monitor where f_create_time='{ten_minutes_ago_format}'
        and f_model_id in ({",".join(model_ids)})"""

        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def get_monitor_data(self, model_id, connection, cursor):
        sql = f"""select f_create_time,f_model_id, f_average_first_token_time,f_generation_token_speed,f_total_token_speed 
        from t_model_monitor where  f_model_id = {model_id} order by f_create_time desc limit 36"""

        # 使用executemany处理多行数据

        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def delete_model_monitor_data(self, connection, cursor):
        now = datetime.datetime.now()
        thirty_days_ago = now - datetime.timedelta(days=30)

        # 使用参数化查询
        sql = "DELETE FROM t_model_monitor WHERE f_create_time < %s"

        cursor.execute(sql, thirty_days_ago)

    @connect_execute_close_db
    def get_quota_by_user_and_model(self, user_id, model_id, connection, cursor):
        # 获取当前日期的月初一号
        now_month = datetime.datetime.now().replace(day=1).strftime('%Y-%m-%d')
        sql = f"""SELECT
                    uqc.f_input_tokens, COALESCE(monthly_usage.used_input_tokens,0) AS used_input_tokens,mqc.f_billing_type,
                    uqc.f_output_tokens,COALESCE(monthly_usage.used_output_tokens,0) AS used_output_tokens,uqc.f_num_type
                FROM
                    t_model_quota_config mqc join t_user_quota_config uqc on uqc.f_model_conf=mqc.f_id
                LEFT JOIN (
                    SELECT
                        f_user_id,
                        f_model_id,
                        SUM(f_input_tokens) AS used_input_tokens,
                        SUM(f_output_tokens) AS used_output_tokens
                    FROM
                        t_model_op_detail
                    WHERE
                        f_user_id = '{user_id}'
                        AND f_model_id = '{model_id}'
                        AND f_create_time >= '{now_month}'
                    GROUP BY
                        f_user_id, f_model_id
                ) AS monthly_usage
                ON
                    uqc.f_user_id = monthly_usage.f_user_id
                    AND mqc.f_model_id = monthly_usage.f_model_id
                WHERE
                    uqc.f_user_id = '{user_id}'
                    AND mqc.f_model_id = '{model_id}';"""

        cursor.execute(sql)
        res = cursor.fetchall()

        # 如果没有配额信息，直接返回空结果
        if not res:
            return res

        # 定义单位换算列表，与model_used_audit_dao.py中保持一致
        num_type_list = [0, 1000, 10000, 100000000, 1000000, 10000000]

        # 获取配额配置信息
        quota_info = res[0]
        f_billing_type = quota_info["f_billing_type"]
        f_num_type = json.loads(quota_info["f_num_type"])

        # 获取配置的总配额（未使用单位换算）
        config_input_tokens = quota_info["f_input_tokens"]
        config_output_tokens = quota_info["f_output_tokens"]

        # 获取已使用的配额
        used_input_tokens = quota_info["used_input_tokens"]
        used_output_tokens = quota_info["used_output_tokens"]

        # 根据单位换算配置计算总配额（使用单位换算后的值）
        total_input_tokens = int(config_input_tokens * num_type_list[f_num_type[0]])
        total_output_tokens = int(config_output_tokens * num_type_list[f_num_type[1]])

        # 计算剩余额度
        if f_billing_type == 1:
            # 分别计算输入和输出的剩余额度
            remaining_input_tokens = total_input_tokens - used_input_tokens
            remaining_output_tokens = total_output_tokens - used_output_tokens
        else:
            # 合并计算输入和输出的剩余额度，从输入配额中扣除
            total_used_tokens = used_input_tokens + used_output_tokens
            remaining_input_tokens = total_input_tokens - total_used_tokens
            remaining_output_tokens = True

        # 将计算后的剩余额度添加到返回结果中
        res[0]["remaining_input_tokens"] = remaining_input_tokens
        res[0]["remaining_output_tokens"] = remaining_output_tokens

        # 保留原有的配额配置信息
        res[0]["total_input_tokens"] = total_input_tokens
        res[0]["total_output_tokens"] = total_output_tokens

        return res

    @connect_execute_close_db
    def get_data_from_default_model(self, connection, cursor):
        sql = """select f_create_by,f_model,f_model_config,f_model_id,f_quota,
                                f_model_name,f_model_series,f_model_type,f_update_by,f_max_model_len, f_model_parameters 
                                from t_llm_model where f_default=1"""

        cursor.execute(sql)
        res = cursor.fetchall()
        return res


llm_model_dao = ModelDao()
