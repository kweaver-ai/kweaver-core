from fastapi import APIRouter, Header, Query, Request, Body
from app.interfaces import logics
from app.controller import small_model_controller
from app.utils.common import get_user_info
from app.utils.permission_manager import permission_manager

small_model_router = APIRouter()


@small_model_router.post("/small-model/reranker")
async def model_used(request: logics.UsedReranker, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    headers = head_request.headers
    func_module = headers.get('x-func-module', "")
    return await small_model_controller.reranker_model_used(request, userId, language, role, func_module, private=False)


@small_model_router.post("/small-model/embedding")
async def model_used(request: logics.UsedEmbedding, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    headers = head_request.headers
    func_module = headers.get('x-func-module', "")
    return await small_model_controller.embedding_model_used(request, userId, language, role, func_module,
                                                             private=False)


@small_model_router.post("/small-model/embeddings")
async def model_used(request: logics.UsedEmbedding, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    headers = head_request.headers
    func_module = headers.get('x-func-module', "")
    return await small_model_controller.embedding_model_used(request, userId, language, role, func_module,
                                                             private=False)
