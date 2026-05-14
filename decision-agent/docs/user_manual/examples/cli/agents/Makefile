SHELL := /bin/bash

.PHONY: list personal-list category-list template-list get get-by-key template-get create update update-knowledge-network publish unpublish delete flow

# 智能体生命周期示例：每个目标对应一个 kweaver agent 子命令。
list:
	./list.sh

personal-list:
	./personal-list.sh

category-list:
	./category-list.sh

template-list:
	./template-list.sh

get:
	./get.sh

get-by-key:
	./get-by-key.sh

template-get:
	./template-get.sh

create:
	./create.sh

update:
	./update.sh

update-knowledge-network:
	./update-knowledge-network.sh

publish:
	./publish.sh

unpublish:
	./unpublish.sh

delete:
	./delete.sh

# 完整流程会创建、发布、取消发布并删除临时 Agent。
flow:
	./flow.sh
