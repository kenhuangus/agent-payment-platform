package database

import (
	"time"

	"gorm.io/gorm"
)

// Repository defines the interface for data access
type Repository interface {
	PartyRepository() PartyRepository
	AgentRepository() AgentRepository
	ConsentRepository() ConsentRepository
	RiskDecisionRepository() RiskDecisionRepository
	PaymentWorkflowRepository() PaymentWorkflowRepository
	PaymentExecutionRepository() PaymentExecutionRepository
	AccountRepository() AccountRepository
	TransactionRepository() TransactionRepository
	PostingRepository() PostingRepository
	OutboxEventRepository() OutboxEventRepository
	AuditEntryRepository() AuditEntryRepository
	HealthCheck() error
	Migrate() error
}

// PartyRepository defines operations for Party entity
type PartyRepository interface {
	Create(party *Party) error
	GetByID(id string) (*Party, error)
	List() ([]*Party, error)
	Update(party *Party) error
	Delete(id string) error
}

// AgentRepository defines operations for Agent entity
type AgentRepository interface {
	Create(agent *Agent) error
	GetByID(id string) (*Agent, error)
	List() ([]*Agent, error)
	ListByOwnerPartyID(ownerPartyID string) ([]*Agent, error)
	Update(agent *Agent) error
	Delete(id string) error
}

// ConsentRepository defines operations for Consent entity
type ConsentRepository interface {
	Create(consent *Consent) error
	GetByID(id string) (*Consent, error)
	List() ([]*Consent, error)
	ListByAgentID(agentID string) ([]*Consent, error)
	ListByOwnerPartyID(ownerPartyID string) ([]*Consent, error)
	Update(consent *Consent) error
	Delete(id string) error
}

// RiskDecisionRepository defines operations for RiskDecision entity
type RiskDecisionRepository interface {
	Create(riskDecision *RiskDecision) error
	GetByID(id string) (*RiskDecision, error)
	List() ([]*RiskDecision, error)
	ListByAgentID(agentID string) ([]*RiskDecision, error)
	Update(riskDecision *RiskDecision) error
	Delete(id string) error
}

// PaymentWorkflowRepository defines operations for PaymentWorkflow entity
type PaymentWorkflowRepository interface {
	Create(workflow *PaymentWorkflow) error
	GetByID(id string) (*PaymentWorkflow, error)
	List() ([]*PaymentWorkflow, error)
	ListByAgentID(agentID string) ([]*PaymentWorkflow, error)
	ListByStatus(status string) ([]*PaymentWorkflow, error)
	Update(workflow *PaymentWorkflow) error
	Delete(id string) error
}

// PaymentExecutionRepository defines operations for PaymentExecution entity
type PaymentExecutionRepository interface {
	Create(execution *PaymentExecution) error
	GetByID(id string) (*PaymentExecution, error)
	List() ([]*PaymentExecution, error)
	ListByAgentID(agentID string) ([]*PaymentExecution, error)
	ListByStatus(status string) ([]*PaymentExecution, error)
	Update(execution *PaymentExecution) error
	Delete(id string) error
}

// AccountRepository defines operations for Account entity
type AccountRepository interface {
	Create(account *Account) error
	GetByID(id string) (*Account, error)
	List() ([]*Account, error)
	ListByAgentID(agentID string) ([]*Account, error)
	ListByType(accountType string) ([]*Account, error)
	ListByAgentIDAndType(agentID, accountType string) ([]*Account, error)
	Update(account *Account) error
	Delete(id string) error
}

// TransactionRepository defines operations for Transaction entity
type TransactionRepository interface {
	Create(transaction *Transaction) error
	GetByID(id string) (*Transaction, error)
	List() ([]*Transaction, error)
	ListByAgentID(agentID string) ([]*Transaction, error)
	Update(transaction *Transaction) error
	Delete(id string) error
}

// PostingRepository defines operations for Posting entity
type PostingRepository interface {
	Create(posting *Posting) error
	GetByID(id string) (*Posting, error)
	List() ([]*Posting, error)
	ListByTransactionID(transactionID string) ([]*Posting, error)
	ListByAccountID(accountID string) ([]*Posting, error)
	Update(posting *Posting) error
	Delete(id string) error
}

