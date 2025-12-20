#!/bin/bash
# Verify Prometheus monitoring stack in KinD cluster
# Usage: ./prometheus-verify.sh

set -e

export KUBECONFIG=/tmp/kind-aura.kubeconfig

echo "=== Prometheus Monitoring Stack Verification ==="
echo

# Check namespace
echo "1. Monitoring Namespace:"
kubectl get namespace monitoring --no-headers && echo "   ✓ Namespace exists" || echo "   ✗ Namespace not found"
echo

# Check deployment
echo "2. Prometheus Deployment:"
if kubectl get deployment -n monitoring prometheus &>/dev/null; then
    READY=$(kubectl get deployment -n monitoring prometheus --no-headers | awk '{print $2}')
    echo "   ✓ Deployment found (Ready: $READY)"
else
    echo "   ✗ Deployment not found"
fi
echo

# Check pod
echo "3. Prometheus Pod:"
POD_STATUS=$(kubectl get pods -n monitoring -l app=prometheus --no-headers 2>/dev/null | awk '{print $3}')
if [ "$POD_STATUS" = "Running" ]; then
    echo "   ✓ Pod running"
else
    echo "   ✗ Pod not running (Status: ${POD_STATUS:-Not found})"
fi
echo

# Check service
echo "4. Prometheus Service:"
if kubectl get svc -n monitoring prometheus &>/dev/null; then
    CLUSTER_IP=$(kubectl get svc -n monitoring prometheus --no-headers | awk '{print $3}')
    echo "   ✓ Service found (ClusterIP: $CLUSTER_IP)"
else
    echo "   ✗ Service not found"
fi
echo

# Check health
echo "5. Health Checks:"
if kubectl exec -n monitoring deployment/prometheus -- wget -q -O- http://localhost:9090/-/healthy 2>/dev/null | grep -q "Healthy"; then
    echo "   ✓ Health check passed"
else
    echo "   ✗ Health check failed"
fi

if kubectl exec -n monitoring deployment/prometheus -- wget -q -O- http://localhost:9090/-/ready 2>/dev/null | grep -q "Ready"; then
    echo "   ✓ Readiness check passed"
else
    echo "   ✗ Readiness check failed"
fi
echo

# Check ServiceMonitor CRD
echo "6. ServiceMonitor CRD:"
if kubectl get crd servicemonitors.monitoring.coreos.com &>/dev/null; then
    echo "   ✓ CRD installed"
else
    echo "   ✗ CRD not found"
fi
echo

# Check ServiceMonitor instance
echo "7. ServiceMonitor in aura namespace:"
if kubectl get servicemonitor -n aura aura-nodes &>/dev/null; then
    echo "   ✓ ServiceMonitor exists"
else
    echo "   ✗ ServiceMonitor not found"
fi
echo

# Check RBAC
echo "8. RBAC Configuration:"
kubectl get sa -n monitoring prometheus &>/dev/null && echo "   ✓ ServiceAccount exists" || echo "   ✗ ServiceAccount not found"
kubectl get clusterrole prometheus &>/dev/null && echo "   ✓ ClusterRole exists" || echo "   ✗ ClusterRole not found"
kubectl get clusterrolebinding prometheus &>/dev/null && echo "   ✓ ClusterRoleBinding exists" || echo "   ✗ ClusterRoleBinding not found"
echo

# Check network policies
echo "9. Network Policies:"
kubectl get networkpolicy -n monitoring allow-monitoring-ingress &>/dev/null && echo "   ✓ Monitoring ingress policy exists" || echo "   ✗ Monitoring ingress policy not found"
kubectl get networkpolicy -n aura allow-monitoring &>/dev/null && echo "   ✓ Aura monitoring policy exists" || echo "   ✗ Aura monitoring policy not found"
echo

echo "=== Verification Complete ==="
echo
echo "To access Prometheus UI:"
echo "  kubectl port-forward -n monitoring svc/prometheus 9090:9090"
echo "  Then visit: http://localhost:9090"
