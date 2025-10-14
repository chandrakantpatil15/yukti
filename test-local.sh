#!/bin/bash

echo "🚀 Testing Yukti on Minikube"

# Build and deploy
echo "📦 Building Docker image..."
eval $(minikube docker-env)
docker build -t cloud-cost-optimizer:latest .

echo "🚢 Deploying to minikube..."
kubectl apply -f k8s/deployment.yaml

echo "⏳ Waiting for deployment..."
kubectl wait --for=condition=available --timeout=300s deployment/cloud-cost-optimizer

echo "📊 Deployment status:"
kubectl get pods,svc -l app=cloud-cost-optimizer

echo "🌐 Service URL:"
minikube service cloud-cost-optimizer-service --url

echo "✅ Test complete! Use 'make logs' to view application logs"