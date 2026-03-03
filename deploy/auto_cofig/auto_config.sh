#!/bin/bash

#############################################
# Auto environment setup script
# Usage:
#   ./auto_config.sh [agent_json_file] [knowledge_network_json_file] [dataflow_json_file] [operator_file] [toolbox_file] [mcp_file]  # run all steps
#   ./auto_config.sh --step [step_number] [file]  # run a single step (file depends on step)
# Notes: Load settings from config.env or environment variables and build datasource requests dynamically.
#        Decide whether to include the "schema" field based on datasource type.
# Steps:
#   1: Get token
#   2: Create datasource and scan
#   3: Import business knowledge network
#   4: Import DataAgent
#   5: Import data flow
#   6: Import operator
#   7: Import toolbox
#   8: Import MCP
#############################################

set -e

# Colored output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to show help
show_help() {
  cat << EOF
自动配置环境脚本

用法:
  $0 [选项] [agent_json_file] [knowledge_network_json_file] [dataflow_json_file] [operator_file] [toolbox_file] [mcp_file]  # 执行全部步骤
  $0 --step [step_number] [file]  # 执行单步（不同步骤需要的 file 不同）

选项:
  -h, --help                    显示此帮助信息
  --step [step_number]          执行单个步骤 (1-8)

参数:
  agent_json_file              Agent配置文件
  knowledge_network_json_file  业务知识网络JSON文件
  dataflow_json_file           数据流JSON文件
  operator_file                算子文件（.adp 或 .json/.yaml/.yml）
  toolbox_file                 工具文件（.adp 或 .json/.yaml/.yml）
  mcp_file                     MCP 文件（.adp）

步骤说明:
  1: 获取token
  2: 创建数据源并扫描
  3: 导入业务知识网络
  4: 导入DataAgent
  5: 导入数据流
  6: 导入算子
  7: 导入工具
  8: 导入MCP

示例:
  $0 my_agent.json my_knowledge.json my_dataflow.json    # 执行全部步骤，使用自定义文件
  $0 --step 1                           # 仅执行步骤1：获取token
  $0 --step 2                           # 仅执行步骤2：创建数据源并扫描
  $0 --step 3 my_knowledge.json         # 仅执行步骤3：导入业务知识网络
  $0 --step 4 my_agent.json             # 仅执行步骤4：导入DataAgent
  $0 --step 5 my_dataflow.json          # 仅执行步骤5：导入数据流
  $0 --step 6 operator.adp              # 仅执行步骤6：导入算子（.adp 或 .json/.yaml/.yml）
  $0 --step 7 toolbox.adp               # 仅执行步骤7：导入工具（.adp 或 .json/.yaml/.yml）
  $0 --step 8 mcp.adp                   # 仅执行步骤8：导入MCP（.adp）

注意:
  - 脚本会读取 config.env 或环境变量中的设置
  - 需要在配置文件中设置必要的认证信息
EOF
}

# Initialize variables
STEP_MODE=false
STEP_NUMBER=0

# Check for help flag
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
  show_help
  exit 0
fi

# Parse arguments
if [ "$1" = "--step" ]; then
  STEP_MODE=true
  STEP_NUMBER=$2
  shift 2  # remove the first two args (--step and the number)
fi

