# -*- coding: utf-8 -*-
"""
前置资源注册（按 config.ini）：
- 小模型：embedding / reranker（先 list 再 add）
- OSS 存储：storage（先 list 再 add）
- 大模型：llm（先 list，再按需 add 并返回 id）

单条用例可在 suites YAML 写 prepare_models，由 test_run 在发请求前调用。
"""
from __future__ import annotations

import socket
import time
from typing import Any, Dict, Optional, Tuple
from urllib.parse import urlparse

import requests
try:
    import paramiko
except Exception:
    paramiko = None

from common import at_env

# 同一会话内每种类型只真正请求一次（多 case 声明 prepare_models 时避免重复 list/add）
_SESSION_MODEL_TYPES: set = set()
_OSS_PORT_FORWARD_READY = False
_SESSION_LLM_MODEL_ID: Optional[str] = None


def _to_int(raw: Any, default: int) -> int:
    try:
        return int(str(raw).strip())
    except Exception:
        return default


def _to_bool(raw: Any, default: bool = False) -> bool:
    if raw is None:
        return default
    s = str(raw).strip().lower()
    if s in ("1", "true", "yes", "on"):
        return True
    if s in ("0", "false", "no", "off"):
        return False
    return default


def _is_tcp_open(host: str, port: int, timeout: float = 0.5) -> bool:
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except Exception:
        return False


def _ensure_oss_port_forward_via_ssh(config: Dict[str, Dict[str, str]]) -> bool:
    """
    通过 SSH 到后台机器执行 kubectl port-forward。
    读取 [server]：
      - host / base_url（主机）
      - ssh_port（默认 22）
      - ssh_user
      - ssh_password
    """
    global _OSS_PORT_FORWARD_READY
    if _OSS_PORT_FORWARD_READY:
        return True

    host = _server_host(config)
    if not host:
        print("[oss_setup] Missing server host, cannot setup ssh port-forward")
        return False
    if _is_tcp_open(host, 8081):
        _OSS_PORT_FORWARD_READY = True
        return True

    if paramiko is None:
        print("[oss_setup] paramiko not available, cannot setup ssh port-forward")
        return False

    srv = config.get("server") or {}
    ssh_port = _to_int(srv.get("ssh_port", "22"), 22)
    ssh_user = (srv.get("ssh_user") or "").strip()
    ssh_password = (srv.get("ssh_password") or "").strip()
    if not ssh_user or not ssh_password:
        print("[oss_setup] Missing ssh_user/ssh_password in [server], cannot setup ssh port-forward")
        return False

    remote_cmd = (
        "nohup kubectl port-forward svc/oss-gateway-backend -n kweaver "
        "--address 0.0.0.0 8081:8080 >/tmp/oss_pf.log 2>&1 &"
    )

    client = None
    try:
        client = paramiko.SSHClient()
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        client.connect(
            hostname=host,
            port=ssh_port,
            username=ssh_user,
            password=ssh_password,
            timeout=10,
        )
        client.exec_command(remote_cmd)
    except Exception as e:
        print("[oss_setup] SSH start port-forward failed: %s" % e)
        return False
    finally:
        try:
            if client is not None:
                client.close()
        except Exception:
            pass

    for _ in range(10):
        if _is_tcp_open(host, 8081, timeout=1.0):
            _OSS_PORT_FORWARD_READY = True
            return True
        time.sleep(0.5)
    print("[oss_setup] Remote port-forward not ready on %s:8081" % host)
    return False


def server_api_origin(config: Dict[str, Dict[str, str]]) -> str:
    """
    从 [server].base_url 解析 scheme://host[:port]（不含路径）。
    base_url 无效时回退 at_env.resolve_request_target（scheme + host）。
    """
    srv = config.get("server") or {}
    base = (srv.get("base_url") or "").strip()
    if base:
        if "://" not in base:
            base = "https://" + base
        p = urlparse(base)
        if p.scheme and p.netloc:
            return "%s://%s" % (p.scheme, p.netloc.rstrip("/"))
    scheme, host = at_env.resolve_request_target(config)
    host = (host or "").strip()
    if host:
        return "%s://%s" % (scheme, host.rstrip("/"))
    return ""


