import pypandoc

from common.configs import InsertType
from .base_handler import BaseFileHandler
from models.file_operator import FileOperatorRequest
from pypdf import PdfWriter, PdfReader
from errors.errors import *


class PDFHandler(BaseFileHandler):
    """PDF文件处理器"""

    SUPPORT_INSERT_TYPES = [
        InsertType.COVER.value,
    ]

    def __init__(self):
        """
        初始化转换器
        """

        self.extra_args = [
            '--pdf-engine=xelatex',  # 使用 xelatex 支持中文
            '-V', 'mainfont=DejaVu Sans',
            '-V', 'CJKmainfont=Noto Serif CJK SC',  # Linux 中文字体
            '-V', 'monofont=DejaVu Sans Mono', # 等宽字体
            '-V', 'geometry:margin=2cm',
            '-V', 'papersize=a4',
            '-V', 'fontsize=12pt',
            '-V', 'highlight_style=tango',
            '-V', 'papersize=a4',
        ]

    def get_file_extension(self) -> str:
        return '.pdf'

    def needs_temp_file(self) -> bool:
        return True

    def supports_operation(self, request: FileOperatorRequest) -> bool:
        """检查是否支持该操作"""
        # PDF只支持文本内容插入，不支持行列操作
        return request.new_type is None and request.insert_type in self.SUPPORT_INSERT_TYPES

    def _do_update(self, source_path: str, target_path: str, request: FileOperatorRequest) -> None:
        """
        统一的更新接口
        
        Args:
            file_path: 文件路径
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
            self.convert_text(target_path, data)
        else:
            write_pdf_file_path = source_path + ".tmp"
            self.convert_text(write_pdf_file_path, data)   
            pdf_list= [source_path, write_pdf_file_path]         
            merger = PdfWriter()

            # 合并所有PDF
            for pdf in pdf_list:
                reader = PdfReader(pdf)
                merger.append(reader)

            # 先写入临时文件
            merger.write(target_path)
            merger.close()

    def convert_text(self, target_path, data):
        """
        将文本转换为PDF文件
        """

        pypandoc.convert_text(
            data.replace('\\n', '\n').replace('\\t', '\t').replace('\\r', '\r'),
            'pdf',
            format='md',
            outputfile=target_path,
            extra_args=self.extra_args
        )
