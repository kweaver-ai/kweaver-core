#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright The kweaver.ai Authors.
#
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the project root for details.

"""
OpenSearch索引清理脚本
功能：检查所有OpenSearch中的索引，删除不在失败或取消的job对应的task中的索引
逻辑：从失败或取消的job查找对应的task，取出这些task中的索引作为有效索引
"""

import os
import sys
import logging
import argparse
from typing import Set, List, Tuple, Dict
import pymysql
from opensearchpy import OpenSearch, RequestsHttpConnection

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)],
)
logger = logging.getLogger(__name__)


class OpenSearchIndexCleaner:
    """OpenSearch索引清理器"""

    def __init__(self, db_config: dict, os_config: dict, dry_run: bool = False):
        """
        初始化清理器

        Args:
            db_config: 数据库配置
            os_config: OpenSearch配置
            dry_run: 是否为试运行模式（不实际删除）
        """
        self.db_config = db_config
        self.os_config = os_config
        self.dry_run = dry_run
        self.db_conn = None
        self.os_client = None

    def connect_database(self) -> bool:
        """连接数据库"""
        try:
            self.db_conn = pymysql.connect(
                host=self.db_config["host"],
                port=self.db_config["port"],
                user=self.db_config["user"],
                password=self.db_config["password"],
                database="adp",
                charset="utf8mb4",
                cursorclass=pymysql.cursors.DictCursor,
            )
            logger.info("数据库连接成功")
            return True
        except Exception as e:
            logger.error(f"数据库连接失败: {e}")
            return False

    def connect_opensearch(self) -> bool:
        """连接OpenSearch"""
        try:
            self.os_client = OpenSearch(
                hosts=[
                    {"host": self.os_config["host"], "port": self.os_config["port"]}
                ],
                http_auth=(
                    self.os_config.get("user", ""),
                    self.os_config.get("password", ""),
                ),
                use_ssl=self.os_config.get("protocol", "http") == "https",
                verify_certs=False,
                ssl_show_warn=False,
                connection_class=RequestsHttpConnection,
                timeout=30,
            )

            # 测试连接
            if self.os_client.ping():
                logger.info("OpenSearch连接成功")
                return True
            else:
                logger.error("OpenSearch连接失败: ping超时")
                return False
        except Exception as e:
            logger.error(f"OpenSearch连接失败: {e}")
            return False

    def get_valid_indexes_from_db(self) -> Set[str]:
        """
        从数据库获取有效索引列表
        逻辑：从运行中的job的task中获取索引，以及object_type_status中的索引

        Returns:
            有效索引名称集合
        """
        valid_indexes = set()

        try:
            with self.db_conn.cursor() as cursor:
                # 从运行中的job的task中获取索引，以及object_type_status中的索引
                sql = """
                SELECT t.f_index
                FROM t_kn_task t
                INNER JOIN t_kn_job j ON t.f_job_id = j.f_id
                WHERE j.f_state IN ('running')
                AND t.f_index != ''
                UNION ALL
                SELECT f_index FROM t_object_type_status
                WHERE f_index != ''
                """
                cursor.execute(sql)
                results = cursor.fetchall()

                for row in results:
                    index_name = row["f_index"]
                    if index_name:
                        valid_indexes.add(index_name)

                logger.info(f"从数据库中获取到 {len(valid_indexes)} 个有效索引")
                return valid_indexes

        except Exception as e:
            logger.error(f"从数据库获取索引列表失败: {e}")
            return set()

    def get_all_opensearch_indexes_with_size(self) -> Dict[str, dict]:
        """
        使用cat API获取OpenSearch中所有索引及其大小信息

        Returns:
            索引信息字典 {index_name: {'docs': count, 'size': bytes}}
        """
        index_info = {}

        try:
            # 使用cat API一次性获取索引信息，bytes='b'直接返回字节数
            response = self.os_client.cat.indices(
                index="adp-kn_ot_index-*,dip-kn_ot_index-*",
                format="json",
                h="index,docsCount,storeSize",
                bytes="b",
            )

            for item in response:
                index_name = item["index"]

                index_info[index_name] = {
                    "docs": int(item.get("docsCount", 0)),
                    "size": int(item.get("storeSize", 0)),
                }

            logger.info(f"从OpenSearch获取到 {len(index_info)} 个业务知识网络相关索引")
            return index_info

        except Exception as e:
            logger.error(f"从OpenSearch获取索引列表失败: {e}")
            return {}

    def find_orphan_indexes(
        self,
        db_valid_indexes: Set[str],
        os_indexes: Set[str],
    ) -> Set[str]:
        """
        找出需要删除的索引（在OpenSearch中存在且不在使用的索引）

        Args:
            db_valid_indexes: 从数据库中获取的有效索引
            os_indexes: OpenSearch中的所有索引

        Returns:
            需要删除的索引集合（在OpenSearch中存在且不在使用的索引）
        """
        # 从 os_indexes 中排除 db_valid_indexes 中的索引
        orphan_indexes = os_indexes - db_valid_indexes
        logger.info(f"发现 {len(orphan_indexes)} 个需要删除的索引")

        return orphan_indexes

    def delete_indexes(
        self, indexes: List[str], index_info: Dict[str, dict]
    ) -> Tuple[int, int, int]:
        """
        删除指定的索引

        Args:
            indexes: 要删除的索引列表
            index_info: 索引信息字典 {index_name: {'docs': count, 'size': bytes}}

        Returns:
            (成功删除数量, 失败数量, 回收的磁盘大小(字节))
        """
        success_count = 0
        failed_count = 0
        total_recovered_size = 0

        for index_name in indexes:
            try:
                # 从预获取的索引信息中获取大小
                info = index_info.get(index_name, {})
                index_size = info.get("size", 0)

                if self.dry_run:
                    logger.info(
                        f"[DRY RUN] 将删除索引: {index_name} (大小: {self._format_bytes(index_size)})"
                    )
                    success_count += 1
                    total_recovered_size += index_size
                else:
                    if self.os_client.indices.exists(index=index_name):
                        response = self.os_client.indices.delete(index=index_name)
                        if response.get("acknowledged", False):
                            logger.info(
                                f"成功删除索引: {index_name} (回收: {self._format_bytes(index_size)})"
                            )
                            success_count += 1
                            total_recovered_size += index_size
                        else:
                            logger.warning(f"删除索引失败（未确认）: {index_name}")
                            failed_count += 1
                    else:
                        logger.warning(f"索引不存在: {index_name}")
                        failed_count += 1
            except Exception as e:
                logger.error(f"删除索引 {index_name} 失败: {e}")
                failed_count += 1

        return success_count, failed_count, total_recovered_size

    def cleanup(self) -> bool:
        """执行清理操作"""
        logger.info("=" * 60)
        logger.info("开始OpenSearch索引清理")
        logger.info(f"试运行模式: {'是' if self.dry_run else '否'}")
        logger.info("=" * 60)

        # 连接数据库
        if not self.connect_database():
            return False

        # 连接OpenSearch
        if not self.connect_opensearch():
            return False

        # 获取从数据库中获取的有效索引
        db_valid_indexes = self.get_valid_indexes_from_db()

        # 获取OpenSearch中的所有索引及其大小信息
        index_info = self.get_all_opensearch_indexes_with_size()
        if not index_info:
            logger.warning("OpenSearch中没有找到任何索引")
            return False

        # 从索引信息中提取索引名称集合
        os_indexes = set(index_info.keys())

        # 找出孤立的索引
        orphan_indexes = self.find_orphan_indexes(db_valid_indexes, os_indexes)

        if not orphan_indexes:
            logger.info("没有发现需要删除的索引，清理完成")
            return True

        # 显示需要删除的索引信息
        logger.info("\n需要删除的索引列表:")
        logger.info("-" * 60)
        for index_name in sorted(orphan_indexes):
            info = index_info.get(index_name, {})
            logger.info(
                f"  - {index_name} (文档数: {info.get('docs', 0):,}, 存储大小: {self._format_bytes(info.get('size', 0))})"
            )
        logger.info("-" * 60)

        # 删除索引
        logger.info(f"\n开始删除 {len(orphan_indexes)} 个索引...")
        success_count, failed_count, recovered_size = self.delete_indexes(
            list(orphan_indexes), index_info
        )

        # 输出结果
        logger.info("\n" + "=" * 60)
        logger.info("清理完成")
        logger.info(f"成功删除: {success_count} 个")
        logger.info(f"删除失败: {failed_count} 个")
        logger.info(f"回收磁盘空间: {self._format_bytes(recovered_size)}")
        logger.info("=" * 60)

        return True

    def _format_bytes(self, bytes_size: int) -> str:
        """
        格式化字节大小

        Args:
            bytes_size: 字节大小

        Returns:
            格式化后的字符串
        """
        for unit in ["B", "KB", "MB", "GB", "TB"]:
            if bytes_size < 1024.0:
                return f"{bytes_size:.2f} {unit}"
            bytes_size /= 1024.0
        return f"{bytes_size:.2f} PB"

    def close(self):
        """关闭连接"""
        if self.db_conn:
            self.db_conn.close()
            logger.info("数据库连接已关闭")


