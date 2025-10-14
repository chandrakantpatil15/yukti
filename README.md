# Yukti - Cloud Cost Optimizer

A reactive Spring Boot application for AWS cost optimization using plugin architecture.

## Quick Start
```bash
# Local development
make local

# Build Docker image
make docker-build

# Deploy to Kubernetes
make k8s-deploy

# View logs
make logs
```

## Architecture
- **Framework**: Spring Boot 3.2.0 with WebFlux
- **Java**: 17 with Virtual Threads
- **Monitoring**: Prometheus + Grafana
- **Deployment**: Kubernetes with HPA

## Project Structure
```
src/main/java/com/cloudcostoptimizer/
├── core/          # Framework components
├── plugins/       # Service-specific optimizers
└── web/          # REST controllers
```

## Development
- Use `.prompts/` for project context
- Follow reactive programming patterns
- Implement new services as plugins