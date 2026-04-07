# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an Agent AT (Acceptance Testing) project using pytest framework for API, functional, and performance testing of data-agent and data-operator-hub systems. The project follows Chinese development practices and is designed for enterprise-level testing in a Kubernetes environment.

## Common Development Commands

### Running Tests
```bash
# Run all tests
python3 -m pytest

# Run specific test categories (NEW STRUCTURE)
python3 -m pytest ./testcases/data-agent/api/
python3 -m pytest ./testcases/data-agent/functional/
python3 -m pytest ./testcases/data-operator-hub/api/
python3 -m pytest ./testcases/data-operator-hub/functional/
python3 -m pytest ./testcases/data-flow/api/
python3 -m pytest ./testcases/data-flow/functional/

# Run specific test modules (NEW PATHS)
python3 -m pytest ./testcases/data-agent/api/test_agent_factory_v3.py
python3 -m pytest ./testcases/data-agent/functional/test_agent_chat.py
python3 -m pytest ./testcases/data-operator-hub/api/category/test_create_category.py

# Run all tests for a specific service
python3 -m pytest ./testcases/data-agent/
python3 -m pytest ./testcases/data-operator-hub/
python3 -m pytest ./testcases/data-flow/

# Run with Allure reporting
python3 -m pytest ./testcases/data-operator-hub --alluredir ./report/xml --clean-alluredir
allure generate ./report/xml -o ./report/html --clean

# Run single test with markers
python3 -m pytest -m smoke
python3 -m pytest -m api
python3 -m pytest -m slow
```

### Environment Setup
```bash
# Install dependencies
pip install -r requirements/requirements.txt

# Run from main script
python3 main.py

# Local development with Docker
docker run -it --name agent-at --net=host \
  -v /root/agent-AT:/app \
  -v /root/.kube/config:/root/.kube/config \
  -v /usr/bin/kubectl:/usr/bin/kubectl \
  acr.aishu.cn/dip/agent-at:20250418 bash
```

### Configuration Management
- Environment configuration: `./config/env.ini`
- Kubernetes ConfigMap: `./cm.yaml` (managed automatically)
- Test configuration follows enterprise standards with unified parameter naming

### Pipeline and Scripts
```bash
# Execute automated tests with pipeline
./scripts/operator-run.sh  # Exposes DB ports, labels nodes, runs helm deployment

# Pipeline parameters include:
- host: Server SSH IP (default: 192.168.232.15)
- password: Server SSH password (default: eisoo.com123)
- testdir: Test case directory to run (default: testcases/data-operator-hub/api - UPDATED)
```

### Additional Pytest Configuration
The `pytest.ini` includes:
- Chinese test markers: `smoke` (冒烟测试), `slow` (耗时测试), `api` (API 测试)
- Default arguments: `-v -s --tb=short`
- Real-time logging enabled with INFO level

## Architecture and Structure

### Core Components

**Test Organization:**
- **NEW STRUCTURE**: `testcases/[service]/[test-type]/test_xxx.py` (Updated 2025-12-17)
- `testcases/data-agent/` - Data agent testing
  - `api/` - Agent API tests (factory V3, app interfaces)
  - `functional/` - End-to-end agent functional tests
  - `performance/` - Agent performance tests
- `testcases/data-operator-hub/` - Operator hub testing
  - `api/` - Operator hub API tests (CRUD, MCP, categories, permissions)
  - `functional/` - End-to-end operator functional tests
  - `performance/` - Operator performance tests
- `testcases/data-flow/` - Data flow testing
  - `api/` - Data flow API tests
  - `functional/` - End-to-end data flow functional tests
  - `performance/` - Data flow performance tests

*Previous structure: `testcases/[api|functional|performance]/[service]/test_xxx.py` (deprecated)*

**Library Architecture (`lib/`):**
- `data-agent/` - Data agent specific clients
  - `agent_factory.py` - Agent Factory V3 API client (40+ methods)
  - `agent_app.py` - Agent App external API client (chat, debug, file operations)
- `operator.py` - Operator management APIs
- `operator_internal.py` - Internal operator APIs
- `mcp.py` - MCP (Model Context Protocol) handling
- `mcp_internal.py` - Internal MCP APIs
- `impex.py` - Import/export functionality
- `permission.py` - Permission management
- `tool_box.py` - Tool management utilities
- `relations.py` - Relationship management
- `dataflow_como_operator.py` - Data flow COMO operator integration

**Common Utilities (`common/`):**
- `get_token.py` - Authentication token management
- `create_user.py` - Test user provisioning
- `delete_user.py` - Test user cleanup
- `request.py` - Common HTTP request patterns
- `assert_tools.py` - Custom assertion utilities

### Session-level Fixtures

