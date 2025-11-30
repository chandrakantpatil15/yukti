# Yukti - Cloud Cost Optimizer Project Context

## Project Overview
- **Name**: Cloud Cost Optimizer (Yukti)
- **Type**: Spring Boot WebFlux Application
- **Java Version**: 17 with Virtual Threads
- **Framework**: Spring Boot 3.2.0
- **Architecture**: Reactive, Plugin-based

## Tech Stack
- Spring Boot WebFlux (Reactive)
- Spring Boot Actuator (Monitoring)
- Micrometer + Prometheus (Metrics)
- Maven (Build Tool)
- Docker (Containerization)
- Kubernetes (Orchestration)

## Current Architecture
```
src/main/java/com/cloudcostoptimizer/
├── core/
│   ├── aws/          # AWS service integrations
│   ├── config/       # Configuration classes
│   └── plugin/       # Plugin framework
├── plugins/
│   └── ec2/          # EC2 cost optimization plugin
├── web/
│   └── controller/   # REST controllers
└── CloudCostOptimizerApplication.java
```

## Infrastructure Setup
- **Kubernetes**: Complete deployment with HPA, monitoring
- **Monitoring**: Prometheus + Grafana stack
- **Docker**: Multi-stage build with Alpine Linux
- **Makefile**: Build automation commands

## Current Progress
- ✅ Basic Spring Boot application structure
- ✅ Health controller endpoint (/api/health)
- ✅ Kubernetes deployment configuration (HPA enabled)
- ✅ Monitoring stack (Prometheus/Grafana)
- ✅ Docker containerization (multi-stage Alpine build)
- ✅ Application configuration (Virtual threads, WebFlux)
- ✅ Build automation (Makefile with local/docker/k8s commands)
- 🔄 Plugin architecture (directories created, ready for implementation)
- ⏳ AWS cost optimization plugins
- ⏳ Core business logic implementation

## Key Features Planned
- Plugin-based architecture for different cloud services
- Real-time cost monitoring and optimization
- Reactive programming for high performance
- Kubernetes-native deployment
- Comprehensive monitoring and alerting

## Current Implementation Status
- **Application**: Spring Boot WebFlux app running on port 8080
- **Endpoints**: Health check at /api/health with detailed status
- **Configuration**: Virtual threads enabled, reactive base path /api
- **Monitoring**: Prometheus metrics exposed, health details enabled
- **Deployment**: Ready for Kubernetes with HPA scaling
- **Next Phase**: Plugin framework and AWS integration

## Development Workflow
- Local development: `make local`
- Docker build: `make docker-build`
- K8s deployment: `make k8s-deploy`
- Logs: `make logs`