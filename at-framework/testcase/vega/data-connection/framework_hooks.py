"""Session cleanup hooks for vega/data-connection module."""

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

    for key in ("items", "list", "rows", "entries", "data"):
        val = payload.get(key)
        if isinstance(val, list):
            return [x for x in val if isinstance(x, dict)]
        if isinstance(val, dict):
            for sub_key in ("items", "list", "rows", "entries"):
                sub_val = val.get(sub_key)
                if isinstance(sub_val, list):
                    return [x for x in sub_val if isinstance(x, dict)]
    return []


def _should_cleanup(name: str) -> bool:
    n = (name or "").strip()
    return n.startswith("test-opensearch-") or n.startswith("test-opensearch-updated")


def session_clean_up(config: Dict[str, Dict[str, str]], allure) -> None:
    scheme, host = at_env.resolve_request_target(config)
    base_url = f"{scheme}://{host}"
    headers = {"Content-Type": "application/json", "Accept": "application/json"}

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
        msg = f"[data-connection cleanup] list datasource failed: {exc}"
        print(msg)
        try:
            allure.attach(msg, name="data_connection_cleanup_error")
        except Exception:
            pass
        return

    items = _extract_items(payload)
    deleted = 0
    skipped = 0
    failed = 0

    for item in items:
        ds_id = str(item.get("id") or "").strip()
        ds_name = str(item.get("name") or "")
        if not ds_id:
            continue
        if not _should_cleanup(ds_name):
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
                    "[data-connection cleanup] delete failed: id=%s name=%s code=%s body=%s"
                    % (ds_id, ds_name, del_resp.status_code, del_resp.text[:300])
                )
        except Exception as exc:
            failed += 1
            print(
                "[data-connection cleanup] delete error: id=%s name=%s err=%s"
                % (ds_id, ds_name, exc)
            )

    summary = {
        "total_listed": len(items),
        "deleted": deleted,
        "skipped": skipped,
        "failed": failed,
    }
    print("[data-connection cleanup] %s" % json.dumps(summary, ensure_ascii=False))
    try:
        allure.attach(json.dumps(summary, ensure_ascii=False), name="data_connection_cleanup_summary")
    except Exception:
        pass
