import json
import re
from fractions import Fraction

import func_timeout
from fastapi.exceptions import RequestValidationError
from app.commons.errors import *
from app.dao.llm_model_dao import llm_model_dao


# 新增模型参数校验


async def llm_add_verify(schema_para, userId):
    # model_name
    model_name = schema_para.get("model_name", "")
    if not isinstance(model_name, str) or model_name == "":
        LLMAdd2Error['description'] = "model_name参数不符合规范"
        LLMAdd2Error['detail'] = "model_name必须为字符串类型且不能为空"
        return LLMAdd2Error
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$',
                     schema_para["model_name"]):
        LLMAdd2Error['description'] = "model_name参数不符合规范"
        LLMAdd2Error['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过50字符"
        return LLMAdd2Error
    model_name_list = llm_model_dao.get_model_by_name(model_name)
    if len(model_name_list) > 0:
        if model_name_list[0]["f_create_by"] == userId:
            LLMAdd2Error['description'] = "名称已存在，请修改"
            LLMAdd2Error['detail'] = "名称已存在，请修改"
        else:
            LLMAdd2Error['description'] = "名称已被其他用户占用，请修改"
            LLMAdd2Error['detail'] = "名称已被其他用户占用，请修改"
        error_dict = LLMAdd2Error.copy()
        error_dict["code"] = "ModelFactory.ConnectController.LLMAdd.NameRepeat"
        return error_dict
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9]{,50}$',
                     schema_para['model_config']['api_model']) or len(
        schema_para['model_config']['api_model'].replace(' ', '')) == 0:
        LLMAdd2Error['description'] = "api_model参数不符合规范"
        LLMAdd2Error['detail'] = "当前参数仅支持英文以及键盘上的特殊字符,且不超过50字符"
        return LLMAdd2Error
    if not isinstance(schema_para['max_model_len'], int) or schema_para['max_model_len'] <= 0:
        LLMAdd2Error['description'] = "max_model_len参数不符合规范"
        LLMAdd2Error['detail'] = "max_model_len必须为int类型且大于0"
        return LLMAdd2Error
    if "model_parameters" in schema_para:
        if not isinstance(schema_para['model_parameters'], int) or schema_para['model_parameters'] <= 0:
            LLMAdd2Error['description'] = "model_parameters参数不符合规范"
            LLMAdd2Error['detail'] = "model_parameters必须为int类型且大于0"
            return LLMAdd2Error
    model_series_list = ["tome", "qwen", "openai", "internlm", "deepseek", "qianxun", "claude",
                         "chatglm", "llama", "others", "baidu", "baidu_tianchen"]
    try:
        if schema_para['model_series'] not in model_series_list:
            LLMAdd2Error['description'] = f"参数model_series必须在{model_series_list}中"
            LLMAdd2Error['detail'] = f"参数model_series必须在{model_series_list}中"
            return LLMAdd2Error

        if schema_para['model_series'] == 'openai':
            if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9]+$',
                             schema_para['model_config']['api_key']) or len(
                schema_para['model_config']["api_key"].replace(' ', '')) == 0:
                LLMAdd2Error['description'] = "api_key参数不符合规范"
                LLMAdd2Error['detail'] = "当前参数仅支持英文以及键盘上的特殊字符"
                return LLMAdd2Error
        elif schema_para['model_series'].lower() == 'tome':
            if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9]{,400}$',
                             schema_para['model_config']['api_url']) or len(
                schema_para['model_config']["api_url"].replace(' ', '')) == 0:
                LLMAdd2Error['description'] = "api_url参数不符合规范"
                LLMAdd2Error['detail'] = "当前参数仅支持英文以及键盘上的特殊字符,且不超过400字符"
                return LLMAdd2Error
        elif schema_para['model_series'].lower() == "others":
            schema_para['model_config']["api_url"] = schema_para["model_config"]["api_url"]
            if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9]{,400}$',
                             schema_para['model_config']['api_url']) or len(
                schema_para['model_config']["api_url"].replace(' ', '')) == 0:
                LLMAdd2Error['description'] = "api_url参数不符合规范"
                LLMAdd2Error['detail'] = "当前参数仅支持英文以及键盘上的特殊字符,且不超过400字符"
                return LLMAdd2Error
        else:
            schema_para['model_config']["api_url"] = schema_para["model_config"]["api_url"]
            if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9]{,400}$',
                             schema_para['model_config']['api_url']) or len(
                schema_para['model_config']["api_url"].replace(' ', '')) == 0:
                LLMAdd2Error['description'] = "api_url参数不符合规范"
                LLMAdd2Error['detail'] = "当前参数仅支持英文以及键盘上的特殊字符,且不超过400字符"
                return LLMAdd2Error
            if "api_key" in schema_para['model_config']:
                if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9]+$',
                                 schema_para['model_config']['api_key']) or len(
                    schema_para['model_config']["api_key"].replace(' ', '')) == 0:
                    LLMAdd2Error['description'] = "api_key参数不符合规范"
                    LLMAdd2Error['detail'] = "当前参数仅支持英文以及键盘上的特殊字符"
                    return LLMAdd2Error
        if "quota" in schema_para.keys() and schema_para["quota"] is not True and schema_para["quota"] is not False:
            raise RequestValidationError([{"loc": ('body', "quota"), "type": "value_error.type_error"}])
        api_key = schema_para['model_config'].get("api_key", None)
        if llm_model_dao.check_model_unique(schema_para['model_config']["api_url"],
                                            schema_para['model_config']["api_model"], userId, api_key):
            error_dict = LLMAdd2Error.copy()
            error_dict["code"] = "ModelFactory.ConnectController.LLMAdd.BaseAndModelRepeat"
            error_dict["description"] = "模型api_url、api_model{}重复".format(
                "" if api_key in ["", None] else "、api_key")
            error_dict["detail"] = "模型api_url、api_model{}重复".format("" if api_key in ["", None] else "、api_key")
            return error_dict
    except Exception as e:
        LLMAdd2Error['description'] = "config不符合规范"
        LLMAdd2Error['detail'] = str(e)
        return LLMAdd2Error


# 测试模型参数校验
def llm_test_verify(model_param):
    key_list = ["model_series", "model_config", "model_type"]
    for k in key_list:
        if k not in model_param.keys():
            raise RequestValidationError([{"loc": ('body', k), "type": "value_error.missing"}])
    conf_list = list(model_param['model_config'].keys())
    if 'api_model' not in conf_list or not model_param['model_config']['api_model']:
        LLMTestError['description'] = "api_model 参数不符合规范"
        LLMTestError['detail'] = "请正确输入 api_model"
        return LLMTestError
    elif 'api_url' not in conf_list or not model_param['model_config']['api_url']:
        LLMTestError['description'] = "api_url 参数不符合规范"
        LLMTestError['detail'] = "请正确输入 api_url"
        return LLMTestError

    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9]{,50}$',
                     model_param['model_config']['api_model']):
        LLMTestError['description'] = "api_model参数不符合规范"
        LLMTestError['detail'] = "当前参数仅支持英文以及键盘上的特殊字符,且不超过50字符"
        return LLMTestError
    try:
        if 'api_key' in model_param['model_config']:
            if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9]+$',
                             model_param['model_config']['api_key']):
                LLMTestError['description'] = "api_key参数不符合规范"
                LLMTestError['detail'] = "当前参数仅支持英文以及键盘上的特殊字符"
                return LLMTestError
        if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9]{,400}$',
                         model_param['model_config']['api_url']):
            LLMTestError['description'] = "api_url 参数不符合规范"
            LLMTestError['detail'] = "当前参数仅支持英文以及键盘上的特殊字符,且不超过400字符"
            return LLMTestError
        if model_param["model_type"] not in ('llm', 'rlm','vu'):
            LLMTestError['description'] = "model_type 参数不符合规范"
            LLMTestError['detail'] = "当前参数仅支持 llm|rlm"
            return LLMTestError
    except Exception as e:
        print(e)
        LLMTestError['description'] = "config不符合规范"
        LLMTestError['detail'] = ""
        return LLMTestError


# 编辑模型校验
def llm_edit_verify(model_para):
    key_list = ["model_config", "model_series", "model_name", "model_id", "max_model_len", "model_type"]
    for k in key_list:
        if k not in model_para.keys():
            raise RequestValidationError([{"loc": ('body', k), "type": "value_error.missing"}])
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$',
                     model_para['model_name']):
        LLMEditError['description'] = "model_name参数不符合规范"
        LLMEditError['detail'] = "当前参数仅支中英文以及键盘上的特殊字符,且不超过50字符"
        return LLMEditError
    if not isinstance(model_para['max_model_len'], int) or model_para['max_model_len'] <= 0:
        LLMEditError['description'] = "max_model_len参数不符合规范"
        LLMEditError['detail'] = "max_model_len必须为int类型且大于0"
        return LLMEditError
    if "model_parameters" in model_para:
        if not isinstance(model_para['model_parameters'], int) or model_para['model_parameters'] <= 0:
            LLMEditError['description'] = "model_parameters参数不符合规范"
            LLMEditError['detail'] = "model_parameters必须为int类型且大于0"
            return LLMEditError
    if model_para["model_type"] not in ('llm', 'rlm', 'vu'):
        LLMEditError['description'] = "model_type 参数不符合规范"
        LLMEditError['detail'] = "当前参数仅支持 llm|rlm"
        return LLMEditError


