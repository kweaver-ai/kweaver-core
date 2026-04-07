import json

import aiohttp

from app.commons.errors import *
import concurrent.futures

from app.core.config import base_config
from app.logs.stand_log import StandLogger


class BaiduTianchenClient:
    def __init__(self, url, ClientId, OperationCode):
        self.url = url
        self.ClientId = ClientId
        self.OperationCode = OperationCode

    async def embedding(self, texts):
        params = {
            "batch_text": texts
        }
        headers = {
            "ClientId": self.ClientId,
            "OperationCode": self.OperationCode
        }
        conn = aiohttp.TCPConnector(verify_ssl=False)
        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(self.url, json=params, headers=headers) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status != 200 or "error_msg" in result.keys():
                    raise Exception(result.get("error_msg", res))
        return result["result"]["batch_embedding"]

    async def reranker(self, query, documents):
        if documents == []:
            return []
        params = {
            "query": query,
            "documents": documents
        }
        headers = {
            "ClientId": self.ClientId,
            "OperationCode": self.OperationCode
        }
        conn = aiohttp.TCPConnector(verify_ssl=False)
        res_list = []
        original_res_list = []
        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(self.url, json=params, headers=headers) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status != 200 or "error_msg" in result.keys():
                    raise Exception(result.get("error_msg", res))
                original_res_list = sorted(result["results"], key=lambda x: x["index"])
        for item in original_res_list:
            res_list.append(item["relevance_score"])
        return res_list


class BaiduClient:
    def __init__(self, url, api_key, secret_key):
        self.url = url
        self.api_key = api_key
        self.secret_key = secret_key
        self.access_token = ""

    async def get_access_token(self):
        headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        }
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
                    if "errors" in tmp_map.keys() and "message" in tmp_map["errors"].keys():
                        error_dict["detail"] = tmp_map["errors"]["message"]
                    return error_dict
                access_token = result["access_token"]
        return access_token

    async def embedding_thread(self, texts_slice, index):
        params = {
            "input": texts_slice
        }

        conn = aiohttp.TCPConnector(verify_ssl=False)
        res_list = []

        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(
                    self.url + f"?access_token={self.access_token}",
                    json=params) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status != 200 or "error_msg" in result.keys():
                    raise Exception(result.get("error_msg", res))
                for item in result["data"]:
                    res_list.append(item["embedding"])
        return {index: res_list}

    async def embedding(self, texts):
        self.access_token = await self.get_access_token()

        for i in range(0, len(texts)):
            if texts[i] == "":
                texts[i] = " "
            if len(texts[i]) > 380:
                texts[i] = texts[i][:380]
        slice_list = [texts[i:i + 16] for i in range(0, len(texts), 16)]
        index = 0
        request_dict = {}
        with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
            futures = []
            for texts_slice in slice_list:
                request_dict[index] = texts_slice  # 存储顺序信息，便于与返回内容匹配
                future = executor.submit(self.embedding_thread, texts_slice, index)
                index += 1
                futures.append(future)
        original_result_list = [await future.result() for future in futures]
        original_result_dict = {}
        for item in original_result_list:
            original_result_dict = original_result_dict | item
        res_list = []

        for i in range(0, index):
            for item in original_result_dict[i]:
                res_list.append(item)
        return res_list

    async def reranker_thread(self, query, documents, index):
        params = {
            "query": query,
            "documents": documents
        }
        conn = aiohttp.TCPConnector(verify_ssl=False)
        res_list = []
        original_res_list = []
        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(
                    self.url + f"?access_token={self.access_token}",
                    json=params) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status != 200 or "error_msg" in result.keys():
                    raise Exception(result.get("error_msg", res))
                original_res_list = sorted(result["results"], key=lambda x: x["index"])
        for item in original_res_list:
            res_list.append(item["relevance_score"])
        return {index: res_list}

    async def reranker(self, query, documents):
        self.access_token = await self.get_access_token()
        if len(query) > 1590:
            query = query[:1590]
        for i in range(0, len(documents)):
            if documents[i] == "":
                documents[i] = " "
            if len(documents[i]) > 4000:
                documents[i] = documents[i][:4000]
        slice_list = [documents[i:i + 64] for i in range(0, len(documents), 64)]
        index = 0
        request_dict = {}
        with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
            futures = []
            for texts_slice in slice_list:
                request_dict[index] = texts_slice  # 存储顺序信息，便于与返回内容匹配
                future = executor.submit(self.reranker_thread, query, texts_slice, index)
                index += 1
                futures.append(future)
        original_result_list = [await future.result() for future in futures]
        original_result_dict = {}
        for item in original_result_list:
            original_result_dict = original_result_dict | item
        res_list = []
        for i in range(0, index):
            for item in original_result_dict[i]:
                res_list.append(item)
        return res_list


class BaishengClient:
    def __init__(self, url, api_key, model):
        self.url = url
        self.api_key = api_key
        self.model = model

    async def embedding(self, texts):
        params = {
            "model": self.model,
            "input": texts
        }
        headers = {
            "Authorization": f"Bearer {self.api_key}"
        }
        conn = aiohttp.TCPConnector(verify_ssl=False)
        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(
                    self.url,
                    json=params, headers=headers) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status != 200 or "error_msg" in result.keys():
                    raise Exception(result.get("error_msg", res))
        res_list = []
        for item in result["data"]:
            res_list.append(item["embedding"])
        return res_list

    async def reranker(self, query, documents):
        params = {
            "model": self.model,
            "query": query,
            "sentences": documents
        }
        headers = {
            "Authorization": f"Bearer {self.api_key}"
        }
        conn = aiohttp.TCPConnector(verify_ssl=False)
        async with aiohttp.ClientSession(connector=conn, timeout=base_config.aiohttp_timeout) as session:
            async with session.post(
                    self.url,
                    json=params, headers=headers) as resp:
                res = await resp.text()
                result = json.loads(res)
                if resp.status != 200 or "error_msg" in result.keys():
                    raise Exception(result.get("error_msg", res))
        return result["scores"]


