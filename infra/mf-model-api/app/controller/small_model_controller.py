from fastapi.responses import JSONResponse

from app.commons.errors.codes import ParamValidationErrors
from app.commons.i18n import get_error_message
from app.commons.snow_id import worker
from app.dao.small_model_dao import small_model_dao
from app.interfaces import dbaccess, logics
from app.logs.stand_log import StandLogger
from app.utils.external_small_model_utils import BaiduClient, BaishengClient, InnerClient, BaiduTianchenClient
from app.utils.observability.observability_log import get_logger
from app.utils.param_verify_utils import *
from app.utils.reshape_utils import *
from fastapi import Response, status
from app.mydb.ConnectUtil import redis_util, get_redis_util
from app.utils.permission_manager import PermissionManager, permission_manager


async def add_model(request: logics.AddExternalSmallModel, userId, language, role):
    try:
        model_id = str(worker.get_id())
        config_info = dbaccess.AddExternalSmallModelInfo(
            model_id=model_id,
            model_name=request.model_name,
            model_type=request.model_type,
            model_config=request.model_config,
            adapter=request.adapter,
            adapter_code=request.adapter_code
        )
        name_list = small_model_dao.name_check(request.model_name)
        if len(name_list) > 0:
            StandLogger.error(ModelFactory_ExternalSmallModel_AddModel_RepeatedNames_Error["description"])
            return JSONResponse(status_code=400,
                                content=ModelFactory_ExternalSmallModel_AddModel_RepeatedNames_Error)
        user_infos = await get_username_by_ids([userId])
        permission = await permission_manager.check_single_permission(user_id=userId, resource_id="*",
                                                                      operations="create",
                                                                      resource_type="small_model",
                                                                      role=role)
        if not permission:
            return JSONResponse(status_code=403, content=NotPermissionError)
        status = await permission_manager.add_permission(
            user_id=userId,
            resource_id=model_id,
            resource_name="小模型",
            resource_type="small_model",
            user_name=user_infos[userId],
            role=role
        )
        if not status:
            raise Exception("add permission failed")
        small_model_dao.add_model_info(config_info, userId)
        content = {"status": "ok", "id": model_id}
        return JSONResponse(status_code=200, content=content)
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_ExternalSmallModel_UnknownError.copy()
        error_dict["detail"] = str(e)
        return JSONResponse(status_code=400, content=error_dict)


async def test_model(request, userId, language, role):
    try:
        model_id = request.model_id
        if model_id:
            resource_id = model_id
            model_info = small_model_dao.get_model_info_by_id(model_id)
            if not model_info:
                return JSONResponse(status_code=400,
                                    content=ModelFactory_SmallModelController_ModelApiDoc_ModelNotFoundError)
            config_info = json.loads(model_info[0]["f_model_config"])
            model_type = model_info[0]["f_model_type"]
            adapter = model_info[0]["f_adapter"]
            adapter_code = model_info[0]["f_adapter_code"]
            api_model = config_info.get("api_model", "")
        else:
            resource_id = "*"
            config_info = request.model_config
            model_type = request.model_type
            api_model = config_info.get("api_model", "")
            adapter = request.adapter,
            adapter_code = request.adapter_code
        try:
            # permission = await permission_manager.check_single_permission(user_id=userId, resource_id=resource_id,
            #                                                               operations="execute",
            #                                                               resource_type="small_model")
            # if not permission:
            #     return JSONResponse(status_code=403, content=NotPermissionError)
            client = InnerClient(url=config_info.get("api_url", ""), model_name=api_model,
                                 api_key=config_info.get("api_key", ""),
                                 adapter=adapter, adapter_code=adapter_code)
            if model_type == "embedding":
                texts = ["hello"]
                await client.test_embedding(texts=texts)
            else:
                query = "test"
                documents = ["test"]
                await client.test_reranker(query=query, documents=documents)
        except Exception as e:
            StandLogger.error(str(e))
            error_dict = ModelFactory_ModelController_TestModel_Error_Error.copy()
            error_dict["detail"] = f"{e}"
            return JSONResponse(status_code=400, content=error_dict)
        content = {"status": "ok", "id": model_id}
        return JSONResponse(status_code=200, content=content)
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_ExternalSmallModel_UnknownError.copy()
        return JSONResponse(status_code=400, content=error_dict)


