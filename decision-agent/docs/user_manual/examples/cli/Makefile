# 指定使用的 shell。
SHELL := /bin/bash

# 包含环境变量配置文件。
-include ../.env

# 计算 SDK 包来源；strip 用来消除 .env 行内注释前的尾部空格。
EXAMPLES_ROOT := $(abspath ..)
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

# 优先使用当前示例目录安装出的 kweaver 命令。
export PATH := $(CURDIR)/node_modules/.bin:$(PATH)

.PHONY: install help list personal-list category-list template-list get get-by-key template-get create update update-knowledge-network publish unpublish delete detail-examples chat-examples skill-examples quick-check smoke flow clean  # 声明伪目标

# 安装用户视角的 kweaver CLI；根据 KWEAVER_SDK_PACKAGE_SOURCE 选择远程或本地包。
install:
	npm uninstall --no-save "$(KWEAVER_SDK_REMOTE_PACKAGE)" >/dev/null 2>&1 || true  # 卸载旧 SDK 包，方便远程和本地切换
	rm -rf node_modules/@kweaver-ai/kweaver-sdk  # 清理旧 SDK 包目录
	npm install --no-save "$(SDK_INSTALL_SPEC)"  # 安装选定的 SDK 包到当前示例目录

help: install  # 帮助信息
	$(MAKE) -C install help  # 执行帮助命令

list: install  # 列出 Agent
	$(MAKE) -C agents list  # 执行 Agent 列表查询

personal-list: install  # 列出个人 Agent
	$(MAKE) -C agents personal-list  # 执行个人 Agent 列表查询

category-list: install  # 列出分类
	$(MAKE) -C agents category-list  # 执行分类列表查询

template-list: install  # 列出模板
	$(MAKE) -C agents template-list  # 执行模板列表查询

get: install  # 获取 Agent
	$(MAKE) -C agents get  # 执行 Agent 获取

get-by-key: install  # 通过 key 获取 Agent
	$(MAKE) -C agents get-by-key  # 执行通过 key 获取 Agent

template-get: install  # 获取模板
	$(MAKE) -C agents template-get  # 执行模板获取

create: install  # 创建 Agent
	$(MAKE) -C agents create  # 执行 Agent 创建

update: install  # 更新 Agent
	$(MAKE) -C agents update  # 执行 Agent 更新

publish: install  # 发布 Agent
	$(MAKE) -C agents publish  # 执行 Agent 发布

unpublish: install  # 取消发布 Agent
	$(MAKE) -C agents unpublish  # 执行 Agent 取消发布

delete: install  # 删除 Agent
	$(MAKE) -C agents delete  # 执行 Agent 删除

detail-examples: get get-by-key template-get  # 详情示例

chat-examples: install  # 聊天示例
	$(MAKE) -C chat all  # 执行所有聊天示例

skill-examples: install  # 技能示例
	$(MAKE) -C skills all  # 执行所有技能示例

update-knowledge-network: install  # 更新知识网络
	$(MAKE) -C agents update-knowledge-network  # 执行知识网络更新

# 低风险快速检查：验证 CLI 入口和列表类命令，不创建数据。
quick-check: help list personal-list category-list template-list  # 快速检查

# 历史兼容别名，语义等同于 quick-check。
smoke: quick-check  # 烟雾测试

# 完整流程会创建、发布、取消发布并删除临时 Agent。
flow: install  # 完整流程
	$(MAKE) -C agents flow  # 执行 Agent 完整流程

# 清理 CLI 示例本地依赖和临时状态。
clean:
	rm -rf node_modules package-lock.json .tmp  # 删除依赖和临时文件
