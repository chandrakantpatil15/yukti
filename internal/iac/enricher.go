package iac

import (
	"yukti/internal/hiddencosts"
)

type Enricher struct {
	terraformGen      *TerraformGenerator
	cloudFormationGen *CloudFormationGenerator
}

func NewEnricher(region string) *Enricher {
	return &Enricher{
		terraformGen:      NewTerraformGenerator("aws", region),
		cloudFormationGen: NewCloudFormationGenerator(region),
	}
}

func (e *Enricher) EnrichFindings(findings []hiddencosts.Finding, useTerraform bool) []hiddencosts.Finding {
	for i := range findings {
		recommendation := &OptimizationRecommendation{
			ResourceID:       findings[i].ResourceARN,
			Action:           mapDetectorToAction(findings[i].DetectorName),
			EstimatedSavings: findings[i].EstimatedSavings,
			Confidence:       findings[i].Confidence,
			Reasoning:        findings[i].Recommendation,
		}

		var script *IaCScript
		if useTerraform {
			script = e.terraformGen.GenerateEC2Optimization(recommendation)
		} else {
			script = e.cloudFormationGen.GenerateEC2Optimization(recommendation)
		}

		findings[i].IaCCode = script.Code
	}

	return findings
}

func (e *Enricher) EnrichFinding(finding *hiddencosts.Finding, useTerraform bool) error {
	recommendation := &OptimizationRecommendation{
		ResourceID:       finding.ResourceARN,
		Action:           mapDetectorToAction(finding.DetectorName),
		EstimatedSavings: finding.EstimatedSavings,
		Confidence:       finding.Confidence,
		Reasoning:        finding.Recommendation,
	}

	var script *IaCScript
	if useTerraform {
		script = e.terraformGen.GenerateEC2Optimization(recommendation)
	} else {
		script = e.cloudFormationGen.GenerateEC2Optimization(recommendation)
	}

	finding.IaCCode = script.Code
	return nil
}

func mapDetectorToAction(detectorName string) string {
	actionMap := map[string]string{
		"ebs_gp2_vs_gp3":              "downsize",
		"ec2_burstable_t2_vs_t3":      "downsize",
		"idle_load_balancer":          "terminate",
		"unattached_eip":              "terminate",
		"detailed_monitoring_waste":   "terminate",
		"k8s_spot_opportunity":        "spot_conversion",
		"spot_instance_opportunity":   "spot_conversion",
		"ec2_previous_gen":            "downsize",
	}

	if action, exists := actionMap[detectorName]; exists {
		return action
	}
	return "downsize"
}
