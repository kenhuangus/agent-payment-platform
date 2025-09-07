package main

import (
	"log"
	"net/http"
	"time"

	"github.com/example/agent-payments/internal/database"
	"github.com/example/agent-payments/internal/types"
	"github.com/example/agent-payments/libs/common"
	"github.com/gin-gonic/gin"
)

var repo database.Repository

type CreateConsentRequest struct {
	AgentID             string           `json:"agentId" binding:"required"`
	OwnerPartyID        string           `json:"ownerPartyId" binding:"required"`
	Rails               []string         `json:"rails"`
	CounterpartiesAllow []string         `json:"counterpartiesAllow"`
	Limits              ConsentLimitsReq `json:"limits"`
	PolicyBundleVersion string           `json:"policyBundleVersion"`
	CosignRule          CosignRuleReq    `json:"cosignRule"`
}

type ConsentLimitsReq struct {
	SingleTxnUSD float64         `json:"singleTxnUSD"`
	DailyUSD     float64         `json:"dailyUSD"`
	Velocity     VelocityCapsReq `json:"velocity"`
}

type VelocityCapsReq struct {
	MaxTxnPerHour int `json:"maxTxnPerHour"`
}

type CosignRuleReq struct {
	ThresholdUSD  float64 `json:"thresholdUSD"`
	ApproverGroup string  `json:"approverGroup"`
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

	r := gin.Default()

	// Setup common middleware
	common.SetupCommonMiddleware(r, func() error {
		return repo.HealthCheck()
	})

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// Consent management
		v1.POST("/consents", createConsent)
		v1.GET("/consents/:id", getConsent)
		v1.GET("/consents", listConsents)
		v1.PUT("/consents/:id/revoke", revokeConsent)

		// Consent validation
		v1.POST("/consents/validate", validateConsent)
	}

	common.Info("Consent service running on :8082")
	log.Fatal(r.Run(":8082"))
}

func createConsent(c *gin.Context) {
	var req CreateConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate required fields
	if req.AgentID == "" || req.OwnerPartyID == "" {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "agentId and ownerPartyId are required"))
		return
	}

	// Verify agent exists
	if _, err := repo.AgentRepository().GetByID(req.AgentID); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Agent not found"))
		return
	}

	// Verify owner party exists
	if _, err := repo.PartyRepository().GetByID(req.OwnerPartyID); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Owner party not found"))
		return
	}

	// Convert request to database model
	consent := &database.Consent{
		AgentID:             req.AgentID,
		OwnerPartyID:        req.OwnerPartyID,
		Rails:               "[]", // Would serialize req.Rails to JSON in production
		CounterpartiesAllow: "[]", // Would serialize req.CounterpartiesAllow to JSON in production
		PolicyBundleVersion: req.PolicyBundleVersion,
		Revoked:             false,
	}

	// Convert nested structures to JSON strings (simplified for now)
	if req.Limits.SingleTxnUSD > 0 || req.Limits.DailyUSD > 0 || req.Limits.Velocity.MaxTxnPerHour > 0 {
		consent.Limits = "{}" // Would serialize to JSON in production
	}

	if req.CosignRule.ThresholdUSD > 0 || req.CosignRule.ApproverGroup != "" {
		consent.CosignRule = "{}" // Would serialize to JSON in production
	}

	if err := repo.ConsentRepository().Create(consent); err != nil {
		common.Error("Failed to create consent: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to create consent"))
		return
	}

	// Convert to API response format
	response := &types.Consent{
		ID:                  consent.ID,
		AgentID:             consent.AgentID,
		OwnerPartyID:        consent.OwnerPartyID,
		Rails:               []string{}, // Would deserialize from consent.Rails JSON in production
		CounterpartiesAllow: []string{}, // Would deserialize from consent.CounterpartiesAllow JSON in production
		PolicyBundleVersion: consent.PolicyBundleVersion,
		CreatedAt:           consent.CreatedAt.Format(time.RFC3339),
		Revoked:             consent.Revoked,
	}

	common.Info("Created consent: %s for agent %s", consent.ID, consent.AgentID)
	c.JSON(http.StatusCreated, common.NewSuccessResponse(response))
}

