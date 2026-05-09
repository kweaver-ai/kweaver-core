#!/usr/bin/env python3
"""Patch a kubectl-style ConfigMap JSON in-place: enable + name a default small
embedding model under server.defaultSmallModelEnabled / defaultSmallModelName.

Usage:  onboard_patch_bkn_cm.py <cm-json-path> <embedding-model-name>

Tries 'yq' (mikefarah or kislyuk) first via subprocess. Falls back to PyYAML
if yq is missing or fails. Exits 2 if neither is available; exits 1 on any
other error and dumps a full traceback to stderr.

Always emits a [onboard-cm-debug] line on stderr at startup so callers can
diagnose silent exits even if the script crashes immediately afterward.

Compatible with Python 3.6+ (uses typing.List, no capture_output / PEP 563).
"""
import json
import os
import subprocess
import sys
import traceback
from typing import List

try:
    import yaml  # type: ignore[import-not-found]
    HAVE_YAML = True
    YAML_VER = getattr(yaml, "__version__", "?")
except Exception as e:  # pragma: no cover - defensive
    yaml = None
    HAVE_YAML = False
    YAML_VER = "missing (%s)" % e


def yq_subprocess_ok() -> bool:
    try:
        r = subprocess.run(
            ["yq", "--version"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=True,
            timeout=4,
        )
        return r.returncode == 0
    except Exception:
        return False


HAVE_YQ = yq_subprocess_ok()


def emit_debug(msg: str) -> None:
    print("[onboard-cm-debug] " + msg, file=sys.stderr, flush=True)


def main(argv: List[str]) -> int:
    emit_debug(
        "python=%s pyyaml=%s yq=%s argv=%r"
        % (sys.executable, YAML_VER, HAVE_YQ, argv[1:])
    )
    if len(argv) < 3:
        print("usage: %s <cm-json> <embedding-name>" % argv[0], file=sys.stderr)
        return 1
    if not HAVE_YAML and not HAVE_YQ:
        print(
            "Neither 'yq' nor PyYAML is available. Install one and re-run onboard:\n"
            "  sudo apt-get install -y python3-yaml                       # Debian/Ubuntu\n"
            "  sudo dnf install -y python3-pyyaml                         # Fedora/RHEL/openEuler\n"
            "  pip3 install --user --break-system-packages pyyaml         # any host with pip3\n"
            "  sudo curl -fsSL -o /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 \\\n"
            "    && sudo chmod +x /usr/local/bin/yq                       # mikefarah yq",
            file=sys.stderr,
        )
        return 2

    path, dname = argv[1], argv[2]
    with open(path, "r", encoding="utf-8") as f:
        j = json.load(f)
    data = j.get("data") or {}
    if not data:
        print("ConfigMap has empty data", file=sys.stderr)
        return 1

    usekey = None
    for k in data:
        if k.endswith("-config.yaml"):
            usekey = k
            break
    if not usekey:
        for k in data:
            if k.endswith(".yaml"):
                usekey = k
                break
    if not usekey:
        usekey = next(iter(data))
    emit_debug("using key=%s" % usekey)

    raw = data.get(usekey) or ""
    if not str(raw).strip():
        print("empty yaml in key %s" % usekey, file=sys.stderr)
        return 1

    newyml = None
    if HAVE_YQ:
        try:
            p = subprocess.run(
                [
                    "yq",
                    ".server.defaultSmallModelEnabled = true | "
                    '.server.defaultSmallModelName = "%s"' % dname,
                ],
                input=raw.encode("utf-8", errors="replace"),
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
            if p.returncode == 0 and p.stdout.strip():
                newyml = p.stdout.decode("utf-8", errors="replace")
            else:
                emit_debug(
                    "yq filter failed rc=%d stderr=%s"
                    % (p.returncode, p.stderr.decode(errors="replace").strip())
                )
        except Exception as e:
            emit_debug("yq invocation raised: %r" % e)

    if newyml is None:
        if not HAVE_YAML:
            print(
                "yq path failed and PyYAML not available — install python3-yaml or mikefarah yq.",
                file=sys.stderr,
            )
            return 2
        c = yaml.safe_load(raw) or {}
        c.setdefault("server", {})
        c["server"]["defaultSmallModelEnabled"] = True
        c["server"]["defaultSmallModelName"] = dname
        # Old PyYAML (e.g. on CentOS 7) has no sort_keys= — omit for compatibility.
        newyml = yaml.dump(c, default_flow_style=False, allow_unicode=True)

    j["data"][usekey] = newyml
    j.pop("status", None)
    md = j.get("metadata", {})
    if md:
        for k in list(md.keys()):
            if k in (
                "uid",
                "resourceVersion",
                "selfLink",
                "managedFields",
                "creationTimestamp",
                "generation",
                "deletionTimestamp",
            ):
                md.pop(k, None)

    out = json.dumps(j)
    emit_debug("writing %d bytes of patched JSON to stdout" % len(out))
    sys.stdout.write(out)
    sys.stdout.flush()
    emit_debug("done.")
    return 0


if __name__ == "__main__":
    try:
        rc = main(sys.argv)
    except SystemExit:
        raise
    except Exception:
        print("[onboard-cm-debug] unhandled exception:", file=sys.stderr)
        traceback.print_exc()
        sys.exit(1)
    sys.exit(rc or 0)