// OutboxEventRepository defines operations for OutboxEvent entity
type OutboxEventRepository interface {
	Create(outboxEvent *OutboxEvent) error
	GetByID(id string) (*OutboxEvent, error)
	ListPending(limit int) ([]*OutboxEvent, error)
	Update(outboxEvent *OutboxEvent) error
	Delete(id string) error
}

// AuditEntryRepository defines operations for AuditEntry entity
type AuditEntryRepository interface {
	Create(auditEntry *AuditEntry) error
	GetByID(id string) (*AuditEntry, error)
	Query(filters AuditQueryFilters) ([]*AuditEntry, error)
	Count() (int64, error)
	CountMissingTimestamps() (int64, error)
	CountDuplicates() (int64, error)
	Archive(beforeDate time.Time) error
}

// repository implements Repository interface
type repository struct {
	db                   *gorm.DB
	partyRepo            PartyRepository
	agentRepo            AgentRepository
	consentRepo          ConsentRepository
	riskDecisionRepo     RiskDecisionRepository
	paymentWorkflowRepo  PaymentWorkflowRepository
	paymentExecutionRepo PaymentExecutionRepository
	accountRepo          AccountRepository
	transactionRepo      TransactionRepository
	postingRepo          PostingRepository
	outboxEventRepo      OutboxEventRepository
	auditEntryRepo       AuditEntryRepository
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db:                   db,
		partyRepo:            &partyRepository{db: db},
		agentRepo:            &agentRepository{db: db},
		consentRepo:          &consentRepository{db: db},
		riskDecisionRepo:     &riskDecisionRepository{db: db},
		paymentWorkflowRepo:  &paymentWorkflowRepository{db: db},
		paymentExecutionRepo: &paymentExecutionRepository{db: db},
		accountRepo:          &accountRepository{db: db},
		transactionRepo:      &transactionRepository{db: db},
		postingRepo:          &postingRepository{db: db},
		outboxEventRepo:      &outboxEventRepository{db: db},
		auditEntryRepo:       &auditEntryRepository{db: db},
	}
}

func (r *repository) PartyRepository() PartyRepository {
	return r.partyRepo
}

func (r *repository) AgentRepository() AgentRepository {
	return r.agentRepo
}

func (r *repository) ConsentRepository() ConsentRepository {
	return r.consentRepo
}

func (r *repository) RiskDecisionRepository() RiskDecisionRepository {
	return r.riskDecisionRepo
}

func (r *repository) PaymentWorkflowRepository() PaymentWorkflowRepository {
	return r.paymentWorkflowRepo
}

func (r *repository) PaymentExecutionRepository() PaymentExecutionRepository {
	return r.paymentExecutionRepo
}

func (r *repository) AccountRepository() AccountRepository {
	return r.accountRepo
}

func (r *repository) TransactionRepository() TransactionRepository {
	return r.transactionRepo
}

func (r *repository) PostingRepository() PostingRepository {
	return r.postingRepo
}

func (r *repository) OutboxEventRepository() OutboxEventRepository {
	return r.outboxEventRepo
}

func (r *repository) AuditEntryRepository() AuditEntryRepository {
	return r.auditEntryRepo
}

func (r *repository) HealthCheck() error {
	return HealthCheck(r.db)
}

func (r *repository) Migrate() error {
	return Migrate(r.db)
}

// partyRepository implements PartyRepository
type partyRepository struct {
	db *gorm.DB
}

func (r *partyRepository) Create(party *Party) error {
	return r.db.Create(party).Error
}

