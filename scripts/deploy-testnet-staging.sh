#!/usr/bin/env bash
set -euo pipefail

SCRIPT_NAME="$(basename "$0")"

function ensure_context() {
  if ! kubectl config current-context >/dev/null 2>&1; then
    echo >&2 "[${SCRIPT_NAME}] no Kubernetes context configured; please set KUBECONFIG before deploying"
    exit 1
  fi
}

function ensure_namespace() {
  # Create namespace if missing
  kubectl get namespace aura-testnet >/dev/null 2>&1 || \
    kubectl create namespace aura-testnet
}

ensure_context
ensure_namespace

kubectl apply -k k8s/overlays/staging/ --namespace aura-testnet --validate=false

cat <<'EOF'
Deployment initiated. Ensure the following prerequisites are in place:
- ConfigMap `aura-config` (genesis/config/app toml) is valid for aura-testnet-1.
- Secret `aura-testnet-tls` contains TLS certs for api.testnet.aura.network, rpc.testnet.aura.network, grpc.testnet.aura.network.
- The `aequitas/aura:testnet` image includes the latest `aurad` binary built with genesis helpers.

After apply:
  kubectl -n aura-testnet get pods
  kubectl -n aura-testnet logs deployment/aura-node
  kubectl -n aura-testnet port-forward svc/aura-node-api 1317:1317
Then run `curl https://api.testnet.aura.network/health` once DNS + Cloudflare are configured.
EOF