The `conftest.py` provides critical session-level fixtures:
- `APrepare` - Creates test organization, department, and user
- `Headers` - Bearer token authentication for external APIs
- `UserHeaders` - Internal API authentication
- `RoleMember` - Manages AI administrator role assignments
- `ModifyCM` - Configures Kubernetes ConfigMap for testing
- `AgentImport` - Imports agents from `data/data-agent/import/` directory with dynamic configuration
- `ModelCheck` - Validates model existence and connectivity for testing

### Configuration Architecture

**Environment Configuration (`config/env.ini`):**
```ini
[server]
host = 192.168.232.11
port = 443
db_port = 3330
db_user = anyshare
db_pwd = eisoo.com123

[requests]
protocol = https

[hydra]
svc_name = hydra-admin
svc_port = 4445
```

**Agent Configuration (`data/data-agent/cfg.json`):**
```json
[
    {
        "agent_config": "ttft_simple_chat_explore.json",
        "need_conf_llm": true,
        "need_conf_doc_resource": false,
        "need_conf_graph_resource": false,
        "test_queries": {
            "single_turn": ["你好，请介绍一下你自己"],
            "multi_turn": []
        }
    }
]
```

**Agent Factory V3 Test Data (`data/data-agent/test_agent_factory_v3_data.json`):**
```json
{
  "agent_config": {
    "name": "测试智能体",
    "profile": "这是一个用于测试的智能体",
    "key": "test-agent-{timestamp}",
    "avatar_type": 1,
    "product_key": "dip",
    "config": {
      "input": {"fields": [{"name": "query", "type": "string"}]},
      "system_prompt": "你是一个测试助手",
      "llms": [{
        "is_default": true,
        "llm_config": {
          "id": "1901307417048780800",
          "name": "aaron-model-dv3",
          "temperature": 1
        }
      }]
    }
  }
}
```

**Kubernetes Integration:**
- Automatic ConfigMap management for operator settings
- Pod restart handling during configuration changes
- Integration with Kubernetes API via mounted config

## Key Development Patterns

### API Testing Pattern
```python
@pytest.mark.api
def test_operator_registration(self, Headers):
    # Use the Headers fixture for authentication
    client = OperatorClient(host=host)
    result = client.RegisterOperator(data, Headers)
    assert result[0] == 201
```

### Common Utilities Pattern
Extract common API request patterns to reduce duplication:
```python
def _make_api_request(self, method, ip, token, endpoint, body=None, path_param=None):
    # Common request handling with Allure reporting
```

### Library Client Pattern
All API clients follow a consistent pattern using configuration from `env.ini` and common request utilities:
```python
class Operator():
    def __init__(self):
        file = GetContent("./config/env.ini")
        self.config = file.config()
        self.base_url = self.config["requests"]["protocol"] + "://" + self.config["server"]["host"] + ":" + self.config["server"]["port"] + "/api/agent-operator-integration/v1/operator"
```

### Data-Driven Agent Testing Pattern
The agent testing framework uses a data-driven approach with configuration-based test queries:
```python
# Test configuration in cfg.json
{
    "test_queries": {
        "single_turn": ["你好，请介绍一下你自己", "你能做什么？"],
        "multi_turn": ["什么是机器学习？", "详细说明", "举个例子"]
    }
}

# Test methods iterate through configured queries
for query in test_queries["single_turn"]:
    result = agent.ChatCompletion(agent_id, {"query": query}, headers)
    validate_response(result)
```

### Test Data Management
- `data/` directory for structured test data (JSON, YAML)
- `data/data-agent/import/` - Agent configuration files for testing
- `data/data-agent/cfg.json` - Agent test configuration with query datasets
- `resource/` directory for document resources (managed via FTP)
- Test data follows enterprise naming conventions

### Branch Management Strategy
- Development branches: `feature/[需求号]`
- Main branch: `MISSION`
- Release branches: `release/DIP-[大版本号]`

## Environment and Deployment

### Testing Environment
- Uses Docker image: `acr.aishu.cn/dip/agent-at:20250418`
- Contains: Python 3.9, pytest, allure, JDK8, thriftAPI
- Kubernetes integration with cluster access

### Dependencies
Key dependencies include:
- pytest==6.2.3 for test framework
- allure-pytest==2.8.40 for reporting
- requests==2.31.0 for HTTP client
- thrift==0.20.0 for RPC communication
- paramiko==3.4.0 for SSH operations
- jsonschema==4.21.1 for schema validation
- M2Crypto==0.40.0 for cryptographic operations
- Company internal PyPI repository: `http://repository.aishu.cn:8081/repository/pypi/simple`

### Docker Environment
- Official testing image: `acr.aishu.cn/dip/agent-at:20250418`
- Contains Python 3.9, pytest, allure, JDK8, and thriftAPI
- Kubernetes integration with mounted config and kubectl access
- DNS configuration required for service resolution within cluster

## Testing Standards

