import datetime
import time
import asyncio
import ast

import func_timeout
import tiktoken

from fastapi.responses import JSONResponse
from app.commons.errors.codes import *
from app.commons.snow_id import snow_id
from app.commons.snow_id import worker
from app.controller.llm_controller import used_model_stream
from app.logs.stand_log import StandLogger
from app.utils import llm_utils
from app.utils.comment_utils import error_log
from app.utils.llm_utils import model_config, get_context_size
from app.utils.param_verify_utils import *
from app.utils.reshape_utils import *
from app.dao.prompt_dao import prompt_dao
from app.dao.llm_model_dao import llm_model_dao


# 获取prompt项目列表
async def source_prompt_item_endpoint(request, prompt_item_name, prompt_name):
    headers = request.headers
    content = prompt_source_item_verify(prompt_item_name)
    if content:
        StandLogger.error(content)
        return JSONResponse(status_code=400, content=content)
    try:
        item_id_list = []
        data = []
        id_name = {}
        item_type_list = prompt_dao.get_all_from_prompt_item_list(None)
        user_ids = await get_userid_by_search(item_type_list)
        user_infos = await get_username_by_ids(user_ids)
        if prompt_name == '':
            for item in item_type_list:
                if item["f_prompt_item_id"] in set(item_id_list):
                    continue
                item_id_list.append(item["f_prompt_item_id"])
                data_item = {
                    "prompt_item_id": item["f_prompt_item_id"],
                    "prompt_item_name": item["f_prompt_item_name"],
                    "prompt_item_types": [],
                    "create_time": item["f_create_time"].strftime('%Y-%m-%d %H:%M:%S'),
                    "create_by": item["f_create_by"],
                    "update_time": item["f_update_time"].strftime('%Y-%m-%d %H:%M:%S'),
                    "update_by": item["f_update_by"],
                    "is_built_in": True if item["f_built_in"] == 1 else False
                }
                prompt_item_types = []
                for i in item_type_list:
                    if i["f_prompt_item_id"] == item["f_prompt_item_id"] and i["f_type_is_delete"] == 0 and i[
                        "f_prompt_item_type"] != "":
                        prompt_item_types.append(i)
                for item_type in prompt_item_types:
                    data_item_type = {
                        "id": item_type["f_prompt_item_type_id"],
                        "name": item_type["f_prompt_item_type"],
                        "prompt_info": []
                    }
                    data_item["prompt_item_types"].append(data_item_type)
                if data_item["create_by"] not in id_name:
                    id_name[data_item["create_by"]] = user_infos.get(data_item["create_by"], "")
                if data_item["update_by"] not in id_name:
                    id_name[data_item["update_by"]] = user_infos.get(data_item["update_by"], "")
                data_item["create_by"] = id_name[data_item["create_by"]]
                data_item["update_by"] = id_name[data_item["update_by"]]
                data.append(data_item)
        else:
            fuzzy_item_type_list = prompt_dao.get_prompt_item_list_by_prompt_name_fuzzy(prompt_name)
            user_ids = await get_userid_by_search(fuzzy_item_type_list)
            user_infos = await get_username_by_ids(user_ids)
            for item in fuzzy_item_type_list:
                # if item["f_prompt_item_id"] in set(item_id_list):
                #     continue
                item_id_list.append(item["f_prompt_item_id"])
                data_item = {
                    "prompt_item_id": item["f_prompt_item_id"],
                    "prompt_item_name": item["f_prompt_item_name"],
                    "prompt_item_types": [],
                    "create_time": item["f_create_time"].strftime('%Y-%m-%d %H:%M:%S'),
                    "create_by": item["f_create_by"],
                    "update_time": item["f_update_time"].strftime('%Y-%m-%d %H:%M:%S'),
                    "update_by": item["f_update_by"],
                    "is_built_in": True if item["f_built_in"] == 1 else False,
                }
                prompt_item_types = []
                for i in item_type_list:
                    if i["f_prompt_item_id"] == item["f_prompt_item_id"] and i["f_type_is_delete"] == 0:
                        prompt_item_types.append(i)
                for item_type in prompt_item_types:
                    # 先收集该item_type下的所有prompt记录
                    prompt_infos = []
                    for prompt_item in fuzzy_item_type_list:
                        if prompt_item['f_prompt_item_type_id'] == item_type['f_prompt_item_type_id']:
                            prompt_infos.append({
                                "prompt_id": prompt_item["f_prompt_id"],
                                "prompt_name": prompt_item["f_prompt_name"],
                                "prompt_desc": prompt_item["f_prompt_desc"],
                                "messages": prompt_item["f_messages"],
                                "variables": json.loads(prompt_item["f_variables"]),
                                "icon": prompt_item["f_icon"],
                                "prompt_type": prompt_item["f_prompt_type"],
                                "create_time": prompt_item["prompt_create_time"].strftime('%Y-%m-%d %H:%M:%S') if
                                prompt_item["prompt_create_time"] else "",
                                "update_time": prompt_item["prompt_update_time"].strftime('%Y-%m-%d %H:%M:%S') if
                                prompt_item["prompt_update_time"] else "",
                                "create_by": user_infos.get(prompt_item["prompt_create_by"], ""),
                                "update_by": user_infos.get(prompt_item["prompt_update_by"], ""),
                                "is_built_in": user_infos.get(prompt_item["prompt_built_in"], ""),
                                "prompt_item_id": prompt_item["f_prompt_item_id"],
                                "prompt_item_name": prompt_item["f_prompt_item_name"],
                                "prompt_item_type": prompt_item["f_prompt_item_type"],
                                "prompt_item_type_id": prompt_item["f_prompt_item_id"]

                            })
                    if prompt_infos:
                        data_item_type = {
                            "id": item_type["f_prompt_item_type_id"],
                            "name": item_type["f_prompt_item_type"],
                            "prompt_info": prompt_infos  # 将收集的所有prompt记录放入数组
                        }
                        data_item["prompt_item_types"].append(data_item_type)
                if data_item["create_by"] not in id_name:
                    id_name[data_item["create_by"]] = user_infos.get(data_item["create_by"], "")
                if data_item["update_by"] not in id_name:
                    id_name[data_item["update_by"]] = user_infos.get(data_item["update_by"], "")
                data_item["create_by"] = id_name[data_item["create_by"]]
                data_item["update_by"] = id_name[data_item["update_by"]]
                data.append(data_item)
        res = {
            "res": {
                "total": len(data),
                "searchTotal": len(data),
                "data": data
            }
        }
        return JSONResponse(status_code=200, content=res)
    except Exception as e:
        print(e)
        StandLogger.error(e.args)
        return JSONResponse(status_code=500, content=DataBaseError)


