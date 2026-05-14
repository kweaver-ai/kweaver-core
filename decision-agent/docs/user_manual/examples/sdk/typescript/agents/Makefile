SHELL := /bin/bash  # 指定使用的 shell

.PHONY: list detail create publish unpublish delete flow  # 声明伪目标

list:  # 列出 Agent
	npm run agents:list  # 运行 Agent 列表脚本

detail:  # Agent 详情
	npm run agents:detail  # 运行 Agent 详情脚本

create:  # 创建 Agent
	npm run agents:create  # 运行 Agent 创建脚本

publish:  # 发布 Agent
	npm run agents:publish  # 运行 Agent 发布脚本

unpublish:  # 取消发布 Agent
	npm run agents:unpublish  # 运行 Agent 取消发布脚本

delete:  # 删除 Agent
	npm run agents:delete  # 运行 Agent 删除脚本

flow:  # 完整流程
	npm run agents:flow  # 运行 Agent 完整流程脚本
