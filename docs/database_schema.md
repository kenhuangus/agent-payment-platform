# Database Schema

## Overview

The Agent Payment Platform uses PostgreSQL as its primary database with a carefully designed schema that supports the complex financial operations, audit requirements, and scalability needs of the system.

## Core Business Entities

### Agents Table
```sql
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(100) UNIQUE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'inactive')),
    kyc_status VARCHAR(50) DEFAULT 'pending' CHECK (kyc_status IN ('pending', 'approved', 'rejected', 'expired')),
    risk_score INTEGER DEFAULT 0 CHECK (risk_score >= 0 AND risk_score <= 100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- Indexes
CREATE INDEX idx_agents_email ON agents(email);
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_kyc_status ON agents(kyc_status);
CREATE INDEX idx_agents_created_at ON agents(created_at);
CREATE INDEX idx_agents_risk_score ON agents(risk_score);
```

### Payments Table
```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(100) UNIQUE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    counterparty_id UUID REFERENCES agents(id),
    counterparty_email VARCHAR(255),
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded')),
    rail VARCHAR(50) CHECK (rail IN ('ach', 'card', 'wire', 'check')),
    rail_transaction_id VARCHAR(255),
    description TEXT,
    metadata JSONB DEFAULT '{}',
    risk_score INTEGER CHECK (risk_score >= 0 AND risk_score <= 100),
    fee_amount DECIMAL(10,2) DEFAULT 0,
    fee_currency VARCHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    created_by UUID,
    updated_by UUID
);

-- Indexes
CREATE INDEX idx_payments_agent_id ON payments(agent_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_rail ON payments(rail);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_completed_at ON payments(completed_at);
CREATE INDEX idx_payments_risk_score ON payments(risk_score);
CREATE INDEX idx_payments_counterparty_email ON payments(counterparty_email);
```

### Accounts Table (Chart of Accounts)
```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    account_number VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
    subtype VARCHAR(50),
    parent_account_id UUID REFERENCES accounts(id),
    balance DECIMAL(15,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- Indexes
CREATE INDEX idx_accounts_agent_id ON accounts(agent_id);
CREATE INDEX idx_accounts_type ON accounts(type);
CREATE INDEX idx_accounts_parent ON accounts(parent_account_id);
CREATE INDEX idx_accounts_active ON accounts(is_active);
CREATE INDEX idx_accounts_account_number ON accounts(account_number);
```

### Transactions Table (Double-Entry Bookkeeping)
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID REFERENCES payments(id),
    journal_entry_id UUID,
    debit_account_id UUID NOT NULL REFERENCES accounts(id),
    credit_account_id UUID NOT NULL REFERENCES accounts(id),
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1.0,
    description TEXT,
    transaction_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status VARCHAR(50) DEFAULT 'posted' CHECK (status IN ('pending', 'posted', 'reversed')),
    reversal_of UUID REFERENCES transactions(id),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID,
    updated_by UUID,

    -- Ensure debit and credit accounts are different
    CONSTRAINT different_accounts CHECK (debit_account_id != credit_account_id)
);

-- Indexes
CREATE INDEX idx_transactions_payment_id ON transactions(payment_id);
CREATE INDEX idx_transactions_debit_account ON transactions(debit_account_id);
CREATE INDEX idx_transactions_credit_account ON transactions(credit_account_id);
CREATE INDEX idx_transactions_date ON transactions(transaction_date);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_journal_entry ON transactions(journal_entry_id);
```

## Risk Management Schema

### Risk Profiles Table
```sql
CREATE TABLE risk_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    overall_score INTEGER NOT NULL CHECK (overall_score >= 0 AND overall_score <= 100),
    velocity_score INTEGER CHECK (velocity_score >= 0 AND velocity_score <= 100),
    geographic_score INTEGER CHECK (geographic_score >= 0 AND geographic_score <= 100),
    amount_score INTEGER CHECK (amount_score >= 0 AND amount_score <= 100),
    counterparty_score INTEGER CHECK (counterparty_score >= 0 AND counterparty_score <= 100),
    flags JSONB DEFAULT '[]',
    last_assessment TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    next_assessment TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_risk_profiles_agent_id ON risk_profiles(agent_id);