async def edit_model(request: logics.EditExternalSmallModel, userId, language, role):
    try:
        config_info = dbaccess.AddExternalSmallModelInfo(
            model_id=request.model_id,
            model_name=request.model_name,
            model_type=request.model_type,
            model_config=request.model_config,
            adapter=request.adapter,
            adapter_code=request.adapter_code
        )
        try:
            model_info = small_model_dao.get_model_info_by_id(request.model_id)
            if len(model_info) == 0:
                StandLogger.error(ModelFactory_ExternalSmallModel_EditModel_IdNotExist_Error["description"])
                return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_EditModel_IdNotExist_Error)
            name_list = small_model_dao.name_check(request.model_name)
            if len(name_list) > 0 and name_list[0]["f_model_id"] != request.model_id:
                StandLogger.error(ModelFactory_ExternalSmallModel_EditModel_RepeatedNames_Error["description"])
                return JSONResponse(status_code=400,
                                    content=ModelFactory_ExternalSmallModel_EditModel_RepeatedNames_Error)
            permission = await permission_manager.check_single_permission(user_id=userId, resource_id=request.model_id,
                                                                          operations="modify",
                                                                          resource_type="small_model",
                                                                          role=role)
            if not permission:
                return JSONResponse(status_code=403, content=NotPermissionError)
            small_model_dao.edit_model_info(config_info, userId)
            cache_key = f"dip:model-api:small-model:{request.model_name}:list"
            global redis_util
            if redis_util is None:
                redis_util = await get_redis_util()
            await redis_util.delete_str(cache_key)
            content = {"status": "ok", "id": request.model_id}
            return JSONResponse(status_code=200, content=content)
        except Exception as e:
            StandLogger.error(ModelFactory_MyPymysqlPool_Connection_ConnectError_Error["description"])
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_ExternalSmallModel_UnknownError.copy()
        return JSONResponse(status_code=400, content=error_dict)


async def get_info_list(order, rule, page, size, model_name, model_type, model_series, user_id, role):
    try:
        permission_ids = await permission_manager.get_permission_ids(user_id=user_id,
                                                                     operation="display",
                                                                     resource_type="small_model",
                                                                     resource_name="小模型",
                                                                     role=role)
        total = 0
        res_list = []
        if permission_ids:
            try:
                original_res = small_model_dao.get_model_info_list(page, size, order, rule, model_name, model_type,
                                                                   model_series, permission_ids)
                total = len(small_model_dao.get_model_info_total(model_name, model_type, model_series, permission_ids))
            except Exception as e:
                StandLogger.error(e.args)
                return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
            user_ids = await get_userid_by_search(original_res)
            user_infos = await get_username_by_ids(user_ids)
            res_list = []
            for item in original_res:
                res_list.append({
                    "model_id": item["f_model_id"],
                    "model_name": item["f_model_name"],
                    "model_type": item["f_model_type"],
                    "model_config": json.loads(
                        re.sub(r'"api_key":\s*"[^"]*"', '"api_key": "******************************"',
                               item["f_model_config"])),
                    "create_time": item["f_create_time"].strftime("%Y-%m-%d %H:%M:%S"),
                    "update_time": item["f_update_time"].strftime("%Y-%m-%d %H:%M:%S"),
                    "create_by": user_infos.get(item["f_create_by"], ""),
                    "update_by": user_infos.get(item["f_update_by"], ""),
                    "adapter": True if item["f_adapter"] else False,
                    "adapter_code": item["f_adapter_code"],
                })
        content = {"count": total, "data": res_list}
        return JSONResponse(status_code=200, content=content)
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_ExternalSmallModel_UnknownError.copy()
        return JSONResponse(status_code=400, content=error_dict)