# 大模型列表校验
def llm_source_verify(order, page, size, rule, series, name, model_type):
    if not re.search(r'^[1-9]\d*$', page):
        LLMSourceError['description'] = "page 参数不符合规范"
        LLMSourceError['detail'] = "当前参数仅支持大于零的整数"
        return LLMSourceError
    if not re.search(r'^[1-9]\d*$', size):
        LLMSourceError['description'] = "size 参数不符合规范"
        LLMSourceError['detail'] = "当前参数仅支持大于零的整数"
        return LLMSourceError
    if not re.search(r'^(asc|desc)$', order):
        LLMSourceError['description'] = "order 参数不符合规范"
        LLMSourceError['detail'] = "当前参数仅支持 asc|desc"
        return LLMSourceError
    if not re.search(r'^(update_time|create_time|model_name)$', rule):
        LLMSourceError['description'] = "rule 参数不符合规范"
        LLMSourceError['detail'] = "当前参数仅支持 update_time|create_time|model_name"
        return LLMSourceError
    if not series:
        LLMSourceError['description'] = "series 参数不符合规范"
        LLMSourceError['detail'] = "series 不能为空"
        return LLMSourceError
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$|^$',
                     name):
        LLMSourceError['description'] = "name 参数不符合规范"
        LLMSourceError['detail'] = "当前参数仅支持中英文以及键盘上的特殊字符,且不超过50字符"
    if model_type and model_type not in ('llm', 'rlm', 'vu'):
        LLMSourceError['description'] = "model_type 参数不符合规范"
        LLMSourceError['detail'] = "当前参数仅支持 llm|rlm"
        return LLMSourceError
    return False


def verify_icon_color_config(s):
    valid_colors = [
        "icon-color-pz-019688",
        "icon-color-pz-F759AB",
        "icon-color-pz-FADB14",
        "icon-color-pz-FF8501",
        "icon-color-pz-F75959",
        "icon-color-pz-8C8C8C",
        "icon-color-pz-126EE3",
        "icon-color-pz-13C2C2",
        "icon-color-pz-52C41A",
        "icon-color-pz-9254DE"
    ]
    if s in valid_colors:
        return True
    else:
        return False


def verify_icon_color_config_metric(s):
    valid_colors = [
        "icon-color-zbk-FF8501",
        "icon-color-zbk-13C2C2",
        "icon-color-zbk-FADB14",
        "icon-color-zbk-019688",
        "icon-color-zbk-9254DE",
        "icon-color-zbk-8C8C8C",
        "icon-color-zbk-126EE3",
        "icon-color-zbk-52C41A",
        "icon-color-zbk-F759AB",
        "icon-color-zbk-F75959"
    ]
    if s in valid_colors:
        return True
    else:
        return False


# 检查字段是否不含表情等字符，是否不超过指定长度
def verify_text_field(s, max_len):
    if not isinstance(s, str):
        return False
    if len(s) > max_len:
        return False
    if not re.match(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！·￥…（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]*$', s):
        return False
    return True


def verify_id(s):
    if not isinstance(s, str):
        return False
    # 只能有数字
    if not re.match(r'^[0-9]{18}$', s) and not re.match(r'^[0-9]{19}$', s):
        return False
    return True


def include_dataset_id(dataset_version_id_list, dataset_id):
    try:
        for item in dataset_version_id_list:
            if item.split("/")[0] == dataset_id:
                return True
        return False
    except Exception:
        return False
def prompt_source_item_verify(prompt_item_name):
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$|^$',
                     prompt_item_name):
        PromptItemSourceError['description'] = "搜索名称参数不符合规范"
        PromptItemSourceError['detail'] = "当前参数仅支持中英文以及键盘上的特殊字符,且不超过50字符"
        return PromptItemSourceError

async def prompt_source_verify(prompt_item_id, prompt_item_type_id, page, size, prompt_name, order,
                               rule, deploy, prompt_type):
    if not prompt_item_id and prompt_item_id != '':
        PromptSourceError['description'] = "prompt_item_id 参数不符合规范"
        PromptSourceError['detail'] = "缺失参数prompt_item_id"
        return PromptSourceError
    # prompt_item_type_id_list = [cell.f_prompt_item_type_id for cell in await PromptItemList.all()]
    # if prompt_item_type_id not in prompt_item_type_id_list and prompt_item_type_id != '':
    #     PromptSourceError['description'] = "prompt_item_type_id 参数不符合规范"
    #     PromptSourceError['detail'] = "当前参数仅支持数据库中已有id"
    #     return PromptSourceError
    if not re.search(r'^[1-9]\d*$', page):
        PromptSourceError['description'] = "page 参数不符合规范"
        PromptSourceError['detail'] = "当前参数仅支持大于零的整数"
        return PromptSourceError
    if not re.search(r'^[1-9]\d*$', size):
        PromptSourceError['description'] = "size 参数不符合规范"
        PromptSourceError['detail'] = "当前参数仅支持大于零的整数"
        return PromptSourceError
    if not re.search(r'^(asc|desc)$', order):
        PromptSourceError['description'] = "order 参数不符合规范"
        PromptSourceError['detail'] = "当前参数仅支持 asc|desc"
        return PromptSourceError
    if not re.search(r'^(update_time|create_time|prompt_name)$', rule):
        PromptSourceError['description'] = "rule 参数不符合规范"
        PromptSourceError['detail'] = "当前参数仅支持 update_time|create_time|prompt_name"
        return PromptSourceError
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$|^$',
                     prompt_name):
        PromptSourceError['description'] = "搜索名称参数不符合规范"
        PromptSourceError['detail'] = "当前参数仅支持中英文以及键盘上的特殊字符,且不超过50字符"
        return PromptSourceError
    if not re.search(r'^(yes|no|all|^$)$', deploy):
        PromptSourceError['description'] = "deploy 参数不符合规范"
        PromptSourceError['detail'] = "当前参数仅支持 yes|no|all"
        return PromptSourceError
    if not re.search(r'^(chat|completion|all|^$)$', prompt_type):
        PromptSourceError['description'] = "prompt_type 参数不符合规范"
        PromptSourceError['detail'] = "当前参数仅支持 chat|completion|all"
        return PromptSourceError
def prompt_llm_source_verify(types):
    if not re.search(r'^(chat|completion|^$)$', types):
        PromptLLMSourceError['description'] = "types 参数不符合规范"
        PromptLLMSourceError['detail'] = "当前参数仅支持 chat|completion|空"
        return PromptLLMSourceError


# 获取提示词模板校验
def prompt_template_verify(prompt_name, prompt_type):
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$|^$',
                     prompt_name):
        PromptTemplateSource['description'] = "搜索名称参数不符合规范"
        PromptTemplateSource['detail'] = "当前参数仅支持中英文以及键盘上的特殊字符,且不超过50字符"
        return PromptTemplateSource
    if not re.search(r'^(chat|completion|^$)$', prompt_type):
        PromptLLMSourceError['description'] = "prompt_type 参数不符合规范"
        PromptLLMSourceError['detail'] = "当前参数仅支持 chat|completion|空"
        return PromptLLMSourceError


# 查看提示词校验
async def check_prompt_verify(prompt_id):
    prompt_id_list = [cell["f_prompt_id"] for cell in prompt_dao.get_all_data_from_prompt_list()]
    if prompt_id not in prompt_id_list:
        # PromptCheck['description'] = "prompt_id 参数不符合规范"
        # PromptCheck['detail'] = "当前参数仅支持数据库中已有prompt_id"
        return {'res': []}


# 调用大模型校验对prompt_id 单独校验
async def used_prompt_id_verify(prompt_service_id):
    prompt_service_id_list = prompt_dao.get_all_prompt_service_id()
    # prompt不存在校验
    if prompt_service_id not in prompt_service_id_list:
        PromptUsed['description'] = "prompt_service_id 参数不符合规范"
        PromptUsed['detail'] = "当前参数仅支持数据库中已有 prompt_service_id"
        return PromptUsed


# 新建提示词项目参数校验
async def item_add_verify(model_para):
    if "prompt_item_name" not in model_para:
        PromptItemAddError1['description'] = "参数错误"
        PromptItemAddError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptItemAddError1]
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$',
                     model_para["prompt_item_name"]) or len(model_para["prompt_item_name"].replace(' ', '')) == 0:
        PromptItemAddError1['description'] = "prompt_item_name参数不符合规范"
        PromptItemAddError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过50字符"
        return [400, PromptItemAddError1]
    name = model_para['prompt_item_name']
    item_name_list = [ids["f_prompt_item_name"] for ids in prompt_dao.get_all_prompt_item_list_distinct(None)]
    if name in item_name_list:
        PromptItemAddError2['description'] = "prompt_item_name参数不符合规范"
        PromptItemAddError2['detail'] = "提示词项目名称重复"
        return [500, PromptItemAddError2]


