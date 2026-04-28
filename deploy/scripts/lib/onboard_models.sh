#!/usr/bin/env bash
# =============================================================================
# KWeaver onboard: model registration + BKN ConfigMap (sourced by onboard.sh)
# =============================================================================

onboard_log_info() { echo -e "${GREEN}[onboard]${NC} $*"; }
onboard_log_warn() { echo -e "${YELLOW}[onboard]${NC} $*"; }
onboard_log_err() { echo -e "${RED}[onboard]${NC} $*"; }

# ---- JSON: extract all model_name values (API list response) ----
onboard_list_model_names() {
    python3 -c '
import json, sys

def walk(o, out):
    if isinstance(o, dict):
        for k, v in o.items():
            if k == "model_name" and isinstance(v, str):
                out.add(v)
            else:
                walk(v, out)
    elif isinstance(o, list):
        for x in o:
            walk(x, out)

j = json.load(sys.stdin)
out = set()
walk(j, out)
for n in sorted(out):
    print(n)
' 2>/dev/null
}

onboard_get_existing_llm_names() {
    kweaver call "/api/mf-model-manager/v1/llm/list?page=1&size=500" 2>/dev/null | onboard_list_model_names
}

onboard_get_existing_small_model_names() {
    kweaver call "/api/mf-model-manager/v1/small-model/list?page=1&size=500" 2>/dev/null | onboard_list_model_names
}

# Read .server.defaultSmallModelName from a *-config.yaml inside the ConfigMap. Empty when not set.
onboard_bkn_cm_default_small_model_name() {
    local ns="$1" cmname="$2"
    if ! kubectl get "cm/${cmname}" -n "${ns}" &>/dev/null; then
        return 0
    fi
    local jtmp
    jtmp="$(mktemp)"
    if ! kubectl get "cm/${cmname}" -n "${ns}" -o json > "${jtmp}" 2>/dev/null; then
        rm -f "${jtmp}"
        return 0
    fi
    python3 - "${jtmp}" <<'PY' 2>/dev/null || true
import json, sys
try:
    import yaml
except ImportError:
    sys.exit(0)
with open(sys.argv[1]) as f:
    j = json.load(f)
data = j.get("data") or {}
key = next((k for k in data if k.endswith("-config.yaml")), None) \
    or next((k for k in data if k.endswith(".yaml")), None) \
    or (next(iter(data), None) if data else None)
if not key:
    sys.exit(0)
raw = data.get(key) or ""
try:
    cfg = yaml.safe_load(raw) or {}
except Exception:
    sys.exit(0)
srv = (cfg.get("server") or {})
if srv.get("defaultSmallModelEnabled") is True:
    name = srv.get("defaultSmallModelName")
    if isinstance(name, str) and name.strip():
        print(name.strip())
PY
    rm -f "${jtmp}"
}

# True iff both BKN ConfigMaps in this namespace already have defaultSmallModelEnabled+Name.
onboard_bkn_cm_already_patched() {
    local ns="${1:-kweaver}"
    local n1 n2
    n1="$(onboard_bkn_cm_default_small_model_name "${ns}" bkn-backend-cm | head -n1)"
    n2="$(onboard_bkn_cm_default_small_model_name "${ns}" ontology-query-cm | head -n1)"
    [[ -n "${n1}" && -n "${n2}" ]]
}

# Print the BKN default small-model name when both ConfigMaps agree; empty otherwise.
onboard_bkn_cm_current_default_name() {
    local ns="${1:-kweaver}"
    local n1 n2
    n1="$(onboard_bkn_cm_default_small_model_name "${ns}" bkn-backend-cm | head -n1)"
    n2="$(onboard_bkn_cm_default_small_model_name "${ns}" ontology-query-cm | head -n1)"
    if [[ -n "${n1}" && "${n1}" == "${n2}" ]]; then
        echo "${n1}"
    fi
}

