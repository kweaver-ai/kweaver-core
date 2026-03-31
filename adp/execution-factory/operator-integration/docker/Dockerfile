ARG baseImage=acr.aishu.cn/public/ubuntu:22.04.20241202

FROM ${baseImage} AS prod

WORKDIR /app/operator-integration

RUN ls -a
RUN ls ..
RUN ls ./
COPY ./operator-integration /app/operator-integration

# 创建非root用户
RUN useradd -m -r -s /bin/bash aishu \
    && chown -R aishu:aishu /app/operator-integration

# 切换到非root用户
USER aishu

CMD ["./operator-integration"]