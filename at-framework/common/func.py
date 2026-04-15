#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
@Time   : 2024/7/10 11:09
@Author : Leopold.yu
@File   : func.py
"""
import configparser
import copy
import json
import os
import random
import re
import string
import time
import uuid

from genson import SchemaBuilder

try:
    import yaml

    _YAML_AVAILABLE = True
except ImportError:
    _YAML_AVAILABLE = False

# 高频常用汉字表（用于随机中文生成）
_COMMON_HANZI = (
    "的一是在不了有和人这中大为上个国我以要他时来用们生到作地于出就分对成会可主发年动同工也能下过子说产种面而方后多定行学法所民得经十"
    "三之进等部度家电力里如水化高自二理起小物现实现加量都两体制机当使点从业本去把性好应开它合还因由其些然前外天政四日那主义事平形相全表间样与关各重新线内数正心反你明看原又么利比或但质气第向道路命此变条只没结解决问意建月公无系军很情者最立代想已通并提直题党程展五果料象员革位入常文总次品式活设置及管特件长求老头基资边流路级少图山统接知较将组见计别她手角期根论运农指几九区强放决西被干做必战先回则任选取据处队南给色光门即保治北造百规热领七海口东导器压志世金增争济阶油思术极交受联什认六共权收证改清念建"
)


def _strict_missing_api():
    """环境变量 AT_STRICT_LOAD_APIS=1/true 时，apis.yaml 中不存在的接口名将导致加载失败而非静默跳过。"""
    v = (os.environ.get("AT_STRICT_LOAD_APIS") or "").strip().lower()
    return v in ("1", "true", "yes", "on")


def _read_yaml(path):
    if not os.path.isfile(path):
        return {}

    with open(path, "r", encoding="utf-8") as f:
        return yaml.safe_load(f)


# def load_global_manifest(base_dir):
#     """
#     加载全局变量清单（global_manifest.yaml），供智能体提取：有哪些变量、如何引用。
#     返回 list[dict]，每项含 name、description（可选）、ref（可选，默认 ${name}）。
#     """
#     if not _YAML_AVAILABLE:
#         raise ImportError("需要安装 PyYAML")
#     base_dir = os.path.abspath(base_dir)
#     config_dir = os.path.join(base_dir, "_config")
#     if not os.path.isdir(config_dir):
#         config_dir = base_dir
#     path = os.path.join(config_dir, "global_manifest.yaml")
#     data = _read_yaml(path)
#     variables = data.get("variables", [])
#     out = []
#     for v in variables:
#         if isinstance(v, dict):
#             item = {"name": v.get("name", ""), "description": v.get("description", ""),
#                     "ref": v.get("ref", "${%s}" % v.get("name", ""))}
#             if "used_in" in v:
#                 item["used_in"] = v["used_in"]
#             out.append(item)
#         else:
#             out.append({"name": str(v), "description": "", "ref": "${%s}" % v})
#     return out

