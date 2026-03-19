#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
模块级钩子：框架与业务解耦，各 ADP 子模块在 testcase/<模块>/framework_hooks.py 中按需实现。
"""
import hashlib
import importlib.util
import os


def load_session_clean_up(base_dir):
    """
    若存在 testcase/<模块>/framework_hooks.py 且定义了 session_clean_up，则返回该可调用对象；否则返回 None。
    """
    base_dir = os.path.abspath(base_dir)
    path = os.path.join(base_dir, "framework_hooks.py")
    if not os.path.isfile(path):
        return None
    mod_name = "fw_hooks_%s" % hashlib.md5(base_dir.encode("utf-8")).hexdigest()[:12]
    spec = importlib.util.spec_from_file_location(mod_name, path)
    if spec is None or spec.loader is None:
        return None
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return getattr(mod, "session_clean_up", None)
