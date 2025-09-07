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

type CreateAgentRequest struct {
	DisplayName  string `json:"displayName" binding:"required"`
	OwnerPartyID string `json:"ownerPartyId" binding:"required"`
	IdentityMode string `json:"identityMode" binding:"required"`
}

type CreatePartyRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"`
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

	// Health check endpoint
	r.GET("/healthz", func(c *gin.Context) {
		if err := repo.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "database unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "identity service ok"})
	})

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// Party management
		v1.POST("/parties", createParty)
		v1.GET("/parties/:id", getParty)

		// Agent management
		v1.POST("/agents", createAgent)
		v1.GET("/agents/:id", getAgent)
		v1.GET("/agents", listAgents)
	}

	log.Println("Identity service running on :8081")
	log.Fatal(r.Run(":8081"))
}

func createParty(c *gin.Context) {
	var req CreatePartyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate using common validation
	if errors := common.ValidateParty(req.Name, req.Type); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, common.NewValidationErrorResponse(errors))
		return
	}

	party := &database.Party{
		Name: req.Name,
		Type: req.Type,
	}

	if err := repo.PartyRepository().Create(party); err != nil {
		common.Error("Failed to create party: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to create party"))
		return
	}

	// Convert to API response format
	response := &types.Party{
		ID:        party.ID,
		Name:      party.Name,
		Type:      party.Type,
		CreatedAt: party.CreatedAt.Format(time.RFC3339),
	}

	common.Info("Created party: %s (%s)", party.Name, party.ID)
	c.JSON(http.StatusCreated, common.NewSuccessResponse(response))
}

func getParty(c *gin.Context) {
	id := c.Param("id")
	party, err := repo.PartyRepository().GetByID(id)
	if err != nil {
		log.Printf("Failed to get party: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "party not found"})
		return
	}

	// Convert to API response format
	response := &types.Party{
		ID:        party.ID,
		Name:      party.Name,
		Type:      party.Type,
		CreatedAt: party.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

func createAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate using common validation
	if errors := common.ValidateAgent(req.DisplayName, req.OwnerPartyID, req.IdentityMode); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, common.NewValidationErrorResponse(errors))
		return
	}

	// Verify owner party exists
	if _, err := repo.PartyRepository().GetByID(req.OwnerPartyID); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Owner party not found"))
		return
	}

	agent := &database.Agent{
		DisplayName:  req.DisplayName,
		OwnerPartyID: req.OwnerPartyID,
		IdentityMode: req.IdentityMode,
	}

	if err := repo.AgentRepository().Create(agent); err != nil {
		common.Error("Failed to create agent: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to create agent"))
		return
	}

	// Convert to API response format
	response := &types.Agent{
		ID:           agent.ID,
		DisplayName:  agent.DisplayName,
		OwnerPartyID: agent.OwnerPartyID,
		IdentityMode: agent.IdentityMode,
		CreatedAt:    agent.CreatedAt.Format(time.RFC3339),
	}

	common.Info("Created agent: %s (%s) for party %s", agent.DisplayName, agent.ID, agent.OwnerPartyID)
	c.JSON(http.StatusCreated, common.NewSuccessResponse(response))
}

func getAgent(c *gin.Context) {
	id := c.Param("id")
	agent, err := repo.AgentRepository().GetByID(id)
	if err != nil {
		log.Printf("Failed to get agent: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	// Convert to API response format
	response := &types.Agent{
		ID:           agent.ID,
		DisplayName:  agent.DisplayName,
		OwnerPartyID: agent.OwnerPartyID,
		IdentityMode: agent.IdentityMode,
		CreatedAt:    agent.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

func listAgents(c *gin.Context) {
	ownerPartyID := c.Query("ownerPartyId")

	var agents []*database.Agent
	var err error

	if ownerPartyID != "" {
		agents, err = repo.AgentRepository().ListByOwnerPartyID(ownerPartyID)
	} else {
		agents, err = repo.AgentRepository().List()
	}

	if err != nil {
		log.Printf("Failed to list agents: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list agents"})
		return
	}

	// Convert to API response format
	var result []*types.Agent
	for _, agent := range agents {
		result = append(result, &types.Agent{
			ID:           agent.ID,
			DisplayName:  agent.DisplayName,
			OwnerPartyID: agent.OwnerPartyID,
			IdentityMode: agent.IdentityMode,
			CreatedAt:    agent.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": result,
		"count":  len(result),
	})
}
