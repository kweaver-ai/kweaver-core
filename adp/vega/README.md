# Vega 数据管理平台

Vega是一个企业级数据管理平台，旨在提供统一的数据连接、元数据管理和数据模型构建能力。该项目采用微服务架构，支持多种数据源，并提供统一查询接口。

## 项目结构

- [data-connection](./data-connection/) - 统一数据连接服务，支持多种数据库类型的连接和认证
- [mdl-data-model](./mdl-data-model/) - 数据模型定义与管理服务
- [mdl-data-model-job](./mdl-data-model-job/) - 数据模型相关任务调度服务
- [mdl-uniquery](./mdl-uniquery/) - 统一查询服务，提供标准化的数据访问接口
- [vega-gateway](./vega-gateway/) - API网关服务
- [vega-gateway-pro](./vega-gateway-pro/) - 高级API网关功能

## 功能特性

- **多数据源支持**：支持DM8、MariaDB等多种数据库
- **统一接口**：提供标准化的API接口访问不同数据源
- **元数据管理**：完整的元数据生命周期管理
- **安全认证**：集成多种认证方式，保障数据安全
- **可扩展架构**：基于微服务的架构，易于扩展

## 快速开始

### 环境要求

- Java 8+ (对于Java项目)
- Go 1.19+ (对于Go项目)
- Maven 3.6+
- Docker & Docker Compose
- Kubernetes (生产环境)

### 构建项目

每个子项目都有独立的构建配置，请参考各个项目的README进行构建。

```bash
# 进入具体项目目录构建
cd data-connection
mvn clean install

# 或运行Docker镜像
docker build -t vega-service-name -f ./Dockerfile .
```

## 部署

Vega使用Helm Charts进行Kubernetes部署，各组件均有对应的Helm包。

```bash
# 部署所有组件
helm install vega ./helm/vega-gateway
helm install data-connection ./helm/data-connection
helm install vega-gateway-pro ./helm/vega-gateway-pro
```

## 贡献

我们欢迎社区贡献！请查看 [CONTRIBUTING.md](./CONTRIBUTING.md) 文件了解如何参与项目开发。

## 许可证

本项目采用 Apache License 2.0 许可证。详情请参见 [LICENSE](./LICENSE) 文件。