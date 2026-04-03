import datetime
import json
from typing import List
from app.commons.snow_id import worker
from app.mydb.my_pymysql_pool import connect_execute_commit_close_db, connect_execute_close_db
from app.logs.stand_log import StandLogger
from app.interfaces import dbaccess, logics


class ModelQuotaDao():
    # 新建模型配额设置
    @connect_execute_commit_close_db
    def add_new_model_quota_config(self, config: dbaccess.ModelQuotaInfo, connection, cursor):
        sql = """insert into t_model_quota_config (f_id, f_model_id, f_billing_type, f_input_tokens, f_output_tokens, f_referprice_in,
                f_referprice_out, f_currency_type, f_create_time, f_update_time, f_num_type, f_price_type) values(%s, %s,
                %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"""
        value_list = [config.conf_id, config.model_id, config.billing_type, config.input_tokens, config.output_tokens,
                      config.referprice_in,
                      config.referprice_out, config.currency_type, datetime.datetime.today(), datetime.datetime.today(),
                      json.dumps(config.num_type), json.dumps(config.price_type)]
        cursor.execute(sql, value_list)

    # 不需要配额的模型的配额设置添加
    @connect_execute_commit_close_db
    def add_no_model_quota_config(self, conf_id, model_id, user_id, connection, cursor):
        sql = """insert into t_model_quota_config (f_id, f_model_id, f_billing_type, f_input_tokens, f_output_tokens, f_referprice_in,
                        f_referprice_out, f_currency_type, f_create_time, f_update_time, f_num_type, f_price_type) values(%s, %s,
                        %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"""
        value_list = [conf_id, model_id, 1, -1, -1, -1, -1, 0, datetime.datetime.today(), datetime.datetime.today(),
                      "[1, 1]", '["thousand", "thousand"]']
        
        cursor.execute(sql, value_list)

    @connect_execute_commit_close_db
    def edit_model_quota_config(self, config: dbaccess.ModelQuotaInfo, connection, cursor):
        sql = """update t_model_quota_config set f_input_tokens = %s, f_output_tokens = %s, f_referprice_in = %s,
                f_referprice_out = %s, f_currency_type = %s, f_update_time = %s, f_num_type = %s, f_billing_type = %s, 
                f_price_type = %s 
                where f_id = %s"""
        value_list = [config.input_tokens, config.output_tokens, config.referprice_in,
                      config.referprice_out, config.currency_type, config.update_time, json.dumps(config.num_type),
                      config.billing_type, json.dumps(config.price_type), config.conf_id]
        
        cursor.execute(sql, value_list)

    # 检查模型是否已经绑定
    @connect_execute_close_db
    def check_model_is_exist(self, model_id, connection, cursor):
        sql = """select count(f_model_id) from t_model_quota_config where f_model_id = %s"""
        
        cursor.execute(sql, model_id)
        res = cursor.fetchall()
        if "COUNT(f_model_id)" in res[0].keys():
            res[0]["count(f_model_id)"] = res[0]["COUNT(f_model_id)"]
        if res[0]["count(f_model_id)"] > 0:
            return True
        return False

    # 检查模型是否存在
    @connect_execute_close_db
    def check_model_conf_is_exist(self, model_id, connection, cursor):
        sql = """select count(*) from t_model_quota_config where f_id = %s"""
        
        cursor.execute(sql, model_id)
        res = cursor.fetchall()
        if "COUNT(*)" in res[0].keys():
            res[0]["count(*)"] = res[0]["COUNT(*)"]
        if res[0]["count(*)"] > 0:
            return True
        return False

    # 获取限额类型
    @connect_execute_close_db
    def get_model_billing_type(self, conf_id, connection, cursor):
        sql = """select `mq`.`f_id`,  `mq`.`f_billing_type` from t_model_quota_config as mq  where f_id = %s"""
        cursor.execute(sql, conf_id)
        res = cursor.fetchall()
        return res

    # 获取计费配置
    @connect_execute_close_db
    def get_model_billing_info_by_model_id(self, model_id, connection, cursor):
        sql = """select `f_id`, `f_billing_type`, `f_referprice_in`, `f_input_tokens`, 
                `f_referprice_out`, `f_currency_type`, `f_price_type` from t_model_quota_config  where f_model_id = %s"""

        
        cursor.execute(sql, model_id)
        res = cursor.fetchall()
        return res

    # 获取指定模型配额配置
    @connect_execute_close_db
    def get_model_config(self, conf_id, connection, cursor):
        sql = """select `mq`.`f_id`, `mq`.`f_model_id` , `mq`.`f_billing_type`, `mq`.`f_input_tokens`, `mq`.`f_output_tokens`, `mq`.`f_referprice_in`,
                `mq`.`f_referprice_out`, `mq`.`f_currency_type`, `mq`.`f_create_time`, `mq`.`f_update_time`, `mq`.`f_num_type`, 
                `mq`.`f_price_type`, t_llm_model.f_model_name
                from t_model_quota_config as mq join t_llm_model on mq.f_model_id=t_llm_model.f_model_id where f_id = %s"""

        
        cursor.execute(sql, conf_id)
        res = cursor.fetchall()
        return res

    # 获取所有模型配额配置
    @connect_execute_close_db
    def get_all_model_quota_config(self, connection, cursor):
        sql = """select `t_llm_model`.`f_model_id` , `mq`.`f_billing_type`, `mq`.`f_input_tokens`, `mq`.`f_output_tokens`, `mq`.`f_referprice_in`,
                `mq`.`f_referprice_out`, `mq`.`f_currency_type`, `mq`.`f_num_type`, `mq`.`f_price_type`
                from t_model_quota_config as mq right join t_llm_model on mq.f_model_id=t_llm_model.f_model_id limit 1000"""
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    # 获取模型配额配置列表
    @connect_execute_close_db
    def get_model_config_list(self, config: logics.GetModelQuotaList, connection, cursor):
        admin_user_id = "266c6a42-6131-4d62-8f39-853e7093701c"
        sql = """select `mq`.`f_id`, `t_llm_model`.`f_model_id` , `mq`.`f_billing_type`, `mq`.`f_input_tokens`, `mq`.`f_output_tokens`, `mq`.`f_referprice_in`,
                `mq`.`f_referprice_out`, `mq`.`f_currency_type`, COALESCE (`mq`.`f_create_time`, t_llm_model.f_create_time) as f_create_time, 
                COALESCE (`mq`.`f_update_time`, t_llm_model.f_update_time) as f_update_time, `mq`.`f_num_type`, 
                t_llm_model.f_model_name, 
                COALESCE ((`mq`.`f_input_tokens` * `mq`.`f_referprice_in` + `mq`.`f_output_tokens` * `mq`.`f_referprice_out`), 0) as total_price,
                t_llm_model.f_model_series, 
                `mq`.`f_price_type`  
                from t_model_quota_config as mq right join t_llm_model on t_llm_model.f_model_id = `mq`.`f_model_id` 
                where (mq.f_input_tokens >= 0 or mq.f_input_tokens is null) 
                and t_llm_model.f_quota = 1 """
        value_list = []
        if config.name != "":
            config.name = "%" + config.name + "%"
            sql += " and (t_llm_model.f_model_name like %s or t_llm_model.f_model like %s ) "
            value_list.append(config.name)
            value_list.append(config.name)
        if config.api_model != "":
            sql += " and "
            sql += " BINARY  t_llm_model.f_model = %s "
            value_list.append(config.api_model)
        if config.rule == "total_price":
            order_str = "order by total_price {0}  ".format(config.order)
        else:
            order_str = " order by {0} {1} ".format(logics.model_quota_list_dict[config.rule], config.order)
        if config.page != -1:
            limit_str = "limit {0} offset {1}".format(config.size, (config.page - 1) * config.size)
            sql = sql + order_str + limit_str
        else:
            sql = sql + order_str
        if value_list == []:

            cursor.execute(sql)
        else:

            cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def delete_model_quota_config(self, conf_id, connection, cursor):
        # 先删除用户配置表中的数据
        user_model_sql = """delete from t_user_quota_config where f_model_conf = %s"""
        value_list = [conf_id]
        cursor.execute(user_model_sql, value_list)

        # 删除模型配额表中数据
        model_sql = """delete from t_model_quota_config where f_id = %s"""
        value_list = [conf_id]
        cursor.execute(model_sql, value_list)

    # 新建模型配额设置
    @connect_execute_commit_close_db
    def add_new_user_model_quota_config(self, config: dbaccess.ModelUserQuotaInfo, userid: str, connection, cursor):
        sql = """insert into t_user_quota_config (f_id, f_model_conf, f_user_id, f_input_tokens, f_output_tokens, f_create_time, f_update_time)
        values(%s, %s, %s, %s, %s, %s, %s)"""
        value_list = [config.conf_id, config.model_id, userid, config.input_tokens, config.output_tokens,
                      config.create_time, config.update_time]
        
        cursor.execute(sql, value_list)

    @connect_execute_commit_close_db
    def edit_user_model_quota_config(self, config: dbaccess.ModelUserQuotaInfo, connection, cursor):
        sql = """update t_user_quota_config set  f_input_tokens = %s, f_output_tokens = %s, f_update_time = %s where f_id = %s"""
        value_list = [config.input_tokens, config.output_tokens, config.update_time, config.conf_id]
        
        cursor.execute(sql, value_list)

    # 检查用户配置基于当前大模型配置id是否已经存在
    @connect_execute_close_db
    def check_user_model_is_exist(self, model_conf_id, user_id, connection, cursor):
        sql = """select count(f_id) from `t_user_quota_config` where `f_model_conf` = %s and `f_user_id` = %s"""
        value_list = [model_conf_id, user_id]
        
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        if "COUNT(f_id)" in res[0].keys():
            res[0]["count(f_id)"] = res[0]["COUNT(f_id)"]
        if res[0]["count(f_id)"] > 0:
            return True
        return False

    # 检查用户配置是否存在
    @connect_execute_close_db
    def check_user_model_conf_is_exist(self, model_id, connection, cursor):
        sql = """select count(*) from `t_user_quota_config` where f_id = %s"""
        
        cursor.execute(sql, model_id)
        res = cursor.fetchall()
        if "COUNT(*)" in res[0].keys():
            res[0]["count(*)"] = res[0]["COUNT(*)"]
        if res[0]["count(*)"] > 0:
            return True
        return False

    # 获取当前大模型已分配配额
    @connect_execute_close_db
    def get_model_assign_total(self, model_conf_id, connection, cursor):
        sql = """select `f_input_tokens`, `f_output_tokens` from `t_user_quota_config` where `f_model_conf` = %s"""
        cursor.execute(sql, model_conf_id)
        res = cursor.fetchall()

        input_total = 0
        output_total = 0

        for val in res:
            input_total += val["f_input_tokens"]
            output_total += val["f_output_tokens"]
        return input_total, output_total

    # 获取用户配额绑定的大模型配额id
    @connect_execute_close_db
    def get_model_id_by_user_conf_id(self, conf_id, connection, cursor):
        sql = """select `f_model_conf` from `t_user_quota_config` where `f_id` = %s"""

        
        cursor.execute(sql, conf_id)
        res = cursor.fetchall()
        return res

    # 获取指定模型配额配置
    @connect_execute_close_db
    def get_user_model_config(self, conf_id, connection, cursor):
        sql = """select `umq`.`f_id`, `umq`.`f_model_conf` , `umq`.`f_input_tokens`, `umq`.`f_output_tokens`,
        `umq`.`f_user_id`, `umq`.`f_create_time`, `umq`.`f_update_time`, `umq`.`f_num_type`  
        from t_user_quota_config as umq where `umq`.`f_id`= %s"""

        
        cursor.execute(sql, conf_id)
        res = cursor.fetchall()
        return res

    # 获取模型配额配置列表
    @connect_execute_close_db
    def get_user_model_config_list(self, conf_id: str, config: logics.GetModelQuotaList, connection, cursor):
        sql = """select f_id, f_model_conf, f_input_tokens, f_output_tokens, f_create_time, f_update_time, f_user_id, f_num_type  
                from t_user_quota_config where f_model_conf = %s """ % (conf_id)

        order_str = "order by {0} {1} ".format(logics.model_quota_list_dict[config.rule], config.order)
        if config.page != -1:
            limit_str = "limit {0} offset {1}".format(config.size, (config.page - 1) * config.size)
            sql = sql + order_str + limit_str
        else:
            sql = sql + order_str

        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def delete_user_model_quota_config_by_id_list(self, conf_id_list, connection, cursor):
        # 先删除用户配置表中的数据
        user_model_sql = """delete from t_user_quota_config where f_id in ({})"""
        user_model_sql = user_model_sql.format(",".join(conf_id_list))
        cursor.execute(user_model_sql)

    @connect_execute_close_db
    def get_count_and_model_by_user_id(self, user_id, connection, cursor):
        sql = """select f_model, count(t_llm_model.f_model_id) as count_num from t_llm_model 
            inner join t_model_quota_config on t_llm_model.f_model_id = t_model_quota_config.f_model_id 
            inner join t_user_quota_config on t_user_quota_config.f_model_conf = t_model_quota_config.f_model_id 
            
            where t_user_quota_config.f_user_id = '%s'""" % user_id
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_user_quota_model_by_user_id(self, userId, page, size, api_model, order, rule, quota, model_type, connection,
                                        cursor):
        # 构建基础WHERE条件
        if model_type:
            where_clause = f"""where (t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
       or (t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}')"""
        else:
            where_clause = f"""where (t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
       or (t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}')"""
        
        sql = f"""select t_user_quota_config.f_id, t_user_quota_config.f_model_conf, t_user_quota_config.f_user_id, 
                t_user_quota_config.f_input_tokens, t_user_quota_config.f_output_tokens, t_llm_model.f_model, 
                t_llm_model.f_model, t_llm_model.f_model_config, t_llm_model.f_model_id, 
                t_llm_model.f_model_name, t_llm_model.f_max_model_len, t_llm_model.f_model_parameters, 
                t_llm_model.f_create_time, t_llm_model.f_update_time, t_llm_model.f_model_type, 
                t_llm_model.f_model_series, t_llm_model.f_model,t_llm_model.f_create_by,t_llm_model.f_update_by, 
                t_model_quota_config.f_billing_type, t_user_quota_config.f_num_type,
                t_llm_model.f_quota    
                from t_model_quota_config 
                left join t_user_quota_config on t_user_quota_config.f_model_conf = t_model_quota_config.f_id 
                right join t_llm_model on t_llm_model.f_model_id = t_model_quota_config.f_model_id
                {where_clause}"""
        if api_model is not None and api_model != "":
            sql += " and BINARY t_llm_model.f_model = %s " % api_model
        if quota is not None:
            if quota:
                sql += " and t_llm_model.f_quota = 1 "
            else:
                sql += " and t_llm_model.f_quota = 0 "
        if rule != "":
            sql += " order by t_llm_model.f_" + rule
        if order == "desc":
            sql += " desc "
        sql += " limit {},{}".format(str((int(page) - 1) * int(size)), str(size))
        # 修复日志记录，正确使用参数化查询
        
        cursor.execute(sql)
        res1 = cursor.fetchall()
        
        # 构建第二个查询的WHERE条件
        if model_type:
            where_clause2 = f"""where (t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                        or (t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}')"""
        else:
            where_clause2 = f"""where (t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                        or (t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}')"""
        
        sql = f"""select t_user_quota_config.f_id, t_user_quota_config.f_model_conf, t_user_quota_config.f_user_id, 
                        t_user_quota_config.f_input_tokens, t_user_quota_config.f_output_tokens, t_llm_model.f_model, 
                        t_llm_model.f_model_name, t_user_quota_config.f_create_time, t_llm_model.f_model_id, 
                        t_model_quota_config.f_billing_type, t_user_quota_config.f_num_type,
                        t_llm_model.f_quota   
                        from t_model_quota_config 
                        left join t_user_quota_config on t_user_quota_config.f_model_conf = t_model_quota_config.f_id 
                        right join t_llm_model on t_llm_model.f_model_id = t_model_quota_config.f_model_id
                        {where_clause2}"""
        if api_model is not None and api_model != "":
            sql += "and BINARY t_llm_model.f_model = '%s' " % api_model
        if quota is not None:
            if quota:
                sql += "and t_llm_model.f_quota = 1 "
            else:
                sql += "and t_llm_model.f_quota = 0 "
        if rule != "":
            sql += " order by t_llm_model.f_" + rule
        if order == "desc":
            sql += " desc "
        sql += " limit 0,999999999"
        
        cursor.execute(sql)
        res2 = cursor.fetchall()
        
        # 构建第三个查询的WHERE条件
        if model_type:
            where_clause3 = f"""where (t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                                        or (t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}')"""
        else:
            where_clause3 = f"""where (t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                                        or (t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}')"""
        
        sql = f"""select t_user_quota_config.f_id, t_user_quota_config.f_model_conf, t_user_quota_config.f_user_id, 
                                        t_user_quota_config.f_input_tokens, t_user_quota_config.f_output_tokens, t_llm_model.f_model, 
                                        t_llm_model.f_model_name, t_user_quota_config.f_create_time, t_llm_model.f_model_id, 
                                        t_model_quota_config.f_billing_type, t_user_quota_config.f_num_type
                                        from t_model_quota_config 
                                        left join t_user_quota_config on t_user_quota_config.f_model_conf = t_model_quota_config.f_id 
                                        right join t_llm_model on t_llm_model.f_model_id = t_model_quota_config.f_model_id
                                        {where_clause3}"""
        if rule != "":
            sql += " order by t_llm_model.f_" + rule
        if order == "desc":
            sql += " desc "
        sql += " limit 0, 999999999"
        cursor.execute(sql)
        res = cursor.fetchall()
        model_id_list = []
        for item in res1:
            if item["f_model_id"] not in model_id_list:
                model_id_list.append(item["f_model_id"])
        if model_id_list == []:
            res3 = []
        else:
            sql = """select sum(f_input_tokens) as sum_input, sum(f_output_tokens) as sum_output, f_model_id from t_model_op_detail

                                     where f_user_id = %s 
                                            and f_model_id in ({}) group by f_model_id"""
            sql = sql.format(",".join(model_id_list))
            
            cursor.execute(sql, userId)
            res3 = cursor.fetchall()
        return res1, res, len(res2), res3

    @connect_execute_close_db
    def get_user_quota_model_by_user_id_name_fuzzy(self, userId, page, size, name, api_model, order, rule, quota,
                                                   model_type,
                                                   connection, cursor):
        admin_id = "266c6a42-6131-4d62-8f39-853e7093701c"
        
        # 构建基础WHERE条件
        if model_type:
            where_clause = f"""where ((t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                         or (t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}')) 
                         and (t_llm_model.f_model_name like '%{name}%' or t_llm_model.f_model like '%{name}%')"""
        else:
            where_clause = f"""where ((t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                         or (t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}')) 
                         and (t_llm_model.f_model_name like '%{name}%' or t_llm_model.f_model like '%{name}%')"""
        
        sql = f"""select t_user_quota_config.f_id, t_user_quota_config.f_model_conf, t_user_quota_config.f_user_id, 
                        t_user_quota_config.f_input_tokens, t_user_quota_config.f_output_tokens, t_llm_model.f_model, 
                        t_llm_model.f_model_config, 
                        t_llm_model.f_model_name, t_llm_model.f_max_model_len, t_llm_model.f_model_parameters, 
                        t_llm_model.f_create_time, t_llm_model.f_update_time, t_llm_model.f_model_type, 
                        t_llm_model.f_model_id, 
                        t_llm_model.f_model_series, t_llm_model.f_model,t_llm_model.f_create_by,t_llm_model.f_update_by, 
                        t_model_quota_config.f_billing_type, t_user_quota_config.f_num_type,
                        t_llm_model.f_quota   
                        from t_model_quota_config 
                        left join t_user_quota_config on t_user_quota_config.f_model_conf = t_model_quota_config.f_id 
                        right join t_llm_model on t_llm_model.f_model_id = t_model_quota_config.f_model_id 
                        {where_clause}"""
        if api_model is not None and api_model != "":
            sql += " and BINARY t_llm_model.f_model = '%s' " % api_model
        if rule != "":
            sql += " order by t_llm_model.f_" + rule
        if order == "desc":
            sql += " desc "
        sql += f" limit {int(size) * (int(page) - 1)}, {size}"
        cursor.execute(sql)
        res1 = cursor.fetchall()
        
        # 构建第二个查询的WHERE条件
        if model_type:
            where_clause2 = f"""where ((t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                         or (t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}'))  and (t_llm_model.f_model_name like '%{name}%' or 
                                t_llm_model.f_model like '%{name}%')"""
        else:
            where_clause2 = f"""where ((t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                         or (t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}'))  and (t_llm_model.f_model_name like '%{name}%' or 
                                t_llm_model.f_model like '%{name}%')"""
        
        sql = f"""select t_user_quota_config.f_id, t_user_quota_config.f_model_conf, t_user_quota_config.f_user_id, 
                                t_user_quota_config.f_input_tokens, t_user_quota_config.f_output_tokens, t_llm_model.f_model, 
                                t_llm_model.f_model_name, t_user_quota_config.f_create_time, t_llm_model.f_model_id, 
                                t_model_quota_config.f_billing_type, t_user_quota_config.f_num_type,
                                t_llm_model.f_quota   
                                from t_model_quota_config 
                                left join t_user_quota_config on t_user_quota_config.f_model_conf = t_model_quota_config.f_id 
                                right join t_llm_model on t_llm_model.f_model_id = t_model_quota_config.f_model_id 
                                {where_clause2}"""
        if api_model is not None and api_model != "":
            sql += "and BINARY t_llm_model.f_model = '%s' " % api_model
        if rule != "":
            sql += " order by t_llm_model.f_" + rule
        if order == "desc":
            sql += " desc "
        sql += f" limit {int(size) * (int(page) - 1)}, {size}"
        cursor.execute(sql)
        res2 = cursor.fetchall()
        
        # 构建第三个查询的WHERE条件
        if model_type:
            where_clause3 = f"""where ((t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                         or (t_llm_model.f_model_type = '{model_type}' and t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}'))"""
        else:
            where_clause3 = f"""where ((t_llm_model.f_quota = 0 and (t_user_quota_config.f_user_id is null or t_user_quota_config.f_user_id = '{userId}'))
                         or (t_llm_model.f_quota = 1 and t_user_quota_config.f_user_id = '{userId}'))"""
        
        sql = f"""select t_user_quota_config.f_id, t_user_quota_config.f_model_conf, t_user_quota_config.f_user_id, 
                                t_user_quota_config.f_input_tokens, t_user_quota_config.f_output_tokens, t_llm_model.f_model, 
                                t_llm_model.f_model_name, t_user_quota_config.f_create_time, t_llm_model.f_model_id, 
                                t_model_quota_config.f_billing_type, t_user_quota_config.f_num_type,
                                from t_model_quota_config 
                                left join t_user_quota_config on t_user_quota_config.f_model_conf = t_model_quota_config.f_id 
                                right join t_llm_model on t_llm_model.f_model_id = t_model_quota_config.f_model_id 
                                {where_clause3}"""
        if rule != "":
            sql += " order by t_llm_model.f_" + rule
        if order == "desc":
            sql += " desc "
        sql += f" limit 0, 999999999"
        res = cursor.fetchall()
        model_id_list = []
        for item in res1:
            if item["f_model_id"] not in model_id_list:
                model_id_list.append(item["f_model_id"])
        if model_id_list == []:
            res3 = []
        else:
            sql = """select sum(f_input_tokens) as sum_input, sum(f_output_tokens) as sum_output, f_model_id from t_model_op_detail

                                     where f_user_id = %s 
                                            and f_model_id in ({}) group by f_model_id"""
            sql = sql.format(",".join(model_id_list))
            
            cursor.execute(sql, userId)
            res3 = cursor.fetchall()
        return res1, res, len(res2), res3

    @connect_execute_close_db
    def get_user_quota_count_by_user_id(self, userId, name, api_model, connection, cursor):
        sql = """select count(*) from t_user_quota_config where f_user_id = %s """
        
        cursor.execute(sql, userId)
        res = cursor.fetchall()
        if "COUNT(*)" in res[0].keys():
            res[0]["count(*)"] = res[0]["COUNT(*)"]
        return res[0]["count(*)"]

    @connect_execute_close_db
    def get_user_quota_token_used(self, user_id, model_id_list, connection, cursor):
        sql = """select sum(f_input_tokens) as sum_input, sum(f_output_tokens) as sum_output, f_model_id from t_model_op_detail
        
         where f_user_id = %s 
                and f_model_id in ({}) group by f_model_id"""
        sql = sql.format(",".join(model_id_list))
        
        cursor.execute(sql, user_id)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_info_from_model_quota_config_by_model_id(self, model_id, connection, cursor):
        sql = """select f_id, f_model_id, f_billing_type, f_input_tokens, f_output_tokens, f_referprice_in, 
                f_referprice_out, f_currency_type, f_create_time, f_update_time, f_num_type, f_price_type  
                from t_model_quota_config where f_model_id = %s"""
        
        cursor.execute(sql, model_id)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_list_from_user_quota_config_by_conf_id(self, model_conf, connection, cursor):
        sql = """select f_create_time, f_id, f_input_tokens, f_model_conf, f_output_tokens, f_update_time, f_user_id, f_num_type  
                from t_user_quota_config where f_model_conf = %s"""
        cursor.execute(sql, model_conf)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def edit_user_quota_config_by_edit_list(self, edit_list: List[logics.AddUserModelQuota], connection, cursor):
        if not edit_list:
            return
        sql = ""
        for item in edit_list:
            sql = """update t_user_quota_config set f_input_tokens = %s, f_output_tokens = %s, f_update_time = '%s', f_num_type = '%s' 
                    where f_user_id = '%s' and f_model_conf = %s;
                    """ % (item.input_tokens, item.output_tokens,
                           datetime.datetime.today().strftime('%Y-%m-%d %H:%M:%S'), json.dumps(item.num_type),
                           item.user_id, item.model_quota_id)
            
            cursor.execute(sql)

    @connect_execute_commit_close_db
    def add_user_quota_config_by_add_list(self, add_list: List[logics.AddUserModelQuota], connection, cursor):
        if not add_list:
            return
        sql = """insert into t_user_quota_config(f_id, f_model_conf, f_user_id, f_input_tokens, 
                f_output_tokens, f_create_time, f_update_time, f_num_type) values"""
        for item in add_list:
            sql += "(%s, %s, '%s', %s, %s, '%s', '%s', '%s')," % (
                str(worker.get_id()), item.model_quota_id, item.user_id,
                item.input_tokens, item.output_tokens,
                datetime.datetime.today().strftime('%Y-%m-%d %H:%M:%S'),
                datetime.datetime.today().strftime('%Y-%m-%d %H:%M:%S'),
                json.dumps(item.num_type))
        sql = sql[0:-1] + ";"
        
        cursor.execute(sql)

    @connect_execute_close_db
    def check_billing_type_edit(self, billing_type, model_quota_id, connection, cursor):
        sql = """select t_user_quota_config.f_model_conf, t_model_quota_config.f_billing_type from t_user_quota_config 
                inner join t_model_quota_config on t_user_quota_config.f_model_conf = t_model_quota_config.f_id 
                where t_user_quota_config.f_model_conf = %s"""
        value_list = [model_quota_id]
        
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        if res == [] or res == ():
            return True
        return res[0]["f_billing_type"] == billing_type

    # 通过user_id和model_id获取用户模型配额信息
    @connect_execute_close_db
    def get_user_model_quota_by_user_id_and_model_id(self, user_id, model_id, connection, cursor):
        sql = """select t_user_quota_config.f_id, t_user_quota_config.f_model_conf, t_user_quota_config.f_input_tokens, 
                t_user_quota_config.f_output_tokens, t_user_quota_config.f_update_time, t_model_quota_config.f_billing_type, 
                t_user_quota_config.f_num_type  
                from t_user_quota_config inner join t_model_quota_config on t_user_quota_config.f_model_conf = t_model_quota_config.f_model_id 
                where t_user_quota_config.f_user_id = %s and t_model_quota_config.f_model_id = %s"""
        value_list = [user_id, model_id]
        
        cursor.execute(sql, value_list)
        return cursor.fetchall()

    # 模型编辑时通过model_id编辑模型配额表中信息(quota改变时)
    @connect_execute_commit_close_db
    def edit_model_quota_by_model_id_in_model_edit(self, model_id, quota, connection, cursor):
        if quota is False:
            sql = """select * from t_model_quota_config where f_model_id = %s"""
            
            cursor.execute(sql, model_id)
            model_conf = cursor.fetchall()
            if len(model_conf) == 0:
                conf_id = str(worker.get_id())
                sql = """insert into t_model_quota_config (f_id, f_model_id, f_billing_type, f_input_tokens, f_output_tokens, f_referprice_in,
                                        f_referprice_out, f_currency_type, f_create_time, f_update_time, f_num_type, f_price_type) values(%s, %s,
                                        %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"""
                value_list = [conf_id, model_id, 1, -1, -1, 0, 0, 0, datetime.datetime.today(),
                              datetime.datetime.today(),
                              "[1, 1]", '["thousand", "thousand"]']
            else:
                sql = """update t_model_quota_config set f_input_tokens = %s, f_output_tokens = %s, f_referprice_in = %s, 
                                        f_referprice_out = %s, f_currency_type = %s, f_update_time = %s, f_num_type = %s,
                                        f_billing_type = %s     
                                        where f_model_id = %s"""
                value_list = [-1, -1, 0, 0, 0, datetime.datetime.today(), "[1, 1]", 1, model_id]
        else:
            # sql = """update t_model_quota_config set f_input_tokens = %s, f_output_tokens = %s, f_referprice_in = %s,
            #             f_referprice_out = %s, f_currency_type = %s, f_update_time = %s, f_num_type = %s
            #             where f_model_id = %s"""
            # value_list = [9999, 9999, 0, 0, 0, datetime.datetime.today(), "[3, 3]", model_id]
            sql = """delete from t_model_quota_config where f_model_id = %s"""
            value_list = [model_id]
        
        cursor.execute(sql, value_list)
        if quota is False:
            sql = """select f_id from t_model_quota_config where f_model_id = %s"""
            
            cursor.execute(sql, model_id)
            res = cursor.fetchall()
            if res == [] or res == ():
                return
            model_quota_id = res[0]["f_id"]
            sql = """delete from t_user_quota_config where f_model_conf = %s """
            cursor.execute(sql, model_quota_id)

    @connect_execute_commit_close_db
    def delete_previous_month_model_quota(self, connection, cursor):
        now = datetime.datetime.now()
        # 获取本月第一天
        first_day_of_current_month = now.replace(day=1, hour=0, minute=0, second=0, microsecond=0)

        # 删除表t_model_op_detail中本月之前的所有数据
        user_model_sql = """delete from t_model_op_detail where f_create_time < %s"""
        StandLogger.info_log("Deleting model op detail records before: %s" % first_day_of_current_month)
        cursor.execute(user_model_sql, (first_day_of_current_month,))

    @connect_execute_commit_close_db
    def delete_model_quota_by_model_id(self, model_ids, connection, cursor):
        # 先删除用户配置表中的数据
        user_quota_sql = """delete from t_user_quota_config where f_model_conf in (select f_id from t_model_quota_config where f_model_id in ({}))"""
        user_quota_sql = user_quota_sql.format(",".join(model_ids))
        cursor.execute(user_quota_sql)
        # 删除限额配置
        model_quota_sql = """delete from t_model_quota_config where f_model_id in ({})"""
        model_quota_sql = model_quota_sql.format(",".join(model_ids))
        cursor.execute(model_quota_sql)

    @connect_execute_close_db
    def get_model_name_by_quota_config_id(self, conf_id_list, connection, cursor):
        conf_id_list = ",".join(conf_id_list)
        sql = f"""select f_model_name from t_llm_model where f_model_id in (
                    select model_quota.f_model_id from t_model_quota_config model_quota join t_user_quota_config user_quota 
                              on user_quota.f_model_conf=model_quota.f_id
                              where user_quota.f_id in ({conf_id_list})
                              group by model_quota.f_model_id)"""
        cursor.execute(sql)
        return cursor.fetchall()

model_quota_dao = ModelQuotaDao()
