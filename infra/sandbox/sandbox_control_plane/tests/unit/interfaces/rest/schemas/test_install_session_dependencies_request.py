"""
增量安装依赖请求模型测试。
"""

import pytest
from pydantic import ValidationError

from src.interfaces.rest.schemas.request import InstallSessionDependenciesRequest


def test_install_session_dependencies_request_default_install_timeout():
    request = InstallSessionDependenciesRequest(
        dependencies=[
            {
                "name": "requests",
                "version": "2.31.0",
            }
        ],
    )

    assert request.install_timeout == 300


@pytest.mark.parametrize("install_timeout", [30, 900, 1800])
def test_install_session_dependencies_request_accepts_valid_install_timeout(
    install_timeout,
):
    request = InstallSessionDependenciesRequest(
        dependencies=[
            {
                "name": "requests",
                "version": "2.31.0",
            }
        ],
        install_timeout=install_timeout,
    )

    assert request.install_timeout == install_timeout


@pytest.mark.parametrize("install_timeout", [29, 1801])
def test_install_session_dependencies_request_rejects_invalid_install_timeout(
    install_timeout,
):
    with pytest.raises(ValidationError):
        InstallSessionDependenciesRequest(
            dependencies=[
                {
                    "name": "requests",
                    "version": "2.31.0",
                }
            ],
            install_timeout=install_timeout,
        )
