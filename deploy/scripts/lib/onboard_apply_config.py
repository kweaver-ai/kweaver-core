#!/usr/bin/env python3
"""
Load deploy/conf models YAML (env ${VAR} expansion), register models via kweaver, optionally BKN.
Invoked from onboard.sh: python3 onboard_apply_config.py <yaml_path> <namespace> <skip_bkn: true/false>
"""
from __future__ import annotations

import json
import os
import re
import shlex
import subprocess
import sys
from typing import Any, Optional

try:
    import yaml
except ImportError as e:
    print("pip3 install pyyaml", file=sys.stderr)
    raise e


def expand_env(s: str) -> str:
    return re.sub(
        r"\$\{([^}]+)\}", lambda m: os.environ.get(m.group(1), ""), s
    )


def kweaver_call(*args: str) -> int:
    return subprocess.call(["kweaver", *args], stdin=subprocess.DEVNULL)


def jcall(path: str, body: dict) -> int:
    return kweaver_call("call", path, "-d", json.dumps(body, ensure_ascii=False))


def get_json(path: str) -> Any:
    p = subprocess.run(
        ["kweaver", "call", path],
        check=False,
        capture_output=True,
        text=True,
    )
    if p.returncode != 0:
        print(p.stderr, file=sys.stderr)
        return None
    raw = (p.stdout or "").strip()
    if not raw:
        return None
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        print(raw[:500], file=sys.stderr)
        return None


def print_completion_report_config_yaml(namespace: str) -> None:
    """Emit the same summary as deploy/scripts/lib/onboard_report.sh after successful --config run."""
    here = os.path.dirname(os.path.abspath(__file__))
    rpt = os.path.join(here, "onboard_report.sh")
    if not os.path.isfile(rpt):
        return
    env = os.environ.copy()
    env["NAMESPACE"] = namespace
    env["ONBOARD_REPORT_MAIN_MODE"] = "config-yaml"
    env["ONBOARD_REPORT_ISF_TEST_USER"] = (
        "not run: --config mode has no ISF test-user wizard"
    )
    env["ONBOARD_REPORT_CONTEXT_LOADER"] = (
        "not run: no probe step; run deploy/onboard.sh separately for ADP toolbox import"
    )
    subprocess.run(
        [
            "bash",
            "-c",
            f"source {shlex.quote(rpt)} && onboard_print_completion_report",
        ],
        env=env,
        check=False,
    )


def find_model_id(payload: Any, name: str) -> str:
    """Find first model_id for the given model_name in a list response tree."""

    def walk(o) -> Optional[str]:
        if isinstance(o, dict):
            if o.get("model_name") == name and o.get("model_id"):
                return o["model_id"]
            for v in o.values():
                r = walk(v)
                if r:
                    return r
        elif isinstance(o, list):
            for x in o:
                r = walk(x)
                if r:
                    return r
        return None

    return walk(payload) or ""


def main() -> int:
    if len(sys.argv) < 4:
        print(
            "usage: onboard_apply_config.py <config.yaml> <namespace> <skip_bkn>",
            file=sys.stderr,
        )
        return 1

    path, namespace, skip_bkn = sys.argv[1], sys.argv[2], sys.argv[3] == "true"
    with open(path, encoding="utf-8", errors="replace") as f:
        text = expand_env(f.read())
    cfg = yaml.safe_load(text) or {}
    if isinstance(cfg.get("namespace"), str) and cfg.get("namespace", "").strip():
        namespace = cfg["namespace"].strip()

    llm_list = list(cfg.get("llm") or [])
    emb_list = list(cfg.get("embedding") or [])
    rrk_list = list(cfg.get("reranker") or [])

    bkn_default: str = ""

    # --- register LLM ---
    for e in llm_list:
        mname = e["model_name"]
        j = get_json("/api/mf-model-manager/v1/llm/list?page=1&size=500")
        if j and find_model_id(j, mname):
            print(f"[onboard] LLM already exists, skip: {mname}")
            continue
        body = {
            "model_name": mname,
            "model_series": e.get("model_series", "others"),
            "max_model_len": int(e.get("max_model_len", 8192)),
            "model_type": e.get("model_type", "llm"),
            "model_config": {
                "api_key": e["api_key"],
                "api_model": e["api_model"],
                "api_url": e["api_url"],
            },
        }
        r = jcall("/api/mf-model-manager/v1/llm/add", body)
        if r != 0:
            return r
        print(f"[onboard] Registered LLM: {mname}")

    def register_small(mname: str, mtype: str, e: dict) -> int:
        j = get_json(
            "/api/mf-model-manager/v1/small-model/list?page=1&size=500"
        )
        if j and find_model_id(j, mname):
            print(f"[onboard] small model exists, skip: {mname}")
            return 0
        sm: dict = {
            "model_name": mname,
            "model_type": mtype,
            "model_config": {
                "api_key": e["api_key"],
                "api_url": e["api_url"],
                "api_model": e["api_model"],
            },
            "batch_size": int(e.get("batch_size", 32)),
            "max_tokens": int(e.get("max_tokens", 512)),
        }
        if mtype == "embedding":
            sm["embedding_dim"] = int(e.get("embedding_dim", 1024))
        r = jcall("/api/mf-model-manager/v1/small-model/add", sm)
        if r == 0:
            print(f"[onboard] Registered small: {mname} ({mtype})")
        return r

    # --- register embedding ---
    for e in emb_list:
        mname = e["model_name"]
        mtype = e.get("model_type") or "embedding"
        r = register_small(mname, mtype, e)
        if r != 0:
            return r
        if mtype == "embedding" and e.get("set_as_bkn_default"):
            bkn_default = mname

    # --- register reranker ---
    for e in rrk_list:
        mname = e["model_name"]
        mtype = e.get("model_type") or "reranker"
        r = register_small(mname, mtype, e)
        if r != 0:
            return r

    if not bkn_default and emb_list:
        bkn_default = emb_list[0]["model_name"]

    # tests
    for e in llm_list:
        mid = find_model_id(
            get_json("/api/mf-model-manager/v1/llm/list?page=1&size=500"),
            e["model_name"],
        )
        if mid:
            jcall(
                "/api/mf-model-manager/v1/llm/test",
                {"model_id": str(mid)},
            )
    for e in emb_list:
        mid = find_model_id(
            get_json(
                "/api/mf-model-manager/v1/small-model/list?page=1&size=500"
            ),
            e["model_name"],
        )
        if mid:
            jcall(
                "/api/mf-model-manager/v1/small-model/test",
                {"model_id": str(mid)},
            )

    if skip_bkn or not bkn_default:
        print(
            f"[onboard] skip BKN (skip_bkn={skip_bkn}, default='{bkn_default}')"
        )
        print_completion_report_config_yaml(namespace)
        return 0

    r = patch_bkn_cms_and_rollout(namespace, bkn_default)
    if r == 0:
        print_completion_report_config_yaml(namespace)
    return r


