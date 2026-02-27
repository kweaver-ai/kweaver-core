#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
@Time   : 2024/7/1 9:54
@Author : Leopold.yu
@File   : test_run.py
"""
import json
from json import JSONDecodeError

import allure
import jsonpath
import pytest
from jinja2 import Template
from jsonschema.validators import validate

from common.func import replace_params, genson
from conftest import config, case_list, BEARER_AUTH
from request.http_client import HTTPClient

resp_values = {}


@allure.title("{case_name}")
@pytest.mark.parametrize("feature, story, case_name, case_info",
                         [(x["feature"], x["story"], x["name"], x) for x in case_list])
def test_case(feature, story, case_name, case_info):
    print("run case: %s @ %s.%s" % (case_name, story, feature))
    allure.attach(
        body="run case: %s @ %s.%s" % (case_name, story, feature), name="用例名称"
    )

    allure.dynamic.feature(feature)
    allure.dynamic.story(story)

    if case_info["prev_case"]:
        with allure.step("执行前置用例执行"):
            for x in case_list:
                if x["name"] == case_info["prev_case"]:
                    test_case(x["feature"], x["story"], x["name"], x)
                    # 若存在同名用例，仅执行第一个匹配项
                    break

    with allure.step("加载用例参数"):
        # 更新前序用例提取的变量
        case_info = replace_params(case_info, **resp_values)

        # jinja2渲染变量
        case_info = {k: Template(v).render() for k, v in case_info.items()}

        # 替换path参数
        if case_info["path_params"] != '':
            case_path_params = json.loads(case_info["path_params"])
            case_info = replace_params(case_info, **case_path_params)

        # 参数格式转换
        case_headers = json.loads(case_info["headers"]) if case_info["headers"] != '' else {}
        if case_info.get("header_params") and case_info["header_params"] != '':
            try:
                case_headers.update(json.loads(case_info["header_params"]))
            except JSONDecodeError:
                pass
        case_headers["Authorization"] = BEARER_AUTH
        case_query_params = json.loads(case_info["query_params"]) if case_info["query_params"] != '' else {}
        case_body_params = json.loads(case_info["body_params"]) if case_info["body_params"] != '' else {}
        try:
            case_form_params = json.loads(case_info["form_params"]) if case_info["form_params"] != '' else {}
        except JSONDecodeError:
            # 忽略格式转换异常，适配fetch接口输入
            case_form_params = None
        case_cookie_params = json.loads(case_info["cookie_params"]) if case_info.get("cookie_params") and case_info["cookie_params"] != '' else None

        allure.attach(
            body="url: %s\nheaders: %s\nquery_params: %s\nbody_params: %s\nform_params: %s" % (
                case_info["url"], case_headers, case_query_params, case_body_params, case_form_params),
            name="请求参数"
        )

    with allure.step("发送请求"):
        client = HTTPClient(url="https://%s%s" % (config["env"]["host"], case_info["url"]),
                            method=case_info["method"], headers=case_headers)
        send_kw = dict(params=case_query_params, json=case_body_params, data=case_form_params)
        if case_cookie_params:
            send_kw["cookies"] = case_cookie_params
        client.send(**send_kw)

        allure.attach(
            body="url: %s\nhttp_code: %s\nresponse: %s" % (
                client.url, client.resp_code(), json.dumps(client.resp_body(), ensure_ascii=False)),
            name="请求响应"
        )

    # 结果断言
    if case_info["code_check"]:
        assert str(client.resp_code()) == case_info["code_check"]

    if case_info.get("resp_headers_check") and case_info["resp_headers_check"] != '':
        try:
            for k, v in json.loads(case_info["resp_headers_check"]).items():
                assert client.resp.headers.get(k) == v, "resp header %s: expect %s, got %s" % (k, v, client.resp.headers.get(k))
        except JSONDecodeError:
            pass

    if case_info["resp_check"]:
        for k, v in json.loads(case_info["resp_check"]).items():
            assert jsonpath.jsonpath(client.resp_body(), k)[0] == v

    if case_info["resp_schema"]:
        json_schema = genson(json.loads(case_info["resp_schema"]))
        validate(instance=client.resp_body(), schema=json_schema)

    # 提取响应中的变量
    if case_info["resp_values"]:
        for k, v in json.loads(case_info["resp_values"]).items():
            param = jsonpath.jsonpath(client.resp_body(), v)[0]
            if isinstance(param, list) or isinstance(param, dict):
                # 将对象存储为字符串,加载用例参数时再转换为JSON格式
                resp_values[k] = json.dumps(param, ensure_ascii=False)
            else:
                resp_values[k] = param


if __name__ == '__main__':
    pass
