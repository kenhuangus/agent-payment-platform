# Kubernetes Deployment Guide

## Overview

This guide provides comprehensive instructions for deploying the Agent Payment Platform on Kubernetes. The platform is designed for cloud-native deployment with horizontal scaling, high availability, and automated operations.

## Prerequisites

### Kubernetes Cluster Requirements
- **Kubernetes**: Version 1.24 or higher
- **Helm**: Version 3.8 or higher
- **kubectl**: Configured for cluster access
- **Storage Class**: For persistent volumes
- **Load Balancer**: For external access
- **Ingress Controller**: NGINX or Traefik

### Cluster Resources
- **CPU**: Minimum 4 cores, recommended 8+ cores
- **Memory**: Minimum 8GB RAM, recommended 16GB+
- **Storage**: Minimum 50GB persistent storage
- **Network**: Calico or Flannel CNI

### Required Tools
```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl && sudo mv kubectl /usr/local/bin/

# Install Helm
curl https://get.helm.sh/helm-v3.8.0-linux-amd64.tar.gz -o helm.tar.gz
tar -zxvf helm.tar.gz && sudo mv linux-amd64/helm /usr/local/bin/

# Verify installations
kubectl version --client
helm version
```

## Namespace Setup

### Create Namespace
```bash
# Create namespace for the application
kubectl create namespace agent-payment

# Set as default namespace for current context
kubectl config set-context --current --namespace=agent-payment
```

### Namespace Labels
```yaml
# namespace.yml
apiVersion: v1
kind: Namespace
metadata:
  name: agent-payment
  labels:
    name: agent-payment
    environment: production
    managed-by: distributedapps-ai
    created-by: ken-huang
```

## Secrets Management

### Create Secrets
```bash
# Database credentials
kubectl create secret generic db-secret \
  --from-literal=database-url="postgresql://agentpay:secure-password@postgres:5432/agent_payments" \
  --from-literal=db-password="secure-password"

# JWT secrets
kubectl create secret generic jwt-secret \
  --from-literal=jwt-secret="your-256-bit-jwt-secret-key-here"

# Redis credentials
kubectl create secret generic redis-secret \
  --from-literal=redis-url="redis://redis:6379"

# API keys
kubectl create secret generic api-keys \
  --from-literal=stripe-key="sk_live_your_stripe_key" \
  --from-literal=plaid-client-id="your_plaid_client_id" \
  --from-literal=plaid-secret="your_plaid_secret"
```

### TLS Certificates
```bash
# Create TLS secret for HTTPS
kubectl create secret tls agentpay-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key

# Or use cert-manager for automatic certificates
kubectl apply -f cert-manager-issuer.yml
```

## PostgreSQL Deployment

### PostgreSQL StatefulSet
```yaml
# postgres-statefulset.yml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  labels:
    app: postgres
    component: database
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
        component: database
    spec:
      containers:
      - name: postgres
        image: postgres:13-alpine
        ports:
        - containerPort: 5432
          name: postgres
        env:
        - name: POSTGRES_DB
          value: "agent_payments"
        - name: POSTGRES_USER
          value: "agentpay"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: db-password
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        - name: postgres-init
          mountPath: /docker-entrypoint-initdb.d
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - agentpay
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - agentpay
          initialDelaySeconds: 5
          periodSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: standard
      resources:
        requests:
          storage: 50Gi
  - metadata:
      name: postgres-init
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: standard
      resources:
        requests:
          storage: 1Gi
```

### PostgreSQL Service
```yaml
# postgres-service.yml
apiVersion: v1
kind: Service
metadata:
  name: postgres
  labels:
    app: postgres
    component: database
spec:
  selector:
    app: postgres
  ports:
  - name: postgres
    port: 5432
    targetPort: 5432
  clusterIP: None  # Headless service for StatefulSet
```

## Redis Deployment

### Redis Deployment
```yaml
# redis-deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  labels:
    app: redis
    component: cache
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
        component: cache
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
          name: redis
        command: ["redis-server", "--appendonly", "yes", "--maxmemory", "512mb", "--maxmemory-policy", "allkeys-lru"]
        volumeMounts:
        - name: redis-storage
          mountPath: /data
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - redis-cli
            - ping
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: redis-storage
        persistentVolumeClaim:
          claimName: redis-pvc
```

### Redis Service
```yaml
# redis-service.yml
apiVersion: v1
kind: Service
metadata:
  name: redis
  labels:
    app: redis
    component: cache
spec:
  selector:
    app: redis
  ports:
  - name: redis
    port: 6379
    targetPort: 6379
```

