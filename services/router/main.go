package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/example/agent-payments/internal/database"
	"github.com/example/agent-payments/internal/types"
	"github.com/example/agent-payments/libs/common"
	"github.com/gin-gonic/gin"
)

var repo database.Repository

type PaymentExecutionRequest struct {
	AgentID      string  `json:"agentId" binding:"required"`
	AmountUSD    float64 `json:"amountUSD" binding:"required"`
	Counterparty string  `json:"counterparty" binding:"required"`
	Rail         string  `json:"rail,omitempty"` // Optional - if not provided, will be auto-selected
	Description  string  `json:"description"`
	Priority     string  `json:"priority,omitempty"` // "fast", "cheap", "reliable"
}

type RailOption struct {
	Rail        string  `json:"rail"`
	Name        string  `json:"name"`
	CostUSD     float64 `json:"costUSD"`
	SpeedHours  int     `json:"speedHours"`
	Reliability float64 `json:"reliability"` // 0.0 to 1.0
	Available   bool    `json:"available"`
}

type RoutingDecision struct {
	SelectedRail  string       `json:"selectedRail"`
	Reason        string       `json:"reason"`
	EstimatedCost float64      `json:"estimatedCost"`
	EstimatedTime int          `json:"estimatedTime"` // hours
	Alternatives  []RailOption `json:"alternatives"`
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
		c.JSON(http.StatusOK, gin.H{"status": "router service ok"})
	})

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// Payment routing and execution
		v1.POST("/payments/execute", executePayment)
		v1.GET("/payments/:id/status", getPaymentStatus)
		v1.POST("/routing/quote", getRoutingQuote)
		v1.GET("/rails", listAvailableRails)
	}

	common.Info("Router service running on :8085")
	log.Fatal(r.Run(":8085"))
}

func executePayment(c *gin.Context) {
	var req PaymentExecutionRequest
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

	// Determine payment rail if not specified
	selectedRail := req.Rail
	if selectedRail == "" {
		routingDecision := selectOptimalRail(req)
		selectedRail = routingDecision.SelectedRail
	}

	// Create payment execution record
	paymentExecution := &database.PaymentExecution{
		AgentID:      req.AgentID,
		AmountUSD:    req.AmountUSD,
		Counterparty: req.Counterparty,
		Rail:         selectedRail,
		Description:  req.Description,
		Status:       "pending",
		Priority:     req.Priority,
	}

	if err := repo.PaymentExecutionRepository().Create(paymentExecution); err != nil {
		common.Error("Failed to create payment execution: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to create payment execution"))
		return
	}

	// Execute payment asynchronously
	go executePaymentAsync(paymentExecution)

	// Convert to API response format
	response := &types.PaymentExecution{
		ID:           paymentExecution.ID,
		AgentID:      paymentExecution.AgentID,
		AmountUSD:    paymentExecution.AmountUSD,
		Counterparty: paymentExecution.Counterparty,
		Rail:         paymentExecution.Rail,
		Description:  paymentExecution.Description,
		Status:       paymentExecution.Status,
		Priority:     paymentExecution.Priority,
		CreatedAt:    paymentExecution.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    paymentExecution.UpdatedAt.Format(time.RFC3339),
	}

	common.Info("Payment execution initiated: %s for agent %s via %s", paymentExecution.ID, req.AgentID, selectedRail)
	c.JSON(http.StatusCreated, common.NewSuccessResponse(response))
}

func selectOptimalRail(req PaymentExecutionRequest) RoutingDecision {
	availableRails := getAvailableRails(req.AmountUSD)

	if len(availableRails) == 0 {
		return RoutingDecision{
			SelectedRail: "none",
			Reason:       "No suitable payment rails available",
			Alternatives: []RailOption{},
		}
	}

	// Select optimal rail based on priority
	var selectedRail RailOption
	var reason string

	switch req.Priority {
	case "fast":
		selectedRail, reason = selectFastestRail(availableRails)
	case "cheap":
		selectedRail, reason = selectCheapestRail(availableRails)
	case "reliable":
		selectedRail, reason = selectMostReliableRail(availableRails)
	default:
		selectedRail, reason = selectBalancedRail(availableRails)
	}

	return RoutingDecision{
		SelectedRail:  selectedRail.Rail,
		Reason:        reason,
		EstimatedCost: selectedRail.CostUSD,
		EstimatedTime: selectedRail.SpeedHours,
		Alternatives:  availableRails,
	}
}

func getAvailableRails(amount float64) []RailOption {
	rails := []RailOption{
		{
			Rail:        "ach",
			Name:        "ACH Transfer",
			CostUSD:     0.50,
			SpeedHours:  24,
			Reliability: 0.95,
			Available:   true,
		},
		{
			Rail:        "wire",
			Name:        "Wire Transfer",
			CostUSD:     25.00,
			SpeedHours:  2,
			Reliability: 0.99,
			Available:   true,
		},
		{
			Rail:        "card",
			Name:        "Card Payment",
			CostUSD:     amount * 0.029, // 2.9% fee
			SpeedHours:  0,
			Reliability: 0.90,
			Available:   amount <= 5000, // Cards limited to $5k
		},
		{
			Rail:        "instant",
			Name:        "Instant Payment",
			CostUSD:     amount * 0.015, // 1.5% fee
			SpeedHours:  0,
			Reliability: 0.85,
			Available:   amount <= 1000, // Instant limited to $1k
		},
	}

	// Filter available rails
	var available []RailOption
	for _, rail := range rails {
		if rail.Available {
			available = append(available, rail)
		}
	}

	return available
}