# 编辑提示词项目参数校验
async def item_edit_verify(model_para):
    if "prompt_item_id" not in model_para or "prompt_item_name" not in model_para:
        PromptItemEditError1['description'] = "参数错误"
        PromptItemEditError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptItemEditError1]
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$',
                     model_para["prompt_item_name"]) or len(model_para["prompt_item_name"].replace(' ', '')) == 0:
        PromptItemEditError1['description'] = "prompt_item_name参数不符合规范"
        PromptItemEditError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过50字符"
        return [400, PromptItemEditError1]
    ids_list = [ids["f_prompt_item_id"] for ids in prompt_dao.get_all_prompt_item_list_distinct(None)]
    # info = await PromptItemList.filter(f_prompt_item_id=model_para['prompt_item_id'])
    info = prompt_dao.get_all_info_from_prompt_item_list_by_item_id(model_para['prompt_item_id'])
    if model_para['prompt_item_id'] not in ids_list or info[0]["f_item_is_delete"] == 1:
        PromptItemEditError1['description'] = "prompt_item_id参数不符合规范"
        PromptItemEditError1['detail'] = "该提示词项目不存在"
        return [400, PromptItemEditError1]
    name_new = model_para['prompt_item_name']
    itemid_list1 = [ids["f_prompt_item_id"] for ids in prompt_dao.get_all_prompt_item_list_distinct(None)]
    itemid_list = [num for num in itemid_list1 if num != model_para['prompt_item_id']]
    item_name_list = []
    for i in itemid_list:
        info = prompt_dao.get_all_info_from_prompt_item_list_by_item_id(i)
        item_name_list.append(info[0]["f_prompt_item_name"])
    if name_new in item_name_list:
        PromptItemEditError2['description'] = "prompt_item_name参数不符合规范"
        PromptItemEditError2['detail'] = "提示词项目名称重复"
        return [500, PromptItemEditError2]


# 新建提示词分类参数校验
async def type_add_verify(model_para):
    if "prompt_item_id" not in model_para or "prompt_item_type" not in model_para:
        PromptTypeAddError1['description'] = "参数错误"
        PromptTypeAddError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptTypeAddError1]
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$',
                     model_para["prompt_item_type"]) or len(model_para["prompt_item_type"].replace(' ', '')) == 0:
        PromptTypeAddError1['description'] = "prompt_item_type参数不符合规范"
        PromptTypeAddError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过50字符"
        return [400, PromptTypeAddError1]
    item_type = model_para['prompt_item_type']
    item_type_list = [ids["f_prompt_item_type"] for ids in
                    prompt_dao.get_all_info_from_prompt_item_list_by_item_id(model_para["prompt_item_id"])]
    ids_list = [ids["f_prompt_item_id"] for ids in prompt_dao.get_all_prompt_item_list_distinct(None)]
    info = prompt_dao.get_all_info_from_prompt_item_list_by_item_id(model_para["prompt_item_id"])
    if model_para['prompt_item_id'] not in ids_list or info[0]["f_item_is_delete"] == 1:
        PromptTypeAddError1['description'] = "prompt_item_id参数不符合规范"
        PromptTypeAddError1['detail'] = "提示词项目不存在"
        return [400, PromptTypeAddError1]
    if item_type in item_type_list:
        PromptTypeAddError2['description'] = "prompt_item_type参数不符合规范"
        PromptTypeAddError2['detail'] = "提示词项目分类名称重复"
        return [500, PromptTypeAddError2]


# 编辑提示词分类参数校验
async def type_edit_verify(model_para):
    if "prompt_item_type" not in model_para or "prompt_item_type_id" not in model_para:
        PromptTypeEditError1['description'] = "参数错误"
        PromptTypeEditError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptTypeEditError1]
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥·…（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$',
                     model_para["prompt_item_type"]) or len(model_para["prompt_item_type"].replace(' ', '')) == 0:
        PromptTypeEditError1['description'] = "prompt_item_type参数不符合规范"
        PromptTypeEditError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过50字符"
        return [400, PromptTypeEditError1]
    info = prompt_dao.get_data_from_prompt_item_list_by_type_id(model_para["prompt_item_type_id"])
    if info == () or info[0]["f_item_is_delete"] == 1:
        PromptTypeEditError1['description'] = "prompt_item_type_id参数不符合规范"
        PromptTypeEditError1['detail'] = "提示词项目不存在"
        return [400, PromptTypeEditError1]
    item_id = info[0]["f_prompt_item_id"]  # 所属项目id
    # type_name_list = [ids.f_prompt_item_type for ids in await PromptItemList.filter(f_prompt_item_id=item_id)] #项目下所有分类
    type_id_list = [ids["f_prompt_item_type_id"] for ids in
                    prompt_dao.get_all_info_from_prompt_item_list_by_item_id(item_id)]  # 项目下所有分类id
    if model_para['prompt_item_type_id'] not in type_id_list or info[0]["f_item_is_delete"] == 1:
        PromptTypeEditError1['description'] = "prompt_item_type_id参数不符合规范"
        PromptTypeEditError1['detail'] = "提示词项目分类不存在"
        return [400, PromptTypeEditError1]
    item_type = model_para['prompt_item_type']
    type_id_list1 = [num for num in type_id_list if num != model_para['prompt_item_type_id']]
    type_name_list = []
    for i in type_id_list1:
        info1 = prompt_dao.get_data_from_prompt_item_list_by_type_id(i)
        type_name_list.append(info1[0]["f_prompt_item_type"])
    if item_type in type_name_list:
        PromptTypeEditError2['description'] = "prompt_item_type参数不符合规范"
        PromptTypeEditError2['detail'] = "提示词项目分类名称重复"
        return [500, PromptTypeEditError2]
    # if info[0].f_prompt_item_type == '未分类':
    #     PromptTypeEditError1['description'] = "prompt_item_type_id参数不符合规范"
    #     PromptTypeEditError1['detail'] = "该分类名称不能编辑"
    #     return [400, PromptTypeEditError1]