async def get_info(model_id, user_id, role):
    try:
        try:
            permission = await permission_manager.check_single_permission(user_id=user_id, resource_id=model_id,
                                                                          operations="display",
                                                                          resource_type="small_model",
                                                                          role=role)
            if not permission:
                return JSONResponse(status_code=403, content=NotPermissionError)
            original_res = small_model_dao.get_model_info_by_id(model_id)
        except Exception as e:
            StandLogger.error(ModelFactory_MyPymysqlPool_Connection_ConnectError_Error["description"])
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        if len(original_res) == 0:
            StandLogger.error(ModelFactory_ExternalSmallModel_GetInfo_IdNotExist_Error["description"])
            return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_GetInfo_IdNotExist_Error)
        item = original_res[0]
        res = {
            "model_id": item["f_model_id"],
            "model_name": item["f_model_name"],
            "model_type": item["f_model_type"],
            "model_config": json.loads(item["f_model_config"]),
            "create_time": item["f_create_time"].strftime("%Y-%m-%d %H:%M:%S"),
            "update_time": item["f_update_time"].strftime("%Y-%m-%d %H:%M:%S"),
            "adapter": True if item["f_adapter"] else False,
            "adapter_code": item["f_adapter_code"],
        }
        return JSONResponse(status_code=200, content=res)
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_ExternalSmallModel_UnknownError.copy()
        return JSONResponse(status_code=400, content=error_dict)


async def delete_model(model_para, userId, language, role):
    try:
        if "model_ids" not in model_para:
            code = ParamValidationErrors.ParamMissing
            content = await get_error_message(code, language)
            content["detail"] = "missing parameters: model_ids"
            return JSONResponse(status_code=400, content=content)
        model_ids = model_para['model_ids']
        if not isinstance(model_ids, list):
            code = ParamValidationErrors.ParamTypeError
            content = await get_error_message(code, language)
            content["detail"] = "parameters type error: model_ids"
            return JSONResponse(status_code=400, content=content)
        if not model_ids:
            return JSONResponse(status_code=400, content=IdValueIsEmpty)
        try:
            original_res = small_model_dao.get_model_info_by_ids(model_ids)
        except Exception as e:
            StandLogger.error(ModelFactory_MyPymysqlPool_Connection_ConnectError_Error["description"])
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        if len(original_res) != len(model_ids):
            StandLogger.error(ModelFactory_ExternalSmallModel_GetInfo_IdNotExist_Error["description"])
            return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_GetInfo_IdNotExist_Error)
        permission_ids = await permission_manager.get_permission_ids(user_id=userId,
                                                                     operation="delete",
                                                                     resource_type="small_model",
                                                                     resource_name="小模型",
                                                                     role=role)
        # permission_delete_ids = [model_id for model_id in model_ids if model_id in permission_ids]
        for model_id in model_ids:
            if model_id not in permission_ids:
                error_dict = NotPermissionError.copy()
                error_dict["detail"] = "部分模型无删除权限"
                return JSONResponse(status_code=403, content=NotPermissionError)
        try:
            # if permission_delete_ids:
            status = await permission_manager.delete_permission(resource_type="small_model",
                                                                resource_ids=model_ids)
            if not status:
                return JSONResponse(status_code=500, content=DeletePermissionResuorceError)
            small_model_dao.delete_model_info_by_ids(model_ids)
            cache_key_list = []
            for model_info in original_res:
                model_name = model_info["f_model_name"]
                cache_key_list.append(f"dip:model-api:small-model:{model_name}:list")
            global redis_util
            if redis_util is None:
                redis_util = await get_redis_util()
            await redis_util.delete_str(cache_key_list)
        except Exception as e:
            StandLogger.error(str(e))
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        content = {"status": "ok", "id": model_ids}
        return JSONResponse(status_code=200, content=content)
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_ExternalSmallModel_UnknownError.copy()
        return JSONResponse(status_code=400, content=error_dict)


