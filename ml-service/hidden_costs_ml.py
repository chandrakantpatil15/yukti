"""
ML-Enhanced Hidden Cost Detection
Provides prediction, anomaly detection, and confidence scoring for cost optimization
"""

import numpy as np
import pandas as pd
from sklearn.ensemble import IsolationForest, RandomForestRegressor
from sklearn.preprocessing import StandardScaler
from typing import Dict, List, Tuple
import joblib
from datetime import datetime, timedelta

class DataTransferPredictor:
    """Predict cross-AZ/region data transfer costs based on topology"""
    
    def __init__(self):
        self.model = RandomForestRegressor(n_estimators=100, random_state=42)
        self.scaler = StandardScaler()
        self.is_trained = False
    
    def train(self, historical_data: pd.DataFrame):
        """
        Train on historical CloudWatch metrics + Cost Explorer data
        Features: resource types, AZs, regions, network topology
        """
        features = self._extract_features(historical_data)
        target = historical_data['data_transfer_cost']
        
        X_scaled = self.scaler.fit_transform(features)
        self.model.fit(X_scaled, target)
        self.is_trained = True
    
    def predict(self, resource_topology: Dict) -> float:
        """Predict monthly data transfer cost"""
        if not self.is_trained:
            return 0.0
        
        features = self._topology_to_features(resource_topology)
        X_scaled = self.scaler.transform([features])
        predicted_cost = self.model.predict(X_scaled)[0]
        
        return max(0, predicted_cost)  # Ensure non-negative
    
    def _extract_features(self, df: pd.DataFrame) -> np.ndarray:
        """Extract features from historical data"""
        return df[['resource_count', 'cross_az_resources', 'cross_region_resources',
                   'avg_data_volume_gb', 'unique_azs', 'unique_regions']].values
    
    def _topology_to_features(self, topology: Dict) -> List[float]:
        """Convert topology to feature vector"""
        return [
            topology.get('resource_count', 0),
            topology.get('cross_az_resources', 0),
            topology.get('cross_region_resources', 0),
            topology.get('avg_data_volume_gb', 0),
            topology.get('unique_azs', 1),
            topology.get('unique_regions', 1)
        ]


class CostAnomalyDetector:
    """Detect unusual cost patterns using Isolation Forest"""
    
    def __init__(self, contamination=0.1):
        self.model = IsolationForest(contamination=contamination, random_state=42)
        self.scaler = StandardScaler()
        self.is_trained = False
    
    def train(self, cost_timeseries: pd.DataFrame):
        """Train on historical cost data"""
        features = self._extract_features(cost_timeseries)
        X_scaled = self.scaler.fit_transform(features)
        self.model.fit(X_scaled)
        self.is_trained = True
    
    def detect(self, cost_timeseries: pd.DataFrame) -> List[Dict]:
        """
        Detect anomalies in cost patterns
        Returns: List of anomalies with timestamps and scores
        """
        if not self.is_trained:
            return []
        
        features = self._extract_features(cost_timeseries)
        X_scaled = self.scaler.transform(features)
        
        # -1 for anomalies, 1 for normal
        predictions = self.model.predict(X_scaled)
        scores = self.model.score_samples(X_scaled)
        
        anomalies = []
        for i, (pred, score) in enumerate(zip(predictions, scores)):
            if pred == -1:
                anomalies.append({
                    'timestamp': cost_timeseries.iloc[i]['timestamp'],
                    'cost': cost_timeseries.iloc[i]['cost'],
                    'anomaly_score': abs(score),
                    'severity': self._score_to_severity(score),
                    'description': self._generate_description(cost_timeseries.iloc[i])
                })
        
        return anomalies
    
    def _extract_features(self, df: pd.DataFrame) -> np.ndarray:
        """Extract features from cost timeseries"""
        df = df.copy()
        df['day_of_week'] = pd.to_datetime(df['timestamp']).dt.dayofweek
        df['hour_of_day'] = pd.to_datetime(df['timestamp']).dt.hour
        df['cost_diff'] = df['cost'].diff().fillna(0)
        df['cost_pct_change'] = df['cost'].pct_change().fillna(0)
        
        return df[['cost', 'day_of_week', 'hour_of_day', 'cost_diff', 'cost_pct_change']].values
    
    def _score_to_severity(self, score: float) -> str:
        """Convert anomaly score to severity level"""
        abs_score = abs(score)
        if abs_score > 0.5:
            return 'Critical'
        elif abs_score > 0.3:
            return 'High'
        elif abs_score > 0.1:
            return 'Medium'
        return 'Low'
    
    def _generate_description(self, row: pd.Series) -> str:
        """Generate human-readable description"""
        cost = row['cost']
        timestamp = row['timestamp']
        return f"Unusual cost spike of ${cost:.2f} detected at {timestamp}"


