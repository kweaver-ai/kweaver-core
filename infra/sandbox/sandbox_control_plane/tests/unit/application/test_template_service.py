"""
模板应用服务单元测试

测试 TemplateService 的用例编排逻辑。
"""
import pytest
from unittest.mock import Mock, AsyncMock

from src.application.services.template_service import TemplateService
from src.application.commands.create_template import CreateTemplateCommand
from src.application.commands.update_template import UpdateTemplateCommand
from src.application.queries.get_template import GetTemplateQuery
from src.domain.entities.template import Template
from src.domain.repositories.template_repository import ITemplateRepository
from src.shared.errors.domain import NotFoundError, ValidationError
from tests.helpers import create_mock_template


class TestTemplateService:
    """模板应用服务测试"""

    @pytest.fixture
    def template_repo(self):
        """模拟模板仓储"""
        repo = Mock()
        repo.save = AsyncMock()
        repo.find_by_id = AsyncMock()
        repo.find_by_name = AsyncMock()
        repo.find_all = AsyncMock()
        repo.delete = AsyncMock()
        return repo

    @pytest.fixture
    def service(self, template_repo):
        """创建模板服务"""
        return TemplateService(template_repo=template_repo)

    @pytest.mark.asyncio
    async def test_create_template_success(self, service, template_repo):
        """测试成功创建模板"""
        template_repo.find_by_id.return_value = None
        template_repo.find_by_name.return_value = None

        command = CreateTemplateCommand(
            template_id="python-datascience",
            name="Python Data Science",
            image_url="python:3.11-datascience",
            runtime_type="docker",
            default_cpu_cores=1,
            default_memory_mb=512,
            default_disk_mb=1024,
            default_timeout_sec=300
        )

        result = await service.create_template(command)

        assert result.id == "python-datascience"
        assert result.name == "Python Data Science"
        template_repo.save.assert_called_once()

    @pytest.mark.asyncio
    async def test_create_template_duplicate_name(self, service, template_repo):
        """测试创建重名模板"""
        existing_template = create_mock_template(
            template_id="existing-id",
            name="Python Data Science"
        )
        template_repo.find_by_id.return_value = None
        template_repo.find_by_name.return_value = existing_template

        command = CreateTemplateCommand(
            template_id="new-id",
            name="Python Data Science",
            image_url="python:3.11",
            runtime_type="docker",
            default_cpu_cores=1,
            default_memory_mb=512,
            default_disk_mb=1024,
            default_timeout_sec=300
        )

        with pytest.raises(ValidationError, match="Template name already exists"):
            await service.create_template(command)

    @pytest.mark.asyncio
    async def test_get_template_success(self, service, template_repo):
        """测试成功获取模板"""
        template = create_mock_template(
            template_id="python-datascience",
            name="Python Data Science",
            image="python:3.11-datascience"
        )
        template_repo.find_by_id.return_value = template

        query = GetTemplateQuery(template_id="python-datascience")
        result = await service.get_template(query)

        assert result.id == "python-datascience"
        assert result.name == "Python Data Science"

    @pytest.mark.asyncio
    async def test_get_template_not_found(self, service, template_repo):
        """测试获取不存在的模板"""
        template_repo.find_by_id.return_value = None

        query = GetTemplateQuery(template_id="non-existent")

        with pytest.raises(NotFoundError, match="Template not found"):
            await service.get_template(query)

    @pytest.mark.asyncio
    async def test_list_templates(self, service, template_repo):
        """测试列出所有模板"""
        templates = [
            create_mock_template("python-basic", "Python Basic", "python:3.11"),
            create_mock_template("python-datascience", "Python Data Science", "python:3.11-datascience"),
        ]
        template_repo.find_all.return_value = templates

        results = await service.list_templates(limit=50, offset=0)

        assert len(results) == 2
        assert results[0].id == "python-basic"
        assert results[1].id == "python-datascience"

    @pytest.mark.asyncio
    async def test_list_templates_with_pagination(self, service, template_repo):
        """测试分页列出模板"""
        all_templates = [
            create_mock_template(f"template-{i}", f"Template {i}", "python:3.11")
            for i in range(5, 10)
        ]
        template_repo.find_all.return_value = all_templates

        results = await service.list_templates(limit=5, offset=5)

        assert len(results) == 5
        assert results[0].id == "template-5"

    @pytest.mark.asyncio
    async def test_update_template_success(self, service, template_repo):
        """测试成功更新模板"""
        template = create_mock_template("python-basic", "Python Basic", "python:3.11")
        template_repo.find_by_id.return_value = template
        template_repo.find_by_name.return_value = None

        command = UpdateTemplateCommand(
            template_id="python-basic",
            image_url="python:3.12"
        )

        result = await service.update_template(command)

        assert result.id == "python-basic"
        assert template.image == "python:3.12"
        template_repo.save.assert_called_once()

    @pytest.mark.asyncio
    async def test_update_template_not_found(self, service, template_repo):
        """测试更新不存在的模板"""
        template_repo.find_by_id.return_value = None

        command = UpdateTemplateCommand(
            template_id="non-existent",
            image_url="python:3.12"
        )

        with pytest.raises(NotFoundError, match="Template not found"):
            await service.update_template(command)

    @pytest.mark.asyncio
    async def test_update_template_duplicate_name(self, service, template_repo):
        """测试更新模板为重名"""
        template = create_mock_template("template-1", "Template 1")
        other_template = create_mock_template("template-2", "Template 2")
        template_repo.find_by_id.return_value = template
        template_repo.find_by_name.return_value = other_template

        command = UpdateTemplateCommand(
            template_id="template-1",
            name="Template 2"
        )

        with pytest.raises(ValidationError, match="Template name already exists"):
            await service.update_template(command)

    @pytest.mark.asyncio
    async def test_update_template_resources(self, service, template_repo):
        """测试更新模板资源"""
        template = create_mock_template("python-basic", "Python Basic", "python:3.11")
        template_repo.find_by_id.return_value = template
        template_repo.find_by_name.return_value = None

        command = UpdateTemplateCommand(
            template_id="python-basic",
            default_cpu_cores=2,
            default_memory_mb=1024,
            default_disk_mb=2048
        )

        result = await service.update_template(command)

        assert template.default_resources.cpu == "2"
        assert template.default_resources.memory == "1024Mi"
        assert template.default_resources.disk == "2048Mi"

    @pytest.mark.asyncio
    async def test_delete_template_success(self, service, template_repo):
        """测试成功删除模板"""
        template = create_mock_template("python-basic", "Python Basic", "python:3.11")
        template_repo.find_by_id.return_value = template

        await service.delete_template("python-basic")

        template_repo.delete.assert_called_once_with("python-basic")

    @pytest.mark.asyncio
    async def test_delete_template_not_found(self, service, template_repo):
        """测试删除不存在的模板"""
        template_repo.find_by_id.return_value = None

        with pytest.raises(NotFoundError, match="Template not found"):
            await service.delete_template("non-existent")

    @pytest.mark.asyncio
    async def test_update_template_same_name(self, service, template_repo):
        """测试更新模板为相同名称（不应报错）"""
        template = create_mock_template("python-basic", "Python Basic", "python:3.11")
        template_repo.find_by_id.return_value = template
        template_repo.find_by_name.return_value = template

        command = UpdateTemplateCommand(
            template_id="python-basic",
            name="Python Basic"
        )

        result = await service.update_template(command)
        assert result.id == "python-basic"

    @pytest.mark.asyncio
    async def test_update_template_name_to_new_value(self, service, template_repo):
        """测试成功更新模板名称为新值"""
        template = create_mock_template("python-basic", "Python Basic", "python:3.11")
        template_repo.find_by_id.return_value = template
        template_repo.find_by_name.return_value = None  # No collision with new name

        new_name = "Python Advanced"
        command = UpdateTemplateCommand(
            template_id="python-basic",
            name=new_name
        )

        result = await service.update_template(command)

        # Verify name was updated on the template entity
        assert template.name == new_name
        assert result.name == new_name
        template_repo.save.assert_called_once()

    @pytest.mark.asyncio
    async def test_update_template_timeout_to_new_value(self, service, template_repo):
        """测试成功更新模板超时时间为新值"""
        template = create_mock_template("python-basic", "Python Basic", "python:3.11")
        template_repo.find_by_id.return_value = template

        new_timeout = 60
        command = UpdateTemplateCommand(
            template_id="python-basic",
            default_timeout_sec=new_timeout
        )

        result = await service.update_template(command)

        # Verify timeout was updated on the template entity
        assert template.default_timeout_sec == new_timeout
        assert result.default_timeout_sec == new_timeout
        template_repo.save.assert_called_once()

    @pytest.mark.asyncio
    async def test_update_template_timeout_to_zero(self, service, template_repo):
        """测试更新模板超时时间为0（边界情况）"""
        template = create_mock_template("python-basic", "Python Basic", "python:3.11")
        template_repo.find_by_id.return_value = template

        new_timeout = 0
        command = UpdateTemplateCommand(
            template_id="python-basic",
            default_timeout_sec=new_timeout
        )

        result = await service.update_template(command)

        # Verify timeout was updated to 0 (allowed)
        assert template.default_timeout_sec == 0
        assert result.default_timeout_sec == 0
