"""
Docker 容器调度器单元测试

测试 DockerScheduler 类的功能。
"""
from unittest.mock import AsyncMock, Mock, patch

import pytest
from aiodocker.exceptions import DockerError

from src.infrastructure.container_scheduler.base import ContainerConfig
from src.infrastructure.container_scheduler.docker_scheduler import DockerScheduler


class TestDockerScheduler:
    """Docker 容器调度器测试"""

    @pytest.fixture
    def mock_docker(self):
        """模拟 Docker 客户端"""
        docker = Mock()
        # 版本信息同步调用
        docker.version.return_value = {"Version": "20.10.0"}
        images_mock = Mock()
        images_mock.inspect = AsyncMock(return_value={})
        images_mock.pull = AsyncMock()
        docker.images = images_mock
        return docker

    @pytest.fixture
    def scheduler(self, mock_docker):
        """创建 Docker 调度器"""
        sched = DockerScheduler()
        sched._docker = mock_docker
        sched._initialized = True
        return sched

    @pytest.fixture
    def basic_config(self):
        """基础容器配置"""
        return ContainerConfig(
            image="python:3.11",
            name="test-container",
            cpu_limit="1",
            memory_limit="512Mi",
            disk_limit="1Gi",
            env_vars={"TEST": "value"},
            labels={"test": "label"},
            network_name="sandbox_network",
            workspace_path="/workspace"
        )

    def test_parse_s3_workspace_valid(self, scheduler):
        """测试解析有效的 S3 workspace 路径"""
        result = scheduler._parse_s3_workspace("s3://my-bucket/sessions/sess_123/")

        assert result is not None
        assert result["bucket"] == "my-bucket"
        assert result["prefix"] == "sessions/sess_123/"

    def test_parse_s3_workspace_invalid(self, scheduler):
        """测试解析无效的 S3 workspace 路径"""
        result = scheduler._parse_s3_workspace("/local/path/workspace")

        assert result is None

    def test_parse_memory_to_bytes_gi(self, scheduler):
        """测试解析内存限制（Gi）"""
        result = scheduler._parse_memory_to_bytes("1Gi")

        assert result == 1024 * 1024 * 1024

    def test_parse_memory_to_bytes_mi(self, scheduler):
        """测试解析内存限制（Mi）"""
        result = scheduler._parse_memory_to_bytes("512Mi")

        assert result == 512 * 1024 * 1024

    def test_parse_memory_to_bytes_ki(self, scheduler):
        """测试解析内存限制（Ki）"""
        result = scheduler._parse_memory_to_bytes("256Ki")

        assert result == 256 * 1024

    def test_parse_memory_to_bytes_default(self, scheduler):
        """测试解析内存限制（默认单位为 MB）"""
        result = scheduler._parse_memory_to_bytes("1024")

        assert result == 1024 * 1024 * 1024

    def test_parse_disk_to_bytes(self, scheduler):
        """测试解析磁盘限制"""
        result = scheduler._parse_disk_to_bytes("10Gi")

        assert result == 10 * 1024 * 1024 * 1024

    @pytest.mark.asyncio
    async def test_create_container_basic(self, scheduler, mock_docker, basic_config):
        """测试创建基本容器"""
        mock_container = Mock()
        mock_container.id = "container-123"

        # 创建 containers mock
        containers_mock = Mock()
        containers_mock.create = AsyncMock(return_value=mock_container)
        mock_docker.containers = containers_mock

        container_id = await scheduler.create_container(basic_config)

        assert container_id == "container-123"
        containers_mock.create.assert_called_once()
        mock_docker.images.inspect.assert_awaited_once_with("python:3.11")
        mock_docker.images.pull.assert_not_called()

    @pytest.mark.asyncio
    async def test_create_container_pulls_missing_image(self, scheduler, mock_docker, basic_config):
        """测试本地镜像缺失时会先拉取远端镜像。"""
        mock_container = Mock()
        mock_container.id = "container-123"

        mock_docker.images.inspect = AsyncMock(side_effect=DockerError(404, "No such image"))
        mock_docker.images.pull = AsyncMock(return_value={})

        containers_mock = Mock()
        containers_mock.create = AsyncMock(return_value=mock_container)
        mock_docker.containers = containers_mock

        container_id = await scheduler.create_container(basic_config)

        assert container_id == "container-123"
        mock_docker.images.inspect.assert_awaited_once_with("python:3.11")
        mock_docker.images.pull.assert_awaited_once_with("python:3.11")
        containers_mock.create.assert_called_once()

    @pytest.mark.asyncio
    async def test_create_container_with_s3_workspace(self, scheduler, mock_docker):
        """测试创建带 S3 workspace 的容器"""
        config = ContainerConfig(
            image="python:3.11",
            name="test-container",
            cpu_limit="1",
            memory_limit="512Mi",
            disk_limit="1Gi",
            env_vars={},
            labels={},
            network_name="sandbox_network",
            workspace_path="s3://my-bucket/sessions/sess_123/"
        )

        mock_container = Mock()
        mock_container.id = "container-123"

        containers_mock = Mock()
        containers_mock.create = AsyncMock(return_value=mock_container)
        mock_docker.containers = containers_mock

        container_id = await scheduler.create_container(config)

        assert container_id == "container-123"

        # 验证容器配置包含 S3 相关配置
        call_args = containers_mock.create.call_args
        container_config = call_args[0][0]
        assert container_config["User"] == "root"  # S3 模式需要 root
        assert "SYS_ADMIN" in container_config["HostConfig"]["CapAdd"]

    @pytest.mark.asyncio
    async def test_start_container(self, scheduler, mock_docker):
        """测试启动容器"""
        mock_container = Mock()
        mock_container.start = AsyncMock()

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        await scheduler.start_container("container-123")

        containers_mock.container.assert_called_once_with("container-123")
        mock_container.start.assert_called_once()

    @pytest.mark.asyncio
    async def test_stop_container(self, scheduler, mock_docker):
        """测试停止容器"""
        mock_container = Mock()
        mock_container.stop = AsyncMock()

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        await scheduler.stop_container("container-123", timeout=10)

        mock_container.stop.assert_called_once_with(timeout=10)

    @pytest.mark.asyncio
    async def test_remove_container(self, scheduler, mock_docker):
        """测试删除容器"""
        mock_container = Mock()
        mock_container.delete = AsyncMock()

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        await scheduler.remove_container("container-123", force=True)

        mock_container.delete.assert_called_once_with(force=True)

    @pytest.mark.asyncio
    async def test_get_container_status_running(self, scheduler, mock_docker):
        """测试获取运行中容器状态"""
        mock_container = Mock()
        mock_container.show = AsyncMock(return_value={
            "Id": "container-123",
            "Name": "/test-container",
            "State": {
                "Status": "running",
                "Paused": False,
                "ExitCode": 0,
                "Running": True,
                "StartedAt": "2024-01-15T10:00:01Z",
                "FinishedAt": "0001-01-01T00:00:00Z"
            },
            "Config": {
                "Image": "python:3.11"
            },
            "NetworkSettings": {
                "IPAddress": "172.17.0.2"
            },
            "Created": "2024-01-15T10:00:00Z"
        })

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        status = await scheduler.get_container_status("container-123")

        assert status.id == "container-123"
        assert status.status == "running"
        assert status.ip_address == "172.17.0.2"

    @pytest.mark.asyncio
    async def test_is_container_running_true(self, scheduler, mock_docker):
        """测试检查容器是否运行中（运行中）"""
        mock_container = Mock()
        mock_container.show = AsyncMock(return_value={
            "Id": "container-123",
            "Name": "/test-container",
            "State": {
                "Status": "running",
                "Running": True,
                "Paused": False,
                "ExitCode": 0,
                "StartedAt": "2024-01-15T10:00:01Z",
                "FinishedAt": "0001-01-01T00:00:00Z"
            },
            "Config": {
                "Image": "python:3.11"
            },
            "NetworkSettings": {
                "IPAddress": "172.17.0.2"
            },
            "Created": "2024-01-15T10:00:00Z"
        })

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        is_running = await scheduler.is_container_running("container-123")

        assert is_running is True

    @pytest.mark.asyncio
    async def test_get_container_logs(self, scheduler, mock_docker):
        """测试获取容器日志"""
        mock_container = Mock()
        mock_container.log = AsyncMock(return_value=["log line 1\n", "log line 2\n"])

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        logs = await scheduler.get_container_logs("container-123", tail=100)

        assert logs == "log line 1\nlog line 2\n"
        mock_container.log.assert_called_once()

    @pytest.mark.asyncio
    async def test_wait_container_success(self, scheduler, mock_docker):
        """测试等待容器完成（成功）"""
        mock_container = Mock()
        mock_container.wait = AsyncMock(return_value={"StatusCode": 0})
        mock_container.log = AsyncMock(return_value=["output\n"])

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        result = await scheduler.wait_container("container-123")

        assert result.status == "completed"
        assert result.exit_code == 0
        assert result.stdout == "output\n"

    @pytest.mark.asyncio
    async def test_wait_container_timeout(self, scheduler, mock_docker):
        """测试等待容器完成（超时）"""
        mock_container = Mock()
        mock_container.wait = AsyncMock(side_effect=TimeoutError())

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        result = await scheduler.wait_container("container-123", timeout=5)

        assert result.status == "timeout"
        assert "timed out" in result.stderr.lower()

    @pytest.mark.asyncio
    async def test_ping_success(self, scheduler, mock_docker):
        """测试 ping 成功"""
        mock_docker.version = AsyncMock(return_value={"Version": "20.10.0"})
        result = await scheduler.ping()
        assert result is True

    @pytest.mark.asyncio
    async def test_close(self, scheduler, mock_docker):
        """测试关闭连接"""
        mock_docker.close = AsyncMock()

        await scheduler.close()

        mock_docker.close.assert_called_once()
        assert scheduler._initialized is False

    def test_build_s3_mount_entrypoint(self, scheduler):
        """测试构建 S3 挂载入口脚本"""
        script = scheduler._build_s3_mount_entrypoint(
            s3_bucket="test-bucket",
            s3_prefix="sessions/sess_123",
            s3_endpoint_url="http://localhost:9000",
            s3_access_key="minioadmin",
            s3_secret_key="minioadmin",
            dependencies=None
        )

        assert "s3fs" in script
        assert "test-bucket" in script
        assert "sessions/sess_123" in script
        assert 'mount --bind "$SESSION_PATH" /workspace' in script
        assert "Workspace bind mounted" in script
        assert "workspace-old" not in script
        assert "ln -s" not in script

    def test_build_s3_mount_entrypoint_with_dependencies(self, scheduler):
        """测试构建带依赖的 S3 挂载入口脚本"""
        dependencies = [{"name": "requests", "version": "==2.31.0"}]
        script = scheduler._build_s3_mount_entrypoint(
            s3_bucket="test-bucket",
            s3_prefix="sessions/sess_123",
            s3_endpoint_url="http://localhost:9000",
            s3_access_key="minioadmin",
            s3_secret_key="minioadmin",
            dependencies=dependencies
        )

        assert "pip3 install" in script
        assert "requests==2.31.0" in script
        assert 'mount --bind "$SESSION_PATH" /workspace' in script

    def test_build_dependency_install_entrypoint(self, scheduler):
        """测试构建依赖安装入口脚本"""
        dependencies = [{"name": "pandas", "version": ">=2.0"}]
        script = scheduler._build_dependency_install_entrypoint(dependencies)

        assert "pip3 install" in script
        assert "pandas>=2.0" in script

    def test_build_dependency_install_entrypoint_no_deps(self, scheduler):
        """测试构建无依赖的入口脚本"""
        script = scheduler._build_dependency_install_entrypoint(None)

        # 应不包含 pip install
        assert "pip3 install" not in script

    @pytest.mark.asyncio
    async def test_is_container_running_false(self, scheduler, mock_docker):
        """测试检查容器是否运行中（未运行）"""
        mock_container = Mock()
        mock_container.show = AsyncMock(return_value={
            "Id": "container-123",
            "Name": "/test-container",
            "State": {
                "Status": "exited",
                "Running": False,
                "Paused": False,
                "ExitCode": 0,
                "StartedAt": "2024-01-15T10:00:01Z",
                "FinishedAt": "2024-01-15T10:05:00Z"
            },
            "Config": {
                "Image": "python:3.11"
            },
            "NetworkSettings": {
                "IPAddress": ""
            },
            "Created": "2024-01-15T10:00:00Z"
        })

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        is_running = await scheduler.is_container_running("container-123")

        assert is_running is False

    @pytest.mark.asyncio
    async def test_get_container_status_stopped(self, scheduler, mock_docker):
        """测试获取已停止容器状态"""
        mock_container = Mock()
        mock_container.show = AsyncMock(return_value={
            "Id": "container-123",
            "Name": "/test-container",
            "State": {
                "Status": "exited",
                "Paused": False,
                "ExitCode": 1,
                "Running": False,
                "StartedAt": "2024-01-15T10:00:01Z",
                "FinishedAt": "2024-01-15T10:05:00Z"
            },
            "Config": {
                "Image": "python:3.11"
            },
            "NetworkSettings": {
                "IPAddress": ""
            },
            "Created": "2024-01-15T10:00:00Z"
        })

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        status = await scheduler.get_container_status("container-123")

        assert status.status == "exited"
        assert status.exit_code == 1

    @pytest.mark.asyncio
    async def test_wait_container_failure(self, scheduler, mock_docker):
        """测试等待容器完成（失败）"""
        mock_container = Mock()
        mock_container.wait = AsyncMock(return_value={"StatusCode": 1})
        mock_container.log = AsyncMock(return_value=["error output\n"])

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        result = await scheduler.wait_container("container-123")

        assert result.status == "failed"
        assert result.exit_code == 1

    @pytest.mark.asyncio
    async def test_ping_failure(self, scheduler, mock_docker):
        """测试 ping 失败"""
        mock_docker.version = AsyncMock(side_effect=Exception("Connection failed"))
        result = await scheduler.ping()
        assert result is False

    @pytest.mark.asyncio
    async def test_ensure_docker_initialization(self, mock_docker):
        """测试 Docker 客户端初始化"""
        scheduler = DockerScheduler()
        scheduler._initialized = False

        with patch('src.infrastructure.container_scheduler.docker_scheduler.Docker', return_value=mock_docker):
            mock_docker.version = AsyncMock(return_value={"Version": "20.10.0"})

            await scheduler._ensure_docker()

            assert scheduler._initialized is True

    @pytest.mark.asyncio
    async def test_ensure_docker_already_initialized(self, scheduler, mock_docker):
        """测试 Docker 客户端已初始化"""
        result = await scheduler._ensure_docker()

        assert result is mock_docker

    def test_parse_s3_workspace_empty(self, scheduler):
        """测试解析空 S3 workspace 路径"""
        result = scheduler._parse_s3_workspace("")

        assert result is None

    def test_parse_s3_workspace_none(self, scheduler):
        """测试解析 None S3 workspace 路径"""
        result = scheduler._parse_s3_workspace(None)

        assert result is None

    def test_parse_disk_to_bytes_mb(self, scheduler):
        """测试解析磁盘限制（MB）"""
        result = scheduler._parse_disk_to_bytes("512Mi")

        assert result == 512 * 1024 * 1024

    def test_parse_memory_to_bytes_gib(self, scheduler):
        """测试解析内存限制（GiB）"""
        result = scheduler._parse_memory_to_bytes("2Gi")

        assert result == 2 * 1024 * 1024 * 1024

    def test_parse_memory_to_bytes_plain_number(self, scheduler):
        """测试解析内存限制（纯数字）"""
        result = scheduler._parse_memory_to_bytes("2048")

        # 默认单位为 MB
        assert result == 2048 * 1024 * 1024

    @pytest.mark.asyncio
    async def test_create_container_with_labels(self, scheduler, mock_docker):
        """测试创建带标签的容器"""
        config = ContainerConfig(
            image="python:3.11",
            name="test-container",
            cpu_limit="1",
            memory_limit="512Mi",
            disk_limit="1Gi",
            env_vars={},
            labels={"session_id": "sess-123", "template_id": "python-test"},
            network_name="sandbox_network",
            workspace_path="/workspace"
        )

        mock_container = Mock()
        mock_container.id = "container-123"

        containers_mock = Mock()
        containers_mock.create = AsyncMock(return_value=mock_container)
        mock_docker.containers = containers_mock

        container_id = await scheduler.create_container(config)

        assert container_id == "container-123"

        # 验证标签被传递
        call_args = containers_mock.create.call_args
        container_config = call_args[0][0]
        assert container_config["Labels"]["session_id"] == "sess-123"
        assert container_config["Labels"]["template_id"] == "python-test"

    @pytest.mark.asyncio
    async def test_get_container_status_with_name(self, scheduler, mock_docker):
        """测试获取容器状态（带名称）"""
        mock_container = Mock()
        mock_container.show = AsyncMock(return_value={
            "Id": "container-123",
            "Name": "/sandbox-sess-123",
            "State": {
                "Status": "running",
                "Paused": False,
                "ExitCode": 0,
                "Running": True,
                "StartedAt": "2024-01-15T10:00:01Z",
                "FinishedAt": "0001-01-01T00:00:00Z"
            },
            "Config": {
                "Image": "python:3.11"
            },
            "NetworkSettings": {
                "IPAddress": "172.17.0.2"
            },
            "Created": "2024-01-15T10:00:00Z"
        })

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        status = await scheduler.get_container_status("container-123")

        assert status.name == "sandbox-sess-123"

    @pytest.mark.asyncio
    async def test_get_container_logs_with_stderr(self, scheduler, mock_docker):
        """测试获取容器日志（包含 stderr）"""
        mock_container = Mock()
        mock_container.log = AsyncMock(return_value=["stdout\n", "stderr\n"])

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        logs = await scheduler.get_container_logs("container-123", tail=100)

        assert "stdout" in logs
        assert "stderr" in logs

    @pytest.mark.asyncio
    async def test_stop_container_default_timeout(self, scheduler, mock_docker):
        """测试停止容器（默认超时）"""
        mock_container = Mock()
        mock_container.stop = AsyncMock()

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        await scheduler.stop_container("container-123")

        mock_container.stop.assert_called_once()

    @pytest.mark.asyncio
    async def test_remove_container_no_force(self, scheduler, mock_docker):
        """测试删除容器（不强制）"""
        mock_container = Mock()
        mock_container.delete = AsyncMock()

        containers_mock = Mock()
        containers_mock.container = Mock(return_value=mock_container)
        mock_docker.containers = containers_mock

        await scheduler.remove_container("container-123", force=False)

        mock_container.delete.assert_called_once_with(force=False)

    @pytest.mark.asyncio
    async def test_create_container_with_env_vars(self, scheduler, mock_docker):
        """测试创建带环境变量的容器"""
        config = ContainerConfig(
            image="python:3.11",
            name="test-container",
            cpu_limit="1",
            memory_limit="512Mi",
            disk_limit="1Gi",
            env_vars={"DEBUG": "true", "API_KEY": "secret"},
            labels={},
            network_name="sandbox_network",
            workspace_path="/workspace"
        )

        mock_container = Mock()
        mock_container.id = "container-123"

        containers_mock = Mock()
        containers_mock.create = AsyncMock(return_value=mock_container)
        mock_docker.containers = containers_mock

        container_id = await scheduler.create_container(config)

        assert container_id == "container-123"

        # 验证环境变量被传递
        call_args = containers_mock.create.call_args
        container_config = call_args[0][0]
        assert "DEBUG=true" in container_config["Env"]
        assert "API_KEY=secret" in container_config["Env"]
