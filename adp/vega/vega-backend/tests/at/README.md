# VEGA Manager - AT（Acceptance Test）测试

## 概述

这是VEGA Manager的验收测试（AT测试），采用**纯黑盒端到端测试**方式，通过HTTP API验证服务功能的正确性。

## 测试特点

- **黑盒测试**: 只通过HTTP API与VEGA Manager交互，不访问内部代码或数据库
- **端到端**: 测试完整的请求-响应流程，包括路由、业务逻辑、数据持久化
- **独立运行**: 测试不启动服务，需要用户预先启动VEGA Manager
- **真实场景**: 测试实际部署环境的行为

## 目录结构

```
tests/
├── at/                              # AT测试
│   ├── catalog/                     # Catalog测试模块（独立）
│   │   ├── physical/                # Physical Catalog测试
│   │   │   ├── internal/            # 模块内部共享代码
│   │   │   │   ├── helpers.go
│   │   │   │   ├── fixtures.go
│   │   │   │   ├── suite.go
│   │   │   │   ├── test_cases.go
│   │   │   │   ├── mariadb_builder.go
│   │   │   │   └── opensearch_builder.go
│   │   │   ├── mariadb/               # MariaDB Catalog测试
│   │   │   │   ├── catalog_test.go
│   │   │   │   ├── mariadb_specific_test.go
│   │   │   │   └── README.md
│   │   │   └── opensearch/          # OpenSearch Catalog测试
│   │   │       ├── catalog_test.go
│   │   │       ├── opensearch_specific_test.go
│   │   │       └── README.md
│   │   └── logical/                 # Logical Catalog测试
│   │       ├── internal/
│   │       │   └── helpers.go
│   │       ├── catalog_test.go
│   │       └── README.md
│   │
│   ├── resource/                    # Resource测试模块（独立）
│   │   ├── internal/                # 模块内部共享代码
│   │   │   ├── helpers.go
│   │   │   ├── fixtures.go
│   │   │   ├── suite.go
│   │   │   ├── test_cases.go
│   │   │   └── catalog_builder.go
│   │   ├── mariadb/
│   │   │   ├── resource_test.go
│   │   │   └── README.md
│   │   └── opensearch/
│   │       ├── resource_test.go
│   │       └── README.md
│   │
│   ├── setup/                       # 测试配置（共享）
│   │   └── config.go                # 配置加载
│   ├── testdata/                    # 测试配置文件（共享）
│   │   ├── test-config.yaml.example # 配置模板
│   │   └── test-config.yaml         # 实际配置（不提交到Git）
│   └── README.md                    # 本文件
│
└── testutil/                        # 测试工具包（共享）
    └── http_client.go               # HTTP客户端封装
```


## 运行测试前的准备

### 1. 启动VEGA Manager服务

```bash
cd /mnt/c/aishu_code/vega-backend/server
go run main.go

# 或使用配置文件
go run main.go -config ./config/vega-backend-config.yaml

# 确保服务成功启动，默认监听 http://localhost:8080
```

### 2. 准备测试目标MariaDB

AT测试需要一个可访问的MariaDB实例来测试catalog连接功能。

#### 使用Docker启动MariaDB（推荐）

```bash
docker run -d --name test-mariadb \
  -e MARIADB_ROOT_PASSWORD=testpass123 \
  -e MARIADB_DATABASE=testdb \
  -p 3306:3306 \
  mariadb:10.6
```

#### 或使用已有MariaDB实例

确保MariaDB可访问，并记录以下信息：
- Host
- Port
- Database
- Username
- Password

### 3. 配置测试文件

```bash
cd /mnt/c/aishu_code/vega-backend/server/tests/at/testdata

# 复制配置模板
cp test-config.yaml.example test-config.yaml

# 编辑test-config.yaml，填入实际配置
# 1. vega_manager.base_url: VEGA Manager服务地址（默认 http://localhost:8080）
# 2. target_mariadb: MariaDB连接信息
nano test-config.yaml
```

配置示例：

```yaml
vega_manager:
  base_url: http://localhost:8080

target_mariadb:
  host: localhost
  port: 3306
  database: testdb
  username: root
  password: testpass123
```

## 运行测试

### 使用Makefile（推荐）

```bash
cd /mnt/c/aishu_code/vega-backend

# 运行所有Catalog AT测试
make test-at-catalog

# 运行所有Resource AT测试
make test-at-resource

# 运行所有AT测试
make test-at

# 检查测试配置是否就绪
make test-at-setup
```

### 直接使用go test