## Application Deployments

### Identity Service
```yaml
# identity-deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: identity-service
  labels:
    app: agentpay
    component: identity
spec:
  replicas: 2
  selector:
    matchLabels:
      app: agentpay
      component: identity
  template:
    metadata:
      labels:
        app: agentpay
        component: identity
    spec:
      containers:
      - name: identity
        image: kenhuangus/agent-payment-platform:identity-v1.0.0
        ports:
        - containerPort: 8081
          name: http
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
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: jwt-secret
        - name: SERVICE_PORT
          value: "8081"
        - name: ENVIRONMENT
          value: "production"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Router Service
```yaml
# router-deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: router-service
  labels:
    app: agentpay
    component: router
spec:
  replicas: 3
  selector:
    matchLabels:
      app: agentpay
      component: router
  template:
    metadata:
      labels:
        app: agentpay
        component: router
    spec:
      containers:
      - name: router
        image: kenhuangus/agent-payment-platform:router-v1.0.0
        ports:
        - containerPort: 8082
          name: http
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
        - name: STRIPE_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: stripe-key
        - name: SERVICE_PORT
          value: "8082"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8082
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Ledger Service
```yaml
# ledger-deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ledger-service
  labels:
    app: agentpay
    component: ledger
spec:
  replicas: 2
  selector:
    matchLabels:
      app: agentpay
      component: ledger
  template:
    metadata:
      labels:
        app: agentpay
        component: ledger
    spec:
      containers:
      - name: ledger
        image: kenhuangus/agent-payment-platform:ledger-v1.0.0
        ports:
        - containerPort: 8083
          name: http
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
        - name: SERVICE_PORT
          value: "8083"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8083
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8083
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Risk Service
```yaml
# risk-deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: risk-service
  labels:
    app: agentpay
    component: risk
spec:
  replicas: 2
  selector:
    matchLabels:
      app: agentpay
      component: risk
  template:
    metadata:
      labels:
        app: agentpay
        component: risk
    spec:
      containers:
      - name: risk
        image: kenhuangus/agent-payment-platform:risk-v1.0.0
        ports:
        - containerPort: 8084
          name: http
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
        - name: SERVICE_PORT
          value: "8084"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8084
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8084
          initialDelaySeconds: 5
          periodSeconds: 5
```

### API Gateway
```yaml
# api-gateway-deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
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
        image: kenhuangus/agent-payment-platform:gateway-v1.0.0
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: IDENTITY_SERVICE_URL
          value: "http://identity-service:8081"
        - name: ROUTER_SERVICE_URL
          value: "http://router-service:8082"
        - name: LEDGER_SERVICE_URL
          value: "http://ledger-service:8083"
        - name: RISK_SERVICE_URL
          value: "http://risk-service:8084"
        - name: SERVICE_PORT
          value: "8080"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
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
```

## Services

### Internal Services
```yaml
# internal-services.yml
apiVersion: v1
kind: Service
metadata:
  name: identity-service
  labels:
    app: agentpay
    component: identity
spec:
  selector:
    app: agentpay
    component: identity
  ports:
  - name: http
    port: 8081
    targetPort: 8081
  type: ClusterIP

---
apiVersion: v1
kind: Service
metadata:
  name: router-service
  labels:
    app: agentpay
    component: router
spec:
  selector:
    app: agentpay
    component: router
  ports:
  - name: http
    port: 8082
    targetPort: 8082
  type: ClusterIP

---
apiVersion: v1
kind: Service
metadata:
  name: ledger-service
  labels:
    app: agentpay
    component: ledger
spec:
  selector:
    app: agentpay
    component: ledger
  ports:
  - name: http
    port: 8083
    targetPort: 8083
  type: ClusterIP

---
apiVersion: v1
kind: Service
metadata:
  name: risk-service
  labels:
    app: agentpay
    component: risk
spec:
  selector:
    app: agentpay
    component: risk
  ports:
  - name: http
    port: 8084
    targetPort: 8084
  type: ClusterIP
```

### External Service
```yaml
# external-service.yml
apiVersion: v1
kind: Service
metadata:
  name: api-gateway-external
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
  - name: https
    port: 443
    targetPort: 8080
    protocol: TCP
  type: LoadBalancer
  loadBalancerIP: "YOUR_LOAD_BALANCER_IP"
