# -*- coding:UTF-8 -*-

import requests
import allure

from urllib3 import disable_warnings
from urllib3.exceptions import InsecureRequestWarning
disable_warnings(InsecureRequestWarning)

class Request():
    def query(self, url, params, headers):
        '''封装get query接口'''
        allure.attach(url, name="Request URL")
        allure.attach(str(params), name="Request params")

        resp = requests.get(url, params=params, headers=headers, verify=False, allow_redirects=False)
        # print(resp.url)
        # print(resp.status_code, resp.text)
        # import pdb; pdb.set_trace();
        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        try:
            return [resp.status_code, resp.json()]
        except:
            # 如果响应不是有效的JSON，返回原始文本
            return [resp.status_code, resp.text]

    def get(self, url, headers):
        '''封装get接口'''
        allure.attach(url, name="Request URL")

        resp = requests.get(url, headers=headers, verify=False, allow_redirects=False)
        # print(url)
        # print(resp.text)
        # import pdb; pdb.set_trace();
        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        try:
            return [resp.status_code, resp.json()]
        except:
            # 如果响应不是有效的JSON，返回原始文本
            return [resp.status_code, resp.text]

    def post(self, url, data, headers):
        '''封装post接口'''
        allure.attach(url, name="Request URL")
        allure.attach(str(data), name="Request Body")
        # print(url)
        resp = requests.post(url, json=data, headers=headers, verify=False, allow_redirects=False)
        # print(resp.status_code, resp.text)

        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        if resp.text == "":
            return [resp.status_code, resp.text]
        else:
            try:
                return [resp.status_code, resp.json()]
            except:
                # 如果响应不是有效的JSON，返回原始文本
                return [resp.status_code, resp.text]

    def query_post(self, url, params, data, headers):
        '''封装post接口，带query参数'''
        allure.attach(url, name="Request URL")
        allure.attach(str(params), name="Request Params")
        allure.attach(str(data), name="Request Body")
        
        resp = requests.post(url, params=params, json=data, headers=headers, verify=False, allow_redirects=False)

        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        if resp.text == "":
            return [resp.status_code, resp.text]
        else:
            try:
                return [resp.status_code, resp.json()]
            except:
                # 如果响应不是有效的JSON，返回原始文本
                return [resp.status_code, resp.text]

    def post_multipart(self, url, files, data, headers, params=None):
        '''封装支持 Multipart 的 POST 接口'''
        allure.attach(url, name="Request URL")
        
        # 深度拷贝并清理 headers，防止 Content-Type 冲突
        request_headers = headers.copy()
        if "Content-Type" in request_headers:
            del request_headers["Content-Type"]
            
        if data:
            allure.attach(str(data), name="Form Fields")
        if params:
            allure.attach(str(params), name="Query Params")
            
        resp = requests.post(url, files=files, data=data, headers=request_headers, params=params, verify=False, allow_redirects=False)
        # print(resp.status_code, resp.text)
        
        if resp.status_code == 500:
            print(f"DEBUG: 500 Error Response Body: {resp.text}")

        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        if resp.text == "":
            return [resp.status_code, resp.text]
        else:
            try:
                return [resp.status_code, resp.json()]
            except:
                # 如果响应不是有效的JSON，返回原始文本
                return [resp.status_code, resp.text]

    def put(self, url, data, headers):
        '''封装put接口'''
        allure.attach(url, name="Request URL")
        allure.attach(str(data), name="Request Body")

        resp = requests.put(url, json=data, headers=headers, verify=False, allow_redirects=False)
        # print(url)
        # print(url, resp.status_code, resp.text)

        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        if resp.text == "":
            return [resp.status_code, resp.text]
        else:
            try:
                return [resp.status_code, resp.json()]
            except:
                # 如果响应不是有效的JSON，返回原始文本
                return [resp.status_code, resp.text]

    def delete(self, url, data, headers):
        '''封装delete接口'''
        allure.attach(url, name="Request URL")
        allure.attach(str(data), name="Request Body")

        resp = requests.delete(url, json=data, headers=headers, verify=False, allow_redirects=False)
        # print(resp.status_code,resp.text)

        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        if resp.text == "":
            return [resp.status_code, resp.text]
        else:
            try:
                return [resp.status_code, resp.json()]
            except:
                # 如果响应不是有效的JSON，返回原始文本
                return [resp.status_code, resp.text]

    def upload_file(self, url, files, data, headers):
        '''封装文件上传接口'''
        allure.attach(url, name="Request URL")
        allure.attach(str(data), name="Request Data")

        resp = requests.post(url, files=files, data=data, headers=headers, verify=False, allow_redirects=False)

        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        if resp.text == "":
            return [resp.status_code, resp.text]
        else:
            try:
                return [resp.status_code, resp.json()]
            except:
                # 如果响应不是有效的JSON，返回原始文本
                return [resp.status_code, resp.text]

    def post_with_timeout(self, url, data, headers, timeout):
        '''封装带超时的post接口'''
        allure.attach(url, name="Request URL")
        allure.attach(str(data), name="Request Body")
        allure.attach(f"Timeout: {timeout}s", name="Request Timeout")

        resp = requests.post(url, json=data, headers=headers, verify=False, allow_redirects=False, timeout=timeout)

        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        if resp.text == "":
            return [resp.status_code, resp.text]
        else:
            try:
                return [resp.status_code, resp.json()]
            except:
                # 如果响应不是有效的JSON，返回原始文本
                return [resp.status_code, resp.text]

    def post_with_retry(self, url, data, headers, timeout=60, max_retries=2, retry_status_codes=[500, 502, 503, 504]):
        '''
        封装带超时和重试的post接口
        :param url: 请求URL
        :param data: 请求数据
        :param headers: 请求头
        :param timeout: 超时时间（秒），默认60秒
        :param max_retries: 最大重试次数，默认2次
        :param retry_status_codes: 需要重试的状态码列表，默认[500, 502, 503, 504]
        :return: (status_code, response_data)
        '''
        import time
        
        allure.attach(url, name="Request URL")
        allure.attach(str(data), name="Request Body")
        allure.attach(f"Timeout: {timeout}s, Max Retries: {max_retries}", name="Request Config")
        
        last_exception = None
        last_status_code = None
        last_response = None
        
        for attempt in range(max_retries + 1):
            try:
                if attempt > 0:
                    # 重试前等待，使用指数退避
                    wait_time = min(2 ** attempt, 10)  # 最多等待10秒
                    print(f"重试第 {attempt} 次，等待 {wait_time} 秒后重试...")
                    time.sleep(wait_time)
                    allure.attach(f"Retry attempt {attempt}", name="Retry Info")
                
                resp = requests.post(url, json=data, headers=headers, verify=False, allow_redirects=False, timeout=timeout)
                
                allure.attach(str(resp.status_code), name="Response Code")
                allure.attach(resp.text, name="Response Result")
                
                # 如果状态码不在重试列表中，直接返回
                if resp.status_code not in retry_status_codes:
                    if resp.text == "":
                        return [resp.status_code, resp.text]
                    else:
                        try:
                            return [resp.status_code, resp.json()]
                        except:
                            return [resp.status_code, resp.text]
                
                # 如果状态码需要重试，记录并继续
                last_status_code = resp.status_code
                if resp.text == "":
                    last_response = resp.text
                else:
                    try:
                        last_response = resp.json()
                    except:
                        last_response = resp.text
                
                print(f"请求返回状态码 {resp.status_code}，需要重试（尝试 {attempt + 1}/{max_retries + 1}）")
                
            except requests.exceptions.Timeout:
                last_exception = f"Timeout after {timeout}s"
                print(f"请求超时（尝试 {attempt + 1}/{max_retries + 1}）")
                if attempt < max_retries:
                    continue
                else:
                    return [504, {"error": f"Request timeout after {timeout}s", "retries": max_retries + 1}]
            except requests.exceptions.RequestException as e:
                last_exception = str(e)
                print(f"请求异常: {e}（尝试 {attempt + 1}/{max_retries + 1}）")
                if attempt < max_retries:
                    continue
                else:
                    return [500, {"error": str(e), "retries": max_retries + 1}]
        
        # 所有重试都失败，返回最后一次的结果
        if last_status_code:
            return [last_status_code, last_response]
        else:
            return [500, {"error": last_exception or "Unknown error", "retries": max_retries + 1}]

    def get_with_retry(self, url, headers, timeout=60, max_retries=2, retry_status_codes=[500, 502, 503, 504]):
        '''
        封装带超时和重试的get接口
        :param url: 请求URL
        :param headers: 请求头
        :param timeout: 超时时间（秒），默认60秒
        :param max_retries: 最大重试次数，默认2次
        :param retry_status_codes: 需要重试的状态码列表，默认[500, 502, 503, 504]
        :return: (status_code, response_data)
        '''
        import time
        
        allure.attach(url, name="Request URL")
        allure.attach(f"Timeout: {timeout}s, Max Retries: {max_retries}", name="Request Config")
        
        last_exception = None
        last_status_code = None
        last_response = None
        
        for attempt in range(max_retries + 1):
            try:
                if attempt > 0:
                    # 重试前等待，使用指数退避
                    wait_time = min(2 ** attempt, 10)  # 最多等待10秒
                    print(f"重试第 {attempt} 次，等待 {wait_time} 秒后重试...")
                    time.sleep(wait_time)
                    allure.attach(f"Retry attempt {attempt}", name="Retry Info")
                
                resp = requests.get(url, headers=headers, verify=False, allow_redirects=False, timeout=timeout)
                
                allure.attach(str(resp.status_code), name="Response Code")
                allure.attach(resp.text, name="Response Result")
                
                # 如果状态码不在重试列表中，直接返回
                if resp.status_code not in retry_status_codes:
                    try:
                        return [resp.status_code, resp.json()]
                    except:
                        return [resp.status_code, resp.text]
                
                # 如果状态码需要重试，记录并继续
                last_status_code = resp.status_code
                try:
                    last_response = resp.json()
                except:
                    last_response = resp.text
                
                print(f"请求返回状态码 {resp.status_code}，需要重试（尝试 {attempt + 1}/{max_retries + 1}）")
                
            except requests.exceptions.Timeout:
                last_exception = f"Timeout after {timeout}s"
                print(f"请求超时（尝试 {attempt + 1}/{max_retries + 1}）")
                if attempt < max_retries:
                    continue
                else:
                    return [504, {"error": f"Request timeout after {timeout}s", "retries": max_retries + 1}]
            except requests.exceptions.RequestException as e:
                last_exception = str(e)
                print(f"请求异常: {e}（尝试 {attempt + 1}/{max_retries + 1}）")
                if attempt < max_retries:
                    continue
                else:
                    return [500, {"error": str(e), "retries": max_retries + 1}]
        
        # 所有重试都失败，返回最后一次的结果
        if last_status_code:
            return [last_status_code, last_response]
        else:
            return [500, {"error": last_exception or "Unknown error", "retries": max_retries + 1}]

    def pathdelete(self, url, headers):
        '''封装delete接口，path传参'''
        allure.attach(url, name="Request URL")

        resp = requests.delete(url, headers=headers, verify=False, allow_redirects=False)
        # print(url)
        # print(resp.text)

        allure.attach(str(resp.status_code), name="Response Code")
        allure.attach(resp.text, name="Response Result")

        if resp.text == "":
            return [resp.status_code, resp.text]
        else:
            try:
                return [resp.status_code, resp.json()]
            except:
                # 如果响应不是有效的JSON，返回原始文本
                return [resp.status_code, resp.text]
