# Monitoring Setup Guide

## Overview

This guide provides comprehensive instructions for setting up monitoring, observability, and alerting for the Agent Payment Platform. The platform uses a multi-layered monitoring approach combining metrics, logs, traces, and alerts to ensure high availability and performance.

## Monitoring Architecture

### Monitoring Stack
```
┌─────────────────────────────────────────────────────────────────┐
│                    Monitoring Architecture                       │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ Prometheus  │  │   Grafana  │  │  AlertManager│             │
│  │   Metrics   │  │ Dashboards │  │   Alerts    │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ ELK Stack  │  │   Jaeger    │  │   Loki      │             │
│  │  Logging    │  │   Tracing  │  │   Logs      │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Health    │  │   Metrics   │  │   Tracing   │             │
│  │   Checks    │  │   Export    │  │   Export    │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

## Prometheus Setup

### Prometheus Configuration
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'agentpay-identity'
    static_configs:
      - targets: ['identity:8081']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'agentpay-router'
    static_configs:
      - targets: ['router:8082']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'agentpay-ledger'
    static_configs:
      - targets: ['ledger:8083']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'agentpay-risk'
    static_configs:
      - targets: ['risk:8084']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'agentpay-gateway'
    static_configs:
      - targets: ['api-gateway:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres:9187']
    scrape_interval: 10s

  - job_name: 'redis'
    static_configs:
      - targets: ['redis:9121']
    scrape_interval: 10s
```

### Alert Rules
```yaml
# alert_rules.yml
groups:
  - name: agentpay
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }}% for {{ $labels.service }}"

      - alert: ServiceDown
        expr: up == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.job }} is down"
          description: "Service {{ $labels.job }} has been down for more than 2 minutes"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
          description: "95th percentile latency is {{ $value }}s for {{ $labels.service }}"

      - alert: DatabaseConnectionHigh
        expr: pg_stat_activity_count{datname="agent_payments"} > 50
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High database connections"
          description: "Database has {{ $value }} active connections"

      - alert: PaymentFailureRate
        expr: rate(payment_failures_total[5m]) / rate(payment_attempts_total[5m]) > 0.01
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High payment failure rate"
          description: "Payment failure rate is {{ $value }}%"

      - alert: RiskScoreHigh
        expr: avg_over_time(risk_score[5m]) > 75
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High average risk score"
          description: "Average risk score is {{ $value }} over last 5 minutes"
```

## Grafana Dashboards

### Application Dashboard
```json
{
  "dashboard": {
    "title": "AgentPay Application Metrics",
    "tags": ["agentpay", "application"],
    "timezone": "UTC",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{service}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m]) * 100",
            "legendFormat": "{{service}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "{{service}} 95th percentile"
          }
        ]
      }
    ]
  }
}
```

### Business Metrics Dashboard
```json
{
  "dashboard": {
    "title": "AgentPay Business Metrics",
    "tags": ["agentpay", "business"],
    "timezone": "UTC",
    "panels": [
      {
        "title": "Payment Volume",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(payment_attempts_total[5m])",
            "legendFormat": "Payment Attempts"
          },
          {
            "expr": "rate(payment_success_total[5m])",
            "legendFormat": "Payment Success"
          }
        ]
      },
      {
        "title": "Payment Success Rate",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(payment_success_total[1h]) / rate(payment_attempts_total[1h]) * 100",
            "format": "percent"
          }
        ]
      },
      {
        "title": "Average Risk Score",
        "type": "graph",
        "targets": [
          {
            "expr": "avg_over_time(risk_score[5m])",
            "legendFormat": "Risk Score"
          }
        ]
      }
    ]
  }
}
```

### Infrastructure Dashboard
```json
{
  "dashboard": {
    "title": "AgentPay Infrastructure",
    "tags": ["agentpay", "infrastructure"],
    "timezone": "UTC",
    "panels": [
      {
        "title": "CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(container_cpu_usage_seconds_total{pod=~\"agentpay-.*\"}[5m])",
            "legendFormat": "{{pod}}"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "container_memory_usage_bytes{pod=~\"agentpay-.*\"}",
            "legendFormat": "{{pod}}"
          }
        ]
      },
      {
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "pg_stat_activity_count{datname=\"agent_payments\"}",
            "legendFormat": "Active Connections"
          }
        ]
      }
    ]
  }
}
```

## ELK Stack Setup

### Elasticsearch Configuration
```yaml
# elasticsearch.yml
cluster.name: agentpay-cluster
node.name: ${HOSTNAME}
path.data: /usr/share/elasticsearch/data
path.logs: /usr/share/elasticsearch/logs

network.host: 0.0.0.0
http.port: 9200

discovery.type: single-node

xpack.security.enabled: true
xpack.security.transport.ssl.enabled: true
xpack.security.http.ssl.enabled: true
```