# 获取提示词列表信息
async def source_prompt_endpoint(request, prompt_item_id, prompt_item_type_id,
                                 page, size, prompt_name, order, rule, deploy, prompt_type):
    headers = request.headers
    error = await prompt_source_verify(prompt_item_id, prompt_item_type_id,
                                       page, size, prompt_name, order, rule, deploy, prompt_type)
    if error:
        StandLogger.error(error)
        return JSONResponse(status_code=400, content=error)
    try:
        order_list, rule_dict = key_value()
        if deploy == 'all' or not deploy:
            deploy = ''
        elif deploy == "yes":
            deploy = 1
        else:
            deploy = 0
        if prompt_type == 'all':
            prompt_type = ''
        # prompt_item_type_id_list = [cell["f_prompt_item_type_id"] for cell in prompt_dao.get_all_from_prompt_item_list()]
        # if prompt_item_type_id not in prompt_item_type_id_list and prompt_item_type_id:
        #     return JSONResponse(status_code=200, content={"res": {'total': 0, 'data': []}})
        prompts = prompt_dao.get_prompt_list(prompt_item_id, prompt_item_type_id,
                                             page, size, prompt_name, order, rule, deploy, prompt_type)
        models = llm_model_dao.get_model_default_paras()
        data = []
        id_name = {}
        user_ids = await get_userid_by_search(prompts)
        user_infos = await get_username_by_ids(user_ids)
        for prompt in prompts:
            data_prompt = {}
            data_prompt["prompt_item_id"] = prompt["f_prompt_item_id"]
            data_prompt["prompt_item_name"] = prompt["f_prompt_item_name"]
            data_prompt["prompt_item_type_id"] = prompt["f_prompt_item_type_id"]
            data_prompt["prompt_item_type"] = prompt["f_prompt_item_type"]
            data_prompt["prompt_id"] = prompt["f_prompt_id"]
            data_prompt["prompt_service_id"] = prompt["f_prompt_id"]
            data_prompt["prompt_name"] = prompt["f_prompt_name"]
            if prompt["f_model_id"]:
                data_prompt["model_name"] = models[prompt["f_model_id"]]["model_name"]
                data_prompt["model_series"] = models[prompt["f_model_id"]]["model_series"]
            else:
                data_prompt["model_name"] = ""
                data_prompt["model_series"] = ""
            data_prompt["prompt_type"] = prompt["f_prompt_type"]
            data_prompt["prompt_desc"] = prompt["f_prompt_desc"]
            data_prompt["prompt_deploy"] = False if prompt["f_is_deploy"] == 0 else True
            data_prompt["icon"] = prompt["f_icon"]
            data_prompt["model_id"] = prompt["f_model_id"]
            data_prompt["create_by"] = prompt["f_create_by"]
            data_prompt["update_by"] = prompt["f_update_by"]
            data_prompt["is_built_in"] = True if prompt["f_built_in"] == 1 else False
            data_prompt["messages"] = prompt["f_messages"]
            data_prompt["variables"] = json.loads(prompt["f_variables"])
            data_prompt["model_para"] = json.loads(prompt["f_model_para"].replace("'", '"'))
            if prompt["f_create_time"] == None:
                data_prompt["create_time"] = ""
            else:
                data_prompt["create_time"] = prompt["f_create_time"].strftime('%Y-%m-%d %H:%M:%S')
            if prompt["f_update_time"] == None:
                data_prompt["update_time"] = ""
            else:
                data_prompt["update_time"] = prompt["f_update_time"].strftime('%Y-%m-%d %H:%M:%S')
            if data_prompt["create_by"] not in id_name:
                id_name[data_prompt["create_by"]] = user_infos.get(data_prompt["create_by"], "")
            if data_prompt["update_by"] not in id_name:
                id_name[data_prompt["update_by"]] = user_infos.get(data_prompt["update_by"], "")
            data_prompt["create_by"] = id_name[data_prompt["create_by"]]
            data_prompt["update_by"] = id_name[data_prompt["update_by"]]
            data.append(data_prompt)
        res = {
            "res": {
                "total": len(prompts),
                "data": data
            }
        }
        return JSONResponse(status_code=200, content=res)
    except Exception as e:
        print(e)
        StandLogger.error(e.args)
        return JSONResponse(status_code=500, content=DataBaseError)


