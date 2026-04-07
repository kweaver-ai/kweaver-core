#!/usr/bin/python3
# -*- coding:utf-8 -*-

import io
import os
import openpyxl

from pypdf import PdfWriter
from docx import Document


class TemplateManager:
    """模板文件管理器"""
    
    @staticmethod
    def create_empty_file(file_type: str, source_path: str):
        """
        创建空文件内容
        
        Args:
            file_type: 文件类型
            
        Returns:
            空文件字节内容
        """
        # contents = b''
        if file_type in 'xlsx':
            contents = TemplateManager._create_empty_excel(file_type)
        elif file_type == 'docx':
            contents = TemplateManager._create_empty_docx()
        elif file_type == 'pdf':
            contents = TemplateManager._create_empty_pdf()
        elif file_type in 'md':
            contents = TemplateManager._create_empty_markdown()
        
        dir_path = os.path.dirname(source_path)
        if not os.path.exists(dir_path):
            os.makedirs(dir_path)

        with open(source_path, 'wb') as f:
            f.write(contents)
    
    @staticmethod
    def _create_empty_excel(file_type: str) -> bytes:
        """创建空Excel文件"""
        wb = openpyxl.Workbook()
        output = io.BytesIO()
        wb.save(output)
        return output.getvalue()
    
    @staticmethod
    def _create_empty_docx() -> bytes:
        """创建空Word文档"""
        doc = Document()
        output = io.BytesIO()
        doc.save(output)
        return output.getvalue()
    
    @staticmethod
    def _create_empty_pdf() -> bytes: 
        writer = PdfWriter()
        writer.add_blank_page(width=595, height=842)
        with io.BytesIO() as buffer:
            writer.write(buffer)
            return buffer.getvalue()
    
    @staticmethod
    def _create_empty_markdown() -> bytes:
        """创建空Markdown文件"""
        return b''