func selectFastestRail(rails []RailOption) (RailOption, string) {
	fastest := rails[0]
	for _, rail := range rails[1:] {
		if rail.SpeedHours < fastest.SpeedHours {
			fastest = rail
		}
	}
	return fastest, fmt.Sprintf("Selected %s for fastest processing (%d hours)", fastest.Name, fastest.SpeedHours)
}

func selectCheapestRail(rails []RailOption) (RailOption, string) {
	cheapest := rails[0]
	for _, rail := range rails[1:] {
		if rail.CostUSD < cheapest.CostUSD {
			cheapest = rail
		}
	}
	return cheapest, fmt.Sprintf("Selected %s for lowest cost ($%.2f)", cheapest.Name, cheapest.CostUSD)
}

func selectMostReliableRail(rails []RailOption) (RailOption, string) {
	mostReliable := rails[0]
	for _, rail := range rails[1:] {
		if rail.Reliability > mostReliable.Reliability {
			mostReliable = rail
		}
	}
	return mostReliable, fmt.Sprintf("Selected %s for highest reliability (%.1f%%)", mostReliable.Name, mostReliable.Reliability*100)
}

func selectBalancedRail(rails []RailOption) (RailOption, string) {
	// Score each rail based on balanced criteria
	type scoredRail struct {
		rail  RailOption
		score float64
	}

	var scored []scoredRail
	for _, rail := range rails {
		// Normalize scores (lower cost/time = higher score, higher reliability = higher score)
		costScore := 1.0 / (1.0 + rail.CostUSD/10.0)             // Normalize cost
		timeScore := 1.0 / (1.0 + float64(rail.SpeedHours)/24.0) // Normalize time
		reliabilityScore := rail.Reliability

		totalScore := (costScore * 0.4) + (timeScore * 0.3) + (reliabilityScore * 0.3)
		scored = append(scored, scoredRail{rail: rail, score: totalScore})
	}

	// Find highest scored rail
	best := scored[0]
	for _, s := range scored[1:] {
		if s.score > best.score {
			best = s
		}
	}

	return best.rail, fmt.Sprintf("Selected %s for balanced cost/speed/reliability", best.rail.Name)
}

func executePaymentAsync(execution *database.PaymentExecution) {
	common.Info("Executing payment %s via %s", execution.ID, execution.Rail)

	// Update status to processing
	execution.Status = "processing"
	if err := repo.PaymentExecutionRepository().Update(execution); err != nil {
		common.Error("Failed to update payment execution status: %v", err)
		return
	}

	// Simulate payment processing based on rail
	var processingTime time.Duration
	switch execution.Rail {
	case "instant":
		processingTime = 100 * time.Millisecond
	case "card":
		processingTime = 200 * time.Millisecond
	case "wire":
		processingTime = 500 * time.Millisecond
	case "ach":
		processingTime = 1 * time.Second
	default:
		processingTime = 500 * time.Millisecond
	}

	time.Sleep(processingTime)

	// Simulate success/failure (90% success rate)
	if time.Now().Unix()%10 != 0 { // 90% success rate
		execution.Status = "completed"
		common.Info("Payment %s completed successfully", execution.ID)
	} else {
		execution.Status = "failed"
		common.Error("Payment %s failed", execution.ID)
	}

	execution.UpdatedAt = time.Now()
	if err := repo.PaymentExecutionRepository().Update(execution); err != nil {
		common.Error("Failed to update payment execution final status: %v", err)
	}
}

func getPaymentStatus(c *gin.Context) {
	id := c.Param("id")
	execution, err := repo.PaymentExecutionRepository().GetByID(id)
	if err != nil {
		log.Printf("Failed to get payment execution: %v", err)
		c.JSON(http.StatusNotFound, common.NewErrorResponse("NOT_FOUND", "Payment execution not found"))
		return
	}

	// Convert to API response format
	response := &types.PaymentExecution{
		ID:           execution.ID,
		AgentID:      execution.AgentID,
		AmountUSD:    execution.AmountUSD,
		Counterparty: execution.Counterparty,
		Rail:         execution.Rail,
		Description:  execution.Description,
		Status:       execution.Status,
		Priority:     execution.Priority,
		CreatedAt:    execution.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    execution.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func getRoutingQuote(c *gin.Context) {
	var req PaymentExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	if req.AmountUSD <= 0 {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "amountUSD must be greater than 0"))
		return
	}

	decision := selectOptimalRail(req)
	c.JSON(http.StatusOK, common.NewSuccessResponse(decision))
}

func listAvailableRails(c *gin.Context) {
	amountStr := c.Query("amount")
	amount := 100.0 // Default amount
	if amountStr != "" {
		if parsed, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil || parsed != 1 {
			c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid amount parameter"))
			return
		}
	}

	rails := getAvailableRails(amount)
	c.JSON(http.StatusOK, common.NewSuccessResponse(map[string]interface{}{
		"rails":  rails,
		"amount": amount,
	}))
}
