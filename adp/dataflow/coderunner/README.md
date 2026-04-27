# CodeRunner Service

CodeRunner is a Python-based code execution and data processing service that provides sandboxed code execution, data analysis, document processing, and OCR capabilities. It consists of two main modules: the CodeRunner module for secure code execution and the DataFlow Tools module for data manipulation and file operations.

[中文文档](README_zh.md)

## Core Architecture

CodeRunner follows a microservices architecture with two independent modules:

```
┌─────────────────────────────────────────────────────────┐
│                   CodeRunner Module                     │
│         (Sandboxed Code Execution Service)              │
│  - RestrictedPython code execution                      │
│  - Python package management                            │
│  - Secure execution environment                         │
│  - Health monitoring                                    │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                 DataFlow Tools Module                   │
│          (Data Processing & Analysis)                   │
│  - Data analysis and transformation                     │
│  - Document processing (Word, Excel, PDF)               │
│  - OCR text extraction                                  │
│  - File operations and management                       │
└─────────────────────────────────────────────────────────┘
```

### Directory Structure

- **`coderunner/`**: Sandboxed code execution module
  - `src/driveradapters/`: HTTP API handlers (runner, pypkg, health)
  - `src/drivenadapters/`: External service integrations
  - `src/logics/`: Core business logic
  - `src/utils/`: Utility functions
- **`dataflowtools/`**: Data processing and analysis module
  - `src/driveradapters/`: HTTP API handlers (data_analysis, file, ocr, tag, py_pkg)
  - `src/drivenadapters/`: External service integrations
  - `src/logics/`: File handlers and analysis logic
- **`deps/`**: Python package dependencies
- **`helm/`**: Kubernetes deployment Helm charts
- **`migrations/`**: Database migration scripts

## Key Features

### CodeRunner Module

#### 1. Sandboxed Code Execution
- **RestrictedPython**: Secure Python code execution with restricted capabilities
- **Isolated Environment**: Execute user code in a sandboxed environment
- **Timeout Control**: Configurable execution timeout limits
- **Resource Limits**: Memory and CPU usage constraints

#### 2. Python Package Management
- **Package Installation**: Install Python packages dynamically
- **Package Listing**: Query installed packages
- **Package Isolation**: Separate package environments per execution
- **Dependency Management**: Handle package dependencies automatically

#### 3. Security Features
- **Code Sandboxing**: Restricted access to system resources
- **Import Control**: Whitelist-based module imports
- **Execution Isolation**: Isolated execution contexts
- **Error Handling**: Safe error reporting without exposing system details

### DataFlow Tools Module

#### 1. Data Analysis
- **Data Transformation**: Process and transform data using pandas
- **Statistical Analysis**: Perform data analysis operations
- **Data Validation**: Validate data against schemas
- **Custom Operations**: Execute custom data processing logic

#### 2. Document Processing
- **Word Documents**: Read and write .doc and .docx files
- **Excel Files**: Process .xls and .xlsx spreadsheets
- **PDF Files**: Extract text and data from PDF documents
- **Template Processing**: Generate documents from templates

#### 3. OCR Capabilities
- **Text Extraction**: Extract text from images and scanned documents
- **Multi-format Support**: Support for various image formats
- **Barcode/QR Code**: Decode barcodes and QR codes
- **Document Scanning**: Process scanned documents

#### 4. File Operations
- **File Upload/Download**: Handle file transfers
- **Format Conversion**: Convert between different file formats
- **File Metadata**: Extract and manage file metadata
- **Batch Processing**: Process multiple files in batch

## Tech Stack

### CodeRunner Module
- **Language**: Python 3.9
- **Web Framework**: Tornado 6.5
- **Sandboxing**: RestrictedPython 8.0
- **Data Processing**: pandas 2.0.3, numpy 1.24.4
- **Document Libraries**: python-docx, openpyxl, PyMuPDF, pdfplumber
- **OCR**: opencv-python, pyzbar
- **HTTP Client**: httpx, requests

### DataFlow Tools Module
- **Language**: Python 3.9
- **Web Framework**: Tornado 6.4.2
- **Database**: PyMySQL, SQLAlchemy
- **Document Processing**: python-docx, openpyxl, pypandoc
- **Data Analysis**: pandas (via coderunner)
- **Caching**: Redis

## Prerequisites

- Python 3.9+
- Docker (for containerized deployment)
- Redis (for caching in dataflowtools)
- MySQL (for data storage in dataflowtools)

## Configuration

Both modules use environment variables for configuration (loaded from `.env` file):

### CodeRunner Module Configuration

```bash
# Server Configuration
PORT=8080
HOST=0.0.0.0

# Execution Limits
MAX_EXECUTION_TIME=30
MAX_MEMORY_MB=512

# Package Management
ALLOW_PACKAGE_INSTALL=true
PACKAGE_INDEX_URL=https://pypi.org/simple
```

### DataFlow Tools Module Configuration

