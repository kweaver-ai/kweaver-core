"""默认 seed 数据测试。"""

from src.infrastructure.persistence.seed.default_data import (
    get_default_templates,
    get_project_version,
)


def test_default_template_images_follow_project_version(monkeypatch):
    """测试默认模板镜像 tag 跟随项目 VERSION。"""
    monkeypatch.delenv("DEFAULT_TEMPLATE_IMAGE", raising=False)
    monkeypatch.delenv("DEFAULT_MULTI_LANGUAGE_TEMPLATE_IMAGE", raising=False)
    monkeypatch.delenv("PROJECT_VERSION", raising=False)
    monkeypatch.delenv("TEMPLATE_IMAGE_TAG", raising=False)

    templates = get_default_templates()
    template_map = {template.f_id: template for template in templates}
    project_version = get_project_version()

    assert template_map["python-basic"].f_image_url == (
        f"sandbox-template-python-basic:{project_version}"
    )
    assert template_map["multi-language"].f_image_url == (
        f"sandbox-template-multi-language:{project_version}"
    )


def test_default_template_images_can_use_template_image_tag(monkeypatch):
    """测试可通过 TEMPLATE_IMAGE_TAG 覆盖默认模板镜像 tag。"""
    monkeypatch.delenv("DEFAULT_TEMPLATE_IMAGE", raising=False)
    monkeypatch.delenv("DEFAULT_MULTI_LANGUAGE_TEMPLATE_IMAGE", raising=False)
    monkeypatch.setenv("TEMPLATE_IMAGE_TAG", "9.9.9")

    templates = get_default_templates()
    template_map = {template.f_id: template for template in templates}

    assert template_map["python-basic"].f_image_url == "sandbox-template-python-basic:9.9.9"
    assert template_map["multi-language"].f_image_url == "sandbox-template-multi-language:9.9.9"


def test_default_template_images_can_be_overridden_independently(monkeypatch):
    """测试部署时可分别覆盖两个默认模板镜像。"""
    monkeypatch.setenv("DEFAULT_TEMPLATE_IMAGE", "registry.local/python-basic:dev")
    monkeypatch.setenv(
        "DEFAULT_MULTI_LANGUAGE_TEMPLATE_IMAGE",
        "registry.local/multi-language:dev",
    )

    templates = get_default_templates()
    template_map = {template.f_id: template for template in templates}

    assert template_map["python-basic"].f_image_url == "registry.local/python-basic:dev"
    assert template_map["multi-language"].f_image_url == "registry.local/multi-language:dev"


def test_default_templates_include_multi_language(monkeypatch):
    """测试默认模板包含多语言复合模板。"""
    monkeypatch.delenv("DEFAULT_TEMPLATE_IMAGE", raising=False)
    monkeypatch.setenv(
        "DEFAULT_MULTI_LANGUAGE_TEMPLATE_IMAGE",
        "custom-multi-language:latest",
    )

    templates = get_default_templates()
    template_map = {template.f_id: template for template in templates}

    assert "python-basic" in template_map
    assert "multi-language" in template_map
    assert template_map["multi-language"].f_image_url == "custom-multi-language:latest"
    assert template_map["multi-language"].f_runtime_type == "multi"
    assert template_map["multi-language"].f_is_active == 1