```bash
cd /mnt/c/aishu_code/vega-backend/server

# 运行所有Catalog测试（包含MariaDB和OpenSearch）
go test -v ./tests/at/catalog/...

# 运行所有Resource测试（包含MariaDB和OpenSearch）
go test -v ./tests/at/resource/...

# 运行MariaDB Catalog测试
go test -v ./tests/at/catalog/mariadb/...

# 运行OpenSearch Catalog测试
go test -v ./tests/at/catalog/opensearch/...

# 运行MariaDB Resource测试
go test -v ./tests/at/resource/mariadb/...

# 运行特定测试（例如：MariaDB特定测试）
go test -v ./tests/at/catalog/mariadb/... -run TestMariaDBSpecific

# 详细输出（禁用缓存）
go test -v -count=1 ./tests/at/catalog/... ./tests/at/resource/...

# 设置超时
go test -v ./tests/at/catalog/... -timeout 5m
```

## 测试用例清单

### 测试文件概览

| 测试文件 | 测试内容 | 用例数量 |
|----------|----------|----------|
| catalog_create_test.go | Catalog创建测试（MariaDB） | 30个 |
| catalog_read_test.go | Catalog查询测试 | 11个 |
| catalog_update_test.go | Catalog更新测试 | 13个 |
| catalog_delete_test.go | Catalog删除测试 | 13个 |
| catalog_opensearch_test.go | OpenSearch Catalog完整测试 | 16个 |
| **总计** | **完整CRUD + 多数据源** | **83个** |

### Catalog创建测试（30个用例）- catalog_create_test.go

#### 正向测试（10个）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| TC001  | 创建MariaDB physical catalog - 基本场景 | 201 Created，返回完整catalog对象 |
| TC002  | 创建catalog - 最小字段（仅name和type） | 201 Created |
| TC003  | 创建catalog - 完整字段 | 201 Created，所有字段正确存储 |
| TC004  | 创建后立即查询 | 创建成功，查询返回一致数据 |
| TC005  | 创建带options的MariaDB catalog | 201 Created |
| TC006  | 创建logical类型catalog | 201 Created |
| TC007  | 创建后测试连接成功 | 连接测试成功，返回version和latency |
| TC008  | 获取catalog状态 | 200 OK，返回health_check_status |
| TC009  | 创建多个catalog，列表查询 | 创建3个，列表查询返回正确total |
| TC010  | Tags数组测试 | 空数组、单个、多个tags均成功 |

#### 负向测试（10个）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| TC101  | 缺少必填字段 - name | 400 Bad Request |
| TC102  | 缺少必填字段 - type | 400 Bad Request |
| TC103  | 无效的type值 | 400 Bad Request，错误信息包含有效值 |
| TC104  | 重复的catalog名称 | 第二次创建返回409 Conflict |
| TC105  | 无效JSON格式 | 400 Bad Request |
| TC106  | 错误的Content-Type | 400/406/415 |
| TC107  | 超长name字段（>255字符） | 400 Bad Request |
| TC108  | 超长description字段（>1000字符） | 400 Bad Request或允许 |
| TC109  | 错误的MariaDB连接密码 | test-connector失败 |
| TC110  | 不存在的MariaDB数据库 | test-connector失败 |

#### 边界测试（5个）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| TC201  | name长度255字符（最大允许） | 201 Created |
| TC202  | name长度256字符（边界） | 201或400（根据实际限制） |
| TC203  | 空tags数组 | 201 Created |
| TC204  | 大量tags（100个） | 201或400（根据实际限制） |
| TC205  | 特殊字符名称（中文、连字符等） | 201 Created |

#### 并发测试（2个）

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| TC301  | 并发创建10个不同catalog | 全部成功 |
| TC302  | 并发创建5个相同名称catalog | 1个成功，4个返回409 |

### Catalog查询测试（11个用例）- catalog_read_test.go

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| TC401  | 查询单个catalog - 成功场景 | 200 OK，返回完整catalog信息 |
| TC402  | 查询不存在的catalog | 404 Not Found |
| TC403  | 查询catalog - 验证所有字段返回 | 所有字段完整返回 |
| TC404  | 列表查询 - 基本场景 | 200 OK，返回items和total |
| TC405  | 列表查询 - 分页测试 | 正确分页，每页数据符合预期 |
| TC406  | 列表查询 - 按type过滤 | 只返回指定type的catalog |
| TC407  | 列表查询 - 空结果测试 | 200 OK，items为空 |
| TC408  | 列表查询 - 无效分页参数 | 返回400或自动纠正 |
| TC409  | 列表查询 - 默认分页参数 | 使用默认分页，正常返回 |
| TC410  | 查询catalog状态 | 200 OK，返回health_check_status |
| TC411  | 查询不存在catalog的状态 | 404 Not Found |

