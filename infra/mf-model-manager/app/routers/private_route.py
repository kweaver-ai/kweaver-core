# 只放接口
from fastapi import APIRouter, Body, Request, Query

from app.controller import small_model_controller, llm_controller
from app.controller.llm_controller import *
from app.controller.prompt_controller import run_prompt_endpoint_stream, source_prompt_item_endpoint, \
    source_prompt_endpoint, prompt_list_endpoint, template_source_prompt_endpoint, check_prompt_endpoint, \
    add_prompt_item_endpoint, edit_prompt_item_endpoint, add_prompt_type_endpoint, edit_prompt_type_endpoint, \
    add_prompt_endpoint, name_edit_prompt_endpoint, edit_prompt_endpoint, edit_template_prompt_endpoint, \
    delete_prompt_endpoint, move_prompt_endpoint, batch_add_prompt_endpoint
from app.interfaces import logics
from app.interfaces.logics import LLMUsedOpenAI
from app.utils.common import get_user_info

private_route = APIRouter()


# 大模型调用内部接口
# @private_route.post("/chat/completions")
async def llm_used_openai2(request: LLMUsedOpenAI, head_request: Request):
    '''
    openai风格大模型调用接口
    ---
    operationId: llm_used_openai
    requestBody:
        description: 'request body'
        content:
            application/json:
                schema:
                type: 'object'
                required:
                    - model
                    - messages
                properties:
                    model:
                        type: string
                        format: string
                        description: '需要调用的模型的名称'
                        example: 'deepseek-chat'
                    top_p:
                        type: float
                        format: float
                        description: '核采样，取值0-1，默认为1'
                        example: 0.7
                    temperature:
                        type: float
                        format: float
                        description: '模型在做出下一个词预测时的确定性和随机性程度。取值0-2，默认为1'
                        example: 0
                    presence_penalty:
                        type: float
                        format: float
                        description: '话题新鲜度，取值-2~2，默认为0'
                        example: 0
                    frequency_penalty:
                        type: float
                        format: float
                        description: '频率惩罚度，取值-2~2，默认为0'
                        example: 0
                    max_tokens:
                        type: integer
                        format: integer
                        description: '单次回复限制，取值10-该模型最大tokens数，默认为1000'
                        example: 1000
                    messages:
                        type: object
                        properties:
                            role:
                                type: string
                                format: string
                                description: '角色：system, assistant, user'
                                example: 'user'
                            content:
                                type: string
                                format: string
                                description: '对话内容'
                                example: '你是谁'
                    stream:
                        type: boolean
                        description: '是否流式返回，默认为否'
                        example: true
                    top_k:
                        type: integer
                        description: '取值大于等于1或者为-1，默认为1',
                        example: 1

    '''
    userId, language, role = await get_user_info(head_request)
    headers = head_request.headers
    func_module = headers.get('x-func-module', "")
    return await used_model_openai(request.dict(), userId, language, func_module)


