#!/bin/bash

# Docker Compose 配置验证脚本
# 用于验证环境变量和端口配置

echo "=== Docker Compose 配置验证 ==="
echo ""

# 检查 .env 文件是否存在
if [ ! -f ".env" ]; then
    echo "❌ .env 文件不存在"
    exit 1
fi

echo "✅ .env 文件存在"

# 定义需要检查的环境变量
required_vars=(
    "MYSQL_ROOT_PASSWORD:MariaDB root 密码"
    "MYSQL_DATABASE:MariaDB 数据库名"
    "MYSQL_USER:MariaDB 用户名"
    "MYSQL_PASSWORD:MariaDB 密码"
    "MARIADB_HOST_PORT:MariaDB 主机端口"
    "REDIS_PASSWORD:Redis 密码"
    "REDIS_HOST_PORT:Redis 主机端口"
    "DB_HOST:数据库主机"
    "DB_PORT:数据库端口"
    "REDIS_HOST:Redis 主机"
    "REDIS_PORT:Redis 端口"
    "AGENT_FACTORY_HOST_PORT:Agent Factory 主机端口"
    "AGENT_EXECUTOR_HOST_PORT:Agent Executor 主机端口"
    "AGENT_MEMORY_HOST_PORT:Agent Memory 主机端口"
    "AGENT_WEB_HOST_PORT:Agent Web 主机端口"
    "BACKEND_URL:后端 URL"
    "TZ:时区"
    "LOG_LEVEL:日志级别"
)

echo ""
echo "=== 检查环境变量 ==="
missing_vars=0

for item in "${required_vars[@]}"; do
    var="${item%%:*}"
    desc="${item##*:}"
    if grep -q "^${var}=" .env; then
        value=$(grep "^${var}=" .env | cut -d'=' -f2)
        echo "✅ ${var} (${desc}) = ${value}"
    else
        echo "❌ ${var} (${desc}) 未设置"
        missing_vars=$((missing_vars + 1))
    fi
done

if [ $missing_vars -gt 0 ]; then
    echo ""
    echo "❌ 发现 $missing_vars 个缺失的环境变量"
    exit 1
fi

echo ""
echo "=== 验证 Docker Compose 配置 ==="
if docker-compose config >/dev/null 2>&1; then
    echo "✅ Docker Compose 配置语法正确"
else
    echo "❌ Docker Compose 配置语法错误"
    docker-compose config
    exit 1
fi

echo ""
echo "=== 检查服务状态 ==="
if docker-compose ps --format "table {{.Name}}\t{{.Status}}" | grep -q "Up"; then
    echo "✅ 服务正在运行"
    docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
else
    echo "⚠️  没有运行的服务"
fi

echo ""
echo "=== 提示 ==="
echo "💡 如需调整端口配置，请修改 .env 文件中的 *_HOST_PORT 变量"

echo ""
echo "=== 验证完成 ==="
