#!/usr/bin/python3
# -*- coding:utf-8 -*-

from abc import ABC, abstractmethod
import os
import tempfile

from jsonschema import ValidationError
from models.file_operator import FileOperatorRequest


class BaseFileHandler(ABC):
    """文件处理器基类"""
    
    def __init__(self):
        self.temp_file_path = None
    
    def update(self, file_path: str, request: FileOperatorRequest) -> None:
        """
        统一的文件更新接口 - 使用安全更新包装器
        
        Args:
            file_path: 要修改的文件路径
            request: 更新请求参数
        """
        if not self.supports_operation(request):
            raise ValidationError(f"{self.__class__.__name__} does not support operation: new_type={request.new_type}, insert_type={request.insert_type}")
        
        # 使用安全更新包装器
        self.safe_update_file(file_path, request)

    def safe_update_file(self, file_path: str, request: FileOperatorRequest) -> None:
        """
        安全地更新文件（公共方法）
        
        策略：
        1. 创建临时文件
        2. 调用子类的具体更新逻辑
        3. 原子性替换原文件
        
        Args:
            file_path: 目标文件路径
            request: 更新请求
        """
        try:
            # 获取文件扩展名
            file_ext = self.get_file_extension()

            if self.needs_temp_file():
                dir_name = os.path.dirname(file_path) or '.'

                temp_file = tempfile.NamedTemporaryFile(
                    mode='wb',
                    dir=dir_name,
                    suffix=file_ext,
                    prefix=f'.tmp_',
                    delete=False
                )
                temp_path = temp_file.name
                temp_file.close()
            
                # 调用子类的具体更新逻辑
                # 子类应该实现这个方法，将更新后的内容写入temp_path
                self._do_update(file_path, temp_path, request)
                
                # 原子性地替换原文件
                os.replace(temp_path, file_path)   
            else:
                self._do_update(file_path, '', request)        
        except Exception as e:
            # 清理临时文件
            if os.path.exists(temp_path):
                try:
                    os.remove(temp_path)
                except:
                    pass
            raise e
    
    @abstractmethod
    def _do_update(self, source_path: str, target_path: str, request: FileOperatorRequest) -> None:
        """
        执行具体的文件更新逻辑（子类实现）
        
        Args:
            source_path: 源文件路径（读取）
            target_path: 目标文件路径（写入）
            request: 更新请求
            
        Note:
            子类应该：
            1. 从 source_path 读取原始内容
            2. 根据 request 进行修改
            3. 将结果写入 target_path
        """
        pass
    
    @abstractmethod
    def get_file_extension(self) -> str:
        """获取文件扩展名"""
        pass
    
    def needs_temp_file(self) -> bool:
        """是否需要临时文件"""
        pass
    
    @abstractmethod
    def supports_operation(self, request: FileOperatorRequest) -> bool:
        """
        检查是否支持该操作
        
        Args:
            request: 更新请求参数
            
        Returns:
            是否支持
        """
        pass