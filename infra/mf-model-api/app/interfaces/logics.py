from fastapi import APIRouter, Body, Header
from typing import Optional, List, Dict, Union
from pydantic import StrictFloat, StrictStr, Field, StrictInt, conint, validator, confloat, conlist, constr, \
    root_validator
from fastapi.exceptions import RequestValidationError
from pydantic import BaseModel


class ModelConf(BaseModel):
    api_model: StrictStr = Field(description="api model")
    api_base: StrictStr = Field(description="api base")
    api_key: StrictStr = Field(description="api key")


class AddModelUsedAudit(BaseModel):
    model_id: StrictStr = Field(description="模型id", default="")
    user_id: StrictStr = Field(description="用户id", default="")
    input_tokens: StrictInt = Field(description="使用tokens量", default=0)
    output_tokens: StrictInt = Field(description="输出tokens量", default=0)
    resourece_type: StrictStr = Field(description="资源类型", default="unknown")
    func_module: StrictStr = Field(description="调用模块", default="unknown")
    total_time: StrictFloat = Field(description="调用总时间", default=0.0)
    first_time: StrictFloat = Field(description="首字时间", default=0.0)
    status: StrictStr = Field(description="调用状态", default="failed")

class ConfIdList(BaseModel):
    conf_id_list: List


class ModelIdList(BaseModel):
    model_id_list: conlist(constr(min_length=19, max_length=19))


class Message(BaseModel):
    role: StrictStr = Field(description="", regex=r'^(user|assistant|system|tool)$')
    content: Union[StrictStr, List[Dict[str, Union[str, Dict[str, str]]]]]
    tool_calls: List[dict] = Field(default=None)
    tool_call_id: StrictStr = Field(default=None)

    @validator('content', check_fields=False)
    def validate_content(cls, v):
        if isinstance(v, list):
            image_url_count = 0
            video_url_count = 0
            for item in v:
                if not isinstance(item, dict):
                    raise ValueError("List items must be dictionaries")
                if 'type' not in item:
                    raise ValueError("Each item must have a 'type' field")

                if item['type'] == 'text' and 'text' not in item:
                    raise ValueError("Text items must have 'text' field")

                if item['type'] == 'image_url':
                    if 'image_url' not in item:
                        raise ValueError("Image items must have 'image_url' field")
                    image_url_count += 1
                    if image_url_count > 5:
                        raise ValueError("Maximum 5 image_url items allowed")

                if item['type'] == 'video_url':
                    if 'video_url' not in item:
                        raise ValueError("Video items must have 'video_url' field")
                    video_url_count += 1
                    if video_url_count > 1:
                        raise ValueError("Maximum 1 video_url item allowed")
        return v


class LLMToolsFunctionParameters(BaseModel):
    type: StrictStr
    properties: Dict
    required: List[str] = Field(default=[])


class LLMToolsFunction(BaseModel):
    name: StrictStr
    description: StrictStr = Field(default="")
    parameters: LLMToolsFunctionParameters


class LLMTool(BaseModel):
    type: StrictStr
    function: LLMToolsFunction


