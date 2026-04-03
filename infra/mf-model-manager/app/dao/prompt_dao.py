import datetime

from app.logs.stand_log import StandLogger
from app.mydb.my_pymysql_pool import connect_execute_commit_close_db, connect_execute_close_db


class PromptDao():
    # 从数据库中查询所有提示词项目列表
    @connect_execute_close_db
    def get_all_prompt_item_list_distinct(self, is_management, connection, cursor):
        sql = """select f_prompt_item_id,f_prompt_item_name,f_create_time,f_create_by,f_update_time,f_update_by,
                    count(distinct f_prompt_item_id) from t_prompt_item_list 
                    group by f_prompt_item_id """
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    # 数有几个item_id
    @connect_execute_close_db
    def count_prompt_item_id(self, connection, cursor):
        sql = """select count(distinct f_prompt_item_id) from t_prompt_item_list"""
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res[0]["count(distinct f_prompt_item_id)"]

    # 从数据库中模糊查询提示词项目列表
    @connect_execute_close_db
    def get_prompt_item_list_by_fuzzy_query_distinct(self, prompt_item_name, connection, cursor):
        sql = """select f_prompt_item_id,f_prompt_item_name,f_create_time,f_create_by,f_update_time,f_update_by,
                    count(distinct f_prompt_item_id) from t_prompt_item_list 
                    where (f_prompt_item_name like %s or f_prompt_item_type like %s) 
                    group by f_prompt_item_id"""
        value_list = [f'%{prompt_item_name}%', f'%{prompt_item_name}%']
        
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    # 通过联查提示词名称，模糊查询对应的提示词项目列表
    @connect_execute_close_db
    def get_prompt_item_list_by_prompt_name_fuzzy(self, prompt_name,connection, cursor):
        sql = """select prompt_item.f_prompt_item_id,prompt_item.f_prompt_item_name,prompt_item.f_prompt_item_type_id,prompt_item.f_prompt_item_type,
                    prompt_item.f_create_time,prompt_item.f_create_by,prompt_item.f_update_time,prompt_item.f_built_in,
                    prompt_item.f_update_by,prompt.f_prompt_id,prompt.f_prompt_name,prompt.f_prompt_desc,prompt.f_messages,
                    prompt.f_variables,prompt.f_icon,prompt.f_create_time as prompt_create_time,prompt.f_update_time  as prompt_update_time,
                    prompt.f_create_by  as prompt_create_by,prompt.f_update_by as prompt_update_by,prompt.f_prompt_type,
                     prompt.f_built_in as prompt_built_in
                    from t_prompt_item_list as prompt_item join t_prompt_list as prompt
                    on prompt_item.f_prompt_item_type_id = prompt.f_prompt_item_type_id 
                    where prompt.f_prompt_name like %s"""
        value_list = [f'%{prompt_name}%']
        
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_prompt_item_type_by_item_id(self, item_id, connection, cursor):
        sql = """select f_prompt_item_type_id, f_prompt_item_type from t_prompt_item_list 
                            where f_prompt_item_id='{}'""".format(item_id)
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    # 从数据库中获取提示词列表信息
    @connect_execute_close_db
    def get_prompt_list(self, prompt_item_id, prompt_item_type_id,
                        page, size, prompt_name, order, rule, deploy, prompt_type, connection, cursor):

        sql = """select * from t_prompt_list join t_prompt_item_list 
                           on t_prompt_list.f_prompt_item_type_id=t_prompt_item_list.f_prompt_item_type_id where 
                           f_prompt_name like %s and 
                           t_prompt_list.f_is_delete=0
                           """
        value_list = [f"%{prompt_name}%"]
        if prompt_item_id:
            sql += " and t_prompt_list.f_prompt_item_id=%s "
            value_list.append(prompt_item_id)
        if prompt_item_type_id:
            sql += " and t_prompt_list.f_prompt_item_type_id=%s "
            value_list.append(prompt_item_type_id)
        if deploy != '':
            sql += " and t_prompt_list.f_is_deploy=%s "
            value_list.append(deploy)
        if prompt_type != '':
            sql += " and t_prompt_list.f_prompt_type=%s "
            value_list.append(prompt_type)
        if rule != "":
            sql += " order by t_prompt_list.f_" + rule
        if order == "desc":
            sql += " desc"
        sql += " limit {},{}".format(str(int(page) - 1), str(size))
        
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    # 获取指定提示词项目的名称
    @connect_execute_close_db
    def get_prompt_item_name_by_id(self, prompt_item_id, connection, cursor):
        sql = "select f_prompt_item_name from t_prompt_item_list where f_prompt_item_id={} limit 1".format(
            prompt_item_id)
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res[0]["f_prompt_item_name"]

    # 获取指定提示词类型的名称
    @connect_execute_close_db
    def get_prompt_item_type_by_type_id(self, prompt_item_type_id, connection, cursor):
        sql = "select f_prompt_item_type from t_prompt_item_list where f_prompt_item_type_id={}".format(
            prompt_item_type_id)
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res[0]["f_prompt_item_type"]

    # 从数据库中获取提示词模板列表信息
    @connect_execute_close_db
    def get_prompt_template_list(self, prompt_type, prompt_name, connection, cursor):
        sql = """select f_icon,f_prompt_id,f_messages,f_prompt_name,f_prompt_type,f_prompt_desc,f_opening_remarks,
                       f_input,f_variables from t_prompt_template_list """
        where_list = []
        value_list = []
        if prompt_type is not None and prompt_type != "":
            where_list.append(" f_prompt_type=%s ")
            value_list.append(prompt_type)

        if prompt_name != None and prompt_name != "":
            # sql += " and f_prompt_name like '%{}%'".format(prompt_name)
            where_list.append(" f_prompt_name like %s ")
            value_list.append("%" + prompt_name + "%")
        if len(where_list) > 0:
            sql += 'where ' + ' and '.join(where_list)
        
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    # 通过id获取提示词信息
    @connect_execute_close_db
    def get_prompt_by_id(self, prompt_id, connection, cursor):
        sql = """select f_prompt_id,f_prompt_name,f_model_id,f_prompt_item_id,f_prompt_service_id,f_prompt_item_type_id,
                    f_messages,f_opening_remarks,f_variables,f_prompt_type,f_prompt_desc,f_is_deploy,f_icon,f_model_para 
                    ,f_is_delete,f_prompt_deploy_url from t_prompt_list where f_prompt_id='{}'""".format(prompt_id)
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    # 获取所有prompt_service_id
    @connect_execute_close_db
    def get_all_prompt_service_id(self, connection, cursor):
        sql = "select f_prompt_service_id from t_prompt_list"
        
        cursor.execute(sql)
        res = cursor.fetchall()
        res_list = []
        for item in res:
            res_list.append(item["f_prompt_service_id"])
        return res_list

    @connect_execute_close_db
    def get_prompt_by_service_id(self, prompt_service_id, connection, cursor):
        sql = """select f_messages,f_model_id,f_is_deploy,f_variables,f_model_para from t_prompt_list 
                       where f_prompt_service_id={}""".format(prompt_service_id)
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_all_info_from_prompt_item_list_by_item_id(self, prompt_item_id, connection, cursor):
        sql = """select f_create_by,f_create_time,f_id,f_item_is_delete,f_prompt_item_id,f_prompt_item_name,
                            f_prompt_item_type,f_prompt_item_type_id,f_type_is_delete,f_update_by,f_update_time 
                            from t_prompt_item_list where f_prompt_item_id='{}' and f_type_is_delete=0""".format(
            prompt_item_id)
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    # 添加提示词项目
    @connect_execute_commit_close_db
    def add_prompt_item(self, id, prompt_item_id, prompt_item_type_id, prompt_item_type, create_by, update_by,
                        prompt_item_name, connection, cursor):
        sql = """insert into t_prompt_item_list(f_id,f_prompt_item_id,f_prompt_item_type_id,f_prompt_item_type,
                            f_create_by,f_update_by,f_update_time,f_create_time,f_prompt_item_name)
                    values(%s, %s, %s, %s, %s, %s, %s, %s, %s)"""
        value_list = [id, prompt_item_id, prompt_item_type_id, prompt_item_type, create_by, update_by,
                      datetime.datetime.today(), datetime.datetime.today(), prompt_item_name]
        
        cursor.execute(sql, value_list)

    # 编辑已有提示词项目
    @connect_execute_commit_close_db
    def edit_prompt_item(self, prompt_item_id, prompt_item_name, userId, connection, cursor):
        sql = """update t_prompt_item_list set f_prompt_item_name=%s, f_update_by=%s, f_update_time=%s 
                            where f_prompt_item_id=%s"""
        value_list = [prompt_item_name, userId, datetime.datetime.today(), prompt_item_id]
        
        cursor.execute(sql, value_list)

    # 通过type_id获取prompt_item_list表中数据
    @connect_execute_close_db
    def get_data_from_prompt_item_list_by_type_id(self, prompt_item_type_id, connection, cursor):
        sql = """select f_create_by,f_create_time,f_id,f_item_is_delete,f_prompt_item_id,f_prompt_item_name,
                            f_prompt_item_type,f_prompt_item_type_id,f_type_is_delete,f_update_by,f_update_time 
                            from t_prompt_item_list where f_prompt_item_type_id='{}'""".format(prompt_item_type_id)
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    # 编辑已有提示词项目分组
    @connect_execute_commit_close_db
    def edit_prompt_item_type(self, prompt_item_type_id, prompt_item_type, userId, connection, cursor):
        sql = """update t_prompt_item_list set f_prompt_item_type=%s, f_update_by=%s, 
                            f_update_time=%s where f_prompt_item_type_id=%s"""
        value_list = [prompt_item_type, userId, datetime.datetime.today(), prompt_item_type_id]
        
        cursor.execute(sql, value_list)

    @connect_execute_close_db
    def get_data_from_prompt_item_list_by_id_and_type_id(self, prompt_item_id, prompt_item_type_id, connection, cursor):
        sql = """select f_create_by,f_create_time,f_id,f_item_is_delete,f_prompt_item_id,f_prompt_item_name,
                            f_prompt_item_type,f_prompt_item_type_id,f_type_is_delete,f_update_by,f_update_time 
                            from t_prompt_item_list where f_prompt_item_type_id='{}' 
                            and f_prompt_item_id='{}'""".format(prompt_item_type_id, prompt_item_id)
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def get_data_from_prompt_list_by_item_type_id(self, item_type_id, connection, cursor):
        sql = """select f_create_by,f_create_time,f_icon,f_is_delete,f_is_deploy,f_messages,f_model_id,
                                f_model_para,f_opening_remarks,f_prompt_deploy_api,f_prompt_deploy_url,f_prompt_desc,
                                f_prompt_id,f_prompt_item_id,f_prompt_item_type_id,f_prompt_name,f_prompt_service_id,
                                f_prompt_type,f_update_by,f_update_time,f_variables from t_prompt_list where 
                                f_prompt_item_type_id='{}'""".format(item_type_id)
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def add_prompt_to_prompt_list(self, prompt_id, prompt_item_id, prompt_item_type_id, prompt_type, prompt_service_id,
                                  prompt_name, prompt_desc, icon, variables, model_id, model_para, opening_remarks,
                                  create_by, update_by, messages, connection, cursor):
        is_built_in = 0
        sql = """insert into t_prompt_list(f_prompt_id,f_prompt_item_id,f_prompt_item_type_id,f_prompt_type,
                            f_prompt_service_id,f_prompt_name,f_prompt_desc,f_icon,f_variables,f_model_id,f_model_para,
                            f_opening_remarks,f_create_by,f_update_by,f_messages,f_update_time,f_create_time,f_built_in) values 
                            (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"""
        value_list = [prompt_id, prompt_item_id, prompt_item_type_id, prompt_type, prompt_service_id, prompt_name,
                      prompt_desc, icon, variables, model_id, model_para, opening_remarks, create_by, update_by,
                      messages, datetime.datetime.today(), datetime.datetime.today(), is_built_in]
        
        cursor.execute(sql, value_list)

    @connect_execute_close_db
    def get_all_data_from_prompt_list(self, connection, cursor):
        sql = """select f_create_by,f_create_time,f_icon,f_is_delete,f_is_deploy,f_messages,f_model_id,
                                f_model_para,f_opening_remarks,f_prompt_deploy_api,f_prompt_deploy_url,f_prompt_desc,
                                f_prompt_id,f_prompt_item_id,f_prompt_item_type_id,f_prompt_name,f_prompt_service_id,
                                f_prompt_type,f_update_by,f_update_time,f_variables from t_prompt_list"""
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def edit_name_in_prompt_list(self, prompt_id, prompt_name, model_id, icon, prompt_desc, update_by, prompt_item_id,
                                 prompt_item_type_id, connection, cursor):
        sql = """update t_prompt_list set f_prompt_name=%s,f_model_id=%s,f_icon=%s,f_prompt_desc=%s,
                            f_update_by=%s,f_update_time=%s,f_prompt_item_id=%s,f_prompt_item_type_id=%s 
                            where f_prompt_id=%s"""
        value_list = [prompt_name, model_id, icon, prompt_desc, update_by, datetime.datetime.today(), prompt_item_id,
                      prompt_item_type_id, prompt_id]
        
        cursor.execute(sql, value_list)

    @connect_execute_commit_close_db
    def edit_prompt_list(self, prompt_id, model_id, model_para, messages, variables, opening_remarks, update_by,
                         connection, cursor):
        sql = """update t_prompt_list set f_model_id=%s,f_model_para=%s,f_messages=%s,f_variables=%s,
                            f_opening_remarks=%s,f_update_by=%s,f_update_time=%s 
                            where f_prompt_id=%s"""
        value_list = [model_id, model_para, messages, variables, opening_remarks, update_by, datetime.datetime.today(),
                      prompt_id]
        
        cursor.execute(sql, value_list)

    @connect_execute_commit_close_db
    def deploy_prompt(self, prompt_id, service_id, connection, cursor):
        sql = """update t_prompt_list set f_is_deploy=1,
                            f_prompt_deploy_url='/api/model-factory/v1/prompt/{}/used' 
                            where f_prompt_id='{}'""".format(service_id, prompt_id)
        
        cursor.execute(sql)

    @connect_execute_commit_close_db
    def undeploy_prompt(self, prompt_id, connection, cursor):
        sql = """update t_prompt_list set f_is_deploy=0,
                                    f_prompt_deploy_url='' 
                                    where f_prompt_id='{}'""".format(prompt_id)
        
        cursor.execute(sql)

    @connect_execute_close_db
    def get_all_from_prompt_item_list(self, is_management, connection, cursor):
        condition = ""
        if is_management is not None:
            condition += "where f_is_management={}".format(is_management)

        sql = """select f_create_by,f_create_time,f_id,f_item_is_delete,f_prompt_item_id,f_prompt_item_name,f_built_in,
                            f_prompt_item_type,f_prompt_item_type_id,f_type_is_delete,f_update_by,f_update_time 
                            from t_prompt_item_list """ + condition
        
        cursor.execute(sql)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def delete_from_prompt_list_by_prompt_id(self, prompt_id, connection, cursor):
        sql = """delete from t_prompt_list where f_prompt_id='{}'""".format(prompt_id)
        
        cursor.execute(sql)

    @connect_execute_commit_close_db
    def delete_from_prompt_list_by_type_id(self, prompt_item_type_id, connection, cursor):
        sql = """delete from t_prompt_list where f_prompt_item_type_id='{}'""".format(prompt_item_type_id)
        
        cursor.execute(sql)

    @connect_execute_commit_close_db
    def delete_from_prompt_item_list_by_type_id(self, prompt_item_type_id, connection, cursor):
        sql = """update t_prompt_item_list set f_type_is_delete=1 where f_prompt_item_type_id='{}'""".format(
            prompt_item_type_id)
        
        cursor.execute(sql)

    @connect_execute_commit_close_db
    def delete_from_prompt_list_by_item_id(self, prompt_item_id, connection, cursor):
        sql = """delete from t_prompt_list where f_prompt_item_id='{}'""".format(prompt_item_id)
        
        cursor.execute(sql)

    @connect_execute_commit_close_db
    def delete_from_prompt_item_list_by_item_id(self, prompt_item_id, connection, cursor):
        sql = """delete from t_prompt_item_list where f_prompt_item_id='{}'""".format(prompt_item_id)
        
        cursor.execute(sql)

    @connect_execute_commit_close_db
    def move_prompt(self, prompt_id, prompt_name, type_id, item_id, update_by, connection, cursor):
        sql = """update t_prompt_list set f_prompt_name=%s,f_prompt_item_type_id=%s,f_prompt_item_id=%s,
                            f_update_by=%s,f_update_time=%s 
                            where f_prompt_id=%s"""
        value_list = [prompt_name, type_id, item_id, update_by, datetime.datetime.today(), prompt_id]
        
        cursor.execute(sql, value_list)

    @connect_execute_commit_close_db
    def edit_template_prompt_list(self, prompt_id, prompt_name, messages, variables, opening_remarks, icon, prompt_desc,
                                  prompt_item_type_id, prompt_item_id, connection, cursor):
        sql = """update t_prompt_list set f_prompt_name=%s, f_messages=%s, f_variables=%s, f_opening_remarks=%s,
                    f_icon=%s, f_prompt_desc=%s, f_prompt_item_type_id=%s, f_prompt_item_id=%s, f_update_time=%s 
                    where f_prompt_id=%s;"""
        value_list = [prompt_name, messages, variables, opening_remarks, icon, prompt_desc, prompt_item_type_id,
                      prompt_item_id, datetime.datetime.today(), prompt_id]
        
        cursor.execute(sql, value_list)

    @connect_execute_close_db
    def check_item_and_item_type(self, item_id, item_type_id, connection, cursor):
        sql = """select f_id from t_prompt_item_list 
                    where f_prompt_item_id=%s and f_prompt_item_type_id=%s and f_type_is_delete=0;"""
        value_list = [item_id, item_type_id]
        
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def check_item_and_item_type_by_name(self, item_name, item_type_name, connection, cursor):
        sql = """select f_prompt_item_id, f_prompt_item_type_id from t_prompt_item_list 
                    where f_prompt_item_name=%s and f_prompt_item_type=%s and f_type_is_delete=0;"""
        value_list = [item_name, item_type_name]
        
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    @connect_execute_close_db
    def check_item_by_name(self, item_name, connection, cursor):
        sql = """select f_prompt_item_id from t_prompt_item_list where f_prompt_item_name=%s and f_type_is_delete=0;"""
        value_list = [item_name]
        
        cursor.execute(sql, value_list)
        res = cursor.fetchall()
        return res

    @connect_execute_commit_close_db
    def add_prompt_batch(self, values, connection, cursor):
        sql = """insert into t_prompt_list(f_prompt_id,f_prompt_item_id,f_prompt_item_type_id,f_prompt_type,
                            f_prompt_service_id,f_prompt_name,f_prompt_desc,f_icon,f_variables,f_model_id,f_model_para,
                            f_opening_remarks,f_create_by,f_update_by,f_messages,f_update_time,f_create_time) values 
                            (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"""
        par = """ , (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) """
        for i in range(len(values) - 1):
            sql += par
        value_list = []
        for value in values:
            value_list += value
        
        cursor.execute(sql, value_list)

    # 根据prompt_id_list删除t_prompt_list表中对应的prompt
    @connect_execute_commit_close_db
    def delete_prompt_by_id_list(self, prompt_id_list, connection, cursor):
        sql = "delete from t_prompt_list where f_prompt_id in ({})".format(",".join(prompt_id_list))
        
        cursor.execute(sql)


prompt_dao = PromptDao()
