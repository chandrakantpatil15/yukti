package optimization

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RealtimeMonitor provides real-time cost monitoring and alerting
type RealtimeMonitor struct {
	alerts       chan CostAlert
	thresholds   map[string]AlertThreshold
	subscribers  []AlertSubscriber
	isRunning    bool
	stopChan     chan bool
	mu           sync.RWMutex
}

// NewRealtimeMonitor creates a new real-time monitor
func NewRealtimeMonitor() *RealtimeMonitor {
	return &RealtimeMonitor{
		alerts:      make(chan CostAlert, 100),
		thresholds:  make(map[string]AlertThreshold),
		subscribers: make([]AlertSubscriber, 0),
		stopChan:    make(chan bool),
	}
}

// AlertThreshold defines cost alert thresholds
type AlertThreshold struct {
	ResourceID    string  `json:"resource_id"`
	DailyCost     float64 `json:"daily_cost_threshold"`
	MonthlyCost   float64 `json:"monthly_cost_threshold"`
	PercentChange float64 `json:"percent_change_threshold"`
	Enabled       bool    `json:"enabled"`
}

// AlertSubscriber defines alert notification interface
type AlertSubscriber interface {
	NotifyAlert(alert CostAlert) error
}

// CostAlert represents a real-time cost alert
type CostAlert struct {
	ID           string    `json:"id"`
	ResourceID   string    `json:"resource_id"`
	AlertType    string    `json:"alert_type"`
	Severity     string    `json:"severity"`
	CurrentCost  float64   `json:"current_cost"`
	ThresholdCost float64  `json:"threshold_cost"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
	Acknowledged bool      `json:"acknowledged"`
}

// Start starts the real-time monitoring
func (rm *RealtimeMonitor) Start(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if rm.isRunning {
		return fmt.Errorf("monitor is already running")
	}
	
	rm.isRunning = true
	
	// Start monitoring goroutine
	go rm.monitorLoop(ctx)
	
	// Start alert processing goroutine
	go rm.processAlerts(ctx)
	
	return nil
}

// Stop stops the real-time monitoring
func (rm *RealtimeMonitor) Stop() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if !rm.isRunning {
		return fmt.Errorf("monitor is not running")
	}
	
	rm.isRunning = false
	rm.stopChan <- true
	close(rm.alerts)
	
	return nil
}

// SetThreshold sets an alert threshold for a resource
func (rm *RealtimeMonitor) SetThreshold(resourceID string, threshold AlertThreshold) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	threshold.ResourceID = resourceID
	rm.thresholds[resourceID] = threshold
}

// Subscribe adds an alert subscriber
func (rm *RealtimeMonitor) Subscribe(subscriber AlertSubscriber) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.subscribers = append(rm.subscribers, subscriber)
}

// CheckCost checks if a cost metric triggers any alerts
func (rm *RealtimeMonitor) CheckCost(metric ResourceMetric) {
	rm.mu.RLock()
	threshold, exists := rm.thresholds[metric.ResourceID]
	rm.mu.RUnlock()
	
	if !exists || !threshold.Enabled {
		return
	}
	
	// Check daily cost threshold
	if threshold.DailyCost > 0 && metric.Cost > threshold.DailyCost {
		alert := CostAlert{
			ID:           fmt.Sprintf("alert-%s-%d", metric.ResourceID, time.Now().Unix()),
			ResourceID:   metric.ResourceID,
			AlertType:    "daily_cost_exceeded",
			Severity:     rm.calculateSeverity(metric.Cost, threshold.DailyCost),
			CurrentCost:  metric.Cost,
			ThresholdCost: threshold.DailyCost,
			Message:      fmt.Sprintf("Daily cost $%.2f exceeds threshold $%.2f", metric.Cost, threshold.DailyCost),
			Timestamp:    time.Now(),
		}
		rm.sendAlert(alert)
	}
	
	// Check monthly cost threshold (extrapolated)
	monthlyCost := metric.Cost * 30
	if threshold.MonthlyCost > 0 && monthlyCost > threshold.MonthlyCost {
		alert := CostAlert{
			ID:           fmt.Sprintf("alert-%s-monthly-%d", metric.ResourceID, time.Now().Unix()),
			ResourceID:   metric.ResourceID,
			AlertType:    "monthly_cost_projection",
			Severity:     rm.calculateSeverity(monthlyCost, threshold.MonthlyCost),
			CurrentCost:  monthlyCost,
			ThresholdCost: threshold.MonthlyCost,
			Message:      fmt.Sprintf("Projected monthly cost $%.2f exceeds threshold $%.2f", monthlyCost, threshold.MonthlyCost),
			Timestamp:    time.Now(),
		}
		rm.sendAlert(alert)
	}
}

// calculateSeverity calculates alert severity
func (rm *RealtimeMonitor) calculateSeverity(current, threshold float64) string {
	ratio := current / threshold
	
	if ratio > 2.0 {
		return "critical"
	} else if ratio > 1.5 {
		return "high"
	} else if ratio > 1.2 {
		return "medium"
	}
	return "low"
}

// sendAlert sends an alert to the channel
func (rm *RealtimeMonitor) sendAlert(alert CostAlert) {
	select {
	case rm.alerts <- alert:
		// Alert sent successfully
	default:
		// Channel is full, log error (in real implementation)
		fmt.Printf("Alert channel full, dropping alert: %s\n", alert.ID)
	}
}

// monitorLoop runs the main monitoring loop
func (rm *RealtimeMonitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-rm.stopChan:
			return
		case <-ticker.C:
			rm.performPeriodicChecks()
		}
	}
}

// performPeriodicChecks performs periodic monitoring checks
func (rm *RealtimeMonitor) performPeriodicChecks() {
	// In a real implementation, this would:
	// 1. Query current resource costs from AWS APIs
	// 2. Compare against historical baselines
	// 3. Detect anomalies and trends
	// 4. Generate predictive alerts
	
	// For now, simulate some monitoring activity
	rm.checkBudgetAlerts()
	rm.checkUsageSpikes()
}

// checkBudgetAlerts checks for budget threshold violations
func (rm *RealtimeMonitor) checkBudgetAlerts() {
	// Simulate budget checking logic
	// In real implementation, this would query AWS Budgets API
}

// checkUsageSpikes checks for unusual usage spikes
func (rm *RealtimeMonitor) checkUsageSpikes() {
	// Simulate spike detection logic
	// In real implementation, this would analyze CloudWatch metrics
}

// processAlerts processes alerts from the channel
func (rm *RealtimeMonitor) processAlerts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case alert, ok := <-rm.alerts:
			if !ok {
				return // Channel closed
			}
			rm.notifySubscribers(alert)
		}
	}
}

// notifySubscribers notifies all subscribers of an alert
func (rm *RealtimeMonitor) notifySubscribers(alert CostAlert) {
	rm.mu.RLock()
	subscribers := make([]AlertSubscriber, len(rm.subscribers))
	copy(subscribers, rm.subscribers)
	rm.mu.RUnlock()
	
	for _, subscriber := range subscribers {
		go func(sub AlertSubscriber) {
			if err := sub.NotifyAlert(alert); err != nil {
				fmt.Printf("Failed to notify subscriber: %v\n", err)
			}
		}(subscriber)
	}
}

// GetActiveAlerts returns all active (unacknowledged) alerts
func (rm *RealtimeMonitor) GetActiveAlerts() []CostAlert {
	// In a real implementation, this would query a persistent store
	// For now, return empty slice
	return []CostAlert{}
}

// AcknowledgeAlert acknowledges an alert
func (rm *RealtimeMonitor) AcknowledgeAlert(alertID string) error {
	// In a real implementation, this would update the alert status
	// in a persistent store
	return nil
}

// EmailSubscriber implements AlertSubscriber for email notifications
type EmailSubscriber struct {
	Email string
}

// NotifyAlert sends email notification
func (es *EmailSubscriber) NotifyAlert(alert CostAlert) error {
	// In a real implementation, this would send an email
	fmt.Printf("EMAIL ALERT to %s: %s - %s\n", es.Email, alert.AlertType, alert.Message)
	return nil
}

// SlackSubscriber implements AlertSubscriber for Slack notifications
type SlackSubscriber struct {
	WebhookURL string
	Channel    string
}

// NotifyAlert sends Slack notification
func (ss *SlackSubscriber) NotifyAlert(alert CostAlert) error {
	// In a real implementation, this would send to Slack webhook
	fmt.Printf("SLACK ALERT to %s: %s - %s\n", ss.Channel, alert.AlertType, alert.Message)
	return nil
}

// WebhookSubscriber implements AlertSubscriber for webhook notifications
type WebhookSubscriber struct {
	URL string
}

// NotifyAlert sends webhook notification
func (ws *WebhookSubscriber) NotifyAlert(alert CostAlert) error {
	// In a real implementation, this would POST to webhook URL
	fmt.Printf("WEBHOOK ALERT to %s: %s - %s\n", ws.URL, alert.AlertType, alert.Message)
	return nil
}