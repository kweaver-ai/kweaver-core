#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
@Time   : 2024/7/1 9:57
@Author : Leopold.yu
@File   : main.py
"""
import argparse
import os

from common import constant


def _build_pytest_cmd(args):
    cmd = "pytest -s -q --alluredir %s --clean-alluredir --junit-xml=%s" % (
        constant.XML_REPORT_DIR,
        constant.JUNIT_REPORT_FILE,
    )
    cmd += " test_run.py"
    if args.case_file:
        cmd += ' --case-file "%s"' % args.case_file
    if args.scope:
        cmd += ' --scope "%s"' % args.scope
    if args.tags:
        cmd += ' --tags "%s"' % args.tags
    if args.api_name:
        cmd += ' --api-name "%s"' % args.api_name
    if args.api_path:
        cmd += ' --api-path "%s"' % args.api_path
    if args.case_names:
        cmd += ' --case-names "%s"' % args.case_names
    if args.suite:
        cmd += ' --suite "%s"' % args.suite
    return cmd


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description="Run AT framework test cases with flexible filters")
    parser.add_argument("--case-file", default="", help="模块目录，如 ./testcase/vega")
    parser.add_argument("--scope", default="", help="scope id")
    parser.add_argument("--tags", default="", help="逗号分隔 tags")
    parser.add_argument("--api-name", default="", help="API 名称（apis.yaml 中 name）")
    parser.add_argument("--api-path", default="", help="API 路径")
    parser.add_argument("--case-names", default="", help="逗号分隔 case 名称列表")
    parser.add_argument("--suite", default="", help="套件 story 或文件名（不含 .yaml）")
    args = parser.parse_args()

    pytest_cmd = _build_pytest_cmd(args)
    os.system(pytest_cmd)

    allure_cmd = "allure generate %s -o %s --clean" % (constant.XML_REPORT_DIR, constant.HTML_REPORT_DIR)
    os.system(allure_cmd)

    # open_cmd = "allure open %s" % constant.HTML_REPORT_DIR
    # os.system(open_cmd)

