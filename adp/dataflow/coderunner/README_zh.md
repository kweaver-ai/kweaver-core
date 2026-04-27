# CodeRunner Service (代码执行服务)

CodeRunner 是一个基于 Python 的代码执行和数据处理服务，提供沙箱化代码执行、数据分析、文档处理和 OCR 功能。它由两个主要模块组成：用于安全代码执行的 CodeRunner 模块和用于数据处理和文件操作的 DataFlow Tools 模块。

[English Documentation](README.md)

## 核心架构

CodeRunner 采用微服务架构，包含两个独立模块：

```
┌─────────────────────────────────────────────────────────┐
│                   CodeRunner 模块                       │
│              (沙箱化代码执行服务)                        │
│  - RestrictedPython 代码执行                            │
│  - Python 包管理                                        │
│  - 安全执行环境                                          │
│  - 健康监控                                              │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                 DataFlow Tools 模块                     │
│              (数据处理与分析)                            │
│  - 数据分析和转换                                        │
│  - 文档处理 (Word, Excel, PDF)                          │
│  - OCR 文本提取                                          │
│  - 文件操作和管理                                        │
└─────────────────────────────────────────────────────────┘
```

### 目录结构说明

- **`coderunner/`**: 沙箱化代码执行模块
  - `src/driveradapters/`: HTTP API 处理器 (runner, pypkg, health)
  - `src/drivenadapters/`: 外部服务集成
  - `src/logics/`: 核心业务逻辑
  - `src/utils/`: 工具函数
- **`dataflowtools/`**: 数据处理和分析模块
  - `src/driveradapters/`: HTTP API 处理器 (data_analysis, file, ocr, tag, py_pkg)
  - `src/drivenadapters/`: 外部服务集成
  - `src/logics/`: 文件处理器和分析逻辑
- **`deps/`**: Python 包依赖
- **`helm/`**: Kubernetes 部署 Helm Charts
- **`migrations/`**: 数据库迁移脚本

## 主要功能

### CodeRunner 模块

#### 1. 沙箱化代码执行
- **RestrictedPython**: 使用受限功能安全执行 Python 代码
- **隔离环境**: 在沙箱环境中执行用户代码
- **超时控制**: 可配置的执行超时限制
- **资源限制**: 内存和 CPU 使用约束

#### 2. Python 包管理
- **包安装**: 动态安装 Python 包
- **包列表**: 查询已安装的包
- **包隔离**: 每次执行使用独立的包环境
- **依赖管理**: 自动处理包依赖关系

#### 3. 安全特性
- **代码沙箱**: 限制对系统资源的访问
- **导入控制**: 基于白名单的模块导入
- **执行隔离**: 隔离的执行上下文
- **错误处理**: 安全的错误报告，不暴露系统细节

### DataFlow Tools 模块

#### 1. 数据分析
- **数据转换**: 使用 pandas 处理和转换数据
- **统计分析**: 执行数据分析操作
- **数据验证**: 根据模式验证数据
- **自定义操作**: 执行自定义数据处理逻辑

#### 2. 文档处理
- **Word 文档**: 读写 .doc 和 .docx 文件
- **Excel 文件**: 处理 .xls 和 .xlsx 电子表格
- **PDF 文件**: 从 PDF 文档中提取文本和数据
- **模板处理**: 从模板生成文档

#### 3. OCR 功能
- **文本提取**: 从图像和扫描文档中提取文本
- **多格式支持**: 支持各种图像格式
- **条形码/二维码**: 解码条形码和二维码
- **文档扫描**: 处理扫描文档

#### 4. 文件操作
- **文件上传/下载**: 处理文件传输
- **格式转换**: 在不同文件格式之间转换
- **文件元数据**: 提取和管理文件元数据
- **批量处理**: 批量处理多个文件

## 技术栈

### CodeRunner 模块
- **语言**: Python 3.9
- **Web 框架**: Tornado 6.5
- **沙箱**: RestrictedPython 8.0
- **数据处理**: pandas 2.0.3, numpy 1.24.4
- **文档库**: python-docx, openpyxl, PyMuPDF, pdfplumber
- **OCR**: opencv-python, pyzbar
- **HTTP 客户端**: httpx, requests

### DataFlow Tools 模块
- **语言**: Python 3.9
- **Web 框架**: Tornado 6.4.2
- **数据库**: PyMySQL, SQLAlchemy
- **文档处理**: python-docx, openpyxl, pypandoc
- **数据分析**: pandas (通过 coderunner)
- **缓存**: Redis

## 前置要求

- Python 3.9+
- Docker (用于容器化部署)
- Redis (用于 dataflowtools 缓存)
- MySQL (用于 dataflowtools 数据存储)

## 配置说明

两个模块都使用环境变量进行配置（从 `.env` 文件加载）：

### CodeRunner 模块配置

```bash
# 服务器配置
PORT=8080
HOST=0.0.0.0

# 执行限制
MAX_EXECUTION_TIME=30
MAX_MEMORY_MB=512

# 包管理
ALLOW_PACKAGE_INSTALL=true
PACKAGE_INDEX_URL=https://pypi.org/simple
```

### DataFlow Tools 模块配置

```bash
# 服务器配置
PORT=8081
HOST=0.0.0.0

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_NAME=dataflow
DB_USER=root
DB_PASSWORD=password

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379

# 外部服务
OCR_SERVICE_URL=http://ocr-service:8080
```

