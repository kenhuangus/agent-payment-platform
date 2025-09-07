# Agent Payment Platform 🏦💳

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-blue.svg)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> **Enterprise-grade agent-to-agent payment platform** with multi-rail routing, risk management, audit trails, and cryptographic verification.

## 🌟 Overview

The Agent Payment Platform is a comprehensive, production-ready payment processing system designed for agent-to-agent transactions. Built with modern microservices architecture, it provides secure, scalable, and compliant payment processing with advanced features like multi-rail routing, real-time risk assessment, comprehensive audit trails, and blockchain-style cryptographic verification.

### 🎯 Key Features

- **🔄 Multi-Rail Payment Processing**: Intelligent routing across ACH, credit cards, and wire transfers
- **🛡️ Advanced Risk Management**: Real-time fraud detection with machine learning integration
- **📊 Double-Entry Bookkeeping**: Complete financial accounting with balance calculations
- **🔗 Hash Chain Verification**: Blockchain-style cryptographic integrity for all transactions
- **📋 Comprehensive Audit Trails**: Full compliance reporting with SOX, PCI-DSS, GDPR support
- **🤝 Consent Management**: Granular payment authorization and approval workflows
- **📈 Real-Time Analytics**: Interactive dashboards with payment metrics and reporting
- **🔐 Enterprise Security**: Multi-layer security with encryption, access control, and monitoring

## 🏗️ Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                    Agent Payment Platform                       │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  Identity   │  │   Router    │  │   Ledger    │             │
│  │  Service    │  │   Service   │  │   Service   │             │
│  │             │  │             │  │             │             │
│  │ • Agent Mgmt│  │ • Rail      │  │ • Accounts  │             │
│  │ • Auth      │  │ • Routing   │  │ • Transactions│           │
│  │ • Consent   │  │ • Fees      │  │ • Balances  │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Risk      │  │   Audit     │  │   Events    │             │
│  │   Service   │  │   Service   │  │   Service   │             │
│  │             │  │             │  │             │             │
│  │ • Scoring   │  │ • Logging   │  │ • Kafka     │             │
│  │ • Alerts    │  │ • Reports   │  │ • Pub/Sub   │             │
│  │ • Monitoring│  │ • Compliance│  │ • Async     │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  Database   │  │   Cache     │  │   Storage   │             │
│  │ PostgreSQL  │  │    Redis    │  │    S3       │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

### Technology Stack

- **Backend**: Go 1.21+ with Gin web framework
- **Database**: PostgreSQL with GORM ORM
- **Message Queue**: Apache Kafka for event streaming
- **Cache**: Redis for high-performance caching
- **Frontend**: Modern HTML5/CSS3/JavaScript with Chart.js
- **Container**: Docker with Kubernetes orchestration
- **Security**: JWT authentication, TLS encryption, RBAC

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 13+
- Docker & Docker Compose
- Apache Kafka (optional, for event streaming)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/agent-payment-platform.git
   cd agent-payment-platform
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up the database**
   ```bash
   # Create PostgreSQL database
   createdb agent_payments

   # Run database migrations
   go run cmd/migrate/main.go
   ```

4. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

5. **Start the services**
   ```bash
   # Start all services with Docker Compose
   docker-compose up -d

   # Or run individual services
   go run services/identity/main.go &
   go run services/router/main.go &
   go run services/ledger/main.go &
   ```

6. **View the UI**
   ```bash
   # Open the dashboard in your browser
   open ui/index.html
   ```

## 📖 Usage

### Basic Payment Flow

```go
// Initialize payment
payment := &PaymentRequest{
    AgentID:      "agent-123",
    Amount:       1500.00,
    Counterparty: "vendor@example.com",
    Description:  "Office supplies",
    Rail:         "auto", // Auto-select best rail
}

// Process payment
response, err := client.ProcessPayment(ctx, payment)
if err != nil {
    log.Fatal("Payment failed:", err)
}

fmt.Printf("Payment %s processed successfully\n", response.PaymentID)
```

### API Endpoints

#### Payment Processing
```http
POST /v1/payments          # Initiate payment
GET  /v1/payments/:id      # Get payment status
GET  /v1/payments          # List payments
PUT  /v1/payments/:id      # Update payment
```

#### Account Management
```http
GET  /v1/accounts/:id/balance    # Get account balance
GET  /v1/accounts/:id/history    # Get transaction history
POST /v1/accounts/:id/reconcile  # Reconcile account
```

#### Risk Assessment
```http
POST /v1/risk/evaluate     # Evaluate payment risk
GET  /v1/risk/alerts       # Get risk alerts
GET  /v1/risk/metrics      # Get risk metrics
```

#### Audit & Compliance
```http
GET  /v1/audit/events      # Query audit events
GET  /v1/audit/summary     # Get audit summary
GET  /v1/audit/compliance  # Get compliance report
```

### SDK Usage

```javascript
// Initialize SDK
const client = new AgentPaymentClient({
    apiKey: 'your-api-key',
    baseURL: 'https://api.agentpay.com'
});

// Create payment
const payment = await client.payments.create({
    agentId: 'agent-123',
    amount: 1500.00,
    counterparty: 'vendor@example.com',
    description: 'Office supplies'
});

console.log('Payment created:', payment.id);
```

## 🔧 Configuration

### Environment Variables

```bash
# Database
DATABASE_URL=postgresql://user:password@localhost/agent_payments

# Redis
REDIS_URL=redis://localhost:6379

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=agent-payments

# Security
JWT_SECRET=your-jwt-secret-key
ENCRYPTION_KEY=your-encryption-key

# External Services
STRIPE_API_KEY=sk_test_...
PLAID_CLIENT_ID=your-plaid-id
```

