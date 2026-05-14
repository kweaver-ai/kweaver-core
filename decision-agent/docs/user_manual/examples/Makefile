SHELL := /bin/bash

-include .env

.PHONY: quick-check smoke api-quick-check cli-quick-check sdk-quick-check api-smoke cli-smoke sdk-smoke flow api-flow cli-flow sdk-flow clean

# 低风险快速检查：只验证服务、安装入口和列表类接口。
quick-check: api-quick-check cli-quick-check sdk-quick-check

# 历史兼容别名，语义等同于 quick-check。
smoke: quick-check

api-quick-check:
	$(MAKE) -C api quick-check

cli-quick-check:
	$(MAKE) -C cli quick-check

sdk-quick-check:
	$(MAKE) -C sdk/typescript quick-check

api-smoke: api-quick-check

cli-smoke: cli-quick-check

sdk-smoke: sdk-quick-check

# 完整流程会创建、发布、取消发布并删除临时 Agent，请显式调用。
flow: api-flow cli-flow sdk-flow

api-flow:
	$(MAKE) -C api flow

cli-flow:
	$(MAKE) -C cli flow

sdk-flow:
	$(MAKE) -C sdk/typescript flow

# 清理示例运行产生的临时状态和本地依赖。
clean:
	rm -rf .tmp
	$(MAKE) -C api clean
	$(MAKE) -C cli clean
	$(MAKE) -C sdk/typescript clean