def mf_model_manager_urls(
    config: Dict[str, Dict[str, str]],
) -> Tuple[Optional[str], Optional[str]]:
    """mf-model-manager 的 add/list 地址。"""
    mm = config.get("model_manager") or {}
    _origin = server_api_origin(config)
    _add_default = (
        ("%s/api/mf-model-manager/v1/small-model/add" % _origin) if _origin else ""
    )
    _list_default = (
        ("%s/api/mf-model-manager/v1/small-model/list" % _origin) if _origin else ""
    )
    add_url = (mm.get("small_model_add_url") or "").strip() or _add_default
    list_url = (mm.get("small_model_list_url") or "").strip() or _list_default
    if not add_url or not list_url:
        return None, None
    return add_url, list_url


def llm_manager_urls(
    config: Dict[str, Dict[str, str]],
) -> Tuple[Optional[str], Optional[str]]:
    """mf-model-manager 的 llm add/list 地址。"""
    mm = config.get("model_manager") or {}
    _origin = server_api_origin(config)
    _add_default = ("%s/api/mf-model-manager/v1/llm/add" % _origin) if _origin else ""
    _list_default = ("%s/api/mf-model-manager/v1/llm/list" % _origin) if _origin else ""
    add_url = (mm.get("llm_add_url") or "").strip() or _add_default
    list_url = (mm.get("llm_list_url") or "").strip() or _list_default
    if not add_url or not list_url:
        return None, None
    return add_url, list_url


def _server_host(config: Dict[str, Dict[str, str]]) -> str:
    srv = config.get("server") or {}
    host = (srv.get("host") or "").strip()
    if host:
        return host
    base = (srv.get("base_url") or "").strip()
    if base:
        if "://" not in base:
            base = "https://" + base
        p = urlparse(base)
        if p.hostname:
            return p.hostname
    return ""


def storage_urls(config: Dict[str, Dict[str, str]]) -> Tuple[Optional[str], Optional[str]]:
    """
    OSS 存储 list/add 地址。
    支持 [oss_info].storage_api_base_url 覆盖，例如 http://192.168.40.62:8081
    未配置时，先通过 SSH 到后台执行 kubectl port-forward，再使用 http://<server.host>:8081/api/v1/storages
    """
    oss = config.get("oss_info") or {}
    base = (oss.get("storage_api_base_url") or "").strip()
    if not base:
        host = _server_host(config)
        if not host:
            return None, None
        if not _ensure_oss_port_forward_via_ssh(config):
            return None, None
        base = "http://%s:8081" % host
    list_url = base.rstrip("/") + "/api/v1/storages"
    add_url = list_url
    return add_url, list_url


def ensure_small_model(
    config: Dict[str, Dict[str, str]],
    model_name: str,
    payload: dict,
    log_prefix: str,
) -> None:
    """
    add 前先查 small-model/list，同名已存在则跳过。
    list 非 2xx 时仍尝试 add。
    """
    add_url, list_url = mf_model_manager_urls(config)
    if not add_url or not list_url:
        print(
            "[%s] Cannot resolve mf-model-manager URL from [server].base_url, skip"
            % log_prefix
        )
        return

    mm = config.get("model_manager") or {}
    timeout = _to_int(mm.get("timeout", "20"), 20)
    verify_ssl = _to_bool(mm.get("verify_ssl", "false"), False)

    try:
        list_resp = requests.get(list_url, timeout=timeout, verify=verify_ssl)
        if 200 <= list_resp.status_code < 300:
            try:
                body = list_resp.json()
            except Exception:
                body = {}
            candidates = []
            if isinstance(body, list):
                candidates = body
            elif isinstance(body, dict):
                for key in ("data", "items", "list", "models", "entries"):
                    val = body.get(key)
                    if isinstance(val, list):
                        candidates = val
                        break
            if any(
                isinstance(x, dict) and x.get("model_name") == model_name
                for x in candidates
            ):
                print(
                    "[%s] Model already exists, skip add: %s" % (log_prefix, model_name)
                )
                return
        else:
            print(
                "[%s] Query model list failed: status=%s body=%s"
                % (log_prefix, list_resp.status_code, list_resp.text[:1000])
            )

        resp = requests.post(
            add_url,
            headers={"Content-Type": "application/json"},
            json=payload,
            timeout=timeout,
            verify=verify_ssl,
        )
        if 200 <= resp.status_code < 300:
            print("[%s] Add success: %s" % (log_prefix, model_name))
            return
        print(
            "[%s] Add failed: status=%s body=%s"
            % (log_prefix, resp.status_code, resp.text[:1000])
        )
    except Exception as e:
        print("[%s] Request error: %s" % (log_prefix, e))


