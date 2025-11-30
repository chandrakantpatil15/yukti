# Ports Explained - Simple Guide

## ✅ How It Works (Simple Version)

### 1. You Define Ports in ONE File

```bash
# File: .env.ports
BACKEND_PORT=8081
FRONTEND_PORT=3000
```

### 2. Docker Compose Loads This File

```yaml
# File: docker-compose.yml
backend:
  env_file:
    - .env.ports  # ← Loads BACKEND_PORT=8081
  environment:
    PORT: ${BACKEND_PORT:-8081}  # ← Uses it here
```

### 3. Code Reads From Environment

```go
// File: cmd/main.go
port := os.Getenv("PORT")  // Gets "8081"
```

---

## 🔍 Proof It Works

### Test 1: Check .env.ports
```bash
$ cat .env.ports
BACKEND_PORT=8081
FRONTEND_PORT=3000
```

### Test 2: Check Docker Compose Reads It
```bash
$ docker-compose config | grep "PORT:"
PORT: "8081"
BACKEND_PORT: "8081"
FRONTEND_PORT: "3000"
```

### Test 3: Check Container Has It
```bash
$ docker exec yukti-backend env | grep PORT
PORT=8081
BACKEND_PORT=8081
```

### Test 4: Check Application Uses It
```bash
$ docker-compose logs backend | grep "Server starting"
[INFO] Server starting on port 8081
```

---

## 📝 To Change Ports

### Step 1: Edit .env.ports
```bash
nano .env.ports

# Change:
BACKEND_PORT=9081
FRONTEND_PORT=9000
```

### Step 2: Rebuild
```bash
docker-compose down
docker-compose up -d --build
```

### Step 3: Verify
```bash
curl http://localhost:9081/health
open http://localhost:9000
```

---

## ❓ FAQ

### Q: Where are ports hardcoded?
**A:** Only in `.env.ports` file. Nowhere else!

### Q: How does code know to read from .env.ports?
**A:** 
1. Docker Compose reads `.env.ports`
2. Sets environment variables in container
3. Code reads environment variables

### Q: What if .env.ports is missing?
**A:** Fallback values in `docker-compose.yml` are used:
```yaml
PORT: ${BACKEND_PORT:-8081}
#                    ↑ fallback
```

### Q: Can I override ports without editing .env.ports?
**A:** Yes, temporarily:
```bash
BACKEND_PORT=9081 docker-compose up -d
```

### Q: Do I need to rebuild after changing .env.ports?
**A:** Yes:
```bash
docker-compose down
docker-compose up -d --build
```

---

## 🎯 Key Points

1. **One file controls all ports**: `.env.ports`
2. **Docker Compose loads it**: `env_file: - .env.ports`
3. **Code reads from environment**: `os.Getenv("PORT")`
4. **Change once, apply everywhere**: Edit `.env.ports` → Rebuild

---

## 🚀 Quick Reference

| Action | Command |
|--------|---------|
| View ports | `cat .env.ports` |
| Change ports | `nano .env.ports` |
| Apply changes | `docker-compose up -d --build` |
| Verify | `docker-compose ps` |
| Check env vars | `docker exec yukti-backend env \| grep PORT` |
| View logs | `docker-compose logs backend` |

---

**Last Updated**: Session 13
