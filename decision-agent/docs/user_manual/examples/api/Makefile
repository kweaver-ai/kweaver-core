SHELL := /bin/bash

.PHONY: health list agent-config auxiliary chat-examples conversations conversation-session import-export quick-check smoke flow clean

health:
	./scripts/health.sh

list:
	./scripts/list-agents.sh

agent-config:
	./scripts/agent-config.sh

auxiliary:
	./scripts/auxiliary.sh all

chat-examples:
	./scripts/chat.sh all

conversations:
	./scripts/conversations.sh

conversation-session:
	./scripts/conversation-session.sh

import-export:
	./scripts/import-export.sh

quick-check: health list auxiliary

smoke: quick-check

flow:
	./scripts/agent-flow.sh

clean:
	rm -rf .tmp