CREATE INDEX idx_risk_profiles_overall_score ON risk_profiles(overall_score);
CREATE INDEX idx_risk_profiles_last_assessment ON risk_profiles(last_assessment);
```

### Risk Alerts Table
```sql
CREATE TABLE risk_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID REFERENCES payments(id),
    agent_id UUID NOT NULL REFERENCES agents(id),
    alert_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    score INTEGER CHECK (score >= 0 AND score <= 100),
    threshold INTEGER CHECK (threshold >= 0 AND threshold <= 100),
    description TEXT,
    details JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'resolved', 'dismissed')),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_risk_alerts_payment_id ON risk_alerts(payment_id);
CREATE INDEX idx_risk_alerts_agent_id ON risk_alerts(agent_id);
CREATE INDEX idx_risk_alerts_type ON risk_alerts(alert_type);
CREATE INDEX idx_risk_alerts_severity ON risk_alerts(severity);
CREATE INDEX idx_risk_alerts_status ON risk_alerts(status);
CREATE INDEX idx_risk_alerts_created_at ON risk_alerts(created_at);
```

## Consent Management Schema

### Consents Table
```sql
CREATE TABLE consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    consent_type VARCHAR(100) NOT NULL,
    scope JSONB NOT NULL DEFAULT '{}',
    granted_to UUID REFERENCES agents(id),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'expired')),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_by UUID,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_consents_agent_id ON consents(agent_id);
CREATE INDEX idx_consents_type ON consents(consent_type);
CREATE INDEX idx_consents_status ON consents(status);
CREATE INDEX idx_consents_granted_to ON consents(granted_to);
CREATE INDEX idx_consents_expires_at ON consents(expires_at);
```

### Consent Requests Table
```sql
CREATE TABLE consent_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id),
    consent_type VARCHAR(100) NOT NULL,
    scope JSONB NOT NULL DEFAULT '{}',
    requested_by UUID NOT NULL REFERENCES agents(id),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
    approved_at TIMESTAMP WITH TIME ZONE,
    approved_by UUID,
    rejected_at TIMESTAMP WITH TIME ZONE,
    rejected_by UUID,
    expires_at TIMESTAMP WITH TIME ZONE,
    reason TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_consent_requests_agent_id ON consent_requests(agent_id);
CREATE INDEX idx_consent_requests_requested_by ON consent_requests(requested_by);
CREATE INDEX idx_consent_requests_status ON consent_requests(status);
CREATE INDEX idx_consent_requests_created_at ON consent_requests(created_at);
```

## Audit Schema

### Audit Events Table
```sql
CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    user_id UUID,
    session_id VARCHAR(255),
    action VARCHAR(50) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    location JSONB,
    risk_score INTEGER,
    compliance_flags JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_entity_type ON audit_events(entity_type);
CREATE INDEX idx_audit_events_entity_id ON audit_events(entity_id);
CREATE INDEX idx_audit_events_user_id ON audit_events(user_id);
CREATE INDEX idx_audit_events_created_at ON audit_events(created_at);
CREATE INDEX idx_audit_events_ip_address ON audit_events(ip_address);
```

### Compliance Reports Table
```sql
CREATE TABLE compliance_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_type VARCHAR(50) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'generating' CHECK (status IN ('generating', 'completed', 'failed')),
    data JSONB,
    summary JSONB,
    generated_at TIMESTAMP WITH TIME ZONE,
    generated_by UUID,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_compliance_reports_type ON compliance_reports(report_type);