# 获取所有提示词的id和name
async def prompt_list_endpoint():
    all_prompt = prompt_dao.get_all_data_from_prompt_list()
    return {prompt['f_prompt_id']: prompt['f_prompt_name'] for prompt in all_prompt}


# 获取提示词模板列表信息
async def template_source_prompt_endpoint(request, prompt_type, prompt_name):
    headers = request.headers
    error = prompt_template_verify(prompt_name, prompt_type)
    if error:
        StandLogger.error(error)
        return JSONResponse(status_code=400, content=error)
    try:
        templates = prompt_dao.get_prompt_template_list(prompt_type, prompt_name)
        data = []
        for template in templates:
            data_template = {
                "icon": template["f_icon"],
                "prompt_id": template["f_prompt_id"],
                "messages": template["f_messages"],
                "prompt_name": template["f_prompt_name"],
                "prompt_type": template["f_prompt_type"],
                "prompt_desc": template["f_prompt_desc"],
                "opening_remarks": template["f_opening_remarks"],
                "input": template["f_input"],
                "variables": json.loads(template["f_variables"])
            }
            data.append(data_template)
        res = {
            "res": {
                "total": len(data),
                "data": data
            }
        }
        return JSONResponse(status_code=200, content=res)
    except Exception as e:
        print(e)
        StandLogger.error(e.args)
        return JSONResponse(status_code=500, content=DataBaseError)


# 提示词查看
async def check_prompt_endpoint(request, prompt_id):
    headers = request.headers
    error = await check_prompt_verify(prompt_id)
    if error:
        StandLogger.error(error)
        return JSONResponse(status_code=200, content=error)
    try:
        prompt = prompt_dao.get_prompt_by_id(prompt_id)[0]
        model_id = prompt["f_model_id"]
        model_name = llm_model_dao.get_model_name_by_id(prompt["f_model_id"]) if model_id else ""
        model_series = llm_model_dao.get_model_series_by_id(prompt["f_model_id"]) if model_id else ""
        res = {
            "res": {
                "prompt_id": prompt["f_prompt_id"],
                "prompt_name": prompt["f_prompt_name"],
                "model_id": model_id,
                "model_name": model_name,
                "prompt_item_id": prompt["f_prompt_item_id"],
                "prompt_service_id": prompt["f_prompt_service_id"],
                "prompt_item_name": prompt_dao.get_prompt_item_name_by_id(prompt["f_prompt_item_id"]),
                "prompt_item_type_id": prompt["f_prompt_item_type_id"],
                "prompt_item_type": prompt_dao.get_prompt_item_type_by_type_id(prompt["f_prompt_item_type_id"]),
                "messages": prompt["f_messages"],
                "opening_remarks": prompt["f_opening_remarks"],
                "variables": json.loads(prompt["f_variables"]),
                "prompt_type": prompt["f_prompt_type"],
                "prompt_desc": prompt["f_prompt_desc"],
                "prompt_deploy": False if prompt["f_is_deploy"] == 0 else True,
                "model_series": model_series,
                "icon": prompt["f_icon"],
                "model_para": eval(prompt["f_model_para"])
            }
        }
        return JSONResponse(status_code=200, content=res)
    except Exception as e:
        print(e)
        StandLogger.error(e.args)
        return JSONResponse(status_code=500, content=DataBaseError)


# 新建提示词项目
async def add_prompt_item_endpoint(userId, params):
    error = await item_add_verify(params)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            f_id = worker.get_id()
            f_prompt_item_id = worker.get_id()
            f_prompt_item_type_id = worker.get_id()
            try:
                prompt_dao.add_prompt_item(f_id, f_prompt_item_id, f_prompt_item_type_id, '默认分组',
                                           userId, userId, params['prompt_item_name'])
                return JSONResponse(status_code=200, content={"res": str(f_prompt_item_id)})
            except Exception as e:
                print(e)
                PromptItemAddError2['description'] = "提示词项目名称重复"
                PromptItemAddError2['detail'] = "prompt_item_name参数不符合规范"
                return JSONResponse(status_code=500, content=PromptItemAddError2)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 编辑提示词项目
