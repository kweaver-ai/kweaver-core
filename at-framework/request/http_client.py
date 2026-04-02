#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
@Time   : 2024/7/10 17:19
@Author : Leopold.yu
@File   : http_client.py
"""
import requests
import urllib3

urllib3.disable_warnings(urllib3.connectionpool.InsecureRequestWarning)


class HTTPClient:
    def __init__(self, url, method, headers):
        self.session = requests.session()

        # 请求参数
        self.url = url
        self.method = method
        self.headers = headers

        # 响应对象
        self.resp = None

    def send(self, **kwargs):
        kwargs["url"] = self.url
        kwargs["method"] = self.method
        kwargs["headers"] = self.headers
        self.resp = self.session.request(verify=False, allow_redirects=True, **kwargs)

        print("request: %s" % kwargs)
        print("code: %s" % self.resp_code())
        print("response: %s" % self.resp_body())

    def resp_code(self):
        return self.resp.status_code

    def resp_body(self):
        """解析 JSON；204/空 body 或非 JSON（如 HTML 错误页）时返回 dict，避免 json() 抛错掩盖状态码。"""
        if self.resp is None:
            return {}
        if self.resp.status_code == 204:
            return {}
        try:
            return self.resp.json()
        except ValueError:
            t = (self.resp.text or "").strip()
            return {"_parse_error": "non_json_body", "_text_preview": t[:2000]} if t else {}


if __name__ == '__main__':
    pass
