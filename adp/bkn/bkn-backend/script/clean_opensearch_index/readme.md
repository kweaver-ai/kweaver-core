# OpenSearch索引清理脚本

## 项目简介

OpenSearch索引清理脚本用于自动清理OpenSearch中的孤立索引。该脚本会检查所有OpenSearch中的索引，删除不在失败或取消的job对应的task中的索引，帮助释放存储空间并保持索引的整洁。

## 功能特性

- ✅ **智能识别**：从失败或取消的job查找对应的task，提取有效索引列表
- ✅ **安全删除**：支持试运行模式（Dry Run），在删除前预览将要删除的索引
- ✅ **详细信息**：显示每个索引的文档数和存储大小
- ✅ **空间统计**：统计并显示回收的磁盘空间总量
- ✅ **灵活配置**：支持环境变量和命令行参数两种配置方式
- ✅ **详细日志**：提供完整的操作日志，便于追踪和审计
- ✅ **高效查询**：使用JOIN语句优化数据库查询，减少往返次数

## 环境要求

- Python 3.10+
- MySQL/MariaDB 数据库
- OpenSearch 集群

## 安装依赖

```bash
pip install pymysql opensearch-py
```

## 配置说明

### 环境变量配置

创建 `.env` 文件或设置系统环境变量：

```bash
# 数据库配置
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password

# OpenSearch配置
export OPENSEARCH_HOST=localhost
export OPENSEARCH_PORT=9200
export OPENSEARCH_PROTOCOL=http
export OPENSEARCH_USER=admin
export OPENSEARCH_PASSWORD=your_password
```

### 命令行参数配置

也可以直接通过命令行参数指定配置：

```bash
python3 script.py \
  --db-host localhost \
  --db-port 3306 \
  --db-user root \
  --db-password your_password \
  --os-host localhost \
  --os-port 9200 \
  --os-protocol http \
  --os-user admin \
  --os-password your_password
```

## 使用方法

### 1. 试运行模式（推荐先使用）

在正式删除前，使用试运行模式查看将要删除的索引：

```bash
python script.py --dry-run
```

### 2. 正式执行

确认无误后，执行实际的删除操作：

```bash
python script.py
```

### 3. 查看帮助

```bash
python script.py --help
```

## 输出示例

### 试运行模式输出

```
============================================================
开始OpenSearch索引清理
试运行模式: 是
============================================================
数据库连接成功
OpenSearch连接成功
从失败或取消的job的task中获取到 5 个有效索引
从OpenSearch获取到 10 个业务知识网络相关索引
发现 5 个孤立索引

孤立的索引列表:
------------------------------------------------------------
  - adp-kn_ot_index-20240101
    文档数: 1,234
    存储大小: 5.23 MB
  - dip-kn_ot_index-20240102
    文档数: 567
    存储大小: 2.15 MB
  - adp-kn_ot_index-20240103
    文档数: 890
    存储大小: 3.45 MB
  - dip-kn_ot_index-20240104
    文档数: 432
    存储大小: 1.87 MB
  - adp-kn_ot_index-20240105
    文档数: 1,024
    存储大小: 4.56 MB
------------------------------------------------------------

开始删除 5 个孤立索引...
[DRY RUN] 将删除索引: adp-kn_ot_index-20240101 (大小: 5.23 MB)
[DRY RUN] 将删除索引: dip-kn_ot_index-20240102 (大小: 2.15 MB)
[DRY RUN] 将删除索引: adp-kn_ot_index-20240103 (大小: 3.45 MB)
[DRY RUN] 将删除索引: dip-kn_ot_index-20240104 (大小: 1.87 MB)
[DRY RUN] 将删除索引: adp-kn_ot_index-20240105 (大小: 4.56 MB)

============================================================
清理完成
成功删除: 5 个
删除失败: 0 个
回收磁盘空间: 17.26 MB
============================================================
```

### 正式执行输出

```
============================================================
开始OpenSearch索引清理
试运行模式: 否
============================================================
数据库连接成功
OpenSearch连接成功
从失败或取消的job的task中获取到 5 个有效索引
从OpenSearch获取到 10 个业务知识网络相关索引
发现 5 个孤立索引

孤立的索引列表:
------------------------------------------------------------
  - adp-kn_ot_index-20240101
    文档数: 1,234
    存储大小: 5.23 MB
  - dip-kn_ot_index-20240102
    文档数: 567
    存储大小: 2.15 MB
  - adp-kn_ot_index-20240103
    文档数: 890
    存储大小: 3.45 MB
  - dip-kn_ot_index-20240104
    文档数: 432
    存储大小: 1.87 MB
  - adp-kn_ot_index-20240105
    文档数: 1,024
    存储大小: 4.56 MB
------------------------------------------------------------

开始删除 5 个孤立索引...
成功删除索引: adp-kn_ot_index-20240101 (回收: 5.23 MB)
成功删除索引: dip-kn_ot_index-20240102 (回收: 2.15 MB)
成功删除索引: adp-kn_ot_index-20240103 (回收: 3.45 MB)
成功删除索引: dip-kn_ot_index-20240104 (回收: 1.87 MB)
成功删除索引: adp-kn_ot_index-20240105 (回收: 4.56 MB)

============================================================
清理完成
成功删除: 5 个
删除失败: 0 个
回收磁盘空间: 17.26 MB
============================================================
```

## 工作原理

1. **连接数据库和OpenSearch**：建立与MySQL数据库和OpenSearch集群的连接
2. **获取有效索引**：从数据库中查询状态为'failed'或'canceled'的job，通过JOIN语句获取这些job对应的task中的索引名称
3. **获取OpenSearch索引**：使用cat API获取所有匹配模式的索引（`adp-kn_ot_index-*` 和 `dip-kn_ot_index-*`）及其大小信息
4. **识别孤立索引**：对比有效索引列表和OpenSearch中的索引，找出孤立索引
5. **删除索引**：在非试运行模式下，删除所有孤立索引
6. **统计结果**：输出删除成功/失败的数量和回收的磁盘空间

## 注意事项

⚠️ **重要提示**：
- 在正式执行前，务必先使用 `--dry-run` 参数进行试运行
- 确保数据库和OpenSearch的连接配置正确
- 脚本只会删除匹配模式 `adp-kn_ot_index-*` 和 `dip-kn_ot_index-*` 的索引
- 建议在业务低峰期执行清理操作
- 执行前建议备份重要的索引数据

## 故障排查

### 数据库连接失败
```
数据库连接失败: [Errno 111] Connection refused
```
**解决方案**：检查数据库地址、端口、用户名和密码是否正确

### OpenSearch连接失败
```
OpenSearch连接失败: ConnectionError
```
**解决方案**：检查OpenSearch服务是否正常运行，确认地址、端口和认证信息

### 没有找到孤立索引
```
没有发现孤立索引，清理完成
```
**说明**：这是正常情况，表示所有OpenSearch中的索引都在有效索引列表中

## 许可证

本脚本为内部工具，请遵循公司相关规定使用。

## 联系方式

如有问题或建议，请联系开发团队。