async def edit_prompt_item_endpoint(userId, model_para):
    error = await item_edit_verify(model_para)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            prompt_dao.edit_prompt_item(model_para["prompt_item_id"], model_para["prompt_item_name"], userId)
            return JSONResponse(status_code=200, content={"res": True})
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 新建提示词分类
async def add_prompt_type_endpoint(userId, model_para):
    error = await type_add_verify(model_para)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            f_prompt_item_type_id = snow_id()
            time.sleep(0.01)
            f_id = snow_id()
            item_name = prompt_dao.get_prompt_item_name_by_id(model_para["prompt_item_id"])
            prompt_dao.add_prompt_item(f_id, model_para["prompt_item_id"], f_prompt_item_type_id,
                                       model_para["prompt_item_type"], userId, userId, item_name)
            return JSONResponse(status_code=200, content={"res": str(f_prompt_item_type_id)})
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 编辑提示词分类
async def edit_prompt_type_endpoint(userId, model_para):
    error = await type_edit_verify(model_para)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            prompt_dao.edit_prompt_item_type(model_para["prompt_item_type_id"], model_para["prompt_item_type"], userId)
            return JSONResponse(status_code=200, content={"res": True})
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 新增提示词
async def add_prompt_endpoint(userId, model_para):
    if "prompt_desc" not in model_para:
        model_para["prompt_desc"] = ''
    if "variables" not in model_para:
        model_para["variables"] = []
    if "opening_remarks" not in model_para:
        model_para["opening_remarks"] = ''
    if "model_id" not in model_para:
        model_para["model_id"] = ""
    if "model_para" not in model_para:
        model_para["model_para"] = {}
    error = await prompt_add_verify(model_para)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            if len(model_para["model_para"]) != 0:
                model_para["model_para"]["temperature"] = round(model_para["model_para"]["temperature"], 2)
                model_para["model_para"]["top_p"] = round(model_para["model_para"]["top_p"], 2)
                model_para["model_para"]["presence_penalty"] = round(model_para["model_para"]["presence_penalty"], 2)
                model_para["model_para"]["frequency_penalty"] = round(model_para["model_para"]["frequency_penalty"], 2)
            idx = snow_id()
            time.sleep(0.01)
            prompt_service_id = snow_id()
            prompt_dao.add_prompt_to_prompt_list(idx, model_para["prompt_item_id"], model_para["prompt_item_type_id"],
                                                 model_para["prompt_type"], prompt_service_id,
                                                 model_para["prompt_name"], model_para["prompt_desc"],
                                                 model_para["icon"], json.dumps(model_para["variables"]),
                                                 model_para["model_id"], json.dumps(model_para["model_para"]),
                                                 model_para["opening_remarks"], userId, userId, model_para["messages"])
            return JSONResponse(status_code=200, content={"res": {"prompt_id": str(idx)}})
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 提示词名称编辑
async def name_edit_prompt_endpoint(userId, model_para):
    if "prompt_desc" not in model_para:
        model_para["prompt_desc"] = ''
    if "model_id" not in model_para:
        model_para["model_id"] = ""
    error = await prompt_name_verify(model_para)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            # 获取原来的提示词分类id
            info = prompt_dao.get_prompt_by_id(model_para["prompt_id"])
            type_id = info[0]["f_prompt_item_type_id"]
            # 如果两次的分类ID不一样
            if model_para['prompt_item_type_id'] != type_id:
                # 去查询第二次分类的所有prompt名称
                info = prompt_dao.get_data_from_prompt_list_by_item_type_id(model_para['prompt_item_type_id'])
                prompt_name_in_new_type = [cell["f_prompt_name"] for cell in info]
                # 如果名称重复, 进入循环阶段，为名称加后缀，直到不重复
                if model_para['prompt_name'] in prompt_name_in_new_type:
                    num = 1
                    while True:
                        new_name = model_para['prompt_name'] + "_{}".format(num)
                        if new_name not in prompt_name_in_new_type:
                            break
                        num += 1
                    model_para['prompt_name'] = new_name
            prompt_dao.edit_name_in_prompt_list(model_para["prompt_id"], model_para["prompt_name"],
                                                model_para["model_id"],
                                                model_para["icon"], model_para["prompt_desc"], userId,
                                                model_para["prompt_item_id"], model_para["prompt_item_type_id"])
            return JSONResponse(status_code=200, content={"res": True})
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 提示词编辑
async def edit_prompt_endpoint(userId, model_para):
    if "variables" not in model_para:
        model_para["variables"] = []
    if "opening_remarks" not in model_para:
        model_para["opening_remarks"] = ''
    if "model_id" not in model_para:
        model_para["model_id"] = ""
    if "model_para" not in model_para:
        model_para["model_para"] = {}
    error = await prompt_edit_verify(model_para)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            prompt_dao.edit_prompt_list(model_para["prompt_id"], model_para["model_id"],
                                        json.dumps(model_para["model_para"]), model_para["messages"],
                                        json.dumps(model_para["variables"]), model_para["opening_remarks"],
                                        userId)
            return JSONResponse(status_code=200, content={"res": True})
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 提示词管理中编辑提示词接口
async def edit_template_prompt_endpoint(userId, params):
    if "variables" not in params:
        params["variables"] = []
    if "opening_remarks" not in params:
        params["opening_remarks"] = ''
    if "prompt_desc" not in params:
        params["prompt_desc"] = ''
    error = await prompt_template_edit_verify(params)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    try:
        # 获取原来的提示词分类id
        info = prompt_dao.get_prompt_by_id(params["prompt_id"])
        type_id = info[0]["f_prompt_item_type_id"]

        # 如果两次的分类ID一样
        if params['prompt_item_type_id'] == type_id:
            prompt_name_list = [ids["f_prompt_name"] for ids in
                                prompt_dao.get_data_from_prompt_list_by_item_type_id(info[0]["f_prompt_item_type_id"])]
            if params["prompt_name"] in prompt_name_list and params["prompt_name"] != info[0]["f_prompt_name"]:
                PromptNameEditError2['description'] = "参数错误"
                PromptNameEditError2['detail'] = "提示词名称重复"
                StandLogger.error(PromptNameEditError2)
                return JSONResponse(status_code=500, content=PromptNameEditError2)
        else:
            # 去查询第二次分类的所有prompt名称
            info = prompt_dao.get_data_from_prompt_list_by_item_type_id(params['prompt_item_type_id'])
            prompt_name_in_new_type = [cell["f_prompt_name"] for cell in info]
            # 如果名称重复, 进入循环阶段，为名称加后缀，直到不重复
            if params['prompt_name'] in prompt_name_in_new_type:
                num = 1
                while True:
                    new_name = params['prompt_name'] + "_{}".format(num)
                    if new_name not in prompt_name_in_new_type:
                        break
                    num += 1
                params['prompt_name'] = new_name
        prompt_dao.edit_template_prompt_list(params["prompt_id"], params["prompt_name"], params["messages"],
                                             json.dumps(params["variables"]), params["opening_remarks"],
                                             params["icon"], params["prompt_desc"], params["prompt_item_type_id"],
                                             params["prompt_item_id"])
        return JSONResponse(status_code=200, content={"res": True})
    except Exception as e:
        StandLogger.error(e.args)
        return JSONResponse(status_code=500, content=DataBaseError)


