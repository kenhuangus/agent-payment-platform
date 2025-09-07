# Docker Deployment Guide

## Overview

This guide provides comprehensive instructions for deploying the Agent Payment Platform using Docker and related container technologies. The platform is designed for easy containerization with optimized Docker images and orchestration support.

## Prerequisites

### System Requirements
- **Docker**: Version 20.10 or higher
- **Docker Compose**: Version 2.0 or higher
- **Memory**: Minimum 4GB RAM, recommended 8GB+
- **Storage**: Minimum 20GB free space
- **CPU**: Multi-core processor recommended

### Network Requirements
- **Ports**: 8080 (API), 5432 (PostgreSQL), 6379 (Redis)
- **DNS**: Resolvable hostnames for all services
- **Firewall**: Open required ports between containers

## Quick Start with Docker Compose

### Single-Command Deployment
```bash
# Clone the repository
git clone https://github.com/kenhuangus/agent-payment-platform.git
cd agent-payment-platform

# Start all services
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

### Docker Compose Configuration
```yaml
# docker-compose.yml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:13-alpine
    container_name: agentpay-postgres
    environment:
      POSTGRES_DB: agent_payments
      POSTGRES_USER: agentpay
      POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme123}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    networks:
      - agentpay-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U agentpay"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: agentpay-redis
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    networks:
      - agentpay-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Identity Service
  identity:
    build:
      context: .
      dockerfile: Dockerfile.identity
    container_name: agentpay-identity
    environment:
      - DATABASE_URL=postgresql://agentpay:${DB_PASSWORD:-changeme123}@postgres:5432/agent_payments
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET:-your-jwt-secret-key}
      - SERVICE_PORT=8081
    ports:
      - "8081:8081"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - agentpay-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Router Service
  router:
    build:
      context: .
      dockerfile: Dockerfile.router
    container_name: agentpay-router
    environment:
      - DATABASE_URL=postgresql://agentpay:${DB_PASSWORD:-changeme123}@postgres:5432/agent_payments
      - REDIS_URL=redis://redis:6379
      - SERVICE_PORT=8082
    ports:
      - "8082:8082"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - agentpay-network
    restart: unless-stopped

  # Ledger Service
  ledger:
    build:
      context: .
      dockerfile: Dockerfile.ledger
    container_name: agentpay-ledger
    environment:
      - DATABASE_URL=postgresql://agentpay:${DB_PASSWORD:-changeme123}@postgres:5432/agent_payments
      - REDIS_URL=redis://redis:6379
      - SERVICE_PORT=8083
    ports:
      - "8083:8083"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - agentpay-network
    restart: unless-stopped

  # Risk Service
  risk:
    build:
      context: .
      dockerfile: Dockerfile.risk
    container_name: agentpay-risk
    environment:
      - DATABASE_URL=postgresql://agentpay:${DB_PASSWORD:-changeme123}@postgres:5432/agent_payments
      - REDIS_URL=redis://redis:6379
      - SERVICE_PORT=8084
    ports:
      - "8084:8084"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - agentpay-network
    restart: unless-stopped

  # API Gateway
  api-gateway:
    build:
      context: .
      dockerfile: Dockerfile.gateway
    container_name: agentpay-gateway
    environment:
      - IDENTITY_SERVICE_URL=http://identity:8081
      - ROUTER_SERVICE_URL=http://router:8082
      - LEDGER_SERVICE_URL=http://ledger:8083
      - RISK_SERVICE_URL=http://risk:8084
      - SERVICE_PORT=8080
    ports:
      - "8080:8080"
    depends_on:
      - identity
      - router
      - ledger
      - risk
    networks:
      - agentpay-network
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:

networks:
  agentpay-network:
    driver: bridge
```

## Environment Configuration

### Environment Variables
```bash
# .env file
# Database Configuration
DB_PASSWORD=your-secure-database-password
DATABASE_URL=postgresql://agentpay:${DB_PASSWORD}@postgres:5432/agent_payments

# Redis Configuration
REDIS_URL=redis://redis:6379

# JWT Configuration
JWT_SECRET=your-256-bit-jwt-secret-key-here
JWT_EXPIRATION=15m
REFRESH_TOKEN_EXPIRATION=24h

# Service Configuration
SERVICE_PORT=8080
LOG_LEVEL=info
ENVIRONMENT=production

