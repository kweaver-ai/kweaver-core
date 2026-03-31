#!/usr/bin/python3
# -*- coding:utf-8 -*-

from typing import Any, Optional
from dataclasses import dataclass

from common.configs import DataSourceType, FileOperatorType, InsertType


@dataclass
class FileOperatorRequest:
    """文件更新请求参数"""
    # 接口请求类型 create, update
    op: str
    
    # 文件类型：xlsx, xls, docx
    file_type: str
    
    # 插入类型：cover, append, append_before, append_after
    insert_type: str
    
    # 插入内容
    content: Any

    # 文件名（用于create_or_*模式）
    doc_name: Optional[str] = None
    
    # 可选参数
    new_type: Optional[str] = None  # new_row, new_col (仅Excel)
    insert_pos: Optional[int] = None  # 插入位置
    
    # 上传参数
    doc_id: Optional[str] = None
    ondup: Optional[int] = None

    doc_name_with_ext: Optional[str] = None

    # source_type 文件内容来源，full_text、url
    source_type: Optional[str] = None
    
    @classmethod
    def from_dict_update(cls, data: dict) -> 'FileOperatorRequest':
        """从字典创建请求对象"""
        return cls(
            op=FileOperatorType.UPDATE.value,
            file_type=data.get('type'),
            insert_type=data.get('insert_type'),
            content=data.get('content'),
            doc_name=data.get('name'),
            new_type=data.get('new_type'),
            insert_pos=data.get('insert_pos'),
            doc_id=data.get('docid'),
            ondup=data.get('ondup', 3)
        )
    
    @classmethod
    def from_dict_create(cls, data: dict) -> 'FileOperatorRequest':
        return cls(
            op=FileOperatorType.CREATE.value,
            file_type=data.get('type'),
            content=data.get('content'),
            doc_id=data.get('docid'),
            doc_name=data.get('name'),
            ondup=data.get('ondup'),
            new_type=data.get('new_type'),
            source_type=data.get('source_type', DataSourceType.FULLTEXT.value),
            insert_type=InsertType.COVER.value
        )
    
    def validate(self):
        """验证请求参数"""
        if not self.file_type:
            raise ValueError("file_type is required")
        if not self.insert_type:
            raise ValueError("insert_type is required")
        if self.content is None:
            raise ValueError("content is required")
        
        # 如果是create_or_*类型，需要文件名
        if self.insert_type in [InsertType.CREATE_OR_COVER.value]:
            if not self.doc_name:
                raise ValueError(f"doc_name is required for {self.insert_type} mode")
            
    def is_create_mode(self) -> bool:
        """是否是创建模式"""
        return self.insert_type in [
            InsertType.CREATE_OR_COVER.value
        ]
    
    def get_actual_insert_type(self) -> str:
        """获取实际的插入类型（去掉create_or_前缀）"""
        if self.insert_type == InsertType.CREATE_OR_COVER.value:
            return InsertType.COVER.value
        else:
            return self.insert_type