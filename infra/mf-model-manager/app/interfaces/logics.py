from fastapi import APIRouter, Body, Header
from typing import Optional, List, Dict, Union
from pydantic import StrictFloat, StrictStr, Field, StrictInt, conint, validator, confloat, conlist, constr, \
    root_validator
from fastapi.exceptions import RequestValidationError
from app.controller.ossclient_controller import *

model_quota_router = APIRouter()

model_quota_list_dict = {
    "create_time": "f_create_time",
    "update_time": "f_update_time",
    "model_name": "f_model_name",
    "total_price": "f_total_price",
    "user_name": "username"
}

"""
正则仅支持中英文、数字和键盘上的特殊字符
^：匹配字符串的开始
[\w\u4e00-\u9fa5~!@#$%^&*()-_=+[]{}\|;:'",<.>/?]`：匹配一个字符，这个字符可以是：
\w：一个字母、数字或下划线
\u4e00-\u9fa5：一个中文字符
~!@#$%^&*()-_=+[]{}\|;:'",<.>/?`：一个键盘上的特殊字符
+：前面的模式可以出现一次或多次
$：匹配字符串的结束
"""
reg0 = r'^[\w\u4e00-\u9fa5~`!@#$%^&*()\-_=+\[\]{}\\|;:\'",<.>/?]+$'

"""
正则仅支持输入英文、数字及键盘上的特殊字符号
^：匹配字符串的开始
[\w~!@#$%^&*()-_=+[]{}\|;:'",<.>/?]`：匹配一个字符，这个字符可以是：
\w：一个字母、数字或下划线
~!@#$%^&*()-_=+[]{}\|;:'",<.>/?`：一个键盘上的特殊字符
+：前面的模式可以出现一次或多次
$：匹配字符串的结束
"""
reg1 = r'^[\w~`!@#$%^&*()\-_=+\[\]{}\\|;:\'",<.>/?]+$'


# 添加大模型配额请求体
class AddModelQuota(BaseModel):
    model_id: StrictStr = Field(description="绑定的模型id", max_length=19, min_length=19)
    billing_type: conint(ge=0, le=1)  # 计费类型 0:统一计费 1:input output单独计费
    input_tokens: confloat(le=9999.9999, ge=1.0) = Field(description="输入tokens总额度")
    output_tokens: Optional[confloat(le=9999.9999, ge=0.0)] = Field(description="输出tokens总额度", default=0.0)
    currency_type: conint(ge=0, le=1)  # 计费单价货币类型,0:人名币 1:美元
    referprice_in: confloat(ge=0.00, le=1000) = Field(description="输入tokens计费单价")
    referprice_out: Optional[confloat(ge=0.00, le=1000)] = Field(description="输出tokens计费单价", default=0)
    num_type: conlist(conint(ge=0, le=5), min_items=2, max_items=2)
    price_type: conlist(constr(regex=r'^(thousand|million)$'), min_items=2, max_items=2)

    # 当billing_type == 1时，output_tokens必传
    @validator('output_tokens', always=True)
    def check_output_tokens(cls, v, values):
        if 'billing_type' in values and values['billing_type'] == 1 and v is None:
            raise RequestValidationError([{"loc": ('body', "output_tokens"), "type": "value_error.missing"}])
        return v

    # 当billing_type == 1时，referprice_out必传
    @validator('referprice_out', always=True)
    def check_referprice_out(cls, v, values):
        if 'billing_type' in values and values['billing_type'] == 1 and v is None:
            raise RequestValidationError([{"loc": ('body', "referprice_out"), "type": "value_error.missing"}])
        return v

    @validator('input_tokens', pre=True, allow_reuse=True)
    def check_input_tokens(cls, v, values):
        return round(v, 3)

    @validator('output_tokens', pre=True, allow_reuse=True)
    def check_output_tokens(cls, v, values):
        return round(v, 3)


# 编辑大模型配额请求体
class EditModelQuota(BaseModel):
    input_tokens: confloat(le=9999.9999, ge=1.0) = Field(description="输入tokens总额度")
    output_tokens: Optional[confloat(le=9999.9999, ge=0.0)] = Field(description="输出tokens总额度", default=0.0)
    currency_type: conint(ge=0, le=1)  # 计费单价货币类型,0:人名币 1:美元
    referprice_in: confloat(ge=0.00, le=1000) = Field(description="输入tokens计费单价")
    referprice_out: Optional[confloat(ge=0.00, le=1000)] = Field(description="输出tokens计费单价", default=0.0)
    billing_type: conint(ge=0, le=1)  # 计费类型 0:统一计费 1:input output单独计费
    num_type: conlist(conint(ge=0, le=5), min_items=2, max_items=2)
    price_type: conlist(constr(regex=r'^(thousand|million)$'), min_items=2, max_items=2)

    @validator('input_tokens', pre=True, allow_reuse=True)
    def check_input_tokens(cls, v, values):
        return round(v, 3)

    @validator('output_tokens', pre=True, allow_reuse=True)
    def check_output_tokens(cls, v, values):
        return round(v, 3)


