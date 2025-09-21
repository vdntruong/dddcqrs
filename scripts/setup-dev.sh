#!/bin/bash

set -e

echo "🚀 Setting up CQRS development environment..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.25 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.25"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or later."
    exit 1
fi

echo "✅ Go version $GO_VERSION is compatible"

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "📦 Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
    echo "✅ golangci-lint installed"
else
    echo "✅ golangci-lint is already installed"
fi

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo "📦 Installing pre-commit..."
    pip install pre-commit
    echo "✅ pre-commit installed"
else
    echo "✅ pre-commit is already installed"
fi

# Setup Go workspace
echo "🔧 Setting up Go workspace..."
go work init
go work use ./shared ./order-management-service ./order-reporting-service
echo "✅ Go workspace configured"

# Download dependencies
echo "📥 Downloading dependencies..."
go mod download
echo "✅ Dependencies downloaded"

# Install pre-commit hooks
echo "🔧 Installing pre-commit hooks..."
pre-commit install
echo "✅ Pre-commit hooks installed"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "⚠️  Docker is not installed. You'll need Docker to run the full stack."
    echo "   Install Docker Desktop or Docker Engine to use docker-compose commands."
else
    echo "✅ Docker is installed"
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "⚠️  docker-compose is not installed. You'll need it to run the full stack."
    echo "   Install docker-compose to use the make docker-up command."
else
    echo "✅ docker-compose is installed"
fi

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "📋 Available commands:"
echo "  make lint          - Run golangci-lint on all modules"
echo "  make test          - Run tests for all modules"
echo "  make build         - Build all services"
echo "  make docker-up     - Start all services with Docker Compose"
echo "  make docker-down   - Stop all services"
echo "  make run-command-service - Run command service locally"
echo "  make run-query-service   - Run query service locally"
echo ""
echo "🔗 Service URLs (when running):"
echo "  Command Service:  http://localhost:8080"
echo "  Query Service:    http://localhost:8081"
echo "  PostgreSQL:       localhost:5432"
echo "  Redis:            localhost:6379"
echo "  Kafka:            localhost:9092"
echo ""
echo "📚 Next steps:"
echo "  1. Run 'make docker-up' to start the infrastructure"
echo "  2. Run 'make run-command-service' in one terminal"
echo "  3. Run 'make run-query-service' in another terminal"
echo "  4. Test the API endpoints"
