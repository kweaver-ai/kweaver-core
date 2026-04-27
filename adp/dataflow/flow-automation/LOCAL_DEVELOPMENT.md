# Local Development Setup Guide

This guide explains how to run the flow-automation service locally with mock dependencies.

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for MongoDB, Redis, MariaDB, Kafka)
- Port availability: 8082, 8083, 9703, 30980, 28000, 6379, 3330, 9097

## Quick Start

### 1. Start Infrastructure Services

Start MongoDB, Redis, MariaDB, and Kafka using Docker Compose:

```bash
docker-compose up -d
```

Verify services are running:
```bash
docker-compose ps
```

### 2. Start Mock Services

The mock services simulate Kubernetes cluster services that aren't available locally:

```bash
./start-mock-services.sh
```

Or manually:
```bash
cd mock-server
go run main.go
```

This will start:
- **User Management Service** on port 30980
- **Deploy Service** on port 9703

### 3. Start the Main Application

In a new terminal:

```bash
go run main.go
```

The application will:
- Load configuration from `.env`
- Override with `.env.local` for local development
- Connect to local MongoDB, Redis, MariaDB
- Use mock services for Kubernetes dependencies

## Configuration

### Environment Files

- **`.env`**: Base configuration (committed to git)
- **`.env.local`**: Local overrides (gitignored, for development)

The `.env.local` file overrides Kubernetes service addresses with localhost:

```bash
DEPLOYSERVICE_HOST=localhost
USERMANAGEMENT_PRIVATE_HOST=localhost
REDIS_HOST=localhost
REDIS_CLUSTER_MODE=standalone
```

### Service Ports

| Service | Port | Purpose |
|---------|------|---------|
| Main API | 8082 | Public API server |
| Private API | 8083 | Private API server |
| MongoDB | 28000 | Database |
| Redis | 6379 | Cache |
| MariaDB | 3330 | Relational database |
| Kafka | 9097 | Message queue |
| Zookeeper | 2181 | Kafka coordination |
| User Management Mock | 30980 | Mock service |
| Deploy Service Mock | 9703 | Mock service |

## Troubleshooting

### Issue: "no such host" errors

**Problem**: Application trying to connect to Kubernetes services

**Solution**: 
1. Ensure `.env.local` exists and overrides the service hosts
2. Ensure mock services are running (`./start-mock-services.sh`)

### Issue: "connection refused" to MongoDB/Redis/MariaDB

**Problem**: Docker services not running

**Solution**:
```bash
docker-compose up -d
docker-compose ps  # Check status
docker-compose logs  # Check logs
```

### Issue: Port already in use

**Problem**: Another service is using the required port

**Solution**:
```bash
# Find what's using the port (example for port 8082)
lsof -i :8082

# Kill the process or change the port in .env.local
```

### Issue: Kafka/Zookeeper unhealthy

**Problem**: Kafka container fails to start

**Solution**:
```bash
# Check logs
docker-compose logs kafka zookeeper

# Restart services
docker-compose restart kafka zookeeper

# If still failing, recreate
docker-compose down
docker-compose up -d
```

### Issue: Missing config file errors

**Problem**: Application looking for `/conf/flow_o11y_data.yaml`

**Solution**: The file is created in `./conf/flow_o11y_data.yaml`. If you see errors, ensure the `conf` directory exists in the project root.

## Adding More Mock Services

If you encounter more "no such host" errors for other Kubernetes services:

1. Add the service mock to `mock-server/main.go`
2. Update `.env.local` to point to localhost
3. Restart the mock server

Example:
```go
func startNewServiceMock() {
    router := gin.New()
    router.GET("/api/endpoint", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })
    router.Run(":PORT")
}
```

## Development Workflow

1. Make code changes
2. Restart `go run main.go` (or use air for hot reload)
3. Test your changes
4. Mock services and Docker containers can keep running

## Stopping Services

```bash
# Stop main application: Ctrl+C in the terminal

# Stop mock services: Ctrl+C in the mock server terminal

# Stop Docker services:
docker-compose down

# Stop and remove volumes (clean slate):
docker-compose down -v
```

## Health Checks

- Main API: http://localhost:8082/api/automation/v1/health
- Mock User Management: http://localhost:30980/health
- Mock Deploy Service: http://localhost:9703/health

## Next Steps

- Set up hot reload with [air](https://github.com/cosmtrek/air)
- Configure your IDE for debugging
- Review the mock service responses in `mock-server/main.go` and adjust as needed