# 获取大模型配额列表请求体
class GetModelQuotaList(BaseModel):
    page: conint(ge=-1)  # 分页
    size: conint(ge=0)  # 数量
    rule: StrictStr = Field(description="根据指定字段排序",
                             regex=r'^(create_time|update_time|model_name|total_price)$')
    order: StrictStr = Field(description="排序规则", regex=r'^(desc|asc)$')
    name: StrictStr = Field(default="")
    api_model: StrictStr = Field(default="")

    # 当page > 0 时，size必传
    @validator('size', always=True)
    def check_output_tokens(cls, v, values):
        if 'page' in values and values['page'] > -1 and v is None:
            raise RequestValidationError([{"loc": ('body', "size"), "type": "value_error.missing"}])
        return v


# 添加用户使用大模型配额请求体
class AddUserModelQuota(BaseModel):
    model_quota_id: StrictStr = Field(description="绑定的模型配额配置id", max_length=19, min_length=19)
    input_tokens: confloat(ge=1.0, le=9999.9999) = Field(description="输入tokens总额度")
    output_tokens: Optional[confloat(ge=0.0, le=9999.9999)] = Field(description="输出tokens总额度", default=0.0)
    user_id: StrictStr
    num_type: conlist(conint(ge=0, le=5), min_items=2, max_items=2)

    @validator('input_tokens', pre=True)
    def check_input_tokens(cls, v, values):
        return round(v, 3)

    @validator('output_tokens', pre=True)
    def check_output_tokens(cls, v, values):
        return round(v, 3)


class AddUserModelQuotaList(BaseModel):
    list: List[AddUserModelQuota]


# 编辑用户使用大模型配额请求体
class EditUserModelQuota(BaseModel):
    input_tokens: StrictFloat = Field(description="输入tokens总额度")
    output_tokens: Optional[StrictFloat] = Field(description="输出tokens总额度", default=0.0)

    @validator('input_tokens', pre=True)
    def check_input_tokens(cls, v, values):
        return round(v, 3)

    @validator('output_tokens', pre=True)
    def check_output_tokens(cls, v, values):
        return round(v, 3)


# 获取用户使用大模型配额列表请求体
class GetUserModelQuotaList(BaseModel):
    page: conint(ge=-1)  # 分页
    size: conint(ge=0)  # 数量
    order: StrictStr = Field(description="根据指定字段排序", regex=r'^(create_time|update_time)$')
    rule: StrictStr = Field(description="排序规则", regex=r'^(desc|asc)$')

    # 当page > 0 时，size必传
    @validator('size', always=True)
    def check_output_tokens(cls, v, values):
        if 'page' in values and values['page'] > -1 and v is None:
            raise RequestValidationError([{"loc": ('body', "size"), "type": "value_error.missing"}])
        return v


# 获取大模型配额列表请求体
class GetModelOpList(BaseModel):
    page: conint(ge=-1)  # 分页
    size: conint(ge=0)  # 数量
    order: StrictStr = Field(description="根据指定字段排序", regex=r'^(create_time|update_time|total_price|user_name)$')
    rule: StrictStr = Field(description="排序规则", regex=r'^(desc|asc)$')
    # query:Optional[StrictStr]= Field(description="搜索关键字",default="")
    user_id: StrictStr = Field(description="用户id")
    api_model: StrictStr = Field(description="模型")

    # filter_type:StrictStr= Field(description="筛选类型",regex=r'^(user_name|model_name|all)$')

    # 当page > 0 时，size必传
    @validator('size', always=True)
    def check_output_tokens(cls, v, values):
        if 'page' in values and values['page'] > -1 and v is None:
            raise RequestValidationError([{"loc": ('body', "size"), "type": "value_error.missing"}])
        return v


class ModelConf(BaseModel):
    api_model: StrictStr = Field(description="api model")
    api_base: StrictStr = Field(description="api base")
    api_key: StrictStr = Field(description="api key")


class AddLLM(BaseModel):
    model_name: StrictStr = Field(description="模型名称")
    model_series: StrictStr = Field(description="协议，选项：AISHU/OpenAI")
    model_conf: ModelConf = Field(description="模型配置")
    model_type: StrictStr = Field(description="模型类型")
    icon: StrictStr = Field(description="模型配置")
    model_quota: AddModelQuota = Field(description="tokens配额配置")


