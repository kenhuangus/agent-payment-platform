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

type AccountRequest struct {
	AgentID     string `json:"agentId" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"` // "asset", "liability", "equity", "revenue", "expense"
	Description string `json:"description"`
	Currency    string `json:"currency,omitempty"` // Default to USD
}

type TransactionRequest struct {
	AgentID     string           `json:"agentId" binding:"required"`
	Description string           `json:"description" binding:"required"`
	ReferenceID string           `json:"referenceId,omitempty"`
	Postings    []PostingRequest `json:"postings" binding:"required"`
}

type PostingRequest struct {
	AccountID string  `json:"accountId" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"` // Positive for debit, negative for credit
	Currency  string  `json:"currency,omitempty"`
}

type BalanceResponse struct {
	AccountID   string  `json:"accountId"`
	AccountName string  `json:"accountName"`
	Balance     float64 `json:"balance"`
	Currency    string  `json:"currency"`
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
		c.JSON(http.StatusOK, gin.H{"status": "ledger service ok"})
	})

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// Account management
		v1.POST("/accounts", createAccount)
		v1.GET("/accounts/:id", getAccount)
		v1.GET("/accounts", listAccounts)
		v1.GET("/accounts/:id/balance", getAccountBalance)

		// Transaction management
		v1.POST("/transactions", createTransaction)
		v1.GET("/transactions/:id", getTransaction)
		v1.GET("/transactions", listTransactions)

		// Balance queries
		v1.GET("/balances", getBalances)
		v1.GET("/balances/agent/:agentId", getAgentBalances)
	}

	common.Info("Ledger service running on :8086")
	log.Fatal(r.Run(":8086"))
}

func createAccount(c *gin.Context) {
	var req AccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate required fields
	if req.AgentID == "" || req.Name == "" || req.Type == "" {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "agentId, name, and type are required"))
		return
	}

	// Validate account type
	validTypes := map[string]bool{
		"asset": true, "liability": true, "equity": true,
		"revenue": true, "expense": true,
	}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid account type"))
		return
	}

	// Verify agent exists
	if _, err := repo.AgentRepository().GetByID(req.AgentID); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Agent not found"))
		return
	}

	// Set default currency
	if req.Currency == "" {
		req.Currency = "USD"
	}

	// Create account
	account := &database.Account{
		AgentID:     req.AgentID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Currency:    req.Currency,
		Balance:     0.0,
	}

	if err := repo.AccountRepository().Create(account); err != nil {
		common.Error("Failed to create account: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to create account"))
		return
	}

	// Convert to API response format
	response := &types.Account{
		ID:          account.ID,
		AgentID:     account.AgentID,
		Name:        account.Name,
		Type:        account.Type,
		Description: account.Description,
		Currency:    account.Currency,
		Balance:     account.Balance,
		CreatedAt:   account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
	}

	common.Info("Account created: %s (%s) for agent %s", account.Name, account.ID, req.AgentID)
	c.JSON(http.StatusCreated, common.NewSuccessResponse(response))
}