### Catalog更新测试（13个用例）- catalog_update_test.go

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| TC501  | 更新catalog name | 200 OK，name更新成功 |
| TC502  | 更新catalog description | 200 OK，description更新成功 |
| TC503  | 更新catalog tags | 200 OK，tags更新成功 |
| TC504  | 更新catalog connector_config | 200 OK，配置更新成功 |
| TC505  | 同时更新多个字段 | 200 OK，所有字段更新成功 |
| TC506  | 部分字段更新 | 200 OK，只更新指定字段 |
| TC507  | 更新不存在的catalog | 404 Not Found |
| TC508  | 更新name为已存在的名称 | 409 Conflict |
| TC509  | 更新name为空字符串 | 400 Bad Request |
| TC510  | 更新name超长 | 400或413 |
| TC511  | 更新空payload | 200或400 |
| TC512  | 验证update_time更新 | update_time自动更新 |
| TC513  | 验证create_time不变 | create_time保持不变 |

### Catalog删除测试（13个用例）- catalog_delete_test.go

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| TC601  | 删除catalog - 成功场景 | 204 No Content |
| TC602  | 删除不存在的catalog | 404 Not Found |
| TC603  | 重复删除同一个catalog | 第二次返回404 |
| TC604  | 删除后不能更新 | 更新返回404 |
| TC605  | 删除后可以创建同名catalog | 201 Created，新ID |
| TC606  | 删除physical类型catalog | 204 No Content |
| TC607  | 删除logical类型catalog | 204 No Content |
| TC608  | 删除包含完整字段的catalog | 204 No Content |
| TC609  | 批量删除多个catalog | 所有删除成功 |
| TC610  | 删除后列表中不再显示 | 列表中不存在 |
| TC611  | 使用无效ID删除 | 400或404 |
| TC612  | 删除catalog后状态不可查 | 404 Not Found |
| TC613  | 删除catalog后不能测试连接 | 404 Not Found |

### OpenSearch Catalog测试（16个用例）- catalog_opensearch_test.go

| 用例ID | 测试场景 | 预期结果 |
|--------|----------|----------|
| TC701  | 创建OpenSearch catalog - 基本场景 | 201 Created，connector_type为opensearch |
| TC702  | 创建OpenSearch catalog - 完整字段 | 201 Created，所有字段正确 |
| TC703  | 创建OpenSearch catalog - 带options | 201 Created |
| TC704  | 创建OpenSearch catalog - SSL配置 | 201或400/500（取决于环境） |
| TC705  | 查询OpenSearch catalog | 200 OK，返回opensearch类型catalog |
| TC706  | 列表查询包含OpenSearch catalog | 200 OK，列表包含opensearch类型 |
| TC707  | 更新OpenSearch catalog name | 200 OK，name更新成功 |
| TC708  | 更新OpenSearch catalog connector_config | 200 OK，配置更新成功 |
| TC709  | 更新OpenSearch catalog tags | 200 OK，tags更新成功 |
| TC710  | 删除OpenSearch catalog | 204 No Content |
| TC711  | OpenSearch catalog测试连接 - 成功 | 200 OK，success=true |
| TC712  | OpenSearch catalog测试连接 - 错误密码 | success=false |
| TC713  | OpenSearch catalog测试连接 - 无效host | success=false |
| TC714  | OpenSearch完整CRUD流程 | 完整流程成功 |
| TC715  | MariaDB和OpenSearch catalog共存 | 两者共存，列表可查询 |
| TC716  | OpenSearch catalog状态查询 | 200 OK，返回状态 |

## 测试输出示例

成功运行示例：

```
=== RUN   TestCatalogCreateATSuite
✓ AT测试环境就绪，VEGA Manager: http://localhost:8080
=== RUN   TestCatalogCreateATSuite/TestTC001_CreateMariaDBCatalog_BasicScenario
=== RUN   TestCatalogCreateATSuite/TestTC002_CreateCatalog_MinimalFields
=== RUN   TestCatalogCreateATSuite/TestTC003_CreateCatalog_FullFields
...
--- PASS: TestCatalogCreateATSuite (45.23s)
    --- PASS: TestCatalogCreateATSuite/TestTC001_CreateMariaDBCatalog_BasicScenario (1.20s)
    --- PASS: TestCatalogCreateATSuite/TestTC002_CreateCatalog_MinimalFields (0.85s)
    --- PASS: TestCatalogCreateATSuite/TestTC003_CreateCatalog_FullFields (1.10s)
PASS
ok      vega-backend/tests/at/catalog   45.234s
```

## 故障排查

### 错误: 无法连接到VEGA Manager服务

```
无法连接到VEGA Manager服务: http://localhost:8080，请确保服务已启动
```

**解决方法**:
1. 确认VEGA Manager服务已启动: `ps aux | grep vega-backend`
2. 检查服务监听端口: `netstat -tuln | grep 8080`
3. 验证配置文件中的base_url是否正确