async def delete_model_by_name(model_para, userId, language):
    try:
        if "model_names" not in model_para:
            code = ParamValidationErrors.ParamMissing
            content = await get_error_message(code, language)
            content["detail"] = "missing parameters: model_names"
            return JSONResponse(status_code=status.HTTP_400_BAD_REQUEST, content=content)
        model_names = model_para['model_names']
        if not isinstance(model_names, list):
            code = ParamValidationErrors.ParamTypeError
            content = await get_error_message(code, language)
            content["detail"] = "parameters type error: model_names"
            return JSONResponse(status_code=status.HTTP_400_BAD_REQUEST, content=content)
        if not model_names:
            return JSONResponse(status_code=400, content=IdValueIsEmpty)
        try:
            original_res = small_model_dao.get_model_info_by_names(model_names)
        except Exception as e:
            StandLogger.error(ModelFactory_MyPymysqlPool_Connection_ConnectError_Error["description"])
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        if len(original_res) != len(model_names):
            StandLogger.error(ModelFactory_ExternalSmallModel_GetInfo_IdNotExist_Error["description"])
            return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_GetInfo_IdNotExist_Error)
        try:
            model_ids = [line["f_model_id"] for line in original_res]
            small_model_dao.delete_model_info_by_ids(model_ids)
            cache_key_list = []
            for model_info in original_res:
                model_name = model_info["f_model_name"]
                cache_key_list.append(f"dip:model-api:small-model:{model_name}:list")
            global redis_util
            if redis_util is None:
                redis_util = await get_redis_util()
            await redis_util.delete_str(cache_key_list)
        except Exception as e:
            StandLogger.error(str(e))
            return JSONResponse(status_code=500, content=ModelFactory_MyPymysqlPool_Connection_ConnectError_Error)
        content = {"status": "ok", "id": model_ids}
        return JSONResponse(status_code=200, content=content)
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_ExternalSmallModel_UnknownError.copy()
        return JSONResponse(status_code=400, content=error_dict)