func getConsent(c *gin.Context) {
	_ = c.Param("id")

	// For now, we'll need to add GetByID to the repository
	// This is a placeholder - would need to implement in repository
	c.JSON(http.StatusNotImplemented, common.NewErrorResponse("NOT_IMPLEMENTED", "Get consent by ID not yet implemented"))
}

func listConsents(c *gin.Context) {
	_ = c.Query("agentId")
	_ = c.Query("ownerPartyId")

	// For now, return empty list - would need to implement in repository
	result := []*types.Consent{}
	interfaces := make([]interface{}, len(result))
	for i, v := range result {
		interfaces[i] = v
	}
	response := common.NewListResponse(interfaces, 1, 10, len(result))

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func revokeConsent(c *gin.Context) {
	_ = c.Param("id")

	// For now, return not implemented - would need to implement in repository
	c.JSON(http.StatusNotImplemented, common.NewErrorResponse("NOT_IMPLEMENTED", "Revoke consent not yet implemented"))
}

type ValidateConsentRequest struct {
	AgentID      string  `json:"agentId" binding:"required"`
	OwnerPartyID string  `json:"ownerPartyId" binding:"required"`
	AmountUSD    float64 `json:"amountUSD" binding:"required"`
	Counterparty string  `json:"counterparty" binding:"required"`
	Rail         string  `json:"rail" binding:"required"`
}

type ConsentValidationResponse struct {
	Valid            bool   `json:"valid"`
	ConsentID        string `json:"consentId,omitempty"`
	Reason           string `json:"reason,omitempty"`
	RequiresApproval bool   `json:"requiresApproval,omitempty"`
	ApproverGroup    string `json:"approverGroup,omitempty"`
}

func validateConsent(c *gin.Context) {
	var req ValidateConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate required fields
	if req.AgentID == "" || req.OwnerPartyID == "" || req.AmountUSD <= 0 || req.Counterparty == "" || req.Rail == "" {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "All fields are required and amount must be positive"))
		return
	}

	// Find active consents for this agent and owner party
	consents, err := repo.ConsentRepository().ListByAgentID(req.AgentID)
	if err != nil {
		common.Error("Failed to list consents: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to retrieve consents"))
		return
	}

	// Filter active (non-revoked) consents
	var activeConsents []*database.Consent
	for _, consent := range consents {
		if !consent.Revoked {
			activeConsents = append(activeConsents, consent)
		}
	}

	if len(activeConsents) == 0 {
		response := &ConsentValidationResponse{
			Valid:  false,
			Reason: "No active consent found for this agent and owner party",
		}
		c.JSON(http.StatusOK, common.NewSuccessResponse(response))
		return
	}

	// Check each consent for validity
	for _, consent := range activeConsents {
		validation := validateConsentRules(consent, req)

		if validation.Valid {
			response := &ConsentValidationResponse{
				Valid:            true,
				ConsentID:        consent.ID,
				RequiresApproval: validation.RequiresApproval,
				ApproverGroup:    validation.ApproverGroup,
			}
			common.Info("Consent validation passed for agent %s, amount %.2f", req.AgentID, req.AmountUSD)
			c.JSON(http.StatusOK, common.NewSuccessResponse(response))
			return
		}
	}

	// No valid consent found
	response := &ConsentValidationResponse{
		Valid:  false,
		Reason: "No consent allows this transaction",
	}
	common.Info("Consent validation failed for agent %s, amount %.2f", req.AgentID, req.AmountUSD)
	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

type ConsentValidationResult struct {
	Valid            bool
	RequiresApproval bool
	ApproverGroup    string
	Reason           string
}

func validateConsentRules(consent *database.Consent, req ValidateConsentRequest) *ConsentValidationResult {
	result := &ConsentValidationResult{
		Valid: true,
	}

	// Check if rail is allowed (simplified - would parse JSON in production)
	// For now, assume all rails are allowed if no specific restrictions

	// Check if counterparty is allowed (simplified - would parse JSON in production)
	// For now, assume all counterparties are allowed if no specific restrictions

	// Check amount limits (simplified - would parse JSON in production)
	// For now, assume no limits if not specified

	// Check cosign rules (simplified - would parse JSON in production)
	// For now, check if amount exceeds threshold requiring approval
	if req.AmountUSD > 10000 { // Example threshold
		result.RequiresApproval = true
		result.ApproverGroup = "senior_approvers"
	}

	return result
}
