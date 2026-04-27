# 清理异常日志文件 (log/exceptions 目录)

# 检测当前目录，设置正确的LOG_PATH
LOG_PATH := log/exceptions
ifeq ($(shell basename $(CURDIR)),scripts)
    LOG_PATH := ../log/exceptions
endif

ifeq ($(shell basename $(CURDIR)),logging-tools)
    LOG_PATH := ../../log/exceptions
endif

.PHONY: clean clean-all help list backup rotate full-cleanup

# 默认目标
help:
	@echo "可用的make命令（异常日志）："
	@echo "  make full-cleanup - 完整清理（清理旧日志 + 备份并清空当前日志）"
	@echo "  make rotate      - 执行日志轮转（备份当前日志并清空）"
	@echo "  make clean       - 清理包含日期的旧日志文件"
	@echo "  make clean-all   - 清理所有日志文件（包括当前日志）"
	@echo "  make backup      - 备份当前日志文件到当天日期"
	@echo "  make list        - 列出当前日志文件"
	@echo "  make help        - 显示帮助信息"

# 步骤1: 清理包含日期的日志文件
clean:
	@echo "清理包含日期的旧异常日志文件..."
	@find $(LOG_PATH) -name "*.log.*" -type f -delete 2>/dev/null || true
	@echo "清理完成"

# 完整清理：清理旧日志 + 备份并清空当前日志
full-cleanup: clean rotate

# 步骤2+3: 备份当前最新的日志文件到当天日期，然后清空
rotate:
	@echo "执行异常日志轮转..."
	@mkdir -p $(LOG_PATH)
	@cd $(LOG_PATH) && DATE=$$(date +%Y-%m-%d); \
	for log in exception_simple.log exception_detailed.log; do \
		if [ -f "$$log" ] && [ -s "$$log" ]; then \
			echo "备份 $$log 到 $$log.$$DATE"; \
			cp "$$log" "$$log.$$DATE"; \
			echo "清空 $$log"; \
			> "$$log"; \
		else \
			echo "文件 $$log 不存在或为空，跳过"; \
		fi; \
	done
	@echo "异常日志轮转完成"

# 仅备份当前日志文件（不清空）
backup:
	@echo "备份当前异常日志文件..."
	@mkdir -p $(LOG_PATH)
	@cd $(LOG_PATH) && DATE=$$(date +%Y-%m-%d); \
	for log in exception_simple.log exception_detailed.log; do \
		if [ -f "$$log" ] && [ -s "$$log" ]; then \
			echo "备份 $$log 到 $$log.$$DATE"; \
			cp "$$log" "$$log.$$DATE"; \
		else \
			echo "文件 $$log 不存在或为空，跳过"; \
		fi; \
	done
	@echo "备份完成"

# 清理所有日志文件
clean-all:
	@echo "清理所有异常日志文件..."
	@cd $(LOG_PATH) && rm -f *.log *.log.* 2>/dev/null || true
	@echo "清理完成"

# 列出当前日志文件
list:
	@echo "当前异常日志文件："
	@cd $(LOG_PATH) && ls -la *.log *.log.* 2>/dev/null || echo "没有找到异常日志文件"
