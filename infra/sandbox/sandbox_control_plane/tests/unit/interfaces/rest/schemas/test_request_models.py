import pytest
from pydantic import ValidationError

from src.interfaces.rest.schemas.request import (
    CreateSessionRequest,
    CreateTemplateRequest,
    DependencySpec,
    ExecuteCodeRequest,
    InstallSessionDependenciesRequest,
    TerminateSessionRequest,
    UpdateTemplateRequest,
)


class TestDependencySpec:
    @pytest.mark.parametrize(
        "name",
        [
            "../requests",
            "/requests",
            "https://example.com/pkg.whl",
            "requests[security]",
            "pandas2.3.3",
        ],
    )
    def test_rejects_unsafe_or_mixed_package_names(self, name):
        with pytest.raises(ValidationError):
            DependencySpec(name=name)

    @pytest.mark.parametrize(
        ("version", "expected"),
        [
            (None, "requests"),
            ("2.32.0", "requests==2.32.0"),
            ("==2.32.0", "requests==2.32.0"),
            (">=2.0", "requests>=2.0"),
            ("~=2.0", "requests~=2.0"),
            ("!=2.1", "requests!=2.1"),
        ],
    )
    def test_to_pip_spec_formats_versions(self, version, expected):
        assert DependencySpec(name="requests", version=version).to_pip_spec() == expected


class TestCreateSessionRequest:
    def test_defaults_and_dependency_fields(self):
        request = CreateSessionRequest(
            dependencies=[DependencySpec(name="requests", version="2.32.0")]
        )

        assert request.timeout == 300
        assert request.cpu == "1"
        assert request.dependencies[0].to_pip_spec() == "requests==2.32.0"
        assert request.install_timeout == 300
        assert request.fail_on_dependency_error is True
        assert request.allow_version_conflicts is False

    @pytest.mark.parametrize("cpu", ["abc", "", "one"])
    def test_rejects_invalid_cpu(self, cpu):
        with pytest.raises(ValidationError):
            CreateSessionRequest(cpu=cpu)

    def test_accepts_custom_python_package_index_url(self):
        request = CreateSessionRequest(
            python_package_index_url="https://mirror.example.com/simple/"
        )

        assert request.python_package_index_url == "https://mirror.example.com/simple/"


class TestExecuteCodeRequest:
    @pytest.mark.parametrize(
        ("raw", "expected"),
        [
            ("tools/build", "tools/build"),
            ("./tools/build", "tools/build"),
            (" tools/build ", "tools/build"),
            ("skill/mini-wiki", "skill/mini-wiki"),
        ],
    )
    def test_normalizes_relative_working_directory(self, raw, expected):
        request = ExecuteCodeRequest(code="print(1)", language="python", working_directory=raw)

        assert request.working_directory == expected

    @pytest.mark.parametrize(
        "working_directory",
        ["", "   ", "/workspace", "../secret", "tools/../secret", r"C:\workspace", r"a\b"],
    )
    def test_rejects_unsafe_working_directory(self, working_directory):
        with pytest.raises(ValidationError):
            ExecuteCodeRequest(
                code="print(1)",
                language="python",
                working_directory=working_directory,
            )

    def test_schema_examples_are_declared(self):
        examples = ExecuteCodeRequest.model_config["json_schema_extra"]["examples"]

        assert any(example.get("working_directory") == "skill/mini-wiki" for example in examples)


class TestOtherRequestModels:
    def test_install_session_dependencies_requires_non_empty_dependency_list(self):
        with pytest.raises(ValidationError):
            InstallSessionDependenciesRequest(dependencies=[])

    def test_install_session_dependencies_accepts_valid_payload(self):
        request = InstallSessionDependenciesRequest(
            dependencies=[DependencySpec(name="numpy")],
            python_package_index_url="https://pypi.org/simple/",
            install_timeout=1800,
        )

        assert request.dependencies[0].to_pip_spec() == "numpy"
        assert request.install_timeout == 1800

    def test_terminate_session_reason_is_optional(self):
        assert TerminateSessionRequest().reason is None
        assert TerminateSessionRequest(reason="cleanup").reason == "cleanup"

    def test_create_template_request_accepts_supported_runtime(self):
        request = CreateTemplateRequest(
            id="tpl-1",
            name="Go",
            image_url="sandbox-go:latest",
            runtime_type="go1.21",
        )

        assert request.default_cpu_cores == 0.5
        assert request.default_memory_mb == 512
        assert request.default_disk_mb == 1024

    def test_create_template_request_rejects_unknown_runtime(self):
        with pytest.raises(ValidationError):
            CreateTemplateRequest(
                id="tpl-1",
                name="Ruby",
                image_url="sandbox-ruby:latest",
                runtime_type="ruby3.3",
            )

    def test_update_template_request_allows_partial_updates(self):
        request = UpdateTemplateRequest(name="Updated", default_cpu_cores=1.5)

        assert request.name == "Updated"
        assert request.image_url is None
        assert request.default_cpu_cores == 1.5
