package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/example/agent-payments/internal/database"
	"github.com/google/uuid"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// Authentication Events
	AuditLoginSuccess     AuditEventType = "auth.login.success"
	AuditLoginFailed      AuditEventType = "auth.login.failed"
	AuditLogout           AuditEventType = "auth.logout"
	AuditPasswordChange   AuditEventType = "auth.password.change"
	AuditPermissionGrant  AuditEventType = "auth.permission.grant"
	AuditPermissionRevoke AuditEventType = "auth.permission.revoke"

	// Payment Events
	AuditPaymentInitiated   AuditEventType = "payment.initiated"
	AuditPaymentAuthorized  AuditEventType = "payment.authorized"
	AuditPaymentRiskChecked AuditEventType = "payment.risk_checked"
	AuditPaymentRouted      AuditEventType = "payment.routed"
	AuditPaymentExecuted    AuditEventType = "payment.executed"
	AuditPaymentCompleted   AuditEventType = "payment.completed"
	AuditPaymentFailed      AuditEventType = "payment.failed"
	AuditPaymentCancelled   AuditEventType = "payment.cancelled"

	// Account Events
	AuditAccountCreated    AuditEventType = "account.created"
	AuditAccountUpdated    AuditEventType = "account.updated"
	AuditAccountDeleted    AuditEventType = "account.deleted"
	AuditBalanceChanged    AuditEventType = "account.balance_changed"
	AuditAccountReconciled AuditEventType = "account.reconciled"

	// Transaction Events
	AuditTransactionPosted    AuditEventType = "transaction.posted"
	AuditTransactionVoided    AuditEventType = "transaction.voided"
	AuditTransactionCorrected AuditEventType = "transaction.corrected"

	// Agent Events
	AuditAgentCreated   AuditEventType = "agent.created"
	AuditAgentUpdated   AuditEventType = "agent.updated"
	AuditAgentSuspended AuditEventType = "agent.suspended"
	AuditAgentActivated AuditEventType = "agent.activated"

	// Consent Events
	AuditConsentCreated AuditEventType = "consent.created"
	AuditConsentRevoked AuditEventType = "consent.revoked"
	AuditConsentUpdated AuditEventType = "consent.updated"

	// System Events
	AuditSystemConfigChanged AuditEventType = "system.config.changed"
	AuditDataExport          AuditEventType = "system.data.export"
	AuditBackupCreated       AuditEventType = "system.backup.created"
	AuditSecurityAlert       AuditEventType = "system.security.alert"
)

// AuditSeverity represents the severity level of an audit event
type AuditSeverity string

