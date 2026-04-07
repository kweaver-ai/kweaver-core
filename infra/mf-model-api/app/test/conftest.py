"""测试配置和公共fixture"""
import sys
import os
from unittest.mock import Mock, AsyncMock, MagicMock, patch

# 添加项目根目录到sys.path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../..')))

# Mock私有依赖模块（在导入项目代码之前）
# Mock tlogging
tlogging_mock = MagicMock()
tlogging_mock.SamplerLogger = MagicMock()
sys.modules['tlogging'] = tlogging_mock

# Mock rdsdriver
rdsdriver_mock = MagicMock()
sys.modules['rdsdriver'] = rdsdriver_mock

# Mock opentelemetry相关模块
opentelemetry_mock = MagicMock()
sys.modules['opentelemetry'] = opentelemetry_mock
sys.modules['opentelemetry.trace'] = MagicMock()
sys.modules['opentelemetry.propagate'] = MagicMock()
sys.modules['opentelemetry.sdk'] = MagicMock()
sys.modules['opentelemetry.sdk.trace'] = MagicMock()
sys.modules['opentelemetry.sdk.trace.export'] = MagicMock()
sys.modules['opentelemetry.sdk.resources'] = MagicMock()
sys.modules['opentelemetry.exporter'] = MagicMock()
sys.modules['opentelemetry.exporter.otlp'] = MagicMock()
sys.modules['opentelemetry.exporter.otlp.proto'] = MagicMock()
sys.modules['opentelemetry.exporter.otlp.proto.grpc'] = MagicMock()
sys.modules['opentelemetry.exporter.otlp.proto.grpc.trace_exporter'] = MagicMock()
sys.modules['opentelemetry.instrumentation'] = MagicMock()
sys.modules['opentelemetry.instrumentation.fastapi'] = MagicMock()

# Mock arrow (日期时间库)
arrow_mock = MagicMock()
arrow_mock.now = MagicMock(return_value=MagicMock())
arrow_mock.get = MagicMock(return_value=MagicMock())
sys.modules['arrow'] = arrow_mock

# Mock dbutilsx (私有数据库连接池)
dbutilsx_mock = MagicMock()
dbutilsx_mock.pooled_db = MagicMock()
dbutilsx_mock.pooled_db.PooledDB = MagicMock()
dbutilsx_mock.pooled_db.PooledDBInfo = MagicMock()
sys.modules['dbutilsx'] = dbutilsx_mock
sys.modules['dbutilsx.pooled_db'] = dbutilsx_mock.pooled_db

# Mock exporter (私有日志导出器)
exporter_mock = MagicMock()
sys.modules['exporter'] = exporter_mock
sys.modules['exporter.resource'] = MagicMock()
sys.modules['exporter.resource.resource'] = MagicMock()
sys.modules['exporter.ar_log'] = MagicMock()
sys.modules['exporter.ar_log.log_exporter'] = MagicMock()
sys.modules['exporter.ar_trace'] = MagicMock()
sys.modules['exporter.ar_trace.trace_exporter'] = MagicMock()
sys.modules['exporter.public'] = MagicMock()
sys.modules['exporter.public.client'] = MagicMock()
sys.modules['exporter.public.public'] = MagicMock()

# Mock llmadapter (私有LLM适配器)
llmadapter_mock = MagicMock()
sys.modules['llmadapter'] = llmadapter_mock
sys.modules['llmadapter.llms'] = MagicMock()
sys.modules['llmadapter.llms.llm_factory'] = MagicMock()
sys.modules['llmadapter.schema'] = MagicMock()
# Mock常用的Message类
llmadapter_mock.schema.HumanMessage = MagicMock
llmadapter_mock.schema.AIMessage = MagicMock

# Mock func_timeout
func_timeout_mock = MagicMock()
func_timeout_mock.func_timeout = MagicMock(side_effect=lambda timeout, func, args=(), kwargs=None: func(*args, **(kwargs or {})))
func_timeout_mock.FunctionTimedOut = Exception
sys.modules['func_timeout'] = func_timeout_mock

