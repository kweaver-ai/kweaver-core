import asyncio
import json
import re
import time
import threading

import aiohttp
import tiktoken
import requests
from fastapi.responses import JSONResponse
from func_timeout import func_set_timeout
from llmadapter.llms.llm_factory import llm_factory
from llmadapter.schema import HumanMessage, AIMessage

from app.commons.errors import ModelError, LLMParamError, ModelTimeoutError, \
    ModelFactory_ModelController_Model_Error_Error
from app.controller.model_audit_controller import add_llm_model_call_log
from app.core.config import base_config
from app.dao.llm_model_dao import llm_model_dao
from app.interfaces import logics
from app.logs.stand_log import StandLogger
from app.utils.observability.observability_log import get_logger

from app.utils.str_util import generate_random_string, has_common_substring

# 全局tokenizer缓存
_TOKENIZER_CACHE = {}
_TOKENIZER_LOCK = threading.RLock()
_MAX_CACHE_SIZE = 5  # 限制最大缓存数量


class OpenAIClient:
    def __init__(self, api_key, api_model, temperature, top_p, top_k, frequency_penalty,
                 presence_penalty, max_tokens, base_url, stop=None, tools=None, tool_choice=None):
        self.llm_type = "openai",
        self.api_type = "azure",
        self.api_version = "2023-03-15-preview",
        self.openai_api_base = base_url,
        self.openai_api_key = api_key,
        self.engine = api_model,
        self.temperature = temperature
        self.top_p = top_p
        self.frequency_penalty = frequency_penalty
        self.presence_penalty = presence_penalty
        self.max_tokens = max_tokens
        self.top_k = top_k
        self.stop = stop
        self.tools = tools
        self.tool_choice = tool_choice

    async def openai_chat_completion(self, message):
        start = time.time()
        llm = llm_factory.create_llm(
            llm_type=self.llm_type[0],
            api_type=self.api_type[0],
            api_version=self.api_version[0],
            openai_api_base=self.openai_api_base[0],
            openai_api_key=self.openai_api_key[0],
            engine=self.engine[0],
            temperature=self.temperature,
            top_p=self.top_p,
            frequency_penalty=self.frequency_penalty,
            presence_penalty=self.presence_penalty,
            max_tokens=self.max_tokens,
            tools=self.tools,
            tool_choice=self.tool_choice
        )
        res = llm.predict(message)
        import os

        cache_key = "9b5ad71b2ce5302211f9c61530b329a4922fc6a4"
        tiktoken_cache_dir = os.path.dirname(os.path.dirname(os.path.realpath(__file__))) + "/utils/tiktoken_cache"
        os.environ["TIKTOKEN_CACHE_DIR"] = tiktoken_cache_dir

        assert os.path.exists(os.path.join(tiktoken_cache_dir, cache_key))

        encoding = tiktoken.get_encoding("cl100k_base")
        tokens = encoding.encode(res)
        token_len = len(tokens)
        result = {
            "res":
                {
                    "time": time.time() - start,
                    "token_len": token_len,
                    "data": res
                }
        }
        return result

    async def openai_chat_completion_stream(self, message, llm_id, user_id, prompt_tokens, cache=False):
        llm = llm_factory.create_llm(
            llm_type=self.llm_type[0],
            api_type=self.api_type[0],
            api_version=self.api_version[0],
            openai_api_base=self.openai_api_base[0],
            openai_api_key=self.openai_api_key[0],
            engine=self.engine[0],
            temperature=self.temperature,
            top_p=self.top_p,
            frequency_penalty=self.frequency_penalty,
            presence_penalty=self.presence_penalty,
            max_tokens=self.max_tokens,
            stop=self.stop,
            tools=self.tools,
            tool_choice=self.tool_choice)
        res_mess = ""
        start_time = time.time()
        token_num = 0
        for token in llm.stream_generator(message, max_tokens=self.max_tokens):
            yield token
            token_num += 1
            res_mess += token
        # yield res_mess
        end_time = time.time()

        import os
        cache_key = "9b5ad71b2ce5302211f9c61530b329a4922fc6a4"
        tiktoken_cache_dir = os.path.dirname(os.path.dirname(os.path.realpath(__file__))) + "/utils/tiktoken_cache"
        os.environ["TIKTOKEN_CACHE_DIR"] = tiktoken_cache_dir

        assert os.path.exists(os.path.join(tiktoken_cache_dir, cache_key))

        encoding = tiktoken.get_encoding("cl100k_base")
        tokens = encoding.encode(res_mess)
        token_len = len(tokens)

        yield "--info--" + str(json.dumps({"time": str(end_time - start_time), "token_len": token_len}))
        log_info = logics.AddModelUsedAudit(
            model_id=llm_id, user_id=user_id, input_tokens=prompt_tokens, output_tokens=token_num)

        await add_llm_model_call_log(log_info)
        # print(res_mess)


