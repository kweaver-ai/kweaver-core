# 创建概念索引删除脚本

## 概述

创建一个 Python 脚本用于删除 OpenSearch 中的概念索引 `adp-kn_concept`。脚本将放在 `script/` 目录下。

## 实现细节

### 文件结构

- `script/delete_concept_index.py` - 主脚本文件

### 功能特性

1. **OpenSearch 连接**

   - 参考 `script/clean_opensearch_index/script.py` 的连接方式
   - 支持 HTTP/HTTPS 协议
   - 支持用户名密码认证
   - 从环境变量或命令行参数读取配置

2. **索引删除**

   - 删除概念索引 `adp-kn_concept`（定义在 `server/interfaces/common.go` 中的 `KN_CONCEPT_INDEX_NAME`）
   - 检查索引是否存在后再删除
   - 显示索引信息（文档数、存储大小）

3. **安全特性**

   - 支持 `--dry-run` 模式，预览操作而不实际删除
   - 删除前显示索引信息并要求确认（可选）
   - 详细的日志输出

4. **配置方式**

   - 环境变量：`OPENSEARCH_HOST`, `OPENSEARCH_PORT`, `OPENSEARCH_PROTOCOL`, `OPENSEARCH_USER`, `OPENSEARCH_PASSWORD`
   - 命令行参数覆盖环境变量
   - 默认值：localhost:9200, http 协议

## 环境要求

- Python 3.10+
- OpenSearch 集群

## 安装依赖

```bash
pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple opensearch-py
```

### 使用示例

```bash
# 使用环境变量
export OPENSEARCH_HOST=localhost
export OPENSEARCH_PORT=9200
export OPENSEARCH_USER=test
export OPENSEARCH_PASSWORD=testpwd
python3 script/delete_concept_index.py

# 使用命令行参数
python3 script/delete_concept_index.py --os-host localhost --os-port 9200 --os-user test --os-password testpwd

# 试运行模式
python3 script/delete_concept_index.py --dry-run
```

## 参考文件

- `script/clean_opensearch_index/script.py` - OpenSearch 连接和索引删除的实现方式
- `server/interfaces/common.go` - 概念索引名称定义 (`KN_CONCEPT_INDEX_NAME = "adp-kn_concept"`)
- `server/common/setting.go` - OpenSearch 配置结构