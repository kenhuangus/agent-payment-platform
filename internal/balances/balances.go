package balances

import (
	"fmt"
	"sort"
	"time"

	"github.com/example/agent-payments/internal/database"
)

// BalanceCalculator provides comprehensive balance calculation functionality
type BalanceCalculator struct {
	repo database.Repository
}

// NewBalanceCalculator creates a new balance calculator
func NewBalanceCalculator(repo database.Repository) *BalanceCalculator {
	return &BalanceCalculator{repo: repo}
}

// AccountBalance represents a detailed account balance
type AccountBalance struct {
	AccountID        string    `json:"accountId"`
	AccountName      string    `json:"accountName"`
	AccountType      string    `json:"accountType"`
	CurrentBalance   float64   `json:"currentBalance"`
	AvailableBalance float64   `json:"availableBalance"`
	Currency         string    `json:"currency"`
	LastUpdated      time.Time `json:"lastUpdated"`
}

// BalanceSheet represents a complete balance sheet
type BalanceSheet struct {
	AgentID          string           `json:"agentId"`
	AsOfDate         time.Time        `json:"asOfDate"`
	Assets           []AccountBalance `json:"assets"`
	TotalAssets      float64          `json:"totalAssets"`
	Liabilities      []AccountBalance `json:"liabilities"`
	TotalLiabilities float64          `json:"totalLiabilities"`
	Equity           []AccountBalance `json:"equity"`
	TotalEquity      float64          `json:"totalEquity"`
	NetWorth         float64          `json:"netWorth"`
}

// TrialBalance represents a trial balance report
type TrialBalance struct {
	AgentID        string           `json:"agentId"`
	PeriodStart    time.Time        `json:"periodStart"`
	PeriodEnd      time.Time        `json:"periodEnd"`
	DebitBalances  []AccountBalance `json:"debitBalances"`
	CreditBalances []AccountBalance `json:"creditBalances"`
	TotalDebits    float64          `json:"totalDebits"`
	TotalCredits   float64          `json:"totalCredits"`
	IsBalanced     bool             `json:"isBalanced"`
}

// BalanceReconciliation represents balance reconciliation data
type BalanceReconciliation struct {
	AccountID          string            `json:"accountId"`
	BookBalance        float64           `json:"bookBalance"`
	BankBalance        float64           `json:"bankBalance"`
	OutstandingChecks  []OutstandingItem `json:"outstandingChecks"`
	DepositsInTransit  []OutstandingItem `json:"depositsInTransit"`
	ReconciledBalance  float64           `json:"reconciledBalance"`
	ReconciliationDate time.Time         `json:"reconciliationDate"`
	IsReconciled       bool              `json:"isReconciled"`
}

// OutstandingItem represents an outstanding transaction item
type OutstandingItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Type        string    `json:"type"` // "check", "deposit"
}

// GetAccountBalance gets detailed balance information for an account
func (bc *BalanceCalculator) GetAccountBalance(accountID string) (*AccountBalance, error) {
	account, err := bc.repo.AccountRepository().GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %s", accountID)
	}

	balance := &AccountBalance{
		AccountID:        account.ID,
		AccountName:      account.Name,
		AccountType:      account.Type,
		CurrentBalance:   account.Balance,
		AvailableBalance: account.Balance, // For now, same as current
		Currency:         account.Currency,
		LastUpdated:      account.UpdatedAt,
	}

	return balance, nil
}

