package main

import (
	"bytes"
	"context" // Used for HTTP request timeouts
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/example/agent-payments/internal/database"
	"github.com/example/agent-payments/internal/events"
	"github.com/example/agent-payments/internal/types"
	"github.com/example/agent-payments/libs/common"
	"github.com/gin-gonic/gin"
)

var repo database.Repository
var railSelector *types.RailSelector

// Placeholder for event publishing - will be implemented later
var _ = func() interface{} {
	// This prevents the events import from being removed by the formatter
	_ = events.EventPublisher{}
	return nil
}()

type PaymentRequest struct {
	AgentID      string           `json:"agentId" binding:"required"`
	AmountUSD    float64          `json:"amountUSD" binding:"required"`
	Counterparty string           `json:"counterparty" binding:"required"`
	Rail         string           `json:"rail,omitempty"` // Optional - will auto-select if not provided
	Description  string           `json:"description"`
	Preferences  *RailPreferences `json:"preferences,omitempty"`
}

type RailPreferences struct {
	Priority          string   `json:"priority,omitempty"`          // "speed", "cost", "security"
	MaxProcessingTime string   `json:"maxProcessingTime,omitempty"` // Duration string like "30m", "2h"
	MaxSettlementTime string   `json:"maxSettlementTime,omitempty"` // Duration string like "24h", "7d"
	PreferredRails    []string `json:"preferredRails,omitempty"`
	ExcludeRails      []string `json:"excludeRails,omitempty"`
	International     bool     `json:"international,omitempty"`
}

type PaymentWorkflow struct {
	ID           string         `json:"id"`
	AgentID      string         `json:"agentId"`
	AmountUSD    float64        `json:"amountUSD"`
	Counterparty string         `json:"counterparty"`
	Rail         string         `json:"rail"`
	Description  string         `json:"description"`
	Status       string         `json:"status"` // "pending", "processing", "completed", "failed"
	Steps        []WorkflowStep `json:"steps"`
	RiskDecision *RiskDecision  `json:"riskDecision,omitempty"`
	ConsentCheck *ConsentCheck  `json:"consentCheck,omitempty"`
	CreatedAt    string         `json:"createdAt"`
	UpdatedAt    string         `json:"updatedAt"`
}

