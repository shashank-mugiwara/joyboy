import pytest
import requests


BASE_URL = "http://localhost:8070"


class TestHealthEndpoint:
    """Test suite for the REST API health check endpoint"""

    def test_health_endpoint_returns_200(self):
        """Test that the /health endpoint returns HTTP 200"""
        response = requests.get(f"{BASE_URL}/health")
        assert response.status_code == 200

    def test_health_endpoint_returns_json(self):
        """Test that the /health endpoint returns JSON content"""
        response = requests.get(f"{BASE_URL}/health")
        assert response.headers.get("Content-Type", "").startswith("application/json")

    def test_health_endpoint_returns_healthy_status(self):
        """Test that the /health endpoint returns a healthy status"""
        response = requests.get(f"{BASE_URL}/health")
        assert response.status_code == 200
        data = response.json()
        assert "status" in data
        assert data["status"] == "healthy"

    def test_health_endpoint_accessible_without_auth(self):
        """Test that the /health endpoint is accessible without authentication"""
        response = requests.get(f"{BASE_URL}/health")
        assert response.status_code == 200
        assert response.status_code != 401
        assert response.status_code != 403
