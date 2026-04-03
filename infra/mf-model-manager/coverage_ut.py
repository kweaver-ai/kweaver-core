import unittest
from os import path

import coverage
import xmlrunner

# 实例化一个对象
cov = coverage.coverage()
cov.start()

# 测试套件
test_dir = path.join(path.dirname(path.abspath(__file__)), 'app/test')

suite = unittest.defaultTestLoader.discover(test_dir, "test_*.py")
# unittest.TextTestRunner().run(suite)
runner = xmlrunner.XMLTestRunner(output="coverage_result")  # 测试结果报告
runner.run(suite)

# 结束分析
cov.stop()

# 结果保存
cov.save()

# 命令行模式展示结果
cov.report()

# 生成HTML覆盖率报告
# cov.html_report(directory='result_html')
cov.xml_report(outfile="coverage.xml")  # 代码覆盖率报告