### 错误: 测试配置文件不存在

```
请确保 tests/at/testdata/test-config.yaml 已配置
```

**解决方法**:
```bash
cd server/tests/at/testdata
cp test-config.yaml.example test-config.yaml
# 编辑test-config.yaml填入实际配置
```

### 错误: 测试连接失败

如果TC007或TC109测试失败，检查：

1. MariaDB是否正常运行: `docker ps | grep test-mariadb`
2. MariaDB连接信息是否正确
3. 网络是否可达: `telnet localhost 3306`
4. 数据库是否存在: `mariadb -h localhost -u root -p -e "SHOW DATABASES;"`

### 错误: 测试超时

```
panic: test timed out after 2m0s
```

**解决方法**:
- 增加超时时间: `go test -v ./tests/at/catalog/... -timeout 10m`
- 检查VEGA Manager服务响应是否正常
- 检查MariaDB连接是否稳定

## 环境变量覆盖配置

可以使用环境变量覆盖配置文件中的值：

```bash
export VEGA_TEST_VEGA_MANAGER_BASE_URL=http://localhost:8080
export VEGA_TEST_TARGET_MARIADB_HOST=localhost
export VEGA_TEST_TARGET_MARIADB_PORT=3306
export VEGA_TEST_TARGET_MARIADB_DATABASE=testdb
export VEGA_TEST_TARGET_MARIADB_USERNAME=root
export VEGA_TEST_TARGET_MARIADB_PASSWORD=testpass123

# 运行测试
go test -v ./tests/at/catalog/...
```

## 测试数据清理

AT测试创建的catalog使用唯一名称（带时间戳），避免冲突。测试完成后，可以手动清理：

```bash
# 通过API删除（推荐）
curl -X DELETE http://localhost:8080/api/vega-backend/v1/catalogs/{catalog_id} \
  -H "X-Account-ID: test-user-001"

# 或直接操作数据库
mariadb -h localhost -u root -p -e "DELETE FROM testdb.t_catalog WHERE name LIKE 'test-%';"
```

## CI/CD集成

在CI环境中运行AT测试：

```yaml
# .github/workflows/at-test.yml
name: AT Tests

on: [push, pull_request]

jobs:
  at-test:
    runs-on: ubuntu-latest

    services:
      mariadb:
        image: mariadb:10.6
        env:
          MARIADB_ROOT_PASSWORD: testpass123
          MARIADB_DATABASE: testdb
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mariadbadmin ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=3

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Start VEGA Manager
        run: |
          cd server
          go run main.go &
          sleep 5

      - name: Configure AT tests
        run: |
          cd server/tests/at/testdata
          cat > test-config.yaml <<EOF
          vega_manager:
            base_url: http://localhost:8080
          target_mariadb:
            host: 127.0.0.1
            port: 3306
            database: testdb
            username: root
            password: testpass123
          EOF

      - name: Run AT tests
        run: |
          cd server
          go test -v ./tests/at/... -timeout 5m
```

## 测试框架

测试使用 **GoConvey** 框架，提供：
- 清晰的BDD（行为驱动开发）风格测试
- 嵌套的Convey结构组织测试场景
- So断言进行验证

### GoConvey断言示例

```go
// 相等断言
So(resp.StatusCode, ShouldEqual, 201)

// 非空断言
So(resp.Body["id"], ShouldNotBeEmpty)
So(resp.Body, ShouldNotBeNil)
So(resp.Error, ShouldBeNil)

// 数值比较
So(latency, ShouldBeGreaterThan, 0.0)
So(count, ShouldBeGreaterThanOrEqualTo, 3)

// 集合断言
So(tags, ShouldHaveLength, 4)
So([]int{400, 415}, ShouldContain, resp.StatusCode)

// 布尔断言
So(success, ShouldBeTrue)
So(failed, ShouldBeFalse)
```

## 扩展测试

### 添加新的测试用例

在 `catalog_create_test.go` 中添加新的Convey场景：

```go
Convey("TC011: 你的测试描述", func() {
    payload := testutil.BuildBasicMariaDBPayload(config.TargetMariaDB)

    resp := client.POST("/api/vega-backend/v1/catalogs", payload)

    So(resp.StatusCode, ShouldEqual, 201)
    So(resp.Body["id"], ShouldNotBeEmpty)
    // 更多断言...
})
```

### 添加新的数据源类型测试

创建新文件 `catalog_create_postgresql_test.go`：

```go
package catalog

// 复用相同的测试框架
// 使用testutil.BuildPostgreSQLPayload()生成payload
```

## 参考资料

- [计划文档](../../../docs/AT_TEST_PLAN.md)
- [VEGA Manager API文档](../../../docs/API.md)
- [GoConvey文档](https://github.com/smartystreets/goconvey)
