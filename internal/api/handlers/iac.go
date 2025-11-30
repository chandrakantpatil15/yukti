package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"yukti/internal/iac"
)

type IaCHandler struct {
	terraformGen      *iac.TerraformGenerator
	cloudFormationGen *iac.CloudFormationGenerator
}

func NewIaCHandler(region string) *IaCHandler {
	return &IaCHandler{
		terraformGen:      iac.NewTerraformGenerator("aws", region),
		cloudFormationGen: iac.NewCloudFormationGenerator(region),
	}
}

type GenerateIaCRequest struct {
	FindingID string `json:"finding_id"`
	Format    string `json:"format"`
	Action    string `json:"action"`
}

type GenerateIaCResponse struct {
	Code         string   `json:"code"`
	RollbackCode string   `json:"rollback_code"`
	Format       string   `json:"format"`
	Instructions []string `json:"instructions"`
}

func (h *IaCHandler) GenerateIaC(w http.ResponseWriter, r *http.Request) {
	var req GenerateIaCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recommendation := &iac.OptimizationRecommendation{
		ResourceID:              req.FindingID,
		Action:                  req.Action,
		RecommendedInstanceType: "t3.medium",
		EstimatedSavings:        25.0,
		Confidence:              0.90,
		Reasoning:               "Cost optimization opportunity",
	}

	var script *iac.IaCScript
	if req.Format == "terraform" {
		script = h.terraformGen.GenerateEC2Optimization(recommendation)
	} else {
		script = h.cloudFormationGen.GenerateEC2Optimization(recommendation)
	}

	response := GenerateIaCResponse{
		Code:         script.Code,
		RollbackCode: script.RollbackCode,
		Format:       script.Type,
		Instructions: script.Instructions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type BulkGenerateRequest struct {
	FindingIDs []string `json:"finding_ids"`
	Format     string   `json:"format"`
}

type BulkGenerateResponse struct {
	Files []IaCFile `json:"files"`
}

type IaCFile struct {
	Filename string `json:"filename"`
	Code     string `json:"code"`
}

func (h *IaCHandler) BulkGenerate(w http.ResponseWriter, r *http.Request) {
	var req BulkGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var files []IaCFile

	for i, findingID := range req.FindingIDs {
		recommendation := &iac.OptimizationRecommendation{
			ResourceID:       findingID,
			Action:           "downsize",
			EstimatedSavings: 20.0,
		}

		var script *iac.IaCScript
		if req.Format == "terraform" {
			script = h.terraformGen.GenerateEC2Optimization(recommendation)
		} else {
			script = h.cloudFormationGen.GenerateEC2Optimization(recommendation)
		}

		files = append(files, IaCFile{
			Filename: fmt.Sprintf("optimized_%d.%s", i, getExtension(req.Format)),
			Code:     script.Code,
		})
	}

	response := BulkGenerateResponse{Files: files}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getExtension(format string) string {
	if format == "terraform" {
		return "tf"
	}
	return "yaml"
}
