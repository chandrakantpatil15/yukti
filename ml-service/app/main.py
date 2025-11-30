from fastapi import FastAPI, HTTPException, Depends, Header
from pydantic import BaseModel
from typing import List, Optional
import numpy as np
from datetime import datetime, timedelta
import redis
import json
import hashlib

app = FastAPI(title="Yukti ML Service", version="1.0.0")

# Redis for caching predictions
redis_client = redis.Redis(host='localhost', port=6379, db=0, decode_responses=True)

class CostDataPoint(BaseModel):
    date: str
    cost: float

class PredictionRequest(BaseModel):
    tenant_id: int
    historical_data: List[CostDataPoint]
    forecast_days: int = 30

class AnomalyRequest(BaseModel):
    tenant_id: int
    historical_data: List[CostDataPoint]
    threshold: float = 2.0

class RecommendationRequest(BaseModel):
    tenant_id: int
    resource_type: str
    current_cost: float
    usage_pattern: dict

def verify_token(authorization: str = Header(None)):
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Missing or invalid token")
    return authorization.replace("Bearer ", "")

@app.get("/health")
def health_check():
    return {"status": "healthy", "service": "ml-service"}

@app.post("/api/v1/ml/forecast")
def forecast_costs(request: PredictionRequest, token: str = Depends(verify_token)):
    """Cost forecasting using linear regression"""
    
    # Check cache
    cache_key = f"forecast:{request.tenant_id}:{request.forecast_days}"
    cached = redis_client.get(cache_key)
    if cached:
        return json.loads(cached)
    
    # Prepare data
    dates = [datetime.strptime(d.date, "%Y-%m-%d") for d in request.historical_data]
    costs = [d.cost for d in request.historical_data]
    
    # Simple linear regression
    X = np.array([(d - dates[0]).days for d in dates]).reshape(-1, 1)
    y = np.array(costs)
    
    # Calculate slope and intercept
    X_mean = X.mean()
    y_mean = y.mean()
    slope = ((X - X_mean) * (y - y_mean)).sum() / ((X - X_mean) ** 2).sum()
    intercept = y_mean - slope * X_mean
    
    # Forecast
    last_day = (dates[-1] - dates[0]).days
    forecast = []
    for i in range(1, request.forecast_days + 1):
        future_day = last_day + i
        predicted_cost = slope * future_day + intercept
        forecast_date = dates[-1] + timedelta(days=i)
        forecast.append({
            "date": forecast_date.strftime("%Y-%m-%d"),
            "predicted_cost": float(max(0, predicted_cost)),
            "confidence": 0.85
        })
    
    result = {
        "tenant_id": request.tenant_id,
        "forecast": forecast,
        "total_predicted_cost": sum(f["predicted_cost"] for f in forecast),
        "trend": "increasing" if slope > 0 else "decreasing",
        "model": "linear_regression"
    }
    
    # Cache for 1 hour
    redis_client.setex(cache_key, 3600, json.dumps(result))
    
    return result

@app.post("/api/v1/ml/anomaly-detect")
def detect_anomalies(request: AnomalyRequest, token: str = Depends(verify_token)):
    """Anomaly detection using statistical methods"""
    
    costs = np.array([d.cost for d in request.historical_data])
    
    # Calculate moving average and std
    window = min(7, len(costs) // 2)
    moving_avg = np.convolve(costs, np.ones(window)/window, mode='valid')
    
    # Detect anomalies (Z-score method)
    mean = costs.mean()
    std = costs.std()
    z_scores = (costs - mean) / std
    
    anomalies = []
    for i, (data_point, z_score) in enumerate(zip(request.historical_data, z_scores)):
        if abs(z_score) > request.threshold:
            anomalies.append({
                "date": data_point.date,
                "cost": data_point.cost,
                "expected_cost": float(mean),
                "deviation": float(abs(data_point.cost - mean)),
                "severity": "high" if abs(z_score) > 3 else "medium",
                "z_score": float(z_score)
            })
    
    return {
        "tenant_id": request.tenant_id,
        "anomalies": anomalies,
        "anomaly_count": len(anomalies),
        "baseline_cost": float(mean),
        "std_deviation": float(std)
    }

@app.post("/api/v1/ml/recommend")
def generate_recommendations(request: RecommendationRequest, token: str = Depends(verify_token)):
    """ML-powered optimization recommendations"""
    
    recommendations = []
    
    # Rule-based ML recommendations (can be enhanced with actual ML models)
    if request.resource_type == "ec2":
        cpu_avg = request.usage_pattern.get("cpu_avg", 50)
        memory_avg = request.usage_pattern.get("memory_avg", 50)
        
        if cpu_avg < 20 and memory_avg < 30:
            recommendations.append({
                "type": "downsize",
                "confidence": 0.92,
                "potential_savings": request.current_cost * 0.5,
                "reason": "Low CPU and memory utilization detected"
            })
        elif cpu_avg < 10:
            recommendations.append({
                "type": "terminate",
                "confidence": 0.88,
                "potential_savings": request.current_cost,
                "reason": "Extremely low utilization - consider termination"
            })
        
        if request.usage_pattern.get("spot_compatible", False):
            recommendations.append({
                "type": "spot_instance",
                "confidence": 0.85,
                "potential_savings": request.current_cost * 0.7,
                "reason": "Workload suitable for spot instances"
            })
    
    return {
        "tenant_id": request.tenant_id,
        "resource_type": request.resource_type,
        "recommendations": recommendations,
        "total_potential_savings": sum(r["potential_savings"] for r in recommendations)
    }

@app.post("/api/v1/ml/batch-predict")
def batch_predictions(requests: List[PredictionRequest], token: str = Depends(verify_token)):
    """Batch processing for multiple tenants"""
    results = []
    for req in requests:
        try:
            result = forecast_costs(req, token)
            results.append({"tenant_id": req.tenant_id, "success": True, "data": result})
        except Exception as e:
            results.append({"tenant_id": req.tenant_id, "success": False, "error": str(e)})
    
    return {"batch_results": results, "total": len(results)}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
