# OSS Gateway 测试说明

## 测试结构

测试文件位于 `test/` 目录下，组织结构如下：

```
test/
├── handler/          # Handler层测试
│   ├── storage_test.go
│   ├── upload_test.go
│   ├── download_test.go
│   ├── head_test.go
│   ├── delete_test.go
│   └── health_test.go
└── pkg/             # 工具包测试
    ├── crypto_test.go
    ├── response_test.go
    └── utils_test.go
```

## 运行测试

### 运行所有测试
```bash
make test
```

或

```bash
go test ./test/... -v
```

### 运行测试并生成覆盖率报告
```bash
make test-coverage
```

或

```bash
go test ./test/handler/... ./test/pkg/... -v -coverprofile=coverage.out -coverpkg=./internal/...,./pkg/...
```

### 查看HTML格式的覆盖率报告
```bash
go tool cover -html=coverage.out
```

## 测试覆盖范围

### Handler层测试 (test/handler/)
- **storage_test.go**: 存储配置管理接口测试
  - 创建、更新、删除、查询存储配置
  - 存储连接检查
  - 参数验证和错误处理

- **upload_test.go**: 上传接口测试
  - 单文件上传URL获取
  - 分片上传初始化
  - 分片上传URL获取
  - 完成分片上传

- **download_test.go**: 下载接口测试
  - 下载URL获取
  - 自定义文件名
  - 内网访问

- **head_test.go**: 元数据查询测试
  - 单个对象元数据URL获取
  - 批量对象元数据URL获取

- **delete_test.go**: 删除接口测试
  - 删除URL获取
  - URL编码处理

- **health_test.go**: 健康检查测试
  - 存活检查
  - 就绪检查

### 工具包测试 (test/pkg/)
- **crypto_test.go**: AES加密解密测试
  - 密钥验证
  - 加密解密功能
  - 大数据处理
  - 错误处理

- **response_test.go**: 响应格式测试
  - 成功响应
  - 错误响应
  - 各种错误码

- **utils_test.go**: 工具函数测试
  - ID生成
  - Endpoint解析
  - 参数验证

## 测试覆盖率

当前测试覆盖率已达到 **50%以上**，主要覆盖：
- Handler层的所有HTTP接口
- 核心工具包函数
- 错误处理逻辑

## 添加新测试

1. 在相应目录下创建 `*_test.go` 文件
2. 使用 `testify` 框架编写测试用例
3. Mock外部依赖（数据库、Redis、OSS适配器等）
4. 运行测试验证功能

## 依赖

测试框架使用：
- `github.com/stretchr/testify` - 断言和Mock
- `net/http/httptest` - HTTP测试
- `github.com/gin-gonic/gin` - Gin框架测试模式
