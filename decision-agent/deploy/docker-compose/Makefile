# Decision Agent Docker Compose Makefile

.PHONY: help up down build logs ps validate clean clean-all restart shell

# 默认目标
help:
	@echo "Decision Agent Docker Compose 命令"
	@echo ""
	@echo "可用命令："
	@echo "  make up      - 启动所有服务"
	@echo "  make down    - 停止所有服务"
	@echo "  make build   - 构建所有镜像"
	@echo "  make logs    - 查看所有服务日志"
	@echo "  make ps      - 查看服务状态"
	@echo "  make validate- 验证配置文件"
	@echo "  make clean   - 停止服务并删除容器"
	@echo "  make clean-all - 停止服务并删除所有数据"
	@echo "  make restart - 重启所有服务"
	@echo "  make shell   - 进入 agent-backend 容器"
	@echo ""
	@echo "示例："
	@echo "  make up              # 启动服务"
	@echo "  make logs backend    # 查看后端日志"
	@echo "  make validate        # 验证配置"

# 启动所有服务
up:
	@echo "启动 Decision Agent 服务..."
	docker-compose up -d
	@echo "服务已启动，使用 'make ps' 查看状态"

# 停止所有服务
down:
	@echo "停止 Decision Agent 服务..."
	docker-compose down

# 构建所有镜像
build:
	@echo "构建 Decision Agent 镜像..."
	docker-compose build --no-cache

# 查看日志
logs:
	@if [ -n "$(service)" ]; then \
		echo "查看 $(service) 服务日志..."; \
		docker-compose logs -f $(service); \
	else \
		echo "查看所有服务日志..."; \
		docker-compose logs -f; \
	fi

# 查看服务状态
ps:
	@echo "Decision Agent 服务状态："
	docker-compose ps

# 验证配置
validate:
	@echo "验证配置文件..."
	./validate-config.sh

# 清理容器和网络
clean:
	@echo "停止服务并删除容器..."
	docker-compose down -v --remove-orphans
	docker system prune -f

# 完全清理（包括数据卷）
clean-all:
	@echo "警告：这将删除所有数据，包括数据库数据！"
	@read -p "确认删除？[y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	docker-compose down -v --remove-orphans
	docker volume prune -f
	docker system prune -af

# 重启服务
restart: down up

# 进入容器
shell:
	@if [ -n "$(service)" ]; then \
		echo "进入 $(service) 容器..."; \
		docker-compose exec $(service) /bin/bash; \
	else \
		echo "进入 agent-backend 容器..."; \
		docker-compose exec agent-backend /bin/bash; \
	fi

# 快速部署（首次使用）
deploy: validate up
	@echo ""
	@echo "部署完成！"
	@echo "前端地址: http://localhost$$(grep AGENT_WEB_HOST_PORT .env 2>/dev/null | cut -d'=' -f2 || echo ":1101")/agent-web/my-agents.html"
	@echo "使用 'make logs' 查看日志"
	@echo "使用 'make ps' 查看服务状态"

# 导出数据库
export-db:
	@echo "导出数据库..."
	docker-compose exec mariadb mysqldump -u root -p$$(grep MYSQL_ROOT_PASSWORD .env | cut -d'=' -f2) adp > backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "数据库已导出"

# 导入数据库
import-db:
	@if [ -z "$(file)" ]; then \
		echo "用法: make import-db file=backup.sql"; \
		exit 1; \
	fi
	@echo "导入数据库 $(file)..."
	docker-compose exec -T mariadb mysql -u root -p$$(grep MYSQL_ROOT_PASSWORD .env | cut -d'=' -f2) adp < $(file)
	@echo "数据库已导入"