### Code Quality
- Follow PEP 8 naming conventions
- Test classes use PascalCase: `TestOperatorRegistration`
- Test methods use camelCase: `testOperatorRegistrationSuccess`
- Comprehensive Allure step reporting with `@allure.title` decorators
- Session-scoped fixtures for efficient resource management
- Allure reporting integrated with detailed test step documentation

### Test Data and Resource Management
- Structured test data in `data/` directory (JSON, YAML files)
- Document resources in `resource/` directory (PDF, DOC, ZIP files)
- Resource files managed via FTP - not committed to repository
- Test data follows enterprise naming conventions and versioning
- Use `GetContent` utility class for loading test data files

### Authentication and Authorization
- Two authentication token types: external API (Bearer token) and internal API (x-user header)
- Session-level fixtures handle token management automatically
- AI administrator role assignment for comprehensive testing
- Automatic user/organization creation and cleanup in session fixtures

### CI/CD Integration
- YAML pipeline configurations in `pipeline/` directory
- Unified parameter structure for pipeline inheritance
- Support for FTP-based resource management
- Automatic test report generation with Allure
- Helm deployment for Kubernetes integration
- Script-based database port exposure and node labeling

### Development Workflow
1. **Local Development**: Use Docker container with volume mounts for code and k8s config
2. **Testing**: Run tests with pytest markers for different test categories
3. **Reporting**: Generate Allure reports for test execution analysis
4. **Integration**: Kubernetes ConfigMap management for test configuration
5. **Cleanup**: Automatic resource cleanup via session fixtures

### Security Considerations
- Authentication via Bearer tokens
- User management with proper cleanup
- Permission-based access control testing
- Secure credential handling via environment configuration
- Role-based testing with AI administrator permissions

## Current Development Status

- **Active branch**: `feature/5004`
- **Major Update 2025-12-17**: Directory structure reorganization completed
  - **NEW STRUCTURE**: `testcases/[service]/[test-type]/test_xxx.py`
  - **Previous**: `testcases/[api|functional|performance]/[service]/test_xxx.py`
  - **Verification**: All test suites validated and working correctly after restructuring

- **Recent focus**: Enhanced testing framework with new agent-factory and agent-app modules
- **Key updates**:
  - **Directory Structure Reorganization**: Service-first organization for better test management
  - **Agent Factory V3 Testing**: Comprehensive test suite for agent-factory/v3 API endpoints
  - **Agent App External APIs**: Full coverage of agent-app external interfaces
  - **Enhanced Data Management**: Structured test data with JSON configuration files
  - **Improved Logging**: Detailed request/response logging with emoji indicators
  - **Better Error Handling**: Comprehensive error messages and skip conditions
  - **Multi-turn Conversation Testing**: Advanced agent conversation validation

### New Test Modules

**Agent Factory V3 Tests** (`testcases/data-agent/api/test_agent_factory_v3.py`):
- Agent CRUD operations (Create, Read, Update, Delete)
- Agent copying and template management
- Publishing and lifecycle management
- Permission and access control testing
- Product management APIs
- Personal space and recent visit features
- ✅ **Verified**: 28 test cases collected and passing

**Agent App Tests** (`testcases/data-agent/api/test_agent_app.py`):
- Agent application details retrieval
- Chat completion and conversation management
- Debug and API chat interfaces
- File checking and document processing
- Resume and terminate chat functionality

**Agent Chat Functional Tests** (`testcases/data-agent/functional/test_agent_chat.py`):
- End-to-end agent conversation testing
- Multi-turn dialogue validation
- Agent configuration and data source testing
- ✅ **Verified**: 2 test cases including complex multi-turn conversations

**Data Operator Hub Tests** (`testcases/data-operator-hub/api/category/test_create_category.py`):
- Category CRUD operations
- Parameterized testing for different category types
- ✅ **Verified**: 48 test cases collected and passing

### New Library Components

**Agent Factory Client** (`lib/data_agent/agent_factory.py`):
- 40+ API methods for complete agent lifecycle management
- Dynamic skill and data source configuration
- Model connectivity testing and validation
- Advanced publishing and permission management

**Agent App Client** (`lib/data_agent/agent_app.py`):
- Full chat completion support with timeout handling
- Conversation management (resume, terminate)
- Debug interface for development
- File checking and API documentation

### Enhanced Features

**Configuration-Driven Testing**:
- `test_agent_factory_v3_data.json`: Structured test data for factory operations
- Dynamic timestamp replacement for unique test runs
- Comprehensive request/response validation

**Error Reporting and Debugging**:
- Detailed error messages with context
- Allure integration for test reporting
- Intelligent server availability checking
- Graceful handling of missing test data

**Data Source and Skill Management**:
- Dynamic tool ID resolution for search tools
- Configurable agent skills (zhipu_search_tool, Plan_Agent, Summary_Agent)
- Document and knowledge graph resource configuration
- Memory, related question, and plan mode features

- Configuration updates: Environment settings and test data management
- Key areas: Data-driven testing, agent conversation functionality, multi-turn dialogue testing