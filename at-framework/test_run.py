#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
@Time   : 2024/7/1 9:54
@Author : Leopold.yu
@File   : test_run.py
"""
import json
import os
import random
import re
import string as _str
import time
from datetime import datetime, timedelta

import allure
import jsonpath
import pytest
import requests
import urllib3
from jinja2 import Environment
from jsonschema.validators import validate

from common import at_env
from common.func import replace_params, replace_placeholders, genson
from common.model_registry import (
    ensure_llm_model_and_get_id,
    register_embedding_model,
    register_rerank_model,
    ensure_oss_storage_and_get_id,
)
from conftest import config, compute_case_list
from request.http_client import HTTPClient

resp_values = {}
urllib3.disable_warnings(urllib3.connectionpool.InsecureRequestWarning)


def _truthy(val):
    if val is None:
        return False
    if isinstance(val, bool):
        return val
    s = str(val).strip().lower()
    return s in ("1", "true", "yes", "on", "y")


def _default_bearer_auth():
    """
    默认 Bearer：
    - AT_AUTH_SOURCE=login 时尝试 get_token
    - 否则使用静态 token（API_ACCESS_TOKEN / test_data.application_token）
    """
    source = at_env.auth_token_source(config)
    if source == "login":
        try:
            from src.common.token_provider import get_token, clear_token_cache

            user, pwd = at_env.admin_credentials(config)
            if user:
                clear_token_cache(user)
                tok = at_env.normalize_bearer_token_value(get_token(user, pwd, force_refresh=True) or "")
                if tok:
                    return "Bearer %s" % tok
        except Exception as ex:
            allure.attach(str(ex), name="get_token 异常")

    static_token = at_env.static_access_token(config)
    if static_token:
        return "Bearer %s" % static_token
    # isf=true 时要求默认都带 Authorization 头，token 缺失时保留 Bearer 空值
    return "Bearer "


def _resolve_authorization(case_info):
    """
    按 case token_source 决定是否强制 login 获取 token，否则走默认 Bearer。
    """
    src = (case_info.get("_token_source") or "").strip().lower()
    if src in ("login", "get_token", "token_provider"):
        try:
            from src.common.token_provider import get_token, clear_token_cache

            user, pwd = at_env.admin_credentials(config)
            if user:
                clear_token_cache(user)
                tok = at_env.normalize_bearer_token_value(get_token(user, pwd, force_refresh=True) or "")
                if tok:
                    return "Bearer %s" % tok
        except Exception as ex:
            allure.attach(str(ex), name="get_token 异常")
    return _default_bearer_auth()


def _parse_prepare_models(case_info):
    """
    从单条用例读取前置模型需求。
    - prepare_models: 字符串 "embedding" / "reranker" / "oss" / "llm" / 逗号组合 或 YAML 列表
    - prepare_embedding / prepare_reranker / prepare_oss / prepare_llm: 可选布尔，等价于列入 prepare_models
    """
    names = []
    raw = case_info.get("prepare_models")
    if raw is None or raw == "":
        parts = []
    elif isinstance(raw, str):
        parts = [x.strip() for x in re.split(r"[,;\s]+", raw.strip()) if x.strip()]
    elif isinstance(raw, list):
        parts = [str(x).strip() for x in raw if str(x).strip()]
    else:
        parts = []

    if _truthy(case_info.get("prepare_embedding")):
        parts.append("embedding")
    if _truthy(case_info.get("prepare_reranker")):
        parts.append("reranker")
    if _truthy(case_info.get("prepare_oss")):
        parts.append("oss")
    if _truthy(case_info.get("prepare_llm")):
        parts.append("llm")

    seen = set()
    for p in parts:
        pl = p.lower()
        if pl in ("embedding", "embed"):
            if "embedding" not in seen:
                seen.add("embedding")
                names.append("embedding")
        elif pl in ("reranker", "rerank"):
            if "reranker" not in seen:
                seen.add("reranker")
                names.append("reranker")
        elif pl in ("oss", "storage"):
            if "oss" not in seen:
                seen.add("oss")
                names.append("oss")
        elif pl in ("llm", "large"):
            if "llm" not in seen:
                seen.add("llm")
                names.append("llm")
    return names


def _run_prepare_models_for_case(case_info):
    """按单条用例声明，向 mf-model-manager 注册所需模型（配置仍来自 config.ini）。"""
    for kind in _parse_prepare_models(case_info):
        if kind == "embedding":
            embedding_id = register_embedding_model(config)
            if embedding_id:
                resp_values["embedding_config_id"] = str(embedding_id)
            embedding_name = ((config.get("embedding_info") or {}).get("embedding_model_name") or "").strip()
            if embedding_name:
                resp_values["embedding_config_name"] = embedding_name
        elif kind == "reranker":
            reranker_id = register_rerank_model(config)
            if reranker_id:
                resp_values["reranker_config_id"] = str(reranker_id)
            reranker_name = ((config.get("rerank_info") or {}).get("rerank_model_name") or "").strip()
            if reranker_name:
                resp_values["reranker_config_name"] = reranker_name
        elif kind == "oss":
            storage_id = ensure_oss_storage_and_get_id(config)
            if storage_id:
                resp_values["storage_config_id"] = str(storage_id)
            storage_name = ((config.get("oss_info") or {}).get("storage_name") or "").strip()
            if storage_name:
                resp_values["storage_config_name"] = storage_name
        elif kind == "llm":
            llm_id = ensure_llm_model_and_get_id(config)
            if llm_id:
                resp_values["llm_config_id"] = str(llm_id)
            llm_name = ((config.get("llm_info") or {}).get("llm_name") or "").strip()
            if llm_name:
                resp_values["llm_config_name"] = llm_name


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
    - size: "size(n)" - 长度检查

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
            found = any(str(item) == target or (isinstance(item, dict) and item.get("stage") == target) for item in
                        actual_value)
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
            return False, "comparison error: %s (actual=%s, compare=%s)" % (str(e), repr(actual_value),
                                                                            repr(compare_value))

    # 3. exists 检查
    if expr == "exists":
        return actual_value is not None and actual_value != "", "expect exists (not None/empty), got %s" % repr(
            actual_value)

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

    # size 长度检查
    len_match = re.match(r"^len\((.+)\)$", expr)
    if len_match:
        len_num = int(len_match.group(1).strip().lower())
        return len(actual_value) == len_num, "expect size '%d', got '%d'" % (len_num, len(actual_value))

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


def _jinja_random_string(length=8, model=8):
    """
    生成指定长度的随机字符串
    :param length: 字符串长度
    :param model: 按位指定字符串类型，
                  1=string.whitespace
                  2=string.ascii_lowercase
                  4=string.ascii_uppercase
                  8=string.digits
                  16=string.hexdigits
                  32=string.octdigits
                  64=string.punctuation
                  128=常见中文
    """

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

    return json.dumps(''.join(random.choices(chars, k=length))).strip('"')


def _jinja_timestamp(scale="s", delta=0, precision=0):
    """
    Jinja2 全局函数 ``timestamp`` 的实现：返回 Unix 时间戳（本地日历语义下的整秒时刻 + 差值，
    可选按粒度对齐后，再按秒或毫秒输出）。

    计算顺序：
    1. 取 ``int(datetime.now().timestamp())``（当前本地时间的整秒）作为基准；
    2. 加上由 ``delta`` 换算得到的偏移秒数 ``offset_sec``；
    3. 若 ``precision`` 为下表中的取值之一，则将该时刻对齐到对应粒度的区间起点（本地时区），
       否则不对齐；
    4. 将对齐后的时刻取 ``int`` 再乘以倍率：秒模式 ×1，毫秒模式 ×1000。

    :param scale: 输出倍率。仅当取值为 ``ms`` 或 ``1000`` 时为毫秒时间戳（返回值 ×1000），
        其余任意值（含默认 ``s``）均为秒级时间戳（×1）。
    :param delta: 时间偏移。
        - ``datetime.timedelta``：使用 ``total_seconds()``，单位为秒，**不再**除以倍率；
        - ``int`` / ``float``：先除以当前 ``scale`` 对应倍率（秒模式 ÷1，毫秒模式 ÷1000），
          即秒模式下该数值表示「秒」，毫秒模式下该数值表示「毫秒」再换成秒参与运算；
        - 其他类型：视为 0。
    :param precision: 对齐粒度（须与下列取值**完全一致**，大小写敏感；未命中则不对齐）：
        - 分钟起点：``M``、``minute``、``minutes``、``分``（注意分钟为**大写** ``M``，与月份 ``m`` 区分）；
        - 小时起点：``H``、``hour``、``hours``、``时``；
        - 日起点（当天 0 点）：``d``、``day``、``days``、``天``；
        - 月起始（当月 1 日 0 点）：``m``、``month``、``months``、``月``；
        - 年起始（当年 1 月 1 日 0 点）：``y``、``year``、``years``、``年``；
        - 默认 ``0`` 及其他取值：不做日历对齐，仅使用步骤 2 得到的整秒时刻。

    模板示例：``{{ timestamp('ms', 0, 'day') }}``、``{{ timestamp('s', 3600, 'd') }}``。
    """
    factor = 1000 if scale in ("ms", "1000") else 1

    if isinstance(delta, timedelta):
        offset_sec = delta.total_seconds()
    elif isinstance(delta, int) or isinstance(delta, float):
        offset_sec = delta / factor
    else:
        offset_sec = 0

    # 单次采样，避免取值漂移
    t = int(datetime.now().timestamp()) + offset_sec

    # 颗粒度对齐
    dt = datetime.fromtimestamp(t)
    if precision in ("M", "minute", "minutes", "分"):
        floored = dt.replace(second=0, microsecond=0).timestamp()
    elif precision in ("H", "hour", "hours", "时"):
        floored = dt.replace(minute=0, second=0, microsecond=0).timestamp()
    elif precision in ("d", "day", "days", "天"):
        floored = dt.replace(hour=0, minute=0, second=0, microsecond=0).timestamp()
    elif precision in ("m", "month", "months", "月"):
        floored = dt.replace(day=1, hour=0, minute=0, second=0, microsecond=0).timestamp()
    elif precision in ("y", "year", "years", "年"):
        floored = dt.replace(month=1, day=1, hour=0, minute=0, second=0, microsecond=0).timestamp()
    else:
        floored = dt.timestamp()

    return int(floored) * factor


_jinja_env = Environment()
_jinja_env.globals["random_string"] = _jinja_random_string
_jinja_env.globals["timestamp"] = _jinja_timestamp
_jinja_env.globals["now"] = datetime.now
_jinja_env.globals["timedelta"] = timedelta


def _render_jinja_fields(case_info):
    out = {}
    for k, v in case_info.items():
        out[k] = _jinja_env.from_string(v).render() if isinstance(v, str) else v
    return out


def _contains_unresolved_placeholder(value):
    if isinstance(value, str):
        return "${" in value and "}" in value
    if isinstance(value, dict):
        return any(_contains_unresolved_placeholder(v) for v in value.values())
    if isinstance(value, (list, tuple)):
        return any(_contains_unresolved_placeholder(v) for v in value)
    return False


def _build_multipart_files(case_info):
    """
    解析 file_params，构建 requests 的 files 参数。
    当前支持：
    - {"file": "relative/or/abs/path.txt"}
    - {"fieldA": "pathA", "fieldB": "pathB"}
    返回: (files_dict, opened_file_handles)
    """
    raw = case_info.get("file_params", "")
    if not raw:
        return None, []
    try:
        file_params = json.loads(raw)
    except Exception:
        return None, []
    if not isinstance(file_params, dict) or not file_params:
        return None, []

    files = {}
    handles = []
    for field_name, file_path in file_params.items():
        if not field_name or not file_path:
            continue
        p = str(file_path)
        if not os.path.isabs(p):
            p = os.path.abspath(p)
        fh = open(p, "rb")
        handles.append(fh)
        files[str(field_name)] = (os.path.basename(p), fh)
    return files or None, handles


def _parse_teardown_operator_names(case_info):
    """
    解析单条用例声明的算子清理名称。
    支持：
      - teardown_operator_name: "name1,name2"
      - teardown_operator_name: ["name1", "name2"]
    """
    raw = case_info.get("teardown_operator_name")
    if raw is None or raw == "":
        return []
    if isinstance(raw, str):
        names = [x.strip() for x in re.split(r"[,;\n]+", raw) if x.strip()]
    elif isinstance(raw, (list, tuple, set)):
        names = [str(x).strip() for x in raw if str(x).strip()]
    else:
        names = [str(raw).strip()] if str(raw).strip() else []
    # 去重并保持顺序
    out = []
    seen = set()
    for n in names:
        if n not in seen:
            seen.add(n)
            out.append(n)
    return out


def _teardown_delete_operators_by_names(names):
    """按名称清理算子：先下架再删除。"""
    if not names:
        return
    scheme, host = at_env.resolve_request_target(config)
    base_url = "%s://%s" % (scheme, host)
    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json",
        "x-business-domain": "bd_public",
    }

    try:
        list_url = "%s/api/agent-operator-integration/v1/operator/info/list" % base_url
        resp = requests.get(list_url, params={"all": "true"}, headers=headers, verify=False, timeout=30)
        if resp.status_code != 200:
            print("[teardown] query operator list failed: %s" % resp.status_code)
            return
        operators = (resp.json() or {}).get("data", [])
    except Exception as e:
        print("[teardown] query operator list error: %s" % e)
        return

    name_to_ids = {}
    for op in operators:
        op_name = (op.get("name") or op.get("operator_info", {}).get("name") or "").strip()
        op_id = str(op.get("operator_id") or "").strip()
        if op_name and op_id:
            name_to_ids.setdefault(op_name, []).append(op_id)

    for name in names:
        ids = name_to_ids.get(name, [])
        if not ids:
            print("[teardown] operator not found by name: %s" % name)
            continue
        for op_id in ids:
            try:
                requests.post(
                    "%s/api/agent-operator-integration/v1/operator/status" % base_url,
                    json=[{"operator_id": op_id, "status": "offline"}],
                    headers=headers,
                    verify=False,
                    timeout=30,
                )
            except Exception:
                pass
            try:
                del_resp = requests.delete(
                    "%s/api/agent-operator-integration/v1/operator/delete" % base_url,
                    json=[{"operator_id": op_id}],
                    headers=headers,
                    verify=False,
                    timeout=30,
                )
                if del_resp.status_code == 200:
                    print("[teardown] [OK] deleted operator by name: %s (%s)" % (name, op_id))
                else:
                    print(
                        "[teardown] [FAIL] delete operator failed: %s (%s) - %s" % (name, op_id, del_resp.status_code))
            except Exception as e:
                print("[teardown] [FAIL] delete operator error: %s (%s) - %s" % (name, op_id, e))


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
    raw_url_template = case_info.get("url", "")

    if case_info.get("prev_case", "") != '':
        with allure.step("执行前置用例执行"):
            for x in compute_case_list():
                if x["name"] == case_info["prev_case"]:
                    try:
                        test_case(x["feature"], x["story"], x["name"], x)
                    except Exception as e:
                        # 前置失败时当前用例跳过，避免根因用例在依赖链中重复计失败。
                        pytest.skip("prev_case '%s' failed: %s" % (case_info["prev_case"], e))
                    # 若存在同名用例，仅执行第一个匹配项
                    break

    with allure.step("加载用例参数"):
        # 更新前序用例提取的变量
        case_info = replace_params(case_info, **resp_values)

        # 替换 DP_AT 风格占位符（${ts_uuid}, ${random_str} 等）
        case_info = {k: replace_placeholders(v) for k, v in case_info.items()}
        case_info = _render_jinja_fields(case_info)

        # 替换path参数
        case_path_params = json.loads(case_info.get("path_params", "{}"))
        case_info = replace_params(case_info, **case_path_params)
        # URL 以原始模板为基准再做一次 path_params 替换，确保 path_params 覆盖优先于 resp_values。
        if raw_url_template:
            case_info["url"] = replace_params({"url": raw_url_template}, **case_path_params)["url"]

        with allure.step("前置：小模型（prepare_models）"):
            _run_prepare_models_for_case(case_info)

        # 小模型前置执行后，再次替换新增的变量（如 embedding_config_id）
        case_info = replace_params(case_info, **resp_values)

        # 参数格式转换
        case_header_params = {
            **json.loads(case_info.get("headers", "{}")),
            **json.loads(case_info.get("header_params", "{}")),
        }
        # 双模式：isf=true 默认注入 token，isf=false 强制不带 token。
        if at_env.isf_enabled(config):
            case_header_params["Authorization"] = _resolve_authorization(case_info)
        else:
            case_header_params.pop("Authorization", None)
        case_cookie_params = json.loads(case_info.get("cookie_params", "{}"))
        case_query_params = json.loads(case_info.get("query_params", "{}"))
        case_body_params = json.loads(case_info.get("body_params", "{}"))
        case_form_params = json.loads(case_info.get("form_params", "{}"))
        case_file_params = json.loads(case_info.get("file_params", "{}")) if case_info.get("file_params") else {}
        is_multipart = _truthy(case_info.get("multipart"))
        url = case_info.get("url", "")
        unresolved_fields = []
        for _name, _val in (
                ("query_params", case_query_params),
                ("body_params", case_body_params),
                ("form_params", case_form_params),
                ("file_params", case_file_params),
                ("url", url),
        ):
            if _contains_unresolved_placeholder(_val):
                unresolved_fields.append(_name)
        if unresolved_fields:
            msg = "Unresolved placeholders found before request send: %s" % ", ".join(unresolved_fields)
            if case_info.get("prev_case", ""):
                pytest.skip("%s (likely caused by prev_case failure)" % msg)
            raise AssertionError(msg)

        allure.attach(
            body="url: %s\nheaders: %s\ncookies：%s\n"
                 "query_params: %s\nbody_params: %s\nform_params: %s\nfile_params: %s\nmultipart: %s" % (
                     url, case_header_params, case_cookie_params,
                     case_query_params, case_body_params, case_form_params, case_file_params, is_multipart),
            name="请求参数"
        )

    with allure.step("发送请求"):
        _scheme, _host = at_env.resolve_request_target(config)
        # 如果 url 已经是完整的 URL（包含协议），则直接使用，否则拼接基础 URL
        if url.lower().startswith("http://") or url.lower().startswith("https://"):
            full_url = url
        else:
            full_url = "%s://%s%s" % (_scheme, _host, url)
        client = HTTPClient(url=full_url,
                            method=case_info.get("method", ""), headers=case_header_params)
        send_kw = dict(params=case_query_params, json=case_body_params, data=case_form_params)
        if case_cookie_params:
            send_kw["cookies"] = case_cookie_params
        opened_files = []
        try:
            files, opened_files = _build_multipart_files(case_info)
            if files is not None:
                # multipart 上传由 requests 自动设置 Content-Type + boundary。
                send_kw.pop("json", None)
                if "Content-Type" in case_header_params:
                    case_header_params.pop("Content-Type", None)
                send_kw["files"] = files
            client.send(**send_kw)
        finally:
            for fh in opened_files:
                try:
                    fh.close()
                except Exception:
                    pass

        resp_code = client.resp_code()
        resp_body = client.resp_body()

        allure.attach(
            body="url: %s\nhttp_code: %s\nresponse: %s" % (
                client.url, resp_code, json.dumps(resp_body, ensure_ascii=False)),
            name="请求响应"
        )

    # 结果断言
    if "code_check" in case_info:
        assert str(resp_code) == case_info["code_check"]

    if "resp_headers_check" in case_info:
        for k, v in json.loads(case_info["resp_headers_check"]).items():
            assert client.resp.headers.get(k) == v

    if "resp_check" in case_info:
        for jsonpath_key, expected_value in json.loads(case_info["resp_check"]).items():
            # 获取实际值
            actual_values = jsonpath.jsonpath(resp_body, jsonpath_key)
            if not actual_values:
                assert False, "jsonpath '%s' not found in response" % jsonpath_key

            # 使用扩展的表达式解析函数进行验证
            passed, error_msg = _parse_check_expression(expected_value, actual_values[0], jsonpath_key)
            assert passed, "resp_check failed for '%s': %s" % (jsonpath_key, error_msg)

    if "resp_schema" in case_info:
        if resp_code in case_info["resp_schema"]:
            resp_schema = genson(json.loads(case_info["resp_schema"][resp_code]))
            validate(instance=resp_body, schema=resp_schema)

    # 提取响应中的变量
    if "resp_values" in case_info:
        for k, v in json.loads(case_info["resp_values"]).items():
            param = jsonpath.jsonpath(resp_body, v)[0]
            if isinstance(param, list) or isinstance(param, dict):
                # 将对象存储为字符串,加载用例参数时再转换为JSON格式
                resp_values[k] = json.dumps(param, ensure_ascii=False)
            else:
                resp_values[k] = param

    # 用例后置清理：按名称删除算子（示例：teardown_operator_name: duplicate_test_title_fixed）
    teardown_names = _parse_teardown_operator_names(case_info)
    if teardown_names:
        _teardown_delete_operators_by_names(teardown_names)

    # 执行后置用例
    if case_info.get("next_case", "") != '':
        with allure.step("执行后置用例"):
            for x in compute_case_list():
                if x["name"] == case_info["next_case"]:
                    try:
                        test_case(x["feature"], x["story"], x["name"], x)
                    except Exception as e:
                        # 后置用例失败不影响当前用例结果，仅记录警告
                        print("WARNING: next_case '%s' failed: %s" % (case_info["next_case"], e))
                    # 若存在同名用例，仅执行第一个匹配项
                    break

    # 针对部分异步用例，设置等待时长，单位s
    if "next_wait" in case_info:
        time.sleep(int(case_info["next_wait"]))


if __name__ == '__main__':
    pass