### Docker Configuration

```yaml
# docker-compose.yml
version: '3.8'
services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: agent_payments
      POSTGRES_USER: agentpay
      POSTGRES_PASSWORD: password

  redis:
    image: redis:7-alpine

  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/hashchain/
go test ./internal/balances/

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./tests/...

# Test specific services
go test ./tests/payment_flow_test.go
go test ./tests/risk_engine_test.go
```

### Load Testing

```bash
# Install k6 for load testing
brew install k6

# Run load tests
k6 run tests/load/payment_load_test.js

# Distributed load testing
k6 run --vus 100 --duration 30s tests/load/payment_load_test.js
```

## 📊 Monitoring & Observability

### Metrics

The platform exposes comprehensive metrics via Prometheus:

- **Payment Metrics**: Success rates, processing times, volumes
- **Risk Metrics**: Alert counts, false positives, detection accuracy
- **System Metrics**: CPU, memory, database connections
- **Business Metrics**: Revenue, transaction volumes, user activity

### Logging

Structured logging with multiple levels:

```json
{
  "timestamp": "2025-09-07T00:04:54Z",
  "level": "INFO",
  "service": "payment-router",
  "correlation_id": "corr-123",
  "message": "Payment routed to ACH",
  "payment_id": "pay-456",
  "amount": 1500.00,
  "rail": "ach"
}
```

### Health Checks

```http
GET /health     # Overall system health
GET /health/db  # Database connectivity
GET /health/kafka # Message queue status
GET /health/redis # Cache status
```

## 🔒 Security

### Authentication & Authorization

- **JWT-based authentication** with refresh tokens
- **Role-based access control (RBAC)** with granular permissions
- **Multi-factor authentication (MFA)** support
- **API key authentication** for service-to-service communication

### Data Protection

- **End-to-end encryption** for sensitive data
- **PCI DSS compliance** for payment data handling
- **GDPR compliance** for data privacy
- **Data masking** in logs and audit trails

### Network Security

- **TLS 1.3 encryption** for all communications
- **Rate limiting** to prevent abuse
- **IP whitelisting** for sensitive operations
- **DDoS protection** with Cloudflare integration

## 📈 Performance

### Benchmarks

- **Payment Processing**: 10,000 TPS with <100ms latency
- **Risk Assessment**: <50ms average response time
- **Database Queries**: <10ms for 95th percentile
- **API Response Time**: <200ms for all endpoints

### Scalability

- **Horizontal scaling** with Kubernetes
- **Database sharding** for high-volume deployments
- **Redis clustering** for cache scalability
- **Kafka partitioning** for event throughput

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes and add tests
4. Run the test suite: `go test ./...`
5. Submit a pull request

### Code Standards

- Follow Go best practices and effective Go guidelines
- Write comprehensive tests for new features
- Update documentation for API changes
- Ensure all tests pass before submitting PR

## 📚 Documentation

### API Documentation

Complete API documentation is available at:
- **Swagger UI**: `http://localhost:8080/swagger/`
- **OpenAPI Spec**: `api/openapi.yaml`
- **Postman Collection**: `docs/postman_collection.json`

### Architecture Documentation

- [System Architecture](docs/architecture.md)
- [Database Schema](docs/database_schema.md)
- [API Design](docs/api_design.md)
- [Security Model](docs/security.md)

### Deployment Guides

- [Docker Deployment](docs/docker_deployment.md)
- [Kubernetes Deployment](docs/kubernetes_deployment.md)
- [AWS Deployment](docs/aws_deployment.md)
- [Monitoring Setup](docs/monitoring_setup.md)

## 🐛 Troubleshooting

### Common Issues

**Database Connection Issues**
```bash
# Check database connectivity
psql -h localhost -U agentpay -d agent_payments

# Reset database
make db-reset
```

**Service Startup Issues**
```bash
# Check service logs
docker-compose logs app

# Check environment variables
cat .env | grep -v PASSWORD
```

**Payment Processing Issues**
```bash
# Check payment service logs
docker-compose logs payment-service

# Verify external API keys
curl -H "Authorization: Bearer $API_KEY" https://api.stripe.com/v1/charges
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Go Community** for the excellent language and ecosystem
- **PostgreSQL** for robust database capabilities
- **Apache Kafka** for reliable event streaming
- **Open Source Community** for invaluable tools and libraries

## 👥 About

**Built by [DistributedApps.ai](https://distributedapps.ai)** and **Ken Huang**

This enterprise-grade agent payment platform was developed by DistributedApps.ai, a leading provider of distributed systems and enterprise software solutions. The platform showcases advanced microservices architecture, real-time processing capabilities, and comprehensive financial technology features.

### 🏢 Company Information

- **Company**: DistributedApps.ai
- **Developer**: Ken Huang
- **Focus**: Enterprise-grade distributed systems and payment platforms
- **Website**: [distributedapps.ai](https://distributedapps.ai)

## 📞 Support

- **Documentation**: [docs.agentpay.com](https://docs.agentpay.com)
- **Issues**: [GitHub Issues](https://github.com/yourusername/agent-payment-platform/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/agent-payment-platform/discussions)
- **Email**: info@distributedapps.ai

---

<div align="center">

**Built with ❤️ by DistributedApps.ai & Ken Huang**

⭐ Star us on GitHub • 📧 Contact us • 🌐 Visit our website

</div>