# Argument validation
if [ $# -gt 6 ]; then
  echo -e "${RED}错误: 参数过多${NC}"
  echo "用法: $0 [agent_json_file] [knowledge_network_json_file] [dataflow_json_file] [operator_file] [toolbox_file] [mcp_file]"
  echo "或: $0 --step [step_number] [file]"
  echo "步骤编号: 1-获取token, 2-创建数据源并扫描, 3-导入知识网络, 4-导入DataAgent, 5-导入数据流, 6-导入算子, 7-导入工具, 8-导入MCP"
  echo ""
  echo "使用 -h 或 --help 参数查看详细帮助信息"
  exit 1
elif [ $# -lt 3 ] && [ "$STEP_MODE" = false ]; then
  echo -e "${RED}错误: 参数不足${NC}"
  echo "用法: $0 [agent_json_file] [knowledge_network_json_file] [dataflow_json_file]"
  echo "或: $0 --step [step_number] [file]"
  echo "步骤编号: 1-获取token, 2-创建数据源并扫描, 3-导入知识网络, 4-导入DataAgent, 5-导入数据流, 6-导入算子, 7-导入工具, 8-导入MCP"
  echo ""
  echo "使用 -h 或 --help 参数查看详细帮助信息"
  exit 1
fi

if [ "$STEP_MODE" = false ]; then
  AGENT_FILE=$1
  KNOWLEDGE_NETWORK_FILE=$2
  DATAFLOW_FILE=$3
  OPERATOR_FILE=${4:-""}
  TOOLBOX_FILE=${5:-""}
  MCP_FILE=${6:-""}
else
  # In step mode, after `--step N` is shifted out, $1 is the first file argument
  AGENT_FILE=${1:-"agent.json"}
  KNOWLEDGE_NETWORK_FILE=${2:-"业务知识网络.json"}
  DATAFLOW_FILE=${3:-"dataflow.json"}
  OPERATOR_FILE=${1:-""}
  TOOLBOX_FILE=${1:-""}
  MCP_FILE=${1:-""}
fi

# Load datasource config (prefer config.env; otherwise use env/defaults)
DS_CONFIG_FILE="${DS_CONFIG_FILE:-config.env}"
if [ -f "$DS_CONFIG_FILE" ]; then
  echo "读取数据源配置: $DS_CONFIG_FILE"
  # shellcheck disable=SC1090
  . "$DS_CONFIG_FILE"
fi

# Detect local IP
echo -e "${YELLOW}正在获取本机IP地址...${NC}"
if [ -n "$IP_ADDRESS" ]; then
  echo "使用环境变量中的IP地址: $IP_ADDRESS"
else
  IP_ADDRESS=$(ip route get 8.8.8.8 2>/dev/null | grep -oP 'src \K\S+' | head -1)
  if [ -z "$IP_ADDRESS" ]; then
    IP_ADDRESS=$(hostname -I | awk '{print $1}' 2>/dev/null)
  fi
  if [ -z "$IP_ADDRESS" ]; then
    echo -e "${RED}错误: 无法自动获取本机IP地址${NC}"
    echo "提示: 可以通过设置环境变量 IP_ADDRESS 来指定IP地址"
    echo "例如: export IP_ADDRESS=192.168.1.100"
    exit 1
  fi
fi

BASE_URL="https://${IP_ADDRESS}"

# Get auth settings from config
USERNAME="${USERNAME:-admin}"
PASSWORD="${PASSWORD:-}"

# Validate required settings
if [ -z "$PASSWORD" ]; then
  echo -e "${RED}错误: 未设置用户密码${NC}"
  echo "请在配置文件中设置 PASSWORD"
  exit 1
fi

# Define two different public keys
LOGIN_PUBLIC_KEY='-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsyOstgbYuubBi2PUqeVj
GKlkwVUY6w1Y8d4k116dI2SkZI8fxcjHALv77kItO4jYLVplk9gO4HAtsisnNE2o
wlYIqdmyEPMwupaeFFFcg751oiTXJiYbtX7ABzU5KQYPjRSEjMq6i5qu/mL67XTk
hvKwrC83zme66qaKApmKupDODPb0RRkutK/zHfd1zL7sciBQ6psnNadh8pE24w8O
2XVy1v2bgSNkGHABgncR7seyIg81JQ3c/Axxd6GsTztjLnlvGAlmT1TphE84mi99
fUaGD2A1u1qdIuNc+XuisFeNcUW6fct0+x97eS2eEGRr/7qxWmO/P20sFVzXc2bF
1QIDAQAB
-----END PUBLIC KEY-----'

DATABASE_PUBLIC_KEY='-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA22GOSQ1jeDhpdzxhJddS
f+U10F4Ivut7giYhchFAIJgRonMamDT86MSqQUc8DdTFdPGLm7M3GUKcsG1qbC3S
qk4XJ9NjmQXbs7IMWyWEWQrN7Iv7S2QjDYJI+ppvIN03I0Km3WKsmnrle2bLzT/V
G8e72YX69dfXAeiX6uDhht1va/JxZVFMIV3pHa6AQQ9gn5SAUTX2akEhRfe1bPJj
fVyoM+dfNtvgdfaraqV1rOhVDEqd0NlOWt2RHwETQwU8gIJib2baj2MtyIAY+fQw
KlKWxUs1GcFbECnhVPiVN6BEhXD7OhRt9QE/cuYl5v4a6ypugGaMBK6VKOqFHDvf
mwIDAQAB
-----END PUBLIC KEY-----'

# Determine required files based on execution mode
if [ "$STEP_MODE" = false ]; then
  # Run-all mode: all files required
  if [ ! -f "$AGENT_FILE" ]; then
    echo -e "${RED}错误: Agent 配置文件不存在: $AGENT_FILE${NC}"
    exit 1
  fi

  if [ ! -f "$KNOWLEDGE_NETWORK_FILE" ]; then
    echo -e "${RED}错误: 业务知识网络JSON文件不存在: $KNOWLEDGE_NETWORK_FILE${NC}"
    exit 1
  fi
  
  if [ ! -f "$DATAFLOW_FILE" ]; then
    echo -e "${RED}错误: 数据流JSON文件不存在: $DATAFLOW_FILE${NC}"
    exit 1
  fi

  # Read knowledge network content
  KNOWLEDGE_NETWORK_JSON_CONTENT=$(cat "$KNOWLEDGE_NETWORK_FILE")
  # Read data flow content
  DATAFLOW_FILE_JSON_CONTENT=$(cat "$DATAFLOW_FILE")
else
  # Step mode: required files depend on the selected step
  if [ "$STEP_NUMBER" -eq 1 ]; then
    # Step 1 only needs base config; no extra files
    :
  elif [ "$STEP_NUMBER" -eq 2 ]; then
    # Step 2 only needs base config; no extra files (create datasource)
    :
  elif [ "$STEP_NUMBER" -eq 3 ]; then
    # Step 3 requires the knowledge network file
    KNOWLEDGE_NETWORK_FILE=${1:-"业务知识网络.json"}
    if [ ! -f "$KNOWLEDGE_NETWORK_FILE" ]; then
      echo -e "${RED}错误: 业务知识网络JSON文件不存在: $KNOWLEDGE_NETWORK_FILE${NC}"
      exit 1
    fi
    # Read knowledge network content
    KNOWLEDGE_NETWORK_JSON_CONTENT=$(cat "$KNOWLEDGE_NETWORK_FILE")
  elif [ "$STEP_NUMBER" -eq 4 ]; then
    # Step 4 requires the agent file
    if [ ! -f "$AGENT_FILE" ]; then
      echo -e "${RED}错误: Agent 配置文件不存在: $AGENT_FILE${NC}"
      exit 1
    fi
  elif [ "$STEP_NUMBER" -eq 5 ]; then
    # Step 5 requires the data flow file
    DATAFLOW_FILE=${1:-"dataflow.json"}
    if [ ! -f "$DATAFLOW_FILE" ]; then
      echo -e "${RED}错误: 数据流JSON文件不存在: $DATAFLOW_FILE${NC}"
      exit 1
    fi
    # Read data flow content
    DATAFLOW_FILE_JSON_CONTENT=$(cat "$DATAFLOW_FILE")
  elif [ "$STEP_NUMBER" -eq 6 ]; then
    # Step 6 requires the operator file
    OPERATOR_FILE=${1:-""}
    if [ -z "$OPERATOR_FILE" ] || [ ! -f "$OPERATOR_FILE" ]; then
      echo -e "${RED}错误: 算子文件不存在或未指定: $OPERATOR_FILE${NC}"
      exit 1
    fi
  elif [ "$STEP_NUMBER" -eq 7 ]; then
    # Step 7 requires the toolbox file
    TOOLBOX_FILE=${1:-""}
    if [ -z "$TOOLBOX_FILE" ] || [ ! -f "$TOOLBOX_FILE" ]; then
      echo -e "${RED}错误: 工具文件不存在或未指定: $TOOLBOX_FILE${NC}"
      exit 1
    fi
  elif [ "$STEP_NUMBER" -eq 8 ]; then
    # Step 8 requires the MCP file
    MCP_FILE=${1:-""}
    if [ -z "$MCP_FILE" ] || [ ! -f "$MCP_FILE" ]; then
      echo -e "${RED}错误: MCP文件不存在或未指定: $MCP_FILE${NC}"
      exit 1
    fi
  fi
fi

BUSINESS_DOMAIN="bd_public"

echo -e "${GREEN}开始自动配置环境...${NC}"
echo "IP地址: $IP_ADDRESS (自动获取)"
echo "用户名: $USERNAME"
echo "Agent文件: $AGENT_FILE"
echo "知识网络文件: $KNOWLEDGE_NETWORK_FILE"
echo "数据流文件: $DATAFLOW_FILE"
echo ""

# Get token
get_token() {
  local temp_suffix=$$
  echo -e "${YELLOW}[1/8] 自动获取登录token...${NC}"

  # Create temp public key files for openssl
  LOGIN_PUBKEY_FILE="/tmp/login_public_key_${temp_suffix}"
  DATABASE_PUBKEY_FILE="/tmp/database_public_key_${temp_suffix}"

  echo "$LOGIN_PUBLIC_KEY" > "$LOGIN_PUBKEY_FILE"
  echo "$DATABASE_PUBLIC_KEY" > "$DATABASE_PUBKEY_FILE"

  # Encrypt login password with openssl
  local ENCRYPTED_LOGIN_PASSWORD=""
  if command -v openssl >/dev/null 2>&1 && [ -n "$PASSWORD" ]; then
    ENCRYPTED_LOGIN_PASSWORD=$(printf "%s" "$PASSWORD" | openssl rsautl -encrypt -pubin -inkey "$LOGIN_PUBKEY_FILE" 2>/dev/null | base64 | tr -d '\n')
  fi

  if [ -z "$ENCRYPTED_LOGIN_PASSWORD" ]; then
    echo -e "${RED}错误: 登录密码加密失败${NC}"
    rm -f "$LOGIN_PUBKEY_FILE" "$DATABASE_PUBKEY_FILE"
    return 1
  fi

  # Fetch CSRF token and challenge from the login page
  local LOGIN_PAGE_URL="${BASE_URL}/interface/studioweb/login?lang=zh-cn&state=EjVex8mfXS&x-forwarded-prefix=&integrated=false&product=adp&_t=$(date +%s)000"
  local LOGIN_PAGE_RESPONSE=$(curl -s -k -c "/tmp/session_cookies_${temp_suffix}.txt" -L "$LOGIN_PAGE_URL")

  if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 无法访问登录页面${NC}"
    rm -f "$LOGIN_PUBKEY_FILE" "$DATABASE_PUBKEY_FILE" "/tmp/session_cookies_${temp_suffix}.txt"
    return 1
  fi

  # Extract CSRF token and challenge (more flexible regex)
  CSRFTOKEN_REGEX='"csrftoken"[[:space:]]*:[[:space:]]*"[^"]*"'
  CHALLENGE_REGEX='"challenge"[[:space:]]*:[[:space:]]*"[^"]*"'

  CSRFTOKEN_LINE=$(echo "$LOGIN_PAGE_RESPONSE" | grep -oP "$CSRFTOKEN_REGEX" | head -1)
  CHALLENGE_LINE=$(echo "$LOGIN_PAGE_RESPONSE" | grep -oP "$CHALLENGE_REGEX" | head -1)

  # Extract token/challenge values
  local CSRF_TOKEN=""
  local CHALLENGE=""
  if [ -n "$CSRFTOKEN_LINE" ]; then
    CSRF_TOKEN=$(echo "$CSRFTOKEN_LINE" | cut -d'"' -f4)
  fi

  if [ -n "$CHALLENGE_LINE" ]; then
    CHALLENGE=$(echo "$CHALLENGE_LINE" | cut -d'"' -f4)
  fi

  if [ -z "$CSRF_TOKEN" ] || [ -z "$CHALLENGE" ]; then
    echo -e "${RED}错误: 无法从登录页面获取CSRF token或challenge${NC}"
    echo "CSRF Token: $CSRF_TOKEN"
    echo "Challenge: $CHALLENGE"
    rm -f "$LOGIN_PUBKEY_FILE" "$DATABASE_PUBKEY_FILE" "/tmp/session_cookies_${temp_suffix}.txt"
    return 1
  fi

  # Build login payload
  local SIGNIN_URL="${BASE_URL}/oauth2/signin"
  local LOGIN_PAYLOAD=$(cat <<EOF
{
  "_csrf": "$CSRF_TOKEN",
  "challenge": "$CHALLENGE",
  "account": "$USERNAME",
  "password": "$ENCRYPTED_LOGIN_PASSWORD",
  "vcode": {"id": "", "content": ""},
  "dualfactorauthinfo": {"validcode": {"vcode": ""}, "OTP": {"OTP": ""}},
  "remember": false,
  "device": {"name": "", "description": "", "client_type": "console_web", "udids": []}
}
EOF
  )

  # Send login request
  local SIGNIN_RESPONSE=$(curl -s -k -XPOST \
    -H "Content-Type: application/json" \
    -b "/tmp/session_cookies_${temp_suffix}.txt" \
    -c "/tmp/session_cookies_after_signin_${temp_suffix}.txt" \
    "$SIGNIN_URL" \
    -d "$LOGIN_PAYLOAD")

  if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 登录请求发送失败${NC}"
    rm -f "$LOGIN_PUBKEY_FILE" "$DATABASE_PUBKEY_FILE" "/tmp/session_cookies_${temp_suffix}.txt" "/tmp/session_cookies_after_signin_${temp_suffix}.txt"
    return 1
  fi

  # Validate login response
  local REDIRECT_URL=$(echo "$SIGNIN_RESPONSE" | grep -oP '"redirect"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | cut -d'"' -f4 | sed 's/\\//g')

  if [ -z "$REDIRECT_URL" ]; then
    echo -e "${RED}错误: 登录失败，响应内容: $SIGNIN_RESPONSE${NC}"
    rm -f "$LOGIN_PUBKEY_FILE" "$DATABASE_PUBKEY_FILE" "/tmp/session_cookies_${temp_suffix}.txt" "/tmp/session_cookies_after_signin_${temp_suffix}.txt"
    return 1
  fi

  # Follow redirect to complete login
  local REDIRECT_RESPONSE=$(curl -s -k -L GET \
    -b "/tmp/session_cookies_after_signin_${temp_suffix}.txt" \
    -c "/tmp/final_cookies_${temp_suffix}.txt" \
    "$REDIRECT_URL")

  if [ $? -ne 0 ]; then
    echo -e "${RED}错误: 重定向URL访问失败${NC}"
    rm -f "$LOGIN_PUBKEY_FILE" "$DATABASE_PUBKEY_FILE" "/tmp/session_cookies_${temp_suffix}.txt" "/tmp/session_cookies_after_signin_${temp_suffix}.txt" "/tmp/final_cookies_${temp_suffix}.txt"
    return 1
  fi

  # Find token in cookie file
  local TOKEN=$(awk '$6 ~ /studio\.oauth2_token/ {print $7}' "/tmp/final_cookies_${temp_suffix}.txt" 2>/dev/null | head -1)

  if [ -z "$TOKEN" ]; then
    # If not found, try extracting from response headers
    local RESPONSE_WITH_HEADERS=$(curl -s -k -D - -L \
      -b "/tmp/session_cookies_after_signin_${temp_suffix}.txt" \
      -c "/tmp/final_cookies_v2_${temp_suffix}.txt" \
      "$REDIRECT_URL")

    # Extract Set-Cookie from headers
    TOKEN=$(echo "$RESPONSE_WITH_HEADERS" | grep -i 'Set-Cookie:' | grep 'studio.oauth2_token' | grep -oP 'studio\.oauth2_token=\K[^;]*' | head -1)
    # Still not found? Try the v2 cookie file
    if [ -z "$TOKEN" ]; then
      TOKEN=$(awk '$6 ~ /studio\.oauth2_token/ {print $7}' "/tmp/final_cookies_v2_${temp_suffix}.txt" 2>/dev/null | head -1)
    fi
  fi

  if [ -z "$TOKEN" ]; then
    echo -e "${RED}错误: 无法从cookie中提取token${NC}"
    echo "检查cookie文件内容："
    echo "=== session_cookies_${temp_suffix}.txt ==="
    cat "/tmp/session_cookies_${temp_suffix}.txt"
    echo -e "\n=== session_cookies_after_signin_${temp_suffix}.txt ==="
    cat "/tmp/session_cookies_after_signin_${temp_suffix}.txt"
    echo -e "\n=== final_cookies_${temp_suffix}.txt ==="
    cat "/tmp/final_cookies_${temp_suffix}.txt"
    rm -f "$LOGIN_PUBKEY_FILE" "$DATABASE_PUBKEY_FILE" "/tmp/session_cookies_${temp_suffix}.txt" "/tmp/session_cookies_after_signin_${temp_suffix}.txt" "/tmp/final_cookies_${temp_suffix}.txt" "/tmp/final_cookies_v2_${temp_suffix}.txt"
    return 1
  fi

  # Remove cookie temp files
  rm -f "/tmp/cookies*${temp_suffix}.txt"

  # Write token to a temp env file for later steps
  echo "export TOKEN=\"$TOKEN\"" > "/tmp/token_${temp_suffix}.env"

  echo -e "${GREEN}✓ Token获取成功${NC}"
  echo "Token: ${TOKEN}"

  # Cleanup temp files
  rm -f "$LOGIN_PUBKEY_FILE" "$DATABASE_PUBKEY_FILE" "/tmp/session_cookies_${temp_suffix}.txt" "/tmp/session_cookies_after_signin_${temp_suffix}.txt" "/tmp/final_cookies_${temp_suffix}.txt" "/tmp/final_cookies_v2_${temp_suffix}.txt"

  return 0
}

# Helper: ensure token exists
ensure_token_exists() {
  local temp_suffix=$$
  if [ -f "/tmp/token_${temp_suffix}.env" ]; then
    source "/tmp/token_${temp_suffix}.env"
  elif [ -n "$TOKEN" ]; then
    # TOKEN already exists; keep using it
    return 0
  else
    echo -e "${YELLOW}未找到认证令牌，正在获取...${NC}"
    if get_token; then
      source "/tmp/token_${temp_suffix}.env"
      return 0
    else
      echo -e "${RED}获取认证令牌失败${NC}"
      return 1
    fi
  fi
}

# Create datasource
create_datasource() {
  local temp_suffix=$$
  echo -e "${YELLOW}[2/8] 根据配置创建数据源...${NC}"

  # Ensure token exists
  ensure_token_exists || return 1

  DS_TYPE="${DS_TYPE:-mysql}"
  DS_HOST="${DS_HOST:-127.0.0.1}"
  DS_PORT=${DS_PORT:-3306}
  DS_DATABASE_NAME="${DS_DATABASE_NAME:-tem}"
  DS_USERNAME="${DS_USERNAME:-root}"
  DS_PASSWORD="${DS_PASSWORD:-}"
  DS_SCHEMA_NAME="${DS_SCHEMA_NAME:-}"
  DS_NAME="${DS_NAME:-HD供应链业务分析}"
  DS_COMMENT="${DS_COMMENT:-存储业务分析智能体分析结果}"
  NEEDS_SCHEMA=0
  case "$DS_TYPE" in
  oracle|postgresql|sqlserver|hologres|opengauss|gaussdb)
    NEEDS_SCHEMA=1
    ;;
  esac

  echo -e "数据源类型: $DS_TYPE, 主机: $DS_HOST:$DS_PORT, 数据库: $DS_DATABASE_NAME"

  # Create a temp public key file for openssl
  DATABASE_PUBKEY_FILE="/tmp/database_public_key_${temp_suffix}"
  echo "$DATABASE_PUBLIC_KEY" > "$DATABASE_PUBKEY_FILE"

  # Encrypt datasource password with database public key
  local ENC_PASSWORD=""
  if command -v openssl >/dev/null 2>&1 && [ -n "$DS_PASSWORD" ]; then
    ENC_PASSWORD=$(printf "%s" "$DS_PASSWORD" | openssl rsautl -encrypt -pubin -inkey "$DATABASE_PUBKEY_FILE" 2>/dev/null | base64 | tr -d '\n')
  fi
  if [ -z "$ENC_PASSWORD" ]; then
    ENC_PASSWORD="$DS_PASSWORD"
  fi

  local BIN_DATA_JSON="{\"database_name\":\"$DS_DATABASE_NAME\",\"connect_protocol\":\"jdbc\",\"host\":\"$DS_HOST\",\"port\":$DS_PORT"
  if [ "$NEEDS_SCHEMA" -eq 1 ] && [ -n "$DS_SCHEMA_NAME" ]; then
    BIN_DATA_JSON+=",\"schema\":\"$DS_SCHEMA_NAME\""
  fi
  BIN_DATA_JSON+=",\"account\":\"$DS_USERNAME\",\"password\":\"$ENC_PASSWORD\"}"

  local DATASOURCE_PAYLOAD=$(cat <<EOF
{
  "name": "${DS_NAME}",
  "bin_data": ${BIN_DATA_JSON},
  "comment": "${DS_COMMENT}",
  "type": "${DS_TYPE}"
}
EOF
  )

  local DATASOURCE_RESPONSE=$(curl -s -k -X POST \
    "${BASE_URL}/api/data-connection/v1/datasource" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    -d "$DATASOURCE_PAYLOAD")

  local DATASOURCE_ID=$(echo $DATASOURCE_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
  local DATASOURCE_NAME=$(echo $DATASOURCE_RESPONSE | grep -o '"name":"[^"]*"' | cut -d'"' -f4)

  if [ -z "$DATASOURCE_ID" ] || [ -z "$DATASOURCE_NAME" ]; then
    echo -e "${RED}错误: 无法从响应中提取数据源ID或名称${NC}"
    echo "响应内容: $DATASOURCE_RESPONSE"
    rm -f "$DATABASE_PUBKEY_FILE" 
    return 1
  else
    # Save datasource info for scan step
    echo "export DATASOURCE_ID=\"$DATASOURCE_ID\"" > "/tmp/datasource_${temp_suffix}.env"
    echo "export DATASOURCE_NAME=\"$DATASOURCE_NAME\"" >> "/tmp/datasource_${temp_suffix}.env"

    echo -e "${GREEN}✓ 数据源添加成功${NC}"
    echo "  数据源ID: $DATASOURCE_ID"
    echo "  数据源名称: $DATASOURCE_NAME"
    rm -f "$DATABASE_PUBKEY_FILE"
  fi

  # Scan immediately after creating the datasource
  scan_datasource_by_id "$DATASOURCE_ID" "$DATASOURCE_NAME"
  local SCAN_RESULT=$?

  return $SCAN_RESULT
}

# Scan datasource by ID (called right after creation)
scan_datasource_by_id() {
  local datasource_id=$1
  local datasource_name=$2
  local temp_suffix=$$
  echo -e "${YELLOW}  扫描数据源...${NC}"

  # Ensure token exists
  ensure_token_exists || return 1

  DS_TYPE="${DS_TYPE:-mysql}"

  # Build scan request payload
  local SCAN_PAYLOAD=$(cat <<EOF
{
  "scan_name": "${datasource_name}",
  "ds_info": {
    "ds_id": "${datasource_id}",
    "ds_type": "${DS_TYPE}",
    "scan_strategy": []
  },
  "type": 0,
  "cron_expression": {},
  "status": "open",
  "use_default_template": true,
  "field_list_when_change": [],
  "use_multi_threads": true,
  "tables": []
}
EOF
  )

  local SCAN_RESPONSE=$(curl -s -k -X POST \
    "${BASE_URL}/api/data-connection/v1/metadata/scan" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    -d "$SCAN_PAYLOAD")

  local SCAN_ID=$(echo $SCAN_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
  local SCAN_STATUS=$(echo $SCAN_RESPONSE | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

  if [ -z "$SCAN_ID" ]; then
    echo -e "${RED}错误: 无法从响应中提取扫描ID${NC}"
    echo "响应内容: $SCAN_RESPONSE"
    return 1
  else
    # Save scan ID to temp file
    echo "export SCAN_ID=\"$SCAN_ID\"" > "/tmp/scan_${temp_suffix}.env"

    echo -e "${GREEN}✓ 数据源扫描已启动${NC}"
    echo "  扫描ID: $SCAN_ID"
    echo "  状态: $SCAN_STATUS"
    return 0
  fi
}

# Import business knowledge network
import_knowledge_network() {
  local temp_suffix=$$
  echo -e "${YELLOW}[3/8] 导入业务知识网络...${NC}"

  # Ensure knowledge network file exists
  if [ ! -f "$KNOWLEDGE_NETWORK_FILE" ]; then
    echo -e "${RED}错误: 业务知识网络JSON文件不存在: $KNOWLEDGE_NETWORK_FILE${NC}"
    return 1
  fi

  # Ensure token exists
  ensure_token_exists || return 1

  # Import knowledge network (submit the JSON payload directly via API)
  local KN_RESPONSE=$(curl -s -k -X POST \
    "${BASE_URL}/api/ontology-manager/v1/knowledge-networks" \
    -H "Content-Type: application/json" \
    -H "x-business-domain: ${BUSINESS_DOMAIN}" \
    -H "Authorization: Bearer ${TOKEN}" \
    --data-binary "@${KNOWLEDGE_NETWORK_FILE}")

  local KN_ID=$(echo $KN_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
  if [ -z "$KN_ID" ]; then
    echo -e "${RED}错误: 无法从知识网络响应中提取ID${NC}"
    echo "响应内容: $KN_RESPONSE"
    return 1
  else
    # Save knowledge network ID to temp file
    echo "export KN_ID=\"$KN_ID\"" > "/tmp/knowledge_network_${temp_suffix}.env"

    echo -e "${GREEN}✓ 业务知识网络导入成功${NC}"
    echo "  知识网络ID: $KN_ID"
    return 0
  fi
}

# Import DataAgent
import_agent() {
  local temp_suffix=$$
  echo -e "${YELLOW}[4/8] 导入DataAgent...${NC}"

  # Ensure agent file exists
  if [ ! -f "$AGENT_FILE" ]; then
    echo -e "${RED}错误: Agent 配置文件不存在: $AGENT_FILE${NC}"
    return 1
  fi

  # Ensure token exists
  ensure_token_exists || return 1

  local AGENT_RESPONSE=$(curl -s -k -X POST \
    "${BASE_URL}/api/agent-factory/v3/agent-inout/import" \
    -H "x-business-domain: ${BUSINESS_DOMAIN}" \
    -H "Authorization: Bearer ${TOKEN}" \
    -F "file=@${AGENT_FILE};type=application/json" \
    -F "import_type=create")

  local IS_SUCCESS=$(echo $AGENT_RESPONSE | grep -o '"is_success":[^,}]*' | cut -d':' -f2)
  if [ "$IS_SUCCESS" = "true" ]; then
    echo -e "${GREEN}✓ DataAgent导入成功${NC}"
    return 0
  else
    echo -e "${RED}警告: DataAgent导入可能失败${NC}"
    echo "响应内容: $AGENT_RESPONSE"
    return 1
  fi
}

# Import data flow
import_data_flow() {
  local temp_suffix=$$
  echo -e "${YELLOW}[5/8] 导入数据流...${NC}"

  # Ensure data flow file exists
  if [ ! -f "$DATAFLOW_FILE" ]; then
    echo -e "${RED}错误: 数据流JSON文件不存在: $DATAFLOW_FILE${NC}"
    return 1
  fi
  
  # Read data flow content
  local DATAFLOW_FILE_JSON_CONTENT=$(cat "$DATAFLOW_FILE")
  
  # Ensure token exists
  ensure_token_exists || return 1

  local DATA_FLOW_RESPONSE=$(curl -s -k -X POST \
    "${BASE_URL}/api/automation/v1/data-flow/flow" \
    -H "x-business-domain: ${BUSINESS_DOMAIN}" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "$DATAFLOW_FILE_JSON_CONTENT")

 
  local DF_ID=$(echo $DATA_FLOW_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
  
  if [ -z "$DF_ID" ]; then
    echo -e "${RED}错误: 无法从数据流响应中提取ID${NC}"
    echo "响应内容: $DATA_FLOW_RESPONSE"
    return 1
  else
    # Save data flow ID to temp file
    echo "export DF_ID=\"$DF_ID\"" > "/tmp/dataflow_${temp_suffix}.env"
    
    echo -e "${GREEN}✓ 数据流导入成功${NC}"
    echo "  数据流ID: $DF_ID"
    return 0
  fi
}

# Import operator (adp/openapi)
import_operator() {
  local temp_suffix=$$
  echo -e "${YELLOW}[6/8] 导入算子...${NC}"

  if [ -z "$OPERATOR_FILE" ] || [ ! -f "$OPERATOR_FILE" ]; then
    echo -e "${RED}错误: 算子文件不存在或未指定: $OPERATOR_FILE${NC}"
    return 1
  fi

  ensure_token_exists || return 1

  local FILE_LOWER
  FILE_LOWER=$(echo "$OPERATOR_FILE" | tr '[:upper:]' '[:lower:]')

  local RESULT=""
  if [[ "$FILE_LOWER" == *.adp ]]; then
    RESULT=$(curl -s -k -w "\n%{http_code}" -X POST \
      "${BASE_URL}/api/agent-operator-integration/v1/impex/import/operator" \
      -H "x-business-domain: ${BUSINESS_DOMAIN}" \
      -H "Authorization: Bearer ${TOKEN}" \
      -F "data=@${OPERATOR_FILE}" \
      -F "mode=upsert")
  else
    RESULT=$(curl -s -k -w "\n%{http_code}" -X POST \
      "${BASE_URL}/api/agent-operator-integration/v1/operator/register" \
      -H "x-business-domain: ${BUSINESS_DOMAIN}" \
      -H "Authorization: Bearer ${TOKEN}" \
      -F "data=@${OPERATOR_FILE}" \
      -F "operator_metadata_type=openapi")
  fi

  local HTTP_CODE
  HTTP_CODE=$(echo "$RESULT" | tail -n 1)
  local BODY
  BODY=$(echo "$RESULT" | sed '$d')

  if [[ "$HTTP_CODE" =~ ^2[0-9][0-9]$ ]]; then
    echo -e "${GREEN}✓ 算子导入成功${NC}"
    return 0
  else
    echo -e "${RED}错误: 算子导入失败 (HTTP $HTTP_CODE)${NC}"
    echo "响应内容: $BODY"
    return 1
  fi
}

# Import toolbox (adp/openapi)
import_toolbox() {
  local temp_suffix=$$
  echo -e "${YELLOW}[7/8] 导入工具...${NC}"

  if [ -z "$TOOLBOX_FILE" ] || [ ! -f "$TOOLBOX_FILE" ]; then
    echo -e "${RED}错误: 工具文件不存在或未指定: $TOOLBOX_FILE${NC}"
    return 1
  fi

  ensure_token_exists || return 1

  local FILE_LOWER
  FILE_LOWER=$(echo "$TOOLBOX_FILE" | tr '[:upper:]' '[:lower:]')

  local RESULT=""
  if [[ "$FILE_LOWER" == *.adp ]]; then
    RESULT=$(curl -s -k -w "\n%{http_code}" -X POST \
      "${BASE_URL}/api/agent-operator-integration/v1/impex/import/toolbox" \
      -H "x-business-domain: ${BUSINESS_DOMAIN}" \
      -H "Authorization: Bearer ${TOKEN}" \
      -F "data=@${TOOLBOX_FILE}" \
      -F "mode=upsert")
  else
    RESULT=$(curl -s -k -w "\n%{http_code}" -X POST \
      "${BASE_URL}/api/agent-operator-integration/v1/tool-box" \
      -H "x-business-domain: ${BUSINESS_DOMAIN}" \
      -H "Authorization: Bearer ${TOKEN}" \
      -F "data=@${TOOLBOX_FILE}" \
      -F "metadata_type=openapi")
  fi

  local HTTP_CODE
  HTTP_CODE=$(echo "$RESULT" | tail -n 1)
  local BODY
  BODY=$(echo "$RESULT" | sed '$d')

  if [[ "$HTTP_CODE" =~ ^2[0-9][0-9]$ ]]; then
    local BOX_ID
    BOX_ID=$(echo "$BODY" | grep -o '"box_id":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$BOX_ID" ]; then
      echo "export TOOLBOX_BOX_ID=\"$BOX_ID\"" > "/tmp/toolbox_${temp_suffix}.env"
      echo "  box_id: $BOX_ID"
    fi
    echo -e "${GREEN}✓ 工具导入成功${NC}"
    return 0
  else
    echo -e "${RED}错误: 工具导入失败 (HTTP $HTTP_CODE)${NC}"
    echo "响应内容: $BODY"
    return 1
  fi
}

# Import MCP (adp)
import_mcp() {
  local temp_suffix=$$
  echo -e "${YELLOW}[8/8] 导入MCP...${NC}"

  if [ -z "$MCP_FILE" ] || [ ! -f "$MCP_FILE" ]; then
    echo -e "${RED}错误: MCP文件不存在或未指定: $MCP_FILE${NC}"
    return 1
  fi

  ensure_token_exists || return 1

  local RESULT
  RESULT=$(curl -s -k -w "\n%{http_code}" -X POST \
    "${BASE_URL}/api/agent-operator-integration/v1/impex/import/mcp" \
    -H "x-business-domain: ${BUSINESS_DOMAIN}" \
    -H "Authorization: Bearer ${TOKEN}" \
    -F "data=@${MCP_FILE}" \
    -F "mode=upsert")

  local HTTP_CODE
  HTTP_CODE=$(echo "$RESULT" | tail -n 1)
  local BODY
  BODY=$(echo "$RESULT" | sed '$d')

  if [[ "$HTTP_CODE" =~ ^2[0-9][0-9]$ ]]; then
    echo -e "${GREEN}✓ MCP导入成功${NC}"
    return 0
  else
    echo -e "${RED}错误: MCP导入失败 (HTTP $HTTP_CODE)${NC}"
    echo "响应内容: $BODY"
    return 1
  fi
}

# Main execution
if [ "$STEP_MODE" = true ]; then
  # Step mode
  case $STEP_NUMBER in
    1)
      get_token
      ;;
    2)
      create_datasource
      ;;
    3)
      import_knowledge_network
      ;;
    4)
      import_agent
      ;;
    5)
      import_data_flow
      ;;
    6)
      import_operator
      ;;
    7)
      import_toolbox
      ;;
    8)
      import_mcp
      ;;
    *)
      echo -e "${RED}错误: 无效的步骤编号${NC}"
      echo "有效步骤编号: 1-获取token, 2-创建数据源并扫描, 3-导入知识网络, 4-导入DataAgent, 5-导入数据流, 6-导入算子, 7-导入工具, 8-导入MCP"
      exit 1
      ;;
  esac
  # Cleanup temp files
  rm -rf /tmp/datasource_*.env /tmp/knowledge_network_*.env /tmp/agent_*.env /tmp/dataflow_*.env /tmp/scan_*.env /tmp/token_*.env /tmp/toolbox_*.env
  exit $?
