# API Design

## Overview

The Agent Payment Platform provides a comprehensive REST API designed for enterprise-grade payment processing. The API follows RESTful principles with consistent resource naming, HTTP status codes, and JSON responses.

## API Principles

### 1. RESTful Design
- **Resource-Based URLs**: Clear, hierarchical resource identification
- **HTTP Methods**: Proper use of GET, POST, PUT, DELETE, PATCH
- **Stateless**: Each request contains all necessary information
- **Uniform Interface**: Consistent request/response formats

### 2. Versioning
- **URL Versioning**: `/v1/` prefix for API versioning
- **Backward Compatibility**: New versions don't break existing clients
- **Deprecation Notices**: Advance warning for deprecated endpoints
- **Sunset Policies**: Clear migration timelines

### 3. Authentication
- **JWT Tokens**: Bearer token authentication
- **API Keys**: Service-to-service authentication
- **Multi-Factor**: Optional MFA for sensitive operations
- **Session Management**: Secure token lifecycle management

## Base URL

```
Production: https://api.agentpay.com/v1
Sandbox:     https://sandbox.agentpay.com/v1
```

## Authentication

### JWT Token Authentication

#### Request
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Token Structure
```json
{
  "sub": "agent-123",
  "exp": 1638360000,
  "iat": 1638273600,
  "iss": "agentpay",
  "aud": "api",
  "role": "agent",
  "permissions": ["payment.create", "payment.read"]
}
```

### API Key Authentication

#### Request
```http
X-API-Key: apikey_1234567890abcdef
```

## Core Resources

### Agents

#### Get Agent Profile
```http
GET /v1/agents/{id}
```

**Response:**
```json
{
  "id": "agent-123",
  "name": "John Doe",
  "email": "john@example.com",
  "status": "active",
  "kyc_status": "approved",
  "risk_score": 15,
  "created_at": "2025-09-01T10:00:00Z",
  "updated_at": "2025-09-06T14:30:00Z"
}
```

#### Update Agent Profile
```http
PUT /v1/agents/{id}
Content-Type: application/json

{
  "name": "John Smith",
  "phone": "+1-555-0123"
}
```

### Payments

#### Create Payment
```http
POST /v1/payments
Content-Type: application/json
Authorization: Bearer {token}

{
  "agent_id": "agent-123",
  "amount": 1500.00,
  "currency": "USD",
  "counterparty_email": "vendor@example.com",
  "description": "Office supplies",
  "rail": "auto",
  "metadata": {
    "invoice_number": "INV-2025-001",
    "department": "IT"
  }
}
```

**Response:**
```json
{
  "id": "pay-456",
  "status": "processing",
  "estimated_completion": "2025-09-07T12:05:00Z",
  "fee_amount": 2.50,
  "created_at": "2025-09-07T12:00:00Z"
}
```

#### Get Payment Status
```http
GET /v1/payments/{id}
```

**Response:**
```json
{
  "id": "pay-456",
  "agent_id": "agent-123",
  "amount": 1500.00,
  "currency": "USD",
  "status": "completed",
  "rail": "ach",
  "rail_transaction_id": "ach_789xyz",
  "completed_at": "2025-09-07T12:02:30Z",
  "fee_amount": 2.50,
  "risk_score": 15,
  "description": "Office supplies",
  "metadata": {
    "invoice_number": "INV-2025-001"
  }
}
```

#### List Payments
```http
GET /v1/payments?agent_id=agent-123&status=completed&limit=20&offset=0
```

**Response:**
```json
{
  "payments": [
    {
      "id": "pay-456",
      "amount": 1500.00,
      "status": "completed",
      "created_at": "2025-09-07T12:00:00Z"
    }
  ],
  "total_count": 1,
  "has_more": false
}
```

#### Cancel Payment
```http
DELETE /v1/payments/{id}
```

### Accounts

#### Get Account Balance
```http
GET /v1/accounts/{id}/balance
```

**Response:**
```json
{
  "account_id": "acc-789",
  "balance": 15750.00,
  "currency": "USD",
  "available_balance": 14750.00,
  "pending_transactions": 1000.00,
  "last_updated": "2025-09-07T12:00:00Z"
}
```

#### Get Transaction History
```http
GET /v1/accounts/{id}/transactions?start_date=2025-09-01&end_date=2025-09-07&limit=50
```