CREATE INDEX idx_compliance_reports_period ON compliance_reports(period_start, period_end);
CREATE INDEX idx_compliance_reports_status ON compliance_reports(status);
CREATE INDEX idx_compliance_reports_generated_at ON compliance_reports(generated_at);
```

## Hash Chain Schema

### Hash Chain Blocks Table
```sql
CREATE TABLE hashchain_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    block_index BIGINT UNIQUE NOT NULL,
    previous_hash VARCHAR(64),
    current_hash VARCHAR(64) UNIQUE NOT NULL,
    transaction_count INTEGER DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    merkle_root VARCHAR(64),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'orphaned', 'invalidated')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_hashchain_blocks_index ON hashchain_blocks(block_index);
CREATE INDEX idx_hashchain_blocks_previous_hash ON hashchain_blocks(previous_hash);
CREATE INDEX idx_hashchain_blocks_current_hash ON hashchain_blocks(current_hash);
CREATE INDEX idx_hashchain_blocks_timestamp ON hashchain_blocks(timestamp);
CREATE INDEX idx_hashchain_blocks_status ON hashchain_blocks(status);
```

### Block Transactions Table
```sql
CREATE TABLE block_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    block_id UUID NOT NULL REFERENCES hashchain_blocks(id) ON DELETE CASCADE,
    transaction_id UUID NOT NULL,
    transaction_hash VARCHAR(64) NOT NULL,
    transaction_data JSONB NOT NULL,
    sequence_number INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(block_id, transaction_id)
);

-- Indexes
CREATE INDEX idx_block_transactions_block_id ON block_transactions(block_id);
CREATE INDEX idx_block_transactions_transaction_id ON block_transactions(transaction_id);
CREATE INDEX idx_block_transactions_hash ON block_transactions(transaction_hash);
```

## Supporting Tables

### API Keys Table
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    permissions JSONB DEFAULT '[]',
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'revoked')),
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_api_keys_agent_id ON api_keys(agent_id);
CREATE INDEX idx_api_keys_status ON api_keys(status);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);
```

### Webhooks Table
```sql
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    events JSONB DEFAULT '[]',
    secret VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'failed')),
    failure_count INTEGER DEFAULT 0,
    last_failure_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_webhooks_agent_id ON webhooks(agent_id);
CREATE INDEX idx_webhooks_status ON webhooks(status);
CREATE INDEX idx_webhooks_events ON webhooks USING GIN(events);
```

## Database Constraints and Triggers

### Balance Update Trigger
```sql
CREATE OR REPLACE FUNCTION update_account_balance()
RETURNS TRIGGER AS $$
BEGIN
    -- Update debit account balance
    UPDATE accounts
    SET balance = balance - NEW.amount,
        updated_at = NOW()
    WHERE id = NEW.debit_account_id;

    -- Update credit account balance
    UPDATE accounts
    SET balance = balance + NEW.amount,
        updated_at = NOW()
    WHERE id = NEW.credit_account_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_account_balance
    AFTER INSERT ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_account_balance();
```

### Audit Trigger
```sql
CREATE OR REPLACE FUNCTION audit_table_changes()
RETURNS TRIGGER AS $$
DECLARE
    old_row JSONB;
    new_row JSONB;
    changes JSONB := '{}';
BEGIN
    IF TG_OP = 'DELETE' THEN
        old_row := row_to_json(OLD)::JSONB;
        INSERT INTO audit_events (event_type, entity_type, entity_id, action, old_values)
        VALUES (TG_TABLE_NAME || '.deleted', TG_TABLE_NAME, OLD.id, 'DELETE', old_row);
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        old_row := row_to_json(OLD)::JSONB;
        new_row := row_to_json(NEW)::JSONB;

        -- Calculate changes
        FOR key IN SELECT jsonb_object_keys(new_row)
        LOOP
            IF old_row->key IS DISTINCT FROM new_row->key THEN
                changes := changes || jsonb_build_object(key, jsonb_build_object('old', old_row->key, 'new', new_row->key));
            END IF;
        END LOOP;

        INSERT INTO audit_events (event_type, entity_type, entity_id, action, old_values, new_values, changes)
        VALUES (TG_TABLE_NAME || '.updated', TG_TABLE_NAME, NEW.id, 'UPDATE', old_row, new_row, changes);
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        new_row := row_to_json(NEW)::JSONB;
        INSERT INTO audit_events (event_type, entity_type, entity_id, action, new_values)
        VALUES (TG_TABLE_NAME || '.created', TG_TABLE_NAME, NEW.id, 'CREATE', new_row);
        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
```

