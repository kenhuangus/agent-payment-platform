package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/example/agent-payments/internal/database"
	"github.com/segmentio/kafka-go"
)

// EventHandler defines the interface for handling events
type EventHandler interface {
	HandleEvent(ctx context.Context, event *Event) error
	CanHandle(eventType EventType) bool
}

// EventConsumer handles consuming and processing events from Kafka
type EventConsumer struct {
	reader       *kafka.Reader
	handlers     []EventHandler
	repo         database.Repository
	topic        string
	groupID      string
	wg           sync.WaitGroup
	shutdownChan chan struct{}
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(repo database.Repository, kafkaBrokers []string, topic, groupID string) *EventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  kafkaBrokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &EventConsumer{
		reader:       reader,
		handlers:     []EventHandler{},
		repo:         repo,
		topic:        topic,
		groupID:      groupID,
		shutdownChan: make(chan struct{}),
	}
}

// RegisterHandler registers an event handler
func (c *EventConsumer) RegisterHandler(handler EventHandler) {
	c.handlers = append(c.handlers, handler)
}

// Start starts consuming events
func (c *EventConsumer) Start(ctx context.Context) error {
	c.wg.Add(1)
	go c.consumeEvents(ctx)
	log.Printf("Event consumer started for topic: %s, group: %s", c.topic, c.groupID)
	return nil
}

// Stop stops the event consumer
func (c *EventConsumer) Stop() error {
	close(c.shutdownChan)
	c.wg.Wait()
	return c.reader.Close()
}

// consumeEvents continuously consumes events from Kafka
func (c *EventConsumer) consumeEvents(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-c.shutdownChan:
			log.Println("Event consumer shutting down")
			return
		case <-ctx.Done():
			log.Println("Event consumer context cancelled")
			return
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			if err := c.processMessage(ctx, &message); err != nil {
				log.Printf("Error processing message: %v", err)
				// In a production system, you might want to send failed messages to a dead letter queue
			}

			// Commit the message offset
			if err := c.reader.CommitMessages(ctx, message); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

// processMessage processes a single Kafka message
func (c *EventConsumer) processMessage(ctx context.Context, message *kafka.Message) error {
	// Parse the event from the message
	var event Event
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %v", err)
	}

	log.Printf("Processing event: %s (%s)", event.Type, event.ID)

	// Find and execute appropriate handlers
	handled := false
	for _, handler := range c.handlers {
		if handler.CanHandle(event.Type) {
			if err := handler.HandleEvent(ctx, &event); err != nil {
				log.Printf("Handler error for event %s: %v", event.Type, err)
				return err
			}
			handled = true
		}
	}

	if !handled {
		log.Printf("No handler found for event type: %s", event.Type)
	}

	return nil
}

// PaymentEventHandler handles payment-related events
type PaymentEventHandler struct {
	repo database.Repository
}

// NewPaymentEventHandler creates a new payment event handler
func NewPaymentEventHandler(repo database.Repository) *PaymentEventHandler {
	return &PaymentEventHandler{repo: repo}
}

// CanHandle returns true if this handler can handle the given event type
func (h *PaymentEventHandler) CanHandle(eventType EventType) bool {
	switch eventType {
	case EventPaymentInitiated, EventPaymentAuthorized, EventPaymentRiskEvaluated,
		EventPaymentRouted, EventPaymentExecuted, EventPaymentCompleted, EventPaymentFailed:
		return true
	default:
		return false
	}
}

// HandleEvent handles payment events
func (h *PaymentEventHandler) HandleEvent(ctx context.Context, event *Event) error {
	switch event.Type {
	case EventPaymentInitiated:
		return h.handlePaymentInitiated(ctx, event)
	case EventPaymentRiskEvaluated:
		return h.handlePaymentRiskEvaluated(ctx, event)
	case EventPaymentRouted:
		return h.handlePaymentRouted(ctx, event)
	case EventPaymentExecuted:
		return h.handlePaymentExecuted(ctx, event)
	case EventPaymentCompleted:
		return h.handlePaymentCompleted(ctx, event)
	case EventPaymentFailed:
		return h.handlePaymentFailed(ctx, event)
	default:
		return fmt.Errorf("unsupported event type: %s", event.Type)
	}
}

