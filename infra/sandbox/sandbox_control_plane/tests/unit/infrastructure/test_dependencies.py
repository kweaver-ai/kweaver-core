from unittest.mock import Mock

from src.infrastructure.dependencies import _create_scheduler_for_state_sync
from src.infrastructure.schedulers.k8s_scheduler_service import K8sSchedulerService


class TestStateSyncDependencyWiring:
    def test_create_scheduler_for_state_sync_uses_real_template_repo_in_k8s(
        self,
        monkeypatch,
    ):
        template_repo = Mock()
        container_scheduler = Mock()

        monkeypatch.setattr("src.infrastructure.dependencies.USE_MOCK_SCHEDULER", False)
        monkeypatch.setattr("src.infrastructure.dependencies.IS_IN_KUBERNETES", True)
        monkeypatch.setenv("POD_NAME", "sandbox-control-plane-test")
        monkeypatch.setenv("POD_UID", "sandbox-control-plane-uid")
        monkeypatch.setattr(
            "src.infrastructure.dependencies.get_settings",
            lambda: Mock(
                control_plane_url="http://sandbox-control-plane.test:8000",
                kubernetes_namespace="sandbox-system",
                disable_bwrap=True,
            ),
        )

        scheduler = _create_scheduler_for_state_sync(
            container_scheduler=container_scheduler,
            template_repo=template_repo,
        )

        assert isinstance(scheduler, K8sSchedulerService)
        assert scheduler._template_repo is template_repo