### Hash Chain Trigger
```sql
CREATE OR REPLACE FUNCTION maintain_hash_chain()
RETURNS TRIGGER AS $$
DECLARE
    prev_hash VARCHAR(64);
    block_data TEXT;
    current_hash VARCHAR(64);
BEGIN
    -- Get previous block hash
    SELECT current_hash INTO prev_hash
    FROM hashchain_blocks
    WHERE block_index = NEW.block_index - 1;

    -- Create block data string
    block_data := NEW.block_index || '|' || COALESCE(prev_hash, '') || '|' ||
                  NEW.transaction_count || '|' || NEW.total_amount || '|' ||
                  COALESCE(NEW.merkle_root, '') || '|' || NEW.timestamp;

    -- Generate current hash
    current_hash := encode(digest(block_data, 'sha256'), 'hex');

    -- Update the block with hashes
    NEW.previous_hash := prev_hash;
    NEW.current_hash := current_hash;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_maintain_hash_chain
    BEFORE INSERT ON hashchain_blocks
    FOR EACH ROW
    EXECUTE FUNCTION maintain_hash_chain();
```

## Performance Optimizations

### Partitioning Strategy
```sql
-- Partition audit_events by month
CREATE TABLE audit_events_y2025m09 PARTITION OF audit_events
    FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');

-- Partition payments by status and date
CREATE TABLE payments_active PARTITION OF payments
    FOR VALUES FROM ('active') TO ('processing')
    PARTITION BY RANGE (created_at);
```

### Materialized Views
```sql
-- Daily payment summary
CREATE MATERIALIZED VIEW daily_payment_summary AS
SELECT
    DATE(created_at) as date,
    COUNT(*) as total_payments,
    SUM(amount) as total_amount,
    AVG(amount) as avg_amount,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_payments
FROM payments
WHERE created_at >= CURRENT_DATE - INTERVAL '90 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Agent balance summary
CREATE MATERIALIZED VIEW agent_balance_summary AS
SELECT
    a.id,
    a.name,
    a.email,
    SUM(CASE WHEN acc.type = 'asset' THEN acc.balance ELSE 0 END) as total_assets,
    SUM(CASE WHEN acc.type = 'liability' THEN acc.balance ELSE 0 END) as total_liabilities,
    SUM(CASE WHEN acc.type = 'equity' THEN acc.balance ELSE 0 END) as total_equity
FROM agents a
LEFT JOIN accounts acc ON a.id = acc.agent_id
GROUP BY a.id, a.name, a.email;
```

## Backup and Recovery

### Automated Backup Strategy
```sql
-- Daily full backup
0 2 * * * pg_dump -U agentpay -h localhost agent_payments > /backups/daily/$(date +\%Y\%m\%d)_full.sql

-- Hourly incremental backup
0 * * * * pg_dump -U agentpay -h localhost --data-only --exclude-table=audit_events agent_payments > /backups/hourly/$(date +\%Y\%m\%d\%H)_incremental.sql
```

### Point-in-Time Recovery
```sql
-- Enable WAL archiving
ALTER SYSTEM SET wal_level = replica;
ALTER SYSTEM SET archive_mode = on;
ALTER SYSTEM SET archive_command = 'cp %p /archive/%f';

-- Create recovery configuration
restore_command = 'cp /archive/%f %p'
recovery_target_time = '2025-09-07 14:30:00'
```

---

*This database schema is maintained by DistributedApps.ai and Ken Huang. Last updated: September 7, 2025*