# Args: model_name, model_series, max_model_len, api_key, api_model, api_url, [model_type]
onboard_ensure_llm() {
    local name="$1" series="$2" mlen="$3" akey="$4" amodel="$5" aurl="$6"
    local mtype="${7:-llm}"
    if printf '%s\n' "${_POSTI_EXISTING_LLM}" | grep -qFx "${name}"; then
        onboard_log_info "LLM already registered, skip: ${name}"
        return 0
    fi
    local body
    body=$(
        python3 -c '
import json, sys
j = {
  "model_name": sys.argv[1],
  "model_series": sys.argv[2],
  "max_model_len": int(sys.argv[3]),
  "model_type": sys.argv[7],
  "model_config": {
    "api_key": sys.argv[4],
    "api_model": sys.argv[5],
    "api_url": sys.argv[6]
  }
}
print(json.dumps(j))
' "${name}" "${series}" "${mlen}" "${akey}" "${amodel}" "${aurl}" "${mtype}" 2>/dev/null
    ) || {
        onboard_log_err "Failed to build LLM json"
        return 1
    }

    if kweaver call /api/mf-model-manager/v1/llm/add -d "${body}"; then
        onboard_log_info "Registered LLM: ${name}"
    else
        onboard_log_err "Failed to register LLM: ${name}"
        return 1
    fi
    _POSTI_EXISTING_LLM="${_POSTI_EXISTING_LLM}
${name}"
}

# Args: name, type, api_key, api_url, api_model, batch, max_tok, emb_dim
onboard_ensure_small_model() {
    local name="$1" stype="$2" akey="$3" aurl="$4" amodel="$5" batch="${6:-32}" maxtok="${7:-512}" embdim="${8:-1024}"
    if printf '%s\n' "${_POSTI_EXISTING_SM}" | grep -qFx "${name}"; then
        onboard_log_info "Small model already registered, skip: ${name}"
        return 0
    fi
    local body
    body=$(
        python3 -c '
import json, sys
j = {
  "model_name": sys.argv[1],
  "model_type": sys.argv[2],
  "model_config": {
    "api_key": sys.argv[3],
    "api_url": sys.argv[4],
    "api_model": sys.argv[5]
  },
  "batch_size": int(sys.argv[6]),
  "max_tokens": int(sys.argv[7])
}
t = j["model_type"]
if t == "embedding" and len(sys.argv) > 8 and sys.argv[8] not in ("", "0"):
    j["embedding_dim"] = int(sys.argv[8])
print(json.dumps(j))
' "${name}" "${stype}" "${akey}" "${aurl}" "${amodel}" "${batch}" "${maxtok}" "${embdim}" 2>/dev/null
    ) || {
        onboard_log_err "Failed to build small-model json"
        return 1
    }

    if kweaver call /api/mf-model-manager/v1/small-model/add -d "${body}"; then
        onboard_log_info "Registered small model: ${name} (${stype})"
    else
        onboard_log_err "Failed to register small model: ${name}"
        return 1
    fi
    _POSTI_EXISTING_SM="${_POSTI_EXISTING_SM}
${name}"
}

onboard_get_id_for_llm() {
    local mname="$1"
    kweaver call "/api/mf-model-manager/v1/llm/list?page=1&size=500" 2>/dev/null | python3 -c "
import json, sys
n = sys.argv[1]
j = json.load(sys.stdin)

def find(o):
    if isinstance(o, dict):
        if o.get('model_name') == n and o.get('model_id'):
            print(o['model_id'])
            return True
        for v in o.values():
            if find(v):
                return True
    elif isinstance(o, list):
        for x in o:
            if find(x):
                return True
    return False

find(j)
" "${mname}" 2>/dev/null | head -1
}

onboard_get_id_for_small() {
    local mname="$1"
    kweaver call "/api/mf-model-manager/v1/small-model/list?page=1&size=500" 2>/dev/null | python3 -c "
import json, sys
n = sys.argv[1]
j = json.load(sys.stdin)

def find(o):
    if isinstance(o, dict):
        if o.get('model_name') == n and o.get('model_id'):
            print(o['model_id'])
            return True
        for v in o.values():
            if find(v):
                return True
    elif isinstance(o, list):
        for x in o:
            if find(x):
                return True
    return False

find(j)
" "${mname}" 2>/dev/null | head -1
}

