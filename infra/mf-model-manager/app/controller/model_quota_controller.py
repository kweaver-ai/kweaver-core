import datetime
import time
from fastapi.responses import JSONResponse
from app.commons.get_user_info import get_username_by_ids, get_userid_by_search
from app.commons.snow_id import worker
from app.dao.model_quota_dao import model_quota_dao
from app.dao.model_used_audit_dao import model_op_dao
from app.logs.stand_log import StandLogger
from app.mydb.ConnectUtil import redis_util, get_redis_util
from app.utils.config_cache import quota_config_cache_tree, ModelConfigNode
from app.utils.param_verify_utils import *
from app.interfaces import logics, dbaccess
num_type_list = [0, 1000, 10000, 100000000, 1000000, 10000000]


# 新建模型配额配置
async def add_model_quota_config(para: logics.AddModelQuota, user_id: str):
    try:
        try:
            # 校验该模型是否存在
            exist = llm_model_dao.check_model_is_exist(para.model_id)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        if not exist:
            StandLogger.error(ModelFactory_ModelQuotaController_ModelConfig_ModelNotFound_Error["description"])
            return JSONResponse(status_code=400,
                                content=ModelFactory_ModelQuotaController_ModelConfig_ModelNotFound_Error)
        try:
            # 校验该模型是否已经设置过
            exist = model_quota_dao.check_model_is_exist(para.model_id)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        if exist:
            StandLogger.error(
                ModelFactory_ModelQuotaController_ModelQuotaConfig_ModelIsAlreadyExist_Error["description"])
            return JSONResponse(status_code=400,
                                content=ModelFactory_ModelQuotaController_ModelQuotaConfig_ModelIsAlreadyExist_Error)
        try:
            new_id = str(worker.get_id())
            db_config = dbaccess.ModelQuotaInfo(
                conf_id=new_id,
                model_id=para.model_id,
                billing_type=para.billing_type,
                input_tokens=para.input_tokens,
                output_tokens=para.output_tokens,
                currency_type=para.currency_type,
                referprice_in=para.referprice_in,
                referprice_out=para.referprice_out,
                create_time=datetime.datetime.today().strftime("%Y-%m-%d %H:%M:%S"),
                update_time=datetime.datetime.today().strftime("%Y-%m-%d %H:%M:%S"),
                num_type=para.num_type,
                price_type=para.price_type
            )
            config_dict = {
                "f_model_id": para.model_id,
                "f_billing_type": para.billing_type,
                "f_input_tokens": para.input_tokens,
                "f_output_tokens": para.output_tokens,
                "f_currency_type": para.currency_type,
                "f_referprice_in": para.referprice_in,
                "f_referprice_out": para.referprice_out,
                "f_num_type": para.num_type,
                "f_price_type": para.price_type
            }
            quota_config_cache_tree.update(config_dict)
            model_quota_dao.add_new_model_quota_config(db_config)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        return JSONResponse(status_code=200, content={"res": new_id})
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
        error_dict["detail"] = str(e.args)
        return JSONResponse(status_code=400,
                            content=error_dict)


