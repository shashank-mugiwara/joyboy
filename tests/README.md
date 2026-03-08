# Joyboy API Tests

This directory contains pytest-based integration tests for the Joyboy REST API.

## Setup

1. Install Python dependencies:
```bash
pip install -r requirements.txt
```

## Running the Tests

### Prerequisites
The Joyboy server must be running on `http://localhost:8070` before running the tests.

1. Start the Joyboy server:
```bash
go run main.go
```

2. In a separate terminal, run the tests:
```bash
pytest tests/
```

Or run with verbose output:
```bash
pytest tests/ -v
```

## Test Coverage

### Health Check Tests (`test_health.py`)
- `test_health_endpoint_returns_200`: Verifies the `/health` endpoint returns HTTP 200
- `test_health_endpoint_returns_json`: Verifies the response is JSON format
- `test_health_endpoint_returns_healthy_status`: Verifies the response contains `{"status": "healthy"}`
- `test_health_endpoint_accessible_without_auth`: Verifies the endpoint doesn't require authentication

## Configuration

The base URL for the API is configured in each test file. By default, it's set to `http://localhost:8070`.
To test against a different environment, modify the `BASE_URL` variable in the test files.