class EditLLM(BaseModel):
    model_id: StrictStr = Field(description="模型id", max_length=19, min_length=19)
    model_name: StrictStr = Field(description="模型名称")
    icon: StrictStr = Field(description="模型配置")
    model_quota: EditModelQuota = Field(description="tokens配额配置")


class AddModelUsedAudit(BaseModel):
    model_id: StrictStr = Field(description="模型id", default="")
    user_id: StrictStr = Field(description="用户id", default="")
    input_tokens: StrictInt = Field(description="使用tokens量", default=0)
    output_tokens: StrictInt = Field(description="输出tokens量", default=0)


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
    model: StrictStr = Field(description="",
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
    batch_size: int = Field(description="批处理大小")
    max_tokens: Optional[int] = Field(default=None, description="最大token数")
    embedding_dim: Optional[int] = Field(default=None, description="嵌入维度")
    is_private: Optional[bool] = Field(default=False, description="是否为私有路由，用于控制校验逻辑")

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

    @root_validator
    def validate_embedding_fields(cls, values):
        model_type = values.get('model_type')
        max_tokens = values.get('max_tokens')
        embedding_dim = values.get('embedding_dim')
        is_private = values.get('is_private', False)
        
        # 对于私有路由，跳过 max_tokens 和 embedding_dim 的校验
        # 私有路由会在控制器中自动设置默认值
        if model_type == 'embedding' and not is_private:
            if max_tokens is None:
                raise RequestValidationError([{"loc": ('body', "max_tokens"), "type": "value_error.missing", "msg": "当model_type为embedding时，max_tokens为必填字段"}])
            if embedding_dim is None:
                raise RequestValidationError([{"loc": ('body', "embedding_dim"), "type": "value_error.missing", "msg": "当model_type为embedding时，embedding_dim为必填字段"}])
        
        return values


class TestSmallModel(BaseModel):
    model_id: Optional[StrictStr] = Field(None, description="配置id", min_length=19, max_length=19)
    model_name: Optional[StrictStr] = Field(default="", description="模型名称")
    model_type: Optional[StrictStr] = Field(default="", description="模型类型", regex=r'^(reranker|embedding)$')
    model_config: Optional[dict] = Field(default={}, description="第三方模型服务配置")
    adapter: Optional[bool] = Field(default=False, description="是否开启适配服务")
    adapter_code: Optional[StrictStr] = Field(default=None, description="适配代码")
    batch_size: Optional[int] = Field(description="批处理大小")
    max_tokens: Optional[int] = Field(default=None, description="最大token数")
    embedding_dim: Optional[int] = Field(default=None, description="嵌入维度")
    change: Optional[bool] = Field(default=False, description="model_config中的api_key是否做过更改")
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
            model_type = values.get('model_type')
            max_tokens = values.get('max_tokens')
            batch_size = values.get('batch_size')
            embedding_dim = values.get('embedding_dim')
            if batch_size is None:
                raise RequestValidationError([{"loc": ('body', "batch_size"), "type": "value_error.missing",
                                               "msg": "batch_size为必填字段"}])
            if model_type == 'embedding':
                if max_tokens is None:
                    raise RequestValidationError([{"loc": ('body', "max_tokens"), "type": "value_error.missing",
                                                   "msg": "当model_type为embedding时，max_tokens为必填字段"}])
                if embedding_dim is None:
                    raise RequestValidationError([{"loc": ('body', "embedding_dim"), "type": "value_error.missing",
                                                   "msg": "当model_type为embedding时，embedding_dim为必填字段"}])

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
    batch_size: int = Field(description="批处理大小")
    max_tokens: Optional[int] = Field(default=None, description="最大token数")
    embedding_dim: Optional[int] = Field(default=None, description="嵌入维度")
    change: Optional[bool] = Field(default=False, description="model_config中的api_key是否做过更改")

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

    @root_validator
    def validate_embedding_fields(cls, values):
        model_type = values.get('model_type')
        max_tokens = values.get('max_tokens')
        embedding_dim = values.get('embedding_dim')

        if model_type == 'embedding':
            if max_tokens is None:
                raise RequestValidationError([{"loc": ('body', "max_tokens"), "type": "value_error.missing",
                                               "msg": "当model_type为embedding时，max_tokens为必填字段"}])
            if embedding_dim is None:
                raise RequestValidationError([{"loc": ('body', "embedding_dim"), "type": "value_error.missing",
                                               "msg": "当model_type为embedding时，embedding_dim为必填字段"}])

        return values


class UsedReranker(BaseModel):
    model: StrictStr = Field(description="模型名称")
    query: StrictStr = Field(description="需要排序的问题")
    documents: list = Field(description="排序的内容列表")


class UsedEmbedding(BaseModel):
    model: StrictStr = Field(description="模型名称")
    input: list = Field(description="向量化的内容列表")


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
