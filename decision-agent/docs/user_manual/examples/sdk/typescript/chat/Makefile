SHELL := /bin/bash  # 指定使用的 shell

.PHONY: non-stream stream all  # 声明伪目标

non-stream:  # 非流式聊天
	npm run chat:non-stream  # 运行非流式聊天脚本

stream:  # 流式聊天
	npm run chat:stream  # 运行流式聊天脚本

all: non-stream stream  # 执行所有聊天示例