onboard_test_llm() {
    local mid="$1"
    [[ -z "${mid}" ]] && return 0
    if kweaver call /api/mf-model-manager/v1/llm/test -d "{\"model_id\": \"${mid}\"}" 2>/dev/null; then
        onboard_log_info "LLM test ok for id ${mid}"
    else
        onboard_log_warn "LLM test failed for id ${mid} (check upstream / network)"
    fi
}

onboard_test_small() {
    local mid="$1"
    [[ -z "${mid}" ]] && return 0
    if kweaver call /api/mf-model-manager/v1/small-model/test -d "{\"model_id\": \"${mid}\"}" 2>/dev/null; then
        onboard_log_info "Small model test ok for id ${mid}"
    else
        onboard_log_warn "Small model test failed for id ${mid}"
    fi
}

# Make sure the embedded ConfigMap-patch python script can run: it needs either
# yq (mikefarah) on PATH, or PyYAML importable. On Ubuntu 24.04 / RHEL 9 /
# openEuler 23 hosts neither is present out of the box. Try the cheapest
# install paths in order:
#   1) pip3 --user (no sudo required)
#   2) sudo apt-get install python3-yaml
#   3) sudo dnf/yum install python3-pyyaml
# If everything fails, return non-zero and let the caller print copy-paste
# install commands. Idempotent — short-circuits when the deps are already there.
onboard_ensure_yaml_dep() {
    if command -v yq >/dev/null 2>&1; then
        return 0
    fi
    if command -v python3 >/dev/null 2>&1 && python3 -c 'import yaml' 2>/dev/null; then
        return 0
    fi

    onboard_log_warn "Neither 'yq' nor PyYAML found — the BKN ConfigMap patch needs one of them. Trying to install PyYAML automatically..."

    if command -v pip3 >/dev/null 2>&1; then
        if pip3 install --user --break-system-packages pyyaml >/dev/null 2>&1 \
            || pip3 install --user pyyaml >/dev/null 2>&1; then
            if python3 -c 'import yaml' 2>/dev/null; then
                onboard_log_info "Installed PyYAML via pip3 --user"
                return 0
            fi
        fi
    fi

    if command -v sudo >/dev/null 2>&1; then
        if command -v apt-get >/dev/null 2>&1; then
            onboard_log_info "Trying: sudo apt-get install -y python3-yaml (may prompt for sudo password)"
            if sudo -n apt-get install -y python3-yaml >/dev/null 2>&1 \
                || sudo apt-get install -y python3-yaml; then
                python3 -c 'import yaml' 2>/dev/null && {
                    onboard_log_info "Installed python3-yaml via apt"
                    return 0
                }
            fi
        elif command -v dnf >/dev/null 2>&1; then
            onboard_log_info "Trying: sudo dnf install -y python3-pyyaml"
            if sudo -n dnf install -y python3-pyyaml >/dev/null 2>&1 \
                || sudo dnf install -y python3-pyyaml; then
                python3 -c 'import yaml' 2>/dev/null && {
                    onboard_log_info "Installed python3-pyyaml via dnf"
                    return 0
                }
            fi
        elif command -v yum >/dev/null 2>&1; then
            onboard_log_info "Trying: sudo yum install -y python3-pyyaml"
            if sudo -n yum install -y python3-pyyaml >/dev/null 2>&1 \
                || sudo yum install -y python3-pyyaml; then
                python3 -c 'import yaml' 2>/dev/null && {
                    onboard_log_info "Installed python3-pyyaml via yum"
                    return 0
                }
            fi
        fi
    fi

    onboard_log_err "Could not auto-install PyYAML or yq. Install one manually and re-run onboard:"
    onboard_log_err "  sudo apt-get install -y python3-yaml                       # Debian/Ubuntu"
    onboard_log_err "  sudo dnf install -y python3-pyyaml                         # Fedora/RHEL/openEuler"
    onboard_log_err "  pip3 install --user --break-system-packages pyyaml         # any host with pip3"
    onboard_log_err "  sudo curl -fsSL -o /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && sudo chmod +x /usr/local/bin/yq"
    return 1
}

