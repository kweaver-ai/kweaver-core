#清理流式响应日志文件

# 检测当前目录，设置正确的LOG_PATH
LOG_PATH := log/streaming_responses
ifeq ($(shell basename $(CURDIR)),scripts)
    LOG_PATH := ../../log/streaming_responses
endif
ifeq ($(shell basename $(CURDIR)),logging-tools)
    LOG_PATH := ../../log/streaming_responses
endif

.PHONY: clean clean-all help list cleanup-by-days

# 默认目标
help:
	@echo "可用的make命令："
	@echo "  make cleanup-by-days N=7 - 清理N天前的日志文件（默认7天）"
	@echo "  make clean              - 清理7天前的日志文件"
	@echo "  make clean-all          - 清理所有日志文件"
	@echo "  make list               - 列出当前日志文件"
	@echo "  make help               - 显示帮助信息"

# 清理N天前的日志文件
cleanup-by-days:
	@echo "清理 $${N:-7} 天前的日志文件..."
	@if [ -z "$$N" ]; then \
		DAYS=7; \
	else \
		DAYS=$$N; \
	fi; \
	cd $(LOG_PATH) && find . -name "*.log" -type f -mtime +$$DAYS -delete 2>/dev/null || true; \
	echo "清理完成"

# 清理7天前的日志文件
clean: cleanup-by-days

# 清理所有日志文件
clean-all:
	@echo "清理所有流式响应日志文件..."
	@cd $(LOG_PATH) && rm -f *.log 2>/dev/null || true
	@echo "清理完成"

# 列出当前日志文件
list:
	@echo "当前流式响应日志文件："
	@cd $(LOG_PATH) && echo "文件总数: $$(ls -1 *.log 2>/dev/null | wc -l)"; \
	echo "总大小: $$(du -sh *.log 2>/dev/null | cut -f1 | tail -1 || echo '0B')"; \
	echo ""; \
	echo "按日期分组："; \
	ls -1 *.log 2>/dev/null | cut -d'_' -f1 | sort | uniq -c | sort -nr || echo "没有找到日志文件"; \
	echo ""; \
	echo "最近的10个文件："; \
	ls -lt *.log 2>/dev/null | head -11 || echo "没有找到日志文件"