**Response:**
```json
{
  "transactions": [
    {
      "id": "txn-123",
      "payment_id": "pay-456",
      "amount": 1500.00,
      "type": "debit",
      "description": "Office supplies",
      "balance_after": 15750.00,
      "created_at": "2025-09-07T12:02:30Z"
    }
  ],
  "total_count": 1,
  "has_more": false
}
```

### Risk Assessment

#### Evaluate Payment Risk
```http
POST /v1/risk/evaluate
Content-Type: application/json

{
  "agent_id": "agent-123",
  "amount": 25000.00,
  "counterparty_email": "supplier@corp.com",
  "description": "Equipment purchase"
}
```

**Response:**
```json
{
  "risk_score": 35,
  "risk_level": "medium",
  "factors": {
    "amount_score": 60,
    "velocity_score": 25,
    "geographic_score": 10,
    "counterparty_score": 45
  },
  "recommendations": [
    "Additional verification required",
    "Consider splitting into smaller transactions"
  ],
  "expires_at": "2025-09-07T13:00:00Z"
}
```

#### Get Risk Alerts
```http
GET /v1/risk/alerts?agent_id=agent-123&severity=high&status=active
```

**Response:**
```json
{
  "alerts": [
    {
      "id": "alert-789",
      "type": "high_velocity",
      "severity": "high",
      "score": 85,
      "description": "Unusual transaction velocity detected",
      "payment_id": "pay-456",
      "created_at": "2025-09-07T12:00:00Z"
    }
  ],
  "total_count": 1
}
```

### Consent Management

#### Create Consent Request
```http
POST /v1/consents
Content-Type: application/json

{
  "agent_id": "agent-123",
  "consent_type": "payment_approval",
  "scope": {
    "max_amount": 5000,
    "currency": "USD",
    "valid_counterparties": ["vendor@example.com"],
    "expires_in_days": 30
  },
  "description": "Monthly vendor payment approval"
}
```

**Response:**
```json
{
  "id": "consent-101",
  "status": "pending",
  "expires_at": "2025-10-07T12:00:00Z",
  "created_at": "2025-09-07T12:00:00Z"
}
```

#### Approve Consent Request
```http
POST /v1/consents/{id}/approve
Content-Type: application/json

{
  "approved_by": "admin-456",
  "notes": "Approved for monthly vendor payments"
}
```

### Audit & Compliance

#### Query Audit Events
```http
GET /v1/audit/events?entity_type=payment&action=created&start_date=2025-09-01&end_date=2025-09-07
```

**Response:**
```json
{
  "events": [
    {
      "id": "audit-123",
      "event_type": "payment.created",
      "entity_type": "payment",
      "entity_id": "pay-456",
      "user_id": "agent-123",
      "action": "created",
      "old_values": null,
      "new_values": {
        "amount": 1500.00,
        "status": "pending"
      },
      "ip_address": "192.168.1.100",
      "created_at": "2025-09-07T12:00:00Z"
    }
  ],
  "total_count": 1,
  "has_more": false
}
```

#### Generate Compliance Report
```http
POST /v1/audit/compliance-reports
Content-Type: application/json

{
  "report_type": "pci_dss",
  "period_start": "2025-09-01",
  "period_end": "2025-09-07",
  "format": "pdf"
}
```

**Response:**
```json
{
  "report_id": "report-789",
  "status": "generating",
  "estimated_completion": "2025-09-07T12:05:00Z",
  "download_url": "https://api.agentpay.com/v1/reports/report-789/download"
}
```

## Error Handling

### Standard Error Response
```json
{
  "error": {
    "code": "PAYMENT_FAILED",
    "message": "Payment processing failed",
    "details": "Insufficient funds in account",
    "correlation_id": "corr-123-456-789",
    "timestamp": "2025-09-07T12:00:00Z"
  }
}
```

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict |
| 422 | Unprocessable Entity | Validation failed |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service temporarily unavailable |

### Error Codes

| Code | Description |
|------|-------------|
| `INVALID_REQUEST` | Malformed request |
| `AUTHENTICATION_FAILED` | Invalid credentials |
| `INSUFFICIENT_PERMISSIONS` | Access denied |
| `RESOURCE_NOT_FOUND` | Entity doesn't exist |
| `PAYMENT_FAILED` | Payment processing error |
| `RISK_BLOCKED` | Payment blocked by risk engine |
| `CONSENT_REQUIRED` | Authorization needed |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `SERVICE_UNAVAILABLE` | Temporary service outage |

## Rate Limiting

