package events

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event
type EventType string

const (
	// Payment Events
	EventPaymentInitiated     EventType = "payment.initiated"
	EventPaymentAuthorized    EventType = "payment.authorized"
	EventPaymentRiskEvaluated EventType = "payment.risk_evaluated"
	EventPaymentRouted        EventType = "payment.routed"
	EventPaymentExecuted      EventType = "payment.executed"
	EventPaymentCompleted     EventType = "payment.completed"
	EventPaymentFailed        EventType = "payment.failed"

	// Agent Events
	EventAgentCreated EventType = "agent.created"
	EventAgentUpdated EventType = "agent.updated"

	// Consent Events
	EventConsentCreated EventType = "consent.created"
	EventConsentRevoked EventType = "consent.revoked"

	// Ledger Events
	EventTransactionPosted EventType = "transaction.posted"
	EventAccountCreated    EventType = "account.created"
	EventBalanceUpdated    EventType = "balance.updated"
)

// Event represents a domain event
type Event struct {
	ID            string                 `json:"id"`
	Type          EventType              `json:"type"`
	AggregateID   string                 `json:"aggregateId"`
	AggregateType string                 `json:"aggregateType"`
	Data          map[string]interface{} `json:"data"`
	Metadata      EventMetadata          `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
	Version       int                    `json:"version"`
}

// EventMetadata contains event metadata
type EventMetadata struct {
	Source        string            `json:"source"`
	UserID        string            `json:"userId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	CorrelationID string            `json:"correlationId"`
	CausationID   string            `json:"causationId,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, aggregateID, aggregateType string, data map[string]interface{}) *Event {
	return &Event{
		ID:            uuid.New().String(),
		Type:          eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Data:          data,
		Metadata: EventMetadata{
			CorrelationID: uuid.New().String(),
		},
		Timestamp: time.Now().UTC(),
		Version:   1,
	}
}

// NewEventWithMetadata creates a new event with metadata
func NewEventWithMetadata(eventType EventType, aggregateID, aggregateType string, data map[string]interface{}, metadata EventMetadata) *Event {
	event := NewEvent(eventType, aggregateID, aggregateType, data)
	event.Metadata = metadata
	return event
}

// PaymentInitiatedEventData represents data for payment initiated events
type PaymentInitiatedEventData struct {
	PaymentID    string  `json:"paymentId"`
	AgentID      string  `json:"agentId"`
	AmountUSD    float64 `json:"amountUSD"`
	Counterparty string  `json:"counterparty"`
	Rail         string  `json:"rail"`
	Description  string  `json:"description"`
}

// PaymentRiskEvaluatedEventData represents data for risk evaluation events
type PaymentRiskEvaluatedEventData struct {
	PaymentID   string   `json:"paymentId"`
	Decision    string   `json:"decision"`
	Score       float64  `json:"score"`
	RiskFactors []string `json:"riskFactors"`
	Reason      string   `json:"reason"`
}

// PaymentRoutedEventData represents data for payment routing events
type PaymentRoutedEventData struct {
	PaymentID     string  `json:"paymentId"`
	SelectedRail  string  `json:"selectedRail"`
	Reason        string  `json:"reason"`
	EstimatedCost float64 `json:"estimatedCost"`
	EstimatedTime int     `json:"estimatedTime"`
}

// PaymentExecutedEventData represents data for payment execution events
type PaymentExecutedEventData struct {
	PaymentID    string `json:"paymentId"`
	Rail         string `json:"rail"`
	Status       string `json:"status"`
	ReferenceID  string `json:"referenceId,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// TransactionPostedEventData represents data for transaction posting events
type TransactionPostedEventData struct {
	TransactionID string             `json:"transactionId"`
	AgentID       string             `json:"agentId"`
	Description   string             `json:"description"`
	Postings      []PostingEventData `json:"postings"`
}

// PostingEventData represents data for individual postings
type PostingEventData struct {
	AccountID string  `json:"accountId"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

// AgentCreatedEventData represents data for agent creation events
type AgentCreatedEventData struct {
	AgentID      string `json:"agentId"`
	DisplayName  string `json:"displayName"`
	OwnerPartyID string `json:"ownerPartyId"`
	IdentityMode string `json:"identityMode"`
}

// ConsentCreatedEventData represents data for consent creation events
type ConsentCreatedEventData struct {
	ConsentID     string   `json:"consentId"`
	AgentID       string   `json:"agentId"`
	OwnerPartyID  string   `json:"ownerPartyId"`
	Rails         []string `json:"rails"`
	PolicyVersion string   `json:"policyVersion"`
}
