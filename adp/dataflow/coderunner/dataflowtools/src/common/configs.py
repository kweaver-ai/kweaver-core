import os
from enum import Enum
import tempfile


db_configs = {
    "type": os.getenv("DB_TYPE", "MariaDB"),
    "name": os.getenv("DB_NAME", "kweaver"),
    "host": os.getenv("DB_HOST", "mariadb-mariadb-master.resource"),
    "port": os.getenv("DB_PORT", "3330"),
    "user": os.getenv("DB_USER", "anyshare"),
    "password": os.getenv("DB_PASSWORD", "xxxxxxx"),
    "charset": os.getenv("DB_CHARSET", "utf8mb4")
}

hydra_configs = {
    "admin_host": os.getenv("OAUTH_ADMIN_HOST", "hydra-admin.anyshare"),
    "admin_port": os.getenv("OAUTH_ADMIN_PORT", "4445"),
}

user_management_configs = {
    "private_host": os.getenv("USERMANAGEMENT_PRIVATE_HOST", "user-management-private.anyshare"),
    "private_port": os.getenv("USERMANAGEMENT_PRIVATE_PORT", "30980"),
}

oss_management_configs = {
    "private_host": os.getenv("OSSGATEWAY_PRIVATE_HOST", "ossgatewaymanager-private.anyshare"),
    "private_port": os.getenv("OSSGATEWAY_PRIVATE_PORT", "9002"),
}

t4th_configs = {
    "protocol": os.getenv("T4TH_PORTOCOL", "http"),
    "host": os.getenv("T4TH_HOST", ""),
    "port": os.getenv("T4TH_PORT", ""),
}

open_doc_configs = {
    "host": os.getenv("OPENDOC_PUBLIC_HOST", "open-doc-public.anyshare"),
    "port": os.getenv("OPENDOC_PUBLIC_PORT", "30998"),
}

docset_configs = {
    "private_host": os.getenv("DOCSET_PRIVATE_HOST", "docset-private.anyshare"),
    "private_port": os.getenv("DOCSET_PRIVATE_PORT", "32597"),
}

mq_configs = {
    "host": os.getenv("MQ_HOST", "proton-mq-nsq-nsqd.resource"),
    "port": os.getenv("MQ_PORT", "4151"),
    "lookupd_host": os.getenv("MQ_LOOKUPD_HOST", "proton-mq-nsq-nsqlookupd.resource"),
    "lookupd_port": os.getenv("MQ_LOOKUPD_PORT", "4161"),
    "connector_type": os.getenv("MQ_CONNECTOR_TYPE", "nsq"),
}

redis_configs = {
    "redis_cluster_mode": os.getenv("REDIS_CLUSTER_MODE", "sentinel"),
    "redis_host": os.getenv("REDIS_HOST", "proton-redis-proton-redis-sentinel.resource"),
    "redis_port": os.getenv("REDIS_PORT", "26379"),
    "redis_user": os.getenv("REDIS_USERNAME", "root"),
    "redis_password": os.getenv("REDIS_PASSWORD", "xxxxxx"),
    "redis_sentinel_host": os.getenv("REDIS_SENTINEL_HOST", "proton-redis-proton-redis-sentinel.resource"),
    "redis_sentinel_port": os.getenv("REDIS_SENTINEL_PORT", "26379"),
    "redis_sentinel_user": os.getenv("REDIS_SENTINEL_USERNAME", "root"),
    "redis_sentinel_password": os.getenv("REDIS_SENTINEL_PASSWORD", "xxxxxx"),
    "redis_master_name": os.getenv("REDIS_MASTER_GROUPNAME", "mymaster"),
}

def is_auth_enable():
    return os.getenv("AUTH_ENABLED", "false").lower() == "true"

# RedisMode Redis 连接模式
class RedisMode(Enum):
    """Redis 连接模式枚举"""
    STANDALONE = "standalone"
    SENTINEL = "sentinel"
    CLUSTER = "cluster"
    MASTER_SLAVE = "master_slave"


# MAX_FILE_SIZE 更新文件操作时文件大小限制参数
MAX_FILE_SIZE= 100 * 1024 * 1024
# DEFAULT_CSF_LEVEL 上传文件时的默认等级
DEFAULT_CSF_LEVEL= 0


# InsertType 插入操作类型
class InsertType(Enum):
    COVER = 'cover'
    APPEND = 'append'
    APPEND_AFTER = 'append_after'
    APPEND_BEFORE = 'append_before'
    CREATE_OR_APPEND = 'create_or_append'
    CREATE_OR_COVER = 'create_or_cover'


# DataType 操作类型
class DataType(Enum):
    NEW_ROW = 'new_row'
    NEW_COL = 'new_col'
    NEW_CONTENT = 'new_content'


# FileMimeType 文件类型配置
class FileMimeType(Enum):
    XLSX = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    XLS = 'application/vnd.ms-excel'
    DOCX = 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    PDF = 'application/pdf'
    MD = 'text/markdown'

    @classmethod
    def get_mimetype(cls, file_type):
        return getattr(cls, file_type.upper(), cls.XLSX)


# CacheConfig 缓存配置
class CacheConfig:
    # 缓存根目录
    CACHE_ROOT_DIR = os.path.join(tempfile.gettempdir(), 'file_cache')

    # 缓存过期时间（秒）
    CACHE_EXPIRE_TIME = 600  # 5分钟

    CACHE_FILE_META_NAME = 'file_meta.json'


class DataSourceType(Enum):
    FULLTEXT = 'full_text'
    URL = 'url'

class FileOperatorType(Enum):
    UPDATE = 'update'
    CREATE = 'create'