### Logstash Configuration
```conf
# logstash.conf
input {
  beats {
    port => 5044
  }
}

filter {
  if [kubernetes] {
    mutate {
      add_field => {
        "container_name" => "%{[kubernetes][container][name]}"
        "namespace" => "%{[kubernetes][namespace]}"
        "pod_name" => "%{[kubernetes][pod][name]}"
      }
    }
  }

  if [message] =~ /^{.*}$/ {
    json {
      source => "message"
    }
  }

  date {
    match => ["timestamp", "ISO8601"]
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "agentpay-%{+YYYY.MM.dd}"
    user => "elastic"
    password => "${ELASTIC_PASSWORD}"
  }
}
```

### Kibana Configuration
```yaml
# kibana.yml
server.name: kibana
server.host: "0"
elasticsearch.hosts: ["http://elasticsearch:9200"]
elasticsearch.username: "kibana_system"
elasticsearch.password: "${KIBANA_PASSWORD}"
xpack.security.enabled: true
xpack.encryptedSavedObjects.encryptionKey: "${ENCRYPTION_KEY}"
```

## Jaeger Tracing

### Jaeger Configuration
```yaml
# jaeger-config.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: jaeger-config
data:
  collector: |
    SPAN_STORAGE_TYPE=memory
    COLLECTOR_OTLP_ENABLED=true
    COLLECTOR_ZIPKIN_HOST_PORT=:9411

  query: |
    SPAN_STORAGE_TYPE=memory

  agent: |
    REPORTER_GRPC_HOST_PORT=jaeger-collector:14250
```

### Application Tracing Setup
```go
// tracing.go
import (
    "github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
    "github.com/uber/jaeger-client-go/config"
)

func InitTracing(serviceName string) (opentracing.Tracer, io.Closer) {
    cfg := &config.Configuration{
        ServiceName: serviceName,
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans:           true,
            LocalAgentHostPort: "jaeger-agent:6831",
        },
    }

    tracer, closer, err := cfg.NewTracer()
    if err != nil {
        log.Fatal("Failed to initialize tracer:", err)
    }

    opentracing.SetGlobalTracer(tracer)
    return tracer, closer
}

// Middleware for tracing HTTP requests
func TracingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        span := opentracing.StartSpan(c.Request.Method + " " + c.Request.URL.Path)
        defer span.Finish()

        span.SetTag("http.method", c.Request.Method)
        span.SetTag("http.url", c.Request.URL.Path)

        c.Next()

        span.SetTag("http.status_code", c.Writer.Status())
    }
}
```

## AlertManager Setup

### AlertManager Configuration
```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@agentpay.com'
  smtp_auth_username: 'alerts@agentpay.com'
  smtp_auth_password: '${SMTP_PASSWORD}'

route:
  group_by: ['alertname', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'team-email'
  routes:
  - match:
      severity: critical
    receiver: 'team-pager'
  - match:
      severity: warning
    receiver: 'team-slack'

receivers:
- name: 'team-email'
  email_configs:
  - to: 'team@distributedapps.ai'
    send_resolved: true

- name: 'team-pager'
  pagerduty_configs:
  - service_key: '${PAGERDUTY_KEY}'
    send_resolved: true

- name: 'team-slack'
  slack_configs:
  - api_url: '${SLACK_WEBHOOK}'
    channel: '#alerts'
    send_resolved: true
    title: '{{ .GroupLabels.alertname }}'
    text: '{{ .CommonAnnotations.description }}'
```

## Application Metrics

### Go Application Metrics
```go
// metrics.go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    paymentAttemptsTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "payment_attempts_total",
            Help: "Total number of payment attempts",
        },
    )

    paymentSuccessTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "payment_success_total",
            Help: "Total number of successful payments",
        },
    )

    riskScore = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "risk_score",
            Help: "Current risk score",
        },
    )
)

// Metrics middleware
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.Request.URL.Path,
        ))
        defer timer.ObserveDuration()

        c.Next()

        httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.Request.URL.Path,
            strconv.Itoa(c.Writer.Status()),
        ).Inc()
    }
}

// Metrics endpoint
func MetricsHandler() gin.HandlerFunc {
    h := promhttp.Handler()
    return func(c *gin.Context) {
        h.ServeHTTP(c.Writer, c.Request)
    }
}
```