# 新增提示词
async def prompt_add_verify(para):
    if "prompt_item_id" not in para or "prompt_item_type_id" not in para or "prompt_name" not in para or "prompt_type" not in para or "model_id" not in para or "icon" not in para or "model_para" not in para or "messages" not in para:
        PromptAddError1['description'] = "参数错误"
        PromptAddError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptAddError1]
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$',
                     para["prompt_name"]) or len(para["prompt_name"].replace(' ', '')) == 0:
        PromptAddError1['description'] = "prompt_name参数不符合规范"
        PromptAddError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过50字符"
        return [400, PromptAddError1]
    if not re.search(r'^(chat|completion)$', para["prompt_type"]):
        PromptAddError1['description'] = "prompt_type参数不符合规范"
        PromptAddError1['detail'] = "当前参数仅支持 chat|completion"
        return [400, PromptAddError1]
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，\na-zA-Z0-9\u4e00-\u9fa5]{,255}$',
                     para["prompt_desc"]):
        PromptAddError1['description'] = "prompt_desc参数不符合规范"
        PromptAddError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过255字符"
        return [400, PromptAddError1]
    name = para['prompt_name']
    # info1 = await PromptItemList.filter(f_prompt_item_type_id=para['prompt_item_type_id'])
    info1 = prompt_dao.get_data_from_prompt_item_list_by_type_id(para['prompt_item_type_id'])
    if info1 == () or info1[0]["f_item_is_delete"] == 1:
        PromptAddError1['description'] = "prompt_item_type_id参数错误"
        PromptAddError1['detail'] = "提示词项目分类不存在"
        return [400, PromptAddError1]
    # info = await PromptItemList.filter(f_prompt_item_type_id=para['prompt_item_type_id'],
    #                                    f_prompt_item_id=para['prompt_item_id'])
    info = prompt_dao.get_data_from_prompt_item_list_by_id_and_type_id(para["prompt_item_id"],
                                                                       para["prompt_item_type_id"])
    if info == () or info[0]["f_item_is_delete"] == 1:
        PromptAddError1['description'] = "prompt_item_id参数错误"
        PromptAddError1['detail'] = "提示词项目不存在"
        return [500, PromptAddError1]
    prompt_name_list = [ids["f_prompt_name"] for ids in
                        prompt_dao.get_data_from_prompt_list_by_item_type_id(para["prompt_item_type_id"])]
    if name in prompt_name_list:
        PromptAddError2['description'] = "prompt_name参数错误"
        PromptAddError2['detail'] = "提示词名称重复"
        return [500, PromptAddError2]

    if para["model_id"] != "":
        ids_list = [ids["f_model_id"] for ids in llm_model_dao.get_all_model_list()]
        info2 = llm_model_dao.get_data_from_model_list_by_id(para["model_id"])
        if para["model_id"] not in ids_list:
            PromptAddError1['description'] = "参数错误"
            PromptAddError1['detail'] = "选择的大模型不存在"
            return [400, PromptAddError1]
        model = info2[0]["f_model"]
        if model == 'gpt-35-turbo-16k':
            if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                    para["model_para"][
                        "max_tokens"] > 16384 or para["model_para"]["max_tokens"] < 10:
                PromptAddError1['description'] = "max_tokens参数错误"
                PromptAddError1['detail'] = "该模型当前参数仅支持10~16384的整数"
                return [400, PromptAddError1]
        if model == 'text-davinci-002':
            if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                    para["model_para"][
                        "max_tokens"] > 4097 or para["model_para"]["max_tokens"] < 10:
                PromptAddError1['description'] = "max_tokens参数错误"
                PromptAddError1['detail'] = "该模型当前参数仅支持10~4097的整数"
                return [400, PromptAddError1]
        if model == 'baichuan2':
            if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                    para["model_para"][
                        "max_tokens"] > 4096 or para["model_para"]["max_tokens"] < 10:
                PromptAddError1['description'] = "max_tokens参数错误"
                PromptAddError1['detail'] = "该模型当前参数仅支持10~4096的整数"
                return [400, PromptAddError1]
        else:
            if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int:
                PromptAddError1['description'] = "max_tokens参数错误"
                PromptAddError1['detail'] = "max_tokens仅支持int类型"
                return [400, PromptAddError1]

    if type(para["icon"]) != str:
        PromptAddError1['description'] = "icon参数错误"
        PromptAddError1['detail'] = "当前参数仅支持字符串类型"
        return [400, PromptAddError1]
    icon = int(para["icon"])
    if icon not in list(range(0, 10)):
        PromptAddError1['description'] = "icon参数错误"
        PromptAddError1['detail'] = "颜色配置错误"
        return [400, PromptAddError1]
    for ceil in para["variables"]:
        if ceil["field_type"] == 'text':
            if "max_len" not in ceil or ceil["max_len"] < 1 or ceil["max_len"] > 256:
                PromptAddError1['description'] = "text参数错误"
                PromptAddError1['detail'] = "当前参数仅支持大于1且小于256"
                return [400, PromptAddError1]
        if ceil["field_type"] == 'number':
            if "range" not in ceil or len(ceil["range"]) != 2:
                PromptAddError1['description'] = "number参数错误"
                PromptAddError1['detail'] = "当前参数仅支持整型"
                return [400, PromptAddError1]
            if ceil["value_type"] == 'i':
                if (ceil["range"][0] == None or type(ceil["range"][0]) == int) and (
                        ceil["range"][1] == None or type(ceil["range"][1]) == int):
                    pass
                else:
                    PromptAddError1['description'] = "number参数错误"
                    PromptAddError1['detail'] = "当前参数仅支持整型"
                    return [400, PromptAddError1]
            if ceil["value_type"] == 'f':
                if (ceil["range"][0] == None or isinstance(ceil["range"][0], (int, float))) and (
                        ceil["range"][1] == None or isinstance(ceil["range"][1], (int, float))):
                    pass
                else:
                    PromptAddError1['description'] = "number参数错误"
                    PromptAddError1['detail'] = "当前参数仅支持数字"
                    return [400, PromptAddError1]

    if len(para["model_para"]) > 0:
        if type(para["model_para"]) != dict:
            PromptAddError1['description'] = "model_para参数错误"
            PromptAddError1['detail'] = "当前参数仅支持字典类型"
            return [400, PromptAddError1]
        if "temperature" not in para["model_para"] or not isinstance(para["model_para"]["temperature"],
                                                                     (int, float)) or isinstance(
            para["model_para"]["temperature"], Fraction) or para["model_para"]["temperature"] > 2 or para["model_para"][
            "temperature"] < 0:
            PromptAddError1['description'] = "temperature参数错误"
            PromptAddError1['detail'] = "当前参数仅支持0~2的小数"
            return [400, PromptAddError1]
        if "top_p" not in para["model_para"] or not isinstance(para["model_para"]["top_p"], (int, float)) or isinstance(
                para["model_para"]["top_p"], Fraction) or para["model_para"]["top_p"] > 1 or para["model_para"][
            "top_p"] < 0:
            PromptAddError1['description'] = "top_p参数错误"
            PromptAddError1['detail'] = "当前参数仅支持0~1的小数"
            return [400, PromptAddError1]
        if "presence_penalty" not in para["model_para"] or not isinstance(para["model_para"]["presence_penalty"],
                                                                          (int, float)) or isinstance(
            para["model_para"]["presence_penalty"], Fraction) or para["model_para"]["presence_penalty"] > 2 or \
                para["model_para"]["presence_penalty"] < -2:
            PromptAddError1['description'] = "presence_penalty参数错误"
            PromptAddError1['detail'] = "当前参数仅支持-2~2的小数"
            return [400, PromptAddError1]
        if "frequency_penalty" not in para["model_para"] or not isinstance(para["model_para"]["frequency_penalty"],
                                                                           (int, float)) or isinstance(
            para["model_para"]["frequency_penalty"], Fraction) or para["model_para"]["frequency_penalty"] > 2 or \
                para["model_para"]["frequency_penalty"] < -2:
            PromptAddError1['description'] = "frequency_penalty参数错误"
            PromptAddError1['detail'] = "当前参数仅支持-2~2的小数"
            return [400, PromptAddError1]


# 提示词名称编辑
async def prompt_name_verify(model_para):
    if "prompt_id" not in model_para or "prompt_name" not in model_para or "model_id" not in model_para or "icon" not in model_para:
        PromptNameEditError1['description'] = "参数错误"
        PromptNameEditError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptNameEditError1]
    prompt_id_list = [ids["f_prompt_id"] for ids in prompt_dao.get_all_data_from_prompt_list()]
    info = prompt_dao.get_prompt_by_id(model_para["prompt_id"])
    if model_para["prompt_id"] not in prompt_id_list:
        PromptNameEditError1['description'] = "参数错误"
        PromptNameEditError1['detail'] = "提示词不存在"
        return [400, PromptNameEditError1]
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$',
                     model_para["prompt_name"]) or len(model_para["prompt_name"].replace(' ', '')) == 0:
        PromptNameEditError1['description'] = "prompt_name参数不符合规范"
        PromptNameEditError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过50字符"
        return [400, PromptNameEditError1]
    prompt_name_list = [ids["f_prompt_name"] for ids in
                        prompt_dao.get_data_from_prompt_list_by_item_type_id(info[0]["f_prompt_item_type_id"])]
    if model_para["prompt_name"] in prompt_name_list and model_para["prompt_name"] != info[0]["f_prompt_name"]:
        PromptNameEditError2['description'] = "参数错误"
        PromptNameEditError2['detail'] = "提示词名称重复"
        return [500, PromptNameEditError2]

    ids_list = [ids["f_model_id"] for ids in llm_model_dao.get_all_model_list()]
    info1 = llm_model_dao.get_data_from_model_list_by_id(model_para['model_id'])
    if model_para["model_id"] != "" and model_para["model_id"] not in ids_list:
        PromptNameEditError1['description'] = "参数错误"
        PromptNameEditError1['detail'] = "选择的大模型不存在"
        return [500, PromptNameEditError1]
    if type(model_para["icon"]) != str:
        PromptNameEditError1['description'] = "icon参数错误"
        PromptNameEditError1['detail'] = "当前参数仅支持字符串类型"
        return [400, PromptNameEditError1]
    elif model_para["icon"] == "":
        PromptNameEditError1['description'] = "icon参数错误"
        PromptNameEditError1['detail'] = "当前参数不可为空"
        return [400, PromptNameEditError1]
    icon = int(model_para["icon"])
    if icon not in list(range(0, 10)):
        PromptNameEditError1['description'] = "参数错误"
        PromptNameEditError1['detail'] = "颜色配置错误"
        return [400, PromptNameEditError1]
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，\na-zA-Z0-9\u4e00-\u9fa5]{,255}$',
                     model_para["prompt_desc"]):
        PromptNameEditError1['description'] = "prompt_desc参数不符合规范"
        PromptNameEditError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过255字符"
        return [400, PromptNameEditError1]



