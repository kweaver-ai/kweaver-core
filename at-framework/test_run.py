#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
@Time   : 2024/7/1 9:54
@Author : Leopold.yu
@File   : test_run.py
"""
import json
import random
import re
import string as _str
from datetime import datetime, timedelta
from json import JSONDecodeError

import allure
import jsonpath
import pytest
from jinja2 import Environment
from jsonschema.validators import validate

from common import at_env
from common.func import replace_params, replace_placeholders, genson
from conftest import config, compute_case_list, BEARER_AUTH
from request.http_client import HTTPClient

resp_values = {}


def _parse_check_expression(expected_value, actual_value, jsonpath_key=""):
    """
    解析并执行 resp_check 表达式验证

    支持的表达式格式：
    - 精确匹配: "value" (字符串), 123 (数字), true/false (布尔), null
    - 比较操作: ">0", "<10", ">=5", "<=100", "!=null"
    - contains: "contains('skill')" - 检查数组中是否包含元素，或字符串是否包含子串
    - exists: "exists" - 检查字段是否存在（非空）
    - regex: "regex('^abc.*')" - 正则匹配
    - type: "type('array')" - 类型检查

    返回: (是否通过, 错误消息)
    """
    if expected_value is None:
        # 期望值为 null，检查实际值是否为 None
        return actual_value is None, "expect null, got %s" % repr(actual_value)

    if isinstance(expected_value, (int, float, bool)):
        # 数字或布尔值：精确匹配
        return actual_value == expected_value, "expect %s, got %s" % (expected_value, repr(actual_value))

    if not isinstance(expected_value, str):
        # 其他类型（如直接传入的 dict/list）：精确匹配
        return actual_value == expected_value, "expect %s, got %s" % (repr(expected_value), repr(actual_value))

    # 字符串表达式解析
    expr = expected_value.strip()

    # 1. contains 操作
    contains_match = re.match(r"^contains\((.+)\)$", expr)
    if contains_match:
        target = contains_match.group(1).strip()
        # 移除可能的引号
        if target.startswith("'") and target.endswith("'"):
            target = target[1:-1]
        elif target.startswith('"') and target.endswith('"'):
            target = target[1:-1]

        if isinstance(actual_value, (list, tuple)):
            # 数组包含检查
            found = any(str(item) == target or (isinstance(item, dict) and item.get("stage") == target) for item in actual_value)
            # 特殊处理：检查数组元素的字段值
            if not found and actual_value and isinstance(actual_value[0], dict):
                # 尝试检查每个元素的特定字段
                found = any(item.get("stage") == target for item in actual_value if isinstance(item, dict))
            return found, "expect array/string contains '%s', got %s" % (target, repr(actual_value))
        elif isinstance(actual_value, str):
            # 字符串包含检查
            return target in actual_value, "expect string contains '%s', got %s" % (target, repr(actual_value))
        else:
            return False, "contains() requires array or string, got %s" % type(actual_value).__name__

    # 2. 比较操作符 (>、<、>=、<=、!=、==)
    # 注意：需要先匹配 >= 和 <=，再匹配 > 和 <，避免部分匹配
    compare_match = re.match(r"^(>=|<=|==|!=|>|<)(.+)$", expr)
    if compare_match:
        op = compare_match.group(1)
        compare_value_str = compare_match.group(2).strip()

        # 处理特殊值
        if compare_value_str.lower() == "null":
            compare_value = None
        elif compare_value_str.lower() == "true":
            compare_value = True
        elif compare_value_str.lower() == "false":
            compare_value = False
        else:
            try:
                # 尝试转换为数字
                if '.' in compare_value_str:
                    compare_value = float(compare_value_str)
                else:
                    compare_value = int(compare_value_str)
            except ValueError:
                compare_value = compare_value_str

        # 执行比较
        try:
            if op == ">":
                result = actual_value > compare_value
            elif op == "<":
                result = actual_value < compare_value
            elif op == ">=":
                result = actual_value >= compare_value
            elif op == "<=":
                result = actual_value <= compare_value
            elif op == "==":
                result = actual_value == compare_value
            elif op == "!=":
                result = actual_value != compare_value
            else:
                return False, "unknown operator: %s" % op

            if result:
                return result, "OK"
            else:
                return result, "expect actual %s %s %s, but got %s (actual=%s)" % (
                    repr(actual_value), op, repr(compare_value),
                    "not satisfied" if op in (">", "<", ">=", "<=") else "different",
                    repr(actual_value)
                )
        except TypeError as e:
            return False, "comparison error: %s (actual=%s, compare=%s)" % (str(e), repr(actual_value), repr(compare_value))

    # 3. exists 检查
    if expr == "exists":
        return actual_value is not None and actual_value != "", "expect exists (not None/empty), got %s" % repr(actual_value)

    # 4. not_exists 检查
    if expr == "not_exists" or expr == "null":
        return actual_value is None or actual_value == "", "expect not_exists/null, got %s" % repr(actual_value)

    # 5. regex 正则匹配
    regex_match = re.match(r"^regex\((.+)\)$", expr)
    if regex_match:
        pattern = regex_match.group(1).strip()
        # 移除可能的引号
        if pattern.startswith("'") and pattern.endswith("'"):
            pattern = pattern[1:-1]
        elif pattern.startswith('"') and pattern.endswith('"'):
            pattern = pattern[1:-1]

        if actual_value is None:
            return False, "regex match failed: actual value is None"

        try:
            result = re.search(pattern, str(actual_value)) is not None
            return result, "expect regex match '%s', got %s" % (pattern, repr(actual_value))
        except re.error as e:
            return False, "regex pattern error: %s" % str(e)

    # 6. type 类型检查
    type_match = re.match(r"^type\((.+)\)$", expr)
    if type_match:
        expected_type = type_match.group(1).strip().lower()
        # 移除可能的引号
        if expected_type.startswith("'") and expected_type.endswith("'"):
            expected_type = expected_type[1:-1]
        elif expected_type.startswith('"') and expected_type.endswith('"'):
            expected_type = expected_type[1:-1]

        type_map = {
            "string": str, "str": str,
            "integer": int, "int": int,
            "float": float, "number": (int, float),
            "boolean": bool, "bool": bool,
            "array": (list, tuple), "list": (list, tuple),
            "object": dict, "dict": dict,
            "null": type(None)
        }

        expected_type_class = type_map.get(expected_type)
        if expected_type_class is None:
            return False, "unknown type: %s" % expected_type

        result = isinstance(actual_value, expected_type_class)
        return result, "expect type '%s', got '%s'" % (expected_type, type(actual_value).__name__)

    # 7. 精确匹配（默认行为）
    # 处理字符串形式的特殊值
    if expr.lower() == "null":
        return actual_value is None, "expect null, got %s" % repr(actual_value)
    elif expr.lower() == "true":
        return actual_value == True, "expect true, got %s" % repr(actual_value)
    elif expr.lower() == "false":
        return actual_value == False, "expect false, got %s" % repr(actual_value)

    # 尝试数字转换
    try:
        if '.' in expr:
            num_expected = float(expr)
            return actual_value == num_expected, "expect %s, got %s" % (num_expected, repr(actual_value))
        else:
            num_expected = int(expr)
            return actual_value == num_expected, "expect %s, got %s" % (num_expected, repr(actual_value))
    except ValueError:
        pass

    # 字符串精确匹配
    return actual_value == expr, "expect '%s', got %s" % (expr, repr(actual_value))


def _resolve_authorization(case_info):
    """默认 Bearer 见 conftest（环境变量优先）；套件 token_source: login 时再调 get_token。"""
    src = (case_info.get("_token_source") or "").strip().lower()
    if src not in ("login", "get_token", "token_provider"):
        return BEARER_AUTH
    try:
        from src.common.token_provider import get_token, clear_token_cache

        user, pwd = at_env.admin_credentials(config)
        if not user:
            allure.attach("token_source=login 但未配置 test_data.admin_user，已回退默认 Bearer", name="鉴权说明")
            return BEARER_AUTH
        # 清理缓存并强制刷新token，避免使用过期的缓存token
        clear_token_cache(user)
        tok = get_token(user, pwd, force_refresh=True)
        if tok:
            return "Bearer %s" % tok
        allure.attach(
            "get_token 返回空，已回退默认 Bearer（API_ACCESS_TOKEN / test_data.application_token）",
            name="鉴权说明",
        )
    except Exception as ex:
        allure.attach(str(ex), name="get_token 异常")
    return BEARER_AUTH


def _jinja_random_string(length=8, model=8):
    """与 etrino 用例中 `random_string(6, 8)` 一致，供 Jinja 渲染；不依赖 faker。"""
    chars = ""
    if model & 1:
        chars += _str.whitespace
    if (model >> 1) & 1:
        chars += _str.ascii_lowercase
    if (model >> 2) & 1:
        chars += _str.ascii_uppercase
    if (model >> 3) & 1:
        chars += _str.digits
    if (model >> 4) & 1:
        chars += _str.hexdigits
    if (model >> 5) & 1:
        chars += _str.octdigits
    if (model >> 6) & 1:
        chars += _str.punctuation + r"·！￥……（）【】、：；“‘《》，。？、"
    if not chars:
        chars = _str.ascii_lowercase + _str.digits

    return json.dumps(''.join(random.choices(chars, k=max(1, int(length))))).strip('"')


_jinja_env = Environment()
_jinja_env.globals["random_string"] = _jinja_random_string
_jinja_env.globals["now"] = datetime.now
_jinja_env.globals["timedelta"] = timedelta


def _render_jinja_fields(case_info):
    out = {}
    for k, v in case_info.items():
        out[k] = _jinja_env.from_string(v).render() if isinstance(v, str) else v
    return out


def pytest_generate_tests(metafunc):
    if metafunc.function.__name__ != "test_case":
        return
    cl = compute_case_list()
    argvals = [(x["feature"], x["story"], x["name"], x) for x in cl]
    metafunc.parametrize("feature, story, case_name, case_info", argvals)


@allure.title("{case_name}")
def test_case(feature, story, case_name, case_info):
    print("run case: %s @ %s.%s" % (case_name, story, feature))
    allure.attach(
        body="run case: %s @ %s.%s" % (case_name, story, feature), name="用例名称"
    )

    allure.dynamic.feature(feature)
    allure.dynamic.story(story)

    if case_info["prev_case"]:
        with allure.step("执行前置用例执行"):
            for x in compute_case_list():
                if x["name"] == case_info["prev_case"]:
                    test_case(x["feature"], x["story"], x["name"], x)
                    # 若存在同名用例，仅执行第一个匹配项
                    break

    with allure.step("加载用例参数"):
        # 更新前序用例提取的变量
        case_info = replace_params(case_info, **resp_values)

        # 替换 DP_AT 风格占位符（${ts_uuid}, ${random_str} 等）
        case_info = {k: replace_placeholders(v) for k, v in case_info.items()}
        case_info = _render_jinja_fields(case_info)

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
        case_headers["Authorization"] = _resolve_authorization(case_info)
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
        _scheme, _host = at_env.resolve_request_target(config)
        client = HTTPClient(url="%s://%s%s" % (_scheme, _host, case_info["url"]),
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
        resp_body = client.resp_body()
        for jsonpath_key, expected_value in json.loads(case_info["resp_check"]).items():
            # 获取实际值
            actual_values = jsonpath.jsonpath(resp_body, jsonpath_key)
            # jsonpath.jsonpath() 返回 False 或 None 表示未找到匹配项
            if actual_values is False or actual_values is None or (isinstance(actual_values, list) and len(actual_values) == 0):
                # jsonpath 未找到匹配项
                assert False, "jsonpath '%s' not found in response" % jsonpath_key

            actual_value = actual_values[0]

            # 使用扩展的表达式解析函数进行验证
            passed, error_msg = _parse_check_expression(expected_value, actual_value, jsonpath_key)
            assert passed, "resp_check failed for '%s': %s" % (jsonpath_key, error_msg)

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
