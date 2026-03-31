#!/bin/sh
# 容器启动脚本：用环境变量替换 nginx 配置模板中的占位符，然后启动 nginx

# 设置默认值
BACKEND_URL="${BACKEND_URL:-https://dip.aishu.cn}"

# 使用 envsubst 替换模板中的环境变量，仅替换 BACKEND_URL 避免影响 nginx 自身变量
envsubst '${BACKEND_URL}' < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf

echo "==> BACKEND_URL: ${BACKEND_URL}"
echo "==> nginx 配置已生成，启动 nginx..."

# 启动 nginx（前台运行）
exec nginx -g 'daemon off;'
