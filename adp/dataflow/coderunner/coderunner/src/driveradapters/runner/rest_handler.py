from tornado.web import RequestHandler
from tornado.web import HTTPError
import tornado.ioloop
import tornado.web
import json
import logging
import sys
import resource
import requests
from requests.adapters import HTTPAdapter
from RestrictedPython import safe_builtins
from RestrictedPython import limited_builtins
from RestrictedPython import utility_builtins
import builtins
from requests.exceptions import RequestException
from concurrent.futures import ThreadPoolExecutor
import asyncio
import os
from urllib3.util.retry import Retry

from logics.check_module import CheckModule

session = requests.Session()
retry = Retry(connect=10, backoff_factor=3)
adapter = HTTPAdapter(max_retries=retry, pool_connections=100, pool_maxsize=100)
session.mount('http://', adapter)
session.mount('https://', adapter)
workers = os.getenv("PYTHON_MAX_WORKERS", "300")
executor = ThreadPoolExecutor(max_workers=int(workers))
executor2 = ThreadPoolExecutor(max_workers=int(workers))
check_module = CheckModule()

class RestrictedImport:
    not_allowed_packages = ['os', 'subprocess', 'pickle', 'sys']

    def __call__(self, name, globals=None, locals=None, fromlist=(), level=0):
        if name not in self.not_allowed_packages:
            return builtins.__import__(name, globals, locals, fromlist, level)
        else:
            raise Exception(f"Importing package '{name}' is not allowed")

__all__ = [
    'RunnerHandler',
]

# 设置资源限制的函数
def set_resource_limits():
    # 设置内存限制
    resource.setrlimit(resource.RLIMIT_AS, (80 * 1024 * 1024, 80 * 1024 * 1024))

class RunnerHandler(RequestHandler):

    def set_default_headers(self):
        self.set_header("Content-Type", "application/json")

    async def post(self):
        # set_resource_limits()

        # 执行请求处理的代码
        # await self.process_request()
        result = await tornado.ioloop.IOLoop.current().run_in_executor(executor2, self.process_request)
        status = result[1]
        try:
            json.dumps(result[0])
            res = result[0]
        except Exception as e:
            res = {"detail": str(result[0]), "code": 500}
            status = 500
        self.set_status(status)
        self.write(res)
        
    def process_request(self):
        try:
            headers = self.request.headers
            host, access_token = str(headers.get('x-as-address')), str(headers.get('x-authorization'))
            if access_token.startswith("Bearer "):
                access_token = access_token[len("Bearer "):]
            data = json.loads(self.request.body)
            input_params = data["input_params"]
            output_params = data["output_params"]
            user_code = data["code"]
            funcName = data.get("func", "main")

            if "-private" in user_code or "hydra-admin" in user_code:
                raise RequestException(f"there may be security vulnerabilities in the code")

            args = []

            for param in input_params:
                _type = param["type"]
                _value = str(param["value"])
                if _type == "string":
                    args.append(_value)
                elif _type == "int":
                    value = int(float(_value))
                    args.append(value)
                elif _type == "array" or _type == "object":
                    try:
                        value = json.loads(_value or 'null')
                        args.append(value)
                    except Exception as e:
                        args.append(_value)
                        
            results = []

            def exec_code():
                try:
                    sys.stdout = None
                    check_module.check_and_install_modules(user_code)
                    # 将代码字符串执行到模块中
                    code = """
try:
    from aishu_anyshare_api import ApiClient
    ApiClient.set_global_host("{}")
    ApiClient.set_global_access_token("{}")
except ImportError as e:
    pass
{}
                        """.format(host, access_token, user_code)

                    restricted_import = RestrictedImport()

                    ALLOWED_BUILTINS = {}
                    INNER_BUILDINS = {
                        '__import__': restricted_import
                    }
                    # ALLOWED_BUILTINS.update(builtins.__dict__)
                    ALLOWED_BUILTINS.update(safe_builtins)
                    ALLOWED_BUILTINS.update(limited_builtins)
                    ALLOWED_BUILTINS.update(utility_builtins)
                    ALLOWED_BUILTINS.update(INNER_BUILDINS)
                    restricted_globals = dict(__builtins__=ALLOWED_BUILTINS)
                    restricted_globals['__name__'] = 'RunnerHandler'
                    exec(code, restricted_globals)
                    # 获取模块的所有属性
                    r = restricted_globals[funcName](*args)
                    results.append(r)
                except TimeoutError as e:
                    raise e
                except MemoryError as e:
                    raise e
                except Exception  as e:
                    raise e

            exec_code()

            result = results[0]
            res = {}
            for index in range(len(output_params)):
                if type(result) is tuple:
                    res[output_params[index]["key"]] = result[index]
                else:
                    res[output_params[0]["key"]] = result
                    break

            return res, 200
        except Exception as e:
            logging.warning(e)
            # self.set_status(500)
            # self.write(json.dumps({"detail": str(e), "code": 500}))
            return json.dumps({"detail": str(e), "code": 500}), 500
        
