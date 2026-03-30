"""项目路径与 config.ini 读取（与 DP_AT 对齐）；ConfigParser 关闭插值以免 log 格式中的 % 被误解析。"""
import configparser
import os

# 项目根目录
BASE_DIR = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# 配置文件目录
CONFIG_DIR = os.path.join(BASE_DIR, "config")
CONFIG_FILE = os.path.join(CONFIG_DIR, "config.ini")

# 测试数据目录
DATA_DIR = os.path.join(BASE_DIR, "test_data")

# 日志目录
LOG_DIR = os.path.join(BASE_DIR, "logs")
os.makedirs(LOG_DIR, exist_ok=True)

# 测试报告目录
REPORT_DIR = os.path.join(BASE_DIR, "reports")
os.makedirs(REPORT_DIR, exist_ok=True)

# 资源目录（rsa_public.key 等）
RESOURCE_DIR = os.path.join(BASE_DIR, "resource")
os.makedirs(RESOURCE_DIR, exist_ok=True)

# Excel 用例数据读取配置
SHEET_TITLE_MAPPING = {
    "global": {
        "sheet": "全局变量",
        "mapping": {
            "name": "变量名",
            "value": "变量值"
        },
        "index": "name"
    },
    "api": {
        "sheet": "接口信息",
        "mapping": {
            "name": "name",
            "url": "url",
            "method": "method",
            "headers": "headers",
            "response_2xx": "response_2xx",
            "response_4xx": "response_4xx"
        },
        "index": "name"
    },
    "suite": {
        "sheet": "执行套件",
        "mapping": {
            "feature": "测试功能",
            "story": "测试场景",
            "switch": "测试开关"
        },
        "index": "name"
    },
    "case": {
        "mapping": {
            "name": "测试用例名称",
            "url": "调用接口",
            "prev_case": "前置用例",
            "query_params": "query参数",
            "body_params": "body参数",
            "form_params": "form参数",
            "path_update": "path参数更替",
            "body_update": "body参数更替",
            "resp_values": "响应提取",
            "code_check": "状态码断言",
            "resp_check": "响应断言"
        },
        "index": "name"
    }
}


class Config:
    def __init__(self):
        self.config = configparser.ConfigParser(interpolation=None)
        self.config.read(CONFIG_FILE, encoding="utf-8")

    def get(self, section, option):
        return self.config.get(section, option)

    def getboolean(self, section, option):
        return self.config.getboolean(section, option)

    def getint(self, section, option):
        return self.config.getint(section, option)

    def getfloat(self, section, option):
        return self.config.getfloat(section, option)

    def set(self, section, option, value):
        if section not in self.config:
            self.config[section] = {}
        self.config.set(section, option, str(value))

    def save(self):
        with open(CONFIG_FILE, "w", encoding="utf-8") as config_file:
            self.config.write(config_file)


config = Config()

if __name__ == "__main__":
    print(REPORT_DIR)