### Business Metrics
```go
// business_metrics.go
func RecordPaymentAttempt(amount float64, currency string) {
    paymentAttemptsTotal.Inc()

    // Record payment amount
    paymentAmount.WithLabelValues(currency).Observe(amount)
}

func RecordPaymentSuccess(amount float64, currency string, processingTime time.Duration) {
    paymentSuccessTotal.Inc()

    // Record successful payment
    paymentSuccessAmount.WithLabelValues(currency).Observe(amount)
    paymentProcessingTime.Observe(processingTime.Seconds())
}

func UpdateRiskScore(score float64) {
    riskScore.Set(score)
}

func RecordAuditEvent(eventType, userID string) {
    auditEventsTotal.WithLabelValues(eventType, userID).Inc()
}
```

## Health Checks

### Application Health Checks
```go
// health.go
type HealthChecker struct {
    db     *sql.DB
    redis  *redis.Client
    kafka  *kafka.Client
}

func (hc *HealthChecker) Check() HealthStatus {
    status := HealthStatus{
        Status:  "healthy",
        Checks:  make(map[string]CheckResult),
        Version: version,
        Uptime:  time.Since(startTime).String(),
    }

    // Database check
    if err := hc.db.Ping(); err != nil {
        status.Status = "unhealthy"
        status.Checks["database"] = CheckResult{
            Status:  "unhealthy",
            Message: err.Error(),
        }
    } else {
        status.Checks["database"] = CheckResult{
            Status: "healthy",
        }
    }

    // Redis check
    if _, err := hc.redis.Ping().Result(); err != nil {
        status.Status = "degraded"
        status.Checks["redis"] = CheckResult{
            Status:  "unhealthy",
            Message: err.Error(),
        }
    } else {
        status.Checks["redis"] = CheckResult{
            Status: "healthy",
        }
    }

    // Kafka check
    if err := hc.kafka.HealthCheck(); err != nil {
        status.Status = "degraded"
        status.Checks["kafka"] = CheckResult{
            Status:  "unhealthy",
            Message: err.Error(),
        }
    } else {
        status.Checks["kafka"] = CheckResult{
            Status: "healthy",
        }
    }

    return status
}

// Health check endpoint
func HealthHandler(hc *HealthChecker) gin.HandlerFunc {
    return func(c *gin.Context) {
        status := hc.Check()

        if status.Status == "unhealthy" {
            c.JSON(503, status)
            return
        }

        c.JSON(200, status)
    }
}
```

### Kubernetes Health Checks
```yaml
# deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentpay-api-gateway
spec:
  template:
    spec:
      containers:
      - name: api-gateway
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        startupProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 6
```

## Log Aggregation

### Structured Logging
```go
// logger.go
import (
    "github.com/sirupsen/logrus"
    "github.com/google/uuid"
)

type Logger struct {
    *logrus.Logger
    serviceName string
}

func NewLogger(serviceName string) *Logger {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: time.RFC3339,
    })

    return &Logger{
        Logger:      logger,
        serviceName: serviceName,
    }
}

func (l *Logger) WithContext(correlationID, userID string) *logrus.Entry {
    return l.WithFields(logrus.Fields{
        "service":        l.serviceName,
        "correlation_id": correlationID,
        "user_id":        userID,
        "timestamp":      time.Now().UTC().Format(time.RFC3339),
    })
}

// Request logging middleware
func LoggingMiddleware(logger *Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        correlationID := c.GetHeader("X-Correlation-ID")
        if correlationID == "" {
            correlationID = uuid.New().String()
        }

        c.Set("correlation_id", correlationID)
        c.Header("X-Correlation-ID", correlationID)

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        logger.WithContext(correlationID, c.GetString("user_id")).WithFields(logrus.Fields{
            "method":      c.Request.Method,
            "path":        c.Request.URL.Path,
            "status":      status,
            "latency":     latency.String(),
            "ip":          c.ClientIP(),
            "user_agent":  c.Request.UserAgent(),
        }).Info("HTTP Request")
    }
}
```

## Performance Monitoring

### APM Integration
```go
// apm.go
import (
    "go.elastic.co/apm"
    "go.elastic.co/apm/module/apmgin"
    "go.elastic.co/apm/module/apmgorm"
)

func InitAPM(serviceName, serviceVersion string) {
    // Initialize APM
    tracer, err := apm.NewTracer(serviceName, serviceVersion)
    if err != nil {
        log.Fatal("Failed to initialize APM:", err)
    }
    defer tracer.Close()

    // Set as global tracer
    apm.SetGlobalTracer(tracer)
}

// APM middleware
func APMMiddleware() gin.HandlerFunc {
    return apmgin.Middleware()
}

// Database APM
func InitDBAPM(db *gorm.DB) {
    apmgorm.Instrument(db)
}
```

