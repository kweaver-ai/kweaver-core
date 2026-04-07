import json

from fastapi import APIRouter, Header, Query, Request
from starlette.responses import JSONResponse

from app.commons.errors import ModelFactory_Router_ParamError_TypeError_Error, \
    ModelFactory_Router_ParamError_FormatError_Error
from app.controller.ossclient_controller import *
from app.interfaces import logics
from app.controller import model_quota_controller
from fastapi.exceptions import RequestValidationError
from app.interfaces.logics import GetModelQuotaList
from app.utils.common import get_user_info

model_quota_router = APIRouter()


@model_quota_router.get("/user-quota/remain-check")
async def remain_check(request: Request, model_id_list: str):
    userId, language, role = await get_user_info(request)
    try:
        model_id_list = json.loads(model_id_list)
    except Exception as e:
        error_dict = ModelFactory_Router_ParamError_TypeError_Error.copy()
        error_dict["detail"] = "model_id_list " + error_dict["detail"]
        StandLogger.error(error_dict["detail"])
        return JSONResponse(content=error_dict, status_code=400)
    return await model_quota_controller.remain_check(userId, model_id_list)


# 获取用户配额模型列表
# @model_quota_router.get("/user-quota/model-list")
async def get_user_quote_model_list(request: Request, page: int = Query(ge=0), size: int = Query(ge=0, default=20),
                                    name: str = Query(default=""), api_model: str = Query(default=""),
                                    order: str = Query(regex=r'^(create_time|model_name)$'),
                                    rule: str = Query(regex=r'^(desc|asc)$'), quota: bool = Query(default=None)):
    userId, language, role = await get_user_info(request)
    return await model_quota_controller.get_user_quote_model_list(userId, page, size, name, api_model, order, rule,
                                                                  quota)


# 删除用户配额配置(支持批量)
@model_quota_router.post('/user-quota/delete')
async def delete_user_model_quota_config(request: Request, conf_id_list: logics.ConfIdList):
    userId, language, role = await get_user_info(request)
    return await model_quota_controller.delete_user_model_quota_config(conf_id_list.conf_id_list, userId)


# 添加大模型配额配置
@model_quota_router.post('/model-quota')
async def add_model_quota(request: logics.AddModelQuota, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    if (request.billing_type == 1 and request.output_tokens == 0) or (
            request.billing_type == 0 and request.output_tokens != 0):
        error_dict = ModelFactory_Router_ParamError_FormatError_Error.copy()
        error_dict["detail"] = "output_tokens " + error_dict["detail"]
        StandLogger.error(error_dict["detail"])
        return JSONResponse(status_code=400, content=error_dict)
    return await model_quota_controller.add_model_quota_config(request, userId)


# 编辑大模型配额配置
@model_quota_router.post('/model-quota/{conf_id}')
async def edit_model_quota(conf_id, request: logics.EditModelQuota, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    return await model_quota_controller.edit_model_quota_config(request, conf_id, userId)


# 获取大模型配额配置列表
@model_quota_router.get('/model-quota/list')
async def get_model_quota_list(request: Request, page: int = Query(ge=0), size: int = Query(ge=0, default=20),
                               rule: str = Query(regex=r'^(create_time|update_time|model_name|total_price)$'),
                               order: str = Query(regex=r'^(desc|asc)$'),
                               name: str = Query(default=""),
                               api_model: str = Query(default="")):
    userId, language, role = await get_user_info(request)
    if page > -1 and size == None:
        raise RequestValidationError([{"loc": ('body', "size"), "type": "value_error.missing"}])
    request = GetModelQuotaList(
        page=page,
        size=size,
        order=order,
        rule=rule,
        name=name,
        api_model=api_model
    )
    return await model_quota_controller.get_model_quota_config_list(request, userId)


# 获取指定大模型配额配置
@model_quota_router.get('/model-quota/{conf_id}')
async def get_model_quota(conf_id: str, request: Request):
    userId, language, role = await get_user_info(request)
    return await model_quota_controller.get_model_quota_config(conf_id, userId)


# 添加用户使用指定大模型配额
@model_quota_router.post('/user-quota')
async def add_user_model_quota(request: logics.AddUserModelQuotaList, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    return await model_quota_controller.add_user_model_quota_config(request, userId)


# 获取大模型配额配置列表
@model_quota_router.get('/user-quota/list')
async def get_user_quota_list(head_request: Request,
                              conf_id: str = Query(min_length=19, max_length=19, regex=r'^[0-9]*$'),
                              page: int = Query(ge=0), size: int = Query(ge=0, default=20),
                              rule: str = Query(regex=r'^(create_time|update_time)$'),
                              order: str = Query(regex=r'^(desc|asc)$')):
    userId, language, role = await get_user_info(head_request)
    if page > -1 and size == None:
        raise RequestValidationError([{"loc": ('body', "size"), "type": "value_error.missing"}])
    request = GetModelQuotaList(
        page=page,
        size=size,
        order=order,
        rule=rule
    )
    return await model_quota_controller.get_user_model_quota_config_list(conf_id, request, userId)