# 编辑模型配额配置
async def edit_model_quota_config(para: logics.EditModelQuota, conf_id: str, user_id: str):
    try:
        try:
            # 校验该模型配置是否存在
            exist = model_quota_dao.get_model_config(conf_id)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        if not exist:
            StandLogger.error(
                ModelFactory_ModelQuotaController_ModelQuotaConfig_ModelConfigNotFound_Error["description"])
            return JSONResponse(status_code=400,
                                content=ModelFactory_ModelQuotaController_ModelQuotaConfig_ModelConfigNotFound_Error)
        try:
            check = model_quota_dao.check_billing_type_edit(para.billing_type, conf_id)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        if not check:
            StandLogger.error("该模型存在用户配额，不支持修改billing_type")
            error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
            error_dict["description"] = error_dict["detail"] = error_dict["solution"] = \
                "该模型存在用户配额，不支持修改billing_type"
            return JSONResponse(status_code=400, content=error_dict)
        # try:
        #     # 获取限额类型
        #     info = model_quota_dao.get_model_billing_type(conf_id)
        # except Exception as e:
        #     StandLogger.error(e.args)
        #     return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)

        if para.billing_type == 1:
            if para.output_tokens is None:
                content = {"code": "ModelFactory.Router.ParamError.ParamMissing",
                           "description": "参数缺失",
                           "detail": "output_tokens 参数缺失",
                           "solution": "请检查填写的参数是否正确。",
                           "link": ""}
                return JSONResponse(status_code=400,
                                    content=content)
            if para.referprice_out is None:
                content = {"code": "ModelFactory.Router.ParamError.ParamMissing",
                           "description": "参数缺失",
                           "detail": "referprice_out 参数缺失",
                           "solution": "请检查填写的参数是否正确。",
                           "link": ""}
                return JSONResponse(status_code=400,
                                    content=content)

        try:
            db_config = dbaccess.ModelQuotaInfo(
                conf_id=conf_id,
                model_id="",
                billing_type=para.billing_type,
                input_tokens=float(para.input_tokens),
                output_tokens=float(para.output_tokens),
                currency_type=para.currency_type,
                referprice_in=para.referprice_in,
                referprice_out=para.referprice_out,
                create_time=datetime.datetime.today().strftime("%Y-%m-%d %H:%M:%S"),
                update_time=datetime.datetime.today().strftime("%Y-%m-%d %H:%M:%S"),
                num_type=para.num_type,
                price_type=para.price_type
            )
            model_quota_dao.edit_model_quota_config(db_config)
            model_id = exist[0]["f_model_id"]
            config_dict = {
                "f_model_id": model_id,
                "f_billing_type": para.billing_type,
                "f_input_tokens": para.input_tokens,
                "f_output_tokens": para.output_tokens,
                "f_currency_type": para.currency_type,
                "f_referprice_in": para.referprice_in,
                "f_referprice_out": para.referprice_out,
                "f_num_type": para.num_type,
                "f_price_type": para.price_type
            }
            quota_config_cache_tree.update(config_dict)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        return JSONResponse(status_code=200, content={"res": conf_id})
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
        error_dict["detail"] = str(e.args)
        return JSONResponse(status_code=400,
                            content=error_dict)


# 获取指定模型配额配置
async def get_model_quota_config(conf_id, user_id: str):
    try:
        res = model_quota_dao.get_model_config(conf_id)

        if res == () or res == []:
            StandLogger.error("配置id不存在")
            return JSONResponse(status_code=400,
                                content=ModelFactory_ModelQuotaController_ModelQuotaConfig_ModelConfigNotFound_Error)
        model_config_info = dbaccess.ModelQuotaInfo(
            conf_id=res[0]["f_id"],
            model_id=res[0]["f_model_id"],
            billing_type=res[0]["f_billing_type"],
            input_tokens=res[0]["f_input_tokens"],
            output_tokens=res[0]["f_output_tokens"],
            currency_type=res[0]["f_currency_type"],
            referprice_in=float(res[0]["f_referprice_in"]),
            referprice_out=float(res[0]["f_referprice_out"]),
            create_time=res[0]["f_create_time"].strftime("%Y-%m-%d %H:%M:%S"),
            update_time=res[0]["f_update_time"].strftime("%Y-%m-%d %H:%M:%S"),
            num_type=json.loads(res[0]["f_num_type"]),
            price_type=json.loads(res[0]["f_price_type"])
        )
        model_data = llm_model_dao.get_data_from_model_list_by_id(model_config_info.model_id)
        if res == () or res == []:
            StandLogger.error("配置id不存在")
            return JSONResponse(status_code=400,
                                content=ModelFactory_ModelQuotaController_ModelQuotaConfig_ModelConfigNotFound_Error)
        res_dict = model_config_info.dict()
        res_dict["model_name"] = model_data[0]["f_model_name"]
        res_dict["model"] = model_data[0]["f_model"]
    except Exception as e:
        StandLogger.error(e.args)
        return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
    return JSONResponse(status_code=200, content={"res": res_dict})