class LLMUsedOpenAI(BaseModel):
    model: Optional[StrictStr] = Field(description="",
                             regex=r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$')
    top_p: confloat(gt=0, le=1) = Field(default=1)
    top_k: conint(ge=1) = Field(default=1)
    temperature: confloat(ge=0, le=2) = Field(default=1)
    presence_penalty: confloat(ge=-2, le=2) = Field(default=0)
    frequency_penalty: confloat(ge=-2, le=2) = Field(default=0)
    max_tokens: conint(ge=10) = Field(default=1024)
    messages: List[Message]
    response_format: Dict = Field(default={})
    stream: bool = Field(default=False)
    cache: bool = Field(default=False)
    stop: Union[str, List[str]] = Field(default=None)
    system: List[dict] = Field(default=[])
    tools: List[LLMTool] = Field(default=None)
    tool_choice: Union[str, dict] = Field(default=None)
    model_id: Optional[StrictStr] = Field(description="", default="")

    @validator('stream', pre=True)
    def check_stream(cls, v, values):
        if not v is True and not v is False:
            raise ValueError("stream must be a Boolean value of true or false")
        return v

    @validator('cache', pre=True)
    def check_cache(cls, v, values):
        if not v is True and not v is False:
            raise ValueError("cache must be a Boolean value of true or false")
        return v

    @validator('max_tokens', pre=True)
    def check_max_tokens(cls, v, values):
        if not isinstance(v, int) or v < 10:
            raise ValueError("max_tokens must be an integer and greater than or equal to 10")
        return v

    @validator('top_p', pre=True)
    def check_top_p(cls, v, values):
        if v <= 0 or v > 1:
            raise ValueError("top_p must be an float and the range of values is 0 < top_p ≤ 1")
        return v

    @validator('top_k', pre=True)
    def validate_top_k(cls, v):
        if not isinstance(v, int) or v < 1:
            raise ValueError("top_k must be an integer and greater than or equal to 1")
        return v

    @validator('presence_penalty', pre=True)
    def check_presence_penalty(cls, v, values):
        if v < -2 or v > 2:
            raise ValueError("presence_penalty must be an float and the range of values is -2 ≤ presence_penalty ≤ 2")
        return v

    @validator('frequency_penalty', pre=True)
    def check_frequency_penalty(cls, v, values):
        if v < -2 or v > 2:
            raise ValueError("frequency_penalty must be an float and the range of values is -2 ≤ frequency_penalty ≤ 2")
        return v

    @validator('temperature', pre=True)
    def check_temperature(cls, v, values):
        if v < 0 or v > 2:
            raise ValueError("temperature must be an float and the range of values is 0 ≤ temperature ≤ 2")
        return v


class ModelPara(BaseModel):
    top_p: confloat(ge=0, le=1)
    top_k: conint(ge=1) = Field(default=1)
    temperature: confloat(ge=0, le=2)
    presence_penalty: confloat(ge=-2, le=2)
    frequency_penalty: confloat(ge=-2, le=2)
    max_tokens: conint(ge=10)


class PromptRunPara(BaseModel):
    model_id: constr(min_length=19, max_length=19)
    model_para: ModelPara
    messages: constr()
    inputs: Dict
    variables: List
    history_dia: List
    type: constr()


# 添加外部小模型请求体
class AddExternalSmallModel(BaseModel):
    model_name: StrictStr = Field(description="模型名称")
    model_type: StrictStr = Field(description="模型类型", regex=r'^(reranker|embedding)$')
    model_config: Optional[dict] = Field(default={}, description="第三方模型服务配置")
    adapter: Optional[bool] = Field(default=False, description="是否开启适配服务")
    adapter_code: Optional[StrictStr] = Field(default=None, description="适配代码")

    @root_validator
    def validate_mutually_exclusive_groups(cls, values):
        model_config = values.get('model_config', {})
        adapter = values.get('adapter', False)
        adapter_code = values.get('adapter_code')

        if model_config and (adapter or adapter_code):
            raise ValueError("model_config和adapter/adapter_code不能同时设置")
        if not model_config and not (adapter or adapter_code):
            raise ValueError("必须设置model_config或adapter/adapter_code中的一组")
        return values

    @validator('model_config', pre=False)
    def check_output_tokens(cls, v, values):
        key_list = ["api_url", "api_model"]
        for k in key_list:
            if k not in v.keys():
                raise RequestValidationError([{"loc": ('body', k), "type": "value_error.missing"}])
        return v

    @validator('adapter_code', pre=False)
    def check_adapter_code(cls, v, values):
        if 'adapter' in values and values['adapter'] is True and not v:
            raise RequestValidationError([{"loc": ('body', "adapter_code"), "type": "value_error.missing"}])
        return v


class TestSmallModel(BaseModel):
    model_id: Optional[StrictStr] = Field(None, description="配置id", min_length=19, max_length=19)
    model_name: Optional[StrictStr] = Field(default="", description="模型名称")
    model_type: Optional[StrictStr] = Field(default="", description="模型类型", regex=r'^(reranker|embedding)$')
    model_config: Optional[dict] = Field(default={}, description="第三方模型服务配置")
    adapter: Optional[bool] = Field(default=False, description="是否开启适配服务")
    adapter_code: Optional[StrictStr] = Field(default=None, description="适配代码")

    @root_validator
    def check_fields(cls, values):
        model_id = values.get('model_id')
        if model_id is None:
            model_config = values.get('model_config', {})
            adapter = values.get('adapter', False)
            adapter_code = values.get('adapter_code')
            if model_config and (adapter or adapter_code):
                raise ValueError("model_config和adapter/adapter_code不能同时设置")
            if not model_config and not (adapter or adapter_code):
                raise ValueError("必须设置model_config或adapter/adapter_code中的一组")
            # 如果没有提供model_id，则必须提供其他三个字段
            required_fields = ['model_name', 'model_type']
            for field in required_fields:
                if values.get(field) is None:
                    raise RequestValidationError([{"loc": ('body', field), "type": "value_error.missing"}])
        else:
            # 如果提供了model_id，则其他字段可以为None
            pass
        return values

    @validator('model_config', pre=False)
    def check_output_tokens(cls, v, values):
        key_list = ["api_url", "api_model"]
        for k in key_list:
            if k not in v.keys():
                raise RequestValidationError([{"loc": ('body', k), "type": "value_error.missing"}])
        return v


class EditExternalSmallModel(BaseModel):
    model_id: StrictStr = Field(description="配置id", min_length=19, max_length=19)
    model_name: StrictStr = Field(description="模型名称")
    model_type: StrictStr = Field(description="模型类型", regex=r'^(reranker|embedding)$')
    model_config: Optional[dict] = Field(default={}, description="第三方模型服务配置")
    adapter: Optional[bool] = Field(default=False, description="是否开启适配服务")
    adapter_code: Optional[StrictStr] = Field(default=None, description="适配代码")

    @root_validator
    def validate_mutually_exclusive_groups(cls, values):
        model_config = values.get('model_config', {})
        adapter = values.get('adapter', False)
        adapter_code = values.get('adapter_code')

        if model_config and (adapter or adapter_code):
            raise ValueError("model_config和adapter/adapter_code不能同时设置")
        if not model_config and not (adapter or adapter_code):
            raise ValueError("必须设置model_config或adapter/adapter_code中的一组")
        return values

    @validator('model_config', pre=False)
    def check_output_tokens(cls, v, values):
        key_list = ["api_url", "api_model"]
        for k in key_list:
            if k not in v.keys():
                raise RequestValidationError([{"loc": ('body', k), "type": "value_error.missing"}])
        return v


class UsedReranker(BaseModel):
    model: Optional[StrictStr] = Field(description="模型名称")
    query: StrictStr = Field(description="需要排序的问题")
    documents: list = Field(description="排序的内容列表")
    model_id: Optional[StrictStr] = Field(description="模型id", default="")


class UsedEmbedding(BaseModel):
    model: Optional[StrictStr] = Field(description="模型名称")
    input: list = Field(description="向量化的内容列表")
    model_id: Optional[StrictStr] = Field(description="模型id", default="")

class AuthInfo(BaseModel):
    userid: Optional[str]
    appid: Optional[str]
    authorization: Optional[str]
    token: Optional[str]


def get_auth(userid: Optional[str] = Header(None),
             token: Optional[str] = Header(None), appid: Optional[str] = Header(None),
             authorization: Optional[str] = Header(None)):
    return AuthInfo(userid=userid, token=token, appid=appid, authorization=authorization)


class LLMGenerateReq(BaseModel):
    # model_name: str = Body("",description="")
    input: Optional[str] = Body(...)
    model_name: Optional[str] = Body(...)
    type: Optional[str] = Body(...)
    retry: Optional[bool] = Body(False)