type WorkflowStep struct {
	Name      string `json:"name"`
	Status    string `json:"status"` // "pending", "running", "completed", "failed"
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type RiskDecision struct {
	Decision    string   `json:"decision"`
	Score       float64  `json:"score"`
	Reason      string   `json:"reason"`
	RiskFactors []string `json:"riskFactors"`
}

type ConsentCheck struct {
	Valid     bool   `json:"valid"`
	Reason    string `json:"reason"`
	ConsentID string `json:"consentId,omitempty"`
}

func main() {
	// Initialize database
	config := database.NewConfig()
	db, err := database.Connect(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repository
	repo = database.NewRepository(db)

	// Initialize rail selector for multi-rail routing
	railSelector = types.NewRailSelector()
	common.Info("Initialized rail selector with %d payment rails", len(railSelector.GetAvailableRails()))

	r := gin.Default()

	// Setup common middleware
	common.SetupCommonMiddleware(r, func() error {
		return repo.HealthCheck()
	})

	// Health check endpoint
	r.GET("/healthz", func(c *gin.Context) {
		if err := repo.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "database unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "orchestration service ok"})
	})

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// Payment orchestration
		v1.POST("/payments", initiatePayment)
		v1.GET("/payments/:id", getPaymentStatus)
		v1.GET("/payments", listPayments)
		v1.POST("/payments/:id/process", processPayment)

		// Rail information
		v1.GET("/rails", getAvailableRails)
		v1.POST("/rails/select", selectRail)
	}

	common.Info("Orchestration service running on :8084")
	log.Fatal(r.Run(":8084"))
}

func initiatePayment(c *gin.Context) {
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate required fields
	if req.AgentID == "" || req.AmountUSD <= 0 || req.Counterparty == "" {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "agentId, amountUSD, and counterparty are required"))
		return
	}

	// Verify agent exists
	if _, err := repo.AgentRepository().GetByID(req.AgentID); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Agent not found"))
		return
	}

	// Handle rail selection - auto-select if not provided
	selectedRail := req.Rail
	if selectedRail == "" {
		// Convert API preferences to internal format
		var prefs *types.RailPreferences
		if req.Preferences != nil {
			prefs = &types.RailPreferences{
				Priority:      req.Preferences.Priority,
				International: req.Preferences.International,
			}

			// Convert preferred rails
			for _, railStr := range req.Preferences.PreferredRails {
				prefs.PreferredRails = append(prefs.PreferredRails, types.PaymentRail(railStr))
			}

			// Convert excluded rails
			for _, railStr := range req.Preferences.ExcludeRails {
				prefs.ExcludeRails = append(prefs.ExcludeRails, types.PaymentRail(railStr))
			}

			// Parse duration strings
			if req.Preferences.MaxProcessingTime != "" {
				if duration, err := time.ParseDuration(req.Preferences.MaxProcessingTime); err == nil {
					prefs.MaxProcessingTime = duration
				}
			}
			if req.Preferences.MaxSettlementTime != "" {
				if duration, err := time.ParseDuration(req.Preferences.MaxSettlementTime); err == nil {
					prefs.MaxSettlementTime = duration
				}
			}
		}

		// Auto-select the best rail
		rail, _, err := railSelector.SelectRail(req.AmountUSD, req.Counterparty, prefs)
		if err != nil {
			common.Error("Failed to select rail for amount %.2f: %v", req.AmountUSD, err)
			c.JSON(http.StatusBadRequest, common.NewErrorResponse("RAIL_SELECTION_ERROR", fmt.Sprintf("No suitable rail found: %v", err)))
			return
		}
		selectedRail = string(rail)
		common.Info("Auto-selected rail %s for payment amount %.2f", selectedRail, req.AmountUSD)
	} else {
		// Validate manually specified rail
		if err := railSelector.ValidateRail(types.PaymentRail(selectedRail), req.AmountUSD); err != nil {
			c.JSON(http.StatusBadRequest, common.NewErrorResponse("RAIL_VALIDATION_ERROR", err.Error()))
			return
		}
	}

	// Create payment workflow
	workflow := &database.PaymentWorkflow{
		AgentID:      req.AgentID,
		AmountUSD:    req.AmountUSD,
		Counterparty: req.Counterparty,
		Rail:         selectedRail,
		Description:  req.Description,
		Status:       "pending",
		Steps:        "[]", // Will be populated with workflow steps
	}

	if err := repo.PaymentWorkflowRepository().Create(workflow); err != nil {
		common.Error("Failed to create payment workflow: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to create payment workflow"))
		return
	}

	// Convert to API response format
	response := &types.PaymentWorkflow{
		ID:           workflow.ID,
		AgentID:      workflow.AgentID,
		AmountUSD:    workflow.AmountUSD,
		Counterparty: workflow.Counterparty,
		Rail:         workflow.Rail,
		Description:  workflow.Description,
		Status:       workflow.Status,
		Steps:        []types.WorkflowStep{}, // Would deserialize from workflow.Steps JSON in production
		CreatedAt:    workflow.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    workflow.UpdatedAt.Format(time.RFC3339),
	}

	common.Info("Payment workflow initiated: %s for agent %s using rail %s", workflow.ID, req.AgentID, selectedRail)
	c.JSON(http.StatusCreated, common.NewSuccessResponse(response))
}

