# ADR: Event Backbone — Apache Kafka (MSK)

## Status
Accepted — 2025-09-06

## Context
We need a durable, high-throughput event backbone with exactly-once semantics for financial events, support for the outbox pattern, idempotency, DLQs, and schema governance.

## Decision
Adopt Apache Kafka (AWS MSK managed) with:
- Producers using the outbox pattern (transactional writes DB→outbox→Kafka) for effectively-once publishing.
- Consumer idempotency via idempotency keys and compacted offset stores.
- Per-tenant topics and DLQs to avoid cross-tenant commingling.
- Schema Registry (Confluent-compatible) with Avro/Proto schemas and compatibility policies.

## Consequences
- Operational maturity and ecosystem (Connectors, MirrorMaker, Flink/Spark streaming) accelerate data platform.
- Requires disciplined idempotency in consumers and backpressure handling.
- MSK adds AWS coupling; alternative (Pulsar) remains a future option.
