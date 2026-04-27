"""
S3 存储单元测试

测试 S3Storage 类的功能。
"""
import pytest
from unittest.mock import Mock, AsyncMock, patch, MagicMock
from datetime import datetime
import asyncio

from src.infrastructure.storage.s3_storage import S3Storage


class TestS3Storage:
    """S3 存储测试"""

    @pytest.fixture
    def mock_boto_client(self):
        """模拟 boto3 客户端"""
        client = Mock()
        return client

    @pytest.fixture
    def mock_settings(self):
        """模拟配置"""
        settings = Mock()
        settings.s3_endpoint_url = "http://localhost:9000"
        settings.s3_access_key_id = "minioadmin"
        settings.s3_secret_access_key = "minioadmin"
        settings.s3_region = "us-east-1"
        settings.s3_bucket = "sandbox-workspace"
        return settings

    @pytest.fixture
    def storage(self, mock_boto_client, mock_settings):
        """创建 S3 存储实例"""
        with patch('src.infrastructure.storage.s3_storage.boto3.client', return_value=mock_boto_client):
            with patch('src.infrastructure.storage.s3_storage.get_settings', return_value=mock_settings):
                return S3Storage()

    def test_parse_s3_path_with_prefix(self, storage):
        """测试解析带 s3:// 前缀的路径"""
        bucket, key = storage._parse_s3_path("s3://my-bucket/path/to/file.txt")

        assert bucket == "my-bucket"
        assert key == "path/to/file.txt"

    def test_parse_s3_path_without_prefix(self, storage):
        """测试解析不带 s3:// 前缀的路径"""
        bucket, key = storage._parse_s3_path("path/to/file.txt")

        assert bucket == storage._bucket
        assert key == "path/to/file.txt"

    def test_parse_s3_path_with_leading_slash(self, storage):
        """测试解析带前导斜杠的路径"""
        bucket, key = storage._parse_s3_path("/path/to/file.txt")

        assert bucket == storage._bucket
        assert key == "path/to/file.txt"

    def test_build_s3_path(self, storage):
        """测试构建 S3 路径"""
        path = storage._build_s3_path("my-bucket", "path/to/file.txt")

        assert path == "s3://my-bucket/path/to/file.txt"

    @pytest.mark.asyncio
    async def test_upload_file_small(self, storage, mock_boto_client):
        """测试上传小文件"""
        mock_boto_client.head_object.return_value = {}
        mock_boto_client.put_object.return_value = {}

        content = b"hello world"
        await storage.upload_file("s3://test-bucket/test.txt", content, "text/plain")

        # 验证调用了 put_object
        mock_boto_client.put_object.assert_called_once()
        call_args = mock_boto_client.put_object.call_args
        assert call_args[1]["Bucket"] == "test-bucket"
        assert call_args[1]["Key"] == "test.txt"
        assert call_args[1]["Body"] == content
        assert call_args[1]["ContentType"] == "text/plain"

    @pytest.mark.asyncio
    async def test_upload_file_with_bucket_check(self, storage, mock_boto_client):
        """测试上传文件时检查 bucket 存在"""
        mock_boto_client.head_bucket.return_value = {}
        mock_boto_client.head_object.side_effect = Exception("Not Found")
        mock_boto_client.put_object.return_value = {}

        await storage.upload_file("s3://test-bucket/test.txt", b"content")

        # 应该检查 bucket
        mock_boto_client.head_bucket.assert_called()

    @pytest.mark.asyncio
    async def test_download_file(self, storage, mock_boto_client):
        """测试下载文件"""
        mock_response = {
            'Body': Mock(read=Mock(return_value=b'file content'))
        }
        mock_boto_client.get_object.return_value = mock_response

        content = await storage.download_file("s3://test-bucket/test.txt")

        assert content == b'file content'
        mock_boto_client.get_object.assert_called_once_with(
            Bucket="test-bucket",
            Key="test.txt"
        )

    @pytest.mark.asyncio
    async def test_file_exists_true(self, storage, mock_boto_client):
        """测试检查文件存在（存在）"""
        mock_boto_client.head_object.return_value = {}

        exists = await storage.file_exists("s3://test-bucket/test.txt")

        assert exists is True

    @pytest.mark.asyncio
    async def test_file_exists_false(self, storage, mock_boto_client):
        """测试检查文件存在（不存在）"""
        from botocore.exceptions import ClientError
        error_response = {'Error': {'Code': '404'}}
        mock_boto_client.head_object.side_effect = ClientError(error_response, 'HeadObject')

        exists = await storage.file_exists("s3://test-bucket/test.txt")

        assert exists is False

    @pytest.mark.asyncio
    async def test_file_exists_error(self, storage, mock_boto_client):
        """测试检查文件存在时出错"""
        from botocore.exceptions import ClientError
        error_response = {'Error': {'Code': '403'}}
        mock_boto_client.head_object.side_effect = ClientError(error_response, 'HeadObject')

        with pytest.raises(ClientError):
            await storage.file_exists("s3://test-bucket/test.txt")

    @pytest.mark.asyncio
    async def test_get_file_info(self, storage, mock_boto_client):
        """测试获取文件信息"""
        mock_response = {
            'ContentLength': 1024,
            'ContentType': 'text/plain',
            'LastModified': datetime.now(),
            'ETag': '"abc123"'
        }
        mock_boto_client.head_object.return_value = mock_response

        info = await storage.get_file_info("s3://test-bucket/test.txt")

        assert info["size"] == 1024
        assert info["content_type"] == "text/plain"
        assert "last_modified" in info
        assert info["etag"] == "abc123"

    @pytest.mark.asyncio
    async def test_get_file_info_default_content_type(self, storage, mock_boto_client):
        """测试获取文件信息（缺少 content_type）"""
        mock_response = {
            'ContentLength': 2048,
            'LastModified': datetime.now(),
            'ETag': '"def456"'
        }
        mock_boto_client.head_object.return_value = mock_response

        info = await storage.get_file_info("s3://test-bucket/test.bin")

        assert info["content_type"] == "application/octet-stream"

    @pytest.mark.asyncio
    async def test_generate_presigned_url(self, storage, mock_boto_client):
        """测试生成预签名 URL"""
        mock_boto_client.generate_presigned_url.return_value = "https://s3.amazonaws.com/..."

        url = await storage.generate_presigned_url("s3://test-bucket/test.txt", 3600)

        assert url == "https://s3.amazonaws.com/..."
        mock_boto_client.generate_presigned_url.assert_called_once()

    @pytest.mark.asyncio
    async def test_delete_file(self, storage, mock_boto_client):
        """测试删除文件"""
        mock_boto_client.delete_object.return_value = {}

        await storage.delete_file("s3://test-bucket/test.txt")

        mock_boto_client.delete_object.assert_called_once_with(
            Bucket="test-bucket",
            Key="test.txt"
        )

    @pytest.mark.asyncio
    async def test_delete_prefix(self, storage, mock_boto_client):
        """测试删除指定前缀的所有文件"""
        # 模拟分页器
        mock_paginator = Mock()
        mock_page_iterator = [
            {
                'Contents': [
                    {'Key': f'sessions/sess_123/file{i}.txt'} for i in range(5)
                ]
            },
            {
                'Contents': [
                    {'Key': f'sessions/sess_123/file{i}.txt'} for i in range(5, 8)
                ]
            }
        ]
        mock_paginator.paginate.return_value = mock_page_iterator
        mock_paginator.return_value = mock_paginator

        mock_boto_client.get_paginator.return_value = mock_paginator
        mock_boto_client.delete_objects.return_value = {}

        deleted_count = await storage.delete_prefix("s3://test-bucket/sessions/sess_123/")

        assert deleted_count == 8

    @pytest.mark.asyncio
    async def test_delete_prefix_with_bucket_in_prefix(self, storage, mock_boto_client):
        """测试删除带 bucket 的前缀"""
        mock_paginator = Mock()
        mock_paginator.paginate.return_value = [{
            'Contents': [
                {'Key': 'sessions/sess_123/file1.txt'}
            ]
        }]
        mock_boto_client.get_paginator.return_value = mock_paginator
        mock_boto_client.delete_objects.return_value = {}

        deleted_count = await storage.delete_prefix("s3://test-bucket/sessions/sess_123/")

        assert deleted_count == 1

    @pytest.mark.asyncio
    async def test_list_files(self, storage, mock_boto_client):
        """测试列出文件"""
        mock_paginator = Mock()
        mock_paginator.paginate.return_value = [{
            'Contents': [
                {
                    'Key': 'test/file1.txt',
                    'Size': 1024,
                    'LastModified': datetime.now(),
                    'ETag': '"abc123"'
                },
                {
                    'Key': 'test/file2.txt',
                    'Size': 2048,
                    'LastModified': datetime.now(),
                    'ETag': '"def456"'
                }
            ]
        }]
        mock_boto_client.get_paginator.return_value = mock_paginator

        files = await storage.list_files("s3://test-bucket/test/", limit=100)

        assert len(files) == 2
        assert files[0]['key'] == 'test/file1.txt'
        assert files[1]['key'] == 'test/file2.txt'

    @pytest.mark.asyncio
    async def test_list_files_with_limit(self, storage, mock_boto_client):
        """测试列出文件（限制数量）"""
        mock_paginator = Mock()
        mock_paginator.paginate.return_value = [{
            'Contents': [
                {'Key': f'test/file{i}.txt', 'Size': 100 * i, 'LastModified': datetime.now(), 'ETag': f'"{i}"'}
                for i in range(20)
            ]
        }]
        mock_boto_client.get_paginator.return_value = mock_paginator

        files = await storage.list_files("s3://test-bucket/test/", limit=5)

        assert len(files) == 5

    @pytest.mark.asyncio
    async def test_list_files_with_bucket_in_prefix(self, storage, mock_boto_client):
        """测试列出文件（带 bucket）"""
        mock_paginator = Mock()
        mock_paginator.paginate.return_value = [{
            'Contents': [
                {'Key': 'sessions/sess_123/file.txt', 'Size': 1024, 'LastModified': datetime.now(), 'ETag': '"abc"'}
            ]
        }]
        mock_boto_client.get_paginator.return_value = mock_paginator

        files = await storage.list_files("s3://test-bucket/sessions/sess_123/")

        assert len(files) == 1
        # key 应该是相对于 bucket 的路径，包含 sessions/sess_123/file.txt
        assert 'sessions/' in files[0]['key']

    @pytest.mark.asyncio
    async def test_initialize_success(self, storage, mock_boto_client):
        """测试初始化成功"""
        mock_boto_client.head_bucket.return_value = {}

        await storage.initialize()

        mock_boto_client.head_bucket.assert_called_once()

    @pytest.mark.asyncio
    async def test_initialize_creates_bucket(self, storage, mock_boto_client):
        """测试初始化时创建 bucket"""
        from botocore.exceptions import ClientError

        # head_bucket 抛出 404 错误
        error_response = {'Error': {'Code': '404'}}
        mock_boto_client.head_bucket.side_effect = ClientError(error_response, 'HeadBucket')
        mock_boto_client.meta.region_name = 'us-west-2'
        mock_boto_client.create_bucket.return_value = {}

        await storage.initialize()

        mock_boto_client.create_bucket.assert_called_once()

    @pytest.mark.asyncio
    async def test_initialize_creates_bucket_us_east_1(self, storage, mock_boto_client):
        """测试初始化时创建 bucket（us-east-1 区域）"""
        from botocore.exceptions import ClientError

        error_response = {'Error': {'Code': '404'}}
        mock_boto_client.head_bucket.side_effect = ClientError(error_response, 'HeadBucket')
        mock_boto_client.meta.region_name = 'us-east-1'
        mock_boto_client.create_bucket.return_value = {}

        await storage.initialize()

        # us-east-1 不需要 LocationConstraint
        call_args = mock_boto_client.create_bucket.call_args
        assert 'CreateBucketConfiguration' not in call_args[1]

    @pytest.mark.asyncio
    async def test_initialize_bucket_create_error(self, storage, mock_boto_client):
        """测试初始化时创建 bucket 失败"""
        from botocore.exceptions import ClientError

        error_response = {'Error': {'Code': '404'}}
        mock_boto_client.head_bucket.side_effect = ClientError(error_response, 'HeadBucket')
        mock_boto_client.meta.region_name = 'us-west-2'

        create_error = ClientError({'Error': {'Code': '403'}}, 'CreateBucket')
        mock_boto_client.create_bucket.side_effect = create_error

        # 不应该抛出异常，只是记录错误
        await storage.initialize()

    @pytest.mark.asyncio
    async def test_upload_large_file(self, storage, mock_boto_client):
        """测试上传大文件（分片上传）"""
        mock_boto_client.head_bucket.return_value = {}
        mock_boto_client.head_object.side_effect = Exception("Not Found")
        mock_boto_client.upload_file = Mock()

        # 创建大于 5MB 的内容
        large_content = b"x" * (6 * 1024 * 1024)  # 6MB

        await storage.upload_file("s3://test-bucket/large.bin", large_content)

        # 大文件应该使用 upload_file
        mock_boto_client.upload_file.assert_called_once()

    @pytest.mark.asyncio
    async def test_upload_file_removes_directory_marker(self, storage, mock_boto_client):
        """测试上传文件时删除目录标记"""
        mock_boto_client.head_bucket.return_value = {}
        mock_boto_client.put_object.return_value = {}

        # 目录标记存在
        mock_boto_client.head_object.return_value = {}
        mock_boto_client.delete_object.return_value = {}

        await storage.upload_file("s3://test-bucket/dir/file.txt", b"content")

        # 应该删除目录标记
        mock_boto_client.delete_object.assert_called()

    @pytest.mark.asyncio
    async def test_upload_file_directory_marker_not_exists(self, storage, mock_boto_client):
        """测试上传文件时目录标记不存在"""
        from botocore.exceptions import ClientError

        mock_boto_client.head_bucket.return_value = {}
        mock_boto_client.put_object.return_value = {}

        # 目录标记不存在
        error_response = {'Error': {'Code': '404'}}
        mock_boto_client.head_object.side_effect = ClientError(error_response, 'HeadObject')

        await storage.upload_file("s3://test-bucket/dir/file.txt", b"content")

        # 不应该删除目录标记
        mock_boto_client.delete_object.assert_not_called()

    @pytest.mark.asyncio
    async def test_delete_prefix_empty(self, storage, mock_boto_client):
        """测试删除空前缀"""
        mock_paginator = Mock()
        mock_paginator.paginate.return_value = [{}]  # 没有 Contents
        mock_boto_client.get_paginator.return_value = mock_paginator

        deleted_count = await storage.delete_prefix("sessions/empty/")

        assert deleted_count == 0

    @pytest.mark.asyncio
    async def test_delete_prefix_error(self, storage, mock_boto_client):
        """测试删除前缀时出错"""
        from botocore.exceptions import ClientError

        mock_paginator = Mock()
        mock_paginator.paginate.side_effect = ClientError(
            {'Error': {'Code': '403'}}, 'ListObjectsV2'
        )
        mock_boto_client.get_paginator.return_value = mock_paginator

        deleted_count = await storage.delete_prefix("sessions/test/")

        assert deleted_count == 0

    @pytest.mark.asyncio
    async def test_list_files_empty(self, storage, mock_boto_client):
        """测试列出空目录"""
        mock_paginator = Mock()
        mock_paginator.paginate.return_value = [{}]  # 没有 Contents
        mock_boto_client.get_paginator.return_value = mock_paginator

        files = await storage.list_files("empty/")

        assert files == []

    @pytest.mark.asyncio
    async def test_list_files_error(self, storage, mock_boto_client):
        """测试列出文件时出错"""
        from botocore.exceptions import ClientError

        mock_paginator = Mock()
        mock_paginator.paginate.side_effect = ClientError(
            {'Error': {'Code': '403'}}, 'ListObjectsV2'
        )
        mock_boto_client.get_paginator.return_value = mock_paginator

        files = await storage.list_files("test/")

        assert files == []

    @pytest.mark.asyncio
    async def test_delete_prefix_relative_path(self, storage, mock_boto_client):
        """测试删除相对路径前缀"""
        mock_paginator = Mock()
        mock_paginator.paginate.return_value = [{
            'Contents': [
                {'Key': 'sessions/sess_123/file.txt'}
            ]
        }]
        mock_boto_client.get_paginator.return_value = mock_paginator
        mock_boto_client.delete_objects.return_value = {}

        deleted_count = await storage.delete_prefix("sessions/sess_123/")

        assert deleted_count == 1

    @pytest.mark.asyncio
    async def test_list_files_relative_path(self, storage, mock_boto_client):
        """测试列出相对路径文件"""
        mock_paginator = Mock()
        mock_paginator.paginate.return_value = [{
            'Contents': [
                {'Key': 'sessions/sess_123/file.txt', 'Size': 1024, 'LastModified': datetime.now(), 'ETag': '"abc"'}
            ]
        }]
        mock_boto_client.get_paginator.return_value = mock_paginator

        files = await storage.list_files("sessions/sess_123/")

        assert len(files) == 1
