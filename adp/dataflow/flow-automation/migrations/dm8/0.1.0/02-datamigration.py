#!/usr/bin/env python3
"""
MongoDB到MariaDB数据迁移脚本
从MongoDB的flow_dag集合查询特定条件的数据，迁移到MariaDB的t_bd_resource_r表
"""

import os
from typing import Dict, List, Generator, Tuple
from pymongo import MongoClient
import rdsdriver
import logging
import sys

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)


class DatabaseManager:
    """数据库管理器"""
    
    def __init__(self):
        """
        初始化数据库管理器
        
        Args:
            config: 配置文件字典
        """
        self.mongo_client = None
        self.mariadb_conn = None
        self.mariadb_cursor = None
        
    def connect_mongodb(self) -> MongoClient:
        """
        连接MongoDB
        
        Returns:
            MongoDB客户端连接
        """
        try:
            mongodb_host = os.environ["MONGODB_HOST"]
            mongodb_port = os.environ["MONGODB_PORT"]
            mongodb_user = os.environ["MONGODB_USER"]
            mongodb_pwd = os.environ["MONGODB_PASSWORD"]
            mongodb_auth_source = os.environ["MONGODB_AUTH_SOURCE"]
            
            dns = f"mongodb://{mongodb_user}:{mongodb_pwd}@{mongodb_host}:{mongodb_port}?authSource={mongodb_auth_source}"
            
            # 创建连接
            client = MongoClient(dns, serverSelectionTimeoutMS=5000)
            
            # 测试连接
            client.admin.command('ping')
            logger.info(f"MongoDB连接成功: {mongodb_host}:{mongodb_port}")
            
            self.mongo_client = client
            return client
            
        except Exception as e:
            logger.error(f"MongoDB连接失败: {e}")
            raise
    
    def connect_mariadb(self):
        """
        连接MariaDB
        
        Returns:
            (MariaDB连接, 游标)
        """
        try:
            # 构建MariaDB连接参数
            connection_params = {
                'host': os.environ["DB_HOST"],
                'port': int(os.environ["DB_PORT"]),
                'user': os.environ["DB_USER"],
                'password': os.environ["DB_PASSWD"],
                'autocommit': True,
                'charset': 'utf8mb4'
            }
            
            # 创建连接
            conn = rdsdriver.connect(**connection_params)
            cursor = conn.cursor()
            
            logger.info(f"MariaDB连接成功: {os.environ['DB_HOST']}:{os.environ['DB_PORT']}")
            
            self.mariadb_conn = conn
            self.mariadb_cursor = cursor
            return conn, cursor
            
        except Exception as e:
            logger.error(f"MariaDB连接失败: {e}")
            raise
    
    def close_connections(self):
        """关闭所有数据库连接"""
        try:
            if self.mariadb_cursor:
                self.mariadb_cursor.close()
            if self.mariadb_conn:
                self.mariadb_conn.close()
            if self.mongo_client:
                self.mongo_client.close()
            logger.info("所有数据库连接已关闭")
        except Exception as e:
            logger.warning(f"关闭数据库连接时出错: {e}")


