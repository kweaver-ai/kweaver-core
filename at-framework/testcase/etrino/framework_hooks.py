"""Session cleanup hooks for etrino module."""

from __future__ import annotations

import json
from typing import Any, Dict, List

import requests
import urllib3

from common import at_env


urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


def _extract_items(payload: Any) -> List[Dict[str, Any]]:
    if isinstance(payload, list):
        return [x for x in payload if isinstance(x, dict)]
    if not isinstance(payload, dict):
        return []
    for key in ("items", "list", "rows", "data"):
        val = payload.get(key)
        if isinstance(val, list):
            return [x for x in val if isinstance(x, dict)]
        if isinstance(val, dict):
            for sub_key in ("items", "list", "rows"):
                sub_val = val.get(sub_key)
                if isinstance(sub_val, list):
                    return [x for x in sub_val if isinstance(x, dict)]
    return []


def _is_test_created_datasource(name: str) -> bool:
    # Keep cleanup conservative: only delete entries created by AT suites.
    return "AutoCase" in (name or "")


def session_clean_up(config: Dict[str, Dict[str, str]], allure) -> None:
    scheme, host = at_env.resolve_request_target(config)
    base_url = f"{scheme}://{host}"
    headers = {"Content-Type": "application/json", "Accept": "application/json"}

    built_in_id = (config.get("global") or {}).get("buildin_catalog_id", "").strip()

    try:
        resp = requests.get(
            f"{base_url}/api/data-connection/v1/datasource",
            params={"limit": -1, "offset": 0},
            headers=headers,
            timeout=20,
            verify=False,
        )
        payload = resp.json()
    except Exception as exc:
        msg = f"[etrino cleanup] query datasource list failed: {exc}"
        print(msg)
        try:
            allure.attach(msg, name="etrino_cleanup_error")
        except Exception:
            pass
        return

    items = _extract_items(payload)
    deleted, skipped, failed = 0, 0, 0

    for item in items:
        ds_id = str(item.get("id") or "").strip()
        name = str(item.get("name") or "")
        if not ds_id:
            continue
        if built_in_id and ds_id == built_in_id:
            skipped += 1
            continue
        if not _is_test_created_datasource(name):
            skipped += 1
            continue

        try:
            del_resp = requests.delete(
                f"{base_url}/api/data-connection/v1/datasource/{ds_id}",
                headers=headers,
                timeout=20,
                verify=False,
            )
            if del_resp.status_code in (200, 204):
                deleted += 1
            else:
                failed += 1
                print(
                    "[etrino cleanup] delete failed: id=%s name=%s code=%s body=%s"
                    % (ds_id, name, del_resp.status_code, del_resp.text[:300])
                )
        except Exception as exc:
            failed += 1
            print("[etrino cleanup] delete error: id=%s name=%s err=%s" % (ds_id, name, exc))

    summary = {
        "total_listed": len(items),
        "deleted": deleted,
        "skipped": skipped,
        "failed": failed,
    }
    print("[etrino cleanup] %s" % json.dumps(summary, ensure_ascii=False))
    try:
        allure.attach(json.dumps(summary, ensure_ascii=False), name="etrino_cleanup_summary")
    except Exception:
        pass
