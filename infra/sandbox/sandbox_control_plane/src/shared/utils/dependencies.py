"""
依赖解析工具

用于处理 Python 依赖包的格式转换和解析。
"""
import json
import re
from typing import List, Optional, Union


DEFAULT_PYTHON_PACKAGE_INDEX_URL = "https://pypi.org/simple/"


def normalize_python_package_index_url(index_url: Optional[str]) -> str:
    """规范化 Python 包仓库地址。"""
    if not index_url:
        return DEFAULT_PYTHON_PACKAGE_INDEX_URL
    return index_url.strip() or DEFAULT_PYTHON_PACKAGE_INDEX_URL


def parse_pip_spec(spec: Union[str, dict]) -> dict[str, Optional[str]]:
    """
    解析 pip requirement spec。

    返回:
        {"name": "requests", "version": "==2.31.0"}
    """
    if isinstance(spec, dict):
        return {
            "name": spec.get("name", ""),
            "version": spec.get("version") or None,
        }

    match = re.match(r"^\s*([A-Za-z0-9._-]+)\s*(.*)\s*$", spec or "")
    if not match:
        return {"name": spec.strip(), "version": None}

    name = match.group(1)
    version = match.group(2).strip() or None
    return {"name": name, "version": version}


def merge_pip_specs(existing: List[str], incoming: List[str]) -> List[str]:
    """按包名合并 pip spec，同名包以后者覆盖前者。"""
    merged: dict[str, str] = {}

    for spec in existing + incoming:
        parsed = parse_pip_spec(spec)
        if parsed["name"]:
            merged[parsed["name"].lower()] = spec

    return list(merged.values())


def parse_dependencies_to_pip_specs(dependencies: Optional[List[Union[str, dict]]]) -> List[str]:
    """
    将依赖列表转换为 pip 规范格式

    Args:
        dependencies: 依赖列表，元素可以是字符串或字典
            - 字符串格式: "requests==2.31.0" 或 "requests"
            - 字典格式: {"name": "requests", "version": "==2.31.0"}

    Returns:
        pip 规范列表，如 ["requests==2.31.0", "pandas>=2.0"]
    """
    if not dependencies:
        return []

    pip_specs = []
    for dep in dependencies:
        if isinstance(dep, dict):
            name = dep.get("name", "")
            version = dep.get("version", "")
            pip_specs.append(f"{name}{version}" if version else name)
        elif isinstance(dep, str):
            pip_specs.append(dep)

    return pip_specs


def format_dependencies_for_script(dependencies: Optional[List[Union[str, dict]]]) -> tuple[str, str]:
    """
    格式化依赖列表用于 shell 脚本

    Args:
        dependencies: 依赖列表

    Returns:
        (deps_json, deps_list) 元组
        - deps_json: JSON 字符串格式的依赖列表
        - deps_list: 空格分隔的 pip 规范字符串，用于 shell 脚本
    """
    if not dependencies:
        return "", ""

    pip_specs = parse_dependencies_to_pip_specs(dependencies)
    deps_json = json.dumps(dependencies)
    deps_list = " ".join(f'"{spec}"' for spec in pip_specs)

    return deps_json, deps_list


def build_dependency_install_script() -> str:
    """
    构建通用的 Python 依赖安装脚本片段

    Returns:
        Shell 脚本字符串，用于安装依赖到 /opt/sandbox-venv
    """
    return """
# ========== 安装 Python 依赖 ==========
echo "📦 Installing dependencies: {deps_json}"
echo "📦 Pip specs: {pip_specs}"

# 将依赖安装到容器本地文件系统（而非 S3 挂载点）
# S3 挂载点是网络文件系统，不适合作为 pip 安装目标
VENV_DIR="/opt/sandbox-venv"
mkdir -p $VENV_DIR
mkdir -p /tmp/pip-cache

echo "Installing dependencies to local filesystem: $VENV_DIR"

if pip3 install \\
    --target $VENV_DIR \\
    --cache-dir /tmp/pip-cache \\
    --no-cache-dir \\
    --no-warn-script-location \\
    --disable-pip-version-check \\
    --index-url https://pypi.org/simple/ \\
    {deps_list}; then
    echo "✅ Dependencies installed successfully to $VENV_DIR"
    # 修改属主为 sandbox 用户（gosu 切换前以 root 安装）
    chown -R sandbox:sandbox $VENV_DIR
    # 清理缓存
    rm -rf /tmp/pip-cache
else
    echo "❌ Failed to install dependencies"
    exit 1
fi
"""


def format_dependency_install_script_for_shell(dependencies: Optional[List[Union[str, dict]]]) -> str:
    """
    格式化依赖安装脚本用于 shell 执行

    Args:
        dependencies: 依赖列表

    Returns:
        Shell 脚本字符串
    """
    if not dependencies:
        return ""

    deps_json, deps_list = format_dependencies_for_script(dependencies)
    pip_specs_quoted = " ".join(f'"{spec}"' for spec in deps_list.split() if spec)

    return f"""
# ========== 安装 Python 依赖 ==========
echo "📦 Installing dependencies: {deps_json}"
echo "📦 Pip specs: {pip_specs_quoted}"

# 将依赖安装到容器本地文件系统
VENV_DIR="/opt/sandbox-venv"
mkdir -p $VENV_DIR
mkdir -p /tmp/pip-cache

echo "Installing dependencies to: $VENV_DIR"

if pip3 install \\
    --target $VENV_DIR \\
    --cache-dir /tmp/pip-cache \\
    --no-cache-dir \\
    --no-warn-script-location \\
    --disable-pip-version-check \\
    --index-url https://pypi.org/simple/ \\
    {deps_list}; then
    echo "✅ Dependencies installed successfully"
    # 清理缓存
    rm -rf /tmp/pip-cache
else
    echo "❌ Failed to install dependencies"
    exit 1
fi
"""
