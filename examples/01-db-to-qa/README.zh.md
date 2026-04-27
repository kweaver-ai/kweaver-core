# 01 · 从数据库到智能问答

> 你的数据库终于能直接回答问题了——用自然语言，不用写 SQL。

## 场景背景

供应链分析师有多年的采购和库存数据存在 MySQL 里。每次需要回答业务问题——
"哪些供应商最可靠？""哪些物料有断货风险？"——都要找 DBA 写 SQL，一来一回耗费半天。

这个示例把数据库接入知识网络和 Agent，用自然语言提问，答案来自你真实的数据。

## 示例流程

```
MySQL 数据库
     │
     ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│  数据源连接   │────▶│   知识网络    │────▶│  上下文加载器    │
│  (ds connect)│     │   (KN)       │     │   语义搜索       │
└─────────────┘     └──────────────┘     └─────────────────┘
                           │
                           ▼
                    ┌──────────────┐     ┌─────────────────┐
                    │  Schema 探索  │     │   Agent 对话     │
                    └──────────────┘     └─────────────────┘
```

0. **导入数据** — 将示例数据（虚构的智能家居供应链）导入 MySQL
1. **连接数据源** — 将 MySQL 注册到平台
2. **创建知识网络** — 自动发现表结构并构建
3. **探索 Schema** — 查看对象类型和属性
4. **语义搜索** — 用自然语言检索知识网络
5. **Agent 对话** — 向 Agent 提问业务问题

## 前置条件

```bash
# 1. 安装 KWeaver CLI
npm install -g @kweaver-ai/kweaver-sdk

# 2. 安装 MySQL 客户端（Step 0 在本机执行 mysql 导入 seed.sql）
#    macOS:  brew install mysql-client
#    Ubuntu: sudo apt install -y mysql-client

# 3. 登录 KWeaver 平台
kweaver auth login https://<platform-url>

# 4. 准备一个平台可访问的 MySQL 数据库
#    DB 用户需要 CREATE TABLE / INSERT / SELECT 权限
```

## 快速开始

```bash
cp env.sample .env
# 填写 DB_HOST、DB_NAME、DB_USER、DB_PASS（见 env.sample 中的注释）
vim .env
./run.sh
```

> **安全提示：** `.env` 已被 gitignore 排除。请勿将含有真实凭据的 `.env` 提交到版本控制。

## 配置说明

**`DB_HOST` 与 `DB_HOST_SEED`**
Step 0 的 `mysql` 在本机运行，Step 1 的 `ds connect` 由平台发起连接。
如果本机需要公网 IP 而平台需要内网 IP，分别设置：`DB_HOST`（内网）和 `DB_HOST_SEED`（公网）。
不设 `DB_HOST_SEED` 时默认使用 `DB_HOST`。

**`DEBUG=1`** 打印详细诊断信息（API 响应、配置等），不会泄露密码。

## 关键命令

```bash
kweaver ds connect mysql $DB_HOST $DB_PORT $DB_NAME \
  --account $DB_USER --password $DB_PASS --name "my-datasource"

kweaver bkn create-from-ds <datasource-id> --name "my-kn" --build

kweaver bkn object-type list <kn-id>
kweaver context-loader kn-search "供应链" --kn-id <kn-id>
kweaver agent chat <agent-id> -m "主要供应商有哪些？"
```

## 常见问题

**`ERROR 1044 Access denied`** — DB 用户对 `DB_NAME` 没有权限。请 DBA 执行：
`GRANT ALL ON your_db.* TO 'your_user'@'%';`

## 清理

脚本退出时自动清理（知识网络、数据源）。手动清理：
```bash
kweaver bkn delete <kn-id> -y
kweaver ds delete <datasource-id> -y
```
