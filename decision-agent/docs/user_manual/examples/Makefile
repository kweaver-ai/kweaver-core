SHELL := /bin/bash

.PHONY: quick-check smoke api-quick-check cli-quick-check sdk-quick-check api-smoke cli-smoke sdk-smoke flow api-flow cli-flow sdk-flow clean

quick-check: api-quick-check cli-quick-check sdk-quick-check

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

flow: api-flow cli-flow sdk-flow

api-flow:
	$(MAKE) -C api flow

cli-flow:
	$(MAKE) -C cli flow

sdk-flow:
	$(MAKE) -C sdk/typescript flow

clean:
	$(MAKE) -C api clean
	$(MAKE) -C cli clean
	$(MAKE) -C sdk/typescript clean
