# ===== Stage 1: 编译阶段 =====
FROM node:22-alpine AS builder

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存层加速重复构建
COPY agent-web/package.json agent-web/yarn.lock ./

# 安装依赖
RUN yarn install --frozen-lockfile

# 复制源码并编译
COPY agent-web/ ./
RUN yarn build

# ===== Stage 2: 运行阶段 =====
FROM nginx:1.22-alpine

# 给 nginx 添加权限
RUN chown root /usr/sbin/nginx && \
    chmod 757 -R /usr/sbin/nginx && \
    chmod u+s /usr/sbin/nginx && \
    chmod 777 -R /var/log/nginx/* && \
    sed -i "s@www-data@root@g" /etc/nginx/nginx.conf && \
    sed -i "s@error_log.*@error_log /dev/null;@g"  /etc/nginx/nginx.conf && \
    sed -i "s@access_log.*@access_log off;@g"  /etc/nginx/nginx.conf && \
    sed -i "s@# server_tokens.*@server_tokens off;@g" /etc/nginx/nginx.conf

EXPOSE 1101

# 复制 nginx 配置模板和启动脚本，并转换 CRLF 为 LF（避免 Windows 换行符问题）
COPY agent-web/nginx.conf.template /etc/nginx/conf.d/default.conf.template
COPY agent-web/entrypoint.sh /entrypoint.sh
RUN sed -i 's/\r$//' /entrypoint.sh /etc/nginx/conf.d/default.conf.template && \
    chmod +x /entrypoint.sh

# 从编译阶段复制构建产物
COPY --from=builder /app/dist /app/dist

ENTRYPOINT ["/entrypoint.sh"]
