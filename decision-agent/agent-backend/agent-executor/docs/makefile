# Generate HTML from openapi.json
openapi-html:
	curl -X GET http://127.0.0.1:30778/openapi.json | python -m json.tool > openapi.json
	redocly build-docs openapi.json -o openapi.html

openapi-html-open:
	open openapi.html

# Clean generated files
clean:
	rm -f openapi.json openapi.html .openapi.json.prev .openapi.json.new openapi.json.tmp

openapi-html-watch:
	@echo "Starting watch mode (every 5 seconds)... Press Ctrl+C to stop"
	@touch .openapi.json.prev || true
	@while true; do \
		curl -sf http://127.0.0.1:30778/openapi.json > .openapi.json.new 2>/dev/null || { \
			echo "API not available, retrying..."; \
			sleep 5; \
			continue; \
		}; \
		if ! diff -q .openapi.json.prev .openapi.json.new >/dev/null 2>&1; then \
			echo "Changes detected, regenerating..."; \
			cp .openapi.json.new openapi.json; \
			python -m json.tool openapi.json > openapi.json.tmp && mv openapi.json.tmp openapi.json; \
			redocly build-docs openapi.json -o openapi.html; \
			cp .openapi.json.new .openapi.json.prev; \
		fi; \
		sleep 5; \
	done
	@rm -f .openapi.json.new .openapi.json.prev

.PHONY: openapi-html clean openapi-html-open openapi-html-watch
