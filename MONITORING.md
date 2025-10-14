# 🚀 Yukti Monitoring Setup

## ✅ Successfully Deployed!

### 📊 Your Custom Dashboard
- **URL**: http://localhost:32511/dashboard
- **Features**: Real-time JVM metrics, CPU usage, HTTP requests
- **Time Ranges**: 1m, 5m, 15m, 1h, 6h, 1d, 1w
- **Auto-refresh**: Every 5 seconds

### 📈 Prometheus Metrics
- **URL**: http://localhost:31417
- **Raw metrics**: http://localhost:32511/actuator/prometheus

### 🎛️ Kubernetes Dashboard
- **URL**: Run `kubectl proxy` then visit http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/
- **Token**: `eyJhbGciOiJSUzI1NiIsImtpZCI6IjZkUEVrckEwcFFNblJ2cHNfbml1UDgxOHhkQ2xRWmYxWmI1ZHdseEUxWDgifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiXSwiZXhwIjoxNzYwMzQ2MTQ1LCJpYXQiOjE3NjAzNDI1NDUsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiY2JmYmFjZWQtNGFiOS00OWIwLTgwOTQtYjQyNDIzYTg4MjBmIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJrdWJlcm5ldGVzLWRhc2hib2FyZCIsInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJhZG1pbi11c2VyIiwidWlkIjoiN2QyOWM5NGYtNzk0ZS00YjcyLWE4ZTAtMWRiODEyMWI4Y2Y4In19LCJuYmYiOjE3NjAzNDI1NDUsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDprdWJlcm5ldGVzLWRhc2hib2FyZDphZG1pbi11c2VyIn0.G5D-ibg_D4upAJ0mRiq74rTGQ6Xjw9fLvSrTBXQh0yZg2nciGmdHXrGaq7oQVZaRRb-zZfCpn6YCdzCKWTnVYU3cAGJvy-VjgiM9Enr-x8Yfq2Vy7L6ZDyK6SLjdk4pb0ldeWOalffhpEb4mWijdLo490hRaPZxjJv1Pm50CwtrcmpkCVweHYujk0OJMR6NtXP2Pi7dRouYE6Bq_qsl-b6jl17JHwP9zxaiTimu2-3Q5zHel_3JizLZULONB0hBHdPS2YxH76op4MdEoerC8YB17qlQPQ5osiC0WFtzmpNnDuYuED1vf9bYjet3w9cnL7aQ8c_8xEZvQH4GmByGa1w`

## 🎯 Quick Access Commands

```bash
# View all services
make urls

# View logs
make logs

# Access Kubernetes dashboard
kubectl proxy

# Port forward services locally
make port-forward
```

## 📊 What You're Monitoring

### JVM Metrics
- Heap memory usage and utilization
- Non-heap memory usage
- Garbage collection stats

### System Metrics  
- CPU usage percentage
- Available processors
- Process-level metrics

### Application Metrics
- HTTP request counts
- Response times
- Health status

## 💡 Resource Usage
- **Yukti App**: 256Mi-512Mi memory, 250m-500m CPU
- **Prometheus**: 512Mi-1Gi memory, 200m-500m CPU
- **Total**: ~1.5GB memory, ~1 CPU core

## 🔧 Troubleshooting

If pods aren't starting:
```bash
kubectl get pods -A
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

## 🎉 Success!
Your Yukti app is now running with comprehensive monitoring on Minikube!