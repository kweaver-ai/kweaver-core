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
        return self.resp.json()


if __name__ == '__main__':
    pass