const (
	SeverityLow      AuditSeverity = "low"
	SeverityMedium   AuditSeverity = "medium"
	SeverityHigh     AuditSeverity = "high"
	SeverityCritical AuditSeverity = "critical"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID            string                 `json:"id"`
	EventType     AuditEventType         `json:"eventType"`
	Severity      AuditSeverity          `json:"severity"`
	UserID        string                 `json:"userId,omitempty"`
	AgentID       string                 `json:"agentId,omitempty"`
	ResourceID    string                 `json:"resourceId,omitempty"`
	ResourceType  string                 `json:"resourceType,omitempty"`
	Action        string                 `json:"action"`
	Description   string                 `json:"description"`
	IPAddress     string                 `json:"ipAddress,omitempty"`
	UserAgent     string                 `json:"userAgent,omitempty"`
	OldValues     map[string]interface{} `json:"oldValues,omitempty"`
	NewValues     map[string]interface{} `json:"newValues,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	SessionID     string                 `json:"sessionId,omitempty"`
	CorrelationID string                 `json:"correlationId,omitempty"`
}

// AuditTrail manages audit logging and reporting
type AuditTrail struct {
	repo database.Repository
}

// NewAuditTrail creates a new audit trail manager
func NewAuditTrail(repo database.Repository) *AuditTrail {
	return &AuditTrail{repo: repo}
}

// LogEvent logs an audit event
func (at *AuditTrail) LogEvent(ctx context.Context, entry *AuditEntry) error {
	// Set defaults
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	// Convert to database format
	auditRecord := &database.AuditEntry{
		ID:            entry.ID,
		EventType:     string(entry.EventType),
		Severity:      string(entry.Severity),
		UserID:        entry.UserID,
		AgentID:       entry.AgentID,
		ResourceID:    entry.ResourceID,
		ResourceType:  entry.ResourceType,
		Action:        entry.Action,
		Description:   entry.Description,
		IPAddress:     entry.IPAddress,
		UserAgent:     entry.UserAgent,
		SessionID:     entry.SessionID,
		CorrelationID: entry.CorrelationID,
		Timestamp:     entry.Timestamp,
	}

	// Serialize complex fields to JSON
	if entry.OldValues != nil {
		oldValuesJSON, _ := json.Marshal(entry.OldValues)
		auditRecord.OldValues = string(oldValuesJSON)
	}
	if entry.NewValues != nil {
		newValuesJSON, _ := json.Marshal(entry.NewValues)
		auditRecord.NewValues = string(newValuesJSON)
	}
	if entry.Metadata != nil {
		metadataJSON, _ := json.Marshal(entry.Metadata)
		auditRecord.Metadata = string(metadataJSON)
	}

	return at.repo.AuditEntryRepository().Create(auditRecord)
}

// LogPaymentEvent logs a payment-related audit event
func (at *AuditTrail) LogPaymentEvent(ctx context.Context, eventType AuditEventType, paymentID, agentID, userID string, details map[string]interface{}) error {
	entry := &AuditEntry{
		EventType:    eventType,
		Severity:     SeverityMedium,
		UserID:       userID,
		AgentID:      agentID,
		ResourceID:   paymentID,
		ResourceType: "payment",
		Action:       string(eventType),
		Description:  fmt.Sprintf("Payment %s: %s", eventType, paymentID),
		Metadata:     details,
	}

	return at.LogEvent(ctx, entry)
}

// LogAccountEvent logs an account-related audit event
func (at *AuditTrail) LogAccountEvent(ctx context.Context, eventType AuditEventType, accountID, agentID, userID string, oldValues, newValues map[string]interface{}) error {
	entry := &AuditEntry{
		EventType:    eventType,
		Severity:     SeverityMedium,
		UserID:       userID,
		AgentID:      agentID,
		ResourceID:   accountID,
		ResourceType: "account",
		Action:       string(eventType),
		Description:  fmt.Sprintf("Account %s: %s", eventType, accountID),
		OldValues:    oldValues,
		NewValues:    newValues,
	}

	return at.LogEvent(ctx, entry)
}

// LogTransactionEvent logs a transaction-related audit event
func (at *AuditTrail) LogTransactionEvent(ctx context.Context, eventType AuditEventType, transactionID, agentID, userID string, details map[string]interface{}) error {
	entry := &AuditEntry{
		EventType:    eventType,
		Severity:     SeverityMedium,
		UserID:       userID,
		AgentID:      agentID,
		ResourceID:   transactionID,
		ResourceType: "transaction",
		Action:       string(eventType),
		Description:  fmt.Sprintf("Transaction %s: %s", eventType, transactionID),
		Metadata:     details,
	}

	return at.LogEvent(ctx, entry)
}

// LogSecurityEvent logs a security-related audit event
func (at *AuditTrail) LogSecurityEvent(ctx context.Context, eventType AuditEventType, userID, ipAddress, userAgent string, details map[string]interface{}) error {
	severity := SeverityMedium
	if eventType == AuditLoginFailed || eventType == AuditSecurityAlert {
		severity = SeverityHigh
	}

	entry := &AuditEntry{
		EventType:   eventType,
		Severity:    severity,
		UserID:      userID,
		Action:      string(eventType),
		Description: fmt.Sprintf("Security event: %s", eventType),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Metadata:    details,
	}

	return at.LogEvent(ctx, entry)
}

// QueryAuditTrail queries audit entries with filters
func (at *AuditTrail) QueryAuditTrail(ctx context.Context, filters AuditQueryFilters) ([]*AuditEntry, error) {
	// Convert filters to database query
	dbFilters := database.AuditQueryFilters{
		UserID:       filters.UserID,
		AgentID:      filters.AgentID,
		ResourceID:   filters.ResourceID,
		ResourceType: filters.ResourceType,
		EventType:    string(filters.EventType),
		Severity:     string(filters.Severity),
		StartDate:    filters.StartDate,
		EndDate:      filters.EndDate,
		IPAddress:    filters.IPAddress,
		Limit:        filters.Limit,
		Offset:       filters.Offset,
	}

	auditRecords, err := at.repo.AuditEntryRepository().Query(dbFilters)
	if err != nil {
		return nil, err
	}

	// Convert to API format
	var entries []*AuditEntry
	for _, record := range auditRecords {
		entry := &AuditEntry{
			ID:            record.ID,
			EventType:     AuditEventType(record.EventType),
			Severity:      AuditSeverity(record.Severity),
			UserID:        record.UserID,
			AgentID:       record.AgentID,
			ResourceID:    record.ResourceID,
			ResourceType:  record.ResourceType,
			Action:        record.Action,
			Description:   record.Description,
			IPAddress:     record.IPAddress,
			UserAgent:     record.UserAgent,
			SessionID:     record.SessionID,
			CorrelationID: record.CorrelationID,
			Timestamp:     record.Timestamp,
		}

		// Deserialize JSON fields
		if record.OldValues != "" {
			json.Unmarshal([]byte(record.OldValues), &entry.OldValues)
		}
		if record.NewValues != "" {
			json.Unmarshal([]byte(record.NewValues), &entry.NewValues)
		}
		if record.Metadata != "" {
			json.Unmarshal([]byte(record.Metadata), &entry.Metadata)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// AuditQueryFilters represents filters for querying audit entries
type AuditQueryFilters struct {
	UserID       string         `json:"userId,omitempty"`
	AgentID      string         `json:"agentId,omitempty"`
	ResourceID   string         `json:"resourceId,omitempty"`
	ResourceType string         `json:"resourceType,omitempty"`
	EventType    AuditEventType `json:"eventType,omitempty"`
	Severity     AuditSeverity  `json:"severity,omitempty"`
	StartDate    *time.Time     `json:"startDate,omitempty"`
	EndDate      *time.Time     `json:"endDate,omitempty"`
	IPAddress    string         `json:"ipAddress,omitempty"`
	Limit        int            `json:"limit,omitempty"`
	Offset       int            `json:"offset,omitempty"`
}

// GetAuditSummary generates an audit summary report
func (at *AuditTrail) GetAuditSummary(ctx context.Context, startDate, endDate time.Time) (*AuditSummary, error) {
	entries, err := at.QueryAuditTrail(ctx, AuditQueryFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err != nil {
		return nil, err
	}

	summary := &AuditSummary{
		PeriodStart:      startDate,
		PeriodEnd:        endDate,
		TotalEvents:      len(entries),
		EventCounts:      make(map[AuditEventType]int),
		SeverityCounts:   make(map[AuditSeverity]int),
		UserActivity:     make(map[string]int),
		ResourceActivity: make(map[string]int),
	}

	for _, entry := range entries {
		// Count by event type
		summary.EventCounts[entry.EventType]++

		// Count by severity
		summary.SeverityCounts[entry.Severity]++

		// Count by user
		if entry.UserID != "" {
			summary.UserActivity[entry.UserID]++
		}

		// Count by resource
		if entry.ResourceID != "" {
			summary.ResourceActivity[entry.ResourceID]++
		}
	}

	return summary, nil
}

// AuditSummary represents a summary of audit activity
type AuditSummary struct {
	PeriodStart      time.Time              `json:"periodStart"`
	PeriodEnd        time.Time              `json:"periodEnd"`
	TotalEvents      int                    `json:"totalEvents"`
	EventCounts      map[AuditEventType]int `json:"eventCounts"`
	SeverityCounts   map[AuditSeverity]int  `json:"severityCounts"`
	UserActivity     map[string]int         `json:"userActivity"`
	ResourceActivity map[string]int         `json:"resourceActivity"`
}

// GetComplianceReport generates a compliance report
func (at *AuditTrail) GetComplianceReport(ctx context.Context, startDate, endDate time.Time) (*ComplianceReport, error) {
	entries, err := at.QueryAuditTrail(ctx, AuditQueryFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err != nil {
		return nil, err
	}

	report := &ComplianceReport{
		PeriodStart:        startDate,
		PeriodEnd:          endDate,
		TotalEvents:        len(entries),
		SecurityEvents:     []AuditEntry{},
		PaymentEvents:      []AuditEntry{},
		FailedLogins:       0,
		SuspiciousActivity: []AuditEntry{},
		ComplianceIssues:   []string{},
	}

	for _, entry := range entries {
		switch entry.EventType {
		case AuditLoginFailed:
			report.FailedLogins++
			report.SecurityEvents = append(report.SecurityEvents, *entry)
		case AuditPaymentFailed, AuditPaymentCancelled:
			report.PaymentEvents = append(report.PaymentEvents, *entry)
		case AuditSecurityAlert:
			report.SuspiciousActivity = append(report.SuspiciousActivity, *entry)
			report.SecurityEvents = append(report.SecurityEvents, *entry)
		}

		// Check for compliance issues
		if entry.Severity == SeverityCritical {
			report.ComplianceIssues = append(report.ComplianceIssues,
				fmt.Sprintf("Critical event: %s - %s", entry.EventType, entry.Description))
		}
	}

	return report, nil
}

// ComplianceReport represents a compliance audit report
type ComplianceReport struct {
	PeriodStart        time.Time    `json:"periodStart"`
	PeriodEnd          time.Time    `json:"periodEnd"`
	TotalEvents        int          `json:"totalEvents"`
	SecurityEvents     []AuditEntry `json:"securityEvents"`
	PaymentEvents      []AuditEntry `json:"paymentEvents"`
	FailedLogins       int          `json:"failedLogins"`
	SuspiciousActivity []AuditEntry `json:"suspiciousActivity"`
	ComplianceIssues   []string     `json:"complianceIssues"`
}

// ArchiveOldEntries archives audit entries older than the specified date
func (at *AuditTrail) ArchiveOldEntries(ctx context.Context, beforeDate time.Time) error {
	// This would typically move old entries to archive storage
	// For now, we'll just mark them as archived in the database
	return at.repo.AuditEntryRepository().Archive(beforeDate)
}

// GetChangeHistory gets the change history for a specific resource
func (at *AuditTrail) GetChangeHistory(ctx context.Context, resourceID, resourceType string, limit int) ([]*AuditEntry, error) {
	return at.QueryAuditTrail(ctx, AuditQueryFilters{
		ResourceID:   resourceID,
		ResourceType: resourceType,
		Limit:        limit,
	})
}

// ValidateAuditIntegrity performs integrity checks on the audit trail
func (at *AuditTrail) ValidateAuditIntegrity(ctx context.Context) (map[string]interface{}, error) {
	validation := map[string]interface{}{
		"totalEntries": 0,
		"issues":       []string{},
		"isValid":      true,
	}

	// Get total count
	totalCount, err := at.repo.AuditEntryRepository().Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count audit entries: %v", err)
	}
	validation["totalEntries"] = totalCount

	// Check for missing timestamps
	missingTimestamps, err := at.repo.AuditEntryRepository().CountMissingTimestamps()
	if err != nil {
		return nil, fmt.Errorf("failed to check missing timestamps: %v", err)
	}
	if missingTimestamps > 0 {
		validation["issues"] = append(validation["issues"].([]string),
			fmt.Sprintf("Found %d entries with missing timestamps", missingTimestamps))
		validation["isValid"] = false
	}

	// Check for duplicate IDs
	duplicateCount, err := at.repo.AuditEntryRepository().CountDuplicates()
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicates: %v", err)
	}
	if duplicateCount > 0 {
		validation["issues"] = append(validation["issues"].([]string),
			fmt.Sprintf("Found %d duplicate audit entries", duplicateCount))
		validation["isValid"] = false
	}

	return validation, nil
}