def patch_bkn_cms_and_rollout(namespace: str, dname: str) -> int:
    """Set server.defaultSmallModel* in *-config.yaml in bkn-backend-cm and ontology-query-cm."""
    for cm in ("bkn-backend-cm", "ontology-query-cm"):
        r = _patch_one_cm(namespace, cm, dname)
        if r != 0:
            return r
    for dep in ("bkn-backend", "ontology-query"):
        p = subprocess.run(
            [
                "kubectl",
                "rollout",
                "restart",
                f"deployment/{dep}",
                "-n",
                namespace,
            ],
            capture_output=True,
            text=True,
        )
        if p.returncode != 0:
            print(f"[onboard] rollout restart {dep}: {p.stderr}", file=sys.stderr)
        else:
            print(f"[onboard] rolled out {dep}")
    subprocess.call(
        [
            "kubectl",
            "rollout",
            "status",
            f"deployment/bkn-backend",
            "-n",
            namespace,
            "--timeout=300s",
        ]
    )
    subprocess.call(
        [
            "kubectl",
            "rollout",
            "status",
            f"deployment/ontology-query",
            "-n",
            namespace,
            "--timeout=300s",
        ]
    )
    return 0


def _patch_one_cm(namespace: str, name: str, dname: str) -> int:
    p = subprocess.run(
        ["kubectl", "get", f"cm/{name}", "-n", namespace, "-o", "json"],
        capture_output=True,
        text=True,
    )
    if p.returncode != 0:
        print(p.stderr, file=sys.stderr)
        return p.returncode
    j = json.loads(p.stdout)
    data = j.get("data") or {}
    if not data:
        print(f"[onboard] {name}: empty data", file=sys.stderr)
        return 1
    usekey = next(
        (k for k in data if k.endswith("-config.yaml")),
        next((k for k in data if k.endswith(".yaml")), list(data.keys())[0]),
    )
    raw = data.get(usekey) or ""
    if not str(raw).strip():
        return 1
    c = yaml.safe_load(raw) or {}
    c.setdefault("server", {})
    c["server"]["defaultSmallModelEnabled"] = True
    c["server"]["defaultSmallModelName"] = dname
    newyml = yaml.dump(
        c, default_flow_style=False, allow_unicode=True, sort_keys=False
    )
    j["data"][usekey] = newyml
    if "metadata" in j:
        md = j["metadata"]
        for k in "uid", "resourceVersion", "selfLink", "managedFields", "creationTimestamp":
            md.pop(k, None)
    j.pop("status", None)
    pr = subprocess.run(
        ["kubectl", "apply", "-f", "-"],
        input=json.dumps(j, ensure_ascii=False).encode("utf-8"),
        capture_output=True,
    )
    if pr.returncode != 0:
        print(pr.stderr.decode(), file=sys.stderr)
    else:
        print(f"[onboard] Patched {name} for defaultSmallModelName={dname}")
    return pr.returncode


if __name__ == "__main__":
    sys.exit(main())
