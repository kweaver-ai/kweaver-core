# -*- coding:UTF-8 -*-
import os

from common.get_content import GetContent
from common.request import Request

class Impex():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"] + "/api/agent-operator-integration/v1/impex"

    '''导出'''
    def export(self, component_type, component_id, headers):
        url = f"{self.base_url}/export/{component_type}/{component_id}"
        return Request.get(self, url, headers)


    '''从文件路径导入（自动处理文件打开和关闭）'''
    def import_from_file(self, type, file_path, data, headers):
        """
        统一导入入口：完全模拟 WebKit 报文顺序和构造。
        1. data 部分在前，包含 filename 和 Content-Type。
        2. mode 部分在后，不包含 Content-Type。
        """
        form_data = data.copy() if data else {}
        mode = form_data.pop("mode", "create")
        
        with open(file_path, "rb") as f:
            # 使用元组列表确保顺序：data 在前，mode 在后
            files = [
                ("data", (os.path.basename(file_path), f, "application/octet-stream")),
                ("mode", (None, mode))
            ]
            return self.importation(type, files, {}, headers)

    '''导入底层调用'''
    def importation(self, type, files, data, headers, params=None):
        url = f"{self.base_url}/import/{type}"
        return Request.post_multipart(self, url, files, data, headers, params=params)