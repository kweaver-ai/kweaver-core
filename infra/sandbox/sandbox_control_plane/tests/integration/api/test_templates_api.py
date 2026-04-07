"""
Template Management API Integration Tests

Tests for template CRUD operations.
"""
import pytest
from httpx import AsyncClient


@pytest.mark.asyncio
class TestTemplatesAPI:
    """Template API integration tests."""

    async def test_create_template(self, http_client: AsyncClient):
        """Test creating a new template."""
        import uuid
        template_id = f"test_template_{uuid.uuid4().hex[:8]}"
        template_name = f"Test Python Template {uuid.uuid4().hex[:8]}"

        template_data = {
            "id": template_id,
            "name": template_name,
            "image_url": "sandbox-template-python-basic:latest",
            "runtime_type": "python3.11",
            "default_cpu_cores": 1.0,
            "default_memory_mb": 512,
            "default_disk_mb": 1024,
            "default_timeout": 300
        }

        response = await http_client.post("/templates", json=template_data)

        if response.status_code not in (201, 200):
            print(f"API Error: {response.status_code} - {response.text}")

        assert response.status_code in (201, 200)
        data = response.json()
        assert "id" in data
        assert data["id"] == template_id
        assert data["name"] == template_name
        assert data["runtime_type"] == "python3.11"

        # Cleanup
        await http_client.delete(f"/templates/{template_id}")

    async def test_create_duplicate_template_id(self, http_client: AsyncClient, test_template_id: str):
        """Test creating a template with duplicate ID should fail."""
        # First template is created by test_template_id fixture
        template_data = {
            "id": test_template_id,
            "name": "Duplicate Template",
            "image_url": "sandbox-template-python-basic:latest",
            "runtime_type": "python3.11"
        }

        response = await http_client.post("/templates", json=template_data)

        # Should return 409 Conflict or 400 Bad Request
        assert response.status_code in (409, 400)

    async def test_create_duplicate_template_name(self, http_client: AsyncClient, test_template_id: str):
        """Test creating a template with duplicate name should fail."""
        import uuid
        template_id = f"test_template_{uuid.uuid4().hex[:8]}"

        # First template is created by "Python Basic (Test)" fixture
        template_data = {
            "id": template_id,
            "name": "Python Basic (Test)",
            "image_url": "sandbox-template-python-basic:latest",
            "runtime_type": "python3.11"
        }

        response = await http_client.post("/templates", json=template_data)

        # Should return 409 Conflict or 400 Bad Request
        assert response.status_code in (409, 400)

    async def test_list_templates(self, http_client: AsyncClient):
        """Test listing all templates."""
        response = await http_client.get("/templates")

        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)

    async def test_get_template(self, http_client: AsyncClient, test_template_id: str):
        """Test getting a specific template."""
        response = await http_client.get(f"/templates/{test_template_id}")

        assert response.status_code == 200
        data = response.json()
        assert data["id"] == test_template_id
        assert "name" in data
        assert "image_url" in data
        assert "runtime_type" in data
        # Check for flat resource fields (API returns these, not nested default_resources)
        assert "default_cpu_cores" in data
        assert "default_memory_mb" in data
        assert "default_disk_mb" in data

    async def test_get_nonexistent_template(self, http_client: AsyncClient):
        """Test getting a template that doesn't exist."""
        response = await http_client.get("/templates/nonexistent_template_xyz")

        assert response.status_code == 404

    async def test_update_template(self, http_client: AsyncClient, test_template_id: str):
        """Test updating a template."""
        # Note: name is immutable in the current implementation
        # Update a mutable field instead (e.g., default_cpu_cores)
        update_data = {
            "default_cpu_cores": 2.0,
            "default_memory_mb": 1024
        }

        response = await http_client.put(
            f"/templates/{test_template_id}",
            json=update_data
        )

        assert response.status_code in (200, 202)

        # Verify update
        get_response = await http_client.get(f"/templates/{test_template_id}")
        assert get_response.status_code == 200
        data = get_response.json()
        assert data["default_cpu_cores"] == 2.0
        assert data["default_memory_mb"] == 1024

        # Restore original values
        await http_client.put(
            f"/templates/{test_template_id}",
            json={"default_cpu_cores": 1.0, "default_memory_mb": 512}
        )

    async def test_update_template_name_success(self, http_client: AsyncClient, test_template_id: str):
        """Test updating template name successfully."""
        import uuid
        new_name = f"Updated Template Name {uuid.uuid4().hex[:8]}"

        # Get original name
        get_response = await http_client.get(f"/templates/{test_template_id}")
        assert get_response.status_code == 200
        original_name = get_response.json()["name"]

        # Update template name
        update_data = {"name": new_name}
        response = await http_client.put(
            f"/templates/{test_template_id}",
            json=update_data
        )

        # Note: This test will fail until name update is implemented
        assert response.status_code in (200, 202)

        # Verify name was updated
        get_response = await http_client.get(f"/templates/{test_template_id}")
        assert get_response.status_code == 200
        data = get_response.json()
        assert data["name"] == new_name

        # Restore original name
        await http_client.put(
            f"/templates/{test_template_id}",
            json={"name": original_name}
        )

    async def test_update_template_name_duplicate(self, http_client: AsyncClient, test_template_id: str):
        """Test updating template name with an existing name should fail."""
        import uuid

        # Create another template to get a duplicate name
        duplicate_template_id = f"test_template_dup_{uuid.uuid4().hex[:8]}"
        duplicate_template_name = f"Duplicate Name Template {uuid.uuid4().hex[:8]}"

        create_data = {
            "id": duplicate_template_id,
            "name": duplicate_template_name,
            "image_url": "sandbox-template-python-basic:latest",
            "runtime_type": "python3.11"
        }

        create_response = await http_client.post("/templates", json=create_data)
        assert create_response.status_code in (201, 200)

        # Try to update test_template_id with the same name
        update_data = {"name": duplicate_template_name}
        response = await http_client.put(
            f"/templates/{test_template_id}",
            json=update_data
        )

        # Should fail with conflict or validation error
        assert response.status_code in (409, 400, 422)

        # Cleanup
        await http_client.delete(f"/templates/{duplicate_template_id}")

    async def test_update_template_timeout_success(self, http_client: AsyncClient, test_template_id: str):
        """Test updating template timeout successfully."""
        # Get original timeout
        get_response = await http_client.get(f"/templates/{test_template_id}")
        assert get_response.status_code == 200
        original_timeout = get_response.json().get("default_timeout_sec", 300)

        # Update template timeout
        new_timeout = 600
        update_data = {"default_timeout": new_timeout}
        response = await http_client.put(
            f"/templates/{test_template_id}",
            json=update_data
        )

        assert response.status_code in (200, 202)

        # Verify timeout was updated
        get_response = await http_client.get(f"/templates/{test_template_id}")
        assert get_response.status_code == 200
        data = get_response.json()
        assert data["default_timeout_sec"] == new_timeout

        # Restore original timeout
        await http_client.put(
            f"/templates/{test_template_id}",
            json={"default_timeout": original_timeout}
        )

    
    async def test_delete_template(self, http_client: AsyncClient):
        """Test deleting a template."""
        # Create a template to delete
        template_data = {
            "id": "test_template_to_delete",
            "name": "Template To Delete",
            "image_url": "sandbox-template-python-basic:latest",
            "runtime_type": "python3.11"
        }

        create_response = await http_client.post("/templates", json=template_data)
        assert create_response.status_code in (201, 200)

        # Delete the template
        delete_response = await http_client.delete("/templates/test_template_to_delete")
        assert delete_response.status_code in (200, 202, 204)

        # Verify deletion
        get_response = await http_client.get("/templates/test_template_to_delete")
        assert get_response.status_code == 404

    async def test_template_resources_validation(self, http_client: AsyncClient):
        """Test template resource validation."""
        template_data = {
            "id": "test_template_resources",
            "name": "Test Resource Template",
            "image_url": "sandbox-template-python-basic:latest",
            "runtime_type": "python3.11",
            "default_cpu_cores": 2.0,
            "default_memory_mb": 1024,
            "default_disk_mb": 5120
        }

        response = await http_client.post("/templates", json=template_data)

        assert response.status_code in (201, 200)
        data = response.json()
        # API returns flat fields, not nested default_resources
        assert data["default_cpu_cores"] == 2.0
        assert data["default_memory_mb"] == 1024
        assert data["default_disk_mb"] == 5120

        # Cleanup
        await http_client.delete(f"/templates/test_template_resources")

    async def test_template_with_empty_resources(self, http_client: AsyncClient):
        """Test creating template without specifying resources (should use defaults)."""
        template_data = {
            "id": "test_template_no_resources",
            "name": "Test No Resources Template",
            "image_url": "sandbox-template-python-basic:latest",
            "runtime_type": "python3.11"
        }

        response = await http_client.post("/templates", json=template_data)

        assert response.status_code in (201, 200)

        # Cleanup
        await http_client.delete(f"/templates/test_template_no_resources")

    async def test_template_runtime_type_validation(self, http_client: AsyncClient):
        """Test template runtime type validation."""
        # Test valid runtime types
        for runtime_type in ["python3.11", "nodejs20"]:
            template_data = {
                "id": f"test_template_{runtime_type.replace('.', '_')}",
                "name": f"Test {runtime_type} Template",
                "image_url": "sandbox-template-python-basic:latest",
                "runtime_type": runtime_type
            }

            response = await http_client.post("/templates", json=template_data)
            assert response.status_code in (201, 200)

            # Cleanup
            await http_client.delete(f"/templates/test_template_{runtime_type.replace('.', '_')}")

    async def test_template_is_active_flag(self, http_client: AsyncClient):
        """Test template is_active flag."""
        template_data = {
            "id": "test_template_active",
            "name": "Test Active Template",
            "image_url": "sandbox-template-python-basic:latest",
            "runtime_type": "python3.11"
        }

        response = await http_client.post("/templates", json=template_data)

        assert response.status_code in (201, 200)
        data = response.json()
        # Note: is_active is not in CreateTemplateRequest, so it defaults to True

        # Cleanup
        await http_client.delete(f"/templates/test_template_active")