class AsyncRunnerHandler(RequestHandler):

    def set_default_headers(self):
        self.set_header("Content-Type", "application/json")

    async def post(self):
        self.set_status(202)  # 使用HTTP 202 Accepted状态码
        self.write({"status": "Task started"})
        self.finish()  # 结束当前请求处理

        await self.async_task()

    async def async_task(self):
        # 使用线程执行器来异步执行exec
        loop = asyncio.get_running_loop()
        await loop.run_in_executor(executor, self.process_request)
            

    def process_request(self):
        callback = ""
        try:
            headers = self.request.headers
            host, access_token = str(headers.get('x-as-address')), str(headers.get('x-authorization'))
            if access_token.startswith("Bearer "):
                access_token = access_token[len("Bearer "):]
            data = json.loads(self.request.body)
            input_params = data["input_params"]
            output_params = data["output_params"]
            user_code = data["code"]
            funcName = data.get("func", "main")
            callback = data["callback"]

            if "-private" in user_code:
                raise RequestException(f"there may be security vulnerabilities in the code")

            args = []

            for param in input_params:
                _type = param["type"]
                _value = str(param["value"])
                if _type == "string":
                    args.append(_value)
                elif _type == "int":
                    value = int(_value)
                    args.append(value)
                elif _type == "array" or _type == "object":
                    try:
                        value = json.loads(_value or 'null')
                        args.append(value)
                    except Exception as e:
                        args.append(_value)

            results = []

            def exec_code():
                try:
                    sys.stdout = None
                    check_module.check_and_install_modules(user_code)
                    # 将代码字符串执行到模块中
                    code = """
try:
    from aishu_anyshare_api import ApiClient
    ApiClient.set_global_host("{}")
    ApiClient.set_global_access_token("{}")
except ImportError as e:
    pass
{}
                        """.format(host, access_token, user_code)

                    restricted_import = RestrictedImport()

                    ALLOWED_BUILTINS = {}
                    INNER_BUILDINS = {
                        '__import__': restricted_import
                    }
                    # ALLOWED_BUILTINS.update(builtins.__dict__)
                    ALLOWED_BUILTINS.update(safe_builtins)
                    ALLOWED_BUILTINS.update(limited_builtins)
                    ALLOWED_BUILTINS.update(utility_builtins)
                    ALLOWED_BUILTINS.update(INNER_BUILDINS)
                    restricted_globals = dict(__builtins__=ALLOWED_BUILTINS)
                    restricted_globals['__name__'] = 'RunnerHandler'
                    exec(code, restricted_globals)
                    # 获取模块的所有属性
                    r = restricted_globals[funcName](*args)
                    results.append(r)
                except TimeoutError as e:
                    raise e
                except MemoryError as e:
                    raise e
                except Exception  as e:
                    raise e

            exec_code()
            result = results[0]
            res = {}
            for index in range(len(output_params)):
                if type(result) is tuple:
                    res[output_params[index]["key"]] = result[index]
                else:
                    res[output_params[0]["key"]] = result
                    break
            try:
                session.post(callback, json={"res": res, "type": "pythoncode"}, verify=False, timeout=10)
            except Exception as e:
                res = {"detail": str(res), "code": 500}
                session.post(callback, json={"res": res, "type": "pythoncode", "code": 500}, verify=False, timeout=10)
                logging.warning(f"send exec python result failed, callback: {callback}, detail: {e}")
        except Exception as e:
            logging.warning(f"exec python code failed, detail: {e}")
            # self.set_status(500)
            # self.write(json.dumps({"detail": str(e), "code": 500}))
            res = {"detail": str(e), "code": 500}
            logging.info(f"callback url: {callback}")
            try:
                session.post(callback, json={'res': res, 'type': 'pythoncode', 'code': 500}, timeout=100)
                return json.dumps({"detail": str(e), "code": 500}), 500
            except Exception as e:
                logging.warning(f"exec callback failed, callback: {callback}, detail: {e}")
