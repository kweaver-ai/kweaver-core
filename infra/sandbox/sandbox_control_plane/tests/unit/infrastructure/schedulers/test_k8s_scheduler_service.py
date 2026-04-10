"""
Kubernetes 调度服务单元测试

测试 K8sSchedulerService 的功能。
"""
import pytest
from unittest.mock import Mock, AsyncMock, patch

from src.infrastructure.schedulers.k8s_scheduler_service import K8sSchedulerService
from src.domain.services.scheduler import RuntimeNode, ScheduleRequest
from src.domain.value_objects.resource_limit import ResourceLimit
from tests.helpers import create_mock_template


class TestK8sSchedulerService:
    """Kubernetes 调度服务测试"""

    @pytest.fixture
    def container_scheduler(self):
        """模拟容器调度器"""
        scheduler = Mock()
        scheduler.create_container = AsyncMock(return_value="sandbox-sess-123")
        scheduler.start_container = AsyncMock()
        scheduler.stop_container = AsyncMock()
        scheduler.remove_container = AsyncMock()
        scheduler.get_container_status = AsyncMock()
        scheduler._namespace = "sandbox-runtime"
        scheduler._core_v1 = Mock()
        return scheduler

    @pytest.fixture
    def template_repo(self):
        """模拟模板仓储"""
        repo = Mock()
        repo.find_by_id = AsyncMock()
        return repo

    @pytest.fixture
    def executor_client(self):
        """模拟执行器客户端"""
        client = Mock()
        client.submit_execution = AsyncMock(return_value="exec-123")
        client.health_check = AsyncMock()
        return client

    @pytest.fixture
    def service(self, container_scheduler, template_repo, executor_client):
        """创建 K8s 调度服务"""
        return K8sSchedulerService(
            container_scheduler=container_scheduler,
            template_repo=template_repo,
            executor_client=executor_client,
            executor_port=8080,
            control_plane_url="http://sandbox-control-plane.sandbox-system.svc.cluster.local:8000",
            disable_bwrap=True,
        )

    @pytest.fixture
    def schedule_request(self):
        """创建调度请求"""
        return ScheduleRequest(
            session_id="sess-123",
            template_id="python-test",
            resource_limit=ResourceLimit.default(),
        )

    @pytest.fixture
    def template(self):
        """创建模板"""
        return create_mock_template(template_id="python-test")

    @pytest.mark.asyncio
    async def test_schedule_returns_cluster_node(self, service, schedule_request):
        """测试调度返回集群节点"""
        result = await service.schedule(schedule_request)

        assert result is not None
        assert result.id == "k8s-cluster"
        assert result.type == "kubernetes"

    @pytest.mark.asyncio
    async def test_get_node_cluster_node(self, service):
        """测试获取集群节点"""
        result = await service.get_node("k8s-cluster")

        assert result is not None
        assert result.id == "k8s-cluster"

    @pytest.mark.asyncio
    async def test_get_node_not_found(self, service):
        """测试获取不存在的节点"""
        result = await service.get_node("non-existent")

        assert result is None

    @pytest.mark.asyncio
    async def test_get_healthy_nodes(self, service):
        """测试获取健康节点列表"""
        result = await service.get_healthy_nodes()

        # K8s always returns the single cluster node
        assert len(result) == 1
        assert result[0].id == "k8s-cluster"

    @pytest.mark.asyncio
    async def test_mark_node_unhealthy(self, service):
        """测试标记节点为不健康（K8s 环境下不执行操作）"""
        # Should not raise error
        await service.mark_node_unhealthy("any-node")

    @pytest.mark.asyncio
    async def test_create_container_for_session_success(
        self, service, container_scheduler, template_repo, template
    ):
        """测试成功创建 Pod"""
        template_repo.find_by_id.return_value = template

        result = await service.create_container_for_session(
            session_id="sess-123",
            template_id="python-test",
            image="python:3.11",
            resource_limit=ResourceLimit.default(),
            env_vars={"TEST": "value"},
            workspace_path="s3://bucket/sessions/sess-123",
            node_id="k8s-cluster",
        )

        assert result == "sandbox-sess-123"
        container_scheduler.create_container.assert_called_once()

    @pytest.mark.asyncio
    async def test_create_container_template_not_found(
        self, service, template_repo
    ):
        """测试模板不存在时创建 Pod 失败"""
        template_repo.find_by_id.return_value = None

        with pytest.raises(RuntimeError, match="Template not found"):
            await service.create_container_for_session(
                session_id="sess-123",
                template_id="non-existent",
                image="python:3.11",
                resource_limit=ResourceLimit.default(),
                env_vars={},
                workspace_path="s3://bucket/sessions/sess-123",
                node_id="k8s-cluster",
            )

    @pytest.mark.asyncio
    async def test_create_container_with_dependencies(
        self, service, container_scheduler, template_repo, template
    ):
        """测试创建带依赖的 Pod"""
        template_repo.find_by_id.return_value = template

        result = await service.create_container_for_session(
            session_id="sess-123",
            template_id="python-test",
            image="python:3.11",
            resource_limit=ResourceLimit.default(),
            env_vars={},
            workspace_path="s3://bucket/sessions/sess-123",
            node_id="k8s-cluster",
            dependencies=[{"name": "requests", "version": ">=2.28.0"}],
        )

        assert result == "sandbox-sess-123"

    @pytest.mark.asyncio
    async def test_create_container_with_error(
        self, service, container_scheduler, template_repo, template
    ):
        """测试创建 Pod 失败"""
        template_repo.find_by_id.return_value = template
        container_scheduler.create_container.side_effect = RuntimeError("Create failed")

        with pytest.raises(RuntimeError, match="Create failed"):
            await service.create_container_for_session(
                session_id="sess-123",
                template_id="python-test",
                image="python:3.11",
                resource_limit=ResourceLimit.default(),
                env_vars={},
                workspace_path="s3://bucket/sessions/sess-123",
                node_id="k8s-cluster",
            )

    @pytest.mark.asyncio
    async def test_destroy_container_success(self, service, container_scheduler):
        """测试成功销毁 Pod"""
        await service.destroy_container("sandbox-sess-123")

        container_scheduler.stop_container.assert_called_once()
        container_scheduler.remove_container.assert_called_once()

    @pytest.mark.asyncio
    async def test_destroy_container_with_error(self, service, container_scheduler):
        """测试销毁 Pod 时出错"""
        container_scheduler.stop_container.side_effect = RuntimeError("Stop failed")

        with pytest.raises(RuntimeError):
            await service.destroy_container("sandbox-sess-123")

    @pytest.mark.asyncio
    async def test_get_container_info(self, service, container_scheduler):
        """测试获取 Pod 信息"""
        container_info = Mock()
        container_scheduler.get_container_status.return_value = container_info

        result = await service.get_container_info("sandbox-sess-123")

        assert result is container_info
        container_scheduler.get_container_status.assert_called_once_with("sandbox-sess-123")

    @pytest.mark.asyncio
    async def test_execute_success(
        self, service, container_scheduler, executor_client
    ):
        """测试成功执行代码"""
        # Mock K8s API response
        mock_pod_info = Mock()
        mock_pod_info.status.pod_ip = "10.0.0.100"

        container_scheduler._core_v1.read_namespaced_pod = Mock(
            return_value=mock_pod_info
        )

        from src.domain.value_objects.execution_request import ExecutionRequest
        execution_request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={},
            working_directory="src/jobs",
        )

        # Patch asyncio.to_thread to make it synchronous in tests
        with patch('asyncio.to_thread', new_callable=AsyncMock) as mock_to_thread:
            mock_to_thread.return_value = mock_pod_info

            result = await service.execute(
                session_id="sess-123",
                container_id="sandbox-sess-123",
                execution_request=execution_request,
            )

            assert result == "exec-123"
            executor_client.submit_execution.assert_called_once()
            assert executor_client.submit_execution.call_args.kwargs["working_directory"] == "src/jobs"

    @pytest.mark.asyncio
    async def test_execute_pod_no_ip(
        self, service, container_scheduler, executor_client
    ):
        """测试 Pod 没有 IP 地址"""
        mock_pod_info = Mock()
        mock_pod_info.status.pod_ip = None

        from src.domain.value_objects.execution_request import ExecutionRequest
        execution_request = ExecutionRequest(
            code="print('hello')",
            language="python",
            event={},
            timeout=60,
            env_vars={},
        )

        with patch('asyncio.to_thread', new_callable=AsyncMock) as mock_to_thread:
            mock_to_thread.return_value = mock_pod_info

            with pytest.raises(RuntimeError, match="does not have an IP address"):
                await service.execute(
                    session_id="sess-123",
                    container_id="sandbox-sess-123",
                    execution_request=execution_request,
                )

    def test_cluster_node_properties(self, service):
        """测试集群节点属性"""
        assert service._cluster_node.id == "k8s-cluster"
        assert service._cluster_node.type == "kubernetes"
        assert service._cluster_node.status == "healthy"
        assert service._cluster_node.max_sessions == 1000

    def test_default_disable_bwrap(self, container_scheduler, template_repo, executor_client):
        """测试默认禁用 bwrap"""
        service = K8sSchedulerService(
            container_scheduler=container_scheduler,
            template_repo=template_repo,
            executor_client=executor_client,
        )

        assert service._disable_bwrap is True

    def test_custom_disable_bwrap(self, container_scheduler, template_repo, executor_client):
        """测试自定义禁用 bwrap"""
        service = K8sSchedulerService(
            container_scheduler=container_scheduler,
            template_repo=template_repo,
            executor_client=executor_client,
            disable_bwrap=False,
        )

        assert service._disable_bwrap is False