# 获取模型配额配置列表
async def get_model_quota_config_list(request, user_id: str):
    try:
        get_all_config = logics.GetModelQuotaList(
            page=-1,
            size=0,
            order="asc",
            rule="create_time",
            name=request.name,
            api_model=request.api_model
        )
        total = len(model_quota_dao.get_model_config_list(get_all_config))
        model_data_list = llm_model_dao.get_all_model_list()
        model_data_dict = {}
        api_model_list = []
        for model_data in model_data_list:
            model_data_dict[model_data["f_model_id"]] = {
                "model_name": model_data["f_model_name"],
                "model": model_data["f_model"]
            }
            if model_data["f_model"] not in api_model_list:
                api_model_list.append(model_data["f_model"])
        res = model_quota_dao.get_model_config_list(request)

        if res == () or res == []:
            return JSONResponse(status_code=200, content={"res": [], "total": total, "model_list": api_model_list})

        model_quota_list = []
        for info in res:
            try:
                model_config_info = dbaccess.ModelQuotaInfo(
                    conf_id=info["f_id"] if info["f_id"] != None else "",
                    model_id=info["f_model_id"],
                    billing_type=info["f_billing_type"] if info["f_billing_type"] != None else -1,
                    input_tokens=info["f_input_tokens"] if info["f_input_tokens"] != None else -1.0,
                    output_tokens=info["f_output_tokens"] if info["f_output_tokens"] != None else -1.0,
                    currency_type=info["f_currency_type"] if info["f_currency_type"] != None else -1,
                    referprice_in=float(info["f_referprice_in"]) if info["f_referprice_in"] != None else -1.0,
                    referprice_out=float(info["f_referprice_out"]) if info["f_referprice_out"] != None else -1.0,
                    create_time=info["f_create_time"].strftime('%Y-%m-%d %H:%M:%S') if info[
                                                                                           "f_billing_type"] != None else "",
                    update_time=info["f_update_time"].strftime('%Y-%m-%d %H:%M:%S') if info[
                                                                                           "f_billing_type"] != None else "",
                    num_type=json.loads(info["f_num_type"]) if info["f_num_type"] != None else [1, 0],
                    price_type=json.loads(info["f_price_type"]) if info["f_price_type"] != None else ["thousand",
                                                                                                      "thousand"]
                ).dict()
                model_config_info["model"] = model_data_dict[info["f_model_id"]]["model"]
                model_config_info["model_name"] = model_data_dict[info["f_model_id"]]["model_name"]
                model_config_info["model_series"] = info["f_model_series"]
                model_quota_list.append(model_config_info)
            except Exception as e:
                continue

    except Exception as e:
        StandLogger.error(e.args)
        return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
    return JSONResponse(status_code=200,
                        content={"res": model_quota_list, "total": total, "model_list": api_model_list})


# 删除模型陪陪配置
async def delete_model_quota_config(conf_id, user_id: str):
    try:
        try:
            # 校验该模型配置是否存在
            exist = model_quota_dao.check_model_conf_is_exist(conf_id)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        if not exist:
            StandLogger.error(
                ModelFactory_ModelQuotaController_UserModelQuotaConfig_ModelConfigNotFound_Error["description"])
            return JSONResponse(status_code=400,
                                content=ModelFactory_ModelQuotaController_UserModelQuotaConfig_ModelConfigNotFound_Error)

        try:
            model_quota_dao.delete_model_quota_config(conf_id)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        return JSONResponse(status_code=200, content={"res": conf_id})
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
        error_dict["detail"] = str(e.args)
        return JSONResponse(status_code=400,
                            content=error_dict)


