SHELL := /bin/bash

KWEAVER_SDK_TS_DIR ?= ../../../../../../../kweaver-sdk/packages/typescript
SDK_PACKAGE ?= $(KWEAVER_SDK_TS_DIR)

.PHONY: install import-smoke client-setup list agent-detail chat chat-stream conversations quick-check smoke flow clean

install:
	npm install
	npm install --no-save "$(SDK_PACKAGE)"

import-smoke: install
	npm run import-smoke

list: install
	npm run list

client-setup: install
	npm run client-setup

agent-detail: install
	npm run agent-detail

chat: install
	npm run chat

chat-stream: install
	npm run chat-stream

conversations: install
	npm run conversations

quick-check: import-smoke client-setup list

smoke: quick-check

flow: install
	npm run flow

clean:
	rm -rf node_modules package-lock.json .tmp