func getPaymentStatus(c *gin.Context) {
	id := c.Param("id")
	workflow, err := repo.PaymentWorkflowRepository().GetByID(id)
	if err != nil {
		log.Printf("Failed to get payment workflow: %v", err)
		c.JSON(http.StatusNotFound, common.NewErrorResponse("NOT_FOUND", "Payment workflow not found"))
		return
	}

	// Convert to API response format
	response := &types.PaymentWorkflow{
		ID:           workflow.ID,
		AgentID:      workflow.AgentID,
		AmountUSD:    workflow.AmountUSD,
		Counterparty: workflow.Counterparty,
		Rail:         workflow.Rail,
		Description:  workflow.Description,
		Status:       workflow.Status,
		Steps:        []types.WorkflowStep{}, // Would deserialize from workflow.Steps JSON in production
		CreatedAt:    workflow.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    workflow.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func listPayments(c *gin.Context) {
	agentID := c.Query("agentId")
	status := c.Query("status")

	var workflows []*database.PaymentWorkflow
	var err error

	if agentID != "" {
		workflows, err = repo.PaymentWorkflowRepository().ListByAgentID(agentID)
	} else if status != "" {
		workflows, err = repo.PaymentWorkflowRepository().ListByStatus(status)
	} else {
		workflows, err = repo.PaymentWorkflowRepository().List()
	}

	if err != nil {
		log.Printf("Failed to list payment workflows: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to list payment workflows"))
		return
	}

	// Convert to API response format
	var result []*types.PaymentWorkflow
	for _, wf := range workflows {
		result = append(result, &types.PaymentWorkflow{
			ID:           wf.ID,
			AgentID:      wf.AgentID,
			AmountUSD:    wf.AmountUSD,
			Counterparty: wf.Counterparty,
			Rail:         wf.Rail,
			Description:  wf.Description,
			Status:       wf.Status,
			Steps:        []types.WorkflowStep{}, // Would deserialize from wf.Steps JSON in production
			CreatedAt:    wf.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    wf.UpdatedAt.Format(time.RFC3339),
		})
	}

	response := common.NewListResponse(make([]interface{}, len(result)), 1, 10, len(result))
	for i, wf := range result {
		response.Items[i] = wf
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func processPayment(c *gin.Context) {
	id := c.Param("id")

	// Get the payment workflow
	workflow, err := repo.PaymentWorkflowRepository().GetByID(id)
	if err != nil {
		log.Printf("Failed to get payment workflow: %v", err)
		c.JSON(http.StatusNotFound, common.NewErrorResponse("NOT_FOUND", "Payment workflow not found"))
		return
	}

	// Only process if status is pending
	if workflow.Status != "pending" {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("INVALID_STATUS", "Payment workflow is not in pending status"))
		return
	}

	// Update status to processing
	workflow.Status = "processing"
	if err := repo.PaymentWorkflowRepository().Update(workflow); err != nil {
		common.Error("Failed to update payment workflow status: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to update workflow status"))
		return
	}

	// Process the payment workflow asynchronously
	go processPaymentWorkflow(workflow)

	c.JSON(http.StatusOK, common.NewSuccessResponse(map[string]string{
		"message":    "Payment processing started",
		"workflowId": workflow.ID,
	}))
}

func processPaymentWorkflow(workflow *database.PaymentWorkflow) {
	common.Info("Starting payment processing for workflow %s", workflow.ID)

	// Step 1: Risk Evaluation
	if err := performRiskEvaluation(workflow); err != nil {
		common.Error("Risk evaluation failed for workflow %s: %v", workflow.ID, err)
		updateWorkflowStatus(workflow, "failed", "Risk evaluation failed")
		return
	}

	// Step 2: Consent Validation
	if err := performConsentValidation(workflow); err != nil {
		common.Error("Consent validation failed for workflow %s: %v", workflow.ID, err)
		updateWorkflowStatus(workflow, "failed", "Consent validation failed")
		return
	}

	// Step 3: Compliance Check (placeholder)
	if err := performComplianceCheck(workflow); err != nil {
		common.Error("Compliance check failed for workflow %s: %v", workflow.ID, err)
		updateWorkflowStatus(workflow, "failed", "Compliance check failed")
		return
	}

	// Step 4: Payment Execution (placeholder)
	if err := executePayment(workflow); err != nil {
		common.Error("Payment execution failed for workflow %s: %v", workflow.ID, err)
		updateWorkflowStatus(workflow, "failed", "Payment execution failed")
		return
	}

	// Mark as completed
	updateWorkflowStatus(workflow, "completed", "Payment processed successfully")
	common.Info("Payment processing completed for workflow %s", workflow.ID)
}

func performRiskEvaluation(workflow *database.PaymentWorkflow) error {
	common.Info("Performing risk evaluation for workflow %s", workflow.ID)

	// Call Risk Service
	riskRequest := map[string]interface{}{
		"agentId":      workflow.AgentID,
		"amountUSD":    workflow.AmountUSD,
		"counterparty": workflow.Counterparty,
		"rail":         workflow.Rail,
	}

	riskResponse, err := callService("http://localhost:8083/v1/risk/evaluate", riskRequest)
	if err != nil {
		return fmt.Errorf("failed to call risk service: %v", err)
	}

	// Parse risk evaluation result
	riskData := riskResponse.Data.(map[string]interface{})
	decision := riskData["decision"].(string)
	score := riskData["score"].(float64)
	reason := riskData["reason"].(string)

	// Check if payment should be blocked based on risk decision
	if decision == "deny" {
		return fmt.Errorf("payment denied by risk evaluation: %s", reason)
	}

	// For review decisions, we could implement approval workflow
	if decision == "review" {
		common.Warn("Payment %s requires manual review: %s", workflow.ID, reason)
		// In production, this would trigger approval workflow
	}

	// Store risk decision in workflow
	workflow.RiskDecision = "{}" // Would serialize riskResponse.Data to JSON

	common.Info("Risk evaluation completed for workflow %s: %s (score: %.2f)", workflow.ID, decision, score)
	return repo.PaymentWorkflowRepository().Update(workflow)
}

func performConsentValidation(workflow *database.PaymentWorkflow) error {
	common.Info("Performing consent validation for workflow %s", workflow.ID)

	// Get agent details to find owner party
	agent, err := repo.AgentRepository().GetByID(workflow.AgentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %v", err)
	}

	// Call Consent Service to validate
	consentRequest := map[string]interface{}{
		"agentId":      workflow.AgentID,
		"ownerPartyId": agent.OwnerPartyID,
		"amountUSD":    workflow.AmountUSD,
		"counterparty": workflow.Counterparty,
		"rail":         workflow.Rail,
	}

	consentResponse, err := callService("http://localhost:8082/v1/consents/validate", consentRequest)
	if err != nil {
		return fmt.Errorf("failed to call consent service: %v", err)
	}

	// Parse consent validation result
	consentData := consentResponse.Data.(map[string]interface{})
	valid := consentData["valid"].(bool)

	if !valid {
		reason := "Consent validation failed"
		if reasonVal, exists := consentData["reason"]; exists {
			reason = reasonVal.(string)
		}
		return fmt.Errorf("consent validation failed: %s", reason)
	}

	// Store consent validation result
	workflow.ConsentCheck = "{}" // Would serialize consentResponse.Data to JSON

	common.Info("Consent validation passed for workflow %s", workflow.ID)
	return repo.PaymentWorkflowRepository().Update(workflow)
}

func performComplianceCheck(workflow *database.PaymentWorkflow) error {
	common.Info("Performing compliance check for workflow %s", workflow.ID)

	// Placeholder for compliance check
	// Would call Compliance Service in production
	time.Sleep(100 * time.Millisecond) // Simulate processing time

	return nil
}

func executePayment(workflow *database.PaymentWorkflow) error {
	common.Info("Executing payment for workflow %s", workflow.ID)

	// Placeholder for payment execution
	// Would call Ledger/Router services in production
	time.Sleep(200 * time.Millisecond) // Simulate processing time

	return nil
}

func updateWorkflowStatus(workflow *database.PaymentWorkflow, status, message string) {
	workflow.Status = status
	workflow.UpdatedAt = time.Now()
	if err := repo.PaymentWorkflowRepository().Update(workflow); err != nil {
		common.Error("Failed to update workflow status: %v", err)
	}
}

func getAvailableRails(c *gin.Context) {
	rails := railSelector.GetAvailableRails()

	// Convert to API response format
	var result []map[string]interface{}
	for rail, characteristics := range rails {
		result = append(result, map[string]interface{}{
			"rail":           string(rail),
			"name":           characteristics.Name,
			"description":    characteristics.Description,
			"minAmount":      characteristics.MinAmount,
			"maxAmount":      characteristics.MaxAmount,
			"processingTime": characteristics.ProcessingTime.String(),
			"settlementTime": characteristics.SettlementTime.String(),
			"feeStructure": map[string]interface{}{
				"fixedFee":   characteristics.FeeStructure.FixedFee,
				"percentFee": characteristics.FeeStructure.PercentFee,
				"minFee":     characteristics.FeeStructure.MinFee,
				"maxFee":     characteristics.FeeStructure.MaxFee,
			},
			"riskLevel":            characteristics.RiskLevel,
			"reversibility":        characteristics.Reversibility,
			"internationalSupport": characteristics.InternationalSupport,
			"requiresVerification": characteristics.RequiresVerification,
		})
	}

	response := common.NewListResponse(make([]interface{}, len(result)), 1, len(result), len(result))
	for i, rail := range result {
		response.Items[i] = rail
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

type RailSelectionRequest struct {
	AmountUSD    float64          `json:"amountUSD" binding:"required"`
	Counterparty string           `json:"counterparty"`
	Preferences  *RailPreferences `json:"preferences,omitempty"`
}

func selectRail(c *gin.Context) {
	var req RailSelectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate amount
	if req.AmountUSD <= 0 {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "amountUSD must be greater than 0"))
		return
	}

	// Convert API preferences to internal format
	var prefs *types.RailPreferences
	if req.Preferences != nil {
		prefs = &types.RailPreferences{
			Priority:      req.Preferences.Priority,
			International: req.Preferences.International,
		}

		// Convert preferred rails
		for _, railStr := range req.Preferences.PreferredRails {
			prefs.PreferredRails = append(prefs.PreferredRails, types.PaymentRail(railStr))
		}

		// Convert excluded rails
		for _, railStr := range req.Preferences.ExcludeRails {
			prefs.ExcludeRails = append(prefs.ExcludeRails, types.PaymentRail(railStr))
		}

		// Parse duration strings
		if req.Preferences.MaxProcessingTime != "" {
			if duration, err := time.ParseDuration(req.Preferences.MaxProcessingTime); err == nil {
				prefs.MaxProcessingTime = duration
			}
		}
		if req.Preferences.MaxSettlementTime != "" {
			if duration, err := time.ParseDuration(req.Preferences.MaxSettlementTime); err == nil {
				prefs.MaxSettlementTime = duration
			}
		}
	}

	// Select the best rail
	selectedRail, characteristics, err := railSelector.SelectRail(req.AmountUSD, req.Counterparty, prefs)
	if err != nil {
		common.Error("Failed to select rail for amount %.2f: %v", req.AmountUSD, err)
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("RAIL_SELECTION_ERROR", fmt.Sprintf("No suitable rail found: %v", err)))
		return
	}

	// Calculate estimated fee (using the unexported method through a workaround)
	// In production, this method should be exported
	fee := characteristics.FeeStructure.FixedFee + (req.AmountUSD * characteristics.FeeStructure.PercentFee)
	if fee < characteristics.FeeStructure.MinFee {
		fee = characteristics.FeeStructure.MinFee
	}
	if fee > characteristics.FeeStructure.MaxFee {
		fee = characteristics.FeeStructure.MaxFee
	}

	response := map[string]interface{}{
		"selectedRail": string(selectedRail),
		"railInfo": map[string]interface{}{
			"name":                 characteristics.Name,
			"description":          characteristics.Description,
			"processingTime":       characteristics.ProcessingTime.String(),
			"settlementTime":       characteristics.SettlementTime.String(),
			"estimatedFee":         fee,
			"riskLevel":            characteristics.RiskLevel,
			"reversibility":        characteristics.Reversibility,
			"internationalSupport": characteristics.InternationalSupport,
			"requiresVerification": characteristics.RequiresVerification,
		},
		"amountUSD": req.AmountUSD,
	}

	common.Info("Selected rail %s for payment amount %.2f", selectedRail, req.AmountUSD)
	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func callService(url string, payload interface{}) (*common.APIResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse common.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	if !apiResponse.Success {
		return nil, fmt.Errorf("service call failed: %v", apiResponse.Error)
	}

	return &apiResponse, nil
}