### Custom Metrics
```go
// custom_metrics.go
var (
    activeUsers = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "active_users_total",
        Help: "Number of active users",
    })

    paymentVolume = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "payment_volume_total",
            Help: "Total payment volume by currency",
        },
        []string{"currency"},
    )

    apiLatency = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "api_latency_seconds",
            Help:    "API request latency in seconds",
            Buckets: []float64{.001, .005, .01, .05, .1, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint", "status"},
    )
)

func RecordActiveUsers(count int) {
    activeUsers.Set(float64(count))
}

func RecordPaymentVolume(amount float64, currency string) {
    paymentVolume.WithLabelValues(currency).Add(amount)
}

func RecordAPILatency(method, endpoint string, status int, duration time.Duration) {
    apiLatency.WithLabelValues(method, endpoint, strconv.Itoa(status)).Observe(duration.Seconds())
}
```

## Alerting

### Alert Channels
```yaml
# alert-channels.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alert-channels
data:
  email: |
    subject: "AgentPay Alert: {{ .GroupLabels.alertname }}"
    body: |
      Alert: {{ .GroupLabels.alertname }}
      Severity: {{ .GroupLabels.severity }}
      Description: {{ .CommonAnnotations.description }}
      Value: {{ .Value }}
      Time: {{ .StartsAt }}

  slack: |
    {
      "channel": "#alerts",
      "username": "AgentPay Alert",
      "icon_emoji": ":warning:",
      "attachments": [
        {
          "color": "{{ if eq .GroupLabels.severity \"critical\" }}danger{{ else if eq .GroupLabels.severity \"warning\" }}warning{{ else }}good{{ end }}",
          "title": "{{ .GroupLabels.alertname }}",
          "text": "{{ .CommonAnnotations.description }}",
          "fields": [
            {
              "title": "Severity",
              "value": "{{ .GroupLabels.severity }}",
              "short": true
            },
            {
              "title": "Value",
              "value": "{{ .Value }}",
              "short": true
            }
          ]
        }
      ]
    }
```

### Monitoring Dashboard

### Main Dashboard
```json
{
  "dashboard": {
    "title": "AgentPay Monitoring Overview",
    "tags": ["agentpay", "overview"],
    "timezone": "UTC",
    "refresh": "30s",
    "panels": [
      {
        "title": "System Health",
        "type": "stat",
        "targets": [
          {
            "expr": "up",
            "legendFormat": "{{job}}"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "mappings": [
              {
                "options": {
                  "0": {
                    "text": "DOWN",
                    "color": "red"
                  },
                  "1": {
                    "text": "UP",
                    "color": "green"
                  }
                },
                "type": "value"
              }
            ]
          }
        }
      },
      {
        "title": "Active Alerts",
        "type": "stat",
        "targets": [
          {
            "expr": "ALERTS{alertstate=\"firing\"}",
            "legendFormat": "{{alertname}}"
          }
        ]
      },
      {
        "title": "Payment Success Rate",
        "type": "gauge",
        "targets": [
          {
            "expr": "rate(payment_success_total[5m]) / rate(payment_attempts_total[5m]) * 100",
            "format": "percent"
          }
        ]
      }
    ]
  }
}
```

## Docker Compose Monitoring

### Monitoring Stack
```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false

  alertmanager:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./monitoring/alertmanager.yml:/etc/alertmanager/alertmanager.yml
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.17.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  logstash:
    image: docker.elastic.co/logstash/logstash:7.17.0
    volumes:
      - ./monitoring/logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    depends_on:
      - elasticsearch

  kibana:
    image: docker.elastic.co/kibana/kibana:7.17.0
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch

volumes:
  prometheus_data:
  grafana_data:
  elasticsearch_data:
```

## Production Checklist

### Monitoring Setup
- [ ] Prometheus configured and running
- [ ] Grafana dashboards created
- [ ] AlertManager configured
- [ ] Application metrics exported
- [ ] Health checks implemented
- [ ] Log aggregation configured

### Alerting Setup
- [ ] Alert rules defined
- [ ] Notification channels configured
- [ ] Escalation policies defined
- [ ] On-call schedules established
- [ ] Alert testing completed

### Observability
- [ ] Distributed tracing enabled
- [ ] Log correlation implemented
- [ ] Performance monitoring active
- [ ] Business metrics tracked
- [ ] User experience monitoring

### Maintenance
- [ ] Monitoring documentation updated
- [ ] Alert response procedures documented
- [ ] Dashboard maintenance scheduled
- [ ] Metric retention policies defined
- [ ] Backup monitoring data configured

---

*This monitoring setup guide is maintained by DistributedApps.ai and Ken Huang. Last updated: September 7, 2025*