// handlePaymentInitiated handles payment initiated events
func (h *PaymentEventHandler) handlePaymentInitiated(ctx context.Context, event *Event) error {
	var data PaymentInitiatedEventData
	if err := json.Unmarshal(event.Data["data"].(json.RawMessage), &data); err != nil {
		return fmt.Errorf("failed to unmarshal payment initiated data: %v", err)
	}

	log.Printf("Payment initiated: %s, Agent: %s, Amount: %.2f USD",
		data.PaymentID, data.AgentID, data.AmountUSD)

	// Create payment workflow record
	workflow := &database.PaymentWorkflow{
		ID:           data.PaymentID,
		AgentID:      data.AgentID,
		AmountUSD:    data.AmountUSD,
		Counterparty: data.Counterparty,
		Rail:         data.Rail,
		Description:  data.Description,
		Status:       "initiated",
	}

	return h.repo.PaymentWorkflowRepository().Create(workflow)
}

// handlePaymentRiskEvaluated handles risk evaluation events
func (h *PaymentEventHandler) handlePaymentRiskEvaluated(ctx context.Context, event *Event) error {
	var data PaymentRiskEvaluatedEventData
	if err := json.Unmarshal(event.Data["data"].(json.RawMessage), &data); err != nil {
		return fmt.Errorf("failed to unmarshal risk evaluation data: %v", err)
	}

	log.Printf("Payment risk evaluated: %s, Decision: %s, Score: %.2f",
		data.PaymentID, data.Decision, data.Score)

	// Update payment workflow with risk decision
	workflow, err := h.repo.PaymentWorkflowRepository().GetByID(data.PaymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment workflow: %v", err)
	}

	riskData := map[string]interface{}{
		"decision":    data.Decision,
		"score":       data.Score,
		"riskFactors": data.RiskFactors,
		"reason":      data.Reason,
	}

	riskJSON, _ := json.Marshal(riskData)
	workflow.RiskDecision = string(riskJSON)

	return h.repo.PaymentWorkflowRepository().Update(workflow)
}

// handlePaymentRouted handles payment routing events
func (h *PaymentEventHandler) handlePaymentRouted(ctx context.Context, event *Event) error {
	var data PaymentRoutedEventData
	if err := json.Unmarshal(event.Data["data"].(json.RawMessage), &data); err != nil {
		return fmt.Errorf("failed to unmarshal routing data: %v", err)
	}

	log.Printf("Payment routed: %s, Rail: %s, Cost: %.2f, Time: %d",
		data.PaymentID, data.SelectedRail, data.EstimatedCost, data.EstimatedTime)

	// Update payment workflow with routing decision
	workflow, err := h.repo.PaymentWorkflowRepository().GetByID(data.PaymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment workflow: %v", err)
	}

	workflow.Rail = data.SelectedRail
	workflow.Status = "routed"

	return h.repo.PaymentWorkflowRepository().Update(workflow)
}

// handlePaymentExecuted handles payment execution events
func (h *PaymentEventHandler) handlePaymentExecuted(ctx context.Context, event *Event) error {
	var data PaymentExecutedEventData
	if err := json.Unmarshal(event.Data["data"].(json.RawMessage), &data); err != nil {
		return fmt.Errorf("failed to unmarshal execution data: %v", err)
	}

	log.Printf("Payment executed: %s, Status: %s, Reference: %s",
		data.PaymentID, data.Status, data.ReferenceID)

	// Create payment execution record
	execution := &database.PaymentExecution{
		ID:           data.PaymentID,
		AgentID:      "", // This should be populated from the workflow
		AmountUSD:    0,  // This should be populated from the workflow
		Counterparty: "", // This should be populated from the workflow
		Rail:         data.Rail,
		Status:       data.Status,
		ReferenceID:  data.ReferenceID,
		ErrorMessage: data.ErrorMessage,
	}

	return h.repo.PaymentExecutionRepository().Create(execution)
}

// handlePaymentCompleted handles payment completion events
func (h *PaymentEventHandler) handlePaymentCompleted(ctx context.Context, event *Event) error {
	log.Printf("Payment completed: %s", event.AggregateID)

	// Update payment workflow status
	workflow, err := h.repo.PaymentWorkflowRepository().GetByID(event.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get payment workflow: %v", err)
	}

	workflow.Status = "completed"
	return h.repo.PaymentWorkflowRepository().Update(workflow)
}

// handlePaymentFailed handles payment failure events
func (h *PaymentEventHandler) handlePaymentFailed(ctx context.Context, event *Event) error {
	log.Printf("Payment failed: %s", event.AggregateID)

	// Update payment workflow status
	workflow, err := h.repo.PaymentWorkflowRepository().GetByID(event.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get payment workflow: %v", err)
	}

	workflow.Status = "failed"
	return h.repo.PaymentWorkflowRepository().Update(workflow)
}