# 新建用户使用模型配额配置
async def add_user_model_quota_config(para: logics.AddUserModelQuotaList, user_id: str):
    try:
        if para.list == []:
            error_dict = ModelFactory_Router_ParamError_FormatError_Error.copy()
            error_dict["detail"] = "list不得为空列表"
            StandLogger.error(error_dict["detail"])
            return JSONResponse(content=error_dict, status_code=400)
        user_id_list = []
        try:
            # 校验非必传参数
            conf = model_quota_dao.get_model_config(para.list[0].model_quota_id)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        # 校验当前模型配额配置是否存在
        if conf is None or len(conf) == 0:
            StandLogger.error(
                ModelFactory_ModelQuotaController_ModelQuotaConfig_ModelConfigNotFound_Error["description"])
            return JSONResponse(status_code=400,
                                content=ModelFactory_ModelQuotaController_ModelQuotaConfig_ModelConfigNotFound_Error)
        if conf[0]["f_billing_type"] == 1 and para.list[0].output_tokens == 0:
            StandLogger.error(ModelFactory_Router_ParamError_ParamMissing_Error["description"])
            return JSONResponse(status_code=400,
                                content={"code": "ModelFactory.Router.ParamError.ParamMissing",
                                         "description": "参数缺失",
                                         "detail": "output_tokens 参数缺失",
                                         "solution": "请检查填写的参数是否正确。",
                                         "link": ""})
        if conf[0]["f_billing_type"] == 1:
            for item in para.list:
                if item.output_tokens == 0:
                    error_dict = ModelFactory_Router_ParamError_FormatError_Error.copy()
                    error_dict["detail"] = "output_tokens " + error_dict["detail"]
                    StandLogger.error(error_dict["detail"])
                    return JSONResponse(status_code=400, content=error_dict)
        else:
            for item in para.list:
                if item.output_tokens != 0:
                    error_dict = ModelFactory_Router_ParamError_FormatError_Error.copy()
                    error_dict["detail"] = "output_tokens " + error_dict["detail"]
                    StandLogger.error(error_dict["detail"])
                    return JSONResponse(status_code=400, content=error_dict)

        input_tokens = conf[0]["f_input_tokens"] * num_type_list[json.loads(conf[0]["f_num_type"])[0]]
        output_tokens = conf[0]["f_output_tokens"] * num_type_list[json.loads(conf[0]["f_num_type"])[1]]
        for item in para.list:
            input_tokens -= item.input_tokens * num_type_list[item.num_type[0]]
            output_tokens -= item.output_tokens * num_type_list[item.num_type[1]]
        if conf[0]["f_billing_type"] == 0:
            if input_tokens < 0 or output_tokens < 0:
                StandLogger.error(ModelFactory_ModelQuotaController_UserModelConfig_NoLeftSpace_Error["description"])
                return JSONResponse(status_code=400,
                                    content=ModelFactory_ModelQuotaController_UserModelConfig_NoLeftSpace_Error)
        else:
            if input_tokens < 0:
                StandLogger.error(
                    ModelFactory_ModelQuotaController_UserModelConfig_InputNoLeftSpace_Error["description"])
                return JSONResponse(status_code=400,
                                    content=ModelFactory_ModelQuotaController_UserModelConfig_InputNoLeftSpace_Error)
            elif output_tokens < 0:
                StandLogger.error(
                    ModelFactory_ModelQuotaController_UserModelConfig_OutputNoLeftSpace_Error["description"])
                return JSONResponse(status_code=400,
                                    content=ModelFactory_ModelQuotaController_UserModelConfig_OutputNoLeftSpace_Error)
        old_user_conf_list = model_quota_dao.get_list_from_user_quota_config_by_conf_id(
            para.list[0].model_quota_id)  # 用于对比是新增还是编辑
        old_user_id_list = []
        for item in old_user_conf_list:
            old_user_id_list.append(item["f_user_id"])
        # 区分添加和编辑
        add_list = []
        edit_list = []
        only_old_user_id_list = old_user_id_list.copy()
        user_ids = []
        for item in para.list:
            user_ids.append(item.user_id)
            if item.user_id in user_id_list:
                error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
                error_dict["detail"] = error_dict["description"] = error_dict["solution"] = "添加了重复的用户"
                StandLogger.error(error_dict["description"])
                return JSONResponse(status_code=400, content=error_dict)
            else:
                user_id_list.append(item.user_id)
            if item.user_id in old_user_id_list:
                edit_list.append(item)
                only_old_user_id_list.remove(item.user_id)
            else:
                add_list.append(item)
        for item in old_user_conf_list:
            if item["f_user_id"] in only_old_user_id_list:
                input_tokens -= item["f_input_tokens"] * num_type_list[json.loads(item["f_num_type"])[0]]
                output_tokens -= item["f_output_tokens"] * num_type_list[json.loads(item["f_num_type"])[1]]
        if conf[0]["f_billing_type"] == 0:
            if input_tokens < 0 or output_tokens < 0:
                StandLogger.error(ModelFactory_ModelQuotaController_UserModelConfig_NoLeftSpace_Error["description"])
                return JSONResponse(status_code=400,
                                    content=ModelFactory_ModelQuotaController_UserModelConfig_NoLeftSpace_Error)
        else:
            if input_tokens < 0:
                StandLogger.error(
                    ModelFactory_ModelQuotaController_UserModelConfig_InputNoLeftSpace_Error["description"])
                return JSONResponse(status_code=400,
                                    content=ModelFactory_ModelQuotaController_UserModelConfig_InputNoLeftSpace_Error)
            elif output_tokens < 0:
                StandLogger.error(
                    ModelFactory_ModelQuotaController_UserModelConfig_OutputNoLeftSpace_Error["description"])
                return JSONResponse(status_code=400,
                                    content=ModelFactory_ModelQuotaController_UserModelConfig_OutputNoLeftSpace_Error)
        try:
            model_quota_dao.edit_user_quota_config_by_edit_list(edit_list)
            model_quota_dao.add_user_quota_config_by_add_list(add_list)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        global redis_util
        if redis_util is None:
            redis_util = await get_redis_util()
        model_name = conf[0]["f_model_name"]
        for user_id in user_ids:
            quota_cache_key = f"{user_id}:dip:model-api:llm-quota:{model_name}:list"
            await redis_util.delete_str(quota_cache_key)
        return JSONResponse(status_code=200, content={"res": "success"})
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
        error_dict["detail"] = str(e.args)
        return JSONResponse(status_code=400,
                            content=error_dict)