func getAccount(c *gin.Context) {
	id := c.Param("id")
	account, err := repo.AccountRepository().GetByID(id)
	if err != nil {
		log.Printf("Failed to get account: %v", err)
		c.JSON(http.StatusNotFound, common.NewErrorResponse("NOT_FOUND", "Account not found"))
		return
	}

	// Convert to API response format
	response := &types.Account{
		ID:          account.ID,
		AgentID:     account.AgentID,
		Name:        account.Name,
		Type:        account.Type,
		Description: account.Description,
		Currency:    account.Currency,
		Balance:     account.Balance,
		CreatedAt:   account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func listAccounts(c *gin.Context) {
	agentID := c.Query("agentId")
	accountType := c.Query("type")

	var accounts []*database.Account
	var err error

	if agentID != "" && accountType != "" {
		accounts, err = repo.AccountRepository().ListByAgentIDAndType(agentID, accountType)
	} else if agentID != "" {
		accounts, err = repo.AccountRepository().ListByAgentID(agentID)
	} else if accountType != "" {
		accounts, err = repo.AccountRepository().ListByType(accountType)
	} else {
		accounts, err = repo.AccountRepository().List()
	}

	if err != nil {
		log.Printf("Failed to list accounts: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to list accounts"))
		return
	}

	// Convert to API response format
	var result []*types.Account
	for _, acc := range accounts {
		result = append(result, &types.Account{
			ID:          acc.ID,
			AgentID:     acc.AgentID,
			Name:        acc.Name,
			Type:        acc.Type,
			Description: acc.Description,
			Currency:    acc.Currency,
			Balance:     acc.Balance,
			CreatedAt:   acc.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   acc.UpdatedAt.Format(time.RFC3339),
		})
	}

	response := common.NewListResponse(make([]interface{}, len(result)), 1, 10, len(result))
	for i, acc := range result {
		response.Items[i] = acc
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func getAccountBalance(c *gin.Context) {
	id := c.Param("id")
	account, err := repo.AccountRepository().GetByID(id)
	if err != nil {
		log.Printf("Failed to get account: %v", err)
		c.JSON(http.StatusNotFound, common.NewErrorResponse("NOT_FOUND", "Account not found"))
		return
	}

	response := BalanceResponse{
		AccountID:   account.ID,
		AccountName: account.Name,
		Balance:     account.Balance,
		Currency:    account.Currency,
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func createTransaction(c *gin.Context) {
	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Invalid request format"))
		return
	}

	// Validate required fields
	if req.AgentID == "" || req.Description == "" || len(req.Postings) == 0 {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "agentId, description, and postings are required"))
		return
	}

	// Verify agent exists
	if _, err := repo.AgentRepository().GetByID(req.AgentID); err != nil {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Agent not found"))
		return
	}

	// Validate double-entry bookkeeping (debits must equal credits)
	var totalDebit, totalCredit float64
	for _, posting := range req.Postings {
		if posting.Amount > 0 {
			totalDebit += posting.Amount
		} else {
			totalCredit += -posting.Amount
		}
	}

	if fmt.Sprintf("%.2f", totalDebit) != fmt.Sprintf("%.2f", totalCredit) {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Debits must equal credits"))
		return
	}

	// Create transaction
	transaction := &database.Transaction{
		AgentID:     req.AgentID,
		Description: req.Description,
		ReferenceID: req.ReferenceID,
		Status:      "posted",
	}

	if err := repo.TransactionRepository().Create(transaction); err != nil {
		common.Error("Failed to create transaction: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to create transaction"))
		return
	}

	// Create postings
	for _, postingReq := range req.Postings {
		// Verify account exists and belongs to agent
		account, err := repo.AccountRepository().GetByID(postingReq.AccountID)
		if err != nil {
			common.Error("Account not found: %s", postingReq.AccountID)
			c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Account not found"))
			return
		}

		if account.AgentID != req.AgentID {
			common.Error("Account %s does not belong to agent %s", postingReq.AccountID, req.AgentID)
			c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "Account does not belong to agent"))
			return
		}

		posting := &database.Posting{
			TransactionID: transaction.ID,
			AccountID:     postingReq.AccountID,
			Amount:        postingReq.Amount,
			Currency:      postingReq.Currency,
		}

		if posting.Currency == "" {
			posting.Currency = account.Currency
		}

		if err := repo.PostingRepository().Create(posting); err != nil {
			common.Error("Failed to create posting: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to create posting"))
			return
		}

		// Update account balance
		account.Balance += posting.Amount
		if err := repo.AccountRepository().Update(account); err != nil {
			common.Error("Failed to update account balance: %v", err)
			c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to update account balance"))
			return
		}
	}

	// Convert to API response format
	response := &types.Transaction{
		ID:          transaction.ID,
		AgentID:     transaction.AgentID,
		Description: transaction.Description,
		ReferenceID: transaction.ReferenceID,
		Status:      transaction.Status,
		CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
	}

	common.Info("Transaction posted: %s for agent %s", transaction.ID, req.AgentID)
	c.JSON(http.StatusCreated, common.NewSuccessResponse(response))
}

