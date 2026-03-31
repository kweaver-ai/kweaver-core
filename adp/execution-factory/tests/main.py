# -*- coding: UTF-8 -*-

import os

def main():
    # --clean-alluredir 会清空历史执行记录
    os.system('python3 -m pytest ./testcases/api/data-operator-hub --alluredir ./report/xml --clean-alluredir')
    os.system('allure generate ./report/xml -o ./report/html --clean')
    # python3 -m pytest ./testcase/DeployInstaller --alluredir ./report/xml --clean-alluredir
    # allure generate ./report/xml -o ./report/html --clean
    # docker exec -it agent-at-new bash
    # export PYTHONPATH=$PYTHONPATH:/app/agent-AT


if __name__ == '__main__':
    main()