# 提示词编辑
async def prompt_edit_verify(para):
    if "prompt_id" not in para or "model_para" not in para or "messages" not in para or "model_id" not in para:
        PromptEditError1['description'] = "参数错误"
        PromptEditError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptEditError1]
    prompt_id_list = [ids["f_prompt_id"] for ids in prompt_dao.get_all_data_from_prompt_list()]
    info = prompt_dao.get_prompt_by_id(para["prompt_id"])
    if para["prompt_id"] not in prompt_id_list:
        PromptEditError1['description'] = "参数错误"
        PromptEditError1['detail'] = "提示词不存在"
        return [400, PromptEditError1]
    ids_list = [ids["f_model_id"] for ids in llm_model_dao.get_all_model_list()]
    info2 = llm_model_dao.get_data_from_model_list_by_id(para["model_id"])
    if para["model_id"] != "" and (para["model_id"] not in ids_list):
        PromptEditError1['description'] = "参数错误"
        PromptEditError1['detail'] = "选择的大模型不存在"
        return [400, PromptEditError1]
    for ceil in para["variables"]:
        if ceil["field_type"] == 'text':
            if "max_len" not in ceil or ceil["max_len"] < 1 or ceil["max_len"] > 256:
                PromptEditError1['description'] = "text参数错误"
                PromptEditError1['detail'] = "当前参数仅支持大于1且小于256"
                return [400, PromptEditError1]
        if ceil["field_type"] == 'number':
            if "range" not in ceil or len(ceil["range"]) != 2:
                PromptEditError1['description'] = "number参数错误"
                PromptEditError1['detail'] = "当前参数仅支持整型"
                return [400, PromptEditError1]
            if ceil["value_type"] == 'i':
                if (ceil["range"][0] == None or type(ceil["range"][0]) == int) and (
                        ceil["range"][1] == None or type(ceil["range"][1]) == int):
                    pass
                else:
                    PromptEditError1['description'] = "number参数错误"
                    PromptEditError1['detail'] = "当前参数仅支持整型"
                    return [400, PromptEditError1]
            if ceil["value_type"] == 'f':
                if (ceil["range"][0] == None or isinstance(ceil["range"][0], (int, float))) and (
                        ceil["range"][1] == None or isinstance(ceil["range"][1], (int, float))):
                    pass
                else:
                    PromptEditError1['description'] = "number参数错误"
                    PromptEditError1['detail'] = "当前参数仅支持数字"
                    return [400, PromptEditError1]
    if type(para["model_para"]) != dict:
        PromptEditError1['description'] = "model_para参数错误"
        PromptEditError1['detail'] = "当前参数仅支持字典类型"
        return [400, PromptEditError1]
    if para["model_para"] == {}:
        return
    if "temperature" not in para["model_para"] or not isinstance(para["model_para"]["temperature"],
                                                                 (int, float)) or isinstance(
        para["model_para"]["temperature"], Fraction) or para["model_para"]["temperature"] > 2 or para["model_para"][
        "temperature"] < 0:
        PromptEditError1['description'] = "temperature参数错误"
        PromptEditError1['detail'] = "当前参数仅支持0~2的小数"
        return [400, PromptEditError1]
    if "top_p" not in para["model_para"] or not isinstance(para["model_para"]["top_p"], (int, float)) or isinstance(
            para["model_para"]["top_p"], Fraction) or para["model_para"]["top_p"] > 1 or para["model_para"][
        "top_p"] < 0:
        PromptEditError1['description'] = "top_p参数错误"
        PromptEditError1['detail'] = "当前参数仅支持0~1的小数"
        return [400, PromptEditError1]
    if "presence_penalty" not in para["model_para"] or not isinstance(para["model_para"]["presence_penalty"],
                                                                      (int, float)) or isinstance(
        para["model_para"]["presence_penalty"], Fraction) or para["model_para"]["presence_penalty"] > 2 or \
            para["model_para"]["presence_penalty"] < -2:
        PromptEditError1['description'] = "presence_penalty参数错误"
        PromptEditError1['detail'] = "当前参数仅支持-2~2的小数"
        return [400, PromptEditError1]
    if "frequency_penalty" not in para["model_para"] or not isinstance(para["model_para"]["frequency_penalty"],
                                                                       (int, float)) or isinstance(
        para["model_para"]["frequency_penalty"], Fraction) or para["model_para"]["frequency_penalty"] > 2 or \
            para["model_para"]["frequency_penalty"] < -2:
        PromptEditError1['description'] = "frequency_penalty参数错误"
        PromptEditError1['detail'] = "当前参数仅支持-2~2的小数"
        return [400, PromptEditError1]
    try:
        model = info2[0]["f_model"]
    except Exception:
        PromptEditError1["description"] = "model_id参数错误"
        PromptEditError1["detail"] = "当model_para传入时，model_id也需要传入"
        return [400, PromptEditError1]
    if model == 'gpt-35-turbo-16k':
        if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                para["model_para"][
                    "max_tokens"] > 16384 or para["model_para"]["max_tokens"] < 10:
            PromptEditError1['description'] = "max_tokens参数错误"
            PromptEditError1['detail'] = "该模型当前参数仅支持10~16384的整数"
            return [400, PromptEditError1]
    if model == 'text-davinci-002':
        if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                para["model_para"][
                    "max_tokens"] > 4097 or para["model_para"]["max_tokens"] < 10:
            PromptEditError1['description'] = "max_tokens参数错误"
            PromptEditError1['detail'] = "该模型当前参数仅支持10~4097的整数"
            return [400, PromptEditError1]
    if model == 'baichuan2':
        if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                para["model_para"][
                    "max_tokens"] > 4096 or para["model_para"]["max_tokens"] < 10:
            PromptEditError1['description'] = "max_tokens参数错误"
            PromptEditError1['detail'] = "该模型当前参数仅支持10~4096的整数"
            return [400, PromptEditError1]


# 提示词管理编辑提示词
async def prompt_template_edit_verify(para):
    if "prompt_id" not in para or "prompt_name" not in para or "messages" not in para or "variables" not in para or\
            "icon" not in para or "prompt_item_type_id" not in para or "prompt_item_id" not in para:
        PromptTemplateEditError1['description'] = "参数错误"
        PromptTemplateEditError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptTemplateEditError1]

    # 校验prompt_id
    prompt_id = para["prompt_id"]
    if not isinstance(prompt_id, str) or prompt_id == "":
        PromptTemplateEditError1['description'] = "参数错误"
        PromptTemplateEditError1['detail'] = "提示词id必须为字符串类型且不能为空"
        return [400, PromptTemplateEditError1]
    prompt_id_list = [ids["f_prompt_id"] for ids in prompt_dao.get_all_data_from_prompt_list()]
    info = prompt_dao.get_prompt_by_id(para["prompt_id"])
    if para["prompt_id"] not in prompt_id_list:
        PromptTemplateEditError1['description'] = "参数错误"
        PromptTemplateEditError1['detail'] = "提示词不存在"
        return [400, PromptTemplateEditError1]

    # 校验variables
    if not isinstance(para["variables"], list):
        PromptTemplateEditError1['description'] = "variables参数不符合规范"
        PromptTemplateEditError1['detail'] = "当前参数必须为list类型"
        return [400, PromptTemplateEditError1]
    for ceil in para["variables"]:
        if ceil["field_type"] == 'text':
            if "max_len" not in ceil or ceil["max_len"] < 1 or ceil["max_len"] > 256:
                PromptEditError1['description'] = "text参数错误"
                PromptEditError1['detail'] = "当前参数仅支持大于1且小于256"
                return [400, PromptEditError1]
        if ceil["field_type"] == 'number':
            if "range" not in ceil or len(ceil["range"]) != 2:
                PromptEditError1['description'] = "number参数错误"
                PromptEditError1['detail'] = "当前参数仅支持整型"
                return [400, PromptEditError1]
            if ceil["value_type"] == 'i':
                if (ceil["range"][0] == None or type(ceil["range"][0]) == int) and (
                        ceil["range"][1] == None or type(ceil["range"][1]) == int):
                    pass
                else:
                    PromptEditError1['description'] = "number参数错误"
                    PromptEditError1['detail'] = "当前参数仅支持整型"
                    return [400, PromptEditError1]
            if ceil["value_type"] == 'f':
                if (ceil["range"][0] == None or isinstance(ceil["range"][0], (int, float))) and (
                        ceil["range"][1] == None or isinstance(ceil["range"][1], (int, float))):
                    pass
                else:
                    PromptEditError1['description'] = "number参数错误"
                    PromptEditError1['detail'] = "当前参数仅支持数字"
                    return [400, PromptEditError1]

    # 校验prompt_name
    if not isinstance(para["prompt_name"], str):
        PromptTemplateEditError1['description'] = "prompt_name参数不符合规范"
        PromptTemplateEditError1['detail'] = "当前参数必须为字符串类型"
        return [400, PromptTemplateEditError1]

    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]{,50}$',
                     para["prompt_name"]) or len(para["prompt_name"].replace(' ', '')) == 0:
        PromptTemplateEditError1['description'] = "prompt_name参数不符合规范"
        PromptTemplateEditError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过50字符"
        return [400, PromptTemplateEditError1]

    # messages
    if not isinstance(para["messages"], str):
        PromptTemplateEditError1['description'] = "messages参数不符合规范"
        PromptTemplateEditError1['detail'] = "当前参数必须为字符串类型"
        return [400, PromptTemplateEditError1]

    # opening_remarks
    if not isinstance(para["opening_remarks"], str):
        PromptTemplateEditError1['description'] = "opening_remarks参数不符合规范"
        PromptTemplateEditError1['detail'] = "当前参数必须为字符串类型"
        return [400, PromptTemplateEditError1]

    # icon
    if type(para["icon"]) != str or para["icon"] == "":
        PromptTemplateEditError1['description'] = "icon参数错误"
        PromptTemplateEditError1['detail'] = "当前参数仅支持字符串类型且不能为空"
        return [400, PromptTemplateEditError1]
    icon = int(para["icon"])
    if icon not in list(range(0, 10)):
        PromptTemplateEditError1['description'] = "参数错误"
        PromptTemplateEditError1['detail'] = "颜色配置错误"
        return [400, PromptTemplateEditError1]

    # 校验prompt_desc
    if not isinstance(para["prompt_desc"], str):
        PromptTemplateEditError1['description'] = "prompt_desc参数不符合规范"
        PromptTemplateEditError1['detail'] = "当前参数必须为字符串类型"
        return [400, PromptTemplateEditError1]
    if not re.search(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！￥…·（）—。【 】‘“’”：；、《》？，\na-zA-Z0-9\u4e00-\u9fa5]{,255}$',
                     para["prompt_desc"]):
        PromptTemplateEditError1['description'] = "prompt_desc参数不符合规范"
        PromptTemplateEditError1['detail'] = "当前参数仅支中英文、数字以及键盘上的特殊字符,且不超过255字符"
        return [400, PromptTemplateEditError1]

    # prompt_item_type_id
    if not isinstance(para["prompt_item_type_id"], str) or para["prompt_item_type_id"] == "":
        PromptTemplateEditError1['description'] = "prompt_item_type_id参数不符合规范"
        PromptTemplateEditError1['detail'] = "当前参数必须为字符串类型且不能为空"
        return [400, PromptTemplateEditError1]

    # prompt_item_id
    if not isinstance(para["prompt_item_id"], str) or para["prompt_item_id"] == "":
        PromptTemplateEditError1['description'] = "prompt_item_id参数不符合规范"
        PromptTemplateEditError1['detail'] = "当前参数必须为字符串类型且不能为空"
        return [400, PromptTemplateEditError1]

    if len(prompt_dao.check_item_and_item_type(para["prompt_item_id"], para["prompt_item_type_id"])) == 0:
        PromptTemplateEditError1['description'] = "参数错误"
        PromptTemplateEditError1['detail'] = "prompt_item_id或prompt_item_type_id不存在"
        return [400, PromptTemplateEditError1]