# 提示词发布
async def deploy_prompt_endpoint(userId, model_para):
    error = await prompt_deploy_verify(model_para)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            info = prompt_dao.get_prompt_by_id(model_para["prompt_id"])
            service_id = info[0]["f_prompt_service_id"]
            prompt_dao.deploy_prompt(model_para["prompt_id"], service_id)
            return JSONResponse(status_code=200, content={"res": True})
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 提示词取消发布
async def undeploy_prompt_endpoint(userId, model_para):
    error = await prompt_undeploy_verify(model_para)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            prompt_dao.undeploy_prompt(model_para['prompt_id'])
            return JSONResponse(status_code=200, content={"res": True})
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 填充提示词
async def completion_prompt_endpoint(userId, prompt_id, inputs):
    if type(inputs) is str:
        inputs = json.loads(inputs.replace("'", '"'))
    error = await completion_prompt_verify(prompt_id, inputs)
    if error:
        StandLogger.error(error)
        return JSONResponse(status_code=400, content=error)
    else:
        try:
            info = prompt_dao.get_prompt_by_id(prompt_id)
            messages = info[0]["f_messages"]
            messages = messages.replace("{{", "{").replace('}}', "}").format(**inputs)
            return JSONResponse(status_code=200, content=messages)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 查看代码
async def code_prompt_endpoint(userId, model_id, prompt_id):
    error = await prompt_code_verify(model_id, prompt_id)
    if error:
        StandLogger.error(error)
        return JSONResponse(status_code=400, content=error)
    try:
        if len(prompt_id.replace(' ', '')) == 0:
            result_model = llm_model_dao.get_data_from_model_list_by_id(model_id)
            result = {
                "res": {
                    "model_series": result_model[0]["f_model_series"],
                    "model_type": result_model[0]["f_model_type"],
                    "model_config": json.loads((result_model[0]["f_model_config"]).replace("'", '"')),
                    "prompt_deploy_url": ""
                }
            }
            return JSONResponse(status_code=200, content=result)
        else:
            result_prompt = prompt_dao.get_prompt_by_id(prompt_id)
            result_model = llm_model_dao.get_data_from_model_list_by_id(model_id)
            result = {
                "res": {
                    "model_series": result_model[0]["f_model_series"],
                    "model_type": result_model[0]["f_model_type"],
                    "model_config": json.loads((result_model[0]["f_model_config"]).replace("'", '"')),
                    "prompt_deploy_url": ""
                }
            }
            if result_prompt[0]["f_is_deploy"] == 1:
                result["res"]["prompt_deploy_url"] = result_prompt[0]["f_prompt_deploy_url"]
            return JSONResponse(status_code=200, content=result)
    except Exception as e:
        StandLogger.error(e.args)
        print(e)
        return JSONResponse(status_code=500, content=DataBaseError)


# 获取api文档接口
async def api_doc_prompt_endpoint(prompt_service_id):
    from app.commons.restful_api import get_prompt_restful_api_document
    info = prompt_dao.get_prompt_by_service_id(prompt_service_id)
    prompt = info[0]["f_messages"]
    var = json.loads(info[0]["f_variables"].replace("'", '"'))
    if var:
        var = [cell['var_name'] for cell in var]
        var_dict = dict()
        for cell in var:
            var_dict[cell] = ''
    else:
        var_dict = ''
    res = get_prompt_restful_api_document(prompt_service_id, var_dict, prompt)
    return JSONResponse(status_code=200, content={'res': res})


# 删除功能函数
async def delete_prompt_endpoint(userId, delete_id):
    error = await prompt_delete_verify(delete_id)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    else:
        try:
            if "prompt_id" in delete_id:
                prompt_dao.delete_from_prompt_list_by_prompt_id(delete_id["prompt_id"])
                return JSONResponse(status_code=200, content={"res": True})
            elif "type_id" in delete_id:
                prompt_dao.delete_from_prompt_list_by_type_id(delete_id["type_id"])
                info = prompt_dao.get_data_from_prompt_item_list_by_type_id(delete_id["type_id"])
                info_is_end = prompt_dao.get_prompt_item_type_by_item_id(info[0]["f_prompt_item_id"])
                if len(info) == 1:
                    item_name = info[0]["f_prompt_item_name"]
                    item_id = info[0]["f_prompt_item_id"]
                    create_by = info[0]["f_create_by"]
                    create_time = info[0]["f_create_time"]
                prompt_dao.delete_from_prompt_item_list_by_type_id(delete_id["type_id"])
                if len(info) == 1:
                    if len(info_is_end) == 1:
                        prompt_dao.add_prompt_item(snow_id(), item_id, '', '',
                                                   create_by, userId, item_name)
                return JSONResponse(status_code=200, content={"res": True})
            elif "item_id" in delete_id:
                prompt_dao.delete_from_prompt_list_by_item_id(delete_id["item_id"])
                prompt_dao.delete_from_prompt_item_list_by_item_id(delete_id["item_id"])
                return JSONResponse(status_code=200, content={"res": True})
            elif "prompt_id_list" in delete_id:
                prompt_dao.delete_prompt_by_id_list(delete_id["prompt_id_list"])
                return JSONResponse(status_code=200, content={"res": True})
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=DataBaseError)


# 获取id模块
async def get_id_endpoint(userId):
    return {"res": str(snow_id())}


