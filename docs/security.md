# Security Model

## Overview

The Agent Payment Platform implements a comprehensive, multi-layered security architecture designed to protect sensitive financial data, prevent unauthorized access, and ensure compliance with industry standards including PCI DSS, SOX, and GDPR.

## Security Principles

### 1. Defense in Depth
- **Multiple Security Layers**: No single point of failure
- **Least Privilege**: Minimum required permissions
- **Fail-Safe Defaults**: Secure by default configuration
- **Zero Trust**: Never trust, always verify

### 2. Data Protection
- **Encryption at Rest**: AES-256 encryption for stored data
- **Encryption in Transit**: TLS 1.3 for all communications
- **Data Classification**: Appropriate protection based on sensitivity
- **Data Minimization**: Collect only necessary data

### 3. Access Control
- **Role-Based Access Control (RBAC)**: Granular permissions
- **Multi-Factor Authentication (MFA)**: Additional verification layer
- **Session Management**: Secure token lifecycle
- **Audit Logging**: Complete activity tracking

## Authentication & Authorization

### JWT Token Authentication

#### Token Structure
```json
{
  "sub": "agent-123",
  "exp": 1638360000,
  "iat": 1638273600,
  "iss": "agentpay",
  "aud": "api",
  "role": "agent",
  "permissions": [
    "payment.create",
    "payment.read",
    "account.read"
  ],
  "scope": "api",
  "jti": "unique-token-id"
}
```

#### Token Security Features
- **Short Expiration**: 15-minute token lifetime
- **Refresh Tokens**: Separate long-lived refresh tokens
- **Token Revocation**: Immediate invalidation capability
- **Signature Verification**: RSA-256 signature validation

### Multi-Factor Authentication (MFA)

#### MFA Implementation
```json
{
  "mfa_required": true,
  "mfa_methods": ["totp", "sms", "email"],
  "backup_codes": ["12345678", "87654321"],
  "recovery_email": "recovery@example.com"
}
```

#### MFA Flow
```
1. Primary Authentication (Username/Password)
2. MFA Challenge (TOTP/SMS/Email)
3. MFA Verification
4. Session Establishment
5. Continuous Verification (Risk-Based)
```

### API Key Authentication

#### Key Management
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY,
    agent_id UUID NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    permissions JSONB DEFAULT '[]',
    status VARCHAR(50) DEFAULT 'active',
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    usage_count INTEGER DEFAULT 0,
    ip_whitelist JSONB DEFAULT '[]',
    rate_limit INTEGER DEFAULT 1000,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Key Security Features
- **Hashed Storage**: bcrypt hashing of API keys
- **IP Whitelisting**: Restrict key usage to specific IPs
- **Rate Limiting**: Per-key rate limit enforcement
- **Usage Monitoring**: Track key usage patterns
- **Automatic Rotation**: Scheduled key rotation

## Data Security

### Encryption Standards

#### At Rest Encryption
```sql
-- Field-level encryption for sensitive data
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    encrypted_amount BYTEA,  -- AES-256 encrypted
    amount_hash VARCHAR(64), -- SHA-256 hash for integrity
    encryption_key_id UUID,  -- Reference to encryption key
    encrypted_at TIMESTAMP WITH TIME ZONE,
    decrypted_at TIMESTAMP WITH TIME ZONE
);

-- Encryption key management
CREATE TABLE encryption_keys (
    id UUID PRIMARY KEY,
    key_data BYTEA,  -- Encrypted master key
    algorithm VARCHAR(50) DEFAULT 'AES-256-GCM',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);
```

#### In Transit Encryption
- **TLS 1.3**: Latest TLS version with perfect forward secrecy
- **Certificate Pinning**: Prevent man-in-the-middle attacks
- **HSTS**: HTTP Strict Transport Security headers
- **Secure Cookies**: HttpOnly, Secure, SameSite flags

### Data Classification

#### Classification Levels
| Level | Description | Encryption | Access Control | Retention |
|-------|-------------|------------|----------------|-----------|
| Public | Marketing materials | None | Open | Unlimited |
| Internal | Business operations | TLS | Role-based | 7 years |
| Confidential | Customer data | AES-256 | MFA required | 7 years |
| Restricted | Financial data | AES-256 + HSM | Dual authorization | 7 years |
| Critical | Encryption keys | HSM-protected | Physical access | Unlimited |

#### Data Handling Procedures
```json
{
  "data_classification": "restricted",
  "encryption_required": true,
  "mfa_required": true,
  "audit_required": true,
  "retention_years": 7,
  "backup_frequency": "daily",
  "disposal_method": "cryptographic_erase"
}
```

