# 自动配置脚本使用说明

## 简介

一个自动化配置脚本 `auto_config.sh`，用于快速配置数据源、导入业务知识网络和DataAgent。

## 配置文件说明

在执行脚本前，需要先配置 `config.env` 文件，该文件包含了认证信息和数据源配置。

### 配置文件

`config.env`

### 配置项说明

```env
# 认证信息
USERNAME=test        # 登录用户名
PASSWORD=111111      # 登录密码

# 数据源配置

# 数据库类型
DS_TYPE=mysql        # 数据库类型，支持：mysql、maria等
# 连接名称
DS_NAME=数据源名称  # 数据源名称
# 数据库名称
DS_DATABASE_NAME=data    # 数据库名称
# 模式名称
DS_SCHEMA_NAME=         # 数据库模式名称（如PostgreSQL的schema）
# 连接方式
DS_CONNECT_PROTOCOL=jdbc  # 连接协议
# 连接地址
DS_HOST=127.0.0.1        # 数据库主机地址
# 端口
DS_PORT=3306             # 数据库端口
# 用户名
DS_USERNAME=root         # 数据库登录用户名
# 密码
DS_PASSWORD=your_password  # 数据库登录密码
# 备注
DS_COMMENT=备注信息  # 数据源备注信息
```

## 环境准备

### 前置条件
1. 登录系统工作台，创建kweaver账号(如：test),访问 https://ip/studio 登录页，登录test用户（默认密码：123456）后下方提示修改密码，进行密码修改，修改密码后才可登录。
2. 为当前用户添加对应权限：添加数据源、导入业务知识网络、导入DataAgent权限 导入数据流。

### 1. 安装依赖（无需执行）

- **Git Bash**：用于执行Shell脚本
- **curl**：用于发送HTTP请求
- **openssl**：用于密码加密

### 2. 赋予脚本执行权限

在Git Bash或WSL中执行以下命令：

```bash
chmod +x auto_config.sh
```

## 执行脚本

### 执行方式

在Bash中执行脚本，支持两种执行模式：

#### 1. 完整执行模式

执行全部4个步骤：获取token、创建数据源并扫描、导入业务知识网络、导入DataAgent、导入数据流。

```bash
./auto_config.sh agent.json HD供应链业务知识网络.json dataflow.json
```

### 执行示例

1. 赋予脚本执行权限：

```bash
chmod +x auto_config.sh
```

2. 执行完整流程：

```bash
./auto_config.sh agent.json HD供应链业务知识网络.json dataflow.json
```

3. 登录 Studio 做后续配置：

- 使用 `test` 账号登录 Studio
- 编辑已导入的 **HD供应链Agent** 并发布
- 编辑业务知识网络，在“对象类”中选择新建的数据源并关联其中对应的视图

#### 2. 单步执行模式

执行指定的单个步骤：

```bash
# 仅执行步骤1：获取token
./auto_config.sh --step 1

# 仅执行步骤2：创建数据源并扫描
./auto_config.sh --step 2

# 仅执行步骤3：导入业务知识网络
./auto_config.sh --step 3 业务知识网络.json

# 仅执行步骤4：导入DataAgent
./auto_config.sh --step 4 agent.json

# 仅执行步骤5：导入数据流
./auto_config.sh --step 5 数据流.json
```

### 参数说明

- `agent.json`：DataAgent配置文件
- `业务知识网络.json`：业务知识网络JSON文件
- `数据流.json`：数据流JSON文件

## 执行流程

1. **获取token**：登录系统获取认证令牌
2. **创建数据源并扫描**：根据配置文件创建数据源并执行扫描
3. **导入业务知识网络**：导入业务知识网络配置
4. **导入DataAgent**：导入DataAgent配置
5. **导入数据流**：导入数据流配置

## 注意事项

1. 确保配置文件中的数据库连接信息正确无误
2. 确保目标系统可以正常访问
3. 执行脚本时需要有足够的权限
4. 脚本会自动获取本机IP地址，用于构建API请求URL

## 常见问题

### 1. 无法获取本机IP地址

**解决方法**：手动设置环境变量 `IP_ADDRESS`

```bash
export IP_ADDRESS=192.168.1.100  # 替换为实际IP地址
```

**删除手动设置环境变量**：  `unset IP_ADDRESS`

### 2. 登录失败

**解决方法**：检查 `config.env` 文件中的用户名和密码是否正确

### 3. 数据源创建失败

**解决方法**：检查数据库连接信息是否正确，确保数据库服务正常运行

### 4. 脚本执行权限不足

**解决方法**：执行 `chmod +x auto_config.sh` 赋予执行权限

## 支持的数据库类型

- mysql
- maria


