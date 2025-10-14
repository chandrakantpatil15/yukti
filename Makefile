# Yukti - Cloud Cost Optimizer Makefile

.PHONY: local docker-build k8s-deploy logs monitoring-deploy monitoring-clean

# Local development
local:
	mvn spring-boot:run

# Build Docker image with cached layers
docker-build:
	mvn clean package -DskipTests
	docker build -t cloud-cost-optimizer:latest --build-arg BUILDKIT_INLINE_CACHE=1 .

# Deploy app to Kubernetes
k8s-deploy: docker-build
	kubectl apply -f k8s/deployment.yaml

# Deploy monitoring stack (Prometheus only)
monitoring-deploy:
	kubectl apply -f k8s/monitoring/namespace.yml
	kubectl apply -f k8s/monitoring/rbac.yml
	kubectl apply -f k8s/monitoring/prometheus-config.yml
	kubectl apply -f k8s/monitoring/prometheus.yml
	@echo "Monitoring stack deployed!"
	@echo "Prometheus: http://localhost:$$(kubectl get svc prometheus -n monitoring -o jsonpath='{.spec.ports[0].nodePort}')"

# Clean monitoring stack
monitoring-clean:
	kubectl delete namespace monitoring --ignore-not-found=true

# View application logs
logs:
	kubectl logs -l app=yukti -f

# View monitoring logs
monitoring-logs:
	kubectl logs -l app=prometheus -n monitoring -f

# Port forward services
port-forward:
	kubectl port-forward svc/cloud-cost-optimizer-service 8090:8090 &
	kubectl port-forward svc/grafana -n monitoring 3000:3000 &
	kubectl port-forward svc/prometheus -n monitoring 9090:9090 &

# Docker Compose commands
docker-up:
	docker-compose up -d
	@echo "✅ Services started!"
	@echo "📊 Yukti App: http://localhost:8081"
	@echo "📊 Dashboard: http://localhost:8081/dashboard"
	@echo "📈 Prometheus: http://localhost:9090"

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f yukti-app

# Get service URLs
urls:
	@echo "App: http://localhost:$$(kubectl get svc cloud-cost-optimizer-service -o jsonpath='{.spec.ports[0].nodePort}')"
	@echo "Dashboard: http://localhost:$$(kubectl get svc cloud-cost-optimizer-service -o jsonpath='{.spec.ports[0].nodePort}')/dashboard"
	@echo "Prometheus: http://localhost:$$(kubectl get svc prometheus -n monitoring -o jsonpath='{.spec.ports[0].nodePort}')"