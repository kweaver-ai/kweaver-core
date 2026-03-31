FROM acr.aishu.cn/public/nginx:1.22.1-alpine-latest

# 给nginx添加权限
RUN chown root /usr/sbin/nginx && \
    chmod 757 -R /usr/sbin/nginx && \
    chmod u+s /usr/sbin/nginx && \
    chmod 777 -R /var/log/nginx/* && \
    sed -i "s@www-data@root@g" /etc/nginx/nginx.conf && \
    sed -i "s@error_log.*@error_log /dev/null;@g"  /etc/nginx/nginx.conf && \
    sed -i "s@access_log.*@access_log off;@g"  /etc/nginx/nginx.conf && \
    sed -i "s@# server_tokens.*@server_tokens off;@g" /etc/nginx/nginx.conf

EXPOSE 1101

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY dist /app/dist
