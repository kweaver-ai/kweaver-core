# MF-Model-API 测试文档

## 概述
本测试套件为 mf-model-api 项目提供全面的单元测试，目标覆盖率至少75%。

## 测试结构

```
app/test/
├── __init__.py                     # 测试包初始化
├── conftest.py                     # pytest配置和公共fixtures
├── test_commons_response.py        # 测试response模块
├── test_commons_snow_id.py         # 测试snow_id模块
├── test_core_config.py             # 测试配置模块
├── test_dao_llm_model_dao.py       # 测试DAO层
├── test_utils_str_util.py          # 测试字符串工具
├── test_utils_comment_utils.py     # 测试日志工具
├── test_utils_param_verify.py      # 测试参数验证
├── test_restful_api.py             # 测试RESTful API文档生成
├── test_controller_llm.py          # 测试LLM控制器
├── test_routers.py                 # 测试路由
├── test_app_utils.py               # 测试应用工具
└── README.md                       # 本文件
```

## 安装依赖

```bash
pip install -r requirements-test.txt
```

## 运行测试

### 运行所有测试
```bash
pytest
```

### 运行特定模块测试
```bash
pytest app/test/test_commons_snow_id.py
```

### 生成覆盖率报告
```bash
pytest --cov=app --cov-report=html --cov-report=term-missing
```

### 查看HTML覆盖率报告
```bash
# 在浏览器中打开
htmlcov/index.html
```

## 测试覆盖的模块

### Commons 模块
- ✅ `response.py` - 响应处理函数
- ✅ `snow_id.py` - 雪花ID生成器
- ✅ `restful_api.py` - RESTful API文档生成

### Utils 模块
- ✅ `str_util.py` - 字符串工具函数
- ✅ `comment_utils.py` - 日志工具函数
- ✅ `param_verify_utils.py` - 参数验证函数
- ✅ `app_utils.py` - 应用工具函数

### Core 模块
- ✅ `config.py` - 配置管理

### DAO 模块
- ✅ `llm_model_dao.py` - LLM模型数据访问

### Controller 模块
- ✅ `llm_controller.py` - LLM控制器逻辑

### Routers 模块
- ✅ `llm_router.py` - LLM路由

## 测试覆盖率目标

- **整体覆盖率**: ≥ 75%
- **关键模块覆盖率**: ≥ 80%
  - Commons: ≥ 80%
  - Utils: ≥ 80%
  - Controller: ≥ 75%

## 测试类型

### 单元测试
测试单个函数或类的功能，使用Mock隔离依赖。

### 集成测试
测试多个模块间的交互。

### 异步测试
使用 `pytest-asyncio` 测试异步函数。

## 常见问题

### Q: 如何Mock数据库连接？
A: 使用 `conftest.py` 中的 `mock_db_connection` fixture。

### Q: 如何Mock Redis？
A: 使用 `conftest.py` 中的 `mock_redis` fixture。

### Q: 如何测试异步函数？
A: 使用 `@pytest.mark.asyncio` 装饰器。

## Mock策略

1. **数据库**: 使用Mock对象模拟数据库连接和游标
2. **Redis**: 使用AsyncMock模拟Redis操作
3. **HTTP请求**: 使用aioresponses或patch aiohttp.ClientSession
4. **环境变量**: 使用patch.dict模拟环境变量

## 持续集成

测试可以集成到CI/CD流程中：

```yaml
# 示例 .github/workflows/test.yml
- name: Run tests
  run: |
    pip install -r requirements-test.txt
    pytest --cov=app --cov-report=xml
    
- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.xml
```

## 贡献指南

1. 新功能必须包含相应的测试
2. 测试覆盖率不应低于75%
3. 所有测试必须通过才能合并
4. 遵循现有的测试命名和结构规范

## 联系方式

如有问题，请联系开发团队。

