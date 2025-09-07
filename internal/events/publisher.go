package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/example/agent-payments/internal/database"
	"github.com/segmentio/kafka-go"
)

// EventPublisher handles publishing events using the outbox pattern
type EventPublisher struct {
	repo        database.Repository
	kafkaWriter *kafka.Writer
	topic       string
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(repo database.Repository, kafkaBrokers []string, topic string) *EventPublisher {
	kafkaWriter := &kafka.Writer{
		Addr:         kafka.TCP(kafkaBrokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	return &EventPublisher{
		repo:        repo,
		kafkaWriter: kafkaWriter,
		topic:       topic,
	}
}

// PublishEvent publishes an event using the outbox pattern
func (p *EventPublisher) PublishEvent(ctx context.Context, event *Event) error {
	// Convert event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal event metadata: %v", err)
	}

	// Create outbox event
	outboxEvent := &database.OutboxEvent{
		EventType:     string(event.Type),
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		Payload:       string(payload),
		Metadata:      string(metadata),
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save to outbox (this should be in the same transaction as the business logic)
	if err := p.repo.OutboxEventRepository().Create(outboxEvent); err != nil {
		return fmt.Errorf("failed to save event to outbox: %v", err)
	}

	log.Printf("Event saved to outbox: %s (%s)", event.Type, event.ID)
	return nil
}

// ProcessOutbox processes pending events from the outbox and publishes them to Kafka
func (p *EventPublisher) ProcessOutbox(ctx context.Context) error {
	// Get pending events
	pendingEvents, err := p.repo.OutboxEventRepository().ListPending(10) // Process up to 10 events at a time
	if err != nil {
		return fmt.Errorf("failed to get pending events: %v", err)
	}

	for _, outboxEvent := range pendingEvents {
		if err := p.publishToKafka(ctx, outboxEvent); err != nil {
			log.Printf("Failed to publish event %s to Kafka: %v", outboxEvent.ID, err)
			p.markEventFailed(outboxEvent, err.Error())
			continue
		}

		// Mark as published
		now := time.Now()
		outboxEvent.Status = "published"
		outboxEvent.PublishedAt = &now
		outboxEvent.UpdatedAt = now

		if err := p.repo.OutboxEventRepository().Update(outboxEvent); err != nil {
			log.Printf("Failed to update outbox event status: %v", err)
		}
	}

	return nil
}

// publishToKafka publishes an event to Kafka
func (p *EventPublisher) publishToKafka(ctx context.Context, outboxEvent *database.OutboxEvent) error {
	message := kafka.Message{
		Key:   []byte(outboxEvent.AggregateID),
		Value: []byte(outboxEvent.Payload),
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(outboxEvent.EventType)},
			{Key: "aggregate-type", Value: []byte(outboxEvent.AggregateType)},
			{Key: "outbox-id", Value: []byte(outboxEvent.ID)},
		},
		Time: time.Now(),
	}

	err := p.kafkaWriter.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %v", err)
	}

	log.Printf("Event published to Kafka: %s (%s)", outboxEvent.EventType, outboxEvent.ID)
	return nil
}

// markEventFailed marks an event as failed and increments retry count
func (p *EventPublisher) markEventFailed(outboxEvent *database.OutboxEvent, errorMsg string) {
	outboxEvent.Status = "failed"
	outboxEvent.ErrorMessage = errorMsg
	outboxEvent.RetryCount++
	outboxEvent.UpdatedAt = time.Now()

	if err := p.repo.OutboxEventRepository().Update(outboxEvent); err != nil {
		log.Printf("Failed to update failed event status: %v", err)
	}
}

// Close closes the Kafka writer
func (p *EventPublisher) Close() error {
	return p.kafkaWriter.Close()
}

// EventPublisherInterface defines the interface for event publishing
type EventPublisherInterface interface {
	PublishEvent(ctx context.Context, event *Event) error
	ProcessOutbox(ctx context.Context) error
	Close() error
}
