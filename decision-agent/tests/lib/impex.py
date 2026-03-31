# -*- coding:UTF-8 -*-

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

    '''导入'''
    def importation(self, type, files, data, headers):
        url = f"{self.base_url}/import/{type}"
        return Request.post_multipart(self, url, files, data, headers)