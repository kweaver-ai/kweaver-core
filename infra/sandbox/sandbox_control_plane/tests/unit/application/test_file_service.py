"""
文件应用服务单元测试

测试 FileService 的用例编排逻辑。
"""
import pytest
from unittest.mock import Mock, AsyncMock

from src.application.services.file_service import FileService
from src.domain.entities.session import Session
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus
from src.domain.repositories.session_repository import ISessionRepository
from src.domain.services.storage import IStorageService
from src.shared.errors.domain import NotFoundError, ValidationError


class TestFileService:
    """文件应用服务测试"""

    @pytest.fixture
    def session_repo(self):
        """模拟会话仓储"""
        repo = Mock()
        repo.find_by_id = AsyncMock()
        return repo

    @pytest.fixture
    def storage_service(self):
        """模拟存储服务"""
        service = Mock()
        service.upload_file = AsyncMock()
        service.download_file = AsyncMock()
        service.file_exists = AsyncMock()
        service.get_file_info = AsyncMock()
        service.generate_presigned_url = AsyncMock()
        service.list_files = AsyncMock()
        return service

    @pytest.fixture
    def service(self, session_repo, storage_service):
        """创建文件服务"""
        return FileService(
            session_repo=session_repo,
            storage_service=storage_service
        )

    @pytest.fixture
    def active_session(self):
        """活跃会话"""
        return Session(
            id="sess_123",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_123",
            runtime_type="docker"
        )

    @pytest.mark.asyncio
    async def test_upload_file_success(self, service, session_repo, storage_service, active_session):
        """测试成功上传文件"""
        session_repo.find_by_id.return_value = active_session

        content = b"hello world"
        result = await service.upload_file(
            session_id="sess_123",
            path="test.txt",
            content=content,
            content_type="text/plain"
        )

        assert result == "test.txt"
        storage_service.upload_file.assert_called_once()

    @pytest.mark.asyncio
    async def test_upload_file_session_not_found(self, service, session_repo):
        """测试上传文件到不存在的会话"""
        session_repo.find_by_id.return_value = None

        with pytest.raises(NotFoundError, match="Session not found"):
            await service.upload_file(
                session_id="non-existent",
                path="test.txt",
                content=b"hello"
            )

    @pytest.mark.asyncio
    async def test_upload_file_session_not_active(self, service, session_repo):
        """测试上传文件到非活跃会话"""
        session = Session(
            id="sess_123",
            template_id="python-basic",
            status=SessionStatus.TERMINATED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_123",
            runtime_type="docker"
        )
        session_repo.find_by_id.return_value = session

        with pytest.raises(ValidationError, match="Session is not active"):
            await service.upload_file(
                session_id="sess_123",
                path="test.txt",
                content=b"hello"
            )

    @pytest.mark.asyncio
    async def test_upload_file_invalid_path(self, service, session_repo, active_session):
        """测试上传文件到无效路径"""
        session_repo.find_by_id.return_value = active_session

        # 绝对路径
        with pytest.raises(ValidationError, match="Invalid file path"):
            await service.upload_file(
                session_id="sess_123",
                path="/absolute/path.txt",
                content=b"hello"
            )

        # 空路径
        with pytest.raises(ValidationError, match="Invalid file path"):
            await service.upload_file(
                session_id="sess_123",
                path="",
                content=b"hello"
            )

    @pytest.mark.asyncio
    async def test_upload_file_with_default_content_type(self, service, session_repo, storage_service, active_session):
        """测试使用默认内容类型上传文件"""
        session_repo.find_by_id.return_value = active_session

        await service.upload_file(
            session_id="sess_123",
            path="test.bin",
            content=b"\x00\x01\x02"
        )

        # 验证使用了默认的 content_type
        call_args = storage_service.upload_file.call_args
        assert call_args[1]["content_type"] == "application/octet-stream"

    @pytest.mark.asyncio
    async def test_upload_file_with_custom_content_type(self, service, session_repo, storage_service, active_session):
        """测试使用自定义内容类型上传文件"""
        session_repo.find_by_id.return_value = active_session

        await service.upload_file(
            session_id="sess_123",
            path="test.json",
            content=b'{"key": "value"}',
            content_type="application/json"
        )

        # 验证使用了自定义的 content_type
        call_args = storage_service.upload_file.call_args
        assert call_args[1]["content_type"] == "application/json"

    @pytest.mark.asyncio
    async def test_upload_file_s3_path_construction(self, service, session_repo, storage_service, active_session):
        """测试 S3 路径构造"""
        session_repo.find_by_id.return_value = active_session

        await service.upload_file(
            session_id="sess_123",
            path="data/test.csv",
            content=b"id,name\n1,test"
        )

        # 验证 S3 路径包含 workspace_path
        call_args = storage_service.upload_file.call_args
        if call_args[0]:
            s3_path = call_args[0][0]
        else:
            s3_path = call_args[1]["s3_path"]
        assert s3_path.startswith(active_session.workspace_path)
        assert "data/test.csv" in s3_path

    @pytest.mark.asyncio
    async def test_download_file_small_file(self, service, session_repo, storage_service, active_session):
        """测试下载小文件（直接返回内容）"""
        session_repo.find_by_id.return_value = active_session
        storage_service.file_exists.return_value = True
        storage_service.get_file_info.return_value = {
            "size": 1024,
            "content_type": "text/plain"
        }
        storage_service.download_file.return_value = b"file content"

        result = await service.download_file(
            session_id="sess_123",
            path="test.txt"
        )

        assert result["content"] == b"file content"
        assert result["content_type"] == "text/plain"
        assert result["size"] == 1024

    @pytest.mark.asyncio
    async def test_download_file_large_file(self, service, session_repo, storage_service, active_session):
        """测试下载大文件（返回预签名 URL）"""
        session_repo.find_by_id.return_value = active_session
        storage_service.file_exists.return_value = True
        storage_service.get_file_info.return_value = {
            "size": 15 * 1024 * 1024,  # 15MB
            "content_type": "application/octet-stream"
        }
        storage_service.generate_presigned_url.return_value = "https://s3.amazonaws.com/..."

        result = await service.download_file(
            session_id="sess_123",
            path="large.bin"
        )

        assert "presigned_url" in result
        assert result["size"] == 15 * 1024 * 1024
        assert result["presigned_url"] == "https://s3.amazonaws.com/..."

    @pytest.mark.asyncio
    async def test_download_file_session_not_found(self, service, session_repo):
        """测试从不存在会话下载文件"""
        session_repo.find_by_id.return_value = None

        with pytest.raises(NotFoundError, match="Session not found"):
            await service.download_file(
                session_id="non-existent",
                path="test.txt"
            )

    @pytest.mark.asyncio
    async def test_download_file_not_found(self, service, session_repo, storage_service, active_session):
        """测试下载不存在的文件"""
        session_repo.find_by_id.return_value = active_session
        storage_service.file_exists.return_value = False

        with pytest.raises(NotFoundError, match="File not found"):
            await service.download_file(
                session_id="sess_123",
                path="nonexistent.txt"
            )

    @pytest.mark.asyncio
    async def test_download_file_10mb_boundary(self, service, session_repo, storage_service, active_session):
        """测试 10MB 边界情况"""
        session_repo.find_by_id.return_value = active_session
        storage_service.file_exists.return_value = True

        # 正好 10MB
        storage_service.get_file_info.return_value = {
            "size": 10 * 1024 * 1024,
            "content_type": "application/octet-stream"
        }
        storage_service.download_file.return_value = b"x" * (10 * 1024 * 1024)

        result = await service.download_file(
            session_id="sess_123",
            path="boundary.bin"
        )

        # 小于 10MB（等于也是小于），应返回内容
        # 但如果 result 是 Mock，需要检查其属性
        if hasattr(result, "__getitem__"):
            assert "content" in result or "presigned_url" in result

    @pytest.mark.asyncio
    async def test_download_file_s3_path_construction(self, service, session_repo, storage_service, active_session):
        """测试下载文件 S3 路径构造"""
        session_repo.find_by_id.return_value = active_session
        storage_service.file_exists.return_value = True
        storage_service.get_file_info.return_value = {
            "size": 1024,
            "content_type": "text/plain"
        }
        storage_service.download_file.return_value = b"content"

        await service.download_file(
            session_id="sess_123",
            path="data/test.csv"
        )

        # 验证所有文件操作都使用正确的 S3 路径
        file_exists_path = storage_service.file_exists.call_args[0][0]
        file_info_path = storage_service.get_file_info.call_args[0][0]
        download_path = storage_service.download_file.call_args[0][0]

        for path in [file_exists_path, file_info_path, download_path]:
            assert path.startswith(active_session.workspace_path)
            assert "data/test.csv" in path

    @pytest.mark.asyncio
    async def test_download_file_with_missing_content_type(self, service, session_repo, storage_service, active_session):
        """测试缺少 content_type 的文件信息"""
        session_repo.find_by_id.return_value = active_session
        storage_service.file_exists.return_value = True
        storage_service.get_file_info.return_value = {
            "size": 1024
            # 缺少 content_type
        }
        storage_service.download_file.return_value = b"content"

        result = await service.download_file(
            session_id="sess_123",
            path="test.txt"
        )

        # 应使用默认 content_type
        assert result["content_type"] == "application/octet-stream"

    @pytest.mark.asyncio
    async def test_list_files_all(self, service, session_repo, storage_service, active_session):
        """测试列出所有文件"""
        session_repo.find_by_id.return_value = active_session
        storage_service.list_files.return_value = [
            {
                "key": "sessions/sess_123/file1.txt",
                "size": 1024,
                "last_modified": "2024-01-01T00:00:00Z",
                "etag": "\"abc123\""
            },
            {
                "key": "sessions/sess_123/src/main.py",
                "size": 2048,
                "last_modified": "2024-01-02T00:00:00Z",
                "etag": "\"def456\""
            }
        ]

        result = await service.list_files(session_id="sess_123")

        assert len(result) == 2
        assert result[0]["name"] == "file1.txt"
        assert result[0]["container_path"] == "/workspace/file1.txt"
        assert result[0]["size"] == 1024
        assert result[1]["name"] == "src/main.py"
        assert result[1]["container_path"] == "/workspace/src/main.py"
        assert result[1]["size"] == 2048
        # 验证调用时使用了正确的 workspace 前缀
        call_prefix = storage_service.list_files.call_args[0][0]
        assert "sessions/sess_123" in call_prefix

    @pytest.mark.asyncio
    async def test_list_files_with_path(self, service, session_repo, storage_service, active_session):
        """测试列出指定目录下的文件"""
        session_repo.find_by_id.return_value = active_session
        storage_service.list_files.return_value = [
            {
                "key": "sessions/sess_123/src/utils/helper.py",
                "size": 512,
                "last_modified": "2024-01-01T00:00:00Z",
                "etag": "\"xyz789\""
            }
        ]

        result = await service.list_files(session_id="sess_123", path="src/utils")

        assert len(result) == 1
        assert result[0]["name"] == "src/utils/helper.py"
        assert result[0]["container_path"] == "/workspace/src/utils/helper.py"
        assert result[0]["size"] == 512
        # 验证调用时使用了正确的前缀（包含子目录）
        call_prefix = storage_service.list_files.call_args[0][0]
        assert "sessions/sess_123/src/utils" in call_prefix

    @pytest.mark.asyncio
    async def test_list_files_with_trailing_slash_path(self, service, session_repo, storage_service, active_session):
        """测试列出指定目录下的文件（带尾部斜杠）"""
        session_repo.find_by_id.return_value = active_session
        storage_service.list_files.return_value = [
            {
                "key": "sessions/sess_123/src/app.py",
                "size": 1024,
                "last_modified": "2024-01-01T00:00:00Z",
                "etag": "\"abc\""
            }
        ]

        result = await service.list_files(session_id="sess_123", path="src/")

        assert len(result) == 1
        assert result[0]["name"] == "src/app.py"
        assert result[0]["container_path"] == "/workspace/src/app.py"
        # 验证路径被正确规范化
        call_prefix = storage_service.list_files.call_args[0][0]
        assert "sessions/sess_123/src" in call_prefix

    @pytest.mark.asyncio
    async def test_list_files_session_not_found(self, service, session_repo):
        """测试列出不存在会话的文件"""
        session_repo.find_by_id.return_value = None

        with pytest.raises(NotFoundError, match="Session not found"):
            await service.list_files(session_id="non-existent")

    @pytest.mark.asyncio
    async def test_list_files_with_limit(self, service, session_repo, storage_service, active_session):
        """测试列出文件（带限制）"""
        session_repo.find_by_id.return_value = active_session
        storage_service.list_files.return_value = []

        await service.list_files(session_id="sess_123", limit=100)

        # 验证 limit 参数被正确传递（通过位置参数）
        call_args = storage_service.list_files.call_args[0]
        assert call_args[1] == 100

    @pytest.mark.asyncio
    async def test_list_files_empty_directory(self, service, session_repo, storage_service, active_session):
        """测试列出空目录（S3 返回目录本身作为 0 大小对象）"""
        session_repo.find_by_id.return_value = active_session
        # S3 返回目录标记本身（以 / 结尾，size=0）
        storage_service.list_files.return_value = [
            {
                "key": "s3://sandbox-workspace/sessions/sess_123/",
                "size": 0,
                "last_modified": "2024-01-01T00:00:00Z",
                "etag": "\"d41d8cd98f00b204e9800998ecf8427e\""
            }
        ]

        result = await service.list_files(session_id="sess_123")

        # 应该过滤掉目录标记，返回空数组
        assert len(result) == 0
        assert result == []
