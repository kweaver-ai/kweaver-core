#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
@Time   : 2024/7/10 17:19
@Author : Leopold.yu
@File   : http_client.py
"""
import requests
import urllib3
import json

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

        self._safe_print("request: %s" % self._to_log_text(kwargs))
        self._safe_print("code: %s" % self.resp_code())
        self._safe_print("response: %s" % self._to_log_text(self.resp_body()))

    @staticmethod
    def _to_log_text(value):
        """将任意对象序列化为稳定文本，避免 repr 触发编码问题。"""
        try:
            return json.dumps(value, ensure_ascii=False, default=str)
        except Exception:
            return str(value)

    @staticmethod
    def _safe_print(text):
        """
        兼容 Windows(GBK/CP936) 与 Linux(UTF-8) 控制台输出：
        当 stdout 编码不支持某些字符时，自动 replace，避免中断测试流程。
        """
        try:
            print(text)
            return
        except UnicodeEncodeError:
            pass

        # 兜底路径：按当前 stdout 编码进行降级替换后再输出
        try:
            import sys
            enc = (getattr(sys.stdout, "encoding", None) or "utf-8")
            safe_text = text.encode(enc, errors="replace").decode(enc, errors="replace")
            print(safe_text)
        except Exception:
            # 最终兜底，保证任何环境都不因日志打印中断
            print(str(text).encode("ascii", errors="replace").decode("ascii"))

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
