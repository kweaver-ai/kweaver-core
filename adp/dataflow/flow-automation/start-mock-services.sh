#!/bin/bash

# Start mock services for local development

echo "Starting mock services..."
cd "$(dirname "$0")/mock-server"

# Check if go.mod exists in mock-server directory
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module for mock server..."
    go mod init mock-server
    go get github.com/gin-gonic/gin
fi

# Run the mock server
echo "Running mock server..."
go run main.go
