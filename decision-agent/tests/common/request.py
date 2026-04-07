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

    def post_multipart(self, url, files, data, headers):
        '''封装post接口'''
        allure.attach(url, name="Request URL")
        allure.attach(str(data), name="Request Body")
        # print(url)
        resp = requests.post(url, files=files, data=data, headers=headers, verify=False, allow_redirects=False)
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
