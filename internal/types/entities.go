package types

// Party represents a legal entity (individual or organization)
type Party struct {
	ID        string
	Name      string
	Type      string // individual, organization
	CreatedAt string
}

// Agent represents an AI agent
type Agent struct {
	ID           string
	DisplayName  string
	OwnerPartyID string
	IdentityMode string // did, oauth
	CreatedAt    string
}

// Consent artifact
type Consent struct {
	ID                  string
	AgentID             string
	OwnerPartyID        string
	Rails               []string
	CounterpartiesAllow []string
	Limits              ConsentLimits
	PolicyBundleVersion string
	CosignRule          CosignRule
	CreatedAt           string
	Revoked             bool
}

type ConsentLimits struct {
	SingleTxnUSD float64
	DailyUSD     float64
	Velocity     VelocityCaps
}

type VelocityCaps struct {
	MaxTxnPerHour int
}

type CosignRule struct {
	ThresholdUSD  float64
	ApproverGroup string
}

// RiskDecision represents a risk evaluation result
type RiskDecision struct {
	ID           string
	AgentID      string
	AmountUSD    float64
	Counterparty string
	Rail         string
	Decision     string  // "approve", "deny", "review"
	Score        float64 // 0.0 to 1.0, higher is riskier
	Reason       string
	Threshold    float64
	RiskFactors  []string
	CreatedAt    string
}

// PaymentWorkflow represents a payment processing workflow
type PaymentWorkflow struct {
	ID           string
	AgentID      string
	AmountUSD    float64
	Counterparty string
	Rail         string
	Description  string
	Status       string // "pending", "processing", "completed", "failed"
	Steps        []WorkflowStep
	RiskDecision *RiskDecision
	ConsentCheck *ConsentCheck
	CreatedAt    string
	UpdatedAt    string
}

// WorkflowStep represents a step in the payment workflow
type WorkflowStep struct {
	Name      string
	Status    string // "pending", "running", "completed", "failed"
	Message   string
	Timestamp string
}

// ConsentCheck represents the result of a consent validation
type ConsentCheck struct {
	Valid     bool
	Reason    string
	ConsentID string
}

// PaymentExecution represents a payment execution through a specific rail
type PaymentExecution struct {
	ID           string
	AgentID      string
	AmountUSD    float64
	Counterparty string
	Rail         string
	Description  string
	Status       string // "pending", "processing", "completed", "failed"
	Priority     string
	ReferenceID  string
	ErrorMessage string
	CreatedAt    string
	UpdatedAt    string
}

// Account represents a ledger account for double-entry bookkeeping
type Account struct {
	ID          string
	AgentID     string
	Name        string
	Type        string // "asset", "liability", "equity", "revenue", "expense"
	Description string
	Currency    string
	Balance     float64
	CreatedAt   string
	UpdatedAt   string
}

// Transaction represents a financial transaction in the ledger
type Transaction struct {
	ID          string
	AgentID     string
	Description string
	ReferenceID string
	Status      string // "pending", "posted", "failed"
	CreatedAt   string
	UpdatedAt   string
}

// TransactionDetail represents a transaction with its postings
type TransactionDetail struct {
	ID          string
	AgentID     string
	Description string
	ReferenceID string
	Status      string
	Postings    []*Posting
	CreatedAt   string
}

// Posting represents an individual entry in a transaction (debit or credit)
type Posting struct {
	ID            string
	TransactionID string
	AccountID     string
	Amount        float64 // Positive = debit, negative = credit
	Currency      string
	CreatedAt     string
}

// Account, Counterparty, Transaction, Posting, Decision, Case can be added similarly