def register_embedding_model(config: Dict[str, Dict[str, str]]) -> None:
    """
    读取 [embedding_info]，向 mf-model-manager 注册 embedding 模型。
    """
    if "embedding" in _SESSION_MODEL_TYPES:
        return
    emb = config.get("embedding_info") or {}
    if not emb:
        print("[embedding_setup] Missing [embedding_info], skip add embedding")
        return

    model_name = (emb.get("embedding_model_name") or "").strip()
    model_type = (emb.get("embedding_model_type") or "").strip()
    api_url = (emb.get("embedding_api_url") or "").strip()
    api_model = (emb.get("embedding_api_model") or "").strip()
    api_key = (emb.get("embedding_api_key") or "").strip()

    if not all([model_name, model_type, api_url, api_model, api_key]):
        print("[embedding_setup] Incomplete embedding_info fields, skip add embedding")
        return

    payload = {
        "model_name": model_name,
        "model_type": model_type,
        "model_config": {
            "api_url": api_url,
            "api_model": api_model,
            "api_key": api_key,
        },
        "batch_size": _to_int(emb.get("batch_size", "10"), 10),
        "max_tokens": _to_int(emb.get("max_tokens", "8192"), 8192),
        "embedding_dim": _to_int(emb.get("embedding_dim", "1024"), 1024),
    }
    ensure_small_model(config, model_name, payload, "embedding_setup")
    _SESSION_MODEL_TYPES.add("embedding")


def register_rerank_model(config: Dict[str, Dict[str, str]]) -> None:
    """
    读取 [rerank_info]，向 mf-model-manager 注册 reranker。
    """
    if "reranker" in _SESSION_MODEL_TYPES:
        return
    rr = config.get("rerank_info") or {}
    if not rr:
        print("[rerank_setup] Missing [rerank_info], skip add reranker")
        return

    model_name = (rr.get("rerank_model_name") or "").strip()
    model_type = (rr.get("rerank_model_type") or "").strip()
    api_url = (rr.get("rerank_api_url") or "").strip()
    api_model = (rr.get("rerank_api_model") or "").strip()
    api_key = (rr.get("rerank_api_key") or "").strip()

    if not all([model_name, model_type, api_url, api_model]):
        print("[rerank_setup] Incomplete rerank_info fields, skip add reranker")
        return

    model_config = {"api_url": api_url, "api_model": api_model}
    if api_key:
        model_config["api_key"] = api_key

    payload = {
        "model_name": model_name,
        "model_type": model_type,
        "model_config": model_config,
        "batch_size": _to_int(rr.get("rerank_batch_size", "2048"), 2048),
    }
    ensure_small_model(config, model_name, payload, "rerank_setup")
    _SESSION_MODEL_TYPES.add("reranker")