## Network Security

### Firewall Configuration

#### Web Application Firewall (WAF)
```nginx
# WAF Rules for API Protection
location /v1/ {
    # Rate limiting
    limit_req zone=api burst=10 nodelay;

    # Block common attacks
    if ($request_uri ~* "(<|>|('|)|(%3C)|(%3E)|(%27)|(%22))") {
        return 403;
    }

    # SQL injection protection
    if ($query_string ~* "(union|select|insert|cast|declare|drop|update|md5|benchmark|script|javascript|vbscript|onload|onerror)") {
        return 403;
    }

    proxy_pass http://api_backend;
}
```

#### Network Segmentation
```
┌─────────────────────────────────────────────────┐
│                 DMZ Zone                        │
├─────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐             │
│  │   Load      │  │    WAF     │             │
│  │  Balancer   │  │            │             │
│  └─────────────┘  └─────────────┘             │
├─────────────────────────────────────────────────┤
│                 Application Zone                │
├─────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐             │
│  │   API       │  │   Web      │             │
│  │  Gateway    │  │   App      │             │
│  └─────────────┘  └─────────────┘             │
├─────────────────────────────────────────────────┤
│                 Data Zone                       │
├─────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐             │
│  │ PostgreSQL  │  │    Redis   │             │
│  │            │  │            │             │
│  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────┘
```

### DDoS Protection

#### Rate Limiting Implementation
```go
// Rate limiting middleware
func RateLimitMiddleware() gin.HandlerFunc {
    limiter := tollbooth.NewLimiter(10, nil) // 10 requests per second
    limiter.SetIPLookups([]string{"X-Real-IP", "X-Forwarded-For"})
    limiter.SetMethods([]string{"GET", "POST", "PUT", "DELETE"})

    return func(c *gin.Context) {
        httpError := tollbooth.LimitByRequest(limiter, c.Writer, c.Request)
        if httpError != nil {
            c.JSON(httpError.StatusCode, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": httpError.Message,
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

#### Cloudflare Integration
```json
{
  "rules": [
    {
      "description": "Block suspicious traffic",
      "expression": "(http.request.uri.path contains \"/admin\" and ip.src in $suspicious_ips)",
      "action": "block"
    },
    {
      "description": "Rate limit API calls",
      "expression": "(http.request.uri.path contains \"/v1/\")",
      "action": "rate_limit",
      "rate_limit": {
        "requests_per_period": 100,
        "period": 60
      }
    }
  ]
}
```

## Application Security

### Input Validation & Sanitization

#### Request Validation
```go
type PaymentRequest struct {
    Amount      decimal.Decimal `json:"amount" validate:"required,gt=0,lt=1000000"`
    Currency    string         `json:"currency" validate:"required,len=3,oneof=USD EUR GBP"`
    Description string         `json:"description" validate:"required,min=1,max=500"`
    Email       string         `json:"counterparty_email" validate:"required,email"`
}

// Validation middleware
func ValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        var req PaymentRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{
                "error": "Validation failed",
                "details": err.Error(),
            })
            c.Abort()
            return
        }

        // Sanitize input
        req.Description = sanitize.HTML(req.Description)

        c.Set("validated_request", req)
        c.Next()
    }
}
```

#### SQL Injection Prevention
```go
// Safe parameterized queries
func GetPaymentByID(ctx context.Context, paymentID string) (*Payment, error) {
    var payment Payment

    query := `
        SELECT id, agent_id, amount, status, created_at
        FROM payments
        WHERE id = $1 AND deleted_at IS NULL
    `

    err := db.QueryRowContext(ctx, query, paymentID).Scan(
        &payment.ID,
        &payment.AgentID,
        &payment.Amount,
        &payment.Status,
        &payment.CreatedAt,
    )

    return &payment, err
}
```

### XSS Protection

#### Content Security Policy
```http
Content-Security-Policy: default-src 'self';
                         script-src 'self' 'unsafe-inline' https://cdn.example.com;
                         style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
                         img-src 'self' data: https:;
                         font-src 'self' https://fonts.gstatic.com;
                         connect-src 'self' https://api.example.com;
                         frame-ancestors 'none';
```

#### Output Encoding
```go
// HTML entity encoding
func SanitizeOutput(input string) string {
    return html.EscapeString(input)
}

