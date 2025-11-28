# AURA Testnet Faucet - Deployment Guide

This guide provides step-by-step instructions for deploying the AURA Testnet Faucet to various environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development Setup](#local-development-setup)
3. [Docker Deployment](#docker-deployment)
4. [Production Deployment](#production-deployment)
5. [Kubernetes Deployment](#kubernetes-deployment)
6. [Configuration Management](#configuration-management)
7. [Monitoring and Logging](#monitoring-and-logging)
8. [Backup and Recovery](#backup-and-recovery)
9. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software
- Docker 20.10+ and Docker Compose 2.0+
- Go 1.23.1+ (for local development)
- Git
- Access to an AURA testnet node
- hCaptcha account (production only)

### Required Accounts
- hCaptcha account for bot protection
- Domain name (for production)
- SSL certificate (for production)

### Network Requirements
- Port 8080 (faucet backend)
- Port 5432 (PostgreSQL, internal)
- Port 6379 (Redis, internal)
- Port 80/443 (nginx, production)
- Access to AURA node RPC endpoint

## Local Development Setup

### Step 1: Clone and Setup

```bash
cd aura/faucet-service
cp .env.example .env
```

### Step 2: Configure Environment

Edit `.env`:
```env
# Development configuration
ENVIRONMENT=development
NODE_RPC=http://localhost:26657
CHAIN_ID=aura-local-1
FAUCET_MNEMONIC=your-test-mnemonic-here
FAUCET_ADDRESS=aura1testaddress...
HCAPTCHA_SECRET=  # Leave empty for development
```

### Step 3: Start Dependencies

```bash
docker-compose up -d postgres redis
```

Wait for services to be healthy:
```bash
docker-compose ps
```

### Step 4: Run Backend

```bash
cd backend
go mod download
go run main.go
```

### Step 5: Access Application

- Frontend: http://localhost:8080
- API: http://localhost:8080/api/v1
- Health: http://localhost:8080/api/v1/health

## Docker Deployment

### Development Environment

1. **Prepare Configuration**:
```bash
cp .env.example .env
# Edit .env with development values
```

2. **Build and Start**:
```bash
docker-compose build
docker-compose up -d
```

3. **Verify Deployment**:
```bash
docker-compose ps
docker-compose logs -f faucet-backend
```

4. **Test Endpoints**:
```bash
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/faucet/info
```

### Staging Environment

1. **Configure for Staging**:
```env
ENVIRONMENT=production
NODE_RPC=https://rpc.staging.aura-chain.com
CHAIN_ID=aura-staging-1
LOG_LEVEL=debug
RATE_LIMIT_PER_IP=20
RATE_LIMIT_PER_ADDRESS=2
```

2. **Deploy**:
```bash
docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d
```

## Production Deployment

### Pre-deployment Checklist

#### Security
- [ ] Generate strong passwords for all services
- [ ] Obtain SSL certificates (Let's Encrypt recommended)
- [ ] Configure firewall rules
- [ ] Set up secrets management
- [ ] Enable database encryption
- [ ] Configure network security groups

#### Configuration
- [ ] Update all environment variables
- [ ] Configure hCaptcha keys
- [ ] Set correct AURA node RPC endpoint
- [ ] Configure CORS origins
- [ ] Set appropriate rate limits
- [ ] Configure log levels

#### Infrastructure
- [ ] Provision server/cloud instance
- [ ] Set up monitoring
- [ ] Configure alerts
- [ ] Set up log aggregation
- [ ] Configure backup system
- [ ] Set up load balancer (if needed)

### Step 1: Server Setup

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Create deploy user
sudo useradd -m -s /bin/bash faucet
sudo usermod -aG docker faucet
```

### Step 2: Deploy Application

```bash
# Switch to deploy user
sudo su - faucet

# Clone repository
git clone https://github.com/aura-chain/aura.git
cd aura/faucet-service

# Configure environment
cp .env.example .env
nano .env  # Edit with production values
```

### Step 3: Configure SSL

```bash
# Create SSL directory
mkdir -p ssl

# Using Let's Encrypt (recommended)
sudo apt install certbot
sudo certbot certonly --standalone -d faucet.aura-chain.com

# Copy certificates
sudo cp /etc/letsencrypt/live/faucet.aura-chain.com/fullchain.pem ssl/cert.pem
sudo cp /etc/letsencrypt/live/faucet.aura-chain.com/privkey.pem ssl/key.pem
sudo chown -R faucet:faucet ssl/
```

### Step 4: Deploy with Nginx

```bash
# Update nginx.conf with your domain
nano nginx.conf

# Start all services with production profile
docker-compose --profile production up -d

# Verify deployment
docker-compose ps
docker-compose logs -f
```

### Step 5: Verify Production Deployment

```bash
# Check health
curl https://faucet.aura-chain.com/api/v1/health

# Check SSL
curl -I https://faucet.aura-chain.com

# Test faucet info
curl https://faucet.aura-chain.com/api/v1/faucet/info
```

### Production Environment Variables

```env
# Production configuration
ENVIRONMENT=production
PORT=8080
CORS_ORIGINS=https://faucet.aura-chain.com

# AURA Node
NODE_RPC=https://rpc.aura-testnet.com
CHAIN_ID=aura-testnet-1
FAUCET_MNEMONIC=production-mnemonic-use-secrets-manager
FAUCET_ADDRESS=aura1productionaddress...
DENOM=uaura
AMOUNT_PER_REQUEST=100000000

# Database (use strong password)
DATABASE_URL=postgres://faucet:STRONG_RANDOM_PASSWORD@postgres:5432/faucet?sslmode=require

# Redis
REDIS_URL=redis://redis:6379/0

# Rate Limiting (adjust based on needs)
RATE_LIMIT_PER_IP=10
RATE_LIMIT_PER_ADDRESS=1
RATE_LIMIT_WINDOW_HOURS=24

# Security
HCAPTCHA_SECRET=your-production-hcaptcha-secret

# Transaction
GAS_LIMIT=200000
GAS_PRICE=0.025uaura
TRANSACTION_MEMO=AURA Testnet Faucet

# Logging
LOG_LEVEL=info
```

## Kubernetes Deployment

### Prerequisites
- Kubernetes cluster (1.20+)
- kubectl configured
- Helm 3+ (optional)

### Step 1: Create Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: aura-faucet
```

```bash
kubectl apply -f namespace.yaml
```

### Step 2: Create Secrets

```bash
# Create secret for environment variables
kubectl create secret generic faucet-secrets \
  --from-literal=faucet-mnemonic='your-mnemonic-here' \
  --from-literal=hcaptcha-secret='your-hcaptcha-secret' \
  --from-literal=database-password='strong-password' \
  -n aura-faucet

# Create secret for TLS
kubectl create secret tls faucet-tls \
  --cert=ssl/cert.pem \
  --key=ssl/key.pem \
  -n aura-faucet
```

### Step 3: Deploy PostgreSQL

```yaml
# postgres-deployment.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: aura-faucet
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: aura-faucet
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        env:
        - name: POSTGRES_USER
          value: "faucet"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: faucet-secrets
              key: database-password
        - name: POSTGRES_DB
          value: "faucet"
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: aura-faucet
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

### Step 4: Deploy Redis

```yaml
# redis-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: aura-faucet
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: aura-faucet
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
```

### Step 5: Deploy Faucet Backend

```yaml
# faucet-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: faucet-backend
  namespace: aura-faucet
spec:
  replicas: 3
  selector:
    matchLabels:
      app: faucet-backend
  template:
    metadata:
      labels:
        app: faucet-backend
    spec:
      containers:
      - name: faucet
        image: aura-faucet:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: NODE_RPC
          value: "https://rpc.aura-testnet.com"
        - name: CHAIN_ID
          value: "aura-testnet-1"
        - name: FAUCET_MNEMONIC
          valueFrom:
            secretKeyRef:
              name: faucet-secrets
              key: faucet-mnemonic
        - name: HCAPTCHA_SECRET
          valueFrom:
            secretKeyRef:
              name: faucet-secrets
              key: hcaptcha-secret
        - name: DATABASE_URL
          value: "postgres://faucet:$(DATABASE_PASSWORD)@postgres:5432/faucet?sslmode=disable"
        - name: REDIS_URL
          value: "redis://redis:6379/0"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: faucet-backend
  namespace: aura-faucet
spec:
  selector:
    app: faucet-backend
  ports:
  - port: 8080
    targetPort: 8080
```

### Step 6: Configure Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: faucet-ingress
  namespace: aura-faucet
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "10"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - faucet.aura-chain.com
    secretName: faucet-tls
  rules:
  - host: faucet.aura-chain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: faucet-backend
            port:
              number: 8080
```

### Step 7: Deploy All Components

```bash
kubectl apply -f postgres-deployment.yaml
kubectl apply -f redis-deployment.yaml
kubectl apply -f faucet-deployment.yaml
kubectl apply -f ingress.yaml
```

### Step 8: Verify Kubernetes Deployment

```bash
# Check pods
kubectl get pods -n aura-faucet

# Check services
kubectl get svc -n aura-faucet

# Check ingress
kubectl get ingress -n aura-faucet

# View logs
kubectl logs -f deployment/faucet-backend -n aura-faucet
```

## Configuration Management

### Using Environment Variables

Best for:
- Development
- Single-instance deployments
- Quick testing

Example:
```bash
export NODE_RPC=http://localhost:26657
export CHAIN_ID=aura-testnet-1
go run main.go
```

### Using .env Files

Best for:
- Local development
- Team development
- Docker Compose deployments

Example:
```bash
cp .env.example .env
# Edit .env
docker-compose up -d
```

### Using Secrets Management

Best for:
- Production environments
- Multiple environments
- Team deployments

#### AWS Secrets Manager

```bash
# Store secret
aws secretsmanager create-secret \
  --name aura-faucet/prod/mnemonic \
  --secret-string "your-mnemonic-here"

# Retrieve in application
export FAUCET_MNEMONIC=$(aws secretsmanager get-secret-value \
  --secret-id aura-faucet/prod/mnemonic \
  --query SecretString \
  --output text)
```

#### HashiCorp Vault

```bash
# Store secret
vault kv put secret/aura-faucet/prod \
  mnemonic="your-mnemonic-here" \
  hcaptcha_secret="your-secret"

# Retrieve in application
vault kv get -field=mnemonic secret/aura-faucet/prod
```

## Monitoring and Logging

### Application Monitoring

#### Prometheus Setup

```yaml
# prometheus-config.yaml
scrape_configs:
  - job_name: 'aura-faucet'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

#### Grafana Dashboard

Key metrics to monitor:
- Request rate
- Success/failure rate
- Response times
- Faucet balance
- Database connections
- Redis memory usage

### Log Aggregation

#### Using ELK Stack

```yaml
# filebeat.yml
filebeat.inputs:
- type: container
  paths:
    - '/var/lib/docker/containers/*/*.log'
  processors:
    - add_docker_metadata: ~

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
```

#### Using Loki

```yaml
# promtail-config.yaml
clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
```

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/api/v1/health

# Detailed health monitoring
while true; do
  curl -s http://localhost:8080/api/v1/health | jq
  sleep 30
done
```

## Backup and Recovery

### Database Backup

#### Automated Daily Backups

```bash
# backup.sh
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
docker-compose exec -T postgres pg_dump -U faucet faucet > "$BACKUP_DIR/faucet_$DATE.sql"
# Compress backup
gzip "$BACKUP_DIR/faucet_$DATE.sql"
# Keep only last 30 days
find $BACKUP_DIR -name "faucet_*.sql.gz" -mtime +30 -delete
```

#### Cron Job Setup

```bash
# Add to crontab
crontab -e

# Run daily at 2 AM
0 2 * * * /path/to/backup.sh
```

### Database Restore

```bash
# Restore from backup
gunzip < backup_20251120.sql.gz | docker-compose exec -T postgres psql -U faucet faucet
```

### Redis Backup

```bash
# Save Redis snapshot
docker-compose exec redis redis-cli SAVE

# Copy snapshot
docker cp aura-faucet-redis:/data/dump.rdb ./redis-backup.rdb
```

## Troubleshooting

### Common Issues and Solutions

#### Issue: Cannot connect to database

**Symptoms**:
```
Failed to connect to database: dial tcp 127.0.0.1:5432: connect: connection refused
```

**Solutions**:
1. Check PostgreSQL is running:
   ```bash
   docker-compose ps postgres
   ```

2. Check database logs:
   ```bash
   docker-compose logs postgres
   ```

3. Verify DATABASE_URL:
   ```bash
   echo $DATABASE_URL
   ```

4. Test connection manually:
   ```bash
   docker-compose exec postgres psql -U faucet -d faucet
   ```

#### Issue: Transactions failing

**Symptoms**:
- Transactions return errors
- "Insufficient funds" errors

**Solutions**:
1. Check faucet balance:
   ```bash
   curl http://localhost:8080/api/v1/faucet/info | jq .balance
   ```

2. Verify node is synced:
   ```bash
   curl http://localhost:26657/status | jq .result.sync_info
   ```

3. Check faucet address:
   ```bash
   echo $FAUCET_ADDRESS
   ```

#### Issue: High memory usage

**Solutions**:
1. Check Redis memory:
   ```bash
   docker-compose exec redis redis-cli INFO memory
   ```

2. Clear old rate limit keys:
   ```bash
   docker-compose exec redis redis-cli --scan --pattern "ratelimit:*" | head -1000 | xargs docker-compose exec redis redis-cli DEL
   ```

3. Optimize database:
   ```bash
   docker-compose exec postgres psql -U faucet -d faucet -c "VACUUM ANALYZE;"
   ```

### Performance Tuning

#### Database Connection Pool

Update in code:
```go
conn.SetMaxOpenConns(100)
conn.SetMaxIdleConns(25)
conn.SetConnMaxLifetime(5 * time.Minute)
```

#### Rate Limiting Optimization

Adjust for your load:
```env
RATE_LIMIT_PER_IP=20
RATE_LIMIT_PER_ADDRESS=2
RATE_LIMIT_WINDOW_HOURS=12
```

#### Caching

Add caching for faucet info:
```go
// Cache faucet balance for 1 minute
```

## Support

For deployment issues:
1. Check logs: `docker-compose logs -f`
2. Review configuration
3. Check GitHub issues
4. Contact AURA team

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-20