# Apply embedded YAML in ConfigMap (bkn-backend-cm, ontology-query-cm); see helm *-config.yaml keys
onboard_upsert_cm_embedded_yaml() {
    local ns="$1" cmname="$2" dname="$3" # ymlkey optional 4th not needed — auto-detect

    if ! kubectl get "cm/${cmname}" -n "${ns}" &>/dev/null; then
        onboard_log_err "ConfigMap ${cmname} not found in ${ns}"
        return 1
    fi

    onboard_ensure_yaml_dep || return 1

    # Belt-and-suspenders: confirm with THIS python3 that PyYAML imports, or
    # that mikefarah-style yq is on PATH. (apt's `yq` on Ubuntu is the Python
    # `python-yq` wrapper around jq — its filter syntax is jq, which is also
    # accepted by our filter, so either is fine.)
    if ! python3 -c 'import yaml' 2>/dev/null && ! command -v yq >/dev/null 2>&1; then
        onboard_log_err "PyYAML still not importable by $(command -v python3) and 'yq' not on PATH after auto-install. python3 -c 'import yaml' must succeed (or install yq). Sometimes pip3 installs into a different python3 than the one on PATH; try: python3 -m pip install --user --break-system-packages pyyaml"
        return 1
    fi

    local jtmp errtmp rc
    jtmp="$(mktemp)"
    errtmp="$(mktemp)"
    kubectl get "cm/${cmname}" -n "${ns}" -o json > "${jtmp}" || {
        rm -f "${jtmp}" "${errtmp}"
        return 1
    }

    if ! OUT_JSON=$(
        python3 - "${jtmp}" "${dname}" 2> "${errtmp}" <<'PY'
import json, subprocess, sys, traceback

try:
    import yaml
    HAVE_YAML = True
    YAML_VER = getattr(yaml, "__version__", "?")
except Exception as e:
    HAVE_YAML = False
    YAML_VER = "missing (%s)" % e


def yq_subprocess_ok():
    try:
        r = subprocess.run(
            ["yq", "--version"],
            capture_output=True,
            check=True,
            timeout=4,
        )
        return r.returncode == 0
    except Exception:
        return False


HAVE_YQ = yq_subprocess_ok()
print(
    "[onboard-cm-debug] python=%s pyyaml=%s yq=%s argv=%r"
    % (sys.executable, YAML_VER, HAVE_YQ, sys.argv[1:]),
    file=sys.stderr,
)
try:
    pass
except Exception:
    pass
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
    sys.exit(2)


try:
    path, dname = sys.argv[1], sys.argv[2]
    with open(path) as f:
        j = json.load(f)
    data = j.get("data") or {}
    if not data:
        print("ConfigMap has empty data", file=sys.stderr)
        sys.exit(1)

    # choose *-config.yaml
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
        usekey = list(data.keys())[0]
    print("[onboard-cm-debug] using key=%s" % usekey, file=sys.stderr)

    raw = data.get(usekey) or ""
    if not str(raw).strip():
        print("empty yaml in key", usekey, file=sys.stderr)
        sys.exit(1)

    newyml = None
    if HAVE_YQ:
        # Detect mikefarah/yq vs python-yq (kislyuk). mikefarah supports
        # `.foo = "bar"`; python-yq is jq-based and accepts the same filter.
        # If yq fails for ANY reason, fall through to PyYAML when available.
        try:
            p = subprocess.run(
                [
                    "yq",
                    ".server.defaultSmallModelEnabled = true | "
                    ".server.defaultSmallModelName = \"%s\"" % dname,
                ],
                input=raw.encode("utf-8", errors="replace"),
                capture_output=True,
            )
            if p.returncode == 0 and p.stdout.strip():
                newyml = p.stdout.decode("utf-8", errors="replace")
            else:
                print(
                    "[onboard-cm-debug] yq filter failed rc=%d stderr=%s"
                    % (p.returncode, p.stderr.decode(errors="replace").strip()),
                    file=sys.stderr,
                )
        except Exception as e:
            print("[onboard-cm-debug] yq invocation raised: %r" % e, file=sys.stderr)

    if newyml is None:
        if not HAVE_YAML:
            print(
                "yq path failed and PyYAML not available — install python3-yaml or mikefarah yq.",
                file=sys.stderr,
            )
            sys.exit(2)
        c = yaml.safe_load(raw) or {}
        c.setdefault("server", {})
        c["server"]["defaultSmallModelEnabled"] = True
        c["server"]["defaultSmallModelName"] = dname
        newyml = yaml.dump(
            c, default_flow_style=False, allow_unicode=True, sort_keys=False
        )

    j["data"][usekey] = newyml
    j.pop("status", None)
    md = j.get("metadata", {})
    if md:
        for k in list(md.keys()):
            if k in (
                "uid", "resourceVersion", "selfLink", "managedFields",
                "creationTimestamp", "generation", "deletionTimestamp",
            ):
                try:
                    del md[k]
                except KeyError:
                    pass
    print(json.dumps(j))
except SystemExit:
    raise
except Exception:
    print("[onboard-cm-debug] unhandled exception:", file=sys.stderr)
    traceback.print_exc()
    sys.exit(1)
PY
    ); then
        rm -f "${jtmp}" "${errtmp}"
    else
        rc=$?
        local err_body=""
        [[ -s "${errtmp}" ]] && err_body="$(cat "${errtmp}")"
        # Keep both inputs around for post-mortem on failure.
        local keep_json="/tmp/onboard-bkn-cm-${cmname}.json"
        local keep_err="/tmp/onboard-bkn-cm-${cmname}.stderr"
        cp -f "${jtmp}" "${keep_json}" 2>/dev/null || true
        cp -f "${errtmp}" "${keep_err}" 2>/dev/null || true
        rm -f "${jtmp}" "${errtmp}"
        if [[ "${rc}" -eq 2 ]]; then
            onboard_log_err "Failed to build patched ConfigMap JSON for ${cmname}: yq or PyYAML required"
        else
            onboard_log_err "Failed to build patched ConfigMap JSON for ${cmname} (python exit=${rc})"
        fi
        onboard_log_err "  python3=$(command -v python3) yq=$(command -v yq 2>/dev/null || echo none)"
        onboard_log_err "  PyYAML import: $(python3 -c 'import yaml; print(yaml.__version__)' 2>&1 | head -n1)"
        onboard_log_err "  Saved ConfigMap JSON: ${keep_json}"
        onboard_log_err "  Saved python stderr:  ${keep_err}"
        if [[ -n "${err_body}" ]]; then
            while IFS= read -r _line; do
                onboard_log_err "  ${_line}"
            done <<< "${err_body}"
        else
            onboard_log_err "  (python wrote nothing to stderr — see saved files above; consider re-running:  python3 ${SCRIPT_DIR:-deploy/scripts/lib}/onboard_apply_config.py  or inspect ${keep_json})"
        fi
        return "${rc}"
    fi

    echo "${OUT_JSON}" | kubectl apply -f - || return 1
    onboard_log_info "Applied ${cmname}: defaultSmallModelName=${dname}"
    return 0
}

# Restart BKN / ontology-query after ConfigMap change
onboard_bkn_rollout() {
    local ns="$1"
    kubectl rollout restart "deployment/bkn-backend" -n "${ns}" 2>/dev/null || onboard_log_warn "deployment/bkn-backend missing or not restartable"
    kubectl rollout restart "deployment/ontology-query" -n "${ns}" 2>/dev/null || onboard_log_warn "deployment/ontology-query missing or not restartable"
    kubectl rollout status "deployment/bkn-backend" -n "${ns}" --timeout=300s 2>/dev/null || true
    kubectl rollout status "deployment/ontology-query" -n "${ns}" --timeout=300s 2>/dev/null || true
    onboard_log_info "Rollout signalled for bkn-backend and ontology-query"
}