async def get_user_model_quota_config(conf_id, user_id: str):
    try:
        try:
            res = model_quota_dao.get_user_model_config(conf_id)

            if res == () or res == []:
                StandLogger.error("配置id不存在")
                return JSONResponse(status_code=400,
                                    content=ModelFactory_ModelQuotaController_UserModelQuotaConfig_ModelConfigNotFound_Error)
            model_conf = model_quota_dao.get_model_config(res[0]["f_model_conf"])
            if model_conf == () or model_conf == []:
                StandLogger.error("配置id不存在")
                return JSONResponse(status_code=400,
                                    content=ModelFactory_ModelQuotaController_UserModelQuotaConfig_ModelConfigNotFound_Error)
            model_config_info = dbaccess.ModelUserQuotaInfo(
                conf_id=res[0]["f_id"],
                model_id=model_conf[0]["f_model_id"],
                user_id=res[0]["f_user_id"],
                input_tokens=res[0]["f_input_tokens"],
                output_tokens=res[0]["f_output_tokens"],
                create_time=res[0]["f_create_time"].strftime("%Y-%m-%d %H:%M:%S"),
                update_time=res[0]["f_update_time"].strftime("%Y-%m-%d %H:%M:%S"),
                num_type=json.loads(res[0]["f_num_type"])
            )
            res = model_config_info.dict()
            user_infos = await get_username_by_ids([res["user_id"]])
            res["user_name"] = user_infos[res["user_id"]]
            res["model_name"] = llm_model_dao.get_model_name_by_id(res["model_id"])
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        return JSONResponse(status_code=200, content={"res": res})
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
        error_dict["detail"] = str(e.args)
        return JSONResponse(status_code=400,
                            content=error_dict)