# 提示词发布
async def prompt_deploy_verify(model_para):
    if "prompt_id" not in model_para:
        PromptDeployError1['description'] = "参数错误"
        PromptDeployError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptDeployError1]
    prompt_id_list = [ids["f_prompt_id"] for ids in prompt_dao.get_all_data_from_prompt_list()]
    info = prompt_dao.get_prompt_by_id(model_para["prompt_id"])
    if model_para["prompt_id"] not in prompt_id_list:
        PromptDeployError1['description'] = "参数错误"
        PromptDeployError1['detail'] = "提示词不存在"
        return [400, PromptDeployError1]
    # if info[0].f_is_deploy is True:
    #     PromptDeployError1['description'] = "参数错误"
    #     PromptDeployError1['detail'] = "该提示词已发布"
    #     return [400, PromptDeployError1]


# 提示词取消发布
async def prompt_undeploy_verify(model_para):
    if "prompt_id" not in model_para:
        PromptUndeployError1['description'] = "参数错误"
        PromptUndeployError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptUndeployError1]
    prompt_id_list = [ids["f_prompt_id"] for ids in prompt_dao.get_all_data_from_prompt_list()]
    info = prompt_dao.get_prompt_by_id(model_para["prompt_id"])
    if model_para["prompt_id"] not in prompt_id_list:
        PromptUndeployError1['description'] = "参数错误"
        PromptUndeployError1['detail'] = "提示词不存在"
        return [400, PromptUndeployError1]
    if info[0]["f_is_deploy"] == 0:
        PromptDeployError1['description'] = "参数错误"
        PromptDeployError1['detail'] = "该提示词未发布"
        return [400, PromptUndeployError1]


# 提示词运行
async def prompt_run_verify(para):
    if "model_id" not in para or "model_para" not in para:
        PromptRunError1['description'] = "参数错误"
        PromptRunError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptRunError1]
    model_list = llm_model_dao.get_all_model_list()
    ids_list = [ids["f_model_id"] for ids in model_list]
    info2 = {}
    for model in model_list:
        if model["f_model_id"] == para["model_id"]:
            info2 = model
            break
    if para["model_id"] not in ids_list:
        PromptRunError1['description'] = "参数错误"
        PromptRunError1['detail'] = "选择的大模型不存在"
        return [400, PromptRunError1]
    # if len(para["messages"].replace(' ', '')) == 0 or len(para["messages"]) > 5000:
    #     PromptRunError1['description'] = "messages参数错误"
    #     PromptRunError1['detail'] = "messages为空或超上限"
    #     return [400, PromptRunError1]

    for var in para["variables"]:  # variables prompt里面的变量，inputs是输入的变量
        var_name = var['var_name']
        # 变量为必填字段，但是输入中没有当前变量
        if not var['optional'] and var_name not in para["inputs"].keys():
            PromptRunError1['description'] = "提示词变量输入异常"
            PromptRunError1['detail'] = "必填项{}缺失".format(var_name)
            return [500, PromptRunError1]
        # 变量为必填或非必填字段，输入中存在当前变量
        if var_name in para["inputs"].keys():
            var_value = para["inputs"][var_name]
            # 校验 <文本型> 变量
            if var['field_type'] == 'text':
                if type(var_value) is not str:
                    PromptRunError1['description'] = "提示词变量输入异常"
                    PromptRunError1['detail'] = f"字段{var_name}仅支持字符型"
                    return [500, PromptRunError1]
                if len(var_value) > var['max_len']:
                    PromptRunError1['description'] = "提示词变量输入异常"
                    PromptRunError1['detail'] = f"字段{var_name}长度最大值为{var['max_len']}"
                    return [500, PromptRunError1]
            # 校验 <数值型> 变量
            if var['field_type'] == 'number':
                if var['value_type'] == 'i' and type(var_value) is not int:
                    PromptRunError1['description'] = "提示词变量输入异常"
                    PromptRunError1['detail'] = f"字段{var_name}仅支持整数"
                    return [500, PromptRunError1]
                if var['value_type'] == 'f':
                    if type(var_value) is not float and type(var_value) is not int:
                        PromptRunError1['description'] = "提示词变量输入异常"
                        PromptRunError1['detail'] = f"字段{var_name}仅支持浮点数"
                        return [500, PromptRunError1]
                if var['range'][0] is not None:
                    if var_value < var['range'][0]:
                        PromptRunError1['description'] = "提示词变量输入异常"
                        PromptRunError1['detail'] = f"字段{var_name}的取值范围为{var['range']}"
                        return [500, PromptRunError1]
                if var['range'][1] is not None:
                    if var_value > var['range'][1]:
                        PromptRunError1['description'] = "提示词变量输入异常"
                        PromptRunError1['detail'] = f"字段{var_name}的取值范围为{var['range']}"
                        return [500, PromptRunError1]
            # 校验 <下拉型> 变量
            if var['field_type'] == 'selector':
                if var_value not in var['options']:
                    PromptRunError1['description'] = "提示词变量输入异常"
                    PromptRunError1['detail'] = f"字段{var_name}的仅支持{var['options']}"
                    return PromptUsed
    if type(para["history_dia"]) != list:
        PromptRunError1['description'] = "历史对话变量输入异常"
        PromptRunError1['detail'] = "历史对话变量格式异常"
        return [400, PromptRunError1]
    from app.utils.llm_utils import model_config, get_context_size
    if para["history_dia"]:
        from app.utils.llm_utils import get_context_size
        if info2["f_model_series"].lower() == "aishu":
            config = json.loads(info2["f_model_config"].replace("'", '"'))
            try:
                model_config.add_model_context_size(info2["f_model_id"], info2["f_model_name"], get_context_size(info2["f_model_series"], config["api_base"], config["api_model"]))
            except Exception:
                model_config.init_model_config()
                model_config.add_model_context_size(info2["f_model_id"], info2["f_model_name"], get_context_size(info2["f_model_series"], config["api_base"], config["api_model"]))
        if info2["f_model_series"].lower() in ["others", "claude"]:
            try:
                model_config.add_model_context_size(info2["f_model_id"], info2["f_model_name"], get_context_size("others", "", info2["f_model"]))
            except Exception:
                model_config.init_model_config()
                model_config.add_model_context_size(info2["f_model_id"], info2["f_model_name"], get_context_size("others", "", info2["f_model"]))
    if type(para["model_para"]) != dict:
        PromptAddError1['description'] = "model_para参数错误"
        PromptAddError1['detail'] = "当前参数仅支持字典类型"
        return [400, PromptAddError1]
    if "temperature" not in para["model_para"] or not isinstance(para["model_para"]["temperature"],
                                                                 (int, float)) or isinstance(
        para["model_para"]["temperature"], Fraction) or para["model_para"]["temperature"] > 2 or para["model_para"][
        "temperature"] < 0:
        PromptAddError1['description'] = "temperature参数错误"
        PromptAddError1['detail'] = "当前参数仅支持0~2的小数"
        return [400, PromptAddError1]
    if "top_p" not in para["model_para"] or not isinstance(para["model_para"]["top_p"], (int, float)) or isinstance(
            para["model_para"]["top_p"], Fraction) or para["model_para"]["top_p"] > 1 or para["model_para"][
        "top_p"] < 0:
        PromptAddError1['description'] = "top_p参数错误"
        PromptAddError1['detail'] = "当前参数仅支持0~1的小数"
        return [400, PromptAddError1]
    if "presence_penalty" not in para["model_para"] or not isinstance(para["model_para"]["presence_penalty"],
                                                                      (int, float)) or isinstance(
        para["model_para"]["presence_penalty"], Fraction) or para["model_para"]["presence_penalty"] > 2 or \
            para["model_para"]["presence_penalty"] < -2:
        PromptAddError1['description'] = "presence_penalty参数错误"
        PromptAddError1['detail'] = "当前参数仅支持-2~2的小数"
        return [400, PromptAddError1]
    if "frequency_penalty" not in para["model_para"] or not isinstance(para["model_para"]["frequency_penalty"],
                                                                       (int, float)) or isinstance(
        para["model_para"]["frequency_penalty"], Fraction) or para["model_para"]["frequency_penalty"] > 2 or \
            para["model_para"]["frequency_penalty"] < -2:
        PromptAddError1['description'] = "frequency_penalty参数错误"
        PromptAddError1['detail'] = "当前参数仅支持-2~2的小数"
        return [400, PromptAddError1]

    if "max_tokens" not in para["model_para"] or not isinstance(para["model_para"]["frequency_penalty"],
                                                                       (int, float)) or isinstance(
        para["model_para"]["frequency_penalty"], Fraction) or para["model_para"]["frequency_penalty"] > 2 or \
            para["model_para"]["frequency_penalty"] < -2:
        PromptAddError1['description'] = "frequency_penalty参数错误"
        PromptAddError1['detail'] = "当前参数仅支持-2~2的小数"
        return [400, PromptAddError1]
    model = info2["f_model"]
    if model == 'gpt-35-turbo-16k':
        if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                para["model_para"][
                    "max_tokens"] > 16384 or para["model_para"]["max_tokens"] < 10:
            PromptAddError1['description'] = "max_tokens参数错误"
            PromptAddError1['detail'] = "该模型当前参数仅支持10~16384的整数"
            return [400, PromptAddError1]
    if model == 'text-davinci-002':
        if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                para["model_para"][
                    "max_tokens"] > 4097 or para["model_para"]["max_tokens"] < 10:
            PromptAddError1['description'] = "max_tokens参数错误"
            PromptAddError1['detail'] = "该模型当前参数仅支持10~4097的整数"
            return [400, PromptAddError1]
    if model == 'baichuan2':
        if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                para["model_para"][
                    "max_tokens"] > 4096 or para["model_para"]["max_tokens"] < 10:
            PromptAddError1['description'] = "max_tokens参数错误"
            PromptAddError1['detail'] = "该模型当前参数仅支持10~4096的整数"
            return [400, PromptAddError1]


