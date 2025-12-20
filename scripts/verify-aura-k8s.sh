#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-aura}"
LINKERD_NS="${LINKERD_NS:-$NAMESPACE}"

echo "=== Aura Kubernetes Deployment Verification (${NAMESPACE}) ==="

step() {
  printf "\n[%s] %s\n" "$1" "$2"
}

step "1/12" "Checking cluster nodes..."
kubectl get nodes

step "2/12" "Checking namespace..."
kubectl get ns "$NAMESPACE"

step "3/12" "Checking pods..."
kubectl get pods -n "$NAMESPACE" -o wide

step "4/12" "Checking services..."
kubectl get svc -n "$NAMESPACE"

step "5/12" "Checking persistent volumes..."
kubectl get pvc -n "$NAMESPACE"

step "6/12" "Checking external secrets..."
kubectl get externalsecret -n "$NAMESPACE"

step "7/12" "Checking Linkerd mesh..."
if command -v linkerd >/dev/null 2>&1; then
  linkerd check --proxy -n "$LINKERD_NS" || echo "Linkerd proxy check failed (mesh issues?)"
else
  echo "linkerd CLI not found"
fi

step "8/12" "Checking network policies..."
kubectl get networkpolicy -n "$NAMESPACE" || true

step "9/12" "Checking autoscaling..."
kubectl get hpa -n "$NAMESPACE" || true
kubectl get vpa -n "$NAMESPACE" || true

# RPC/REST probe via service to validate in-cluster reachability
step "10/12" "Probing RPC/REST services (ClusterIP)..."
RPC_PROBE_POD="aura-rpc-probe"
kubectl delete pod "$RPC_PROBE_POD" -n "$NAMESPACE" --ignore-not-found >/dev/null 2>&1 || true
cat <<EOF | kubectl apply -f - >/dev/null
apiVersion: v1
kind: Pod
metadata:
  name: ${RPC_PROBE_POD}
  namespace: ${NAMESPACE}
  annotations:
    linkerd.io/inject: "disabled"
spec:
  restartPolicy: Never
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    runAsGroup: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: probe
    image: curlimages/curl:8.11.1
    command:
      - /bin/sh
      - -c
      - >
        echo "RPC /status" &&
        curl -sS --max-time 5 http://aura-rpc:26657/status 2>&1 | head -10 || echo "RPC probe failed"; echo "---";
        echo "REST /health" &&
        curl -sS --max-time 5 http://aura-api:1317/health 2>&1 | head -10 || echo "REST probe failed";
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
      runAsNonRoot: true
      runAsUser: 1000
      runAsGroup: 1000
EOF
kubectl wait --for=condition=Ready --timeout=30s pod/"$RPC_PROBE_POD" -n "$NAMESPACE" >/dev/null 2>&1 || true
kubectl logs "$RPC_PROBE_POD" -n "$NAMESPACE" || true
kubectl delete pod "$RPC_PROBE_POD" -n "$NAMESPACE" --ignore-not-found >/dev/null 2>&1 || true
# Fallback: probe inside existing validator sidecar (rpc-probe) if present
FIRST_VALIDATOR=$(kubectl get pod -n "$NAMESPACE" -l app=aura,component=validator -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
if [[ -n "$FIRST_VALIDATOR" ]]; then
  echo "Validator in-pod probe ($FIRST_VALIDATOR):"
  kubectl exec "$FIRST_VALIDATOR" -n "$NAMESPACE" -c rpc-probe -- sh -c 'echo "RPC /status (loopback)"; curl -s --max-time 3 http://127.0.0.1:26657/status | head -5 && echo "---"; echo "REST /health (loopback)"; curl -s --max-time 3 http://127.0.0.1:1317/health | head -5 && echo "---"; echo "RPC via service (aura-rpc)"; curl -s --max-time 3 http://aura-rpc:26657/status | head -5 && echo "---"; echo "REST via service (aura-api)"; curl -s --max-time 3 http://aura-api:1317/health | head -5' || echo "  rpc-probe container unavailable"
fi

step "11/12" "Checking validator health..."
for i in 0 1 2 3; do
  POD="aura-validator-$i"
  if kubectl get pod "$POD" -n "$NAMESPACE" &>/dev/null; then
    echo "- $POD status:"
    STATUS_OUTPUT=$(kubectl exec "$POD" -n "$NAMESPACE" -c aura -- sh -c 'aurad status --home /home/aura/.aura --node tcp://127.0.0.1:26657 --output json' 2>&1) || STATUS_RC=$?
    if [[ ${STATUS_RC:-0} -eq 0 ]]; then
      echo "$STATUS_OUTPUT" | head -5
    else
      echo "  aurad status failed (rc=${STATUS_RC:-1}); attempting HTTP /status..."
      HTTP_OUTPUT=$(kubectl exec "$POD" -n "$NAMESPACE" -c aura -- sh -c 'if command -v curl >/dev/null 2>&1; then curl -s http://127.0.0.1:26657/status; elif command -v wget >/dev/null 2>&1; then wget -qO- http://127.0.0.1:26657/status; fi' 2>/dev/null) || true
      if [[ -n "$HTTP_OUTPUT" ]]; then
        echo "  RPC /status response:"
        echo "$HTTP_OUTPUT" | head -10
      fi
      echo "  falling back to node ID:"
      kubectl exec "$POD" -n "$NAMESPACE" -c aura -- aurad tendermint show-node-id 2>/dev/null || echo "  node-id check failed"
    fi
  fi
done

step "12/12" "Capturing recent validator logs..."
for i in 0 1 2 3; do
  POD="aura-validator-$i"
  if kubectl get pod "$POD" -n "$NAMESPACE" &>/dev/null; then
    echo "- $POD logs (last 10 lines):"
    kubectl logs "$POD" -n "$NAMESPACE" -c aura --tail=10 || true
  fi
done

echo -e "\n=== Verification Complete ==="
