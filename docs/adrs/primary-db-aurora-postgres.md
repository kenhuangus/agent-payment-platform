# ADR: Primary OLTP — Aurora PostgreSQL

## Status
Accepted — 2025-09-06

## Context
We need a relational store for service metadata, identities, consents, workflow state, and materialized views that complement the event log.

## Decision
Use Amazon Aurora PostgreSQL for transactional workloads.
- Strong consistency, read scaling, native JSONB for flexible metadata.
- Extensions: pgcrypto for hashing, logical replication for outbox.
- Multi-AZ and automated backups for resilience.

## Consequences
- Suits transactional semantics and SQL ergonomics.
- True global active-active is limited; for multi-region active-active consider CockroachDB in v1+ if needed.
