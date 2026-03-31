# 算子集成

## 1. 项目结构
项目的目录结构如下：

```
.
├── azure-pipelines.yml
├── docker
│   ├── Dockerfile
│   └── Makefile
├── docs
│   ├── apis
│   ├── build.sh
│   ├── data
│   └── preview.sh
├── go.mod
├── go.sum
├── helm
│   └── agent-operator-integration
├── migrations
│   ├── 1.0.0
│   ├── init.sql
│   └── readme.md
├── project.sh
├── README.md
├── server
│   ├── dbaccess
│   ├── drivenadapters
│   ├── driveradapters
│   ├── infra
│   ├── interfaces
│   ├── logics
│   ├── main.go
│   ├── mocks
│   ├── tests
│   └── utils
├── sonar-scanner.properties
└── VERSION
```


## 2. 项目依赖
- 数据库
- hydra
- user-management

## 3. 项目构建
### 3.1 编译
```shell
go build -o agent-operator-integration ./server/main.go
```
### 3.2 打包
```shell
docker build -t agent-operator-integration:latest .
```
### 3.3 部署
```shell
helm install agent-operator-integration ./helm/agent-operator-integration
```

## 4. 项目测试
- [注册查询相关](./server/tests/http/register接口相关测试数据.md)
- [更新删除相关](./server/tests/http/operator.http)

## 5. 项目运行
 - [编译运行相关脚本](./project.sh)

# API文档
## HTML 格式
- [外部接口文档](./docs/apis/api_public/operator.html)
- [内部接口文档](./docs/apis/api_private/operator.html)
## YAML 格式
- [外部接口文档](./docs/apis/api_public/operator.yaml)
- [内部接口文档](./docs/apis/api_private/operator.yaml)

# 测试数据

- [注册查询相关](./server/tests/http/register接口相关测试数据.md)
- [更新删除相关](./server/tests/http/operator.http)
- [流程场景测试](./server/tests/http/scenario.md)

## 测试文件
### JSON 文件
- [json/auth.json](./server/tests/file/json/auth.json)
- [json/file_decrypt.json](./server/tests/file/json/file_decrypt.json)
- [full_text_subdoc.jso](./server/tests/file/json/full_text_subdoc.json)

### YAML 文件
- [template.yaml](./server/tests/file/yaml/template.yaml)
- [test.yaml](./server/tests/file/yaml/test.yaml)

# 认证头参数

## 概述
微服务内部接口调用需要在HTTP头中传递认证参数，用于标识调用方的身份信息。

## 认证头参数
- `x-account-id`: 账户ID，标识调用方的唯一身份
- `x-account-type`: 账户类型，支持以下类型：
  - `user`: 用户账户
  - `app`: 应用账户
  - `anonymous`: 匿名访问

## 使用方式

### 1. 在HTTP请求中设置认证头
```go
req.Header.Set("x-account-id", "user-123")
req.Header.Set("x-account-type", "user")
```

# 算子操作工具

该工具用于对算子进行注册、更新、查询、删除、发布、下线等操作。

[工具简介](./server/tests/tool/README.MD)