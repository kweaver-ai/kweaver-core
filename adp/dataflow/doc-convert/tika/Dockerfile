ARG BASE_IMAGE=acr.aishu.cn/dip/tika:3.3.0.0-20260325-1
FROM ${BASE_IMAGE}

USER root

# 安装开源字体（覆盖常见 PDF 字体需求）
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    fonts-liberation \
    fonts-liberation2 \
    fonts-dejavu-core \
    fonts-noto-core \
    fonts-noto-cjk \
    fonts-noto-cjk-extra \
    fonts-wqy-zenhei \
    fonts-wqy-microhei \
    fontconfig \
    && fc-cache -fv \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# 把新增的自定义配置文件复制到镜像内部的执行目录
COPY tika-config.xml /config/tika-config.xml
ENV TIKA_CONFIG="/config/tika-config.xml"

USER 35002:35002

# 显式声明启动命令参数（结合原始的 Entrypoint 一同执行启动服务）
CMD ["-c", "/config/tika-config.xml"]
