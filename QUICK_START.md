# 🚀 Yukti FinOps - Quick Start Guide

## **One Command Setup**

```bash
make start-all
```

That's it! This will:
- ✅ Sync AWS data
- ✅ Start API server (port 8080)
- ✅ Start ML service (port 8000)  
- ✅ Start React UI (port 3000)

## **Access Your Application**

- **📱 Main UI**: http://localhost:3000
- **🔧 API**: http://localhost:8080
- **🤖 ML Service**: http://localhost:8000

## **Common Commands**

```bash
# Check what's running
make status

# Stop everything
make stop-all

# Just start API
make start-api

# Just start UI
make start-ui

# Sync fresh AWS data
make sync-data

# Run tests
make test
```

## **First Time Setup**

```bash
# Setup database (one time only)
make setup

# Then start everything
make start-all
```

## **Troubleshooting**

```bash
# Clean and restart
make stop-all
make clean
make start-all

# Check service status
make status

# Test API endpoints
make test-api
```

## **Development Shortcuts**

```bash
make dev        # Same as start-all
make local      # Same as start-all  
make full-stack # Same as start-all
```

**That's it! Your FinOps platform is ready in one command!** 🎯