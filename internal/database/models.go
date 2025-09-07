package database

import (
	"time"

	"gorm.io/gorm"
)

// Party represents a legal entity (individual or organization) in the database
type Party struct {
	ID        string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string `gorm:"not null;size:255"`
	Type      string `gorm:"not null;check:type IN ('individual', 'organization')"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relationships
	Agents []Agent `gorm:"foreignKey:OwnerPartyID"`
}

// Agent represents an AI agent in the database
type Agent struct {
	ID           string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DisplayName  string `gorm:"not null;size:255"`
	OwnerPartyID string `gorm:"type:uuid;not null"`
	IdentityMode string `gorm:"not null;check:identity_mode IN ('did', 'oauth')"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// Relationships
	OwnerParty Party `gorm:"foreignKey:OwnerPartyID;references:ID"`
}

// Consent represents a consent artifact in the database
type Consent struct {
	ID                  string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID             string `gorm:"type:uuid;not null"`
	OwnerPartyID        string `gorm:"type:uuid;not null"`
	Rails               string `gorm:"type:jsonb"` // JSON array of strings
	CounterpartiesAllow string `gorm:"type:jsonb"` // JSON array of strings
	Limits              string `gorm:"type:jsonb"` // JSON object for ConsentLimits
	PolicyBundleVersion string `gorm:"size:100"`
	CosignRule          string `gorm:"type:jsonb"` // JSON object for CosignRule
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
	Revoked             bool           `gorm:"default:false"`

	// Relationships
	Agent      Agent `gorm:"foreignKey:AgentID;references:ID"`
	OwnerParty Party `gorm:"foreignKey:OwnerPartyID;references:ID"`
}

// RiskDecision represents a risk evaluation decision in the database
type RiskDecision struct {
	ID           string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID      string  `gorm:"type:uuid;not null"`
	AmountUSD    float64 `gorm:"type:decimal(15,2);not null"`
	Counterparty string  `gorm:"not null;size:255"`
	Rail         string  `gorm:"not null;size:50"`
	Decision     string  `gorm:"not null;check:decision IN ('approve', 'deny', 'review')"`
	Score        float64 `gorm:"type:decimal(3,2);not null;check:score >= 0 AND score <= 1"`
	Reason       string  `gorm:"not null;size:500"`
	Threshold    float64 `gorm:"type:decimal(3,2);not null"`
	RiskFactors  string  `gorm:"type:jsonb"` // JSON array of strings
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// Relationships
	Agent Agent `gorm:"foreignKey:AgentID;references:ID"`
}

// PaymentWorkflow represents a payment processing workflow in the database
type PaymentWorkflow struct {
	ID           string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID      string  `gorm:"type:uuid;not null"`
	AmountUSD    float64 `gorm:"type:decimal(15,2);not null"`
	Counterparty string  `gorm:"not null;size:255"`
	Rail         string  `gorm:"not null;size:50"`
	Description  string  `gorm:"size:500"`
	Status       string  `gorm:"not null;check:status IN ('pending', 'processing', 'completed', 'failed')"`
	Steps        string  `gorm:"type:jsonb"`    // JSON array of workflow steps
	RiskDecision string  `gorm:"type:jsonb"`    // JSON object for risk decision
	ConsentCheck string  `gorm:"type:jsonb"`    // JSON object for consent check
	Hash         string  `gorm:"size:64;index"` // SHA-256 hash of payment data
	PreviousHash string  `gorm:"size:64;index"` // Previous payment hash for chain
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// Relationships
	Agent Agent `gorm:"foreignKey:AgentID;references:ID"`
}

// PaymentExecution represents a payment execution through a specific rail
type PaymentExecution struct {
	ID           string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID      string  `gorm:"type:uuid;not null"`
	AmountUSD    float64 `gorm:"type:decimal(15,2);not null"`
	Counterparty string  `gorm:"not null;size:255"`
	Rail         string  `gorm:"not null;size:50"`
	Description  string  `gorm:"size:500"`
	Status       string  `gorm:"not null;check:status IN ('pending', 'processing', 'completed', 'failed')"`
	Priority     string  `gorm:"size:50"`  // "fast", "cheap", "reliable"
	ReferenceID  string  `gorm:"size:255"` // External reference from payment processor
	ErrorMessage string  `gorm:"size:500"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// Relationships
	Agent Agent `gorm:"foreignKey:AgentID;references:ID"`
}

// Account represents a ledger account for double-entry bookkeeping
type Account struct {
	ID          string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID     string  `gorm:"type:uuid;not null"`
	Name        string  `gorm:"not null;size:255"`
	Type        string  `gorm:"not null;check:type IN ('asset', 'liability', 'equity', 'revenue', 'expense')"`
	Description string  `gorm:"size:500"`
	Currency    string  `gorm:"not null;size:3;default:'USD'"`
	Balance     float64 `gorm:"type:decimal(15,2);not null;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships
	Agent Agent `gorm:"foreignKey:AgentID;references:ID"`
}

