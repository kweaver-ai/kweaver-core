# 只放接口
import datetime

from fastapi import APIRouter, Body, Request, Query
from app.controller.llm_controller import *
from app.interfaces.logics import LLMUsedOpenAI
from app.utils.common import get_user_info
from app.utils.observability.observability_log import get_logger

llm_route = APIRouter()
health_route = APIRouter()
from fastapi.security import HTTPBearer

bearer_scheme = HTTPBearer()


@health_route.get("/health/ready", include_in_schema=False)
async def health_ready():
    return {"res": 0}


@health_route.get("/health/alive", include_in_schema=False)
async def health_alive():
    return {"res": 0}


@llm_route.post("/chat/completions")
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