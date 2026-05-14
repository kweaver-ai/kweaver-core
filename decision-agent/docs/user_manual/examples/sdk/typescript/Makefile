# 指定使用的 shell。
SHELL := /bin/bash

# 包含环境变量配置文件。
-include ../../.env

# 计算 SDK 包来源；strip 用来消除 .env 行内注释前的尾部空格。
EXAMPLES_ROOT := $(abspath ../..)
KWEAVER_SDK_PACKAGE_SOURCE ?= remote
KWEAVER_SDK_PACKAGE_SOURCE := $(strip $(KWEAVER_SDK_PACKAGE_SOURCE))
KWEAVER_SDK_REMOTE_PACKAGE ?= @kweaver-ai/kweaver-sdk
KWEAVER_SDK_REMOTE_PACKAGE := $(strip $(KWEAVER_SDK_REMOTE_PACKAGE))
KWEAVER_SDK_TS_DIR ?= ../../../../../kweaver-sdk/packages/typescript
KWEAVER_SDK_TS_DIR := $(strip $(KWEAVER_SDK_TS_DIR))
ifneq ($(filter /%,$(KWEAVER_SDK_TS_DIR)),$(KWEAVER_SDK_TS_DIR))
# 如果 SDK 目录不是绝对路径，则转换为绝对路径。
KWEAVER_SDK_TS_DIR := $(EXAMPLES_ROOT)/$(KWEAVER_SDK_TS_DIR)
endif
ifeq ($(KWEAVER_SDK_PACKAGE_SOURCE),remote)
# 使用远程 npm 包安装。
SDK_INSTALL_SPEC := $(KWEAVER_SDK_REMOTE_PACKAGE)
else ifeq ($(KWEAVER_SDK_PACKAGE_SOURCE),local)
# 使用本地 SDK 项目安装。
SDK_INSTALL_SPEC := $(KWEAVER_SDK_TS_DIR)
else
$(error KWEAVER_SDK_PACKAGE_SOURCE must be remote or local, got "$(KWEAVER_SDK_PACKAGE_SOURCE)")
endif

.PHONY: install import-smoke client-setup list agent-detail create publish unpublish delete chat chat-stream conversations conversation-messages quick-check smoke flow clean  # 声明伪目标

# 安装示例依赖和用户视角的 TypeScript SDK 包。
install:
	npm install  # 安装 npm 依赖
	npm uninstall --no-save "$(KWEAVER_SDK_REMOTE_PACKAGE)" >/dev/null 2>&1 || true  # 卸载旧 SDK 包，方便远程和本地切换
	rm -rf node_modules/@kweaver-ai/kweaver-sdk  # 清理旧 SDK 包目录
	npm install --no-save "$(SDK_INSTALL_SPEC)"  # 安装选定的 SDK 包到当前示例目录

import-smoke: install  # 导入烟雾测试依赖
	$(MAKE) -C client import-smoke  # 执行客户端导入烟雾测试

list: install  # 列出 Agent
	$(MAKE) -C agents list  # 执行 Agent 列表查询

client-setup: install  # 客户端设置
	$(MAKE) -C client setup  # 执行客户端设置

agent-detail: install  # Agent 详情
	$(MAKE) -C agents detail  # 执行 Agent 详情查询

create: install  # 创建 Agent
	$(MAKE) -C agents create  # 执行 Agent 创建

publish: install  # 发布 Agent
	$(MAKE) -C agents publish  # 执行 Agent 发布

unpublish: install  # 取消发布 Agent
	$(MAKE) -C agents unpublish  # 执行 Agent 取消发布

delete: install  # 删除 Agent
	$(MAKE) -C agents delete  # 执行 Agent 删除

chat: install  # 聊天（非流式）
	$(MAKE) -C chat non-stream  # 执行非流式聊天

chat-stream: install  # 聊天（流式）
	$(MAKE) -C chat stream  # 执行流式聊天

conversations: install  # 对话列表
	$(MAKE) -C conversations list  # 执行对话列表查询

conversation-messages: install  # 对话消息
	$(MAKE) -C conversations messages  # 执行对话消息查询

# 低风险快速检查：验证导入、初始化和列表接口。
quick-check: import-smoke client-setup list  # 快速检查

# 历史兼容别名，语义等同于 quick-check。
smoke: quick-check  # 烟雾测试

# 完整流程会创建、发布、取消发布并删除临时 Agent。
flow: install  # 完整流程
	$(MAKE) -C agents flow  # 执行 Agent 完整流程

# 清理 SDK 示例本地依赖和临时状态。
clean:
	rm -rf node_modules package-lock.json .tmp  # 删除依赖和临时文件