### Rate Limit Headers
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1638360000
X-RateLimit-Retry-After: 60
```

### Rate Limit Response
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded",
    "details": "Too many requests",
    "retry_after": 60
  }
}
```

## Pagination

### Standard Pagination
```http
GET /v1/payments?limit=20&offset=0
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "total_count": 150,
    "limit": 20,
    "offset": 0,
    "has_more": true,
    "next_offset": 20,
    "prev_offset": null
  }
}
```

### Cursor-Based Pagination
```http
GET /v1/audit/events?cursor=eyJpZCI6ImF1ZGl0LTEyMyJ9&limit=50
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "has_more": true,
    "next_cursor": "eyJpZCI6ImF1ZGl0LTE3MyJ9",
    "prev_cursor": null
  }
}
```

## Webhooks

### Webhook Configuration
```http
POST /v1/webhooks
Content-Type: application/json

{
  "url": "https://example.com/webhooks/agentpay",
  "events": ["payment.completed", "payment.failed", "risk.alert"],
  "secret": "whsec_webhook_secret_key",
  "description": "Payment status notifications"
}
```

### Webhook Payload
```json
{
  "id": "wh-123",
  "event_type": "payment.completed",
  "created_at": "2025-09-07T12:02:30Z",
  "data": {
    "payment": {
      "id": "pay-456",
      "amount": 1500.00,
      "status": "completed",
      "completed_at": "2025-09-07T12:02:30Z"
    }
  },
  "signature": "v1,g0hM9SsE+OTPJTGt/tmIKtSyZlE3uFJELVlNIOLJ1OE="
}
```

### Webhook Signature Verification
```javascript
const crypto = require('crypto');

function verifyWebhook(payload, signature, secret) {
  const expectedSignature = crypto
    .createHmac('sha256', secret)
    .update(payload, 'utf8')
    .digest('hex');

  return signature === `v1,${expectedSignature}`;
}
```

## SDKs and Libraries

### Official SDKs

#### Go SDK
```go
import "github.com/agentpay/agentpay-go"

client := agentpay.NewClient(&agentpay.Config{
    APIKey: "your-api-key",
    BaseURL: "https://api.agentpay.com/v1",
})

// Create payment
payment, err := client.Payments.Create(&agentpay.PaymentRequest{
    Amount: 1500.00,
    CounterpartyEmail: "vendor@example.com",
    Description: "Office supplies",
})
```

#### JavaScript SDK
```javascript
import { AgentPay } from '@agentpay/sdk';

const client = new AgentPay({
    apiKey: 'your-api-key',
    baseURL: 'https://api.agentpay.com/v1'
});

// Create payment
const payment = await client.payments.create({
    amount: 1500.00,
    counterpartyEmail: 'vendor@example.com',
    description: 'Office supplies'
});
```

#### Python SDK
```python
from agentpay import Client

client = Client(
    api_key='your-api-key',
    base_url='https://api.agentpay.com/v1'
)

# Create payment
payment = client.payments.create(
    amount=1500.00,
    counterparty_email='vendor@example.com',
    description='Office supplies'
)
```

## API Versioning

### Version Headers
```http
Accept: application/vnd.agentpay.v1+json
X-API-Version: v1
```

### Version Deprecation
```http
Warning: 299 - "API v1 deprecated, migrate to v2 by 2026-01-01"
X-API-Deprecation-Date: 2026-01-01
X-API-Sunset-Date: 2026-06-01
```

## Testing

### Sandbox Environment
```bash
# Use sandbox for testing
curl -X POST https://sandbox.agentpay.com/v1/payments \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"amount": 100.00, "counterparty_email": "test@example.com"}'
```

### Test Data
- **Test Cards**: Use `4242424242424242` for successful payments
- **Test Amounts**: Amounts ending in `.00` succeed, `.01` fail
- **Test Emails**: Use `test@agentpay.com` for testing
- **Test Webhooks**: Use webhook testing tools like ngrok

## Support

### Documentation
- **API Reference**: https://docs.agentpay.com/api
- **SDK Documentation**: https://docs.agentpay.com/sdks
- **Integration Guides**: https://docs.agentpay.com/guides

### Support Channels
- **Email**: api-support@distributedapps.ai
- **GitHub Issues**: For SDK and API issues
- **Status Page**: https://status.agentpay.com
- **Community Forum**: https://community.agentpay.com

---

*This API design document is maintained by DistributedApps.ai and Ken Huang. Last updated: September 7, 2025*
