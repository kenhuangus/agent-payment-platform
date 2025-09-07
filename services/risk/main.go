package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/example/agent-payments/internal/database"
	"github.com/example/agent-payments/internal/types"
	"github.com/example/agent-payments/libs/common"
	"github.com/gin-gonic/gin"
)

var repo database.Repository

type RiskEvaluationRequest struct {
	AgentID      string  `json:"agentId" binding:"required"`
	AmountUSD    float64 `json:"amountUSD" binding:"required"`
	Counterparty string  `json:"counterparty" binding:"required"`
	Rail         string  `json:"rail" binding:"required"`
}

type RiskDecision struct {
	Decision    string   `json:"decision"` // "approve", "deny", "review"
	Score       float64  `json:"score"`    // 0.0 to 1.0, higher is riskier
	Reason      string   `json:"reason"`
	Threshold   float64  `json:"threshold"`
	RiskFactors []string `json:"riskFactors"`
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

	// Health check endpoint
	r.GET("/healthz", func(c *gin.Context) {
		if err := repo.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "database unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "risk service ok"})
	})

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// Risk evaluation
		v1.POST("/risk/evaluate", evaluateRisk)
		v1.GET("/risk/decisions/:id", getRiskDecision)
		v1.GET("/risk/decisions", listRiskDecisions)
	}

	common.Info("Risk service running on :8083")
	log.Fatal(r.Run(":8083"))
}

func evaluateRisk(c *gin.Context) {
	var req RiskEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate required fields
	if req.AgentID == "" || req.AmountUSD <= 0 || req.Counterparty == "" || req.Rail == "" {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "agentId, amountUSD, counterparty, and rail are required"))
		return
	}

	// Verify agent exists
	if _, err := repo.AgentRepository().GetByID(req.AgentID); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Agent not found"))
		return
	}

	// Perform risk evaluation
	decision := evaluateRiskLogic(req)

	// Store risk decision in database
	riskDecision := &database.RiskDecision{
		AgentID:      req.AgentID,
		AmountUSD:    req.AmountUSD,
		Counterparty: req.Counterparty,
		Rail:         req.Rail,
		Decision:     decision.Decision,
		Score:        decision.Score,
		Reason:       decision.Reason,
		Threshold:    decision.Threshold,
		RiskFactors:  "[]", // Would serialize decision.RiskFactors to JSON in production
	}

	if err := repo.RiskDecisionRepository().Create(riskDecision); err != nil {
		common.Error("Failed to create risk decision: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to create risk decision"))
		return
	}

	// Convert to API response format
	response := &types.RiskDecision{
		ID:           riskDecision.ID,
		AgentID:      riskDecision.AgentID,
		AmountUSD:    riskDecision.AmountUSD,
		Counterparty: riskDecision.Counterparty,
		Rail:         riskDecision.Rail,
		Decision:     riskDecision.Decision,
		Score:        riskDecision.Score,
		Reason:       riskDecision.Reason,
		Threshold:    riskDecision.Threshold,
		RiskFactors:  []string{}, // Would deserialize from riskDecision.RiskFactors JSON in production
		CreatedAt:    riskDecision.CreatedAt.Format(time.RFC3339),
	}

	common.Info("Risk evaluation completed: %s for agent %s, decision: %s", riskDecision.ID, req.AgentID, decision.Decision)
	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func evaluateRiskLogic(req RiskEvaluationRequest) RiskDecision {
	score := 0.0
	riskFactors := []string{}
	threshold := 0.7 // Configurable threshold

	// Amount-based risk
	if req.AmountUSD > 25000 {
		score += 0.4
		riskFactors = append(riskFactors, "very_high_amount")
	} else if req.AmountUSD > 10000 {
		score += 0.3
		riskFactors = append(riskFactors, "high_amount")
	} else if req.AmountUSD > 1000 {
		score += 0.15
		riskFactors = append(riskFactors, "medium_amount")
	}

	// Counterparty risk (simplified - in production would check against sanctions lists, etc.)
	counterpartyLower := strings.ToLower(req.Counterparty)
	if strings.Contains(counterpartyLower, "suspicious") ||
		strings.Contains(counterpartyLower, "unknown") ||
		len(req.Counterparty) < 3 {
		score += 0.25
		riskFactors = append(riskFactors, "suspicious_counterparty")
	} else if strings.Contains(counterpartyLower, "new") ||
		strings.Contains(counterpartyLower, "unverified") {
		score += 0.1
		riskFactors = append(riskFactors, "unverified_counterparty")
	}

	// Rail risk
	switch strings.ToLower(req.Rail) {
	case "wire":
		score += 0.2
		riskFactors = append(riskFactors, "wire_transfer")
	case "international":
		score += 0.3
		riskFactors = append(riskFactors, "international_transfer")
	case "card":
		score += 0.05
		riskFactors = append(riskFactors, "card_payment")
	}

	// Cap score at 1.0
	if score > 1.0 {
		score = 1.0
	}

	// Determine decision
	decision := "approve"
	reason := "Transaction approved - low risk"

	if score >= threshold {
		decision = "deny"
		reason = "Transaction denied - risk score exceeds threshold"
	} else if score >= threshold*0.8 {
		decision = "review"
		reason = "Transaction requires manual review"
	}

	return RiskDecision{
		Decision:    decision,
		Score:       score,
		Reason:      reason,
		Threshold:   threshold,
		RiskFactors: riskFactors,
	}
}

func getRiskDecision(c *gin.Context) {
	id := c.Param("id")
	riskDecision, err := repo.RiskDecisionRepository().GetByID(id)
	if err != nil {
		log.Printf("Failed to get risk decision: %v", err)
		c.JSON(http.StatusNotFound, common.NewErrorResponse("NOT_FOUND", "Risk decision not found"))
		return
	}

	// Convert to API response format
	response := &types.RiskDecision{
		ID:           riskDecision.ID,
		AgentID:      riskDecision.AgentID,
		AmountUSD:    riskDecision.AmountUSD,
		Counterparty: riskDecision.Counterparty,
		Rail:         riskDecision.Rail,
		Decision:     riskDecision.Decision,
		Score:        riskDecision.Score,
		Reason:       riskDecision.Reason,
		Threshold:    riskDecision.Threshold,
		RiskFactors:  []string{}, // Would deserialize from riskDecision.RiskFactors JSON in production
		CreatedAt:    riskDecision.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func listRiskDecisions(c *gin.Context) {
	agentID := c.Query("agentId")

	var riskDecisions []*database.RiskDecision
	var err error

	if agentID != "" {
		riskDecisions, err = repo.RiskDecisionRepository().ListByAgentID(agentID)
	} else {
		riskDecisions, err = repo.RiskDecisionRepository().List()
	}

	if err != nil {
		log.Printf("Failed to list risk decisions: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to list risk decisions"))
		return
	}

	// Convert to API response format
	var result []*types.RiskDecision
	for _, rd := range riskDecisions {
		result = append(result, &types.RiskDecision{
			ID:           rd.ID,
			AgentID:      rd.AgentID,
			AmountUSD:    rd.AmountUSD,
			Counterparty: rd.Counterparty,
			Rail:         rd.Rail,
			Decision:     rd.Decision,
			Score:        rd.Score,
			Reason:       rd.Reason,
			Threshold:    rd.Threshold,
			RiskFactors:  []string{}, // Would deserialize from rd.RiskFactors JSON in production
			CreatedAt:    rd.CreatedAt.Format(time.RFC3339),
		})
	}

	response := common.NewListResponse(make([]interface{}, len(result)), 1, 10, len(result))
	for i, rd := range result {
		response.Items[i] = rd
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}