func getTransaction(c *gin.Context) {
	id := c.Param("id")
	transaction, err := repo.TransactionRepository().GetByID(id)
	if err != nil {
		log.Printf("Failed to get transaction: %v", err)
		c.JSON(http.StatusNotFound, common.NewErrorResponse("NOT_FOUND", "Transaction not found"))
		return
	}

	// Get postings for this transaction
	postings, err := repo.PostingRepository().ListByTransactionID(id)
	if err != nil {
		log.Printf("Failed to get postings: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to get postings"))
		return
	}

	// Convert postings to API format
	var postingResponses []*types.Posting
	for _, p := range postings {
		postingResponses = append(postingResponses, &types.Posting{
			ID:            p.ID,
			TransactionID: p.TransactionID,
			AccountID:     p.AccountID,
			Amount:        p.Amount,
			Currency:      p.Currency,
			CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		})
	}

	// Convert to API response format
	response := &types.TransactionDetail{
		ID:          transaction.ID,
		AgentID:     transaction.AgentID,
		Description: transaction.Description,
		ReferenceID: transaction.ReferenceID,
		Status:      transaction.Status,
		Postings:    postingResponses,
		CreatedAt:   transaction.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func listTransactions(c *gin.Context) {
	agentID := c.Query("agentId")

	var transactions []*database.Transaction
	var err error

	if agentID != "" {
		transactions, err = repo.TransactionRepository().ListByAgentID(agentID)
	} else {
		transactions, err = repo.TransactionRepository().List()
	}

	if err != nil {
		log.Printf("Failed to list transactions: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to list transactions"))
		return
	}

	// Convert to API response format
	var result []*types.Transaction
	for _, tx := range transactions {
		result = append(result, &types.Transaction{
			ID:          tx.ID,
			AgentID:     tx.AgentID,
			Description: tx.Description,
			ReferenceID: tx.ReferenceID,
			Status:      tx.Status,
			CreatedAt:   tx.CreatedAt.Format(time.RFC3339),
		})
	}

	response := common.NewListResponse(make([]interface{}, len(result)), 1, 10, len(result))
	for i, tx := range result {
		response.Items[i] = tx
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(response))
}

func getBalances(c *gin.Context) {
	agentID := c.Query("agentId")

	if agentID == "" {
		c.JSON(http.StatusBadRequest, common.NewErrorResponse("VALIDATION_ERROR", "agentId parameter is required"))
		return
	}

	accounts, err := repo.AccountRepository().ListByAgentID(agentID)
	if err != nil {
		log.Printf("Failed to get accounts: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to get accounts"))
		return
	}

	var balances []BalanceResponse
	for _, account := range accounts {
		balances = append(balances, BalanceResponse{
			AccountID:   account.ID,
			AccountName: account.Name,
			Balance:     account.Balance,
			Currency:    account.Currency,
		})
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(map[string]interface{}{
		"agentId":  agentID,
		"balances": balances,
	}))
}

func getAgentBalances(c *gin.Context) {
	agentID := c.Param("agentId")

	accounts, err := repo.AccountRepository().ListByAgentID(agentID)
	if err != nil {
		log.Printf("Failed to get accounts: %v", err)
		c.JSON(http.StatusInternalServerError, common.NewErrorResponse("DATABASE_ERROR", "Failed to get accounts"))
		return
	}

	var balances []BalanceResponse
	for _, account := range accounts {
		balances = append(balances, BalanceResponse{
			AccountID:   account.ID,
			AccountName: account.Name,
			Balance:     account.Balance,
			Currency:    account.Currency,
		})
	}

	c.JSON(http.StatusOK, common.NewSuccessResponse(map[string]interface{}{
		"agentId":  agentID,
		"balances": balances,
	}))
}
