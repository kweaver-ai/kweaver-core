# -*- coding:UTF-8 -*-
import configparser
import json
import yaml

class GetContent():
    def __init__(self, filename):
        self.filename = filename

    '''读取ini配置文件'''
    def config(self):
        config = configparser.ConfigParser()
        config.read(self.filename)
        return config

        # config = configparser.ConfigParser()
        # config.read(self.filename)
        # data = {}
        # for section in config.sections():
        #     data[section] = {}
        #     for key, value in config.items(section):
        #         data[section][key] = value
        # return data

    '''读取json文件'''
    def jsonfile(self):
        with open(self.filename, 'r', encoding="utf-8") as f:
            data = json.load(f)
            # raw_content = f.read().strip()
            # parsed_data = json.loads(raw_content)
            # data = json.dumps(parsed_data, indent=4, sort_keys=True)
        return data

    '''读取yaml文件'''
    def yamlfile(self):
        with open(self.filename, 'r', encoding="utf-8") as f:
            data = yaml.safe_load(f)
        return data

if __name__ == "__main__":
    # 使用示例
    data = GetContent("./resource/openapi/non-compliant/title_not_exists.yaml").yamlfile()
    print(data)