# 函数执行综合测试用例说明

## 概述

本文档描述了 `execute_function_comprehensive_data.json` 中的综合测试用例，这些用例基于框架测试指南整理而成，涵盖了函数执行的各个方面。

## 测试数据文件

- **文件路径**: `data/data-operator-hub/agent-operator-integration/execute_function_comprehensive_data.json`
- **测试文件**: `testcases/data-operator-hub/api/function/test_execute_function_comprehensive.py`

## 测试用例分类

### 1. 基础功能测试（用例 1-7）
- 正常执行 + stdout/stderr 捕获
- 业务抛异常
- 代码里没有handler函数
- 返回值不可JSON序列化
- 超时测试
- 网络访问测试（默认禁网）
- 文件系统写入测试

### 2. 参数校验测试（用例 8-12）
- 缺少handler_code参数
- 缺少event参数
- event类型错误
- handler_code为空字符串
- Python语法错误

### 3. 边界情况测试（用例 13-14）
- 返回值缺少statusCode/body
- stderr大量输出
- stdout大量输出

### 4. 数据处理测试（用例 15-19）
- 基础数据处理（字符串处理）
- 数值计算（数学运算）
- 列表处理（数据转换）
- 字典数据处理（JSON操作）
- 条件判断和流程控制

### 5. 异常处理测试（用例 20-21）
- 异常处理和错误返回
- 日志输出测试（stdout/stderr）

### 6. 标准库测试（用例 22-28）
- JSON库使用
- datetime库使用
- re库使用（正则表达式）
- math库使用
- random库使用
- collections库使用
- base64和hashlib库使用
- itertools库使用
- functools库使用

### 7. 业务场景测试（用例 29-35）
- 用户信息处理
- 数据验证和转换
- API响应格式化
- 条件过滤和排序
- 错误处理和默认值
- 数据聚合统计

## 测试用例格式

每个测试用例包含以下字段：

```json
{
    "title": "测试用例标题",
    "code": "Python代码字符串",
    "event": {
        "参数名": "参数值"
    },
    "expected_status": 200,
    "description": "测试用例描述"
}
```

## 字段说明

- **title**: 测试用例标题，用于Allure报告显示
- **code**: Python代码字符串，包含handler函数定义
- **event**: 传递给handler函数的参数对象
- **expected_status**: 期望的HTTP状态码（200, 400, 500等）
- **description**: 测试用例的详细描述（可选）

## 运行测试

### 运行所有综合测试用例

```bash
cd /root/agent-at/agent-AT
pytest testcases/data-operator-hub/api/function/test_execute_function_comprehensive.py -v
```

### 运行特定测试用例

```bash
# 运行第1个测试用例
pytest testcases/data-operator-hub/api/function/test_execute_function_comprehensive.py::TestExecuteFunctionComprehensive::test_execute_function_comprehensive[0] -v

# 运行前5个测试用例
pytest testcases/data-operator-hub/api/function/test_execute_function_comprehensive.py::TestExecuteFunctionComprehensive::test_execute_function_comprehensive[0-4] -v
```

### 生成Allure报告

```bash
pytest testcases/data-operator-hub/api/function/test_execute_function_comprehensive.py --alluredir=./allure-results
allure serve ./allure-results
```

## 特殊用例说明

### 1. 语法错误测试
- **预期**: 返回400或500错误
- **实际行为**: 如果返回200，测试会记录警告但不会失败
- **原因**: 后端可能捕获了语法错误但返回了成功状态码

### 2. 超时测试
- **预期**: 返回500错误
- **注意**: 实际超时时间取决于后端配置，可能不会立即超时

### 3. 网络访问测试
- **预期**: 返回500错误（网络访问被禁止）
- **实际行为**: 如果返回200，说明沙箱网络未禁用

### 4. 文件系统写入测试
- **预期**: 返回500错误（文件写入被禁止）
- **实际行为**: 如果返回200，说明文件系统隔离未生效

### 5. 序列化错误测试
- **预期**: 返回500错误
- **实际行为**: 如果返回200，后端可能处理了序列化错误

## 注意事项

1. **handler函数签名**: 所有测试用例使用 `handler(event)` 签名，不使用 `context` 参数
2. **返回值格式**: 建议返回 `{"statusCode": 200, "body": {...}}` 格式，但不是必须的
3. **错误处理**: 某些错误场景可能返回200但包含错误信息（stderr字段）
4. **超时时间**: 超时测试的实际行为取决于后端配置的超时时间

## 扩展测试用例

如果需要添加新的测试用例，请按照以下步骤：

1. 在 `execute_function_comprehensive_data.json` 中添加新的测试用例对象
2. 确保JSON格式正确（可以使用JSON验证工具）
3. 运行测试验证新用例是否正常工作

## 相关文档

- 框架测试指南: 参考原始测试指南文档
- API文档: `/v1/function/execute`
- 现有测试用例: `testcases/data-operator-hub/api/function/test_execute_function.py`
