from fastapi import APIRouter, Body, Request
from app.controller.prompt_controller import *
from app.utils.common import get_user_info

prompt_route = APIRouter()


# 获取提示词项目列表信息接口
@prompt_route.get('/prompt-item-source')
async def source_prompt_item(request: Request, prompt_item_name='', prompt_name=''):
    return await source_prompt_item_endpoint(request, prompt_item_name, prompt_name)


# 获取提示词列表信息接口
@prompt_route.get('/prompt-source')
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
@prompt_route.get('/prompt-list')
async def prompt_list():
    return await prompt_list_endpoint()


# 获取提示词模板列表信息接口
@prompt_route.get('/prompt-template-source')
async def template_source_prompt(request: Request, prompt_type='', prompt_name=''):
    return await template_source_prompt_endpoint(request, prompt_type, prompt_name)


# 提示词查看接口
@prompt_route.get('/prompt/{prompt_id}')
async def check_prompt(request: Request, prompt_id):
    return await check_prompt_endpoint(request, prompt_id)


# 新建提示词项目接口
@prompt_route.post("/prompt-item-add")
async def add_prompt_item(request: Request, params: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await add_prompt_item_endpoint(userId, params)


# 编辑提示词项目接口
@prompt_route.post("/prompt-item-edit")
async def edit_prompt_item(request: Request, model_para: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await edit_prompt_item_endpoint(userId, model_para)


# 新建提示词分类接口
@prompt_route.post("/prompt-type-add")
async def add_prompt_type(request: Request, model_para: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await add_prompt_type_endpoint(userId, model_para)


# 编辑提示词分类接口
@prompt_route.post("/prompt-type-edit")
async def edit_prompt_type(request: Request, model_para: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await edit_prompt_type_endpoint(userId, model_para)


# 新增提示词接口
@prompt_route.post("/prompt-add")
async def add_prompt(request: Request, model_para: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await add_prompt_endpoint(userId, model_para)


# 编辑提示词名称接口
@prompt_route.post("/prompt-name-edit")
async def name_edit_prompt(request: Request, model_para: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await name_edit_prompt_endpoint(userId, model_para)


# 编辑提示词接口
@prompt_route.post("/prompt-edit")
async def edit_prompt(request: Request, model_para: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await edit_prompt_endpoint(userId, model_para)


# 提示词管理中编辑提示词接口
@prompt_route.post("/prompt-template-edit")
async def edit_prompt(request: Request, params: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await edit_template_prompt_endpoint(userId, params)


# 提示词发布接口
# @prompt_route .post("/prompt-deploy")
# async def deploy_prompt(request: Request, model_para: dict = Body(...)):
#     userId, language, role = await get_user_info(request)
#     return await deploy_prompt_endpoint(userId, model_para)


# 提示词取消发布接口
# @prompt_route.post("/prompt-undeploy")
# async def undeploy_prompt(request: Request, model_para: dict = Body(...)):
#     userId, language, role = await get_user_info(request)
#     return await undeploy_prompt_endpoint(userId, model_para)


# 填充提示词open接口
# @prompt_route.get('/open/prompt_completion/{prompt_id}')
# async def completion_prompt(request: Request, prompt_id, inputs=''):
#     userId, language, role = await get_user_info(request)
#     return await completion_prompt_endpoint(userId, prompt_id, inputs)
#
#
# # 代码查看接口
# @prompt_route.get('/prompt-code')
# async def code_prompt(request: Request, model_id, prompt_id=''):
#     userId, language, role = await get_user_info(request)
#     return await code_prompt_endpoint(userId, model_id, prompt_id)


# 获取api文档接口
@prompt_route.get("/prompt-api-doc")
async def api_doc_prompt(request: Request, service_id):
    return await api_doc_prompt_endpoint(service_id)


# 删除接口
@prompt_route.post("/delete-prompt")
async def delete_prompt(request: Request, delete_id: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await delete_prompt_endpoint(userId, delete_id)


# 获取服务id接口
@prompt_route.get("/get-id")
async def get_id(request: Request):
    userId, language, role = await get_user_info(request)
    return await get_id_endpoint(userId)


# 提示词移动接口
@prompt_route.post("/prompt/move")
async def move_prompt(request: Request, move_param: dict = Body(...)):
    userId, language, role = await get_user_info(request)
    return await move_prompt_endpoint(userId, move_param)


# 提示词批量创建接口
@prompt_route.post("/prompt/batch_add")
async def batch_add_prompt(request: Request, model_para: list = Body(...)):
    userId, language, role = await get_user_info(request)
    return await batch_add_prompt_endpoint(userId, model_para)

# # 提示词优化
# @prompt_route .post("/llm_generate")
# async def llm_generate_router(request: Request, req: LLMGenerateReq = Body(...)):
#     userId, language, role = await get_user_info(request)
#     return await llm_generate(userId, language, req)
