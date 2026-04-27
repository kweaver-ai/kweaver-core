# 02 · 从 CSV 文件到知识网络

> 散落的表格，连接起来。不需要写 SQL。

## 场景背景

HR 总监的员工、部门、项目数据散落在三张表格里。想搞清楚"谁向谁汇报"
或"哪些项目人手不足"，需要手动 VLOOKUP 串联多个文件，费时又容易出错。

这个示例把这些 CSV 文件导入知识网络。关系自动发现，你可以遍历组织架构、
语义搜索，并向 Agent 提问关于你的人员和项目的业务问题。

## 示例流程

```
本地 CSV 文件
     │
     ▼
┌─────────────────────┐     ┌──────────────┐
│  bkn create-from-csv │────▶│   知识网络    │
│  （导入 + 构建）      │     └──────┬───────┘
└─────────────────────┘            │
              ┌────────────────────┼───────────────────┐
              ▼                    ▼                   ▼
       ┌────────────┐     ┌──────────────┐    ┌───────────────┐
       │  Schema 探索 │     │   子图遍历   │    │  Agent 问答   │
       └────────────┘     └──────────────┘    └───────────────┘
```

0. **连接** MySQL 数据源（用于存放导入的表）
1. **一条命令**导入 CSV 并构建知识网络
2. **探索**自动发现的对象类型和属性
3. **查询**对象实例
4. **子图遍历**（depth=2，多跳关系查询）
5. **语义搜索**（通过 context-loader）
6. **导出** KN 定义文件
7. **Agent 对话**（注入 schema 上下文）

### 示例数据

| 文件 | 内容 |
|------|------|
| `departments.csv` | 5 个部门，含预算和人数 |
| `employees.csv` | 16 名员工，含职级、薪资、汇报关系 |
| `projects.csv` | 8 个项目，含状态、预算、负责人 |

## 前置条件

```bash
# 1. 安装 KWeaver CLI
npm install -g @kweaver-ai/kweaver-sdk

# 2. 登录 KWeaver 平台
kweaver auth login https://<platform-url>

# 3. 准备一个平台可访问的 MySQL 数据库
#    （脚本自动创建表，无需手动建表）
```

## 快速开始

```bash
cp env.sample .env
# 填写 DB_HOST、DB_NAME、DB_USER、DB_PASS（见 env.sample 中的注释）
vim .env
./run.sh
```

> **安全提示：** `.env` 已被 gitignore 排除。请勿将含有真实凭据的 `.env` 提交到版本控制。

### 使用自己的 CSV 文件

将 `data/` 目录中的文件替换为你自己的 CSV 即可：
- 第一行为列名（header）
- 文件名成为表名和对象类型名
- 所有列自动导入，数值列自动识别类型

## 关键命令

| 命令 | 作用 |
|------|------|
| `kweaver ds connect mysql ...` | 注册 MySQL 数据源 |
| `kweaver bkn create-from-csv <ds-id> --files data/*.csv --build` | 导入 CSV 并构建知识网络 |
| `kweaver bkn object-type list <kn-id>` | 列出自动发现的对象类型 |
| `kweaver bkn object-type query <kn-id> <ot-id> --limit 5` | 查询实例 |
| `kweaver bkn subgraph <kn-id> <instance-id> --depth 2` | 子图遍历 |
| `kweaver context-loader kn-search "..." --only-schema` | 语义搜索 |
| `kweaver bkn export <kn-id>` | 导出知识网络定义 |
| `kweaver agent chat <agent-id> -m "..."` | 对话（含 schema 上下文） |

## 与示例 01 的区别

| | 01-db-to-qa | 02-csv-to-kn |
|---|---|---|
| 数据来源 | 已有 MySQL 数据库 | 本地 CSV 文件 |
| 数据导入 | `ds connect` + `create-from-ds` | `create-from-csv`（一步完成） |
| Schema 准备 | 编写 SQL seed 文件 | 直接带 CSV |
| 网络特性展示 | 语义搜索 + 问答 | **子图遍历** + 导出 |
| 数据领域 | 供应链（BOM、采购订单） | **HR（员工、部门、项目）** |
