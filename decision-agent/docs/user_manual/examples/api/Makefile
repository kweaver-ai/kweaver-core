SHELL := /bin/bash  # 指定使用的 shell

.PHONY: health list agent-config agent-config-minimal agent-config-advanced auxiliary chat-examples conversations conversation-session import-export quick-check smoke flow clean  # 声明伪目标
.PHONY: chat-non-stream chat-stream chat-incremental-stream chat-debug chat-terminate chat-resume  # 声明伪目标

# 基础健康检查，确认本地 Decision Agent 服务可访问。
health:
	$(MAKE) -C health ready  # 执行健康检查

list:  # 列出已发布 Agent
	$(MAKE) -C agents list-published  # 执行已发布 Agent 列表查询

agent-config: agent-config-minimal agent-config-advanced  # Agent 配置

agent-config-minimal:  # 最小 Agent 配置
	$(MAKE) -C agent-config minimal  # 执行最小配置生成

agent-config-advanced:  # 高级 Agent 配置
	$(MAKE) -C agent-config advanced  # 执行高级配置生成

auxiliary:  # 辅助接口
	$(MAKE) -C auxiliary all  # 执行所有辅助接口示例

chat-non-stream:  # 非流式聊天
	$(MAKE) -C chat non-stream  # 执行非流式聊天

chat-stream:  # 流式聊天
	$(MAKE) -C chat stream  # 执行流式聊天

chat-incremental-stream:  # 增量流式聊天
	$(MAKE) -C chat incremental-stream  # 执行增量流式聊天

chat-debug:  # Debug 聊天
	$(MAKE) -C chat debug  # 执行 Debug 聊天

chat-terminate:  # 终止聊天
	$(MAKE) -C chat terminate  # 执行终止聊天

chat-resume:  # 恢复聊天
	$(MAKE) -C chat resume  # 执行恢复聊天

chat-examples:  # 聊天示例
	$(MAKE) -C chat all  # 执行所有聊天示例

conversations:  # 对话
	$(MAKE) -C conversations all  # 执行所有对话示例

conversation-session:  # 对话会话
	$(MAKE) -C session manage  # 执行会话管理

import-export:  # 导入导出
	$(MAKE) -C import-export roundtrip  # 执行导入导出

# 低风险快速检查：不创建或删除 Agent。
quick-check: health list auxiliary  # 快速检查

# 历史兼容别名，语义等同于 quick-check。
smoke: quick-check  # 烟雾测试

# 完整流程会创建、发布、取消发布并删除临时 Agent。
flow:
	$(MAKE) -C agents flow  # 执行 Agent 完整流程

# 清理 API 示例生成的临时文件。
clean:
	rm -rf .tmp  # 删除临时文件