# 提示词移动
async def move_prompt_endpoint(userId, move_param):
    error = await prompt_move_verify(move_param)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    try:
        prompt_id = move_param['prompt_id']
        new_item_id = move_param['prompt_item_id']
        new_type_id = move_param['prompt_item_type_id']
        # 获取原来的 ID 和 name
        info = prompt_dao.get_prompt_by_id(prompt_id)
        old_type_id = info[0]["f_prompt_item_type_id"]
        old_name = info[0]["f_prompt_name"]
        # 对重名的进行改名
        if new_type_id != old_type_id:
            info = prompt_dao.get_data_from_prompt_list_by_item_type_id(new_type_id)
            prompt_name_in_new_type = [cell["f_prompt_name"] for cell in info]
            if old_name in prompt_name_in_new_type:
                num = 1
                while True:
                    new_name = old_name + "_{}".format(num)
                    if new_name not in prompt_name_in_new_type:
                        break
                    num += 1
                old_name = new_name
        # 写数据
        prompt_dao.move_prompt(prompt_id, old_name, new_type_id, new_item_id, userId)
        return JSONResponse(status_code=200, content={"res": True})
    except Exception as e:
        StandLogger.error(e.args)
        return JSONResponse(status_code=500, content=DataBaseError)


async def batch_add_prompt_endpoint(userId, params_list):
    error = await batch_add_prompt_endpoint_verify(params_list)
    if error:
        StandLogger.error(error[1])
        return JSONResponse(status_code=error[0], content=error[1])
    try:
        res = []
        for params in params_list:

            prompt_item_name = params["prompt_item_name"]
            prompt_item_type_name = params["prompt_item_type_name"]
            prompt_item = prompt_dao.check_item_and_item_type_by_name(prompt_item_name, prompt_item_type_name)
            if len(prompt_item) == 0:
                items = prompt_dao.check_item_by_name(prompt_item_name)
                if len(items) > 0:
                    prompt_item_id = items[0]["f_prompt_item_id"]
                else:
                    prompt_item_id = worker.get_id()
                f_id = worker.get_id()
                prompt_item_type_id = worker.get_id()
                prompt_dao.add_prompt_item(f_id, prompt_item_id, prompt_item_type_id, prompt_item_type_name,
                                           userId, userId, params['prompt_item_name'])
            else:
                prompt_item_id = prompt_item[0]["f_prompt_item_id"]
                prompt_item_type_id = prompt_item[0]["f_prompt_item_type_id"]
            prompt_name_id_dict = {prompt["f_prompt_name"]: prompt["f_prompt_id"] for prompt in
                                   prompt_dao.get_data_from_prompt_list_by_item_type_id(prompt_item_type_id)}

            prompt_id_list = {}
            insert_values = []
            for prompt in params["prompt_list"]:
                if not isinstance(prompt["messages"], str):
                    error_dict = DataBaseError.copy()
                    error_dict["description"] = error_dict["detail"] = error_dict["solution"] = "messages必须是字符串类型"
                    return JSONResponse(status_code=500, content=error_dict)
                prompt_name = prompt["prompt_name"]
                if prompt_name in prompt_name_id_dict.keys():
                    prompt_id = prompt_name_id_dict[prompt_name]
                    prompt_dao.edit_template_prompt_list(prompt_id, prompt_name, prompt["messages"],
                                                         json.dumps(prompt.get("variables", [])),
                                                         prompt.get("opening_remarks", ""),
                                                         prompt["icon"], prompt.get("prompt_desc", ""),
                                                         prompt_item_type_id, prompt_item_id)
                else:
                    prompt_id = snow_id()
                    time.sleep(0.01)
                    prompt_service_id = snow_id()
                    insert_values.append([prompt_id, prompt_item_id, prompt_item_type_id, prompt["prompt_type"],
                                          prompt_service_id, prompt_name, prompt.get("prompt_desc", ""),
                                          prompt["icon"], json.dumps(prompt.get("variables", [])),
                                          prompt.get("model_id", ""), json.dumps(prompt.get("model_para", {})),
                                          prompt.get("opening_remarks", ""), userId, userId, prompt["messages"],
                                          datetime.datetime.today(), datetime.datetime.today()])
                prompt_id_list[prompt_name] = str(prompt_id)
            if len(insert_values) > 0:
                prompt_dao.add_prompt_batch(insert_values)
            res.append({"prompt_item_name": prompt_item_name,
                        "prompt_item_type_name": prompt_item_type_name,
                        "prompt_list": prompt_id_list})
        return JSONResponse(status_code=200, content={"res": res})
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = DataBaseError.copy()

        error_dict["description"] = "发生了未知错误"
        error_dict["detail"] = str(e.args[0])
        error_dict["solution"] = str(e.args[0])
        if "dict can not" in str(e.args[0]):
            error_dict["description"] = error_dict["detail"] = error_dict["solution"] = "messages必须是字符串类型"
        return JSONResponse(status_code=500, content=error_dict)


