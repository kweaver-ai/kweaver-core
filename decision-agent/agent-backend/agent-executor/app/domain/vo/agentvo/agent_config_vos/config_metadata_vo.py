# -*- coding:utf-8 -*-
from typing import Optional

from pydantic import BaseModel, validator


class ConfigMetadataVo(BaseModel):
    """配置元数据模型 - 系统维护的agent配置元数据"""

    config_tpl_version: str = ""
    config_last_set_timestamp: Optional[int] = None

    @validator("config_tpl_version")
    def validate_config_tpl_version(cls, v):
        """验证配置模板版本"""
        return v

    @validator("config_last_set_timestamp")
    def validate_config_last_set_timestamp(cls, v):
        """验证配置最后设置时间戳"""
        if v is not None and not isinstance(v, int):
            return 0
        return v

    @property
    def config_last_set_timestamp_str(self) -> str:
        """配置最后设置时间戳的字符串表示"""
        if self.config_last_set_timestamp is None:
            return ""
        return str(self.config_last_set_timestamp)