# External Services
STRIPE_API_KEY=sk_live_your_stripe_key_here
PLAID_CLIENT_ID=your_plaid_client_id
PLAID_SECRET=your_plaid_secret

# Security
ENCRYPTION_KEY=your-32-byte-encryption-key
API_KEY_SALT=your-api-key-salt

# Email Configuration (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
```

### Secrets Management
```yaml
# Docker secrets (recommended for production)
echo "your-db-password" | docker secret create db_password -
echo "your-jwt-secret" | docker secret create jwt_secret -

# Use secrets in docker-compose.yml
services:
  postgres:
    secrets:
      - db_password
secrets:
  db_password:
    external: true
```

## Custom Docker Images

### Multi-Stage Dockerfile
```dockerfile
# Dockerfile.identity
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o identity ./services/identity

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Copy binary from builder
COPY --from=builder /app/identity /app/identity

# Change ownership
RUN chown appuser:appuser /app/identity

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8081/health || exit 1

# Run the application
CMD ["/app/identity"]
```

### Optimized Image Features
```dockerfile
# Security hardening
FROM alpine:latest
RUN apk add --no-cache ca-certificates && \
    adduser -D -s /bin/sh appuser && \
    mkdir -p /app && \
    chown -R appuser:appuser /app

# Multi-stage build for smaller images
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /app/main
USER 1000
EXPOSE 8080
HEALTHCHECK CMD ["/app/main", "--health"]
CMD ["/app/main"]
```

## Docker Compose Overrides

### Development Environment
```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  postgres:
    ports:
      - "5432:5432"
    volumes:
      - ./scripts/dev-init.sql:/docker-entrypoint-initdb.d/dev-init.sql

  api-gateway:
    environment:
      - LOG_LEVEL=debug
      - ENVIRONMENT=development
    volumes:
      - ./logs:/app/logs
    ports:
      - "8080:8080"
```

### Production Environment
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  postgres:
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '1.0'
        reservations:
          memory: 1G
          cpus: '0.5'
    environment:
      - POSTGRES_PASSWORD_FILE=/run/secrets/db_password

  redis:
    command: redis-server --appendonly yes --maxmemory 512mb --maxmemory-policy allkeys-lru

  api-gateway:
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
    environment:
      - LOG_LEVEL=warn
      - ENVIRONMENT=production
```

## Kubernetes Deployment

### Kubernetes Manifests
```yaml
# k8s/deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentpay-api-gateway
  labels:
    app: agentpay
    component: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: agentpay
      component: api-gateway
  template:
    metadata:
      labels:
        app: agentpay
        component: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: kenhuangus/agent-payment-platform:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: redis-url
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### Service Definition
```yaml
# k8s/service.yml
apiVersion: v1
kind: Service
metadata:
  name: agentpay-api-gateway
  labels:
    app: agentpay
    component: api-gateway
spec:
  selector:
    app: agentpay
    component: api-gateway
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  type: LoadBalancer
```

### Ingress Configuration
```yaml
# k8s/ingress.yml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: agentpay-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - api.agentpay.com
    secretName: agentpay-tls
  rules:
  - host: api.agentpay.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: agentpay-api-gateway
            port:
              number: 80
```

## Monitoring and Logging

### Container Logs
```bash
# View all container logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f api-gateway

# Follow logs with timestamps
docker-compose logs -f --timestamps

# Export logs for analysis
docker-compose logs > logs_$(date +%Y%m%d_%H%M%S).txt
```

### Health Checks
```bash
# Check container health
docker-compose ps

# Manual health check
curl http://localhost:8080/health

# Database connectivity check
docker-compose exec postgres pg_isready -U agentpay

# Redis connectivity check
docker-compose exec redis redis-cli ping
```

### Resource Monitoring
```bash
# Container resource usage
docker stats

# Specific container stats
docker stats agentpay-api-gateway

# Disk usage
docker system df

# Clean up unused resources
docker system prune -a --volumes
```

## Backup and Recovery

### Database Backup
```bash
# Backup PostgreSQL data
docker-compose exec postgres pg_dump -U agentpay agent_payments > backup_$(date +%Y%m%d_%H%M%S).sql

