"""
数据库相关配置
"""

from dataclasses import dataclass
from typing import Optional


@dataclass
class RdsConfig:
    """关系型数据库配置"""

    host: Optional[str] = None
    port: int = 3330
    dbname: Optional[str] = None
    user: Optional[str] = None
    password: Optional[str] = None
    db_type: str = ""

    @classmethod
    def from_dict(cls, data: dict) -> "RdsConfig":
        """从字典创建配置对象"""
        return cls(
            host=data.get("host"),
            port=int(data.get("port", 3330)),
            dbname=data.get("dbname"),
            user=data.get("user"),
            password=data.get("password"),
            db_type=data.get("db_type", ""),
        )


@dataclass
class RedisConfig:
    """Redis数据库配置"""

    # 连接模式: standalone, sentinel, master-slave
    cluster_mode: str = ""

    # 通用配置
    host: str = ""
    port: str = ""
    user: str = ""
    password: str = ""

    # Sentinel模式配置
    sentinel_master: str = ""
    sentinel_user: str = ""
    sentinel_password: str = ""

    # Master-Slave模式配置
    read_host: str = ""
    read_port: str = ""
    read_user: str = ""
    read_password: str = ""
    write_host: str = ""
    write_port: str = ""
    write_user: str = ""
    write_password: str = ""

    @classmethod
    def from_dict(cls, data: dict) -> "RedisConfig":
        """从字典创建配置对象"""
        return cls(
            cluster_mode=str(data.get("cluster_mode", "")),
            host=data.get("host", ""),
            port=str(data.get("port", "")),
            user=data.get("user", ""),
            password=str(data.get("password", "")),
            sentinel_master=str(data.get("sentinel_master", "")),
            sentinel_user=str(data.get("sentinel_user", "")),
            sentinel_password=str(data.get("sentinel_password", "")),
            read_host=data.get("read_host", ""),
            read_port=str(data.get("read_port", "")),
            read_user=data.get("read_user", ""),
            read_password=str(data.get("read_password", "")),
            write_host=data.get("write_host", ""),
            write_port=str(data.get("write_port", "")),
            write_user=data.get("write_user", ""),
            write_password=str(data.get("write_password", "")),
        )


@dataclass
class GraphDBConfig:
    """图数据库配置"""

    host: str = ""
    port: str = ""
    type: str = "nebulaGraph"
    read_only_user: str = ""
    read_only_password: str = ""

    @classmethod
    def from_dict(cls, data: dict) -> "GraphDBConfig":
        """从字典创建配置对象"""
        return cls(
            host=data.get("host", ""),
            port=str(data.get("port", "")),
            type=data.get("type", "nebulaGraph"),
            read_only_user=data.get("read_only_user", ""),
            read_only_password=data.get("read_only_password", ""),
        )


@dataclass
class OpenSearchConfig:
    """OpenSearch配置"""

    host: str = ""
    port: str = ""
    user: str = ""
    password: str = ""

    @classmethod
    def from_dict(cls, data: dict) -> "OpenSearchConfig":
        """从字典创建配置对象"""
        return cls(
            host=data.get("host", ""),
            port=str(data.get("port", "")),
            user=data.get("user", ""),
            password=data.get("password", ""),
        )
