ARG baseImage=acr.aishu.cn/public/ubuntu:22.04.20241202

FROM ${baseImage} AS prod

WORKDIR /app/agent-retrieval

COPY ./agent-retrieval /app/agent-retrieval

# 创建非root用户
RUN useradd -m -r -s /bin/bash aishu \
    && chown -R aishu:aishu /app/agent-retrieval

# 切换到非root用户
USER aishu

CMD ["./agent-retrieval"]