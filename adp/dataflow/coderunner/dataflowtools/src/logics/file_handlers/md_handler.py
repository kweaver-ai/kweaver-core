from common.configs import InsertType
from .base_handler import BaseFileHandler
from models.file_operator import FileOperatorRequest

class MDHandler(BaseFileHandler):
    """MD文件处理器"""

    SUPPORT_INSERT_TYPES = [
        InsertType.APPEND.value,
        InsertType.APPEND_BEFORE.value,
        InsertType.APPEND_AFTER.value,
        InsertType.COVER.value,
    ]
    
    def get_file_extension(self) -> str:
        return '.md'

    def needs_temp_file(self) -> bool:
        return False

    def supports_operation(self, request: FileOperatorRequest) -> bool:
        """检查是否支持该操作"""
        # MD只支持文本内容插入，不支持行列操作
        return request.new_type is None and request.insert_type in self.SUPPORT_INSERT_TYPES

    def _do_update(self, source_path: str, target_path: str, request: FileOperatorRequest) -> None:
        """
        统一的更新接口
        
        Args:
            file_path: 临时文件路径
            request: 更新请求
            
        Returns:
            更新后的数据
        """
        
        content = request.content
        data = ''
        if isinstance(content, str):
            data = content
        elif isinstance(content, list):
            data = '\n'.join(str(item) for item in content)
        else:
            data = str(content)

        # 处理文档内容
        if request.insert_type == InsertType.COVER.value:
            self._write(source_path, data, mode='w')
        else:
            self._write(source_path, data, mode='a')
    
    @staticmethod
    def _write(file_path: str, content: str, mode: str = 'w'):
        """写入内容到 MD 文件
        
        Args:
            content: 要写入的内容
            mode: 'w' 覆盖写入, 'a' 追加写入
        """
        with open(file_path, mode, encoding='utf-8') as f:
            f.write(content)