# BKN 导入导出集成测试

验证 BKN 文件（tar 格式）的导入、导出及数据一致性。

## 技术栈

- **语言**: Go 1.24
- **测试框架**: [GoConvey](https://github.com/smartystreets/goconvey) (BDD 风格)
- **配置管理**: [Viper](https://github.com/spf13/viper) (YAML + 环境变量)

## 目录结构

```
bkn/
├── bkn_test.go                               # 主测试文件（10 个用例）
├── README.md
└── helpers/
    ├── bkn_helpers.go                        # 测试辅助函数
    └── examples/                             # 测试数据
        └── k8s-network/                      # K8s 网络拓扑示例
            ├── CHECKSUM
            ├── SKILL.md
            ├── network.bkn
            ├── concept_groups/k8s.bkn
            ├── object_types/                 # node, pod, service
            ├── relation_types/               # pod_belongs_node, service_routes_pod
            ├── action_types/                 # restart_pod, cordon_node
            └── risk_types/                   # restart_pod_high_risk
```

## 运行方式

### 配置

测试配置文件位于 `../testdata/test-config.yaml`，也可通过 `BKN_TEST_` 前缀的环境变量覆盖：

```bash
export BKN_TEST_BKN_BACKEND_BASE_URL=http://my-server:8080
```

### 执行

```bash
# 从 tests 目录运行
cd bkn/bkn-backend/tests
go test ./integration_tests/bkn/ -v
```

## 测试用例清单

### 导入测试 (BKN101-BKN102)

| 编号 | 描述 | 验证点 |
|------|------|--------|
| BKN101 | 导入 k8s-network 示例 | tar 构建成功、上传返回 200、响应包含 kn_id |
| BKN102 | 导入后验证各类型已创建 | 导入 k8s-network → GET object-types、relation-types、action-types、concept-groups 列表均非空 |

### 导出测试 (BKN121-BKN122)

| 编号 | 描述 | 验证点 |
|------|------|--------|
| BKN121 | 基本导出场景 | 返回 200、响应体非空、IsValidTar 验证通过 |
| BKN122 | 验证 Content-Disposition 包含 kn_id | Content-Disposition 头非空、包含 knID |

### 负向测试 (BKN201-BKN204)

| 编号 | 描述 | 验证点 |
|------|------|--------|
| BKN201 | 导入无效文件格式（纯文本） | 返回 >= 400 |
| BKN202 | 导出不存在的知识网络 | 返回 >= 400 |
| BKN203 | 导入空文件 | 返回 >= 400 |
| BKN204 | 导入缺少 network.bkn 的 tar 包 | 合法 tar 但缺少必要文件，返回 >= 400 |

### 复杂数据测试 (BKN221)

| 编号 | 描述 | 验证点 |
|------|------|--------|
| BKN221 | 导出包含复杂结构的 BKN | 导出内容包含 object_types、relation_types、action_types 字节 |

## 测试数据说明

### k8s-network

Kubernetes 集群网络拓扑，包含 3 种对象、2 种关系、2 种行动、1 种风险。

```
Service ──routes──> Pod ──belongs──> Node
                     │                │
              restart_pod        cordon_node
              (action)           (action)
```

## 辅助函数 (helpers/bkn_helpers.go)

| 函数 | 用途 |
|------|------|
| `IsValidTar(data)` | 验证字节数据是否为合法 tar 格式 |
| `GenerateUniqueName(prefix)` | 生成带时间戳的唯一名称 |
| `BuildStringWithLength(char, length)` | 构建指定长度的字符串 |
| `DeleteTestKN(client, knID, branch, t)` | 删除测试知识网络 |
| `CleanupKNs(client, t)` | 清理所有测试 KN |
| `BuildTarFromExamplesDir(exampleName)` | 从 examples 目录构建 tar 包 |
| `BuildSimpleBKNTar(knID)` | 构建简单 tar（network.bkn + 1 个 object_type） |
| `BuildFullBKNTar(knID)` | 构建完整 tar（2 个 object_type + 1 个 relation + 1 个 action） |
| `BuildTarWithoutNetworkBKN()` | 构建缺少 network.bkn 的 tar（负向测试用） |
| `GetExampleNames()` | 获取可用示例名称列表 |
| `CreateTestObjectType(client, knID, t)` | 通过 API 创建测试对象类型 |
| `CreateTestRelationType(...)` | 通过 API 创建测试关系类型 |
| `CreateTestActionType(...)` | 通过 API 创建测试行动类型 |
| `VerifyObjectTypesExist(client, knID, t)` | 验证对象类型是否存在 |
| `VerifyRelationTypesExist(client, knID, t)` | 验证关系类型是否存在 |
| `VerifyActionTypesExist(client, knID, t)` | 验证行动类型是否存在 |
| `VerifyConceptGroupsExist(client, knID, t)` | 验证概念分组是否存在 |

## API 端点覆盖

| 端点 | 方法 | 用途 | 测试覆盖 |
|------|------|------|----------|
| `/api/bkn-backend/v1/bkns` | POST | 导入 BKN | BKN101-102 |
| `/api/bkn-backend/v1/bkns/{knID}` | GET | 导出 BKN | BKN121-122, BKN221 |
| `/api/bkn-backend/v1/knowledge-networks/{knID}/object-types` | GET | 列出对象类型 | BKN102 |
| `/api/bkn-backend/v1/knowledge-networks/{knID}/relation-types` | GET | 列出关系类型 | BKN102 |
| `/api/bkn-backend/v1/knowledge-networks/{knID}/action-types` | GET | 列出行动类型 | BKN102 |
| `/api/bkn-backend/v1/knowledge-networks/{knID}/concept-groups` | GET | 列出概念分组 | BKN102 |
| `/api/bkn-backend/v1/knowledge-networks` | GET | 列出 KN | 清理（teardown） |
| `/api/bkn-backend/v1/knowledge-networks/{knID}` | DELETE | 删除 KN | 清理（teardown） |

## 建议补充的测试场景

| 场景 | 优先级 | 说明 |
|------|--------|------|
| 幂等性 | 高 | 同一 BKN 导入两次，验证无报错且不产生重复数据 |
| 覆盖/更新语义 | 中 | 已有数据的 KN 再次导入，验证更新而非重复创建 |
| 默认分支行为 | 中 | 不传 branch 参数时的默认行为 |