def prompt_template_run_verify(para):
    from app.utils.llm_utils import model_config, get_context_size
    if "model_name" not in para or "model_para" not in para or "prompt_id" not in para:
        PromptTemplateRunError1['description'] = "参数错误"
        PromptTemplateRunError1['detail'] = "传入参数名称或数量错误"
        return [400, PromptTemplateRunError1]
    prompt_id = para["prompt_id"]
    if not isinstance(prompt_id, str) or prompt_id == "":
        PromptTemplateRunError1['description'] = "参数错误"
        PromptTemplateRunError1['detail'] = "提示词id必须为字符串类型且不能为空"
        return [400, PromptTemplateRunError1]

    prompt = prompt_dao.get_prompt_by_id(para["prompt_id"])
    if len(prompt) == 0:
        PromptTemplateRunError1['description'] = "参数错误"
        PromptTemplateRunError1['detail'] = "提示词id不存在"
        return [400, PromptTemplateRunError1]

    model_list = llm_model_dao.get_all_model_list()
    info2 = {}
    for model in model_list:
        if model["f_model_name"] == para["model_name"]:
            info2 = model
            break

    if type(para["history_dia"]) != list:
        PromptTemplateRunError1['description'] = "历史对话变量输入异常"
        PromptTemplateRunError1['detail'] = "历史对话变量格式异常"
        return [400, PromptTemplateRunError1]
    if "context_size" not in model_config.get_model_config("id", model["f_model_id"]).keys():
        config = json.loads(model["f_model_config"].replace("'", '"'))
        if "api_url" in config.keys():
            config["api_base"] = config["api_url"]
        try:
            try:
                context_size = get_context_size(model["f_model_series"], config["api_base"], config["api_model"])
            except func_timeout.exceptions.FunctionTimedOut:
                context_size = 4096
        except Exception as e:
            context_size = 4096
        try:
            model_config.add_model_context_size(model["f_model_id"], model["f_model_name"], context_size)
        except Exception:
            model_config.init_model_config()
            model_config.add_model_context_size(model["f_model_id"], model["f_model_name"], context_size)
    max_tokens = model_config.get_model_config("name", para["model_name"])["context_size"]
    if para["history_dia"]:

        history_str = ''
        for cell in para["history_dia"]:
            history_str += cell['message']
        if len(history_str) > max_tokens:
            PromptTemplateRunError1['description'] = "历史信息字数超过限制"
            PromptTemplateRunError1['detail'] = "历史信息字数不超过{}字符".format(max_tokens)
            return [400, PromptTemplateRunError1]
    if type(para["model_para"]) != dict:
        PromptTemplateRunError1['description'] = "model_para参数错误"
        PromptTemplateRunError1['detail'] = "当前参数仅支持字典类型"
        return [400, PromptTemplateRunError1]
    if "temperature" not in para["model_para"] or not isinstance(para["model_para"]["temperature"],
                                                                 (int, float)) or isinstance(
        para["model_para"]["temperature"], Fraction) or para["model_para"]["temperature"] > 2 or para["model_para"][
        "temperature"] < 0:
        PromptTemplateRunError1['description'] = "temperature参数错误"
        PromptTemplateRunError1['detail'] = "当前参数仅支持0~2的小数"
        return [400, PromptTemplateRunError1]
    if "top_p" not in para["model_para"] or not isinstance(para["model_para"]["top_p"], (int, float)) or isinstance(
            para["model_para"]["top_p"], Fraction) or para["model_para"]["top_p"] > 1 or para["model_para"][
        "top_p"] < 0:
        PromptTemplateRunError1['description'] = "top_p参数错误"
        PromptTemplateRunError1['detail'] = "当前参数仅支持0~1的小数"
        return [400, PromptTemplateRunError1]
    if "top_k" in para["model_para"] and (not isinstance(para["model_para"]["top_k"], int) or isinstance(
            para["model_para"]["top_p"], Fraction) or para["model_para"]["top_k"] < 1):
        PromptTemplateRunError1['description'] = "top_k参数错误"
        PromptTemplateRunError1['detail'] = "当前参数仅支持大于等于1的整数"
        return [400, PromptTemplateRunError1]
    if "presence_penalty" not in para["model_para"] or not isinstance(para["model_para"]["presence_penalty"],
                                                                      (int, float)) or isinstance(
        para["model_para"]["presence_penalty"], Fraction) or para["model_para"]["presence_penalty"] > 2 or \
            para["model_para"]["presence_penalty"] < -2:
        PromptTemplateRunError1['description'] = "presence_penalty参数错误"
        PromptTemplateRunError1['detail'] = "当前参数仅支持-2~2的小数"
        return [400, PromptTemplateRunError1]
    if "frequency_penalty" not in para["model_para"] or not isinstance(para["model_para"]["frequency_penalty"],
                                                                       (int, float)) or isinstance(
        para["model_para"]["frequency_penalty"], Fraction) or para["model_para"]["frequency_penalty"] > 2 or \
            para["model_para"]["frequency_penalty"] < -2:
        PromptTemplateRunError1['description'] = "frequency_penalty参数错误"
        PromptTemplateRunError1['detail'] = "当前参数仅支持-2~2的小数"
        return [400, PromptTemplateRunError1]

    model = info2.get("f_model", "")
    if model == 'gpt-35-turbo-16k':
        if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                para["model_para"][
                    "max_tokens"] > 16384 or para["model_para"]["max_tokens"] < 10:
            PromptAddError1['description'] = "max_tokens参数错误"
            PromptAddError1['detail'] = "该模型当前参数仅支持10~16384的整数"
            return [400, PromptAddError1]
    else:
        if "max_tokens" not in para["model_para"] or type(para["model_para"]["max_tokens"]) != int or \
                para["model_para"][
                    "max_tokens"] > max_tokens or para["model_para"]["max_tokens"] < 10:
            PromptAddError1['description'] = "max_tokens参数错误"
            PromptAddError1['detail'] = "该模型当前参数仅支持10~" + str(max_tokens) + "的整数"
            return [400, PromptAddError1]