func (r *partyRepository) GetByID(id string) (*Party, error) {
	var party Party
	err := r.db.First(&party, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &party, nil
}

func (r *partyRepository) List() ([]*Party, error) {
	var parties []*Party
	err := r.db.Find(&parties).Error
	return parties, err
}

func (r *partyRepository) Update(party *Party) error {
	return r.db.Save(party).Error
}

func (r *partyRepository) Delete(id string) error {
	return r.db.Delete(&Party{}, "id = ?", id).Error
}

// agentRepository implements AgentRepository
type agentRepository struct {
	db *gorm.DB
}

func (r *agentRepository) Create(agent *Agent) error {
	return r.db.Create(agent).Error
}

func (r *agentRepository) GetByID(id string) (*Agent, error) {
	var agent Agent
	err := r.db.Preload("OwnerParty").First(&agent, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *agentRepository) List() ([]*Agent, error) {
	var agents []*Agent
	err := r.db.Preload("OwnerParty").Find(&agents).Error
	return agents, err
}

func (r *agentRepository) ListByOwnerPartyID(ownerPartyID string) ([]*Agent, error) {
	var agents []*Agent
	err := r.db.Preload("OwnerParty").Where("owner_party_id = ?", ownerPartyID).Find(&agents).Error
	return agents, err
}

func (r *agentRepository) Update(agent *Agent) error {
	return r.db.Save(agent).Error
}

func (r *agentRepository) Delete(id string) error {
	return r.db.Delete(&Agent{}, "id = ?", id).Error
}

// consentRepository implements ConsentRepository
type consentRepository struct {
	db *gorm.DB
}

func (r *consentRepository) Create(consent *Consent) error {
	return r.db.Create(consent).Error
}

func (r *consentRepository) GetByID(id string) (*Consent, error) {
	var consent Consent
	err := r.db.Preload("Agent").Preload("OwnerParty").First(&consent, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &consent, nil
}

func (r *consentRepository) List() ([]*Consent, error) {
	var consents []*Consent
	err := r.db.Preload("Agent").Preload("OwnerParty").Find(&consents).Error
	return consents, err
}

func (r *consentRepository) ListByAgentID(agentID string) ([]*Consent, error) {
	var consents []*Consent
	err := r.db.Preload("Agent").Preload("OwnerParty").Where("agent_id = ?", agentID).Find(&consents).Error
	return consents, err
}

func (r *consentRepository) ListByOwnerPartyID(ownerPartyID string) ([]*Consent, error) {
	var consents []*Consent
	err := r.db.Preload("Agent").Preload("OwnerParty").Where("owner_party_id = ?", ownerPartyID).Find(&consents).Error
	return consents, err
}

func (r *consentRepository) Update(consent *Consent) error {
	return r.db.Save(consent).Error
}

func (r *consentRepository) Delete(id string) error {
	return r.db.Delete(&Consent{}, "id = ?", id).Error
}

// riskDecisionRepository implements RiskDecisionRepository
type riskDecisionRepository struct {
	db *gorm.DB
}

func (r *riskDecisionRepository) Create(riskDecision *RiskDecision) error {
	return r.db.Create(riskDecision).Error
}

func (r *riskDecisionRepository) GetByID(id string) (*RiskDecision, error) {
	var riskDecision RiskDecision
	err := r.db.Preload("Agent").First(&riskDecision, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &riskDecision, nil
}

func (r *riskDecisionRepository) List() ([]*RiskDecision, error) {
	var riskDecisions []*RiskDecision
	err := r.db.Preload("Agent").Find(&riskDecisions).Error
	return riskDecisions, err
}

func (r *riskDecisionRepository) ListByAgentID(agentID string) ([]*RiskDecision, error) {
	var riskDecisions []*RiskDecision
	err := r.db.Preload("Agent").Where("agent_id = ?", agentID).Find(&riskDecisions).Error
	return riskDecisions, err
}

func (r *riskDecisionRepository) Update(riskDecision *RiskDecision) error {
	return r.db.Save(riskDecision).Error
}

func (r *riskDecisionRepository) Delete(id string) error {
	return r.db.Delete(&RiskDecision{}, "id = ?", id).Error
}

// paymentWorkflowRepository implements PaymentWorkflowRepository
type paymentWorkflowRepository struct {
	db *gorm.DB
}

func (r *paymentWorkflowRepository) Create(workflow *PaymentWorkflow) error {
	return r.db.Create(workflow).Error
}

func (r *paymentWorkflowRepository) GetByID(id string) (*PaymentWorkflow, error) {
	var workflow PaymentWorkflow
	err := r.db.Preload("Agent").First(&workflow, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *paymentWorkflowRepository) List() ([]*PaymentWorkflow, error) {
	var workflows []*PaymentWorkflow
	err := r.db.Preload("Agent").Find(&workflows).Error
	return workflows, err
}

func (r *paymentWorkflowRepository) ListByAgentID(agentID string) ([]*PaymentWorkflow, error) {
	var workflows []*PaymentWorkflow
	err := r.db.Preload("Agent").Where("agent_id = ?", agentID).Find(&workflows).Error
	return workflows, err
}

func (r *paymentWorkflowRepository) ListByStatus(status string) ([]*PaymentWorkflow, error) {
	var workflows []*PaymentWorkflow
	err := r.db.Preload("Agent").Where("status = ?", status).Find(&workflows).Error
	return workflows, err
}

func (r *paymentWorkflowRepository) Update(workflow *PaymentWorkflow) error {
	return r.db.Save(workflow).Error
}

func (r *paymentWorkflowRepository) Delete(id string) error {
	return r.db.Delete(&PaymentWorkflow{}, "id = ?", id).Error
}

// paymentExecutionRepository implements PaymentExecutionRepository
type paymentExecutionRepository struct {
	db *gorm.DB
}

func (r *paymentExecutionRepository) Create(execution *PaymentExecution) error {
	return r.db.Create(execution).Error
}

func (r *paymentExecutionRepository) GetByID(id string) (*PaymentExecution, error) {
	var execution PaymentExecution
	err := r.db.Preload("Agent").First(&execution, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

func (r *paymentExecutionRepository) List() ([]*PaymentExecution, error) {
	var executions []*PaymentExecution
	err := r.db.Preload("Agent").Find(&executions).Error
	return executions, err
}

func (r *paymentExecutionRepository) ListByAgentID(agentID string) ([]*PaymentExecution, error) {
	var executions []*PaymentExecution
	err := r.db.Preload("Agent").Where("agent_id = ?", agentID).Find(&executions).Error
	return executions, err
}

func (r *paymentExecutionRepository) ListByStatus(status string) ([]*PaymentExecution, error) {
	var executions []*PaymentExecution
	err := r.db.Preload("Agent").Where("status = ?", status).Find(&executions).Error
	return executions, err
}

func (r *paymentExecutionRepository) Update(execution *PaymentExecution) error {
	return r.db.Save(execution).Error
}

func (r *paymentExecutionRepository) Delete(id string) error {
	return r.db.Delete(&PaymentExecution{}, "id = ?", id).Error
}

// accountRepository implements AccountRepository
type accountRepository struct {
	db *gorm.DB
}

func (r *accountRepository) Create(account *Account) error {
	return r.db.Create(account).Error
}

func (r *accountRepository) GetByID(id string) (*Account, error) {
	var account Account
	err := r.db.Preload("Agent").First(&account, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepository) List() ([]*Account, error) {
	var accounts []*Account
	err := r.db.Preload("Agent").Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) ListByAgentID(agentID string) ([]*Account, error) {
	var accounts []*Account
	err := r.db.Preload("Agent").Where("agent_id = ?", agentID).Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) ListByType(accountType string) ([]*Account, error) {
	var accounts []*Account
	err := r.db.Preload("Agent").Where("type = ?", accountType).Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) ListByAgentIDAndType(agentID, accountType string) ([]*Account, error) {
	var accounts []*Account
	err := r.db.Preload("Agent").Where("agent_id = ? AND type = ?", agentID, accountType).Find(&accounts).Error
	return accounts, err
}

func (r *accountRepository) Update(account *Account) error {
	return r.db.Save(account).Error
}

func (r *accountRepository) Delete(id string) error {
	return r.db.Delete(&Account{}, "id = ?", id).Error
}

// transactionRepository implements TransactionRepository
type transactionRepository struct {
	db *gorm.DB
}

func (r *transactionRepository) Create(transaction *Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) GetByID(id string) (*Transaction, error) {
	var transaction Transaction
	err := r.db.Preload("Agent").Preload("Postings").First(&transaction, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) List() ([]*Transaction, error) {
	var transactions []*Transaction
	err := r.db.Preload("Agent").Preload("Postings").Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) ListByAgentID(agentID string) ([]*Transaction, error) {
	var transactions []*Transaction
	err := r.db.Preload("Agent").Preload("Postings").Where("agent_id = ?", agentID).Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) Update(transaction *Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *transactionRepository) Delete(id string) error {
	return r.db.Delete(&Transaction{}, "id = ?", id).Error
}

// postingRepository implements PostingRepository
type postingRepository struct {
	db *gorm.DB
}

func (r *postingRepository) Create(posting *Posting) error {
	return r.db.Create(posting).Error
}

func (r *postingRepository) GetByID(id string) (*Posting, error) {
	var posting Posting
	err := r.db.Preload("Transaction").Preload("Account").First(&posting, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &posting, nil
}

func (r *postingRepository) List() ([]*Posting, error) {
	var postings []*Posting
	err := r.db.Preload("Transaction").Preload("Account").Find(&postings).Error
	return postings, err
}

func (r *postingRepository) ListByTransactionID(transactionID string) ([]*Posting, error) {
	var postings []*Posting
	err := r.db.Preload("Transaction").Preload("Account").Where("transaction_id = ?", transactionID).Find(&postings).Error
	return postings, err
}

func (r *postingRepository) ListByAccountID(accountID string) ([]*Posting, error) {
	var postings []*Posting
	err := r.db.Preload("Transaction").Preload("Account").Where("account_id = ?", accountID).Find(&postings).Error
	return postings, err
}

func (r *postingRepository) Update(posting *Posting) error {
	return r.db.Save(posting).Error
}

func (r *postingRepository) Delete(id string) error {
	return r.db.Delete(&Posting{}, "id = ?", id).Error
}

// outboxEventRepository implements OutboxEventRepository
type outboxEventRepository struct {
	db *gorm.DB
}

func (r *outboxEventRepository) Create(outboxEvent *OutboxEvent) error {
	return r.db.Create(outboxEvent).Error
}

func (r *outboxEventRepository) GetByID(id string) (*OutboxEvent, error) {
	var outboxEvent OutboxEvent
	err := r.db.First(&outboxEvent, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &outboxEvent, nil
}

func (r *outboxEventRepository) ListPending(limit int) ([]*OutboxEvent, error) {
	var outboxEvents []*OutboxEvent
	err := r.db.Where("status = ?", "pending").Order("created_at ASC").Limit(limit).Find(&outboxEvents).Error
	return outboxEvents, err
}

func (r *outboxEventRepository) Update(outboxEvent *OutboxEvent) error {
	return r.db.Save(outboxEvent).Error
}

func (r *outboxEventRepository) Delete(id string) error {
	return r.db.Delete(&OutboxEvent{}, "id = ?", id).Error
}

// auditEntryRepository implements AuditEntryRepository
type auditEntryRepository struct {
	db *gorm.DB
}

func (r *auditEntryRepository) Create(auditEntry *AuditEntry) error {
	return r.db.Create(auditEntry).Error
}

func (r *auditEntryRepository) GetByID(id string) (*AuditEntry, error) {
	var auditEntry AuditEntry
	err := r.db.First(&auditEntry, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &auditEntry, nil
}

func (r *auditEntryRepository) Query(filters AuditQueryFilters) ([]*AuditEntry, error) {
	query := r.db.Model(&AuditEntry{})

	// Apply filters
	if filters.UserID != "" {
		query = query.Where("user_id = ?", filters.UserID)
	}
	if filters.AgentID != "" {
		query = query.Where("agent_id = ?", filters.AgentID)
	}
	if filters.ResourceID != "" {
		query = query.Where("resource_id = ?", filters.ResourceID)
	}
	if filters.ResourceType != "" {
		query = query.Where("resource_type = ?", filters.ResourceType)
	}
	if filters.EventType != "" {
		query = query.Where("event_type = ?", filters.EventType)
	}
	if filters.Severity != "" {
		query = query.Where("severity = ?", filters.Severity)
	}
	if filters.StartDate != nil {
		query = query.Where("timestamp >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("timestamp <= ?", *filters.EndDate)
	}
	if filters.IPAddress != "" {
		query = query.Where("ip_address = ?", filters.IPAddress)
	}

	// Apply ordering and limits
	query = query.Order("timestamp DESC")
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var auditEntries []*AuditEntry
	err := query.Find(&auditEntries).Error
	return auditEntries, err
}

func (r *auditEntryRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&AuditEntry{}).Count(&count).Error
	return count, err
}

func (r *auditEntryRepository) CountMissingTimestamps() (int64, error) {
	var count int64
	err := r.db.Model(&AuditEntry{}).Where("timestamp IS NULL").Count(&count).Error
	return count, err
}

func (r *auditEntryRepository) CountDuplicates() (int64, error) {
	var count int64
	// This is a simplified duplicate check - in practice, you'd define what constitutes a duplicate
	err := r.db.Raw(`
		SELECT COUNT(*) FROM (
			SELECT COUNT(*) as cnt FROM audit_entries
			GROUP BY event_type, user_id, agent_id, resource_id, timestamp
			HAVING COUNT(*) > 1
		) duplicates
	`).Scan(&count).Error
	return count, err
}

func (r *auditEntryRepository) Archive(beforeDate time.Time) error {
	// Mark entries as archived instead of deleting them
	return r.db.Model(&AuditEntry{}).Where("timestamp < ?", beforeDate).Update("archived", true).Error
}
