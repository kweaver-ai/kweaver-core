SHELL := /bin/bash  # 指定使用的 shell

.PHONY: list messages all  # 声明伪目标

list:  # 对话列表
	npm run conversations:list  # 运行对话列表脚本

messages:  # 对话消息
	npm run conversations:messages  # 运行对话消息脚本

all: list messages  # 执行所有对话示例
