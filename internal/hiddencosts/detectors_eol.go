package hiddencosts

import "time"

// End-of-Life (EOL) Software Detectors - Critical for security and compliance


type EC2EOLDetector struct{}

func (d *EC2EOLDetector) Name() string { return "ec2_eol_os" }
func (d *EC2EOLDetector) Category() Category { return CategoryEOL }

func (d *EC2EOLDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	eolOS := map[string]string{
		"Windows Server 2012":     "2023-10-10",
		"Windows Server 2012 R2":  "2023-10-10",
		"Amazon Linux 1":          "2023-12-31",
		"Ubuntu 16.04":            "2021-04-30",
		"CentOS 7":                "2024-06-30",
		"RHEL 7":                  "2024-06-30",
	}

	for _, r := range resources {
		if r.Type == "ec2" {
			osNameVal, ok := r.Metadata["os_name"]
			if !ok || osNameVal == nil {
				continue
			}
			osName, ok := osNameVal.(string)
			if !ok {
				continue
			}
			if eolDate, exists := eolOS[osName]; exists {
				eol, _ := time.Parse("2006-01-02", eolDate)
				if time.Now().After(eol) {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityCritical,
						Title:            "EC2 running end-of-life operating system",
						Description:      osName + " reached EOL on " + eolDate + " - no security patches available",
						ResourceARN:      r.ARN,
						EstimatedCost:    0,
						EstimatedSavings: 0,
						Confidence:       1.0,
						Recommendation:   "Upgrade to supported OS version immediately (security risk)",
					})
				}
			}
		}
	}
	return findings, nil
}

type RDSEOLDetector struct{}

func (d *RDSEOLDetector) Name() string { return "rds_eol_engine" }
func (d *RDSEOLDetector) Category() Category { return CategoryEOL }

func (d *RDSEOLDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	eolEngines := map[string]string{
		"mysql-5.6":      "2021-02-05",
		"mysql-5.7":      "2023-10-31",
		"postgres-9.6":   "2021-11-11",
		"postgres-10":    "2022-11-10",
		"postgres-11":    "2023-11-09",
		"mariadb-10.3":   "2023-05-25",
		"oracle-se1-11":  "2020-12-31",
	}

	for _, r := range resources {
		if r.Type == "rds" {
			engineVal, ok := r.Metadata["engine"]
			if !ok || engineVal == nil {
				continue
			}
			engine, ok := engineVal.(string)
			if !ok {
				continue
			}
			versionVal, ok := r.Metadata["engine_version"]
			if !ok || versionVal == nil {
				continue
			}
			version, ok := versionVal.(string)
			if !ok {
				continue
			}
			engineKey := engine + "-" + version

			if eolDate, exists := eolEngines[engineKey]; exists {
				eol, _ := time.Parse("2006-01-02", eolDate)
				if time.Now().After(eol) {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityCritical,
						Title:            "RDS running end-of-life database engine",
						Description:      engineKey + " reached EOL on " + eolDate + " - security vulnerabilities",
						ResourceARN:      r.ARN,
						EstimatedCost:    0,
						EstimatedSavings: 0,
						Confidence:       1.0,
						Recommendation:   "Upgrade to supported engine version (compliance requirement)",
					})
				}
			}
		}
	}
	return findings, nil
}

type LambdaEOLRuntimeDetector struct{}

func (d *LambdaEOLRuntimeDetector) Name() string { return "lambda_eol_runtime" }
func (d *LambdaEOLRuntimeDetector) Category() Category { return CategoryEOL }

func (d *LambdaEOLRuntimeDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	eolRuntimes := map[string]string{
		"python3.6":  "2022-07-18",
		"python3.7":  "2023-11-27",
		"nodejs12.x": "2023-03-31",
		"nodejs14.x": "2023-11-27",
		"dotnet3.1":  "2023-04-03",
		"ruby2.7":    "2023-12-07",
		"java8":      "2024-01-08",
	}

	for _, r := range resources {
		if r.Type == "lambda" {
			runtimeVal, ok := r.Metadata["runtime"]
			if !ok || runtimeVal == nil {
				continue
			}
			runtime, ok := runtimeVal.(string)
			if !ok {
				continue
			}
			if eolDate, exists := eolRuntimes[runtime]; exists {
				eol, _ := time.Parse("2006-01-02", eolDate)
				if time.Now().After(eol) {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityCritical,
						Title:            "Lambda using deprecated runtime",
						Description:      runtime + " reached EOL on " + eolDate + " - function will stop working",
						ResourceARN:      r.ARN,
						EstimatedCost:    0,
						EstimatedSavings: 0,
						Confidence:       1.0,
						Recommendation:   "Upgrade to supported runtime immediately (function will be disabled)",
					})
				}
			}
		}
	}
	return findings, nil
}

type EKSEOLVersionDetector struct{}

func (d *EKSEOLVersionDetector) Name() string { return "eks_eol_version" }
func (d *EKSEOLVersionDetector) Category() Category { return CategoryEOL }

func (d *EKSEOLVersionDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	eolVersions := map[string]string{
		"1.21": "2023-02-15",
		"1.22": "2023-06-04",
		"1.23": "2023-10-11",
		"1.24": "2024-01-31",
	}

	for _, r := range resources {
		if r.Type == "eks_cluster" {
			versionVal, ok := r.Metadata["kubernetes_version"]
			if !ok || versionVal == nil {
				continue
			}
			version, ok := versionVal.(string)
			if !ok {
				continue
			}
			if eolDate, exists := eolVersions[version]; exists {
				eol, _ := time.Parse("2006-01-02", eolDate)
				if time.Now().After(eol) {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityCritical,
						Title:            "EKS cluster running unsupported Kubernetes version",
						Description:      "Kubernetes " + version + " reached EOL on " + eolDate,
						ResourceARN:      r.ARN,
						EstimatedCost:    0,
						EstimatedSavings: 0,
						Confidence:       1.0,
						Recommendation:   "Upgrade to supported Kubernetes version (security patches unavailable)",
					})
				}
			}
		}
	}
	return findings, nil
}

type ElastiCacheEOLVersionDetector struct{}

func (d *ElastiCacheEOLVersionDetector) Name() string { return "elasticache_eol_version" }
func (d *ElastiCacheEOLVersionDetector) Category() Category { return CategoryEOL }

func (d *ElastiCacheEOLVersionDetector) Detect(resources []Resource) ([]Finding, error) {
	var findings []Finding
	eolVersions := map[string]string{
		"redis-5.0":      "2022-04-30",
		"redis-6.0":      "2024-09-30",
		"memcached-1.4":  "2021-12-31",
		"memcached-1.5":  "2023-06-30",
	}

	for _, r := range resources {
		if r.Type == "elasticache" {
			engineVal, ok := r.Metadata["engine"]
			if !ok || engineVal == nil {
				continue
			}
			engine, ok := engineVal.(string)
			if !ok {
				continue
			}
			versionVal, ok := r.Metadata["engine_version"]
			if !ok || versionVal == nil {
				continue
			}
			version, ok := versionVal.(string)
			if !ok {
				continue
			}
			engineKey := engine + "-" + version

			if eolDate, exists := eolVersions[engineKey]; exists {
				eol, _ := time.Parse("2006-01-02", eolDate)
				if time.Now().After(eol) {
					findings = append(findings, Finding{
						DetectorName:     d.Name(),
						Category:         d.Category(),
						Severity:         SeverityHigh,
						Title:            "ElastiCache running end-of-life engine version",
						Description:      engineKey + " reached EOL on " + eolDate,
						ResourceARN:      r.ARN,
						EstimatedCost:    0,
						EstimatedSavings: 0,
						Confidence:       1.0,
						Recommendation:   "Upgrade to supported engine version",
					})
				}
			}
		}
	}
	return findings, nil
}