# def load_case_from_yaml(base_dir):
#     """
#     从 YAML 目录加载用例，返回与 load_case 相同结构的 case_list。
#     目录结构建议：
#       base_dir/
#         _config/
#           global.yaml   # 全局变量 name: value
#           apis.yaml    # 接口定义 name -> {name, url, method, headers}
#         suites/
#           *.yaml       # 每个文件: feature, story, switch, cases (list)
#     依赖 PyYAML，未安装时抛出 ImportError。
#     """
#     if not _YAML_AVAILABLE:
#         raise ImportError("YAML 用例加载需要安装 PyYAML: pip install PyYAML")
#
#     base_dir = os.path.abspath(base_dir)
#     config_dir = os.path.join(base_dir, "_config")
#     suites_dir = os.path.join(base_dir, "suites")
#
#     # 全局变量：支持 {name: value} 或 [{name, value}, ...]
#     raw_global = _read_yaml(os.path.join(config_dir, "global.yaml"), {})
#     if isinstance(raw_global, list):
#         global_params = {item["name"]: item["value"] for item in raw_global}
#     else:
#         global_params = dict(raw_global)
#     # 嵌套引用一次替换
#     params = dict(global_params)
#     global_flat = {k: string.Template(str(v)).safe_substitute(**params) for k, v in params.items()}
#
#     # 接口信息：name -> {name, url, method, headers}
#     apis_list = _read_yaml(os.path.join(config_dir, "apis.yaml"), {})
#     if isinstance(apis_list, list):
#         api_params = {item["name"]: item for item in apis_list}
#     else:
#         api_params = {k: v if isinstance(v, dict) else {"name": k, "url": v, "method": "GET", "headers": "{}"}
#                      for k, v in apis_list.items()}
#     for k, v in api_params.items():
#         if "name" not in v:
#             v["name"] = k
#         if "url" not in v:
#             v["url"] = ""
#         if "method" not in v:
#             v["method"] = "GET"
#         if "headers" not in v:
#             v["headers"] = "{}"
#
#     # 用例字段默认值（与 YAML case 结构一致，含 OpenAPI 对齐的 header/cookie/resp_headers）
#     case_keys = ["name", "url", "prev_case", "path_params", "query_params", "header_params", "cookie_params",
#                  "body_params", "form_params", "resp_values", "code_check", "resp_headers_check",
#                  "resp_schema", "resp_check", "description"]
#
#     def _normalize_case(c):
#         out = {}
#         for key in case_keys:
#             val = c.get(key, "")
#             if isinstance(val, (dict, list)):
#                 val = json.dumps(val, ensure_ascii=False)
#             out[key] = str(val) if val is not None else ""
#         return out
#
#     case_list = []
#     skipped_missing_api = []
#     if not os.path.isdir(suites_dir):
#         return case_list
#
#     for fn in sorted(os.listdir(suites_dir)):
#         if not fn.endswith((".yaml", ".yml")):
#             continue
#         path = os.path.join(suites_dir, fn)
#         if not os.path.isfile(path):
#             continue
#         with open(path, "r", encoding="utf-8") as f:
#             suite = yaml.safe_load(f) or {}
#         feature = suite.get("feature", "")
#         story = suite.get("story", fn.replace(".yaml", "").replace(".yml", ""))
#         if str(suite.get("switch", "")).lower() != "y":
#             continue
#         suite_tags = suite.get("tags")
#         if suite_tags is None:
#             suite_tags = []
#         if not isinstance(suite_tags, list):
#             suite_tags = [suite_tags] if suite_tags else []
#         for c in suite.get("cases", []):
#             case = _normalize_case(c)
#             api_name = case["url"]
#             if api_name not in api_params:
#                 cname = (case.get("name") or "").strip() or "(unnamed)"
#                 skipped_missing_api.append("  suite=%s case=%s api_name=%r" % (fn, cname, api_name))
#                 continue
#             case["feature"] = feature
#             case["story"] = story
#             case["_suite_file"] = fn
#             # 仅从 apis 合并 url/method/headers，避免 API 项的 name 覆盖用例的 name
#             api_row = api_params[api_name]
#             case["url"] = api_row.get("url", "")
#             case["method"] = api_row.get("method", "GET")
#             case["headers"] = api_row.get("headers", "{}")
#             case = replace_params(case, **global_flat)
#             case["api_name"] = api_name
#             case["_suite_file"] = fn
#             ct = c.get("tags")
#             if isinstance(ct, list) and ct:
#                 case["tags"] = list(ct)
#             else:
#                 case["tags"] = list(suite_tags)
#             # 套件或单条用例可设 token_source: login，执行时走 token_provider.get_token
#             case["_token_source"] = str(
#                 c.get("token_source") or suite.get("token_source", "") or ""
#             ).strip()
#             case_list.append(case)

def _resolve_scope_to_tags(base_dir: str, scope):
    """根据 path_scope_mapping 将 scope(id) 解析为 tags 列表。"""
    config_dir = os.path.join(os.path.abspath(base_dir), "_config")
    mapping = _read_yaml(os.path.join(config_dir, "path_scope_mapping.yaml"))
    for sub in mapping.get("subsystems", []):
        if sub.get("id") == scope:
            tags = list(sub.get("scope_tags", []))
            if sub.get("smoke_tags"):
                tags.extend(sub["smoke_tags"])
            return list(set(tags))
    return []


def load_case_by_suite(file_path: str):
    """
    从suite配置中加载用例信息
    """
    with open(file_path, "r", encoding="utf-8") as fp:
        suite = yaml.safe_load(fp)

    if suite["switch"] not in ('y', 'Y', '1', 'ON'):
        return []

    case_list = []
    for case in suite["cases"]:
        case_info = {
            "feature": suite.get("feature"),
            "story": suite.get("story"),
            "switch": suite.get("switch"),
            "description": suite.get("description"),
            "tags": suite.get("tags", []),
            "_suite_file": os.path.basename(file_path),
            "_token_source": suite.get("token_source", "")
        }

        # 追加case级别tag并去重
        case_info["tags"].extend(case.get("tags", []))
        case_info["tags"] = sorted(set(case_info["tags"]))
        case.pop("tags", "")

        # 除tag外，同名字段case覆盖suite配置
        case_info.update(case)

        case_list.append(case_info)

    return case_list