# async def encode_endpoint(params_json):
#     try:
#         if "text" not in params_json:
#             EncodeParaError["description"] = EncodeParaError["detail"] = "缺少参数text"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#         elif not isinstance(params_json["text"], str):
#             EncodeParaError["description"] = EncodeParaError["detail"] = "text必须为字符串类型"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#
#         if "model_name" not in params_json:
#             EncodeParaError["description"] = EncodeParaError["detail"] = "缺少参数model_name"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#         elif not isinstance(params_json["model_name"], str):
#             EncodeParaError["description"] = EncodeParaError["detail"] = "model_name必须为字符串类型"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#         elif params_json["model_name"] == "":
#             EncodeParaError["description"] = EncodeParaError["detail"] = "model_name不可以为空"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#
#         model_name = params_json["model_name"]
#         text = params_json["text"]
#         model_info = llm_utils.model_config.get_model_config("name", model_name)
#         if len(model_info) <= 0:
#             model_info = llm_model_dao.get_model_by_name(model_name)
#             if len(model_info) <= 0:
#                 EncodeParaError['description'] = EncodeParaError['detail'] = "模型名称不存在"
#                 return JSONResponse(status_code=500, content=EncodeParaError)
#             model_info = model_info[0]
#             llm_utils.model_config.add_model_config(model_info["f_model_id"], model_info["f_model_series"],
#                                                     model_info["f_model_type"], model_info["f_model_name"],
#                                                     model_info["f_model"], model_info["f_model_url"],
#                                                     model_info["f_model_config"])
#         token_ids, token_count = await llm_utils.encode(model_info["f_model_series"], text)
#         return JSONResponse(status_code=200, content={"res": {"token_ids": token_ids, "count": token_count}})
#     except Exception as e:
#         StandLogger.error(repr(e))
#         UnknownError["description"] = UnknownError["detail"] = repr(e)
#         return JSONResponse(status_code=500, content=UnknownError)


# async def encode_endpoint_v2(params_json):
#     try:
#         if "text" not in params_json:
#             EncodeParaError["description"] = EncodeParaError["detail"] = "缺少参数text"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#         elif not isinstance(params_json["text"], str):
#             EncodeParaError["description"] = EncodeParaError["detail"] = "text必须为字符串类型"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#
#         if "model_name" not in params_json:
#             EncodeParaError["description"] = EncodeParaError["detail"] = "缺少参数model_name"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#         elif not isinstance(params_json["model_name"], str):
#             EncodeParaError["description"] = EncodeParaError["detail"] = "model_name必须为字符串类型"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#         elif params_json["model_name"] == "":
#             EncodeParaError["description"] = EncodeParaError["detail"] = "model_name不可以为空"
#             return JSONResponse(status_code=400, content=EncodeParaError)
#
#         model_name = params_json["model_name"]
#         text = params_json["text"]
#         model_info = llm_utils.model_config.get_model_config("name", model_name)
#         if len(model_info) <= 0:
#             model_info = llm_model_dao.get_model_by_name(model_name)
#             if len(model_info) <= 0:
#                 EncodeParaError['description'] = EncodeParaError['detail'] = "模型名称不存在"
#                 return JSONResponse(status_code=500, content=EncodeParaError)
#             model_info = model_info[0]
#             llm_utils.model_config.add_model_config(model_info["f_model_id"], model_info["f_model_series"],
#                                                     model_info["f_model_type"], model_info["f_model_name"],
#                                                     model_info["f_model"], model_info["f_model_url"],
#                                                     model_info["f_model_config"])
#         config = json.loads(model_info["f_model_config"])
#         token_ids, token_count = await llm_utils.encode(model_info["f_model_series"], text, model_info["f_model"],
#                                                         config.get("api_key", ""), config.get("secret_key", ""))
#         return JSONResponse(status_code=200, content={"res": {"count": token_count}})
#     except Exception as e:
#         StandLogger.error(repr(e))
#         UnknownError["description"] = UnknownError["detail"] = repr(e)
#         return JSONResponse(status_code=500, content=UnknownError)


async def decode_endpoint(params_json):
    try:
        if "token_ids" not in params_json:
            DecodeParaError["description"] = DecodeParaError["detail"] = "缺少参数token_ids"
            return JSONResponse(status_code=400, content=DecodeParaError)
        elif not isinstance(params_json["token_ids"], list):
            DecodeParaError["description"] = DecodeParaError["detail"] = "token_ids必须为list类型"
            return JSONResponse(status_code=400, content=DecodeParaError)

        if "model_name" not in params_json:
            DecodeParaError["description"] = DecodeParaError["detail"] = "缺少参数model_name"
            return JSONResponse(status_code=400, content=DecodeParaError)
        elif not isinstance(params_json["model_name"], str):
            DecodeParaError["description"] = DecodeParaError["detail"] = "model_name必须为字符串类型"
            return JSONResponse(status_code=400, content=DecodeParaError)
        elif params_json["model_name"] == "":
            DecodeParaError["description"] = DecodeParaError["detail"] = "model_name不可以为空"
            return JSONResponse(status_code=400, content=DecodeParaError)

        model_name = params_json["model_name"]
        token_ids = params_json["token_ids"]
        model_info = llm_utils.model_config.get_model_config("name", model_name)
        if len(model_info) <= 0:
            model_info = llm_model_dao.get_model_by_name(model_name)
            if len(model_info) <= 0:
                DecodeParaError['description'] = DecodeParaError['detail'] = "模型名称不存在"
                return JSONResponse(status_code=500, content=DecodeParaError)
            model_info = model_info[0]
            llm_utils.model_config.add_model_config(model_info["f_model_id"], model_info["f_model_series"],
                                                    model_info["f_model_type"], model_info["f_model_name"],
                                                    model_info["f_model"], model_info["f_model_url"],
                                                    model_info["f_model_config"])
        if model_info["f_model_series"] == "openai":
            import os

            cache_key = "9b5ad71b2ce5302211f9c61530b329a4922fc6a4"
            tiktoken_cache_dir = os.path.dirname(os.path.dirname(os.path.realpath(__file__))) + "/utils/tiktoken_cache"
            os.environ["TIKTOKEN_CACHE_DIR"] = tiktoken_cache_dir

            assert os.path.exists(os.path.join(tiktoken_cache_dir, cache_key))

            encoding = tiktoken.get_encoding("cl100k_base")
            text = encoding.decode(token_ids)
            return JSONResponse(status_code=200, content={"res": {"text": text, "count": len(text)}})
        elif model_info["f_model_series"].lower() == "aishu":
            model_config = eval(model_info["f_model_config"])
            res = llm_utils.decode(model_config["api_base"], model_config["api_model"], token_ids)
            return JSONResponse(status_code=200, content={"res": res})
        else:
            StandLogger.error(DecodeError["description"])
            return JSONResponse(status_code=500, content=DecodeError)
    except Exception as e:
        StandLogger.error(repr(e))
        UnknownError["description"] = UnknownError["detail"] = repr(e)
        return JSONResponse(status_code=500, content=UnknownError)