// JSON response sanitization
func SafeJSONResponse(data interface{}) gin.H {
    return gin.H{
        "data": data,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "request_id": generateRequestID(),
    }
}
```

## Audit & Compliance

### Comprehensive Audit Logging

#### Audit Event Structure
```json
{
  "id": "audit-123",
  "timestamp": "2025-09-07T12:00:00Z",
  "event_type": "payment.created",
  "entity_type": "payment",
  "entity_id": "pay-456",
  "user_id": "agent-123",
  "session_id": "sess-789",
  "action": "CREATE",
  "old_values": null,
  "new_values": {
    "amount": 1500.00,
    "status": "pending"
  },
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "location": {
    "country": "US",
    "city": "New York",
    "coordinates": [40.7128, -74.0060]
  },
  "risk_score": 15,
  "compliance_flags": ["pci_dss", "gdpr"],
  "correlation_id": "corr-123-456-789"
}
```

#### Audit Triggers
```sql
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_events (
        event_type,
        entity_type,
        entity_id,
        user_id,
        action,
        old_values,
        new_values,
        ip_address
    ) VALUES (
        TG_TABLE_NAME || '.' || LOWER(TG_OP),
        TG_TABLE_NAME,
        COALESCE(NEW.id, OLD.id),
        current_user_id(),
        TG_OP,
        CASE WHEN TG_OP != 'INSERT' THEN row_to_json(OLD) ELSE NULL END,
        CASE WHEN TG_OP != 'DELETE' THEN row_to_json(NEW) ELSE NULL END,
        current_ip_address()
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
```

### Compliance Reporting

#### SOX Compliance
```sql
-- SOX audit trail query
SELECT
    ae.timestamp,
    ae.event_type,
    ae.user_id,
    ae.action,
    ae.old_values,
    ae.new_values,
    ae.ip_address,
    u.name as user_name,
    u.role as user_role
FROM audit_events ae
LEFT JOIN users u ON ae.user_id = u.id
WHERE ae.timestamp >= '2025-01-01'
  AND ae.event_type LIKE 'payment.%'
ORDER BY ae.timestamp DESC;
```

#### PCI DSS Compliance
```sql
-- PCI DSS sensitive data access audit
SELECT
    ae.timestamp,
    ae.user_id,
    ae.action,
    ae.entity_type,
    ae.entity_id,
    CASE
        WHEN ae.entity_type = 'payment' THEN p.amount
        WHEN ae.entity_type = 'account' THEN a.balance
        ELSE NULL
    END as sensitive_data_accessed,
    ae.ip_address,
    ae.location
FROM audit_events ae
LEFT JOIN payments p ON ae.entity_id = p.id AND ae.entity_type = 'payment'
LEFT JOIN accounts a ON ae.entity_id = a.id AND ae.entity_type = 'account'
WHERE ae.timestamp >= CURRENT_DATE - INTERVAL '1 year'
  AND ae.entity_type IN ('payment', 'account')
ORDER BY ae.timestamp DESC;
```

## Incident Response

### Security Incident Response Plan

#### Phase 1: Detection & Assessment
```yaml
detection:
  - Automated monitoring alerts
  - User reports
  - Security scanning results
  - Third-party notifications

assessment:
  - Incident classification (severity, impact)
  - Evidence collection
  - Containment strategy
  - Communication plan
```

#### Phase 2: Containment
```yaml
containment:
  - Isolate affected systems
  - Block malicious traffic
  - Revoke compromised credentials
  - Implement emergency patches

short_term:
  - Temporary security measures
  - Traffic filtering
  - Access restrictions
  - Monitoring enhancement
```

#### Phase 3: Recovery
```yaml
recovery:
  - System restoration from backups
  - Security patch application
  - Configuration verification
  - Service testing

long_term:
  - Root cause analysis
  - Security improvements
  - Process updates
  - Training enhancements
```

### Incident Response Team
- **Security Lead**: Overall incident coordination
- **Technical Lead**: Technical response and containment
- **Legal Counsel**: Regulatory compliance and reporting
- **Communications**: Internal/external communications
- **Business Continuity**: Operational impact management

## Security Monitoring

### Real-Time Monitoring

#### SIEM Integration
```json
{
  "event": {
    "timestamp": "2025-09-07T12:00:00Z",
    "source": "agent-payment-platform",
    "severity": "high",
    "category": "authentication",
    "message": "Multiple failed login attempts",
    "details": {
      "user_id": "agent-123",
      "ip_address": "192.168.1.100",
      "attempts": 5,
      "time_window": "5 minutes"
    }
  }
}
```

#### Security Dashboards
```sql
-- Real-time security metrics
CREATE VIEW security_dashboard AS
SELECT
    DATE_TRUNC('hour', created_at) as hour,
    COUNT(CASE WHEN event_type = 'auth.failed' THEN 1 END) as failed_logins,
    COUNT(CASE WHEN event_type = 'payment.blocked' THEN 1 END) as blocked_payments,
    COUNT(CASE WHEN severity = 'high' THEN 1 END) as high_severity_events,
    COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_events,
    AVG(risk_score) as avg_risk_score
FROM audit_events
WHERE created_at >= CURRENT_DATE
GROUP BY DATE_TRUNC('hour', created_at)
ORDER BY hour DESC;
```

### Automated Alerts

#### Alert Rules
```yaml
alerts:
  - name: "Brute Force Attack"
    condition: "failed_logins > 10 AND time_window = '5 minutes'"
    severity: "high"
    action: "block_ip, notify_security_team"

  - name: "Unusual Payment Pattern"
    condition: "payment_amount > avg_amount * 3 AND same_user = true"
    severity: "medium"
    action: "require_additional_verification"

  - name: "Data Exfiltration Attempt"
    condition: "large_data_download AND unusual_time = true"
    severity: "critical"
    action: "immediate_block, security_incident_response"
```

## Compliance Standards

### PCI DSS Compliance
- **Requirement 1**: Network security controls
- **Requirement 2**: System password policies
- **Requirement 3**: Cardholder data protection
- **Requirement 4**: Encrypted transmission
- **Requirement 5**: Anti-malware protection
- **Requirement 6**: Secure application development
- **Requirement 7**: Access control
- **Requirement 8**: User identification
- **Requirement 9**: Physical access control
- **Requirement 10**: Logging and monitoring
- **Requirement 11**: Regular testing
- **Requirement 12**: Security policy

### GDPR Compliance
- **Data Protection by Design**: Privacy considerations in system design
- **Data Minimization**: Collect only necessary personal data
- **Purpose Limitation**: Clear purpose for data processing
- **Storage Limitation**: Data retention policies
- **Data Subject Rights**: Access, rectification, erasure rights
- **Breach Notification**: 72-hour breach reporting requirement
- **Data Protection Impact Assessment**: DPIA for high-risk processing

### SOX Compliance
- **Access Controls**: Prevent unauthorized access to financial data
- **Change Management**: Controlled system changes
- **Segregation of Duties**: Prevent single points of failure
- **Audit Trails**: Complete transaction logging
- **Financial Reporting**: Accurate financial statement generation

## Security Testing

### Penetration Testing
```bash
# Automated security scanning
nikto -h https://api.agentpay.com

# SQL injection testing
sqlmap -u "https://api.agentpay.com/v1/payments" --data="amount=100"

# XSS testing
xsser --url="https://app.agentpay.com" --auto
```

### Vulnerability Scanning
```yaml
# Vulnerability scan configuration
scanner:
  schedule: "0 2 * * *"  # Daily at 2 AM
  targets:
    - "https://api.agentpay.com"
    - "https://app.agentpay.com"
  checks:
    - ssl_expiry
    - certificate_validity
    - known_vulnerabilities
    - misconfigurations
    - exposed_services
  notifications:
    - email: security@distributedapps.ai
    - slack: "#security-alerts"
```

### Security Code Review
```go
// Security code review checklist
func SecurityReview(payment *Payment) error {
    // Input validation
    if payment.Amount <= 0 {
        return errors.New("invalid payment amount")
    }

    // Authorization check
    if !hasPermission(payment.AgentID, "payment.create") {
        return errors.New("insufficient permissions")
    }

    // Business rule validation
    if payment.Amount > getDailyLimit(payment.AgentID) {
        return errors.New("amount exceeds daily limit")
    }

    // Fraud detection
    if riskScore := calculateRiskScore(payment); riskScore > 80 {
        return errors.New("payment blocked by risk engine")
    }

    return nil
}
```

## Security Training & Awareness

### Employee Training Program
- **Security Awareness Training**: Annual mandatory training
- **Role-Specific Training**: Specialized training by job function
- **Incident Response Training**: Simulated security incidents
- **Policy Training**: Regular policy review and updates

### Third-Party Risk Management
```json
{
  "vendor_assessment": {
    "vendor_name": "Payment Processor Inc",
    "assessment_date": "2025-09-01",
    "risk_level": "low",
    "controls_reviewed": [
      "access_controls",
      "encryption",
      "incident_response",
      "business_continuity"
    ],
    "contractual_safeguards": [
      "data_processing_agreement",
      "security_requirements",
      "breach_notification",
      "right_to_audit"
    ]
  }
}
```

---

*This security model document is maintained by DistributedApps.ai and Ken Huang. Last updated: September 7, 2025*