# 获取用户使用模型配额配置列表
async def get_user_model_quota_config_list(conf_id, request: logics.GetModelQuotaList, user_id: str):
    try:
        try:
            delete_id_list = []

            try:
                # 获取模型配置
                confInfo = model_quota_dao.get_model_config(conf_id)
            except Exception as e:
                StandLogger.error(e.args)
                return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
            if confInfo == () or confInfo == []:
                return JSONResponse(status_code=200, content={"res": [],
                                                              "input_tokens_remain": 0,
                                                              "output_tokens_remain": 0})
            input_tokens_remain = confInfo[0]["f_input_tokens"] * num_type_list[
                json.loads(confInfo[0]["f_num_type"])[0]]  # 剩余输入额度(未分配额度)
            output_tokens_remain = confInfo[0]["f_output_tokens"] * num_type_list[
                json.loads(confInfo[0]["f_num_type"])[1]]  # 剩余输出额度(未分配额度)
            res = model_quota_dao.get_user_model_config_list(conf_id, request)

            if res == () or res == []:
                return JSONResponse(status_code=200, content={"res": [],
                                                              "input_tokens_remain": input_tokens_remain,
                                                              "output_tokens_remain": output_tokens_remain})
            model_quota_list = []

            user_id_list = []
            for info in res:
                user_id_list.append(info["f_user_id"])
            user_dict = await get_username_by_ids(user_id_list)
            now = datetime.datetime.now()
            log_list = model_op_dao.get_model_used_logs_list_within_specified_timeframe(
                datetime.datetime(year=now.year, month=now.month, day=1), now)
            log_user_dict = {}
            for item in log_list:
                if item["f_model_id"] != confInfo[0]["f_model_id"]:
                    continue
                log_user_dict[item["f_user_id"]] = {
                    "input_tokens": item["f_input_tokens"],
                    "output_tokens": item["f_output_tokens"]
                }
            for info in res:
                # 计算剩余配额
                if confInfo[0]["f_billing_type"] == 0:
                    input_tokens_remain -= info["f_input_tokens"] * num_type_list[json.loads(info["f_num_type"])[0]]
                else:
                    input_tokens_remain -= info["f_input_tokens"] * num_type_list[json.loads(info["f_num_type"])[0]]
                    output_tokens_remain -= info["f_output_tokens"] * num_type_list[json.loads(info["f_num_type"])[1]]
                if info["f_user_id"] not in log_user_dict.keys():
                    log_user_dict[info["f_user_id"]] = {
                        "input_tokens": 0,
                        "output_tokens": 0
                    }
                if info["f_user_id"] not in user_dict.keys():
                    delete_id_list.append(info["f_id"])
                    continue
                inputs_left = int(info["f_input_tokens"] * num_type_list[json.loads(info["f_num_type"])[0]])
                outputs_left = int(info["f_output_tokens"] * num_type_list[json.loads(info["f_num_type"])[1]])
                if confInfo[0]["f_billing_type"] == 0:
                    inputs_left -= log_user_dict[info["f_user_id"]]["input_tokens"] + log_user_dict[info["f_user_id"]][
                        "output_tokens"]
                else:
                    inputs_left -= log_user_dict[info["f_user_id"]]["input_tokens"]
                    outputs_left -= log_user_dict[info["f_user_id"]]["output_tokens"]
                model_config_info = {
                    "user_quota_id": info["f_id"],
                    "model_quota_id": info["f_model_conf"],
                    "input_tokens": info["f_input_tokens"],
                    "output_tokens": info["f_output_tokens"],
                    "inputs_left": int(inputs_left),
                    "outputs_left": int(outputs_left),
                    "create_time": info["f_create_time"].strftime("%Y-%m-%d %H:%M:%S"),
                    "update_time": info["f_update_time"].strftime("%Y-%m-%d %H:%M:%S"),
                    "user_id": info["f_user_id"],
                    "user_name": user_dict.get(info["f_user_id"], ""),
                    "num_type": json.loads(info["f_num_type"])
                }
                model_quota_list.append(model_config_info)
            if delete_id_list != []:
                model_quota_dao.delete_user_model_quota_config_by_id_list(delete_id_list)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        return JSONResponse(status_code=200, content={"res": model_quota_list,
                                                      "input_tokens_remain": int(input_tokens_remain),
                                                      "output_tokens_remain": int(output_tokens_remain)})

    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
        error_dict["detail"] = str(e.args)
        return JSONResponse(status_code=400,
                            content=error_dict)


