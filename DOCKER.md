# Docker Deployment Guide

This guide explains how to build and run the Pacman Game application using Docker.

## Prerequisites

- Docker 20.10+ installed
- Docker Compose 2.0+ (optional, for orchestration)
- Make (optional, for convenience commands)

## Quick Start

### Option 1: Using Make (Recommended)

```bash
# Build and run with Docker
make docker-run

# View logs
make docker-logs

# Stop container
make docker-stop
```

### Option 2: Using Docker Compose

```bash
# Start the application
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the application
docker-compose down
```

### Option 3: Using Docker CLI

```bash
# Build the image
docker build -t pacman-game:latest .

# Run the container
docker run -d \
  --name pacman-game \
  -p 8080:8080 \
  pacman-game:latest

# View logs
docker logs -f pacman-game

# Stop the container
docker stop pacman-game
docker rm pacman-game
```

## Accessing the Application

Once running, access the application at:
- **Game UI**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

## Docker Image Features

### Multi-Stage Build
- **Builder stage**: Compiles the Go application
- **Runtime stage**: Minimal Alpine Linux image for small footprint

### Security Best Practices
- ✅ Non-root user execution
- ✅ Minimal base image (Alpine Linux)
- ✅ No unnecessary packages
- ✅ Security patches via latest Alpine
- ✅ Read-only filesystem compatible

### Optimizations
- Layer caching for dependencies
- Stripped binaries (`-ldflags="-w -s"`)
- Removed debug symbols
- Small final image size (~20MB)

### Health Checks
- Automatic health monitoring every 30 seconds
- Checks `/health` endpoint
- 3 retry attempts before marking unhealthy

## Environment Variables

Configure the application using environment variables:

```bash
# Server configuration
SERVER_PORT=8080              # Port to listen on (default: 8080)
SERVER_MODE=release           # Gin mode: debug, release (default: release)

# Logging
LOG_LEVEL=info               # Log level: debug, info, warn, error

# Observability
SERVICE_NAME=pacman-game     # Service name for tracing
SERVICE_VERSION=1.0.0        # Service version
ENVIRONMENT=production       # Environment: development, staging, production

# OpenTelemetry (optional)
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
OTEL_SERVICE_NAME=pacman-game
```

### Example with Custom Port

```bash
docker run -d \
  --name pacman-game \
  -p 9090:9090 \
  -e SERVER_PORT=9090 \
  pacman-game:latest
```

## Available Make Commands

```bash
make help              # Show all available commands

# Local development
make build             # Build binary locally
make run               # Run application locally
make test              # Run tests
make coverage          # Generate coverage report
make tidy              # Clean up dependencies
make lint              # Run linters

# Docker commands
make docker-build      # Build Docker image
make docker-run        # Build and run container
make docker-stop       # Stop container
make docker-logs       # View container logs
make docker-shell      # Open shell in container
make docker-clean      # Remove image and container

# Docker Compose commands
make compose-up        # Start with docker-compose
make compose-down      # Stop docker-compose services
make compose-logs      # View compose logs
make compose-rebuild   # Rebuild and restart

# Cleanup
make clean             # Clean build artifacts
make clean-all         # Clean everything
```

## Docker Compose Features

The `docker-compose.yml` includes:

### Main Service
- Automatic container restart
- Health checks
- Network isolation
- Environment variable configuration

### Optional Services (Commented Out)
Uncomment in `docker-compose.yml` to enable:

#### Jaeger (Distributed Tracing)
```yaml
# Uncomment the jaeger service in docker-compose.yml
```

Access Jaeger UI at http://localhost:16686

## Troubleshooting

### Container won't start
```bash
# Check logs
docker logs pacman-game

# Check if port is already in use
lsof -i :8080

# Use a different port
docker run -d -p 9090:8080 pacman-game:latest
```

### Permission issues
The container runs as non-root user (UID 1000). If you need to debug:
```bash
# Run as root
docker run --user root -it pacman-game:latest /bin/sh
```

### Health check failing
```bash
# Check health status
docker inspect pacman-game | grep -A 10 Health

# Test health endpoint manually
docker exec pacman-game wget -qO- http://localhost:8080/health
```

### Image size too large
```bash
# Check image size
docker images pacman-game

# Optimize by cleaning Docker cache
docker system prune -a
```

## Building for Different Architectures

### ARM64 (Apple Silicon, Raspberry Pi)
```bash
docker build --platform linux/arm64 -t pacman-game:arm64 .
```

### Multi-platform build
```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t pacman-game:multiarch .
```

## Production Considerations

### 1. Use Specific Tags
```bash
docker build -t pacman-game:v1.0.0 .
```

### 2. Resource Limits
```bash
docker run -d \
  --name pacman-game \
  --memory="256m" \
  --cpus="0.5" \
  -p 8080:8080 \
  pacman-game:latest
```

### 3. Logging
```bash
# Use JSON logging driver
docker run -d \
  --name pacman-game \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  -p 8080:8080 \
  pacman-game:latest
```

### 4. Read-only Filesystem
```bash
docker run -d \
  --name pacman-game \
  --read-only \
  --tmpfs /tmp \
  -p 8080:8080 \
  pacman-game:latest
```

### 5. Security Scanning
```bash
# Scan for vulnerabilities
docker scout cves pacman-game:latest

# Or use Trivy
trivy image pacman-game:latest
```

## CI/CD Integration

### GitHub Actions Example
```yaml
- name: Build Docker image
  run: docker build -t pacman-game:${{ github.sha }} .

- name: Run tests in container
  run: docker run pacman-game:${{ github.sha }} go test ./...
```

### Push to Registry
```bash
# Tag for registry
docker tag pacman-game:latest registry.example.com/pacman-game:v1.0.0

# Push to registry
docker push registry.example.com/pacman-game:v1.0.0
```

## Kubernetes Deployment

The application includes Kubernetes manifests in the `k8s/` and `helm/` directories. After building the image:

```bash
# Update image in k8s/deployment.yaml
kubectl apply -f k8s/

# Or use Helm
helm install pacman-game ./helm/todo-app
```

## Next Steps

- Configure observability with Jaeger (see `docker-compose.yml`)
- Set up persistent storage if needed
- Configure ingress/load balancer
- Implement CI/CD pipeline
- Add monitoring and alerting

## Support

For issues or questions:
1. Check logs: `docker logs pacman-game`
2. Verify health: `curl http://localhost:8080/health`
3. Review configuration in `docker-compose.yml`
4. Check application logs for errors