```bash
# Server Configuration
PORT=8081
HOST=0.0.0.0

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_NAME=dataflow
DB_USER=root
DB_PASSWORD=password

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# External Services
OCR_SERVICE_URL=http://ocr-service:8080
```

## API Documentation

### CodeRunner Module Endpoints

#### Code Execution
- `POST /api/runner/execute` - Execute Python code in sandboxed environment
  - Request body: `{"code": "print('Hello')", "timeout": 30}`
  - Response: `{"result": "Hello\n", "status": "success"}`

#### Package Management
- `POST /api/pypkg/install` - Install Python package
- `GET /api/pypkg/list` - List installed packages
- `DELETE /api/pypkg/uninstall` - Uninstall package

#### Health Check
- `GET /api/health` - Service health status

### DataFlow Tools Module Endpoints

#### Data Analysis
- `POST /api/data_analysis/analyze` - Perform data analysis
- `POST /api/data_analysis/transform` - Transform data

#### File Operations
- `POST /api/file/upload` - Upload file
- `GET /api/file/download/{file_id}` - Download file
- `POST /api/file/convert` - Convert file format
- `DELETE /api/file/{file_id}` - Delete file

#### OCR
- `POST /api/ocr/extract` - Extract text from image
- `POST /api/ocr/barcode` - Decode barcode/QR code

#### Tag Management
- `POST /api/tag/create` - Create tag
- `GET /api/tag/list` - List tags
- `PUT /api/tag/update` - Update tag
- `DELETE /api/tag/delete` - Delete tag

#### Package Management
- `POST /api/py_pkg/install` - Install Python package
- `GET /api/py_pkg/list` - List packages

#### Health Check
- `GET /api/health` - Service health status

## Building and Running

### Local Development

#### CodeRunner Module

```bash
cd coderunner
pip install -r requirements.txt
python src/main.py
```

#### DataFlow Tools Module

```bash
cd dataflowtools
pip install -r requirements.txt
python src/main.py
```

### Docker Build

#### CodeRunner Module

```bash
docker build -t coderunner:latest -f Dockerfile.coderunner .
docker run -p 8080:8080 coderunner:latest
```

#### DataFlow Tools Module

```bash
docker build -t dataflowtools:latest -f Dockerfile.dataflowtools .
docker run -p 8081:8081 dataflowtools:latest
```

### Package Initialization

The `init_site_packages.sh` script initializes Python site packages for isolated execution environments:

```bash
./init_site_packages.sh
```

## Deployment

### Using Helm

```bash
cd helm
helm install coderunner . -f values.yaml
```

### Environment Variables

Key environment variables for deployment:

**CodeRunner:**
- `PORT`: Server port (default: 8080)
- `MAX_EXECUTION_TIME`: Maximum code execution time in seconds
- `ALLOW_PACKAGE_INSTALL`: Enable/disable package installation

**DataFlow Tools:**
- `PORT`: Server port (default: 8081)
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port

## Security Considerations

### Code Execution Security

1. **RestrictedPython**: All code execution uses RestrictedPython to prevent dangerous operations
2. **Import Restrictions**: Only whitelisted modules can be imported
3. **Resource Limits**: Execution time and memory limits prevent resource exhaustion
4. **Isolated Environments**: Each execution runs in an isolated context
5. **No File System Access**: Restricted access to the file system

### Best Practices

- Always validate user input before execution
- Set appropriate timeout limits for code execution
- Monitor resource usage and set limits
- Regularly update dependencies for security patches
- Use read-only file systems where possible

## Development Guide

### Project Structure

Both modules follow a hexagonal architecture pattern:

```
src/
├── driveradapters/     # Inbound adapters (HTTP handlers)
├── drivenadapters/     # Outbound adapters (external services)
├── logics/             # Core business logic
├── common/             # Shared utilities
├── errors/             # Error definitions
└── utils/              # Helper functions
```

### Adding New Features

1. **Add API Handler**: Create handler in `driveradapters/`
2. **Implement Logic**: Add business logic in `logics/`
3. **Add Dependencies**: Update `requirements.txt`
4. **Update Documentation**: Document new endpoints

### Testing

```bash
# Run tests (if available)
pytest tests/

# Code quality checks
pylint src/
black src/
```

## Architecture Details

### Hexagonal Architecture

Both modules implement hexagonal architecture (ports and adapters):
- **Driver Adapters**: HTTP API handlers that receive requests
- **Driven Adapters**: Integrations with external services (databases, message queues, etc.)
- **Core Logic**: Business logic independent of external interfaces

### RestrictedPython Integration

The CodeRunner module uses RestrictedPython to provide secure code execution:
- Compiles code with restricted builtins
- Prevents access to dangerous functions
- Provides safe execution environment
- Handles errors gracefully

### File Processing Pipeline

DataFlow Tools implements a factory pattern for file processing:
- **Base Handler**: Common interface for all file types
- **Specific Handlers**: Word, Excel, PDF, Markdown handlers
- **Template Support**: Generate documents from templates
- **Caching**: Cache processed results for performance
