package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EnhancedClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewEnhancedClient(baseURL string) *EnhancedClient {
	return &EnhancedClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// EnhanceFinding adds ML predictions to a finding
func (c *EnhancedClient) EnhanceFinding(ctx context.Context, finding map[string]interface{}, context map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"finding": finding,
		"context": context,
	}

	resp, err := c.post(ctx, "/enhance-finding", payload)
	if err != nil {
		return nil, err
	}

	return resp["enhanced_finding"].(map[string]interface{}), nil
}

// DetectAnomalies detects cost anomalies in timeseries
func (c *EnhancedClient) DetectAnomalies(ctx context.Context, costData []map[string]interface{}) ([]map[string]interface{}, error) {
	payload := map[string]interface{}{
		"data": costData,
	}

	resp, err := c.post(ctx, "/detect-anomalies", payload)
	if err != nil {
		return nil, err
	}

	anomalies := resp["anomalies"].([]interface{})
	result := make([]map[string]interface{}, len(anomalies))
	for i, a := range anomalies {
		result[i] = a.(map[string]interface{})
	}

	return result, nil
}

// ClassifyWorkload classifies resource workload type
func (c *EnhancedClient) ClassifyWorkload(ctx context.Context, resource map[string]interface{}) (string, error) {
	payload := map[string]interface{}{
		"resource": resource,
	}

	resp, err := c.post(ctx, "/classify-workload", payload)
	if err != nil {
		return "", err
	}

	return resp["workload_type"].(string), nil
}

// PredictDataTransfer predicts monthly data transfer costs
func (c *EnhancedClient) PredictDataTransfer(ctx context.Context, topology map[string]interface{}) (float64, error) {
	payload := map[string]interface{}{
		"topology": topology,
	}

	resp, err := c.post(ctx, "/predict-data-transfer", payload)
	if err != nil {
		return 0, err
	}

	return resp["predicted_cost"].(float64), nil
}

// TrainModels trains all ML models with historical data
func (c *EnhancedClient) TrainModels(ctx context.Context, trainingData map[string]interface{}) error {
	_, err := c.post(ctx, "/train-models", trainingData)
	return err
}

// GetModelStats returns statistics about trained models
func (c *EnhancedClient) GetModelStats(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/model-stats", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *EnhancedClient) post(ctx context.Context, endpoint string, payload interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if success, ok := result["success"].(bool); !ok || !success {
		return nil, fmt.Errorf("ML service returned error")
	}

	return result, nil
}
