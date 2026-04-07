# -*- coding:UTF-8 -*-
import json
import os
from pathlib import Path

class FileProcess():
    def write_json_to_file(self, data, filename, indent=4):
        """
        将JSON数据写入文件
        
        参数:
            data: 要写入的JSON数据（可以是字典、列表等可序列化对象）
            filename: 要保存的文件名
        """
        try:
             # 获取文件所在的目录路径
            dir_path = os.path.dirname(filename)
            
            # 如果目录不存在，则创建目录（包括所有父目录）
            if dir_path and not os.path.exists(dir_path):
                # 使用Path.mkdir()创建目录，parents=True表示创建所有父目录
                # exist_ok=True表示如果目录已存在也不报错
                Path(dir_path).mkdir(parents=True, exist_ok=True)
                print(f"已创建目录: {dir_path}")
            # 使用with语句打开文件，确保操作完成后正确关闭文件
            # indent参数用于格式化输出，使JSON内容更易读
            with open(filename, 'w', encoding='utf-8') as file:
                json.dump(data, file, ensure_ascii=False, indent=4)
            print(f"JSON数据已成功写入文件: {filename}")
        except Exception as e:
            print(f"写入文件时发生错误: {e}")

# 示例用法
if __name__ == "__main__":
    # 要写入的JSON数据（字典形式）
    sample_data = {
        "name": "张三",
        "age": 30,
        "is_student": False,
        "hobbies": ["阅读", "旅行", "编程"],
        "address": {
            "city": "北京",
            "district": "海淀区"
        }
    }
    
    # 调用函数将数据写入文件
    FileProcess().write_json_to_file(sample_data, "output.json")