// GetAgentBalances gets all balances for an agent
func (bc *BalanceCalculator) GetAgentBalances(agentID string) ([]AccountBalance, error) {
	accounts, err := bc.repo.AccountRepository().ListByAgentID(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts for agent %s: %v", agentID, err)
	}

	var balances []AccountBalance
	for _, account := range accounts {
		balance := AccountBalance{
			AccountID:        account.ID,
			AccountName:      account.Name,
			AccountType:      account.Type,
			CurrentBalance:   account.Balance,
			AvailableBalance: account.Balance,
			Currency:         account.Currency,
			LastUpdated:      account.UpdatedAt,
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

// GenerateBalanceSheet creates a complete balance sheet for an agent
func (bc *BalanceCalculator) GenerateBalanceSheet(agentID string, asOfDate time.Time) (*BalanceSheet, error) {
	accounts, err := bc.repo.AccountRepository().ListByAgentID(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts for agent %s: %v", agentID, err)
	}

	bs := &BalanceSheet{
		AgentID:     agentID,
		AsOfDate:    asOfDate,
		Assets:      []AccountBalance{},
		Liabilities: []AccountBalance{},
		Equity:      []AccountBalance{},
	}

	for _, account := range accounts {
		balance := AccountBalance{
			AccountID:        account.ID,
			AccountName:      account.Name,
			AccountType:      account.Type,
			CurrentBalance:   account.Balance,
			AvailableBalance: account.Balance,
			Currency:         account.Currency,
			LastUpdated:      account.UpdatedAt,
		}

		switch account.Type {
		case "asset":
			bs.Assets = append(bs.Assets, balance)
			bs.TotalAssets += account.Balance
		case "liability":
			bs.Liabilities = append(bs.Liabilities, balance)
			bs.TotalLiabilities += account.Balance
		case "equity":
			bs.Equity = append(bs.Equity, balance)
			bs.TotalEquity += account.Balance
		}
	}

	bs.NetWorth = bs.TotalAssets - bs.TotalLiabilities

	return bs, nil
}

// GenerateTrialBalance creates a trial balance report
func (bc *BalanceCalculator) GenerateTrialBalance(agentID string, startDate, endDate time.Time) (*TrialBalance, error) {
	accounts, err := bc.repo.AccountRepository().ListByAgentID(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts for agent %s: %v", agentID, err)
	}

	tb := &TrialBalance{
		AgentID:        agentID,
		PeriodStart:    startDate,
		PeriodEnd:      endDate,
		DebitBalances:  []AccountBalance{},
		CreditBalances: []AccountBalance{},
	}

	for _, account := range accounts {
		balance := AccountBalance{
			AccountID:        account.ID,
			AccountName:      account.Name,
			AccountType:      account.Type,
			CurrentBalance:   account.Balance,
			AvailableBalance: account.Balance,
			Currency:         account.Currency,
			LastUpdated:      account.UpdatedAt,
		}

		// Classify based on normal balance type
		switch account.Type {
		case "asset", "expense":
			// Assets and expenses normally have debit balances
			if account.Balance >= 0 {
				tb.DebitBalances = append(tb.DebitBalances, balance)
				tb.TotalDebits += account.Balance
			} else {
				tb.CreditBalances = append(tb.CreditBalances, balance)
				tb.TotalCredits += -account.Balance
			}
		case "liability", "equity", "revenue":
			// Liabilities, equity, and revenue normally have credit balances
			if account.Balance >= 0 {
				tb.CreditBalances = append(tb.CreditBalances, balance)
				tb.TotalCredits += account.Balance
			} else {
				tb.DebitBalances = append(tb.DebitBalances, balance)
				tb.TotalDebits += -account.Balance
			}
		}
	}

	// Check if trial balance is balanced
	tb.IsBalanced = fmt.Sprintf("%.2f", tb.TotalDebits) == fmt.Sprintf("%.2f", tb.TotalCredits)

	return tb, nil
}

// ReconcileAccount performs account reconciliation
func (bc *BalanceCalculator) ReconcileAccount(accountID string, bankBalance float64, reconciliationDate time.Time) (*BalanceReconciliation, error) {
	account, err := bc.repo.AccountRepository().GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %s", accountID)
	}

	// Get outstanding checks (unpresented checks)
	outstandingChecks, err := bc.getOutstandingChecks(accountID, reconciliationDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get outstanding checks: %v", err)
	}

	// Get deposits in transit
	depositsInTransit, err := bc.getDepositsInTransit(accountID, reconciliationDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get deposits in transit: %v", err)
	}

	// Calculate reconciled balance
	reconciledBalance := account.Balance
	for _, check := range outstandingChecks {
		reconciledBalance += check.Amount // Add back outstanding checks
	}
	for _, deposit := range depositsInTransit {
		reconciledBalance -= deposit.Amount // Subtract deposits in transit
	}

	reconciliation := &BalanceReconciliation{
		AccountID:          accountID,
		BookBalance:        account.Balance,
		BankBalance:        bankBalance,
		OutstandingChecks:  outstandingChecks,
		DepositsInTransit:  depositsInTransit,
		ReconciledBalance:  reconciledBalance,
		ReconciliationDate: reconciliationDate,
		IsReconciled:       fmt.Sprintf("%.2f", reconciledBalance) == fmt.Sprintf("%.2f", bankBalance),
	}

	return reconciliation, nil
}

// GetBalanceHistory gets historical balance data for an account
func (bc *BalanceCalculator) GetBalanceHistory(accountID string, startDate, endDate time.Time) ([]AccountBalance, error) {
	// This would typically query transaction history to reconstruct balances
	// For now, return current balance as a placeholder
	currentBalance, err := bc.GetAccountBalance(accountID)
	if err != nil {
		return nil, err
	}

	// In a real implementation, this would reconstruct balance history
	// by replaying transactions from the start date
	history := []AccountBalance{*currentBalance}

	return history, nil
}

// CalculateAccountAging performs account aging analysis
func (bc *BalanceCalculator) CalculateAccountAging(accountID string, asOfDate time.Time) (map[string]float64, error) {
	// This would analyze outstanding receivables/payables by age
	// For now, return a placeholder structure
	aging := map[string]float64{
		"current":  0.0, // 0-30 days
		"30_days":  0.0, // 31-60 days
		"60_days":  0.0, // 61-90 days
		"90_days":  0.0, // 91-120 days
		"120_plus": 0.0, // 120+ days
	}

	// In a real implementation, this would query transaction history
	// and calculate aging based on due dates and payment status

	return aging, nil
}

// ValidateBalanceIntegrity performs balance integrity checks
func (bc *BalanceCalculator) ValidateBalanceIntegrity(agentID string) (map[string]interface{}, error) {
	accounts, err := bc.repo.AccountRepository().ListByAgentID(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %v", err)
	}

	validation := map[string]interface{}{
		"totalAccounts": len(accounts),
		"issues":        []string{},
		"isValid":       true,
	}

	// Check for accounts with invalid balances
	for _, account := range accounts {
		if account.Balance < -1000000 || account.Balance > 1000000 {
			validation["issues"] = append(validation["issues"].([]string),
				fmt.Sprintf("Account %s has suspicious balance: %.2f", account.ID, account.Balance))
			validation["isValid"] = false
		}
	}

	// Check for missing required accounts
	requiredTypes := []string{"asset", "liability", "equity", "revenue", "expense"}
	typeCount := make(map[string]int)
	for _, account := range accounts {
		typeCount[account.Type]++
	}

	for _, requiredType := range requiredTypes {
		if typeCount[requiredType] == 0 {
			validation["issues"] = append(validation["issues"].([]string),
				fmt.Sprintf("Missing required account type: %s", requiredType))
			validation["isValid"] = false
		}
	}

	return validation, nil
}

// Helper functions

func (bc *BalanceCalculator) getOutstandingChecks(accountID string, asOfDate time.Time) ([]OutstandingItem, error) {
	// In a real implementation, this would query for unpresented checks
	// For now, return empty slice
	return []OutstandingItem{}, nil
}

func (bc *BalanceCalculator) getDepositsInTransit(accountID string, asOfDate time.Time) ([]OutstandingItem, error) {
	// In a real implementation, this would query for uncleared deposits
	// For now, return empty slice
	return []OutstandingItem{}, nil
}

// SortBalancesByType sorts account balances by account type
func SortBalancesByType(balances []AccountBalance) {
	sort.Slice(balances, func(i, j int) bool {
		// Define sort order for account types
		typeOrder := map[string]int{
			"asset":     1,
			"liability": 2,
			"equity":    3,
			"revenue":   4,
			"expense":   5,
		}

		iOrder := typeOrder[balances[i].AccountType]
		jOrder := typeOrder[balances[j].AccountType]

		if iOrder != jOrder {
			return iOrder < jOrder
		}

		// Within same type, sort by account name
		return balances[i].AccountName < balances[j].AccountName
	})
}

// FormatCurrency formats a balance as currency string
func FormatCurrency(amount float64, currency string) string {
	return fmt.Sprintf("%s %.2f", currency, amount)
}

// CalculatePercentage calculates percentage of a value relative to total
func CalculatePercentage(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (value / total) * 100
}
