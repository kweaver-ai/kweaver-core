# 日志管理工具

本目录包含用于管理项目日志的makefile脚本，提供各种日志清理、备份和管理功能。

## 文件说明

- `logs.mk` - 主日志管理脚本（管理 log/ 目录下的日志文件）
- `streaming-logs.mk` - 流式响应日志管理脚本（管理 log/streaming_responses/ 目录下的日志文件）

## 使用方法

### 主日志管理 (logs.mk)
```bash
# 从项目根目录使用
make -f scripts/logging-tools/logs.mk help
make -f scripts/logging-tools/logs.mk clean
make -f scripts/logging-tools/logs.mk backup
make -f scripts/logging-tools/logs.mk rotate

# 从scripts目录使用
make -f logging-tools/logs.mk help
```

### 流式日志管理 (streaming-logs.mk)
```bash
# 从项目根目录使用
make -f scripts/logging-tools/streaming-logs.mk help
make -f scripts/logging-tools/streaming-logs.mk clean
make -f scripts/logging-tools/streaming-logs.mk cleanup-by-days N=7

# 从scripts目录使用
make -f logging-tools/streaming-logs.mk help
```

## 功能特性

### logs.mk 功能
- 完整清理（清理旧日志 + 备份并清空当前日志）
- 日志轮转（备份当前日志并清空）
- 清理包含日期的旧日志文件
- 清理所有日志文件
- 备份当前日志文件到当天日期
- 列出当前日志文件

### streaming-logs.mk 功能
- 按天数清理日志文件
- 清理7天前的日志文件
- 清理所有流式日志文件
- 列出当前流式日志文件

## 注意事项

- 脚本会自动检测当前目录，正确设置日志路径
- 支持从不同目录执行命令
- 建议定期执行清理操作以避免日志文件过大

### 主日志管理

```bash
# 在项目根目录下执行
make -f scripts/logs.mk help          # 查看帮助
make -f scripts/logs.mk list          # 列出日志文件
make -f scripts/logs.mk full-cleanup  # 完整清理（推荐）
make -f scripts/logs.mk clean         # 清理旧日志
make -f scripts/logs.mk rotate        # 备份并清空当前日志
make -f scripts/logs.mk backup        # 仅备份当前日志
make -f scripts/logs.mk clean-all     # 清理所有日志
```

### 流式响应日志管理

```bash
# 在项目根目录下执行
make -f scripts/streaming-logs.mk help                    # 查看帮助
make -f scripts/streaming-logs.mk list                    # 列出日志统计
make -f scripts/streaming-logs.mk clean                   # 清理7天前的日志
make -f scripts/streaming-logs.mk cleanup-by-days N=3     # 清理3天前的日志
make -f scripts/streaming-logs.mk clean-all               # 清理所有日志
```

## 快捷命令

推荐创建以下快捷命令：

```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
alias logs-clean='make -f scripts/logs.mk full-cleanup'
alias streaming-clean='make -f scripts/streaming-logs.mk clean'
alias logs-list='make -f scripts/logs.mk list && echo "---" && make -f scripts/streaming-logs.mk list'
```

## 注意事项

1. 执行清理操作前建议先使用 `list` 命令查看当前日志状态
2. `full-cleanup` 会备份当前日志后再清空，相对安全
3. 流式响应日志文件名格式：`YYYYMMDD_HHMMSS_UUID.log`
4. 支持多种执行方式：
   - 在项目根目录：`make -f scripts/logs.mk list`
   - 在scripts目录：`make -f logs.mk list`
   - 使用绝对路径：`make -f /path/to/scripts/logs.mk -C /path/to/scripts list`
5. 自动检测执行目录并调整路径，无需手动修改
