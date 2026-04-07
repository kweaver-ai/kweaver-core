# -*- coding:UTF-8 -*-

import pytest
import allure
import os
import subprocess

# 添加项目根目录到路径
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))


@pytest.fixture(scope="session", autouse=True)
def ModifyCM():
    '''
    修改cm配置，变更算子配置默认值：
        批量注册算子修改为10个
        最大文件大小修改为50k
        描述字符长度限制修改为255
    '''
    os.system("kubectl -n anyshare get cm agent-operator-integration -o yaml > cm.yaml")
    os.system("sed -i 's/import_file_size_limit: [0-9.e+-]*/import_file_size_limit: 51200/g; \
              s/import_operator_max_count: [0-9]*/import_operator_max_count: 10/g; \
              s/operator_description_length_limit: [0-9]*/operator_description_length_limit: 255/g' cm.yaml")
    os.system("kubectl replace -f cm.yaml")
    os.system("kubectl -n anyshare delete pod $(kubectl get pod -n anyshare | grep agent-operator-integration | awk '{print $1}')")
    command = "kubectl get pod -n anyshare | grep agent-operator-integration | awk '{print $2}'"
    for i in range(60):
        result = subprocess.run(command, shell=True, capture_output=True, text=True)
        print("标准输出：", result.stdout)
        print("标准错误：", result.stderr)
        print("返回码：", result.returncode)
        if "1/1" in result.stdout:
            break
        else:
            os.system("sleep 5")