else
  # Run-all mode
  echo "执行全部步骤..."

  # 1. Get token
  if get_token; then
    DATASOURCE_SUCCESS=0

    # 2. Create datasource
    if create_datasource; then
      DATASOURCE_SUCCESS=1
    else
      DATASOURCE_SUCCESS=0
      SCAN_SUCCESS=0
      echo -e "${YELLOW}⚠ 数据源创建失败，跳过扫描${NC}"
    fi

    # 4. Import business knowledge network
    if import_knowledge_network; then
      KN_SUCCESS=1
    else
      KN_SUCCESS=0
      echo -e "${YELLOW}⚠ 知识网络导入失败${NC}"
    fi

    # 5. Import DataAgent
    if import_agent; then
      AGENT_SUCCESS=1
    else
      AGENT_SUCCESS=0
      echo -e "${YELLOW}⚠ DataAgent导入失败${NC}"
    fi
    
    # 6. Import data flow
    if import_data_flow; then
      DATAFLOW_SUCCESS=1
    else
      DATAFLOW_SUCCESS=0
      echo -e "${YELLOW}⚠ 数据流导入失败${NC}"
    fi

    # 7. Import operator (optional)
    if [ -n "$OPERATOR_FILE" ]; then
      if [ -f "$OPERATOR_FILE" ]; then
        if import_operator; then
          OPERATOR_SUCCESS=1
        else
          OPERATOR_SUCCESS=0
          echo -e "${YELLOW}⚠ 算子导入失败${NC}"
        fi
      else
        echo -e "${YELLOW}⚠ 未找到算子文件，跳过：$OPERATOR_FILE${NC}"
      fi
    else
      echo -e "${YELLOW}⚠ 未指定算子文件，跳过导入算子${NC}"
    fi

    # 8. Import toolbox (optional)
    if [ -n "$TOOLBOX_FILE" ]; then
      if [ -f "$TOOLBOX_FILE" ]; then
        if import_toolbox; then
          TOOLBOX_SUCCESS=1
        else
          TOOLBOX_SUCCESS=0
          echo -e "${YELLOW}⚠ 工具导入失败${NC}"
        fi
      else
        echo -e "${YELLOW}⚠ 未找到工具文件，跳过：$TOOLBOX_FILE${NC}"
      fi
    else
      echo -e "${YELLOW}⚠ 未指定工具文件，跳过导入工具${NC}"
    fi

    # 9. Import MCP (optional)
    if [ -n "$MCP_FILE" ]; then
      if [ -f "$MCP_FILE" ]; then
        if import_mcp; then
          MCP_SUCCESS=1
        else
          MCP_SUCCESS=0
          echo -e "${YELLOW}⚠ MCP导入失败${NC}"
        fi
      else
        echo -e "${YELLOW}⚠ 未找到MCP文件，跳过：$MCP_FILE${NC}"
      fi
    else
      echo -e "${YELLOW}⚠ 未指定MCP文件，跳过导入MCP${NC}"
    fi
  else
    echo -e "${RED}获取token失败，终止执行${NC}"
    exit 1
  fi
fi

# Cleanup temp files
rm -rf /tmp/datasource_*.env /tmp/knowledge_network_*.env /tmp/agent_*.env /tmp/dataflow_*.env /tmp/scan_*.env /tmp/token_*.env /tmp/toolbox_*.env
exit 0