def load_case(path: str):
    """
    从 YAML 用例目录加载用例（唯一支持方式）。
    返回 case_list，供 pytest parametrize 使用。
    path目录结构：
      path/
        _config/
          global.yaml   # 全局变量 name: value
          apis.yaml    # 接口定义 name -> {name, url, method, headers}
        suites/
          *.yaml       # 每个文件: feature, story, switch, cases (list)
    """
    path = os.path.abspath(path)
    if not os.path.isdir(path):
        raise ValueError("case_file 需为模块用例目录路径，例如 ./testcase/vega")
    if not _YAML_AVAILABLE:
        raise ImportError("YAML 用例加载需要安装 PyYAML: pip install PyYAML")

    # 加载全局变量
    global_flat = _read_yaml(os.path.join(path, "_config", "global.yaml"))

    # 加载接口信息
    # 为方便查找，转换index格式：name -> {name, url, method, headers, response: {200: {}, 400: {}}}
    api_list = _read_yaml(os.path.join(path, "_config", "apis.yaml"))
    api_params = {item["name"]: item for item in api_list}

    case_list = []
    # 加载suite
    for root, dirs, files in os.walk(os.path.join(path, "suites")):
        for f in files:
            if not f.endswith((".yaml", ".yml")):
                continue

            full_path = os.path.join(root, f)
            case_list.extend(load_case_by_suite(full_path))

    # 终止运行并抛出异常case
    if _strict_missing_api():
        exception_case = [
            "suite=%s，case=%s，api_name=%s" % (x["_suite_file"], x["name"], x["url"])
            for x in case_list if x["url"] not in api_params
        ]

        msg = (
                "以下用例引用的接口名未在 apis.yaml 中定义，已跳过加载（请补全 apis 或修正 case 的 url 字段）：\n%s"
                % "\n".join(exception_case)
        )
        raise ValueError(msg)

    # 更新api信息
    case_list = [{**x, **api_params[x["url"]]} for x in case_list if x["url"] in api_params]

    # 替换全局变量
    case_list = [replace_params(x, **global_flat) for x in case_list]

    return case_list


def get_cases(base_dir: str, scope=None, tags=None, suite=None, name=None, names=None, api_name=None, api_path=None):
    """
    按需提取用例（粒度到单条），供执行器或智能体在有新提交时根据内容灵活筛选用例。
    - scope: 与 path_scope_mapping 的 subsystem.id 一致，解析为 tags 再按 case.tags 过滤
    - tags: 列表，单条 case 的 tags 与其中任一相交则保留（case 级 tags）
    - suite: 套件 story 或文件名，精确到该套件
    - name: 用例 name 精确匹配（单条）
    - names: 用例 name 列表，只保留 name 在此列表中的用例；智能体根据提交内容选出要回归的 case 名后传此参数
    - api_name: 接口名称（与 apis.yaml 的 name 一致），只保留调用该接口的用例
    - api_path: 接口路径（如 /api/...），只保留 url 与该路径匹配的用例
    返回与 load_case_from_yaml 同结构的 case 列表（含 tags、api_name、_suite_file）。
    """
    path = os.path.abspath(base_dir)
    if not os.path.isdir(path):
        raise ValueError("base_dir 需为模块用例目录路径，例如 ./testcase/vega")

    full_list = load_case(path)
    out = full_list

    # 按用例 name 列表筛选（智能体根据提交内容灵活选出要回归的 case 名）
    if names is not None and len(names) > 0:
        names_set = set(n.strip() for n in names if n and str(n).strip())
        if names_set:
            out = [c for c in out if (c.get("name") or "").strip() in names_set]
            return out

    requested_tags = []
    if scope:
        requested_tags.extend(_resolve_scope_to_tags(path, scope))
    if tags:
        if isinstance(tags, str):
            tags = [t.strip() for t in tags.split(",") if t.strip()]
        requested_tags.extend(tags)
    requested_tags = list(set(requested_tags))
    if requested_tags:
        out = [c for c in out if c.get("tags") and set(c["tags"]) & set(requested_tags)]
    if suite:
        suite_clean = str(suite).replace(".yaml", "").replace(".yml", "")
        out = [c for c in out if c.get("story") == suite_clean or (
                c.get("_suite_file", "").replace(".yaml", "").replace(".yml", "") == suite_clean)]
    if name:
        out = [c for c in out if c.get("name") == name]
    if api_name:
        out = [c for c in out if c.get("api_name") == api_name]
    if api_path:
        ap = str(api_path).strip()
        out = [c for c in out if (c.get("url") or "").strip() == ap or ap in (c.get("url") or "")]
    return out


