#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
校验测试模块目录结构是否可被框架加载（adp 各子模块接入前可跑此脚本）。
"""
import argparse
import os
import sys

# 保证可 import common
_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if _ROOT not in sys.path:
    sys.path.insert(0, _ROOT)

from common.func import load_case_from_yaml  # noqa: E402
from common.hooks import load_session_clean_up  # noqa: E402


def main():
    ap = argparse.ArgumentParser(description="Validate testcase module directory for AT framework")
    ap.add_argument("base_dir", help="e.g. testcase/vega or testcase/dataflow")
    ap.add_argument(
        "--strict-apis",
        action="store_true",
        help="若存在 case 引用 apis.yaml 未定义的接口名则退出码 1",
    )
    args = ap.parse_args()
    base = os.path.abspath(args.base_dir)
    errs = []
    if not os.path.isdir(base):
        errs.append("not a directory: %s" % base)
    cfg = os.path.join(base, "_config")
    if not os.path.isdir(cfg):
        errs.append("missing _config/")
    else:
        for name in ("apis.yaml", "global.yaml"):
            p = os.path.join(cfg, name)
            if not os.path.isfile(p):
                errs.append("missing %s" % p)
    suites = os.path.join(base, "suites")
    if not os.path.isdir(suites):
        errs.append("missing suites/")
    if errs:
        print("validate_module: FAIL\n" + "\n".join(errs))
        return 1

    prev = os.environ.get("AT_STRICT_LOAD_APIS")
    try:
        if args.strict_apis:
            os.environ["AT_STRICT_LOAD_APIS"] = "1"
        else:
            os.environ.pop("AT_STRICT_LOAD_APIS", None)
        n = len(load_case_from_yaml(base))
    except ValueError as e:
        print("validate_module: FAIL (strict API check)\n%s" % e)
        return 1
    finally:
        if prev is None:
            os.environ.pop("AT_STRICT_LOAD_APIS", None)
        else:
            os.environ["AT_STRICT_LOAD_APIS"] = prev

    hook = load_session_clean_up(base)
    hook_info = "session_clean_up registered" if hook else "no framework_hooks.session_clean_up (ok for generic modules)"
    print("validate_module: OK — loaded %d case(s); %s" % (n, hook_info))
    return 0


if __name__ == "__main__":
    sys.exit(main() or 0)