# Backup with compression
docker-compose exec postgres pg_dump -U agentpay agent_payments | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec postgres pg_dump -U agentpay agent_payments > $BACKUP_DIR/backup_$DATE.sql
find $BACKUP_DIR -name "backup_*.sql" -mtime +7 -delete
```

### Database Restore
```bash
# Restore from backup
docker-compose exec -T postgres psql -U agentpay agent_payments < backup_20250907_120000.sql

# Restore from compressed backup
gunzip -c backup_20250907_120000.sql.gz | docker-compose exec -T postgres psql -U agentpay agent_payments
```

### Volume Backup
```bash
# Backup Docker volumes
docker run --rm -v agentpay_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup_$(date +%Y%m%d_%H%M%S).tar.gz -C /data .

# Restore Docker volumes
docker run --rm -v agentpay_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres_backup_20250907_120000.tar.gz -C /data
```

## Scaling and Performance

### Horizontal Scaling
```bash
# Scale specific services
docker-compose up -d --scale api-gateway=3
docker-compose up -d --scale router=2

# Auto-scaling with Docker Swarm
docker service scale agentpay_api-gateway=5
```

### Load Balancing
```yaml
# Load balancer configuration
version: '3.8'
services:
  loadbalancer:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - api-gateway
    networks:
      - agentpay-network
```

### Performance Optimization
```dockerfile
# Optimized Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -a -installsuffix cgo \
    -o main .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /app/main
EXPOSE 8080
CMD ["/app/main"]
```

## Security Best Practices

### Image Security
```dockerfile
# Security scanning
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache ca-certificates
# Add security scanning
# RUN trivy filesystem --exit-code 1 --no-progress /app

FROM alpine:latest
# Use non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser
# Minimal attack surface
RUN apk --no-cache add ca-certificates
```

### Secret Management
```yaml
# Docker secrets
services:
  api-gateway:
    secrets:
      - jwt_secret
      - db_password
    environment:
      - JWT_SECRET_FILE=/run/secrets/jwt_secret
      - DB_PASSWORD_FILE=/run/secrets/db_password

secrets:
  jwt_secret:
    file: ./secrets/jwt_secret.txt
  db_password:
    file: ./secrets/db_password.txt
```

### Network Security
```yaml
# Internal network only
services:
  postgres:
    networks:
      - internal
    ports: []  # No external ports

  api-gateway:
    networks:
      - internal
      - external
    ports:
      - "8080:8080"

networks:
  internal:
    internal: true
  external:
    driver: bridge
```

## Troubleshooting

### Common Issues

#### Container Won't Start
```bash
# Check container logs
docker-compose logs <service-name>

# Check container status
docker-compose ps

# Restart specific service
docker-compose restart <service-name>

# Rebuild and restart
docker-compose up -d --build <service-name>
```

#### Database Connection Issues
```bash
# Check database connectivity
docker-compose exec postgres pg_isready -U agentpay

# Check database logs
docker-compose logs postgres

# Reset database
docker-compose down -v
docker-compose up -d postgres
```

#### Memory Issues
```bash
# Check memory usage
docker stats

# Increase memory limits
docker-compose.yml:
services:
  api-gateway:
    deploy:
      resources:
        limits:
          memory: 1G
        reservations:
          memory: 512M
```

#### Port Conflicts
```bash
# Check port usage
netstat -tulpn | grep :8080

# Change port mapping
docker-compose.yml:
services:
  api-gateway:
    ports:
      - "8081:8080"  # Change host port
```

## Production Deployment Checklist

### Pre-Deployment
- [ ] Environment variables configured
- [ ] Secrets properly set up
- [ ] SSL certificates obtained
- [ ] Domain name configured
- [ ] DNS records updated
- [ ] Firewall rules configured

### Deployment Steps
- [ ] Docker images built and tested
- [ ] docker-compose.yml validated
- [ ] Database migrations prepared
- [ ] Backup of current system taken
- [ ] Deployment scripts tested
- [ ] Rollback plan prepared

### Post-Deployment
- [ ] Services started successfully
- [ ] Health checks passing
- [ ] Logs monitoring configured
- [ ] Performance monitoring active
- [ ] Backup strategy implemented
- [ ] Documentation updated

---

*This Docker deployment guide is maintained by DistributedApps.ai and Ken Huang. Last updated: September 7, 2025*
