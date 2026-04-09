#!/usr/bin/env python3
"""
OpenSearch索引删除脚本
从环境变量获取连接信息，连接OpenSearch并删除指定索引
"""

import os
import logging
from opensearchpy import OpenSearch


INDEX_TO_DELETE = "adp_bkn_concept_dataset"


# 配置日志
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


def get_client(
    host: str, port: int, user: str, password: str, protocol: str
) -> OpenSearch:
    """创建OpenSearch客户端"""
    try:
        client = OpenSearch(
            hosts=[{"host": host, "port": port, "scheme": protocol}],
            http_auth=(user, password),
            use_ssl=protocol == "https",
            verify_ssl=protocol == "https",
            ssl_assert_hostname=False,
            ssl_show_warn=False,
            timeout=30,
        )
        logger.info(f"OpenSearch客户端创建成功: {protocol}://{host}:{port}")

        info = client.info()
        logger.info(
            f"OpenSearch连接成功，版本: {info.get('version', {}).get('number', 'unknown')}"
        )
        return client
    except Exception as e:
        logger.error(f"创建OpenSearch客户端失败: {e}")
        raise e


def delete_index(client: OpenSearch, index_name: str) -> dict:
    """删除指定的索引"""
    logger.info(f"开始删除索引: {index_name}")

    try:
        # 删除索引
        response = client.indices.delete(index=index_name, ignore_unavailable=True)
        if response.get("acknowledged", False):
            logger.info(f"索引删除成功: {index_name}")
        else:
            logger.error(f"索引删除未确认: {index_name}, 响应: {response}")
    except Exception as e:
        logger.error(f"删除索引时发生错误 {index_name}: {e}")
        raise e


if __name__ == "__main__":
    if os.environ.get("CI_MODE") == "true":
        logger.info("CI_MODE 为 true，跳过 当前升级文件")
    else:
        logger.info("正在创建OpenSearch客户端...")
        client = get_client(
            os.environ["OPENSEARCH_HOST"],
            int(os.environ["OPENSEARCH_PORT"]),
            os.environ["OPENSEARCH_USER"],
            os.environ["OPENSEARCH_PASSWORD"],
            os.environ["OPENSEARCH_PROTOCOL"],
        )

        logger.info(f"准备删除索引: {INDEX_TO_DELETE}")
        delete_index(client, INDEX_TO_DELETE)
        logger.info(f"删除索引 {INDEX_TO_DELETE} 操作完成")