async def embedding_model_used(request, userId, language, role, func_module, private=True):
    try:
        model_name = request.model
        model_id = request.model_id
        if not model_name and not model_id:
            error_dict = ModelFactory_Router_ParamError_ParamMissing_Error
            error_dict['detail'] = "model或者model_id至少需要传递一个"
            return JSONResponse(status_code=400, content=error_dict)
        texts = request.input
        if model_name:
            cache_key = f"dip:model-api:small-model:{model_name}:list"
        else:
            cache_key = f"dip:model-api:small-model:{model_id}:list"
        # 确保 redis_util 已初始化
        global redis_util
        if redis_util is None:
            redis_util = await get_redis_util()
        res = await redis_util.get_str(cache_key)
        if res is not None and isinstance(res, (str, bytes)):
            try:
                model_info = eval(res)
            except Exception as e:
                StandLogger.warn(f"解析缓存数据失败: {str(e)}, key={cache_key}, value={res}")
                # 缓存解析失败，回退到数据库查询
                model_info = small_model_dao.get_model_info_by_name_id(model_name, model_id)
                if len(model_info) == 0:
                    return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_Used_NameNotExist)
                # 重新设置缓存
                await redis_util.set_str(key=cache_key, value=str(model_info), expire=3600 * 24)
        else:
            # 缓存不存在或非预期类型，从数据库获取
            model_info = small_model_dao.get_model_info_by_name_id(model_name, model_id)
            if len(model_info) == 0:
                return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_Used_NameNotExist)
            # 设置缓存
            await redis_util.set_str(key=cache_key, value=str(model_info), expire=3600 * 24)
        model_info = model_info[0]
        config_info = json.loads(model_info["f_model_config"])
        adapter = model_info["f_adapter"]
        model_id = model_info["f_model_id"]
        adapter_code = model_info["f_adapter_code"]
        if not private:
            permission = await permission_manager.check_single_permission(user_id=userId, resource_id=model_id,
                                                                          operations="execute",
                                                                          resource_type="small_model",
                                                                          role=role)
            if not permission:
                return JSONResponse(status_code=403, content=NotPermissionError)
        client = InnerClient(url=config_info.get("api_url", ""), model_name=config_info.get("api_model", ""),
                             api_key=config_info.get("api_key", ""), adapter=adapter, adapter_code=adapter_code)
        res_dict = await client.embedding(texts)
        prompt_tokens = res_dict.get("usage", {}).get("prompt_tokens")
        total_tokens = res_dict.get("usage", {}).get("total_tokens")
        if get_logger():
            get_logger().info(
                f'{{"model_name":{model_name},"resourece_type":"embeddings","user_id":{userId},'
                f'"prompt_tokens":{prompt_tokens},"total_tokens":{total_tokens},"func_module":{func_module},"status":"success"}}')
        return res_dict

    except Exception as e:
        StandLogger.error(
            f"call embeddingError,model_name={model_name},model_id={model_id},error_detail={e},body={texts}")
        error_dict = ModelFactory_ExternalSmallModel_UnknownError.copy()
        error_dict["detail"] = str(e)
        if get_logger():
            get_logger().info(
                f'{{"model_name":{model_name},"resourece_type":"embeddings","user_id":{userId},'
                f'"prompt_tokens":0,"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
        return JSONResponse(status_code=400, content=error_dict)


async def reranker_model_used(request, userId, language, role, func_module, private=True):
    try:
        model_name = request.model
        model_id = request.model_id
        if not model_name and not model_id:
            error_dict = ModelFactory_Router_ParamError_ParamMissing_Error
            error_dict['detail'] = "model或者model_id至少需要传递一个"
            return JSONResponse(status_code=400, content=error_dict)
        query = request.query
        documents = request.documents
        if model_name:
            cache_key = f"dip:model-api:small-model:{model_name}:list"
        else:
            cache_key = f"dip:model-api:small-model:{model_id}:list"
        # 确保 redis_util 已初始化
        global redis_util
        if redis_util is None:
            redis_util = await get_redis_util()
        res = await redis_util.get_str(cache_key)
        if res is not None and isinstance(res, (str, bytes)):
            try:
                model_info = eval(res)
            except Exception as e:
                StandLogger.warn(f"解析缓存数据失败: {str(e)}, key={cache_key}, value={res}")
                # 缓存解析失败，回退到数据库查询
                model_info = small_model_dao.get_model_info_by_name_id(model_name, model_id)
                if len(model_info) == 0:
                    return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_Used_NameNotExist)
                # 重新设置缓存
                await redis_util.set_str(key=cache_key, value=str(model_info), expire=3600 * 24)
        else:
            # 缓存不存在或非预期类型，从数据库获取
            model_info = small_model_dao.get_model_info_by_name_id(model_name, model_id)
            if len(model_info) == 0:
                return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_Used_NameNotExist)
            # 设置缓存
            await redis_util.set_str(key=cache_key, value=str(model_info), expire=3600 * 24)

        model_info = model_info[0]
        config_info = json.loads(model_info["f_model_config"])
        adapter = model_info["f_adapter"]
        model_id = model_info["f_model_id"]
        adapter_code = model_info["f_adapter_code"]
        if not private:
            permission = await permission_manager.check_single_permission(user_id=userId, resource_id=model_id,
                                                                          operations="execute",
                                                                          resource_type="small_model",
                                                                          role=role)
            if not permission:
                return JSONResponse(status_code=403, content=NotPermissionError)
        client = InnerClient(url=config_info.get("api_url", ""), model_name=config_info.get("api_model", ""),
                             api_key=config_info.get("api_key", ""), adapter=adapter, adapter_code=adapter_code)
        res_dict = await client.reranker(query=query, documents=documents)
        prompt_tokens = res_dict.get("usage", {}).get("prompt_tokens")
        total_tokens = res_dict.get("usage", {}).get("total_tokens")
        if get_logger():
            get_logger().info(
                f'{{"model_name":{model_name},"resourece_type":"reranker","user_id":{userId},'
                f'"prompt_tokens":{prompt_tokens},"total_tokens":{total_tokens},"func_module":{func_module}}}')
        return res_dict
    except Exception as e:
        StandLogger.error(e.args)
        error_dict = ModelFactory_ExternalSmallModel_UnknownError.copy()
        error_dict["detail"] = str(e)
        if get_logger():
            get_logger().info(
                f'{{"model_name":{model_name},"resourece_type":"embeddings","user_id":{userId},'
                f'"prompt_tokens":0,"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
        return JSONResponse(status_code=400, content=error_dict)
