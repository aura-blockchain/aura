# Prometheus Monitoring Stack Deployment

## Deployed Resources

### 1. Monitoring Namespace
- **Name**: `monitoring`
- **Status**: Active

### 2. Prometheus Server
- **Deployment**: `prometheus` (1 replica)
- **Image**: `prom/prometheus:v2.54.1`
- **Service**: ClusterIP on port 9090
- **Storage**: EmptyDir (ephemeral)
- **Resource Limits**: 500m CPU, 512Mi memory

### 3. RBAC Configuration
- **ServiceAccount**: `prometheus` in monitoring namespace
- **ClusterRole**: `prometheus` with permissions to discover and scrape pods/services
- **ClusterRoleBinding**: Links ServiceAccount to ClusterRole

### 4. ServiceMonitor CRD
- **CRD**: `servicemonitors.monitoring.coreos.com`
- **Instance**: `aura-nodes` in aura namespace
- **Target**: Pods with label `app=aura` on port `prometheus`

### 5. Network Policies
- **monitoring namespace**: `allow-monitoring-ingress` allows ingress from aura namespace on port 9090
- **aura namespace**:
  - `allow-monitoring` allows ingress from monitoring namespace on port 26660 (metrics scraping)
  - `allow-monitoring-scrape` allows ingress from monitoring namespace
  - `allow-egress-to-monitoring-general` allows egress to monitoring namespace on port 9090 (health checks)

## Prometheus Configuration

Scrapes configured for:
- Kubernetes API servers
- Kubernetes nodes
- Kubernetes pods (with prometheus.io/scrape annotation)
- Aura blockchain nodes (dedicated job)

## Verification

```bash
export KUBECONFIG=/tmp/kind-aura.kubeconfig

# Check Prometheus health
kubectl exec -n monitoring deployment/prometheus -- wget -q -O- http://localhost:9090/-/ready

# Check ServiceMonitor
kubectl get servicemonitor -n aura

# View Prometheus targets
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Then visit: http://localhost:9090/targets
```

## Files Created
- `/home/hudson/blockchain-projects/aura/k8s/prometheus-setup.yaml` - Main deployment manifest
- `/home/hudson/blockchain-projects/aura/k8s/servicemonitor-crd.yaml` - ServiceMonitor CRD and instance
- `/home/hudson/blockchain-projects/aura/k8s/allow-egress-to-monitoring.yaml` - Egress policy for monitoring access
- `/home/hudson/blockchain-projects/aura/scripts/prometheus-verify.sh` - Verification script

## Status
✓ Prometheus pod running and healthy
✓ Service accessible within cluster
✓ ServiceMonitor CRD installed
✓ ServiceMonitor instance created in aura namespace
✓ Network policies configured for monitoring access
✓ Network policy tests passing (7 passed, 0 failed, 1 skipped)