# reranker模型调用内部接口
# @private_route.post("/small-model/reranker")
async def model_used(request: logics.UsedReranker, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    headers = head_request.headers
    func_module = headers.get('x-func-module', "")
    return await small_model_controller.reranker_model_used(request, userId, language, role, func_module)


# embedding模型调用内部接口
# @private_route.post("/small-model/embedding")
async def model_used(request: logics.UsedEmbedding, head_request: Request):
    userId, language, role = await get_user_info(head_request)
    headers = head_request.headers
    func_module = headers.get('x-func-module', "")
    return await small_model_controller.embedding_model_used(request, userId, language, role, func_module)


# @private_route.post("/small-model/embeddings")
# async def model_used(request: logics.UsedEmbedding, head_request: Request):
#     userId, language, role = await get_user_info(head_request)
#     headers = head_request.headers
#     func_module = headers.get('x-func-module', "")
#     return await small_model_controller.embedding_model_used(request, userId, language, role, func_module)


@private_route.get("/llm/list")
async def source_llm(request: Request, page, size, order='desc', rule='update_time', series='all', name='',
                     api_model='', model_type='', quota: bool = Query(default=None)):
    userId, language, role = await get_user_info(request)
    return await source_model(userId, language, page, size, name, order, series, rule, api_model, model_type, quota)


# 查看大模型接口
@private_route.get("/llm/get")
async def check_llm_(model_id, request: Request, ):
    userId, language, role = await get_user_info(request)
    return await check_model(model_id, userId, language)


# 流式返回接口
@private_route.post("/prompt-run-stream")
async def run_prompt_stream(request: Request, params: dict = Body(...)):
    userId = ""
    security_token = ""
    return await run_prompt_endpoint_stream(userId, params, security_token)


@private_route.get('/prompt-item-source')
async def source_prompt_item(request: Request, prompt_item_name='', prompt_name=''):
    return await source_prompt_item_endpoint(request, prompt_item_name, prompt_name)


# 获取提示词列表信息接口
@private_route.get('/prompt-source')
async def source_prompt(
        request: Request,
        page, size,
        prompt_item_id='',
        prompt_item_type_id='', prompt_name='',
        order='desc', rule='update_time', deploy='all', prompt_type='all'):
    return await source_prompt_endpoint(
        request,
        prompt_item_id, prompt_item_type_id,
        page, size, prompt_name, order, rule, deploy, prompt_type)


# 获取所有提示词的id和name
@private_route.get('/prompt-list')
async def prompt_list():
    return await prompt_list_endpoint()


# 获取提示词模板列表信息接口
@private_route.get('/prompt-template-source')
async def template_source_prompt(request: Request, prompt_type='', prompt_name=''):
    return await template_source_prompt_endpoint(request, prompt_type, prompt_name)


# 提示词查看接口
@private_route.get('/prompt/{prompt_id}')
async def check_prompt(request: Request, prompt_id):
    return await check_prompt_endpoint(request, prompt_id)


# 新建提示词项目接口
@private_route.post("/prompt-item-add")
async def add_prompt_item(request: Request, params: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await add_prompt_item_endpoint(userId, params)


# 编辑提示词项目接口
@private_route.post("/prompt-item-edit")
async def edit_prompt_item(request: Request, model_para: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await edit_prompt_item_endpoint(userId, model_para)


# 新建提示词分类接口
@private_route.post("/prompt-type-add")
async def add_prompt_type(request: Request, model_para: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await add_prompt_type_endpoint(userId, model_para)


# 编辑提示词分类接口
@private_route.post("/prompt-type-edit")
async def edit_prompt_type(request: Request, model_para: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await edit_prompt_type_endpoint(userId, model_para)


# 新增提示词接口
@private_route.post("/prompt-add")
async def add_prompt(request: Request, model_para: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await add_prompt_endpoint(userId, model_para)


# 编辑提示词名称接口
@private_route.post("/prompt-name-edit")
async def name_edit_prompt(request: Request, model_para: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await name_edit_prompt_endpoint(userId, model_para)


# 编辑提示词接口
@private_route.post("/prompt-edit")
async def edit_prompt(request: Request, model_para: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await edit_prompt_endpoint(userId, model_para)


# 提示词管理中编辑提示词接口
@private_route.post("/prompt-template-edit")
async def edit_prompt(request: Request, params: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await edit_template_prompt_endpoint(userId, params)


#
# # 填充提示词open接口
# @private_route .get('/open/prompt_completion/{prompt_id}')
# async def completion_prompt(request: Request, prompt_id, inputs=''):
#     headers = request.headers
#     userId = headers.get("x-account-id")
#     return await completion_prompt_endpoint(userId, prompt_id, inputs)


# 删除接口
@private_route.post("/delete-prompt")
async def delete_prompt(request: Request, delete_id: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await delete_prompt_endpoint(userId, delete_id)


# 提示词移动接口
@private_route.post("/prompt/move")
async def move_prompt(request: Request, move_param: dict = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await move_prompt_endpoint(userId, move_param)


# 提示词批量创建接口
@private_route.post("/prompt/batch_add")
async def batch_add_prompt(request: Request, model_para: list = Body(...)):
    headers = request.headers
    userId = headers.get("x-account-id")
    return await batch_add_prompt_endpoint(userId, model_para)


# 添加大模型接口
@private_route.post("/llm/add")
async def add_llm(request: Request, model_para: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    # 内部接口写死admin用户
    userId = "266c6a42-6131-4d62-8f39-853e7093701c"
    return await llm_controller.add_model(model_para, userId, language)


@private_route.post("/small-model/add")
async def add_model(head_request: Request, request_data: dict = Body(..., example={"is_private": True})):
    userId, language, role = await get_user_info(head_request)
    # 内部接口写死admin用户
    userId = "266c6a42-6131-4d62-8f39-853e7093701c"
    # 确保is_private默认为True
    if "is_private" not in request_data:
        request_data["is_private"] = True
    # 创建AddExternalSmallModel实例
    request = logics.AddExternalSmallModel(**request_data)
    return await small_model_controller.add_model(request, userId, language, role, private=True)


@private_route.post("/llm/delete_by_name")
async def remove_llm(request: Request, model_names: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await remove_model_by_name(model_names, userId, language)


@private_route.post("/small-model/delete_by_name")
async def delete_model(request: Request, model_para: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await small_model_controller.delete_model_by_name(model_para, userId, language)


@private_route.get("/small-model/get")
async def get_info(request: Request, model_id):
    userId, language, role = await get_user_info(request)
    return await small_model_controller.get_info(model_id, userId, role)


@private_route.get("/small-model/list")
async def get_info_list(request: Request, order: str = Query(regex=r'^(desc|asc)$', default="desc"),
                        rule: str = Query(regex=r'^(create_time|update_time|model_name)$', default="update_time"),
                        page: int = Query(ge=1, default=1), size: int = Query(ge=1, default=20),
                        model_name: str = Query(default=""), model_type: str = Query(default=""),
                        model_series: str = Query(default="")):
    userId, language, role = await get_user_info(request)
    return await small_model_controller.get_info_list(order, rule, page, size, model_name, model_type,
                                                      model_series, userId, role)


@private_route.get("/small-model/get_by_name")
async def get_info(request: Request, model_name):
    userId, language, role = await get_user_info(request)
    return await small_model_controller.get_info_by_name(model_name)