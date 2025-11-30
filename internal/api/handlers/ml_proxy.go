package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type MLProxyHandler struct {
	baseURL string
}

func NewMLProxyHandler() *MLProxyHandler {
	base := os.Getenv("ML_SERVICE_URL")
	if base == "" {
		// Default to docker compose service name and port
		base = "http://ml-service:8000"
	}
	return &MLProxyHandler{baseURL: base}
}

// Proxy for anomaly detection: POST /api/v1/ml/anomaly-detect -> ML: /detect-anomalies
func (h *MLProxyHandler) AnomalyDetect(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url := h.baseURL + "/detect-anomalies"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *MLProxyHandler) Forecast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "No forecast data available yet. Data will be available once ML service is fully integrated.",
		"data":    []interface{}{},
	})
}


