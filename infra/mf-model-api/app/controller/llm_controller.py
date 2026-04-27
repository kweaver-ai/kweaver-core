from fastapi.responses import JSONResponse
from app.logs.stand_log import StandLogger
from app.mydb.ConnectUtil import redis_util, get_redis_util
from app.utils import llm_utils
from app.utils.llm_utils import openai_series_stream, OpenAIClientRequest
from app.utils.param_verify_utils import *
from app.utils.reshape_utils import *
from sse_starlette import EventSourceResponse
import time

eng_dict = {
    "名称已存在，请修改": "Name already exists, please modify",
    "名称已被其他用户占用，请修改": "The name is already taken by another user, please change it"
}


async def used_model_openai(request, user_id, language, func_module):
    if "stream" not in request.keys():
        stream = True
    else:
        stream = request["stream"]
        if not isinstance(stream, bool):
            error_dict = ModelFactory_Router_ParamError_TypeError_Error.copy()
            error_dict["detail"] = "stream " + error_dict["detail"]
            StandLogger.error(error_dict["detail"])
            return JSONResponse(status_code=400, content=error_dict)

    model_name = request["model"]
    model_id = request["model_id"]
    is_default = False
    if not model_name and not model_id:
        model_name = "default_model_3ed523"
        is_default = True
    if model_name:
        cache_key = f"dip:model-api:llm:{model_name}:list"
    else:
        cache_key = f"dip:model-api:llm:{model_id}:list"
    try:
        # 确保 redis_util 已初始化
        global redis_util
        if redis_util is None:
            redis_util = await get_redis_util()
        t1 = time.time()
        res = await redis_util.get_str(cache_key)
        t2 = time.time()
        StandLogger.info_log(f"从redis中获取llm配置耗时：{t2 - t1}s")
        if res is not None and isinstance(res, (str, bytes)):
            try:
                # 使用json.loads替代eval()，更安全且性能更好
                import json
                if isinstance(res, bytes):
                    res = res.decode('utf-8')
                model_info = json.loads(res)
            except (json.JSONDecodeError, ValueError) as e:
                StandLogger.warn(f"解析缓存数据失败: {str(e)}, key={cache_key}, value={res[:100]}...")
                # 缓存解析失败，回退到数据库查询
                if not is_default:
                    model_info = llm_model_dao.get_data_from_model_list_by_name_id(model_name, model_id)
                else:
                    model_info = llm_model_dao.get_data_from_default_model()
                if len(model_info) == 0:
                    if not is_default:
                        return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_Used_NameNotExist)
                    else:
                        return JSONResponse(status_code=400, content=ModelFactory_DedaultModel_NotExist)
                # 重新设置缓存，使用JSON格式
                await redis_util.set_str(key=cache_key, value=json.dumps(model_info), expire=3600 * 24)
        else:
            # 缓存不存在或非预期类型，从数据库获取
            if not is_default:
                model_info = llm_model_dao.get_data_from_model_list_by_name_id(model_name, model_id)
            else:
                model_info = llm_model_dao.get_data_from_default_model()
            if len(model_info) == 0:
                if not is_default:
                    return JSONResponse(status_code=400, content=ModelFactory_ExternalSmallModel_Used_NameNotExist)
                else:
                    return JSONResponse(status_code=400, content=ModelFactory_DedaultModel_NotExist)
            # 设置缓存，使用JSON格式
            import json
            await redis_util.set_str(key=cache_key, value=json.dumps(model_info), expire=3600 * 24)
    except Exception as e:
        StandLogger.error(e.args)
        DataBaseError["detail"] = str(e)
        return JSONResponse(status_code=500, content=DataBaseError)
    model_data = model_info[0]
    model_series = model_data["f_model_series"]
    context_size = model_data["f_max_model_len"]
    model_id = model_data["f_model_id"]
    quota = model_data["f_quota"]
    if quota:
        quota_cache_key = f"{user_id}:dip:model-api:llm-quota:{model_name}:list"
        res = await redis_util.get_str(quota_cache_key)
        if res is not None and isinstance(res, (str, bytes)):
            try:
                model_quota_info = eval(res)
            except Exception as e:
                StandLogger.warn(f"解析缓存数据失败: {str(e)}, key={quota_cache_key}, value={res}")
                # 缓存解析失败，回退到数据库查询
                model_quota_info = llm_model_dao.get_quota_by_user_and_model(user_id, model_id)
                # 重新设置缓存
                await redis_util.set_str(key=quota_cache_key, value=str(model_quota_info), expire=300)
        else:
            # 缓存不存在或非预期类型，从数据库获取
            model_quota_info = llm_model_dao.get_quota_by_user_and_model(user_id, model_id)
            # 设置缓存
            await redis_util.set_str(key=quota_cache_key, value=str(model_quota_info), expire=300)
        if len(model_quota_info) == 0 or model_quota_info[0]["remaining_input_tokens"] <= 0 or model_quota_info[0][
            "remaining_output_tokens"] <= 0:
            error_dict = ModelQuotaControllerUserModelConfigNoLeftSpaceError.copy()
            return JSONResponse(status_code=400, content=error_dict)

    if request["max_tokens"] > context_size * 1000:
        error_dict = ModelFactory_Router_ParamError_FormatError_Error.copy()
        error_dict["detail"] = f"max_tokens超过最大值{context_size}k"
        return JSONResponse(status_code=400, content=error_dict)
    messages = request["messages"]
    message = messages[len(messages) - 1]["content"]
    history_dia = []
    for i in range(0, len(messages)):
        if messages[i]["tool_calls"] is None:
            messages[i].pop("tool_calls")
        if messages[i]["tool_call_id"] is None:
            messages[i].pop("tool_call_id")
        tmp_item = {
            "role": messages[i]["role"],
            "message": messages[i]["content"]
        }
        if tmp_item["role"] == "assistant":
            tmp_item["role"] = "ai"
        elif tmp_item["role"] in ["user", "system"]:
            tmp_item["role"] = "human"
        history_dia.append(tmp_item)

    if model_series == "openai":
        _, mess_str = llm_utils.prompt(message + "", "", "", history_dia.copy())
        config = json.loads(model_data["f_model_config"].replace("'", '"'))
        try:
            openai_client = OpenAIClientRequest(
                api_url=config['api_url'],
                api_model=config['api_model'],
                api_key=config.get("api_key", ""),
                model_id=model_data["f_model_id"],
                temperature=request["temperature"],
                top_p=request["top_p"],
                frequency_penalty=request["frequency_penalty"],
                presence_penalty=request["presence_penalty"],
                max_tokens=request["max_tokens"],
                top_k=request["top_k"],
                response_format=request["response_format"],
                stop=request["stop"],
                tools=request.get("tools", None),
                tool_choice=request.get("tool_choice", None)
            )
            if stream:
                return EventSourceResponse(
                    openai_client.chat_completion_stream_openai(messages, user_id, True, func_module, request["cache"]),
                    ping=3600)
            else:
                res = await openai_client.chat_completion(messages, user_id, func_module, request["cache"])
                if "detail" in res.keys():
                    return JSONResponse(status_code=500, content=res)
                else:
                    return JSONResponse(status_code=200, content=res)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_ModelController_Model_ConnectError_Error)
    elif model_series.lower() == "claude":
        config = json.loads(model_data["f_model_config"].replace("'", '"'))
        from app.utils.llm_utils import ClaudeClient
        try:
            claude_client = ClaudeClient(
                api_url=config['api_url'],
                api_model=config['api_model'],
                api_key=config['api_key'],
                model_id=model_data["f_model_id"],
                temperature=request["temperature"],
                top_p=request["top_p"],
                frequency_penalty=request["frequency_penalty"],
                presence_penalty=request["presence_penalty"],
                max_tokens=request["max_tokens"],
                top_k=request["top_k"],
                system=request.get("system", []),
                tools=request.get("tools", None),
                tool_choice=request.get("tool_choice", None)
            )
            if stream:
                return EventSourceResponse(
                    claude_client.chat_completion_stream_openai(messages, user_id, func_module, request["cache"]),
                    ping=3600)
            else:
                res = await claude_client.chat_completion(messages, user_id, func_module, request["cache"])
                if "detail" in res.keys():
                    return JSONResponse(status_code=500, content=res)
                else:
                    return JSONResponse(status_code=200, content=res)
        except Exception as e:
            StandLogger.error(e.args)
            return JSONResponse(status_code=500, content=ModelFactory_ModelController_Model_ConnectError_Error)
    elif model_series.lower() == "baidu":
        config = json.loads(model_data["f_model_config"].replace("'", '"'))
        from app.utils.llm_utils import BaiduClient
        baidu_client = BaiduClient(
            api_url=config['api_url'],
            api_model=config['api_model'],
            api_key=config['api_key'],
            model_id=model_data["f_model_id"],
            temperature=request["temperature"],
            top_p=request["top_p"],
            frequency_penalty=request["frequency_penalty"],
            presence_penalty=request["presence_penalty"],
            max_tokens=request["max_tokens"],
            top_k=request["top_k"],
            stop=request["stop"],
            secret_key=config["secret_key"]
        )
        if stream:
            return EventSourceResponse(
                baidu_client.chat_completion_stream_openai(messages, user_id, True, func_module, request["cache"]),
                ping=3600)
        else:
            res = await baidu_client.chat_completion(messages, user_id, func_module, request["cache"])
            if "detail" in res.keys():
                return JSONResponse(status_code=500, content=res)
            else:
                return JSONResponse(status_code=200, content=res)
    elif model_series.lower() == "baidu_tianchen":
        config = json.loads(model_data["f_model_config"].replace("'", '"'))
        from app.utils.llm_utils import BaiduTianchenClient
        try:
            baidu_tianchen_client = BaiduTianchenClient(
                api_url=config['api_url'],
                api_model=config['api_model'],
                model_id=model_data["f_model_id"],
                temperature=request["temperature"],
                top_p=request["top_p"],
                frequency_penalty=request["frequency_penalty"],
                presence_penalty=request["presence_penalty"],
                max_tokens=request["max_tokens"],
                top_k=request["top_k"],
                stop=request["stop"],
                OperationCode=config['OperationCode'],
                ClientId=config['ClientId']
            )
            if stream:
                return EventSourceResponse(
                    baidu_tianchen_client.chat_completion_stream_openai(messages, user_id, True, request["cache"]),
                    ping=3600)
            else:
                res = await baidu_tianchen_client.chat_completion(messages, user_id, request["cache"])
                if "detail" in res.keys():
                    return JSONResponse(status_code=500, content=res)
                else:
                    return JSONResponse(status_code=200, content=res)
        except Exception as e:
            return JSONResponse(status_code=500, content=ModelFactory_ModelController_Model_ConnectError_Error)
    else:
        config = json.loads(model_data["f_model_config"].replace("'", '"'))
        from app.utils.llm_utils import OtherClient
        try:
            other_client = OtherClient(
                api_url=config['api_url'],
                api_model=config['api_model'],
                api_key=config.get("api_key", ""),
                model_id=model_data["f_model_id"],
                temperature=request["temperature"],
                top_p=request["top_p"],
                frequency_penalty=request["frequency_penalty"],
                presence_penalty=request["presence_penalty"],
                max_tokens=request["max_tokens"],
                top_k=request["top_k"],
                response_format=request["response_format"],
                stop=request["stop"],
                model_type=model_data["f_model_type"],
                tools=request.get("tools", None),
                tool_choice=request.get("tool_choice", None)
            )
            if stream:
                return EventSourceResponse(
                    other_client.chat_completion_stream_openai(messages, user_id, True, model_data, func_module),
                    ping=3600)
            else:
                res = await other_client.chat_completion(messages, user_id, func_module)
                if "detail" in res.keys():
                    return JSONResponse(status_code=500, content=res)
                else:
                    return JSONResponse(status_code=200, content=res)
        except Exception as e:
            StandLogger.error(f"call llmModelError {config['api_model']} error params={messages},error={e}")
            return JSONResponse(status_code=500, content=ModelFactory_ModelController_Model_ConnectError_Error)