# Mock tiktoken (用于token计数)
tiktoken_mock = MagicMock()
tiktoken_mock.get_encoding = MagicMock(return_value=MagicMock())
# Mock encoding对象的encode方法
mock_encoding = MagicMock()
mock_encoding.encode = MagicMock(return_value=[1, 2, 3, 4, 5])  # 返回假的token列表
tiktoken_mock.get_encoding.return_value = mock_encoding
sys.modules['tiktoken'] = tiktoken_mock

# Mock confluent_kafka (Kafka客户端)
confluent_kafka_mock = MagicMock()
# Mock Producer
mock_producer = MagicMock()
mock_producer.produce = MagicMock()
mock_producer.poll = MagicMock()
mock_producer.flush = MagicMock()
confluent_kafka_mock.Producer = MagicMock(return_value=mock_producer)
# Mock Consumer
mock_consumer = MagicMock()
mock_consumer.subscribe = MagicMock()
mock_consumer.poll = MagicMock(return_value=None)
mock_consumer.close = MagicMock()
confluent_kafka_mock.Consumer = MagicMock(return_value=mock_consumer)
# Mock Admin相关
confluent_kafka_admin_mock = MagicMock()
confluent_kafka_admin_mock.AdminClient = MagicMock()
confluent_kafka_admin_mock.NewTopic = MagicMock()
sys.modules['confluent_kafka'] = confluent_kafka_mock
sys.modules['confluent_kafka.admin'] = confluent_kafka_admin_mock

import pytest
import asyncio
from typing import Dict, Any


@pytest.fixture(scope="session")
def event_loop():
    """创建事件循环"""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
def mock_redis():
    """Mock Redis连接"""
    redis_mock = AsyncMock()
    redis_mock.get_str = AsyncMock(return_value=None)
    redis_mock.set_str = AsyncMock(return_value=True)
    redis_mock.delete_str = AsyncMock(return_value=True)
    return redis_mock


@pytest.fixture
def mock_db_connection():
    """Mock数据库连接"""
    connection = Mock()
    cursor = Mock()
    cursor.execute = Mock()
    cursor.fetchall = Mock(return_value=[])
    cursor.fetchone = Mock(return_value=None)
    return connection, cursor


@pytest.fixture
def mock_user_info():
    """Mock用户信息"""
    return "test_user_id", "zh", "user"


@pytest.fixture
def mock_llm_model_data():
    """Mock LLM模型数据"""
    return [{
        "f_model_id": "123456789012345678",
        "f_model_name": "test_model",
        "f_model_series": "openai",
        "f_model": "gpt-3.5-turbo",
        "f_model_config": '{"api_url": "http://test.com", "api_model": "gpt-3.5-turbo", "api_key": "test_key"}',
        "f_max_model_len": 4096,
        "f_model_parameters": 1000000,
        "f_model_type": "llm",
        "f_quota": False,
        "f_create_by": "user1",
        "f_update_by": "user2",
        "f_create_time": "2024-01-01 00:00:00",
        "f_update_time": "2024-01-02 00:00:00"
    }]


@pytest.fixture
def mock_small_model_data():
    """Mock 小模型数据"""
    return [{
        "f_model_id": "123456789012345679",
        "f_model_name": "test_embedding",
        "f_model_type": "embedding",
        "f_model_config": '{"api_url": "http://test.com", "api_model": "embedding-model", "api_key": "test_key"}',
        "f_adapter": False,
        "f_adapter_code": "",
        "f_create_by": "user1",
        "f_update_by": "user2",
        "f_create_time": "2024-01-01 00:00:00",
        "f_update_time": "2024-01-02 00:00:00"
    }]


@pytest.fixture
def mock_request():
    """Mock FastAPI Request对象"""
    request = Mock()
    request.headers = {
        "Authorization": "Bearer test_token",
        "x-account-id": "test_user",
        "x-account-type": "user",
        "x-func-module": "test"
    }
    request.url = Mock()
    request.url.path = "/api/v1/test"
    return request


@pytest.fixture
def mock_aiohttp_response():
    """Mock aiohttp响应"""
    response = AsyncMock()
    response.status = 200
    response.text = AsyncMock(return_value='{"active": true, "sub": "user123", "client_id": "user123"}')
    return response


@pytest.fixture(autouse=True)
def reset_mocks():
    """每个测试后重置所有Mock"""
    yield
    # 测试后清理工作

