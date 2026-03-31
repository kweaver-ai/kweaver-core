# Mock Services for Local Development

This directory contains mock services that simulate the Kubernetes cluster services required by the flow-automation application.

## Services Mocked

1. **User Management Service** (port 30980)
   - Endpoint: `POST /api/user-management/v1/apps`
   - Purpose: Mock app registration for internal accounts

2. **Deploy Service** (port 9703)
   - Endpoint: `GET /api/deploy-manager/v1/access-addr/app`
   - Purpose: Mock cluster access information

3. **Authorization Service** (port 30920)
   - Endpoint: `POST /api/authorization/v1/operation-check`
   - Purpose: Mock authorization permission checks and policy management

## Usage

### Option 1: Using the start script (recommended)

From the project root directory:

```bash
./start-mock-services.sh
```

### Option 2: Manual start

```bash
cd mock-server
go mod init mock-server  # Only needed first time
go get github.com/gin-gonic/gin  # Only needed first time
go run main.go
```

## Configuration

The mock services use the following default ports:
- User Management: 30980
- Deploy Service: 9703
- Authorization: 30920

You can override these by setting environment variables:
```bash
export USER_MANAGEMENT_MOCK_PORT=30980
export DEPLOY_SERVICE_MOCK_PORT=9703
export AUTHORIZATION_MOCK_PORT=30920
```

## Integration with Main Application

The main application will automatically use these mock services when running locally, thanks to the `.env.local` configuration file which overrides the Kubernetes service addresses with `localhost`.

## Adding More Mock Services

To add more mock services:

1. Add a new function like `startXXXServiceMock()` in `main.go`
2. Start it as a goroutine in the `main()` function
3. Update `.env.local` to point the service to localhost
4. Update this README

## Health Checks

Each mock service provides a health check endpoint:
- User Management: http://localhost:30980/health
- Deploy Service: http://localhost:9703/health
- Authorization: http://localhost:30920/health