def register_oss_storage(config: Dict[str, Dict[str, str]]) -> None:
    """
    读取 [oss_info]，按 storage_name 注册 OSS 存储（存在则跳过）。
    """
    if "oss" in _SESSION_MODEL_TYPES:
        return
    oss = config.get("oss_info") or {}
    if not oss:
        print("[oss_setup] Missing [oss_info], skip add storage")
        return

    storage_name = (oss.get("storage_name") or "").strip()
    vendor_type = (oss.get("vendor_type") or "").strip()
    endpoint = (oss.get("endpoint") or "").strip()
    bucket_name = (oss.get("bucket_name") or "").strip()
    access_key_id = (oss.get("access_key_id") or "").strip()
    access_key_secret = (oss.get("access_key_secret") or "").strip()
    region = (oss.get("region") or "").strip()
    internal_endpoint = (oss.get("internal_endpoint") or "").strip()
    is_default = _to_bool(oss.get("is_default"), True)

    if not all([storage_name, vendor_type, endpoint, bucket_name, access_key_id, access_key_secret]):
        print("[oss_setup] Incomplete [oss_info], skip add storage")
        return

    add_url, list_url = storage_urls(config)
    if not add_url or not list_url:
        print("[oss_setup] Cannot resolve storage API URL, skip")
        return

    payload = {
        "storage_name": storage_name,
        "vendor_type": vendor_type,
        "endpoint": endpoint,
        "bucket_name": bucket_name,
        "access_key_id": access_key_id,
        "access_key_secret": access_key_secret,
        "is_default": is_default,
        "internal_endpoint": internal_endpoint,
        "region": region,
    }

    timeout = _to_int((oss.get("timeout") or "20"), 20)
    verify_ssl = _to_bool(oss.get("verify_ssl"), False)

    try:
        list_resp = requests.get(list_url, timeout=timeout, verify=verify_ssl)
        if 200 <= list_resp.status_code < 300:
            try:
                body = list_resp.json()
            except Exception:
                body = {}
            candidates = []
            if isinstance(body, list):
                candidates = body
            elif isinstance(body, dict):
                for key in ("data", "items", "list", "storages", "entries"):
                    val = body.get(key)
                    if isinstance(val, list):
                        candidates = val
                        break
            if any(isinstance(x, dict) and x.get("storage_name") == storage_name for x in candidates):
                print("[oss_setup] Storage already exists, skip add: %s" % storage_name)
                _SESSION_MODEL_TYPES.add("oss")
                return
        else:
            print(
                "[oss_setup] Query storage list failed: status=%s body=%s"
                % (list_resp.status_code, list_resp.text[:1000])
            )

        resp = requests.post(
            add_url,
            headers={"Content-Type": "application/json"},
            json=payload,
            timeout=timeout,
            verify=verify_ssl,
        )
        if 200 <= resp.status_code < 300:
            print("[oss_setup] Add storage success: %s" % storage_name)
            _SESSION_MODEL_TYPES.add("oss")
            return
        print(
            "[oss_setup] Add storage failed: status=%s body=%s"
            % (resp.status_code, resp.text[:1000])
        )
    except Exception as e:
        print("[oss_setup] Request error: %s" % e)


def _extract_items(body: Any) -> list:
    if isinstance(body, list):
        return body
    if isinstance(body, dict):
        for key in ("data", "items", "list", "models", "entries", "records"):
            val = body.get(key)
            if isinstance(val, list):
                return val
            if isinstance(val, dict):
                for sub_key in ("items", "list", "records"):
                    sub_val = val.get(sub_key)
                    if isinstance(sub_val, list):
                        return sub_val
    return []


def _extract_id_from_body(body: Any) -> Optional[str]:
    if isinstance(body, dict):
        if body.get("id") is not None:
            return str(body.get("id"))
        for key in ("data", "item", "result"):
            sub = body.get(key)
            if isinstance(sub, dict) and sub.get("id") is not None:
                return str(sub.get("id"))
    return None


def _extract_id_from_item(item: Any) -> Optional[str]:
    if not isinstance(item, dict):
        return None
    for key in ("id", "model_id", "llm_id"):
        if item.get(key) is not None:
            return str(item.get(key))
    return None