async def completion_prompt_verify(prompt_id, inputs):
    prompt_id_list = [cell["f_prompt_id"] for cell in prompt_dao.get_all_data_from_prompt_list()]
    if prompt_id not in prompt_id_list:
        PromptConpletionError1['description'] = "prompt_id 参数不符合规范"
        PromptConpletionError1['detail'] = "当前参数仅支持数据库中已有prompt_id"
        return PromptConpletionError1
    info = prompt_dao.get_prompt_by_id(prompt_id)
    messages = info[0]["f_messages"]
    var = re.findall(r'\{\{(.*?)\}\}', messages)  # 提示词中的变量
    inputs_key = inputs.keys()
    if set(var) != set(inputs_key):
        PromptConpletionError1['description'] = "提示词变量输入异常"
        PromptConpletionError1['detail'] = "inputs的变量与提示词中的变量不一致"
        return PromptConpletionError1


# 查看代码
async def prompt_code_verify(model_id, prompt_id):
    ids_list = [ids["f_model_id"] for ids in llm_model_dao.get_all_model_list()]
    is_delete = llm_model_dao.get_data_from_model_list_by_id(model_id)
    if model_id not in ids_list or len(model_id.replace(' ', '')) == 0:
        PromptCodeError1['description'] = "参数错误"
        PromptCodeError1['detail'] = "选择的大模型不存在"
        return PromptCodeError1
    if len(prompt_id.replace(' ', '')) != 0:
        prompt_id_list = [ids["f_prompt_id"] for ids in prompt_dao.get_all_data_from_prompt_list()]
        info = prompt_dao.get_prompt_by_id(prompt_id)
        if prompt_id not in prompt_id_list:
            PromptCodeError1['description'] = "参数错误"
            PromptCodeError1['detail'] = "提示词不存在"
            return PromptCodeError1


# 删除功能
async def prompt_delete_verify(delete_id):
    if len(delete_id.items()) != 1:
        PromptDeleteError1['description'] = "参数错误"
        PromptDeleteError1['detail'] = "传入参数数量仅支持一个"
        return [500, PromptDeleteError1]
    key_list = list(delete_id.keys())
    if not re.search(r'^(prompt_id|type_id|item_id|prompt_id_list)$', key_list[0]):
        PromptDeleteError1['description'] = "参数错误"
        PromptDeleteError1['detail'] = "传入参数仅支持prompt_id|type_id|item_id|prompt_id_list"
        return [500, PromptDeleteError1]
    if "prompt_id" in delete_id:
        prompt_id_list = [cell["f_prompt_id"] for cell in prompt_dao.get_all_data_from_prompt_list()]
        if delete_id["prompt_id"] not in prompt_id_list:
            PromptDeleteError1['description'] = "prompt_id参数不符合规范"
            PromptDeleteError1['detail'] = "该prompt不存在"
            return [500, PromptDeleteError1]
    if "type_id" in delete_id:
        prompt_id_list = [cell["f_prompt_item_type_id"] for cell in prompt_dao.get_all_from_prompt_item_list(None)]
        if delete_id["type_id"] not in prompt_id_list:
            PromptDeleteError1['description'] = "type_id参数不符合规范"
            PromptDeleteError1['detail'] = "该type不存在"
            return [500, PromptDeleteError1]
    if "item_id" in delete_id:
        prompt_id_list = [cell["f_prompt_item_id"] for cell in prompt_dao.get_all_from_prompt_item_list(None)]
        if delete_id["item_id"] not in prompt_id_list:
            PromptDeleteError1['description'] = "item_id参数不符合规范"
            PromptDeleteError1['detail'] = "该item不存在"
            return [500, PromptDeleteError1]
    if "prompt_id_list" in delete_id:
        if len(delete_id["prompt_id_list"]) == 0:
            PromptDeleteError1['description'] = "prompt_id_list参数不符合规范"
            PromptDeleteError1['detail'] = "列表不得为空"
            return [500, PromptDeleteError1]


# 提示词移动校验
async def prompt_move_verify(move_param):
    key_list = list(move_param.keys())
    if 'prompt_id' not in key_list:
        PromptMoveError['description'] = "prompt_id 参数不符合规范"
        PromptMoveError['detail'] = "参数中需提供 prompt_id"
        return [400, PromptMoveError]
    if 'prompt_item_id' not in key_list:
        PromptMoveError['description'] = "prompt_item_id 参数不符合规范"
        PromptMoveError['detail'] = "参数中需提供 prompt_item_id"
        return [400, PromptMoveError]
    if 'prompt_item_type_id' not in key_list:
        PromptMoveError['description'] = "prompt_item_type_id 参数不符合规范"
        PromptMoveError['detail'] = "参数中需提供 prompt_item_type_id"
        return [400, PromptMoveError]
    prompt_id_list = [cell["f_prompt_id"] for cell in prompt_dao.get_all_data_from_prompt_list()]
    if move_param["prompt_id"] not in prompt_id_list:
        PromptMoveError['description'] = "prompt_id 参数不符合规范"
        PromptMoveError['detail'] = "当前 prompt 不存在"
        return [400, PromptMoveError]
    prompt_id_list = [cell["f_prompt_item_type_id"] for cell in prompt_dao.get_all_from_prompt_item_list(None)]
    if move_param["prompt_item_type_id"] not in prompt_id_list:
        PromptMoveError['description'] = "prompt_item_type_id 参数不符合规范"
        PromptMoveError['detail'] = "当前分组不存在"
        return [400, PromptMoveError]
    prompt_id_list = [cell["f_prompt_item_id"] for cell in prompt_dao.get_all_from_prompt_item_list(None)]
    if move_param["prompt_item_id"] not in prompt_id_list:
        PromptMoveError['description'] = "prompt_item_id 参数不符合规范"
        PromptMoveError['detail'] = "当前项目不存在"
        return [400, PromptMoveError]


async def batch_add_prompt_endpoint_verify(params_list):
    if not isinstance(params_list, list):
        BatchAddPromptError['description'] = "参数必须为列表"
        BatchAddPromptError['detail'] = "参数必须为列表"
        return [400, BatchAddPromptError]

    for params in params_list:
        if "prompt_item_name" not in params:
            BatchAddPromptError['description'] = "缺少参数prompt_item_name"
            BatchAddPromptError['detail'] = "缺少参数prompt_item_name"
            return [400, BatchAddPromptError]

        if "prompt_item_type_name" not in params:
            BatchAddPromptError['description'] = "缺少参数prompt_item_type_name"
            BatchAddPromptError['detail'] = "缺少参数prompt_item_type_name"
            return [400, BatchAddPromptError]

        if "prompt_list" not in params:
            BatchAddPromptError['description'] = "缺少参数prompt_list"
            BatchAddPromptError['detail'] = "缺少参数prompt_list"
            return [400, BatchAddPromptError]

        if not isinstance(params["prompt_list"], list):
            BatchAddPromptError['description'] = "参数prompt_list必须为列表"
            BatchAddPromptError['detail'] = "参数prompt_list必须为列表"
            return [400, BatchAddPromptError]


def verify_icon_color_config(s):
    valid_colors = [
        "icon-color-pz-019688",
        "icon-color-pz-F759AB",
        "icon-color-pz-FADB14",
        "icon-color-pz-FF8501",
        "icon-color-pz-F75959",
        "icon-color-pz-8C8C8C",
        "icon-color-pz-126EE3",
        "icon-color-pz-13C2C2",
        "icon-color-pz-52C41A",
        "icon-color-pz-9254DE"
    ]
    if s in valid_colors:
        return True
    else:
        return False

def verify_icon_color_config_metric(s):
    valid_colors = [
        "icon-color-zbk-FF8501",
        "icon-color-zbk-13C2C2",
        "icon-color-zbk-FADB14",
        "icon-color-zbk-019688",
        "icon-color-zbk-9254DE",
        "icon-color-zbk-8C8C8C",
        "icon-color-zbk-126EE3",
        "icon-color-zbk-52C41A",
        "icon-color-zbk-F759AB",
        "icon-color-zbk-F75959"
    ]
    if s in valid_colors:
        return True
    else:
        return False

# 检查字段是否不含表情等字符，是否不超过指定长度
def verify_text_field(s, max_len):
    if not isinstance(s, str):
        return False
    if len(s) > max_len:
        return False
    if not re.match(r'^[=~!@#$&%^*()_+`{}\-\[\];:,.\\?<>\'"|/！·￥…（）—。【 】‘“’”：；、《》？，a-zA-Z0-9\u4e00-\u9fa5]*$', s):
        return False
    return True

def verify_id(s):
    if not isinstance(s, str):
        return False
    # 只能有数字
    if not re.match(r'^[0-9]{18}$', s) and not re.match(r'^[0-9]{19}$', s):
        return False
    return True

def include_dataset_id(dataset_version_id_list, dataset_id):
    try:
        for item in dataset_version_id_list:
            if item.split("/")[0] == dataset_id:
                return True
        return False
    except Exception:
        return False