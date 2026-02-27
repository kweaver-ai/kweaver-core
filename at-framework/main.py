#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
@Time   : 2024/7/1 9:57
@Author : Leopold.yu
@File   : main.py
"""
import os

from common import constant

if __name__ == '__main__':
    pytest_cmd = "pytest -s -q --alluredir %s --clean-alluredir  --junit-xml=%s" % (
        constant.XML_REPORT_DIR, constant.JUNIT_REPORT_FILE)
    os.system(pytest_cmd)

    allure_cmd = "allure generate %s -o %s --clean" % (constant.XML_REPORT_DIR, constant.HTML_REPORT_DIR)
    os.system(allure_cmd)

    # open_cmd = "allure open %s" % constant.HTML_REPORT_DIR
    # os.system(open_cmd)

