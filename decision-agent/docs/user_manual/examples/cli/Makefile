SHELL := /bin/bash

KWEAVER_SDK_TS_DIR ?= ../../../../../../kweaver-sdk/packages/typescript
SDK_PACKAGE ?= $(KWEAVER_SDK_TS_DIR)
export PATH := $(CURDIR)/node_modules/.bin:$(PATH)

.PHONY: install help list personal-list category-list template-list detail-examples chat-examples skill-examples update-knowledge-network quick-check smoke flow clean

install:
	npm install --no-save "$(SDK_PACKAGE)"

help: install
	kweaver --help >/dev/null
	kweaver agent --help >/dev/null
	@echo "kweaver CLI help OK"

list: install
	./scripts/list-agents.sh list

personal-list: install
	./scripts/list-agents.sh personal-list

category-list: install
	./scripts/list-agents.sh category-list

template-list: install
	./scripts/list-agents.sh template-list

detail-examples: install
	./scripts/agent-optional.sh details

chat-examples: install
	./scripts/agent-optional.sh chat

skill-examples: install
	./scripts/agent-optional.sh skill

update-knowledge-network: install
	./scripts/agent-optional.sh update-knowledge-network

quick-check: help list personal-list category-list template-list

smoke: quick-check

flow: install
	./scripts/agent-flow.sh

clean:
	rm -rf node_modules package-lock.json .tmp
