# 🐳 Docker Quick Start Guide

## TL;DR - Run Your Application Now

```bash
# Option 1: Using Make (easiest)
make docker-run

# Option 2: Using Docker Compose
docker-compose up -d

# Option 3: Using Docker CLI
docker build -t pacman-game:latest .
docker run -d -p 8080:8080 --name pacman-game pacman-game:latest
```

Then open your browser: **http://localhost:8080**

---

## What Was Changed

### 1. ✅ Updated Dockerfile
**Location**: `Dockerfile`

**Key Improvements**:
- ✨ Builds from `cmd/server/` (Clean Architecture entry point)
- 🔒 Runs as non-root user for security
- 📦 Multi-stage build for minimal image size (~20MB)
- 🏥 Built-in health checks
- 🚀 Optimized build flags for smaller, faster binaries
- 📁 Includes static files for web frontend

**Before vs After**:
```diff
- RUN go build -o main .              # Built wrong entry point
+ RUN go build -o server ./cmd/server # Builds correct entry point

- CMD ["./main"]                      # Ran as root
+ USER appuser                        # Runs as non-root
+ CMD ["./server"]                    # Secure execution
```

### 2. 🆕 Added .dockerignore
**Location**: `.dockerignore`

**Benefits**:
- Faster builds (excludes unnecessary files)
- Smaller build context
- Better security (excludes sensitive files)

### 3. 🆕 Added docker-compose.yml
**Location**: `docker-compose.yml`

**Features**:
- One-command startup
- Environment variable configuration
- Health monitoring
- Optional Jaeger integration for tracing
- Network isolation

### 4. 🆕 Added Makefile
**Location**: `Makefile`

**Convenience Commands**:
```bash
make help              # Show all commands
make docker-run        # Build and run
make docker-logs       # View logs
make docker-stop       # Stop container
make compose-up        # Start with compose
make test              # Run tests
make clean-all         # Clean everything
```

### 5. 📚 Added Documentation
**Locations**: 
- `DOCKER.md` - Comprehensive Docker guide
- `DOCKER_QUICK_START.md` - This file!

---

## Quick Commands

### Starting the Application
```bash
# Method 1: Make (recommended)
make docker-run

# Method 2: Docker Compose
docker-compose up -d

# Method 3: Docker CLI
docker build -t pacman-game:latest .
docker run -d -p 8080:8080 --name pacman-game pacman-game:latest
```

### Viewing Logs
```bash
# Make
make docker-logs

# Docker Compose
docker-compose logs -f

# Docker CLI
docker logs -f pacman-game
```

### Stopping the Application
```bash
# Make
make docker-stop

# Docker Compose
docker-compose down

# Docker CLI
docker stop pacman-game && docker rm pacman-game
```

### Checking Health
```bash
# Check health endpoint
curl http://localhost:8080/health

# Check Docker health status
docker inspect pacman-game | grep -A 10 Health
```

---

## Environment Variables

Customize the application by setting environment variables:

```bash
docker run -d \
  -p 8080:8080 \
  -e PORT=8080 \
  -e GIN_MODE=release \
  -e LOG_LEVEL=info \
  -e SERVICE_NAME=pacman-game \
  -e ENVIRONMENT=production \
  --name pacman-game \
  pacman-game:latest
```

**Available Variables**:
- `PORT` - Server port (default: 8080)
- `GIN_MODE` - Gin mode: debug/release (default: release)
- `LOG_LEVEL` - Log level: debug/info/warn/error (default: info)
- `LOG_FORMAT` - Log format: json/text (default: json)
- `SERVICE_NAME` - Service name for tracing (default: pacman-game)
- `SERVICE_VERSION` - Service version (default: 1.0.0)
- `ENVIRONMENT` - Environment: development/staging/production (default: development)
- `TRACING_ENABLED` - Enable tracing (default: true)
- `METRICS_ENABLED` - Enable metrics (default: true)

---

## Testing Your Docker Setup

### 1. Build the Image
```bash
docker build -t pacman-game:latest .
```

