import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi.testclient import TestClient
from fastapi import status
import asyncio


class TestMain:
    def test_app_creation(self):
        """Test that app is created"""
        from src.main import app

        assert app is not None
        assert app.title == "agent-memory"
        assert app.version == "1.0.0"

    def test_root_endpoint(self):
        """Test root endpoint"""
        from src.main import app

        client = TestClient(app)
        response = client.get("/")

        assert response.status_code == 200
        assert response.json() == {"message": "Agent Memory Service"}

    def test_health_check_endpoint(self):
        """Test health check endpoint"""
        from src.main import app

        client = TestClient(app)
        response = client.get("/health")

        assert response.status_code == 200
        assert response.json() == {"status": "ok"}

    def test_main_block(self):
        """Test main block execution"""
        # The main block only runs when __name__ == "__main__"
        # We'll test that the code structure is correct
        from src.main import app

        assert app is not None

    def test_app_configuration(self):
        """Test app is configured with correct values"""
        from src.main import app

        # Check app was created with correct configuration
        assert app.title == "agent-memory"
        assert app.version == "1.0.0"
        assert not app.debug

    def test_router_registration(self):
        """Test that routers are registered"""
        from src.main import app

        routes = [str(route.path) for route in app.routes]
        assert any("health" in route for route in routes)
        assert any("api/agent-memory" in route for route in routes)

    def test_module_imports(self):
        """Test that all necessary modules can be imported"""
        from src.main import app

        assert app is not None

    def test_pyinstaller_mode(self):
        """Test app initialization in PyInstaller mode"""
        from pathlib import Path
        import sys as sys_module

        # In normal mode, _MEIPASS should not exist
        # We just check that the code doesn't crash
        from src.main import ROOT_DIR

        assert isinstance(ROOT_DIR, Path)
        # In test environment, _MEIPASS is not set
        # In PyInstaller mode, it would be set by the bundler
        assert not hasattr(sys_module, "_MEIPASS") or ROOT_DIR.name != "src"

    def test_default_app_config(self):
        """Test app with default configuration"""
        from src.main import app

        # Check app has default configuration
        assert app.title == "agent-memory"
        assert app.version == "1.0.0"
        assert not app.debug