class OpenAIClientRequest:
    def __init__(self, api_url, api_model, api_key, model_id,
                 temperature, top_p, frequency_penalty, presence_penalty, max_tokens, top_k=1, response_format={},
                 stop=None, tools=None, tool_choice=None):
        self.api_url = api_url
        self.api_model = api_model
        self.temperature = temperature
        self.top_p = top_p
        self.frequency_penalty = frequency_penalty
        self.presence_penalty = presence_penalty
        self.max_tokens = max_tokens
        self.api_key = api_key
        self.model_id = model_id
        self.top_k = top_k
        self.response_format = response_format
        self.stop = stop
        self.tools = tools
        self.tool_choice = tool_choice

    async def chat_completion(self, messages, user_id, func_module, cache=False):  # 写一版直接请求url的，便于传入工具
        if messages[len(messages) - 1]["role"] != "user" and self.api_model.find("qianxun") != -1:
            error_dict = ModelError.copy()
            error_dict["description"] = error_dict["detail"] = error_dict["solution"] = "千循大模型只支持最后一条消息role为user"
            return error_dict
        start_time = time.time()
        params = {
            "messages": messages,
            "model": self.api_model,
            "stream": False,
            "top_p": self.top_p,
            "temperature": self.temperature,
            "frequency_penalty": self.frequency_penalty,
            "presence_penalty": self.presence_penalty,
            "max_tokens": self.max_tokens
        }
        if self.response_format != {}:
            params["response_format"] = self.response_format
        if self.stop is not None:
            params["stop"] = self.stop
        if self.tools is not None:
            params["tools"] = self.tools
        if self.tool_choice is not None:
            params["tool_choice"] = self.tool_choice
        headers = {
            "api-key": self.api_key
        }
        conn = aiohttp.TCPConnector(verify_ssl=False)
        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(
                    self.api_url + "openai/deployments/{}/chat/completions?api-version=2023-05-15&api-type=azure".format(
                        self.api_model),
                    json=params, headers=headers) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status == 200:
                    log_info = logics.AddModelUsedAudit(
                        model_id=self.model_id, user_id=user_id, input_tokens=result["usage"]["prompt_tokens"],
                        output_tokens=result["usage"]["completion_tokens"])
                    await add_llm_model_call_log(log_info)
                    usage = result.get("usage", {})
                    prompt_tokens = usage.get("prompt_tokens", 0)
                    completion_tokens = usage.get("completion_tokens", 0)
                    total_tokens = usage.get("total_tokens", 0)
                    if get_logger():
                        get_logger().info(
                            f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                            f'"prompt_tokens":{prompt_tokens},"completion_tokens":{completion_tokens},'
                            f'"total_tokens":{total_tokens},"func_module":{func_module},"status":"success"}}')
                    return result
                else:
                    tmp_map = result
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    if "detail" in tmp_map.keys():
                        error_dict["detail"] = tmp_map["detail"]
                    if "error" in tmp_map.keys() and "message" in tmp_map["error"].keys():
                        error_dict["detail"] = tmp_map["error"]["message"]
                    if get_logger():
                        get_logger().info(
                            f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                            f'"prompt_tokens":0,"completion_tokens":0,'
                            f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                    return error_dict

    async def chat_completion_stream_openai(self, messages, user_id, return_info, func_module, cache=False):
        if messages[len(messages) - 1]["role"] != "user" and self.api_model.find("qianxun") != -1:
            error_dict = ModelError.copy()
            error_dict["description"] = error_dict["detail"] = error_dict["solution"] = "千循大模型只支持最后一条消息role为user"
            yield error_dict
        retry_time = 3
        while retry_time > 0:
            retry_time -= 1
            try:
                token_yield = False
                start_time = time.time()
                token_len = 0
                params = {
                    "messages": messages,
                    "model": self.api_model,
                    "stream": True,
                    "top_p": self.top_p,
                    "temperature": self.temperature,
                    "frequency_penalty": self.frequency_penalty,
                    "presence_penalty": self.presence_penalty,
                    "max_tokens": self.max_tokens,
                    "stream_options": {
                        "include_usage": True
                    }
                }
                if self.response_format != {}:
                    params["response_format"] = self.response_format
                if self.stop is not None:
                    params["stop"] = self.stop
                if self.tools is not None:
                    params["tools"] = self.tools
                if self.tool_choice is not None:
                    params["tool_choice"] = self.tool_choice
                for i in range(0, len(messages)):
                    if messages[i]["content"] == "":
                        messages[i]["content"] = " "
                headers = {
                    "api-key": self.api_key
                }
                prompt_tokens = 0
                completion_tokens = 0
                async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                    url = self.api_url + f"openai/deployments/{self.api_model}/chat/completions?api-version=2023-05-15&api-type=azure"
                    async with session.post(url, json=params, headers=headers, ssl=False) as response:
                        response.encoding = 'utf-8'
                        if response.status != 200:
                            error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                            error_dict["description"] = error_dict["detail"] = await response.text()
                            yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                            return
                        ans = ""
                        async for chunk in response.content:
                            chunk = chunk.decode(encoding='utf-8')
                            if chunk.endswith('\n'):
                                chunk = chunk[:-1]
                            elif chunk.endswith('\r\n'):
                                chunk = chunk[:-2]
                            if chunk != "data: [DONE]" and chunk != "":
                                if chunk[0:6] == "data: ":
                                    chunk = chunk[6:]
                                datas = json.loads(chunk)
                                try:
                                    delta = datas["choices"][0]["delta"]
                                    if "content" in delta:
                                        token = delta["content"]
                                        if token is not None:
                                            # yield token
                                            token_len += 1
                                            # print(token)
                                            ans += token
                                except Exception as e:
                                    pass
                                if "usage" in datas.keys() and completion_tokens == 0:
                                    try:
                                        prompt_tokens = datas["usage"]["prompt_tokens"]
                                        completion_tokens = datas["usage"]["completion_tokens"]
                                        token_yield = True
                                    except Exception:
                                        pass
                                yield chunk
                            elif chunk == "data: [DONE]":
                                # end_time = time.time()
                                if return_info:
                                    if completion_tokens == 0:
                                        completion_tokens = token_len
                                        model_info = llm_model_dao.get_data_from_model_list_by_id(self.model_id)
                                        prompt_str = ""
                                        for item_prompt in messages:
                                            prompt_str += item_prompt["content"]
                                        usage_res = {
                                            "id": "chatcmpl-" + generate_random_string(32),
                                            "object": "chat.completion.chunk",
                                            "created": int(time.time()),
                                            "model": self.api_model,
                                            "choices": [
                                                {
                                                    "index": 0,
                                                    "delta": {
                                                        "content": ""
                                                    },
                                                    "finish_reason": "stop",
                                                    "usage": {
                                                        "prompt_tokens": prompt_tokens,
                                                        "completion_tokens": completion_tokens,
                                                        "total_tokens": prompt_tokens + completion_tokens
                                                    }
                                                }
                                            ]

                                        }
                                        if token_yield == False:
                                            yield json.dumps(usage_res, ensure_ascii=False)
                                    # yield "--info--" + str(json.dumps({"time": str(end_time - start_time),
                                    #                                    "token_len": completion_tokens,
                                    #                                    "prompt_tokens": prompt_tokens}))
                                    yield "[DONE]"
                                # yield "--end--"
                        log_info = logics.AddModelUsedAudit(
                            model_id=self.model_id, user_id=user_id, input_tokens=prompt_tokens,
                            output_tokens=completion_tokens)
                        await add_llm_model_call_log(log_info)
                        if get_logger():
                            get_logger().info(
                                f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                                f'"prompt_tokens":{prompt_tokens},"completion_tokens":{completion_tokens},'
                                f'"total_tokens":{prompt_tokens + completion_tokens},"func_module":{func_module},"status":"success"}}')
            except aiohttp.ClientError as e:
                if retry_time <= 0:
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    error_dict["description"] = f"大模型: {self.api_model} 连接失败，请检查该服务是否可用"
                    error_dict["detail"] = str(e)
                    yield json.dumps(error_dict, ensure_ascii=False)
                    StandLogger.error(json.dumps(error_dict, ensure_ascii=False))
                    if get_logger():
                        get_logger().info(
                            f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                            f'"prompt_tokens":0,"completion_tokens":0,'
                            f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                    return
                else:
                    StandLogger.warn(f"大模型: {self.api_model} 连接失败，1秒后重试")
                    await asyncio.sleep(1)
            except Exception as e:
                StandLogger.error(e.args)
                if get_logger():
                    get_logger().info(
                        f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                        f'"prompt_tokens":0,"completion_tokens":0,'
                        f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                raise e


def prompt(ai_system, ai_user, ai_assistant, ai_history):
    messages = []
    mess_str = ''
    if ai_history:
        if type(ai_history) is str:
            ai_history = json.loads(ai_history.replace("'", '"'))
        for cell in ai_history:
            if cell["role"] == "human" and cell["message"].strip():
                messages.append(HumanMessage(content="{}".format(cell["message"])))
                mess_str += cell["message"]
            elif cell["role"] == 'ai' and cell["message"].strip():
                messages.append(AIMessage(content="{}".format(cell["message"])))
                mess_str += cell["message"]
    if ai_system.strip():
        messages.append(HumanMessage(content="{}".format(ai_system)))
        mess_str += ai_system
    if ai_user.strip():
        messages.append(HumanMessage(content="{}".format(ai_user)))
        mess_str += ai_user
    if ai_assistant.strip():
        messages.append(HumanMessage(content="{}".format(ai_assistant)))
        mess_str += ai_assistant
    return messages, mess_str


async def openai_token_num(user_max_token, llm_max_token, message):
    import os

    cache_key = "9b5ad71b2ce5302211f9c61530b329a4922fc6a4"
    tiktoken_cache_dir = os.path.dirname(os.path.dirname(os.path.realpath(__file__))) + "/utils/tiktoken_cache"
    os.environ["TIKTOKEN_CACHE_DIR"] = tiktoken_cache_dir

    assert os.path.exists(os.path.join(tiktoken_cache_dir, cache_key))

    encoding = tiktoken.get_encoding("cl100k_base")
    tokens = encoding.encode(message)
    mess_token = len(tokens)
    if user_max_token + mess_token >= llm_max_token - 50:
        return -1
    return mess_token


async def openai_series_stream(types, api_key, api_model, ai_system, ai_history, ai_user, ai_assistant, base_url,
                               top_p=0.95,
                               top_k=1, temperature=0.7, max_tokens=300, frequency_penalty=0, presence_penalty=0,
                               llm_id=None,
                               user_id="", cache=False, stop=None):
    try:
        # 获取所有输入信息的长度
        messages, mess_str = prompt(ai_system, ai_user, ai_assistant, ai_history)
        max_token = await get_context_size("openai", base_url, api_model)
        llm_max_token = max_token if max_token is not None else 16000
        # 校验 token 有没有超出
        try:
            prompt_tokens = await openai_token_num(max_tokens, llm_max_token, mess_str)
            if prompt_tokens == -1:
                return JSONResponse(status_code=400, content=ModelError)
        except Exception as e:
            StandLogger.error(str(e))
            return JSONResponse(status_code=500, content=LLMParamError)

        llm = OpenAIClient(api_key, api_model, temperature, top_p, top_k, frequency_penalty,
                           presence_penalty, max_tokens, base_url, stop)
        if types == "chat":
            res = llm.openai_chat_completion_stream(messages, llm_id, user_id, prompt_tokens, cache)

            return res
        else:
            res = llm.openai_chat_completion_stream(ai_system, llm_id, user_id, prompt_tokens, cache)

            return res
    except Exception as e:
        return JSONResponse(status_code=500, content=ModelTimeoutError)


class BaiduTianchenClient:
    def __init__(self, api_url, api_model, model_id, temperature, top_p, max_tokens, frequency_penalty,
                 ClientId, OperationCode, presence_penalty, top_k=1, stop=None):
        self.api_url = api_url
        self.api_model = api_model
        self.temperature = temperature if temperature != 0 else 0.1
        self.frequency_penalty = frequency_penalty
        self.presence_penalty = presence_penalty
        self.max_tokens = max_tokens
        self.model_id = model_id
        self.top_k = top_k
        self.stop = stop
        self.top_p = top_p
        self.ClientId = ClientId
        self.OperationCode = OperationCode

    async def chat_completion(self, messages, user_id, cache=False):
        StandLogger.info_log("messages: " + json.dumps(messages, ensure_ascii=False))
        messages_json = json.dumps(messages, ensure_ascii=False)
        system = None
        prompt_str = ""
        new_messages = []
        for i in range(0, len(messages)):
            try:
                prompt_str += messages[i]["content"]
                if messages[i]["role"] == "system":
                    if len(messages) > 1:
                        if system is None:
                            system = messages[i]["content"]
                    else:
                        messages[i]["role"] = "user"
                        new_messages = messages
                else:
                    new_messages.append(messages[i])
            except Exception:
                pass
        params = {
            "messages": new_messages,
            "model": self.api_model,
            "stream": False,
            "top_p": self.top_p,
            "temperature": self.temperature,
            "max_tokens": self.max_tokens,
            "top_k": self.top_k
        }
        if system is not None:
            params["system"] = system
        conn = aiohttp.TCPConnector(verify_ssl=False)
        headers = {
            "ClientId": self.ClientId,
            "OperationCode": self.OperationCode
        }
        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(self.api_url + f"?api_name={self.api_model}", json=params, headers=headers) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status == 200 and "result" in result.keys():
                    log_info = logics.AddModelUsedAudit(
                        model_id=self.model_id, user_id=user_id, input_tokens=int(len(prompt_str) // 1.5),
                        output_tokens=int(len(result["result"]) // 1.5))
                    await add_llm_model_call_log(log_info)
                    res = {
                        "id": "chatcmpl-" + generate_random_string(32),
                        "object": "chat.completion",
                        "created": int(time.time()),
                        "model": self.api_model,
                        "choices": [
                            {
                                "index": 0,
                                "message": {
                                    "role": "assistant",
                                    "content": result["result"]
                                },
                                "finish_reason": "stop"
                            }
                        ],
                        "usage": {
                            "prompt_tokens": int(len(prompt_str) // 1.5),
                            "total_tokens": int(len(prompt_str) // 1.5 + len(result["result"]) // 1.5),
                            "completion_tokens": int(len(result["result"]) // 1.5)
                        }
                    }
                    return res
                else:
                    tmp_map = result
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    if "detail" in tmp_map.keys():
                        error_dict["detail"] = tmp_map["detail"]
                    if "error" in tmp_map.keys() and "message" in tmp_map["error"].keys():
                        error_dict["detail"] = tmp_map["error"]["message"]
                    return error_dict

    async def chat_completion_stream_openai(self, messages, user_id, return_info, cache=False):
        StandLogger.info_log("messages: " + json.dumps(messages, ensure_ascii=False))
        token_yield = False
        messages_json = json.dumps(messages, ensure_ascii=False)
        retry_time = 3
        try:
            while retry_time > 0:
                prompt_str = ""
                retry_time -= 1
                token_len = 0
                system = None
                new_messages = []
                for i in range(0, len(messages)):
                    try:
                        prompt_str += messages[i]["content"]
                        if messages[i]["role"] == "system":
                            if len(messages) > 1:
                                if system is None:
                                    system = messages[i]["content"]
                            else:
                                messages[i]["role"] = "user"
                                new_messages = messages
                        else:
                            new_messages.append(messages[i])
                    except Exception:
                        pass
                params = {
                    "messages": new_messages,
                    "model": self.api_model,
                    "stream": True,
                    "top_p": self.top_p,
                    "temperature": self.temperature,
                    "max_tokens": self.max_tokens,
                    "top_k": self.top_k
                }
                headers = {
                    "ClientId": self.ClientId,
                    "OperationCode": self.OperationCode
                }
                if system is not None:
                    params["system"] = system
                async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                    url = self.api_url + f"?api_name={self.api_model}"
                    async with session.post(url, json=params, headers=headers, ssl=False) as response:
                        response.encoding = 'utf-8'
                        if response.status != 200:
                            error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                            error_dict["description"] = error_dict["detail"] = await response.text()
                            yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                            return
                        ans = ""
                        prompt_tokens = completion_tokens = 0
                        async for chunk in response.content:
                            chunk = chunk.decode(encoding='utf-8')
                            if chunk.endswith('\n'):
                                chunk = chunk[:-1]
                            elif chunk.endswith('\r\n'):
                                chunk = chunk[:-2]
                            if chunk != "data: [DONE]" and chunk != "":
                                if chunk[0:6] == "data: ":
                                    chunk = chunk[6:]
                                datas = json.loads(chunk)
                                if "result" not in datas.keys():
                                    yield chunk
                                    return
                                content = datas["result"]
                                if datas.get("sentence_id", 0) == 0:
                                    format_json = {
                                        "id": datas["id"],
                                        "model": self.api_model,
                                        "choices": [
                                            {
                                                "index": 0,
                                                "delta": {
                                                    "role": "assistant"
                                                },
                                                "finish_reason": None
                                            }
                                        ],
                                        "usage": {
                                            "prompt_tokens": 0,
                                            "completion_tokens": 0,
                                            "total_tokens": 0
                                        }
                                    }
                                    yield json.dumps(format_json, ensure_ascii=False)
                                format_json = {
                                    "id": datas["id"],
                                    "model": self.api_model,
                                    "choices": [
                                        {
                                            "index": 0,
                                            "delta": {
                                                "content": content
                                            },
                                            "finish_reason": None
                                        }
                                    ],
                                    "usage": {
                                        "prompt_tokens": 0,
                                        "completion_tokens": 0,
                                        "total_tokens": 0
                                    }
                                }
                                yield json.dumps(format_json, ensure_ascii=False)
                                ans += content
                                if datas["is_end"] is True:
                                    prompt_tokens = int(len(prompt_str) // 1.5)
                                    completion_tokens = int(len(ans) // 1.5)
                                    format_json = {
                                        "id": datas["id"],
                                        "model": self.api_model,
                                        "choices": [
                                            {
                                                "index": 0,
                                                "delta": {},
                                                "finish_reason": "stop"
                                            }
                                        ],
                                        "usage": {
                                            "prompt_tokens": prompt_tokens,
                                            "completion_tokens": completion_tokens,
                                            "total_tokens": prompt_tokens + completion_tokens
                                        }
                                    }

                                    yield json.dumps(format_json, ensure_ascii=False)
                                    yield "[DONE]"
                                    break
                        log_info = logics.AddModelUsedAudit(
                            model_id=self.model_id, user_id=user_id, input_tokens=prompt_tokens,
                            output_tokens=completion_tokens)
                        await add_llm_model_call_log(log_info)
                        return
        except aiohttp.ClientError as e:
            if retry_time <= 0:
                error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                error_dict["description"] = f"大模型: {self.api_model} 连接失败，请检查该服务是否可用"
                error_dict["detail"] = str(e)
                yield json.dumps(error_dict, ensure_ascii=False)
                StandLogger.error(json.dumps(error_dict, ensure_ascii=False))
                return
            else:
                StandLogger.warn(f"大模型: {self.api_model} 连接失败，1秒后重试")
                await asyncio.sleep(1)
        except Exception as e:
            StandLogger.error(e.args)
            raise e

    async def chat_completion_stream(self, messages, user_id, return_info, cache=False):
        StandLogger.info_log("messages: " + json.dumps(messages, ensure_ascii=False))
        retry_time = 3
        while retry_time > 0:
            prompt_str = ""
            retry_time -= 1
            try:
                start_time = time.time()
                headers = {
                    "ClientId": self.ClientId,
                    "OperationCode": self.OperationCode
                }
                system = None
                new_messages = []
                for i in range(0, len(messages)):
                    try:
                        prompt_str += messages[i]["content"]
                        if messages[i]["role"] == "system":
                            if len(messages) > 1:
                                if system is None:
                                    system = messages[i]["content"]
                            else:
                                messages[i]["role"] = "user"
                                new_messages = messages
                        else:
                            new_messages.append(messages[i])
                    except Exception:
                        pass
                params = {
                    "messages": new_messages,
                    "model": self.api_model,
                    "stream": True,
                    "top_p": self.top_p,
                    "temperature": self.temperature,
                    "max_tokens": self.max_tokens,
                    "top_k": self.top_k
                }
                if system is not None:
                    params["system"] = system
                url = self.api_url + f"?api_name={self.api_model}"
                async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                    async with session.post(url, json=params, headers=headers, ssl=False) as response:
                        response.encoding = 'utf-8'
                        if response.status != 200:
                            error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                            error_dict["description"] = error_dict["detail"] = await response.text()
                            yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                            return
                        ans = ""
                        prompt_tokens = completion_tokens = 0
                        async for chunk in response.content:
                            chunk = chunk.decode(encoding='utf-8')
                            if chunk.endswith('\n'):
                                chunk = chunk[:-1]
                            elif chunk.endswith('\r\n'):
                                chunk = chunk[:-2]
                            if chunk != "data: [DONE]" and chunk != "":
                                if chunk[0:6] == "data: ":
                                    chunk = chunk[6:]
                                datas = json.loads(chunk)
                                if "result" not in datas.keys():
                                    yield "--error--" + chunk
                                    return
                                content = datas["result"]
                                yield content
                                if datas["is_end"] is True:
                                    prompt_tokens = int(len(prompt_str) // 1.5)
                                    completion_tokens = int(len(ans) // 1.5)
                                    break
                        end_time = time.time()
                        if return_info:
                            yield "--info--" + str(json.dumps({"time": str(end_time - start_time),
                                                               "token_len": completion_tokens,
                                                               "prompt_tokens": prompt_tokens}))
                        yield "--end--"
                        log_info = logics.AddModelUsedAudit(
                            model_id=self.model_id, user_id=user_id, input_tokens=prompt_tokens,
                            output_tokens=completion_tokens)
                        await add_llm_model_call_log(log_info)
                        return
            except aiohttp.ClientError as e:
                if retry_time <= 0:
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    error_dict["description"] = f"大模型: {self.api_model} 连接失败，请检查该服务是否可用"
                    error_dict["detail"] = str(e)
                    yield json.dumps(error_dict, ensure_ascii=False)
                    StandLogger.error(json.dumps(error_dict, ensure_ascii=False))
                    return
                else:
                    StandLogger.warn(f"大模型: {self.api_model} 连接失败，1秒后重试")
                    await asyncio.sleep(1)
            except Exception as e:
                StandLogger.error(e.args)
                raise e


# 百度大模型类
class BaiduClient:
    def __init__(self, api_url, api_model, api_key, model_id, temperature, top_p, max_tokens, frequency_penalty,
                 presence_penalty, secret_key, top_k=1, stop=None):
        self.api_url = api_url
        self.api_model = api_model
        self.temperature = temperature
        self.frequency_penalty = frequency_penalty
        self.presence_penalty = presence_penalty
        self.max_tokens = max_tokens
        self.api_key = api_key
        self.model_id = model_id
        self.top_k = top_k
        self.stop = stop
        self.secret_key = secret_key
        self.top_p = top_p

    async def chat_completion(self, messages, user_id, func_module, cache=False):
        messages_json = json.dumps(messages, ensure_ascii=False)
        headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        }
        access_token = ""
        conn = aiohttp.TCPConnector(verify_ssl=False)
        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(
                    f"https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id={self.api_key}&client_secret={self.secret_key}",
                    headers=headers) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status != 200:
                    tmp_map = result
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    if "detail" in tmp_map.keys():
                        error_dict["detail"] = tmp_map["detail"]
                    if "error" in tmp_map.keys() and "message" in tmp_map["error"].keys():
                        error_dict["detail"] = tmp_map["error"]["message"]
                    return error_dict
                access_token = result["access_token"]
        start_time = time.time()
        system = None
        new_messages = []
        for i in range(0, len(messages)):
            if messages[i]["role"] == "system":
                if len(messages) > 1:
                    if system is None:
                        system = messages[i]["content"]
                else:
                    messages[i]["role"] = "user"
                    new_messages = messages
            else:
                new_messages.append(messages[i])
        params = {
            "messages": new_messages,
            "model": self.api_model,
            "stream": False,
            "top_p": self.top_p,
            "temperature": self.temperature,
            "frequency_penalty": self.frequency_penalty,
            "presence_penalty": self.presence_penalty,
            "max_output_tokens": self.max_tokens,
            "top_k": self.top_k
        }
        if system is not None:
            params["system"] = system
        conn = aiohttp.TCPConnector(verify_ssl=False)
        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(self.api_url + f"?access_token={access_token}", json=params) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status == 200 and "result" in result.keys():
                    log_info = logics.AddModelUsedAudit(
                        model_id=self.model_id, user_id=user_id, input_tokens=result["usage"]["prompt_tokens"],
                        output_tokens=result["usage"]["completion_tokens"])
                    await add_llm_model_call_log(log_info)
                    res = {
                        "id": "chatcmpl-" + generate_random_string(32),
                        "object": "chat.completion",
                        "created": result["created"],
                        "model": self.api_model,
                        "choices": [
                            {
                                "index": 0,
                                "message": {
                                    "role": "assistant",
                                    "content": result["result"]
                                },
                                "finish_reason": "stop"
                            }
                        ],
                        "usage": result["usage"]
                    }
                    usage = result.get("usage", {})
                    prompt_tokens = usage.get("prompt_tokens", 0)
                    completion_tokens = usage.get("completion_tokens", 0)
                    total_tokens = usage.get("total_tokens", 0)
                    if get_logger():
                        get_logger().info(
                            f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                            f'"prompt_tokens":{prompt_tokens},"completion_tokens":{completion_tokens},'
                            f'"total_tokens":{total_tokens},"func_module":{func_module},"status":"success"}}')
                    return res
                else:
                    tmp_map = result
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    if "detail" in tmp_map.keys():
                        error_dict["detail"] = tmp_map["detail"]
                    if "error" in tmp_map.keys() and "message" in tmp_map["error"].keys():
                        error_dict["detail"] = tmp_map["error"]["message"]
                    if get_logger():
                        get_logger().info(
                            f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                            f'"prompt_tokens":0,"completion_tokens":0,'
                            f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                    return error_dict

    async def chat_completion_stream_openai(self, messages, user_id, return_info, func_module, cache=False):
        token_yield = False
        messages_json = json.dumps(messages, ensure_ascii=False)
        retry_time = 3
        while retry_time > 0:
            retry_time -= 1
            try:
                headers = {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                }
                async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                    url = f"https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id={self.api_key}&client_secret={self.secret_key}"
                    async with session.post(url, headers=headers, ssl=False) as access_res:
                        if access_res.status != 200:
                            tmp_map = await access_res.json()
                            error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                            if "detail" in tmp_map.keys():
                                error_dict["detail"] = tmp_map["detail"]
                            if "error" in tmp_map.keys() and "message" in tmp_map["error"].keys():
                                error_dict["detail"] = tmp_map["error"]["message"]
                            yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                            return
                        access_info = await access_res.json()
                        access_token = access_info["access_token"]
                start_time = time.time()
                token_len = 0
                system = None
                new_messages = []
                for i in range(0, len(messages)):
                    if messages[i]["role"] == "system":
                        if len(messages) > 1:
                            if system is None:
                                system = messages[i]["content"]
                        else:
                            messages[i]["role"] = "user"
                            new_messages = messages
                    else:
                        new_messages.append(messages[i])
                params = {
                    "messages": new_messages,
                    "model": self.api_model,
                    "stream": True,
                    "top_p": self.top_p,
                    "temperature": self.temperature,
                    "frequency_penalty": self.frequency_penalty,
                    "presence_penalty": self.presence_penalty,
                    "max_output_tokens": self.max_tokens,
                    "top_k": self.top_k,
                    "stream_options": {
                        "include_usage": True
                    }
                }
                if system is not None:
                    params["system"] = system
                async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                    url = self.api_url + f"?access_token={access_token}"
                    async with session.post(url, json=params, headers=headers, ssl=False) as response:
                        response.encoding = 'utf-8'
                        if response.status != 200:
                            error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                            error_dict["description"] = error_dict["detail"] = await response.text()
                            yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                            return
                        ans = ""
                        prompt_tokens = completion_tokens = 0
                        async for chunk in response.content:
                            chunk = chunk.decode(encoding='utf-8')
                            if chunk.endswith('\n'):
                                chunk = chunk[:-1]
                            elif chunk.endswith('\r\n'):
                                chunk = chunk[:-2]
                            if chunk != "data: [DONE]" and chunk != "":
                                if chunk[0:6] == "data: ":
                                    chunk = chunk[6:]
                                datas = json.loads(chunk)
                                if "result" not in datas.keys():
                                    yield chunk
                                    return
                                content = datas["result"]
                                if datas["sentence_id"] == 0:
                                    format_json = {
                                        "id": datas["id"],
                                        "model": self.api_model,
                                        "choices": [
                                            {
                                                "index": 0,
                                                "delta": {
                                                    "role": "assistant"
                                                },
                                                "finish_reason": None
                                            }
                                        ],
                                        "usage": {
                                            "prompt_tokens": 0,
                                            "completion_tokens": 0,
                                            "total_tokens": 0
                                        }
                                    }
                                    yield json.dumps(format_json, ensure_ascii=False)
                                format_json = {
                                    "id": datas["id"],
                                    "model": self.api_model,
                                    "choices": [
                                        {
                                            "index": 0,
                                            "delta": {
                                                "content": content
                                            },
                                            "finish_reason": None
                                        }
                                    ],
                                    "usage": {
                                        "prompt_tokens": 0,
                                        "completion_tokens": 0,
                                        "total_tokens": 0
                                    }
                                }
                                yield json.dumps(format_json, ensure_ascii=False)
                                ans += content
                                if datas["is_end"] is True:
                                    if "prompt_tokens_details" not in datas:
                                        datas["usage"]["prompt_tokens_details"] = {
                                            "cached_tokens": 0
                                        }
                                        # 计算 uncached_tokens
                                    cached_tokens = datas["usage"]["prompt_tokens_details"].get(
                                        "cached_tokens", 0)
                                    datas["usage"]["prompt_tokens_details"][
                                        "uncached_tokens"] = prompt_tokens - cached_tokens
                                    format_json = {
                                        "id": datas["id"],
                                        "model": self.api_model,
                                        "choices": [
                                            {
                                                "index": 0,
                                                "delta": {},
                                                "finish_reason": "stop"
                                            }
                                        ],
                                        "usage": datas["usage"]
                                    }
                                    prompt_tokens = datas["usage"]["prompt_tokens"]
                                    completion_tokens = datas["usage"]["completion_tokens"]
                                    yield json.dumps(format_json, ensure_ascii=False)
                                    yield "[DONE]"
                                    break
                        log_info = logics.AddModelUsedAudit(
                            model_id=self.model_id, user_id=user_id, input_tokens=prompt_tokens,
                            output_tokens=completion_tokens)
                        await add_llm_model_call_log(log_info)
                        if get_logger():
                            get_logger().info(
                                f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                                f'"prompt_tokens":{prompt_tokens},"completion_tokens":{completion_tokens},'
                                f'"total_tokens":{prompt_tokens + completion_tokens},"func_module":{func_module},"status":"success"}}')
                        return
            except aiohttp.ClientError as e:
                if retry_time <= 0:
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    error_dict["description"] = f"大模型: {self.api_model} 连接失败，请检查该服务是否可用"
                    error_dict["detail"] = str(e)
                    if get_logger():
                        get_logger().info(
                            f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                            f'"prompt_tokens":0,"completion_tokens":0,'
                            f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                    yield json.dumps(error_dict, ensure_ascii=False)
                    StandLogger.error(json.dumps(error_dict, ensure_ascii=False))
                    return
                else:
                    StandLogger.warn(f"大模型: {self.api_model} 连接失败，1秒后重试")
                    await asyncio.sleep(1)
            except Exception as e:
                error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                error_dict["detail"] = str(e)
                if get_logger():
                    get_logger().info(
                        f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                        f'"prompt_tokens":0,"completion_tokens":0,'
                        f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                yield json.dumps(error_dict, ensure_ascii=False)
                return

    async def chat_completion_stream(self, messages, user_id, return_info, cache=False):
        messages_json = json.dumps(messages, ensure_ascii=False)
        retry_time = 3
        while retry_time > 0:
            retry_time -= 1
            try:
                headers = {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                }
                baidu_url = f"https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id={self.api_key}&client_secret={self.secret_key}"
                async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                    async with session.post(baidu_url, headers=headers, ssl=False) as access_res:
                        if access_res.status != 200:
                            tmp_map = await access_res.json()
                            error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                            if "detail" in tmp_map.keys():
                                error_dict["detail"] = tmp_map["detail"]
                            if "error" in tmp_map.keys() and "message" in tmp_map["error"].keys():
                                error_dict["detail"] = tmp_map["error"]["message"]
                            yield json.dumps(error_dict, ensure_ascii=False)
                            return
                        access_info = await access_res.json()
                        access_token = access_info["access_token"]
                        start_time = time.time()
                        token_len = 0
                        system = None
                        new_messages = []
                        for i in range(0, len(messages)):
                            if messages[i]["role"] == "system":
                                if len(messages) > 1:
                                    if system is None:
                                        system = messages[i]["content"]
                                else:
                                    messages[i]["role"] = "user"
                                    new_messages = messages
                            else:
                                new_messages.append(messages[i])
                        params = {
                            "messages": new_messages,
                            "model": self.api_model,
                            "stream": True,
                            "top_p": self.top_p,
                            "temperature": self.temperature,
                            "frequency_penalty": self.frequency_penalty,
                            "presence_penalty": self.presence_penalty,
                            "max_output_tokens": self.max_tokens,
                            "top_k": self.top_k
                        }
                        if system is not None:
                            params["system"] = system
                        async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                            url = self.api_url + f"?access_token={access_token}"
                            async with session.post(url, json=params, headers=headers, ssl=False) as response:
                                response.encoding = 'utf-8'
                                if response.status != 200:
                                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                                    error_dict["description"] = error_dict["detail"] = response.content.decode(
                                        encoding="utf-8")
                                    yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                                    return
                                ans = ""
                                prompt_tokens = completion_tokens = 0
                                async for chunk in response.content:
                                    chunk = chunk.decode(encoding='utf-8')
                                    if chunk.endswith('\n'):
                                        chunk = chunk[:-1]
                                    elif chunk.endswith('\r\n'):
                                        chunk = chunk[:-2]
                                    if chunk != "data: [DONE]" and chunk != "":
                                        if chunk[0:6] == "data: ":
                                            chunk = chunk[6:]
                                        datas = json.loads(chunk)
                                        if "result" not in datas.keys():
                                            yield "--error--" + chunk
                                            return
                                        content = datas["result"]
                                        yield content
                                        if datas["is_end"] is True:
                                            prompt_tokens = datas["usage"]["prompt_tokens"]
                                            completion_tokens = datas["usage"]["completion_tokens"]
                                            break
                                end_time = time.time()
                                if return_info:
                                    yield "--info--" + str(json.dumps({"time": str(end_time - start_time),
                                                                       "token_len": completion_tokens,
                                                                       "prompt_tokens": prompt_tokens}))
                                yield "--end--"
                                log_info = logics.AddModelUsedAudit(
                                    model_id=self.model_id, user_id=user_id, input_tokens=prompt_tokens,
                                    output_tokens=completion_tokens)
                                await add_llm_model_call_log(log_info)
                                return
            except aiohttp.ClientError as e:
                if retry_time <= 0:
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    error_dict["description"] = f"大模型: {self.api_model} 连接失败，请检查该服务是否可用"
                    error_dict["detail"] = str(e)
                    yield json.dumps(error_dict, ensure_ascii=False)
                    StandLogger.error(json.dumps(error_dict, ensure_ascii=False))
                    return
                else:
                    StandLogger.warn(f"大模型: {self.api_model} 连接失败，1秒后重试")
                    await asyncio.sleep(1)
            except Exception as e:
                error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                error_dict["detail"] = str(e)
                yield json.dumps(error_dict, ensure_ascii=False)
                return


# 其他模型类
class OtherClient:
    def __init__(self, api_url, api_model, api_key, model_id,
                 temperature, top_p, frequency_penalty, presence_penalty, max_tokens, top_k=1, response_format={},
                 stop=None, model_type="llm", tools=None, tool_choice=None):
        self.api_url = api_url
        self.api_model = api_model
        self.temperature = temperature
        self.top_p = top_p
        self.frequency_penalty = frequency_penalty
        self.presence_penalty = presence_penalty
        self.max_tokens = max_tokens
        self.api_key = api_key
        self.model_id = model_id
        self.top_k = top_k
        self.response_format = response_format
        self.stop = stop
        self.model_type = model_type
        self.tools = tools
        self.tool_choice = tool_choice

    async def chat_completion(self, messages, user_id, func_module):
        retry_time = 3
        while retry_time > 0:
            retry_time -= 1
            try:
                if messages[len(messages) - 1]["role"] != "user" and self.api_model.find("qianxun") != -1:
                    error_dict = ModelError.copy()
                    error_dict["description"] = error_dict["detail"] = error_dict[
                        "solution"] = "千循大模型只支持最后一条消息role为user"
                    return error_dict
                start_time = time.time()
                params = {
                    "messages": messages,
                    "model": self.api_model,
                    "stream": False,
                    "top_p": self.top_p,
                    "temperature": self.temperature,
                    "frequency_penalty": self.frequency_penalty,
                    "presence_penalty": self.presence_penalty,
                    "max_tokens": self.max_tokens,
                    "top_k": self.top_k
                }
                if self.response_format != {}:
                    params["response_format"] = self.response_format
                if self.stop is not None:
                    params["stop"] = self.stop
                if self.tools is not None:
                    params["tools"] = self.tools
                if self.tool_choice is not None:
                    params["tool_choice"] = self.tool_choice

                headers = {
                    "Authorization": f"Bearer {self.api_key}"
                }
                conn = aiohttp.TCPConnector(verify_ssl=False)
                async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
                    async with session.post(self.api_url, json=params, headers=headers) as resp:
                        res = await resp.text()
                        result = json.loads(res)
                        if resp.status == 200:
                            log_info = logics.AddModelUsedAudit(
                                model_id=self.model_id, user_id=user_id, input_tokens=result["usage"]["prompt_tokens"],
                                output_tokens=result["usage"]["completion_tokens"])
                            await add_llm_model_call_log(log_info)
                            if self.model_type == "rlm" and not result["choices"][0]["message"].get("reasoning_content",
                                                                                                    None):
                                # 提取thinking标签
                                pattern = r"<think>(.*?)</think>"
                                match = re.search(pattern, result["choices"][0]["message"]["content"], re.DOTALL)
                                if match:
                                    result["choices"][0]["message"]["reasoning_content"] = match.group(1)
                                    result["choices"][0]["message"]["content"] = re.sub(pattern, "",
                                                                                        result["choices"][0]["message"][
                                                                                            "content"], flags=re.DOTALL)
                                else:
                                    pattern = r"(.*?)</think>"
                                    match = re.search(pattern, result["choices"][0]["message"]["content"], re.DOTALL)
                                    if match:
                                        result["choices"][0]["message"]["reasoning_content"] = match.group(1)
                                        result["choices"][0]["message"]["content"] = re.sub(pattern, "",
                                                                                            result["choices"][0][
                                                                                                "message"]["content"],
                                                                                            flags=re.DOTALL)
                            usage = result.get("usage", {})
                            prompt_tokens = usage.get("prompt_tokens", 0)
                            completion_tokens = usage.get("completion_tokens", 0)
                            total_tokens = usage.get("total_tokens", 0)
                            if "prompt_tokens_details" not in usage:
                                usage["prompt_tokens_details"] = {
                                    "cached_tokens": 0
                                }
                            if not usage["prompt_tokens_details"]:
                                usage["prompt_tokens_details"] = {
                                    "cached_tokens": 0
                                }
                                # 计算 uncached_tokens
                            cached_tokens = usage["prompt_tokens_details"].get(
                                "cached_tokens", 0)
                            usage["prompt_tokens_details"][
                                "uncached_tokens"] = prompt_tokens - cached_tokens
                            result["usage"] = usage
                            if get_logger():
                                get_logger().info(
                                    f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                                    f'"prompt_tokens":{prompt_tokens},"completion_tokens":{completion_tokens},'
                                    f'"total_tokens":{total_tokens},"func_module":{func_module},"status":"success"}}')
                            return result
                        else:
                            tmp_map = result
                            error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                            if "detail" in tmp_map.keys():
                                error_dict["detail"] = tmp_map["detail"]
                            if "error" in tmp_map.keys() and "message" in tmp_map["error"].keys():
                                error_dict["detail"] = tmp_map["error"]["message"]
                            if get_logger():
                                get_logger().info(
                                    f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                                    f'"prompt_tokens":0,"completion_tokens":0,'
                                    f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                            return error_dict
            except Exception as e:
                StandLogger.error(e.args)
                if get_logger():
                    get_logger().info(
                        f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                        f'"prompt_tokens":0,"completion_tokens":0,'
                        f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                if retry_time <= 0:
                    raise e
                else:
                    await asyncio.sleep(0.5)

    async def chat_completion_stream_openai(self, messages, user_id, return_info, model_data, func_module):
        retry_time = 3
        while retry_time > 0:
            retry_time -= 1
            try:
                chunk_id = ""
                if messages[len(messages) - 1]["role"] != "user" and self.api_model.find("qianxun") != -1:
                    error_dict = ModelError.copy()
                    error_dict["description"] = error_dict["detail"] = error_dict[
                        "solution"] = "千循大模型只支持最后一条消息role为user"
                    yield json.dumps(error_dict, ensure_ascii=False)
                token_len = 0
                params = {
                    "messages": messages,
                    "model": self.api_model,
                    "stream": True,
                    "top_p": self.top_p,
                    "temperature": self.temperature,
                    "frequency_penalty": self.frequency_penalty,
                    "presence_penalty": self.presence_penalty,
                    "max_tokens": self.max_tokens,
                    "top_k": self.top_k,
                    "stream_options": {"include_usage": True}
                }
                if self.response_format != {}:
                    params["response_format"] = self.response_format
                if self.stop is not None:
                    params["stop"] = self.stop
                if self.tools is not None:
                    params["tools"] = self.tools
                if self.tool_choice is not None:
                    params["tool_choice"] = self.tool_choice

                for i in range(0, len(messages)):
                    if messages[i]["content"] == "":
                        messages[i]["content"] = " "
                headers = {
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json"
                }
                prompt_tokens = 0
                completion_tokens = 0
                async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                    async with session.post(self.api_url, json=params, headers=headers, ssl=False) as response:
                        response.encoding = 'utf-8'
                        if response.status != 200:
                            error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                            error_dict["description"] = error_dict["detail"] = await response.text()
                            StandLogger.error(json.dumps(error_dict, ensure_ascii=False))
                            yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                            return
                        ans = ""
                        think_str = ""
                        rlm_type = ""
                        think_done = False
                        think_start = False
                        async for chunk in response.content:
                            chunk = chunk.decode(encoding='utf-8')
                            if chunk.endswith('\n'):
                                chunk = chunk[:-1]
                            elif chunk.endswith('\r\n'):
                                chunk = chunk[:-2]
                            if len(chunk) >= 6 and chunk[0:6] != "data: ":
                                continue
                            if chunk != "data: [DONE]" and chunk != "":
                                if self.model_type == "rlm":
                                    if chunk[0:6] == "data: ":
                                        chunk = chunk[6:]
                                    try:
                                        datas = json.loads(chunk)
                                    except Exception:
                                        continue
                                    try:
                                        if chunk_id == "":
                                            chunk_id = datas.get("id", "")
                                        if datas["choices"]:
                                            delta = datas["choices"][0]["delta"]
                                            if "content" in delta and delta["content"] is not None:
                                                token = delta["content"]
                                                if token is not None:
                                                    token_len += 1
                                                    ans += token
                                        else:
                                            ans += ""
                                    except Exception as e:
                                        if "error" in datas.keys():
                                            yield json.dumps(datas, ensure_ascii=False)
                                            return
                                    if rlm_type == "":
                                        if "reasoning_content" in datas["choices"][0]["delta"]:
                                            rlm_type = "saas"
                                        elif datas["choices"][0]["delta"].get("content", "") not in ["", None]:
                                            rlm_type = "inner"
                                    if rlm_type == "saas":
                                        if "usage" in datas.keys() and completion_tokens == 0:
                                            try:
                                                prompt_tokens = datas["usage"]["prompt_tokens"]
                                                completion_tokens = datas["usage"]["completion_tokens"]
                                                total_tokens = datas["usage"]["total_tokens"]
                                                # 处理 usage 信息，添加 uncached_tokens 字段
                                                if "prompt_tokens_details" not in datas["usage"]:
                                                    datas["usage"]["prompt_tokens_details"] = {
                                                        "cache_type": "implicit",
                                                        "cached_tokens": 0
                                                    }
                                                prompt_tokens_details = datas["usage"]["prompt_tokens_details"]
                                                cached_tokens = prompt_tokens_details.get("cached_tokens", 0)
                                                uncached_tokens = prompt_tokens - cached_tokens
                                                datas["usage"]["prompt_tokens_details"][
                                                    "uncached_tokens"] = uncached_tokens
                                            except Exception:
                                                pass
                                        yield json.dumps(datas, ensure_ascii=False)
                                    elif rlm_type == "inner":
                                        if "usage" in datas.keys() and completion_tokens == 0:
                                            try:
                                                prompt_tokens = datas["usage"]["prompt_tokens"]
                                                completion_tokens = datas["usage"]["completion_tokens"]
                                                total_tokens = datas["usage"]["total_tokens"]
                                                # 处理 usage 信息，添加 uncached_tokens 字段
                                                if "prompt_tokens_details" not in datas["usage"]:
                                                    datas["usage"]["prompt_tokens_details"] = {
                                                        "cache_type": "implicit",
                                                        "cached_tokens": 0
                                                    }
                                                prompt_tokens_details = datas["usage"]["prompt_tokens_details"]
                                                cached_tokens = prompt_tokens_details.get("cached_tokens", 0)
                                                uncached_tokens = prompt_tokens - cached_tokens
                                                datas["usage"]["prompt_tokens_details"][
                                                    "uncached_tokens"] = uncached_tokens
                                            except Exception:
                                                pass
                                        if think_done is False:
                                            if think_start is False:
                                                start_pattern = r"<think>(.*?)"
                                                match = re.search(start_pattern, ans, re.DOTALL)
                                                if match:
                                                    think_start = True
                                                    if datas["choices"]:
                                                        datas["choices"][0]["delta"]["reasoning_content"] = re.sub(
                                                            start_pattern, "", ans, flags=re.DOTALL)
                                                        think_str = ans
                                                        datas["choices"][0]["delta"]["content"] = None
                                                        yield json.dumps(datas, ensure_ascii=False)
                                                if not match and len(ans) >= 7:
                                                    if datas["choices"]:
                                                        think_start = True
                                                        datas["choices"][0]["delta"]["reasoning_content"] = ans
                                                        think_str = ans
                                                        datas["choices"][0]["delta"]["content"] = None
                                                        yield json.dumps(datas, ensure_ascii=False)
                                            else:
                                                pattern = r"(.*?)</think>"
                                                match = re.search(pattern, ans, re.DOTALL)
                                                if match:
                                                    think_done = True
                                                    reasoning_datas = datas.copy()
                                                    if datas["choices"]:
                                                        reasoning_datas["choices"][0]["delta"][
                                                            "reasoning_content"] = match.group(1).replace(
                                                            think_str.replace("<think>", ""), "")
                                                        reasoning_datas["choices"][0]["delta"]["content"] = None
                                                        yield json.dumps(reasoning_datas, ensure_ascii=False)
                                                    if datas["choices"]:
                                                        ans = re.sub(pattern, "", ans, flags=re.DOTALL)
                                                        datas["choices"][0]["delta"]["reasoning_content"] = None
                                                        datas["choices"][0]["delta"]["content"] = ans
                                                        yield json.dumps(datas, ensure_ascii=False)
                                                else:
                                                    if not has_common_substring(ans.replace("<think>", ""), "</think>"):
                                                        if datas["choices"]:
                                                            datas["choices"][0]["delta"]["reasoning_content"] = ans.replace(
                                                                think_str, "")
                                                            datas["choices"][0]["delta"]["content"] = None
                                                            think_str = ans
                                                            yield json.dumps(datas, ensure_ascii=False)



                                        else:
                                            if datas["choices"]:
                                                datas["choices"][0]["delta"]["reasoning_content"] = None
                                            yield json.dumps(datas, ensure_ascii=False)
                                else:
                                    if chunk[0:6] == "data: ":
                                        chunk = chunk[6:]
                                    try:
                                        datas = json.loads(chunk)
                                    except Exception as e:
                                        continue
                                    # try:
                                    #     delta = datas["choices"][0]["delta"]
                                    #     if "content" in delta:
                                    #         token = delta["content"]
                                    #         if token is not None:
                                    #             token_len += 1
                                    #             ans += token
                                    # except Exception as e:
                                    #     if "error" in datas.keys():
                                    #         yield json.dumps(datas, ensure_ascii=False)
                                    #         return
                                    if "usage" in datas.keys() and completion_tokens == 0:
                                        try:
                                            prompt_tokens = datas["usage"]["prompt_tokens"]
                                            completion_tokens = datas["usage"]["completion_tokens"]
                                            total_tokens = datas["usage"]["total_tokens"]
                                            # 处理 usage 信息，添加 uncached_tokens 字段
                                            if "prompt_tokens_details" not in datas["usage"]:
                                                datas["usage"]["prompt_tokens_details"] = {
                                                    "cache_type": "implicit",
                                                    "cached_tokens": 0
                                                }
                                            prompt_tokens_details = datas["usage"]["prompt_tokens_details"]
                                            cached_tokens = prompt_tokens_details.get("cached_tokens", 0)
                                            uncached_tokens = prompt_tokens - cached_tokens
                                            datas["usage"]["prompt_tokens_details"]["uncached_tokens"] = uncached_tokens
                                        except Exception:
                                            pass
                                    yield json.dumps(datas, ensure_ascii=False)

                            else:

                                if chunk == "data: [DONE]":
                                    yield "[DONE]"
                        log_info = logics.AddModelUsedAudit(
                            model_id=self.model_id, user_id=user_id, input_tokens=prompt_tokens,
                            output_tokens=completion_tokens)
                        await add_llm_model_call_log(log_info)
                        if get_logger():
                            get_logger().info(
                                f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                                f'"prompt_tokens":{prompt_tokens},"completion_tokens":{completion_tokens},'
                                f'"total_tokens":{total_tokens},"func_module":{func_module},"status":"success"}}')
                        return
            except aiohttp.ClientError as e:
                StandLogger.error(f"call llmModelError {self.api_model} error params={params},headers={headers},error={e}")
                if retry_time <= 0:
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    error_dict["description"] = f"大模型: {self.api_model} 连接失败，请检查该服务是否可用"
                    error_dict["detail"] = str(e)
                    yield json.dumps(error_dict, ensure_ascii=False)
                    StandLogger.error(json.dumps(error_dict, ensure_ascii=False))
                    if get_logger():
                        get_logger().info(
                            f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                            f'"prompt_tokens":0,"completion_tokens":0,'
                            f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                    return
                else:
                    StandLogger.warn(f"大模型: {self.api_model} 连接失败，1秒后重试")
                    await asyncio.sleep(1)
            except Exception as e:
                StandLogger.error(f"call llmModelError {self.api_model} error params={params},headers={headers},error={e}")
                error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                error_dict["detail"] = str(e)
                if get_logger():
                    get_logger().info(
                        f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                        f'"prompt_tokens":0,"completion_tokens":0,'
                        f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                yield json.dumps(error_dict, ensure_ascii=False)
                return

    async def chat_completion_stream(self, messages, user_id, return_info, model_data):
        retry_time = 3
        while retry_time > 0:
            retry_time -= 1
            try:
                if messages[len(messages) - 1]["role"] != "user" and self.api_model.find("qianxun") != -1:
                    error_dict = ModelError.copy()
                    error_dict["description"] = error_dict["detail"] = error_dict[
                        "solution"] = "千循大模型只支持最后一条消息role为user"
                    yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                start_time = time.time()
                token_len = 0
                params = {
                    "messages": messages,
                    "model": self.api_model,
                    "stream": True,
                    "top_p": self.top_p,
                    "temperature": self.temperature,
                    "frequency_penalty": self.frequency_penalty,
                    "presence_penalty": self.presence_penalty,
                    "max_tokens": self.max_tokens,
                    "top_k": self.top_k
                }
                if self.response_format != {}:
                    params["response_format"] = self.response_format
                if self.stop is not None:
                    params["stop"] = self.stop
                if self.tools is not None:
                    params["tools"] = self.tools
                if self.tool_choice is not None:
                    params["tool_choice"] = self.tool_choice

                for i in range(0, len(messages)):
                    if messages[i]["content"] == "":
                        messages[i]["content"] = " "
                headers = {
                    "Authorization": f"Bearer {self.api_key}"
                }
                prompt_tokens = 0
                completion_tokens = 0
                # try:
                async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                    async with session.post(self.api_url, json=params, headers=headers, ssl=False) as response:
                        response.encoding = 'utf-8'
                        if response.status != 200:
                            error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                            res = await response.text()
                            error_dict["description"] = error_dict["detail"] = res.replace('"', "'")
                            yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                            return
                        ans = ""
                        think_end = False
                        if self.model_type == "llm":
                            think_end = True
                        rlm_type = ""
                        async for chunk in response.content:
                            chunk = chunk.decode('utf-8')
                            if chunk.endswith('\n'):
                                chunk = chunk[:-1]
                            elif chunk.endswith('\r\n'):
                                chunk = chunk[:-2]
                            if len(chunk) >= 6 and chunk[0:6] != "data: ":
                                continue
                            # StandLogger.info_log(chunk)
                            if chunk != "data: [DONE]" and chunk != "":
                                if chunk[0:6] == "data: ":
                                    chunk = chunk[6:]
                                try:
                                    datas = json.loads(chunk)
                                except Exception:
                                    continue
                                try:
                                    delta = datas["choices"][0]["delta"]
                                    if rlm_type == "" and self.model_type == "rlm":
                                        if "content" in delta and "reasoning_content" not in delta and delta[
                                            "content"] not in [None, ""]:
                                            rlm_type = "inner"
                                        elif "reasoning_content" in delta:
                                            rlm_type = "saas"
                                            think_end = True

                                    if "content" in delta:
                                        token = delta["content"]
                                        token_len += 1
                                        if token is not None:
                                            ans += token

                                            if think_end:
                                                yield token
                                            if not think_end:
                                                pattern = r"(.*?)</think>"
                                                match = re.search(pattern, ans, re.DOTALL)
                                                if match:
                                                    ans = re.sub(pattern, "", ans, flags=re.DOTALL)
                                                    yield ans
                                                    think_end = True
                                            # print(token)
                                except Exception as e:
                                    if "error" in datas.keys():
                                        yield "--error--" + json.dumps(datas, ensure_ascii=False)
                                        return
                                if "usage" in datas.keys() and completion_tokens == 0:
                                    try:
                                        prompt_tokens = datas["usage"]["prompt_tokens"]
                                        completion_tokens = datas["usage"]["completion_tokens"]

                                    except Exception:
                                        pass
                            elif chunk == "data: [DONE]":
                                if not think_end:
                                    yield ans
                                if ans == "":
                                    params["stream"] = False
                                    response = requests.post(self.api_url, json=params, headers=headers, verify=False)
                                    if response.status_code != 200:
                                        error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                                        error_dict["description"] = error_dict["detail"] = json.dumps(response.json(),
                                                                                                      ensure_ascii=False)
                                        yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                                        return
                                end_time = time.time()
                                if return_info:
                                    if completion_tokens == 0:
                                        completion_tokens = token_len
                                        prompt_str = ""
                                        for item_prompt in messages:
                                            prompt_str += item_prompt["content"]
                                    yield "--info--" + str(json.dumps({"time": str(end_time - start_time),
                                                                       "token_len": completion_tokens,
                                                                       "prompt_tokens": prompt_tokens}))
                                yield "--end--"
                        log_info = logics.AddModelUsedAudit(
                            model_id=self.model_id, user_id=user_id, input_tokens=prompt_tokens,
                            output_tokens=completion_tokens)
                        await add_llm_model_call_log(log_info)
                        return
            except aiohttp.ClientError as e:
                if retry_time <= 0:
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    error_dict["description"] = f"大模型: {self.api_model} 连接失败，请检查该服务是否可用"
                    error_dict["detail"] = str(e)
                    yield json.dumps(error_dict, ensure_ascii=False)
                    StandLogger.error(json.dumps(error_dict, ensure_ascii=False))
                    return
                else:
                    StandLogger.warn(f"大模型: {self.api_model} 连接失败，1秒后重试")
                    await asyncio.sleep(1)
            except Exception as e:
                error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                error_dict["detail"] = str(e)
                yield json.dumps(error_dict, ensure_ascii=False)
                return


class ClaudeClient:
    def __init__(self, api_url, api_model, api_key, model_id,
                 temperature, top_p, frequency_penalty, presence_penalty, max_tokens, top_k=1, system=[],
                 tools=None, tool_choice=None):
        self.api_url = api_url
        self.api_model = api_model
        self.temperature = temperature
        self.top_p = top_p
        self.frequency_penalty = frequency_penalty
        self.presence_penalty = presence_penalty
        self.max_tokens = max_tokens
        self.api_key = api_key
        self.model_id = model_id
        self.top_k = top_k
        self.system = system
        self.tools = tools
        self.tool_choice = tool_choice
        if self.tools:
            for i in range(0, len(self.tools)):
                self.tools[i] = self.tools[i]["function"]
                self.tools[i]["input_schema"] = self.tools[i].pop("parameters")
        if tool_choice == "auto":
            self.tool_choice = {"type": "auto"}
        elif tool_choice == "required":
            self.tool_choice = {"type": "any"}
        elif tool_choice is not None:
            self.tool_choice = {"type": "tool", "name": self.tool_choice["function"]["name"]}

    async def chat_completion(self, messages, user_id, func_module, cache=False):
        messages_json = json.dumps(messages, ensure_ascii=False)
        system = None
        for i in range(0, len(messages)):

            if messages[i]["role"] == "system":
                if len(messages) > 1:
                    if system is None:
                        system = messages[i]["content"]
                    messages.pop(i)
                else:
                    messages[i]["role"] = "user"
            elif messages[i]["role"] == "tool":

                tmp_message = {
                    "role": "user",
                    "content": [
                        {
                            "type": "tool_result",
                            "tool_use_id": messages[i]["tool_call_id"],
                            "content": messages[i]["content"]
                        }
                    ]
                }
                messages[i] = tmp_message
            elif messages[i]["role"] == "assistant":
                if "tool_calls" in messages[i].keys() and messages[i]["tool_calls"] is not None and len(
                        messages[i]["tool_calls"]) > 0:
                    tmp_message = {
                        "role": "assistant",
                        "content": [
                            {
                                "type": "text",
                                "text": messages[i].get("content", "")
                            }
                        ]
                    }
                    if messages[i].get("content", "") == "":
                        tmp_message["content"] = []
                    for j in range(0, len(messages[i]["tool_calls"])):
                        tmp_message["content"].append(
                            {
                                "type": "tool_use",
                                "id": messages[i]["tool_calls"][j]["id"],
                                "name": messages[i]["tool_calls"][j]["function"]["name"],
                                "input": json.loads(messages[i]["tool_calls"][j]["function"]["arguments"])
                            })
                    messages[i] = tmp_message
            if "tool_calls" in messages[i].keys():
                messages[i].pop("tool_calls")
            if "tool_call_id" in messages[i].keys():
                messages[i].pop("tool_call_id")

        params = {
            "messages": messages,
            "model": self.api_model,
            "stream": False,
            "top_p": self.top_p,
            "temperature": self.temperature,
            "max_tokens": self.max_tokens,
            "top_k": self.top_k
        }
        if self.system:
            params["system"] = self.system
        if self.tools:
            params["tools"] = self.tools
        if self.tool_choice:
            params["tool_choice"] = self.tool_choice
        headers = {
            "content-type": "application/json",
            "x-api-key": f"{self.api_key}",
            "anthropic-version": "2023-06-01",
            "anthropic-beta": "prompt-caching-2024-07-31",

        }
        async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
            async with session.post(self.api_url, json=params, headers=headers) as resp:
                result = await resp.text()
                res = json.loads(result)
                if resp.status == 200:

                    log_info = logics.AddModelUsedAudit(
                        model_id=self.model_id, user_id=user_id, input_tokens=res["usage"]["input_tokens"],
                        output_tokens=res["usage"]["output_tokens"])
                    await add_llm_model_call_log(log_info)
                    res_json = res
                    stop_reason = res_json["stop_reason"]
                    finish_reason = None
                    if stop_reason == "end_turn":
                        finish_reason = "stop"
                    elif stop_reason == "max_tokens":
                        finish_reason = "length"
                    elif stop_reason == "stop_sequence":
                        finish_reason = "content_filter"
                    elif stop_reason == "tool_use":
                        finish_reason = "tool_calls"
                    StandLogger.info_log("original_usage:" + str(res_json["usage"]))
                    format_res = {
                        "id": res_json["id"],
                        "object": "chat.completion",
                        "created": int(time.time()),
                        "model": res_json["model"],
                        "choices": [
                            {
                                "index": 0,
                                "message": {
                                    "role": "assistant",
                                    "content": res_json["content"][0].get("text", "")
                                },
                                "finish_reason": finish_reason
                            }
                        ],
                        "usage": {
                            "prompt_tokens": res_json["usage"]["input_tokens"] + res_json["usage"].get(
                                "cache_creation_input_tokens", 0) + res_json["usage"].get("cache_read_input_tokens", 0),
                            "total_tokens": res_json["usage"]["input_tokens"] + res_json["usage"].get(
                                "cache_creation_input_tokens", 0) + res_json["usage"].get("cache_read_input_tokens",
                                                                                          0) + res_json["usage"][
                                                "output_tokens"],
                            "completion_tokens": res_json["usage"]["output_tokens"],
                            "prompt_cache_hit_tokens": res_json["usage"].get("cache_read_input_tokens", 0),
                            "prompt_cache_miss_tokens": res_json["usage"]["input_tokens"] + res_json["usage"].get(
                                "cache_creation_input_tokens", 0)
                        }
                    }
                    if len(res_json["content"]) > 0:
                        format_res["choices"][0]["message"]["tool_calls"] = []
                        for i in range(0, len(res_json["content"])):
                            if res_json["content"][i]["type"] == "tool_use":
                                format_res["choices"][0]["message"]["tool_calls"].append({
                                    "index": len(format_res["choices"][0]["message"]["tool_calls"]),
                                    "id": res_json["content"][i]["id"],
                                    "type": "function",
                                    "function": {
                                        "name": res_json["content"][i]["name"],
                                        "arguments": json.dumps(res_json["content"][i]["input"], ensure_ascii=False)
                                    }
                                })
                    usage = format_res.get("usage", {})
                    prompt_tokens = usage.get("prompt_tokens", 0)
                    completion_tokens = usage.get("completion_tokens", 0)
                    total_tokens = usage.get("total_tokens", 0)
                    if get_logger():
                        get_logger().info(
                            f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                            f'"prompt_tokens":{prompt_tokens},"completion_tokens":{completion_tokens},'
                            f'"total_tokens":{total_tokens},"func_module":{func_module},"status":"success"}}')
                    return format_res
                else:
                    tmp_map = res
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    if "detail" in tmp_map.keys():
                        error_dict["detail"] = tmp_map["detail"]
                    if "error" in tmp_map.keys() and "message" in tmp_map["error"].keys():
                        error_dict["detail"] = tmp_map["error"]["message"]
                    if get_logger():
                        get_logger().info(
                            f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                            f'"prompt_tokens":0,"completion_tokens":0,'
                            f'"total_tokens":0,"func_module":{func_module},"status":"failed"}}')
                    return error_dict

    async def chat_completion_stream_openai(self, messages, user_id, func_module, cache=False):
        system = None
        messages_json = json.dumps(messages, ensure_ascii=False)
        for i in range(0, len(messages)):

            if messages[i]["role"] == "system":
                if len(messages) > 1:
                    if system is None:
                        system = messages[i]["content"]
                    messages.pop(i)
                else:
                    messages[i]["role"] = "user"
            elif messages[i]["role"] == "tool":

                tmp_message = {
                    "role": "user",
                    "content": [
                        {
                            "type": "tool_result",
                            "tool_use_id": messages[i]["tool_call_id"],
                            "content": messages[i]["content"]
                        }
                    ]
                }
                messages[i] = tmp_message
            elif messages[i]["role"] == "assistant":
                if "tool_calls" in messages[i].keys() and messages[i]["tool_calls"] is not None and len(
                        messages[i]["tool_calls"]) > 0:
                    tmp_message = {
                        "role": "assistant",
                        "content": [
                            {
                                "type": "text",
                                "text": messages[i].get("content", "")
                            }
                        ]
                    }
                    if messages[i].get("content", "") == "":
                        tmp_message["content"] = []
                    for j in range(0, len(messages[i]["tool_calls"])):
                        tmp_message["content"].append(
                            {
                                "type": "tool_use",
                                "id": messages[i]["tool_calls"][j]["id"],
                                "name": messages[i]["tool_calls"][j]["function"]["name"],
                                "input": json.loads(messages[i]["tool_calls"][j]["function"]["arguments"])
                            })
                    messages[i] = tmp_message
            if "tool_calls" in messages[i].keys():
                messages[i].pop("tool_calls")
            if "tool_call_id" in messages[i].keys():
                messages[i].pop("tool_call_id")

        params = {
            "messages": messages,
            "model": self.api_model,
            "stream": True,
            "top_p": self.top_p,
            "temperature": self.temperature,
            "max_tokens": self.max_tokens,
            "top_k": self.top_k
        }
        if self.system:
            params["system"] = self.system
        if self.tools:
            params["tools"] = self.tools
        if self.tool_choice:
            params["tool_choice"] = self.tool_choice
        headers = {
            "x-api-key": f"{self.api_key}",
            "anthropic-version": "2023-06-01",
            "anthropic-beta": "prompt-caching-2024-07-31",
            "content-type": "application/json"
        }
        async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
            async with session.post(self.api_url, json=params, headers=headers, ssl=False) as response:
                response.encoding = 'utf-8'
                if response.status != 200:
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    error_dict["description"] = error_dict["detail"] = await response.text()
                    yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                    return
                chunk_id = None
                chunk_model = None
                prompt_tokens = 0
                completion_tokens = 0
                prompt_cache_hit_tokens = 0
                prompt_cache_miss_tokens = 0
                ans = ""
                async for chunk in response.content:
                    chunk = chunk.decode(encoding='utf-8')
                    if chunk.endswith('\n'):
                        chunk = chunk[:-1]
                    elif chunk.endswith('\r\n'):
                        chunk = chunk[:-2]
                    # StandLogger.info_log(chunk)
                    if chunk.find("data:") != -1:
                        chunk = chunk.replace("data:", "")
                    try:
                        chunk_json = json.loads(chunk)
                    except Exception as e:
                        continue
                    chunk_type = chunk_json["type"]
                    if chunk_type == "message_start":
                        format_json = {
                            "id": chunk_json["message"]["id"],
                            "model": chunk_json["message"]["model"],
                            "choices": [
                                {
                                    "index": 0,
                                    "delta": {
                                        "role": "assistant"
                                    },
                                    "finish_reason": None
                                }
                            ],
                            "usage": {
                                "prompt_tokens": 0,
                                "completion_tokens": 0,
                                "total_tokens": 0
                            }
                        }
                        StandLogger.info_log("original_usage:" + str(chunk_json["message"]["usage"]))
                        prompt_tokens += chunk_json["message"]["usage"]["input_tokens"] + chunk_json["message"][
                            "usage"].get("cache_creation_input_tokens", 0) + chunk_json["message"]["usage"].get(
                            "cache_read_input_tokens", 0)
                        completion_tokens += chunk_json["message"]["usage"]["output_tokens"]
                        prompt_cache_hit_tokens += chunk_json["message"]["usage"].get("cache_read_input_tokens", 0)
                        prompt_cache_miss_tokens += chunk_json["message"]["usage"]["input_tokens"] + \
                                                    chunk_json["message"]["usage"].get("cache_creation_input_tokens", 0)
                        chunk_id = chunk_json["message"]["id"]
                        chunk_model = chunk_json["message"]["model"]
                        yield json.dumps(format_json, ensure_ascii=False)
                    elif chunk_type == "content_block_start":
                        if chunk_json.get("content_block", {}).get("type", "") == "tool_use":
                            format_json = {
                                "id": chunk_id,
                                "model": chunk_model,
                                "choices": [
                                    {
                                        "index": 0,
                                        "delta": {
                                            "tool_calls": [
                                                {
                                                    "index": 0,
                                                    "id": chunk_json["content_block"]["id"],
                                                    "type": "function",
                                                    "function": {
                                                        "name": chunk_json["content_block"]["name"],
                                                        "arguments": ""
                                                    }
                                                }
                                            ]
                                        },
                                        "finish_reason": None
                                    }
                                ],
                                "usage": {
                                    "prompt_tokens": 0,
                                    "completion_tokens": 0,
                                    "total_tokens": 0
                                }
                            }
                            yield json.dumps(format_json, ensure_ascii=False)
                    elif chunk_type == "ping":
                        pass
                    elif chunk_type == "content_block_delta":
                        if "text" in chunk_json["delta"].keys():
                            format_json = {
                                "id": chunk_id,
                                "model": chunk_model,
                                "choices": [
                                    {
                                        "index": 0,
                                        "delta": {
                                            "content": chunk_json["delta"]["text"]
                                        },
                                        "finish_reason": None
                                    }
                                ],
                                "usage": {
                                    "prompt_tokens": 0,
                                    "completion_tokens": 0,
                                    "total_tokens": 0
                                }
                            }
                        elif "partial_json" in chunk_json["delta"].keys():
                            format_json = {
                                "id": chunk_id,
                                "model": chunk_model,
                                "choices": [
                                    {
                                        "index": 0,
                                        "delta": {
                                            "tool_calls": [
                                                {
                                                    "index": 0,
                                                    "function": {
                                                        "arguments": chunk_json["delta"]["partial_json"]
                                                    }
                                                }
                                            ]
                                        },
                                        "finish_reason": None
                                    }
                                ],
                                "usage": {
                                    "prompt_tokens": 0,
                                    "completion_tokens": 0,
                                    "total_tokens": 0
                                }
                            }
                        yield json.dumps(format_json, ensure_ascii=False)
                        if "text" in chunk_json["delta"].keys():
                            ans += chunk_json["delta"]["text"]
                    elif chunk_type == "content_block_stop":
                        pass
                    elif chunk_type == "message_delta":
                        StandLogger.info_log("original_usage:" + str(chunk_json["usage"]))
                        completion_tokens += chunk_json["usage"]["output_tokens"]
                    elif chunk_type == "message_stop":
                        pass
                format_json = {
                    "id": chunk_id,
                    "model": chunk_model,
                    "choices": [
                        {
                            "index": 0,
                            "delta": {},
                            "finish_reason": "stop"
                        }
                    ],
                    "usage": {
                        "prompt_tokens": prompt_tokens,
                        "completion_tokens": completion_tokens,
                        "total_tokens": prompt_tokens + completion_tokens,
                        "prompt_cache_miss_tokens": prompt_cache_miss_tokens,
                        "prompt_cache_hit_tokens": prompt_cache_hit_tokens
                    }
                }
                yield json.dumps(format_json, ensure_ascii=False)
                yield "[DONE]"
                log_info = logics.AddModelUsedAudit(
                    model_id=self.model_id, user_id=user_id, input_tokens=prompt_tokens,
                    output_tokens=completion_tokens)
                await add_llm_model_call_log(log_info)
                if get_logger():
                    get_logger().info(
                        f'{{"model_name":{self.api_model},"resourece_type":"LLM","user_id":{user_id},'
                        f'"prompt_tokens":{prompt_tokens},"completion_tokens":{completion_tokens},'
                        f'"total_tokens":{prompt_tokens + completion_tokens},"func_module":{func_module},"status":"success"}}')

    async def chat_completion_stream(self, messages, user_id, return_info, cache=False):
        messages_json = json.dumps(messages, ensure_ascii=False)
        start_time = time.time()
        system = None
        new_messages = []
        for i in range(0, len(messages)):
            if messages[i]["role"] == "system":
                if len(messages) > 1:
                    if system is None:
                        system = messages[i]["content"]
                else:
                    messages[i]["role"] = "user"
                    new_messages = messages
            else:
                new_messages.append(messages[i])
        params = {
            "messages": new_messages,
            "model": self.api_model,
            "stream": True,
            "top_p": self.top_p,
            "temperature": self.temperature,
            "max_tokens": self.max_tokens,
            "top_k": self.top_k
        }
        if self.system:
            params["system"] = self.system
        headers = {
            "x-api-key": f"{self.api_key}",
            "anthropic-version": "2023-06-01",
            "anthropic-beta": "prompt-caching-2024-07-31",
            "content-type": "application/json"
        }
        async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
            async with session.post(self.api_url, json=params, headers=headers, ssl=False) as response:
                response.encoding = 'utf-8'
                if response.status != 200:
                    error_dict = ModelFactory_ModelController_Model_Error_Error.copy()
                    res = await response.text()
                    error_dict["description"] = error_dict["detail"] = res.replace('"', "'")
                    yield "--error--" + json.dumps(error_dict, ensure_ascii=False)
                    return
                prompt_tokens = 0
                completion_tokens = 0
                prompt_cache_hit_tokens = 0
                prompt_cache_miss_tokens = 0

                ans = ""
                async for chunk in response.content:
                    chunk = chunk.decode(encoding='utf-8')
                    if chunk.endswith('\n'):
                        chunk = chunk[:-1]
                    elif chunk.endswith('\r\n'):
                        chunk = chunk[:-2]
                    # StandLogger.info_log(chunk)
                    if chunk.find("data:") != -1:
                        chunk = chunk.replace("data:", "")
                    try:
                        chunk_json = json.loads(chunk)
                    except Exception as e:
                        continue
                    chunk_type = chunk_json["type"]
                    if chunk_type == "message_start":
                        StandLogger.info_log("original_usage:" + str(chunk_json["message"]["usage"]))
                        prompt_tokens += chunk_json["message"]["usage"]["input_tokens"] + chunk_json["message"][
                            "usage"].get("cache_creation_input_tokens", 0) + chunk_json["message"]["usage"].get(
                            "cache_read_input_tokens", 0)
                        completion_tokens += chunk_json["message"]["usage"]["output_tokens"]
                        prompt_cache_hit_tokens += chunk_json["message"]["usage"].get("cache_read_input_tokens", 0)
                        prompt_cache_miss_tokens += chunk_json["message"]["usage"]["input_tokens"] + \
                                                    chunk_json["message"][
                                                        "usage"].get("cache_creation_input_tokens", 0)

                    elif chunk_type == "content_block_start":
                        pass
                    elif chunk_type == "ping":
                        pass
                    elif chunk_type == "content_block_delta":
                        yield chunk_json["delta"]["text"]
                        ans += chunk_json["delta"]["text"]
                    elif chunk_type == "content_block_stop":
                        pass
                    elif chunk_type == "message_delta":
                        StandLogger.info_log("original_usage:" + str(chunk_json["usage"]))
                        completion_tokens += chunk_json["usage"]["output_tokens"]
                    elif chunk_type == "message_stop":
                        pass
                end_time = time.time()
                if return_info:
                    yield "--info--" + str(json.dumps({"time": str(end_time - start_time),
                                                       "token_len": completion_tokens,
                                                       "prompt_tokens": prompt_tokens,
                                                       "prompt_cache_hit_tokens": prompt_cache_hit_tokens,
                                                       "prompt_cache_miss_tokens": prompt_cache_miss_tokens}))
                yield "--end--"
                log_info = logics.AddModelUsedAudit(
                    model_id=self.model_id, user_id=user_id, input_tokens=prompt_tokens,
                    output_tokens=completion_tokens)
                await add_llm_model_call_log(log_info)


async def encode(model_series, text, api_model="", api_key="", secret_key=""):
    return [], len(text) // 4  # 5002beta过度，5003删除该函数
    # import os
    # if model_series == "openai":
    #     cache_key = "9b5ad71b2ce5302211f9c61530b329a4922fc6a4"
    #     tiktoken_cache_dir = os.path.dirname(os.path.dirname(os.path.realpath(__file__))) + "/utils/tiktoken_cache"
    #     os.environ["TIKTOKEN_CACHE_DIR"] = tiktoken_cache_dir
    #
    #     assert os.path.exists(os.path.join(tiktoken_cache_dir, cache_key))
    #
    #     encoding = tiktoken.get_encoding("cl100k_base")
    #     tokens = encoding.encode(text)
    #     return tokens, len(tokens)
    # elif model_series in ["qwen", "internlm", "deepseek", "chatglm", "llama"]:
    #     tokenizer_dir = os.path.dirname(
    #         os.path.dirname(os.path.realpath(__file__))) + f"/utils/tokenizer/{model_series}"
    #     try:
    #         # 使用异步方式加载tokenizer
    #         tokenizer = await get_tokenizer_async(tokenizer_dir)
    #         tokens = tokenizer.encode(text)
    #         return tokens, len(tokens)
    #     except Exception as e:
    #         StandLogger.error(f"Tokenizer加载失败: {str(e)}")
    #         # 出错时使用近似估算
    #         return [], len(text) // 4
    # elif model_series == "tome":
    #     tokenizer_dir = os.path.dirname(
    #         os.path.dirname(os.path.realpath(__file__))) + f"/utils/tokenizer/qwen"
    #     try:
    #         # 使用异步方式加载tokenizer
    #         tokenizer = await get_tokenizer_async(tokenizer_dir)
    #         tokens = tokenizer.encode(text)
    #         return tokens, len(tokens)
    #     except Exception as e:
    #         StandLogger.error(f"Tokenizer加载失败: {str(e)}")
    #         # 出错时使用近似估算
    #         return [], len(text) // 4
    # elif model_series == "baidu":
    #     headers = {
    #         'Content-Type': 'application/json',
    #         'Accept': 'application/json'
    #     }
    #     async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
    #         baidu_url = f"https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id={api_key}&client_secret={secret_key}"
    #         async with session.post(baidu_url, headers=headers, ssl=False) as access_res:
    #             access_info = await access_res.json()
    #             access_token = access_info["access_token"]
    #     params = {
    #         "prompt": text
    #     }
    #     if api_model in ["ernie-4.0-8k", "ernie-3.5-8k", "ernie-speed-8k", "ernie-speed-128k", "ernie-lite-8k",
    #                      "ernie-tiny-8k", "ernie-char-8k"]:
    #         params["model"] = api_model
    #     async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
    #         url = f"https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/tokenizer/erniebot?access_token={access_token}"
    #         async with session.post(url, json=params, headers=headers, ssl=False) as original_res:
    #             original_info = await original_res.json()
    #             return [], original_info["usage"]["total_tokens"]
    # else:
    #     return [], len(text) // 4  # 使用更合理的近似值


async def decode(api_base, api_model, token_ids):
    url = api_base + "/tokenizer/decode"
    params = {
        "token_ids": token_ids,
        "model": api_model
    }
    async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
        async with session.post(url, data=json.dumps(params), ssl=False) as response:
            return response.json()


async def get_context_size(model_series, api_base, api_model):
    model_token_mapping = {
        "gpt-4": 8192,
        "gpt-4o": 128000,
        "gpt-4-0314": 8192,
        "gpt-4-32k": 32768,
        "gpt-4-32k-0314": 32768,
        "gpt-35-turbo-16k": 16384,
        "gpt-3.5-turbo": 4096,
        "gpt-3.5-turbo-0301": 4096,
        "text-ada-001": 2049,
        "ada": 2049,
        "text-babbage-001": 2040,
        "babbage": 2049,
        "text-curie-001": 2049,
        "curie": 2049,
        "davinci": 2049,
        "text-davinci-003": 4097,
        "text-davinci-002": 4097,
        "code-davinci-002": 8001,
        "code-davinci-001": 8001,
        "code-cushman-002": 2048,
        "code-cushman-001": 2048,
        "deepseek-chat": 131072,
        "deepseek-coder": 131072,
        "qianxun-l-8k": 8192,
        "qianxun-l-32k": 32768,
        "qianxun-l-128k": 131072,
        "qwen1.5-1.8b-chat": 32768
    }
    if model_series == "claude":
        return 200 * 1024
    if model_series == "openai" or model_series == "others":
        context_size = model_token_mapping.get(api_model)
        if context_size is None:
            context_size = 32768
        return context_size

    url = api_base + f"/{api_model}/config"
    response = requests.get(url)
    if response.status_code == 200:
        context_size = response.json()["max_tokens_length"]
        return context_size
    return 32768


set_res = {
    "res": {
        "time": 0.406968355178833,
        "token_len": 18,
        "data": "Sorry, I need more information to help you. Please provide more details about your question."
    }
}


class ModelConfig:
    def __init__(self):
        self.config_name = {}
        self.config_id = {}

    async def init_model_config(self):
        models = llm_model_dao.get_all_model_list()
        StandLogger.info_log(models)
        for model in models:
            self.config_name[model["f_model_name"]] = model
            self.config_id[model["f_model_id"]] = model

    async def get_model_config(self, key, value):
        # self.init_model_config()
        try:
            if key == "name":
                res = self.config_name.get(value)
                if res == None:
                    raise Exception("")
                return res
            elif key == "id":

                res = self.config_id.get(value)
                if res == None:
                    raise Exception("")
                return res
        except Exception:
            await self.init_model_config()
            if key == "name":
                return self.config_name.get(value, {})
            elif key == "id":
                return self.config_id.get(value, {})

    async def add_model_config(self, model_id, model_series, model_type, model_name, model, config):
        self.config_id[str(model_id)] = self.config_name[model_name] = \
            {"f_model_id": str(model_id), "f_model_series": model_series,
             "f_model_type": model_type, "f_model_name": model_name, "f_model": model, "f_model_config": config}

    async def add_model_context_size(self, model_id, model_name, context_size):
        try:
            self.config_id[str(model_id)]["context_size"] = context_size
            self.config_name[model_name]["context_size"] = context_size
        except Exception as e:
            self.config_id[str(model_id)] = {}
            self.config_name[model_name] = {}
            self.config_id[str(model_id)]["context_size"] = context_size
            self.config_name[model_name]["context_size"] = context_size

    async def model_config_rename(self, model_id, new_name):
        await self.init_model_config()
        old_name = self.config_id[model_id]["f_model_name"]
        self.config_name[new_name] = self.config_name[old_name]
        self.config_id[model_id]["f_model_name"] = new_name
        self.config_name.pop(old_name)

    @func_set_timeout(2)
    async def get_context_size_from_llm(self, llm):
        try:
            return llm.get_context_size()
        except Exception as e:
            return 4096


model_config = ModelConfig()