# 删除用户使用模型陪陪配置
async def delete_user_model_quota_config(conf_id_list, user_id: str):
    try:

        try:
            res = model_quota_dao.get_model_name_by_quota_config_id(conf_id_list)
            model_quota_dao.delete_user_model_quota_config_by_id_list(conf_id_list)
            global redis_util
            if redis_util is None:
                redis_util = await get_redis_util()
            for line in res:
                model_name = line["f_model_name"]
                quota_cache_key = f"dip:model-api:llm-quota:{model_name}:list"
                await redis_util.delete_str(quota_cache_key)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        return JSONResponse(status_code=200, content={"res": "success"})

    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
        error_dict["detail"] = str(e.args)
        return JSONResponse(status_code=400,
                            content=error_dict)


async def get_user_quote_model_list(userId, page, size, name, api_model, order, rule, quota, model_type):
    try:
        if name is None or name == "":
            user_quote_list, total_list, total, user_quota_token_used_list = model_quota_dao.get_user_quota_model_by_user_id(
                userId, page, size, api_model, order, rule, quota,model_type)
        else:
            user_quote_list, total_list, total, user_quota_token_used_list = model_quota_dao.get_user_quota_model_by_user_id_name_fuzzy(
                userId, page, size, name,
                api_model, order, rule, quota, model_type)
        model_list = []
        for item in total_list:
            if item["f_model"] not in model_list:
                model_list.append(item["f_model"])
        if len(user_quote_list) == 0 and len(user_quota_token_used_list) == 0:
            return {"total": 0, "res": []}
        user_quota_token_used_dict = {}
        for item in user_quote_list:
            if item["f_model_id"] not in user_quota_token_used_dict.keys():
                user_quota_token_used_dict[item["f_model_id"]] = {
                    "input_tokens": 0,
                    "output_tokens": 0
                }
        # 计算已使用额度
        for item in user_quota_token_used_list:
            user_quota_token_used_dict[item["f_model_id"]] = {
                "input_tokens": int(item["sum_input"]),
                "output_tokens": int(item["sum_output"])
            }

        res = []
        user_ids = await get_userid_by_search(user_quote_list)
        user_infos = await get_username_by_ids(user_ids)
        for item in user_quote_list:
            if item["f_num_type"] is not None:
                if item["f_billing_type"] == 0:
                    user_quota_token_used_dict[item["f_model_id"]]["input_tokens"] += \
                        user_quota_token_used_dict[item["f_model_id"]]["output_tokens"]
                res_item = {
                    "input_tokens_remain": int(item["f_input_tokens"] * num_type_list[json.loads(item["f_num_type"])[0]]
                                               - user_quota_token_used_dict[item["f_model_id"]]["input_tokens"]),
                    "input_tokens": int(item["f_input_tokens"] * num_type_list[json.loads(item["f_num_type"])[0]]),
                    "input_tokens_used": user_quota_token_used_dict[item["f_model_id"]]["input_tokens"],
                    "billing_type": item["f_billing_type"],
                    "model_id": item["f_model_id"],
                    "model_name": item["f_model_name"],
                    "model_series": item["f_model_series"],
                    "model": item["f_model"],
                    "create_by": user_infos.get(item["f_create_by"], None),
                    "update_by": user_infos.get(item["f_update_by"], None),
                    "create_time": item["f_create_time"].strftime("%Y-%m-%d %H:%M:%S"),
                    "update_time": item["f_update_time"].strftime("%Y-%m-%d %H:%M:%S"),
                    "max_model_len": item["f_max_model_len"],
                    "model_parameters": item["f_model_parameters"],
                    "model_type": item["f_model_type"],
                    "quota": True if item["f_quota"] == 1 else False,
                    "model_config": json.loads(
                        re.sub(r'"api_key":\s*"[^"]*"', '"api_key": "******************************"',
                               item["f_model_config"]))
                }
                if item["f_billing_type"] == 1:
                    res_item["output_tokens"] = int(item["f_output_tokens"] * num_type_list[
                        json.loads(item["f_num_type"])[1]])
                    res_item["output_tokens_used"] = user_quota_token_used_dict[item["f_model_id"]]["output_tokens"]
                    res_item["output_tokens_remain"] = int(item["f_output_tokens"] * num_type_list[
                        json.loads(item["f_num_type"])[1]] - res_item["output_tokens_used"])
                else:
                    res_item["output_tokens"] = 0
                    res_item["output_tokens_used"] = 0
                    res_item["output_tokens_remain"] = 0
            else:
                res_item = {
                    "input_tokens_remain": None,
                    "input_tokens": None,
                    "input_tokens_used": user_quota_token_used_dict[item["f_model_id"]]["input_tokens"],
                    "billing_type": item["f_billing_type"],
                    "model_id": item["f_model_id"],
                    "model_name": item["f_model_name"],
                    "model_series": item["f_model_series"],
                    "model": item["f_model"],
                    "create_by": user_infos.get(item["f_create_by"], None),
                    "update_by": user_infos.get(item["f_update_by"], None),
                    "create_time": item["f_create_time"].strftime("%Y-%m-%d %H:%M:%S"),
                    "update_time": item["f_update_time"].strftime("%Y-%m-%d %H:%M:%S"),
                    "max_model_len": item["f_max_model_len"],
                    "model_parameters": item["f_model_parameters"],
                    "model_type": item["f_model_type"],
                    "quota": True if item["f_quota"] == 1 else False,
                    "model_config": json.loads(
                        re.sub(r'"api_key":\s*"[^"]*"', '"api_key": "******************************"',
                               item["f_model_config"]))
                }
                if item["f_billing_type"] == 1:
                    res_item["output_tokens"] = None
                    res_item["output_tokens_used"] \
                        = user_quota_token_used_dict[item["f_model_id"]]["output_tokens"]
                    res_item["output_tokens_remain"] = None
                else:
                    res_item["output_tokens"] = 0
                    res_item["output_tokens_used"] = 0
                    res_item["output_tokens_remain"] = 0
            res.append(res_item)
        return {
            "res": res,
            "total": total,
            "model_list": model_list
        }
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
        error_dict["description"] = str(e.args)
        error_dict["detail"] = str(e.args)
        return JSONResponse(status_code=400, content=error_dict)