```

## Ingress

### NGINX Ingress
```yaml
# ingress.yml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: agentpay-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "300"
spec:
  tls:
  - hosts:
    - api.agentpay.com
    - app.agentpay.com
    secretName: agentpay-tls
  rules:
  - host: api.agentpay.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway-external
            port:
              number: 80
  - host: app.agentpay.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-app
            port:
              number: 80
```

## ConfigMaps

### Application Configuration
```yaml
# configmap.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agentpay-config
  labels:
    app: agentpay
data:
  LOG_LEVEL: "info"
  ENVIRONMENT: "production"
  JWT_EXPIRATION: "15m"
  REFRESH_TOKEN_EXPIRATION: "24h"
  RATE_LIMIT_REQUESTS: "1000"
  RATE_LIMIT_WINDOW: "1h"
  CACHE_TTL: "300"
  DATABASE_MAX_CONNECTIONS: "20"
  REDIS_MAX_CONNECTIONS: "10"
  AUDIT_RETENTION_DAYS: "2555"
  BACKUP_RETENTION_DAYS: "30"
```

## Persistent Volumes

### PostgreSQL PVC
```yaml
# postgres-pvc.yml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  labels:
    app: postgres
    component: database
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: standard
  resources:
    requests:
      storage: 50Gi
```

### Redis PVC
```yaml
# redis-pvc.yml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-pvc
  labels:
    app: redis
    component: cache
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: standard
  resources:
    requests:
      storage: 10Gi
```

## Horizontal Pod Autoscaling

### API Gateway HPA
```yaml
# hpa.yml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-gateway-hpa
  labels:
    app: agentpay
    component: api-gateway
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

## Network Policies

### Database Network Policy
```yaml
# network-policy.yml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: database-network-policy
  labels:
    app: agentpay
spec:
  podSelector:
    matchLabels:
      component: database
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: agentpay
    ports:
    - protocol: TCP
      port: 5432
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

## Monitoring and Logging

### Prometheus ServiceMonitor
```yaml
# prometheus-servicemonitor.yml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: agentpay-servicemonitor
  labels:
    app: agentpay
    release: prometheus
