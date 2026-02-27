#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
按需提取用例或全局变量清单，供流水线/智能体使用。智能体可根据提交内容灵活选出要回归的 case name，再通过 CASE_NAMES 执行。
用法:
  # 按用例 name 列表提取（智能体根据内容选出 name 后）
  python scripts/extract_cases.py --names "用例名1,用例名2,..." [--format json|yaml]
  # 导出全部用例的 name/description 等，供智能体根据提交内容筛选
  python scripts/extract_cases.py --list-fields name,description,story [--format json|yaml]
  python scripts/extract_cases.py [--scope SCOPE] [--tags TAGS] [--suite SUITE] [--name NAME] [--api-name API_NAME] [--api-path API_PATH] [--format json|yaml]
  python scripts/extract_cases.py --globals [--format json|yaml]
环境变量: CASE_FILE, SCOPE, TAGS, SUITE, NAME, CASE_NAMES, API_NAME, API_PATH 可替代对应参数。
"""
import argparse
import json
import os
import sys

# 保证可导入 common
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from common.func import get_cases, load_case_from_yaml, load_global_manifest, get_global_flat


def _base_dir_default():
    return os.environ.get("CASE_FILE", "").strip() or os.path.join(os.path.dirname(__file__), "..", "testcase", "vega")


def main():
    ap = argparse.ArgumentParser(description="Extract cases or global manifest by scope/tags/suite/name")
    ap.add_argument("--base-dir", default=_base_dir_default(), help="Module case root, e.g. testcase/vega (default: CASE_FILE or testcase/vega)")
    ap.add_argument("--scope", default=os.environ.get("SCOPE", "").strip(), help="Subsystem scope id from path_scope_mapping")
    ap.add_argument("--tags", default=os.environ.get("TAGS", "").strip(), help="Comma-separated tags")
    ap.add_argument("--suite", default=os.environ.get("SUITE", "").strip(), help="Suite story or file name")
    ap.add_argument("--name", default=os.environ.get("NAME", "").strip(), help="Case name exact match (single)")
    ap.add_argument("--names", default=os.environ.get("CASE_NAMES", "").strip(), help="Comma-separated case names; agent picks names by commit content, then extract/run by this list")
    ap.add_argument("--list-fields", default="", help="Output all cases with only these fields (e.g. name,description,story) for agent to choose names by content; no other filter")
    ap.add_argument("--api-name", default=os.environ.get("API_NAME", "").strip(), help="API name (apis.yaml), only cases calling this API")
    ap.add_argument("--api-path", default=os.environ.get("API_PATH", "").strip(), help="API path (e.g. /api/...), only cases with matching url")
    ap.add_argument("--globals", action="store_true", help="Output global variable manifest instead of cases")
    ap.add_argument("--format", choices=["json", "yaml"], default="json", help="Output format")
    args = ap.parse_args()

    base_dir = os.path.abspath(args.base_dir)
    if not os.path.isdir(base_dir):
        print("Error: base_dir not a directory: %s" % base_dir, file=sys.stderr)
        sys.exit(1)

    if args.globals:
        try:
            manifest = load_global_manifest(base_dir)
            data = {"variables": manifest}
        except Exception as e:
            print("Error loading global manifest: %s" % e, file=sys.stderr)
            sys.exit(1)
    else:
        list_fields = [f.strip() for f in (args.list_fields or "").split(",") if f.strip()]
        if list_fields:
            # 导出全部用例的指定字段，供智能体根据提交内容筛选出要回归的 name
            try:
                all_cases = load_case_from_yaml(base_dir)
                rows = []
                for c in all_cases:
                    row = {}
                    for k in list_fields:
                        if k in c:
                            row[k] = c[k]
                    rows.append(row)
                data = {"count": len(rows), "cases": rows}
            except Exception as e:
                print("Error load_case_from_yaml: %s" % e, file=sys.stderr)
                sys.exit(1)
        else:
            names_str = args.names or None
            names_list = [x.strip() for x in names_str.split(",") if x.strip()] if names_str else None
            scope = args.scope or None
            tags = args.tags or None
            suite = args.suite or None
            name = args.name or None
            api_name = args.api_name or None
            api_path = args.api_path or None
            try:
                cases = get_cases(base_dir, scope=scope, tags=tags, suite=suite, name=name, names=names_list, api_name=api_name, api_path=api_path)
                data = {"count": len(cases), "cases": cases}
            except Exception as e:
                print("Error get_cases: %s" % e, file=sys.stderr)
                sys.exit(1)

    if args.format == "yaml":
        try:
            import yaml
            out = yaml.dump(data, allow_unicode=True, default_flow_style=False, sort_keys=False)
        except ImportError:
            out = json.dumps(data, ensure_ascii=False, indent=2)
    else:
        out = json.dumps(data, ensure_ascii=False, indent=2)
    print(out)


if __name__ == "__main__":
    main()
