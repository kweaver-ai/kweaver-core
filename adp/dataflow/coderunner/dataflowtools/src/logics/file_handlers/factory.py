#!/usr/bin/python3
# -*- coding:utf-8 -*-

from typing import Dict, Type
from .base_handler import BaseFileHandler
from .excel_handlers import XlsxHandler, XlsHandler
from .word_handler import DocxHandler
from .pdf_handler import PDFHandler
from .md_handler import MDHandler
from errors.errors import BadParameterException

class FileHandlerFactory:
    """文件处理器工厂"""
    
    _handlers: Dict[str, Type[BaseFileHandler]] = {
        'xlsx': XlsxHandler,
        'xls': XlsHandler,
        'docx': DocxHandler,
        'pdf': PDFHandler,
        'md': MDHandler
    }
    
    @classmethod
    def register_handler(cls, extension: str, handler_class: Type[BaseFileHandler]):
        """注册新的文件处理器"""
        cls._handlers[extension] = handler_class
    
    @classmethod
    def get_handler(cls, file_extension: str) -> BaseFileHandler:
        """根据文件扩展名获取处理器"""
        handler_class = cls._handlers.get(file_extension)
        if not handler_class:
            raise BadParameterException(detail= f"Unsupported file type: {file_extension}, supported types: {list(cls._handlers.keys())}")
        return handler_class()
    
    @classmethod
    def get_supported_extensions(cls) -> list:
        """获取支持的文件扩展名列表"""
        return list(cls._handlers.keys())