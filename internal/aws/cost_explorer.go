package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

type CostExplorerClient struct {
	client *costexplorer.Client
}

func NewCostExplorerClient(cfg aws.Config) *CostExplorerClient {
	return &CostExplorerClient{
		client: costexplorer.NewFromConfig(cfg),
	}
}

type CostData struct {
	StartDate    string
	EndDate      string
	TotalCost    float64
	ServiceCosts map[string]float64
	DailyCosts   []DailyCost
}

type DailyCost struct {
	Date string
	Cost float64
}

func (c *CostExplorerClient) GetMonthlyCost(ctx context.Context, startDate, endDate string) (*CostData, error) {
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	}

	result, err := c.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, err
	}

	costData := &CostData{
		StartDate:    startDate,
		EndDate:      endDate,
		ServiceCosts: make(map[string]float64),
		DailyCosts:   []DailyCost{},
	}

	for _, resultByTime := range result.ResultsByTime {
		dailyCost := 0.0
		for _, group := range resultByTime.Groups {
			service := ""
			if len(group.Keys) > 0 {
				service = group.Keys[0]
			}
			cost := parseFloat(group.Metrics["UnblendedCost"].Amount)
			costData.ServiceCosts[service] += cost
			dailyCost += cost
		}
		costData.DailyCosts = append(costData.DailyCosts, DailyCost{
			Date: *resultByTime.TimePeriod.Start,
			Cost: dailyCost,
		})
		costData.TotalCost += dailyCost
	}

	return costData, nil
}

func (c *CostExplorerClient) GetCostForecast(ctx context.Context, startDate, endDate string) (float64, error) {
	input := &costexplorer.GetCostForecastInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Metric:      types.MetricUnblendedCost,
		Granularity: types.GranularityMonthly,
	}

	result, err := c.client.GetCostForecast(ctx, input)
	if err != nil {
		return 0, err
	}

	return parseFloat(result.Total.Amount), nil
}

func (c *CostExplorerClient) GetCostAnomaly(ctx context.Context) ([]CostAnomaly, error) {
	now := time.Now()
	startDate := now.AddDate(0, -1, 0).Format("2006-01-02")

	input := &costexplorer.GetAnomaliesInput{
		DateInterval: &types.AnomalyDateInterval{
			StartDate: aws.String(startDate),
		},
	}

	result, err := c.client.GetAnomalies(ctx, input)
	if err != nil {
		return nil, err
	}

	var anomalies []CostAnomaly
	for _, anomaly := range result.Anomalies {
		impact := 0.0
		if anomaly.Impact != nil {
			impact = anomaly.Impact.TotalImpact
		}
		anomalies = append(anomalies, CostAnomaly{
			AnomalyID: *anomaly.AnomalyId,
			Service:   *anomaly.DimensionValue,
			Impact:    impact,
			StartDate: *anomaly.AnomalyStartDate,
			EndDate:   getStringValue(anomaly.AnomalyEndDate),
			Feedback:  string(anomaly.Feedback),
		})
	}

	return anomalies, nil
}

type CostAnomaly struct {
	AnomalyID string
	Service   string
	Impact    float64
	StartDate string
	EndDate   string
	Feedback  string
}

func parseFloat(s *string) float64 {
	if s == nil {
		return 0
	}
	var f float64
	fmt.Sscanf(*s, "%f", &f)
	return f
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
