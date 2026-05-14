#!/usr/bin/env bash
set -euo pipefail  # 设置严格模式：遇到错误立即退出，未定义变量报错，管道命令中任何命令失败都返回失败

EXAMPLES_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"  # 获取示例根目录的绝对路径
EXAMPLES_ENV_FILE="$EXAMPLES_ROOT/.env"  # 设置环境变量配置文件路径
EXAMPLES_STATE_DIR="$EXAMPLES_ROOT/.tmp"  # 设置临时状态目录路径
EXAMPLES_STATE_FILE="$EXAMPLES_STATE_DIR/state.env"  # 设置状态文件路径

# 读取简单的 KEY=VALUE 文件；已在当前 shell 中设置的变量优先级更高。
examples_load_kv_file() {  # 定义函数：加载键值对文件
  local file="$1"  # 获取第一个参数：文件路径
  [[ -f "$file" ]] || return 0  # 如果文件不存在则直接返回

  while IFS= read -r line || [[ -n "$line" ]]; do  # 逐行读取文件内容
    line="${line%$'\r'}"  # 移除Windows风格的回车符
    [[ "$line" =~ ^[[:space:]]*$ ]] && continue  # 跳过空行
    [[ "$line" =~ ^[[:space:]]*# ]] && continue  # 跳过注释行
    [[ "$line" =~ ^[[:space:]]*([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]] || continue  # 匹配 KEY=VALUE 格式

    local key="${BASH_REMATCH[1]}"  # 提取键名
    local value="${BASH_REMATCH[2]}"  # 提取值
    value="${value%%[[:space:]]#*}"  # 移除值后面的注释
    value="${value#"${value%%[![:space:]]*}"}"  # 移除值前导空格
    value="${value%"${value##*[![:space:]]}"}"  # 移除值尾部空格

    if [[ "$value" == \"*\" && "$value" == *\" ]]; then  # 如果值被双引号包围
      value="${value:1:${#value}-2}"  # 移除双引号
    elif [[ "$value" == \'*\' && "$value" == *\' ]]; then  # 如果值被单引号包围
      value="${value:1:${#value}-2}"  # 移除单引号
    fi

    if [[ -z "${!key:-}" ]]; then  # 如果键尚未设置
      export "$key=$value"  # 导出环境变量
    fi
  done < "$file"  # 从文件读取
}

# 加载示例环境，并补齐本地 no-auth 场景的默认值。
examples_load_env() {  # 定义函数：加载示例环境
  mkdir -p "$EXAMPLES_STATE_DIR"  # 创建状态目录
  examples_load_kv_file "$EXAMPLES_ENV_FILE"  # 加载环境变量配置文件
  examples_load_kv_file "$EXAMPLES_STATE_FILE"  # 加载状态文件

  export KWEAVER_BASE_URL="${KWEAVER_BASE_URL:-http://127.0.0.1:13020}"  # 设置 kweaver 基础 URL，默认本地地址
  export KWEAVER_NO_AUTH="${KWEAVER_NO_AUTH:-1}"  # 设置是否跳过认证，默认为1（跳过）
  export KWEAVER_BUSINESS_DOMAIN="${KWEAVER_BUSINESS_DOMAIN:-bd_public}"  # 设置业务域名，默认为 bd_public
  export KWEAVER_TOKEN="${KWEAVER_TOKEN:-}"  # 设置认证令牌，默认为空
  export KWEAVER_SDK_PACKAGE_SOURCE="${KWEAVER_SDK_PACKAGE_SOURCE:-remote}"  # 设置 SDK 包来源，默认为远程 npm 包
  export KWEAVER_SDK_REMOTE_PACKAGE="${KWEAVER_SDK_REMOTE_PACKAGE:-@kweaver-ai/kweaver-sdk}"  # 设置远程 npm 包名
  export KWEAVER_SDK_TS_DIR="${KWEAVER_SDK_TS_DIR:-../../../../../kweaver-sdk/packages/typescript}"  # 设置 TypeScript SDK 目录
  export KWEAVER_LLM_ID="${KWEAVER_LLM_ID:-xxx}"  # 设置 LLM ID，默认为 xxx
  export KWEAVER_LLM_NAME="${KWEAVER_LLM_NAME:-deepseek-v3}"  # 设置 LLM 名称，默认为 deepseek-v3
  export AGENT_VERSION="${AGENT_VERSION:-v0}"  # 设置 Agent 版本，默认为 v0
}

# 把脚本运行过程中产生的状态写入 .tmp/state.env，供后续示例复用。
examples_state_set() {  # 定义函数：设置状态值
  local key="$1"  # 获取第一个参数：键名
  local value="$2"  # 获取第二个参数：值
  mkdir -p "$EXAMPLES_STATE_DIR"  # 创建状态目录
  local tmp_file  # 声明临时文件变量
  tmp_file="$(mktemp "$EXAMPLES_STATE_DIR/state.XXXXXX")"  # 创建临时文件
  if [[ -f "$EXAMPLES_STATE_FILE" ]]; then  # 如果状态文件存在
    grep -v -E "^${key}=" "$EXAMPLES_STATE_FILE" > "$tmp_file" || true  # 移除旧的同名键值
  fi
  printf '%s=%s\n' "$key" "$value" >> "$tmp_file"  # 追加新的键值对
  mv "$tmp_file" "$EXAMPLES_STATE_FILE"  # 替换原状态文件
  export "$key=$value"  # 导出环境变量
}

# 从 .tmp/state.env 中清理指定状态，常用于删除临时 Agent 后收口。
examples_state_clear() {  # 定义函数：清除状态值
  mkdir -p "$EXAMPLES_STATE_DIR"  # 创建状态目录
  local tmp_file  # 声明临时文件变量
  tmp_file="$(mktemp "$EXAMPLES_STATE_DIR/state.XXXXXX")"  # 创建临时文件
  if [[ -f "$EXAMPLES_STATE_FILE" ]]; then  # 如果状态文件存在
    grep -v -E "^($(IFS='|'; echo "$*"))=" "$EXAMPLES_STATE_FILE" > "$tmp_file" || true  # 移除指定的键值对
  fi
  mv "$tmp_file" "$EXAMPLES_STATE_FILE"  # 替换原状态文件
  local key  # 声明键变量
  for key in "$@"; do  # 遍历所有参数键
    unset "$key" || true  # 取消设置环境变量
  done
}

# 校验必需变量；缺少时返回非 0，避免示例“未执行但成功”。
examples_require_env() {  # 定义函数：要求环境变量存在
  local name="$1"  # 获取第一个参数：变量名
  local message="$2"  # 获取第二个参数：错误信息
  if [[ -z "${!name:-}" ]]; then  # 如果变量未设置或为空
    echo "Error: $message" >&2  # 输出错误信息到标准错误
    return 1  # 返回错误码1
  fi
}

# 需要 AGENT_ID 的示例会先读取状态；交互终端可选择立即创建临时 Agent。
examples_ensure_agent_id() {  # 定义函数：确保 AGENT_ID 存在
  local create_script="${1:-}"  # 获取第一个参数：创建脚本路径
  if [[ -n "${AGENT_ID:-}" ]]; then  # 如果 AGENT_ID 已存在
    return 0  # 直接返回成功
  fi

  if [[ -n "$create_script" && -t 0 && -t 1 ]]; then  # 如果提供了创建脚本且在交互终端
    local reply  # 声明回复变量
    read -r -p "AGENT_ID is missing. Create a temporary Agent now? [y/N] " reply  # 询问用户是否创建临时 Agent
    case "$reply" in  # 根据用户回复执行不同操作
      y|Y|yes|YES)  # 如果用户同意
        "$create_script"  # 执行创建脚本
        examples_load_kv_file "$EXAMPLES_STATE_FILE"  # 加载状态文件
        examples_require_env AGENT_ID "Agent was created, but AGENT_ID was not saved to $EXAMPLES_STATE_FILE."  # 确保 AGENT_ID 存在
        return 0  # 返回成功
        ;;
      *)  # 如果用户不同意
        echo "Canceled: AGENT_ID is required." >&2  # 输出取消信息到标准错误
        return 1  # 返回错误码1
        ;;
    esac
  fi

  echo "Error: AGENT_ID is required. Set AGENT_ID, add it to $EXAMPLES_ENV_FILE, or run the create target first." >&2  # 输出错误信息到标准错误
  return 1  # 返回错误码1
}
