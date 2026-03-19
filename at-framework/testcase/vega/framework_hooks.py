#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
VEGA 调试模块：会话级数据清理（数据源 + 原子视图）。
其他 ADP 模块请在本模块目录下实现自己的 framework_hooks.session_clean_up，或留空不实现。
"""
import requests


def session_clean_up(config, allure):
    """
    :param config: load_sys_config 结果，含 env / external 等 section
    :param allure: allure 模块，用于 attach 失败信息
    """
    env = config.get("env", {})
    ext = config.get("external", {})
    host = (env.get("host") or "").strip()
    token = (ext.get("token") or "").strip()
    scheme = (env.get("request_scheme") or "https").strip().rstrip(":/") or "https"
    if not host or not token:
        print("framework_hooks: skip clean_up (missing host or token)")
        return

    base = "%s://%s" % (scheme, host)
    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json",
        "Authorization": "Bearer %s" % token,
    }

    print("VEGA session_clean_up: clearing non-built-in datasources @ %s" % base)

    try:
        r = requests.get("%s/api/data-connection/v1/datasource" % base, verify=False, headers=headers, timeout=120)
    except requests.RequestException as e:
        print("查询数据源列表失败: %s" % e)
        allure.attach(body=str(e), name="清理历史数据失败")
        return

    if r.status_code > 200:
        print("查询数据源列表失败 HTTP %s" % r.status_code)
        try:
            allure.attach(body=r.text, name="清理历史数据失败")
        except Exception:
            pass
        return

    try:
        entries = r.json().get("entries") or []
    except Exception:
        entries = []

    for x in entries:
        if x.get("is_built_in"):
            continue
        ds_id = x.get("id")
        name = x.get("name", ds_id)
        try:
            rv = requests.get(
                "%s/api/mdl-data-model/v1/data-views" % base,
                params={"data_source_id": ds_id, "type": "atomic"},
                verify=False,
                headers=headers,
                timeout=120,
            )
        except requests.RequestException as e:
            print("查询视图失败 @ %s: %s" % (name, e))
            allure.attach(body=str(e), name="清理历史数据失败")
            return
        if rv.status_code > 200:
            print("查询视图失败 @ %s" % name)
            try:
                allure.attach(body=rv.text, name="清理历史数据失败")
            except Exception:
                pass
            return
        try:
            for y in (rv.json().get("entries") or []):
                vid = y.get("id")
                if not vid:
                    continue
                dr = requests.delete(
                    "%s/api/mdl-data-model/v1/data-views/%s" % (base, vid),
                    verify=False,
                    headers=headers,
                    timeout=120,
                )
                if dr.status_code > 300:
                    print("删除视图失败 @ %s" % name)
                    try:
                        allure.attach(body=dr.text, name="清理历史数据失败")
                    except Exception:
                        pass
                    return
        except Exception:
            pass

        try:
            dd = requests.delete(
                "%s/api/data-connection/v1/datasource/%s" % (base, ds_id),
                verify=False,
                headers=headers,
                timeout=120,
            )
        except requests.RequestException as e:
            print("删除数据源失败 @ %s: %s" % (name, e))
            allure.attach(body=str(e), name="清理历史数据失败")
            return
        if dd.status_code > 200:
            print("删除数据源失败 @ %s HTTP %s" % (name, dd.status_code))
            try:
                allure.attach(body=dd.text, name="清理历史数据失败")
            except Exception:
                pass