spec:
  selector:
    matchLabels:
      app: agentpay
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
```

### Fluent Bit ConfigMap
```yaml
# fluent-bit-config.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  labels:
    app: fluent-bit
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush         5
        Log_Level     info
        Daemon        off

    [INPUT]
        Name              tail
        Path              /var/log/containers/*agentpay*.log
        Parser            docker
        Tag               agentpay.*
        Refresh_Interval  5

    [OUTPUT]
        Name  elasticsearch
        Match agentpay.*
        Host  elasticsearch-master
        Port  9200
        Index agentpay
```

## Deployment Strategy

### Rolling Updates
```yaml
# rolling-update.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    spec:
      containers:
      - name: api-gateway
        image: kenhuangus/agent-payment-platform:gateway-v1.0.0
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
```

### Blue-Green Deployment
```yaml
# blue-green-deployment.yml
apiVersion: v1
kind: Service
metadata:
  name: api-gateway-blue-green
spec:
  selector:
    app: agentpay
    component: api-gateway
    version: v1.0.0  # Change to v1.1.0 for green deployment
  ports:
  - name: http
    port: 80
    targetPort: 8080
  type: ClusterIP
```

## Backup and Recovery

### Database Backup CronJob
```yaml
# backup-cronjob.yml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  labels:
    app: postgres
    component: backup
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: postgres-backup
            image: postgres:13-alpine
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h postgres -U agentpay agent_payments | gzip > /backup/backup_$(date +%Y%m%d_%H%M%S).sql.gz
              # Clean up old backups (keep last 30 days)
              find /backup -name "backup_*.sql.gz" -mtime +30 -delete
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: db-password
            volumeMounts:
            - name: backup-storage
              mountPath: /backup
          volumes:
          - name: backup-storage
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
```

## Security

### Pod Security Standards
```yaml
# pod-security.yml
apiVersion: v1
kind: Pod
metadata:
  name: secure-pod
  labels:
    app: agentpay
spec:
  securityContext:
    runAsUser: 1000
    runAsGroup: 1000
    fsGroup: 1000
    runAsNonRoot: true
  containers:
  - name: app
    image: kenhuangus/agent-payment-platform:latest
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      runAsUser: 1000
      capabilities:
        drop:
        - ALL
    volumeMounts:
    - name: tmp-volume
      mountPath: /tmp
  volumes:
  - name: tmp-volume
    emptyDir: {}
```

### RBAC Configuration
```yaml
# rbac.yml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: agentpay-role
  namespace: agent-payment
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: agentpay-rolebinding
  namespace: agent-payment
subjects:
- kind: ServiceAccount
  name: agentpay-sa
  namespace: agent-payment
roleRef:
  kind: Role
  name: agentpay-role
  apiGroup: rbac.authorization.k8s.io
```

## Helm Chart

### Chart Structure
```
agent-payment/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── _helpers.tpl
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── postgres/
│   │   ├── statefulset.yaml
│   │   └── service.yaml
│   ├── redis/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── identity/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── router/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── ledger/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── risk/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── api-gateway/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── ingress.yaml
│   └── hpa.yaml
└── charts/
```

### Installation
```bash
# Add repository
helm repo add agentpay https://kenhuangus.github.io/agent-payment-platform
helm repo update

# Install chart
helm install agentpay agentpay/agent-payment-platform \
  --namespace agent-payment \
  --create-namespace \
  --values values-production.yaml

# Upgrade
helm upgrade agentpay agentpay/agent-payment-platform \
  --namespace agent-payment \
  --values values-production.yaml

# Uninstall
helm uninstall agentpay --namespace agent-payment
```

## Troubleshooting

### Common Issues

#### Pod CrashLoopBackOff
```bash
# Check pod logs
kubectl logs -f pod/agentpay-api-gateway-12345 -n agent-payment

# Check pod events
kubectl describe pod agentpay-api-gateway-12345 -n agent-payment

# Check resource usage
kubectl top pods -n agent-payment
```

#### Service Unavailable
```bash
# Check service endpoints
kubectl get endpoints -n agent-payment

# Check service configuration
kubectl describe service api-gateway -n agent-payment

# Test service connectivity
kubectl exec -it agentpay-api-gateway-12345 -n agent-payment -- curl http://localhost:8080/health
```

#### Database Connection Issues
```bash
# Check database pod status
kubectl get pods -l app=postgres -n agent-payment

# Check database logs
kubectl logs -f postgres-0 -n agent-payment

# Test database connectivity
kubectl exec -it postgres-0 -n agent-payment -- psql -U agentpay -d agent_payments -c "SELECT version();"
```

#### Ingress Issues
```bash
# Check ingress status
kubectl get ingress -n agent-payment

# Check ingress controller logs
kubectl logs -f nginx-ingress-controller-12345 -n ingress-nginx

# Test ingress configuration
curl -H "Host: api.agentpay.com" http://YOUR_LOAD_BALANCER_IP
```

## Performance Optimization

### Resource Optimization
```yaml
# Resource optimization
apiVersion: apps/v1
kind: Deployment
metadata:
  name: optimized-deployment
spec:
  template:
    spec:
      containers:
      - name: app
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        env:
        - name: GOMAXPROCS
          value: "2"
        - name: GOGC
          value: "100"
```

### Affinity and Anti-Affinity
```yaml
# Pod affinity rules
apiVersion: apps/v1
kind: Deployment
metadata:
  name: affinity-deployment
spec:
  template:
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - agentpay
            topologyKey: kubernetes.io/hostname
        podAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - postgres
              topologyKey: kubernetes.io/hostname
```

## Production Checklist

### Pre-Deployment
- [ ] Kubernetes cluster configured
- [ ] Helm installed and configured
- [ ] Storage classes available
- [ ] Load balancer configured
- [ ] SSL certificates obtained
- [ ] DNS records configured
- [ ] Secrets created
- [ ] ConfigMaps prepared

### Deployment
- [ ] Namespace created
- [ ] Secrets applied
- [ ] ConfigMaps applied
- [ ] PostgreSQL deployed and healthy
- [ ] Redis deployed and healthy
- [ ] Services deployed in order
- [ ] Ingress configured
- [ ] SSL/TLS enabled

### Post-Deployment
- [ ] All pods running and healthy
- [ ] Services accessible
- [ ] Database migrations completed
- [ ] Monitoring configured
- [ ] Logging configured
- [ ] Backup strategy implemented
- [ ] Security policies applied

### Validation
- [ ] API endpoints responding
- [ ] Database connections working
- [ ] Authentication functioning
- [ ] Payment processing working
- [ ] Monitoring dashboards accessible
- [ ] Log aggregation working

---

*This Kubernetes deployment guide is maintained by DistributedApps.ai and Ken Huang. Last updated: September 7, 2025*
