package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

type RIRecommendation struct {
	Service              string
	InstanceType         string
	Region               string
	Term                 string
	PaymentOption        string
	EstimatedMonthlyCost float64
	EstimatedSavings     float64
	BreakEvenMonths      int
	RecommendedQuantity  int
}

type SPRecommendation struct {
	Service              string
	Term                 string
	PaymentOption        string
	HourlyCommitment     float64
	EstimatedMonthlyCost float64
	EstimatedSavings     float64
	BreakEvenMonths      int
}

func (c *CostExplorerClient) GetRIRecommendations(ctx context.Context) ([]RIRecommendation, error) {
	input := &costexplorer.GetReservationPurchaseRecommendationInput{
		Service:              aws.String("AmazonEC2"),
		LookbackPeriodInDays: types.LookbackPeriodInDaysThirtyDays,
		TermInYears:          types.TermInYearsOneYear,
		PaymentOption:        types.PaymentOptionNoUpfront,
	}

	result, err := c.client.GetReservationPurchaseRecommendation(ctx, input)
	if err != nil {
		return nil, err
	}

	var recommendations []RIRecommendation
	for _, rec := range result.Recommendations {
		for _, detail := range rec.RecommendationDetails {
			recommendations = append(recommendations, RIRecommendation{
				Service:              "EC2",
				InstanceType:         *detail.InstanceDetails.EC2InstanceDetails.InstanceType,
				Region:               getStringValue(detail.InstanceDetails.EC2InstanceDetails.Region),
				Term:                 "1 Year",
				PaymentOption:        "No Upfront",
				EstimatedMonthlyCost: parseFloat(detail.EstimatedMonthlySavingsAmount) * -1,
				EstimatedSavings:     parseFloat(detail.EstimatedMonthlySavingsAmount),
				BreakEvenMonths:      0,
				RecommendedQuantity:  int(parseFloat(detail.RecommendedNumberOfInstancesToPurchase)),
			})
		}
	}

	return recommendations, nil
}

func (c *CostExplorerClient) GetSavingsPlansRecommendations(ctx context.Context) ([]SPRecommendation, error) {
	input := &costexplorer.GetSavingsPlansPurchaseRecommendationInput{
		SavingsPlansType:     types.SupportedSavingsPlansTypeComputeSp,
		LookbackPeriodInDays: types.LookbackPeriodInDaysThirtyDays,
		TermInYears:          types.TermInYearsOneYear,
		PaymentOption:        types.PaymentOptionNoUpfront,
	}

	result, err := c.client.GetSavingsPlansPurchaseRecommendation(ctx, input)
	if err != nil {
		return nil, err
	}

	var recommendations []SPRecommendation
	for _, rec := range result.SavingsPlansPurchaseRecommendation.SavingsPlansPurchaseRecommendationDetails {
		recommendations = append(recommendations, SPRecommendation{
			Service:              "Compute",
			Term:                 "1 Year",
			PaymentOption:        "No Upfront",
			HourlyCommitment:     parseFloat(rec.HourlyCommitmentToPurchase),
			EstimatedMonthlyCost: parseFloat(rec.EstimatedMonthlySavingsAmount) * -1,
			EstimatedSavings:     parseFloat(rec.EstimatedMonthlySavingsAmount),
			BreakEvenMonths:      0,
		})
	}

	return recommendations, nil
}

type RICoverage struct {
	Service          string
	CoveragePercent  float64
	OnDemandCost     float64
	ReservedCost     float64
	TotalCost        float64
	PotentialSavings float64
}

func (c *CostExplorerClient) GetRICoverage(ctx context.Context) (*RICoverage, error) {
	now := time.Now()
	startDate := now.AddDate(0, -1, 0).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	input := &costexplorer.GetReservationCoverageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity: types.GranularityMonthly,
	}

	result, err := c.client.GetReservationCoverage(ctx, input)
	if err != nil {
		return nil, err
	}

	coverage := &RICoverage{Service: "EC2"}
	if len(result.CoveragesByTime) > 0 && result.Total != nil {
		total := result.Total
		if total.CoverageHours != nil {
			coverage.CoveragePercent = parseFloat(total.CoverageHours.CoverageHoursPercentage)
		}
		if total.CoverageCost != nil {
			coverage.OnDemandCost = parseFloat(total.CoverageCost.OnDemandCost)
		}
		coverage.PotentialSavings = coverage.OnDemandCost * 0.3
	}

	return coverage, nil
}

type SPCoverage struct {
	Service          string
	CoveragePercent  float64
	OnDemandCost     float64
	CoveredCost      float64
	TotalCost        float64
	PotentialSavings float64
}

func (c *CostExplorerClient) GetSPCoverage(ctx context.Context) (*SPCoverage, error) {
	now := time.Now()
	startDate := now.AddDate(0, -1, 0).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	input := &costexplorer.GetSavingsPlansCoverageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity: types.GranularityMonthly,
	}

	result, err := c.client.GetSavingsPlansCoverage(ctx, input)
	if err != nil {
		return nil, err
	}

	coverage := &SPCoverage{Service: "Compute"}
	if len(result.SavingsPlansCoverages) > 0 {
		for _, cov := range result.SavingsPlansCoverages {
			if cov.Coverage != nil {
				coverage.CoveragePercent = parseFloat(cov.Coverage.CoveragePercentage)
				coverage.OnDemandCost = parseFloat(cov.Coverage.OnDemandCost)
				coverage.TotalCost = parseFloat(cov.Coverage.TotalCost)
				coverage.PotentialSavings = coverage.OnDemandCost * 0.4
			}
		}
	}

	return coverage, nil
}