# 余量检查
async def remain_check(userId, model_id_list):
    try:
        res_list = []
        for model_id in model_id_list:
            log_list = model_op_dao.get_model_used_logs_list_within_specified_timeframe_by_model_id(model_id, userId)
            input_used = 0
            output_used = 0
            for item in log_list:
                input_used += item["f_input_tokens"]
                output_used += item["f_output_tokens"]
            user_quota = model_quota_dao.get_user_model_quota_by_user_id_and_model_id(userId, model_id)
            if user_quota == [] or user_quota == ():

                model_info = llm_model_dao.get_data_from_model_list_by_id(model_id)
                if model_info[0]["f_create_by"] == userId:
                    res = {
                        "model_id": model_id,
                        "input_tokens": 999900000000,
                        "input_used": input_used,
                        "output_tokens": 999900000000,
                        "output_used": output_used
                    }
                    res_list.append(res)
                    continue
                elif model_info[0]["f_quota"] == 0:
                    res = {
                        "model_id": model_id,
                        "input_tokens": -1,
                        "input_used": input_used,
                        "output_tokens": -1,
                        "output_used": output_used
                    }
                    res_list.append(res)
                    continue
                else:
                    StandLogger.error("配置不存在")
                    return JSONResponse(status_code=400,
                                        content=ModelFactory_ModelQuotaController_UserModelQuotaConfig_ModelConfigNotFound_Error)

            res = {
                "model_id": model_id,
                "input_tokens": int(user_quota[0]["f_input_tokens"] * num_type_list[
                    json.loads(user_quota[0]["f_num_type"])[0]]),
                "input_used": input_used,
                "output_tokens": int(user_quota[0]["f_output_tokens"] * num_type_list[
                    json.loads(user_quota[0]["f_num_type"])[1]]),
                "output_used": output_used
            }
            if res["output_tokens"] == 0:
                res["input_used"] += res["output_used"]
                res["output_used"] = 0

            res_list.append(res)

        return JSONResponse(status_code=200, content={"res": res_list})

    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_BenchmarkController_ModelConfig_UnknownError_Error.copy()
        error_dict["description"] = str(e.args)
        error_dict["detail"] = str(e.args)
        return JSONResponse(status_code=400, content=error_dict)
