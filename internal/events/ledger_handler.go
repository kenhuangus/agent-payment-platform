package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/example/agent-payments/internal/database"
)

// LedgerEventHandler handles ledger-related events
type LedgerEventHandler struct {
	repo database.Repository
}

// NewLedgerEventHandler creates a new ledger event handler
func NewLedgerEventHandler(repo database.Repository) *LedgerEventHandler {
	return &LedgerEventHandler{repo: repo}
}

// CanHandle returns true if this handler can handle the given event type
func (h *LedgerEventHandler) CanHandle(eventType EventType) bool {
	switch eventType {
	case EventTransactionPosted, EventAccountCreated, EventBalanceUpdated:
		return true
	default:
		return false
	}
}

// HandleEvent handles ledger events
func (h *LedgerEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	switch event.Type {
	case EventTransactionPosted:
		return h.handleTransactionPosted(ctx, event)
	case EventAccountCreated:
		return h.handleAccountCreated(ctx, event)
	case EventBalanceUpdated:
		return h.handleBalanceUpdated(ctx, event)
	default:
		return fmt.Errorf("unsupported event type: %s", event.Type)
	}
}

// handleTransactionPosted handles transaction posting events
func (h *LedgerEventHandler) handleTransactionPosted(ctx context.Context, event *Event) error {
	var data TransactionPostedEventData
	if err := json.Unmarshal(event.Data["data"].(json.RawMessage), &data); err != nil {
		return fmt.Errorf("failed to unmarshal transaction data: %v", err)
	}

	log.Printf("Transaction posted: %s, Agent: %s, Description: %s",
		data.TransactionID, data.AgentID, data.Description)

	// Create transaction record
	transaction := &database.Transaction{
		ID:          data.TransactionID,
		AgentID:     data.AgentID,
		Description: data.Description,
		Status:      "posted",
	}

	if err := h.repo.TransactionRepository().Create(transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %v", err)
	}

	// Create postings for the transaction
	for _, postingData := range data.Postings {
		posting := &database.Posting{
			TransactionID: data.TransactionID,
			AccountID:     postingData.AccountID,
			Amount:        postingData.Amount,
			Currency:      postingData.Currency,
		}

		if err := h.repo.PostingRepository().Create(posting); err != nil {
			return fmt.Errorf("failed to create posting: %v", err)
		}

		// Update account balance
		if err := h.updateAccountBalance(postingData.AccountID, postingData.Amount); err != nil {
			return fmt.Errorf("failed to update account balance: %v", err)
		}
	}

	return nil
}

// handleAccountCreated handles account creation events
func (h *LedgerEventHandler) handleAccountCreated(ctx context.Context, event *Event) error {
	log.Printf("Account created: %s", event.AggregateID)

	// The account should already be created by the service that triggered this event
	// This handler can be used for additional processing like notifications, etc.
	return nil
}

// handleBalanceUpdated handles balance update events
func (h *LedgerEventHandler) handleBalanceUpdated(ctx context.Context, event *Event) error {
	log.Printf("Balance updated for account: %s", event.AggregateID)

	// The balance should already be updated by the transaction posting handler
	// This handler can be used for additional processing like notifications, etc.
	return nil
}

// updateAccountBalance updates the balance of an account
func (h *LedgerEventHandler) updateAccountBalance(accountID string, amount float64) error {
	account, err := h.repo.AccountRepository().GetByID(accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}

	// Update balance (positive amount = debit, negative amount = credit)
	account.Balance += amount

	if err := h.repo.AccountRepository().Update(account); err != nil {
		return fmt.Errorf("failed to update account balance: %v", err)
	}

	log.Printf("Updated balance for account %s: %.2f %s",
		accountID, account.Balance, account.Currency)

	return nil
}