## API 文档

### CodeRunner 模块端点

#### 代码执行
- `POST /api/runner/execute` - 在沙箱环境中执行 Python 代码
  - 请求体: `{"code": "print('Hello')", "timeout": 30}`
  - 响应: `{"result": "Hello\n", "status": "success"}`

#### 包管理
- `POST /api/pypkg/install` - 安装 Python 包
- `GET /api/pypkg/list` - 列出已安装的包
- `DELETE /api/pypkg/uninstall` - 卸载包

#### 健康检查
- `GET /api/health` - 服务健康状态

### DataFlow Tools 模块端点

#### 数据分析
- `POST /api/data_analysis/analyze` - 执行数据分析
- `POST /api/data_analysis/transform` - 转换数据

#### 文件操作
- `POST /api/file/upload` - 上传文件
- `GET /api/file/download/{file_id}` - 下载文件
- `POST /api/file/convert` - 转换文件格式
- `DELETE /api/file/{file_id}` - 删除文件

#### OCR
- `POST /api/ocr/extract` - 从图像中提取文本
- `POST /api/ocr/barcode` - 解码条形码/二维码

#### 标签管理
- `POST /api/tag/create` - 创建标签
- `GET /api/tag/list` - 列出标签
- `PUT /api/tag/update` - 更新标签
- `DELETE /api/tag/delete` - 删除标签

#### 包管理
- `POST /api/py_pkg/install` - 安装 Python 包
- `GET /api/py_pkg/list` - 列出包

#### 健康检查
- `GET /api/health` - 服务健康状态

## 构建和运行

### 本地开发

#### CodeRunner 模块

```bash
cd coderunner
pip install -r requirements.txt
python src/main.py
```

#### DataFlow Tools 模块

```bash
cd dataflowtools
pip install -r requirements.txt
python src/main.py
```

### Docker 构建

#### CodeRunner 模块

```bash
docker build -t coderunner:latest -f Dockerfile.coderunner .
docker run -p 8080:8080 coderunner:latest
```

#### DataFlow Tools 模块

```bash
docker build -t dataflowtools:latest -f Dockerfile.dataflowtools .
docker run -p 8081:8081 dataflowtools:latest
```

### 包初始化

`init_site_packages.sh` 脚本为隔离执行环境初始化 Python site packages：

```bash
./init_site_packages.sh
```

## 部署

### 使用 Helm

```bash
cd helm
helm install coderunner . -f values.yaml
```

### 环境变量

部署的关键环境变量：

**CodeRunner:**
- `PORT`: 服务器端口 (默认: 8080)
- `MAX_EXECUTION_TIME`: 最大代码执行时间（秒）
- `ALLOW_PACKAGE_INSTALL`: 启用/禁用包安装

**DataFlow Tools:**
- `PORT`: 服务器端口 (默认: 8081)
- `DB_HOST`: 数据库主机
- `DB_PORT`: 数据库端口
- `REDIS_HOST`: Redis 主机
- `REDIS_PORT`: Redis 端口

## 安全注意事项

### 代码执行安全

1. **RestrictedPython**: 所有代码执行都使用 RestrictedPython 防止危险操作
2. **导入限制**: 只能导入白名单中的模块
3. **资源限制**: 执行时间和内存限制防止资源耗尽
4. **隔离环境**: 每次执行都在隔离的上下文中运行
5. **无文件系统访问**: 限制对文件系统的访问

### 最佳实践

- 始终在执行前验证用户输入
- 为代码执行设置适当的超时限制
- 监控资源使用并设置限制
- 定期更新依赖项以获取安全补丁
- 尽可能使用只读文件系统

## 开发指南

### 项目结构

两个模块都遵循六边形架构模式：

```
src/
├── driveradapters/     # 入站适配器 (HTTP 处理器)
├── drivenadapters/     # 出站适配器 (外部服务)
├── logics/             # 核心业务逻辑
├── common/             # 共享工具
├── errors/             # 错误定义
└── utils/              # 辅助函数
```

### 添加新功能

1. **添加 API 处理器**: 在 `driveradapters/` 中创建处理器
2. **实现逻辑**: 在 `logics/` 中添加业务逻辑
3. **添加依赖**: 更新 `requirements.txt`
4. **更新文档**: 记录新端点

### 测试

```bash
# 运行测试（如果可用）
pytest tests/

# 代码质量检查
pylint src/
black src/
```

## 架构详解

### 六边形架构

两个模块都实现了六边形架构（端口和适配器）：
- **驱动适配器**: 接收请求的 HTTP API 处理器
- **被驱动适配器**: 与外部服务的集成（数据库、消息队列等）
- **核心逻辑**: 独立于外部接口的业务逻辑

### RestrictedPython 集成

CodeRunner 模块使用 RestrictedPython 提供安全的代码执行：
- 使用受限的内置函数编译代码
- 防止访问危险函数
- 提供安全的执行环境
- 优雅地处理错误

### 文件处理流程

DataFlow Tools 实现了文件处理的工厂模式：
- **基础处理器**: 所有文件类型的通用接口
- **特定处理器**: Word、Excel、PDF、Markdown 处理器
- **模板支持**: 从模板生成文档
- **缓存**: 缓存处理结果以提高性能
