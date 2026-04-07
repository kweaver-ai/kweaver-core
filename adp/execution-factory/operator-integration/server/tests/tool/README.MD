# 算子操作工具

## 工具简介

该工具用于对算子进行注册、更新、查询、删除、发布、下线等操作。

## 环境准备

```bash
# Python 3.10+
python3 -m pip install requests==2.31.0 pyyaml==6.0.1 argparse==1.4.0
```

## 使用方式

本工具支持两种使用方式：**直接运行 Python 脚本** 或 **编译成二进制文件后运行**。

### 方式一：直接使用 Python 脚本（推荐）

```bash
# 查看帮助
python3 operator_client.py -h

# 通用命令格式
python3 operator_client.py --config=<配置文件路径> <命令>
```

### 方式二：编译成二进制文件

如需生成独立可执行文件，可使用 PyInstaller 进行编译：

```bash
# 安装 PyInstaller
python3 -m pip install pyinstaller==6.2.0

# 编译成二进制文件（输出到 dist/ 目录）
pyinstaller --onefile --name operator-cli operator_client.py

# 使用编译后的二进制文件
./dist/operator-cli --config=<配置文件路径> <命令>
```

## 配置说明

> 1. 工具需要配置文件，配置文件格式为 YAML，示例见 [config.yaml](./config.yaml)
> 2. 配置文件中需设置 `token`（用户登录凭证）用于认证
> 3. 配置文件中需设置 `base_url`（平台地址）用于请求
> 4. 其他必填参数详情见配置文件
> 5. API 详情见文档：[DIP算子服务公有接口协议](../../../docs/apis/api_public/operator.html)

## 命令列表

```bash
# 注册算子
python3 operator_client.py --config=config.yaml register

# 更新算子
python3 operator_client.py --config=config.yaml update

# 查询列表
python3 operator_client.py --config=config.yaml list

# 查询详情
python3 operator_client.py --config=config.yaml detail

# 删除算子
python3 operator_client.py --config=config.yaml delete

# 发布算子
python3 operator_client.py --config=config.yaml publish

# 下线算子
python3 operator_client.py --config=config.yaml offline
```
