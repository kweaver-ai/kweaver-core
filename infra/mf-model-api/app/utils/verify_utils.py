import json

import aiohttp
from fastapi.responses import JSONResponse
from func_timeout import func_set_timeout
from llmadapter.llms.llm_factory import llm_factory
from llmadapter.schema import AIMessage
from urllib3.exceptions import MaxRetryError
from app.commons.errors import ModelFactory_ModelController_TestModel_Error_Error, LLMTestError
from app.logs.stand_log import StandLogger


@func_set_timeout(30)
async def llm_test(series, config, llm_id, user_id, model_type):
    message = [AIMessage(content="你是")]
    content = "测试连接失败，请重新检查信息"
    prompt = "你是"
    # 区分openai和其他模型
    if series == 'openai':
        try:
            try:
                if "api_key" not in config.keys():
                    LLMTestError['description'] = "api_key 参数缺失"
                    LLMTestError['detail'] = "openai大模型需要 api_key"
                    return JSONResponse(status_code=500, content=LLMTestError)
                llm = llm_factory.create_llm(llm_type="openai",
                                             api_type="azure",
                                             api_version="2023-03-15-preview",
                                             openai_api_base=config['api_url'],
                                             openai_api_key=config['api_key'],
                                             engine=config['api_model'],
                                             temperature=0.2,
                                             max_tokens=400)
                # if llm_id != "":
                #     log_info = logics.AddModelUsedAudit(
                #         model_id=llm_id, user_id=user_id, input_tokens=5,
                #         output_tokens=10)
                #     await add_llm_model_call_log(log_info)
                return JSONResponse(status_code=200, content={"res": {"status": True, "model_type": "chat"}})

            except Exception as e:
                print(e)
                # if llm_id != "":
                #     log_info = logics.AddModelUsedAudit(
                #         model_id=llm_id, user_id=user_id, input_tokens=5,
                #         output_tokens=10)
                #     await add_llm_model_call_log(log_info)
                return JSONResponse(status_code=200, content={"res": {"status": True, "model_type": "chat"}})
        except Exception as e:
            print(e)
            if isinstance(e.args[0], MaxRetryError):
                content = "无法访问该链接，请检查该链接是否可以访问"
            error_dict = ModelFactory_ModelController_TestModel_Error_Error.copy()
            error_dict["detail"] = str(e.args[0])
            error_dict["description"] = error_dict["solution"] = content
            if not isinstance(e.args[0], MaxRetryError):
                error_dict["description"] = "模型配置错误，请检查模型信息"
            # if error_dict["detail"].strip(" ") != "":
            #     error_dict["description"] = error_dict["detail"]
            # if len(error_dict["description"]) > 500:
            #     error_dict["description"] = error_dict["description"][0:500]
            return JSONResponse(status_code=400, content=error_dict)

    elif series.lower() == "claude":
        try:
            params = {
                "messages": [
                    {
                        "content": "你好",
                        "role": "user"
                    }
                ],
                "model": config["api_model"],
                "stream": False,
                "max_tokens": 1000
            }
            headers = {
                "x-api-key": f"{config['api_key']}",
                "anthropic-version": "2023-06-01",
                "content-type": "application/json"
            }
            async with aiohttp.ClientSession() as session:
                async with session.post(config["api_url"], json=params, headers=headers, ssl=False) as response:
                    response.encoding = 'utf-8'
                    if response.status != 200:
                        error_dict = ModelFactory_ModelController_TestModel_Error_Error.copy()
                        error_dict["detail"] = await response.text()
                        return JSONResponse(status_code=400, content=error_dict)
            content = {"status": "ok", "id": llm_id}
            return JSONResponse(status_code=200, content=content)
        except Exception as e:
            print(e)
            content = "测试连接失败，请重新检查信息"
            if isinstance(e.args[0], MaxRetryError):
                content = "无法访问该链接，请检查该链接是否可以访问"
            error_dict = ModelFactory_ModelController_TestModel_Error_Error.copy()
            error_dict["detail"] = str(e.args[0])
            error_dict["description"] = error_dict["solution"] = content
            if not isinstance(e.args[0], MaxRetryError):
                error_dict["description"] = "模型配置错误，请检查模型信息"
            return JSONResponse(status_code=400, content=error_dict)
    elif series.lower() == "baidu":
        headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        }
        url = f"https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id={config.get('api_key', '')}&client_secret={config.get('secret_key', '')}"
        async with aiohttp.ClientSession() as session:
            async with session.post(url, headers=headers, ssl=False) as response:
                response.encoding = 'utf-8'
                if response.status != 200:
                    error_dict = ModelFactory_ModelController_TestModel_Error_Error.copy()
                    error_dict["detail"] = await response.text()
                    return JSONResponse(status_code=400, content=error_dict)
                access_res = await response.json()
                access_token = access_res["access_token"]
        params = {
            "messages": [
                {
                    "content": "你好",
                    "role": "user"
                }
            ]
        }
        async with aiohttp.ClientSession() as session:
            url = config["api_url"] + f"?access_token={access_token}"
            async with session.post(url, json=params, headers=headers, ssl=False) as response:
                response.encoding = 'utf-8'
                if response.status != 200:
                    error_dict = ModelFactory_ModelController_TestModel_Error_Error.copy()
                    error_dict["detail"] = await response.text()
                    return JSONResponse(status_code=400, content=error_dict)
        content = {"status": "ok", "id": llm_id}
        return JSONResponse(status_code=200, content=content)
    elif series.lower() == "baidu_tianchen":
        params = {
            "messages": [
                {
                    "role": "user",
                    "content": "你好"
                }
            ]
        }
        async with aiohttp.ClientSession() as session:
            url = config["api_url"] + f"?api_name="
            async with session.post(url, json=params,  ssl=False) as response:
                response.encoding = 'utf-8'
                if response.status != 200:
                    error_dict = ModelFactory_ModelController_TestModel_Error_Error.copy()
                    error_dict["detail"] = await response.text()
                    return JSONResponse(status_code=400, content=error_dict)
        content = {"status": "ok", "id": llm_id}
        return JSONResponse(status_code=200, content=content)
    else:
        try:
            params = {
                "messages": [
                    {
                        "content": "你好",
                        "role": "user"
                    }
                ],
                "model": config["api_model"],
                "stream": True,
                "stream_options": {"include_usage": True}
            }
            headers = {
                "Authorization": f"Bearer {config.get('api_key', '')}",
                "Content-Type": "application/json"
            }
            token_len = 0
            prompt_tokens = 0
            completion_tokens = 0
            async with aiohttp.ClientSession() as session:
                async with session.post(config["api_url"], json=params, headers=headers, ssl=False) as response:
                    response.encoding = 'utf-8'
                    if response.status != 200:
                        error_dict = ModelFactory_ModelController_TestModel_Error_Error.copy()
                        error_dict["description"] = error_dict["solution"] = content
                        try:
                            error_dict["detail"] = await response.text()
                        except Exception as e:
                            StandLogger.error(str(e))
                        error_dict["description"] = "模型配置错误，请检查模型信息"
                        return JSONResponse(status_code=400, content=error_dict)
                    async for chunk in response.content:
                        chunk = chunk.decode('utf-8')
                        if chunk.endswith('\n'):
                            chunk = chunk[:-1]
                        elif chunk.endswith('\r\n'):
                            chunk = chunk[:-2]
                        if len(chunk) >= 6 and chunk[0:6] != "data: ":
                            continue
                        if chunk != "data: [DONE]" and chunk != "":
                            if chunk[0:6] == "data: ":
                                chunk = chunk[6:]
                            try:
                                datas = json.loads(chunk)
                            except Exception:
                                continue
                            if "usage" in datas.keys():
                                try:
                                    prompt_tokens = datas["usage"]["prompt_tokens"]
                                    completion_tokens = datas["usage"]["completion_tokens"]
                                except Exception:
                                    pass
                    # if llm_id != "":
                    #     log_info = logics.AddModelUsedAudit(
                    #         model_id=llm_id, user_id=user_id,
                    #         input_tokens=prompt_tokens,
                    #         output_tokens=completion_tokens)
                    #     await add_llm_model_call_log(log_info)
                    content = {"status": "ok", "id": llm_id}
                    return JSONResponse(status_code=200, content=content)
        except Exception as e:
            StandLogger.error(str(e))
            content = "测试连接失败，请重新检查信息"
            if isinstance(e.args[0], MaxRetryError):
                content = "无法访问该链接，请检查该链接是否可以访问"
            error_dict = ModelFactory_ModelController_TestModel_Error_Error.copy()
            error_dict["detail"] = str(e.args[0])
            error_dict["description"] = error_dict["solution"] = content
            if not isinstance(e.args[0], MaxRetryError):
                error_dict["description"] = "模型配置错误，请检查模型信息"
            return JSONResponse(status_code=400, content=error_dict)