def _find_llm_id_by_name(
    config: Dict[str, Dict[str, str]],
    model_name: str,
    model_series: str,
) -> Optional[str]:
    _, list_url = llm_manager_urls(config)
    if not list_url:
        return None

    mm = config.get("model_manager") or {}
    timeout = _to_int(mm.get("timeout", "20"), 20)
    verify_ssl = _to_bool(mm.get("verify_ssl", "false"), False)
    size = _to_int(mm.get("llm_list_size", "50"), 50)
    params = {
        "name": model_name,
        "page": 1,
        "order": "desc",
        "rule": "update_time",
        "series": model_series or "all",
        "size": size,
        "model_type": "llm",
    }
    try:
        resp = requests.get(list_url, params=params, timeout=timeout, verify=verify_ssl)
        if not (200 <= resp.status_code < 300):
            print(
                "[llm_setup] Query llm list failed: status=%s body=%s"
                % (resp.status_code, resp.text[:1000])
            )
            return None
        try:
            body = resp.json()
        except Exception:
            body = {}
        for item in _extract_items(body):
            if not isinstance(item, dict):
                continue
            item_name = str(item.get("model_name") or item.get("name") or "").strip()
            if item_name != model_name:
                continue
            _id = _extract_id_from_item(item)
            if _id:
                return _id
    except Exception as e:
        print("[llm_setup] Query llm list error: %s" % e)
    return None


def ensure_llm_model_and_get_id(config: Dict[str, Dict[str, str]]) -> Optional[str]:
    """
    根据 [llm_info] 获取/创建 llm，并返回模型 id。
    """
    global _SESSION_LLM_MODEL_ID
    if _SESSION_LLM_MODEL_ID:
        return _SESSION_LLM_MODEL_ID

    llm = config.get("llm_info") or {}
    model_name = (llm.get("llm_name") or "").strip()
    model_series = (llm.get("llm_series") or "").strip() or "all"
    max_model_len = _to_int(llm.get("llm_max_model_len", "128"), 128)
    api_model = (llm.get("llm_api_model") or "").strip()
    api_url = (llm.get("llm_api_url") or "").strip()
    api_key = (llm.get("llm_api_key") or "").strip()
    if not all([model_name, api_model, api_url]):
        print("[llm_setup] Incomplete [llm_info], skip llm prepare")
        return None

    existing_id = _find_llm_id_by_name(config, model_name, model_series)
    if existing_id:
        _SESSION_LLM_MODEL_ID = existing_id
        print("[llm_setup] LLM already exists: %s (id=%s)" % (model_name, existing_id))
        return _SESSION_LLM_MODEL_ID

    add_url, _ = llm_manager_urls(config)
    if not add_url:
        print("[llm_setup] Cannot resolve llm add URL, skip")
        return None

    mm = config.get("model_manager") or {}
    timeout = _to_int(mm.get("timeout", "20"), 20)
    verify_ssl = _to_bool(mm.get("verify_ssl", "false"), False)
    payload = {
        "model_config": {
            "api_model": api_model,
            "api_url": api_url,
            "api_key": api_key,
        },
        "model_name": model_name,
        "model_series": model_series,
        "max_model_len": max_model_len,
        "model_type": "llm",
        "quota": False,
        "change": True,
    }
    try:
        resp = requests.post(
            add_url,
            headers={"Content-Type": "application/json"},
            json=payload,
            timeout=timeout,
            verify=verify_ssl,
        )
        if not (200 <= resp.status_code < 300):
            print(
                "[llm_setup] Add llm failed: status=%s body=%s"
                % (resp.status_code, resp.text[:1000])
            )
            return None
        try:
            body = resp.json()
        except Exception:
            body = {}
        created_id = _extract_id_from_body(body)
        if created_id:
            _SESSION_LLM_MODEL_ID = created_id
            print("[llm_setup] Add llm success: %s (id=%s)" % (model_name, created_id))
            return _SESSION_LLM_MODEL_ID

        fallback_id = _find_llm_id_by_name(config, model_name, model_series)
        if fallback_id:
            _SESSION_LLM_MODEL_ID = fallback_id
            print(
                "[llm_setup] Add success and list found id: %s (id=%s)"
                % (model_name, fallback_id)
            )
            return _SESSION_LLM_MODEL_ID

        print("[llm_setup] Add success but id not found in response/list")
    except Exception as e:
        print("[llm_setup] Add llm request error: %s" % e)
    return None