class DataMigrator:
    """数据迁移器"""
    
    def __init__(self, db_manager: DatabaseManager, target_db: str = 'workflow'):
        """
        初始化数据迁移器
        
        Args:
            db_manager: 数据库管理器实例
            target_db: 目标数据库名称
        """
        self.db_manager = db_manager
        self.target_db = target_db
        
    @staticmethod
    def fetch_pages(collection, query: Dict = None, fields: Dict = None, 
                   page_size: int = 100) -> Generator[Tuple[int, List[Dict]], None, None]:
        """
        生成器：逐页返回MongoDB数据
        
        Args:
            collection: MongoDB集合
            query: 查询条件
            fields: 返回字段
            page_size: 每页大小
            
        Yields:
            (页码, 数据列表)
        """
        if query is None:
            query = {}
        if fields is None:
            fields = {}
            
        page = 1
        
        while True:
            skip = (page - 1) * page_size
            cursor = collection.find(query, fields).skip(skip).limit(page_size)
            results = list(cursor)
            
            if not results:
                break
            
            yield page, results
            page += 1
    
    def insert_into_mariadb(self, cursor, resource_ids: List[str], batch_size: int = 1000) -> int:
        """
        将数据插入到MariaDB
        
        Args:
            cursor: MariaDB游标
            resource_ids: 资源ID列表
            batch_size: 批量插入大小
            
        Returns:
            插入的记录数
        """
        if not resource_ids:
            return 0
            
        inserted_count = 0
        
        # 分批处理，避免SQL语句过长
        for i in range(0, len(resource_ids), batch_size):
            batch_ids = resource_ids[i:i + batch_size]
            
            # 检查已存在的记录
            check_sql = f"""
            SELECT f_resource_id FROM {self.target_db}.t_bd_resource_r 
            WHERE f_resource_id IN (%s) AND f_resource_type = 'data_flow'
            """
            
            # 构建参数占位符
            placeholders = ', '.join(['%s'] * len(batch_ids))
            check_sql = check_sql % placeholders
            
            try:
                cursor.execute(check_sql, batch_ids)
                existing_ids = {row[0] for row in cursor.fetchall()}
                
                # 找出需要插入的resource_id
                new_ids = [rid for rid in batch_ids if rid not in existing_ids]
                
                if not new_ids:
                    continue
                
                # 批量插入
                insert_sql = f"""
                INSERT INTO "{self.target_db}"."t_bd_resource_r" 
                ("created_at", "updated_at", "f_bd_id", "f_resource_id", "f_resource_type", "f_create_by") 
                VALUES(current_timestamp(6), current_timestamp(6), 'bd_public', %s, 'data_flow', '-');
                """
                
                cursor.executemany(insert_sql, [(rid,) for rid in new_ids])
                inserted_count += len(new_ids)
                
                logger.info(f"批量插入 {len(new_ids)} 条记录，累计插入 {inserted_count} 条")
                
            except Exception as e:
                logger.error(f"插入数据时出错: {e}")
                # 可以选择记录失败的批次并继续
                continue
        
        return inserted_count
    
    def migrate_data(self, collection_name: str = 'flow_dag'):
        """
        执行数据迁移
        
        Args:
            collection_name: MongoDB集合名称
        """
        try:
            # 获取MongoDB集合
            db = self.db_manager.mongo_client['automation']
            collection = db[collection_name]
            
            logger.info(f"开始迁移数据，源集合: {'automation'}.{collection_name}")
            
            # 构建查询条件
            filter_condition = {
                "$or": [
                    {"biz_domain_id": {"$exists": False}},  # 字段不存在
                    {"biz_domain_id": None},                # 字段值为 None
                    {"biz_domain_id": ""}                   # 字段值为空字符串
                ]
            }
            
            # 查询字段
            fields = {'_id': 1, 'type': 1}
            
            total_records = 0
            total_inserted = 0
            
            # 分页处理数据
            for page, data in self.fetch_pages(collection, filter_condition, fields, page_size=100):
                resource_ids = []
                
                logger.info(f"处理第 {page} 页，共 {len(data)} 条记录")
                
                # 提取resource_id
                for doc in data:
                    _id = doc['_id']
                    doc_type = doc.get('type', 'default')
                    if doc_type == 'combo-operator':
                        continue
                    resource_ids.append(f"{_id}:{doc_type}")
                
                # 插入到MariaDB
                inserted = self.insert_into_mariadb(self.db_manager.mariadb_cursor, resource_ids)
                total_inserted += inserted
                total_records += len(data)
            
            logger.info(f"数据迁移完成！总共处理 {total_records} 条记录，成功插入 {total_inserted} 条记录")
            
        except Exception as e:
            logger.error(f"数据迁移过程中出错: {e}")
            raise


def main():
    """主函数"""
    
    try:     
        # 1. 初始化数据库管理器
        logger.info("初始化数据库连接...")
        db_manager = DatabaseManager()
        
        # 2. 连接数据库
        db_manager.connect_mongodb()
        db_manager.connect_mariadb()
        
        # 3. 执行数据迁移
        logger.info("开始数据迁移...")
        migrator = DataMigrator(db_manager, target_db='model_management')
        migrator.migrate_data(collection_name='flow_dag')
        
        logger.info("数据迁移任务执行成功！")
        
    except Exception as e:
        logger.error(f"程序执行失败: {e}")
        return 1
    finally:
        # 确保关闭数据库连接
        if 'db_manager' in locals():
            db_manager.close_connections()
    
    return 


if __name__ == "__main__":
    exit_code = main()
    sys.exit(exit_code)