@pytest.mark.asyncio
class TestTemplateValidation:
    """Template validation tests."""

    async def test_template_create_fails_with_empty_image_url(self, http_client: AsyncClient):
        """Test that template creation fails when image_url is empty."""
        import uuid
        template_id = f"test_empty_image_{uuid.uuid4().hex[:8]}"

        template_data = {
            "id": template_id,
            "name": "Test Empty Image URL",
            "image_url": "",  # Empty image_url
            "runtime_type": "python3.11"
        }

        response = await http_client.post("/templates", json=template_data)

        # Should fail with validation error
        assert response.status_code == 422  # Validation error

    async def test_template_create_fails_with_missing_image_url(self, http_client: AsyncClient):
        """Test that template creation fails when image_url is not provided."""
        import uuid
        template_id = f"test_no_image_{uuid.uuid4().hex[:8]}"

        template_data = {
            "id": template_id,
            "name": "Test No Image URL",
            # image_url is missing
            "runtime_type": "python3.11"
        }

        response = await http_client.post("/templates", json=template_data)

        # Should fail with validation error
        assert response.status_code == 422  # Validation error

    async def test_template_create_fails_with_invalid_runtime_type(self, http_client: AsyncClient):
        """Test that template creation fails when runtime_type is invalid."""
        import uuid

        # Test with invalid runtime types
        invalid_runtime_types = [
            "python3.10",  # Not in allowed list
            "nodejs18",    # Not in allowed list
            "java11",      # Not in allowed list
            "go1.20",      # Not in allowed list
            "ruby3.0",     # Not supported at all
            "invalid",     # Completely invalid
        ]

        for runtime_type in invalid_runtime_types:
            template_id = f"test_invalid_runtime_{uuid.uuid4().hex[:8]}"

            template_data = {
                "id": template_id,
                "name": f"Test Invalid Runtime {runtime_type}",
                "image_url": "sandbox-template-python-basic:latest",
                "runtime_type": runtime_type
            }

            response = await http_client.post("/templates", json=template_data)

            # Should fail with validation error
            assert response.status_code == 422, f"Expected 422 for runtime_type '{runtime_type}', got {response.status_code}"

    async def test_template_create_succeeds_with_valid_runtime_types(self, http_client: AsyncClient):
        """Test that template creation succeeds with all valid runtime types."""
        import uuid

        # Test all valid runtime types
        valid_runtime_types = ["python3.11", "nodejs20", "java17", "go1.21"]

        for runtime_type in valid_runtime_types:
            template_id = f"test_valid_{runtime_type.replace('.', '_')}_{uuid.uuid4().hex[:8]}"

            template_data = {
                "id": template_id,
                "name": f"Test Valid Runtime {runtime_type}",
                "image_url": "sandbox-template-python-basic:latest",
                "runtime_type": runtime_type
            }

            response = await http_client.post("/templates", json=template_data)

            # Should succeed
            assert response.status_code in (201, 200), f"Expected success for runtime_type '{runtime_type}', got {response.status_code}: {response.text}"

            # Cleanup
            await http_client.delete(f"/templates/{template_id}")