// Transaction represents a financial transaction in the ledger
type Transaction struct {
	ID           string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID      string `gorm:"type:uuid;not null"`
	Description  string `gorm:"not null;size:500"`
	ReferenceID  string `gorm:"size:255"`
	Status       string `gorm:"not null;check:status IN ('pending', 'posted', 'failed')"`
	Hash         string `gorm:"size:64;index"` // SHA-256 hash of transaction data
	PreviousHash string `gorm:"size:64;index"` // Previous transaction hash for chain
	BlockIndex   int    `gorm:"default:0"`     // Block index in hash chain
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// Relationships
	Agent    Agent     `gorm:"foreignKey:AgentID;references:ID"`
	Postings []Posting `gorm:"foreignKey:TransactionID"`
}

// Posting represents an individual entry in a transaction (debit or credit)
type Posting struct {
	ID            string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TransactionID string  `gorm:"type:uuid;not null"`
	AccountID     string  `gorm:"type:uuid;not null"`
	Amount        float64 `gorm:"type:decimal(15,2);not null"` // Positive = debit, negative = credit
	Currency      string  `gorm:"not null;size:3;default:'USD'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relationships
	Transaction Transaction `gorm:"foreignKey:TransactionID;references:ID"`
	Account     Account     `gorm:"foreignKey:AccountID;references:ID"`
}

// OutboxEvent represents an event in the outbox pattern
type OutboxEvent struct {
	ID            string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventType     string `gorm:"not null"`
	AggregateID   string `gorm:"not null"`
	AggregateType string `gorm:"not null"`
	Payload       string `gorm:"type:jsonb;not null"`
	Metadata      string `gorm:"type:jsonb"`
	Status        string `gorm:"not null;check:status IN ('pending', 'published', 'failed')"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PublishedAt   *time.Time
	ErrorMessage  string `gorm:"size:500"`
	RetryCount    int    `gorm:"default:0"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventType     string    `gorm:"not null;index"`
	Severity      string    `gorm:"not null;check:severity IN ('low', 'medium', 'high', 'critical')"`
	UserID        string    `gorm:"index"`
	AgentID       string    `gorm:"index"`
	ResourceID    string    `gorm:"index"`
	ResourceType  string    `gorm:"index"`
	Action        string    `gorm:"not null"`
	Description   string    `gorm:"not null;size:500"`
	IPAddress     string    `gorm:"size:45"`
	UserAgent     string    `gorm:"size:500"`
	OldValues     string    `gorm:"type:jsonb"`
	NewValues     string    `gorm:"type:jsonb"`
	Metadata      string    `gorm:"type:jsonb"`
	SessionID     string    `gorm:"index"`
	CorrelationID string    `gorm:"index"`
	Timestamp     time.Time `gorm:"not null;index"`
	Archived      bool      `gorm:"default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AuditQueryFilters represents filters for querying audit entries
type AuditQueryFilters struct {
	UserID       string
	AgentID      string
	ResourceID   string
	ResourceType string
	EventType    string
	Severity     string
	StartDate    *time.Time
	EndDate      *time.Time
	IPAddress    string
	Limit        int
	Offset       int
}

// TableName specifies the table name for Party
func (Party) TableName() string {
	return "parties"
}

// TableName specifies the table name for Agent
func (Agent) TableName() string {
	return "agents"
}

// TableName specifies the table name for Consent
func (Consent) TableName() string {
	return "consents"
}

// TableName specifies the table name for RiskDecision
func (RiskDecision) TableName() string {
	return "risk_decisions"
}

// TableName specifies the table name for PaymentWorkflow
func (PaymentWorkflow) TableName() string {
	return "payment_workflows"
}

// TableName specifies the table name for PaymentExecution
func (PaymentExecution) TableName() string {
	return "payment_executions"
}

// TableName specifies the table name for OutboxEvent
func (OutboxEvent) TableName() string {
	return "outbox_events"
}

// Migrate creates/updates database tables
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Party{}, &Agent{}, &Consent{}, &RiskDecision{}, &PaymentWorkflow{}, &PaymentExecution{}, &Account{}, &Transaction{}, &Posting{}, &OutboxEvent{}, &AuditEntry{})
}
