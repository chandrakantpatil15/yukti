"""
FastAPI endpoints for ML-enhanced cost detection
"""

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict, List, Optional
import pandas as pd
from hidden_costs_ml import MLEnhancedDetector

app = FastAPI(title="Yukti ML Service", version="2.0")

# Initialize ML detector
ml_detector = MLEnhancedDetector()

# Request/Response Models
class Finding(BaseModel):
    detector_name: str
    resource_arn: str
    estimated_cost: float
    estimated_savings: float
    data_quality_score: Optional[float] = 0.8
    resource_age_days: Optional[int] = 30

class EnhanceFindingRequest(BaseModel):
    finding: Finding
    context: Dict

class CostTimeseriesRequest(BaseModel):
    data: List[Dict]  # [{timestamp, cost}, ...]

class ResourceClassificationRequest(BaseModel):
    resource: Dict

class DataTransferPredictionRequest(BaseModel):
    topology: Dict

class TrainingDataRequest(BaseModel):
    transfer_data: Optional[List[Dict]] = None
    cost_timeseries: Optional[List[Dict]] = None
    historical_findings: Optional[List[Dict]] = None
    labeled_resources: Optional[List[Dict]] = None


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "ml-enhanced-detection",
        "version": "2.0",
        "models_trained": {
            "transfer_predictor": ml_detector.transfer_predictor.is_trained,
            "anomaly_detector": ml_detector.anomaly_detector.is_trained,
            "confidence_estimator": ml_detector.confidence_estimator.is_trained,
            "workload_classifier": ml_detector.workload_classifier.is_trained
        }
    }


@app.post("/enhance-finding")
async def enhance_finding(request: EnhanceFindingRequest):
    """
    Enhance a finding with ML predictions
    Adds confidence score, workload classification, predicted costs
    """
    try:
        finding_dict = request.finding.dict()
        enhanced = ml_detector.enhance_finding(finding_dict, request.context)
        
        return {
            "success": True,
            "enhanced_finding": enhanced
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/detect-anomalies")
async def detect_anomalies(request: CostTimeseriesRequest):
    """
    Detect cost anomalies in timeseries data
    Returns list of anomalies with timestamps and severity
    """
    try:
        df = pd.DataFrame(request.data)
        anomalies = ml_detector.detect_cost_anomalies(df)
        
        return {
            "success": True,
            "anomalies": anomalies,
            "total_anomalies": len(anomalies)
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/classify-workload")
async def classify_workload(request: ResourceClassificationRequest):
    """
    Classify resource workload type
    Returns: production, dev, test, or sandbox
    """
    try:
        workload_type = ml_detector.workload_classifier.classify(request.resource)
        
        return {
            "success": True,
            "workload_type": workload_type,
            "resource_arn": request.resource.get('arn', 'unknown')
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/predict-data-transfer")
async def predict_data_transfer(request: DataTransferPredictionRequest):
    """
    Predict monthly data transfer costs based on topology
    """
    try:
        predicted_cost = ml_detector.transfer_predictor.predict(request.topology)
        
        return {
            "success": True,
            "predicted_cost": round(predicted_cost, 2),
            "confidence": 0.85 if ml_detector.transfer_predictor.is_trained else 0.5
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/train-models")
async def train_models(request: TrainingDataRequest):
    """
    Train all ML models with provided data
    This should be called periodically with historical data
    """
    try:
        training_data = {}
        
        if request.transfer_data:
            training_data['transfer_data'] = pd.DataFrame(request.transfer_data)
        
        if request.cost_timeseries:
            training_data['cost_timeseries'] = pd.DataFrame(request.cost_timeseries)
        
        if request.historical_findings:
            training_data['historical_findings'] = pd.DataFrame(request.historical_findings)
        
        if request.labeled_resources:
            training_data['labeled_resources'] = pd.DataFrame(request.labeled_resources)
        
        ml_detector.train_all_models(training_data)
        
        # Save models
        ml_detector.save_models('/app/models/ml_models.pkl')
        
        return {
            "success": True,
            "message": "Models trained successfully",
            "models_trained": {
                "transfer_predictor": ml_detector.transfer_predictor.is_trained,
                "anomaly_detector": ml_detector.anomaly_detector.is_trained,
                "confidence_estimator": ml_detector.confidence_estimator.is_trained,
                "workload_classifier": ml_detector.workload_classifier.is_trained
            }
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/model-stats")
async def get_model_stats():
    """Get statistics about trained models"""
    return {
        "transfer_predictor": {
            "trained": ml_detector.transfer_predictor.is_trained,
            "type": "RandomForestRegressor",
            "features": 6
        },
        "anomaly_detector": {
            "trained": ml_detector.anomaly_detector.is_trained,
            "type": "IsolationForest",
            "contamination": 0.1
        },
        "confidence_estimator": {
            "trained": ml_detector.confidence_estimator.is_trained,
            "type": "RandomForestRegressor",
            "features": 5
        },
        "workload_classifier": {
            "trained": ml_detector.workload_classifier.is_trained,
            "type": "RandomForestClassifier",
            "classes": ["production", "dev", "test", "sandbox"]
        }
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
