SHELL := /bin/bash  # 指定使用的 shell

.PHONY: import-smoke setup all  # 声明伪目标

import-smoke:  # 导入烟雾测试
	npm run client:import-smoke  # 运行导入烟雾测试脚本

setup:  # 客户端设置
	npm run client:setup  # 运行客户端设置脚本

all: import-smoke setup  # 执行所有客户端示例