**Expected output**:
```
[+] Building 45.2s (18/18) FINISHED
...
=> exporting to image
=> => naming to docker.io/library/pacman-game:latest
```

### 2. Run the Container
```bash
docker run -d -p 8080:8080 --name pacman-game pacman-game:latest
```

**Verify it's running**:
```bash
docker ps | grep pacman-game
```

### 3. Test the Application
```bash
# Health check
curl http://localhost:8080/health

# Expected: {"status":"ok","service":"pacman-game"}

# Start a game
curl -X POST http://localhost:8080/api/game/start

# Open in browser
open http://localhost:8080
```

### 4. View Logs
```bash
docker logs pacman-game
```

**Expected output**:
```json
{"level":"info","time":"...","message":"starting pacman game server"}
{"level":"info","time":"...","message":"server listening","port":"8080"}
```

---

## Troubleshooting

### Problem: Port already in use
```bash
# Check what's using port 8080
lsof -i :8080

# Use a different port
docker run -d -p 9090:8080 --name pacman-game pacman-game:latest
```

### Problem: Container exits immediately
```bash
# Check logs for errors
docker logs pacman-game

# Run interactively to debug
docker run -it --rm pacman-game:latest /bin/sh
```

### Problem: Can't access the application
```bash
# Verify container is running
docker ps

# Check port mapping
docker port pacman-game

# Test from inside container
docker exec pacman-game wget -qO- http://localhost:8080/health
```

### Problem: Build fails
```bash
# Clean Docker cache
docker builder prune -a

# Build with no cache
docker build --no-cache -t pacman-game:latest .
```

---

## Next Steps

1. **Development**:
   - Modify code
   - Rebuild: `make docker-build`
   - Restart: `make docker-run`

2. **Add Observability**:
   - Uncomment Jaeger in `docker-compose.yml`
   - Access Jaeger UI: http://localhost:16686

3. **Production Deployment**:
   - Push to container registry
   - Deploy to Kubernetes (see `k8s/` and `helm/` directories)
   - Configure monitoring and alerting

4. **CI/CD**:
   - Integrate Docker build in your pipeline
   - Add automated testing
   - Set up automated deployments

---

## Performance Characteristics

- **Build Time**: ~30-60 seconds (first build)
- **Image Size**: ~20MB (optimized)
- **Startup Time**: <2 seconds
- **Memory Usage**: ~10-20MB at idle
- **CPU Usage**: Minimal (<1% at idle)

---

## Architecture

```
┌─────────────────────────────────────────┐
│         Docker Container                │
│                                         │
│  ┌────────────────────────────────┐   │
│  │    Alpine Linux (Base)         │   │
│  │                                 │   │
│  │  ┌──────────────────────────┐  │   │
│  │  │   Go Application         │  │   │
│  │  │   (cmd/server)           │  │   │
│  │  │                          │  │   │
│  │  │  ├─ HTTP Server (Gin)   │  │   │
│  │  │  ├─ Game Logic          │  │   │
│  │  │  ├─ Static Files        │  │   │
│  │  │  └─ Health Checks       │  │   │
│  │  └──────────────────────────┘  │   │
│  │                                 │   │
│  │  Port: 8080                     │   │
│  │  User: appuser (non-root)      │   │
│  └────────────────────────────────┘   │
└─────────────────────────────────────────┘
              │
              ↓
        Host Port 8080
```

---

## Security Features

✅ **Non-root User**: Runs as UID 1000  
✅ **Minimal Attack Surface**: Alpine Linux base  
✅ **No Unnecessary Packages**: Only essential tools  
✅ **Health Monitoring**: Automatic health checks  
✅ **Stripped Binaries**: No debug symbols  
✅ **Read-only Compatible**: Can run with read-only filesystem  

---

## Support

For detailed information, see:
- **Comprehensive Guide**: `DOCKER.md`
- **Architecture**: `ARCHITECTURE.md`
- **Main README**: `README.md`

For issues:
1. Check logs: `docker logs pacman-game`
2. Verify health: `curl http://localhost:8080/health`
3. Test endpoints: See API documentation

---

**Happy Dockerizing! 🚀**

