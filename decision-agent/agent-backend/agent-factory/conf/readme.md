## 配置文件说明

### 从helm获取配置文件方法

通过Helm模板生成的配置文件包含了多个YAML配置，执行以下命令可获取完整的配置文件：

```shell
cd helm
make test
```

执行后会在 `helm/helm-template-gen.yaml` 文件中生成以下配置文件：

#### 主要配置文件

1. **agent-factory-secret.yaml** (第10行开始)
   - 路径：`/sysvol/conf/secret/agent-factory-secret.yaml`
   - 包含数据库和Redis的敏感配置信息
   - 内容：数据库连接配置、Redis连接配置等

2. **agent-factory.yaml** (第87行开始)
   - 路径：`/sysvol/conf/agent-factory.yaml`
   - 应用主配置文件
   - 内容：项目基础配置、微服务地址、模型工厂配置等

3. **mq_config.yaml** (第140行开始)
   - 消息队列配置文件
   - 内容：Kafka连接配置、认证信息等

4. **dm_svc.conf** (第80行和第185行)
   - 数据库服务配置文件
   - 内容：时区、语言、数据库连接模式等

5. **mdlconfig.json** (第166行开始)
   - 模型配置文件
   - 内容：Kafka服务连接配置等

#### 使用说明

1. 执行 `make test` 命令生成配置文件
2. 从 `helm-template-gen.yaml` 中提取所需的配置文件内容
3. 将配置文件放到对应的路径下（参考agent-go-common-pkg/cconf/config.go中配置文件路径的获取逻辑)
