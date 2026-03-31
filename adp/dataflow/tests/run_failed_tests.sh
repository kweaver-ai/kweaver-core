#!/bin/bash
# 运行上次失败的测试用例

# 方式1：运行所有上次失败的测试（最简单）
# pytest --lf

# 方式2：运行指定文件的上次失败的测试
# pytest --lf testcases/data-operator-hub/api/operator/test_get_operator_list.py

# 方式3：运行指定目录的上次失败的测试
# pytest --lf testcases/data-operator-hub/api/operator/

# 方式4：运行失败的测试，并显示详细输出
# pytest --lf -v

# 方式5：运行失败的测试，失败后停止
# pytest --lf -x

# 方式6：运行失败的测试，显示简短错误信息
# pytest --lf --tb=short

# 默认使用方式1：运行所有上次失败的测试
pytest --lf -v --tb=short
