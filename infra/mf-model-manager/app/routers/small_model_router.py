from fastapi import APIRouter, Header, Query, Request, Body
from app.interfaces import logics
from app.controller import small_model_controller
from app.utils.common import get_user_info
from app.utils.permission_manager import permission_manager

small_model_router = APIRouter()


@small_model_router.post("/small-model/add")
async def add_model(request: logics.AddExternalSmallModel, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    return await small_model_controller.add_model(request, userId, language, role)


@small_model_router.post("/small-model/edit")
async def edit_model(request: logics.EditExternalSmallModel, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    return await small_model_controller.edit_model(request, userId, language, role)


@small_model_router.get("/small-model/list")
async def get_info_list(request: Request, order: str = Query(regex=r'^(desc|asc)$', default="desc"),
                        rule: str = Query(regex=r'^(create_time|update_time|model_name)$', default="update_time"),
                        page: int = Query(ge=1, default=1), size: int = Query(ge=1, default=20),
                        model_name: str = Query(default=""), model_type: str = Query(default=""),
                        model_series: str = Query(default="")):
    userId, language, role = await get_user_info(request)
    return await small_model_controller.get_info_list(order, rule, page, size, model_name, model_type,
                                                      model_series, userId, role)


@small_model_router.get("/small-model/get")
async def get_info(request: Request, model_id):
    userId, language, role = await get_user_info(request)
    return await small_model_controller.get_info(model_id, userId, role)


@small_model_router.post("/small-model/delete")
async def delete_model(request: Request, model_para: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await small_model_controller.delete_model(model_para, userId, language, role)


# @small_model_router.post("/small-model/reranker")
async def model_used(request: logics.UsedReranker, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    headers = head_request.headers
    func_module = headers.get('x-func-module', "")
    return await small_model_controller.reranker_model_used(request, userId, language, role, func_module, private=False)


# @small_model_router.post("/small-model/embedding")
async def model_used(request: logics.UsedEmbedding, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    headers = head_request.headers
    func_module = headers.get('x-func-module', "")
    return await small_model_controller.embedding_model_used(request, userId, language, role, func_module, private=False)


# @small_model_router.post("/small-model/embeddings")
async def model_used(request: logics.UsedEmbedding, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    headers = head_request.headers
    func_module = headers.get('x-func-module', "")
    return await small_model_controller.embedding_model_used(request, userId, language, role, func_module, private=False)


@small_model_router.post("/small-model/test")
async def model_used(request: logics.TestSmallModel, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    return await small_model_controller.test_model(request, userId, language, role)