class SavingsConfidenceEstimator:
    """Calculate confidence score for savings estimates"""
    
    def __init__(self):
        self.model = RandomForestRegressor(n_estimators=50, random_state=42)
        self.is_trained = False
    
    def train(self, historical_findings: pd.DataFrame):
        """
        Train on historical findings and actual savings realized
        Features: data quality, resource stability, finding type
        """
        features = self._extract_features(historical_findings)
        target = historical_findings['actual_savings'] / historical_findings['estimated_savings']
        target = target.clip(0, 1)  # Confidence between 0 and 1
        
        self.model.fit(features, target)
        self.is_trained = True
    
    def estimate(self, finding: Dict) -> float:
        """
        Calculate confidence score (0.0-1.0) for a finding
        Based on: data quality, historical accuracy, resource stability
        """
        if not self.is_trained:
            return self._rule_based_confidence(finding)
        
        features = self._finding_to_features(finding)
        confidence = self.model.predict([features])[0]
        
        return max(0.0, min(1.0, confidence))  # Clamp to [0, 1]
    
    def _extract_features(self, df: pd.DataFrame) -> np.ndarray:
        """Extract features from historical findings"""
        return df[['data_quality_score', 'resource_age_days', 'metric_completeness',
                   'historical_accuracy', 'resource_stability']].values
    
    def _finding_to_features(self, finding: Dict) -> List[float]:
        """Convert finding to feature vector"""
        return [
            finding.get('data_quality_score', 0.8),
            finding.get('resource_age_days', 30),
            finding.get('metric_completeness', 0.9),
            finding.get('historical_accuracy', 0.85),
            finding.get('resource_stability', 0.9)
        ]
    
    def _rule_based_confidence(self, finding: Dict) -> float:
        """Fallback rule-based confidence when model not trained"""
        base_confidence = 0.85
        
        # Adjust based on data quality
        data_quality = finding.get('data_quality_score', 0.8)
        base_confidence *= data_quality
        
        # Adjust based on resource age (newer = less confident)
        age_days = finding.get('resource_age_days', 30)
        if age_days < 7:
            base_confidence *= 0.7
        elif age_days < 30:
            base_confidence *= 0.85
        
        # Adjust based on finding type
        finding_type = finding.get('detector_name', '')
        if 'eol' in finding_type.lower():
            base_confidence = 1.0  # EOL findings are certain
        
        return max(0.0, min(1.0, base_confidence))