class InnerClient:
    def __init__(self, url, model_name, api_key="", adapter=False, adapter_code=None):
        if not url.startswith(('http://', 'https://')):
            url = f"http://{url}"
        self.url = url
        self.model_name = model_name
        self.api_key = api_key
        self.headers = {
            "Authorization": f"Bearer {self.api_key}"
        }
        self.adapter = adapter
        self.adapter_code = adapter_code

    async def embedding(self, texts):
        if self.adapter and self.adapter_code:
            try:
                global_namespace = {'__builtins__': __builtins__}
                exec(self.adapter_code, global_namespace, global_namespace)
                adapter_func = global_namespace.get('main')
                if not adapter_func or not callable(adapter_func):
                    raise ValueError("Adapter code must define an async function named 'main'")

                result = await adapter_func(texts)
                return result
            except Exception as e:
                raise Exception(f"Adapter execution failed: {str(e)}")

        # 原有逻辑
        params = {
            "model": self.model_name,
            "input": texts
        }
        async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
            async with session.post(
                    self.url,
                    json=params, headers=self.headers, ssl=False) as resp:
                if resp.status != 200:
                    error_msg = await resp.text()
                    StandLogger.error(
                        f"call embeddingError,model_name={self.model_name},error_detail={error_msg},body={texts},status={resp.status}")
                    raise Exception(error_msg)
                res = await resp.text()
                result = json.loads(res)
        return result

    async def reranker(self, query, documents):
        if self.adapter and self.adapter_code:
            try:
                global_namespace = {'__builtins__': __builtins__}
                exec(self.adapter_code, global_namespace, global_namespace)
                reranker_func = global_namespace.get('main')
                if not reranker_func or not callable(reranker_func):
                    raise ValueError("Adapter code must define an async function named 'my_reranker'")
                result = await reranker_func(query, documents)
                return result
            except Exception as e:
                raise Exception(f"Adapter execution failed: {str(e)}")

        # 原有逻辑
        params = {
            "model": self.model_name,
            "query": query,
            "documents": documents
        }
        async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
            async with session.post(
                    self.url,
                    json=params, headers=self.headers, ssl=False) as resp:
                if resp.status != 200:
                    error_msg = await resp.text()
                    StandLogger.error(
                        f"call reranker error,model_name={self.model_name},error_detail={error_msg},query={query}，documents={documents},status={resp.status}")
                    raise Exception(error_msg)
                res = await resp.text()
                result = json.loads(res)
        return result

    async def test_embedding(self, texts):
        if self.adapter and self.adapter_code:
            try:
                global_namespace = {'__builtins__': __builtins__}
                exec(self.adapter_code, global_namespace, global_namespace)
                adapter_func = global_namespace.get('main')
                if not adapter_func or not callable(adapter_func):
                    raise ValueError("Adapter code must define an async function named 'main'")

                result = await adapter_func(texts)

                return result
            except Exception as e:
                raise Exception(f"Adapter execution failed: {str(e)}")

        else:
            params = {
                "model": self.model_name,
                "input": texts
            }
            async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                async with session.post(
                        self.url,
                        json=params, headers=self.headers, ssl=False) as resp:
                    if resp.status == 422:
                        raise Exception("string_too_long,String should have at most 122880 characters")
                    if resp.status != 200:
                        raise Exception(resp.content)
                    res = await resp.text()
                    result = json.loads(res)
        required_keys = ["object", "data", "model", "usage"]
        if not all(key in result for key in required_keys):
            raise ValueError(f"Invalid adapter response format, missing one of: {required_keys}")

        if not isinstance(result["data"], list) or not all(
                isinstance(item, dict) and "embedding" in item and
                isinstance(item["embedding"], list) and len(item["embedding"]) > 0
                for item in result["data"]
        ):
            raise ValueError("Invalid data format in adapter response")
        return result

    async def test_reranker(self, query, documents):
        if self.adapter and self.adapter_code:
            try:
                global_namespace = {'__builtins__': __builtins__}
                exec(self.adapter_code, global_namespace, global_namespace)
                reranker_func = global_namespace.get('main')
                if not reranker_func or not callable(reranker_func):
                    raise ValueError("Adapter code must define an async function named 'my_reranker'")
                result = await reranker_func(query, documents)
            except Exception as e:
                raise Exception(f"Adapter execution failed: {str(e)}")

        else:
            params = {
                "model": self.model_name,
                "query": query,
                "documents": documents
            }
            async with aiohttp.ClientSession(timeout=base_config.aiohttp_timeout) as session:
                async with session.post(
                        self.url,
                        json=params, headers=self.headers, ssl=False) as resp:
                    if resp.status != 200:
                        error_msg = await resp.text()
                        StandLogger.error(
                            f"call reranker error,model_name={self.model_name},error_detail={error_msg},query={query}，documents={documents},status={resp.status}")
                        raise Exception(error_msg)
                    res = await resp.text()
                    result = json.loads(res)
        required_keys = ["object", "results", "model", "usage"]
        if not all(key in result for key in required_keys):
            raise ValueError(f"Invalid adapter response format, missing one of: {required_keys}")

        if not isinstance(result["results"], list) or len(result["results"]) == 0 or not all(
                isinstance(item, dict) and "relevance_score" in item
                for item in result["results"]
        ):
            raise ValueError("Invalid results format in adapter response")
        return result