# async def llm_generate(userId, language, req: LLMGenerateReq):
#     generation_prompts = {
#         'prompt': '1100000000000000038',  # 生成提示词
#         'function': '1100000000000000039'  # 生成代码
#     }
#     if not req.input:
#         AutoGenerationError1['description'] = "参数错误"
#         AutoGenerationError1['detail'] = "input不能为空"
#         return JSONResponse(status_code=400, content=AutoGenerationError1)
#     if req.type not in generation_prompts:
#         AutoGenerationError1['description'] = "参数错误"
#         AutoGenerationError1['detail'] = "type 只能为" + str(list(generation_prompts.keys()))
#         return JSONResponse(status_code=400, content=AutoGenerationError1)
#     prompt = prompt_dao.get_prompt_by_id(generation_prompts[req.type])
#     if not prompt:
#         AutoGenerationError1['description'] = "提示词不存在"
#         AutoGenerationError1['detail'] = "提示词不存在"
#         return JSONResponse(status_code=500, content=AutoGenerationError1)
#     prompt_message = Message(role="system", content=prompt[0]["f_messages"])
#     input_message = Message(role="user", content=req.input)
#     messages = [prompt_message, input_message]
#     if req.retry:
#         temperature = 0.7
#     else:
#         temperature = 0.01
#     llm_req = LLMUsedOpenAI(
#         model=req.model_name,
#         top_p=0.5,
#         top_k=100,
#         temperature=temperature,
#         presence_penalty=0,
#         frequency_penalty=0,
#         max_tokens=2048,
#         messages=messages,
#         stream=True
#     )
#     res = await llm_used_openai2(llm_req, userId, auth_info.token, auth_info.appid, auth_info.authorization, None)
#     return res
async def run_prompt_endpoint_stream(request, model_para, security_token=None):
    if "stream" not in model_para.keys():
        stream = True
    else:
        stream = model_para["stream"]
        if not isinstance(stream, bool):
            error_dict = ModelFactory_Router_ParamError_TypeError_Error.copy()
            error_dict["detail"] = "stream " + error_dict["detail"]
            StandLogger.error(error_dict["detail"])
            return JSONResponse(status_code=400, content=error_dict)
    if "cache" not in model_para.keys():
        cache = False
    else:
        cache = model_para["cache"]
        if not isinstance(cache, bool):
            error_dict = ModelFactory_Router_ParamError_TypeError_Error.copy()
            error_dict["detail"] = "cache " + error_dict["detail"]
            StandLogger.error(error_dict["detail"])
            return JSONResponse(status_code=400, content=error_dict)
    if "inputs" not in model_para:
        model_para["inputs"] = {}
    if "history_dia" not in model_para:
        model_para["history_dia"] = []
    if "variables" not in model_para:
        model_para["variables"] = []
    try:
        inputs = model_para["inputs"]
        prompt_type = model_para.get("type", "completion")
        model_id = model_para["model_id"]
        history_dia = model_para["history_dia"]
        para = model_para["model_para"]
        info = llm_model_dao.get_data_from_model_list_by_id(model_id)
        try:
            messages = model_para["messages"]
            for message_input, value in inputs.items():
                messages = messages.replace("{{" + str(message_input) + "}}", str(value))
        except Exception as e:
            StandLogger.error(e.args)
            PromptTemplateRunError1["description"] = "messages解析错误"
            PromptTemplateRunError1["detail"] = "messages解析错误"
            return JSONResponse(status_code=500, content=PromptTemplateRunError1)
        prompt_tokens = 100
        context_size = info[0]["f_max_model_len"] * 1000
        if prompt_tokens + para['max_tokens'] > context_size - 50:
            return JSONResponse(status_code=500, content=ModelError)

        result = await used_model_stream(
            request=request,
            llm_id=model_id,
            ai_system=messages,
            ai_user='',
            ai_assistant='',
            ai_history=history_dia,
            prompt_type=prompt_type,
            top_p=para['top_p'],
            temperature=para['temperature'],
            max_tokens=para['max_tokens'],
            frequency_penalty=para['frequency_penalty'],
            presence_penalty=para['presence_penalty'],
            return_info=True,
            user_id=request,
            prompt_tokens=prompt_tokens,
            stream=stream,
            top_k=para.get('top_k', 1),
            model_name=info[0]["f_model_name"],
            cache=cache,
            stop=model_para.get('stop', None),
            system=model_para.get("system", []),
            security_token=security_token
        )
        if stream:
            return result
        else:
            return result
    except Exception as e:
        StandLogger.error(e.args)
        return JSONResponse(status_code=500, content=DataBaseError)