class WorkloadClassifier:
    """Classify workloads (production, dev, test, sandbox) using ML"""
    
    def __init__(self):
        from sklearn.ensemble import RandomForestClassifier
        self.model = RandomForestClassifier(n_estimators=100, random_state=42)
        self.is_trained = False
    
    def train(self, labeled_resources: pd.DataFrame):
        """Train on labeled resources"""
        features = self._extract_features(labeled_resources)
        target = labeled_resources['workload_type']
        
        self.model.fit(features, target)
        self.is_trained = True
    
    def classify(self, resource: Dict) -> str:
        """
        Classify resource workload type
        Features: tags, naming, usage patterns
        Output: production, dev, test, sandbox
        """
        if not self.is_trained:
            return self._rule_based_classification(resource)
        
        features = self._resource_to_features(resource)
        prediction = self.model.predict([features])[0]
        
        return prediction
    
    def _extract_features(self, df: pd.DataFrame) -> np.ndarray:
        """Extract features from resources"""
        return df[['has_prod_tag', 'has_dev_tag', 'name_contains_prod',
                   'name_contains_dev', 'uptime_hours', 'cost_monthly']].values
    
    def _resource_to_features(self, resource: Dict) -> List[float]:
        """Convert resource to feature vector"""
        tags = resource.get('tags', {})
        name = resource.get('name', '').lower()
        
        return [
            1 if 'prod' in tags.get('environment', '').lower() else 0,
            1 if 'dev' in tags.get('environment', '').lower() else 0,
            1 if 'prod' in name else 0,
            1 if 'dev' in name or 'test' in name else 0,
            resource.get('uptime_hours', 0),
            resource.get('cost_monthly', 0)
        ]
    
    def _rule_based_classification(self, resource: Dict) -> str:
        """Fallback rule-based classification"""
        tags = resource.get('tags', {})
        name = resource.get('name', '').lower()
        
        # Check tags first
        env = tags.get('environment', '').lower()
        if env in ['production', 'prod']:
            return 'production'
        elif env in ['development', 'dev']:
            return 'dev'
        elif env in ['test', 'testing', 'qa']:
            return 'test'
        elif env in ['sandbox', 'demo']:
            return 'sandbox'
        
        # Check name patterns
        if 'prod' in name:
            return 'production'
        elif 'dev' in name:
            return 'dev'
        elif 'test' in name or 'qa' in name:
            return 'test'
        elif 'sandbox' in name or 'demo' in name:
            return 'sandbox'
        
        # Default to production (conservative)
        return 'production'


class MLEnhancedDetector:
    """Main class integrating all ML models"""
    
    def __init__(self):
        self.transfer_predictor = DataTransferPredictor()
        self.anomaly_detector = CostAnomalyDetector()
        self.confidence_estimator = SavingsConfidenceEstimator()
        self.workload_classifier = WorkloadClassifier()
    
    def enhance_finding(self, finding: Dict, context: Dict) -> Dict:
        """Enhance a finding with ML predictions"""
        enhanced = finding.copy()
        
        # Add confidence score
        enhanced['confidence'] = self.confidence_estimator.estimate(finding)
        
        # Add workload classification
        if 'resource' in context:
            enhanced['workload_type'] = self.workload_classifier.classify(context['resource'])
        
        # Add predicted cost if applicable
        if finding.get('detector_name') == 'cross_az_data_transfer':
            if 'topology' in context:
                enhanced['predicted_cost'] = self.transfer_predictor.predict(context['topology'])
        
        return enhanced
    
    def detect_cost_anomalies(self, cost_timeseries: pd.DataFrame) -> List[Dict]:
        """Detect cost anomalies"""
        return self.anomaly_detector.detect(cost_timeseries)
    
    def train_all_models(self, training_data: Dict):
        """Train all ML models"""
        if 'transfer_data' in training_data:
            self.transfer_predictor.train(training_data['transfer_data'])
        
        if 'cost_timeseries' in training_data:
            self.anomaly_detector.train(training_data['cost_timeseries'])
        
        if 'historical_findings' in training_data:
            self.confidence_estimator.train(training_data['historical_findings'])
        
        if 'labeled_resources' in training_data:
            self.workload_classifier.train(training_data['labeled_resources'])
    
    def save_models(self, path: str):
        """Save trained models to disk"""
        joblib.dump({
            'transfer_predictor': self.transfer_predictor,
            'anomaly_detector': self.anomaly_detector,
            'confidence_estimator': self.confidence_estimator,
            'workload_classifier': self.workload_classifier
        }, path)
    
    def load_models(self, path: str):
        """Load trained models from disk"""
        models = joblib.load(path)
        self.transfer_predictor = models['transfer_predictor']
        self.anomaly_detector = models['anomaly_detector']
        self.confidence_estimator = models['confidence_estimator']
        self.workload_classifier = models['workload_classifier']