def load_config_from_env() -> Tuple[dict, dict]:
    """
    从环境变量加载配置

    Returns:
        (数据库配置, OpenSearch配置)
    """
    # 数据库配置
    db_config = {
        "host": os.getenv("DB_HOST", "localhost"),
        "port": int(os.getenv("DB_PORT", "3306")),
        "user": os.getenv("DB_USER", "root"),
        "password": os.getenv("DB_PASSWORD", ""),
    }

    # OpenSearch配置
    os_config = {
        "host": os.getenv("OPENSEARCH_HOST", "localhost"),
        "port": int(os.getenv("OPENSEARCH_PORT", "9200")),
        "protocol": os.getenv("OPENSEARCH_PROTOCOL", "http"),
        "user": os.getenv("OPENSEARCH_USER", ""),
        "password": os.getenv("OPENSEARCH_PASSWORD", ""),
    }

    return db_config, os_config


def main():
    """主函数"""
    parser = argparse.ArgumentParser(
        description="OpenSearch索引清理工具 - 删除不在失败或取消的job的task中的索引"
    )
    parser.add_argument(
        "--dry-run", action="store_true", help="试运行模式，不实际删除索引"
    )
    parser.add_argument("--db-host", help="数据库主机地址（默认从环境变量DB_HOST读取）")
    parser.add_argument(
        "--db-port", type=int, help="数据库端口（默认从环境变量DB_PORT读取）"
    )
    parser.add_argument("--db-user", help="数据库用户名（默认从环境变量DB_USER读取）")
    parser.add_argument(
        "--db-password", help="数据库密码（默认从环境变量DB_PASSWORD读取）"
    )
    parser.add_argument(
        "--os-host", help="OpenSearch主机地址（默认从环境变量OPENSEARCH_HOST读取）"
    )
    parser.add_argument(
        "--os-port",
        type=int,
        help="OpenSearch端口（默认从环境变量OPENSEARCH_PORT读取）",
    )
    parser.add_argument(
        "--os-user", help="OpenSearch用户名（默认从环境变量OPENSEARCH_USER读取）"
    )
    parser.add_argument(
        "--os-password", help="OpenSearch密码（默认从环境变量OPENSEARCH_PASSWORD读取）"
    )
    parser.add_argument(
        "--os-protocol",
        choices=["http", "https"],
        help="OpenSearch协议（默认从环境变量OPENSEARCH_PROTOCOL读取）",
    )

    args = parser.parse_args()

    # 加载配置
    db_config, os_config = load_config_from_env()

    # 命令行参数覆盖环境变量
    if args.db_host:
        db_config["host"] = args.db_host
    if args.db_port:
        db_config["port"] = args.db_port
    if args.db_user:
        db_config["user"] = args.db_user
    if args.db_password:
        db_config["password"] = args.db_password

    if args.os_host:
        os_config["host"] = args.os_host
    if args.os_port:
        os_config["port"] = args.os_port
    if args.os_user:
        os_config["user"] = args.os_user
    if args.os_password:
        os_config["password"] = args.os_password
    if args.os_protocol:
        os_config["protocol"] = args.os_protocol

    # 验证必需的配置
    if not db_config["password"]:
        logger.error(
            "数据库密码未设置，请设置DB_PASSWORD环境变量或使用--db-password参数"
        )
        sys.exit(1)

    if not os_config["password"]:
        logger.error(
            "OpenSearch密码未设置，请设置OPENSEARCH_PASSWORD环境变量或使用--os-password参数"
        )
        sys.exit(1)

    # 创建清理器并执行清理
    cleaner = OpenSearchIndexCleaner(
        db_config=db_config, os_config=os_config, dry_run=args.dry_run
    )

    try:
        success = cleaner.cleanup()
        sys.exit(0 if success else 1)
    except KeyboardInterrupt:
        logger.info("\n用户中断操作")
        sys.exit(130)
    except Exception as e:
        logger.error(f"清理过程中发生错误: {e}", exc_info=True)
        sys.exit(1)
    finally:
        cleaner.close()


if __name__ == "__main__":
    main()
