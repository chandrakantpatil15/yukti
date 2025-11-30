package ml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MLClient struct {
	baseURL string
	token   string
	client  *http.Client
}

type CostDataPoint struct {
	Date string  `json:"date"`
	Cost float64 `json:"cost"`
}

type ForecastRequest struct {
	TenantID       int             `json:"tenant_id"`
	HistoricalData []CostDataPoint `json:"historical_data"`
	ForecastDays   int             `json:"forecast_days"`
}

type ForecastResponse struct {
	TenantID           int             `json:"tenant_id"`
	Forecast           []ForecastPoint `json:"forecast"`
	TotalPredictedCost float64         `json:"total_predicted_cost"`
	Trend              string          `json:"trend"`
	Model              string          `json:"model"`
}

type ForecastPoint struct {
	Date          string  `json:"date"`
	PredictedCost float64 `json:"predicted_cost"`
	Confidence    float64 `json:"confidence"`
}

type AnomalyRequest struct {
	TenantID       int             `json:"tenant_id"`
	HistoricalData []CostDataPoint `json:"historical_data"`
	Threshold      float64         `json:"threshold"`
}

type AnomalyResponse struct {
	TenantID     int       `json:"tenant_id"`
	Anomalies    []Anomaly `json:"anomalies"`
	AnomalyCount int       `json:"anomaly_count"`
	BaselineCost float64   `json:"baseline_cost"`
	StdDeviation float64   `json:"std_deviation"`
}

type Anomaly struct {
	Date         string  `json:"date"`
	Cost         float64 `json:"cost"`
	ExpectedCost float64 `json:"expected_cost"`
	Deviation    float64 `json:"deviation"`
	Severity     string  `json:"severity"`
	ZScore       float64 `json:"z_score"`
}

func NewMLClient(baseURL, token string) *MLClient {
	return &MLClient{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *MLClient) Forecast(req ForecastRequest) (*ForecastResponse, error) {
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/ml/forecast", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service error: %s", string(bodyBytes))
	}

	var result ForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *MLClient) DetectAnomalies(req AnomalyRequest) (*AnomalyResponse, error) {
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/ml/anomaly-detect", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service error: %s", string(bodyBytes))
	}

	var result AnomalyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *MLClient) HealthCheck() error {
	resp, err := c.client.Get(c.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ML service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