def replace_params(input_case, **kwargs):
    tmp_case = copy.deepcopy(input_case)
    '''
    使用JinJa2.Template会导致未配置的参数项置空
    故此处应用string.Template，仅替换信息
    用例执行时渲染参数
    '''
    output_case = {k: string.Template(str(v)).safe_substitute(kwargs) for k, v in tmp_case.items()}
    return output_case


def load_sys_config(file):
    cfg = configparser.ConfigParser(interpolation=None)
    cfg.read(file, encoding="utf-8")
    return {x: {y[0]: y[1] for y in cfg.items(x)} for x in cfg.sections()}


def genson(data: dict):
    builder = SchemaBuilder()
    builder.add_object(data)
    to_schema = builder.to_schema()
    return to_schema


# ============= DP_AT 风格占位符替换支持 =============

def _random_str(length=8, kind="alnum"):
    """生成随机字符串"""
    if length <= 0:
        return ""
    kind = (kind or "alnum").lower()
    if kind == "digits":
        pool = string.digits
    elif kind == "letters":
        pool = string.ascii_letters
    elif kind == "hex":
        pool = "0123456789abcdef"
    else:
        pool = string.ascii_letters + string.digits
    return "".join(random.choice(pool) for _ in range(length))


def _random_int(min_value=0, max_value=10 ** 9):
    """生成随机整数"""
    if min_value > max_value:
        min_value, max_value = max_value, min_value
    return random.randint(min_value, max_value)


def _random_cn(length=4):
    """生成随机中文"""
    if length <= 0:
        return ""
    return "".join(random.choice(_COMMON_HANZI) for _ in range(length))


def _fake_cn_name():
    """生成中文姓名"""
    return _random_cn(random.randint(2, 3))


def _ts_uuid():
    """生成时间戳 UUID"""
    return f"{int(time.time())}_{_random_str(6)}"


def _gen_uuid():
    """生成标准 UUID"""
    return str(uuid.uuid4())


def _ts_ms():
    """当前时间戳（毫秒）"""
    return int(time.time() * 1000)


def _ts_s():
    """当前时间戳（秒）"""
    return int(time.time())


def replace_placeholders(text):
    """
    替换 DP_AT 风格占位符：${func(args)}
    支持的占位符：
    - ${ts_uuid} -> 时间戳 UUID
    - ${uuid} -> 标准 UUID
    - ${random_str} / ${random_str(n)} -> 随机字符串
    - ${random_int} / ${random_int(min,max)} -> 随机整数
    - ${fake_cn_name} -> 中文姓名
    - ${random_cn} / ${random_cn(n)} -> 随机中文
    - ${ts_ms} -> 毫秒时间戳
    - ${ts_s} -> 秒时间戳
    """
    if not isinstance(text, str):
        return text

    if "${" not in text:
        return text

    pattern = re.compile(r'\$\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*(?:\(([^)]*)\))?\}')

    def _parse_args(arg_str):
        if not arg_str:
            return []
        parts = []
        for raw in arg_str.split(","):
            item = raw.strip()
            if item:
                try:
                    parts.append(int(item))
                except ValueError:
                    parts.append(item)
        return parts

    def repl(m):
        func_name = m.group(1).strip()
        args = _parse_args(m.group(2)) if m.group(2) else []

        func_map = {
            'ts_uuid': lambda: _ts_uuid(),
            'uuid': lambda: _gen_uuid(),
            'random_str': lambda: _random_str(*args) if args else _random_str(),
            'random_int': lambda: _random_int(*args) if args else _random_int(),
            'fake_cn_name': lambda: _fake_cn_name(),
            'random_cn': lambda: _random_cn(*args) if args else _random_cn(),
            'ts_ms': lambda: _ts_ms(),
            'ts_s': lambda: _ts_s(),
        }

        func = func_map.get(func_name)
        if func:
            try:
                return str(func())
            except Exception:
                return m.group(0)
        return m.group(0)

    return pattern.sub(repl, text)


def replace_params_with_placeholders(input_case, **kwargs):
    """
    先替换 global 变量，再替换占位符
    """
    tmp_case = copy.deepcopy(input_case)
    # 第一步：替换 global 变量
    output_case = {k: string.Template(str(v)).safe_substitute(**kwargs) for k, v in tmp_case.items()}
    # 第二步：替换占位符
    output_case = {k: replace_placeholders(v) for k, v in output_case.items()}
    return output_case


# ============= 占位符替换支持结束 =============


if __name__ == "__main__":
    base = os.path.join(os.path.dirname(__file__), "..")
    case_file = os.path.join(base, "testcase", "agent-backend")  # 默认示例模块
    rst = load_case(case_file)
    # for x in rst:
    #     print(x)
