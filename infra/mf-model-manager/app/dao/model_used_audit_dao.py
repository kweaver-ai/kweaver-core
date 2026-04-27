import datetime
import json

from app.mydb.my_pymysql_pool import connect_execute_commit_close_db, connect_execute_close_db

from app.logs.stand_log import StandLogger
from app.interfaces import dbaccess, logics


class ModelOpDao():
    # 新建模型配额设置（单条插入）
    # @connect_execute_commit_close_db
    # def add_model_used_log(self, config: dbaccess.ModelUsedAuditInfo, connection, cursor):
    #     sql = """insert into t_model_op_detail (f_id, f_model_id, f_user_id, f_input_tokens, f_output_tokens, f_total_price,
    #             f_create_time, f_currency_type, f_referprice_in, f_referprice_out, f_price_type) values(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"""
    #
    #     value_list = [config.conf_id, config.model_id, config.user_id, config.input_tokens, config.output_tokens,
    #                   config.total_price,
    #                   datetime.datetime.today(), config.currency_type, config.referprice_in, config.referprice_out,
    #                   json.dumps(config.price_type)]
    #     cursor.execute(sql, value_list)

    # 批量插入模型使用日志
    @connect_execute_commit_close_db
    def batch_add_model_used_log(self, batch_data: list, connection, cursor):
        # 使用 INSERT ... ON DUPLICATE KEY UPDATE 实现幂等性
        sql = """INSERT INTO t_model_op_detail (f_id, f_model_id, f_user_id, f_input_tokens, f_output_tokens, f_total_price,
                f_create_time, f_currency_type, f_referprice_in, f_referprice_out, f_price_type, f_total_count, f_failed_count,
                f_average_total_time, f_average_first_time) 
                VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON DUPLICATE KEY UPDATE
                f_input_tokens = f_input_tokens + VALUES(f_input_tokens),
                f_output_tokens = f_output_tokens + VALUES(f_output_tokens),
                f_total_price = f_total_price + VALUES(f_total_price),
                f_create_time = VALUES(f_create_time),
                f_total_count = f_total_count + VALUES(f_total_count),
                f_failed_count = f_failed_count + VALUES(f_failed_count),
                f_average_total_time = ((f_average_total_time * (f_total_count - f_failed_count)) + (VALUES(f_average_total_time) * (VALUES(f_total_count) - VALUES(f_failed_count)))) / (f_total_count - f_failed_count + VALUES(f_total_count) - VALUES(f_failed_count)),
                f_average_first_time = ((f_average_first_time * (f_total_count - f_failed_count)) + (VALUES(f_average_first_time) * (VALUES(f_total_count) - VALUES(f_failed_count)))) / (f_total_count - f_failed_count + VALUES(f_total_count) - VALUES(f_failed_count))"""
        
        values = []
        for config in batch_data:
            value_list = [config.conf_id, config.model_id, config.user_id, config.input_tokens, config.output_tokens,
                          config.total_price,
                          datetime.datetime.today(), config.currency_type, config.referprice_in, config.referprice_out,
                          json.dumps(config.price_type), config.total_count, config.failed_count,
                          config.average_total_time, config.average_first_time]
            values.append(value_list)
        StandLogger.info_log(f"准备批量入库: rows={len(values)}")
        
        try:
            cursor.executemany(sql, values)
            inserted = cursor.rowcount
            StandLogger.info_log(f"批量入库完成: affected_rows={inserted}")
        except Exception as e:
            StandLogger.error(f"批量入库时出错: {e}")
            # 如果批量操作失败，尝试逐条插入
            inserted = 0
            for value_list in values:
                try:
                    cursor.execute(sql, value_list)
                    inserted += 1
                except Exception as single_e:
                    StandLogger.error(f"单条插入失败: {single_e}, 数据: {value_list}")
        
        return inserted

    # 获取模型配额配置列表
    @connect_execute_close_db
    def get_model_config_list(self, config: logics.GetModelOpList, connection, cursor):
        today = datetime.datetime.today()
        month_start = datetime.datetime(year=today.year, month=today.month, day=1, minute=0, second=0)
        sql = """select `t_model_op_detail`.`f_id`, `t_model_op_detail`.`f_model_id` , `t_model_op_detail`.`f_user_id`, 
                `t_model_op_detail`.`f_input_tokens`, `t_model_op_detail`.`f_output_tokens`, 
                `t_model_op_detail`.`f_total_price`, `t_model_op_detail`.`f_create_time`, 
                `t_model_op_detail`.`f_currency_type`, `t_model_op_detail`.`f_referprice_in`, 
                `t_model_op_detail`.`f_referprice_out`, `account`.`username`, `t_model_op_detail`.`f_price_type`,  
                COALESCE(t_model_quota_config.f_billing_type, 1) as f_billing_type, t_model_list.f_model   
                  from t_model_op_detail inner join account on account.account_id = t_model_op_detail.f_user_id  
                  left join t_model_quota_config on t_model_op_detail.f_model_id = t_model_quota_config.f_model_id 
                  inner join t_model_list on t_model_list.f_model_id=t_model_op_detail.f_model_id
                  where `t_model_op_detail`.`f_create_time` >= '%s' """ % month_start

        if config.api_model != "" and config.user_id != "":
            sql += "and f_user_id = '%s' and BINARY t_model_list.f_model = '%s' " % (config.user_id, config.api_model)
        elif config.api_model != "":
            sql += "and BINARY t_model_list.f_model = '%s' " % config.api_model
        elif config.user_id != "":
            sql += "and f_user_id = '%s' " % config.user_id

        order_str = "order by {0} {1} ".format(logics.model_quota_list_dict[config.order], config.rule)
        if config.page != -1:
            limit_str = "limit {0} offset {1}".format(config.size, (config.page - 1) * config.size)
            sql = sql + order_str + limit_str
        else:
            sql = sql + order_str

        StandLogger.info_log(sql)
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    # 获取指定时间范围内的所有log
    @connect_execute_close_db
    def get_model_used_logs_list_within_specified_timeframe(self, start_time, end_time, connection, cursor):
        sql = """select f_model_id, f_user_id, sum(f_input_tokens) as f_input_tokens, sum(f_output_tokens) as f_output_tokens
                from t_model_op_detail where f_create_time >= %s and f_create_time < %s group by f_model_id, f_user_id"""
        value_list = [start_time, end_time]
        StandLogger.info_log(sql % tuple(value_list))
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    # 生成通过model_id获取当月log列表的函数
    @connect_execute_close_db
    def get_model_used_logs_list_within_specified_timeframe_by_model_id(self, model_id, user_id, connection, cursor):
        sql = """select f_id, f_model_id, f_user_id, f_input_tokens, f_output_tokens, f_total_price, f_create_time, 
                f_currency_type, f_price_type  
               from t_model_op_detail where f_model_id = '%s' and f_user_id = '%s' and f_create_time >= '%s'""" % (
            model_id, user_id,
            datetime.datetime(year=datetime.datetime.today().year, month=datetime.datetime.today().month, day=1,
                              minute=0,
                              second=0))

        StandLogger.info_log(sql)
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    # 删除指定时间范围内的所有log
    @connect_execute_commit_close_db
    def delete_model_used_logs_list_within_specified_timeframe(self, start_time, end_time, connection, cursor):
        sql = """delete from t_model_op_detail where f_create_time >= %s and f_create_time < %s"""
        value_list = [start_time, end_time]
        StandLogger.info_log(sql % tuple(value_list))
        cursor.execute(sql, value_list)

    # 归档表写入新数据
    @connect_execute_commit_close_db
    def add_data_to_model_archiving(self, new_id, model_id, archiving, key, file_size, connection, cursor):
        sql = """insert into t_model_archiving(f_id, f_model_id, f_archiving, f_create_time, f_key, f_file_size) 
                values(%s, %s, %s, %s, %s, %s)"""
        value_list = [new_id, model_id, archiving, datetime.datetime.today(), key, file_size]
        StandLogger.info_log(sql % tuple(value_list))
        cursor.execute(sql, value_list)

    @connect_execute_close_db
    def get_model_archiving_list(self, page, size, model_id, order, rule, name, connection, cursor):
        # if "使用明细_" in name:
        #     name = name.replace("使用明细_", "")
        # if "用明细_" in name:
        #     name = name.replace("用明细_", "")
        # if "明细_" in name:
        #     name = name.replace("明细_", "")
        # if "细_" in name:
        #     name = name.replace("细_", "")
        sql = """select f_id, f_model_id, f_archiving, f_create_time, f_key, f_file_size from t_model_archiving """
        where = False
        if model_id != "":
            where = True
            sql += "where f_model_id = %s " % model_id
        if name != "":
            if not where:
                sql += " where "
            else:
                sql += " and "
            sql += "CONCAT('使用明细_', f_archiving) like '%{}%'".format(name)
        if order == "create_time":
            sql += "order by f_create_time "
        elif order == "file_size":
            sql += "order by f_file_size "
        else:
            sql += "order by f_archiving "
        if rule == "desc":
            sql += "desc "
        sql += "limit {}, {} ".format(size * (page - 1), size)
        StandLogger.info_log(sql)
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_all_token_model(self, connection, cursor):
        today = datetime.datetime.today()
        month_start = datetime.datetime(year=today.year, month=today.month, day=1, minute=0, second=0)
        sql = """select distinct t_model_op_detail.f_model_id, t_model_list.f_model_name, t_model_list.f_icon, t_model_list.f_model from t_model_op_detail 
                inner join t_model_list on t_model_op_detail.f_model_id = t_model_list.f_model_id 
                where t_model_op_detail.f_create_time > '%s' and t_model_op_detail.f_user_id!=''""" % month_start
        StandLogger.info_log(sql)
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def check_remain_tokens(self, model_id, user_id, connection, cursor):
        sql = """select account_id from account where username='admin'"""
        StandLogger.info_log(sql)
        cursor.execute(sql)
        res = cursor.fetchall()
        if res == [] or res == ():
            admin_id = ""
        else:
            admin_id = res[0]["account_id"]
        sql0 = """select f_quota, f_create_by from t_model_list where f_model_id=%s"""
        StandLogger.info_log(sql0 % model_id)
        cursor.execute(sql0, model_id)
        quota = cursor.fetchall()
        if quota[0]["f_create_by"] not in [admin_id, user_id]:
            return False, "permission"
        if quota[0]["f_quota"] == 0:
            return True, ""
        num_type_list = [0, 1000, 10000, 100000000, 1000000, 10000000]
        sql1 = """select f_billing_type, f_id, f_model_id, f_input_tokens, f_output_tokens, f_num_type from t_model_quota_config 
                where f_model_id = %s """ % (model_id)
        StandLogger.info_log(sql1)
        cursor.execute(sql1)
        model_quota_config = cursor.fetchall()
        if model_quota_config == () or model_quota_config == []:  # 未设置配额，相当于还有token余额，可以调用
            if quota[0]["f_create_by"] == user_id:
                return True, ""
            else:
                return False, "permission"
        model_quota_config = model_quota_config[0]
        sql2 = """select f_model_conf, f_user_id, f_input_tokens, f_output_tokens, f_num_type from t_user_quota_config 
                where f_model_conf = %s and f_user_id = '%s'""" % (model_quota_config["f_id"], user_id)
        StandLogger.info_log(sql2)
        cursor.execute(sql2)
        user_quota_config = cursor.fetchall()
        if user_quota_config == () or user_quota_config == []:
            return False, "permission"
        user_quota_config = user_quota_config[0]
        sql3 = """select COALESCE(sum(f_input_tokens), 0) as input_tokens_used, COALESCE(sum(f_output_tokens), 0) as output_tokens_used, 
                f_user_id, f_model_id from t_model_op_detail 
                where f_user_id = '%s' and f_model_id = '%s' """ % (user_id, model_id)
        StandLogger.info_log(sql3)
        cursor.execute(sql3)
        token_list = cursor.fetchall()
        if model_quota_config["f_billing_type"] == 1:
            input_tokens = int(
                user_quota_config["f_input_tokens"] * num_type_list[json.loads(user_quota_config["f_num_type"])[0]])
            output_tokens = int(
                user_quota_config["f_output_tokens"] * num_type_list[json.loads(user_quota_config["f_num_type"])[1]])
            # for item in token_list:
            #     input_tokens -= item["f_input_tokens"]
            #     output_tokens -= item["f_output_tokens"]
            input_tokens -= token_list[0]["input_tokens_used"]
            output_tokens -= token_list[0]["output_tokens_used"]
            return_text = ""
            if input_tokens <= 0:
                return_text = "input"
            elif output_tokens <= 0:
                return_text = "output"
            return return_text == "", return_text
        else:
            input_tokens = int(
                user_quota_config["f_input_tokens"] * num_type_list[json.loads(user_quota_config["f_num_type"])[0]])
            # for item in token_list:
            #     input_tokens -= item["f_input_tokens"]
            input_tokens -= token_list[0]["input_tokens_used"] + token_list[0]["output_tokens_used"]
            return input_tokens > 0, "input"

    @connect_execute_close_db
    def get_all_user_id_in_log(self, connection, cursor):
        today = datetime.datetime.today()
        month_start = datetime.datetime(year=today.year, month=today.month, day=1, minute=0, second=0)
        sql = """select distinct t_model_op_detail.f_user_id, account.username from t_model_op_detail inner join 
                account on t_model_op_detail.f_user_id = account.account_id 
                where t_model_op_detail.f_create_time > '%s'""" % month_start
        cursor.execute(sql)
        tmp_res = cursor.fetchall()
        res = []
        for item in tmp_res:
            res.append({
                "user_id": item["f_user_id"],
                "user_name": item["username"]
            })
        return res

    # @connect_execute_close_db
    # def get_total_log_detail(self, config: logics.GetModelOpList, connection, cursor):
    #     today = datetime.datetime.today()
    #     month_start = datetime.datetime(year=today.year, month=today.month, day=1, minute=0, second=0)
    #     sql1 = """select sum(f_input_tokens) as tokens_used from t_model_op_detail
    #             where f_output_tokens = 0 and `f_create_time` >= '%s' """ % month_start
    #     if config.user_id != "":
    #         sql1 += "and f_user_id = '%s' " % config.user_id
    #     StandLogger.info_log(sql1)
    #     cursor.execute(sql1)
    #     res1 = cursor.fetchall()
    #     sql2 = """"""


model_op_dao = ModelOpDao()
